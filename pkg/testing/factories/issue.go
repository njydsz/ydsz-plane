// Package factories — 测试数据工厂（手写 Builder Pattern）。
//
// 设计目标：
//   - 让集成测试和本地 fixture seeding 复用同一套 factory；
//   - 零外部依赖，不使用反射，保持类型安全；
//   - 默认值覆盖所有必填字段，测试只需 WithXxx 关心的字段。
//
// 用法示例：
//
//	req := factories.NewRequirementFactory().
//	    WithTitle("修复登录页 500").
//	    WithPriority(issue_pkg.PriorityHigh).
//	    Build()
package factories

import (
	"fmt"
	"time"

	issue_pkg "github.com/njydsz/ydsz-plane/internal/application/issue"
)

// factoryMonotonic 用于生成唯一序列值（非并发安全，单测串行足够）。
var factoryMonotonic int64

func nextSeq() int64 {
	factoryMonotonic++
	return factoryMonotonic
}

func ptrInt64(v int64) *int64 { return &v }
func ptrInt(v int) *int       { return &v }
func ptrStr(v string) *string { return &v }

func now() time.Time { return time.Now().UTC() }
func nowPtr() *time.Time {
	t := now()
	return &t
}

// ---------------------------------------------------------------------------
// Requirement Factory
// ---------------------------------------------------------------------------

// RequirementFactory 生成 Requirement mock 数据。
type RequirementFactory struct {
	val issue_pkg.Requirement
}

// NewRequirementFactory 创建 Requirement 工厂，填充合理默认值。
func NewRequirementFactory() *RequirementFactory {
	seq := nextSeq()
	t := now()
	f := RequirementFactory{}
	f.val = issue_pkg.Requirement{
		ID:             seq,
		PublicID:       fmt.Sprintf("REQ-PUB-%d", seq),
		WorkspaceID:    1,
		ProjectID:      1,
		SequenceID:     seq,
		Identifier:     fmt.Sprintf("REQ-%d", seq),
		TypeCode:       issue_pkg.TypeRequirement,
		Depth:          0,
		Name:           fmt.Sprintf("需求-%d", seq),
		StateID:        1,
		Priority:       issue_pkg.PriorityMedium,
		Point:          ptrInt(3),
		Progress:       0,
		IsDraft:        false,
		SortOrder:      float64(seq) * 1000,
		Version:        1,
		Assignees:      []int64{1},
		CreatedBy:      1,
		CreatedAt:      t,
		UpdatedAt:      t,
		Source:         ptrStr("业务方"),
		ReviewStatus:   nil,
		TargetDate:     nowPtr(),
	}
	return &f
}

// WithTitle 设置需求标题。
func (f *RequirementFactory) WithTitle(t string) *RequirementFactory {
	f.val.Name = t
	return f
}

// WithStateID 设置状态 ID。
func (f *RequirementFactory) WithStateID(stateID int64) *RequirementFactory {
	f.val.StateID = stateID
	return f
}

// WithPriority 设置优先级。
func (f *RequirementFactory) WithPriority(p issue_pkg.IssuePriority) *RequirementFactory {
	f.val.Priority = p
	return f
}

// WithWorkspaceID 设置工作空间 ID。
func (f *RequirementFactory) WithWorkspaceID(wsID int64) *RequirementFactory {
	f.val.WorkspaceID = wsID
	return f
}

// WithProjectID 设置项目 ID。
func (f *RequirementFactory) WithProjectID(projID int64) *RequirementFactory {
	f.val.ProjectID = projID
	return f
}

// WithAssignees 设置分配人列表。
func (f *RequirementFactory) WithAssignees(ids []int64) *RequirementFactory {
	f.val.Assignees = ids
	return f
}

// WithLabels 设置标签列表。
func (f *RequirementFactory) WithLabels(ids []int64) *RequirementFactory {
	f.val.Labels = ids
	return f
}

// WithModules 设置模块列表。
func (f *RequirementFactory) WithModules(ids []int64) *RequirementFactory {
	f.val.Modules = ids
	return f
}

// WithPoint 设置故事点。
func (f *RequirementFactory) WithPoint(pt int) *RequirementFactory {
	f.val.Point = ptrInt(pt)
	return f
}

// WithSprintID 设置迭代 ID。
func (f *RequirementFactory) WithSprintID(sprintID int64) *RequirementFactory {
	f.val.SprintID = ptrInt64(sprintID)
	return f
}

// WithDescription 设置描述（HTML）。
func (f *RequirementFactory) WithDescription(html string) *RequirementFactory {
	f.val.DescriptionHTML = html
	return f
}

// WithCreatedBy 设置创建人 ID。
func (f *RequirementFactory) WithCreatedBy(userID int64) *RequirementFactory {
	f.val.CreatedBy = userID
	return f
}

// WithSource 设置来源。
func (f *RequirementFactory) WithSource(src string) *RequirementFactory {
	f.val.Source = ptrStr(src)
	return f
}

// WithReviewStatus 设置评审状态。
func (f *RequirementFactory) WithReviewStatus(status string) *RequirementFactory {
	f.val.ReviewStatus = ptrStr(status)
	return f
}

// WithCompletedAt 设置完成时间（同时标记进度 100、非草稿）。
func (f *RequirementFactory) WithCompletedAt(t time.Time) *RequirementFactory {
	f.val.CompletedAt = &t
	f.val.Progress = 100
	f.val.IsDraft = false
	return f
}

// Build 返回构造好的 Requirement。
func (f *RequirementFactory) Build() issue_pkg.Requirement {
	return f.val
}

// BuildView 返回跨类型只读视图。
func (f *RequirementFactory) BuildView() issue_pkg.WorkitemView {
	return f.val.ToView()
}

// ---------------------------------------------------------------------------
// Task Factory
// ---------------------------------------------------------------------------

// TaskFactory 生成 Task mock 数据。
type TaskFactory struct {
	val issue_pkg.Task
}

// NewTaskFactory 创建 Task 工厂，填充合理默认值。
func NewTaskFactory() *TaskFactory {
	seq := nextSeq()
	t := now()
	f := TaskFactory{}
	f.val = issue_pkg.Task{
		ID:             seq,
		PublicID:       fmt.Sprintf("TSK-PUB-%d", seq),
		WorkspaceID:    1,
		ProjectID:      1,
		SequenceID:     seq,
		Identifier:     fmt.Sprintf("TSK-%d", seq),
		TypeCode:       issue_pkg.TypeTask,
		Depth:          0,
		Name:           fmt.Sprintf("任务-%d", seq),
		StateID:        1,
		Priority:       issue_pkg.PriorityMedium,
		Point:          ptrInt(2),
		Progress:       0,
		IsDraft:        false,
		SortOrder:      float64(seq) * 1000,
		Version:        1,
		Assignees:      []int64{1},
		CreatedBy:      1,
		CreatedAt:      t,
		UpdatedAt:      t,
		Category:       ptrStr("功能开发"),
		TargetDate:     nowPtr(),
	}
	return &f
}

// WithTitle 设置任务标题。
func (f *TaskFactory) WithTitle(t string) *TaskFactory {
	f.val.Name = t
	return f
}

// WithStateID 设置状态 ID。
func (f *TaskFactory) WithStateID(stateID int64) *TaskFactory {
	f.val.StateID = stateID
	return f
}

// WithPriority 设置优先级。
func (f *TaskFactory) WithPriority(p issue_pkg.IssuePriority) *TaskFactory {
	f.val.Priority = p
	return f
}

// WithWorkspaceID 设置工作空间 ID。
func (f *TaskFactory) WithWorkspaceID(wsID int64) *TaskFactory {
	f.val.WorkspaceID = wsID
	return f
}

// WithProjectID 设置项目 ID。
func (f *TaskFactory) WithProjectID(projID int64) *TaskFactory {
	f.val.ProjectID = projID
	return f
}

// WithAssignees 设置分配人列表。
func (f *TaskFactory) WithAssignees(ids []int64) *TaskFactory {
	f.val.Assignees = ids
	return f
}

// WithPoint 设置故事点。
func (f *TaskFactory) WithPoint(pt int) *TaskFactory {
	f.val.Point = ptrInt(pt)
	return f
}

// WithSprintID 设置迭代 ID。
func (f *TaskFactory) WithSprintID(sprintID int64) *TaskFactory {
	f.val.SprintID = ptrInt64(sprintID)
	return f
}

// WithCategory 设置任务分类。
func (f *TaskFactory) WithCategory(cat string) *TaskFactory {
	f.val.Category = ptrStr(cat)
	return f
}

// WithActualEffort 设置实际工时（小时）。
func (f *TaskFactory) WithActualEffort(hours float64) *TaskFactory {
	f.val.ActualEffort = &hours
	return f
}

// WithRemainingEffort 设置剩余工时（小时）。
func (f *TaskFactory) WithRemainingEffort(hours float64) *TaskFactory {
	f.val.RemainingEffort = &hours
	return f
}

// WithDelayReason 设置延期原因。
func (f *TaskFactory) WithDelayReason(reason string) *TaskFactory {
	f.val.DelayReason = ptrStr(reason)
	return f
}

// WithDescription 设置描述（HTML）。
func (f *TaskFactory) WithDescription(html string) *TaskFactory {
	f.val.DescriptionHTML = html
	return f
}

// WithCreatedBy 设置创建人 ID。
func (f *TaskFactory) WithCreatedBy(userID int64) *TaskFactory {
	f.val.CreatedBy = userID
	return f
}

