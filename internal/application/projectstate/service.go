// Package projectstate — 项目状态与更新应用服务。
package projectstate

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrStateNotFound 项目状态记录不存在。
	ErrStateNotFound = errors.New("project state not found")
	// ErrUpdateNotFound 项目更新不存在。
	ErrUpdateNotFound = errors.New("project update not found")
)

// --- 项目状态 ---

// UpsertState 创建或更新项目状态（每项目仅一条状态记录）。
func (s *Service) UpsertState(ctx context.Context, in UpsertStateInput) (*ProjectState, error) {
	if in.ProjectID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	if in.Status == "" || !Status(in.Status).IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "invalid value"})
	}
	if in.Priority == "" {
		in.Priority = string(PriorityMedium)
	}
	if !IsValidPriority(in.Priority) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "priority", Reason: "invalid value"})
	}

	var ps ProjectState
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_states (id, project_id, status, priority, start_date, target_date, owner_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (project_id) DO UPDATE SET
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			start_date = EXCLUDED.start_date,
			target_date = EXCLUDED.target_date,
			owner_id = EXCLUDED.owner_id,
			description = EXCLUDED.description,
			updated_at = now()
		RETURNING id, project_id, status, priority, start_date, target_date, owner_id,
			COALESCE(description,''), updated_at`,
		genID(), in.ProjectID, in.Status, in.Priority, in.StartDate, in.TargetDate,
		in.OwnerID, nullIfEmpty(in.Description),
	).Scan(&ps.ID, &ps.ProjectID, &ps.Status, &ps.Priority, &ps.StartDate,
		&ps.TargetDate, &ps.OwnerID, &ps.Description, &ps.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ps, nil
}

// GetState 查询项目状态。
func (s *Service) GetState(ctx context.Context, projectID int64) (*ProjectState, error) {
	var ps ProjectState
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, status, priority, start_date, target_date, owner_id,
			COALESCE(description,''), updated_at
		FROM project_states WHERE project_id = $1`, projectID,
	).Scan(&ps.ID, &ps.ProjectID, &ps.Status, &ps.Priority, &ps.StartDate,
		&ps.TargetDate, &ps.OwnerID, &ps.Description, &ps.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrStateNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ps, nil
}

// --- 项目更新 ---

// CreateUpdate 发布项目更新。
func (s *Service) CreateUpdate(ctx context.Context, in CreateUpdateInput) (*ProjectUpdate, error) {
	if in.ProjectID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	if in.Status == "" || !Status(in.Status).IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "invalid value"})
	}
	if strings.TrimSpace(in.Content) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "required"})
	}

	var pu ProjectUpdate
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_updates
			(id, workspace_id, project_id, author_id, status, content, progress_percent, initiative_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, workspace_id, project_id, author_id, status, content,
			progress_percent, initiative_id, created_at`,
		genID(), in.WorkspaceID, in.ProjectID, in.AuthorID, in.Status, in.Content,
		in.ProgressPercent, nullIfZero(in.InitiativeID),
	).Scan(&pu.ID, &pu.WorkspaceID, &pu.ProjectID, &pu.AuthorID, &pu.Status,
		&pu.Content, &pu.ProgressPercent, &pu.InitiativeID, &pu.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &pu, nil
}

// GetUpdate 查询更新。
func (s *Service) GetUpdate(ctx context.Context, id int64) (*ProjectUpdate, error) {
	var pu ProjectUpdate
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, author_id, status, content,
			progress_percent, initiative_id, created_at
		FROM project_updates WHERE id = $1`, id,
	).Scan(&pu.ID, &pu.WorkspaceID, &pu.ProjectID, &pu.AuthorID, &pu.Status,
		&pu.Content, &pu.ProgressPercent, &pu.InitiativeID, &pu.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUpdateNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &pu, nil
}

// ListUpdates 列出项目或倡议的更新。
func (s *Service) ListUpdates(ctx context.Context, opts ListUpdatesOptions) ([]*ProjectUpdate, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, author_id, status, content,
			progress_percent, initiative_id, created_at
		FROM project_updates
		WHERE ($1 = 0 OR project_id = $1)
		  AND ($2 = 0 OR initiative_id = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		opts.ProjectID, opts.InitiativeID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*ProjectUpdate
	for rows.Next() {
		var pu ProjectUpdate
		if err := rows.Scan(&pu.ID, &pu.WorkspaceID, &pu.ProjectID, &pu.AuthorID,
			&pu.Status, &pu.Content, &pu.ProgressPercent, &pu.InitiativeID, &pu.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &pu)
	}
	return out, rows.Err()
}

// DeleteUpdate 删除更新（含反应与评论）。
func (s *Service) DeleteUpdate(ctx context.Context, id int64) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM project_update_reactions WHERE update_id = $1`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if _, err := s.db.Exec(ctx, `DELETE FROM project_update_comments WHERE update_id = $1`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	tag, err := s.db.Exec(ctx, `DELETE FROM project_updates WHERE id = $1`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUpdateNotFound
	}
	return nil
}

// --- 反应 ---

// AddReaction 添加/切换表情回应（幂等）。
func (s *Service) AddReaction(ctx context.Context, updateID, userID int64, emoji string) error {
	if emoji == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "emoji", Reason: "required"})
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO project_update_reactions (id, update_id, user_id, emoji)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (update_id, user_id, emoji) DO NOTHING`,
		genID(), updateID, userID, emoji)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// RemoveReaction 移除表情回应。
func (s *Service) RemoveReaction(ctx context.Context, updateID, userID int64, emoji string) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM project_update_reactions
		WHERE update_id = $1 AND user_id = $2 AND emoji = $3`,
		updateID, userID, emoji)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListReactions 列出更新的反应。
func (s *Service) ListReactions(ctx context.Context, updateID int64) ([]ProjectUpdateReaction, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, update_id, user_id, emoji, created_at
		FROM project_update_reactions WHERE update_id = $1 ORDER BY created_at ASC`, updateID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	var out []ProjectUpdateReaction
	for rows.Next() {
		var r ProjectUpdateReaction
		if err := rows.Scan(&r.ID, &r.UpdateID, &r.UserID, &r.Emoji, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- 评论 ---

// AddComment 添加更新评论（支持回复）。
func (s *Service) AddComment(ctx context.Context, updateID, authorID int64, content string, parentID *int64) (*ProjectUpdateComment, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "required"})
	}
	var c ProjectUpdateComment
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_update_comments (id, update_id, author_id, parent_id, content)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, update_id, author_id, parent_id, content, created_at, deleted_at`,
		genID(), updateID, authorID, parentID, content,
	).Scan(&c.ID, &c.UpdateID, &c.AuthorID, &c.ParentID, &c.Content, &c.CreatedAt, &c.DeletedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &c, nil
}

// ListComments 列出更新的评论。
func (s *Service) ListComments(ctx context.Context, updateID int64) ([]ProjectUpdateComment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, update_id, author_id, parent_id, content, created_at, deleted_at
		FROM project_update_comments
		WHERE update_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`, updateID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	var out []ProjectUpdateComment
	for rows.Next() {
		var c ProjectUpdateComment
		if err := rows.Scan(&c.ID, &c.UpdateID, &c.AuthorID, &c.ParentID, &c.Content, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DeleteComment 软删除评论（仅作者）。
func (s *Service) DeleteComment(ctx context.Context, id, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE project_update_comments SET deleted_at = now()
		WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUpdateNotFound
	}
	return nil
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullIfZero(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

var _ = time.Now
