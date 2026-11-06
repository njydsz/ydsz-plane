// Package pagecomment — 页面评论与行内评论应用服务。
//
// 职责：
//   - 页面级 / 行内评论的创建、编辑、软删除；
//   - 线程回复与线程解决 / 取消解决；
//   - 评论表情回应的添加 / 移除 / 列表；
//   - 从纯文本内容中提取 @username 提及（mentions）。
package pagecomment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 页面评论应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建页面评论服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrCommentNotFound 评论不存在或已删除。
	ErrCommentNotFound = errors.New("page comment not found")
	// ErrReactionNotFound 表情回应不存在。
	ErrReactionNotFound = errors.New("page comment reaction not found")
	// ErrNotCommentAuthor 只有评论作者本人可编辑 / 删除该评论。
	ErrNotCommentAuthor = errors.New("only the comment author can modify it")
)

// commentColumns 与 page_comments 表列一一对应（List/Get 共用）。
const commentColumns = `id, workspace_id, page_id, page_type, parent_id, author_id,
	content, block_id, quote_text, resolved_at, resolved_by, mentions,
	hidden, moderated_at, moderated_by, moderation_reason,
	is_edited, created_at, updated_at, deleted_at`

// mentionPattern 匹配纯文本中的 @username 提及。
//
// 允许字母 / 数字 / 下划线 / 点 / 连字符；@ 前必须是行首或空白，避免误伤邮箱地址。
var mentionPattern = regexp.MustCompile(`(?:^|\s)@([A-Za-z0-9_.\-]+)`)

// CreateComment 创建评论（支持回复与行内定位）。
//
// 校验：
//   - page_type 必须是 project / knowledge；
//   - content 不能为空；
//   - 若指定 parent_id，父评论必须存在且属于同一页面（防止跨页面串线程）；
//   - 若指定 block_id，则视为行内评论，quote_text 可选。
func (s *Service) CreateComment(ctx context.Context, in CreateCommentInput) (*PageComment, error) {
	if !in.PageType.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_type", Reason: "页面类型必须是 project 或 knowledge"})
	}
	if strings.TrimSpace(in.Content) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "评论内容不能为空"})
	}
	if in.PageID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_id", Reason: "页面 ID 无效"})
	}

	// 父评论校验：必须存在且属于同一页面 + 同一 page_type
	if in.ParentID != nil && *in.ParentID > 0 {
		var parentPageID int64
		var parentPageType string
		err := s.db.QueryRow(ctx, `
			SELECT page_id, page_type FROM page_comments
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
			*in.ParentID, in.WorkspaceID).Scan(&parentPageID, &parentPageType)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "父评论不存在"})
			}
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.CreateComment parent: %w", err))
		}
		if parentPageID != in.PageID || parentPageType != string(in.PageType) {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "父评论与当前页面不匹配"})
		}
	}

	// 提及提取：优先使用入参提供的 mentions（前端已解析），否则从文本提取。
	mentions := in.Mentions
	if len(mentions) == 0 {
		resolved, err := s.ResolveMentionNames(ctx, in.WorkspaceID, in.Content)
		if err != nil {
			return nil, err
		}
		mentions = resolved
	}
	if mentions == nil {
		mentions = []int64{}
	}

	var c PageComment
	err := scanComment(s.db.QueryRow(ctx, `
		INSERT INTO page_comments
			(id, workspace_id, page_id, page_type, parent_id, author_id,
			 content, block_id, quote_text, mentions)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+commentColumns,
		genID(), in.WorkspaceID, in.PageID, string(in.PageType), in.ParentID, in.AuthorID,
		in.Content, nullIfEmpty(in.BlockID), nullIfEmpty(in.QuoteText), mentions), &c)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.CreateComment: %w", err))
	}

	s.fillAuthor(ctx, &c)
	return &c, nil
}

