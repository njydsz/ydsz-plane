// Package attachment 附件域应用层：管理文件上传、下载与关联。
package attachment

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/storage"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Attachment 附件模型。
type Attachment struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   int64      `json:"project_id"`
	EntityType  string     `json:"entity_type"`
	EntityID    int64      `json:"entity_id"`
	FileName    string     `json:"file_name"`
	FileSize    int64      `json:"file_size"`
	ContentType string     `json:"content_type"`
	StorageKey  string     `json:"storage_key"`
	StorageURL  string     `json:"storage_url,omitempty"`
	ThumbKey    string     `json:"thumb_key,omitempty"`
	UploadedBy  int64      `json:"uploaded_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateInput 创建附件入参。
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

// Service 附件应用服务。
type Service struct {
	db  *pgxpool.Pool
	st  *storage.Client
}

// NewService 创建附件服务。
func NewService(db *pgxpool.Pool, st *storage.Client) *Service {
	return &Service{db: db, st: st}
}

// Storage 返回存储客户端（供 handler 直接调用预签名等底层操作）。
func (s *Service) Storage() *storage.Client { return s.st }

// Create 在数据库中创建附件记录。
func (s *Service) Create(ctx context.Context, input CreateInput) (*Attachment, error) {
	var a Attachment
	err := s.db.QueryRow(ctx, `
		INSERT INTO attachments
			(workspace_id, project_id, entity_type, entity_id,
			 file_name, file_size, content_type, storage_key,
			 uploaded_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
		RETURNING id, workspace_id, project_id, entity_type, entity_id,
			file_name, file_size, content_type, storage_key,
			uploaded_by, created_at, updated_at`,
		input.WorkspaceID, input.ProjectID, input.EntityType, input.EntityID,
		input.FileName, input.FileSize, input.ContentType, input.StorageKey,
		input.UploadedBy,
	).Scan(
		&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityType, &a.EntityID,
		&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey,
		&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Attachment.Create: %w", err)
	}
	return &a, nil
}

// ListByEntity 查询某实体下的所有附件。
func (s *Service) ListByEntity(ctx context.Context, entityType string, entityID int64) ([]Attachment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, entity_type, entity_id,
			file_name, file_size, content_type, storage_key,
			uploaded_by, created_at, updated_at
		FROM attachments
		WHERE entity_type = $1 AND entity_id = $2 AND deleted_at IS NULL
		ORDER BY created_at ASC`, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("Attachment.ListByEntity: %w", err)
	}
	defer rows.Close()

	var atts []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(
			&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityType, &a.EntityID,
			&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey,
			&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("Attachment.ListByEntity scan: %w", err)
		}
		atts = append(atts, a)
	}
	if atts == nil {
		atts = []Attachment{}
	}
	return atts, nil
}

// Get 查询单个附件。
func (s *Service) Get(ctx context.Context, id int64) (*Attachment, error) {
	var a Attachment
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, entity_type, entity_id,
			file_name, file_size, content_type, storage_key,
			uploaded_by, created_at, updated_at
		FROM attachments
		WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(
		&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityType, &a.EntityID,
		&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey,
		&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Attachment.Get: %w", err)
	}
	return &a, nil
}

// Delete 软删除附件（仅标记 deleted_at，不删除实际文件）。
func (s *Service) Delete(ctx context.Context, id, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE attachments SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND uploaded_by = $2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return fmt.Errorf("Attachment.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("ATTACHMENT.NOT_FOUND", "附件不存在或无权删除")
	}
	return nil
}
