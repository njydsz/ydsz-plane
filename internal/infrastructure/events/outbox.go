// Package events implements the transactional outbox pattern:
// domain events are written to the domain_events table inside the business
// transaction, then a background relay publishes them to RabbitMQ
// (EventExchange) and marks them published. Consumers must be idempotent
// on event id.
//
// Messaging stack (post-M1 upgrade):
//   - RabbitMQ: EventExchange (topic, at-least-once delivery with consumer
//     acks, dead-letter routing, message TTL). Handles domain events that
//     must be processed reliably (notifications, webhooks, ES indexing,
//     audit trail, automation rules).
//   - Redis Streams: retained for low-latency real-time push (WebSocket
//     fan-out), rate limiting, distributed locks, and Asynq task broker.
//
// RabbitMQ topology for events:
//   - Exchange "plane.events" (topic, durable)
//   - Routing key: plane.events.<aggregate_type>.<event_type>
//   - Dead-letter exchange "plane.dlx" routes poisoned messages.
//   - Messages are persistent (DeliveryMode=2) and survive broker restarts.
//
// Retry & observability:
//   - Failed Nacks requeue the message (limited by DLX threshold).
//   - Prometheus counters track published/nack counts (added in S3+).
//   - Structured logs carry event_id, worker_id, routing_key for tracing.
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// Event is a domain-event record stored in PostgreSQL outbox (domain_events).
type Event struct {
	ID            int64           `json:"id"`
	WorkspaceID   int64           `json:"workspace_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

// toEnvelope converts an outbox record to an mq.EventEnvelope.
func (e Event) toEnvelope() mq.EventEnvelope {
	return mq.EventEnvelope{
		EventID:       e.ID,
		EventType:     e.EventType,
		WorkspaceID:   e.WorkspaceID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		Payload:       e.Payload,
		OccurredAt:    e.OccurredAt,
		Exchange:      mq.EventExchange,
		RoutingKey:    mq.RoutingKey(e.AggregateType, e.EventType),
	}
}

// Recorder writes events into the outbox within an existing transaction.
type Recorder struct{}

// NewRecorder constructs a Recorder.
func NewRecorder() *Recorder { return &Recorder{} }

// Record inserts an event into domain_events. Must be called inside the
// business transaction so event and state commit atomically.
func (r *Recorder) Record(ctx context.Context, tx pgx.Tx, e Event) error {
	const q = `INSERT INTO domain_events
		(workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(ctx, q, e.WorkspaceID, e.AggregateType, e.AggregateID, e.EventType, e.Payload); err != nil {
		return fmt.Errorf("events: record %s: %w", e.EventType, err)
	}
	return nil
}

// Querier is the minimal DB surface the relay needs (the pool satisfies it).
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Relay polls unpublished events and publishes them to RabbitMQ.
//
// Lifecycle:
//  1. NewRelay(initialises the RabbitMQ client and declares topology).
//  2. Run(ctx) starts the polling loop until ctx is cancelled.
//  3. On shutdown ctx cancel, in-flight batch completes before return.
//
// Observability: each batch cycle logs published count, error, and latency.
type Relay struct {
	db     Querier
	mq     *mq.Client
	log    *zap.Logger
	batch  int
	period time.Duration
}

// NewRelay constructs a Relay with an already-initialised MQ client.
// The caller is responsible for closing the MQ client via Relay.Close
// or by passing a shared client.
func NewRelay(db Querier, mqClient *mq.Client, log *zap.Logger) *Relay {
	return &Relay{db: db, mq: mqClient, log: log, batch: 200, period: 500 * time.Millisecond}
}

// Run starts the polling loop until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	r.log.Info("outbox relay started",
		zap.Int("batch", r.batch),
		zap.Duration("period", r.period),
		zap.String("exchange", mq.EventExchange))

	for {
		select {
		case <-ctx.Done():
			r.log.Info("outbox relay stopped (context cancelled)")
			return
		default:
		}

		if err := r.publishBatch(ctx); err != nil {
			r.log.Error("outbox relay batch failed", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(r.period):
		}
	}
}

// publishBatch polls unpublished events and publishes each to RabbitMQ.
func (r *Relay) publishBatch(ctx context.Context) error {
	rows, err := r.db.Query(ctx, `
		SELECT id, workspace_id, aggregate_type, aggregate_id, event_type, payload, occurred_at
		FROM domain_events
		WHERE published_at IS NULL
		ORDER BY id
		LIMIT $1`, r.batch)
	if err != nil {
		return fmt.Errorf("events: poll: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.AggregateType, &e.AggregateID, &e.EventType, &e.Payload, &e.OccurredAt); err != nil {
			return fmt.Errorf("events: scan: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	for _, e := range events {
		envelope := e.toEnvelope()
		if err := r.mq.PublishEvent(ctx, envelope); err != nil {
			return fmt.Errorf("events: publish %d (%s): %w", e.ID, e.EventType, err)
		}
		if _, err := r.db.Exec(ctx,
			`UPDATE domain_events SET published_at = now() WHERE id = $1`,
			e.ID); err != nil {
			return fmt.Errorf("events: mark published %d: %w", e.ID, err)
		}
	}

	r.log.Debug("outbox relay batch published", zap.Int("count", len(events)))
	return nil
}

// Close releases the underlying RabbitMQ client held by the relay.
// Callers may skip this if the client is shared and closed elsewhere.
func (r *Relay) Close() error {
	if r.mq != nil {
		return r.mq.Close()
	}
	return nil
}

// EnqueueTask publishes a background task envelope to the TaskExchange.
// This is the lightweight path for fire-and-forget async work that does
// not require the full job-queue semantics of Asynq (e.g. event-driven
// reactions in automation rules, real-time indexing fan-out).
func EnqueueTask(ctx context.Context, client *mq.Client, taskType string, workspaceID int64, payload json.RawMessage) error {
	if client == nil {
		return errors.New("events: mq client is nil")
	}
	return client.Publish(ctx, mq.TaskExchange, "plane.tasks."+taskType, mq.EventEnvelope{
		EventType:   taskType,
		WorkspaceID: workspaceID,
		Payload:     payload,
		OccurredAt:  time.Now(),
		Exchange:    mq.TaskExchange,
		RoutingKey:  "plane.tasks." + taskType,
	})
}

// --------------------------------------------------------------------------
// Legacy consumer helpers (kept for the WebSocket fan-out path).
// The RealTime relay that used to consume Redis Streams still lives
// alongside RabbitMQ: WS fan-out favours Redis P/S latency for sub-100ms
// delivery, while RabbitMQ handles the durable, replay-friendly event
// streams.
// --------------------------------------------------------------------------

var _ = sql.ErrNoRows // keep import if SQL helpers added later
