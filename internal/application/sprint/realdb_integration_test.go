//go:build integration

// Package sprint — 迭代查询集成测试（需要真实 PostgreSQL）。
//
// 覆盖：List / GetBacklog / ListSprintIssues 的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package sprint

import (
	"context"
	"testing"
)

// TestSprintQueriesIntegration 验证迭代查询在真实库上的行为。
func TestSprintQueriesIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool)
	ctx := context.Background()

	if _, total, err := s.List(ctx, ListSprintsOptions{WorkspaceID: 1, ProjectID: 1}); err != nil {
		t.Fatalf("List: %v", err)
	} else if total < 0 {
		t.Fatal("List: expected non-negative sprint total")
	}

	if _, _, err := s.GetBacklog(ctx, 1, 1, 10, 0); err != nil {
		t.Fatalf("GetBacklog: %v", err)
	}

	// 不存在的迭代应返回错误。
	if _, _, err := s.ListSprintIssues(ctx, 1, 1, 10, 0); err == nil {
		t.Fatal("ListSprintIssues: expected error for missing sprint")
	}
}
