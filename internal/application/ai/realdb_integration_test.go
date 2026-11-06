//go:build integration

// Package ai — AI 能力集成测试（需要真实 PostgreSQL）。
//
// 覆盖：DetectDuplicates / SmartAssign 的真实 DB 行为。
// 运行方式见 integration_helper_test.go。
package ai

import (
	"context"
	"testing"
)

// TestDetectDuplicatesIntegration 验证重复检测在真实库上的行为。
func TestDetectDuplicatesIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool, Config{Provider: "fallback"})

	got, err := s.DetectDuplicates(context.Background(), 1, "Alpha requirement", "alpha requirement")
	if err != nil {
		t.Fatalf("DetectDuplicates: %v", err)
	}
	if got == nil {
		t.Fatal("DetectDuplicates: expected non-nil candidates")
	}
}

// TestSmartAssignIntegration 验证智能分配在真实库上的行为。
func TestSmartAssignIntegration(t *testing.T) {
	pool := integrationPool(t)
	s := NewService(pool.Pool, Config{Provider: "fallback"})

	if _, err := s.SmartAssign(context.Background(), SmartAssignInput{
		WorkspaceID: 1, ProjectID: 1, IssueTitle: "alpha", IssueDesc: "alpha", TopN: 3,
	}); err != nil {
		t.Fatalf("SmartAssign: %v", err)
	}
}
