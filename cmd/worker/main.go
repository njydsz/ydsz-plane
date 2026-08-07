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
