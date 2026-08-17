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

// WorkloadHeatmapEntry 单成员单日工时汇总。
type WorkloadHeatmapEntry struct {
	UserID          int64   `json:"user_id"`
	SpentDate       string  `json:"spent_date"`
	TotalMinutes    int     `json:"total_minutes"`
	TotalHours      float64 `json:"total_hours"`
	IssueCount      int     `json:"issue_count"`      // 涉及工作项数
	LogCount        int     `json:"log_count"`        // 记录条数
}

// WorkloadSummary 项目工时汇总。
type WorkloadSummary struct {
	TotalHours    float64 `json:"total_hours"`
	TotalMembers  int     `json:"total_members"`
	TotalDays     int     `json:"total_days"`
	DailyAverage  float64 `json:"daily_average_hours"`
}

// WorkloadHeatmapData 热力图数据。
type WorkloadHeatmapData struct {
	Entries  []WorkloadHeatmapEntry `json:"entries"`
	Members  []WorkloadMember       `json:"members"`
	Summary  WorkloadSummary        `json:"summary"`
	DateFrom string                 `json:"date_from"`
	DateTo   string                 `json:"date_to"`
}

// WorkloadMember 参与工时统计的成员信息。
type WorkloadMember struct {
	UserID   int64   `json:"user_id"`
	TotalHours float64 `json:"total_hours"`
	DayCount   int     `json:"day_count"`   // 有记录的天数
}

// GetWorkloadHeatmap 获取项目在指定日期范围内的工时热力图数据（按成员 × 日期聚合）。
func (s *TimeLogService) GetWorkloadHeatmap(ctx context.Context, wsID, projectID int64, dateFrom, dateTo string) (*WorkloadHeatmapData, error) {
	rows, err := s.db.Query(ctx, `
		SELECT user_id, spent_date, SUM(duration_minutes) AS total_minutes,
		       COUNT(*) AS log_count, COUNT(DISTINCT workitem_id) AS issue_count
		FROM (
		    SELECT user_id, spent_date, duration_minutes, task_id AS workitem_id
		    FROM task_timelogs
		    WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		      AND spent_date BETWEEN $3 AND $4
		    UNION ALL
		    SELECT user_id, spent_date, duration_minutes, requirement_id
		    FROM requirement_timelogs
		    WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		      AND spent_date BETWEEN $3 AND $4
		    UNION ALL
		    SELECT user_id, spent_date, duration_minutes, defect_id
		    FROM defect_timelogs
		    WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		      AND spent_date BETWEEN $3 AND $4
		) tl
		GROUP BY user_id, spent_date
		ORDER BY user_id, spent_date`,
		wsID, projectID, dateFrom, dateTo)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var entries []WorkloadHeatmapEntry
	memberMap := make(map[int64]*WorkloadMember)
	dateSet := make(map[string]bool)
	var totalMinutes int

	for rows.Next() {
		var e WorkloadHeatmapEntry
		var spentDate time.Time
		if err := rows.Scan(&e.UserID, &spentDate, &e.TotalMinutes, &e.LogCount, &e.IssueCount); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		e.SpentDate = spentDate.Format("2006-01-02")
		e.TotalHours = float64(e.TotalMinutes) / 60.0
		totalMinutes += e.TotalMinutes
		dateSet[e.SpentDate] = true

		if m, ok := memberMap[e.UserID]; ok {
			m.TotalHours += e.TotalHours
			m.DayCount++
		} else {
			memberMap[e.UserID] = &WorkloadMember{
				UserID:     e.UserID,
				TotalHours: e.TotalHours,
				DayCount:   1,
			}
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 转换为切片并按工时降序
	members := make([]WorkloadMember, 0, len(memberMap))
	for _, m := range memberMap {
		members = append(members, *m)
	}
	// 按总工时降序排列
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			if members[j].TotalHours > members[i].TotalHours {
				members[i], members[j] = members[j], members[i]
			}
		}
	}

	totalHours := float64(totalMinutes) / 60.0
	dailyAvg := 0.0
	if len(dateSet) > 0 {
		dailyAvg = totalHours / float64(len(dateSet))
	}

	return &WorkloadHeatmapData{
		Entries:  entries,
		Members:  members,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Summary: WorkloadSummary{
			TotalHours:   totalHours,
			TotalMembers: len(memberMap),
			TotalDays:    len(dateSet),
			DailyAverage: dailyAvg,
		},
	}, nil
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
