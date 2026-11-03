// Package initiative — 倡议（Initiative）领域模型与枚举。
//
// 对标 Plane / Jira Advanced Roadmaps 的 Initiative 层级：
// 一个倡议是跨项目的战略性目标，可关联多个项目与工作项，
// 并通过聚合这些项目/迭代/版本/工作项的进度来向上汇报整体健康度。
//
// 数据表：
//   - initiatives（主表，id 由应用层 genID 生成）
//   - initiative_projects（倡议 ↔ 项目 多对多）
//   - initiative_work_items（倡议 ↔ 工作项 多对多，工作项按类型区分）
package initiative

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// InitiativeStatus 倡议状态枚举。
type InitiativeStatus string

const (
	// StatusOnTrack 进展正常。
	StatusOnTrack InitiativeStatus = "on_track"
	// StatusAtRisk 存在风险。
	StatusAtRisk InitiativeStatus = "at_risk"
	// StatusOffTrack 偏离轨道。
	StatusOffTrack InitiativeStatus = "off_track"
	// StatusCompleted 已完成。
	StatusCompleted InitiativeStatus = "completed"
)

// IsValid 校验状态值是否合法。
func (s InitiativeStatus) IsValid() bool {
	switch s {
	case StatusOnTrack, StatusAtRisk, StatusOffTrack, StatusCompleted:
		return true
	default:
		return false
	}
}

// InitiativePriority 倡议优先级枚举。
type InitiativePriority string

const (
	// PriorityUrgent 紧急。
	PriorityUrgent InitiativePriority = "urgent"
	// PriorityHigh 高。
	PriorityHigh InitiativePriority = "high"
	// PriorityMedium 中。
	PriorityMedium InitiativePriority = "medium"
	// PriorityLow 低。
	PriorityLow InitiativePriority = "low"
	// PriorityNone 无优先级。
	PriorityNone InitiativePriority = "none"
)

// IsValid 校验优先级值是否合法。
func (p InitiativePriority) IsValid() bool {
	switch p {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		return true
	default:
		return false
	}
}

// 工作项类型常量（与 initiative_work_items.work_item_type 一致）。
const (
	WorkItemTypeRequirement = "requirement"
	WorkItemTypeTask        = "task"
	WorkItemTypeDefect      = "defect"
)

// Initiative 倡议聚合根。
//
// 业务规则：
//   - 一个倡议属于一个工作空间（workspace_id）；
//   - 可通过 initiative_projects 关联多个项目；
//   - 可通过 initiative_work_items 关联多个工作项（需求/任务/缺陷）。
type Initiative struct {
	ID          int64              `json:"id"`
	WorkspaceID int64              `json:"workspace_id"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Status      InitiativeStatus   `json:"status"`
	Priority    InitiativePriority `json:"priority"`
	StartDate   *time.Time         `json:"start_date,omitempty"`
	TargetDate  *time.Time         `json:"target_date,omitempty"`
	OwnerID     *int64             `json:"owner_id,omitempty"`
	CreatedBy   int64              `json:"created_by"`
	ArchivedAt  *time.Time         `json:"archived_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// InitiativeProject 倡议 ↔ 项目 关联（多对多）。
type InitiativeProject struct {
	InitiativeID int64     `json:"initiative_id"`
	ProjectID    int64     `json:"project_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// InitiativeWorkItem 倡议 ↔ 工作项 关联（多对多，带工作项类型）。
type InitiativeWorkItem struct {
	InitiativeID int64     `json:"initiative_id"`
	WorkItemID   int64     `json:"work_item_id"`
	WorkItemType string    `json:"work_item_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// InitiativeProgress 倡议进度聚合（由关联项目/迭代/版本/工作项聚合而来）。
type InitiativeProgress struct {
	// 关联项目数
	ProjectCount int `json:"project_count"`
	// 关联工作项总数 / 已完成数
	TotalWorkItems int `json:"total_work_items"`
	DoneWorkItems  int `json:"done_work_items"`
	// 按工作项类型分布
	ByType map[string]InitiativeTypeProgress `json:"by_type,omitempty"`
	// 按关联项目分布
	ByProject []InitiativeProjectProgress `json:"by_project,omitempty"`
	// 整体完成百分比（0-100）
	CompletionRate float64 `json:"completion_rate"`
}

// InitiativeTypeProgress 某一工作项类型的进度。
type InitiativeTypeProgress struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

// InitiativeProjectProgress 某一关联项目的进度。
type InitiativeProjectProgress struct {
	ProjectID      int64   `json:"project_id"`
	ProjectName    string  `json:"project_name"`
	TotalItems     int     `json:"total_items"`
	DoneItems      int     `json:"done_items"`
	CompletionRate float64 `json:"completion_rate"`
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与 calendar / intake / datajob 包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
// 同一毫秒内可生成 4,194,304 个不重复 ID，时间戳保证跨毫秒单调递增。
//
// 生产环境应由数据库序列或分布式 ID 服务提供，此处保留为无外部依赖的
// 轻量实现，保证应用层在测试与本地环境可用。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
