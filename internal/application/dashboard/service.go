// Package dashboard — 项目仪表盘应用服务。
package dashboard

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供仪表盘应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建仪表盘服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Dashboard Data ---

// GetDashboard 获取项目仪表盘完整数据。
// 优先使用快照数据，若过期 (>5min) 则实时查询。
func (s *Service) GetDashboard(ctx context.Context, wsID, projectID int64) (*DashboardData, error) {
	data := &DashboardData{
		Snapshots: map[string]any{},
	}

	// widget 配置
	widgets, err := s.getWidgets(ctx, projectID)
	if err != nil {
		return nil, err
	}
	data.Widgets = widgets

	// 各 widget 实时/快照数据（此处以实时查询为默认）
	for _, w := range widgets {
		if !w.IsVisible {
			continue
		}
		snap, err := s.refreshWidgetSnapshot(ctx, projectID, w.WidgetType)
		if err == nil && snap != nil {
			data.Snapshots[string(w.WidgetType)] = snap
		}
	}

	// 未解决风险告警
	alerts, err := s.getActiveAlerts(ctx, projectID)
	if err != nil {
		return nil, err
	}
	data.Alerts = alerts

	return data, nil
}

// getWidgets 获取项目 widget 配置（项目默认 + 用户覆盖合并）。
func (s *Service) getWidgets(ctx context.Context, projectID int64) ([]DashboardWidget, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, widget_type, title, grid_x, grid_y, grid_w, grid_h,
		       config, is_visible, sort_order, user_id, created_at, updated_at
		FROM dashboard_widgets
		WHERE project_id = $1
		ORDER BY sort_order`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var widgets []DashboardWidget
	for rows.Next() {
		var w DashboardWidget
		var cfgRaw []byte
		if err := rows.Scan(&w.ID, &w.ProjectID, &w.WidgetType, &w.Title,
			&w.GridX, &w.GridY, &w.GridW, &w.GridH,
			&cfgRaw, &w.IsVisible, &w.SortOrder, &w.UserID,
			&w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		_ = json.Unmarshal(cfgRaw, &w.Config)
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// refreshWidgetSnapshot 刷新单个 widget 的数据快照。
func (s *Service) refreshWidgetSnapshot(ctx context.Context, projectID int64, wt WidgetType) (any, error) {
	switch wt {
	case WidgetProgressOverview:
		return s.getProgressOverview(ctx, projectID)
	case WidgetPrioritySplit:
		return s.getPrioritySplit(ctx, projectID)
	case WidgetStateDistribution:
		return s.getStateDistribution(ctx, projectID)
	case WidgetOverdueList:
		return s.getOverdueList(ctx, projectID)
	case WidgetBlockedList:
		return s.getBlockedList(ctx, projectID)
	case WidgetBurndown:
		return s.getActiveSprintBurndown(ctx, projectID)
	case WidgetTeamWorkload:
		return s.getTeamWorkload(ctx, projectID)
	case WidgetRecentActivity:
		return s.getRecentActivity(ctx, projectID)
	case WidgetVelocity:
		return s.getVelocity(ctx, projectID)
	default:
		// 未知 widget: 尝试读取数据库快照
		var data map[string]any
		err := s.db.QueryRow(ctx,
			`SELECT data FROM dashboard_snapshots WHERE project_id = $1 AND widget_type = $2`,
			projectID, wt).Scan(&data)
		return data, err
	}
}

// getProgressOverview 进度概览。
func (s *Service) getProgressOverview(ctx context.Context, projectID int64) (any, error) {
	var w ProgressOverviewWidget
	err := s.db.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE i.deleted_at IS NULL) AS total,
			count(*) FILTER (WHERE sg."group" = 'completed' AND i.deleted_at IS NULL) AS done,
			count(*) FILTER (WHERE sg."group" = 'started' AND i.deleted_at IS NULL) AS in_progress,
			count(*) FILTER (WHERE i.target_date < CURRENT_DATE AND sg."group" NOT IN ('completed','cancelled') AND i.deleted_at IS NULL) AS overdue,
			count(*) FILTER (WHERE EXISTS (SELECT 1 FROM biz_entity_relation ir WHERE ir.source_id = i.id AND ir.relation_type = 'blocked_by') AND i.deleted_at IS NULL) AS blocked
		FROM (
		    SELECT id, state_id, project_id, deleted_at, target_date FROM task
		    UNION ALL
		    SELECT id, state_id, project_id, deleted_at, target_date FROM requirement
		    UNION ALL
		    SELECT id, state_id, project_id, deleted_at, target_date FROM defect
		) i
		JOIN states sg ON sg.id = i.state_id
		WHERE i.project_id = $1`,
		projectID).Scan(&w.TotalIssues, &w.DoneIssues, &w.InProgress, &w.OverdueIssues, &w.BlockedIssues)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if w.TotalIssues > 0 {
		w.CompletionRate = float64(w.DoneIssues) / float64(w.TotalIssues)
	}
	w.ActiveSprints = s.countActiveSprints(ctx, projectID)
	return &w, nil
}

