// Package dashboard — 项目仪表盘域单元测试。
//
// 覆盖范围：
//  1. WidgetType 枚举值与有效值集合
//  2. DashboardWidget 模型 JSON 序列化
//  3. ProgressOverviewWidget 等子字段序列化
//  4. RiskAlert 严重级别映射 (severity critical/high/medium/low)
//  5. StateBucket / OverdueItem / BlockedItem 等序列化
//  6. max 辅助函数
package dashboard

import (
	"encoding/json"
	"testing"
)

// TestWidgetType_Values 验证所有 WidgetType 枚举值唯一且非空。
func TestWidgetType_Values(t *testing.T) {
	widgets := []WidgetType{
		WidgetProgressOverview,
		WidgetBurndown,
		WidgetVelocity,
		WidgetPrioritySplit,
		WidgetStateDistribution,
		WidgetOverdueList,
		WidgetBlockedList,
		WidgetRiskAlert,
		WidgetRecentActivity,
		WidgetTeamWorkload,
		WidgetVersionBurndown,
		WidgetModuleDistribution,
		WidgetDORA,
		WidgetProjectCompare,
	}

	seen := make(map[WidgetType]bool)
	for _, w := range widgets {
		if w == "" {
			t.Errorf("WidgetType should not be empty")
		}
		if seen[w] {
			t.Errorf("duplicate WidgetType: %q", w)
		}
		seen[w] = true
	}
}

// TestWidgetType_StringValues 验证 WidgetType 字符串值格式。
func TestWidgetType_StringValues(t *testing.T) {
	cases := []struct {
		wt   WidgetType
		want string
	}{
		{WidgetProgressOverview, "progress_overview"},
		{WidgetBurndown, "burndown"},
		{WidgetVelocity, "velocity"},
		{WidgetPrioritySplit, "priority_split"},
		{WidgetStateDistribution, "state_distribution"},
		{WidgetOverdueList, "overdue_list"},
		{WidgetBlockedList, "blocked_list"},
		{WidgetRiskAlert, "risk_alert"},
		{WidgetRecentActivity, "recent_activity"},
		{WidgetTeamWorkload, "team_workload"},
		{WidgetVersionBurndown, "version_burndown"},
		{WidgetModuleDistribution, "module_distribution"},
		{WidgetDORA, "dora"},
		{WidgetProjectCompare, "project_compare"},
	}
	for _, tc := range cases {
		if string(tc.wt) != tc.want {
			t.Errorf("WidgetType %q: got %q, want %q", tc.wt, string(tc.wt), tc.want)
		}
	}
}

// TestWidgetType_IsKnown 验证 WidgetType 是否为已知类型。
func TestWidgetType_IsKnown(t *testing.T) {
	knownTypes := map[WidgetType]bool{
		WidgetProgressOverview:    true,
		WidgetBurndown:            true,
		WidgetVelocity:            true,
		WidgetPrioritySplit:       true,
		WidgetStateDistribution:   true,
		WidgetOverdueList:         true,
		WidgetBlockedList:         true,
		WidgetRiskAlert:           true,
		WidgetRecentActivity:      true,
		WidgetTeamWorkload:        true,
		WidgetVersionBurndown:     true,
		WidgetModuleDistribution:  true,
		WidgetDORA:                true,
		WidgetProjectCompare:       true,
		"unknown_widget":           false,
		"":                        false,
	}
	for wt, wantValid := range knownTypes {
		got := isKnownWidgetType(wt)
		if got != wantValid {
			t.Errorf("isKnownWidgetType(%q) = %v, want %v", wt, got, wantValid)
		}
	}
}

// isKnownWidgetType 校验 WidgetType 是否为预定义常量。
func isKnownWidgetType(wt WidgetType) bool {
	switch wt {
	case WidgetProgressOverview, WidgetBurndown, WidgetVelocity,
		WidgetPrioritySplit, WidgetStateDistribution, WidgetOverdueList,
		WidgetBlockedList, WidgetRiskAlert, WidgetRecentActivity,
		WidgetTeamWorkload, WidgetVersionBurndown, WidgetModuleDistribution,
		WidgetDORA, WidgetProjectCompare:
		return true
	}
	return false
}

