// Package issue — 跨聚合根共享的基础设施函数。
//
// 这些 helper 被 TaskService / RequirementService / DefectService 复用，
// 避免三个服务重复实现连接事务、M2M 写入、领域事件录制等基础能力。
package issue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// workitemEventPayload 跨类型工作项事件的统一 payload 格式。
type workitemEventPayload struct {
	WorkspaceID  int64    `json:"workspace_id"`
	ProjectID    int64    `json:"project_id"`
	WorkitemType string   `json:"workitem_type"` // requirement | task | defect
	WorkitemID   int64    `json:"workitem_id"`
	ActorID      int64    `json:"actor_id"`
	ActorName    string   `json:"actor_name"`
	Identifier   string   `json:"identifier"`
	Name         string   `json:"name"`
	AssigneeIDs  []int64  `json:"assignee_ids"`
	FromState    string   `json:"from_state"`
	ToState      string   `json:"to_state"`
}

// recordWorkitemEvent 在既有事务内将领域事件写入 Outbox（domain_events 表）。
func recordWorkitemEvent(ctx context.Context, tx pgx.Tx, eventType string,
	wsID, projectID, workitemID int64, workitemType IssueTypeCode,
	actorID int64, actorName, identifier, name string, assigneeIDs []int64, fromState, toState string) error {
	payload := workitemEventPayload{
		WorkspaceID: wsID, ProjectID: projectID, WorkitemID: workitemID,
		WorkitemType: string(workitemType), ActorID: actorID, ActorName: actorName,
		Identifier: identifier, Name: name, AssigneeIDs: assigneeIDs,
		FromState: fromState, ToState: toState,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_events (workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)`,
		wsID, workitemType, workitemID, eventType, payloadJSON)
	if err != nil {
		return fmt.Errorf("record workitem event %s: %w", eventType, err)
	}
	return nil
}

// recordCommentEvent 在既有事务内将 Comment 领域事件写入 Outbox。
func recordCommentEvent(ctx context.Context, tx pgx.Tx, workspaceID, issueID, commentID, actorID int64, actorName, content string, assigneeIDs []int64) error {
	payload := struct {
		WorkspaceID int64   `json:"workspace_id"`
		ActorID     int64   `json:"actor_id"`
		ActorName   string  `json:"actor_name"`
		IssueID     int64   `json:"issue_id"`
		CommentID   int64   `json:"comment_id"`
		Content     string  `json:"content"`
		AssigneeIDs []int64 `json:"assignee_ids"`
	}{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		ActorName:   actorName,
		IssueID:     issueID,
		CommentID:   commentID,
		Content:     content,
		AssigneeIDs: assigneeIDs,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal comment payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_events (workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'comment', $2, 'comment.created', $3)`,
		workspaceID, issueID, payloadJSON)
	if err != nil {
		return fmt.Errorf("record comment event: %w", err)
	}
	return nil
}

// loadIntArray 查询整数数组。
func loadIntArray(ctx context.Context, db *pgxpool.Pool, query string, arg interface{}) ([]int64, error) {
	rows, err := db.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// loadIntArrayTx 在事务内查询整数数组。
func loadIntArrayTx(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) []int64 {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err == nil {
			ids = append(ids, v)
		}
	}
	return ids
}

// loadStateName 通过事务查状态名称。
func loadStateName(ctx context.Context, tx pgx.Tx, stateID int64) string {
	var name string
	_ = tx.QueryRow(ctx, `SELECT name FROM states WHERE id = $1`, stateID).Scan(&name)
	return name
}

// getUserNameTx 通过事务从 users 表查询用户显示名。
func getUserNameTx(ctx context.Context, tx pgx.Tx, userID int64) string {
	var name string
	_ = tx.QueryRow(ctx, `SELECT COALESCE(display_name,'') FROM users WHERE id = $1`, userID).Scan(&name)
	return name
}

// workitemM2MPrefix 返回工作项类型对应的 M2M 表前缀（task / requirement / defect）。
// 关联表结构：{prefix}_assignees / {prefix}_labels / {prefix}_modules / {prefix}_watchers。
func workitemM2MPrefix(t IssueTypeCode) string {
	switch t {
	case TypeTask, TypeRequirement, TypeDefect:
		return string(t)
	default:
		return string(TypeTask)
	}
}

// detectWorkitemType 通过三主表推断工作项类型。
// 雪花 ID 全局唯一，仅按 id 命中，无需 workspace_id。
func detectWorkitemType(ctx context.Context, db *pgxpool.Pool, workitemID int64) (IssueTypeCode, error) {
	var tc string
	err := db.QueryRow(ctx, `
		SELECT 'task' FROM task WHERE id = $1
		UNION ALL SELECT 'requirement' FROM requirement WHERE id = $1
		UNION ALL SELECT 'defect' FROM defect WHERE id = $1
		LIMIT 1`, workitemID).Scan(&tc)
	if err != nil {
		return "", errs.ErrNotFound
	}
	return IssueTypeCode(tc), nil
}

// subresourceTable 返回某工作项类型的子资源表名（suffix 形如 "_comments"）。
func subresourceTable(t IssueTypeCode, suffix string) string {
	return workitemM2MPrefix(t) + suffix
}

