// Package persistence 提供 PostgreSQL 连接池与强制行级安全（RLS）的
// 租户上下文辅助函数（见 docs/architecture/04）。
package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 包装 pgxpool.Pool 并提供租户感知辅助函数。
type Pool struct {
	*pgxpool.Pool
}

// NewPool 创建连接池并验证连通性。
//
// 参数：
//   - url：libpq 连接串。
//   - maxConns：最大连接数；<=0 时使用 pgxpool 默认值。
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

// WithTenantTx 在设置了租户上下文的事务内执行 fn，
// 使 RLS 策略（current_setting('app.workspace_id')）按行隔离数据。
// SET LOCAL 仅作用于当前事务，在连接池下是安全的。
func (p *Pool) WithTenantTx(ctx context.Context, workspaceID int64, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistence: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // Commit 后为 no-op

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

// Ping 委托给底层连接池（供 /readyz 探活使用）。
func (p *Pool) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }
