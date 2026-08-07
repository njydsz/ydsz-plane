// Package version — 版本日领域模型与枚举。
//
// 参考: Plane / Jira Fix Version / GitHub Release。
// 版本日聚合跨迭代的交付物，作为发布、发版、Release Notes 的聚合根。
// 状态机: planning → active → released → archived (不可逆)
package version

import "time"

// VersionStatusCode 版本日状态枚举。
type VersionStatusCode string

const (
	VersionPlanning VersionStatusCode = "planning"
	VersionActive   VersionStatusCode = "active"
	VersionReleased VersionStatusCode = "released"
	VersionArchived VersionStatusCode = "archived"
)

// IsValid 判断状态是否合法。
func (s VersionStatusCode) IsValid() bool {
	switch s {
	case VersionPlanning, VersionActive, VersionReleased, VersionArchived:
		return true
	}
	return false
}

// Version 版本日聚合根。
type Version struct {
	ID             int64             `json:"id"`
	WorkspaceID    int64             `json:"workspace_id"`
	ProjectID      int64             `json:"project_id"`
	Name           string            `json:"name"`
	Semver         string            `json:"semver"`
	Description    string            `json:"description,omitempty"`
	Status         VersionStatusCode `json:"status"`
	Checklist      []ChecklistItem   `json:"checklist,omitempty"`
	ReleaseNotes   string            `json:"release_notes,omitempty"`
	DeliveredAt    *time.Time        `json:"delivered_at,omitempty"`
	TargetDate     *string           `json:"target_date,omitempty"`
	ArchivedAt     *time.Time        `json:"archived_at,omitempty"`
	CreatedBy      int64             `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	// 聚合 (按需填充)
	Sprints        []SprintRef      `json:"sprints,omitempty"`
	Progress       *VersionProgress `json:"progress,omitempty"`
	Quality        *QualityMetrics  `json:"quality,omitempty"`
	DeliveryReport *DeliveryReport  `json:"delivery_report,omitempty"`
}

// ChecklistItem 发布检查清单条目。
type ChecklistItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Checked  bool   `json:"checked"`
}

// SprintRef 版本日关联的迭代摘要。
type SprintRef struct {
	SprintID    int64             `json:"sprint_id"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	StartDate   *string           `json:"start_date,omitempty"`
	EndDate     *string           `json:"end_date,omitempty"`
	CompletedAt *string           `json:"completed_at,omitempty"`
	Progress    *SprintProgressRef `json:"progress,omitempty"`
}

// SprintProgressRef 迭代进度摘要。
type SprintProgressRef struct {
	TotalPoints float64 `json:"total_points"`
	DonePoints  float64 `json:"done_points"`
	TotalIssues int     `json:"total_issues"`
	DoneIssues  int     `json:"done_issues"`
}

// VersionProgress 版本日实时进度聚合。
type VersionProgress struct {
	TotalPoints    float64            `json:"total_points"`
	DonePoints     float64            `json:"done_points"`
	TotalIssues    int                `json:"total_issues"`
	DoneIssues     int                `json:"done_issues"`
	CompletionRate float64            `json:"completion_rate"`
	ByStateGroup   map[string]float64 `json:"by_state_group,omitempty"`
	SprintCount    int                `json:"sprint_count"`
	IssueCount     int                `json:"issue_count"`
}

// QualityMetrics 版本日质量指标（准出校验用）。
type QualityMetrics struct {
	TotalBugs       int         `json:"total_bugs"`
	OpenBugs        int         `json:"open_bugs"`
	CriticalBugs    int         `json:"critical_bugs"`
	MajorBugs       int         `json:"major_bugs"`
	BugBySeverity   map[int]int `json:"bug_by_severity,omitempty"`
	FoundBugCount   int         `json:"found_bug_count"`
	FixedBugCount   int         `json:"fixed_bug_count"`
	FixRate         float64     `json:"fix_rate"`
	PassQualityGate bool        `json:"pass_quality_gate"`
}

// DeliveryReport 交付报告数据。
type DeliveryReport struct {
	GeneratedAt       time.Time `json:"generated_at"`
	SprintCount       int       `json:"sprint_count"`
	TotalPoints       float64   `json:"total_points"`
	CompletedPoints   float64   `json:"completed_points"`
	TotalIssues       int       `json:"total_issues"`
	CompletedIssues   int       `json:"completed_issues"`
	BugCount          int       `json:"bug_count"`
	FixedBugCount     int       `json:"fixed_bug_count"`
	PassRate          float64   `json:"pass_rate"`
	EligibleToRelease bool      `json:"eligible_to_release"`
}

// NoteIssueRef Release Notes 中引用的工作项摘要。
type NoteIssueRef struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	StateName  string `json:"state_name"`
}

// ReleaseNotesData Release Notes 渲染数据源。
type ReleaseNotesData struct {
	VersionName      string         `json:"version_name"`
	Semver           string         `json:"semver"`
	RequirementsDone []NoteIssueRef `json:"requirements_done,omitempty"`
	BugsFixed        []NoteIssueRef `json:"bugs_fixed,omitempty"`
	KnownIssues      []NoteIssueRef `json:"known_issues,omitempty"`
}

// --- 输入 ---

// CreateVersionInput 创建版本日的入参。
type CreateVersionInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Semver      string
	Description string
	TargetDate  *string
	Checklist   []ChecklistItem
	CreatedBy   int64
}

// UpdateVersionInput 更新版本日字段的入参。
type UpdateVersionInput struct {
	Name        *string
	Description *string
	Semver      *string
	TargetDate  *string
	Checklist   []ChecklistItem
	Version     int
}

// ListVersionsOptions 查询选项。
type ListVersionsOptions struct {
	WorkspaceID int64
	ProjectID   int64
	Status      *VersionStatusCode
	Limit       int
	Offset      int
}

// ReleaseVersionInput 发布的入参。
type ReleaseVersionInput struct {
	DraftOverride         string
	ForceChecklist        bool
	AddKnownIssuesToNotes bool
}

// AddSprintInput 将迭代挂入版本日。
type AddSprintInput struct {
	VersionID int64
	SprintID  int64
	AddedBy   int64
}

// UpdateIssueVersionsInput 更新工作项的版本字段（found/fix/release）。
type UpdateIssueVersionsInput struct {
	IssueID          int64
	WorkspaceID      int64
	FoundVersionID   *int64
	FixVersionID     *int64
	ReleaseVersionID *int64
}

// BugVersionFilter 缺陷版本过滤条件。
type BugVersionFilter struct {
	WorkspaceID    int64
	ProjectID      int64
	FoundVersionID *int64
	FixVersionID   *int64
	StateGroup     *string
	Severity       *int
	Limit          int
	Offset         int
}

// BugVersionView 缺陷版本视图。
type BugVersionView struct {
	IssueID      int64  `json:"issue_id"`
	Identifier   string `json:"identifier"`
	Name         string `json:"name"`
	Severity     *int   `json:"severity,omitempty"`
	FoundPhase   string `json:"found_phase,omitempty"`
	StateName    string `json:"state_name"`
	StateGroup   string `json:"state_group"`
	FoundVersion string `json:"found_version,omitempty"`
	FixVersion   string `json:"fix_version,omitempty"`
	RootCause    string `json:"root_cause_category,omitempty"`
}
