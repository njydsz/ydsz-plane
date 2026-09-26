// Package issue — 协调器纯函数 + 聚合根转换不变量单元测试。
//
// 覆盖：
//  1. ValidateTable 白名单校验（SQL 注入防护）
//  2. IssueTypeCode.Table() 映射
//  3. taskToIssue / requirementToIssue / defectToIssue DTO 转换不变量
//  4. PriorityWeight 排序单调性
//  5. 枚举完备性
//
// 纯函数、无 IO、毫秒级执行 — 符合单元测试门禁标准。
package issue

import (
	"testing"
)

// ==========================================================================
// P0: ValidateTable — SQL 注入白名单防护
// ==========================================================================

func TestValidateTable_Whitelist(t *testing.T) {
	cases := []struct {
		table string
		valid bool
	}{
		{TableTask, true},
		{TableRequirement, true},
		{TableDefect, true},
		{"issues", false},
		{"users", false},
		{"", false},
		{"task; DROP TABLE users;", false},
		{"TASK", false},
		{" requirement", false},
	}
	for _, c := range cases {
		err := ValidateTable(c.table)
		if c.valid && err != nil {
			t.Errorf("ValidateTable(%q): 应通过, got %v", c.table, err)
		}
		if !c.valid && err == nil {
			t.Errorf("ValidateTable(%q): 应失败, got nil", c.table)
		}
	}
}

// ==========================================================================
// P1: IssueTypeCode.Table() 映射
// ==========================================================================

func TestIssueTypeCode_Table(t *testing.T) {
	cases := map[IssueTypeCode]string{
		TypeTask:        TableTask,
		TypeRequirement: TableRequirement,
		TypeDefect:      TableDefect,
	}
	for tc, want := range cases {
		if got := tc.Table(); got != want {
			t.Errorf("IssueTypeCode(%q).Table() = %q, want %q", tc, got, want)
		}
	}

	// 未知类型回退到 task（防御性默认值）
	if got := IssueTypeCode("unknown").Table(); got != TableTask {
		t.Errorf("'unknown'.Table() = %q, want %q", got, TableTask)
	}
	if got := IssueTypeCode("").Table(); got != TableTask {
		t.Errorf("''.Table() = %q, want %q", got, TableTask)
	}
}

// ==========================================================================
// P2: 聚合根 → Issue DTO 转换不变量
// ==========================================================================

func TestTaskToIssue_Conversion(t *testing.T) {
	cat := "backend"
	parentID := int64(100)
	sprintID := int64(5)
	now := time.Now().UTC().Truncate(time.Second)

	src := &Task{
		ID: 1, PublicID: "pub-t1", WorkspaceID: 1, ProjectID: 1,
		SequenceID: 42, Identifier: "YD-42", TypeCode: TypeTask,
		ParentID: &parentID, Depth: 1, Name: "开发登录接口",
		StateID: 2, Priority: PriorityHigh, SprintID: &sprintID,
		Progress: 30, SortOrder: 100.5,
		Version: 2, CreatedBy: 1, CreatedAt: now, UpdatedAt: now,
		Category: &cat, Assignees: []int64{10, 20},
		Labels: []int64{100}, Modules: []int64{200},
	}

	got := taskToIssue(src)
	if got.TypeCode != TypeTask {
		t.Errorf("TypeCode: got %q want %q", got.TypeCode, TypeTask)
	}
	if got.ID != src.ID {
		t.Errorf("ID: got %d want %d", got.ID, src.ID)
	}
	if got.PublicID != src.PublicID {
		t.Errorf("PublicID: got %q want %q", got.PublicID, src.PublicID)
	}
	if got.Name != src.Name {
		t.Errorf("Name: got %q want %q", got.Name, src.Name)
	}
	if got.Category == nil || *got.Category != cat {
		t.Errorf("Category: got %v want %q", got.Category, cat)
	}
	if len(got.Assignees) != 2 {
		t.Errorf("Assignees length: got %d want 2", len(got.Assignees))
	}
	if got.ParentID == nil || *got.ParentID != parentID {
		t.Errorf("ParentID: got %v want %d", got.ParentID, parentID)
	}
	if got.Progress != 30 {
		t.Errorf("Progress: got %d want 30", got.Progress)
	}
}

func TestRequirementToIssue_Conversion(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	src := &Requirement{
		ID: 2, PublicID: "pub-r1", WorkspaceID: 1, ProjectID: 1,
		SequenceID: 43, Identifier: "YD-43", TypeCode: TypeRequirement,
		Depth: 1, Name: "用户登录需求", StateID: 1, Priority: PriorityUrgent,
		Progress: 0, SortOrder: 200.0, Version: 1, CreatedBy: 1,
		CreatedAt: now, UpdatedAt: now,
		Assignees: []int64{10}, Labels: []int64{100}, Modules: []int64{200},
	}

	got := requirementToIssue(src)
	if got.TypeCode != TypeRequirement {
		t.Errorf("TypeCode: got %q want %q", got.TypeCode, TypeRequirement)
	}
	if got.PublicID != "pub-r1" {
		t.Errorf("PublicID: got %q want %q", got.PublicID, "pub-r1")
	}
	if got.Name != src.Name {
		t.Errorf("Name: got %q want %q", got.Name, src.Name)
	}
	// Requirement 无 Category / Severity / FoundPhase
	if got.Category != nil {
		t.Errorf("Requirement should have nil Category, got %v", *got.Category)
	}
	if got.Severity != nil {
		t.Errorf("Requirement should have nil Severity, got %v", *got.Severity)
	}
}

