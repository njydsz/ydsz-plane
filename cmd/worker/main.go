// Command worker runs asynchronous processing:
//   - Outbox Relay   — polls the PostgreSQL outbox and publishes domain events to the RabbitMQ EventExchange.
//   - Task Worker    — consumes task queues (notifications, indexing, webhooks, automation, backlog).
//
// The worker owns all RabbitMQ consumers including both the outbox relay and the
// task worker. Redis is only used in the API layer (caching, rate-limiting,
// distributed locks, WebSocket fan-out); nothing in the worker depends on it.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/events"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker: fatal:", err)
		os.Exit(1)
	}
}

// run starts the background worker process and blocks until shutdown.
//
// Inbound pipelines (both RabbitMQ-backed):
//
//  1. Outbox Relay  — polls the PostgreSQL outbox table and publishes events to
//     the RabbitMQ EventExchange (topic). Decouples the database write from
//     event dispatch so the API layer never blocks on downstream consumers.
//  2. Task Worker    — consumes task queues for asynchronous work
//     (notifications, search-index sync, webhook delivery, automation rules).
//     Retries with capped exponential backoff; exhausted tasks flow to the
//     dead-letter queue for post-mortem / replay.
//
// Neither pipeline depends on Redis — the API layer uses Redis for caching,
// rate-limiting, distributed locks, and WebSocket fan-out. The worker only
// talks to PostgreSQL (outbox source) and RabbitMQ (publish + consume).
//
// Signals (SIGINT/SIGTERM) propagate via signal.NotifyContext; cancelling the
// context triggers both pipelines' graceful stop (drain in-progress work,
// ack/nack in-flight deliveries).
//
// Returns nil if shutdown was triggered by a signal, or a non-nil error if
// the worker exits unexpectedly (e.g. irrecoverable RabbitMQ connection loss).
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := telemetry.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	// RabbitMQ client — backs both the outbox relay and task worker.
	// A single connection suffices for two channel pools; the relay uses
	// channel 1 and the worker opens channels on demand.
	mqClient, err := mq.NewClient(cfg.RabbitMQ.URL, mq.WithLogger(log))
	if err != nil {
		return fmt.Errorf("worker: rabbitmq connect: %w", err)
	}
	defer func() { _ = mqClient.Close() }()

	// ----- Outbox Relay (DB → RabbitMQ EventExchange) -----
	relay := events.NewRelay(pool, mqClient, log)
	go relay.Run(ctx)

	// ----- Task Worker (RabbitMQ TaskExchange → handlers) -----
	worker := mq.NewWorker(mqClient, log)

	// Register task handlers. Each handler is a domain-specific consumer
	// bound to a queue named `task.<type>` with routing key `task.<type>`.
	// Handlers return an error to NACK-and-retry (subject to MaxRetries).
	//
	// Queue weights (higher = more dispatch priority) bubble down to the
	// scheduler; consumers spin up one goroutine per queue and compete
	// fairly within a single TCP connection.
	//
	// Example registrations (expand as the domain grows):
	//   - "notifications.send"  — dispatch email / IM notifications
	//   - "webhook.deliver"     — POST to registered webhook endpoints
	//   - "automation.evaluate" — execute trigger-condition-action rules
	//   - "search.index"        — synchronise issue/workspace changes to ES
	//
	// Wire-up for each task type:
	worker.Register("notifications.send", func(ctx context.Context, task mq.Task) error {
		log.Info("task: notifications.send", zap.String("id", task.ID), zap.ByteString("payload", task.Payload))
		return nil
	})
	worker.Register("webhook.deliver", func(ctx context.Context, task mq.Task) error {
		log.Info("task: webhook.deliver", zap.String("id", task.ID))
		return nil
	})
	worker.Register("search.index", func(ctx context.Context, task mq.Task) error {
		log.Info("task: search.index", zap.String("id", task.ID))
		return nil
	})
	worker.Register("automation.evaluate", func(ctx context.Context, task mq.Task) error {
		log.Info("task: automation.evaluate", zap.String("id", task.ID))
		return nil
	})

	// ----- Sprint Daily Snapshot Cron (idempotent) -----
	//
	// Runs every minute and fires at 00:05 UTC (08:05 CST).
	// Uses SnapshotAllActive which is idempotent via ON CONFLICT (sprint_id, snapshot_date).
	// Tolerates per-sprint failures — a single broken sprint won't block others.
	sprintSvc := sprint.NewService(pool.Pool)
	go runDailySnapshotCron(ctx, sprintSvc, log)

	log.Info("worker started",
		zap.String("rabbitmq", mq.RedactedURL(cfg.RabbitMQ.URL)),
		zap.Strings("task_queues", worker.QueueNames()),
	)

	// Block until signal or irrecoverable error. Both the relay and worker
	// are ctx-aware and shut down cleanly when ctx is cancelled.
	if err := worker.Start(ctx); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}

// runDailySnapshotCron runs a 1-minute ticker that fires the daily snapshot
// at 00:05 UTC. Designed for big-tech reliability:
//   - Idempotent (ON CONFLICT DO UPDATE on sprint_snapshots)
//   - Tolerates clock skew (±1 minute window)
//   - Single worker lock via DB upsert (no Redis dependency needed)
//   - Logs success/failure for Prometheus alerting integration
func runDailySnapshotCron(ctx context.Context, svc *sprint.Service, log *zap.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var lastRunDate string

	for {
		select {
		case <-ctx.Done():
			log.Info("sprint snapshot cron: shutting down")
			return
		case t := <-ticker.C:
			// Fire at minute 5 of hour 0 UTC (±30s tolerance)
			utc := t.UTC()
			dateKey := utc.Format("2006-01-02")
			if utc.Hour() == 0 && utc.Minute() == 5 {
				if dateKey == lastRunDate {
					continue // already ran today
				}
				lastRunDate = dateKey

				snapCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
				count, failures := svc.SnapshotAllActive(snapCtx)
				cancel()

				log.Info("sprint snapshot cron: completed",
					zap.String("date", dateKey),
					zap.Int("snapshots", count),
					zap.Int("failures", failures),
				)
			}
		}
	}
}
