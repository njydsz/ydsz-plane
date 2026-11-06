// Package aiagent — 工作空间工具执行器实现。
//
// 将 LLM 的工具调用映射到本地领域服务（issue/sprint/module/workspace），
// 保持 aiagent 包对领域包的直接依赖最小化（通过接口注入）。
package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// WorkspaceToolProvider 基于 pgx 的工作空间工具执行器。
//
// 说明：为降低耦合，这里直接使用参数化 SQL 完成查询/写入，
// 与 issue/sprint 服务层的写操作保持相同的字段与约束语义。
type WorkspaceToolProvider struct {
	db *pgxpool.Pool
}

// NewWorkspaceToolProvider 构造执行器。
func NewWorkspaceToolProvider(db *pgxpool.Pool) *WorkspaceToolProvider {
	return &WorkspaceToolProvider{db: db}
}

// ExecuteTool 执行工具调用。
func (p *WorkspaceToolProvider) ExecuteTool(ctx context.Context, workspaceID, userID int64, name string, args json.RawMessage) (any, error) {
	if workspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	var a map[string]any
	if len(args) > 0 {
		if err := json.Unmarshal(args, &a); err != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "arguments", Reason: "invalid JSON"})
		}
	}

	switch name {
	// --- 查询类 ---
	case "list_work_items":
		return p.listWorkItems(ctx, workspaceID, a)
	case "get_work_item":
		return p.getWorkItem(ctx, workspaceID, a)
	case "search_workspace":
		return p.searchWorkspace(ctx, workspaceID, a)
	case "list_cycles":
		return p.listCycles(ctx, workspaceID, a)
	case "list_modules":
		return p.listModules(ctx, workspaceID, a)
	case "list_members":
		return p.listMembers(ctx, workspaceID, a)

	// --- 写入类 ---
	case "create_work_item":
		return p.createWorkItem(ctx, workspaceID, userID, a)
	case "update_work_item":
		return p.updateWorkItem(ctx, workspaceID, userID, a)
	case "transition_work_item":
		return p.transitionWorkItem(ctx, workspaceID, userID, a)
	case "assign_work_item":
		return p.assignWorkItem(ctx, workspaceID, userID, a)
	case "add_comment":
		return p.addComment(ctx, workspaceID, userID, a)
	case "create_cycle":
		return p.createCycle(ctx, workspaceID, userID, a)
	}
	return nil, fmt.Errorf("未知工具：%s", name)
}

// --- 查询实现 ---

// listWorkItems 列出工作项（跨三类表 UNION）。
func (p *WorkspaceToolProvider) listWorkItems(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	projectID := int64Arg(args, "project_id")
	limit := intArg(args, "limit", 50)
	offset := intArg(args, "offset", 0)
	wiType := strings.ToLower(strArg(args, "type"))
	stateName := strArg(args, "state")
	assignee := strings.ToLower(strArg(args, "assignee"))

	// 按类型选择表；未指定时跨三表 UNION
	tables := []string{"requirement", "task", "defect"}
	if wiType != "" {
		if t := workItemTable(wiType); t != "" {
			tables = []string{t}
		}
	}

	var queries []string
	var params []any
	paramIdx := 1
	addParam := func(v any) string {
		params = append(params, v)
		s := "$" + strconv.Itoa(paramIdx)
		paramIdx++
		return s
	}

	for _, table := range tables {
		conds := []string{"w.project_id = " + addParam(projectID), "w.deleted = false"}
		if stateName != "" {
			conds = append(conds, `EXISTS (
				SELECT 1 FROM states s WHERE s.id = w.state_id AND LOWER(s.name) = LOWER(`+addParam(stateName)+`))`)
		}
		if assignee != "" {
			if assignee == "me" {
				at, afk := assigneeTableOf(table)
				conds = append(conds, `EXISTS (
					SELECT 1 FROM `+at+` a WHERE a.`+afk+` = w.id AND a.user_id = `+addParam(userIDFromContext(ctx))+` AND a.del = false)`)
			} else {
				at2, afk2 := assigneeTableOf(table)
				conds = append(conds, `EXISTS (
					SELECT 1 FROM `+at2+` a
					JOIN workspace_members m ON m.user_id = a.user_id
					WHERE a.`+afk2+` = w.id AND a.del = false
					  AND LOWER(COALESCE(m.display_name, '')) = `+addParam(assignee)+`)`)
			}
		}
		queries = append(queries, fmt.Sprintf(`
			SELECT w.id, '%s' AS type, w.title, w.priority, w.state_id,
			       COALESCE((SELECT s.name FROM states s WHERE s.id = w.state_id), '') AS state_name,
			       COALESCE((SELECT s."group" FROM states s WHERE s.id = w.state_id), '') AS state_group,
			       w.created_at, w.updated_at
			FROM %s w WHERE %s`, table, table, strings.Join(conds, " AND ")))
	}

	query := strings.Join(queries, " UNION ALL ") + fmt.Sprintf(`
		ORDER BY updated_at DESC LIMIT %s OFFSET %s`,
		addParam(limit), addParam(offset))

	rows, err := p.db.Query(ctx, query, params...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var id, stateID int64
		var typ, title, priority, stateNameOut, stateGroup string
		var createdAt, updatedAt *string
		if err := rows.Scan(&id, &typ, &title, &priority, &stateID, &stateNameOut, &stateGroup, &createdAt, &updatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, map[string]any{
			"id": id, "type": typ, "title": title, "priority": priority,
			"state_id": stateID, "state": stateNameOut, "state_group": stateGroup,
		})
	}
	return map[string]any{"items": out, "count": len(out)}, rows.Err()
}

