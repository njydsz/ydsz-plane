// Package issue — 工作项领域（Requirement / Task / Defect 各自为独立聚合根，无公共基类）。
//
// 参考: Plane / Linear / Jira 的工作项模型；状态机见 docs/architecture/07。
//
// 设计原则：每个聚合根携带自身完整字段，不抽取 BaseWorkitem 之类公共封装。
// 字段重复是聚合根独立性的代价，换取类型边界清晰、改动互不干扰。
package issue

import "time"

// IssueTypeCode 工作项类型枚举。
//
// 对标 Plane 的 IssueType：epic 作为顶层容器，requirement / task / defect 作为可执行工作项。
// Epic 可包含多个 requirement / task / defect 作为子项（通过 parent_id 关联）。
type IssueTypeCode string

const (
	TypeEpic        IssueTypeCode = "epic"
	TypeRequirement IssueTypeCode = "requirement"
	TypeTask        IssueTypeCode = "task"
	TypeDefect      IssueTypeCode = "defect"
)

// IssuePriority 优先级枚举。
type IssuePriority string

const (
	PriorityUrgent IssuePriority = "urgent"
	PriorityHigh   IssuePriority = "high"
	PriorityMedium IssuePriority = "medium"
	PriorityLow    IssuePriority = "low"
	PriorityNone   IssuePriority = "none"
)

// StateGroup 状态分组（决定工作项在看板中的列位置）。
type StateGroup string

const (
	GroupBacklog   StateGroup = "backlog"
	GroupStarted   StateGroup = "started"
	GroupCompleted StateGroup = "completed"
	GroupCancelled StateGroup = "cancelled"
)

// PriorityWeight 数值越大越紧急（用于排序）。
var PriorityWeight = map[IssuePriority]int{
	PriorityNone:   0,
	PriorityLow:    1,
	PriorityMedium: 2,
	PriorityHigh:   3,
	PriorityUrgent: 4,
}

