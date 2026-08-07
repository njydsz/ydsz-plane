// Package issue — 工作项领域（Issue Aggregate：需求/任务/缺陷统一模型）。
//
// 参考: Plane / Linear / Jira 的工作项模型；状态机见 docs/architecture/07。
package issue

import "time"

// IssueTypeCode 工作项类型枚举。
type IssueTypeCode string

const (
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
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Group       StateGroup `json:"group"`
	Color       string    `json:"color"`
	Sequence    float64   `json:"sequence"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Issue 工作项域模型（API 响应 DTO）。
type Issue struct {
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
	Severity          *int           `json:"severity,omitempty"`
	FoundPhase        *string        `json:"found_phase,omitempty"`
	RootCauseCategory *string        `json:"root_cause_category,omitempty"`
	VerifierID        *int64         `json:"verifier_id,omitempty"`
	Environment       map[string]any `json:"environment,omitempty"`
	ReproduceSteps    map[string]any `json:"reproduce_steps,omitempty"`
	Category          *string        `json:"category,omitempty"`
	ActualEffort      *float64       `json:"actual_effort,omitempty"`
	RemainingEffort   *float64       `json:"remaining_effort,omitempty"`
	DelayReason       *string        `json:"delay_reason,omitempty"`
	Source            *string        `json:"source,omitempty"`
	Point             *int           `json:"point,omitempty"`
	SprintID          *int64         `json:"sprint_id,omitempty"`
	FoundVersionID    *int64         `json:"found_version_id,omitempty"`
	FixVersionID      *int64         `json:"fix_version_id,omitempty"`
	ReleaseVersionID  *int64         `json:"release_version_id,omitempty"`
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
}

// Module 项目模块。
type Module struct {
	ID          int64      `json:"id"`
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
	IssueID     int64          `json:"issue_id"`
	Verb        string         `json:"verb"`
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