// getWorkItem 获取单个工作项。
func (p *WorkspaceToolProvider) getWorkItem(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	id := int64Arg(args, "id")
	if id <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id", Reason: "required"})
	}

	var title, description, priority, stateName, stateGroup string
	var stateID int64
	err := p.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT w.title, COALESCE(w.description,''), w.priority, w.state_id,
		       COALESCE((SELECT s.name FROM states s WHERE s.id = w.state_id), ''),
		       COALESCE((SELECT s."group" FROM states s WHERE s.id = w.state_id), '')
		FROM %s w WHERE w.id = $1 AND w.deleted = false`, table), id,
	).Scan(&title, &description, &priority, &stateID, &stateName, &stateGroup)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("工作项不存在：%s #%d", table, id)
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 指派人
	assignees := []map[string]any{}
	at3, afk3 := assigneeTableOf(table)
	aRows, err := p.db.Query(ctx, fmt.Sprintf(`
		SELECT a.user_id, COALESCE(m.display_name, '') AS name
		FROM %s a LEFT JOIN workspace_members m ON m.user_id = a.user_id
		WHERE a.%s = $1 AND a.del = false`, at3, afk3), id)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var uid int64
			var name string
			if err := aRows.Scan(&uid, &name); err == nil {
				assignees = append(assignees, map[string]any{"id": uid, "name": name})
			}
		}
	}

	return map[string]any{
		"id": id, "type": table, "title": title, "description": description,
		"priority": priority, "state_id": stateID, "state": stateName,
		"state_group": stateGroup, "assignees": assignees,
	}, nil
}

// searchWorkspace 全文搜索工作项。
func (p *WorkspaceToolProvider) searchWorkspace(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	q := strings.TrimSpace(strArg(args, "query"))
	if q == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "query", Reason: "required"})
	}
	limit := intArg(args, "limit", 20)

	rows, err := p.db.Query(ctx, `
		SELECT id, 'requirement' AS type, title, priority FROM requirement
		WHERE workspace_id = $1 AND deleted = false AND (title ILIKE $2 OR description ILIKE $2)
		UNION ALL
		SELECT id, 'task', title, priority FROM task
		WHERE workspace_id = $1 AND deleted = false AND (title ILIKE $2 OR description ILIKE $2)
		UNION ALL
		SELECT id, 'defect', title, priority FROM defect
		WHERE workspace_id = $1 AND deleted = false AND (title ILIKE $2 OR description ILIKE $2)
		ORDER BY 1 LIMIT $3`,
		workspaceID, "%"+q+"%", limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var typ, title, priority string
		if err := rows.Scan(&id, &typ, &title, &priority); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, map[string]any{"id": id, "type": typ, "title": title, "priority": priority})
	}
	return map[string]any{"items": out, "count": len(out), "query": q}, rows.Err()
}

// listCycles 列出迭代。
func (p *WorkspaceToolProvider) listCycles(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	projectID := int64Arg(args, "project_id")
	status := strArg(args, "status")

	conds := []string{"s.workspace_id = $1"}
	params := []any{workspaceID}
	if projectID > 0 {
		params = append(params, projectID)
		conds = append(conds, "s.project_id = $"+strconv.Itoa(len(params)))
	}
	if status != "" {
		params = append(params, status)
		conds = append(conds, "s.status = $"+strconv.Itoa(len(params)))
	}

	rows, err := p.db.Query(ctx, `
		SELECT s.id, s.project_id, s.name, COALESCE(s.goal,''), s.status,
		       s.start_date, s.end_date
		FROM sprints s WHERE `+strings.Join(conds, " AND ")+`
		ORDER BY s.created_at DESC LIMIT 100`, params...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var id, projectIDOut int64
		var name, goal, statusOut string
		var startDate, endDate *string
		if err := rows.Scan(&id, &projectIDOut, &name, &goal, &statusOut, &startDate, &endDate); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, map[string]any{
			"id": id, "project_id": projectIDOut, "name": name, "goal": goal, "status": statusOut,
		})
	}
	return map[string]any{"items": out, "count": len(out)}, rows.Err()
}

// listModules 列出模块。
func (p *WorkspaceToolProvider) listModules(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	projectID := int64Arg(args, "project_id")
	rows, err := p.db.Query(ctx, `
		SELECT id, project_id, name, COALESCE(description,''),
		       COALESCE(workflow_status,'backlog')
		FROM modules
		WHERE workspace_id = $1 AND ($2 = 0 OR project_id = $2) AND deleted = false
		ORDER BY sort_order ASC LIMIT 200`, workspaceID, projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var id, projectIDOut int64
		var name, desc, status string
		if err := rows.Scan(&id, &projectIDOut, &name, &desc, &status); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, map[string]any{
			"id": id, "project_id": projectIDOut, "name": name,
			"description": desc, "status": status,
		})
	}
	return map[string]any{"items": out, "count": len(out)}, rows.Err()
}

// listMembers 列出工作空间成员。
func (p *WorkspaceToolProvider) listMembers(ctx context.Context, workspaceID int64, args map[string]any) (any, error) {
	rows, err := p.db.Query(ctx, `
		SELECT m.user_id, COALESCE(m.display_name, ''), m.email, m.role
		FROM workspace_members m
		WHERE m.tenant_id = $1 AND m.status = 'active'
		ORDER BY m.display_name ASC LIMIT 500`, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var uid int64
		var name, email, role string
		if err := rows.Scan(&uid, &name, &email, &role); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, map[string]any{
			"id": uid, "name": name, "email": email, "role": role,
		})
	}
	return map[string]any{"items": out, "count": len(out)}, rows.Err()
}

// --- 写入实现 ---

// createWorkItem 创建工作项。
func (p *WorkspaceToolProvider) createWorkItem(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	projectID := int64Arg(args, "project_id")
	title := strings.TrimSpace(strArg(args, "title"))
	if projectID <= 0 || title == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id/title", Reason: "required"})
	}
	priority := strArg(args, "priority")
	if priority == "" {
		priority = "medium"
	}

	var id int64
	err := p.db.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, workspace_id, project_id, title, description, priority, state_id,
			created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,
			(SELECT id FROM states WHERE project_id = $3 AND "group" = 'backlog' ORDER BY sort_order LIMIT 1),
			$7,$7)
		RETURNING id`, table),
		genID(), workspaceID, projectID, title, strArg(args, "description"), priority, userID,
	).Scan(&id)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 指派人
	if assigneeID := int64Arg(args, "assignee_id"); assigneeID > 0 {
		at, afk := assigneeTableOf(table)
		_, _ = p.db.Exec(ctx, fmt.Sprintf(`
			INSERT INTO %s (id, workspace_id, %s, user_id)
			VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			at, afk), genID(), workspaceID, id, assigneeID)
	}

	return map[string]any{"id": id, "type": table, "title": title, "status": "created"}, nil
}

// updateWorkItem 更新工作项。
func (p *WorkspaceToolProvider) updateWorkItem(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	id := int64Arg(args, "id")
	if id <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id", Reason: "required"})
	}

	sets := []string{}
	params := []any{}
	add := func(expr string, v any) {
		params = append(params, v)
		sets = append(sets, expr+"$"+strconv.Itoa(len(params)))
	}
	if v := strArg(args, "title"); v != "" {
		add("title = ", v)
	}
	if _, ok := args["description"]; ok {
		add("description = ", strArg(args, "description"))
	}
	if v := strArg(args, "priority"); v != "" {
		add("priority = ", v)
	}
	if v := int64Arg(args, "state_id"); v > 0 {
		add("state_id = ", v)
	}
	if len(sets) == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "fields", Reason: "nothing to update"})
	}
	sets = append(sets, "updated_at = now()")
	params = append(params, id, workspaceID)

	tag, err := p.db.Exec(ctx, fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d AND workspace_id = $%d AND deleted = false",
		table, strings.Join(sets, ", "), len(params)-1, len(params)), params...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("工作项不存在或无权更新：%s #%d", table, id)
	}
	return map[string]any{"id": id, "type": table, "updated": true}, nil
}

// transitionWorkItem 流转状态。
func (p *WorkspaceToolProvider) transitionWorkItem(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	id := int64Arg(args, "id")
	stateID := int64Arg(args, "state_id")
	if id <= 0 || stateID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id/state_id", Reason: "required"})
	}

	tag, err := p.db.Exec(ctx, fmt.Sprintf(
		"UPDATE %s SET state_id = $1, updated_at = now() WHERE id = $2 AND workspace_id = $3 AND deleted = false",
		table), stateID, id, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("工作项不存在或无权流转：%s #%d", table, id)
	}
	return map[string]any{"id": id, "state_id": stateID, "transitioned": true}, nil
}

// assignWorkItem 指派工作项。
func (p *WorkspaceToolProvider) assignWorkItem(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	id := int64Arg(args, "id")
	assigneeID := int64Arg(args, "user_id")
	if id <= 0 || assigneeID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id/user_id", Reason: "required"})
	}

	at5, afk5 := assigneeTableOf(table)
	if _, err := p.db.Exec(ctx, fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1", at5, afk5), id); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	at6, afk6 := assigneeTableOf(table)
	if _, err := p.db.Exec(ctx, fmt.Sprintf(
		"INSERT INTO %s (id, workspace_id, %s, user_id) VALUES ($1,$2,$3,$4)",
		at6, afk6), genID(), workspaceID, id, assigneeID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return map[string]any{"id": id, "assignee_id": assigneeID, "assigned": true}, nil
}

// addComment 添加评论。
func (p *WorkspaceToolProvider) addComment(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	table := workItemTable(strArg(args, "type"))
	if table == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "required"})
	}
	id := int64Arg(args, "id")
	content := strings.TrimSpace(strArg(args, "content"))
	if id <= 0 || content == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id/content", Reason: "required"})
	}

	commentTable, fk := commentTableOf(table)
	if commentTable == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "unsupported type"})
	}

	var commentID int64
	err := p.db.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, workspace_id, project_id, %s, content_json, content_html,
			content_stripped, created_by)
		VALUES ($1,$2,$3,$4,'[]',$5,$6,$7)
		RETURNING id`, commentTable, fk),
		genID(), workspaceID, int64Arg(args, "project_id"), id,
		"<p>"+content+"</p>", content, userID,
	).Scan(&commentID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return map[string]any{"comment_id": commentID, "work_item_id": id, "created": true}, nil
}

// createCycle 创建迭代。
func (p *WorkspaceToolProvider) createCycle(ctx context.Context, workspaceID, userID int64, args map[string]any) (any, error) {
	projectID := int64Arg(args, "project_id")
	name := strings.TrimSpace(strArg(args, "name"))
	if projectID <= 0 || name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id/name", Reason: "required"})
	}

	var id int64
	err := p.db.QueryRow(ctx, `
		INSERT INTO sprints (id, workspace_id, project_id, name, goal, status, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,'planned',$6,$6)
		RETURNING id`, genID(), workspaceID, projectID, name, strArg(args, "goal"), userID,
	).Scan(&id)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return map[string]any{"id": id, "name": name, "status": "planned"}, nil
}

// --- helpers ---

func strArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func int64Arg(args map[string]any, key string) int64 {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	}
	return 0
}

func intArg(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok && f > 0 {
			return int(f)
		}
	}
	return def
}

// fkOf 返回指派人关联表的外键列名。
func fkOf(wiType string) string {
	_, fk := assigneeTableOf(wiType)
	return fk
}