// countActiveSprints 统计活跃迭代数。
func (s *Service) countActiveSprints(ctx context.Context, projectID int64) int {
	var count int
	_ = s.db.QueryRow(ctx,
		`SELECT count(*) FROM sprints WHERE project_id = $1 AND status = 'active' AND deleted_at IS NULL`,
		projectID).Scan(&count)
	return count
}

// getPrioritySplit 优先级分布。
func (s *Service) getPrioritySplit(ctx context.Context, projectID int64) (any, error) {
	w := PrioritySplitWidget{ByPriority: map[string]int{}}
	rows, err := s.db.Query(ctx, `
		SELECT i.priority, count(*) FROM (
		    SELECT priority, project_id, deleted_at FROM task
		    UNION ALL
		    SELECT priority, project_id, deleted_at FROM requirement
		    UNION ALL
		    SELECT priority, project_id, deleted_at FROM defect
		) i
		WHERE i.project_id = $1 AND i.deleted_at IS NULL
		GROUP BY i.priority ORDER BY i.priority`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if priority == "" {
			priority = "none"
		}
		w.ByPriority[priority] = count
		w.Total += count
	}
	return &w, rows.Err()
}

// getStateDistribution 状态分布。
func (s *Service) getStateDistribution(ctx context.Context, projectID int64) (any, error) {
	w := StateDistributionWidget{}
	rows, err := s.db.Query(ctx, `
		SELECT st.id, st.name, st."group", st.color, count(i.id)
		FROM states st
		LEFT JOIN issues i ON i.state_id = st.id AND i.project_id = $1 AND i.deleted_at IS NULL
		WHERE st.workspace_id = (SELECT workspace_id FROM projects WHERE id = $1) AND st.deleted_at IS NULL
		GROUP BY st.id, st.name, st."group", st.color
		ORDER BY st.sequence`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var b StateBucket
		if err := rows.Scan(&b.StateID, &b.StateName, &b.GroupName, &b.Color, &b.Count); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		w.ByState = append(w.ByState, b)
		w.Total += b.Count
	}
	return &w, rows.Err()
}

// getOverdueList 逾期列表（前 N 条）。
func (s *Service) getOverdueList(ctx context.Context, projectID int64) (any, error) {
	w := OverdueListWidget{}
	err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM issues i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.deleted_at IS NULL
			AND i.target_date < CURRENT_DATE AND st."group" NOT IN ('completed','cancelled')`,
		projectID).Scan(&w.Total)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.sequence_id::text, i.name, i.priority, CURRENT_DATE - i.target_date AS overdue_days,
		       (SELECT string_agg(u.display_name, ',') FROM issue_assignees ia JOIN users u ON u.id = ia.user_id WHERE ia.issue_id = i.id) AS assignee
		FROM issues i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.deleted_at IS NULL
			AND i.target_date < CURRENT_DATE AND st."group" NOT IN ('completed','cancelled')
		ORDER BY i.target_date ASC LIMIT 10`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var it OverdueItem
		if err := rows.Scan(&it.ID, &it.Identifier, &it.Title, &it.Priority, &it.OverdueDays, &it.Assignee); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		w.Items = append(w.Items, it)
	}
	return &w, rows.Err()
}

