//go:build integration

// Package metrics — 度量查询集成测试（需要真实 PostgreSQL）。
//
// 覆盖：GetLeadTime / GetQualityMetrics / GetDORA / GetResourceLoadDetail /
// GetWeeklyThroughput / CFDCalculator / ControlChartCalculator。
// 运行方式见 integration_helper_test.go。
package metrics

import (
	"context"
	"testing"
	"time"
)

// TestMetricsQueriesIntegration 验证全部度量查询在真实库上的行为。
func TestMetricsQueriesIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool)
	ctx := context.Background()

	if _, err := s.GetLeadTime(ctx, 1, 1, 30); err != nil {
		t.Fatalf("GetLeadTime: %v", err)
	}
	if _, err := s.GetQualityMetrics(ctx, 1, 1); err != nil {
		t.Fatalf("GetQualityMetrics: %v", err)
	}
	if _, err := s.GetDORA(ctx, 1, 1); err != nil {
		t.Fatalf("GetDORA: %v", err)
	}
	if _, err := s.GetResourceLoadDetail(ctx, 1, 1); err != nil {
		t.Fatalf("GetResourceLoadDetail: %v", err)
	}
	if _, err := s.GetWeeklyThroughput(ctx, 1, 1, 4); err != nil {
		t.Fatalf("GetWeeklyThroughput: %v", err)
	}
	if _, err := NewCFDCalculator(s.db).Calculate(ctx, 1, 1, time.Now().AddDate(0, 0, -3), time.Now()); err != nil {
		t.Fatalf("CFD: %v", err)
	}
	if _, err := NewControlChartCalculator(s.db).Calculate(ctx, 1, 1, 30); err != nil {
		t.Fatalf("ControlChart: %v", err)
	}
}
