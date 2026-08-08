// Package dashboard — 项目仪表盘应用服务。
//
// 对标:
//   - Plane: /api/workspaces/{slug}/projects/{project}/dashboard/
//   - Jira: /rest/api/latest/dashboard + /rest/api/latest/project/{key}/status
//   - Linear: project.dashboard / project.overview
//
// 设计要点:
//   - widgets 以 JSONB 配置，按 project 维度 + 用户级覆盖
//   - 数据通过 dashboard_snapshots 快照加速，worker 定期刷新
//   - 风险预警由独立检测 worker 基于 risk_rules 生成 risk_alerts
package dashboard

import "time"

// --- Widget Types ---

// WidgetType 仪表盘 widget 类型。
type WidgetType string

const (
	WidgetProgressOverview  WidgetType = "progress_overview"
	WidgetBurndown          WidgetType = "burndown"
	WidgetVelocity          WidgetType = "velocity"
	WidgetPrioritySplit     WidgetType = "priority_split"
	WidgetStateDistribution WidgetType = "state_distribution"
	WidgetOverdueList       WidgetType = "overdue_list"
	WidgetBlockedList       WidgetType = "blocked_list"
	WidgetRiskAlert         WidgetType = "risk_alert"
	WidgetRecentActivity    WidgetType = "recent_activity"
	WidgetTeamWorkload      WidgetType = "team_workload"
	WidgetVersionBurndown   WidgetType = "version_burndown"
	WidgetModuleDistribution WidgetType = "module_distribution"
	WidgetDORA              WidgetType = "dora"
	WidgetProjectCompare    WidgetType = "project_compare"
)

// DashboardWidget widget 配置。
type DashboardWidget struct {
	ID         int64          `json:"id"`
	ProjectID  int64          `json:"project_id"`
	WidgetType WidgetType     `json:"widget_type"`
	Title      string         `json:"title"`
	GridX      int            `json:"grid_x"`
	GridY      int            `json:"grid_y"`
	GridW      int            `json:"grid_w"`
	GridH      int            `json:"grid_h"`
	Config     map[string]any `json:"config"`
	IsVisible  bool           `json:"is_visible"`
	SortOrder  int            `json:"sort_order"`
	UserID     *int64         `json:"user_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// DashboardData 仪表盘完整数据（widgets + 快照数据）。
type DashboardData struct {
	Widgets  []DashboardWidget `json:"widgets"`
	Snapshots map[string]any   `json:"snapshots"`  // key = widget_type, value = snapshot data
	Alerts   []RiskAlert       `json:"alerts"`     // 未解决风险告警
}

// --- Snapshot Data ---

// ProgressOverviewWidget 进度概览 widget 数据。
type ProgressOverviewWidget struct {
	TotalIssues    int     `json:"total_issues"`
	DoneIssues     int     `json:"done_issues"`
	InProgress     int     `json:"in_progress"`
	OverdueIssues  int     `json:"overdue_issues"`
	BlockedIssues  int     `json:"blocked_issues"`
	CompletionRate float64 `json:"completion_rate"`
	ActiveSprints  int     `json:"active_sprints"`
}

// PrioritySplitWidget 优先级分布。
type PrioritySplitWidget struct {
	Total    int            `json:"total"`
	ByPriority map[string]int `json:"by_priority"`
}

// StateDistributionWidget 状态分布。
type StateDistributionWidget struct {
	Total    int            `json:"total"`
	ByState  []StateBucket  `json:"by_state"`
}

// StateBucket 状态分桶。
type StateBucket struct {
	StateID   int64  `json:"state_id"`
	StateName string `json:"state_name"`
	GroupName string `json:"group_name"`
	Color     string `json:"color"`
	Count     int    `json:"count"`
}

// BurndownWidget 燃尽图数据（汇总）。
type BurndownWidget struct {
	SprintID     int64        `json:"sprint_id"`
	SprintName   string       `json:"sprint_name"`
	TotalPoints  int          `json:"total_points"`
	BurnedPoints int          `json:"burned_points"`
	TotalIssues  int          `json:"total_issues"`
	BurnedIssues int          `json:"burned_issues"`
	RemainingDays int         `json:"remaining_days"`
	IsActive     bool         `json:"is_active"`
}

// OverdueListWidget 逾期列表（前 N 条）。
type OverdueListWidget struct {
	Total   int          `json:"total"`
	Items   []OverdueItem `json:"items"`
}

// OverdueItem 逾期工作项摘要。
type OverdueItem struct {
	ID          int64  `json:"id"`
	Identifier  string `json:"identifier"`
	Title       string `json:"title"`
	Priority    string `json:"priority"`
	OverdueDays int    `json:"overdue_days"`
	Assignee    string `json:"assignee"`
}

// BlockedListWidget 阻塞列表。
type BlockedListWidget struct {
	Total   int          `json:"total"`
	Items   []BlockedItem `json:"items"`
}

// BlockedItem 阻塞工作项。
type BlockedItem struct {
	ID             int64  `json:"id"`
	Identifier     string `json:"identifier"`
	Title          string `json:"title"`
	BlockedCount   int    `json:"blocked_count"`  // 阻塞下游数量
	BlockerNames   string `json:"blocker_names"`
}

// --- Widget Data: Team Workload ---

// TeamWorkloadWidget 团队工作负载 widget 数据。
type TeamWorkloadWidget struct {
	Members []TeamMemberWorkload `json:"members"`
}

// TeamMemberWorkload 成员工作负载。
type TeamMemberWorkload struct {
	UserID     int64  `json:"user_id"`
	UserName   string `json:"user_name"`
	Avatar     string `json:"avatar,omitempty"`
	Todo       int    `json:"todo"`
	InProgress int    `json:"in_progress"`
	Done       int    `json:"done"`
	Total      int    `json:"total"`
}

// --- Widget Data: Recent Activity ---

// RecentActivityWidget 最近活动 widget 数据。
type RecentActivityWidget struct {
	Items []ActivityItem `json:"items"`
}

// ActivityItem 活动条目。
type ActivityItem struct {
	ID              int64  `json:"id"`
	IssueID         int64  `json:"issue_id"`
	IssueIdentifier string `json:"issue_identifier"`
	ActorID         int64  `json:"actor_id"`
	ActorName       string `json:"actor_name"`
	ActorAvatar     string `json:"actor_avatar,omitempty"`
	Verb            string `json:"verb"`
	TargetState     string `json:"target_state,omitempty"`
	CreatedAt       string `json:"created_at"`
}

// --- Widget Data: Velocity ---

// VelocityWidget 速率图 widget 数据。
type VelocityWidget struct {
	Sprints []VelocityPoint `json:"sprints"`
	Average float64         `json:"average"`
}

// VelocityPoint 单迭代速率数据点。
type VelocityPoint struct {
	SprintID       int64   `json:"sprint_id"`
	SprintName     string  `json:"sprint_name"`
	CompletedCount int     `json:"completed_count"`
	CommittedCount int     `json:"committed_count"`
	CompletionRate float64 `json:"completion_rate"`
}

// --- Risk Alerts ---

// RiskAlert 风险告警。
type RiskAlert struct {
	ID          int64          `json:"id"`
	ProjectID   *int64         `json:"project_id,omitempty"`
	RuleID      int64          `json:"rule_id"`
	Severity    string         `json:"severity"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
	IsResolved  bool           `json:"is_resolved"`
	CreatedAt   time.Time      `json:"created_at"`
}

