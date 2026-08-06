// Package persistence provides the PostgreSQL connection pool and the tenant
// context helper that enforces row-level security (see docs/architecture/04).
package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with tenant-aware helpers.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a connection pool and verifies connectivity.
func NewPool(ctx context.Context, url string, maxConns int32) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("persistence: parse dsn: %w", err)
	}
	if maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("persistence: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("persistence: ping: %w", err)
	}
	return &Pool{pool}, nil
}

// WithTenantTx runs fn inside a transaction with the tenant context set, so
// that RLS policies (current_setting('app.workspace_id')) isolate rows.
// The SET LOCAL is transaction-scoped and safe under connection pooling.
func (p *Pool) WithTenantTx(ctx context.Context, workspaceID int64, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistence: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", workspaceID); err != nil {
		return fmt.Errorf("persistence: set tenant: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistence: commit: %w", err)
	}
	return nil
}

// Ping delegates to the underlying pool (used by /readyz).
func (p *Pool) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }
