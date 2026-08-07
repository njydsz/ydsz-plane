// Package issue — 活动日志服务（问题时间线）。
package issue

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ActivityService 活动日志查询。
type ActivityService struct {
	db *pgxpool.Pool
}

// NewActivityService 创建活动服务。
func NewActivityService(db *pgxpool.Pool) *ActivityService {
	return &ActivityService{db: db}
}

// ListByIssue 获取工作项的活动日志时间线。
func (s *ActivityService) ListByIssue(ctx context.Context, wsID, issueID int64, limit, offset int) ([]IssueActivity, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, issue_id, verb,
		       field, old_value, new_value, old_ref, new_ref,
		       actor_id, actor_email, actor_name, created_at
		FROM issue_activities
		WHERE issue_id = $1 AND workspace_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		issueID, wsID, limit, offset)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var activities []IssueActivity
	for rows.Next() {
		var a IssueActivity
		var field, oldVal, newVal pgx.NullString
		var actorID pgx.NullInt64
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.ProjectID, &a.IssueID, &a.Verb,
			&field, &oldVal, &newVal, &a.OldRef, &a.NewRef,
			&actorID, &a.ActorEmail, &a.ActorName, &a.CreatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		if field.Valid {
			a.Field = &field.String
		}
		if oldVal.Valid {
			a.OldValue = &oldVal.String
		}
		if newVal.Valid {
			a.NewValue = &newVal.String
		}
		if actorID.Valid {
			a.ActorID = &actorID.Int64
		}
		activities = append(activities, a)
	}

	var total int64
	_ = s.db.QueryRow(ctx,
		`SELECT count(*) FROM issue_activities WHERE issue_id = $1 AND workspace_id = $2`,
		issueID, wsID).Scan(&total)

	return activities, total, rows.Err()
}

// appendActivity 写活动日志（辅助）。
func (s *ActivityService) appendActivity(ctx context.Context, issueID, wsID, projectID int64,
	verb string, actorID *int64, oldVal, newVal *string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO issue_activities (workspace_id, project_id, issue_id, verb, actor_id, old_value, new_value)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		wsID, projectID, issueID, verb, actorID, oldVal, newVal)
}
