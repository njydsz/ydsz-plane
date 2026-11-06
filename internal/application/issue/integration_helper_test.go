//go:build integration

// Package issue — 集成测试共享辅助（需要真实 PostgreSQL）。
//
// 所有集成测试统一通过 YDSZ_TEST_DATABASE_URL 获取测试库连接，
// 未设置时自动跳过，保证 `go test ./...`（无 integration 标签）不受影响。
//
// 运行方式：
//
//	YDSZ_TEST_DATABASE_URL="postgres://ydsz:ydsz@localhost:5432/ydsz_plane_test" \
//		go test -tags=integration ./internal/application/issue/...
package issue

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
