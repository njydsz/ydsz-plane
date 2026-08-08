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
