// Package intake — Intake 收件箱：公开表单提交、评审队列、转正流程。
//
// 设计参考：GitHub Issue Forms / Jira Service Management / Linear Inbox。
// 流程：
//   1. 管理员创建 intake_channels（公开表单 → /intake/{slug}）。
//   2. 外部用户免登录提交（限流 20/min/IP + 验证码）。
//   3. intake.submitted 事件触发自动分配规则。
//   4. 管理员在收件箱页审核 → accept（转正需求/缺陷）/ reject / archive。
//   5. 提交者通过 tracking_id 查看处理状态（脱敏视图）。
package intake

import (
	"encoding/json"
	"time"
)

// --- 常量 ---

// IntakeStatus 是 intake_issue 的状态枚举。
type IntakeStatus string

const (
	IntakeStatusOpen     IntakeStatus = "open"
	IntakeStatusAccepted IntakeStatus = "accepted"
	IntakeStatusRejected IntakeStatus = "rejected"
	IntakeStatusArchived IntakeStatus = "archived"
)

// IssueType 转正时选择的工作项类型。
const (
	IssueTypeRequirement = "requirement"
	IssueTypeDefect      = "defect"
	IssueTypeTask        = "task"
)

// --- Channel（公开表单配置）

// Channel 是 intake 提交通道的配置。
type Channel struct {
	ID                 int64           `json:"id"`
	WorkspaceID        int64           `json:"workspace_id"`
	ProjectID          *int64          `json:"project_id,omitempty"`
	Slug               string          `json:"slug"`
	Name               string          `json:"name"`
	Description        string          `json:"description,omitempty"`
	IsPublic           bool            `json:"is_public"`
	DefaultIssueType   string          `json:"default_issue_type"`
	DefaultPriority    int16           `json:"default_priority"`
	AutoAssignRules    json.RawMessage `json:"auto_assign_rules"`
	RateLimitPerMin    int16           `json:"rate_limit_per_min"`
	RequireCaptcha     bool            `json:"require_captcha"`
	CustomFields       json.RawMessage `json:"custom_fields"`
	Branding           json.RawMessage `json:"branding"`
	NotifyOnSubmit     bool            `json:"notify_on_submit"`
	NotifyUsers        []int64         `json:"notify_users"`
	IsActive           bool            `json:"is_active"`
	CreatedBy          int64           `json:"created_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// AutoAssignRule 是单条自动分配规则。
type AutoAssignRule struct {
	Match   AutoMatch `json:"match"`
	AssignTo int64    `json:"assign_to"`
}

// AutoMatch 定义匹配条件。
type AutoMatch struct {
	Keyword   string `json:"keyword,omitempty"`
	IssueType string `json:"issue_type,omitempty"`
	Priority  *int16 `json:"priority,omitempty"`
}

// --- Issue（收件工单）

// Issue 是外部用户提交的收件工单。
type Issue struct {
	ID                int64           `json:"id"`
	ChannelID         int64           `json:"channel_id"`
	WorkspaceID       int64           `json:"workspace_id"`
	ProjectID         *int64          `json:"project_id,omitempty"`
	TrackingID        string          `json:"tracking_id"`
	SubmitterName     string          `json:"submitter_name"`
	SubmitterEmail    string          `json:"submitter_email"`
	SubmitterUserID   *int64          `json:"submitter_user_id,omitempty"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	IssueType         string          `json:"issue_type"`
	Priority          int16           `json:"priority"`
	CustomFields      json.RawMessage `json:"custom_fields,omitempty"`
	AttachmentIDs     []int64         `json:"attachment_ids"`
	Status            string          `json:"status"`
	StatusReason      string          `json:"status_reason,omitempty"`
	ConvertedIssueID  *int64          `json:"converted_issue_id,omitempty"`
	AssignedTo        *int64          `json:"assigned_to,omitempty"`
	ReviewedBy        *int64          `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time      `json:"reviewed_at"`
	NotifyOnStatus    bool            `json:"notify_on_status"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// --- 提交跟踪视图（脱敏）

// PublicView 是提交者查看的处理状态（public tracking portal）。
type PublicView struct {
	TrackingID       string     `json:"tracking_id"`
	Title            string     `json:"title"`
	Status           string     `json:"status"`
	StatusText       string     `json:"status_text"`
	Priority         int16      `json:"priority"`
	IssueType        string     `json:"issue_type"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	StatusReason     string     `json:"status_reason,omitempty"`
	ConvertedIssueID *int64     `json:"converted_issue_id,omitempty"`
}

// statusText 将状态枚举翻译为提交者可读的中文。
func StatusText(status string) string {
	switch status {
	case string(IntakeStatusOpen):
		return "已提交 / 待处理"
	case string(IntakeStatusAccepted):
		return "已接受 / 正在处理"
	case string(IntakeStatusRejected):
		return "已拒绝"
	case string(IntakeStatusArchived):
		return "已归档"
	default:
		return status
	}
}

// IssueTypeText 翻译工作项类型。
func IssueTypeText(t string) string {
	switch t {
	case IssueTypeRequirement:
		return "需求"
	case IssueTypeDefect:
		return "缺陷"
	case IssueTypeTask:
		return "任务"
	default:
		return t
	}
}
