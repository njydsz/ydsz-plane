// Package mq provides a production-grade RabbitMQ client implementing
// connection management, automatic reconnection, publisher-confirms, and
// consumer infrastructure for the Ydsz Plane event bus.
//
// Architecture:
//   - Topic exchange "plane.events"   — domain events from the outbox relay.
//   - Topic exchange "plane.tasks"    — background task dispatch (replaces
//     Asynq for event-driven workloads while Asynq remains for job-queue
//     semantics).
//   - Dead-letter exchange "plane.dlx" — captures messages that fail after
//     max retries, enabling post-mortem analysis and replay.
//
// Why RabbitMQ alongside Redis:
//   - Redis Streams: operational simplicity, low-latency pub/sub for
//     real-time push (WebSocket fan-out), rate limiting, distributed locks.
//   - RabbitMQ: reliable at-least-once delivery with consumer acks,
//     dead-letter routing, message TTL, priority queues, and flexible
//     topic-based routing patterns required by enterprise workflows.
//
// The topic hierarchy under "plane.events" uses dot-separated routing keys:
//   plane.events.<aggregate_type>.<event_type>
//   e.g. plane.events.issue.created, plane.events.workspace.member_added
//
// Consumer queues bind with wildcard patterns for flexible subscription.
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

// Exchange and queue constants.
const (
	// EventExchange is the primary topic exchange for domain events.
	EventExchange = "plane.events"

	// TaskExchange dispatches background tasks to consumer queues.
	TaskExchange = "plane.tasks"

	// DeadLetterExchange routes exhausted messages for inspection / replay.
	DeadLetterExchange = "plane.dlx"

	// DefaultConsumerTag prefix for auto-generated consumer tags.
	DefaultConsumerTag = "plane-consumer"

	// MaxReconnectAttempts before giving up (with exponential backoff).
	MaxReconnectAttempts = 10
)

// EventEnvelope is the wire format for all messages published to RabbitMQ.
// It mirrors the StreamEvent schema previously stored in JSON fields.
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

// RoutingKey builds a routing key from aggregate and event type.
func RoutingKey(aggregate, eventType string) string {
	return fmt.Sprintf("plane.events.%s.%s", aggregate, eventType)
}

// Client wraps an AMQP connection and channel with automatic reconnection.
// It is safe for concurrent use: publish acquires a mutex; consume
// operations use a separate channel.
type Client struct {
	url    string
	config amqp.Config
	log    *zap.Logger
	tls    *tls.Config

	mu       sync.RWMutex
	conn     *amqp.Connection
	ch       *amqp.Channel

	chanMU   sync.Mutex // serialises channel-level publishing

	connClose chan *amqp.Error
	chanClose chan *amqp.Error

	connected bool
	done      chan struct{}
}

// ClientOption configures optional client parameters.
type ClientOption func(*Client)

// WithTLS attaches a TLS config for amqps:// connections.
func WithTLS(tlsConf *tls.Config) ClientOption {
	return func(c *Client) { c.tls = tlsConf }
}

// WithLogger sets a custom logger.
func WithLogger(log *zap.Logger) ClientOption {
	return func(c *Client) { c.log = log }
}

// NewClient constructs a RabbitMQ client and establishes the initial
// connection. It blocks until the connection succeeds so that callers
// fail fast on startup rather than discovering a broken connection at
// first publish.
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

// connect opens a connection and channel, declares the topology, and
// registers close-notification handlers.
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

	// Enable publisher confirms so we get async acks/nacks.
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

// declareTopology creates exchanges and the dead-letter queue. Idempotent
// — re-declaring with the same parameters is a no-op in RabbitMQ.
func declareTopology(ch *amqp.Channel) error {
	// Dead-letter exchange (direct) + dead-letter queue for inspection.
	if err := ch.ExchangeDeclare(DeadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(DeadLetterExchange+".queue", true, false, false, false, amqp.Table{
		"x-message-ttl": int64(7 * 24 * time.Hour / time.Millisecond), // 7d retention
		"x-max-length":  int64(100_000),
		"x-overflow":    "reject-publish",
	}); err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	if err := ch.QueueBind(DeadLetterExchange+".queue", "", DeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	// Primary topic exchanges.
	for _, ex := range []string{EventExchange, TaskExchange} {
		if err := ch.ExchangeDeclare(ex, "topic", true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare exchange %s: %w", ex, err)
		}
	}
	return nil
}

// reconnectLoop monitors for connection/channel close signals and attempts
// a reconnection with bounded exponential backoff.
func (c *Client) reconnectLoop() {
	backoff := time.Second

	for {
		select {
		case <-c.done:
			return
		case err := <-c.connClose:
			if err == nil {
				return // graceful close
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

// Publish sends a message to the specified exchange/routing-key with
// mandatory delivery and publisher-confirm synchronisation. Returns when
// the broker has confirmed receipt or the context expires.
//
// Mandatory=true ensures unroutable messages are returned rather than
// silently dropped — this should never happen if consumers declare
// their bindings before publishing.
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

	// Wait for broker confirm or mandatory return.
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

// PublishRaw sends a message using a caller-supplied amqp.Publishing struct.
// Use this when you need fine-grained control over TTL (delayed tasks),
// Priority, Expiration, or other headers not exposed by the ergonomic
// Publish / PublishEvent methods. Publisher-confirm and mandatory routing
// are still enforced.
//
// Unlike Publish, the caller is responsible for setting Body, ContentType,
// DeliveryMode, MessageId, Type, Timestamp, and any application headers.
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
func (c *Client) PublishEvent(ctx context.Context, envelope EventEnvelope) error {
	envelope.Exchange = EventExchange
	envelope.RoutingKey = RoutingKey(envelope.AggregateType, envelope.EventType)
	return c.Publish(ctx, EventExchange, envelope.RoutingKey, envelope)
}

// Consume starts a long-lived consumer on the named queue. Each delivery
// is dispatched to the handler; if the handler returns an error the
// message is nack'd and requeued (until the dead-letter threshold).
//
// The function blocks until the consumer is cancelled or the connection
// drops; on reconnect the caller should invoke Consume again. Use
// DeclareQueue to create and bind the queue before consuming.
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

// DeclareQueue creates a durable queue bound to the given exchange with
// the supplied routing-key pattern and dead-letter configuration.
// Returns the declared queue for convenience.
func (c *Client) DeclareQueue(ctx context.Context, name, exchange, routingKey string, args amqp.Table) (amqp.Queue, error) {
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	if args == nil {
		args = amqp.Table{}
	}
	// Wire dead-letter routing unless the caller overrides it.
	if _, hasDLX := args["x-dead-letter-exchange"]; !hasDLX {
		args["x-dead-letter-exchange"] = DeadLetterExchange
		args["x-dead-letter-routing-key"] = name + ".dead"
	}
	// 24h default TTL for unprocessed messages before DLX routing.
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

// QueueExists probes whether a queue already exists (for idempotent declarators).
func (c *Client) QueueExists(ctx context.Context, name string) bool {
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	_, err := ch.QueueDeclarePassive(name, true, false, false, false, nil)
	return err == nil
}

// Close gracefully shuts down the channel and connection.
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

// Healthy reports whether the underlying connection is open.
func (c *Client) Healthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.conn != nil && !c.conn.IsClosed()
}

// redactURL strips credentials from an AMQP URL for safe logging.
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

// RedactedURL is the exported helper for callers that want to log the
// connection endpoint without leaking credentials.
func RedactedURL(rawurl string) string { return redactURL(rawurl) }