// State 工作项状态（项目维度）。
type State struct {
	ID              int64      `json:"id"`
	WorkspaceID     int64      `json:"workspace_id"`
	ProjectID       int64      `json:"project_id"`
	Name            string     `json:"name"`
	Group           StateGroup `json:"group"`
	Color           string     `json:"color"`
	Sequence        float64    `json:"sequence"`
	IsDefault       bool       `json:"is_default"`
	ApplicableTypes []string   `json:"applicable_types"` // 适用的工作项类型：all|epic|requirement|task|defect
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// --- 需求 ---

// Requirement 需求工作项（独立聚合根）。
type Requirement struct {
	ID                 int64          `json:"id"`
	PublicID           string         `json:"public_id"`
	WorkspaceID        int64          `json:"workspace_id"`
	ProjectID          int64          `json:"project_id"`
	SequenceID         int64          `json:"sequence_id"`
	Identifier         string         `json:"identifier"`
	TypeCode           IssueTypeCode  `json:"type_code"`
	ParentID           *int64         `json:"parent_id,omitempty"`
	Depth              int            `json:"depth"`
	Name               string         `json:"name"`
	DescriptionJSON    map[string]any `json:"description_json,omitempty"`
	DescriptionHTML    string         `json:"description_html,omitempty"`
	StateID            int64          `json:"state_id"`
	State              *State         `json:"state,omitempty"`
	Priority           IssuePriority  `json:"priority"`
	Point              *int           `json:"point,omitempty"`
	SprintID           *int64         `json:"sprint_id,omitempty"`
	VersionID          *int64         `json:"version_id,omitempty"`
	Progress           int            `json:"progress"`
	StartDate          *time.Time     `json:"start_date,omitempty"`
	TargetDate         *time.Time     `json:"target_date,omitempty"`
	CompletedAt        *time.Time     `json:"completed_at,omitempty"`
	IsDraft            bool           `json:"is_draft"`
	SortOrder          float64        `json:"sort_order"`
	Version            int            `json:"version"`
	Assignees          []int64        `json:"assignees,omitempty"`
	Labels             []int64        `json:"labels,omitempty"`
	Modules            []int64        `json:"modules,omitempty"`
	Watchers           []int64        `json:"watchers,omitempty"`
	CreatedBy          int64          `json:"created_by"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	// 需求专属字段
	Source             *string        `json:"source,omitempty"`
	AcceptanceCriteria map[string]any `json:"acceptance_criteria,omitempty"`
	BusinessValue      string         `json:"business_value,omitempty"`
	ReviewStatus       *string        `json:"review_status,omitempty"`
}

// --- 任务 ---

// Task 任务工作项（独立聚合根）。
type Task struct {
	ID                int64          `json:"id"`
	PublicID          string         `json:"public_id"`
	WorkspaceID       int64          `json:"workspace_id"`
	ProjectID         int64          `json:"project_id"`
	SequenceID        int64          `json:"sequence_id"`
	Identifier        string         `json:"identifier"`
	TypeCode          IssueTypeCode  `json:"type_code"`
	ParentID          *int64         `json:"parent_id,omitempty"`
	Depth             int            `json:"depth"`
	Name              string         `json:"name"`
	DescriptionJSON   map[string]any `json:"description_json,omitempty"`
	DescriptionHTML   string         `json:"description_html,omitempty"`
	StateID           int64          `json:"state_id"`
	State             *State         `json:"state,omitempty"`
	Priority          IssuePriority  `json:"priority"`
	Point             *int           `json:"point,omitempty"`
	SprintID          *int64         `json:"sprint_id,omitempty"`
	VersionID         *int64         `json:"version_id,omitempty"`
	Progress          int            `json:"progress"`
	StartDate         *time.Time     `json:"start_date,omitempty"`
	TargetDate        *time.Time     `json:"target_date,omitempty"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
	IsDraft           bool           `json:"is_draft"`
	SortOrder         float64        `json:"sort_order"`
	Version           int            `json:"version"`
	Assignees         []int64        `json:"assignees,omitempty"`
	Labels            []int64        `json:"labels,omitempty"`
	Modules           []int64        `json:"modules,omitempty"`
	Watchers          []int64        `json:"watchers,omitempty"`
	CreatedBy         int64          `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	// 任务专属字段
	Category          *string  `json:"category,omitempty"`
	ActualEffort      *float64 `json:"actual_effort,omitempty"`
	RemainingEffort   *float64 `json:"remaining_effort,omitempty"`
	DelayReason       *string  `json:"delay_reason,omitempty"`
}

// --- 缺陷 ---

// Defect 缺陷工作项（独立聚合根）。
type Defect struct {
	ID                int64          `json:"id"`
	PublicID          string         `json:"public_id"`
	WorkspaceID       int64          `json:"workspace_id"`
	ProjectID         int64          `json:"project_id"`
	SequenceID        int64          `json:"sequence_id"`
	Identifier        string         `json:"identifier"`
	TypeCode          IssueTypeCode  `json:"type_code"`
	ParentID          *int64         `json:"parent_id,omitempty"`
	Depth             int            `json:"depth"`
	Name              string         `json:"name"`
	DescriptionJSON   map[string]any `json:"description_json,omitempty"`
	DescriptionHTML   string         `json:"description_html,omitempty"`
	StateID           int64          `json:"state_id"`
	State             *State         `json:"state,omitempty"`
	Priority          IssuePriority  `json:"priority"`
	Point             *int           `json:"point,omitempty"`
	SprintID          *int64         `json:"sprint_id,omitempty"`
	VersionID         *int64         `json:"version_id,omitempty"`
	Progress          int            `json:"progress"`
	StartDate         *time.Time     `json:"start_date,omitempty"`
	TargetDate        *time.Time     `json:"target_date,omitempty"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
	IsDraft           bool           `json:"is_draft"`
	SortOrder         float64        `json:"sort_order"`
	Version           int            `json:"version"`
	Assignees         []int64        `json:"assignees,omitempty"`
	Labels            []int64        `json:"labels,omitempty"`
	Modules           []int64        `json:"modules,omitempty"`
	Watchers          []int64        `json:"watchers,omitempty"`
	CreatedBy         int64          `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	// 缺陷专属字段
	Severity          int            `json:"severity"`
	FoundPhase        string         `json:"found_phase"`
	RootCauseCategory *string        `json:"root_cause_category,omitempty"`
	VerifierID        *int64         `json:"verifier_id,omitempty"`
	Environment       map[string]any `json:"environment,omitempty"`
	ReproduceSteps    map[string]any `json:"reproduce_steps"`
	FixSteps          map[string]any `json:"fix_steps,omitempty"`
	RegressionRisk    *string        `json:"regression_risk,omitempty"`
	FoundVersionID    *int64         `json:"found_version_id,omitempty"`
	FixVersionID      *int64         `json:"fix_version_id,omitempty"`
}

// --- 跨类型视图（只读，供列表/搜索等场景返回）---

// WorkitemView 跨类型只读视图 — 从各表汇聚字段的通用投影。
// 用于必须跨类型统一返回的地方（列表页、搜索），各聚合根结构体仍保持独立。
type WorkitemView struct {
	ID          int64         `json:"id"`
	PublicID    string        `json:"public_id"`
	WorkspaceID int64         `json:"workspace_id"`
	ProjectID   int64         `json:"project_id"`
	SequenceID  int64         `json:"sequence_id"`
	Identifier  string        `json:"identifier"`
	TypeCode    IssueTypeCode `json:"type_code"`
	ParentID    *int64        `json:"parent_id,omitempty"`
	Depth       int           `json:"depth"`
	Name        string        `json:"name"`
	StateID     int64         `json:"state_id"`
	State       *State        `json:"state,omitempty"`
	Priority    IssuePriority `json:"priority"`
	SprintID    *int64        `json:"sprint_id,omitempty"`
	VersionID   *int64        `json:"version_id,omitempty"`
	Progress    int           `json:"progress"`
	TargetDate  *time.Time    `json:"target_date,omitempty"`
	IsDraft     bool          `json:"is_draft"`
	SortOrder   float64       `json:"sort_order"`
	Version     int           `json:"version"`
	CreatedBy   int64         `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	// 按类型的可选项
	Severity          *int           `json:"severity,omitempty"`
	FoundPhase        *string        `json:"found_phase,omitempty"`
	Category          *string        `json:"category,omitempty"`
	Source            *string        `json:"source,omitempty"`
	Assignees         []int64        `json:"assignees,omitempty"`
	Labels            []int64        `json:"labels,omitempty"`
	Modules           []int64        `json:"modules,omitempty"`
}

// ToView 将 Requirement 投影为跨类型只读视图。
func (r Requirement) ToView() WorkitemView {
	return WorkitemView{
		ID: r.ID, PublicID: r.PublicID, WorkspaceID: r.WorkspaceID, ProjectID: r.ProjectID,
		SequenceID: r.SequenceID, Identifier: r.Identifier, TypeCode: r.TypeCode,
		ParentID: r.ParentID, Depth: r.Depth, Name: r.Name, StateID: r.StateID, State: r.State,
		Priority: r.Priority, SprintID: r.SprintID, VersionID: r.VersionID, Progress: r.Progress,
		TargetDate: r.TargetDate, IsDraft: r.IsDraft, SortOrder: r.SortOrder, Version: r.Version,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Source: r.Source, Assignees: r.Assignees, Labels: r.Labels, Modules: r.Modules,
	}
}

// ToView 将 Task 投影为跨类型只读视图。
func (t Task) ToView() WorkitemView {
	return WorkitemView{
		ID: t.ID, PublicID: t.PublicID, WorkspaceID: t.WorkspaceID, ProjectID: t.ProjectID,
		SequenceID: t.SequenceID, Identifier: t.Identifier, TypeCode: t.TypeCode,
		ParentID: t.ParentID, Depth: t.Depth, Name: t.Name, StateID: t.StateID, State: t.State,
		Priority: t.Priority, SprintID: t.SprintID, VersionID: t.VersionID, Progress: t.Progress,
		TargetDate: t.TargetDate, IsDraft: t.IsDraft, SortOrder: t.SortOrder, Version: t.Version,
		CreatedBy: t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		Category: t.Category, Assignees: t.Assignees, Labels: t.Labels, Modules: t.Modules,
	}
}

// ToView 将 Defect 投影为跨类型只读视图。
func (d Defect) ToView() WorkitemView {
	severity := d.Severity
	return WorkitemView{
		ID: d.ID, PublicID: d.PublicID, WorkspaceID: d.WorkspaceID, ProjectID: d.ProjectID,
		SequenceID: d.SequenceID, Identifier: d.Identifier, TypeCode: d.TypeCode,
		ParentID: d.ParentID, Depth: d.Depth, Name: d.Name, StateID: d.StateID, State: d.State,
		Priority: d.Priority, SprintID: d.SprintID, VersionID: d.VersionID, Progress: d.Progress,
		TargetDate: d.TargetDate, IsDraft: d.IsDraft, SortOrder: d.SortOrder, Version: d.Version,
		CreatedBy: d.CreatedBy, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
		Severity: &severity, FoundPhase: &d.FoundPhase,
		Assignees: d.Assignees, Labels: d.Labels, Modules: d.Modules,
	}
}

// --- 领域不变量与输入 ---

// CreateRequirementInput 创建需求入参。
type CreateRequirementInput struct {
	WorkspaceID       int64
	ProjectID         int64
	Name              string
	DescriptionHTML   string
	StateID           int64
	Priority          IssuePriority
	ParentID          *int64
	Source            *string
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Point             *int
	StartDate         *time.Time
	TargetDate        *time.Time
	IsDraft           bool
	CreatedBy         int64
}

// CreateTaskInput 创建任务入参。
type CreateTaskInput struct {
	WorkspaceID     int64
	ProjectID       int64
	Name            string
	DescriptionHTML string
	StateID         int64
	Priority        IssuePriority
	ParentID        *int64
	Category        *string
	Assignees       []int64
	Labels          []int64
	Modules         []int64
	Point           *int
	StartDate       *time.Time
	TargetDate      *time.Time
	IsDraft         bool
	CreatedBy       int64
}

// CreateDefectInput 创建缺陷入参。
type CreateDefectInput struct {
	WorkspaceID       int64
	ProjectID         int64
	Name              string
	DescriptionHTML   string
	StateID           int64
	Priority          IssuePriority
	ParentID          *int64
	Severity          int
	FoundPhase        string
	ReproduceSteps    map[string]any
	Environment       map[string]any
	SourceVersionID   *int64
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Point             *int
	StartDate         *time.Time
	TargetDate        *time.Time
	IsDraft           bool
	CreatedBy         int64
}

// UpdateRequirementInput 更新需求入参。
type UpdateRequirementInput struct {
	Name              *string
	DescriptionHTML   *string
	Priority          *IssuePriority
	ParentID          *int64
	Source            *string
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Point             *int
	TargetDate        *time.Time
	Progress          *int
	Version           int
}

// UpdateTaskInput 更新任务入参。
type UpdateTaskInput struct {
	Name              *string
	DescriptionHTML   *string
	Priority          *IssuePriority
	ParentID          *int64
	Category          *string
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Point             *int
	TargetDate        *time.Time
	Progress          *int
	Version           int
}

// UpdateDefectInput 更新缺陷入参。
type UpdateDefectInput struct {
	Name              *string
	DescriptionHTML   *string
	Priority          *IssuePriority
	ParentID          *int64
	Severity          *int
	FoundPhase        *string
	RootCauseCategory *string
	VerifierID        *int64
	ReproduceSteps    map[string]any
	FixSteps          map[string]any
	RegressionRisk    *string
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Point             *int
	TargetDate        *time.Time
	Progress          *int
	FoundVersionID    *int64
	FixVersionID      *int64
	Version           int
}

// ListWorkitemsOptions 跨类型列表查询选项。
type ListWorkitemsOptions struct {
	WorkspaceID      int64
	ProjectID        int64
	StateID          *int64
	Group            *StateGroup
	TypeCode         *IssueTypeCode
	Priority         *IssuePriority
	ParentID         *int64
	Search           string
	SortBy           string
	SortDesc         bool
	Limit            int
	Offset           int
	AssigneeID       *int64
	LabelID          *int64
	ModuleID         *int64
	SprintID         *int64
	StartDateFrom    *string // ISO date string
	TargetDateTo     *string
	SeverityFrom     *int
}

// BatchUpdateInput 批量操作输入。
type BatchUpdateInput struct {
	IDs        []int64
	ToStateID  *int64
	AssigneeID *int64
	Priority   *string
	Delete     bool
}

// BatchResult 批量操作结果。
type BatchResult struct {
	Succeeded int
	Failed    int
}

// --- 其他领域实体 ---

// WorkitemExtension 工作项扩展属性
type WorkitemExtension struct {
	ID          int64          `json:"id"`
	WorkspaceID int64          `json:"workspace_id"`
	ProjectID   int64          `json:"project_id"`
	EntityType  IssueTypeCode  `json:"entity_type"`
	EntityID    int64          `json:"entity_id"`
	FieldName   string         `json:"field_name"`
	FieldValue  map[string]any `json:"field_value"`
	FieldSchema map[string]any `json:"field_schema"`
	CreatedBy   int64          `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// BizEntityRelation 工作项关联关系
type BizEntityRelation struct {
	ID            int64         `json:"id"`
	WorkspaceID   int64         `json:"workspace_id"`
	ProjectID     int64         `json:"project_id"`
	SourceType    IssueTypeCode `json:"source_type"`
	SourceID      int64         `json:"source_id"`
	TargetType    IssueTypeCode `json:"target_type"`
	TargetID      int64         `json:"target_id"`
	RelationType  string        `json:"relation_type"`
	CreatedBy     int64         `json:"created_by"`
	CreatedAt     time.Time     `json:"created_at"`
}

// Module 项目模块。
type Module struct {
	ID          int64      `json:"id"`
	PublicID    string     `json:"public_id,omitempty"`
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   int64      `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	LeadID      *int64     `json:"lead_id,omitempty"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	SortOrder   float64    `json:"sort_order"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Label 项目标签。
type Label struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description,omitempty"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IssueActivity 活动日志条目。
type IssueActivity struct {
	ID          int64          `json:"id"`
	WorkspaceID int64          `json:"workspace_id"`
	ProjectID   int64          `json:"project_id"`
	Verb        string         `json:"verb"`
	IssueID     int64          `json:"issue_id"`
	Field       *string        `json:"field,omitempty"`
	OldValue    *string        `json:"old_value,omitempty"`
	NewValue    *string        `json:"new_value,omitempty"`
	OldRef      map[string]any `json:"old_ref,omitempty"`
	NewRef      map[string]any `json:"new_ref,omitempty"`
	ActorID     *int64         `json:"actor_id,omitempty"`
	ActorEmail  string         `json:"actor_email,omitempty"`
	ActorName   string         `json:"actor_name,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// TimeLog 工时记录。
type TimeLog struct {
	ID              int64     `json:"id"`
	WorkspaceID     int64     `json:"workspace_id"`
	ProjectID       int64     `json:"project_id"`
	IssueID         int64     `json:"issue_id"`
	UserID          int64     `json:"user_id"`
	SpentDate       time.Time `json:"spent_date"`
	DurationMinutes int       `json:"duration_minutes"`
	Description     string    `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// StateTransition 状态流转规则。
type StateTransition struct {
	ID             int64    `json:"id"`
	WorkspaceID    int64    `json:"workspace_id"`
	ProjectID      int64    `json:"project_id"`
	TypeCode       string   `json:"type_code"` // all | requirement | task | defect
	FromStateID    int64    `json:"from_state_id"`
	ToStateID      int64    `json:"to_state_id"`
	RequiredFields []string `json:"required_fields"`
}

// IssueRelation 关联关系。
type IssueRelation struct {
	ID            int64     `json:"id"`
	WorkspaceID   int64     `json:"workspace_id"`
	ProjectID     int64     `json:"project_id"`
	SourceIssueID int64     `json:"source_issue_id"`
	TargetIssueID int64     `json:"target_issue_id"`
	RelationType  string    `json:"relation_type"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// IssueDependency 依赖关系。
type IssueDependency struct {
	ID             int64     `json:"id"`
	WorkspaceID    int64     `json:"workspace_id"`
	ProjectID      int64     `json:"project_id"`
	PredecessorID  int64     `json:"predecessor_id"`
	SuccessorID    int64     `json:"successor_id"`
	DependencyType string    `json:"dependency_type"` // FS | SS | FF | SF
	LagDays        int       `json:"lag_days"`
	CreatedBy      int64     `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

// IssueReaction 工作项表情反应（emoji 轻量反馈）。
type IssueReaction struct {
	ID           int64     `json:"id"`
	WorkspaceID  int64     `json:"workspace_id"`
	ProjectID    int64     `json:"project_id"`
	IssueID      int64     `json:"issue_id"`
	UserID       int64     `json:"user_id"`
	ReactionType string    `json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// IssueVote 工作项投票（赞成/反对）。
type IssueVote struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	IssueID     int64     `json:"issue_id"`
	UserID      int64     `json:"user_id"`
	Vote        int       `json:"vote"` // 1=赞成 -1=反对
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReactionSummary 单种表情的聚合统计（详情/列表展示）。
type ReactionSummary struct {
	ReactionType string `json:"reaction_type"`
	Count        int    `json:"count"`
	Reacted      bool   `json:"reacted"` // 当前用户是否已反应
}

// VoteSummary 投票聚合统计。
type VoteSummary struct {
	Upvotes   int  `json:"upvotes"`
	Downvotes int  `json:"downvotes"`
	Score     int  `json:"score"`
	Voted     *int `json:"voted,omitempty"` // 当前用户投票: 1 / -1 / nil(未投)
}
