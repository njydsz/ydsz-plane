// Package automation — 自动化规则引擎应用服务。
//
// 对标参考:
//   - Jira Automation (trigger → condition → action)
//   - Plane Workflow Engine (event-driven rules)
//   - GitHub Actions (reusable triggers + actions)
//   - n8n / Zapier 的可视化规则 DSL
//
// 设计原则:
//   - 事件驱动：通过 RabbitMQ EventExchange 消费领域事件
//   - 纯函数条件求值：condtion evaluator 无 IO、无副作用
//   - 串行执行：同一工作项的自动化规则通过 Redis 锁串行化，
//     避免并发写冲突
//   - 防循环：执行链路深度 ≤5，超限丢弃并告警
//   - 可观测：每次执行都写 rule_executions 审计日志
package automation

import "time"

// --- Rule Status ---

// RuleStatus 规则生命周期状态。
type RuleStatus string

const (
	RuleStatusDraft    RuleStatus = "draft"
	RuleStatusActive   RuleStatus = "active"
	RuleStatusDisabled RuleStatus = "disabled"
	RuleStatusError    RuleStatus = "error"
)

// ExecutionStatus 规则执行结果状态。
type ExecutionStatus string

const (
	ExecMatched  ExecutionStatus = "matched"
	ExecSkipped  ExecutionStatus = "skipped"
	ExecSuccess  ExecutionStatus = "success"
	ExecFailed   ExecutionStatus = "failed"
	ExecDryRun   ExecutionStatus = "dry_run"
)

// --- DSL Structures ---

// Trigger 规则触发条件。
type Trigger struct {
	Type   string         `json:"type"`   // issue.created | issue.updated | issue.status_changed | issue.commented | sprint.completed | version.released | scheduled
	Filter map[string]any `json:"filter"` // 事件级过滤: { "to_group": "started", "type_code": "defect" }
	Cron   string         `json:"cron"`   // 仅 type=scheduled 时有效
}

// Condition 单条条件表达式。
type Condition struct {
	Field string `json:"field"` // issues 字段路径: severity, state.group, estimate_points
	Op    string `json:"op"`    // eq | ne | gt | gte | lt | lte | contains | in | is_empty | is_not_empty | changed
	Value any    `json:"value"` // 比较值（op=in 时为数组）
}

// ConditionGroup 条件组合（all/any 嵌套）。
type ConditionGroup struct {
	All []ConditionGroup `json:"all"` // 默认: 所有条件均满足
	Any []ConditionGroup `json:"any"` // 任一条件满足
	// 叶节点（当 All/Any 为空时）
	Condition *Condition `json:"-"` // 通过 UnmarshalJSON 自动区分
}

// Action 单个动作定义。
type Action struct {
	Type   string         `json:"type"`   // transition | assign | update_field | notify | create_issue | copy_field | webhook_call | webhook_deliver
	Field  string         `json:"field"`  // 目标字段（transition → state, assign → user_id）
	Value  any            `json:"value"`  // 目标值（支持 ${变量}
	Config map[string]any `json:"config"` // 动作级配置：channel/template/target/strategy 等
}

// ActionType 支持的动作类型常量。
const (
	ActionTransition  = "transition"
	ActionAssign      = "assign"
	ActionUpdateField = "update_field"
	ActionNotify      = "notify"
	ActionCreateIssue = "create_issue"
	ActionCopyField   = "copy_field"
	ActionWebhookCall = "webhook_call"
)

// RuleDSL 完整的规则 DSL 定义。
type RuleDSL struct {
	Trigger    Trigger       `json:"trigger"`
	Conditions []Condition   `json:"conditions,omitempty"` // 简单条件列表（all 语义）
	Actions    []Action      `json:"actions"`
	// 高级条件组（conditions 为空时生效）
	ConditionGroup *ConditionGroup `json:"condition_group,omitempty"`
}

// --- Domain Model ---

