// Package knowledge — 知识库应用服务（Service 层）。
//
// 实现知识库空间 CRUD、文档树管理、乐观锁版本控制、自动版本快照、工作项关联。
package knowledge

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供知识库应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建知识库服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ==========================================================================
// 空间 CRUD
// ==========================================================================

// spaceColumns 与 knowledge_spaces 表列共用。
const spaceColumns = `id, workspace_id, project_id, name, slug, description, owner_id,
	default_permission, is_private, cover_image, created_at, updated_at, deleted`

// ListSpaces 列出空间（过滤 workspace_id / project_id / keyword，支持分页）。
func (s *Service) ListSpaces(ctx context.Context, opts ListSpacesOptions) ([]KnowledgeSpace, int64, error) {
	var sets []string
	var args []interface{}
	arg := 1

	sets = append(sets, fmt.Sprintf("workspace_id = $%d", arg))
	args = append(args, opts.WorkspaceID)
	arg++

	sets = append(sets, "deleted = false")

	if opts.ProjectID != nil {
		sets = append(sets, fmt.Sprintf("project_id = $%d", arg))
		args = append(args, *opts.ProjectID)
		arg++
	}
	if opts.Keyword != "" {
		sets = append(sets, fmt.Sprintf("name ILIKE $%d", arg))
		args = append(args, "%"+opts.Keyword+"%")
		arg++
	}

	// 计数
	countQuery := "SELECT COUNT(*) FROM knowledge_spaces WHERE " + strings.Join(sets, " AND ")
	var total int64
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListSpaces count: %w", err))
	}

	// 分页
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT %s FROM knowledge_spaces WHERE %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		spaceColumns, strings.Join(sets, " AND "), arg, arg+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListSpaces: %w", err))
	}
	defer rows.Close()

	var items []KnowledgeSpace
	for rows.Next() {
		var sp KnowledgeSpace
		if err := scanSpace(rows, &sp); err != nil {
			return nil, 0, err
		}
		items = append(items, sp)
	}
	if items == nil {
		items = []KnowledgeSpace{}
	}
	return items, total, nil
}

