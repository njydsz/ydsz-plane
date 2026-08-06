// Package events implements the transactional outbox pattern:
// domain events are written to the domain_events table inside the business
// transaction, then a background relay publishes them to NATS and marks them
// published. Consumers must be idempotent on event id.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/ydszopen/ydsz-plane/internal/infrastructure/telemetry"
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

// Subject builds the NATS subject, e.g. "ydsz.issue.status_changed".
func Subject(eventType string) string { return "ydsz." + eventType }

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

// Relay polls unpublished events and publishes them to NATS.
type Relay struct {
	db     Querier
	nc     *nats.Conn
	log    *zap.Logger
	batch  int
	period time.Duration
}

// NewRelay constructs a Relay. batch is the poll batch size.
func NewRelay(db Querier, nc *nats.Conn, log *zap.Logger) *Relay {
	return &Relay{db: db, nc: nc, log: log, batch: 200, period: 500 * time.Millisecond}
}

// Run starts the polling loop until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.period)
	defer ticker.Stop()
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
		data, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("events: marshal %d: %w", e.ID, err)
		}
		if err := r.nc.Publish(Subject(e.EventType), data); err != nil {
			telemetry.NATSPublished.WithLabelValues("error").Inc()
			return fmt.Errorf("events: publish %d: %w", e.ID, err)
		}
		if _, err := r.db.Exec(ctx, `UPDATE domain_events SET published_at = now() WHERE id = $1`, e.ID); err != nil {
			return fmt.Errorf("events: mark published %d: %w", e.ID, err)
		}
		telemetry.NATSPublished.WithLabelValues("ok").Inc()
	}
	return nil
}
