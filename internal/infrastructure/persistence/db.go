// Package persistence 提供 PostgreSQL 连接池与强制行级安全（RLS）的
// 租户上下文辅助函数（见 docs/architecture/04）。
package persistence

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

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
//   - maxConns：最大连接数；<=0 时使用启发式默认值（max(4, runtime.NumCPU() * 2)）。
//
// 连接池参数按以下优先级生效：
//  1. 环境变量覆盖（YDSZ_DB_MAX_CONNS / YDSZ_DB_MIN_CONNS / YDSZ_DB_CONN_MAX_LIFETIME /
//     YDSZ_DB_CONN_MAX_IDLE_TIME / YDSZ_DB_HEALTH_CHECK_PERIOD）
//  2. 启发式规则（基于 runtime.NumCPU()）
//  - MaxConns = max(4, runtime.NumCPU() * 2)
//  - MinConns = max(2, runtime.NumCPU())
//  - MaxConnLifetime = 30 * time.Minute
//  - MaxConnIdleTime = 5 * time.Minute
//  - HealthCheckPeriod = 1 * time.Minute
func NewPool(ctx context.Context, url string, maxConns int32) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("persistence: parse dsn: %w", err)
	}

	// --- 连接池参数启发式调优（优先级：环境变量 > 入参 > 启发式默认值）---

	// MaxConns：最大连接数
	if v := os.Getenv("YDSZ_DB_MAX_CONNS"); v != "" {
		if n, parseErr := strconv.ParseInt(v, 10, 32); parseErr == nil && n > 0 {
			cfg.MaxConns = int32(n)
		}
	} else if maxConns > 0 {
		cfg.MaxConns = maxConns
	} else {
		// 启发式：max(4, runtime.NumCPU() * 2)
		cfg.MaxConns = int32(max(4, runtime.NumCPU()*2))
	}

	// MinConns：最小空闲连接数
	if v := os.Getenv("YDSZ_DB_MIN_CONNS"); v != "" {
		if n, parseErr := strconv.ParseInt(v, 10, 32); parseErr == nil && n >= 0 {
			cfg.MinConns = int32(n)
		}
	} else {
		cfg.MinConns = int32(max(2, runtime.NumCPU()))
	}

	// MaxConnLifetime：单条连接最大存活时间
	if v := os.Getenv("YDSZ_DB_CONN_MAX_LIFETIME"); v != "" {
		if d, parseErr := time.ParseDuration(v); parseErr == nil && d > 0 {
			cfg.MaxConnLifetime = d
		} else {
			cfg.MaxConnLifetime = 30 * time.Minute
		}
	} else {
		cfg.MaxConnLifetime = 30 * time.Minute
	}

	// MaxConnIdleTime：连接最大空闲时间
	if v := os.Getenv("YDSZ_DB_CONN_MAX_IDLE_TIME"); v != "" {
		if d, parseErr := time.ParseDuration(v); parseErr == nil && d > 0 {
			cfg.MaxConnIdleTime = d
		} else {
			cfg.MaxConnIdleTime = 5 * time.Minute
		}
	} else {
		cfg.MaxConnIdleTime = 5 * time.Minute
	}

	// HealthCheckPeriod：连接健康检查周期
	if v := os.Getenv("YDSZ_DB_HEALTH_CHECK_PERIOD"); v != "" {
		if d, parseErr := time.ParseDuration(v); parseErr == nil && d > 0 {
			cfg.HealthCheckPeriod = d
		} else {
			cfg.HealthCheckPeriod = 1 * time.Minute
		}
	} else {
		cfg.HealthCheckPeriod = 1 * time.Minute
	}

	// --- Prepared Statement Cache ---
	// QueryExecModeCacheDescribe 模式下，pgx 会在首次执行时 Describe 语句并缓存参数类型
	// 信息，后续执行直接用类型信息构建准备好的语句。
	// 该模式不真正在服务端 Prepare，避免了命名冲突风险，同时仍可享受
	// prepared statement 的性能收益（减少解析开销）。
	// 已有手写 prepared statement（tx.Prepare）不会被破坏——QueryExecModeCacheDescribe
	// 只影响隐式查询路径。
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

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
// 使 RLS 策略（tenant_id = get_tenant()）按行隔离数据。
// SET LOCAL 仅作用于当前事务，在连接池下是安全的。
func (p *Pool) WithTenantTx(ctx context.Context, workspaceID int64, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("persistence: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // Commit 后为 no-op

	if _, err := tx.Exec(ctx, "SELECT set_tenant($1)", workspaceID); err != nil {
		return fmt.Errorf("persistence: set tenant: %w", err)
	}
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", workspaceID); err != nil {
		return fmt.Errorf("persistence: set workspace: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("persistence: commit: %w", err)
	}
	return nil
}

// WithTenant 在设置了租户上下文的连接上执行 fn。
// 用于非事务场景（如单次查询），确保 RLS 策略生效。
// 用法：pool.WithTenant(ctx, tenantID, func(conn *pgx.Conn) error { ... })
func (p *Pool) WithTenant(ctx context.Context, tenantID int64, fn func(conn *pgx.Conn) error) error {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("persistence: acquire conn: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT set_tenant($1)", tenantID); err != nil {
		return fmt.Errorf("persistence: set tenant: %w", err)
	}
	if _, err := conn.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", tenantID); err != nil {
		return fmt.Errorf("persistence: set workspace: %w", err)
	}
	return fn(conn.Conn())
}

// Ping 委托给底层连接池（供 /readyz 探活使用）。
func (p *Pool) Ping(ctx context.Context) error { return p.Pool.Ping(ctx) }

// max 返回 a、b 中较大值（Go 1.21+ 标准库内置，此处为兼容旧版本）。
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
