// Package workbench — 个人工作台应用服务。
//
// 实现要点:
//   - 聚合 5 类数据：我的任务（分桶）、迭代概览、最近访问、快捷操作
//   - 单一 API 调用返回完整首屏数据，避免前端 N+1
//   - 支持按项目过滤或工作空间级全局
//   - 布局持久化为 JSONB（gridstack style）
package workbench

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供工作台应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建工作台服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// GetSummary 获取工作台首屏聚合数据。
func (s *Service) GetSummary(ctx context.Context, wsID, userID int64, projectID *int64) (*WorkbenchSummary, error) {
	var summary WorkbenchSummary
	var err error

	// 我的任务分桶
	bucket, err := s.getMyIssuesBucket(ctx, wsID, userID, projectID)
	if err != nil {
		return nil, err
	}
	summary.MyIssues = bucket

	// 迭代概览
	summary.SprintOverviews, err = s.getSprintOverviews(ctx, wsID, userID, projectID)
	if err != nil {
		return nil, err
	}

	// 最近访问
	summary.RecentItems, err = s.ListRecentItems(ctx, wsID, userID, 10)
	if err != nil {
		return nil, err
	}

	// 统计数
	summary.OverdueCount = len(bucket.Overdue)
	summary.BlockedCount = s.countBlocked(ctx, wsID, userID)

	// 快捷操作
	summary.QuickActions = QuickActionSet{
		CanCreateIssue:   true,
		CanStartSprint:   projectID != nil && *projectID > 0,
		ActiveIssueCount: len(bucket.InProgress),
	}

	return &summary, nil
}

