// Package attachment 附件域模型与输入类型。
//
// 参照: Plane / Linear 附件模型 — 按工作项类型分表存储
// （task_attachments / requirement_attachments / defect_attachments），
// 文件实际存储于 S3/MinIO，DB 仅记录元数据，预签名 URL 按需生成。
//
// 设计原则：禁止 entity_type + entity_id 多态关联，附件按业务类型独立分表。
package attachment

import "time"

// Attachment 附件域模型（跨 per-type 表的统一视图）。
type Attachment struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	// EntityType 由调用方上下文决定，JSON 序列化时回填。
	EntityType string `json:"entity_type,omitempty"`
	// EntityID 为对应 per-type 表中具体 FK 的值（task_id / requirement_id / defect_id）。
	EntityID    int64     `json:"entity_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	StorageKey  string    `json:"storage_key"`
	// StorageURL 为按需生成的预签名下载 URL，不持久化。
	StorageURL string    `json:"storage_url,omitempty"`
	ThumbKey   string    `json:"thumb_key,omitempty"`
	UploadedBy int64     `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// EntityType 支持的附件关联实体类型。
type EntityType string

const (
	EntityTask        EntityType = "task"
	EntityRequirement EntityType = "requirement"
	EntityDefect      EntityType = "defect"
)

// IsValid 校验 entity_type 是否为支持的类型。
func (e EntityType) IsValid() bool {
	switch e {
	case EntityTask, EntityRequirement, EntityDefect:
		return true
	}
	return false
}

// String 返回字符串表示。
func (e EntityType) String() string {
	return string(e)
}

// ResolveEntityType 将字符串解析为 EntityType，支持 "issue" 别名映射。
// "issue" 为前端通用标识，需配合 entityID 查询具体类型后解析。
func ResolveEntityType(s string) (EntityType, bool) {
	switch s {
	case "task":
		return EntityTask, true
	case "requirement":
		return EntityRequirement, true
	case "defect":
		return EntityDefect, true
	}
	return "", false
}

// CreateInput 创建附件记录的内部入参。
type CreateInput struct {
	WorkspaceID int64
	ProjectID   int64
	EntityType  EntityType
	EntityID    int64
	FileName    string
	FileSize    int64
	ContentType string
	StorageKey  string
	UploadedBy  int64
}

// PresignedUploadInput 预签名上传请求入参。
type PresignedUploadInput struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	EntityType  string `json:"entity_type" binding:"required"`
	EntityID    int64  `json:"entity_id" binding:"required"`
	UploadedBy  int64  `json:"-"`
}

// PresignedUploadResult 预签名上传响应。
type PresignedUploadResult struct {
	UploadURL  string      `json:"upload_url"`
	StorageKey string      `json:"storage_key"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// ConfirmUploadInput 上传确认入参。
type ConfirmUploadInput struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
	EntityType  string `json:"entity_type" binding:"required"`
	EntityID    int64  `json:"entity_id" binding:"required"`
	StorageKey  string `json:"storage_key" binding:"required"`
	UploadedBy  int64  `json:"-"`
}

// ConfirmUploadResult 上传确认响应。
type ConfirmUploadResult struct {
	Attachment Attachment `json:"attachment"`
}

// ListResponse 附件列表响应。
type ListResponse struct {
	Results []Attachment `json:"results"`
}
