// Package attachment 附件域模型与输入类型。
//
// 参照: Plane / Linear 附件模型 — 多态关联实体（issue / comment / workspace / project），
// 文件实际存储于 S3/MinIO，DB 仅记录元数据，预签名 URL 按需生成。
package attachment

import "time"

// Attachment 附件域模型。
type Attachment struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	EntityType  string    `json:"entity_type"`
	EntityID    int64     `json:"entity_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	StorageKey  string    `json:"storage_key"`
	// StorageURL 为按需生成的预签名下载 URL，不持久化，JSON omitempty 在设置后保留。
	StorageURL string    `json:"storage_url,omitempty"`
	ThumbKey   string    `json:"thumb_key,omitempty"`
	UploadedBy int64     `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateInput 创建附件记录的内部入参。
type CreateInput struct {
	WorkspaceID int64
	ProjectID   int64
	EntityType  string
	EntityID    int64
	FileName    string
	FileSize    int64
	ContentType string
	StorageKey  string
	UploadedBy  int64
}

// PresignedUploadInput 预签名上传请求入参。
// UploadedBy 不由客户端 JSON 传入，由 handler 从认证上下文中注入。
type PresignedUploadInput struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	EntityType  string `json:"entity_type" binding:"required"`
	EntityID    int64  `json:"entity_id" binding:"required"`
	UploadedBy  int64  `json:"-"`
}

// PresignedUploadResult 预签名上传响应。
// Attachment 在确认上传（ConfirmUpload）之前为 nil，前端不应依赖此字段。
type PresignedUploadResult struct {
	UploadURL  string      `json:"upload_url"`
	StorageKey string      `json:"storage_key"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// ConfirmUploadInput 上传确认入参。
// 客户端在 PUT 成功后提交，服务端校验存储对象存在后写入 DB。
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