// GetSpace 获取单个未删除空间。
func (s *Service) GetSpace(ctx context.Context, id, wsID int64) (*KnowledgeSpace, error) {
	var sp KnowledgeSpace
	err := s.db.QueryRow(ctx, `
		SELECT `+spaceColumns+`
		FROM knowledge_spaces
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		id, wsID).Scan(
		&sp.ID, &sp.WorkspaceID, &sp.ProjectID, &sp.Name, &sp.Slug, &sp.Description,
		&sp.OwnerID, &sp.DefaultPermission, &sp.IsPrivate, &sp.CoverImage,
		&sp.CreatedAt, &sp.UpdatedAt, &sp.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.GetSpace: %w", err))
	}
	return &sp, nil
}

// CreateSpace 创建知识库空间。
func (s *Service) CreateSpace(ctx context.Context, in CreateSpaceInput, actorID int64) (*KnowledgeSpace, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "空间名称不能为空"})
	}
	if strings.TrimSpace(in.Slug) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "slug", Reason: "空间标识不能为空"})
	}

	perm := in.DefaultPermission
	if perm == "" {
		perm = PermissionViewer
	}

	var sp KnowledgeSpace
	err := s.db.QueryRow(ctx, `
		INSERT INTO knowledge_spaces (workspace_id, project_id, name, slug, description, owner_id,
			default_permission, is_private, cover_image)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING `+spaceColumns,
		in.WorkspaceID, in.ProjectID, in.Name, in.Slug, nullableStr(in.Description), in.OwnerID,
		perm, in.IsPrivate, nullableStr(in.CoverImage)).Scan(
		&sp.ID, &sp.WorkspaceID, &sp.ProjectID, &sp.Name, &sp.Slug, &sp.Description,
		&sp.OwnerID, &sp.DefaultPermission, &sp.IsPrivate, &sp.CoverImage,
		&sp.CreatedAt, &sp.UpdatedAt, &sp.DeletedAt)
	if err != nil {
		if strings.Contains(err.Error(), "knowledge_spaces_slug_key") {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "slug", Reason: "该空间标识已被占用"})
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.CreateSpace: %w", err))
	}
	return &sp, nil
}

// UpdateSpace 更新空间（动态 SET）。
func (s *Service) UpdateSpace(ctx context.Context, id, wsID int64, in UpdateSpaceInput) (*KnowledgeSpace, error) {
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "空间名称不能为空"})
	}

	sets, args := buildSpaceUpdateSet(in)
	if len(sets) == 0 {
		return s.GetSpace(ctx, id, wsID)
	}
	sets = append(sets, "updated_at = now()")
	args = append(args, id, wsID)
	idIdx := len(args) - 1
	wsIdx := len(args)

	query := fmt.Sprintf(`UPDATE knowledge_spaces SET %s
		WHERE id = $%d AND workspace_id = $%d AND deleted = false`,
		strings.Join(sets, ", "), idIdx, wsIdx)

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.UpdateSpace: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound
	}
	return s.GetSpace(ctx, id, wsID)
}

// DeleteSpace 软删除空间（设置 deleted）。
func (s *Service) DeleteSpace(ctx context.Context, id, wsID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE knowledge_spaces SET deleted = true, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		id, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("knowledge.DeleteSpace: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ==========================================================================
// 文档 CRUD
// ==========================================================================

// pageColumns 与 knowledge_pages 表列共用。
const pageColumns = `id, workspace_id, space_id, parent_id, lft, rgt, depth,
	title, path, content_md, content_html, version, status, sort_order,
	is_pinned, is_featured, view_count, created_by, updated_by, created_at, updated_at, deleted`

// ListPageTree 获取 space 下所有未删除文档，在 Go 端组装为无限层级树。
func (s *Service) ListPageTree(ctx context.Context, wsID, spaceID int64) ([]KnowledgePageNode, error) {
	rows, err := s.db.Query(ctx, `
		`+pageColumns+`
		FROM knowledge_pages
		WHERE workspace_id = $1 AND space_id = $2 AND deleted = false
		ORDER BY depth ASC, sort_order ASC, created_at ASC`,
		wsID, spaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListPageTree: %w", err))
	}
	defer rows.Close()

	type flatNode struct {
		page     KnowledgePage
		children []*flatNode
	}
	nodeMap := make(map[int64]*flatNode)
	var roots []*flatNode

	for rows.Next() {
		var p KnowledgePage
		if err := scanPage(rows, &p); err != nil {
			return nil, err
		}
		n := &flatNode{page: p}
		nodeMap[p.ID] = n
		if p.ParentID == nil {
			roots = append(roots, n)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListPageTree rows: %w", err))
	}

	// 二次遍历：挂载子节点
	for _, n := range nodeMap {
		if n.page.ParentID != nil {
			if parent, ok := nodeMap[*n.page.ParentID]; ok {
				parent.children = append(parent.children, n)
			}
		}
	}

	// 递归转换为输出格式
	var buildNodes func(nodes []*flatNode) []KnowledgePageNode
	buildNodes = func(nodes []*flatNode) []KnowledgePageNode {
		if nodes == nil {
			return []KnowledgePageNode{}
		}
		result := make([]KnowledgePageNode, 0, len(nodes))
		for _, n := range nodes {
			result = append(result, KnowledgePageNode{
				KnowledgePage: n.page,
				Children:      buildNodes(n.children),
			})
		}
		return result
	}

	return buildNodes(roots), nil
}

// GetPage 获取单个未删除文档。
func (s *Service) GetPage(ctx context.Context, id, wsID int64) (*KnowledgePage, error) {
	var p KnowledgePage
	err := s.db.QueryRow(ctx, `
		SELECT `+pageColumns+`
		FROM knowledge_pages
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		id, wsID).Scan(
		&p.ID, &p.WorkspaceID, &p.SpaceID, &p.ParentID, &p.Lft, &p.Rgt, &p.Depth,
		&p.Title, &p.Path, &p.ContentMD, &p.ContentHTML, &p.Version, &p.Status,
		&p.SortOrder, &p.IsPinned, &p.IsFeatured, &p.ViewCount,
		&p.CreatedBy, &p.UpdatedBy, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.GetPage: %w", err))
	}
	return &p, nil
}