// locateSubresourceTable 在三个分表（task/requirement/defect）中定位子资源记录所属表。
// 用于 Update/Delete 这类仅凭子资源 id 无法直接推断类型的场景。
func locateSubresourceTable(ctx context.Context, db *pgxpool.Pool, suffix string, id int64) (string, error) {
	for _, p := range []string{"task", "requirement", "defect"} {
		tbl := p + suffix
		var found int64
		if err := db.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s WHERE id = $1`, tbl), id).Scan(&found); err == nil {
			return tbl, nil
		}
	}
	return "", errs.ErrNotFound
}

// insertM2M 写入 M2M 关联（assignees / labels / modules），仅当切片非 nil 时操作。
// tenant_id / created_by / updated_by / status / deleted 依赖表 DEFAULT 兜底。
func insertM2M(ctx context.Context, tx pgx.Tx, typeCode IssueTypeCode, wsID, projectID, workitemID int64, assignees, labels, modules []int64) error {
	prefix := workitemM2MPrefix(typeCode)
	idCol := prefix + "_id"
	for _, uid := range assignees {
		if _, err := tx.Exec(ctx, fmt.Sprintf(
			`INSERT INTO %s_assignees (workspace_id, project_id, %s, user_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			prefix, idCol), wsID, projectID, workitemID, uid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, lid := range labels {
		if _, err := tx.Exec(ctx, fmt.Sprintf(
			`INSERT INTO %s_labels (workspace_id, project_id, %s, label_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			prefix, idCol), wsID, projectID, workitemID, lid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, mid := range modules {
		if _, err := tx.Exec(ctx, fmt.Sprintf(
			`INSERT INTO %s_modules (workspace_id, project_id, %s, module_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			prefix, idCol), wsID, projectID, workitemID, mid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// updateM2M 覆盖写入 M2M 关联；assignees / labels / modules 为 nil 表示不处理。
func updateM2M(ctx context.Context, tx pgx.Tx, typeCode IssueTypeCode, wsID, projectID, workitemID int64, assignees, labels, modules []int64) error {
	return sharedUpdateM2M(ctx, tx, typeCode, wsID, projectID, workitemID, assignees, labels, modules)
}

// sharedUpdateM2M 是 package-level 的 M2M 更新实现，被三个服务复用。
func sharedUpdateM2M(ctx context.Context, tx pgx.Tx, typeCode IssueTypeCode, wsID, projectID, workitemID int64, assignees, labels, modules []int64) error {
	prefix := workitemM2MPrefix(typeCode)
	idCol := prefix + "_id"
	if assignees != nil {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s_assignees WHERE %s = $1`, prefix, idCol), workitemID); err != nil {
			return err
		}
		for _, uid := range assignees {
			if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s_assignees (workspace_id, project_id, %s, user_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, prefix, idCol), wsID, projectID, workitemID, uid); err != nil {
				return err
			}
		}
	}
	if labels != nil {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s_labels WHERE %s = $1`, prefix, idCol), workitemID); err != nil {
			return err
		}
		for _, lid := range labels {
			if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s_labels (workspace_id, project_id, %s, label_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, prefix, idCol), wsID, projectID, workitemID, lid); err != nil {
				return err
			}
		}
	}
	if modules != nil {
		if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s_modules WHERE %s = $1`, prefix, idCol), workitemID); err != nil {
			return err
		}
		for _, mid := range modules {
			if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s_modules (workspace_id, project_id, %s, module_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, prefix, idCol), wsID, projectID, workitemID, mid); err != nil {
				return err
			}
		}
	}
	return nil
}

// triggerProgressRollup 递归回写父级进度（保留原语义；epic 容器场景复用）。
func triggerProgressRollup(ctx context.Context, tx pgx.Tx, parentID int64, tableName string) {
	var total, completed int
	err := tx.QueryRow(ctx, fmt.Sprintf(`
		SELECT count(*), count(*) FILTER (WHERE state_id IN (
		    SELECT id FROM states WHERE "group" = 'completed'
		))
		FROM %s WHERE parent_id = $1 AND workspace_id = (
		    SELECT workspace_id FROM %s WHERE id = $1
		) AND deleted = false`, tableName, tableName), parentID).Scan(&total, &completed)
	if err != nil || total == 0 {
		return
	}
	progress := completed * 100 / total
	_, _ = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET progress = $1, updated_at = now() WHERE id = $2`, tableName), progress, parentID)

	var grand sql.NullInt64
	_ = tx.QueryRow(ctx, fmt.Sprintf(`SELECT parent_id FROM %s WHERE id = $1`, tableName), parentID).Scan(&grand)
	if grand.Valid {
		triggerProgressRollup(ctx, tx, grand.Int64, tableName)
	}
}

// --- 为兼容旧调用保留的 Validation helpers ---

// validateWorkitemName 校验工作项名称（三个类型共用）。
func validateWorkitemName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
	}
	if errs := validateNameLen(name); errs != nil {
		return errs
	}
	return nil
}

func validateNameLen(name string) error {
	if len(name) > 500 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能超过 500 字符"})
	}
	return nil
}

// 校验类型字段合法性，各类型可嵌入使用。
func validateTypeCode(tc IssueTypeCode) bool {
	switch tc {
	case TypeEpic, TypeRequirement, TypeTask, TypeDefect:
		return true
	}
	return false
}