// ListComments 列出页面下的评论（扁平列表，前端按 parent_id 组装线程）。
//
// 排序：根评论按 created_at 升序；回复紧跟其后（同样按 created_at 升序），
// 便于前端一次遍历构建树。已解决线程排在未解决之后。
func (s *Service) ListComments(ctx context.Context, opts ListCommentsOptions) ([]PageComment, error) {
	if !opts.PageType.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_type", Reason: "页面类型必须是 project 或 knowledge"})
	}
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(ctx, `
		SELECT `+commentColumns+`
		FROM page_comments
		WHERE workspace_id = $1 AND page_id = $2 AND page_type = $3 AND deleted_at IS NULL
		  AND ($4 = TRUE OR hidden = FALSE)
		  AND ($5 = FALSE OR hidden = TRUE)
		ORDER BY (resolved_at IS NOT NULL) ASC, parent_id NULLS FIRST, created_at ASC
		LIMIT $6 OFFSET $7`,
		opts.WorkspaceID, opts.PageID, string(opts.PageType), opts.IncludeHidden, opts.HiddenOnly, limit, offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ListComments: %w", err))
	}
	defer rows.Close()

	out := make([]PageComment, 0)
	for rows.Next() {
		var c PageComment
		if err := scanComment(rows, &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ListComments rows: %w", err))
	}

	// 批量填充作者信息
	s.fillAuthors(ctx, out)
	return out, nil
}

// GetComment 获取单条评论（含作者信息）。
func (s *Service) GetComment(ctx context.Context, commentID int64) (*PageComment, error) {
	var c PageComment
	if err := scanComment(s.db.QueryRow(ctx, `
		SELECT `+commentColumns+`
		FROM page_comments
		WHERE id = $1 AND deleted_at IS NULL`, commentID), &c); err != nil {
		return nil, err
	}
	s.fillAuthor(ctx, &c)
	return &c, nil
}

// UpdateComment 编辑评论内容（仅作者本人可编辑）。
func (s *Service) UpdateComment(ctx context.Context, commentID, userID int64, in UpdateCommentInput) (*PageComment, error) {
	if in.Content == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "评论内容不能为空"})
	}
	if strings.TrimSpace(*in.Content) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "评论内容不能为空"})
	}

	// 先取当前评论以拿到 workspace_id（用于提及解析的作用域收敛）。
	existing, err := s.GetComment(ctx, commentID)
	if err != nil {
		return nil, err
	}

	mentions := in.Mentions
	if mentions == nil {
		resolved, err := s.ResolveMentionNames(ctx, existing.WorkspaceID, *in.Content)
		if err != nil {
			return nil, err
		}
		mentions = resolved
	}
	if mentions == nil {
		mentions = []int64{}
	}

	var c PageComment
	err = scanComment(s.db.QueryRow(ctx, `
		UPDATE page_comments
		SET content = $3, mentions = $4, is_edited = true, updated_at = now()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL
		RETURNING `+commentColumns,
		commentID, userID, *in.Content, mentions), &c)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 区分「不存在」与「非作者」：先查是否存在
			var exists bool
			_ = s.db.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM page_comments WHERE id = $1 AND deleted_at IS NULL)`,
				commentID).Scan(&exists)
			if !exists {
				return nil, ErrCommentNotFound
			}
			return nil, ErrNotCommentAuthor
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.UpdateComment: %w", err))
	}

	s.fillAuthor(ctx, &c)
	return &c, nil
}

// DeleteComment 软删除评论（仅作者本人可删除）。
func (s *Service) DeleteComment(ctx context.Context, commentID, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE page_comments SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`,
		commentID, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.DeleteComment: %w", err))
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		_ = s.db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM page_comments WHERE id = $1 AND deleted_at IS NULL)`,
			commentID).Scan(&exists)
		if !exists {
			return ErrCommentNotFound
		}
		return ErrNotCommentAuthor
	}
	return nil
}

// ModerateComment 审核评论：隐藏 / 取消隐藏。
//
// action 支持 "hide" | "unhide"；reason 为审核原因（可选）。
// 该操作由具备 comment:moderate 权限的管理员执行，因此不受作者身份限制。
func (s *Service) ModerateComment(ctx context.Context, commentID, moderatorID int64, action string, reason string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	var hidden bool
	switch action {
	case "hide":
		hidden = true
	case "unhide":
		hidden = false
	default:
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "action", Reason: "action 必须是 hide 或 unhide"})
	}

	tag, err := s.db.Exec(ctx, `
		UPDATE page_comments
		SET hidden = $3,
		    moderated_at = now(),
		    moderated_by = $2,
		    moderation_reason = COALESCE(NULLIF($4, ''), moderation_reason),
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`,
		commentID, moderatorID, hidden, nullIfEmpty(reason))
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ModerateComment: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// ResolveThread 将评论所在线程标记为已解决。
//
// 线程的根评论（parent_id IS NULL）为目标；若传入的是回复，则向上追溯到根评论。
func (s *Service) ResolveThread(ctx context.Context, commentID, userID int64) error {
	return s.setResolved(ctx, commentID, userID, true)
}

// UnresolveThread 取消线程的解决状态。
func (s *Service) UnresolveThread(ctx context.Context, commentID, userID int64) error {
	return s.setResolved(ctx, commentID, userID, false)
}

// setResolved 设置线程解决状态。
func (s *Service) setResolved(ctx context.Context, commentID, userID int64, resolved bool) error {
	rootID, err := s.threadRootID(ctx, commentID)
	if err != nil {
		return err
	}

	var tag pgconn.CommandTag
	if resolved {
		tag, err = s.db.Exec(ctx, `
			UPDATE page_comments
			SET resolved_at = now(), resolved_by = $2, updated_at = now()
			WHERE id = $1 AND deleted_at IS NULL`,
			rootID, userID)
	} else {
		tag, err = s.db.Exec(ctx, `
			UPDATE page_comments
			SET resolved_at = NULL, resolved_by = 0, updated_at = now()
			WHERE id = $1 AND deleted_at IS NULL`,
			rootID)
	}
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.setResolved: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// threadRootID 返回评论所在线程的根评论 ID。
func (s *Service) threadRootID(ctx context.Context, commentID int64) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		WITH RECURSIVE thread AS (
			SELECT id, parent_id, 0 AS depth
			FROM page_comments WHERE id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT c.id, c.parent_id, t.depth + 1
			FROM page_comments c
			JOIN thread t ON c.id = t.parent_id
			WHERE c.deleted_at IS NULL AND t.depth < 32
		)
		SELECT id FROM thread WHERE parent_id IS NULL LIMIT 1`,
		commentID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrCommentNotFound
		}
		return 0, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.threadRootID: %w", err))
	}
	return id, nil
}

