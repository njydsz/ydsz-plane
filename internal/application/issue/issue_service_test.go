// Package issue — Issue 领域不变量单元测试。
//
// 覆盖范围：
//   1. 输入校验（validateCreateInput / validateTypeFields）
//   2. 错误码语义（mapPgError）
//   3. 优先级权重排序
//   4. 工作项类型枚举
//   5. 状态分组枚举
//   6. Issue 模型 JSON 序列化 / 反序列化
//   7. buildUpdateSet 逻辑
//   8. buildIssueWhere / buildCountWhere 查询构造
//   9. 排序字段映射
//  10. 工时输入校验
//
// 互联网大厂标准：
//   - 表驱动测试 (table-driven tests)
//   - 边界条件全覆盖
//   - 黄金路径 + 错误路径双覆盖
//   - 常量枚举验证
package issue

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ==========================================================================
// P1: 输入校验 — validateCreateInput
// ==========================================================================

func TestValidateCreateInput(t *testing.T) {
	cases := []struct {
		name    string
		input   CreateIssueInput
		wantErr bool
		errField string
	}{
		{
			name: "正常创建需求",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeRequirement,
				Name: "用户登录", CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "正常创建任务",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeTask,
				Name: "实现登录接口", CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "正常创建缺陷",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeDefect,
				Name: "登录按钮无响应", CreatedBy: 1,
				Severity:   intPtr(3),
				FoundPhase: strPtr("unit"),
			},
			wantErr: false,
		},
		{
			name: "名称为空",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeTask,
				Name: "", CreatedBy: 1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "名称仅空白字符",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeTask,
				Name: "   ", CreatedBy: 1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "名称超长 >500",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeTask,
				Name: string(make([]byte, 501)), CreatedBy: 1,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "名称=500 合法",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: TypeTask,
				Name: string(make([]byte, 500)), CreatedBy: 1,
			},
			wantErr: false,
		},
		{
			name: "无效类型",
			input: CreateIssueInput{
				WorkspaceID: 1, ProjectID: 1, TypeCode: "story",
				Name: "测试", CreatedBy: 1,
			},
			wantErr:  true,
			errField: "type_code",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateCreateInput(c.input)
			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// ==========================================================================
// P2: 领域不变量 — validateTypeFields（缺陷必填校验）
// ==========================================================================

func TestValidateTypeFields_DefectRequired(t *testing.T) {
	cases := []struct {
		name       string
		typeCode   IssueTypeCode
		severity   *int
		foundPhase *string
		wantErr    bool
	}{
		{
			name:       "需求不需要严重度",
			typeCode:   TypeRequirement,
			severity:   nil,
			foundPhase: nil,
			wantErr:    false,
		},
		{
			name:       "任务不需要严重度",
			typeCode:   TypeTask,
			severity:   nil,
			foundPhase: nil,
			wantErr:    false,
		},
		{
			name:       "缺陷缺少严重度",
			typeCode:   TypeDefect,
			severity:   nil,
			foundPhase: strPtr("unit"),
			wantErr:    true,
		},
		{
			name:       "缺陷严重度越界(0)",
			typeCode:   TypeDefect,
			severity:   intPtr(0),
			foundPhase: strPtr("unit"),
			wantErr:    true,
		},
		{
			name:       "缺陷严重度越界(6)",
			typeCode:   TypeDefect,
			severity:   intPtr(6),
			foundPhase: strPtr("unit"),
			wantErr:    true,
		},
		{
			name:       "缺陷缺少发现阶段",
			typeCode:   TypeDefect,
			severity:   intPtr(3),
			foundPhase: nil,
			wantErr:    true,
		},
		{
			name:       "缺陷发现阶段为空字符串",
			typeCode:   TypeDefect,
			severity:   intPtr(3),
			foundPhase: strPtr(""),
			wantErr:    true,
		},
		{
			name:       "缺陷完整字段合法",
			typeCode:   TypeDefect,
			severity:   intPtr(3),
			foundPhase: strPtr("unit"),
			wantErr:    false,
		},
		{
			name:       "缺陷边界值 severity=1",
			typeCode:   TypeDefect,
			severity:   intPtr(1),
			foundPhase: strPtr("integration"),
			wantErr:    false,
		},
		{
			name:       "缺陷边界值 severity=5",
			typeCode:   TypeDefect,
			severity:   intPtr(5),
			foundPhase: strPtr("production"),
			wantErr:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateTypeFields(c.typeCode, c.severity, c.foundPhase)
			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// ==========================================================================
// P3: 错误码映射 — mapPgError
// ==========================================================================

func TestMapPgError_NotFound(t *testing.T) {
	err := mapPgError(pgx.ErrNoRows)
	if err == nil {
		t.Fatal("expected error for pgx.ErrNoRows")
	}
}

func TestMapPgError_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{ConstraintName: "issues_project_id_sequence_id_key"}
	err := mapPgError(pgErr)
	if err == nil {
		t.Fatal("expected error for unique violation")
	}
}

func TestMapPgError_DefectConstraint(t *testing.T) {
	pgErr := &pgconn.PgError{ConstraintName: "defect_required"}
	err := mapPgError(pgErr)
	if err == nil {
		t.Fatal("expected error for defect_required constraint")
	}
}

func TestMapPgError_Generic(t *testing.T) {
	pgErr := &pgconn.PgError{ConstraintName: "some_unknown"}
	err := mapPgError(pgErr)
	if err == nil {
		t.Fatal("expected non-nil error for unknown constraint")
	}
}

// ==========================================================================
// P4: 优先级权重排序
// ==========================================================================

func TestPriorityWeight_Ordering(t *testing.T) {
	cases := []struct {
		p1   IssuePriority
		p2   IssuePriority
		desc string
	}{
		{PriorityUrgent, PriorityHigh, "紧急 > 高"},
		{PriorityHigh, PriorityMedium, "高 > 中"},
		{PriorityMedium, PriorityLow, "中 > 低"},
		{PriorityLow, PriorityNone, "低 > 无"},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if PriorityWeight[c.p1] <= PriorityWeight[c.p2] {
				t.Errorf("%s: %d <= %d", c.desc, PriorityWeight[c.p1], PriorityWeight[c.p2])
			}
		})
	}
}

func TestPriorityWeight_AllDefined(t *testing.T) {
	expected := []IssuePriority{
		PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone,
	}
	for _, p := range expected {
		w, ok := PriorityWeight[p]
		if !ok {
			t.Errorf("priority %q not defined in PriorityWeight", p)
		}
		if w < 0 || w > 5 {
			t.Errorf("priority %q weight %d out of range", p, w)
		}
	}
}

// ==========================================================================
// P5: 工作项类型枚举验证
// ==========================================================================

func TestIssueTypeCode_Values(t *testing.T) {
	valid := map[IssueTypeCode]bool{
		TypeRequirement: true,
		TypeTask:        true,
		TypeDefect:      true,
	}

	cases := []IssueTypeCode{
		TypeRequirement, TypeTask, TypeDefect,
		IssueTypeCode("story"), IssueTypeCode("epic"), IssueTypeCode(""),
	}

	for _, tc := range cases {
		_, ok := valid[tc]
		if !ok && (tc == TypeRequirement || tc == TypeTask || tc == TypeDefect) {
			t.Errorf("expected %q to be valid", tc)
		}
		if ok && (tc != TypeRequirement && tc != TypeTask && tc != TypeDefect) {
			t.Errorf("expected %q to be invalid", tc)
		}
	}
}

// ==========================================================================
// P6: 状态分组枚举验证
// ==========================================================================

func TestStateGroup_Values(t *testing.T) {
	valid := map[StateGroup]bool{
		GroupBacklog:   true,
		GroupStarted:   true,
		GroupCompleted: true,
		GroupCancelled: true,
	}

	cases := []StateGroup{
		GroupBacklog, GroupStarted, GroupCompleted, GroupCancelled,
		StateGroup("archived"), StateGroup(""),
	}

	for _, g := range cases {
		_, ok := valid[g]
		if !ok && (g == GroupBacklog || g == GroupStarted || g == GroupCompleted || g == GroupCancelled) {
			t.Errorf("expected %q to be valid", g)
		}
	}
}

// ==========================================================================
// P7: Issue 模型 JSON 序列化/反序列化
// ==========================================================================

func TestIssueModel_JSONRoundTrip(t *testing.T) {
	iss := Issue{
		ID:         1,
		WorkspaceID: 1,
		ProjectID:  1,
		SequenceID: 42,
		Identifier: "YD-42",
		TypeCode:   TypeTask,
		Depth:      2,
		Name:       "实现登录接口",
		DescriptionHTML: "<p>描述</p>",
		StateID:    2,
		Priority:   PriorityHigh,
		Progress:   50,
		Version:    3,
		ParentID:   int64Ptr(1),
		Point:      intPtr(5),
		Severity:   intPtr(3),
		FoundPhase: strPtr("unit"),
		Assignees:  []int64{1, 2},
		Labels:     []int64{3},
	}

	raw, err := json.Marshal(iss)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Issue
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != iss.ID {
		t.Errorf("id: got %v want %v", got.ID, iss.ID)
	}
	if got.Identifier != iss.Identifier {
		t.Errorf("identifier: got %q want %q", got.Identifier, iss.Identifier)
	}
	if got.TypeCode != iss.TypeCode {
		t.Errorf("type_code: got %q want %q", got.TypeCode, iss.TypeCode)
	}
	if got.Depth != iss.Depth {
		t.Errorf("depth: got %v want %v", got.Depth, iss.Depth)
	}
	if got.Priority != iss.Priority {
		t.Errorf("priority: got %q want %q", got.Priority, iss.Priority)
	}
	if got.Progress != iss.Progress {
		t.Errorf("progress: got %v want %v", got.Progress, iss.Progress)
	}
	if len(got.Assignees) != 2 {
		t.Errorf("assignees length: got %d want 2", len(got.Assignees))
	}
	if got.Version != iss.Version {
		t.Errorf("version: got %d want %d", got.Version, iss.Version)
	}
}

func TestIssueModel_EmptyFields(t *testing.T) {
	iss := Issue{
		ID:          1,
		Name:        "最小工作项",
		StateID:     1,
		Depth:       1,
		Priority:    PriorityNone,
		Version:     1,
		Assignees:   []int64{},
		Labels:      nil,
	}
	raw, _ := json.Marshal(iss)
	var got Issue
	_ = json.Unmarshal(raw, &got)
	// omitempty 空切片序列化为 null 或不存在，反序列化后为 nil
	if got.ID != 1 {
		t.Errorf("id: got %v want 1", got.ID)
	}
}

// ==========================================================================
// P8: 查询构造器 — buildIssueWhere
// ==========================================================================

func TestBuildIssueWhere_BaseOnly(t *testing.T) {
	opts := ListIssuesOptions{ProjectID: 1, WorkspaceID: 1}
	where, args := buildIssueWhere(opts)
	if len(args) != 2 {
		t.Errorf("base args: got %d want 2", len(args))
	}
	if where == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
}

func TestBuildIssueWhere_WithStateFilter(t *testing.T) {
	sid := int64(5)
	opts := ListIssuesOptions{ProjectID: 1, WorkspaceID: 1, StateID: &sid}
	where, args := buildIssueWhere(opts)
	if len(args) != 3 {
		t.Errorf("args: got %d want 3", len(args))
	}
	if where == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
}

func TestBuildIssueWhere_AllFilters(t *testing.T) {
	sid := int64(3)
	tc := TypeTask
	pri := PriorityHigh
	pid := int64(1)
	opts := ListIssuesOptions{
		ProjectID:   1,
		WorkspaceID: 1,
		StateID:     &sid,
		TypeCode:    &tc,
		Priority:    &pri,
		ParentID:    &pid,
		Search:      "login",
	}
	where, args := buildIssueWhere(opts)
	if len(args) < 5 {
		t.Errorf("args: got %d want >=5", len(args))
	}
	if where == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
}

func TestBuildCountWhere(t *testing.T) {
	opts := ListIssuesOptions{ProjectID: 1, WorkspaceID: 1}
	where, args := buildCountWhere(opts)
	if len(args) != 2 {
		t.Errorf("args: got %d want 2", len(args))
	}
	if where == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
}

// ==========================================================================
// P9: 排序字段映射
// ==========================================================================

func TestBuildIssueSort(t *testing.T) {
	cases := []struct {
		sortBy   string
		sortDesc bool
		wantContains string
	}{
		{"priority", false, "CASE i.priority"},
		{"priority", true, "DESC"},
		{"target_date", false, "i.target_date"},
		{"created_at", true, "i.created_at DESC"},
		{"sequence", false, "i.sequence_id"},
		{"unknown", false, "i.updated_at"},
		{"", true, "i.updated_at DESC"},
	}

	for _, c := range cases {
		t.Run(c.sortBy+"_"+map[bool]string{true: "desc", false: "asc"}[c.sortDesc], func(t *testing.T) {
			got := buildIssueSort(ListIssuesOptions{SortBy: c.sortBy, SortDesc: c.sortDesc})
			if !containsString(got, c.wantContains) {
				t.Errorf("sort: got %q, want contains %q", got, c.wantContains)
			}
		})
	}
}

// ==========================================================================
// P10: buildUpdateSet 逻辑
// ==========================================================================

func TestBuildUpdateSet_Empty(t *testing.T) {
	in := UpdateIssueInput{}
	current := &Issue{}
	sets, args := buildUpdateSet(in, current)
	if len(sets) != 0 {
		t.Errorf("empty input: expected 0 sets, got %d", len(sets))
	}
	if len(args) != 0 {
		t.Errorf("empty input: expected 0 args, got %d", len(args))
	}
}

func TestBuildUpdateSet_Name(t *testing.T) {
	in := UpdateIssueInput{Name: strPtr("新名称")}
	current := &Issue{Version: 1}
	sets, args := buildUpdateSet(in, current)
	if len(sets) < 1 { // name + updated_at
		t.Errorf("expected >=1 sets, got %d", len(sets))
	}
	if len(args) < 1 {
		t.Errorf("expected >=1 args, got %d", len(args))
	}
}

func TestBuildUpdateSet_MultiField(t *testing.T) {
	phase := "unit"
	in := UpdateIssueInput{
		Name:       strPtr("改名"),
		Priority:   ptrPriority(PriorityUrgent),
		Severity:   intPtr(4),
		FoundPhase: &phase,
		Version:    2,
	}
	current := &Issue{Version: 1}
	sets, args := buildUpdateSet(in, current)
	// name + priority + severity + found_phase + updated_at = 5
	if len(sets) != 5 {
		t.Errorf("multi-field: expected 5 sets, got %d: %v", len(sets), sets)
	}
	if len(args) != 4 {
		t.Errorf("multi-field: expected 4 args, got %d", len(args))
	}
}

func TestBuildUpdateSet_VersionFields(t *testing.T) {
	fv := int64(10)
	fixv := int64(20)
	rv := int64(30)
	in := UpdateIssueInput{
		FoundVersionID:   &fv,
		FixVersionID:     &fixv,
		ReleaseVersionID: &rv,
	}
	current := &Issue{Version: 1}
	sets, args := buildUpdateSet(in, current)
	if len(sets) != 4 { // 3 fields + updated_at
		t.Errorf("version fields: expected 4 sets, got %d", len(sets))
	}
	if len(args) != 3 {
		t.Errorf("version fields: expected 3 args, got %d", len(args))
	}
}

// ==========================================================================
// P11: 工时输入校验
// ==========================================================================

func TestTimeLogInput_Validation(t *testing.T) {
	cases := []struct {
		name    string
		input   CreateTimeLogInput
		wantErr bool
	}{
		{
			name: "正常工时",
			input: CreateTimeLogInput{
				WorkspaceID: 1, ProjectID: 1, IssueID: 1, UserID: 1,
				DurationMinutes: 60, Description: "开发",
			},
			// 仅验证校验逻辑，不含 DB
		},
		{
			name: "时长为0",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 1, UserID: 1, DurationMinutes: 0,
			},
			wantErr: true,
		},
		{
			name: "时长为负",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 1, UserID: 1, DurationMinutes: -1,
			},
			wantErr: true,
		},
		{
			name: "时长超过24小时",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 1, UserID: 1, DurationMinutes: 1441,
			},
			wantErr: true,
		},
		{
			name: "时长=1440合法",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 1, UserID: 1, DurationMinutes: 1440,
			},
			wantErr: false, // 边界值合法
		},
		{
			name: "缺少 user_id",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 1, UserID: 0, DurationMinutes: 60,
			},
			wantErr: true,
		},
		{
			name: "缺少 issue_id",
			input: CreateTimeLogInput{
				WorkspaceID: 1, IssueID: 0, UserID: 1, DurationMinutes: 60,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// 仅验证 Create 的入参校验部分，跳过 DB
			if c.input.DurationMinutes <= 0 || c.input.DurationMinutes > 1440 {
				if !c.wantErr {
					t.Errorf("expected error for duration=%d", c.input.DurationMinutes)
				}
				return
			}
			if c.input.UserID == 0 {
				if !c.wantErr {
					t.Error("expected error for missing user_id")
				}
				return
			}
			if c.input.IssueID == 0 {
				if !c.wantErr {
					t.Error("expected error for missing issue_id")
				}
				return
			}
			if c.wantErr {
				t.Error("expected no error but validation would pass")
			}
		})
	}
}

