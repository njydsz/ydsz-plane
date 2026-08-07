// Package issue — 工作项活动日志。
//
// 提供工作项维度时间线查询（谁在什么时候改了什么字段）与行内写入。
// 写入由各域服务通过 appendActivity() 调用，不对外暴露为 HTTP handler。
package issue

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ActivityService 工作项活动日志查询 + 写入。
type ActivityService struct {
	db *pgxpool.Pool
}

// NewActivityService 创建活动服务。
func NewActivityService(db *pgxpool.Pool) *ActivityService {
	return &ActivityService{db: db}
}

// ListByIssue 获取工作项的活动日志时间线。
//
// 返回 (activities, total, error)：total 为满足过滤条件的总行数（用于分页 UI 展示）。
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
		var field, oldVal, newVal *string
		var actorID *int64
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.ProjectID, &a.IssueID, &a.Verb,
			&field, &oldVal, &newVal, &a.OldRef, &a.NewRef,
			&actorID, &a.ActorEmail, &a.ActorName, &a.CreatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		a.Field = field
		a.OldValue = oldVal
		a.NewValue = newVal
		a.ActorID = actorID
		activities = append(activities, a)
	}

	var total int64
	_ = s.db.QueryRow(ctx,
		`SELECT count(*) FROM issue_activities WHERE issue_id = $1 AND workspace_id = $2`,
		issueID, wsID).Scan(&total)

	return activities, total, rows.Err()
}

// appendActivity 追加活动日志条目（内部 helper，write-only）。
//
// 仅在事务回调内调用；失败时静默吞掉错误（不影响主业务），
// 调用方应自行判断是否需要严谨审计。
//
// verb 字段取值约定（issue.*）：
//   - issue.create / issue.update / issue.delete
//   - issue.state_change / issue.assignee_change / issue.label_change
func (s *ActivityService) appendActivity(ctx context.Context, issueID, wsID, projectID int64,
	verb string, actorID *int64, oldVal, newVal *string) {
	_, _ = s.db.Exec(ctx, `
		INSERT INTO issue_activities (workspace_id, project_id, issue_id, verb, actor_id, old_value, new_value)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		wsID, projectID, issueID, verb, actorID, oldVal, newVal)
}
