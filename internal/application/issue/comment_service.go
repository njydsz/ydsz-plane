// Package issue 工作项评论应用服务：评论的创建、编辑、删除、
// 列表查询与提及解析，并在事务内录制领域事件。
package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// CreateWithEvent 在事务中创建评论并录制领域事件（comment.created）。
// 供 IssueHandler 调用，确保评论创建与事件录制原子提交。
func (s *CommentService) CreateWithEvent(ctx context.Context, input CreateCommentInput) (*Comment, error) {
	var comment Comment
	err := s.withTx(ctx, input.WorkspaceID, func(tx pgx.Tx) error {
		mentions := input.Mentions
		if mentions == nil {
			mentions = []int64{}
		}
		if input.ContentJSON == nil {
			input.ContentJSON = json.RawMessage("{}")
		}

		err := tx.QueryRow(ctx, `
			INSERT INTO issue_comments
				(workspace_id, project_id, issue_id, content_json, content_html, content_stripped,
				 created_by, mentions, parent_id, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
			RETURNING id, workspace_id, project_id, issue_id, content_json, content_html,
				content_stripped, created_by, mentions, parent_id, is_edited, edited_at,
				created_at, updated_at`,
			input.WorkspaceID, input.ProjectID, input.IssueID,
			input.ContentJSON, input.ContentHTML, input.ContentStripped,
			input.CreatedBy, mentions, input.ParentID,
		).Scan(
			&comment.ID, &comment.WorkspaceID, &comment.ProjectID, &comment.IssueID,
			&comment.ContentJSON, &comment.ContentHTML, &comment.ContentStripped,
			&comment.CreatedBy, &comment.Mentions, &comment.ParentID,
			&comment.IsEdited, &comment.EditedAt, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("CommentService.CreateWithEvent: insert: %w", err)
		}

		// 填充创建者信息
		_ = tx.QueryRow(ctx, `SELECT display_name, COALESCE(avatar_url,'') FROM users WHERE id=$1`,
			comment.CreatedBy).Scan(&comment.CreatorName, &comment.CreatorAvatar)

		// 录制领域事件：comment.created → 通知工作项关注人
		assignees := loadAssigneesForTx(ctx, tx, input.IssueID)
		return recordCommentEvent(ctx, tx, input.WorkspaceID, input.IssueID, comment.ID,
			input.CreatedBy, comment.CreatorName, input.ContentStripped, assignees)
	})
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// withTx 事务辅助（复用 IssueService 的连接）。
func (s *CommentService) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// loadAssigneesForTx 事务内查工作项分配人。
func loadAssigneesForTx(ctx context.Context, tx pgx.Tx, issueID int64) []int64 {
	rows, err := tx.Query(ctx, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, issueID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err == nil {
			ids = append(ids, uid)
		}
	}
	return ids
}
type Comment struct {
	ID              int64           `json:"id"`
	WorkspaceID     int64           `json:"workspace_id"`
	ProjectID       int64           `json:"project_id"`
	IssueID         int64           `json:"issue_id"`
	ContentJSON     json.RawMessage `json:"content_json"`
	ContentHTML     string          `json:"content_html"`
	ContentStripped string          `json:"content_stripped"`
	CreatedBy       int64           `json:"created_by"`
	CreatorName     string          `json:"creator_name,omitempty"`
	CreatorAvatar   string          `json:"creator_avatar,omitempty"`
	Mentions        []int64         `json:"mentions"`
	ParentID        *int64          `json:"parent_id"`
	IsEdited        bool            `json:"is_edited"`
	EditedAt        *time.Time      `json:"edited_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// CreateCommentInput 创建评论入参。
type CreateCommentInput struct {
	IssueID         int64
	WorkspaceID     int64
	ProjectID       int64
	ContentJSON     json.RawMessage
	ContentHTML     string
	ContentStripped string
	CreatedBy       int64
	Mentions        []int64
	ParentID        *int64
}

// UpdateCommentInput 编辑评论入参。
type UpdateCommentInput struct {
	ContentJSON     json.RawMessage
	ContentHTML     string
	ContentStripped string
	Mentions        []int64
}

// CommentService 评论应用服务。
type CommentService struct {
	db *pgxpool.Pool
}

// NewCommentService 创建评论服务。
func NewCommentService(db *pgxpool.Pool) *CommentService {
	return &CommentService{db: db}
}

// Create 创建评论。
func (s *CommentService) Create(ctx context.Context, input CreateCommentInput) (*Comment, error) {
	mentions := input.Mentions
	if mentions == nil {
		mentions = []int64{}
	}
	if input.ContentJSON == nil {
		input.ContentJSON = json.RawMessage("{}")
	}
	// 服务端白名单清洗，防存储型 XSS（content_html 由前端 v-html 渲染）
	input.ContentHTML = sanitizeCommentHTML(input.ContentHTML)

	var c Comment
	err := s.db.QueryRow(ctx, `
		INSERT INTO issue_comments
			(workspace_id, project_id, issue_id, content_json, content_html, content_stripped,
			 created_by, mentions, parent_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
		RETURNING id, workspace_id, project_id, issue_id, content_json, content_html,
			content_stripped, created_by, mentions, parent_id, is_edited, edited_at,
			created_at, updated_at`,
		input.WorkspaceID, input.ProjectID, input.IssueID,
		input.ContentJSON, input.ContentHTML, input.ContentStripped,
		input.CreatedBy, mentions, input.ParentID,
	).Scan(
		&c.ID, &c.WorkspaceID, &c.ProjectID, &c.IssueID,
		&c.ContentJSON, &c.ContentHTML, &c.ContentStripped,
		&c.CreatedBy, &c.Mentions, &c.ParentID,
		&c.IsEdited, &c.EditedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("CommentService.Create: %w", err)
	}
	// 填充创建者名称
	_ = s.db.QueryRow(ctx, `SELECT display_name, COALESCE(avatar_url,'') FROM users WHERE id=$1`,
		c.CreatedBy).Scan(&c.CreatorName, &c.CreatorAvatar)
	return &c, nil
}

// ListByIssue 查询工作项下的所有评论。
func (s *CommentService) ListByIssue(ctx context.Context, issueID int64) ([]Comment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.workspace_id, c.project_id, c.issue_id,
			c.content_json, c.content_html, c.content_stripped,
			c.created_by, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''),
			c.mentions, c.parent_id, c.is_edited, c.edited_at,
			c.created_at, c.updated_at
		FROM issue_comments c
		LEFT JOIN users u ON u.id = c.created_by
		WHERE c.issue_id = $1
		ORDER BY c.created_at ASC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("CommentService.ListByIssue: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(
			&c.ID, &c.WorkspaceID, &c.ProjectID, &c.IssueID,
			&c.ContentJSON, &c.ContentHTML, &c.ContentStripped,
			&c.CreatedBy, &c.CreatorName, &c.CreatorAvatar,
			&c.Mentions, &c.ParentID, &c.IsEdited, &c.EditedAt,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("CommentService.ListByIssue scan: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// Update 编辑评论。
func (s *CommentService) Update(ctx context.Context, commentID, userID int64, input UpdateCommentInput) (*Comment, error) {
	if input.ContentJSON == nil {
		input.ContentJSON = json.RawMessage("{}")
	}
	mentions := input.Mentions
	if mentions == nil {
		mentions = []int64{}
	}
	// 服务端白名单清洗（与 Create 一致）
	input.ContentHTML = sanitizeCommentHTML(input.ContentHTML)

	var c Comment
	err := s.db.QueryRow(ctx, `
		UPDATE issue_comments SET
			content_json = $3, content_html = $4, content_stripped = $5,
			mentions = $6, is_edited = true, edited_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND created_by = $2
		RETURNING id, workspace_id, project_id, issue_id, content_json, content_html,
			content_stripped, created_by, mentions, parent_id, is_edited, edited_at,
			created_at, updated_at`,
		commentID, userID,
		input.ContentJSON, input.ContentHTML, input.ContentStripped, mentions,
	).Scan(
		&c.ID, &c.WorkspaceID, &c.ProjectID, &c.IssueID,
		&c.ContentJSON, &c.ContentHTML, &c.ContentStripped,
		&c.CreatedBy, &c.Mentions, &c.ParentID,
		&c.IsEdited, &c.EditedAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("CommentService.Update: %w", err)
	}
	return &c, nil
}

// Delete 删除评论（软删除级联已由 ON DELETE CASCADE 处理）。
func (s *CommentService) Delete(ctx context.Context, commentID, userID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM issue_comments WHERE id = $1 AND created_by = $2`,
		commentID, userID)
	if err != nil {
		return fmt.Errorf("CommentService.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("COMMENT.NOT_FOUND", "评论不存在或无权删除")
	}
	return nil
}
