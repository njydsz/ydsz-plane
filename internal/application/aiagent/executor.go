// Package aiagent — AI 动作执行器。
//
// 将 Build/Autopilot 模式产生的动作落到实际领域表。
// 所有写操作均走参数化 SQL，字段名来自内部常量，不拼接外部输入。
package aiagent

import (
	"context"
	"strings"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// execCreateWorkItem 创建需求/任务/缺陷。
func (s *Service) execCreateWorkItem(ctx context.Context, p *ActionPayload) error {
	if p.ProjectID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	if strings.TrimSpace(p.Title) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "required"})
	}
	wiType := strings.ToLower(p.WorkItemType)
	if wiType == "" {
		wiType = "task"
	}

	table := workItemTable(wiType)
	if table == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}

	// 统一插入：三张工作项表均有 project_id/title/description/priority/state_id/created_by
	if _, err := s.db.Exec(ctx, `
		INSERT INTO `+table+` (id, project_id, title, description, priority, state_id, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,
			(SELECT id FROM states WHERE project_id = $2 AND "group" = 'backlog' ORDER BY sort_order LIMIT 1),
			$6,$6)`,
		genID(), p.ProjectID, p.Title, p.Description, defaultStr(p.Priority, "medium"),
		p.AssigneeID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// execUpdateWorkItem 更新工作项字段。
func (s *Service) execUpdateWorkItem(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids", Reason: "required"})
	}
	wiType := strings.ToLower(p.WorkItemType)
	if wiType == "" {
		wiType = "task"
	}

	sets := []string{}
	args := []any{}
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if p.Title != "" {
		add("title = ", p.Title)
	}
	if p.Description != "" {
		add("description = ", p.Description)
	}
	if p.Priority != "" {
		add("priority = ", p.Priority)
	}
	if len(sets) == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "fields", Reason: "nothing to update"})
	}
	sets = append(sets, "updated_at = now()")

	table := workItemTable(wiType)
	if table == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}
	args = append(args, p.IDs)
	query := "UPDATE " + table + " SET " + joinComma(sets) +
		" WHERE id = ANY($" + itoa(len(args)) + "::bigint[])"
	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// execUpdateState 流转工作项状态。
func (s *Service) execUpdateState(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 || p.StateID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids/state_id", Reason: "required"})
	}
	wiType := strings.ToLower(p.WorkItemType)
	if wiType == "" {
		wiType = "task"
	}
	table := workItemTable(wiType)
	if table == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}
	if _, err := s.db.Exec(ctx,
		"UPDATE "+table+" SET state_id = $1, updated_at = now() WHERE id = ANY($2::bigint[])",
		p.StateID, p.IDs); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// execAssign 指派工作项。
func (s *Service) execAssign(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 || p.AssigneeID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids/assignee_id", Reason: "required"})
	}
	wiType := strings.ToLower(p.WorkItemType)
	if wiType == "" {
		wiType = "task"
	}
	assigneeTable, fkColumn := assigneeTableOf(wiType)
	if assigneeTable == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}
	if _, err := s.db.Exec(ctx,
		"DELETE FROM "+assigneeTable+" WHERE "+fkColumn+" = ANY($1::bigint[])", p.IDs); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	for _, id := range p.IDs {
		if _, err := s.db.Exec(ctx,
			"INSERT INTO "+assigneeTable+" (id, "+fkColumn+", user_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING",
			genID(), id, p.AssigneeID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// execAddComment 添加评论。
func (s *Service) execAddComment(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 || strings.TrimSpace(p.Content) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids/content", Reason: "required"})
	}
	if p.AssigneeID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "author_id", Reason: "required"})
	}
	wiType := strings.ToLower(defaultStr(p.WorkItemType, "task"))
	commentTable, fkColumn := commentTableOf(wiType)
	if commentTable == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}
	for _, id := range p.IDs {
		if _, err := s.db.Exec(ctx, `
			INSERT INTO `+commentTable+` (id, workspace_id, project_id, `+fkColumn+`,
				content_json, content_html, content_stripped, created_by)
			VALUES ($1,$2,$3,$4,'[]','<p>'||$5||'</p>',$5,$6)`,
			genID(), 0, p.ProjectID, id, p.Content, p.AssigneeID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// execLogTime 记录工时（累加到 task.actual_effort）。
func (s *Service) execLogTime(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 || p.Hours <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids/hours", Reason: "required"})
	}
	// 工时按任务累计到 actual_effort（项目无独立工时表）
	for _, id := range p.IDs {
		if _, err := s.db.Exec(ctx, `
			UPDATE task SET actual_effort = COALESCE(actual_effort, 0) + $1,
				updated_at = now()
			WHERE id = $2`,
			p.Hours, id); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// execCreateModule 创建模块。
func (s *Service) execCreateModule(ctx context.Context, p *ActionPayload) error {
	if p.ProjectID <= 0 || strings.TrimSpace(p.Name) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id/name", Reason: "required"})
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO modules (id, project_id, name, description, lead_id, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$5,$5)`,
		genID(), p.ProjectID, p.Name, p.Description, p.AssigneeID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// execCreateCycle 创建迭代。
func (s *Service) execCreateCycle(ctx context.Context, p *ActionPayload) error {
	if p.ProjectID <= 0 || strings.TrimSpace(p.Name) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id/name", Reason: "required"})
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO sprints (id, project_id, name, goal, status, created_by, updated_by)
		VALUES ($1,$2,$3,$4,'planned',$5,$5)`,
		genID(), p.ProjectID, p.Name, p.Description, p.AssigneeID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// execDeleteWorkItem 删除工作项（软删）。
func (s *Service) execDeleteWorkItem(ctx context.Context, p *ActionPayload) error {
	if len(p.IDs) == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids", Reason: "required"})
	}
	wiType := strings.ToLower(p.WorkItemType)
	if wiType == "" {
		wiType = "task"
	}
	table := workItemTable(wiType)
	if table == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "unsupported type"})
	}
	if _, err := s.db.Exec(ctx,
		"UPDATE "+table+" SET deleted = true, updated_at = now() WHERE id = ANY($1::bigint[])",
		p.IDs); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// --- helpers ---

// workItemTable 返回工作项类型对应的表名。
func workItemTable(wiType string) string {
	switch strings.ToLower(wiType) {
	case "requirement", "req":
		return "requirement"
	case "task":
		return "task"
	case "defect", "bug":
		return "defect"
	}
	return ""
}

// assigneeTableOf 返回指派人关联表与外键列。
func assigneeTableOf(wiType string) (string, string) {
	switch strings.ToLower(wiType) {
	case "requirement", "req":
		return "requirement_assignees", "requirement_id"
	case "task":
		return "task_assignees", "task_id"
	case "defect", "bug":
		return "defect_assignees", "defect_id"
	}
	return "", ""
}

// commentTableOf 返回评论表与外键列。
func commentTableOf(wiType string) (string, string) {
	switch strings.ToLower(wiType) {
	case "requirement", "req":
		return "requirement_comments", "requirement_id"
	case "task":
		return "task_comments", "task_id"
	case "defect", "bug":
		return "defect_comments", "defect_id"
	}
	return "", ""
}

func defaultStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func intOrDefault(f float64, def int) int {
	if f <= 0 {
		return def
	}
	return int(f)
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
