// Package automation — ActionExecutor 的真实实现。
//
// dbActionExecutor 实现 engine.go 中声明的 ActionExecutor SPI，
// 将自动化动作落地为对 Issue / State / Notification 域服务及
// 直接 DB 查询（assignees、tech_lead、least_loaded 负载）的实际副作用。
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/application/issue"
	"github.com/njydsz/ydsz-plane/internal/application/notification"
)

// dbActionExecutor 是 ActionExecutor 的默认实现。
// 持有 db 连接池（用于复杂 SQL）+ 可选的 issue/state 域服务。
type dbActionExecutor struct {
	db       *pgxpool.Pool
	issueSvc *issue.Service
	stateSvc *issue.StateService
}

// newActionExecutor 创建基于真实 DB/域服务的 ActionExecutor。
func newActionExecutor(db *pgxpool.Pool) *dbActionExecutor {
	return &dbActionExecutor{
		db:       db,
		issueSvc: issue.NewService(db),
		stateSvc: issue.NewStateService(db),
	}
}

// compile-time assertion：dbActionExecutor 必须满足 ActionExecutor。
var _ ActionExecutor = (*dbActionExecutor)(nil)

// TransitionIssueStatus 将工作项流转到目标状态（按状态名解析 stateID）。
func (x *dbActionExecutor) TransitionIssueStatus(ctx context.Context, wsID, projectID, issueID int64, targetState string) error {
	if strings.TrimSpace(targetState) == "" {
		return fmt.Errorf("transition: empty target state")
	}
	// 状态名 → ID 索引
	index, err := x.stateSvc.GetStateNameIndex(ctx, wsID, projectID)
	if err != nil {
		return fmt.Errorf("transition: load state index: %w", err)
	}
	toStateID, ok := index[targetState]
	if !ok {
		return fmt.Errorf("transition: unknown target state %q", targetState)
	}

	// 用 issue.CreatedBy 作为流转操作者（userID 仅用于事件 actor 记录）
	iss, err := x.issueSvc.GetByID(ctx, wsID, issueID)
	if err != nil {
		return fmt.Errorf("transition: load issue: %w", err)
	}
	_, err = x.issueSvc.Transition(ctx, wsID, projectID, issueID, toStateID, iss.CreatedBy)
	if err != nil {
		return fmt.Errorf("transition: %w", err)
	}
	return nil
}

// AssignIssue 指派工作项给指定用户。
func (x *dbActionExecutor) AssignIssue(ctx context.Context, wsID, projectID, issueID int64, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("assign: invalid user_id %d", userID)
	}
	_, err := x.db.Exec(ctx, `
		INSERT INTO issue_assignees (issue_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (issue_id, user_id) DO NOTHING`, issueID, userID)
	if err != nil {
		return fmt.Errorf("assign: upsert assignee: %w", err)
	}
	// 可选审计：写 issue_activities（该表启用 RLS，需设置租户上下文）
	newValStr := strconv.FormatInt(userID, 10)
	x.appendActivity(ctx, wsID, projectID, issueID, "updated",
		"assignees", nil, &newValStr)
	return nil
}

