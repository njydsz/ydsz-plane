// Package attachment 附件域应用层：管理文件上传、下载与关联。
//
// 按需求/任务/缺陷类型分表存储（task_attachments / requirement_attachments / defect_attachments），
// 禁止 entity_type + entity_id 多态关联。
package attachment

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/storage"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// entityTable 定义 EntityType 到表名和 FK 列的映射。
type entityTable struct {
	tableName string
	fkColumn  string
}

// entityTableMap 支持的附件关联类型映射表。
var entityTableMap = map[EntityType]entityTable{
	EntityTask:        {tableName: "task_attachments", fkColumn: "task_id"},
	EntityRequirement: {tableName: "requirement_attachments", fkColumn: "requirement_id"},
	EntityDefect:      {tableName: "defect_attachments", fkColumn: "defect_id"},
}

// Service 附件应用服务。
type Service struct {
	db *pgxpool.Pool
	st *storage.Client
}

// NewService 创建附件服务。
func NewService(db *pgxpool.Pool, st *storage.Client) *Service {
	return &Service{db: db, st: st}
}

// Storage 返回存储客户端（供需要直接操作存储的场景使用）。
func (s *Service) Storage() *storage.Client { return s.st }

// resolveTable 根据 EntityType 返回对应的表名和 FK 列。
func (s *Service) resolveTable(et EntityType) (entityTable, error) {
	et2, ok := entityTableMap[et]
	if !ok {
		return entityTable{}, errs.Validation("ATTACHMENT.UNSUPPORTED_ENTITY_TYPE",
			fmt.Sprintf("不支持的附件关联类型: %s", et))
	}
	return et2, nil
}

// List 查询某工作项下的所有附件，并为每条记录生成 24 小时有效期的预签名下载 URL。
func (s *Service) List(ctx context.Context, wsID, projectID int64, entityType EntityType, entityID int64) ([]Attachment, error) {
	et, err := s.resolveTable(entityType)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT id, workspace_id, project_id, %s,
		       file_name, file_size, file_type, storage_path, thumbnail_path,
		       created_by, created_at, updated_at
		FROM %s
		WHERE workspace_id = $1 AND project_id = $2
		  AND %s = $3
		  AND deleted = false
		ORDER BY created_at ASC`,
		et.fkColumn, et.tableName, et.fkColumn),
		wsID, projectID, entityID)
	if err != nil {
		return nil, fmt.Errorf("Attachment.List: %w", err)
	}
	defer rows.Close()

	downloadExpiry := 24 * time.Hour
	atts := make([]Attachment, 0)
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(
			&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityID,
			&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey, &a.ThumbKey,
			&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("Attachment.List scan: %w", err)
		}
		a.EntityType = entityType.String()
		// 生成预签名下载 URL（失败不中断整体流程，仅跳过 URL 填充）
		if url, err := s.st.PresignedDownloadURL(ctx, a.StorageKey, downloadExpiry); err == nil {
			a.StorageURL = url
		}
		atts = append(atts, a)
	}
	return atts, nil
}

// CreatePresignedUpload 生成 UUID 存储 key，预签名 PUT URL（15 分钟）。
// 仅返回上传 URL 与 storage_key，不写入 DB；客户端 PUT 成功后须调用 ConfirmUpload 落库。
func (s *Service) CreatePresignedUpload(ctx context.Context, wsID, projectID int64, entityType EntityType, input PresignedUploadInput) (*PresignedUploadResult, error) {
	if _, err := s.resolveTable(entityType); err != nil {
		return nil, err
	}

	ct := input.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}

	// 生成 UUID 存储 key：{ws}/{project}/{entity_type}/{entity_id}/{uuid}_{filename}
	id := uuid.NewString()
	ext := filepath.Ext(input.FileName)
	base := sanitizeFilename(input.FileName)
	storageKey := fmt.Sprintf("%d/%d/%s/%d/%s_%s%s",
		wsID, projectID, entityType, input.EntityID,
		id, base, ext)

	// 预签名 PUT URL，15 分钟有效期
	uploadURL, err := s.st.PresignedUploadURL(ctx, storageKey, 15*time.Minute, ct)
	if err != nil {
		return nil, fmt.Errorf("Attachment.CreatePresignedUpload: presign: %w", err)
	}

	return &PresignedUploadResult{
		UploadURL:  uploadURL,
		StorageKey: storageKey,
	}, nil
}

// ConfirmUpload 校验对象存储中的文件已存在（stat），然后写入 DB 记录并返回附件。
func (s *Service) ConfirmUpload(ctx context.Context, wsID, projectID int64, entityType EntityType, input ConfirmUploadInput) (*Attachment, error) {
	et, err := s.resolveTable(entityType)
	if err != nil {
		return nil, err
	}

	// 校验对象已上传
	exists, err := s.st.Exists(ctx, input.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("Attachment.ConfirmUpload: storage stat: %w", err)
	}
	if !exists {
		return nil, errs.NotFound("ATTACHMENT.NOT_UPLOADED", "文件尚未上传或已过期，请重新上传")
	}

	// 获取实际存储大小（stat 返回）；客户端传入的 file_size 作为 fallback
	size := input.FileSize
	if actual, err := s.st.Size(ctx, input.StorageKey); err == nil && actual > 0 {
		size = actual
	}

	ct := input.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}

	var a Attachment
	err = s.db.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s
			(workspace_id, project_id, %s,
		     file_name, file_size, file_type, storage_path,
		     created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())
		RETURNING id, workspace_id, project_id, %s,
			file_name, file_size, file_type, storage_path, thumbnail_path,
			created_by, created_at, updated_at`,
		et.tableName, et.fkColumn, et.fkColumn),
		wsID, projectID, input.EntityID,
		input.FileName, size, ct, input.StorageKey,
		input.UploadedBy,
	).Scan(
		&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityID,
		&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey, &a.ThumbKey,
		&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Attachment.ConfirmUpload: insert: %w", err)
	}
	a.EntityType = entityType.String()
	return &a, nil
}

