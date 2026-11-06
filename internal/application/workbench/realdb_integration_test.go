//go:build integration

// Package workbench — 工作台查询集成测试（需要真实 PostgreSQL）。
//
// 覆盖：GetSummary / GetEfficiency / GetFeed 的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package workbench

import (
	"context"
	"testing"
)

// TestGetSummaryIntegration 验证工作台概览查询。
func TestGetSummaryIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool)

	sum, err := s.GetSummary(context.Background(), 1, 1, nil)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if sum == nil {
		t.Fatal("GetSummary: expected non-nil result")
	}
}

// TestGetEfficiencyIntegration 验证效率报表查询。
func TestGetEfficiencyIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool)

	rep, err := s.GetEfficiency(context.Background(), 1, 1, nil)
	if err != nil {
		t.Fatalf("GetEfficiency: %v", err)
	}
	if rep == nil {
		t.Fatal("GetEfficiency: expected non-nil result")
	}
}

// TestGetFeedIntegration 验证动态流查询。
func TestGetFeedIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool)

	feed, err := s.GetFeed(context.Background(), 1, 1, nil, 20, 0)
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if feed == nil {
		t.Fatal("GetFeed: expected non-nil result")
	}
}
