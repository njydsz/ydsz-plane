//go:build integration

// Package version — 集成测试共享辅助（需要真实 PostgreSQL）。
//
// 统一通过 YDSZ_TEST_DATABASE_URL 获取测试库连接，未设置时自动跳过。
package version

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// integrationPool 返回测试库连接；未配置时跳过测试。
//
// 返回 *pgxpool.Pool（而非 *persistence.Pool），与各域 Service 的
// 构造函数签名保持一致，调用方无需二次解引用。
func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
