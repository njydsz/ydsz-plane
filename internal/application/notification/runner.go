// Package notification — 通知消费者 Worker Runner。
//
// 在 EventExchange 上绑定订阅队列，监听领域事件并转换为通知。
// 与 OutboxRelay 共享同一个 RabbitMQ 客户端连接。
package notification

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

// notifQueue 是通知消费者绑定的队列名。
const notifQueue = "notifications.inbox"

// routingPattern 订阅所有领域事件。
// HandleEvent 内部会过滤出与通知相关的事件类型。
// 使用通配符而非精确匹配，便于后续扩展新事件类型而无需修改绑定。
const routingPattern = "plane.events.#"

// RunConsumer 启动阻塞型通知消费者循环。
// 当 ctx 取消时优雅退出；连接断开时自动重连。
func RunConsumer(ctx context.Context, mqClient *mq.Client, db *pgxpool.Pool, log *zap.Logger) {
	log.Info("notification consumer: starting",
		zap.String("queue", notifQueue),
		zap.String("exchange", mq.EventExchange))

	cons := newConsumer(db, log)

	for {
		select {
		case <-ctx.Done():
			log.Info("notification consumer: stopped")
			return
		default:
		}

		if err := runConsumeLoop(ctx, mqClient, cons, log); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Info("notification consumer: stopped (context)")
				return
			}
			log.Warn("notification consumer: connection lost, retrying",
				zap.Error(err), zap.Duration("backoff", 2*time.Second))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// runConsumeLoop 单次消费循环：声明队列 → 消费。
// 返回非 nil 错误时触发外层重连逻辑。
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, cons *consumer, log *zap.Logger) error {
	// 声明队列并绑定通配符路由
	if _, err := mqClient.DeclareQueue(ctx, notifQueue, mq.EventExchange, routingPattern, amqp.Table{
		"x-max-priority": int64(5),
	}); err != nil {
		return errors.New("notification: declare queue: " + err.Error())
	}

	// 开始消费
	return mqClient.Consume(ctx, notifQueue, "notif-consumer", false, func(delivery amqp.Delivery) error {
		var event mq.EventEnvelope
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			log.Warn("notification consumer: bad payload, skipping",
				zap.Error(err))
			// 不能解析的消息直接 ACK 丢弃，避免死信循环
			return nil
		}

		if err := cons.HandleEvent(ctx, event); err != nil {
			log.Warn("notification consumer: handle failed",
				zap.String("type", event.EventType),
				zap.Error(err))
			return err // NACK → 重试
		}
		return nil // ACK
	})
}