// UpdateIssueField 更新工作项字段（按 field 名映射）。
func (x *dbActionExecutor) UpdateIssueField(ctx context.Context, wsID, projectID, issueID int64, field string, value any) error {
	if strings.TrimSpace(field) == "" {
		return fmt.Errorf("update_field: missing field name")
	}
	// 乐观锁：先取当前 version
	current, err := x.issueSvc.GetByID(ctx, wsID, issueID)
	if err != nil {
		return fmt.Errorf("update_field: load issue: %w", err)
	}
	valueStr := toString(value)

	switch field {
	case "name":
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			v := valueStr
			in.Name = &v
		})
	case "description_html", "description":
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			v := valueStr
			in.DescriptionHTML = &v
		})
	case "state_id":
		id, convErr := toInt64(value)
		if convErr != nil {
			return fmt.Errorf("update_field: state_id: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.StateID = &id
		})
	case "priority":
		p := issue.IssuePriority(valueStr)
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.Priority = &p
		})
	case "severity":
		sev, convErr := toInt(value)
		if convErr != nil {
			return fmt.Errorf("update_field: severity: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.Severity = &sev
		})
	case "category":
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			v := valueStr
			in.Category = &v
		})
	case "assignees":
		ids, convErr := toInt64Slice(value)
		if convErr != nil {
			return fmt.Errorf("update_field: assignees: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.Assignees = ids
		})
	case "fix_version_id":
		id, convErr := toInt64(value)
		if convErr != nil {
			return fmt.Errorf("update_field: fix_version_id: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.FixVersionID = &id
		})
	case "found_version_id":
		id, convErr := toInt64(value)
		if convErr != nil {
			return fmt.Errorf("update_field: found_version_id: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.FoundVersionID = &id
		})
	case "release_version_id":
		id, convErr := toInt64(value)
		if convErr != nil {
			return fmt.Errorf("update_field: release_version_id: %w", convErr)
		}
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {
			in.ReleaseVersionID = &id
		})
	case "point", "estimate_points":
		pt, convErr := toInt(value)
		if convErr != nil {
			return fmt.Errorf("update_field: point: %w", convErr)
		}
		return x.updatePoint(ctx, wsID, issueID, current.Version, pt)
	case "started_at":
		t, convErr := toTime(value)
		if convErr != nil {
			return fmt.Errorf("update_field: started_at: %w", convErr)
		}
		return x.updateTimeCol(ctx, wsID, issueID, current.Version, "started_at", t)
	case "completed_at":
		t, convErr := toTime(value)
		if convErr != nil {
			return fmt.Errorf("update_field: completed_at: %w", convErr)
		}
		return x.updateTimeCol(ctx, wsID, issueID, current.Version, "completed_at", t)
	case "target_date":
		t, convErr := toTime(value)
		if convErr != nil {
			return fmt.Errorf("update_field: target_date: %w", convErr)
		}
		return x.updateTimeCol(ctx, wsID, issueID, current.Version, "target_date", t)
	default:
		// 未知字段：为保持确定性不 panic，直接走 Update（无实际 SET 子句）。
		return x.updateViaInput(ctx, wsID, issueID, current.Version, func(in *issue.UpdateIssueInput) {})
	}
}

// updateViaInput 通过 IssueService.Update 更新受支持的字段（乐观锁 version 校验）。
func (x *dbActionExecutor) updateViaInput(ctx context.Context, wsID, issueID int64, version int,
	mutate func(*issue.UpdateIssueInput)) error {
	in := issue.UpdateIssueInput{Version: version}
	mutate(&in)
	if _, err := x.issueSvc.Update(ctx, wsID, issueID, in); err != nil {
		return fmt.Errorf("update_field: %w", err)
	}
	return nil
}

// updatePoint 直接更新 point 字段（UpdateIssueInput 无该字段）。
func (x *dbActionExecutor) updatePoint(ctx context.Context, wsID, issueID int64, version, point int) error {
	return x.directUpdate(ctx, wsID, issueID, version, "point = $1", point)
}

// updateTimeCol 直接更新时间戳字段（UpdateIssueInput 无 started_at/completed_at/target_date）。
func (x *dbActionExecutor) updateTimeCol(ctx context.Context, wsID, issueID int64, version int, col string, t time.Time) error {
	// col 来自受控的 switch 分支，此处使用白名单拼接
	return x.directUpdate(ctx, wsID, issueID, version, col+" = $1", t)
}

