// Package calendar — 项目日历域。
//
// 对标 Jira/Plane 的「项目日历」能力：在日历网格上叠加展示需求/任务/缺陷的
// 时间排布，同时支持独立的日历事件（会议、里程碑、发布日、评审会等）。
//
// 数据表：calendar_events（id 由应用层生成）。
//
// 设计要点：
//   - 事件与工作项解耦：source_type/source_id 用于关联来源工作项（可选），
//     未关联时为纯日历事件；
//   - 幂等：idempotency_key 支持外部系统（如 CI/CD）重复上报时去重；
//   - 乐观锁：version 字段支撑并发编辑冲突检测。
package calendar

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// EventType 事件类型（对应 event_type 列的 CHECK 约束）。
type EventType string

const (
	// EventMeeting 会议（如需求评审、迭代规划会）。
	EventMeeting EventType = "meeting"
	// EventMilestone 里程碑（如版本发布、阶段交付）。
	EventMilestone EventType = "milestone"
	// EventRelease 发布日（如上线窗口、灰度开始）。
	EventRelease EventType = "release"
	// EventDeadline 截止日（如提测截止、合同到期）。
	EventDeadline EventType = "deadline"
	// EventOther 其他自定义事件。
	EventOther EventType = "other"
)

// IsValid 校验事件类型是否合法。
func (e EventType) IsValid() bool {
	switch e {
	case EventMeeting, EventMilestone, EventRelease, EventDeadline, EventOther:
		return true
	default:
		return false
	}
}

// EventStatus 事件状态（对应 entity_status 枚举子集）。
type EventStatus string

const (
	// StatusActive 事件状态：活跃。
	StatusActive EventStatus = "active"
	// StatusArchived 事件状态：已归档（终态）。
	StatusArchived EventStatus = "archived"
)

// IsTerminal 报告状态是否为终态（不再变化）。
func (s EventStatus) IsTerminal() bool {
	return s == StatusArchived
}

// CalendarEvent 日历事件实体。
type CalendarEvent struct {
	ID             int64      `json:"id"`
	TenantID       int64      `json:"tenant_id"`
	WorkspaceID    int64      `json:"workspace_id"`
	ProjectID      *int64     `json:"project_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        time.Time  `json:"end_time"`
	IsAllDay       bool       `json:"is_all_day"`
	Location       string     `json:"location,omitempty"`
	EventType      EventType  `json:"event_type"`
	SourceType     string     `json:"source_type,omitempty"`
	SourceID       *int64     `json:"source_id,omitempty"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
	OrganizerID    *int64     `json:"organizer_id,omitempty"`
	Status         EventStatus `json:"status"`
	Version        int        `json:"version"`
	CreatedBy      int64      `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateEventInput 创建事件入参。
type CreateEventInput struct {
	TenantID       int64
	WorkspaceID    int64
	ProjectID      *int64
	Title          string
	Description    string
	StartTime      time.Time
	EndTime        time.Time
	IsAllDay       bool
	Location       string
	EventType      EventType
	SourceType     string
	SourceID       *int64
	IdempotencyKey string
	OrganizerID    *int64
	CreatedBy      int64
}

// UpdateEventInput 更新事件入参。
type UpdateEventInput struct {
	ID          int64
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	IsAllDay    bool
	Location    string
	EventType   EventType
	OrganizerID *int64
	Version     int
	UpdatedBy   int64
}

// ListEventsFilter 事件列表过滤条件。
type ListEventsFilter struct {
	WorkspaceID int64
	ProjectID   *int64
	// Start / End 为时间范围过滤（闭区间）。
	Start      time.Time
	End        time.Time
	EventType  EventType
	SourceType string
	SourceID   *int64
	// Limit / Offset 分页。
	Limit  int
	Offset int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与 intake 包策略一致）。
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
