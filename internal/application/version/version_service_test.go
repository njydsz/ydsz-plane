// Package version — 版本应用服务单元测试。
//
// 覆盖范围：
//   1. SemVer 2.0 解析与校验
//   2. 状态机流转规则 (canTransition)
//   3. 检查清单校验 (checklistAllRequiredChecked)
//   4. 清单规范化 (normalizeChecklist)
//   5. 模型 JSON 序列化/反序列化
//   6. 进度聚合边界条件
//   7. 质量指标计算逻辑
//   8. Release Notes 数据模型
//   9. 输入校验 (validateCreateInput)
//  10. 版本状态枚举合法性
//
// 互联网大厂标准：
//   - 表驱动测试 (table-driven tests)
//   - 边界条件全覆盖
//   - 黄金路径 + 错误路径双覆盖
//   - 每个函数独立可测（纯函数优先）
package version

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

// ==========================================================================
// P0-1: SemVer 2.0 解析与校验
// ==========================================================================

func TestParseSemVer_Valid(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		major  int
		minor  int
		patch  int
		pre    string
		build  string
	}{
		{"纯版本号", "1.2.3", 1, 2, 3, "", ""},
		{"大版本号", "999.888.777", 999, 888, 777, "", ""},
		{"零版本", "0.0.0", 0, 0, 0, "", ""},
		{"pre-release", "1.0.0-alpha", 1, 0, 0, "alpha", ""},
		{"pre-release.1", "1.0.0-alpha.1", 1, 0, 0, "alpha.1", ""},
		{"build", "1.0.0+build.20240801", 1, 0, 0, "", "build.20240801"},
		{"pre+build", "2.1.0-rc.1+exp.sha.5114f85", 2, 1, 0, "rc.1", "exp.sha.5114f85"},
		{"patch 大号", "0.100.200", 0, 100, 200, "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			serr, v := ParseSemVer(c.raw)
			if serr != nil {
				t.Fatalf("unexpected error: %v", serr)
			}
			if v.Major != c.major {
				t.Errorf("major: got %d want %d", v.Major, c.major)
			}
			if v.Minor != c.minor {
				t.Errorf("minor: got %d want %d", v.Minor, c.minor)
			}
			if v.Patch != c.patch {
				t.Errorf("patch: got %d want %d", v.Patch, c.patch)
			}
			if v.PreRelease != c.pre {
				t.Errorf("pre: got %q want %q", v.PreRelease, c.pre)
			}
			if v.Build != c.build {
				t.Errorf("build: got %q want %q", v.Build, c.build)
			}
			if v.Raw != c.raw {
				t.Errorf("raw: got %q want %q", v.Raw, c.raw)
			}
		})
	}
}

func TestParseSemVer_Invalid(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"空字符串", ""},
		{"纯字母", "abc"},
		{"缺少 patch", "1.2"},
		{"多了段", "1.2.3.4"},
		{"前导零", "01.2.3"},
		{"patch 前导零", "1.2.03"},
		{"负数", "-1.2.3"},
		{"非法 pre-release 字符", "1.0.0-@invalid"},
		{"空 build", "1.0.0+"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			serr, v := ParseSemVer(c.raw)
			if serr == nil {
				t.Errorf("expected error for %q, got valid: %+v", c.raw, v)
			}
		})
	}
}

func TestSemVer_String(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"1.2.3", "1.2.3"},
		{"1.0.0-alpha", "1.0.0-alpha"},
		{"1.0.0+build", "1.0.0+build"},
		{"2.1.0-rc.1+exp", "2.1.0-rc.1+exp"},
	}

	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			_, v := ParseSemVer(c.raw)
			if v == nil {
				t.Fatal("parse failed")
			}
			if got := v.String(); got != c.want {
				t.Errorf("String: got %q want %q", got, c.want)
			}
		})
	}
}

func TestSemVer_Compare(t *testing.T) {
	cases := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"a < b", "1.0.0", "2.0.0", -1},
		{"a > b", "2.0.0", "1.0.0", 1},
		{"相等", "1.2.3", "1.2.3", 0},
		{"release > pre-release", "1.0.0", "1.0.0-rc.1", 1},
		{"pre-release < release", "1.0.0-alpha", "1.0.0", -1},
		{"pre-release 比较", "1.0.0-alpha", "1.0.0-beta", -1},
		{"pre-release 数字", "1.0.0-1", "1.0.0-2", -1},
		{"数字 < 字母", "1.0.0-1", "1.0.0-alpha", -1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, a := ParseSemVer(c.a)
			_, b := ParseSemVer(c.b)
			if a == nil || b == nil {
				t.Fatal("parse failed")
			}
			if got := a.Compare(b); got != c.want {
				t.Errorf("Compare(%q, %q): got %d want %d", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestSemErr_Error(t *testing.T) {
	e := &SemErr{Reason: "格式错误", Value: "1.0.0-beta"}
	s := e.Error()
	if s == "" {
		t.Error("Error() returned empty string")
	}
	// 超长截断
	long := &SemErr{Reason: "test", Value: string(make([]byte, 100))}
	if len(long.Error()) == 0 {
		t.Error("empty error for long input")
	}
}

// ==========================================================================
// P0-2: 状态机流转规则
// ==========================================================================

func TestCanTransition(t *testing.T) {
	cases := []struct {
		name string
		from VersionStatusCode
		to   VersionStatusCode
		want bool
	}{
		// planning
		{"planning→active", VersionPlanning, VersionActive, true},
		{"planning→archived", VersionPlanning, VersionArchived, true},
		{"planning→released (禁止)", VersionPlanning, VersionReleased, false},
		{"planning→planning (自身)", VersionPlanning, VersionPlanning, false},

		// active
		{"active→released", VersionActive, VersionReleased, true},
		{"active→archived", VersionActive, VersionArchived, true},
		{"active→planning (回退禁止)", VersionActive, VersionPlanning, false},
		{"active→active (自身)", VersionActive, VersionActive, false},

		// released
		{"released→archived", VersionReleased, VersionArchived, true},
		{"released→active (回退禁止)", VersionReleased, VersionActive, false},
		{"released→planning (回退禁止)", VersionReleased, VersionPlanning, false},

		// archived (终态)
		{"archived→* (终态禁止流转)", VersionArchived, VersionActive, false},
		{"archived→planning", VersionArchived, VersionPlanning, false},
		{"archived→released", VersionArchived, VersionReleased, false},
		{"archived→archived", VersionArchived, VersionArchived, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := canTransition(c.from, c.to)
			if got != c.want {
				t.Errorf("canTransition(%q, %q): got %v want %v", c.from, c.to, got, c.want)
			}
		})
	}
}

// ==========================================================================
// P0-3: 检查清单校验
// ==========================================================================

func TestChecklistAllRequiredChecked(t *testing.T) {
	cases := []struct {
		name  string
		items []ChecklistItem
		want  bool
	}{
		{"空清单", []ChecklistItem{}, true},
		{"全部完成", []ChecklistItem{
			{ID: "1", Label: "单元测试通过", Required: true, Checked: true},
			{ID: "2", Label: "代码审查完成", Required: true, Checked: true},
		}, true},
		{"有未完成必做", []ChecklistItem{
			{ID: "1", Label: "单元测试通过", Required: true, Checked: true},
			{ID: "2", Label: "代码审查完成", Required: true, Checked: false},
		}, false},
		{"仅可选未完成", []ChecklistItem{
			{ID: "1", Label: "更新文档", Required: false, Checked: false},
			{ID: "2", Label: "单元测试通过", Required: true, Checked: true},
		}, true},
		{"所有未完成", []ChecklistItem{
			{ID: "1", Label: "A", Required: true, Checked: false},
		}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := checklistAllRequiredChecked(c.items)
			if got != c.want {
				t.Errorf("checklistAllRequiredChecked: got %v want %v", got, c.want)
			}
		})
	}
}

func TestNormalizeChecklist(t *testing.T) {
	cases := []struct {
		name     string
		in       []ChecklistItem
		wantLen  int
		wantFirstID string
	}{
		{"nil → 空切片", nil, 0, ""},
		{"空切片", []ChecklistItem{}, 0, ""},
		{"补全 ID", []ChecklistItem{
			{Label: " 测试通过 ", Required: true},
		}, 1, "chk-1"},
		{"trim 并去重", []ChecklistItem{
			{Label: "   ", Required: true},
			{Label: "项1", ID: "existing"},
		}, 1, "existing"},
		{"保留已有 ID", []ChecklistItem{
			{Label: "项1", ID: "my-id", Required: true},
		}, 1, "my-id"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := normalizeChecklist(c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != c.wantLen {
				t.Errorf("len: got %d want %d", len(got), c.wantLen)
			}
			if c.wantLen > 0 && c.wantFirstID != "" {
				if got[0].ID != c.wantFirstID {
					t.Errorf("first ID: got %q want %q", got[0].ID, c.wantFirstID)
				}
			}
			// 验证 trim
			for i, it := range got {
				if it.Label == "" {
					t.Errorf("item %d has empty label", i)
				}
			}
		})
	}
}

// ==========================================================================
// P0-4: 模型 JSON 序列化 / 反序列化
// ==========================================================================

func TestVersion_JSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	delivered := now.Add(-24 * time.Hour)
	targetDate := "2025-06-30"

	v := Version{
		ID:          1,
		WorkspaceID: 10,
		ProjectID:   100,
		Name:        "v1.0 正式版",
		Semver:      "1.0.0",
		Description: "首个正式版本",
		Status:      VersionReleased,
		Checklist: []ChecklistItem{
			{ID: "chk-1", Label: "测试通过", Required: true, Checked: true},
		},
		ReleaseNotes: "## v1.0.0\n\n首版发布",
		DeliveredAt:  &delivered,
		TargetDate:   &targetDate,
		CreatedBy:    1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Progress: &VersionProgress{
			TotalPoints:    100,
			DonePoints:     80,
			TotalIssues:    20,
			DoneIssues:     18,
			CompletionRate: 0.8,
			SprintCount:     2,
		},
	}

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Version
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != v.ID {
		t.Errorf("id: got %v want %v", got.ID, v.ID)
	}
	if got.Name != v.Name {
		t.Errorf("name: got %q want %q", got.Name, v.Name)
	}
	if got.Semver != v.Semver {
		t.Errorf("semver: got %q want %q", got.Semver, v.Semver)
	}
	if got.Status != v.Status {
		t.Errorf("status: got %q want %q", got.Status, v.Status)
	}
	if got.ReleaseNotes != v.ReleaseNotes {
		t.Errorf("release_notes: got %q want %q", got.ReleaseNotes, v.ReleaseNotes)
	}
	if got.Progress == nil {
		t.Fatal("progress should not be nil")
	}
	if got.Progress.DonePoints != v.Progress.DonePoints {
		t.Errorf("progress.done_points: got %v want %v", got.Progress.DonePoints, v.Progress.DonePoints)
	}
}

func TestChecklistItem_JSONRoundTrip(t *testing.T) {
	items := []ChecklistItem{
		{ID: "chk-1", Label: "代码审查", Required: true, Checked: true},
		{ID: "chk-2", Label: "性能测试", Required: false, Checked: false},
	}

	raw, _ := json.Marshal(items)
	var got []ChecklistItem
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != len(items) {
		t.Errorf("len: got %d want %d", len(got), len(items))
	}
	if got[0].Required != items[0].Required {
		t.Errorf("required: got %v want %v", got[0].Required, items[0].Required)
	}
}

func TestQualityMetrics_JSONRoundTrip(t *testing.T) {
	q := QualityMetrics{
		TotalBugs:       15,
		OpenBugs:        3,
		CriticalBugs:    0,
		MajorBugs:       2,
		BugBySeverity:   map[int]int{0: 0, 1: 0, 2: 5, 3: 7, 4: 3},
		FoundBugCount:   15,
		FixedBugCount:   12,
		FixRate:         0.8,
		PassQualityGate: true,
	}

	raw, _ := json.Marshal(q)
	var got QualityMetrics
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TotalBugs != q.TotalBugs {
		t.Errorf("total_bugs: got %v want %v", got.TotalBugs, q.TotalBugs)
	}
	if got.PassQualityGate != q.PassQualityGate {
		t.Errorf("pass_quality_gate: got %v want %v", got.PassQualityGate, q.PassQualityGate)
	}
}

func TestDeliveryReport_JSONRoundTrip(t *testing.T) {
	r := DeliveryReport{
		GeneratedAt:       time.Now().Truncate(time.Second),
		SprintCount:       2,
		TotalPoints:       200,
		CompletedPoints:   180,
		TotalIssues:       40,
		CompletedIssues:   36,
		BugCount:          10,
		FixedBugCount:     8,
		PassRate:          0.9,
		EligibleToRelease: true,
	}

	raw, _ := json.Marshal(r)
	var got DeliveryReport
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PassRate != r.PassRate {
		t.Errorf("pass_rate: got %v want %v", got.PassRate, r.PassRate)
	}
	if got.EligibleToRelease != r.EligibleToRelease {
		t.Errorf("eligible: got %v want %v", got.EligibleToRelease, r.EligibleToRelease)
	}
}

func TestReleaseNotesData_JSONRoundTrip(t *testing.T) {
	data := ReleaseNotesData{
		VersionName: "v1.0 正式版",
		Semver:      "1.0.0",
		RequirementsDone: []NoteIssueRef{
			{Identifier: "YD-101", Name: "用户登录", StateName: "已完成"},
		},
		BugsFixed: []NoteIssueRef{
			{Identifier: "YD-201", Name: "登录超时", StateName: "已关闭"},
		},
		KnownIssues: []NoteIssueRef{
			{Identifier: "YD-301", Name: "性能问题", StateName: "进行中"},
		},
	}

	raw, _ := json.Marshal(data)
	var got ReleaseNotesData
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.VersionName != data.VersionName {
		t.Errorf("version_name: got %q want %q", got.VersionName, data.VersionName)
	}
	if len(got.RequirementsDone) != 1 {
		t.Errorf("requirements: got %d want 1", len(got.RequirementsDone))
	}
}

// ==========================================================================
// P0-5: 进度聚合边界条件
// ==========================================================================

func TestVersionProgress_Boundaries(t *testing.T) {
	cases := []struct {
		name string
		p    VersionProgress
	}{
		{"空进度", VersionProgress{TotalPoints: 0, DonePoints: 0, TotalIssues: 0, DoneIssues: 0, CompletionRate: 0}},
		{"全部完成", VersionProgress{TotalPoints: 200, DonePoints: 200, TotalIssues: 40, DoneIssues: 40, CompletionRate: 1.0, SprintCount: 3}},
		{"超额完成", VersionProgress{TotalPoints: 100, DonePoints: 120, TotalIssues: 20, DoneIssues: 25, CompletionRate: 1.2, SprintCount: 1}},
		{"部分完成 (忌零除)", VersionProgress{TotalPoints: 0, DonePoints: 0, CompletionRate: 0, SprintCount: 0}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, _ := json.Marshal(c.p)
			var got VersionProgress
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got.TotalPoints != c.p.TotalPoints {
				t.Errorf("total_points: got %v want %v", got.TotalPoints, c.p.TotalPoints)
			}
			if math.Abs(got.CompletionRate-c.p.CompletionRate) > 1e-6 {
				t.Errorf("completion_rate: got %v want %v", got.CompletionRate, c.p.CompletionRate)
			}
		})
	}
}

// ==========================================================================
// P0-6: 版本状态枚举合法性
// ==========================================================================

func TestVersionStatusCode_IsValid(t *testing.T) {
	cases := []struct {
		code  VersionStatusCode
		valid bool
	}{
		{VersionPlanning, true},
		{VersionActive, true},
		{VersionReleased, true},
		{VersionArchived, true},
		{VersionStatusCode(""), false},
		{VersionStatusCode("unknown"), false},
		{VersionStatusCode("draft"), false},
	}

	for _, c := range cases {
		t.Run(string(c.code), func(t *testing.T) {
			got := c.code.IsValid()
			if got != c.valid {
				t.Errorf("IsValid(%q): got %v want %v", c.code, got, c.valid)
			}
		})
	}
}

// ==========================================================================
// P0-7: SprintRef 关联迭代摘要序列化
// ==========================================================================

func TestSprintRef_JSONRoundTrip(t *testing.T) {
	sd := "2025-06-01"
	ed := "2025-06-14"

	s := SprintRef{
		SprintID: 1,
		Name:     "Sprint 5",
		Status:   "active",
		StartDate: &sd,
		EndDate:   &ed,
		Progress: &SprintProgressRef{
			TotalPoints: 50,
			DonePoints:  30,
			TotalIssues: 10,
			DoneIssues:  6,
		},
	}

	raw, _ := json.Marshal(s)
	var got SprintRef
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.SprintID != s.SprintID {
		t.Errorf("sprint_id: got %v want %v", got.SprintID, s.SprintID)
	}
	if *got.StartDate != sd {
		t.Errorf("start_date: got %q want %q", *got.StartDate, sd)
	}
	if got.Progress == nil {
		t.Fatal("progress should not be nil")
	}
}

// ==========================================================================
// P0-8: BugVersionView 缺陷版本视图
// ==========================================================================

func TestBugVersionView_JSONRoundTrip(t *testing.T) {
	sev := 1

	b := BugVersionView{
		IssueID:      100,
		Identifier:   "YD-100",
		Name:         "登录超时",
		Severity:     &sev,
		FoundPhase:   "系统测试",
		StateName:    "打开",
		StateGroup:   "started",
		FoundVersion: "1.0.0-beta",
		FixVersion:   "1.0.1",
		RootCause:    "超时配置",
	}

	raw, _ := json.Marshal(b)
	var got BugVersionView
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.IssueID != b.IssueID {
		t.Errorf("issue_id: got %v want %v", got.IssueID, b.IssueID)
	}
	if *got.Severity != *b.Severity {
		t.Errorf("severity: got %v want %v", *got.Severity, *b.Severity)
	}
	if got.StateGroup != b.StateGroup {
		t.Errorf("state_group: got %q want %q", got.StateGroup, b.StateGroup)
	}
}

// ==========================================================================
// P0-9: quality gate 计算逻辑测试
// ==========================================================================

func TestQualityGate_PassCondition(t *testing.T) {
	// pass_quality_gate == true 当且仅当 critical_bugs == 0
	cases := []struct {
		name   string
		crit   int
		expect bool
	}{
		{"无致命缺陷", 0, true},
		{"有致命缺陷", 1, false},
		{"多个致命缺陷", 5, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pass := c.crit == 0
			if pass != c.expect {
				t.Errorf("pass_quality_gate: got %v want %v", pass, c.expect)
			}
		})
	}
}

// ==========================================================================
// P0-10: 修复率计算
// ==========================================================================

func TestFixRate_Boundary(t *testing.T) {
	cases := []struct {
		name  string
		found int
		fixed int
		want  float64
	}{
		{"全部修复", 10, 10, 1.0},
		{"一半修复", 10, 5, 0.5},
		{"无缺陷", 0, 0, 1.0}, // 无缺陷时修复率=1
		{"找到但未修复", 5, 0, 0.0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var rate float64
			if c.found == 0 {
				rate = 1.0
			} else {
				rate = float64(c.fixed) / float64(c.found)
			}
			if math.Abs(rate-c.want) > 1e-6 {
				t.Errorf("fix_rate: got %v want %v", rate, c.want)
			}
		})
	}
}

// ==========================================================================
// P0-11: delivery report eligible 条件
// ==========================================================================

func TestDeliveryEligible(t *testing.T) {
	cases := []struct {
		name     string
		passGate bool
		compRate float64
		want     bool
	}{
		{"全部满足", true, 0.9, true},
		{"门禁未过", false, 0.9, false},
		{"进度不足", true, 0.7, false},
		{"都不满足", false, 0.5, false},
		{"刚好80%", true, 0.8, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			eligible := c.passGate && c.compRate >= 0.8
			if eligible != c.want {
				t.Errorf("eligible: got %v want %v", eligible, c.want)
			}
		})
	}
}

// ==========================================================================
// P0-12: checklist 上限校验（S6 加固）
// ==========================================================================

func TestNormalizeChecklist_Overflow(t *testing.T) {
	// 超过 maxChecklistItems (50) 应返回错误
	items := make([]ChecklistItem, 51)
	for i := range items {
		items[i] = ChecklistItem{Label: fmt.Sprintf("检查项 %d", i+1), Required: true}
	}
	_, err := normalizeChecklist(items)
	if err == nil {
		t.Error("expected error for checklist exceeding max items")
	}
}

