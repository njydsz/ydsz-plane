//go:build integration

// Package version — 版本查询集成测试（需要真实 PostgreSQL）。
//
// 覆盖：List / GetByID 的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package version

import (
	"context"
	"testing"
)

// TestVersionQueriesIntegration 验证版本查询在真实库上的行为。
func TestVersionQueriesIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(Deps{DB: pool.Pool})
	ctx := context.Background()

	versions, total, err := s.List(ctx, ListVersionsOptions{WorkspaceID: 1, ProjectID: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if versions == nil || total < 0 {
		t.Fatal("List: expected version list")
	}

	// 不存在的版本应返回错误。
	if _, err := s.GetByID(ctx, 1, 1); err == nil {
		t.Fatal("GetByID: expected error for missing version")
	}
}
