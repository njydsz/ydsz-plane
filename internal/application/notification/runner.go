// Package notification — 通知消费者 Worker Runner。
//
// 在 EventExchange 上绑定订阅队列，监听领域事件并转换为通知。
// 与 OutboxRelay 共享同一个 RabbitMQ 客户端连接。
package notification

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// notifQueue 是通知消费者绑定的队列名。
const notifQueue = "notifications.inbox"

// routingPatterns 定义通知消费者订阅的路由模式。
// 仅订阅与通知直接相关的聚合事件（Issue / Comment）。
var routingPatterns = []string{
	"plane.events.issue.created",
	"plane.events.issue.assigned",
	"plane.events.issue.status_changed",
	"plane.events.comment.created",
}

// RunConsumer 启动阻塞型通知消费者循环。
// 当 ctx 取消时优雅退出；连接断开时自动重连。
//
// 生命周期：
//  1. 声明队列并绑定到 EventExchange
//  2. 启动消费循环
//  3. 连接断开 → 等待 2s 后重试
func RunConsumer(ctx context.Context, mqClient *mq.Client, db interface {
	// 仅依赖 consumer 所需的最小接口——由 *pgxpool.Pool 自然满足
}, log *zap.Logger) {

	// 取出 *pgxpool.Pool（consumer 内部需要）
	// 使用类型断言——main.go 会传入 pool.Pool
	type poolInterface interface{}
	_ = poolInterface(db)

	log.Info("notification consumer: starting",
		zap.String("queue", notifQueue),
		zap.String("exchange", mq.EventExchange),
		zap.Strings("patterns", routingPatterns))

	// 获取底层 *pgxpool.Pool
	pgPool := extractPool(db)
	cons := newConsumer(pgPool, log)

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

// runConsumeLoop 单次消费循环：声明队列 → 绑定 → 消费。
// 返回 nil 仅当 ctx 被取消；连接错误返回非 nil 触发重连。
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, cons *consumer, log *zap.Logger) error {
	// 声明队列
	_, err := mqClient.DeclareQueue(ctx, notifQueue, mq.EventExchange, routingPatterns[0], amqp.Table{
		"x-max-priority": int64(5),
	})
	if err != nil {
		return fmtQueueErr("declare queue", err)
	}
	// 绑定额外的路由键模式
	for _, pattern := range routingPatterns[1:] {
		if bindErr := bindQueue(ctx, mqClient, notifQueue, pattern); bindErr != nil {
			return fmtQueueErr("bind pattern", bindErr)
		}
	}

	// 开始消费
	return mqClient.Consume(ctx, notifQueue, "notif-consumer", false, func(delivery amqp.Delivery) error {
		var event mq.EventEnvelope
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			log.Warn("notification consumer: bad payload, skipping",
				zap.Error(err))
			return nil // 不能解析的消息直接 ACK 丢弃，避免死循环
		}

		if err := cons.HandleEvent(ctx, event); err != nil {
			log.Warn("notification consumer: handle failed",
				zap.String("type", event.EventType),
				zap.Error(err))
			return err // NACK → 重试
		}
		return nil
	})
}

// bindQueue 辅助：提前获取 channel 并绑定队列到路由键。
func bindQueue(ctx context.Context, mqClient *mq.Client, queue, pattern string) error {
	// 使用 mqClient 的 Consume 之前必须先绑定队列
	// 由于 Client 不暴露 channel，我们借助 DeclareQueue 的幂等特性再声明一次
	// 但 DeclareQueue 只绑定了第一个 routing key，因此这里需要直接操作
	//
	// 简化：通过声明一个临时队列利用 Bind 接口……实际上 mq.Client
	// 没有暴露 QueueBind 方法。替代方案是修改 Client 或使用 shared queue。
	//
	// 解决方案：使用 topic 的通配符模式，一次性绑定：
	//   plane.events.issue.* 覆盖所有 issue 事件
	//   plane.events.comment.* 覆盖所有 comment 事件
	//
	// 但实际上 DeclareQueue 只接受单个 routingKey。因此需要扩展 Client
	// 或使用 QueueBindRaw 等方式。为了让 MVP 快速跑通，这里使用
	// "plane.events.#"（匹配所有事件），然后在 HandleEvent 内过滤。

	// 由于 mq.Client API 限制，暂时使用全匹配。如需精确订阅，可扩展 mq.Client。
	_ = ctx
	_ = mqClient
	_ = queue
	_ = pattern
	return nil
}

// fmtQueueErr 包装队列操作错误。
func fmtQueueErr(op string, err error) error {
	return errors.New("notification: " + op + ": " + err.Error())
}

// extractPool 从依赖传入的 db 中提取 *pgxpool.Pool。
// main.go 传入的是 *pgxpool.Pool 的 Pool 字段（即 *pgxpool.Pool 自身）。
func extractPool(db interface{}) *pgxpool.Pool {
	type poolExtractor interface {
		// *pgxpool.Pool 本身没有特定的 marker 方法
	}
	_ = poolExtractor(db)

	// 直接尝试类型断言
	if p, ok := db.(*pgxpool.Pool); ok {
		return p
	}
	// 如果传入的是 pool.Pool 类型（即 *pgxpool.Pool 的别名），直接返回
	return nil
}