// directUpdate 直接执行带乐观锁的 UPDATE（version 不匹配则报错）。
func (x *dbActionExecutor) directUpdate(ctx context.Context, wsID, issueID int64, version int, setClause string, arg any) error {
	q := "UPDATE issues SET " + setClause + ", version = version + 1, updated_at = now()" +
		" WHERE id = $2 AND workspace_id = $3 AND version = $4 AND deleted_at IS NULL"
	tag, err := x.db.Exec(ctx, q, arg, issueID, wsID, version)
	if err != nil {
		return fmt.Errorf("direct_update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("direct_update: version conflict or not found")
	}
	return nil
}

// SendNotification 发送一条 in-app 通知（跳过操作者本人）。
func (x *dbActionExecutor) SendNotification(ctx context.Context, n NotificationRequest) error {
	if n.ActorID != nil && *n.ActorID == n.RecipientID {
		return nil // 不通知操作者自己
	}
	svc := notification.NewService(x.db)
	_, err := svc.Create(ctx, notification.CreateNotificationInput{
		WorkspaceID: n.WorkspaceID,
		RecipientID: n.RecipientID,
		EventType:   notification.EventType("automation"),
		EntityType:  notification.EntityProject,
		EntityID:    n.ProjectID,
		Title:       n.Title,
		Body:        n.Body,
		ActorID:     n.ActorID,
		ActorName:   n.ActorName,
		Channel:     notification.ChannelInApp,
		Payload:     json.RawMessage(`{}`),
	})
	if err != nil {
		return fmt.Errorf("send_notification: %w", err)
	}
	return nil
}

// CreateIssue 创建工作项，返回新工作项 ID。
func (x *dbActionExecutor) CreateIssue(ctx context.Context, in CreateIssueRequest) (int64, error) {
	typeCode := issue.IssueTypeCode(in.TypeCode)
	if typeCode == "" {
		typeCode = issue.TypeRequirement
	}
	iss, err := x.issueSvc.Create(ctx, issue.CreateIssueInput{
		WorkspaceID:     in.WorkspaceID,
		ProjectID:       in.ProjectID,
		TypeCode:        typeCode,
		Name:            in.Name,
		DescriptionHTML: in.Description,
		CreatedBy:       in.CreatedBy,
	})
	if err != nil {
		return 0, fmt.Errorf("create_issue: %w", err)
	}
	return iss.ID, nil
}

// appendActivity 追加审计日志（RLS 表，需设置租户上下文，失败静默）。
func (x *dbActionExecutor) appendActivity(ctx context.Context, wsID, projectID, issueID int64,
	verb, field string, oldVal, newVal *string) {
	tx, err := x.db.Begin(ctx)
	if err != nil {
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return
	}
	_, _ = tx.Exec(ctx, `
		INSERT INTO issue_activities (workspace_id, project_id, issue_id, verb, field, old_value, new_value)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		wsID, projectID, issueID, verb, field, oldVal, newVal)
	_ = tx.Commit(ctx)
}

// --- 类型转换工具 ---

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return 0, fmt.Errorf("cannot parse int from %q", val)
		}
		return n, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

func toInt64(v any) (int64, error) {
	n, err := toInt(v)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

func toInt64Slice(v any) ([]int64, error) {
	switch val := v.(type) {
	case []any:
		out := make([]int64, 0, len(val))
		for _, item := range val {
			n, err := toInt64(item)
			if err != nil {
				return nil, err
			}
			out = append(out, n)
		}
		return out, nil
	case []int64:
		return val, nil
	case int64:
		return []int64{val}, nil
	default:
		// 单个数值 → 单元素列表
		if n, err := toInt64(v); err == nil {
			return []int64{n}, nil
		}
		return nil, fmt.Errorf("cannot convert %T to int64 slice", v)
	}
}

func toTime(v any) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case *time.Time:
		if val == nil {
			return time.Time{}, fmt.Errorf("nil time")
		}
		return *val, nil
	case string:
		for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05"} {
			if t, err := time.Parse(layout, strings.TrimSpace(val)); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time from %q", val)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}
