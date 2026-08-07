// Package mq 提供生产级 RabbitMQ 客户端，实现连接管理、自动重连、
// publisher-confirm 以及 Ydsz Plane 事件总线的消费者基础设施。
//
// 架构：
//   - topic exchange "plane.events"   —— outbox relay 的领域事件。
//   - topic exchange "plane.tasks"    —— 后台任务分发（事件驱动负载替换 Asynq，
//     Asynq 保留用于任务队列语义）。
//   - 死信 exchange "plane.dlx" —— 捕获超过最大重试次数的消息，
//     便于事后分析与重放。
//
// 为什么 RabbitMQ 与 Redis 并存：
//   - Redis Streams：运维简单、低延迟 pub/sub 用于实时推送（WebSocket 扇出）、
//     限流、分布式锁。
//   - RabbitMQ：可靠的 at-least-once 投递 + 消费者 ack、死信路由、消息 TTL、
//     优先级队列，以及企业工作流所需的灵活 topic 路由模式。
//
// "plane.events" 下的 topic 层级使用点分隔路由键：
//   plane.events.<aggregate_type>.<event_type>
//   例如 plane.events.issue.created、plane.events.workspace.member_added
//
// 消费者队列使用通配符模式绑定，实现灵活订阅。
package mq

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Exchange 与队列常量。
const (
	// EventExchange 是领域事件的主 topic exchange。
	EventExchange = "plane.events"

	// TaskExchange 将后台任务分发到消费者队列。
	TaskExchange = "plane.tasks"

	// DeadLetterExchange 将重试耗尽的消息路由出来供检查/重放。
	DeadLetterExchange = "plane.dlx"

	// DefaultConsumerTag 是自动生成消费者标签的前缀。
	DefaultConsumerTag = "plane-consumer"

	// MaxReconnectAttempts 是放弃前的最大重连次数（指数退避）。
	MaxReconnectAttempts = 10
)

