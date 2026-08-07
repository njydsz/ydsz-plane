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
		WHERE i.workspace_id = $1 AND i.deleted_at IS NULL
			AND EXISTS (SELECT 1 FROM issue_assignees ia WHERE ia.issue_id = i.id AND ia.user_id = $2)
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))`
	args := []interface{}{wsID, userID}
	argIdx := 3
	if projectID != nil && *projectID > 0 {
		whereSQL += fmt.Sprintf(" AND i.project_id = $%d", argIdx)
		args = append(args, *projectID)
		argIdx++
	}

	// 总数
	countSQL := "SELECT count(*) FROM issues i " + whereSQL
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
			EXISTS (SELECT 1 FROM issue_relations ir
				WHERE ir.source_issue_id = i.id AND ir.relation_type = 'blocked_by') AS is_blocked
		FROM issues i
		JOIN states s ON s.id = i.state_id
		JOIN projects p ON p.id = i.project_id
		LEFT JOIN sprints sp ON sp.id = i.sprint_id AND sp.deleted_at IS NULL
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
		SELECT count(*) FROM issues i
		WHERE i.workspace_id = $1 AND i.deleted_at IS NULL
			AND EXISTS (SELECT 1 FROM issue_assignees ia WHERE ia.issue_id = i.id AND ia.user_id = $2)
			AND EXISTS (SELECT 1 FROM issue_relations ir
				WHERE ir.source_issue_id = i.id AND ir.relation_type = 'blocked_by')
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))`,
		wsID, userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// getSprintOverviews 获取参与迭代概览。
func (s *Service) getSprintOverviews(ctx context.Context, wsID, userID int64, projectID *int64) ([]SprintOverview, error) {
	whereSQL := "WHERE s.workspace_id = $1 AND s.status = 'active' AND s.deleted_at IS NULL"
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
			SELECT 1 FROM sprint_issues si
			JOIN issues i ON i.id = si.issue_id
			JOIN issue_assignees ia ON ia.issue_id = i.id
			WHERE si.sprint_id = s.id AND ia.user_id = $%d
		))`, argIdx, argIdx)
	args = append(args, userID)

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT
			s.id, s.name, s.goal, s.status, s.start_date, s.end_date,
			p.id AS project_id, p.name AS project_name,
			(SELECT count(*) FROM sprint_issues si JOIN issues i ON i.id = si.issue_id
				JOIN issue_assignees ia ON ia.issue_id = i.id
				WHERE si.sprint_id = s.id AND ia.user_id = $%d) AS my_count,
			(SELECT coalesce(sum(CASE WHEN sg."group" = 'completed' THEN 1 ELSE 0 END), 0)
			 FROM sprint_issues si JOIN issues i ON i.id = si.issue_id
				JOIN states sg ON sg.id = i.state_id WHERE si.sprint_id = s.id) AS done_count,
			(SELECT count(*) FROM sprint_issues si WHERE si.sprint_id = s.id) AS total_count
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