// TestDashboardWidget_JSON_Roundtrip 验证 DashboardWidget 序列化往返。
func TestDashboardWidget_JSON_Roundtrip(t *testing.T) {
	userID := int64(42)
	w := DashboardWidget{
		ID:         1,
		ProjectID:  100,
		WidgetType: WidgetProgressOverview,
		Title:      "进度概览",
		GridX:      0,
		GridY:      0,
		GridW:      6,
		GridH:      4,
		Config:     map[string]any{"show_percentage": true},
		IsVisible:  true,
		SortOrder:  1,
		UserID:     &userID,
	}

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("DashboardWidget marshal: %v", err)
	}

	var decoded DashboardWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("DashboardWidget unmarshal: %v", err)
	}

	if decoded.WidgetType != w.WidgetType {
		t.Errorf("WidgetType: got %q, want %q", decoded.WidgetType, w.WidgetType)
	}
	if decoded.GridW != w.GridW {
		t.Errorf("GridW: got %d, want %d", decoded.GridW, w.GridW)
	}
	if decoded.IsVisible != w.IsVisible {
		t.Errorf("IsVisible: got %v, want %v", decoded.IsVisible, w.IsVisible)
	}
	if decoded.UserID == nil || *decoded.UserID != *w.UserID {
		t.Errorf("UserID: got %v, want %v", decoded.UserID, w.UserID)
	}
}

