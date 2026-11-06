//go:build integration

// Package sprint — 集成测试共享辅助（需要真实 PostgreSQL）。
//
// 统一通过 YDSZ_TEST_DATABASE_URL 获取测试库连接，未设置时自动跳过。
package sprint

import (
	"context"
	"os"
	"testing"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

// integrationPool 返回测试库连接池；未配置时跳过测试。
func integrationPool(t *testing.T) *persistence.Pool {
	t.Helper()
	dsn := os.Getenv("YDSZ_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("YDSZ_TEST_DATABASE_URL not set, skipping integration test")
	}
	pool, err := persistence.NewPool(context.Background(), dsn, 4)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