// getMyIssuesBucket 获取我的工作项分桶。
func (s *Service) getMyIssuesBucket(ctx context.Context, wsID, userID int64, projectID *int64) (MyIssuesBucket, error) {
	bucket := MyIssuesBucket{}
	today := time.Now().Format("2006-01-02")

	whereSQL := `
		WHERE i.workspace_id = $1 AND i.deleted = false
			AND EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia WHERE ia.issue_id = i.id AND ia.user_id = $2)
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))`
	args := []interface{}{wsID, userID}
	argIdx := 3
	if projectID != nil && *projectID > 0 {
		whereSQL += fmt.Sprintf(" AND i.project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	}

	// 总数
	countSQL := "SELECT count(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i " + whereSQL
	if err := s.db.QueryRow(ctx, countSQL, args...).Scan(&bucket.Total); err != nil {
		return bucket, errs.ErrInternal.Wrap(err)
	}

	// 查询所有分配给我的未完成工作项，按状态分桶
	querySQL := fmt.Sprintf(`
		SELECT
			i.id, i.identifier, i.name, i.type_code, i.priority,
			i.state_id, s.name AS state_name, s.color AS state_color, s."group" AS state_group,
			i.target_date,
			p.id AS project_id, p.name AS project_name,
			sp.id AS sprint_id, sp.name AS sprint_name,
			EXISTS (SELECT 1 FROM biz_entity_relations ir
				WHERE ir.source_id = i.id AND ir.relation_type = 'blocked_by') AS is_blocked
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i
		JOIN states s ON s.id = i.state_id
		JOIN projects p ON p.id = i.project_id
		LEFT JOIN sprints sp ON sp.id = i.sprint_id AND sp.deleted = false
		%s
		ORDER BY
			CASE s."group" WHEN 'started' THEN 0 WHEN 'backlog' THEN 1 ELSE 2 END,
			CASE i.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END,
			i.target_date NULLS LAST`, whereSQL)

	rows, err := s.db.Query(ctx, querySQL, args...)
	if err != nil {
		return bucket, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		var d IssueDigest
		var targetDate *time.Time
		var stateGroup string // 仅用于消耗扫描列
		if err := rows.Scan(
			&d.ID, &d.Identifier, &d.Title, &d.TypeCode, &d.Priority,
			&d.StateID, &d.StateName, &d.StateColor, &stateGroup,
			&targetDate,
			&d.GroupID, &d.ProjectName,
			&d.SprintID, &d.SprintName,
			&d.IsBlocked); err != nil {
			return bucket, errs.ErrInternal.Wrap(err)
		}
		if targetDate != nil {
			t := targetDate.Format("2006-01-02")
			d.TargetDate = &t
		}

		// 根据状态组和 target_date 分桶
		switch {
		case d.TargetDate != nil && *d.TargetDate == today:
			if len(bucket.Today) < 10 {
				bucket.Today = append(bucket.Today, d)
			}
		case d.TargetDate != nil && *d.TargetDate < today:
			if len(bucket.Overdue) < 10 {
				bucket.Overdue = append(bucket.Overdue, d)
			}
		case stateGroup == "started":
			if len(bucket.InProgress) < 10 {
				bucket.InProgress = append(bucket.InProgress, d)
			}
		case d.TargetDate != nil:
			if len(bucket.Upcoming) < 10 {
				bucket.Upcoming = append(bucket.Upcoming, d)
			}
		default:
			if len(bucket.Backlog) < 10 {
				bucket.Backlog = append(bucket.Backlog, d)
			}
		}
	}

	return bucket, rows.Err()
}

// countBlocked 统计阻塞工作项数。
func (s *Service) countBlocked(ctx context.Context, wsID, userID int64) int {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i
		WHERE i.workspace_id = $1 AND i.deleted = false
			AND EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia WHERE ia.issue_id = i.id AND ia.user_id = $2)
			AND EXISTS (SELECT 1 FROM biz_entity_relations ir
				WHERE ir.source_id = i.id AND ir.relation_type = 'blocked_by')
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))`,
		wsID, userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// getSprintOverviews 获取参与迭代概览。
func (s *Service) getSprintOverviews(ctx context.Context, wsID, userID int64, projectID *int64) ([]SprintOverview, error) {
	whereSQL := "WHERE s.workspace_id = $1 AND s.status = 'active' AND s.deleted = false"
	args := []interface{}{wsID}
	argIdx := 2

	if projectID != nil && *projectID > 0 {
		whereSQL += fmt.Sprintf(" AND s.project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	}

	// 限制为与我相关的迭代（我在其中有工作项 OR 我是 owner）
	whereSQL += fmt.Sprintf(`
		AND (s.owner_id = $%d OR EXISTS (
			SELECT 1 FROM (
				SELECT sprint_id, task_id AS issue_id FROM sprint_tasks
				UNION ALL SELECT sprint_id, requirement_id FROM sprint_requirements
				UNION ALL SELECT sprint_id, defect_id FROM sprint_defects
			) si
			JOIN (SELECT id FROM requirement WHERE deleted = false
			    UNION ALL SELECT id FROM task WHERE deleted = false
			    UNION ALL SELECT id FROM defect WHERE deleted = false) i ON i.id = si.issue_id
			JOIN (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia ON ia.issue_id = i.id
			WHERE si.sprint_id = s.id AND ia.user_id = $%d
		))`, argIdx, argIdx)
	args = append(args, userID)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT
			s.id, s.name, s.goal, s.status, s.start_date, s.end_date,
			p.id AS project_id, p.name AS project_name,
			(SELECT count(*) FROM (
				SELECT sprint_id, task_id AS issue_id FROM sprint_tasks
				UNION ALL SELECT sprint_id, requirement_id FROM sprint_requirements
				UNION ALL SELECT sprint_id, defect_id FROM sprint_defects
			) si JOIN (SELECT id FROM requirement WHERE deleted = false
				UNION ALL SELECT id FROM task WHERE deleted = false
				UNION ALL SELECT id FROM defect WHERE deleted = false) i ON i.id = si.issue_id
				JOIN (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia ON ia.issue_id = i.id
				WHERE si.sprint_id = s.id AND ia.user_id = $%d) AS my_count,
			(SELECT coalesce(sum(CASE WHEN sg."group" = 'completed' THEN 1 ELSE 0 END), 0)
			 FROM (
				SELECT sprint_id, task_id AS issue_id FROM sprint_tasks
				UNION ALL SELECT sprint_id, requirement_id FROM sprint_requirements
				UNION ALL SELECT sprint_id, defect_id FROM sprint_defects
			) si JOIN (SELECT id, state_id FROM requirement WHERE deleted = false
				UNION ALL SELECT id, state_id FROM task WHERE deleted = false
				UNION ALL SELECT id, state_id FROM defect WHERE deleted = false) i ON i.id = si.issue_id
				JOIN states sg ON sg.id = i.state_id WHERE si.sprint_id = s.id) AS done_count,
			(SELECT count(*) FROM (
				SELECT sprint_id, task_id AS issue_id FROM sprint_tasks
				UNION ALL SELECT sprint_id, requirement_id FROM sprint_requirements
				UNION ALL SELECT sprint_id, defect_id FROM sprint_defects
			) si WHERE si.sprint_id = s.id) AS total_count
		FROM sprints s
		JOIN projects p ON p.id = s.project_id
		%s
		ORDER BY s.start_date DESC
		LIMIT 5`, argIdx, whereSQL), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var overviews []SprintOverview
	for rows.Next() {
		var ov SprintOverview
		var startDate, endDate *time.Time
		var doneCount, totalCount int
		if err := rows.Scan(
			&ov.SprintID, &ov.SprintName, &ov.Goal, &ov.Status, &startDate, &endDate,
			&ov.ProjectID, &ov.ProjectName,
			&ov.MyIssueCount, &doneCount, &totalCount); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if totalCount > 0 {
			ov.Progress = float64(doneCount) / float64(totalCount)
		}
		if endDate != nil {
			ov.DaysRemaining = int(endDate.Sub(time.Now()).Hours() / 24)
		}
		overviews = append(overviews, ov)
	}
	return overviews, rows.Err()
}

