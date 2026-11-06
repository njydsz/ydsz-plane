//go:build integration

// Package dashboard — 仪表盘集成测试（需要真实 PostgreSQL）。
//
// 覆盖：GetDashboard / SaveWidget 的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package dashboard

import (
	"context"
	"testing"
)

// TestGetDashboardIntegration 验证仪表盘聚合查询。
func TestGetDashboardIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool)

	got, err := s.GetDashboard(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}
	if got == nil {
		t.Fatal("GetDashboard: expected non-nil result")
	}
}

// TestWidgetCRUDIntegration 验证组件的创建与清理。
func TestWidgetCRUDIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool)
	ctx := context.Background()

	widget, err := s.SaveWidget(ctx, SaveWidgetInput{
		ProjectID: 1, UserID: 1, WidgetType: WidgetProgressOverview,
		Title: "Regression", GridX: 1, GridY: 2, GridW: 3, GridH: 4,
		Config: map[string]any{"x": 1},
	})
	if err != nil {
		t.Fatalf("SaveWidget: %v", err)
	}
	if widget.ID == 0 {
		t.Fatal("SaveWidget: expected non-zero id")
	}
	t.Cleanup(func() {
		_, _ = s.db.Exec(ctx, `DELETE FROM dashboard_widgets WHERE id = $1`, widget.ID)
	})
}