// EventEnvelope 是所有发往 RabbitMQ 消息的线上格式。
// 它镜像了之前存储在 JSON 字段中的 StreamEvent schema。
type EventEnvelope struct {
	EventID       int64           `json:"event_id"`
	EventType     string          `json:"event_type"`
	WorkspaceID   int64           `json:"workspace_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Exchange      string          `json:"exchange"`
	RoutingKey    string          `json:"routing_key"`
}

// RoutingKey 由聚合类型与事件类型构建路由键。
func RoutingKey(aggregate, eventType string) string {
	return fmt.Sprintf("plane.events.%s.%s", aggregate, eventType)
}

// Client 包装 AMQP 连接与 channel，支持自动重连。
// 并发安全：发布通过互斥锁串行；消费操作使用独立 channel。
type Client struct {
	url    string
	config amqp.Config
	log    *zap.Logger
	tls    *tls.Config

	mu       sync.RWMutex
	conn     *amqp.Connection
	ch       *amqp.Channel

	chanMU   sync.Mutex // 串行化 channel 级发布

	connClose chan *amqp.Error
	chanClose chan *amqp.Error

	connected bool
	done      chan struct{}
}

// ClientOption 配置可选的客户端参数。
type ClientOption func(*Client)

// WithTLS 为 amqps:// 连接附加 TLS 配置。
func WithTLS(tlsConf *tls.Config) ClientOption {
	return func(c *Client) { c.tls = tlsConf }
}

// WithLogger 设置自定义 logger。
func WithLogger(log *zap.Logger) ClientOption {
	return func(c *Client) { c.log = log }
}

// NewClient 构造 RabbitMQ 客户端并建立初始连接。
// 它会阻塞直至连接成功，使调用方在启动阶段快速失败，
// 而不是在首次发布时才发现连接不可用。
func NewClient(url string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		url:       url,
		log:       zap.NewNop(),
		done:      make(chan struct{}),
		connClose: make(chan *amqp.Error, 1),
		chanClose: make(chan *amqp.Error, 1),
	}
	for _, opt := range opts {
		opt(c)
	}

	if err := c.connect(); err != nil {
		return nil, err
	}
	go c.reconnectLoop()

	return c, nil
}

// connect 打开连接与 channel、声明拓扑，并注册关闭通知处理器。
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cfg := c.config
	if c.tls != nil {
		cfg.TLSClientConfig = c.tls
	}

	conn, err := amqp.DialConfig(c.url, cfg)
	if err != nil {
		return fmt.Errorf("mq: dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mq: open channel: %w", err)
	}

	// 启用 publisher confirms，获取异步 ack/nack。
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("mq: enable confirms: %w", err)
	}

	if err := declareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("mq: declare topology: %w", err)
	}

	c.conn = conn
	c.ch = ch
	c.connected = true
	c.connClose = make(chan *amqp.Error, 1)
	c.chanClose = make(chan *amqp.Error, 1)
	conn.NotifyClose(c.connClose)
	ch.NotifyClose(c.chanClose)

	c.log.Info("mq: connected", zap.String("url", redactURL(c.url)))
	return nil
}

// declareTopology 创建 exchange 与死信队列。幂等 ——
// 在 RabbitMQ 中以相同参数重复声明是 no-op。
func declareTopology(ch *amqp.Channel) error {
	// 死信 exchange（direct）+ 死信队列，供检查使用。
	if err := ch.ExchangeDeclare(DeadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(DeadLetterExchange+".queue", true, false, false, false, amqp.Table{
		"x-message-ttl": int64(7 * 24 * time.Hour / time.Millisecond), // 保留 7 天
		"x-max-length":  int64(100_000),
		"x-overflow":    "reject-publish",
	}); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	if err := ch.QueueBind(DeadLetterExchange+".queue", "", DeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	// 主 topic exchange。
	for _, ex := range []string{EventExchange, TaskExchange} {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", ex, err)
		}
	}
	return nil
}

// reconnectLoop 监听连接/channel 关闭信号，并以有界指数退避重试重连。
func (c *Client) reconnectLoop() {
	backoff := time.Second

	for {
		select {
		case <-c.done:
			return
		case err := <-c.connClose:
			if err == nil {
				return // 优雅关闭
			}
			c.log.Warn("mq: connection closed", zap.Error(err))
			c.markDisconnected()
			c.reconnect(&backoff)
		case err := <-c.chanClose:
			if err == nil {
				continue
			}
			c.log.Warn("mq: channel closed", zap.Error(err))
			c.markDisconnected()
			c.reconnect(&backoff)
		}
	}
}

func (c *Client) markDisconnected() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}

func (c *Client) reconnect(backoff *time.Duration) {
	for attempt := 1; attempt <= MaxReconnectAttempts; attempt++ {
		d := *backoff
		c.log.Info("mq: reconnecting",
			zap.Int("attempt", attempt),
			zap.Duration("backoff", d))
		select {
		case <-c.done:
			return
		case <-time.After(d):
		}
		if err := c.connect(); err == nil {
			*backoff = time.Second
			return
		}
		*backoff = min(*backoff*2, 30*time.Second)
	}
	c.log.Error("mq: reconnect attempts exhausted", zap.String("url", redactURL(c.url)))
}

// Publish 向指定 exchange/路由键发送消息，强制投递并与 publisher-confirm
// 同步。在 broker 确认收到或 context 过期时返回。
//
// Mandatory=true 确保无法路由的消息被退回而不是静默丢弃 ——
// 只要消费者在发布前声明了绑定，这种情况就不应发生。
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, envelope EventEnvelope) error {
	c.chanMU.Lock()
	defer c.chanMU.Unlock()

	c.mu.RLock()
	ch := c.ch
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return errors.New("mq: not connected")
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("mq: marshal envelope: %w", err)
	}

	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	returnCh := ch.NotifyReturn(make(chan amqp.Return, 1))

	if err := ch.PublishWithContext(ctx, exchange, routingKey, true, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		MessageId:    fmt.Sprintf("%d", envelope.EventID),
		Type:         envelope.EventType,
		Body:         body,
		Headers: amqp.Table{
			"workspace_id":   envelope.WorkspaceID,
			"aggregate_type": envelope.AggregateType,
		},
	}); err != nil {
		return fmt.Errorf("mq: publish: %w", err)
	}

	// 等待 broker 确认或 mandatory 退回。
	select {
	case confirm := <-confirmCh:
		if !confirm.Ack {
			return fmt.Errorf("mq: nack received for routing key %s", routingKey)
		}
		return nil
	case ret := <-returnCh:
		return fmt.Errorf("mq: message returned (no route): key=%s reason=%s", routingKey, ret.ReplyText)
	case <-ctx.Done():
		return fmt.Errorf("mq: publish cancelled: %w", ctx.Err())
	}
}

// PublishRaw 使用调用方提供的 amqp.Publishing 结构发送消息。
// 当需要对 TTL（延迟任务）、Priority、Expiration 或其他 Publish /
// PublishEvent 未暴露的 header 做细粒度控制时使用。
// publisher-confirm 与 mandatory 路由仍然强制生效。
//
// 与 Publish 不同，调用方需自行设置 Body、ContentType、DeliveryMode、
// MessageId、Type、Timestamp 及应用 header。
func (c *Client) PublishRaw(ctx context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	c.chanMU.Lock()
	defer c.chanMU.Unlock()
	c.mu.RLock()
	ch := c.ch
	connected := c.connected
	c.mu.RUnlock()
	if !connected {
		return errors.New("mq: not connected")
	}

	confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	returnCh := ch.NotifyReturn(make(chan amqp.Return, 1))

	if err := ch.PublishWithContext(ctx, exchange, routingKey, true, false, msg); err != nil {
		return fmt.Errorf("mq: publish: %w", err)
	}

	select {
	case confirm := <-confirmCh:
		if !confirm.Ack {
			return fmt.Errorf("mq: nack received for routing key %s", routingKey)
		}
		return nil
	case ret := <-returnCh:
		return fmt.Errorf("mq: message returned (no route): key=%s reason=%s", routingKey, ret.ReplyText)
	case <-ctx.Done():
		return fmt.Errorf("mq: publish cancelled: %w", ctx.Err())
	}
}
// PublishEvent 以领域事件方式发布：自动补全 EventExchange 与路由键。
func (c *Client) PublishEvent(ctx context.Context, envelope EventEnvelope) error {
	envelope.Exchange = EventExchange
	envelope.RoutingKey = RoutingKey(envelope.AggregateType, envelope.EventType)
	return c.Publish(ctx, EventExchange, envelope.RoutingKey, envelope)
}

// Consume 在指定队列上启动长生命周期消费者。每条投递分发给 handler；
// 若 handler 返回错误，消息被 nack 并重新入队（直到死信阈值）。
//
// 该函数阻塞直至消费者被取消或连接断开；重连后调用方应重新调用
// Consume。消费前请用 DeclareQueue 创建并绑定队列。
func (c *Client) Consume(ctx context.Context, queue, consumerTag string, autoAck bool, handler func(delivery amqp.Delivery) error) error {
	if consumerTag == "" {
		consumerTag = DefaultConsumerTag
	}

	c.mu.RLock()
	ch := c.ch
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return errors.New("mq: cannot consume: not connected")
	}

	msgs, err := ch.Consume(queue, consumerTag, autoAck, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("mq: consume: %w", err)
	}

	c.log.Info("mq: consumer started", zap.String("queue", queue), zap.String("tag", consumerTag))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-msgs:
			if !ok {
				return errors.New("mq: delivery channel closed")
			}
			if err := handler(delivery); err != nil {
				c.log.Warn("mq: handler error, nacking",
					zap.Error(err),
					zap.String("queue", queue))
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					c.log.Error("mq: nack failed", zap.Error(nackErr))
				}
			} else if !autoAck {
				if ackErr := delivery.Ack(false); ackErr != nil {
					c.log.Error("mq: ack failed", zap.Error(ackErr))
				}
			}
		}
	}
}

// DeclareQueue 创建绑定到指定 exchange 的持久队列，
// 支持路由键模式与死信配置。返回已声明的队列便于后续使用。
func (c *Client) DeclareQueue(ctx context.Context, name, exchange, routingKey string, args amqp.Table) (amqp.Queue, error) {
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	if args == nil {
		args = amqp.Table{}
	}
	// 除非调用方覆盖，否则接入死信路由。
	if _, hasDLX := args["x-dead-letter-exchange"]; !hasDLX {
		args["x-dead-letter-exchange"] = DeadLetterExchange
		args["x-dead-letter-routing-key"] = name + ".dead"
	}
	// 未处理消息在进入 DLX 前默认保留 24h。
	if _, hasTTL := args["x-message-ttl"]; !hasTTL {
		args["x-message-ttl"] = int64(24 * time.Hour / time.Millisecond)
	}

	q, err := ch.QueueDeclare(name, true, false, false, false, args)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("mq: declare queue %s: %w", name, err)
	}
	if err := ch.QueueBind(name, routingKey, exchange, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("mq: bind queue %s: %w", name, err)
	}
	return q, nil
}

// QueueExists 探测队列是否已存在（供幂等声明器使用）。
func (c *Client) QueueExists(ctx context.Context, name string) bool {
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	_, err := ch.QueueDeclarePassive(name, true, false, false, false, nil)
	return err == nil
}

// Close 优雅关闭 channel 与连接。
func (c *Client) Close() error {
	close(c.done)
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	c.connected = false
	if len(errs) > 0 {
		return fmt.Errorf("mq: close: %v", errs)
	}
	return nil
}

// Healthy 报告底层连接是否开启。
func (c *Client) Healthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.conn != nil && !c.conn.IsClosed()
}

// redactURL 从 AMQP URL 中剥离凭据，用于安全日志输出。
func redactURL(rawurl string) string {
	u, err := amqp.ParseURI(rawurl)
	if err != nil {
		return "(unparseable)"
	}
	if u.Password != "" {
		u.Password = "***"
	}
	return fmt.Sprintf("amqp://%s@%s:%d/%s", u.Username, u.Host, u.Port, u.Vhost)
}

// RedactedURL 是导出的辅助函数，供调用方记录连接端点而不泄露凭据。
func RedactedURL(rawurl string) string { return redactURL(rawurl) }