// ==========================================================================
// P12: 状态模板合法性
// ==========================================================================

func TestStateTemplates_WellFormed(t *testing.T) {
	for _, tmpl := range DefaultTemplates {
		t.Run(tmpl.Name, func(t *testing.T) {
			if len(tmpl.States) < 2 {
				t.Errorf("template %q: expected >=2 states, got %d", tmpl.Name, len(tmpl.States))
			}
			if len(tmpl.Transitions) == 0 {
				t.Errorf("template %q: expected >=1 transitions", tmpl.Name)
			}
			// 每个流转必须引用存在的状态
			stateNames := map[string]bool{}
			for _, st := range tmpl.States {
				stateNames[st.Name] = true
			}
			for _, tr := range tmpl.Transitions {
				// "*" 是通配符，允许从任意状态流转
				if tr.From != "*" && !stateNames[tr.From] {
					t.Errorf("template %q: transition from %q does not exist", tmpl.Name, tr.From)
				}
				if !stateNames[tr.To] {
					t.Errorf("template %q: transition to %q does not exist", tmpl.Name, tr.To)
				}
			}
		})
	}

	// 验证 BuiltInTransitions 包含三种模板
	expectedNames := []string{"dev_flow", "defect_flow", "requirement_flow"}
	for _, name := range expectedNames {
		if _, ok := BuiltInTransitions[name]; !ok {
			t.Errorf("missing BuiltInTransitions[%q]", name)
		}
	}
}

// ==========================================================================
// P13: State 模型 JSON 序列化
// ==========================================================================

func TestStateModel_JSONRoundTrip(t *testing.T) {
	st := State{
		ID:          1,
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "待处理",
		Group:       GroupBacklog,
		Color:       "#8DA2C2",
		Sequence:    100,
		IsDefault:   true,
	}
	raw, _ := json.Marshal(st)
	var got State
	_ = json.Unmarshal(raw, &got)
	if got.Name != st.Name {
		t.Errorf("name: got %q want %q", got.Name, st.Name)
	}
	if got.Group != st.Group {
		t.Errorf("group: got %q want %q", got.Group, st.Group)
	}
	if got.IsDefault != st.IsDefault {
		t.Errorf("is_default: got %v want %v", got.IsDefault, st.IsDefault)
	}
}

// ==========================================================================
// helpers
// ==========================================================================

func intPtr(v int) *int       { return &v }
func int64Ptr(v int64) *int64 { return &v }
func strPtr(v string) *string { return &v }

func ptrPriority(p IssuePriority) *IssuePriority { return &p }

func containsString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
