// 本文件属于 Package mq：基于 RabbitMQ 的任务队列实现。
// 提供 Task 定义、TaskClient（入队）与 Worker（消费分发），
// 替代 M0.5/S1 阶段的 Redis 版 Asynq 语义。
package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ----- 处理器回调 -----

// TaskHandler 处理单条投递。返回错误表示让 Worker 执行
// NACK 并重新入队（受每个任务 MaxRetries 限制）。
type TaskHandler func(ctx context.Context, task Task) error

// ----- 任务定义 -----

// Task 是提交到 RabbitMQ 任务队列的一个异步工作单元。
// 构建在 Outbox 事件总线使用的同一套 RabbitMQ 原语之上：
//   - 通过 per-message TTL + DLX 回环实现延迟分发。
//   - 通过队列级 `x-max-priority` header 实现优先级选择。
//   - 封顶指数退避重试。
//   - 对重试耗尽/永久失败的任务做死信路由。
//
// 该类型替代 M0.5 / S1 阶段基于 Redis 的 Asynq 语义。
type Task struct {
	// ID 用于去重与日志的唯一标识。Enqueue 时若为空则自动生成。
	ID string `json:"id"`

	// Type 选择处理器。与队列名一一对应（`task.<Type>`），
	// 以路由键 `task.<Type>` 绑定到 TaskExchange。
	Type string `json:"type"`

	// Payload 是不透明的任务数据。处理器自行解码所需字段。
	Payload json.RawMessage `json:"payload"`

	// Priority 决定同队列待处理任务间的分发顺序。
	// 范围 0-255，越大越先分发。默认 0（无优先级）。
	// 接收方必须在其队列上声明 `x-max-priority`。
	Priority uint8 `json:"priority"`

	// MaxRetries 限制处理器失败后的自动重新入队次数。第 N 次失败后
	// 任务不再入队，而是走队列的死信路由。默认 0 —— 不重试。
	MaxRetries int `json:"max_retries"`

	// RetryCount 由 Worker 在每次重新入队时递增。初始为 0。
	RetryCount int `json:"retry_count"`

	// Delay 是任务变为可分发前的最小等待时长。底层机制是 per-message
	// TTL：任务先经 DeadLetterExchange 路由到临时中转，TTL 到期后
	// 弹回任务队列。受 RabbitMQ 策略限制最大 24 小时。
	Delay time.Duration `json:"delay"`

	// CreatedAt 是任务提交时间戳，用于指标统计。
	CreatedAt time.Time `json:"created_at"`
}

// RoutingKey 返回该任务类型的 AMQP 路由键。
func (t Task) RoutingKey() string {
	return "task." + t.Type
}

// IsZero 报告 Task 是否为空占位。
func (t Task) IsZero() bool { return t.Type == "" }

// ----- TaskClient -----

// TaskClient 向 RabbitMQ 任务队列提交 Task 负载。
// 并发安全：底层 Client 串行化发布。
type TaskClient struct {
	client *Client
	log    *zap.Logger
	mu     sync.Mutex
}

// NewTaskClient 包装既有 *Client 用于任务提交。
// 底层 Client 的关闭仍由调用方负责。
func NewTaskClient(client *Client, log *zap.Logger) *TaskClient {
	if log == nil {
		log = zap.NewNop()
	}
	return &TaskClient{client: client, log: log}
}