// Rule 是聚合根：一条自动化规则的全部信息。
type Rule struct {
	ID                  int64      `json:"id"`
	WorkspaceID         int64      `json:"workspace_id"`
	ProjectID           *int64     `json:"project_id,omitempty"`
	Name                string     `json:"name"`
	Description         string     `json:"description,omitempty"`
	DSL                 RuleDSL    `json:"dsl"`
	TriggerType         string     `json:"trigger_type"`
	ActionCount         int        `json:"action_count"`
	Status              RuleStatus `json:"status"`
	CreatedBy           int64      `json:"created_by"`
	LastRunAt           *time.Time `json:"last_run_at,omitempty"`
	LastError           string     `json:"last_error,omitempty"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	ExecutionCount      int64      `json:"execution_count"`
	SortOrder           int        `json:"sort_order"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// RuleExecution 是一次规则执行的审计记录。
type RuleExecution struct {
	ID             int64           `json:"id"`
	WorkspaceID    int64           `json:"workspace_id"`
	ProjectID      *int64          `json:"project_id,omitempty"`
	RuleID         int64           `json:"rule_id"`
	TriggerEventID *int64          `json:"trigger_event_id,omitempty"`
	Status         ExecutionStatus `json:"status"`
	DurationMs     *int            `json:"duration_ms,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	ContextJSON    map[string]any  `json:"context,omitempty"`
	TriggerDepth   int             `json:"trigger_depth"`
	ViaAutomation  bool            `json:"via_automation"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Template 是预置规则模板（从 automation_templates 表读取）。
type Template struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Slug           string         `json:"slug"`
	Description    string         `json:"description"`
	Category       string         `json:"category"`
	DSLTemplate    RuleDSL        `json:"dsl_template"`
	Icon           string         `json:"icon"`
	SortOrder      int            `json:"sort_order"`
	IsRecommended  bool           `json:"is_recommended"`
}

// --- Input Types ---

// CreateRuleInput 创建规则的入参。
type CreateRuleInput struct {
	WorkspaceID int64    `json:"workspace_id"`
	ProjectID   *int64    `json:"project_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DSL         RuleDSL  `json:"dsl"`
	Status      RuleStatus `json:"status"`
	SortOrder   int      `json:"sort_order"`
	CreatedBy   int64    `json:"-"`
}

// UpdateRuleInput 更新规则的入参。
type UpdateRuleInput struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	DSL         *RuleDSL    `json:"dsl"`
	Status      *RuleStatus `json:"status"`
	SortOrder   *int        `json:"sort_order"`
	Version     int         `json:"version"` // 乐观锁
}

// ListRulesOptions 规则列表查询选项。
type ListRulesOptions struct {
	WorkspaceID int64
	ProjectID   *int64
	Status      *RuleStatus
	TriggerType *string
	Limit       int
	Offset      int
}

// --- Execution Context ---

// ExecutionContext 运行时上下文（供条件求值与动作执行使用）。
type ExecutionContext struct {
	RuleID      int64          `json:"rule_id"`
	WorkspaceID int64          `json:"workspace_id"`
	ProjectID   int64          `json:"project_id"`
	EventType   string         `json:"event_type"`
	// 事件载荷（按事件类型解析）
	Issue       *IssueContext  `json:"issue,omitempty"`
	Sprint      *SprintContext `json:"sprint,omitempty"`
	Version     *VersionContext `json:"version,omitempty"`
	Actor       ActorContext   `json:"actor"`
	// 系统变量（运行时注入）
	Now         time.Time      `json:"now"`
	Depth       int            `json:"depth"`        // 执行链路深度
	ViaAutomation bool         `json:"via_automation"` // 是否由其他规则触发
	// DryRun = true 时不实际执行动作，仅返回"将执行"列表
	DryRun      bool           `json:"dry_run"`
}

// IssueContext 运行时工作项上下文。
type IssueContext struct {
	ID            int64     `json:"id"`
	Identifier    string    `json:"identifier"`
	Name          string    `json:"name"`
	TypeCode      string    `json:"type_code"`
	StateID       int64     `json:"state_id"`
	StateName     string    `json:"state_name"`
	StateGroup    string    `json:"state_group"`
	Priority      string    `json:"priority"`
	Severity      *int      `json:"severity,omitempty"`
	EstimatePoints *int     `json:"estimate_points,omitempty"`
	CreatedBy     int64     `json:"created_by"`
	ParentID      *int64    `json:"parent_id,omitempty"`
	ProjectID     int64     `json:"project_id"`
	IsDeleted     bool      `json:"is_deleted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// SprintContext 运行时迭代上下文。
type SprintContext struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ProjectID   int64     `json:"project_id"`
	Status      string    `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// VersionContext 运行时版本上下文。
type VersionContext struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ProjectID int64  `json:"project_id"`
	Status    string `json:"status"`
}

// ActorContext 操作者上下文。
type ActorContext struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

// --- Execution Result ---

// ExecutionResult 单条规则执行结果。
type ExecutionResult struct {
	RuleID      int64          `json:"rule_id"`
	Status      ExecutionStatus `json:"status"`
	ActionsTaken []Action      `json:"actions_taken"`
	DurationMs  int            `json:"duration_ms"`
	Error       string         `json:"error,omitempty"`
	SkippedReason string       `json:"skipped_reason,omitempty"`
}

// --- Built-in Rule Templates (7 条) ---

// BuiltInTemplates 返回 PRD §9.7 预置的规则模板（共 15 条，覆盖质量/效率/通知/管理场景）。
// 注：这些模板同时存在于 DB automation_templates 表中，
// 此处以 Go 常量形式提供方便单元测试与代码引用。
func BuiltInTemplates() []Template {
	return []Template{
		// --- 质量类 ---
		{
			Slug:        "auto-complete-parent",
			Name:        "子项全部完成后自动完成父项",
			Description: "当所有子工作项都标记为完成时，自动将父工作项状态更新为已完成",
			Category:    "quality",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "issue.status_changed", Filter: map[string]any{"to_group": "completed"}},
				Conditions: []Condition{
					{Field: "parent", Op: "is_not_empty"},
				},
				Actions: []Action{
					{Type: ActionTransition, Value: "completed"},
				},
			},
			Icon:          "git-merge",
			SortOrder:     1,
			IsRecommended: true,
		},
		{
			Slug:        "defect-notify-tech-lead",
			Name:        "新缺陷通知技术负责人",
			Description: "项目里新建高优缺陷时，自动通知项目技术负责人",
			Category:    "quality",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "issue.created", Filter: map[string]any{"type_code": "defect", "priority": "urgent"}},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${project.tech_lead}", "template": "🚨 新建紧急缺陷: [{{issue.identifier}}] {{issue.name}}"}},
				},
			},
			Icon:          "alert-triangle",
			SortOrder:     2,
			IsRecommended: true,
		},
		{
			Slug:        "defect-assign-verifier",
			Name:        "缺陷修复后自动指派验证人",
			Description: "缺陷修复后自动将验证任务指派给创建者",
			Category:    "quality",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "issue.status_changed", Filter: map[string]any{"type_code": "defect", "to_group": "completed"}},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${issue.created_by}", "template": "缺陷 {{issue.identifier}} 已修复，请验证"}},
				},
			},
			Icon:      "check-circle",
			SortOrder: 3,
		},
		// --- 效率类 ---
		{
			Slug:        "auto-start-date",
			Name:        "进入'进行中'时自动填写开始日期",
			Description: "工作项首次进入进行中状态时，自动记录开始时间",
			Category:    "efficiency",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "issue.status_changed", Filter: map[string]any{"to_group": "started"}},
				Conditions: []Condition{{Field: "started_at", Op: "is_empty"}},
				Actions: []Action{
					{Type: ActionUpdateField, Field: "started_at", Value: "${now}"},
				},
			},
			Icon:          "play",
			SortOrder:     4,
			IsRecommended: true,
		},
		{
			Slug:        "auto-assign-least-loaded",
			Name:        "最闲人自动指派",
			Description: "新建工作项时自动分配给当前负载最轻的成员",
			Category:    "efficiency",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "issue.created"},
				Conditions: []Condition{{Field: "assignees", Op: "is_empty"}},
				Actions: []Action{
					{Type: ActionAssign, Config: map[string]any{"strategy": "least_loaded", "role": "member", "scope": "project"}},
				},
			},
			Icon:      "user-plus",
			SortOrder: 5,
		},
		{
			Slug:        "auto-set-priority",
			Name:        "高优需求自动标记",
			Description: "根据关键词自动设置工作项优先级",
			Category:    "efficiency",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "issue.created"},
				Conditions: []Condition{
					{Field: "issue.name", Op: "contains", Value: "紧急"},
				},
				Actions: []Action{
					{Type: ActionUpdateField, Field: "priority", Value: "urgent"},
				},
			},
			Icon:      "zap",
			SortOrder: 6,
		},
		{
			Slug:        "version-release-transition",
			Name:        "版本发布后自动流转工作项",
			Description: "版本发布时，自动将该版本下的工作项状态更新为已完成",
			Category:    "efficiency",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "version.released"},
				Conditions: []Condition{
					{Field: "issue.fix_version", Op: "eq", Value: "${version.id}"},
				},
				Actions: []Action{
					{Type: ActionTransition, Value: "completed"},
				},
			},
			Icon:      "rocket",
			SortOrder: 7,
		},
		{
			Slug:        "epic-points-rollup",
			Name:        "Epic 点数自动汇总",
			Description: "当子工作项点数变更时，自动汇总到 Epic 的聚合点数字段",
			Category:    "efficiency",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "issue.updated", Filter: map[string]any{"field_changes": []string{"estimate_points"}}},
				Conditions: []Condition{{Field: "issue.type_code", Op: "ne", Value: "epic"}},
				Actions: []Action{
					{Type: ActionCopyField, Field: "sum_children_points", Value: "${parent.estimate_points}"},
				},
			},
			Icon:      "layers",
			SortOrder: 8,
		},
		// --- 通知类 ---
		{
			Slug:        "overdue-reminder",
			Name:        "逾期提醒",
			Description: "工作项到期前 1 天自动提醒负责人",
			Category:    "notification",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "scheduled", Cron: "0 9 * * *", Filter: map[string]any{"due_within_hours": 24}},
				Conditions: []Condition{{Field: "state.group", Op: "ne", Value: "completed"}},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${issue.assignees}", "template": "工作项 {{issue.identifier}} 即将到期"}},
				},
			},
			Icon:          "clock",
			SortOrder:     9,
			IsRecommended: true,
		},
		{
			Slug:        "status-change-notify-watchers",
			Name:        "状态变更通知关注人",
			Description: "工作项状态变更时通知所有关注人",
			Category:    "notification",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "issue.status_changed"},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${issue.watchers}", "template": "{{issue.identifier}} 状态变更为 {{issue.state_name}}"}},
				},
			},
			Icon:      "bell",
			SortOrder: 10,
		},
		{
			Slug:        "sprint-complete-summary",
			Name:        "迭代完成自动通知团队",
			Description: "迭代完成时自动通知所有成员并发送总结",
			Category:    "notification",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "sprint.completed"},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${project.members}", "template": "迭代 {{sprint.name}} 已完成"}},
				},
			},
			Icon:      "flag",
			SortOrder: 11,
		},
		// --- 管理类 ---
		{
			Slug:        "sprint-auto-start-issues",
			Name:        "迭代启动后自动开始工作项",
			Description: "迭代启动后，自动将所有待办工作项流转到进行中",
			Category:    "management",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "sprint.started"},
				Conditions: []Condition{{Field: "state.group", Op: "eq", Value: "todo"}},
				Actions: []Action{
					{Type: ActionTransition, Value: "started"},
				},
			},
			Icon:      "play-circle",
			SortOrder: 12,
		},
		{
			Slug:        "auto-archive-old-issues",
			Name:        "长期未更新工作项自动归档",
			Description: "超过 30 天未更新的已完成工作项自动归档",
			Category:    "management",
			DSLTemplate: RuleDSL{
				Trigger: Trigger{Type: "scheduled", Cron: "0 2 * * *"},
				Conditions: []Condition{
					{Field: "state.group", Op: "eq", Value: "completed"},
					{Field: "issue.updated_at", Op: "lt", Value: "now-30d"},
				},
				Actions: []Action{
					{Type: ActionUpdateField, Field: "is_archived", Value: "true"},
				},
			},
			Icon:      "archive",
			SortOrder: 13,
		},
		{
			Slug:        "duplicate-issue-check",
			Name:        "重复工作项提醒",
			Description: "新建工作项时检测可能的重复项并提醒",
			Category:    "management",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "issue.created"},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${issue.created_by}", "template": "⚠️ 检测到可能的重复工作项请确认"}},
				},
			},
			Icon:      "copy",
			SortOrder: 14,
		},
		{
			Slug:        "new-member-welcome",
			Name:        "新成员加入通知",
			Description: "工作空间有新成员加入时通知所有成员",
			Category:    "management",
			DSLTemplate: RuleDSL{
				Trigger:    Trigger{Type: "member.added"},
				Conditions: []Condition{},
				Actions: []Action{
					{Type: ActionNotify, Config: map[string]any{"channel": "in_app", "target": "${workspace.members}", "template": "欢迎 {{actor.user_name}} 加入工作空间"}},
				},
			},
			Icon:      "user-plus",
			SortOrder: 15,
		},
	}
}