// Delete 加载附件并从存储删除文件，然后软删除 DB 记录。
func (s *Service) Delete(ctx context.Context, wsID, projectID int64, entityType EntityType, attachmentID int64, userID int64) error {
	et, err := s.resolveTable(entityType)
	if err != nil {
		return err
	}

	// 加载附件获取 storage_path
	var a Attachment
	err = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, workspace_id, project_id, %s,
		       file_name, file_size, file_type, storage_path,
		       created_by, created_at, updated_at
		FROM %s
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3 AND deleted = false`,
		et.fkColumn, et.tableName),
		attachmentID, wsID, projectID,
	).Scan(
		&a.ID, &a.WorkspaceID, &a.ProjectID, &a.EntityID,
		&a.FileName, &a.FileSize, &a.ContentType, &a.StorageKey,
		&a.UploadedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return errs.NotFound("ATTACHMENT.NOT_FOUND", "附件不存在")
	}

	// 从对象存储删除文件（失败不阻断，记录后由清理任务兜底）
	if err := s.st.Delete(ctx, a.StorageKey); err != nil {
		_ = err
	}

	// 软删除 DB 记录
	tag, err := s.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s SET deleted = NOW(), updated_at = NOW()
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3
		  AND created_by = $4 AND deleted = false`,
		et.tableName),
		attachmentID, wsID, projectID, userID)
	if err != nil {
		return fmt.Errorf("Attachment.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("ATTACHMENT.NOT_FOUND", "附件不存在或无权删除")
	}
	return nil
}

// CountByEntity 统计某工作项下的附件数量（用于总量限制校验）。
func (s *Service) CountByEntity(ctx context.Context, wsID, projectID int64, entityType EntityType, entityID int64) (int, error) {
	et, err := s.resolveTable(entityType)
	if err != nil {
		return 0, err
	}

	var count int
	err = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s
		WHERE workspace_id = $1 AND project_id = $2 AND %s = $3 AND deleted = false`,
		et.tableName, et.fkColumn),
		wsID, projectID, entityID).Scan(&count)
	return count, err
}

// TotalSizeByEntity 统计某工作项下附件总大小（用于容量限制校验）。
func (s *Service) TotalSizeByEntity(ctx context.Context, wsID, projectID int64, entityType EntityType, entityID int64) (int64, error) {
	et, err := s.resolveTable(entityType)
	if err != nil {
		return 0, err
	}

	var total sql.NullInt64
	err = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COALESCE(SUM(file_size), 0) FROM %s
		WHERE workspace_id = $1 AND project_id = $2 AND %s = $3 AND deleted = false`,
		et.tableName, et.fkColumn),
		wsID, projectID, entityID).Scan(&total)
	return total.Int64, err
}

// sanitizeFilename 清理文件名，移除路径分隔符与特殊字符。
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\\' || r == '/' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