// RiskRule 风险规则。
type RiskRule struct {
	ID             int64          `json:"id"`
	WorkspaceID    int64          `json:"workspace_id"`
	ProjectID      *int64         `json:"project_id,omitempty"`
	RuleName       string         `json:"rule_name"`
	RuleType       string         `json:"rule_type"`
	ConditionJSON  map[string]any `json:"condition_json"`
	NotifyChannels []string       `json:"notify_channels"`
	IsActive       bool           `json:"is_active"`
	LastTriggered  *time.Time     `json:"last_triggered,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// --- Templates ---

// DashboardTemplate 仪表盘布局模板。
type DashboardTemplate struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Layout      map[string]any `json:"layout"`
	Icon        string         `json:"icon"`
	Category    string         `json:"category"`
	IsDefault   bool           `json:"is_default"`
	SortOrder   int            `json:"sort_order"`
}

// --- Inputs ---

// SaveWidgetInput 保存 widget 配置。
type SaveWidgetInput struct {
	ProjectID  int64
	UserID     int64
	WidgetType WidgetType
	Title      string
	GridX      int
	GridY      int
	GridW      int
	GridH      int
	Config     map[string]any
}

// UpdateWidgetInput 更新 widget 入参 — 仅更新非 nil 字段。
type UpdateWidgetInput struct {
	WidgetID  int64
	ProjectID int64
	GridX     *int
	GridY     *int
	GridW     *int
	GridH     *int
	Title     *string
	Config    map[string]any
}

// RefreshSnapshotInput 刷新 widget 快照请求。
type RefreshSnapshotInput struct {
	ProjectID  int64
	WidgetType WidgetType
	RealTime   bool  // true = 忽略快照直接查
}

// --- Workspace 级聚合: 多项目对比 ---

// ProjectCompareItem 单个项目的对比数据（需求完成率 + 缺陷数）。
type ProjectCompareItem struct {
	ProjectID       int64   `json:"project_id"`
	ProjectName     string  `json:"project_name"`
	Identifier      string  `json:"identifier"`
	TotalIssues     int     `json:"total_issues"`
	DoneIssues      int     `json:"done_issues"`
	CompletionRate  float64 `json:"completion_rate"`
	DefectCount     int     `json:"defect_count"`
	ActiveSprintCount int   `json:"active_sprint_count"`
}
