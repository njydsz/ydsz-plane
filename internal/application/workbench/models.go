// Package workbench — 个人工作台应用服务。
//
// 对标:
//   - Plane: /api/workspaces/{slug}/users/me/workbench/
//   - Linear: viewer.assignedIssues / viewer.activeProjects
//   - Jira: /rest/api/latest/dashboard + /rest/api/latest/myPermissions
//
// 聚合接口设计：
//   - 一个 /workbench summary 调用返回今日待办、进行中、逾期、阻塞、最近访问、参与迭代概览
//   - 避免前端 N+1 请求，提升首屏加载性能
//   - 基于 workbench_configs 实现个性化布局持久化
package workbench

import "time"

// --- Workbench Summary (主聚合) ---

// WorkbenchSummary 工作台首屏聚合数据。
type WorkbenchSummary struct {
	MyIssues      MyIssuesBucket    `json:"my_issues"`       // 我的任务分桶
	SprintOverviews []SprintOverview `json:"sprint_overviews"` // 参与迭代概览
	RecentItems   []RecentItem      `json:"recent_items"`    // 最近访问
	OverdueCount  int               `json:"overdue_count"`   // 逾期总数
	BlockedCount  int               `json:"blocked_count"`   // 阻塞总数
	QuickActions  QuickActionSet    `json:"quick_actions"`   // 快捷操作
}

// MyIssuesBucket 我的工作项分桶视图。
type MyIssuesBucket struct {
	Total     int           `json:"total"`      // 总计（不含取消/已完成）
	Today     []IssueDigest `json:"today"`      // 今日任务（target_date = 今天）
	Upcoming  []IssueDigest `json:"upcoming"`   // 即将开始（未开始 + 7 天内）
	Overdue   []IssueDigest `json:"overdue"`    // 逾期（target_date < 今天 + 未完成）
	InProgress []IssueDigest `json:"in_progress"` // 进行中
	Backlog   []IssueDigest `json:"backlog"`    // 待规划
}

// IssueDigest 工作项工作台摘要。
type IssueDigest struct {
	ID          int64   `json:"id"`
	Identifier  string  `json:"identifier"`
	Title       string  `json:"title"`
	TypeCode    string  `json:"type_code"`
	Priority    string  `json:"priority"`
	StateID     int64   `json:"state_id"`
	StateName   string  `json:"state_name"`
	StateColor  string  `json:"state_color"`
	GroupID     int64   `json:"group_id"`
	ProjectName string  `json:"project_name"`
	SprintID    *int64  `json:"sprint_id,omitempty"`
	SprintName  string  `json:"sprint_name,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
	IsBlocked   bool    `json:"is_blocked"`
}

// SprintOverview 迭代工作台概览。
type SprintOverview struct {
	SprintID      int64   `json:"sprint_id"`
	SprintName    string  `json:"sprint_name"`
	ProjectID     int64   `json:"project_id"`
	ProjectName   string  `json:"project_name"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`       // 0-1
	MyIssueCount  int     `json:"my_issue_count"` // 该迭代中我的工作项数
	DaysRemaining int     `json:"days_remaining"` // 负数 = 已结束
	Goal          string  `json:"goal"`
}

// RecentItem 最近访问条目。
type RecentItem struct {
	ItemType   string `json:"item_type"`
	ItemID     int64  `json:"item_id"`
	ProjectID  int64  `json:"project_id"`
	Title      string `json:"title"`
	Identifier string `json:"identifier,omitempty"`
	AccessedAt string `json:"accessed_at"`
	URL        string `json:"url"`
}

// QuickActionSet 快捷操作入口。
type QuickActionSet struct {
	CanCreateIssue   bool `json:"can_create_issue"`
	CanStartSprint   bool `json:"can_start_sprint"`
	ActiveIssueCount int  `json:"active_issue_count"` // 进行中的数量（用于 Focus Mode Entry）
}

// --- Layout / Config ---

// WorkbenchConfig 工作台布局配置。
type WorkbenchConfig struct {
	ID           int64          `json:"id"`
	WorkspaceID  int64          `json:"workspace_id"`
	ProjectID    *int64         `json:"project_id,omitempty"`
	UserID       int64          `json:"user_id"`
	Layout       LayoutConfig   `json:"layout"`
	WidgetStates map[string]any `json:"widget_states"`
	FocusEnabled bool           `json:"focus_enabled"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// LayoutConfig 拖拽布局配置。
type LayoutConfig struct {
	Widgets []LayoutWidget `json:"widgets"`
}

// LayoutWidget 单个 Widget 布局。
type LayoutWidget struct {
	Type string `json:"type"`
	W    int    `json:"w"` // 列宽
	H    int    `json:"h"` // 行高
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// SaveLayoutInput 保存布局。
type SaveLayoutInput struct {
	WorkspaceID  int64
	ProjectID    *int64
	UserID       int64
	Layout       LayoutConfig
	WidgetStates map[string]any
	FocusEnabled bool
}

// --- Recent Items ---

// RecordRecentInput 记录最近访问。
type RecordRecentInput struct {
	WorkspaceID int64
	UserID      int64
	ItemType    string
	ItemID      int64
	ProjectID   *int64
	Title       string
	Identifier  string
}

// --- Templates ---

// WorkbenchTemplate 工作台模板。
type WorkbenchTemplate struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Layout      LayoutConfig   `json:"layout"`
	Icon        string         `json:"icon"`
	IsDefault   bool           `json:"is_default"`
	SortOrder   int            `json:"sort_order"`
}

// ApplyTemplateInput 应用模板到工作台。
type ApplyTemplateInput struct {
	WorkspaceID int64
	ProjectID   *int64
	UserID      int64
	TemplateSlug string
}
