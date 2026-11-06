// Package milestone — 项目里程碑（Milestones）领域。
//
// 对标 Jira/Plane 的「里程碑」能力：把跨迭代的关键交付节点聚合为里程碑，
// 通过关联工作项（跨 task/requirement/defect 类型）计算整体进度。
//
// 数据表：
//   - milestones             （id 由应用层 genID 生成）
//   - milestone_work_items   （milestone_id + work_item_id 联合主键）
package milestone

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// MilestoneStatus 里程碑状态。
type MilestoneStatus string

const (
	// StatusNotStarted 未开始。
	StatusNotStarted MilestoneStatus = "not_started"
	// StatusInProgress 进行中。
	StatusInProgress MilestoneStatus = "in_progress"
	// StatusCompleted 已完成。
	StatusCompleted MilestoneStatus = "completed"
	// StatusMissed 已逾期未完成。
	StatusMissed MilestoneStatus = "missed"
)

// IsValid 校验状态是否合法。
func (s MilestoneStatus) IsValid() bool {
	switch s {
	case StatusNotStarted, StatusInProgress, StatusCompleted, StatusMissed:
		return true
	default:
		return false
	}
}

// Milestone 里程碑实体。
type Milestone struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   int64           `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	TargetDate  *time.Time      `json:"target_date,omitempty"`
	Status      MilestoneStatus `json:"status"`
	CreatedBy   int64           `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MilestoneWorkItem 里程碑关联的工作项（跨类型）。
type MilestoneWorkItem struct {
	MilestoneID  int64     `json:"milestone_id"`
	WorkItemID   int64     `json:"work_item_id"`
	WorkItemType string    `json:"work_item_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// MilestoneProgress 里程碑进度聚合结果。
type MilestoneProgress struct {
	Total        int     `json:"total"`         // 关联工作项总数
	Completed    int     `json:"completed"`     // 已完成（状态 group=done）数量
	InProgress   int     `json:"in_progress"`   // 进行中（group=in_progress）数量
	NotStarted   int     `json:"not_started"`   // 未开始（group=todo/backlog）数量
	Percent      float64 `json:"percent"`       // 完成百分比（0-100，保留 2 位）
	ComputedAt   time.Time `json:"computed_at"`
}

// --- Input Types ---

// CreateInput 创建里程碑入参。
type CreateInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Description string
	TargetDate  *time.Time
	Status      MilestoneStatus
	CreatedBy   int64
}

// UpdateInput 更新里程碑入参（指针字段为 nil 时不更新）。
type UpdateInput struct {
	Name        *string
	Description *string
	TargetDate  *time.Time
	Status      *MilestoneStatus
}

// ListOptions 列表查询选项。
type ListOptions struct {
	WorkspaceID int64
	ProjectID   int64
	Status      MilestoneStatus
	Limit       int
	Offset      int
}

// LinkInput 关联工作项入参。
type LinkInput struct {
	MilestoneID  int64
	WorkItemID   int64
	WorkItemType string
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
