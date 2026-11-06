// Package recurring — 重复工作项（Recurring Work Items）领域。
//
// 对标 Plane / Jira 的「循环任务」能力：用户定义一条规则（cron 表达式 +
// 标题/描述模板 + 指派/标签/状态），系统按 cron 周期自动生成工作项。
//
// 数据表：recurring_rules（id 由应用层 genID 生成）。
//
// 设计要点：
//   - 调度器（EvaluateDue）以 next_run_at <= now 为触发条件，避免依赖进程内
//     cron 轮询即可在 worker / API 任意进程中安全执行；
//   - next_run_at 推进使用 robfig/cron 解析下一次触发时间，支持 5 段式表达式；
//   - 生成的工作项通过 issue.Service 创建，复用既有校验 / 状态机 / 通知链路。
package recurring

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// RecurringRule 重复工作项规则。
type RecurringRule struct {
	ID                  int64      `json:"id"`
	WorkspaceID         int64      `json:"workspace_id"`
	ProjectID           int64      `json:"project_id"`
	WorkItemType        string     `json:"work_item_type"` // task | requirement | defect
	TitleTemplate       string     `json:"title_template"`
	DescriptionTemplate string     `json:"description_template,omitempty"`
	CronExpr            string     `json:"cron_expr"`
	NextRunAt           time.Time  `json:"next_run_at"`
	LastRunAt           *time.Time `json:"last_run_at,omitempty"`
	AssigneeIDs         []int64    `json:"assignee_ids,omitempty"`
	LabelIDs            []int64    `json:"label_ids,omitempty"`
	StateID             int64      `json:"state_id,omitempty"`
	Enabled             bool       `json:"enabled"`
	CreatedBy           int64      `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// --- Input Types ---

// CreateRuleInput 创建规则入参。
type CreateRuleInput struct {
	WorkspaceID         int64
	ProjectID           int64
	WorkItemType        string
	TitleTemplate       string
	DescriptionTemplate string
	CronExpr            string
	AssigneeIDs         []int64
	LabelIDs            []int64
	StateID             int64
	Enabled             bool
	CreatedBy           int64
}

// UpdateRuleInput 更新规则入参（指针字段为 nil 时不更新）。
type UpdateRuleInput struct {
	TitleTemplate       *string
	DescriptionTemplate *string
	CronExpr            *string
	AssigneeIDs         []int64
	LabelIDs            []int64
	StateID             *int64
	Enabled             *bool
}

// ListRulesOptions 列表查询选项。
type ListRulesOptions struct {
	WorkspaceID int64
	ProjectID   int64
	Enabled     *bool
	Limit       int
	Offset      int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
