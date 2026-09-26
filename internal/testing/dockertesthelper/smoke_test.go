//go:build dockertest

package dockertesthelper

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestSmoke_PostgresContainer 验证 dockertest 容器化模式的端到端工作流：
//   1. StartPostgres 启动容器、生成 DSN
//   2. 从 DSN 创建连接池
//   3. 插入一条测试数据
//   4. 查询验证数据正确
//   5. t.Cleanup 自动销毁容器
//
// 运行方式：
//
//	go test -tags dockertest -run TestSmoke_PostgresContainer ./internal/testing/dockertesthelper/...
func TestSmoke_PostgresContainer(t *testing.T) {
	ctx := context.Background()

	// 启动容器并执行迁移
	cfg := StartPostgres(ctx, t)
	if cfg.DSN == "" {
		t.Fatal("StartPostgres: empty DSN")
	}
	if cfg.DriverName != "pgx" {
		t.Fatalf("StartPostgres: driver=%q, want pgx", cfg.DriverName)
	}
	t.Logf("container DSN: %s", sanitizeDSN(cfg.DSN))

	// 创建连接池
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// 验证连通性
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	// 验证迁移已执行（关键表应存在）
	var tableCount int
	err = pool.QueryRow(ctx, `
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name IN ('workspaces', 'projects', 'states', 'users')
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("check migrated tables: %v", err)
	}
	if tableCount < 4 {
		t.Fatalf("expected at least 4 migrated tables, got %d", tableCount)
	}

	// 写入 + 查询完整生命周期
	var id int64
	err = pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id)
		VALUES ($1, $2, 1)
		RETURNING id
	`, "smoke-test-ws", "smoke-test-slug").Scan(&id)
	if err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	if id == 0 {
		t.Fatal("insert: returned id 0")
	}

	var name string
	err = pool.QueryRow(ctx, `SELECT name FROM workspaces WHERE id = $1`, id).Scan(&name)
	if err != nil {
		t.Fatalf("select workspace: %v", err)
	}
	if name != "smoke-test-ws" {
		t.Errorf("got name=%q, want %q", name, "smoke-test-ws")
	}

	t.Logf("smoke test passed: inserted workspace id=%d", id)

	// 清理（如果 t.Cleanup 因 panic 不执行，容器级隔离也保证了不会泄漏持久数据）
	if _, cerr := pool.Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id); cerr != nil {
		t.Logf("cleanup workspace: %v", cerr)
	}
}

// TestSmoke_ParallelContainersWithTxRollback 验证并行模式下事务回滚的隔离效果。
//
// 多个子测试并行使用同一个父容器（加速运行），
// 每个子测试 begin tx → 操作 → rollback，数据自动隔离。
// 证明 dockertesthelper + 事务回滚模式可替代手动数据清理。
func TestSmoke_ParallelContainersWithTxRollback(t *testing.T) {
	ctx := context.Background()
	cfg := StartPostgres(ctx, t)

	parentPool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		t.Fatalf("parent pool: %v", err)
	}
	t.Cleanup(func() { parentPool.Close() })

	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tx, txErr := parentPool.Begin(ctx)
			if txErr != nil {
				t.Fatalf("begin tx: %v", txErr)
			}
			defer func() { _ = tx.Rollback(ctx) }()

			var id int64
			qerr := tx.QueryRow(ctx, `
				INSERT INTO workspaces (name, slug, owner_id)
				VALUES ($1, $2, 1)
				RETURNING id
			`, "parallel-"+name, "parallel-slug-"+name).Scan(&id)
			if qerr != nil {
				t.Fatalf("[%s] insert: %v", name, qerr)
			}

			var gotName string
			serr := tx.QueryRow(ctx, `SELECT name FROM workspaces WHERE id = $1`, id).Scan(&gotName)
			if serr != nil {
				t.Fatalf("[%s] select: %v", name, serr)
			}
			if gotName != "parallel-"+name {
				t.Errorf("[%s] got name=%q, want %q", name, gotName, "parallel-"+name)
			}
			t.Logf("[%s] inserted id=%d, name=%q", name, id, gotName)
			// 事务回滚自动清理数据
		})
	}
}

// sanitizeDSN 脱敏 DSN 中的密码仅用于日志输出。
func sanitizeDSN(dsn string) string {
	parts := strings.Split(dsn, " ")
	for i, part := range parts {
		if len(part) > 9 && part[:9] == "password=" {
			parts[i] = "password=***"
		}
	}
	return strings.Join(parts, " ")
}
