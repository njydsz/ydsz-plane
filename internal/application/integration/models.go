// Package integration — 集成（Marketplace + 第三方连接器）领域。
//
// 对标 Plane Integrations / Linear Integrations：
//   - Integration：工作空间级集成实例（GitHub / GitLab / Slack / Sentry / Jira 等），
//     凭证（Credentials）以 AES-256-GCM 加密后存储，绝不下发前端；
//   - IntegrationEvent：集成事件日志（入站 / 出站），用于同步记录与调试；
//   - IntegrationConnector：连接器接口，每个 provider 一个实现，
//     负责凭证校验、连通性测试与工作项同步。
//
// 数据表：integrations / integration_events（id 由应用层 genID 生成）。
package integration

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"time"
)

// --- 枚举 ---

// Provider 标识第三方服务类型。
type Provider string

const (
	// ProviderGitHub GitHub 代码托管。
	ProviderGitHub Provider = "github"
	// ProviderGitLab GitLab 代码托管。
	ProviderGitLab Provider = "gitlab"
	// ProviderSlack Slack 消息通知。
	ProviderSlack Provider = "slack"
	// ProviderSentry Sentry 错误监控。
	ProviderSentry Provider = "sentry"
	// ProviderJira Jira 项目管理。
	ProviderJira Provider = "jira"
)

// IsValid 校验 provider 是否受支持。
func (p Provider) IsValid() bool {
	switch p {
	case ProviderGitHub, ProviderGitLab, ProviderSlack, ProviderSentry, ProviderJira:
		return true
	default:
		return false
	}
}

// Status 标识集成实例的生命周期状态。
type Status string

const (
	// StatusPending 待激活（已创建但凭证尚未通过测试）。
	StatusPending Status = "pending"
	// StatusActive 已激活（凭证有效，同步正常）。
	StatusActive Status = "active"
	// StatusPaused 已暂停（不再触发同步，但凭证保留）。
	StatusPaused Status = "paused"
	// StatusError 错误（最近一次同步 / 测试失败）。
	StatusError Status = "error"
)

// Direction 标识集成事件的方向。
type Direction string

const (
	// DirectionInbound 入站（第三方 → 本系统）。
	DirectionInbound Direction = "inbound"
	// DirectionOutbound 出站（本系统 → 第三方）。
	DirectionOutbound Direction = "outbound"
)

// EventStatus 标识集成事件的处理状态。
type EventStatus string

const (
	// EventSuccess 成功。
	EventSuccess EventStatus = "success"
	// EventFailed 失败。
	EventFailed EventStatus = "failed"
	// EventIgnored 忽略（不匹配过滤条件）。
	EventIgnored EventStatus = "ignored"
)

// --- 领域模型 ---

// Integration 集成实例（工作空间级）。
//
// Credentials 为加密后的 JSON 字符串（AES-256-GCM），字段因 provider 而异，
// 永不在 API 响应中暴露（json:"-"）。Config 为明文 JSON（repo / channel / project 等）。
type Integration struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	Provider    string `json:"provider"` // github / gitlab / slack / sentry / jira
	Name        string `json:"name"`
	Status      string `json:"status"` // active / paused / error / pending
	// Credentials 加密存储的凭证（JSON，字段因 provider 而异）。
	Credentials string `json:"-"`
	// Config 配置（repo、channel、project 等），明文 JSON。
	Config     string     `json:"config,omitempty"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	LastError  string     `json:"last_error,omitempty"`
	CreatedBy  int64      `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// IntegrationEvent 集成事件日志（用于同步记录与调试）。