// CreatePage 创建文档（parent_id 校验同一 space 下，自动计算 depth 和 path）。
func (s *Service) CreatePage(ctx context.Context, in CreatePageInput, actorID int64) (*KnowledgePage, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "文档标题不能为空"})
	}

	// 校验父文档是否存在（同一 space 下）
	var depth int
	var parentPath string
	if in.ParentID != nil {
		var p KnowledgePage
		err := s.db.QueryRow(ctx, `
			SELECT id, depth, path FROM knowledge_pages
			WHERE id = $1 AND workspace_id = $2 AND space_id = $3 AND deleted = false`,
			*in.ParentID, in.WorkspaceID, in.SpaceID).Scan(&p.ID, &p.Depth, &p.Path)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "父文档不存在"})
			}
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.CreatePage parent: %w", err))
		}
		depth = p.Depth + 1
		parentPath = p.Path
	}

	status := in.Status
	if status == "" {
		status = PageStatusDraft
	}

	var p KnowledgePage
	err := s.db.QueryRow(ctx, `
		INSERT INTO knowledge_pages (workspace_id, space_id, parent_id, depth, title,
			path, content_md, content_html, status, sort_order, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING `+pageColumns,
		in.WorkspaceID, in.SpaceID, in.ParentID, depth, in.Title,
		parentPath+"/"+strings.ToLower(strings.ReplaceAll(in.Title, " ", "-")),
		nullableStr(in.ContentMD), nullableStr(in.ContentHTML), status,
		in.SortOrder, actorID, actorID).Scan(
		&p.ID, &p.WorkspaceID, &p.SpaceID, &p.ParentID, &p.Lft, &p.Rgt, &p.Depth,
		&p.Title, &p.Path, &p.ContentMD, &p.ContentHTML, &p.Version, &p.Status,
		&p.SortOrder, &p.IsPinned, &p.IsFeatured, &p.ViewCount,
		&p.CreatedBy, &p.UpdatedBy, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.CreatePage: %w", err))
	}
	return &p, nil
}

// UpdatePage 更新文档（乐观锁 + 内容变更自动创建版本快照）。
func (s *Service) UpdatePage(ctx context.Context, id, wsID, version int64, in UpdatePageInput) (*KnowledgePage, error) {
	// 预先获取当前内容（用于快照）
	var currentTitle, currentMD, currentHTML string
	if in.ContentMD != nil || in.ContentHTML != nil {
		current, err := s.GetPage(ctx, id, wsID)
		if err != nil {
			return nil, err
		}
		currentTitle = current.Title
		currentMD = current.ContentMD
		currentHTML = current.ContentHTML
	}

	sets, args := buildPageUpdateSet(in)
	if len(sets) == 0 {
		return s.GetPage(ctx, id, wsID)
	}
	sets = append(sets, "version = version + 1", "updated_at = now()")

	args = append(args, id, wsID, version)
	idIdx := len(args) - 2
	wsIdx := len(args) - 1
	verIdx := len(args)

	query := fmt.Sprintf(`UPDATE knowledge_pages SET %s
		WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted = false`,
		strings.Join(sets, ", "), idIdx, wsIdx, verIdx)

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.UpdatePage: %w", err))
	}
	if tag.RowsAffected() == 0 {
		// 区分 404 vs 409
		if _, err := s.GetPage(ctx, id, wsID); err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return nil, errs.ErrNotFound
			}
			return nil, err
		}
		return nil, errs.ErrVersionConflict
	}

	// 变更成功：若 content_md / content_html 有更新，创建版本快照
	if in.ContentMD != nil || in.ContentHTML != nil {
		newMD := currentMD
		if in.ContentMD != nil {
			newMD = *in.ContentMD
		}
		newHTML := currentHTML
		if in.ContentHTML != nil {
			newHTML = *in.ContentHTML
		}
		if in.Title != nil {
			currentTitle = *in.Title
		}
		_, _ = s.createVersionSnapshotTx(ctx, id, currentTitle, newMD, newHTML,
			getChangeSummary(in.ChangeSummary))
	}

	return s.GetPage(ctx, id, wsID)
}

