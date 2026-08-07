// Package events implements the transactional outbox pattern:
// domain events are written to the domain_events table inside the business
// transaction, then a background relay publishes them to Redis Streams and
// marks them published. Consumers must be idempotent on event id.
//
// Replaced NATS JetStream with Redis Streams (v8+) for operational simplicity:
// - At-least-once delivery via XACK
// - Consumer group for horizontal scaling
// - Native persistence (AOF/RDB)
// - Zero additional infrastructure (already required by cache layer)
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// StreamKey is the Redis Stream key for outbox events.
	StreamKey = "ydsz:events"
	// ConsumerGroup is the default consumer group name.
	ConsumerGroup = "ydsz-consumers"
)

// Event is a domain event record.
type Event struct {
	ID            int64           `json:"id"`
	WorkspaceID   int64           `json:"workspace_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

// StreamEvent is the event envelope stored in Redis Stream.
type StreamEvent struct {
	EventID       int64           `json:"event_id"`
	EventType     string          `json:"event_type"`
	WorkspaceID   int64           `json:"workspace_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   int64           `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func streamEvent(e Event) StreamEvent {
	return StreamEvent{
		EventID:       e.ID,
		EventType:     e.EventType,
		WorkspaceID:   e.WorkspaceID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		Payload:       e.Payload,
		OccurredAt:    e.OccurredAt,
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

// Relay polls unpublished events and publishes them to Redis Streams.
type Relay struct {
	db     Querier
	rdb    *redis.Client
	log    *zap.Logger
	batch  int
	period time.Duration
}

// NewRelay constructs a Relay. batch is the poll batch size.
func NewRelay(db Querier, rdb *redis.Client, log *zap.Logger) *Relay {
	return &Relay{db: db, rdb: rdb, log: log, batch: 200, period: 500 * time.Millisecond}
}

// EnsureConsumerGroup creates the consumer group if it doesn't exist.
func (r *Relay) EnsureConsumerGroup(ctx context.Context) error {
	err := r.rdb.XGroupCreateMkStream(ctx, StreamKey, ConsumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("events: create consumer group: %w", err)
	}
	return nil
}

// Run starts the polling loop until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.period)
	defer ticker.Stop()

	// Ensure consumer group exists on startup.
	if err := r.EnsureConsumerGroup(ctx); err != nil {
		r.log.Warn("outbox relay: ensure consumer group failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.publishBatch(ctx); err != nil {
				r.log.Error("outbox relay batch failed", zap.Error(err))
			}
		}
	}
}

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

	for _, e := range events {
		se := streamEvent(e)
		// Add to Redis Stream. ID "*" = auto-increment.
		streamID, err := r.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: StreamKey,
			Values: map[string]any{
				"data":   se,
				"type":   e.EventType,
				"ws_id":  e.WorkspaceID,
				"agg_id": e.AggregateID,
			},
		}).Result()
		if err != nil {
			return fmt.Errorf("events: xadd %d: %w", e.ID, err)
		}

		if _, err := r.db.Exec(ctx, `UPDATE domain_events SET published_at = now(), stream_id = $1 WHERE id = $2`, streamID, e.ID); err != nil {
			return fmt.Errorf("events: mark published %d: %w", e.ID, err)
		}
	}
	return nil
}

// ReadEvents reads pending events from the stream (consumer-group aware).
// Used by downstream consumers (notifications, webhooks, real-time push).
func (r *Relay) ReadEvents(ctx context.Context, count int64) ([]StreamEvent, string, error) {
	res, err := r.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: "ydsz-consumer",
		Streams:  []string{StreamKey, ">"},
		Count:    count,
		Block:    time.Second,
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("events: xreadgroup: %w", err)
	}

	if len(res) == 0 {
		return nil, "", nil
	}

	var events []StreamEvent
	var lastID string
	for _, msg := range res[0].Messages {
		lastID = msg.ID
		data, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}
		var se StreamEvent
		if err := json.Unmarshal([]byte(data), &se); err != nil {
			continue
		}
		events = append(events, se)
	}
	return events, lastID, nil
}

// AckEvent acknowledges a processed event.
func (r *Relay) AckEvent(ctx context.Context, streamID string) error {
	return r.rdb.XAck(ctx, StreamKey, ConsumerGroup, streamID).Err()
}
