// Command worker runs asynchronous processing: the outbox relay (DB → Redis
// Streams) and the Asynq task consumers (notifications, indexing, webhooks,
// automation, ...).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/cache"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/events"
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
// The worker runs two concurrent pipelines:
//
//  1. Outbox Relay — polls the PostgreSQL outbox table and publishes events to
//     Redis Streams. Decouples the database write from event dispatch so the
//     API layer never blocks on downstream consumers.
//  2. Asynq Server — consumes task queues for asynchronous work (notifications,
//     indexing, webhooks, automation triggers). The server selects tasks from
//     multiple queues with weighted priority (see Concurrency/Queues below).
//
// Redis DB index is offset by +1 relative to the API's Redis DB to keep Streams
// data isolated from session/cache keys, simplifying operational inspection
// and eviction policies.
//
// Signals (SIGINT/SIGTERM) are delivered via signal.NotifyContext; cancelling the
// context triggers the outbox relay's shutdown and the Asynq server's
// graceful drain (active tasks finish, queued tasks remain for next boot).
//
// Returns nil if shutdown was triggered by a signal, or the Asynq server error
// if it exited unexpectedly.
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

	// Redis client for outbox relay (Streams) + Asynq
	rdb, err := cache.NewClient(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("worker: redis connect: %w", err)
	}
	defer func() { _ = rdb.Close() }()

	// outbox relay: DB -> Redis Streams
	relay := events.NewRelay(pool, rdb, log)
	go relay.Run(ctx)

	// asynq task server (queues defined per domain; consumers mount in S2+)
	//
	// Queue weights (higher = more polling dispatch priority):
	//   - "default":        5 — highest priority; general tasks, issue
	//     activity processing, webhook deliveries. Most latency-sensitive.
	//   - "notifications":  3 — medium priority; email/push notifications.
	//     Users tolerate slight delay; lower weight prevents them from
	//     starving foreground work.
	//   - "automation":     2 — lowest priority; rule-triggered actions (auto-assign,
	//     status transitions). Background processing, rarely user-blocking.
	//
	// Concurrency of 10 means up to 10 tasks are processed simultaneously across
	// all queues. Tuned for moderate workloads; scale with worker replicas or
	// increase if CPU utilization is consistently low.
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB + 1},
		asynq.Config{
			Concurrency: 10,
			Queues:      map[string]int{"default": 5, "notifications": 3, "automation": 2},
		},
	)
	mux := asynq.NewServeMux()
	// mux.HandleFunc(events.TaskX, handler) — consumers are registered per Sprint.

	log.Info("worker started", zap.String("redis", cfg.Redis.Addr))
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run(mux) }()

	select {
	case <-ctx.Done():
		srv.Shutdown()
		return nil
	case err := <-errCh:
		return err
	}
}
