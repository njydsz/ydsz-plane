// Package issue — 工时记录服务。
package issue

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// TimeLogService 工时记录查询 + 写入。
type TimeLogService struct {
	db *pgxpool.Pool
}

// NewTimeLogService 创建工时服务。
func NewTimeLogService(db *pgxpool.Pool) *TimeLogService {
	return &TimeLogService{db: db}
}

// CreateInput 添加入工时。
type CreateTimeLogInput struct {
	WorkspaceID     int64
	ProjectID       int64
	IssueID         int64
	UserID          int64
	SpentDate       time.Time
	DurationMinutes int
	Description     string
}

// Create 添加工时记录（同步更新 issue.actual_effort, remaining_effort）。
func (s *TimeLogService) Create(ctx context.Context, in CreateTimeLogInput) (*TimeLog, error) {
	if in.DurationMinutes <= 0 || in.DurationMinutes > 1440 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "duration_minutes", Reason: "工时必须介于 1-1440 分钟（24h）"})
	}
	if in.UserID == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "user_id", Reason: "必须指定用户"})
	}
	if in.IssueID == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_id", Reason: "必须指定工作项"})
	}

	var tl TimeLog
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO time_logs (workspace_id, project_id, issue_id, user_id, spent_date, duration_minutes, description)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			RETURNING id, created_at, updated_at`,
			in.WorkspaceID, in.ProjectID, in.IssueID, in.UserID,
			in.SpentDate, in.DurationMinutes, in.Description).Scan(
			&tl.ID, &tl.CreatedAt, &tl.UpdatedAt)
		if err != nil {
			return err
		}

		tl.WorkspaceID = in.WorkspaceID
		tl.ProjectID = in.ProjectID
		tl.IssueID = in.IssueID
		tl.UserID = in.UserID
		tl.SpentDate = in.SpentDate
		tl.DurationMinutes = in.DurationMinutes
		tl.Description = in.Description

		// 更新 issue 汇总
		_, err = tx.Exec(ctx, `
			UPDATE issues SET
				actual_effort = coalesce(actual_effort,0) + $1 / 60.0,
				remaining_effort = greatest(coalesce(remaining_effort,0) - $1 / 60.0, 0),
				updated_at = now()
			WHERE id = $2 AND deleted_at IS NULL`,
			in.DurationMinutes, in.IssueID)
		return err
	})

	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &tl, nil
}

// ListByIssue 列出工作项的工时记录。
func (s *TimeLogService) ListByIssue(ctx context.Context, wsID, issueID int64, limit, offset int) ([]TimeLog, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, issue_id, user_id, spent_date, duration_minutes, description, created_at, updated_at
		FROM time_logs
		WHERE issue_id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		ORDER BY spent_date DESC, created_at DESC
		LIMIT $3 OFFSET $4`, issueID, wsID, limit, offset)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var logs []TimeLog
	for rows.Next() {
		var tl TimeLog
		if err := rows.Scan(&tl.ID, &tl.WorkspaceID, &tl.ProjectID, &tl.IssueID, &tl.UserID,
			&tl.SpentDate, &tl.DurationMinutes, &tl.Description, &tl.CreatedAt, &tl.UpdatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		logs = append(logs, tl)
	}

	var total int64
	_ = s.db.QueryRow(ctx, `SELECT count(*) FROM time_logs WHERE issue_id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		issueID, wsID).Scan(&total)

	return logs, total, rows.Err()
}

// Delete 删除工时记录（回写汇总）。
func (s *TimeLogService) Delete(ctx context.Context, wsID, logID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		var issueID int64
		var minutes int
		err := tx.QueryRow(ctx, `
			SELECT issue_id, duration_minutes FROM time_logs WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
			logID, wsID).Scan(&issueID, &minutes)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrNotFound
			}
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE time_logs SET deleted_at = now() WHERE id = $1`, logID); err != nil {
			return err
		}
		// 回写
		_, err = tx.Exec(ctx, `
			UPDATE issues SET
				actual_effort = greatest(coalesce(actual_effort,0) - $1 / 60.0, 0),
				remaining_effort = coalesce(remaining_effort,0) + $1 / 60.0,
				updated_at = now()
			WHERE id = $2`, minutes, issueID)
		return err
	})
}

// WithTx 事务辅助。
func (s *TimeLogService) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