// AddReaction 为评论添加表情回应（同一用户同一 emoji 幂等）。
func (s *Service) AddReaction(ctx context.Context, commentID, userID int64, emoji string) (*PageCommentReaction, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "emoji", Reason: "表情标识不能为空"})
	}
	if len([]rune(emoji)) > 8 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "emoji", Reason: "表情标识过长"})
	}

	var r PageCommentReaction
	err := s.db.QueryRow(ctx, `
		INSERT INTO page_comment_reactions (id, comment_id, user_id, emoji)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (comment_id, user_id, emoji) DO UPDATE SET emoji = EXCLUDED.emoji
		RETURNING id, comment_id, user_id, emoji, created_at`,
		genID(), commentID, userID, emoji).Scan(
		&r.ID, &r.CommentID, &r.UserID, &r.Emoji, &r.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.AddReaction: %w", err))
	}
	return &r, nil
}

// RemoveReaction 移除用户对评论的指定表情回应。
func (s *Service) RemoveReaction(ctx context.Context, commentID, userID int64, emoji string) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM page_comment_reactions
		WHERE comment_id = $1 AND user_id = $2 AND emoji = $3`,
		commentID, userID, emoji)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.RemoveReaction: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrReactionNotFound
	}
	return nil
}

// ListReactions 列出评论的全部表情回应。
func (s *Service) ListReactions(ctx context.Context, commentID int64) ([]PageCommentReaction, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, comment_id, user_id, emoji, created_at
		FROM page_comment_reactions
		WHERE comment_id = $1
		ORDER BY created_at ASC`, commentID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ListReactions: %w", err))
	}
	defer rows.Close()

	out := make([]PageCommentReaction, 0)
	for rows.Next() {
		var r PageCommentReaction
		if err := rows.Scan(&r.ID, &r.CommentID, &r.UserID, &r.Emoji, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ExtractMentionNames 从纯文本内容中提取 @username 提及（去重后的原始用户名列表）。
//
// 说明：
//   - 匹配规则见 mentionPattern（@ 前必须是行首或空白）；
//   - 末尾的句点会被剔除，避免 "@alice." 被误识别为用户名的一部分；
//   - 返回的用户名未经 DB 校验，供前端展示或日志使用。
func ExtractMentionNames(content string) []string {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimSuffix(m[1], ".")
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// ResolveMentionNames 将 @username 列表解析为工作空间内用户 ID。
func (s *Service) ResolveMentionNames(ctx context.Context, workspaceID int64, content string) ([]int64, error) {
	names := ExtractMentionNames(content)
	if len(names) == 0 {
		return nil, nil
	}

	// workspaceID 为 0 时不限制工作空间（内部编辑场景）。
	if workspaceID > 0 {
		rows, err := s.db.Query(ctx, `
			SELECT DISTINCT u.id FROM users u
			JOIN workspace_members wm ON wm.user_id = u.id AND wm.workspace_id = $1
			WHERE u.deleted = false AND u.is_active
			  AND lower(u.display_name) = ANY($2)`,
			workspaceID, lowerAll(names))
		if err != nil {
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ResolveMentionNames: %w", err))
		}
		defer rows.Close()

		out := make([]int64, 0)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, errs.ErrInternal.Wrap(err)
			}
			out = append(out, id)
		}
		return out, rows.Err()
	}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT id FROM users
		WHERE deleted = false AND is_active AND lower(display_name) = ANY($1)`,
		lowerAll(names))
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("pagecomment.ResolveMentionNames: %w", err))
	}
	defer rows.Close()

	out := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// lowerAll 将字符串切片全部转为小写。
func lowerAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}

// fillAuthor 填充评论作者信息。
func (s *Service) fillAuthor(ctx context.Context, c *PageComment) {
	if c == nil || c.AuthorID == 0 {
		return
	}
	var name, avatar string
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(display_name,''), COALESCE(avatar_url,'')
		FROM users WHERE id = $1 AND deleted = false`,
		c.AuthorID).Scan(&name, &avatar); err == nil {
		c.AuthorName = name
		c.AuthorAvatar = avatar
	}
}

