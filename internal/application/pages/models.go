// Package pages — 项目文档页面领域（对标 Plane Pages）。
//
// 每个项目拥有一个文档页面树：页面支持富文本内容（JSON 描述）、父子嵌套与
// 排序；通过 version 字段实现乐观锁，软删除归档。
package pages

import (
	"encoding/json"
	"time"
)

// Page 项目文档页面模型（API 响应 DTO）。
type Page struct {
	ID                  int64           `json:"id"`
	PublicID            string          `json:"public_id"`
	WorkspaceID         int64           `json:"workspace_id"`
	ProjectID           int64           `json:"project_id"`
	Name                string          `json:"name"`
	DescriptionJSON     json.RawMessage `json:"description_json,omitempty"`
	DescriptionHTML     string          `json:"description_html,omitempty"`
	DescriptionStripped string          `json:"description_stripped,omitempty"`
	ParentID            *int64          `json:"parent_id,omitempty"`
	SortOrder           float64         `json:"sort_order"`
	Category            string          `json:"category,omitempty"`
	CreatedBy           int64           `json:"created_by"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
	Version             int32           `json:"version"`
}

// DocumentVersion 文档版本快照模型。
type DocumentVersion struct {
	ID            int64     `json:"id"`
	PageID        int64     `json:"page_id"`
	VersionNumber int32     `json:"version_number"`
	ContentMD     string    `json:"content_md,omitempty"`
	ContentHTML   string    `json:"content_html,omitempty"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// DocumentLink 文档关联模型（关联 Issue / Sprint / Version 等）。
type DocumentLink struct {
	ID           int64     `json:"id"`
	PageID       int64     `json:"page_id"`
	LinkableType string    `json:"linkable_type"`
	LinkableID   int64     `json:"linkable_id"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// PageTemplate 文档模板模型。
// ProjectID = 0 表示工作空间级模板（跨项目共享）；ProjectID > 0 为项目级模板。
type PageTemplate struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ContentHTML string    `json:"content_html,omitempty"`
	Category    string    `json:"category,omitempty"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTemplateInput 创建模板入参。
type CreateTemplateInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ContentHTML string `json:"content_html"`
	Category    string `json:"category"`
}

// UpdateTemplateInput 更新模板入参（指针字段为 nil 时不更新）。
type UpdateTemplateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ContentHTML *string `json:"content_html"`
	Category    *string `json:"category"`
}

// PageShare 文档公开分享链接模型。
type PageShare struct {
	ID           int64      `json:"id"`
	PageID       int64      `json:"page_id"`
	WorkspaceID  int64      `json:"workspace_id"`
	ProjectID    int64      `json:"project_id"`
	Token        string     `json:"token"`
	IsActive     bool       `json:"is_active"`
	PasswordHash string     `json:"-"` // 不序列化到 JSON
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedBy    int64      `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
}

// CreateShareInput 创建分享链接入参。
type CreateShareInput struct {
	Password  string     `json:"password"`             // 可选访问密码（明文，后端哈希）
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // 可选过期时间
}

// UpdateShareInput 更新分享链接入参。
type UpdateShareInput struct {
	IsActive  *bool      `json:"is_active"`
	Password  *string    `json:"password"` // 明文；nil 表示不修改
	ExpiresAt *time.Time `json:"expires_at"`
}

// PublicSharePageView 公开分享页面的视图模型（对外暴露，不含敏感信息）。
type PublicSharePageView struct {
	PageID          int64           `json:"page_id"`
	WorkspaceID     int64           `json:"workspace_id"`
	ProjectID       int64           `json:"project_id"`
	Name            string          `json:"name"`
	DescriptionHTML string          `json:"description_html,omitempty"`
	DescriptionJSON json.RawMessage `json:"description_json,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