type IntegrationEvent struct {
	ID            int64     `json:"id"`
	IntegrationID int64     `json:"integration_id"`
	Direction     string    `json:"direction"` // inbound / outbound
	EventType     string    `json:"event_type"`
	Status        string    `json:"status"` // success / failed / ignored
	Payload       string    `json:"payload,omitempty"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// SyncResult 是一次同步的结果摘要。
type SyncResult struct {
	// Provider 为产生该结果的 provider。
	Provider string `json:"provider"`
	// ItemsFetched 为从第三方拉取的工作项数量。
	ItemsFetched int `json:"items_fetched"`
	// ItemsImported 为成功映射 / 导入本系统的工作项数量。
	ItemsImported int `json:"items_imported"`
	// ItemsSkipped 为因重复或过滤被跳过的数量。
	ItemsSkipped int `json:"items_skipped"`
	// Message 为人类可读的摘要（如 "synced 12 issues from owner/repo"）。
	Message string `json:"message,omitempty"`
	// FetchedAt 为同步发生时间。
	FetchedAt time.Time `json:"fetched_at"`
}

// ExternalWorkItem 是第三方工作项的归一化表示（GitHub issue / GitLab issue 等）。
type ExternalWorkItem struct {
	// ExternalID 为第三方侧的唯一标识（如 GitHub issue number 的字符串形式）。
	ExternalID string `json:"external_id"`
	// Title 为标题。
	Title string `json:"title"`
	// Description 为描述 / 正文。
	Description string `json:"description"`
	// State 为状态（open / closed / ...）。
	State string `json:"state"`
	// URL 为第三方侧的可访问链接。
	URL string `json:"url"`
	// Author 为作者标识（登录名 / 邮箱）。
	Author string `json:"author,omitempty"`
	// Labels 为标签列表。
	Labels []string `json:"labels,omitempty"`
	// Assignees 为指派人列表。
	Assignees []string `json:"assignees,omitempty"`
	// CreatedAt 为第三方侧创建时间。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 为第三方侧更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

// IntegrationConnector 连接器接口（每个 provider 一个实现）。
//
// 契约：
//   - Provider 返回 provider 标识；
//   - ValidateCredentials 做纯本地校验（字段完整性、格式），不发起网络请求；
//   - TestConnection 发起一次真实网络请求验证凭证有效性；
//   - SyncWorkItems 拉取第三方工作项并归一化为 ExternalWorkItem。
type IntegrationConnector interface {
	Provider() string
	ValidateCredentials(creds map[string]string) error
	TestConnection(creds map[string]string) error
	SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error)
}

// --- Input Types ---

// CreateIntegrationInput 创建（连接）集成的入参。
type CreateIntegrationInput struct {
	WorkspaceID int64
	Provider    string
	Name        string
	// Credentials 明文凭证（JSON object），服务端加密后存储。
	Credentials map[string]string
	// Config 明文配置（JSON object），如 repo / channel / project。
	Config    map[string]string
	CreatedBy int64
}

// UpdateIntegrationInput 更新集成的入参（字段为指针，nil 表示不更新）。
type UpdateIntegrationInput struct {
	Name        *string
	Credentials map[string]string
	Config      map[string]string
}

// ListIntegrationsOptions 集成列表查询选项。
type ListIntegrationsOptions struct {
	WorkspaceID int64
	// Provider 过滤（空 = 全部）。
	Provider string
	// Status 过滤（空 = 全部）。
	Status string
	Limit  int
	Offset int
}

// ListEventsOptions 事件日志查询选项。
type ListEventsOptions struct {
	WorkspaceID   int64
	IntegrationID *int64
	Direction     string
	Status        string
	EventType     string
	Limit         int
	Offset        int
}

// ProviderMeta 描述一个可用 provider 的元数据（供 GET /providers 返回）。
type ProviderMeta struct {
	// Provider 为 provider 标识。
	Provider string `json:"provider"`
	// Name 为展示名称。
	Name string `json:"name"`
	// Icon 为图标标识（前端据此渲染）。
	Icon string `json:"icon"`
	// Description 为功能简介。
	Description string `json:"description"`
	// Fields 为该 provider 所需的凭证字段（key → label）。
	Fields []ProviderField `json:"fields"`
	// DocsURL 为文档链接。
	DocsURL string `json:"docs_url"`
	// SupportsSync 标识是否支持 SyncWorkItems。
	SupportsSync bool `json:"supports_sync"`
	// SupportsNotify 标识是否支持出站通知。
	SupportsNotify bool `json:"supports_notify"`
}

// ProviderField 描述 provider 的一个凭证字段。
type ProviderField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"` // true = 密码 / Token 类，前端以密文输入
	Placeholder string `json:"placeholder,omitempty"`
}

// --- ID 生成 ---

// genID 生成应用层 BIGINT 主键（与其它应用层包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
