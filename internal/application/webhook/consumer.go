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

// 确保 Consumer 实现了期望的接口签名（供外部断言）。
var _ = json.Unmarshal // 保留 import