// ListRecentItems 列出最近访问条目。
func (s *Service) ListRecentItems(ctx context.Context, wsID, userID int64, limit int) ([]RecentItem, error) {
	if limit > 20 {
		limit = 20
	}

	rows, err := s.db.Query(ctx, `
		SELECT item_type, item_id, project_id, title, identifier, accessed_at
		FROM recent_items
		WHERE workspace_id = $1 AND user_id = $2
		ORDER BY accessed_at DESC LIMIT $3`,
		wsID, userID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var items []RecentItem
	for rows.Next() {
		var it RecentItem
		var accessedAt time.Time
		if err := rows.Scan(&it.ItemType, &it.ItemID, &it.ProjectID, &it.Title, &it.Identifier, &accessedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		it.AccessedAt = accessedAt.Format(time.RFC3339)
		it.URL = buildItemURL(it.ItemType, it.ItemID, it.ProjectID)
		items = append(items, it)
	}
	return items, rows.Err()
}

// RecordRecent 记录最近访问（upsert，最新访问覆盖最旧）。
func (s *Service) RecordRecent(ctx context.Context, in RecordRecentInput) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO recent_items (workspace_id, user_id, item_type, item_id, project_id, title, identifier, accessed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id, item_type, item_id)
		DO UPDATE SET accessed_at = now(), title = EXCLUDED.title, identifier = EXCLUDED.identifier`,
		in.WorkspaceID, in.UserID, in.ItemType, in.ItemID, in.ProjectID, in.Title, in.Identifier)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	// 修剪保留最近 50 条
	_, _ = s.db.Exec(ctx, `
		DELETE FROM recent_items
		WHERE user_id = $1 AND id NOT IN (
			SELECT id FROM recent_items WHERE user_id = $1 ORDER BY accessed_at DESC LIMIT 50
		)`, in.UserID)

	return nil
}

// --- Layout ---

// GetConfig 获取工作台配置。
func (s *Service) GetConfig(ctx context.Context, wsID, userID int64, projectID *int64) (*WorkbenchConfig, error) {
	var cfg WorkbenchConfig
	var projID interface{}
	if projectID != nil {
		projID = *projectID
	}

	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, user_id, layout, widget_states, focus_enabled, updated_at
		FROM workbench_configs
		WHERE workspace_id = $1 AND user_id = $2 AND project_id IS NOT DISTINCT FROM $3`,
		wsID, userID, projID).Scan(
		&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.UserID,
		&cfg.Layout, &cfg.WidgetStates, &cfg.FocusEnabled, &cfg.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// 返回默认空配置
			return &WorkbenchConfig{
				WorkspaceID:  wsID,
				ProjectID:    projectID,
				UserID:       userID,
				Layout:       LayoutConfig{Widgets: []LayoutWidget{}},
				WidgetStates: map[string]any{},
			}, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &cfg, nil
}

// SaveConfig 保存工作台布局配置。
func (s *Service) SaveConfig(ctx context.Context, in SaveLayoutInput) (*WorkbenchConfig, error) {
	layoutJSON, _ := json.Marshal(in.Layout)
	widgetStatesJSON, _ := json.Marshal(in.WidgetStates)
	var projID interface{}
	if in.ProjectID != nil {
		projID = *in.ProjectID
	}

	var cfg WorkbenchConfig
	err := s.db.QueryRow(ctx, `
		INSERT INTO workbench_configs (workspace_id, project_id, user_id, layout, widget_states, focus_enabled)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, COALESCE(project_id, 0))
		DO UPDATE SET layout = EXCLUDED.layout, widget_states = EXCLUDED.widget_states,
			focus_enabled = EXCLUDED.focus_enabled, updated_at = now()
		RETURNING id, workspace_id, project_id, user_id, layout, widget_states, focus_enabled, updated_at`,
		in.WorkspaceID, projID, in.UserID, layoutJSON, widgetStatesJSON, in.FocusEnabled).Scan(
		&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.UserID,
		&cfg.Layout, &cfg.WidgetStates, &cfg.FocusEnabled, &cfg.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &cfg, nil
}

// --- Templates ---

// ListTemplates 列出工作台模板。
func (s *Service) ListTemplates(ctx context.Context) ([]WorkbenchTemplate, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, slug, description, layout, icon, is_default, sort_order
		FROM workbench_templates ORDER BY sort_order`)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var templates []WorkbenchTemplate
	for rows.Next() {
		var t WorkbenchTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.Layout, &t.Icon, &t.IsDefault, &t.SortOrder); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// ApplyTemplate 应用模板到工作台。
func (s *Service) ApplyTemplate(ctx context.Context, in ApplyTemplateInput) (*WorkbenchConfig, error) {
	var tmpl WorkbenchTemplate
	err := s.db.QueryRow(ctx,
		`SELECT layout FROM workbench_templates WHERE slug = $1`, in.TemplateSlug).Scan(&tmpl.Layout)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	return s.SaveConfig(ctx, SaveLayoutInput{
		WorkspaceID:  in.WorkspaceID,
		ProjectID:    in.ProjectID,
		UserID:       in.UserID,
		Layout:       tmpl.Layout,
		WidgetStates: map[string]any{},
		FocusEnabled: false,
	})
}

// --- Feed ---

// GetFeed 获取关注动态流。
// 合并: watched issues + 评论过的 issue + 被@提及的 issue 的活动记录。
func (s *Service) GetFeed(ctx context.Context, wsID, userID int64, projectID *int64, limit, offset int) ([]FeedItem, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	args := []interface{}{wsID, userID}
	argIdx := 3

	projFilter := ""
	if projectID != nil && *projectID > 0 {
		projFilter = fmt.Sprintf(" AND ia.project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (ia.created_at, ia.id)
			ia.id, ia.workitem_id AS issue_id,
			i.name AS issue_name,
			i.identifier, i.type_code,
			ia.verb, COALESCE(ia.field_name, '') AS field,
			COALESCE(ia.new_value, '') AS new_value,
			COALESCE(u.display_name, '') AS actor_name,
			ia.created_at
		FROM (
		    SELECT id, workspace_id, project_id, task_id AS workitem_id, verb, field_name, old_value, new_value, actor_id, created_at FROM task_activities
		    UNION ALL SELECT id, workspace_id, project_id, requirement_id, verb, field_name, old_value, new_value, actor_id, created_at FROM requirement_activities
		    UNION ALL SELECT id, workspace_id, project_id, defect_id, verb, field_name, old_value, new_value, actor_id, created_at FROM defect_activities
		) ia
		LEFT JOIN users u ON u.id = ia.actor_id
		JOIN (SELECT id, name, identifier, type_code FROM requirement WHERE deleted = false
		    UNION ALL SELECT id, name, identifier, type_code FROM task WHERE deleted = false
		    UNION ALL SELECT id, name, identifier, type_code FROM defect WHERE deleted = false) i ON i.id = ia.workitem_id
		WHERE ia.workspace_id = $1
			AND (ia.actor_id IS NULL OR ia.actor_id != $2)
			AND (
				EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_watchers UNION ALL SELECT requirement_id, user_id FROM requirement_watchers UNION ALL SELECT defect_id, user_id FROM defect_watchers) iw WHERE iw.issue_id = ia.workitem_id AND iw.user_id = $2)
				OR EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, created_by, mentions FROM task_comments UNION ALL SELECT requirement_id, created_by, mentions FROM requirement_comments UNION ALL SELECT defect_id, created_by, mentions FROM defect_comments) ic WHERE ic.issue_id = ia.workitem_id AND ic.created_by = $2)
				OR EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, created_by, mentions FROM task_comments UNION ALL SELECT requirement_id, created_by, mentions FROM requirement_comments UNION ALL SELECT defect_id, created_by, mentions FROM defect_comments) ic WHERE ic.issue_id = ia.workitem_id AND $2 = ANY(ic.mentions))
			)
			%s
		ORDER BY ia.created_at DESC, ia.id DESC
		LIMIT $%d OFFSET $%d`, projFilter, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var items []FeedItem
	for rows.Next() {
		var fi FeedItem
		var createdAt time.Time
		if err := rows.Scan(
			&fi.ID, &fi.IssueID, &fi.IssueName, &fi.Identifier, &fi.TypeCode,
			&fi.Verb, &fi.Field, &fi.NewValue, &fi.ActorName, &createdAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		fi.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, fi)
	}
	if items == nil {
		items = []FeedItem{}
	}
	return items, rows.Err()
}

// --- Efficiency ---

// GetEfficiency 获取个人效率报告。
func (s *Service) GetEfficiency(ctx context.Context, wsID, userID int64, projectID *int64) (*EfficiencyReport, error) {
	now := time.Now()
	weekStart := weekStartDate(now)

	var report EfficiencyReport

	// 本周完成的工作项数量和 story points
	if err := s.db.QueryRow(ctx, `
		SELECT count(*)::int, COALESCE(sum(i.point), 0)::int
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i
		JOIN states s ON s.id = i.state_id
		WHERE s."group" = 'completed'
			AND i.workspace_id = $1
			AND i.deleted = false
			AND i.completed_at IS NOT NULL
			AND i.completed_at::date >= $2
			AND i.completed_at::date <= $3
			AND EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia WHERE ia.issue_id = i.id AND ia.user_id = $4)`,
		wsID, weekStart, now, userID).Scan(&report.WeekIssues, &report.WeekPoints); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 本周记录的工时（分钟 → 小时）
	var weekMinutes int
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(sum(tl.duration_minutes), 0)::int
		FROM (
		    SELECT workspace_id, project_id, task_id AS workitem_id, user_id, spent_date, duration_minutes FROM task_timelogs WHERE deleted = false
		    UNION ALL SELECT workspace_id, project_id, requirement_id, user_id, spent_date, duration_minutes FROM requirement_timelogs WHERE deleted = false
		    UNION ALL SELECT workspace_id, project_id, defect_id, user_id, spent_date, duration_minutes FROM defect_timelogs WHERE deleted = false
		) tl
		WHERE tl.workspace_id = $1
			AND tl.user_id = $2
			AND tl.spent_date >= $3
			AND tl.spent_date <= $4
			AND tl.deleted = false`,
		wsID, userID, weekStart, now).Scan(&weekMinutes); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	report.WeekHours = mathRound(float64(weekMinutes)/60.0, 1)

	// 逾期数量
	if err := s.db.QueryRow(ctx, `
		SELECT count(*)::int FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i
		WHERE i.workspace_id = $1
			AND i.deleted = false
			AND i.target_date IS NOT NULL
			AND i.target_date < CURRENT_DATE
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))
			AND EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia WHERE ia.issue_id = i.id AND ia.user_id = $2)`,
		wsID, userID).Scan(&report.OverdueCount); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 最近 4 周完成趋势
	report.WeeklyTrend = make([]WeeklyTrend, 4)
	for i := 0; i < 4; i++ {
		weekLabel := fmt.Sprintf("W%d", 4-i)
		ws := weekStartDate(now.AddDate(0, 0, -7*i))
		we := ws.AddDate(0, 0, 6)
		var count, points int
		if err := s.db.QueryRow(ctx, `
			SELECT count(*)::int, COALESCE(sum(i.point), 0)::int
			FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i
			JOIN states s ON s.id = i.state_id
			WHERE s."group" = 'completed'
				AND i.workspace_id = $1
				AND i.deleted = false
				AND i.completed_at IS NOT NULL
				AND i.completed_at::date >= $2
				AND i.completed_at::date <= $3
				AND EXISTS (SELECT 1 FROM (SELECT task_id AS issue_id, user_id FROM task_assignees UNION ALL SELECT requirement_id, user_id FROM requirement_assignees UNION ALL SELECT defect_id, user_id FROM defect_assignees) ia WHERE ia.issue_id = i.id AND ia.user_id = $4)`,
			wsID, ws.Format("2006-01-02"), we.Format("2006-01-02"), userID).Scan(&count, &points); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		report.WeeklyTrend[i] = WeeklyTrend{Week: weekLabel, Count: count, Points: points}
	}

	return &report, nil
}

// --- Helpers ---

func buildItemURL(itemType string, itemID int64, projectID int64) string {
	switch itemType {
	case "issue":
		return fmt.Sprintf("/projects/%d/issues/%d", projectID, itemID)
	case "sprint":
		return fmt.Sprintf("/projects/%d/sprints/%d", projectID, itemID)
	case "version":
		return fmt.Sprintf("/projects/%d/versions/%d", projectID, itemID)
	case "project":
		return fmt.Sprintf("/projects/%d/board", itemID)
	}
	return ""
}

// weekStartDate 返回本周一的日期（ISO 周）。
func weekStartDate(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-int(weekday)+1, 0, 0, 0, 0, t.Location())
}

// mathRound 四舍五入到指定小数位。
func mathRound(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}
