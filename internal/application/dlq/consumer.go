// Package dlq — RabbitMQ 死信队列持久化消费者（生产者侧）。
//
// 消费 plane.dlx 交换上的死信消息（重试耗尽的任务/事件），把元数据写入
// dlq_events 表，供管理页查询与重试。若持久化失败仅记日志（不 NACK，
// 避免死信再次进入死信形成循环）。
package dlq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// RunDLXConsumer 启动阻塞型死信持久化消费者，ctx 取消时优雅退出。
// 应在 cmd/worker/main.go 中以独立 goroutine 调用。
func RunDLXConsumer(ctx context.Context, mqClient *mq.Client, db *pgxpool.Pool, log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}
	log.Info("dlq consumer: starting",
		zap.String("exchange", mq.DeadLetterExchange),
		zap.String("queue", mq.DeadLetterExchange+".queue"))

	svc := NewService(db)

	for {
		select {
		case <-ctx.Done():
			log.Info("dlq consumer: stopped")
			return
		default:
		}

		if err := runConsumeLoop(ctx, mqClient, svc, log); err != nil {
			log.Warn("dlq consumer: connection lost, retrying", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// runConsumeLoop 单次消费循环：声明队列 → 消费 → 持久化。
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, svc *Service, log *zap.Logger) error {
	queue := mq.DeadLetterExchange + ".queue"
	if _, err := mqClient.DeclareQueue(ctx, queue, mq.DeadLetterExchange, "", amqp.Table{}); err != nil {
		return err
	}

	return mqClient.Consume(ctx, queue, "dlq-persist", true, func(delivery amqp.Delivery) error {
		if err := persistDeadLetter(ctx, svc, delivery); err != nil {
			// 持久化失败：记日志并 ACK（避免死信重入死信死循环）；
			// 管理面可通过 RabbitMQ 原生控制台兜底。
			log.Warn("dlq consumer: persist failed",
				zap.String("routing_key", delivery.RoutingKey), zap.Error(err))
		}
		return nil // ACK
	})
}

// persistDeadLetter 提取死信元数据并写入 dlq_events。
func persistDeadLetter(ctx context.Context, svc *Service, delivery amqp.Delivery) error {
	var workspaceID int64
	var eventID any // 事件型死信记 event_id；任务型为 NULL（UNIQUE 约束对 NULL 不冲突）

	// 尝试解析为事件信封（EventExchange 死信）
	var env mq.EventEnvelope
	if err := json.Unmarshal(delivery.Body, &env); err == nil && env.EventType != "" {
		workspaceID = env.WorkspaceID
		eventID = env.EventID
	} else {
		// 尝试解析为任务（TaskExchange 死信）：payload 内嵌 workspace_id
		var task mq.Task
		if err := json.Unmarshal(delivery.Body, &task); err == nil {
			workspaceID = extractWorkspaceID(task.Payload)
		}
	}

	queue := delivery.RoutingKey
	if queue == "" {
		queue = "unknown"
	}
	exchange := delivery.Exchange
	if exchange == "" {
		exchange = mq.DeadLetterExchange
	}

	_, err := svc.db.Exec(ctx, `
		INSERT INTO dlq_events (event_id, workspace_id, queue, exchange, routing_key, payload, error_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (event_id, queue) DO NOTHING`,
		eventID, workspaceID, queue, exchange, delivery.RoutingKey,
		delivery.Body, truncateReason(delivery.Headers))
	if err != nil {
		return err
	}
	return nil
}

// extractWorkspaceID 从任务 payload 中提取 workspace_id。
func extractWorkspaceID(payload []byte) int64 {
	var p struct {
		WorkspaceID int64 `json:"workspace_id"`
	}
	_ = json.Unmarshal(payload, &p)
	return p.WorkspaceID
}

// truncateReason 从投递头提取死信原因（x-death 首条 reason），截断至 500 字符。
func truncateReason(headers amqp.Table) string {
	if headers == nil {
		return ""
	}
	if deaths, ok := headers["x-death"].([]any); ok && len(deaths) > 0 {
		if first, ok := deaths[0].(map[string]any); ok {
			if reason, ok := first["reason"].(string); ok {
				return truncate(reason, 500)
			}
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
