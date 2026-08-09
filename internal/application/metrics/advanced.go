// Package metrics — 高级效能度量指标（CFD/控制图/资源负载）。
//
// 对标参考:
//   - DORA 《Accelerate State of DevOps》CFD 标准
//   - 美团研发效能度量规范 v2.0（控制图控制线）
//   - 阿里云效能 CFD 渲染规范
//
// 核心指标:
//   - CFD (累积流图): 按日期分桶统计各状态组工作项数量
//   - 控制图 (Control Chart): 前置时间 P50/P85/P95 + 控制限
//   - 资源负载细化: 每位成员 WIP 分布
//   - 吞吐量: 每周完成工作项数（对齐 SAFe 规范）
package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// --- CFD (Cumulative Flow Diagram) ---

// CFDDataPoint CFD 单条数据点。
type CFDDataPoint struct {
	Date        string `json:"date"`         // YYYY-MM-DD
	Backlog     int    `json:"backlog"`      // state_group=todo
	Todo        int    `json:"todo"`         // state_group=todo (排除 backlog)
	InProgress  int    `json:"in_progress"`  // state_group=started
	Done        int    `json:"done"`         // state_group=completed
	Cancelled   int    `json:"cancelled"`    // state_group=cancelled
	TotalActive int    `json:"total_active"` // 未归档的非 cancelled 总数
}

// CFDCalculator CFD 计算器。
type CFDCalculator struct {
	db *pgxpool.Pool
}

// NewCFDCalculator 创建 CFD 计算器。
func NewCFDCalculator(db *pgxpool.Pool) *CFDCalculator {
	return &CFDCalculator{db: db}
}

