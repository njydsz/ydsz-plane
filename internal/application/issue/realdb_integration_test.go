//go:build integration

// Package issue — 缺陷分析集成测试（需要真实 PostgreSQL）。
//
// 覆盖：GetAnalytics / ExportDefects / GetDefectAge 三个查询的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package issue

import (
	"context"
	"testing"
)

// TestDefectAnalyticsIntegration 验证缺陷分析在真实库上的查询正确性。
func TestDefectAnalyticsIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewDefectAnalyticsService(pool.Pool)
	ctx := context.Background()

	query := AnalyticsQuery{WorkspaceID: 1, ProjectID: 1}

	analytics, err := s.GetAnalytics(ctx, query)
	if err != nil {
		t.Fatalf("GetAnalytics: %v", err)
	}
	if analytics == nil {
		t.Fatal("GetAnalytics: expected non-nil result")
	}

	if _, err := s.ExportDefects(ctx, query, 10); err != nil {
		t.Fatalf("ExportDefects: %v", err)
	}

	age, err := s.GetDefectAge(ctx, query)
	if err != nil {
		t.Fatalf("GetDefectAge: %v", err)
	}
	if age == nil {
		t.Fatal("GetDefectAge: expected non-nil result")
	}
}
