// Package webhook — Webhook 事件消费者。
//
// 订阅 EventExchange 的领域事件，过滤匹配 Webhook 订阅后投递。
// 并行消费（每个订阅互不阻塞），同一订阅串行投递避免竞争。
//
// 消费拓扑：
//   - 队列：message plane.events.#（订阅全事件路由）
//   - 一次消费一条，成功投递目标后 ACK。
//   - 投递失败走 TaskExchange 的中转重试；消费本身 ACK，不阻塞后续事件。
package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// Consumer 是 Webhook 事件消费者。
type Consumer struct {
	dispatcher *Dispatcher
	log        *zap.Logger
}

// NewConsumer 构造 Webhook 消费者。
func NewConsumer(dispatcher *Dispatcher, log *zap.Logger) *Consumer {
	return &Consumer{
		dispatcher: dispatcher,
		log:        log,
	}
}

// HandleEvent 处理单个领域事件。
// 非订阅事件静默 ACK（无匹配订阅也是正常路径）。
func (c *Consumer) HandleEvent(ctx context.Context, envelope mq.EventEnvelope) error {
	if c.dispatcher == nil {
		return nil
	}

	ok, failed := c.dispatcher.DispatchEvent(ctx, envelope)
	if ok > 0 || failed > 0 {
		c.log.Debug("webhook: dispatched",
			zap.String("event", envelope.EventType),
			zap.Int("ok", ok),
			zap.Int("failed", failed))
	}
	return nil
}

// HandleRetryTask 处理重试任务（来自 TaskExchange webhook.retry）。
func (c *Consumer) HandleRetryTask(ctx context.Context, task mq.Task) error {
	if c.dispatcher == nil {
		return nil
	}
	return c.dispatcher.HandleRetry(ctx, task)
}

// consumerQueue 是 webhook 消费者绑定的队列名。
const consumerQueue = "webhook.events"

// routingPattern 订阅所有领域事件（Dispatcher 内部按订阅匹配过滤）。
const routingPattern = "plane.events.#"

// RunConsumer 启动阻塞型 webhook 投递消费者循环。
//
// 订阅 EventExchange 的全部领域事件，逐条交给 Dispatcher 匹配 Webhook 订阅并投递。
// 投递失败的事件由 Dispatcher 异步排入 TaskExchange 的重试队列（webhook.retry），
// 消费本身立即 ACK，不阻塞后续事件。
//
// 当 ctx 取消时优雅退出；连接断开后以指数退避自动重连。
//
// 调用方应在独立 goroutine 中运行：
//
//	go webhook.RunConsumer(ctx, mqClient, dispatcher, log)
func RunConsumer(ctx context.Context, mqClient *mq.Client, dispatcher *Dispatcher, log *zap.Logger) {
	log.Info("webhook consumer: starting",
		zap.String("queue", consumerQueue),
		zap.String("exchange", mq.EventExchange))

	consumer := NewConsumer(dispatcher, log)

	for {
		select {
		case <-ctx.Done():
			log.Info("webhook consumer: stopped")
			return
		default:
		}

		if err := runConsumeLoop(ctx, mqClient, consumer, log); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Info("webhook consumer: stopped (context)")
				return
			}
			log.Warn("webhook consumer: connection lost, retrying",
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
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, consumer *Consumer, log *zap.Logger) error {
	if _, err := mqClient.DeclareQueue(ctx, consumerQueue, mq.EventExchange, routingPattern, amqp.Table{
		"x-max-priority":       int64(5),
		"x-dead-letter-exchange": mq.DeadLetterExchange,
	}); err != nil {
		return errors.New("webhook: declare queue: " + err.Error())
	}

	return mqClient.Consume(ctx, consumerQueue, "webhook-consumer", false, func(delivery amqp.Delivery) error {
		var envelope mq.EventEnvelope
		if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
			log.Warn("webhook consumer: bad payload, skipping", zap.Error(err))
			return nil // 无法解析直接 ACK，避免死信循环
		}

		if err := consumer.HandleEvent(ctx, envelope); err != nil {
			log.Warn("webhook consumer: handle failed",
				zap.String("event_type", envelope.EventType), zap.Error(err))
			return err // NACK → 重试（受 MaxRetries 限制）
		}
		return nil // ACK
	})
}