// Calculate 计算项目在指定日期范围内的 CFD 数据。
// 返回按日期升序排列的 CFD 数据点列表。
func (c *CFDCalculator) Calculate(ctx context.Context, wsID, projectID int64, fromDate, toDate time.Time) ([]CFDDataPoint, error) {
	// 构建每日快照的时间序列
	rows, err := c.db.Query(ctx, `
		WITH date_series AS (
			SELECT generate_series($3::date, $4::date, '1 day'::interval)::date AS snap_date
		),
		daily_counts AS (
			SELECT
				ds.snap_date,
				count(*) FILTER (WHERE st."group" = 'backlog') AS backlog,
				count(*) FILTER (WHERE st."group" = 'todo') AS todo,
				count(*) FILTER (WHERE st."group" = 'started') AS in_progress,
				count(*) FILTER (WHERE st."group" = 'completed' AND i.completed_at <= ds.snap_date + INTERVAL '1 day') AS done,
				count(*) FILTER (WHERE st."group" = 'cancelled') AS cancelled,
				count(*) FILTER (WHERE i.deleted_at IS NULL AND st."group" != 'cancelled') AS total_active
			FROM date_series ds
			LEFT JOIN (
				SELECT id, state_id, project_id, workspace_id, created_at FROM requirement WHERE deleted_at IS NULL
				UNION ALL SELECT id, state_id, project_id, workspace_id, created_at FROM task WHERE deleted_at IS NULL
				UNION ALL SELECT id, state_id, project_id, workspace_id, created_at FROM defect WHERE deleted_at IS NULL
			) i ON i.project_id = $1 AND i.workspace_id = $2
				AND i.created_at <= ds.snap_date + INTERVAL '1 day'
			LEFT JOIN states st ON st.id = i.state_id
			GROUP BY ds.snap_date
			ORDER BY ds.snap_date ASC
		)
		SELECT snap_date, backlog, todo, in_progress, done, cancelled, total_active
		FROM daily_counts`,
		projectID, wsID, fromDate, toDate)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var points []CFDDataPoint
	for rows.Next() {
		var p CFDDataPoint
		if err := rows.Scan(&p.Date, &p.Backlog, &p.Todo, &p.InProgress, &p.Done, &p.Cancelled, &p.TotalActive); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

// --- Control Chart ---

// ControlChartPoint 控制图单条数据点。
type ControlChartPoint struct {
	Date      string  `json:"date"`       // 完成日期
	LeadDays  float64 `json:"lead_days"`  // 该工作项的前置时间
	EpisodeID string  `json:"episode_id"` // 工作项 identifier
}

// ControlChartResult 控制图完整结果。
type ControlChartResult struct {
	Points       []ControlChartPoint `json:"points"`
	P50          float64             `json:"p50"`
	P85          float64             `json:"p85"`
	P95          float64             `json:"p95"`
	UpperControl float64             `json:"upper_control_limit"` // UCL = P85 * 1.5
	MovingAvg    []MovingAverage     `json:"moving_avg_7d"`       // 7 点移动均线
}

// MovingAverage 移动均线数据点。
type MovingAverage struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// ControlChartCalculator 前置时间控制图计算器。
type ControlChartCalculator struct {
	db *pgxpool.Pool
}

// NewControlChartCalculator 创建控制图计算器。
func NewControlChartCalculator(db *pgxpool.Pool) *ControlChartCalculator {
	return &ControlChartCalculator{db: db}
}

// Calculate 计算需求前置时间的控制图（按完成日期排序）。
func (c *ControlChartCalculator) Calculate(ctx context.Context, wsID, projectID int64, days int) (*ControlChartResult, error) {
	if days <= 0 {
		days = 90
	}
	since := time.Now().AddDate(0, 0, -days)

	rows, err := c.db.Query(ctx, `
		SELECT i.identifier, i.completed_at::date, 
		       EXTRACT(EPOCH FROM (i.completed_at - i.created_at)) / 86400.0 AS lead_days
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.workspace_id = $2
		  AND i.type_code = 'requirement'
		  AND st."group" = 'completed'
		  AND i.completed_at >= $3
		  AND i.deleted_at IS NULL
		ORDER BY i.completed_at ASC`,
		projectID, wsID, since)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	result := &ControlChartResult{}
	var allLeadDays []float64
	for rows.Next() {
		var p ControlChartPoint
		if err := rows.Scan(&p.EpisodeID, &p.Date, &p.LeadDays); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result.Points = append(result.Points, p)
		allLeadDays = append(allLeadDays, p.LeadDays)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if len(allLeadDays) > 0 {
		result.P50 = percentile(allLeadDays, 50)
		result.P85 = percentile(allLeadDays, 85)
		result.P95 = percentile(allLeadDays, 95)
		result.UpperControl = result.P85 * 1.5
		result.MovingAvg = computeMovingAverage(result.Points, 7)
	}
	return result, nil
}

// computeMoving_average 计算 7 点移动均线（前置时间局部趋势）。
func computeMovingAverage(points []ControlChartPoint, window int) []MovingAverage {
	if len(points) < window {
		return nil
	}
	var ma []MovingAverage
	for i := window - 1; i < len(points); i++ {
		var sum float64
		for j := i - window + 1; j <= i; j++ {
			sum += points[j].LeadDays
		}
		ma = append(ma, MovingAverage{
			Date:  points[i].Date,
			Value: sum / float64(window),
		})
	}
	return ma
}

// --- Resource Load Detail ---

// MemberLoad 单个成员的资源负载。
type MemberLoad struct {
	UserID       int64   `json:"user_id"`
	UserName     string  `json:"user_name"`
	ActiveIssues int     `json:"active_issues"`
	TotalPoints  int     `json:"total_points"`
	LeadTimeAvg  float64 `json:"lead_time_avg_days"`
}

// ResourceLoadDetail 项目成员级资源负载明细。
type ResourceLoadDetail struct {
	ProjectID    int64        `json:"project_id"`
	Members      []MemberLoad `json:"members"`
	TotalWIP     int          `json:"total_wip"`
	AvgWIPPerMem float64      `json:"avg_wip_per_member"`
	MaxWIP       int          `json:"max_wip"`
	Imbalance    float64      `json:"imbalance_ratio"` // 最大 WIP / 平均 WIP（>2 告警）
}

// GetResourceLoadDetail 查询项目成员级资源负载。
func (s *Service) GetResourceLoadDetail(ctx context.Context, wsID, projectID int64) (*ResourceLoadDetail, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.display_name,
		       COALESCE(load.active_cnt, 0),
		       COALESCE(load.total_pts, 0)
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		LEFT JOIN LATERAL (
			SELECT count(*) AS active_cnt, coalesce(sum(sub.point), 0) AS total_pts
			FROM issue_assignees ia
			JOIN (SELECT id, state_id, point, project_id FROM requirement WHERE deleted_at IS NULL
			    UNION ALL SELECT id, state_id, point, project_id FROM task WHERE deleted_at IS NULL
			    UNION ALL SELECT id, state_id, point, project_id FROM defect WHERE deleted_at IS NULL) sub ON sub.id = ia.issue_id
			JOIN states st ON st.id = sub.state_id
			WHERE ia.user_id = pm.user_id AND sub.project_id = $1
			  AND st."group" NOT IN ('completed', 'cancelled')
		) load ON true
		WHERE pm.project_id = $1 AND pm.workspace_id = $2
		ORDER BY load.active_cnt DESC NULLS LAST`,
		projectID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	result := &ResourceLoadDetail{ProjectID: projectID}
	for rows.Next() {
		var m MemberLoad
		if err := rows.Scan(&m.UserID, &m.UserName, &m.ActiveIssues, &m.TotalPoints); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result.Members = append(result.Members, m)
		result.TotalWIP += m.ActiveIssues
		if m.ActiveIssues > result.MaxWIP {
			result.MaxWIP = m.ActiveIssues
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if len(result.Members) > 0 {
		result.AvgWIPPerMem = float64(result.TotalWIP) / float64(len(result.Members))
		if result.AvgWIPPerMem > 0 {
			result.Imbalance = float64(result.MaxWIP) / result.AvgWIPPerMem
		}
	}
	return result, nil
}

// --- Weekly Throughput ---

// WeeklyThroughput 周吞吐量统计。
type WeeklyThroughput struct {
	WeekStart string `json:"week_start"` // YYYY-MM-DD (周一)
	WeekEnd   string `json:"week_end"`   // YYYY-MM-DD (周日)
	Completed int    `json:"completed"`  // 完成需求数
	Points    int    `json:"points"`     // 完成故事点
}

// GetWeeklyThroughput 查询项目近 N 周的吞吐量。
func (s *Service) GetWeeklyThroughput(ctx context.Context, wsID, projectID int64, weeks int) ([]WeeklyThroughput, error) {
	if weeks <= 0 || weeks > 52 {
		weeks = 12
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			date_trunc('week', i.completed_at)::date AS week_start,
			(date_trunc('week', i.completed_at)::date + INTERVAL '6 days')::date AS week_end,
			count(*) AS completed,
			coalesce(sum(i.point), 0) AS points
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.workspace_id = $2
		  AND i.type_code = 'requirement'
		  AND st."group" = 'completed'
		  AND i.completed_at >= now() - ($3 || ' weeks')::interval
		  AND i.deleted_at IS NULL
		GROUP BY date_trunc('week', i.completed_at)
		ORDER BY week_start DESC`,
		projectID, wsID, fmt.Sprint(weeks))
	if err != nil && err != pgx.ErrNoRows {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []WeeklyThroughput
	for rows.Next() {
		var w WeeklyThroughput
		if err := rows.Scan(&w.WeekStart, &w.WeekEnd, &w.Completed, &w.Points); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, w)
	}
	return result, rows.Err()
}

// 兼容 sql.NullFloat64
var _ = sql.NullFloat64{}
