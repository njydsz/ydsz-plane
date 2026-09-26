// Package txfixture 提供事务型测试夹具，解决集成测试间的数据残留问题。
//
// 核心思想：在测试开始 begin tx，测试结束时 rollback，
// 测试中所有数据变更随回滚自动撤销，无需手动清理。
//
// 与 t.Cleanup 手动清理对比：
//   - txfixture 不依赖清理 SQL，对复杂数据关系不敏感
//   - 每个测试独立事务，天然支持 t.Parallel()
//   - t.Cleanup 保证即使测试 panic 也会执行 rollback
//
// 注意：txfixture 依赖外部传入 *persistence.Pool，
// 不直接绑定 dockertesthelper，可与任意 DSN 搭配使用（包括已存在的环境变量模式）。
package txfixture

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// TxFixture 封装事务型测试夹具。
//
// 通过 Begin + deferred Rollback 保证隔离性。
// 如果在测试中需要显式提交（极少见），可调用 Commit。
type TxFixture struct {
	pool      *persistence.Pool
	tx        pgx.Tx
	committed bool
}

// NewTxFixture 开启新事务作为测试沙箱。
//
// 使用建议：
//  1. 在测试函数顶部调用：fx, _ := txfixture.NewTxFixture(ctx, pool, t)
//  2. 使用 fx.PgTx() 获取 pgx.Tx 进行数据库操作
//  3. 不需要手动清理——t.Cleanup 自动回滚
//
// 示例：
//
//	func TestWithTxIsolation(t *testing.T) {
//	    pool := /* 由 dockertesthelper 或环境变量创建 */
//	    fx, err := txfixture.NewTxFixture(ctx, pool, t)
//	    if err != nil { t.Fatal(err) }
//	    tx := fx.PgTx()
//	    _, _ = tx.Exec(ctx, "INSERT INTO workspaces ...")
//	    // ... 测试结束时 fx 的 t.Cleanup 自动回滚
//	}
func NewTxFixture(ctx context.Context, pool *persistence.Pool, tb testing.TB) (*TxFixture, error) {
	if pool == nil {
		return nil, fmt.Errorf("txfixture: nil pool")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("txfixture: begin tx: %w", err)
	}

	fx := &TxFixture{pool: pool, tx: tx}

	// 注册 t.Cleanup：测试结束时自动回滚。
	// 即使在 panic 场景下，Go runtime 仍会执行已注册的 Cleanup 函数。
	if tb != nil {
		tb.Cleanup(func() {
			if !fx.committed {
				if err := fx.tx.Rollback(context.Background()); err != nil {
					tb.Logf("txfixture: rollback warning: %v", err)
				}
			}
		})
	}

	return fx, nil
}

// PgTx 返回内部 pgx.Tx，供测试中直接操作数据库。
func (f *TxFixture) PgTx() pgx.Tx {
	return f.tx
}

// Pool 返回出处的连接池（通常不需要直接使用，主要用于获取非事务型操作）。
func (f *TxFixture) Pool() *persistence.Pool {
	return f.pool
}

// Commit 提交事务。
//
// 正常测试中不应该需要 Commit，因为你希望数据自动回滚。
// 仅在特定提交验证场景下使用（例如：测试事务内部 commit 后的行为）。
func (f *TxFixture) Commit() error {
	if f.committed {
		return fmt.Errorf("txfixture: already committed")
	}
	if err := f.tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("txfixture: commit: %w", err)
	}
	f.committed = true
	return nil
}

// Rollback 手动回滚事务。
//
// 通常在 t.Cleanup 中自动调用。
// 如果已提交，此操作会返回 error 但不致命（pgx 对已 commit 的 tx rollback 返回 "tx closed"）。
func (f *TxFixture) Rollback() error {
	if f.committed {
		return nil
	}
	if err := f.tx.Rollback(context.Background()); err != nil {
		return fmt.Errorf("txfixture: rollback: %w", err)
	}
	return nil
}

// ParallelTxFixture 为并行子测试快速创建独立事务沙箱。
//
// 用法：
//
//	t.Run("sub", func(t *testing.T) {
//	    t.Parallel()
//	    fx := txfixture.ParallelTxFixture(ctx, pool, t)
//	    tx := fx.PgTx()
//	    // ...
//	})
//
// 每个子测试拥有独立的 tx，回滚互不影响。
// 注意：所有子测试共享同一个底层 pool，需确保 pool.MaxConns > 并行数。
func ParallelTxFixture(ctx context.Context, pool *persistence.Pool, tb testing.TB) *TxFixture {
	tb.Helper()
	fx, err := NewTxFixture(ctx, pool, tb)
	if err != nil {
		tb.Fatalf("ParallelTxFixture: %v", err)
	}
	return fx
}