// Enqueue 将任务提交到其队列。task.ID 为空时自动生成。
// 队列在首次使用时惰性声明（幂等）。
func (tc *TaskClient) Enqueue(ctx context.Context, task Task) error {
	if task.Type == "" {
		return errors.New("mq/task: task type is required")
	}
	if task.ID == "" {
		task.ID = fmt.Sprintf("%s-%d", task.Type, time.Now().UnixNano())
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 确保队列存在（幂等）。
	if err := tc.ensureQueue(ctx, task); err != nil {
		return err
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("mq/task: marshal: %w", err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    task.CreatedAt,
		MessageId:    task.ID,
		Type:         task.Type,
		Body:         body,
		Headers: amqp.Table{
			"max_retries": task.MaxRetries,
		},
	}
	if task.Priority > 0 {
		msg.Priority = task.Priority
	}
	if task.Delay > 0 {
		if task.Delay > 24*time.Hour {
			return fmt.Errorf("mq/task: delay exceeds 24h maximum (%s)", task.Delay)
		}
		// Per-message TTL（需 RabbitMQ ≥ 3.11）。到期后消息路由到队列的 DLX；
		// 任务队列将 DLX 反弹路由键绑定到自身 → 延迟结束后消息回到队首。
		msg.Expiration = fmt.Sprintf("%d", task.Delay.Milliseconds())
	}

	if err := tc.client.PublishRaw(ctx, TaskExchange, task.RoutingKey(), msg); err != nil {
		return fmt.Errorf("mq/task: enqueue: %w", err)
	}

	tc.log.Debug("mq/task: enqueued",
		zap.String("id", task.ID),
		zap.String("type", task.Type),
		zap.String("queue", queueName(task.Type)),
		zap.Uint8("priority", task.Priority),
		zap.Duration("delay", task.Delay),
	)
	return nil
}

// ensureQueue 惰性声明队列并绑定到 TaskExchange。
func (tc *TaskClient) ensureQueue(ctx context.Context, task Task) error {
	qName := queueName(task.Type)
	if tc.client.QueueExists(ctx, qName) {
		return nil
	}
	_, err := tc.client.DeclareQueue(ctx, qName, TaskExchange, task.RoutingKey(), amqp.Table{
		"x-max-priority": int64(10),
	})
	return err
}

// ----- Worker -----

// Worker 从 RabbitMQ 队列消费 Task 投递并分发给已注册的处理器。
// 它替代 asynq.ServeMux 用于 Ydsz Plane 任务子系统，
// 同时共享 Outbox Relay 已验证的连接模型。
type Worker struct {
	client   *Client
	log      *zap.Logger
	handlers map[string]TaskHandler
	mu       sync.RWMutex
}

// NewWorker 将已连接的 *Client 包装为任务队列消费者。
func NewWorker(client *Client, log *zap.Logger) *Worker {
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{client: client, log: log, handlers: make(map[string]TaskHandler)}
}

// Register 为指定任务类型注册处理器。重复注册同类型会覆盖旧处理器。
func (w *Worker) Register(taskType string, handler TaskHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[taskType] = handler
}

// QueueNames 返回当前全部已注册处理器对应的队列名。
func (w *Worker) QueueNames() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	qs := make([]string, 0, len(w.handlers))
	for t := range w.handlers {
		qs = append(qs, queueName(t))
	}
	return qs
}

// Start 为每个已注册队列启动一个消费者 goroutine 并阻塞至 ctx 取消。
// 取消时每个消费者尽可能完成当前投递后返回。若消费者因连接丢失退出，
// Start 携带该错误返回；Client.reconnectLoop 会重建连接，
// 调用方应再次调用 Start。
//
// 幂等检查：通过取消传给 Start 的 ctx 来停止。
func (w *Worker) Start(ctx context.Context) error {
	names := w.QueueNames()
	if len(names) == 0 {
		w.log.Warn("mq/task: no task handlers registered; worker idle")
		<-ctx.Done()
		return ctx.Err()
	}

	w.log.Info("mq/task: worker starting", zap.Int("queues", len(names)))

	errCh := make(chan error, len(names))
	for _, q := range names {
		go func(queue string) {
			errCh <- w.runConsumer(ctx, queue)
		}(q)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (w *Worker) runConsumer(ctx context.Context, queue string) error {
	for {
		err := w.client.Consume(ctx, queue, DefaultConsumerTag+"-"+queue, false, func(delivery amqp.Delivery) error {
			var task Task
			if err := json.Unmarshal(delivery.Body, &task); err != nil {
				w.log.Warn("mq/task: unmarshal error, routing to DLQ",
					zap.Error(err), zap.String("queue", queue))
				// 无法解析 → 不重新入队（返回错误 → NACK）。
				return fmt.Errorf("mq/task: bad payload: %w", err)
			}

			w.mu.RLock()
			handler, ok := w.handlers[task.Type]
			w.mu.RUnlock()
			if !ok {
				w.log.Warn("mq/task: no handler registered",
					zap.String("type", task.Type), zap.String("queue", queue))
				return nil // 防御性 ACK —— 正常情况下不应发生
			}

			if err := handler(ctx, task); err != nil {
				return w.handleFailure(ctx, task, err)
			}
			return nil // 成功即 ACK
		})
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return err
		}
		w.log.Warn("mq/task: consumer dropped, will reconnect",
			zap.String("queue", queue), zap.Error(err))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// handleFailure 决定失败任务是重试还是落入死信队列。
func (w *Worker) handleFailure(ctx context.Context, task Task, handlerErr error) error {
	w.log.Warn("mq/task: handler failed",
		zap.String("type", task.Type),
		zap.String("id", task.ID),
		zap.Int("retry", task.RetryCount),
		zap.Int("max_retries", task.MaxRetries),
		zap.Error(handlerErr),
	)

	if task.RetryCount >= task.MaxRetries {
		// 无剩余重试 → 允许 NACK 路由到 DLX。
		return fmt.Errorf("mq/task: %s exhausted retries: %w", task.Type, handlerErr)
	}

	// 以递增的 RetryCount 与指数退避重新入队。
	task.RetryCount++
	backoff := retryBackoff(task.RetryCount)
	task.Delay = backoff
	task.ID = fmt.Sprintf("%s-r%d-%d", task.Type, task.RetryCount, time.Now().UnixNano())

	tc := NewTaskClient(w.client, w.log)
	if err := tc.Enqueue(ctx, task); err != nil {
		w.log.Error("mq/task: retry enqueue failed", zap.Error(err))
	}
	// 返回原始错误 → 将耗尽投递的这条 NACK 且不重新入队
	// （重试路径已生成带独立信封的新消息）。
	return handlerErr
}

// retryBackoff 计算封顶指数退避：2^attempt 秒，最大 30s。
func retryBackoff(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

// ----- 辅助函数 -----

// queueName 从任务类型推导持久队列标识。
func queueName(taskType string) string {
	return "task." + taskType
}