// fillAuthors 批量填充评论作者信息（按 author_id 去重查询）。
func (s *Service) fillAuthors(ctx context.Context, comments []PageComment) {
	if len(comments) == 0 {
		return
	}
	ids := make(map[int64]struct{}, len(comments))
	for _, c := range comments {
		if c.AuthorID != 0 {
			ids[c.AuthorID] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return
	}
	idList := make([]int64, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, COALESCE(display_name,''), COALESCE(avatar_url,'')
		FROM users WHERE id = ANY($1) AND deleted = false`, idList)
	if err != nil {
		return
	}
	defer rows.Close()

	info := make(map[int64]struct{ name, avatar string }, len(idList))
	for rows.Next() {
		var id int64
		var name, avatar string
		if err := rows.Scan(&id, &name, &avatar); err != nil {
			continue
		}
		info[id] = struct{ name, avatar string }{name: name, avatar: avatar}
	}
	for i := range comments {
		if v, ok := info[comments[i].AuthorID]; ok {
			comments[i].AuthorName = v.name
			comments[i].AuthorAvatar = v.avatar
		}
	}
}

// scanComment 从 pgx 行读取一条 PageComment。
func scanComment(row pgx.Row, c *PageComment) error {
	var parentID sql.NullInt64
	var blockID, quoteText, moderationReason sql.NullString
	var resolvedAt, moderatedAt, deletedAt sql.NullTime
	err := row.Scan(
		&c.ID, &c.WorkspaceID, &c.PageID, &c.PageType, &parentID, &c.AuthorID,
		&c.Content, &blockID, &quoteText, &resolvedAt, &c.ResolvedBy, &c.Mentions,
		&c.Hidden, &moderatedAt, &c.ModeratedBy, &moderationReason,
		&c.IsEdited, &c.CreatedAt, &c.UpdatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCommentNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid {
		v := parentID.Int64
		c.ParentID = &v
	}
	if blockID.Valid {
		c.BlockID = blockID.String
	}
	if quoteText.Valid {
		c.QuoteText = quoteText.String
	}
	if resolvedAt.Valid {
		t := resolvedAt.Time
		c.ResolvedAt = &t
	}
	if moderatedAt.Valid {
		t := moderatedAt.Time
		c.ModeratedAt = &t
	}
	if moderationReason.Valid {
		c.ModerationReason = moderationReason.String
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		c.DeletedAt = &t
	}
	return nil
}

// nullIfEmpty 空字符串转为 nil（落库 NULL），否则原样返回。
func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
