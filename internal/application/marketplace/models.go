// Package marketplace — 应用商店。
//
// 对标 Plane 的 Marketplace：应用/导入器/模板/集成四类目录，
// 支持评分、评论、安装统计与一键安装。
package marketplace

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// Category 应用分类。
type Category string

const (
	// CategoryApps 应用。
	CategoryApps Category = "apps"
	// CategoryImporters 导入器。
	CategoryImporters Category = "importers"
	// CategoryTemplates 模板。
	CategoryTemplates Category = "templates"
	// CategoryIntegrations 集成。
	CategoryIntegrations Category = "integrations"
)

// IsValidCategory 校验分类。
func IsValidCategory(c string) bool {
	switch Category(c) {
	case CategoryApps, CategoryImporters, CategoryTemplates, CategoryIntegrations:
		return true
	}
	return false
}

// Pricing 定价模式。
type Pricing string

const (
	// PricingFree 免费。
	PricingFree Pricing = "free"
	// PricingPaid 付费。
	PricingPaid Pricing = "paid"
	// PricingFreemium 免费增值。
	PricingFreemium Pricing = "freemium"
)

// InstallType 安装方式。
type InstallType string

const (
	// InstallBuiltin 内置（无需配置）。
	InstallBuiltin InstallType = "builtin"
	// InstallOAuth OAuth 授权。
	InstallOAuth InstallType = "oauth"
	// InstallWebhook Webhook 配置。
	InstallWebhook InstallType = "webhook"
	// InstallCLI CLI 工具。
	InstallCLI InstallType = "cli"
)

// Status 发布状态。
type Status string

const (
	// StatusPublished 已发布。
	StatusPublished Status = "published"
	// StatusDraft 草稿。
	StatusDraft Status = "draft"
	// StatusDeprecated 已弃用。
	StatusDeprecated Status = "deprecated"
)

// App 应用目录条目。
type App struct {
	ID            int64          `json:"id"`
	Slug          string         `json:"slug"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Category      Category       `json:"category"`
	Publisher     string         `json:"publisher"`
	PublisherURL  string         `json:"publisher_url,omitempty"`
	IconURL       string         `json:"icon_url,omitempty"`
	Screenshots   []string       `json:"screenshots,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Pricing       Pricing        `json:"pricing"`
	InstallType   InstallType    `json:"install_type"`
	InstallConfig map[string]any `json:"install_config,omitempty"`
	DocsURL       string         `json:"docs_url,omitempty"`
	RatingAvg     float64        `json:"rating_avg"`
	RatingCount   int            `json:"rating_count"`
	InstallCount  int            `json:"install_count"`
	Status        Status         `json:"status"`
	CreatedBy     int64          `json:"created_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Install 安装记录。
type Install struct {
	ID          int64          `json:"id"`
	WorkspaceID int64          `json:"workspace_id"`
	AppID       int64          `json:"app_id"`
	InstalledBy int64          `json:"installed_by"`
	Config      map[string]any `json:"config,omitempty"`
	Enabled     bool           `json:"enabled"`
	InstalledAt time.Time      `json:"installed_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Review 评分与评论。
type Review struct {
	ID          int64     `json:"id"`
	AppID       int64     `json:"app_id"`
	WorkspaceID int64     `json:"workspace_id"`
	UserID      int64     `json:"user_id"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryInfo 分类信息。
type CategoryInfo struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

// CreateAppInput 创建应用入参。
type CreateAppInput struct {
	Slug          string
	Name          string
	Description   string
	Category      Category
	Publisher     string
	PublisherURL  string
	IconURL       string
	Screenshots   []string
	Tags          []string
	Pricing       Pricing
	InstallType   InstallType
	InstallConfig map[string]any
	DocsURL       string
	CreatedBy     int64
}

// UpdateAppInput 更新应用入参。
type UpdateAppInput struct {
	Name          *string
	Description   *string
	PublisherURL  *string
	IconURL       *string
	Screenshots   []string
	Tags          []string
	Pricing       *Pricing
	InstallConfig map[string]any
	DocsURL       *string
	Status        *Status
}

// InstallInput 安装入参。
type InstallInput struct {
	WorkspaceID int64
	AppID       int64
	InstalledBy int64
	Config      map[string]any
}

// ReviewInput 评分入参。
type ReviewInput struct {
	WorkspaceID int64
	AppID       int64
	UserID      int64
	Rating      int
	Comment     string
}

func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