// Build 返回构造好的 Task。
func (f *TaskFactory) Build() issue_pkg.Task {
	return f.val
}

// BuildView 返回跨类型只读视图。
func (f *TaskFactory) BuildView() issue_pkg.WorkitemView {
	return f.val.ToView()
}

// ---------------------------------------------------------------------------
// Defect Factory
// ---------------------------------------------------------------------------

// DefectFactory 生成 Defect mock 数据。
type DefectFactory struct {
	val issue_pkg.Defect
}

// NewDefectFactory 创建 Defect 工厂，填充合理默认值。
func NewDefectFactory() *DefectFactory {
	seq := nextSeq()
	t := now()
	f := DefectFactory{}
	f.val = issue_pkg.Defect{
		ID:                seq,
		PublicID:          fmt.Sprintf("DEF-PUB-%d", seq),
		WorkspaceID:       1,
		ProjectID:         1,
		SequenceID:        seq,
		Identifier:        fmt.Sprintf("DEF-%d", seq),
		TypeCode:          issue_pkg.TypeDefect,
		Depth:             0,
		Name:              fmt.Sprintf("缺陷-%d", seq),
		StateID:           1,
		Priority:          issue_pkg.PriorityMedium,
		Progress:          0,
		IsDraft:           false,
		SortOrder:         float64(seq) * 1000,
		Version:           1,
		Assignees:         []int64{1},
		CreatedBy:         1,
		CreatedAt:         t,
		UpdatedAt:         t,
		Severity:          3,
		FoundPhase:        "系统测试",
		ReproduceSteps:    map[string]any{"steps": "1. 打开登录页\n2. 输入正确凭证\n3. 点击登录按钮"},
		FoundVersionID:    ptrInt64(1),
		TargetDate:        nowPtr(),
	}
	return &f
}

// WithTitle 设置缺陷标题。
func (f *DefectFactory) WithTitle(t string) *DefectFactory {
	f.val.Name = t
	return f
}

// WithStateID 设置状态 ID。
func (f *DefectFactory) WithStateID(stateID int64) *DefectFactory {
	f.val.StateID = stateID
	return f
}

// WithPriority 设置优先级。
func (f *DefectFactory) WithPriority(p issue_pkg.IssuePriority) *DefectFactory {
	f.val.Priority = p
	return f
}

// WithSeverity 设置严重程度（1-5）。
func (f *DefectFactory) WithSeverity(sev int) *DefectFactory {
	f.val.Severity = sev
	return f
}

// WithWorkspaceID 设置工作空间 ID。
func (f *DefectFactory) WithWorkspaceID(wsID int64) *DefectFactory {
	f.val.WorkspaceID = wsID
	return f
}

// WithProjectID 设置项目 ID。
func (f *DefectFactory) WithProjectID(projID int64) *DefectFactory {
	f.val.ProjectID = projID
	return f
}

// WithAssignees 设置分配人列表。
func (f *DefectFactory) WithAssignees(ids []int64) *DefectFactory {
	f.val.Assignees = ids
	return f
}

// WithFoundPhase 设置发现阶段。
func (f *DefectFactory) WithFoundPhase(phase string) *DefectFactory {
	f.val.FoundPhase = phase
	return f
}

// WithRootCauseCategory 设置根因分类。
func (f *DefectFactory) WithRootCauseCategory(cat string) *DefectFactory {
	f.val.RootCauseCategory = ptrStr(cat)
	return f
}

// WithVerifierID 设置验证人 ID。
func (f *DefectFactory) WithVerifierID(userID int64) *DefectFactory {
	f.val.VerifierID = ptrInt64(userID)
	return f
}

// WithDescription 设置描述（HTML）。
func (f *DefectFactory) WithDescription(html string) *DefectFactory {
	f.val.DescriptionHTML = html
	return f
}

// WithCreatedBy 设置创建人 ID。
func (f *DefectFactory) WithCreatedBy(userID int64) *DefectFactory {
	f.val.CreatedBy = userID
	return f
}

// WithRegressionRisk 设置回归风险。
func (f *DefectFactory) WithRegressionRisk(risk string) *DefectFactory {
	f.val.RegressionRisk = ptrStr(risk)
	return f
}

// WithFixVersionID 设置修复版本 ID。
func (f *DefectFactory) WithFixVersionID(verID int64) *DefectFactory {
	f.val.FixVersionID = ptrInt64(verID)
	return f
}

// Build 返回构造好的 Defect。
func (f *DefectFactory) Build() issue_pkg.Defect {
	return f.val
}

// BuildView 返回跨类型只读视图。
func (f *DefectFactory) BuildView() issue_pkg.WorkitemView {
	return f.val.ToView()
}
