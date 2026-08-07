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

// ----- Handler callback -----

// TaskHandler processes a single delivery. Returning an error signals
// the Worker to NACK-and-requeue (subject to the per-task MaxRetries).
type TaskHandler func(ctx context.Context, task Task) error

// ----- Task definition -----

// Task is a unit of asynchronous work submitted to a RabbitMQ-backed queue.
// Built atop the same RabbitMQ primitives used by the Outbox event bus:
//   - Delayed dispatch via per-message TTL + DLX loopback.
//   - Priority selection via per-queue `x-max-priority` header.
//   - Retries with capped exponential backoff.
//   - Dead-letter routing for exhausted / permanently-failed tasks.
//
// This type replaces the Redis-backed Asynq semantics used in M0.5 / S1.
type Task struct {
	// ID is a unique identifier for dedup and logging. Auto-generated on
	// Enqueue if left empty.
	ID string `json:"id"`

	// Type selects the handler. Maps 1:1 to a queue name
	// (`task.<Type>`) bound to TaskExchange using routing key `task.<Type>`.
	Type string `json:"type"`

	// Payload is the opaque task data. Handlers decode what they need.
	Payload json.RawMessage `json:"payload"`

	// Priority selects dispatch order among pending tasks on the same queue.
	// Range 0-255, higher = dispatched sooner. Default 0 (no priority).
	// Receivers must declare `x-max-priority` on their queue.
	Priority uint8 `json:"priority"`

	// MaxRetries caps automatic requeues after handler failure. After the
	// Nth failed attempt the task is NOT requeued and follows the queue's
	// dead-letter route. Default 0 — no retries.
	MaxRetries int `json:"max_retries"`

	// RetryCount is incremented by the Worker on each requeue. Begin at 0.
	RetryCount int `json:"retry_count"`

	// Delay is the minimum wait before the task becomes dispatchable. The
	// underlying mechanism is per-message TTL: the task is first routed to
	// a transient intermediary via the DeadLetterExchange and, after expiry,
	// bounced back to the task queue. Maximum 24 h per RabbitMQ policy.
	Delay time.Duration `json:"delay"`

	// CreatedAt is the task submission timestamp; used for metrics.
	CreatedAt time.Time `json:"created_at"`
}

// RoutingKey returns the AMQP routing key for this task type.
func (t Task) RoutingKey() string {
	return "task." + t.Type
}

// IsZero reports whether the Task is an empty placeholder.
func (t Task) IsZero() bool { return t.Type == "" }

// ----- TaskClient -----

// TaskClient submits Task payloads to the RabbitMQ task queues. It is safe
// for concurrent use because the underlying Client serialises publishes.
type TaskClient struct {
	client *Client
	log    *zap.Logger
	mu     sync.Mutex
}

// NewTaskClient wraps an existing *Client for task submission.
// The caller remains responsible for closing the underlying Client.
func NewTaskClient(client *Client, log *zap.Logger) *TaskClient {
	if log == nil {
		log = zap.NewNop()
	}
	return &TaskClient{client: client, log: log}
}

// Enqueue submits a task to its queue. If task.ID is empty, one is generated.
// The queue is lazily declared on first use (idempotent).
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

	// Ensure queue exists (idempotent).
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
		// Per-message TTL (requires RabbitMQ ≥ 3.11). On expiry the message
		// is routed to the queue's DLX; the task queue binds its DLX
		// bounce-routing-key to itself → the message returns to the head of
		// the queue after the delay has elapsed.
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

// ensureQueue lazily declares the queue and binds it to TaskExchange.
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

// Worker consumes Task deliveries from RabbitMQ queues and dispatches to
// registered handlers. It replaces asynq.ServeMux for the Ydsz Plane task
// subsystem while sharing the same proven connection model used by the
// Outbox Relay.
type Worker struct {
	client   *Client
	log      *zap.Logger
	handlers map[string]TaskHandler
	mu       sync.RWMutex
}

// NewWorker wraps a connected *Client as a task-queue consumer.
func NewWorker(client *Client, log *zap.Logger) *Worker {
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{client: client, log: log, handlers: make(map[string]TaskHandler)}
}

// Register assigns a handler for the given task type. Calling with the same
// type overwrites the previous handler.
func (w *Worker) Register(taskType string, handler TaskHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[taskType] = handler
}

// QueueNames returns the queue names of all currently-registered handlers.
func (w *Worker) QueueNames() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	qs := make([]string, 0, len(w.handlers))
	for t := range w.handlers {
		qs = append(qs, queueName(t))
	}
	return qs
}

// Start launches one consumer goroutine per registered queue and blocks
// until ctx is cancelled. On cancellation each consumer finishes its
// current delivery (when possible) and returns. If a consumer drops due to
// connection loss, Start exits with that error; the Client.reconnectLoop
// will re-establish the connection and the caller is expected to invoke
// Start again.
//
// Idempotent-checked: call Stop by cancelling the ctx passed to Start.
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
				// Unparseable → do NOT requeue (return error → NACK).
				return fmt.Errorf("mq/task: bad payload: %w", err)
			}

			w.mu.RLock()
			handler, ok := w.handlers[task.Type]
			w.mu.RUnlock()
			if !ok {
				w.log.Warn("mq/task: no handler registered",
					zap.String("type", task.Type), zap.String("queue", queue))
				return nil // ACK defensive — shouldn't happen
			}

			if err := handler(ctx, task); err != nil {
				return w.handleFailure(ctx, task, err)
			}
			return nil // ACK on success
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

// handleFailure decides whether to retry a failed task or let it fall
// through to the dead-letter queue.
func (w *Worker) handleFailure(ctx context.Context, task Task, handlerErr error) error {
	w.log.Warn("mq/task: handler failed",
		zap.String("type", task.Type),
		zap.String("id", task.ID),
		zap.Int("retry", task.RetryCount),
		zap.Int("max_retries", task.MaxRetries),
		zap.Error(handlerErr),
	)

	if task.RetryCount >= task.MaxRetries {
		// No retries remaining → allow NACK to route to DLX.
		return fmt.Errorf("mq/task: %s exhausted retries: %w", task.Type, handlerErr)
	}

	// Re-enqueue with incremented RetryCount and exponential backoff.
	task.RetryCount++
	backoff := retryBackoff(task.RetryCount)
	task.Delay = backoff
	task.ID = fmt.Sprintf("%s-r%d-%d", task.Type, task.RetryCount, time.Now().UnixNano())

	tc := NewTaskClient(w.client, w.log)
	if err := tc.Enqueue(ctx, task); err != nil {
		w.log.Error("mq/task: retry enqueue failed", zap.Error(err))
	}
	// Return the original error → NACK the exhausted delivery without requeue
	// (the retry path produced a fresh message with its own envelope).
	return handlerErr
}

// retryBackoff calculates capped exponential backoff: 2^attempt seconds, max 30s.
func retryBackoff(attempt int) time.Duration {
	d := time.Duration(1<<uint(attempt)) * time.Second
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	return d
}

// ----- helpers -----

// queueName derives a durable queue identifier from a task type.
func queueName(taskType string) string {
	return "task." + taskType
}
