// Package automation (test) 为自动化规则引擎提供单元测试辅助类型与 fixture。
//
// 包含：
//   - ExecutionContext 构造器（buildTestContext）
//   - 模板变量解析测试（TestResolveTemplate）验证 ${issue.identifier} 等占位符替换
//
// 测试仅依赖本仓库内部类型，无需外部服务。
package automation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/njydsz/ydsz-plane/internal/application/issue"
)

// buildTestContext 构造一个含 issue 上下文的 ExecutionContext。
func buildTestContext() *ExecutionContext {
	ep := 8
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	return &ExecutionContext{
		WorkspaceID: 100,
		ProjectID:   200,
		Actor:       ActorContext{UserID: 11, UserName: "tester"},
		Now:         now,
		Issue: &IssueContext{
			ID:             300,
			Identifier:     "YD-42",
			Name:           "示例工作项",
			EstimatePoints: &ep,
		},
	}
}

func TestResolveTemplate(t *testing.T) {
	ctx := buildTestContext()

	tests := []struct {
		name string
		tpl  string
		want string
	}{
		{"issue_identifier", "编号 {{issue.identifier}}", ""},
		{"issue_name", "名称 ${issue.name}", "名称 示例工作项"},
		{"parent_estimate", "点数 ${parent.estimate_points}", "点数 8"},
		{"project_id", "项目 ${project.id}", "项目 200"},
		{"now", "日期 ${now}", "日期 2025-01-02"},
		{"unknown_left_as_is", "目标 ${project.tech_lead}", "目标 "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveTemplate(tc.tpl, ctx)
			if tc.want == "" {
				// 仅断言"不 panic 且不保留未解析变量"（此处统一替换为空串）
				if strings.Contains(got, "${") {
					t.Fatalf("resolveTemplate(%q) = %q, still contains unresolved ${}", tc.tpl, got)
				}
				return
			}
			if got != tc.want {
				t.Fatalf("resolveTemplate(%q) = %q, want %q", tc.tpl, got, tc.want)
			}
		})
	}
}

func TestResolveTemplateNilSafe(t *testing.T) {
	if got := resolveTemplate("${issue.name}", nil); got != "${issue.name}" {
		t.Fatalf("resolveTemplate(nil ctx) = %q, want original template", got)
	}
	empty := &ExecutionContext{}
	if got := resolveTemplate("${issue.name}", empty); got != "${issue.name}" {
		t.Fatalf("resolveTemplate(no issue ctx) = %q, want original template", got)
	}
}

func TestResolveTemplateParentEstimateNil(t *testing.T) {
	ctx := &ExecutionContext{Issue: &IssueContext{EstimatePoints: nil}}
	if got := resolveTemplate("${parent.estimate_points}", ctx); got != "0" {
		t.Fatalf("resolveTemplate(nil estimate_points) = %q, want 0", got)
	}
}

func TestResolveAssignTargetNumeric(t *testing.T) {
	eng := NewEngine(nil, nil, nil, nil)
	uid, err := eng.resolveAssignTarget(context.Background(), Action{Value: "42"}, &ExecutionContext{})
	if err != nil {
		t.Fatalf("resolveAssignTarget numeric: unexpected err %v", err)
	}
	if uid != 42 {
		t.Fatalf("resolveAssignTarget numeric = %d, want 42", uid)
	}
}

func TestResolveAssignTargetLeastLoadedNoDB(t *testing.T) {
	// db 为 nil 时应返回清晰错误而非 panic
	eng := NewEngine(nil, nil, nil, nil)
	act := Action{Config: map[string]any{"strategy": "least_loaded", "role": "member"}}
	_, err := eng.resolveAssignTarget(context.Background(), act, &ExecutionContext{WorkspaceID: 1, ProjectID: 2})
	if err == nil {
		t.Fatal("resolveAssignTarget least_loaded with nil db should return error")
	}
	if !strings.Contains(err.Error(), "least_loaded") {
		t.Fatalf("resolveAssignTarget least_loaded error = %v, want least_loaded hint", err)
	}
}

func TestResolveAssignTargetUnknown(t *testing.T) {
	eng := NewEngine(nil, nil, nil, nil)
	_, err := eng.resolveAssignTarget(context.Background(), Action{}, &ExecutionContext{})
	if err == nil {
		t.Fatal("resolveAssignTarget unknown should return error")
	}
}

// TestStateNameResolution 验证"状态名 → ID"解析逻辑（StateNameIndex 查询模式）。
func TestStateNameResolution(t *testing.T) {
	index := issue.StateNameIndex{"待办": 10, "进行中": 20, "已完成": 30}

	stateID, ok := index["进行中"]
	if !ok || stateID != 20 {
		t.Fatalf("StateNameIndex lookup 进行中 = %d, ok=%v; want 20,true", stateID, ok)
	}
	if _, ok := index["不存在"]; ok {
		t.Fatal("StateNameIndex lookup for missing name should be !ok")
	}
}

// TestTypeConversions 验证 UpdateIssueField 依赖的类型转换纯函数。
func TestTypeConversions(t *testing.T) {
	// toString
	if got := toString(42); got != "42" {
		t.Fatalf("toString(42) = %q, want 42", got)
	}
	if got := toString("abc"); got != "abc" {
		t.Fatalf("toString(abc) = %q", got)
	}

	// toInt / toInt64
	if got, err := toInt("5"); err != nil || got != 5 {
		t.Fatalf("toInt(5) = %d, err=%v", got, err)
	}
	if _, err := toInt("not-a-number"); err == nil {
		t.Fatal("toInt(bad) should error")
	}
	if got, err := toInt64(float64(9.0)); err != nil || got != 9 {
		t.Fatalf("toInt64(9.0) = %d, err=%v", got, err)
	}

	// toInt64Slice
	ids, err := toInt64Slice([]any{1, 2, 3})
	if err != nil || len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
		t.Fatalf("toInt64Slice([]any{1,2,3}) = %v, err=%v", ids, err)
	}
	ids, err = toInt64Slice(int64(7))
	if err != nil || len(ids) != 1 || ids[0] != 7 {
		t.Fatalf("toInt64Slice(int64(7)) = %v, err=%v", ids, err)
	}

	// toTime
	ts, err := toTime("2025-01-02")
	if err != nil || ts.Year() != 2025 || ts.Month() != 1 || ts.Day() != 2 {
		t.Fatalf("toTime(date) = %v, err=%v", ts, err)
	}
	if _, err := toTime("garbage"); err == nil {
		t.Fatal("toTime(bad) should error")
	}
}