// getBlockedList 阻塞列表。
func (s *Service) getBlockedList(ctx context.Context, projectID int64) (any, error) {
	w := BlockedListWidget{}
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.sequence_id::text, i.name,
		       (SELECT count(*) FROM issue_relations ir WHERE ir.source_issue_id = i.id AND ir.relation_type = 'blocked_by') AS blocked_count,
		       (SELECT string_agg(u.display_name, ',') FROM issue_relations ir
		           JOIN issues blocker ON blocker.id = ir.target_issue_id
		           JOIN issue_assignees ia ON ia.issue_id = blocker.id
		           JOIN users u ON u.id = ia.user_id
		       WHERE ir.source_issue_id = i.id AND ir.relation_type = 'blocked_by') AS blockers
		FROM issues i
		WHERE i.project_id = $1 AND i.deleted_at IS NULL
			AND EXISTS (SELECT 1 FROM issue_relations ir WHERE ir.source_issue_id = i.id AND ir.relation_type = 'blocked_by')
		ORDER BY blocked_count DESC LIMIT 10`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var it BlockedItem
		if err := rows.Scan(&it.ID, &it.Identifier, &it.Title, &it.BlockedCount, &it.BlockerNames); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		w.Total++
		w.Items = append(w.Items, it)
	}
	return &w, rows.Err()
}

// getActiveSprintBurndown 获取当前活跃迭代的燃尽数据。
func (s *Service) getActiveSprintBurndown(ctx context.Context, projectID int64) (any, error) {
	var w BurndownWidget
	var endDate *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT s.id, s.name, s.end_date,
		       (SELECT COALESCE(sum(i2.point), 0) FROM sprint_issues si2 JOIN issues i2 ON i2.id = si2.issue_id
		           WHERE si2.sprint_id = s.id AND i2.deleted_at IS NULL) AS total_points,
		       (SELECT count(*) FROM sprint_issues si2 JOIN issues i2 ON i2.id = si2.issue_id
		           WHERE si2.sprint_id = s.id AND i2.deleted_at IS NULL) AS total_issues,
		       (SELECT count(*) FROM sprint_issues si2 JOIN issues i2 ON i2.id = si2.issue_id
		           JOIN states sg2 ON sg2.id = i2.state_id
		       WHERE si2.sprint_id = s.id AND sg2."group" = 'completed' AND i2.deleted_at IS NULL) AS burned_issues
		FROM sprints s WHERE s.project_id = $1 AND s.status = 'active' AND s.deleted_at IS NULL
		ORDER BY s.start_date DESC LIMIT 1`,
		projectID).Scan(&w.SprintID, &w.SprintName, &endDate, &w.TotalPoints, &w.TotalIssues, &w.BurnedIssues)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	w.BurnedPoints = w.BurnedIssues * (w.TotalPoints / max(w.TotalIssues, 1))
	if endDate != nil {
		w.RemainingDays = int(endDate.Sub(time.Now()).Hours() / 24)
	}
	w.IsActive = true
	return &w, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getTeamWorkload 按负责人分组统计未完成 issues。
