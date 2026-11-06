// Package projectstate — 项目状态与更新域。
//
// 对标 Plane Project States + Updates：
//   - 项目可拥有状态（on_track / at_risk / off_track / completed）、优先级、起止日期；
//   - 项目可发布「更新」（状态快照 + 文本 + 线程评论 + 表情回应）；
//   - 更新可向上汇入倡议（Initiative）的状态指标。
//
// 数据表：
//   - project_states
//   - project_updates
//   - project_update_reactions
//   - project_update_comments
package projectstate

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// ProjectState 项目状态记录。
type ProjectState struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	Status      string     `json:"status"` // on_track / at_risk / off_track / completed / paused
	Priority    string     `json:"priority"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	OwnerID     int64      `json:"owner_id"`
	Description string     `json:"description"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProjectUpdate 项目更新（状态快照 + 文本）。
type ProjectUpdate struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	AuthorID    int64     `json:"author_id"`
	Status      string    `json:"status"`
	Content     string    `json:"content"`
	// 发布时的进度快照（百分比）
	ProgressPercent float64 `json:"progress_percent"`
	// 可选关联倡议
	InitiativeID    int64     `json:"initiative_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// ProjectUpdateReaction 更新表情回应。
type ProjectUpdateReaction struct {
	ID        int64     `json:"id"`
	UpdateID  int64     `json:"update_id"`
	UserID    int64     `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectUpdateComment 更新评论。
type ProjectUpdateComment struct {
	ID        int64      `json:"id"`
	UpdateID  int64      `json:"update_id"`
	AuthorID  int64      `json:"author_id"`
	ParentID  *int64     `json:"parent_id,omitempty"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// --- 枚举 ---

// Status 项目/更新状态。
type Status string

const (
	// StatusOnTrack 正常推进。
	StatusOnTrack Status = "on_track"
	// StatusAtRisk 存在风险。
	StatusAtRisk Status = "at_risk"
	// StatusOffTrack 偏离轨道。
	StatusOffTrack Status = "off_track"
	// StatusCompleted 已完成。
	StatusCompleted Status = "completed"
	// StatusPaused 已暂停。
	StatusPaused Status = "paused"
)

// IsValid 校验状态值。
func (s Status) IsValid() bool {
	switch s {
	case StatusOnTrack, StatusAtRisk, StatusOffTrack, StatusCompleted, StatusPaused:
		return true
	}
	return false
}

// Priority 优先级。
type Priority string

const (
	PriorityUrgent Priority = "urgent"
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
	PriorityNone   Priority = "none"
)

// IsValidPriority 校验优先级值。
func IsValidPriority(p string) bool {
	switch Priority(p) {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		return true
	}
	return false
}

// --- Input Types ---

// UpsertStateInput 创建/更新项目状态入参。
type UpsertStateInput struct {
	ProjectID   int64
	Status      string
	Priority    string
	StartDate   *time.Time
	TargetDate  *time.Time
	OwnerID     int64
	Description string
}

// CreateUpdateInput 创建项目更新入参。
type CreateUpdateInput struct {
	WorkspaceID     int64
	ProjectID       int64
	AuthorID        int64
	Status          string
	Content         string
	ProgressPercent float64
	InitiativeID    int64
}

// ListUpdatesOptions 更新列表查询选项。
type ListUpdatesOptions struct {
	ProjectID    int64
	InitiativeID int64
	Limit        int
	Offset       int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