// DeletePage 软删除文档；如有子节点则一同软删。
func (s *Service) DeletePage(ctx context.Context, id, wsID int64, soft bool) error {
	if !soft {
		// 硬删除仅作占位，实际统一走软删除
		return s.softDeletePage(ctx, id, wsID)
	}
	return s.softDeletePage(ctx, id, wsID)
}

// softDeletePage 软删除文档及所有子节点。
func (s *Service) softDeletePage(ctx context.Context, id, wsID int64) error {
	// 先获取 space_id 以支持批量删除同 space 子节点
	var spaceID int64
	err := s.db.QueryRow(ctx, `
		SELECT space_id FROM knowledge_pages WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		id, wsID).Scan(&spaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal.Wrap(fmt.Errorf("knowledge.softDeletePage: %w", err))
	}

	// 删除该文档及 space 下所有后代
	tag, err := s.db.Exec(ctx, `
		UPDATE knowledge_pages SET deleted = true, updated_at = now()
		WHERE workspace_id = $1 AND space_id = $2 AND id IN (
			WITH RECURSIVE subtree AS (
				SELECT id FROM knowledge_pages WHERE id = $3 AND workspace_id = $1 AND space_id = $2
				UNION ALL
				SELECT p.id FROM knowledge_pages p
				JOIN subtree st ON p.parent_id = st.id
				WHERE p.workspace_id = $1 AND p.space_id = $2 AND p.deleted = false
			)
			SELECT id FROM subtree
		)`,
		wsID, spaceID, id)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("knowledge.softDeletePage: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ==========================================================================
// 版本快照
// ==========================================================================

// ListVersions 列出文档所有版本快照（按 version DESC）。
func (s *Service) ListVersions(ctx context.Context, wsID, pageID int64) ([]KnowledgePageVersion, error) {
	// 先校验文档属于该 workspace
	if _, err := s.GetPage(ctx, pageID, wsID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, page_id, version, title, content_md, content_html, change_summary, created_by, created_at
		FROM knowledge_page_versions
		WHERE page_id = $1
		ORDER BY version DESC`,
		pageID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListVersions: %w", err))
	}
	defer rows.Close()

	var items []KnowledgePageVersion
	for rows.Next() {
		var v KnowledgePageVersion
		var changeSummary sql.NullString
		if err := rows.Scan(
			&v.ID, &v.PageID, &v.Version, &v.Title, &v.ContentMD, &v.ContentHTML,
			&changeSummary, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListVersions scan: %w", err))
		}
		if changeSummary.Valid {
			v.ChangeSummary = changeSummary.String
		}
		items = append(items, v)
	}
	if items == nil {
		items = []KnowledgePageVersion{}
	}
	return items, rows.Err()
}

