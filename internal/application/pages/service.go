// Package pages — 项目文档页面领域（对标 Plane Pages）。
//
// 项目级页面树：页面通过 parent_id 嵌套，sort_order 排序；
// 更新走乐观锁（version 字段），删除为软删除。
package pages

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// defaultSortOrder 新建页面时的默认排序值（与 issues 表约定一致）。
const defaultSortOrder = 65535.0

// Service 提供页面 CRUD 应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建页面服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// CreatePageInput 创建页面的入参。
type CreatePageInput struct {
	Name                string          `json:"name"`
	DescriptionJSON     json.RawMessage `json:"description_json"`
	DescriptionHTML     string          `json:"description_html"`
	DescriptionStripped string          `json:"description_stripped"`
	ParentID            *int64          `json:"parent_id"`
	SortOrder           *float64        `json:"sort_order"`
	Category            string          `json:"category"`
}

// UpdatePageInput 更新页面的入参。
type UpdatePageInput struct {
	Name                *string         `json:"name"`
	DescriptionJSON     json.RawMessage `json:"description_json"`
	DescriptionHTML     *string         `json:"description_html"`
	DescriptionStripped *string         `json:"description_stripped"`
	ParentID            *int64          `json:"parent_id"`
	SortOrder           *float64        `json:"sort_order"`
	Category            *string         `json:"category"`
	Version             int32           `json:"version"`
}

// pageColumns 与 pages 表列一一对应（List/Get 共用）。
const pageColumns = `id, public_id, workspace_id, project_id, name,
	description_json, description_html, description_stripped,
	parent_id, sort_order, category, created_by, created_at, updated_at, deleted_at, version`

