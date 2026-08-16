// Package knowledge — 知识库领域（Knowledge Base：空间 / 文档 / 版本快照 / 文档关联）。
//
// 对标 Plane / Notion / Confluence 的知识库模型；支持无限层级文档树、乐观锁版本控制、
// 自动版本快照、文档与工作项的双向关联。
package knowledge

import "time"

// --- 枚举类型 ---

// SpacePermission 知识库空间默认权限枚举。
type SpacePermission string

const (
	PermissionViewer SpacePermission = "viewer"
	PermissionEditor SpacePermission = "editor"
	PermissionAdmin  SpacePermission = "admin"
	PermissionOwner  SpacePermission = "owner"
)

// PageStatus 文档状态枚举。
type PageStatus string

const (
	PageStatusDraft     PageStatus = "draft"
	PageStatusPublished PageStatus = "published"
	PageStatusArchived  PageStatus = "archived"
)

// PageRelationType 文档与工作项的关联类型枚举。
type PageRelationType string

const (
	RelationReferenced  PageRelationType = "referenced"
	RelationReferencing PageRelationType = "referencing"
)

// --- 主实体 ---

// KnowledgeSpace 知识库空间。
type KnowledgeSpace struct {
	ID                int64          `json:"id"`
	WorkspaceID       int64          `json:"workspace_id"`
	ProjectID         *int64         `json:"project_id,omitempty"`
	Name              string         `json:"name"`
	Slug              string         `json:"slug"`
	Description       string         `json:"description,omitempty"`
	OwnerID           *int64         `json:"owner_id,omitempty"`
	DefaultPermission SpacePermission `json:"default_permission"`
	IsPrivate         bool           `json:"is_private"`
	CoverImage        string         `json:"cover_image,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         *time.Time     `json:"deleted,omitempty"`
}

// KnowledgePage 知识库文档。
type KnowledgePage struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	SpaceID     int64      `json:"space_id"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Lft         int64      `json:"lft"`
	Rgt         int64      `json:"rgt"`
	Depth       int        `json:"depth"`
	Title       string     `json:"title"`
	Path        string     `json:"path,omitempty"`
	ContentMD   string     `json:"content_md,omitempty"`
	ContentHTML string     `json:"content_html,omitempty"`
	Version     int64      `json:"version"`
	Status      PageStatus `json:"status"`
	SortOrder   int64      `json:"sort_order"`
	IsPinned    bool       `json:"is_pinned"`
	IsFeatured  bool       `json:"is_featured"`
	ViewCount   int64      `json:"view_count"`
	CreatedBy   *int64     `json:"created_by,omitempty"`
	UpdatedBy   *int64     `json:"updated_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted,omitempty"`
}

// KnowledgePageVersion 版本快照。
type KnowledgePageVersion struct {
	ID            int64     `json:"id"`
	PageID        int64     `json:"page_id"`
	Version       int64     `json:"version"`
	Title         string    `json:"title"`
	ContentMD     string    `json:"content_md,omitempty"`
	ContentHTML   string    `json:"content_html,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	CreatedBy     *int64    `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// KnowledgePageRelation 文档与工作项的关联关系。
type KnowledgePageRelation struct {
	ID           int64           `json:"id"`
	PageID       int64           `json:"page_id"`
	IssueID      int64           `json:"issue_id"`
	RelationType PageRelationType `json:"relation_type"`
	CreatedAt    time.Time       `json:"created_at"`
}

// --- 输入 DTO ---

// CreateSpaceInput 创建空间入参。
type CreateSpaceInput struct {
	WorkspaceID       int64          `json:"workspace_id"`
	ProjectID         *int64         `json:"project_id,omitempty"`
	Name              string         `json:"name" binding:"required,max=255"`
	Slug              string         `json:"slug" binding:"required,max=128"`
	Description       string         `json:"description"`
	OwnerID           *int64         `json:"owner_id"`
	DefaultPermission SpacePermission `json:"default_permission"`
	IsPrivate         bool           `json:"is_private"`
	CoverImage        string         `json:"cover_image"`
}

// UpdateSpaceInput 更新空间入参（指针字段为 nil 时不更新）。
type UpdateSpaceInput struct {
	Name              *string         `json:"name,omitempty"`
	Description       *string         `json:"description,omitempty"`
	DefaultPermission *SpacePermission `json:"default_permission,omitempty"`
	IsPrivate         *bool           `json:"is_private,omitempty"`
	CoverImage        *string         `json:"cover_image,omitempty"`
}

// CreatePageInput 创建文档入参。
type CreatePageInput struct {
	WorkspaceID int64      `json:"workspace_id"`
	SpaceID     int64      `json:"space_id"`
	ParentID    *int64     `json:"parent_id"`
	Title       string     `json:"title" binding:"required,max=512"`
	ContentMD   string     `json:"content_md"`
	ContentHTML string     `json:"content_html"`
	Status      PageStatus `json:"status"`
	SortOrder   int64      `json:"sort_order"`
}

// UpdatePageInput 更新文档入参（指针字段为 nil 时不更新）。
type UpdatePageInput struct {
	Title         *string     `json:"title"`
	ContentMD     *string     `json:"content_md"`
	ContentHTML   *string     `json:"content_html"`
	ParentID      *int64      `json:"parent_id"`
	Status        *PageStatus `json:"status"`
	SortOrder     *int64      `json:"sort_order"`
	IsPinned      *bool       `json:"is_pinned"`
	IsFeatured    *bool       `json:"is_featured"`
	Version       int64       `json:"version"`
	ChangeSummary *string     `json:"change_summary"`
}

// CreatePageVersionInput 版本快照入参（UpdatePage 内部使用）。
type CreatePageVersionInput struct {
	PageID        int64  `json:"page_id"`
	Title         string `json:"title"`
	ContentMD     string `json:"content_md"`
	ContentHTML   string `json:"content_html"`
	ChangeSummary string `json:"change_summary"`
}

// AddPageRelationInput 添加关联入参。
type AddPageRelationInput struct {
	PageID       int64           `json:"page_id"`
	IssueID      int64           `json:"issue_id"`
	RelationType PageRelationType `json:"relation_type"`
}

// --- 查询 Options ---

// ListSpacesOptions 空间列表查询选项。
type ListSpacesOptions struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   *int64 `json:"project_id,omitempty"`
	Keyword     string `json:"keyword"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// ListPagesOptions 文档列表查询选项。
type ListPagesOptions struct {
	WorkspaceID int64  `json:"workspace_id"`
	SpaceID     int64  `json:"space_id"`
	ParentID    *int64 `json:"parent_id,omitempty"`
	Status      string `json:"status"`
	Keyword     string `json:"keyword"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

// ListVersionsOptions 版本快照列表查询选项。
type ListVersionsOptions struct {
	PageID int64 `json:"page_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// ListRelationsOptions 关联关系列表查询选项。
type ListRelationsOptions struct {
	PageID int64 `json:"page_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// --- 输出去重 DTO ---

// PageVersionDetail 完整版本快照（查询响应）。
type PageVersionDetail struct {
	ID            int64     `json:"id"`
	PageID        int64     `json:"page_id"`
	Version       int64     `json:"version"`
	Title         string    `json:"title"`
	ContentMD     string    `json:"content_md,omitempty"`
	ContentHTML   string    `json:"content_html,omitempty"`
	ChangeSummary string    `json:"change_summary,omitempty"`
	CreatedBy     *int64    `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// KnowledgePageNode 文档树形节点（包含子节点递归）。
type KnowledgePageNode struct {
	KnowledgePage
	Children []KnowledgePageNode `json:"children"`
}