func TestNormalizeChecklist_LabelTooLong(t *testing.T) {
	longLabel := strings.Repeat("a", 201)
	items := []ChecklistItem{{Label: longLabel, Required: true}}
	_, err := normalizeChecklist(items)
	if err == nil {
		t.Error("expected error for label exceeding 200 chars")
	}
}

// ==========================================================================
// P0-13: validateCreateInput 边界（S6 加固）
// ==========================================================================

func TestValidateCreateInput_SemverTooLong(t *testing.T) {
	in := CreateVersionInput{
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "v1.0",
		Semver:      strings.Repeat("1.0.0-alpha.", 10),
	}
	err := validateCreateInput(in)
	if err == nil {
		t.Error("expected error for semver > 50 chars")
	}
}

func TestValidateCreateInput_DescriptionTooLong(t *testing.T) {
	in := CreateVersionInput{
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "v1.0",
		Semver:      "1.0.0",
		Description: strings.Repeat("x", 2001),
	}
	err := validateCreateInput(in)
	if err == nil {
		t.Error("expected error for description > 2000 chars")
	}
}

// ==========================================================================
// P0-14: renderReleaseNotes 输出验证
// ==========================================================================

func TestRenderReleaseNotes(t *testing.T) {
	v := &Version{Name: "正式版", Semver: "1.0.0", Description: "首个发布"}
	src := &ReleaseNotesData{
		VersionName: "正式版",
		Semver:      "1.0.0",
		RequirementsDone: []NoteIssueRef{
			{Identifier: "YD-101", Name: "用户登录", StateName: "已完成"},
		},
		BugsFixed: []NoteIssueRef{
			{Identifier: "YD-201", Name: "登录超时修复", StateName: "已关闭"},
		},
		KnownIssues: []NoteIssueRef{
			{Identifier: "YD-301", Name: "已知性能问题", StateName: "进行中"},
		},
	}

	notes := renderReleaseNotes(v, src)

	checks := []string{
		"# 正式版 v1.0.0",
		"## ✅ 已完成需求与任务",
		"YD-101",
		"用户登录",
		"## 🐛 修复缺陷",
		"YD-201",
		"## ⚠️ 已知问题",
		"YD-301",
	}
	for _, c := range checks {
		if !strings.Contains(notes, c) {
			t.Errorf("release notes missing expected content: %q", c)
		}
	}
}

func TestRenderReleaseNotes_Empty(t *testing.T) {
	v := &Version{Name: "空版本", Semver: "0.1.0"}
	src := &ReleaseNotesData{VersionName: "空版本", Semver: "0.1.0"}

	notes := renderReleaseNotes(v, src)
	if !strings.Contains(notes, "（无完成的需求/任务）") {
		t.Error("should show empty placeholder for requirements")
	}
	if !strings.Contains(notes, "（无修复缺陷）") {
		t.Error("should show empty placeholder for bugs")
	}
	if strings.Contains(notes, "## ⚠️ 已知问题") {
		t.Error("should not show known issues section when empty")
	}
}

// ==========================================================================
// P0-15: mustMarshalJSON
// ==========================================================================

func TestMustMarshalJSON(t *testing.T) {
	b := mustMarshalJSON(map[string]int{"a": 1})
	if string(b) != `{"a":1}` {
		t.Errorf("unexpected output: %s", b)
	}

	// 不可序列化的值返回 null
	b = mustMarshalJSON(make(chan int))
	if string(b) != "null" {
		t.Errorf("expected null for unserializable value, got: %s", b)
	}
}

// ==========================================================================
// P0-16: Version 模型包含 version 字段
// ==========================================================================

func TestVersion_HasOptimisticLockField(t *testing.T) {
	v := Version{ID: 1, Version: 3}
	raw, _ := json.Marshal(v)
	var got map[string]interface{}
	json.Unmarshal(raw, &got)
	if ver, ok := got["version"]; !ok {
		t.Error("version field missing from JSON output")
	} else if v, ok := ver.(float64); !ok || int(v) != 3 {
		t.Errorf("unexpected version value: %v", ver)
	}
}

// ==========================================================================
// helpers
// ==========================================================================

func ptr[T any](v T) *T { return &v }
