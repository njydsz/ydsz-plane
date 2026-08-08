// Package webhook — Webhook 领域模型、投递与重试逻辑。
//
// 设计参考：GitHub Webhooks / GitLab Hooks / Stripe Event Webhooks。
// 核心契约：
//   - HMAC-SHA256 签名允许接收方验证消息真实性。
//   - 唯一投递 ID 允许接收方幂等处理。
//   - 指数退避重试在目标暂时不可用时最大化最终投递率。
//   - 日志保留 30 天用于对账与调试。
//   - SSRF 防护拒绝内网 / 保留 IP。
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// --- Event Types ---

// 完整事件目录（与 domain_events、RabbitMQ routing key 对齐）。
const (
	// --- Issue 事件 ---
	EventIssueCreated         = "issue.created"
	EventIssueUpdated         = "issue.updated"
	EventIssueDeleted         = "issue.deleted"
	EventIssueStatusChanged   = "issue.status_changed"
	EventIssueCommented       = "issue.commented"
	EventIssueCommentUpdated  = "issue.comment_updated"
	EventIssueCommentDeleted  = "issue.comment_deleted"
	EventIssueAttachmentAdded = "issue.attachment_added"
	EventIssueAttachmentRemoved = "issue.attachment_removed"

	// --- Project 事件 ---
	EventProjectCreated      = "project.created"
	EventProjectUpdated      = "project.updated"
	EventProjectDeleted      = "project.deleted"
	EventProjectMemberAdded  = "project.member_added"
	EventProjectMemberRemoved = "project.member_removed"

	// --- State / Module / Label 事件 ---
	EventStateCreated = "state.created"
	EventStateUpdated = "state.updated"
	EventStateDeleted = "state.deleted"
	EventModuleCreated = "module.created"
	EventModuleUpdated = "module.updated"
	EventModuleDeleted = "module.deleted"
	EventLabelCreated = "label.created"
	EventLabelUpdated = "label.updated"
	EventLabelDeleted = "label.deleted"

	// --- Sprint 事件 ---
	EventSprintCreated     = "sprint.created"
	EventSprintStarted     = "sprint.started"
	EventSprintCompleted   = "sprint.completed"
	EventSprintDeleted     = "sprint.deleted"
	EventSprintIssueAdded  = "sprint.issue_added"
	EventSprintIssueRemoved = "sprint.issue_removed"

	// --- Attachment 事件（Issue 维度已在上方声明） ---

	// --- User / Member 事件 ---
	EventMemberAdded       = "member.added"
	EventMemberRemoved     = "member.removed"
	EventMemberRoleChanged = "member.role_changed"
	EventUserInvited       = "user.invited"
	EventUserRemoved       = "user.removed"
	EventUserRoleUpdated   = "user.role_updated"

	// --- Version 事件 ---
	EventVersionCreated  = "version.created"
	EventVersionUpdated  = "version.updated"
	EventVersionDeleted  = "version.deleted"
	EventVersionReleased = "version.released"

	// --- Intake 事件 ---
	EventIntakeSubmitted = "intake.submitted"
	EventIntakeConverted = "intake.converted"
	EventIntakeCreated   = "intake.created"
	EventIntakeDeleted   = "intake.deleted"
	EventIntakeMerged    = "intake.merged"
)

// AllEvents 返回完整事件目录（用于前端多选框或全订阅）。
func AllEvents() []string {
	return []string{
		// Issue
		EventIssueCreated, EventIssueUpdated, EventIssueDeleted,
		EventIssueStatusChanged, EventIssueCommented,
		EventIssueCommentUpdated, EventIssueCommentDeleted,
		EventIssueAttachmentAdded, EventIssueAttachmentRemoved,
		// Project
		EventProjectCreated, EventProjectUpdated, EventProjectDeleted,
		EventProjectMemberAdded, EventProjectMemberRemoved,
		// State / Module / Label
		EventStateCreated, EventStateUpdated, EventStateDeleted,
		EventModuleCreated, EventModuleUpdated, EventModuleDeleted,
		EventLabelCreated, EventLabelUpdated, EventLabelDeleted,
		// Sprint
		EventSprintCreated, EventSprintStarted, EventSprintCompleted, EventSprintDeleted,
		EventSprintIssueAdded, EventSprintIssueRemoved,
		// User / Member
		EventMemberAdded, EventMemberRemoved, EventMemberRoleChanged,
		EventUserInvited, EventUserRemoved, EventUserRoleUpdated,
		// Version
		EventVersionCreated, EventVersionUpdated, EventVersionDeleted, EventVersionReleased,
		// Intake
		EventIntakeSubmitted, EventIntakeConverted,
		EventIntakeCreated, EventIntakeDeleted, EventIntakeMerged,
	}
}

