// Package sprint — Sprint 领域模型与枚举。
//
// 参考: Plane / Jira Sprint / Scrum guild。
package sprint

import "time"

// SprintStatusCode 迭代状态枚举。
type SprintStatusCode string

const (
	SprintPlanned   SprintStatusCode = "planned"
	SprintActive    SprintStatusCode = "active"
	SprintCompleted SprintStatusCode = "completed"
)

// UnfinishedStrategy 迭代结束时未完成任务的处理策略。
type UnfinishedStrategy string

const (
	UnfinishedBacklog    UnfinishedStrategy = "backlog"     // 退回 Backlog
	UnfinishedNextSprint UnfinishedStrategy = "next_sprint" // 移至下一迭代
	UnfinishedKeep       UnfinishedStrategy = "keep"        // 仅归档工作项
)

// Sprint 迭代聚合根。
// 业务规则: 一个迭代只属于一个版本 (version_id FK)。
type Sprint struct {
	ID             int64            `json:"id"`
	WorkspaceID    int64            `json:"workspace_id"`
	ProjectID      int64            `json:"project_id"`
	VersionID      *int64           `json:"version_id,omitempty"`
	Name           string           `json:"name"`
	Description    string           `json:"description,omitempty"`
	Goal           string           `json:"goal,omitempty"`
	Status         SprintStatusCode `json:"status"`
	StartDate      *time.Time       `json:"start_date,omitempty"`
	EndDate        *time.Time       `json:"end_date,omitempty"`
	Capacity       *float64         `json:"capacity,omitempty"` // 容量（故事点）
	OwnerID        *int64           `json:"owner_id,omitempty"`
	Viewport       map[string]any   `json:"viewport,omitempty"`
	// 进度（实时计算，非持久化字段）
	Progress       SprintProgress   `json:"progress,omitempty"`
	// 复盘数据
	ReviewSnapshot *ReviewSnapshot  `json:"review_snapshot,omitempty"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	CreatedBy      int64            `json:"created_by"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// SprintProgress 迭代的实时进度聚合。
type SprintProgress struct {
	TotalPoints     float64            `json:"total_points"`
	DonePoints      float64            `json:"done_points"`
	TotalIssues     int                `json:"total_issues"`
	DoneIssues      int                `json:"done_issues"`
	ByStateGroup    map[string]float64 `json:"by_state_group,omitempty"`
	Saturation      float64            `json:"saturation,omitempty"` // capacity > 0 时：total_points / capacity
}

// ReviewSnapshot 迭代结束时的复盘数据快照。
type ReviewSnapshot struct {
	CommittedPoints  float64 `json:"committed_points"`  // 启动时承诺的故事点
	CompletedPoints  float64 `json:"completed_points"`  // 结束时完成的故事点
	JoinedPoints     float64 `json:"joined_points"`     // 中途加入的故事点
	RemovedPoints    float64 `json:"removed_points"`    // 中途移除的故事点
	CommittedIssues  int     `json:"committed_issues"`
	CompletedIssues  int     `json:"completed_issues"`
	JoinedIssues     int     `json:"joined_issues"`   // 中途加入的工作项
	RemovedIssues    int     `json:"removed_issues"`  // 中途移除的工作项
	CompletionRate   float64 `json:"completion_rate"` // 完成率
}

// SprintSnapshot 每日快照聚合（用于燃尽图 / 燃起图）。
type SprintSnapshot struct {
	ID           int64          `json:"id"`
	WorkspaceID  int64          `json:"workspace_id"`
	ProjectID    int64          `json:"project_id"`
	SprintID     int64          `json:"sprint_id"`
	SnapshotDate time.Time      `json:"snapshot_date"`
	Data         SnapshotData   `json:"data"`
	CreatedAt    time.Time      `json:"created_at"`
}

// SnapshotData 快照数据结构（DB JSONB）。
type SnapshotData struct {
	TotalPoints  float64            `json:"total_points"`
	DonePoints   float64            `json:"done_points"`
	TotalIssues  int                `json:"total_issues"`
	DoneIssues   int                `json:"done_issues"`
	ByStateGroup map[string]float64 `json:"by_state_group,omitempty"`
	AddedPoints  float64            `json:"added_points"`
	RemovedPoints float64           `json:"removed_points"`
	// IssueIDs 快照时刻迭代内的工作项集合（用于跨快照对比计算 RemovedPoints）。
	IssueIDs []int64 `json:"issue_ids,omitempty"`
}

// BurndownPoint 燃尽图单点。
type BurndownPoint struct {
	Date        time.Time `json:"date"`
	TotalPoints float64   `json:"total_points"`
	DonePoints  float64   `json:"done_points"`
	Remaining   float64   `json:"remaining"`
	IdealLine   float64   `json:"ideal_line"`
}

// VelocityStats 速率统计。
type VelocityStats struct {
	AvgPoints     float64         `json:"avg_points"`
	AvgIssues     float64         `json:"avg_issues"`
	P50           float64         `json:"p50"`
	RecentSprints []SprintVelocity `json:"recent_sprints"`
	Count         int             `json:"count"`
}

// SprintVelocity 单期迭代速率。
type SprintVelocity struct {
	SprintID        int64     `json:"sprint_id"`
	SprintName      string    `json:"sprint_name"`
	CompletedPoints float64   `json:"completed_points"`
	CompletedIssues int       `json:"completed_issues"`
	EndDate         time.Time `json:"end_date"`
}
