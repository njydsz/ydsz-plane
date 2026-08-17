// Package intake — 收件箱（匿名提报）域。
//
// 对标 Jira Service Management / TAPD 提报渠道：
//   - IntakeChannel：入口渠道（公开提交链接，如支持台/需求收集箱/缺陷反馈箱）；
//   - IntakeIssue：渠道收到的工单，支持公开免登提交、提交者跟踪、管理员审核、
//     以及「转正」——把通过审核的工单转为正式需求/任务/缺陷。
//
// 数据表：intake_channels / intake_issues（id 由应用层生成，见 genID）。
package intake

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// ChannelStatus 渠道状态（对应 entity_status 枚举子集）。
type ChannelStatus string

const (
	ChannelActive   ChannelStatus = "active"
	ChannelArchived ChannelStatus = "archived"
)

// IssueStatus 工单状态（对应 intake_issue_status 枚举）。
type IssueStatus string

const (
	IssueOpen     IssueStatus = "open"
	IssueAccepted IssueStatus = "accepted"
	IssueRejected IssueStatus = "rejected"
	IssueArchived IssueStatus = "archived"
)

// IntakeChannel 入口渠道。
type IntakeChannel struct {
	ID          int64           `json:"id"`
	Code        string          `json:"code,omitempty"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   *int64          `json:"project_id,omitempty"`
	Description string          `json:"description,omitempty"`
	IsActive    bool            `json:"is_active"`
	Config      map[string]any  `json:"config,omitempty"`
	Status      string          `json:"status"`
	IssueCount  int             `json:"issue_count,omitempty"`
	CreatedBy   int64           `json:"created_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// IntakeIssue 入口工单。
type IntakeIssue struct {
	ID              int64      `json:"id"`
	Code            string     `json:"code,omitempty"`
	Name            string     `json:"name"`
	WorkspaceID     int64      `json:"workspace_id"`
	ProjectID       *int64     `json:"project_id,omitempty"`
	ChannelID       int64      `json:"channel_id"`
	ChannelName     string     `json:"channel_name,omitempty"`
	TrackingID      string     `json:"tracking_id"`
	SubmitterName   string     `json:"submitter_name,omitempty"`
	SubmitterEmail  string     `json:"submitter_email,omitempty"`
	Description     string     `json:"description,omitempty"`
	Priority        string     `json:"priority"`
	Status          IssueStatus `json:"status"`
	LinkedEntityType  string    `json:"linked_entity_type,omitempty"`
	LinkedEntityID    *int64    `json:"linked_entity_id,omitempty"`
	LinkedEntityIdent string    `json:"linked_entity_identifier,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy      *int64     `json:"resolved_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// genID 生成应用层 BIGINT 主键（雪花式：时间戳高位 + 随机低位）。
// intake 两表 id 无数据库默认值（init SQL 注释：应用层生成），
// 使用毫秒时间戳 <<22 | 22bit 随机，避免并发冲突且全局趋势递增。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