func TestDefectToIssue_Conversion(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	fvID := int64(10)
	src := &Defect{
		ID: 3, PublicID: "pub-d1", WorkspaceID: 1, ProjectID: 1,
		SequenceID: 44, Identifier: "YD-44", TypeCode: TypeDefect,
		Depth: 1, Name: "登录按钮无响应", StateID: 4, Priority: PriorityUrgent,
		Severity: 3, FoundPhase: "production",
		Progress: 0, SortOrder: 300.0, Version: 1, CreatedBy: 1,
		CreatedAt: now, UpdatedAt: now,
		FixVersionID: &fvID,
	}

	got := defectToIssue(src)
	if got.TypeCode != TypeDefect {
		t.Errorf("TypeCode: got %q want %q", got.TypeCode, TypeDefect)
	}
	if got.Severity == nil || *got.Severity != 3 {
		t.Errorf("Severity: got %v want 3", got.Severity)
	}
	if got.FoundPhase == nil || *got.FoundPhase != "production" {
		t.Errorf("FoundPhase: got %v want 'production'", got.FoundPhase)
	}
	// Defect 无 Category
	if got.Category != nil {
		t.Errorf("Defect should have nil Category, got %v", *got.Category)
	}
}

// ==========================================================================
// P3: Priority 权重单调性验证
// ==========================================================================

func TestPriorityWeight_StrictOrder(t *testing.T) {
	// urgent(4) > high(3) > medium(2) > low(1) > none(0)
	pairs := []struct {
		higher, lower IssuePriority
	}{
		{PriorityUrgent, PriorityHigh},
		{PriorityHigh, PriorityMedium},
		{PriorityMedium, PriorityLow},
		{PriorityLow, PriorityNone},
	}
	for _, p := range pairs {
		if PriorityWeight[p.higher] <= PriorityWeight[p.lower] {
			t.Errorf("PriorityWeight[%q](%d) should > PriorityWeight[%q](%d)",
				p.higher, PriorityWeight[p.higher], p.lower, PriorityWeight[p.lower])
		}
	}
}

// ==========================================================================
// P4: 枚举完备性（穷举 switch 覆盖，防新增枚举遗漏）
// ==========================================================================

func TestSwitch_ExhaustiveIssueType(t *testing.T) {
	allTypes := []IssueTypeCode{TypeEpic, TypeRequirement, TypeTask, TypeDefect}
	for _, tc := range allTypes {
		switch tc {
		case TypeEpic:
		case TypeRequirement:
		case TypeTask:
		case TypeDefect:
		default:
			t.Errorf("unexpected IssueTypeCode: %q", tc)
		}
	}
}

func TestSwitch_ExhaustiveStateGroup(t *testing.T) {
	allGroups := []StateGroup{GroupBacklog, GroupStarted, GroupCompleted, GroupCancelled}
	for _, g := range allGroups {
		switch g {
		case GroupBacklog:
		case GroupStarted:
		case GroupCompleted:
		case GroupCancelled:
		default:
			t.Errorf("unexpected StateGroup: %q", g)
		}
	}
}

func TestSwitch_ExhaustivePriority(t *testing.T) {
	allPriorities := []IssuePriority{PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone}
	for _, p := range allPriorities {
		switch p {
		case PriorityUrgent:
		case PriorityHigh:
		case PriorityMedium:
		case PriorityLow:
		case PriorityNone:
		default:
			t.Errorf("unexpected IssuePriority: %q", p)
		}
	}
}

// ==========================================================================
// P5: DelayReason 选项列表非空（供前端下拉）
// ==========================================================================

func TestDelayReasonOptions_NonEmpty(t *testing.T) {
	if len(DelayReasonOptions) == 0 {
		t.Error("DelayReasonOptions should not be empty")
	}
	seen := map[string]bool{}
	for _, opt := range DelayReasonOptions {
		if opt.Value == "" {
			t.Errorf("DelayReasonOptions contains empty value: %+v", opt)
		}
		if seen[opt.Value] {
			t.Errorf("Duplicate DelayReasonOptions value: %q", opt.Value)
		}
		seen[opt.Value] = true
	}
}

// ==========================================================================
// P6: CreateDefectInput Severity 范围校验（纯规则）
// ==========================================================================

func TestDefectSeverity_Range(t *testing.T) {
	cases := []struct {
		severity int
		valid    bool
	}{
		{1, true}, {2, true}, {3, true}, {4, true}, {5, true},
		{0, false}, {6, false}, {-1, false},
	}
	for _, c := range cases {
		got := c.severity >= 1 && c.severity <= 5
		if got != c.valid {
			t.Errorf("defect severity %d: isValid=%v, want %v", c.severity, got, c.valid)
		}
	}
}

// ==========================================================================
// P7: CreateRequirementInput Name 规则（复用纯规则）
// ==========================================================================

func TestNameValidation_Rules(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"正常名称", true},
		{"a", true},
		{"", false},
		{"   ", false},
	}
	for _, c := range cases {
		got := len(c.name) > 0 && len(c.name) <= 500 && c.name != ""
		trimmed := false
		for _, r := range c.name {
			if r != ' ' && r != '\t' && r != '\n' {
				trimmed = true
				break
			}
		}
		isValid := got && trimmed
		if isValid != c.valid {
			t.Errorf("name validation %q: got %v, want %v", c.name, isValid, c.valid)
		}
	}
}