// List 列出项目下全部未删除页面，按 sort_order、created_at 排序。
func (s *Service) List(ctx context.Context, wsID, projectID int64) ([]Page, error) {
	rows, err := s.db.Query(ctx, `
		SELECT `+pageColumns+`
		FROM pages
		WHERE workspace_id = $1 AND project_id = $2 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC`, wsID, projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.List: %w", err))
	}
	defer rows.Close()

	var items []Page
	for rows.Next() {
		var p Page
		if err := scanPage(rows, &p); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// Get 获取单个未删除页面；不存在返回 errs.ErrNotFound。
func (s *Service) Get(ctx context.Context, wsID, projectID, pageID int64) (*Page, error) {
	var p Page
	err := s.db.QueryRow(ctx, `
		SELECT `+pageColumns+`
		FROM pages
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3 AND deleted_at IS NULL`,
		pageID, wsID, projectID).Scan(
		&p.ID, &p.PublicID, &p.WorkspaceID, &p.ProjectID, &p.Name,
		&p.DescriptionJSON, &p.DescriptionHTML, &p.DescriptionStripped,
		&p.ParentID, &p.SortOrder, &p.Category, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&p.DeletedAt, &p.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// Create 创建页面，默认 sort_order 65535；校验名称非空。
func (s *Service) Create(ctx context.Context, wsID, projectID, userID int64, input CreatePageInput) (*Page, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "文档名称不能为空"})
	}

	sortOrder := defaultSortOrder
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}

	// 空 JSON 一律落库为 NULL，避免写入空串
	descJSON := nullableJSON(input.DescriptionJSON)

	var p Page
	err := s.db.QueryRow(ctx, `
		INSERT INTO pages (workspace_id, project_id, name, description_json, description_html,
			description_stripped, parent_id, sort_order, category, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+pageColumns,
		wsID, projectID, input.Name, descJSON, input.DescriptionHTML,
		input.DescriptionStripped, input.ParentID, sortOrder, nullableStr(input.Category), userID).Scan(
		&p.ID, &p.PublicID, &p.WorkspaceID, &p.ProjectID, &p.Name,
		&p.DescriptionJSON, &p.DescriptionHTML, &p.DescriptionStripped,
		&p.ParentID, &p.SortOrder, &p.Category, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&p.DeletedAt, &p.Version)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.Create: %w", err))
	}
	return &p, nil
}

// Update 更新页面（乐观锁：WHERE version = input.Version）。
// 无行受影响时区分两种失败：页面不存在 → ErrNotFound；版本不匹配 → ErrVersionConflict。
// 内容（description_html / description_stripped）变更时自动创建版本快照。
func (s *Service) Update(ctx context.Context, wsID, projectID, pageID, userID int64, input UpdatePageInput) (*Page, error) {
	// 预先获取当前页面内容（用于快照的 content_md 字段）
	var currentHTML, currentStripped string
	if input.DescriptionHTML != nil || input.DescriptionStripped != nil {
		current, err := s.Get(ctx, wsID, projectID, pageID)
		if err != nil {
			return nil, err
		}
		currentHTML = current.DescriptionHTML
		currentStripped = current.DescriptionStripped
	}

	sets, args := buildUpdateSet(input)
	if len(sets) == 0 {
		return s.Get(ctx, wsID, projectID, pageID)
	}

	// 每次成功更新版本号 +1，供下次乐观锁比对
	sets = append(sets, "version = version + 1", "updated_at = now()")

	args = append(args, pageID, wsID, projectID, input.Version)
	idIdx := len(args) - 3
	wsIdx := len(args) - 2
	pidIdx := len(args) - 1
	verIdx := len(args)

	query := fmt.Sprintf(`UPDATE pages SET %s
		WHERE id = $%d AND workspace_id = $%d AND project_id = $%d AND version = $%d AND deleted_at IS NULL`,
		strings.Join(sets, ", "), idIdx, wsIdx, pidIdx, verIdx)

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.Update: %w", err))
	}
	if tag.RowsAffected() == 0 {
		// 区分 404 与 409：先确认页面是否仍然存在
		if _, err := s.Get(ctx, wsID, projectID, pageID); err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return nil, errs.ErrNotFound
			}
			return nil, err
		}
		return nil, errs.ErrVersionConflict
	}

	// 内容变更时自动创建版本快照
	if input.DescriptionHTML != nil || input.DescriptionStripped != nil {
		contentMD := currentStripped
		if input.DescriptionStripped != nil {
			contentMD = *input.DescriptionStripped
		}
		contentHTML := currentHTML
		if input.DescriptionHTML != nil {
			contentHTML = *input.DescriptionHTML
		}
		_, _ = s.CreateVersion(ctx, wsID, projectID, pageID, userID, contentMD, contentHTML)
	}

	return s.Get(ctx, wsID, projectID, pageID)
}

// Delete 软删除页面。
func (s *Service) Delete(ctx context.Context, wsID, projectID, pageID, userID int64) error {
	_ = userID

	tag, err := s.db.Exec(ctx, `
		UPDATE pages SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3 AND deleted_at IS NULL`,
		pageID, wsID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pages.Delete: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// buildUpdateSet 动态构建 SET 子句与参数（指针字段为 nil 时不更新）。
func buildUpdateSet(input UpdatePageInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", arg))
		args = append(args, *input.Name)
		arg++
	}
	if len(input.DescriptionJSON) > 0 {
		sets = append(sets, fmt.Sprintf("description_json = $%d", arg))
		args = append(args, input.DescriptionJSON)
		arg++
	}
	if input.DescriptionHTML != nil {
		sets = append(sets, fmt.Sprintf("description_html = $%d", arg))
		args = append(args, *input.DescriptionHTML)
		arg++
	}
	if input.DescriptionStripped != nil {
		sets = append(sets, fmt.Sprintf("description_stripped = $%d", arg))
		args = append(args, *input.DescriptionStripped)
		arg++
	}
	if input.ParentID != nil {
		sets = append(sets, fmt.Sprintf("parent_id = $%d", arg))
		args = append(args, *input.ParentID)
		arg++
	}
	if input.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", arg))
		args = append(args, *input.SortOrder)
		arg++
	}
	if input.Category != nil {
		sets = append(sets, fmt.Sprintf("category = $%d", arg))
		args = append(args, *input.Category)
		arg++
	}

	return sets, args
}

// nullableStr 空字符串转为 nil（落库 NULL），否则原样返回。
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// scanPage 从 pgx 行读取一条 Page。使用 sql.Null 类型桥接可空列。
func scanPage(row pgx.Row, p *Page) error {
	var parentID sql.NullInt64
	var deletedAt sql.NullTime
	var category sql.NullString
	err := row.Scan(
		&p.ID, &p.PublicID, &p.WorkspaceID, &p.ProjectID, &p.Name,
		&p.DescriptionJSON, &p.DescriptionHTML, &p.DescriptionStripped,
		&parentID, &p.SortOrder, &category, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&deletedAt, &p.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid {
		v := parentID.Int64
		p.ParentID = &v
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	if category.Valid {
		p.Category = category.String
	}
	return nil
}

// nullableJSON 空 RawMessage 转为 nil（落库 NULL），否则原样返回。
func nullableJSON(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

// --- 版本历史 ---

// CreateVersion 为指定页面创建版本快照（在 Update 成功后自动调用）。
func (s *Service) CreateVersion(ctx context.Context, wsID, projectID, pageID, userID int64, contentMD, contentHTML string) (*DocumentVersion, error) {
	// 获取当前最新版本号并 +1
	var nextVersion int32
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(version_number), 0) + 1
		FROM document_versions WHERE page_id = $1`, pageID).Scan(&nextVersion)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.CreateVersion: %w", err))
	}

	var dv DocumentVersion
	err = s.db.QueryRow(ctx, `
		INSERT INTO document_versions (page_id, version_number, content_md, content_html, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, page_id, version_number, content_md, content_html, created_by, created_at`,
		pageID, nextVersion, contentMD, contentHTML, userID).Scan(
		&dv.ID, &dv.PageID, &dv.VersionNumber, &dv.ContentMD,
		&dv.ContentHTML, &dv.CreatedBy, &dv.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.CreateVersion: %w", err))
	}
	return &dv, nil
}

