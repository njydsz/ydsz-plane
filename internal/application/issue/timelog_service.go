// Package issue — 工时记录。
//
// 记录用户在某工作项上花费的耗时（分钟），同步更新 task.actual_effort / remaining_effort 汇总字段。
// 所有写操作均通过 withTx() 在同一事务内完成「插入记录 + 更新汇总」，保证统计一致性。
// 数据按工作项类型存储于 task_timelogs / requirement_timelogs / defect_timelogs。
package issue

import (
	"context"
	"errors"
	"fmt"
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

// CreateTimeLogInput 添加入工时的入参。
type CreateTimeLogInput struct {
	WorkspaceID     int64     // 工作空间 ID（RLS 隔离）。
	ProjectID       int64     // 项目 ID。
	IssueID         int64     // 工作项 ID。
	UserID          int64     // 记录人 user_id。
	SpentDate       time.Time // 花费日期（仅日期部分有意义）。
	DurationMinutes int       // 持续分钟数，取值 [1, 1440]。
	Description     string    // 工作内容描述（可选）。
}

// Create 添加工时记录（同步更新 task.actual_effort, remaining_effort）。
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
		// 按类型写入对应分表（雪花 ID 全局唯一，仅一表命中）
		var inserted bool
		for _, tbl := range []string{"task", "requirement", "defect"} {
			idCol := tbl + "_id"
			err := tx.QueryRow(ctx, `
				INSERT INTO `+tbl+`_timelogs (workspace_id, project_id, `+idCol+`, user_id, spent_date, duration_minutes, description)
				VALUES ($1,$2,$3,$4,$5,$6,$7)
				RETURNING id, created_at, updated_at`,
				in.WorkspaceID, in.ProjectID, in.IssueID, in.UserID,
				in.SpentDate, in.DurationMinutes, in.Description).Scan(
				&tl.ID, &tl.CreatedAt, &tl.UpdatedAt)
			if err == nil {
				inserted = true
				break
			}
		}
		if !inserted {
			return errs.ErrNotFound
		}

		tl.WorkspaceID = in.WorkspaceID
		tl.ProjectID = in.ProjectID
		tl.IssueID = in.IssueID
		tl.UserID = in.UserID
		tl.SpentDate = in.SpentDate
		tl.DurationMinutes = in.DurationMinutes
		tl.Description = in.Description

		// 更新 issue 汇总
		_, err := tx.Exec(ctx, `
			UPDATE task SET
				actual_effort = coalesce(actual_effort,0) + $1 / 60.0,
				remaining_effort = greatest(coalesce(remaining_effort,0) - $1 / 60.0, 0),
				updated_at = now()
			WHERE id = $2 AND deleted = false`,
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
		SELECT id, workspace_id, project_id, workitem_id AS issue_id, user_id, spent_date, duration_minutes, description, created_at, updated_at
		FROM (
		    SELECT id, workspace_id, project_id, task_id AS workitem_id, user_id, spent_date, duration_minutes, description, created_at, updated_at
		    FROM task_timelogs WHERE task_id = $1 AND workspace_id = $2 AND deleted = false
		    UNION ALL
		    SELECT id, workspace_id, project_id, requirement_id, user_id, spent_date, duration_minutes, description, created_at, updated_at
		    FROM requirement_timelogs WHERE requirement_id = $1 AND workspace_id = $2 AND deleted = false
		    UNION ALL
		    SELECT id, workspace_id, project_id, defect_id, user_id, spent_date, duration_minutes, description, created_at, updated_at
		    FROM defect_timelogs WHERE defect_id = $1 AND workspace_id = $2 AND deleted = false
		) tl
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
	_ = s.db.QueryRow(ctx, `
		SELECT count(*) FROM (
		    SELECT 1 FROM task_timelogs WHERE task_id = $1 AND workspace_id = $2 AND deleted = false
		    UNION ALL SELECT 1 FROM requirement_timelogs WHERE requirement_id = $1 AND workspace_id = $2 AND deleted = false
		    UNION ALL SELECT 1 FROM defect_timelogs WHERE defect_id = $1 AND workspace_id = $2 AND deleted = false
		) t`, issueID, wsID).Scan(&total)

	return logs, total, rows.Err()
}

// Update 更新工时记录（差值回写汇总）。
func (s *TimeLogService) Update(ctx context.Context, wsID, logID int64, durationMinutes int, description string, spentDate time.Time) (*TimeLog, error) {
	if durationMinutes <= 0 || durationMinutes > 1440 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "duration_minutes", Reason: "工时必须介于 1-1440 分钟（24h）"})
	}

	var tl TimeLog
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		// 定位所属分表并读取当前值
		var issueID int64
		var oldMinutes int
		tbl := locateTimelogTable(ctx, tx, logID, wsID)
		if tbl == "" {
			return errs.ErrNotFound
		}
		idCol := tbl + "_id"
		err := tx.QueryRow(ctx, `
			SELECT `+idCol+`, duration_minutes FROM `+tbl+`_timelogs WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
			logID, wsID).Scan(&issueID, &oldMinutes)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrNotFound
			}
			return err
		}

		// 更新记录
		err = tx.QueryRow(ctx, `
			UPDATE `+tbl+`_timelogs
			SET duration_minutes = $1, description = $2, spent_date = $3, updated_at = now()
			WHERE id = $4 AND workspace_id = $5 AND deleted = false
			RETURNING id, workspace_id, project_id, `+idCol+`, user_id, spent_date, duration_minutes, description, created_at, updated_at`,
			durationMinutes, description, spentDate, logID, wsID).
			Scan(&tl.ID, &tl.WorkspaceID, &tl.ProjectID, &tl.IssueID, &tl.UserID,
				&tl.SpentDate, &tl.DurationMinutes, &tl.Description, &tl.CreatedAt, &tl.UpdatedAt)
		if err != nil {
			return err
		}

		// 差值回写 issue 汇总
		diff := durationMinutes - oldMinutes
		_, err = tx.Exec(ctx, `
			UPDATE task SET
				actual_effort = greatest(coalesce(actual_effort,0) + $1 / 60.0, 0),
				remaining_effort = greatest(coalesce(remaining_effort,0) - $1 / 60.0, 0),
				updated_at = now()
			WHERE id = $2 AND deleted = false`, diff, issueID)
		return err
	})
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &tl, nil
}

// Delete 删除工时记录（回写汇总）。
func (s *TimeLogService) Delete(ctx context.Context, wsID, logID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		tbl := locateTimelogTable(ctx, tx, logID, wsID)
		if tbl == "" {
			return errs.ErrNotFound
		}
		idCol := tbl + "_id"
		var issueID int64
		var minutes int
		err := tx.QueryRow(ctx, `
			SELECT `+idCol+`, duration_minutes FROM `+tbl+`_timelogs WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
			logID, wsID).Scan(&issueID, &minutes)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrNotFound
			}
			return err
		}
		if _, err := tx.Exec(ctx,
			`UPDATE `+tbl+`_timelogs SET deleted = true WHERE id = $1`, logID); err != nil {
			return err
		}
		// 回写
		_, err = tx.Exec(ctx, `
			UPDATE task SET
				actual_effort = greatest(coalesce(actual_effort,0) - $1 / 60.0, 0),
				remaining_effort = coalesce(remaining_effort,0) + $1 / 60.0,
				updated_at = now()
			WHERE id = $2`, minutes, issueID)
		return err
	})
}

// locateTimelogTable 在三个分表中定位工时记录所属表。
func locateTimelogTable(ctx context.Context, tx pgx.Tx, logID, wsID int64) string {
	for _, tbl := range []string{"task", "requirement", "defect"} {
		var found int64
		if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s_timelogs WHERE id = $1 AND workspace_id = $2`, tbl), logID, wsID).Scan(&found); err == nil {
			return tbl
		}
	}
	return ""
}

// withTx 在事务内执行 fn，并自动设置 RLS app.workspace_id。
//
// 提交由调用方在 fn 中返回 nil 时自动完成；fn 返回 error 时回滚。
// 统一封装以避免各方法重复写 set_config + Begin + Rollback 样板。
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
