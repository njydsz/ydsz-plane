// Package notification — 通知领域模型与服务。
//
// 参考: Linear/GitLab Notification system。
package notification

import (
	"encoding/json"
	"time"
)

// --- Event Types ---

// EventType 通知事件类型。
type EventType string

const (
	EventIssueCreated      EventType = "issue.created"
	EventIssueAssigned     EventType = "issue.assigned"
	EventIssueStatusChanged EventType = "issue.status_changed"
	EventIssueDeleted      EventType = "issue.deleted"
	EventCommentCreated    EventType = "comment.created"
	EventSprintStarted     EventType = "sprint.started"
	EventSprintCompleted   EventType = "sprint.completed"
	EventVersionReleased   EventType = "version.released"
	EventMemberAdded       EventType = "member.added"
	EventMemberRemoved     EventType = "member.removed"
	EventMemberRoleChanged EventType = "member.role_changed"
	EventInvitationSent    EventType = "invitation.sent"
)

// EntityType 通知关联的对象类型。
type EntityType string

const (
	EntityIssue     EntityType = "issue"
	EntitySprint    EntityType = "sprint"
	EntityVersion   EntityType = "version"
	EntityProject   EntityType = "project"
	EntityWorkspace EntityType = "workspace"
	EntityComment   EntityType = "comment"
	EntityMember    EntityType = "member"
)

// Channel 通知渠道。
type Channel string

const (
	ChannelInApp   Channel = "in_app"
	ChannelEmail   Channel = "email"
	ChannelWeCom   Channel = "wecom"
	ChannelDingTalk Channel = "dingtalk"
	ChannelFeishu  Channel = "feishu"
)

// Digest 摘要频率。
type Digest string

const (
	DigestRealtime Digest = "realtime"
	DigestDaily    Digest = "daily"
	DigestWeekly   Digest = "weekly"
	DigestOff      Digest = "off"
)

// --- Domain Models ---

// Notification 通知聚合根。
type Notification struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	RecipientID int64      `json:"recipient_id"`
	EventType   EventType  `json:"event_type"`
	EntityType  EntityType `json:"entity_type"`
	EntityID    int64      `json:"entity_id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	ActionURL   string     `json:"action_url"`
	ActorID     *int64     `json:"actor_id"`
	ActorName   string     `json:"actor_name"`
	IsRead      bool       `json:"is_read"`
	IsArchived  bool       `json:"is_archived"`
	ReadAt      *time.Time `json:"read_at"`
	Channel     Channel    `json:"channel"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NotificationPreference 用户通知偏好。
type NotificationPreference struct {
	ID          int64           `json:"id"`
	UserID      int64           `json:"user_id"`
	WorkspaceID int64           `json:"workspace_id"`
	EventTypes  []string        `json:"event_types"`
	Channels    []string        `json:"channels"`
	Digest      Digest          `json:"digest"`
	DNDEnabled  bool            `json:"dnd_enabled"`
	DNDStart    string          `json:"dnd_start"`
	DNDEnd      string          `json:"dnd_end"`
	IsEnabled   bool            `json:"is_enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateNotificationInput 创建通知的入参。
type CreateNotificationInput struct {
	WorkspaceID int64
	RecipientID int64
	EventType   EventType
	EntityType  EntityType
	EntityID    int64
	Title       string
	Body        string
	ActionURL   string
	ActorID     *int64
	ActorName   string
	Channel     Channel
	Payload     json.RawMessage
}

// --- Event Title Templates ---

// EventTitles 为各事件类型提供中文标题模板。
var EventTitles = map[EventType]string{
	EventIssueCreated:       "创建了工作项",
	EventIssueAssigned:      "将工作项分配给你",
	EventIssueStatusChanged: "变更了工作项状态",
	EventIssueDeleted:       "删除了工作项",
	EventCommentCreated:     "评论了工作项",
	EventSprintStarted:      "启动了迭代",
	EventSprintCompleted:    "完成了迭代",
	EventVersionReleased:    "发布了版本",
	EventMemberAdded:        "加入了工作空间",
	EventMemberRemoved:      "移出了成员",
	EventMemberRoleChanged:  "变更了成员角色",
	EventInvitationSent:     "发送了邀请",
}