// --- Domain Models ---

// Webhook 是工作空间或项目级的 Webhook 订阅配置。
type Webhook struct {
	ID            int64      `json:"id"`
	WorkspaceID   int64      `json:"workspace_id"`
	ProjectID     *int64     `json:"project_id,omitempty"`
	Name          string     `json:"name"`
	TargetURL     string     `json:"target_url"`
	Secret        string     `json:"-"` // 不在 API 响应中暴露；仅在创建时返回一次
	Events        []string   `json:"events"`
	IsActive      bool       `json:"is_active"`
	LastError     string     `json:"last_error,omitempty"`
	LastTriggered *time.Time `json:"last_triggered,omitempty"`
	LastStatus    string     `json:"last_status,omitempty"`
	UnhealthyAt   *time.Time `json:"unhealthy_at,omitempty"`
	CreatedBy     int64      `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// WebhookLog 是单次 Webhook 投递的记录（含请求/响应详情）。
type WebhookLog struct {
	ID              int64           `json:"id"`
	WebhookID       int64           `json:"webhook_id"`
	WorkspaceID     int64           `json:"workspace_id"`
	DeliveryID      string          `json:"delivery_id"`
	EventType       string          `json:"event_type"`
	EventID         *int64          `json:"event_id,omitempty"`
	RequestURL      string          `json:"request_url"`
	RequestMethod   string          `json:"request_method"`
	RequestHeaders  json.RawMessage `json:"request_headers,omitempty"`
	RequestBody     string          `json:"request_body,omitempty"`
	ResponseStatus  *int            `json:"response_status,omitempty"`
	ResponseBody    string          `json:"response_body,omitempty"`
	ResponseHeaders json.RawMessage `json:"response_headers,omitempty"`
	Status          string          `json:"status"`
	Attempt         int16           `json:"attempt"`
	DurationMs      *int            `json:"duration_ms,omitempty"`
	Error           string          `json:"error,omitempty"`
	OccurredAt      time.Time       `json:"occurred_at"`
}

// LogStatus 是投递日志的状态枚举。
const (
	LogStatusDelivered = "delivered"
	LogStatusFailed    = "failed"
	LogStatusRetrying  = "retrying"
)

// WebhookStatus 是聚合体整体健康状态。
const (
	WebhookStatusSuccess   = "success"
	WebhookStatusFailed    = "failed"
	WebhookStatusUnhealthy = "unhealthy"
)

// --- 投递负载 ---

// DeliveryPayload 是发往目标 URL 的 JSON body。
type DeliveryPayload struct {
	Event      string          `json:"event"`
	Workspace  string          `json:"workspace"` // slug
	Project    string          `json:"project,omitempty"`
	Data       json.RawMessage `json:"data"`
	Actor      *ActorInfo      `json:"actor,omitempty"`
	ActorName  string          `json:"actor_name,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}

// ActorInfo 描述触发事件的执行者。
type ActorInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// --- HMAC 签名 ---

// ComputeSignature 计算 payload 的 HMAC-SHA256 签名。
// 签名输入 = timestamp + "." + body（防重放）。
func ComputeSignature(secret string, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature 验证接收方预期的签名。
func VerifySignature(secret string, timestamp int64, body []byte, expectedMAC string) bool {
	actual := ComputeSignature(secret, timestamp, body)
	return hmac.Equal([]byte(actual), []byte(expectedMAC))
}

// SignatureHeader 构造 X-Ydsz-Signature-256 header 值。
func SignatureHeader(secret string, timestamp int64, body []byte) string {
	return "sha256=" + ComputeSignature(secret, timestamp, body)
}