func (s *Service) getTeamWorkload(ctx context.Context, projectID int64) (any, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.display_name, COALESCE(u.avatar_url, ''),
		       count(*) FILTER (WHERE sg."group" IN ('backlog','unstarted')) AS todo,
		       count(*) FILTER (WHERE sg."group" = 'started') AS in_progress,
		       count(*) FILTER (WHERE sg."group" = 'completed') AS done,
		       count(*) AS total
		FROM issue_assignees ia
		JOIN issues i ON i.id = ia.issue_id AND i.project_id = $1 AND i.deleted_at IS NULL
		JOIN states sg ON sg.id = i.state_id
		JOIN users u ON u.id = ia.user_id
		GROUP BY u.id, u.display_name, u.avatar_url
		ORDER BY total DESC
		LIMIT 20`,
		projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	w := TeamWorkloadWidget{}
	for rows.Next() {
		var m TeamMemberWorkload
		if err := rows.Scan(&m.UserID, &m.UserName, &m.Avatar, &m.Todo, &m.InProgress, &m.Done, &m.Total); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		w.Members = append(w.Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if len(w.Members) == 0 {
		return nil, nil
	}
	return &w, nil
}

// getRecentActivity 查询项目最近 20 条活动记录。
func (s *Service) getRecentActivity(ctx context.Context, projectID int64) (any, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.issue_id, i.sequence_id::text, a.actor_id, u.display_name,
		       COALESCE(u.avatar_url, '') AS actor_avatar,
		       a.verb, COALESCE(s.name, '') AS target_state, a.created_at
		FROM issue_activities a
		JOIN issues i ON i.id = a.issue_id
		LEFT JOIN users u ON u.id = a.actor_id
		LEFT JOIN states s ON s.id = i.state_id
		WHERE a.project_id = $1
		ORDER BY a.created_at DESC
		LIMIT 20`,
		projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	w := RecentActivityWidget{}
	for rows.Next() {
		var it ActivityItem
		var createdAt time.Time
		if err := rows.Scan(&it.ID, &it.IssueID, &it.IssueIdentifier, &it.ActorID,
			&it.ActorName, &it.ActorAvatar, &it.Verb, &it.TargetState, &createdAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		verbLabel := it.Verb
		switch it.Verb {
		case "created":
			verbLabel = "创建了"
		case "transitioned":
			verbLabel = "流转到"
		case "commented":
			verbLabel = "评论了"
		case "updated":
			verbLabel = "更新了"
		case "attached":
			verbLabel = "附件"
		default:
			verbLabel = it.Verb
		}
		it.Verb = verbLabel
		it.CreatedAt = createdAt.Format(time.RFC3339)
		w.Items = append(w.Items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if len(w.Items) == 0 {
		return nil, nil
	}
	return &w, nil
}

// getVelocity 查询最近 5 个已完成迭代的工作速率。
func (s *Service) getVelocity(ctx context.Context, projectID int64) (any, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.name,
		       (SELECT count(*) FROM sprint_issues si
		           JOIN issues i ON i.id = si.issue_id
		           JOIN states sg ON sg.id = i.state_id
		       WHERE si.sprint_id = s.id AND sg."group" = 'completed'
		           AND i.deleted_at IS NULL) AS completed_count,
		       (SELECT count(*) FROM sprint_issues si2
		           JOIN issues i2 ON i2.id = si2.issue_id
		       WHERE si2.sprint_id = s.id AND i2.deleted_at IS NULL) AS committed_count
		FROM sprints s
		WHERE s.project_id = $1 AND s.status = 'completed' AND s.deleted_at IS NULL
		ORDER BY s.completed_at DESC
		LIMIT 5`,
		projectID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	w := VelocityWidget{}
	var totalRate float64
	var count int
	for rows.Next() {
		var p VelocityPoint
		if err := rows.Scan(&p.SprintID, &p.SprintName, &p.CompletedCount, &p.CommittedCount); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if p.CommittedCount > 0 {
			p.CompletionRate = float64(p.CompletedCount) / float64(p.CommittedCount)
		}
		w.Sprints = append(w.Sprints, p)
		totalRate += p.CompletionRate
		count++
	}
	if count > 0 {
		w.Average = totalRate / float64(count)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if len(w.Sprints) == 0 {
		return nil, nil
	}
	return &w, nil
}

// --- Risk Alerts ---

// getActiveAlerts 获取未解决风险告警。
func (s *Service) getActiveAlerts(ctx context.Context, projectID int64) ([]RiskAlert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, rule_id, severity, title, description, metadata, is_resolved, created_at
		FROM risk_alerts
		WHERE project_id = $1 AND NOT is_resolved
		ORDER BY
			CASE severity WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END,
			created_at DESC
		LIMIT 20`,
		projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// GetActiveUnresolvedAlerts 跨项目获取未解决告警。
func (s *Service) GetActiveUnresolvedAlerts(ctx context.Context, wsID int64) ([]RiskAlert, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, rule_id, severity, title, description, metadata, is_resolved, created_at
		FROM risk_alerts
		WHERE workspace_id = $1 AND NOT is_resolved
		ORDER BY created_at DESC LIMIT 50`,
		wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	return scanAlerts(rows)
}

func scanAlerts(rows pgx.Rows) ([]RiskAlert, error) {
	var alerts []RiskAlert
	for rows.Next() {
		var a RiskAlert
		var metaRaw []byte
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.RuleID, &a.Severity,
			&a.Title, &a.Description, &metaRaw, &a.IsResolved, &a.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		_ = json.Unmarshal(metaRaw, &a.Metadata)
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// --- Templates ---

// ListTemplates 列出仪表盘模板。
func (s *Service) ListTemplates(ctx context.Context, category string) ([]DashboardTemplate, error) {
	var rows pgx.Rows
	var err error
	if category != "" {
		rows, err = s.db.Query(ctx,
			`SELECT id, name, slug, description, layout, icon, category, is_default, sort_order
			FROM dashboard_templates WHERE category = $1 ORDER BY sort_order`,
			category)
	} else {
		rows, err = s.db.Query(ctx,
			`SELECT id, name, slug, description, layout, icon, category, is_default, sort_order
			FROM dashboard_templates ORDER BY sort_order`)
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var templates []DashboardTemplate
	for rows.Next() {
		var t DashboardTemplate
		var layoutRaw []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description,
			&layoutRaw, &t.Icon, &t.Category, &t.IsDefault, &t.SortOrder); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		_ = json.Unmarshal(layoutRaw, &t.Layout)
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// --- Widget save ---

// SaveWidget 保存 widget 配置。
func (s *Service) SaveWidget(ctx context.Context, in SaveWidgetInput) (*DashboardWidget, error) {
	cfgJSON, _ := json.Marshal(in.Config)
	var w DashboardWidget
	err := s.db.QueryRow(ctx, `
		INSERT INTO dashboard_widgets (project_id, widget_type, title, grid_x, grid_y, grid_w, grid_h, config, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (project_id, widget_type, user_id)
		DO UPDATE SET title = EXCLUDED.title, grid_x = EXCLUDED.grid_x, grid_y = EXCLUDED.grid_y,
			grid_w = EXCLUDED.grid_w, grid_h = EXCLUDED.grid_h, config = EXCLUDED.config, updated_at = now()
		RETURNING id, project_id, widget_type, title, grid_x, grid_y, grid_w, grid_h,
		          config, is_visible, sort_order, user_id, created_at, updated_at`,
		in.ProjectID, in.WidgetType, in.Title, in.GridX, in.GridY, in.GridW, in.GridH, cfgJSON, in.GridY*10+in.GridX).Scan(
		&w.ID, &w.ProjectID, &w.WidgetType, &w.Title,
		&w.GridX, &w.GridY, &w.GridW, &w.GridH,
		&w.Config, &w.IsVisible, &w.SortOrder, &w.UserID,
		&w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &w, nil
}

// DeleteWidget 删除 widget 配置。
func (s *Service) DeleteWidget(ctx context.Context, widgetID, projectID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM dashboard_widgets WHERE id = $1 AND project_id = $2`,
		widgetID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- Resolve Alert ---

// ResolveAlert 解决告警。
func (s *Service) ResolveAlert(ctx context.Context, wsID, alertID, userID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE risk_alerts SET is_resolved = TRUE, resolved_at = now(), resolved_by = $3
		WHERE id = $1 AND workspace_id = $2 AND NOT is_resolved`,
		alertID, wsID, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}
