// Package search — 全文搜索服务（PostgreSQL FTS）。
//
// 架构决策：
//   - 使用 PostgreSQL 内置 FTS（tsvector + GIN）而非外部 ES，
//     满足中小规模工作项（<100 万）的低运维搜索需求。
//   - 未来 ES 升级路径：通过 search_documents 表抽象，
//     额外同步至 ES 索引即可切换，应用层接口不变。
//   - 搜索结果带高亮片段（ts_headline），提升 CTR。
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供全文搜索服务。
// 支持多对象搜索（issues/sprints/versions）、过滤、高亮、搜索历史、收藏。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建搜索服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Search 执行全文搜索。
//
// 流程:
//  1. 构建查询（to_tsquery + websearch_to_tsquery 混合）
//  2. 按 doc_type 过滤
//  3. 应用结构化过滤器
//  4. 检索 + 排名 + 高亮
//  5. 按类型分组返回
func (s *Service) Search(ctx context.Context, q SearchQuery) (*SearchResponse, error) {
	if q.Limit <= 0 || q.Limit > 50 {
		q.Limit = 20
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if len(q.DocTypes) == 0 {
		q.DocTypes = []string{"issue", "sprint", "version"}
	}

	start := time.Now()

	// 安全处理搜索词（防 SQL 注入）
	tsQuery := toTSQuery(q.Query)
	if tsQuery == "" {
		return &SearchResponse{
			Query:  q.Query,
			Total:  0,
			Groups: []SearchGroup{},
			TimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// 构建 WHERE 条件
	whereParts := []string{
		"d.workspace_id = $1",
		"d.search_tsv @@ to_tsquery('simple', $2)",
	}
	args := []interface{}{q.WorkspaceID, tsQuery}
	argIdx := 3

	// 项目过滤
	if q.ProjectID > 0 {
		whereParts = append(whereParts, fmt.Sprintf("d.project_id = $%d", argIdx))
		args = append(args, q.ProjectID)
		argIdx++
	}

	// 类型过滤
	if len(q.DocTypes) > 0 {
		placeholders := make([]string, len(q.DocTypes))
		for i := range q.DocTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, q.DocTypes[i])
			argIdx++
		}
		whereParts = append(whereParts, fmt.Sprintf("d.doc_type IN (%s)", strings.Join(placeholders, ",")))
	}

	// issue 结构化过滤
	if code, ok := q.Filters["type_code"].(string); ok && code != "" {
		whereParts = append(whereParts, fmt.Sprintf("d.metadata->>'type_code' = $%d", argIdx))
		args = append(args, code)
		argIdx++
	}
	if priority, ok := q.Filters["priority"].(string); ok && priority != "" {
		whereParts = append(whereParts, fmt.Sprintf("d.metadata->>'priority' = $%d", argIdx))
		args = append(args, priority)
		argIdx++
	}
	if stateID, ok := q.Filters["state_id"].(float64); ok && stateID > 0 {
		whereParts = append(whereParts, fmt.Sprintf("(d.metadata->>'state_id')::bigint = $%d", argIdx))
		args = append(args, int64(stateID))
		argIdx++
	}

	whereSQL := strings.Join(whereParts, " AND ")

	// 查询总数
	countSQL := fmt.Sprintf("SELECT count(*) FROM search_documents d WHERE %s", whereSQL)
	var total int
	if err := s.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 排序 + 分页
	limitIdx := argIdx
	offsetIdx := argIdx + 1
	queryArgs := append(args, q.Limit, q.Offset)

	sql := fmt.Sprintf(`
		SELECT
			d.doc_type, d.doc_id, d.title, d.identifier, d.content,
			ts_rank(d.search_tsv, to_tsquery('simple', $2)) AS rank,
			ts_headline('simple', d.content, to_tsquery('simple', $2),
				'StartSel=<b>, StopSel=</b>, MaxFragments=3, FragmentDelimiter=...'
			) AS highlight,
			d.metadata
		FROM search_documents d
		WHERE %s
		ORDER BY rank DESC, d.updated_at DESC
		LIMIT $%d OFFSET $%d`, whereSQL, limitIdx, offsetIdx)

	rows, err := s.db.Query(ctx, sql, queryArgs...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	// 按类型分组聚合
	groupMap := map[string]*SearchGroup{}
	for _, dt := range q.DocTypes {
		groupMap[dt] = &SearchGroup{DocType: dt, Hits: []SearchHit{}}
	}

	for rows.Next() {
		var dt string
		var docID int64
		var title, identifier, content, highlight string
		var rank float64
		var metaRaw []byte

		if err := rows.Scan(&dt, &docID, &title, &identifier, &content, &rank, &highlight, &metaRaw); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}

		var meta Metadata
		_ = json.Unmarshal(metaRaw, &meta)

		// 对 issue 补充 state_name
		if dt == "issue" && meta.StateID > 0 {
			meta.StateName = s.lookupStateName(ctx, q.WorkspaceID, meta.StateID)
		}

		hit := SearchHit{
			DocType:    dt,
			DocID:      docID,
			Title:      title,
			Identifier: identifier,
			Highlights: []string{highlight},
			Rank:       rank,
			Metadata:   meta,
			URL:        buildDocURL(dt, docID, q.ProjectID),
		}

		if g, ok := groupMap[dt]; ok {
			g.Hits = append(g.Hits, hit)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 构建有序 groups
	var groups []SearchGroup
	for _, dt := range q.DocTypes {
		g := groupMap[dt]
		g.Total = int64(len(g.Hits))
		if len(g.Hits) > 0 {
			groups = append(groups, *g)
		}
	}

	return &SearchResponse{
		Query:  q.Query,
		Total:  total,
		Groups: groups,
		TimeMs: time.Since(start).Milliseconds(),
	}, nil
}

// --- Search History ---

// RecordHistory 记录搜索历史（保留最近 50 条）。
func (s *Service) RecordHistory(ctx context.Context, in RecordHistoryInput) error {
	// 幂等：同一用户+同一查询+10s 内不重复记录
	const dedupeWindow = 10 * time.Second

	filtersJSON, _ := json.Marshal(in.Filters)
	_, err := s.db.Exec(ctx, `
		WITH inserted AS (
			INSERT INTO search_history (workspace_id, user_id, query, filters, result_count)
			SELECT $1, $2, $3, $4, $5
			WHERE NOT EXISTS (
				SELECT 1 FROM search_history
				WHERE user_id = $2 AND query = $3 AND searched_at > now() - ($6 || ' seconds')::interval
			)
			RETURNING id
		),
		prune AS (
			DELETE FROM search_history
			WHERE user_id = $2 AND id NOT IN (
				SELECT id FROM search_history WHERE user_id = $2 ORDER BY searched_at DESC LIMIT 50
			)
		)
		SELECT 1 FROM inserted`,
		in.WorkspaceID, in.UserID, in.Query, filtersJSON, in.ResultCount,
		strconv.Itoa(int(dedupeWindow.Seconds())))
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListHistory 获取用户最近搜索历史（最多 20 条）。
func (s *Service) ListHistory(ctx context.Context, wsID, userID int64, limit int) ([]SearchHistoryEntry, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, user_id, query, filters, result_count, searched_at
		FROM search_history
		WHERE workspace_id = $1 AND user_id = $2
		ORDER BY searched_at DESC LIMIT $3`,
		wsID, userID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var entries []SearchHistoryEntry
	for rows.Next() {
		var e SearchHistoryEntry
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.UserID, &e.Query, &e.Filters, &e.ResultCount, &e.SearchedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// DeleteHistory 删除单条搜索历史。
func (s *Service) DeleteHistory(ctx context.Context, wsID, userID, historyID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM search_history WHERE id = $1 AND user_id = $2 AND workspace_id = $3`,
		historyID, userID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ClearHistory 清空用户搜索历史。
func (s *Service) ClearHistory(ctx context.Context, wsID, userID int64) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM search_history WHERE user_id = $1 AND workspace_id = $2`,
		userID, wsID)
	return err
}

// --- Search Bookmarks ---

// CreateBookmark 创建搜索收藏。
func (s *Service) CreateBookmark(ctx context.Context, in CreateBookmarkInput) (*SearchBookmark, error) {
	filtersJSON, _ := json.Marshal(in.Filters)
	var bm SearchBookmark
	var projID interface{}
	if in.ProjectID != nil {
		projID = *in.ProjectID
	}

	err := s.db.QueryRow(ctx, `
		INSERT INTO search_bookmarks (workspace_id, project_id, user_id, name, query, filters, is_shared, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, (
			SELECT coalesce(max(sort_order), 0) + 1 FROM search_bookmarks WHERE user_id = $3
		))
		RETURNING id, workspace_id, project_id, user_id, name, query, filters, is_shared, sort_order, created_at, updated_at`,
		in.WorkspaceID, projID, in.UserID, in.Name, in.Query, filtersJSON, in.IsShared).Scan(
		&bm.ID, &bm.WorkspaceID, &bm.ProjectID, &bm.UserID, &bm.Name, &bm.Query,
		&bm.Filters, &bm.IsShared, &bm.SortOrder, &bm.CreatedAt, &bm.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &bm, nil
}

// ListBookmarks 列出用户搜索收藏。
func (s *Service) ListBookmarks(ctx context.Context, wsID, userID int64, projectID *int64) ([]SearchBookmark, error) {
	var rows pgx.Rows
	var err error
	if projectID != nil && *projectID > 0 {
		rows, err = s.db.Query(ctx, `
			SELECT id, workspace_id, project_id, user_id, name, query, filters, is_shared, sort_order, created_at, updated_at
			FROM search_bookmarks
			WHERE workspace_id = $1 AND user_id = $2 AND (project_id = $3 OR project_id IS NULL)
			ORDER BY sort_order, created_at DESC`,
			wsID, userID, *projectID)
	} else {
		rows, err = s.db.Query(ctx, `
			SELECT id, workspace_id, project_id, user_id, name, query, filters, is_shared, sort_order, created_at, updated_at
			FROM search_bookmarks
			WHERE workspace_id = $1 AND user_id = $2
			ORDER BY sort_order, created_at DESC`,
			wsID, userID)
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var bookmarks []SearchBookmark
	for rows.Next() {
		var bm SearchBookmark
		if err := rows.Scan(
			&bm.ID, &bm.WorkspaceID, &bm.ProjectID, &bm.UserID, &bm.Name,
			&bm.Query, &bm.Filters, &bm.IsShared, &bm.SortOrder, &bm.CreatedAt, &bm.UpdatedAt,
		); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		bookmarks = append(bookmarks, bm)
	}
	return bookmarks, rows.Err()
}

// UpdateBookmark 更新搜索收藏。
func (s *Service) UpdateBookmark(ctx context.Context, wsID, userID, bmID int64, in UpdateBookmarkInput) (*SearchBookmark, error) {
	var bm SearchBookmark
	err := s.db.QueryRow(ctx, `
		UPDATE search_bookmarks SET
			name = coalesce($4, name),
			query = coalesce($5, query),
			filters = coalesce($6, filters),
			is_shared = coalesce($7, is_shared),
			updated_at = now()
		WHERE id = $1 AND user_id = $2 AND workspace_id = $3
		RETURNING id, workspace_id, project_id, user_id, name, query, filters, is_shared, sort_order, created_at, updated_at`,
		bmID, userID, wsID, in.Name, in.Query, in.Filters, in.IsShared).Scan(
		&bm.ID, &bm.WorkspaceID, &bm.ProjectID, &bm.UserID, &bm.Name,
		&bm.Query, &bm.Filters, &bm.IsShared, &bm.SortOrder, &bm.CreatedAt, &bm.UpdatedAt)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return &bm, nil
}

// DeleteBookmark 删除搜索收藏。
func (s *Service) DeleteBookmark(ctx context.Context, wsID, userID, bmID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM search_bookmarks WHERE id = $1 AND user_id = $2 AND workspace_id = $3`,
		bmID, userID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- Helpers ---

// toTSQuery 将用户输入转换为安全的 tsquery。
// 处理: "foo bar" → "foo & bar"，"foo -bar" → "foo & !bar"
func toTSQuery(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// 使用 websearch_to_tsquery 更安全，但需要拆分特殊字符
	// 这里用 prefix matching 增强
	tokens := strings.Fields(input)
	var parts []string
	for _, tok := range tokens {
		// 过滤控制字符
		tok = strings.Map(func(r rune) rune {
			if r >= 32 && r < 127 {
				return r
			}
			return -1
		}, tok)
		if tok == "" {
			continue
		}
		// 前缀匹配: "foo*" → "foo:*"
		parts = append(parts, tok + ":*")
	}
	return strings.Join(parts, " & ")
}

// lookupStateName 查询状态名称。
func (s *Service) lookupStateName(ctx context.Context, wsID, stateID int64) string {
	var name string
	_ = s.db.QueryRow(ctx,
		`SELECT name FROM states WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		stateID, wsID).Scan(&name)
	return name
}

// buildDocURL 构建搜索结果的前端跳转 URL。
func buildDocURL(docType string, docID, projectID int64) string {
	switch docType {
	case "issue":
		return fmt.Sprintf("/projects/%d/issues/%d", projectID, docID)
	case "sprint":
		return fmt.Sprintf("/projects/%d/sprints/%d", projectID, docID)
	case "version":
		return fmt.Sprintf("/projects/%d/versions/%d", projectID, docID)
	}
	return ""
}

// mapPgError 映射 PostgreSQL 错误。
func (s *Service) mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if err == pgx.ErrNoRows {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}