// CreateVersionSnapshot 创建版本快照（UpdatePage 内部调用，事务内）。
func (s *Service) createVersionSnapshotTx(ctx context.Context, pageID int64, title, contentMD, contentHTML, changeSummary string) (*KnowledgePageVersion, error) {
	var v KnowledgePageVersion
	var cs interface{}
	if changeSummary != "" {
		cs = changeSummary
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO knowledge_page_versions (page_id, version, title, content_md, content_html, change_summary)
		VALUES ($1, (
			SELECT coalesce(max(version), 0) + 1 FROM knowledge_page_versions WHERE page_id = $1
		), $2, $3, $4, $5)
		RETURNING id, page_id, version, title, content_md, content_html, change_summary, created_by, created_at`,
		pageID, title, nullableStr(contentMD), nullableStr(contentHTML), cs).Scan(
		&v.ID, &v.PageID, &v.Version, &v.Title, &v.ContentMD, &v.ContentHTML,
		&v.ChangeSummary, &v.CreatedBy, &v.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.createVersionSnapshotTx: %w", err))
	}
	return &v, nil
}

// ==========================================================================
// 全文检索（PostgreSQL tsvector + GIN）
// ==========================================================================

// Search 全文检索（PostgreSQL 内置 tsvector）。
// 依赖：knowledge_pages_tsv_trigger（见 knowledge-migrations.sql）已把 title/content_md 同步到 tsv 列。
// 当 spaceID 为 nil 时不限定空间（搜索整个工作空间）。
func (s *Service) Search(ctx context.Context, wsID int64, keyword string, spaceID *int64, limit int) ([]KnowledgePage, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	args := []any{wsID, keyword, limit}
	spaceFilter := ""
	if spaceID != nil {
		spaceFilter = " AND space_id = $4"
		args = append(args, *spaceID)
	}
	var q = `SELECT ` + pageColumns + `
			   FROM knowledge_pages
			   WHERE workspace_id = $1 AND deleted = false
			     AND tsv @@ websearch_to_tsquery('simple', $2)
			   ` + spaceFilter + `
			   ORDER BY ts_rank(tsv, websearch_to_tsquery('simple', $2)) DESC, updated_at DESC
			   LIMIT $3`
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.Search: %w", err))
	}
	defer rows.Close()
	var out []KnowledgePage
	for rows.Next() {
		var p KnowledgePage
		if err := scanPage(rows, &p); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []KnowledgePage{}
	}
	return out, rows.Err()
}

// ==========================================================================
// 文档关联工作项
// ==========================================================================

// AddPageRelation 添加文档与工作项的关联。
func (s *Service) AddPageRelation(ctx context.Context, in AddPageRelationInput) (*KnowledgePageRelation, error) {
	// 校验文档存在
	var wsID int64
	err := s.db.QueryRow(ctx, `
		SELECT workspace_id FROM knowledge_pages
		WHERE id = $1 AND deleted = false`, in.PageID).Scan(&wsID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.AddPageRelation page: %w", err))
	}

	// 校验工作项同属一个 workspace
	var issueWsID int64
	err = s.db.QueryRow(ctx, `
		SELECT workspace_id FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w WHERE id = $1 AND deleted = false`, in.IssueID).Scan(&issueWsID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_id", Reason: "工作项不存在"})
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.AddPageRelation issue: %w", err))
	}
	if wsID != issueWsID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_id", Reason: "工作项与文档不在同一工作空间"})
	}

	relType := in.RelationType
	if relType == "" {
		relType = RelationReferenced
	}

	var rel KnowledgePageRelation
	err = s.db.QueryRow(ctx, `
		INSERT INTO knowledge_page_relations (page_id, workitem_id, relation_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (page_id, workitem_id, relation_type) DO UPDATE SET page_id = EXCLUDED.page_id
		RETURNING id, page_id, workitem_id, relation_type, created_at`,
		in.PageID, in.IssueID, relType).Scan(
		&rel.ID, &rel.PageID, &rel.IssueID, &rel.RelationType, &rel.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.AddPageRelation: %w", err))
	}
	return &rel, nil
}

// RemovePageRelation 移除文档与工作项的关联。
func (s *Service) RemovePageRelation(ctx context.Context, id, pageID, wsID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM knowledge_page_relations
		WHERE id = $1 AND page_id = $2 AND page_id IN (
			SELECT id FROM knowledge_pages WHERE id = $2 AND workspace_id = $3 AND deleted = false
		)`,
		id, pageID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("knowledge.RemovePageRelation: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ListPageRelations 列出文档的全部关联。
func (s *Service) ListPageRelations(ctx context.Context, wsID, pageID int64) ([]KnowledgePageRelation, error) {
	// 校验文档属于该 workspace
	if _, err := s.GetPage(ctx, pageID, wsID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, page_id, workitem_id, relation_type, created_at
		FROM knowledge_page_relations
		WHERE page_id = $1
		ORDER BY created_at DESC`,
		pageID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListPageRelations: %w", err))
	}
	defer rows.Close()

	var items []KnowledgePageRelation
	for rows.Next() {
		var r KnowledgePageRelation
		if err := rows.Scan(&r.ID, &r.PageID, &r.IssueID, &r.RelationType, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("knowledge.ListPageRelations scan: %w", err))
		}
		items = append(items, r)
	}
	if items == nil {
		items = []KnowledgePageRelation{}
	}
	return items, rows.Err()
}

// ==========================================================================
// 辅助函数
// ==========================================================================

// buildSpaceUpdateSet 动态构建空间 SET 子句。
func buildSpaceUpdateSet(in UpdateSpaceInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1

	if in.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", arg))
		args = append(args, *in.Name)
		arg++
	}
	if in.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", arg))
		args = append(args, *in.Description)
		arg++
	}
	if in.DefaultPermission != nil {
		sets = append(sets, fmt.Sprintf("default_permission = $%d", arg))
		args = append(args, string(*in.DefaultPermission))
		arg++
	}
	if in.IsPrivate != nil {
		sets = append(sets, fmt.Sprintf("is_private = $%d", arg))
		args = append(args, *in.IsPrivate)
		arg++
	}
	if in.CoverImage != nil {
		sets = append(sets, fmt.Sprintf("cover_image = $%d", arg))
		args = append(args, *in.CoverImage)
		arg++
	}

	return sets, args
}

// buildPageUpdateSet 动态构建文档 SET 子句。
func buildPageUpdateSet(in UpdatePageInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1

	if in.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", arg))
		args = append(args, *in.Title)
		arg++
	}
	if in.ContentMD != nil {
		sets = append(sets, fmt.Sprintf("content_md = $%d", arg))
		args = append(args, *in.ContentMD)
		arg++
	}
	if in.ContentHTML != nil {
		sets = append(sets, fmt.Sprintf("content_html = $%d", arg))
		args = append(args, *in.ContentHTML)
		arg++
	}
	if in.ParentID != nil {
		sets = append(sets, fmt.Sprintf("parent_id = $%d", arg))
		args = append(args, *in.ParentID)
		arg++
	}
	if in.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", arg))
		args = append(args, string(*in.Status))
		arg++
	}
	if in.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", arg))
		args = append(args, *in.SortOrder)
		arg++
	}
	if in.IsPinned != nil {
		sets = append(sets, fmt.Sprintf("is_pinned = $%d", arg))
		args = append(args, *in.IsPinned)
		arg++
	}
	if in.IsFeatured != nil {
		sets = append(sets, fmt.Sprintf("is_featured = $%d", arg))
		args = append(args, *in.IsFeatured)
		arg++
	}

	return sets, args
}

// getChangeSummary 返回变更摘要字符串（去空格，空时返回 ""）。
func getChangeSummary(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// nullableStr 空字符串转为 nil（落库 NULL）。
func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// scanSpace 从行扫描空间实体。
func scanSpace(row pgx.Row, sp *KnowledgeSpace) error {
	var projectID sql.NullInt64
	var ownerID sql.NullInt64
	var description sql.NullString
	var coverImage sql.NullString
	var deletedAt sql.NullTime

	err := row.Scan(
		&sp.ID, &sp.WorkspaceID, &projectID, &sp.Name, &sp.Slug, &description,
		&ownerID, &sp.DefaultPermission, &sp.IsPrivate, &coverImage,
		&sp.CreatedAt, &sp.UpdatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if projectID.Valid {
		v := projectID.Int64
		sp.ProjectID = &v
	}
	if ownerID.Valid {
		v := ownerID.Int64
		sp.OwnerID = &v
	}
	if description.Valid {
		sp.Description = description.String
	}
	if coverImage.Valid {
		sp.CoverImage = coverImage.String
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		sp.DeletedAt = &t
	}
	return nil
}

// scanPage 从行扫描文档实体。
func scanPage(row pgx.Row, p *KnowledgePage) error {
	var parentID sql.NullInt64
	var path sql.NullString
	var contentMD sql.NullString
	var contentHTML sql.NullString
	var createdBy sql.NullInt64
	var updatedBy sql.NullInt64
	var deletedAt sql.NullTime

	err := row.Scan(
		&p.ID, &p.WorkspaceID, &p.SpaceID, &parentID, &p.Lft, &p.Rgt, &p.Depth,
		&p.Title, &path, &contentMD, &contentHTML, &p.Version, &p.Status,
		&p.SortOrder, &p.IsPinned, &p.IsFeatured, &p.ViewCount,
		&createdBy, &updatedBy, &p.CreatedAt, &p.UpdatedAt, &deletedAt)
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
	if path.Valid {
		p.Path = path.String
	}
	if contentMD.Valid {
		p.ContentMD = contentMD.String
	}
	if contentHTML.Valid {
		p.ContentHTML = contentHTML.String
	}
	if createdBy.Valid {
		v := createdBy.Int64
		p.CreatedBy = &v
	}
	if updatedBy.Valid {
		v := updatedBy.Int64
		p.UpdatedBy = &v
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	return nil
}