// ListVersions 获取指定页面的所有历史版本（按版本号降序）。
func (s *Service) ListVersions(ctx context.Context, pageID int64) ([]DocumentVersion, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, page_id, version_number, content_md, content_html, created_by, created_at
		FROM document_versions WHERE page_id = $1
		ORDER BY version_number DESC`, pageID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.ListVersions: %w", err))
	}
	defer rows.Close()

	var items []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(&v.ID, &v.PageID, &v.VersionNumber,
			&v.ContentMD, &v.ContentHTML, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		items = append(items, v)
	}
	return items, rows.Err()
}

// GetVersion 获取指定版本快照。
func (s *Service) GetVersion(ctx context.Context, pageID int64, versionNumber int32) (*DocumentVersion, error) {
	var v DocumentVersion
	err := s.db.QueryRow(ctx, `
		SELECT id, page_id, version_number, content_md, content_html, created_by, created_at
		FROM document_versions WHERE page_id = $1 AND version_number = $2`,
		pageID, versionNumber).Scan(
		&v.ID, &v.PageID, &v.VersionNumber,
		&v.ContentMD, &v.ContentHTML, &v.CreatedBy, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &v, nil
}

// RollbackToVersion 将页面内容回滚到指定版本。
func (s *Service) RollbackToVersion(ctx context.Context, wsID, projectID, pageID, userID int64, versionNumber int32) (*Page, error) {
	dv, err := s.GetVersion(ctx, pageID, versionNumber)
	if err != nil {
		return nil, err
	}

	// 先获取当前页面
	current, err := s.Get(ctx, wsID, projectID, pageID)
	if err != nil {
		return nil, err
	}

	html := dv.ContentHTML
	md := dv.ContentMD
	return s.Update(ctx, wsID, projectID, pageID, userID, UpdatePageInput{
		DescriptionHTML:     &html,
		DescriptionStripped: &md,
		Version:             current.Version,
	})
}

// SavePageSnapshot 保存当前页面内容为新的版本快照（外部调用入口）。
func (s *Service) SavePageSnapshot(ctx context.Context, wsID, projectID, pageID, userID int64, contentMD, contentHTML string) (*DocumentVersion, error) {
	return s.CreateVersion(ctx, wsID, projectID, pageID, userID, contentMD, contentHTML)
}

// --- 文档关联 ---

// CreateLink 创建文档与外部实体（Issue/Sprint/Version）的关联。
func (s *Service) CreateLink(ctx context.Context, pageID, userID int64, linkableType string, linkableID int64) (*DocumentLink, error) {
	var dl DocumentLink
	err := s.db.QueryRow(ctx, `
		INSERT INTO document_links (page_id, linkable_type, linkable_id, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, page_id, linkable_type, linkable_id, created_by, created_at`,
		pageID, linkableType, linkableID, userID).Scan(
		&dl.ID, &dl.PageID, &dl.LinkableType, &dl.LinkableID,
		&dl.CreatedBy, &dl.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.CreateLink: %w", err))
	}
	return &dl, nil
}

// ListLinks 获取指定页面的所有关联。
func (s *Service) ListLinks(ctx context.Context, pageID int64) ([]DocumentLink, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, page_id, linkable_type, linkable_id, created_by, created_at
		FROM document_links WHERE page_id = $1
		ORDER BY created_at DESC`, pageID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pages.ListLinks: %w", err))
	}
	defer rows.Close()

	var items []DocumentLink
	for rows.Next() {
		var l DocumentLink
		if err := rows.Scan(&l.ID, &l.PageID, &l.LinkableType, &l.LinkableID,
			&l.CreatedBy, &l.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

// DeleteLink 删除文档关联。
func (s *Service) DeleteLink(ctx context.Context, pageID, linkID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM document_links WHERE id = $1 AND page_id = $2`,
		linkID, pageID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pages.DeleteLink: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