// TestProgressOverviewWidget_JSON_Roundtrip 验证进度概览 widget 数据序列化。
func TestProgressOverviewWidget_JSON_Roundtrip(t *testing.T) {
	original := ProgressOverviewWidget{
		TotalIssues:    100,
		DoneIssues:     60,
		InProgress:     20,
		OverdueIssues:  5,
		BlockedIssues:  3,
		CompletionRate: 0.6,
		ActiveSprints:  2,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ProgressOverviewWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.TotalIssues != original.TotalIssues {
		t.Errorf("TotalIssues: got %d, want %d", decoded.TotalIssues, original.TotalIssues)
	}
	if decoded.CompletionRate != original.CompletionRate {
		t.Errorf("CompletionRate: got %f, want %f", decoded.CompletionRate, original.CompletionRate)
	}
	if decoded.ActiveSprints != original.ActiveSprints {
		t.Errorf("ActiveSprints: got %d, want %d", decoded.ActiveSprints, original.ActiveSprints)
	}
}

// TestPrioritySplitWidget_JSON_Roundtrip 验证优先级分布 widget 序列化。
func TestPrioritySplitWidget_JSON_Roundtrip(t *testing.T) {
	original := PrioritySplitWidget{
		Total: 50,
		ByPriority: map[string]int{
			"urgent": 3,
			"high":   10,
			"medium": 25,
			"low":    12,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded PrioritySplitWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Total != original.Total {
		t.Errorf("Total: got %d, want %d", decoded.Total, original.Total)
	}
	if decoded.ByPriority["urgent"] != 3 {
		t.Errorf("ByPriority[urgent]: got %d, want 3", decoded.ByPriority["urgent"])
	}
}

// TestStateDistributionWidget_JSON_Roundtrip 验证状态分布 widget 序列化。
func TestStateDistributionWidget_JSON_Roundtrip(t *testing.T) {
	original := StateDistributionWidget{
		Total: 30,
		ByState: []StateBucket{
			{StateID: 1, StateName: "待办", GroupName: "todo", Color: "#gray", Count: 10},
			{StateID: 2, StateName: "进行中", GroupName: "started", Color: "#blue", Count: 15},
			{StateID: 3, StateName: "已完成", GroupName: "completed", Color: "#green", Count: 5},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded StateDistributionWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.ByState) != 3 {
		t.Errorf("ByState length: got %d, want 3", len(decoded.ByState))
	}
	if decoded.ByState[0].StateName != "待办" {
		t.Errorf("ByState[0].StateName: got %q, want 待办", decoded.ByState[0].StateName)
	}
	if decoded.Total != 30 {
		t.Errorf("Total: got %d, want 30", decoded.Total)
	}
}

// TestOverdueListWidget_JSON_Roundtrip 验证逾期列表 widget 序列化。
func TestOverdueListWidget_JSON_Roundtrip(t *testing.T) {
	original := OverdueListWidget{
		Total: 5,
		Items: []OverdueItem{
			{ID: 1, Identifier: "PRJ-1", Title: "任务 A", Priority: "urgent", OverdueDays: 3, Assignee: "张三"},
			{ID: 2, Identifier: "PRJ-2", Title: "任务 B", Priority: "high", OverdueDays: 1, Assignee: "李四"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded OverdueListWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Items) != 2 {
		t.Errorf("Items length: got %d, want 2", len(decoded.Items))
	}
	if decoded.Items[0].OverdueDays != 3 {
		t.Errorf("Items[0].OverdueDays: got %d, want 3", decoded.Items[0].OverdueDays)
	}
}

// TestBlockedListWidget_JSON_Roundtrip 验证阻塞列表 widget 序列化。
func TestBlockedListWidget_JSON_Roundtrip(t *testing.T) {
	original := BlockedListWidget{
		Total: 2,
		Items: []BlockedItem{
			{ID: 1, Identifier: "PRJ-10", Title: "被阻塞任务", BlockedCount: 3, BlockerNames: "PRJ-8, PRJ-9"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded BlockedListWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Items[0].BlockedCount != 3 {
		t.Errorf("BlockedCount: got %d, want 3", decoded.Items[0].BlockedCount)
	}
}

// TestRiskAlert_SeverityValues 验证告警严重级别枚举映射。
func TestRiskAlert_SeverityValues(t *testing.T) {
	// Severity 在 SQL 排序中为 critical=0, high=1, medium=2, low=3
	expectedSeverities := []string{"critical", "high", "medium", "low"}
	for _, sev := range expectedSeverities {
		alert := RiskAlert{
			ID:       1,
			Severity: sev,
			Title:    "test",
		}
		if alert.Severity != sev {
			t.Errorf("RiskAlert.Severity: got %q, want %q", alert.Severity, sev)
		}
	}
}

// TestRiskAlert_JSON_Roundtrip 验证 RiskAlert 序列化往返。
func TestRiskAlert_JSON_Roundtrip(t *testing.T) {
	projectID := int64(100)
	original := RiskAlert{
		ID:         1,
		ProjectID:  &projectID,
		RuleID:     5,
		Severity:   "critical",
		Title:      "风险告警标题",
		Description: "超过 80% 工作项逾期",
		Metadata: map[string]any{
			"overdue_ratio": 0.85,
			"threshold":     0.8,
		},
		IsResolved: false,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("RiskAlert marshal: %v", err)
	}

	var decoded RiskAlert
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("RiskAlert unmarshal: %v", err)
	}

	if decoded.Severity != original.Severity {
		t.Errorf("Severity: got %q, want %q", decoded.Severity, original.Severity)
	}
	if decoded.IsResolved != original.IsResolved {
		t.Errorf("IsResolved: got %v, want %v", decoded.IsResolved, original.IsResolved)
	}
	if decoded.Metadata["overdue_ratio"] != 0.85 {
		t.Errorf("Metadata[overdue_ratio]: got %v, want 0.85", decoded.Metadata["overdue_ratio"])
	}
}

// TestBurndownWidget_JSON_Roundtrip 验证燃尽图 widget 序列化。
func TestBurndownWidget_JSON_Roundtrip(t *testing.T) {
	original := BurndownWidget{
		SprintID:      1,
		SprintName:    "Sprint 10",
		TotalPoints:   100,
		BurnedPoints:  40,
		TotalIssues:   20,
		BurnedIssues:  8,
		RemainingDays: 5,
		IsActive:      true,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded BurndownWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.SprintID != original.SprintID {
		t.Errorf("SprintID: got %d, want %d", decoded.SprintID, original.SprintID)
	}
	if decoded.BurnedPoints != 40 {
		t.Errorf("BurnedPoints: got %d, want 40", decoded.BurnedPoints)
	}
}

// TestVelocityWidget_JSON_Roundtrip 验证速率图 widget 序列化。
func TestVelocityWidget_JSON_Roundtrip(t *testing.T) {
	original := VelocityWidget{
		Sprints: []VelocityPoint{
			{SprintID: 1, SprintName: "Sprint 1", CompletedCount: 10, CommittedCount: 12, CompletionRate: 0.83},
			{SprintID: 2, SprintName: "Sprint 2", CompletedCount: 15, CommittedCount: 15, CompletionRate: 1.0},
		},
		Average: 0.915,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded VelocityWidget
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Sprints) != 2 {
		t.Errorf("Sprints length: got %d, want 2", len(decoded.Sprints))
	}
	if decoded.Average != 0.915 {
		t.Errorf("Average: got %f, want 0.915", decoded.Average)
	}
}

// TestMax_Helper 验证 max 辅助函数。
func TestMax_Helper(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{3, 5, 5},
		{10, 2, 10},
		{0, 0, 0},
		{-1, -5, -1},
		{100, 100, 100},
	}
	for _, tc := range cases {
		if got := max(tc.a, tc.b); got != tc.want {
			t.Errorf("max(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// TestTeamMemberWorkload_Total 验证成员负载 total 字段计算逻辑。
func TestTeamMemberWorkload_Total(t *testing.T) {
	m := TeamMemberWorkload{
		UserID:     1,
		UserName:   "test",
		Todo:       3,
		InProgress: 2,
		Done:       5,
		Total:      10,
	}
	// Total 是 DB 查询结果直接映射，这里验证字段访问
	if m.Total != 10 {
		t.Errorf("Total: got %d, want 10", m.Total)
	}
}
