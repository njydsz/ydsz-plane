// Package automation — 自动化规则引擎事件消费者 Runner。
//
// 在 EventExchange 上绑定订阅队列，监听领域事件并执行匹配的自动化规则。
// 与 OutboxRelay 共享同一个 RabbitMQ 客户端连接。
//
// 设计要点:
//   - 直接在 EventExchange 上消费，不绕经 TaskExchange（减少一跳延迟）
//   - 错误隔离：单条事件处理失败不影响后续事件（最多重试 MaxRetries 次）
//   - 幂等：引擎内部通过 rule_executions 表的触发事件 ID 防重
//   - 自动重连：连接断开后指数退避重连
package automation

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// consumerQueue 是自动化消费者绑定的队列名。
const consumerQueue = "automation.evaluate"

// routingPattern 订阅所有领域事件。
// Engine.EvaluateEvent 内部会过滤出配置了对应触发器的规则。
const routingPattern = "plane.events.#"

// RunConsumer 启动阻塞型自动化消费者循环。
// 当 ctx 取消时优雅退出；连接断开时自动重连。
//
// 参数:
//   - mqClient: RabbitMQ 客户端（与 OutboxRelay 共享连接）
//   - db: PostgreSQL 连接池（用于加载上下文/规则）
//   - log: 结构化日志
//
// 调用方应在独立 goroutine 中运行：
//
//	go automation.RunConsumer(ctx, mqClient, pool.Pool, log)
func RunConsumer(ctx context.Context, mqClient *mq.Client, db *pgxpool.Pool, log *zap.Logger) {
	log.Info("automation consumer: starting",
		zap.String("queue", consumerQueue),
		zap.String("exchange", mq.EventExchange))

	eng := newEngine(db, log)

	for {
		select {
		case <-ctx.Done():
			log.Info("automation consumer: stopped")
			return
		default:
		}

		if err := runConsumeLoop(ctx, mqClient, eng, log); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Info("automation consumer: stopped (context)")
				return
			}
			log.Warn("automation consumer: connection lost, retrying",
				zap.Error(err), zap.Duration("backoff", 2*time.Second))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// newEngine 创建默认的执行引擎（含内置 SPI 实现）。
func newEngine(db *pgxpool.Pool, log *zap.Logger) *Engine {
	svc := NewService(db)
	prov := NewDefaultContextProvider(db)
	exec := newActionExecutor(db)
	eng := NewEngine(svc, exec, prov, log)
	eng.db = db
	return eng
}

// runConsumeLoop 单次消费循环：声明队列 → 消费。
// 返回非 nil 错误时触发外层重连逻辑。
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, eng *Engine, log *zap.Logger) error {
	// 声明队列并绑定通配符路由
	if _, err := mqClient.DeclareQueue(ctx, consumerQueue, mq.EventExchange, routingPattern, amqp.Table{
		"x-max-priority": int64(5),
		"x-dead-letter-exchange": mq.DeadLetterExchange,
	}); err != nil {
		return errors.New("automation: declare queue: " + err.Error())
	}

	// 开始消费
	return mqClient.Consume(ctx, consumerQueue, "automation-consumer", false, func(delivery amqp.Delivery) error {
		var event mq.EventEnvelope
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			log.Warn("automation consumer: bad payload, skipping",
				zap.Error(err))
			// 不能解析的消息直接 ACK 丢弃，避免死信循环
			return nil
		}

		if err := eng.EvaluateEvent(ctx, event); err != nil {
			log.Warn("automation consumer: handle failed",
				zap.String("event_type", event.EventType),
				zap.Error(err))
			return err // NACK → 重试（受 MaxRetries 限制）
		}
		return nil // ACK
	})
}
