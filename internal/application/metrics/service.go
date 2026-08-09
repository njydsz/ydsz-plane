// Package metrics — 效能度量应用服务。
//
// 参考 docs/architecture/11-仪表盘与效能度量设计.md §2 效能度量
// 指标口径定义（冻结）：
//   - 速度 Velocity: 每迭代完成的故事点（sprint_snapshots 聚合）
//   - 前置时间 Lead Time: 需求 created_at → completed_at（自然日 P50/P85）
//   - 缺陷密度 Defect Density: 缺陷数 / Σ需求故事点
//   - 逃逸率 Escape Rate: found_phase=production|customer 缺陷占比
//   - DORA: 基于 deployment_events 计算 DF/LT/CFR/MTTR
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

// Service 提供效能度量应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建效能度量服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Metric Name Constants ---

const (
	MetricVelocity      = "velocity"
	MetricLeadTimeP50   = "lead_time_p50"
	MetricLeadTimeP85   = "lead_time_p85"
	MetricThroughput    = "throughput"
	MetricWIP           = "wip"
	MetricDefectDensity = "defect_density"
	MetricEscapeRate    = "escape_rate"
	MetricReworkRate    = "rework_rate"
	MetricDORADF        = "dora_deployment_frequency"
	MetricDORALT        = "dora_lead_time"
	MetricDORACFR       = "dora_change_failure_rate"
	MetricDORAMTTR      = "dora_mttr"
)

// --- Velocity ---

// VelocityResult 迭代速率统计。
type VelocityResult struct {
	ProjectID    int64            `json:"project_id"`
	Average      float64          `json:"average_points"`
	SprintCount  int              `json:"sprint_count"`
	LastSprintID *int64           `json:"last_sprint_id,omitempty"`
	LastPoints   int              `json:"last_points"`
	Trend        []SprintVelocity `json:"trend,omitempty"`
}

// SprintVelocity 单次迭代速率。
type SprintVelocity struct {
	SprintID   int64   `json:"sprint_id"`
	SprintName string  `json:"sprint_name"`
	PointsDone float64 `json:"points_done"`
	IssuesDone int     `json:"issues_done"`
}

// GetVelocity 查询项目近 N 个迭代的速率。
func (s *Service) GetVelocity(ctx context.Context, wsID, projectID int64, lastN int) (*VelocityResult, error) {
	if lastN <= 0 || lastN > 20 {
		lastN = 6
	}

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.name,
		       COALESCE((s.review_snapshot->>'completed_points')::numeric, 0) AS done_points,
		       COALESCE((s.review_snapshot->>'completed_issues')::int, 0) AS done_issues
		FROM sprints s
		WHERE s.project_id = $1 AND s.workspace_id = $2 AND s.status = 'completed' AND s.deleted_at IS NULL
		ORDER BY s.completed_at DESC LIMIT $3`,
		projectID, wsID, lastN)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	result := &VelocityResult{ProjectID: projectID}
	var totalPoints float64
	for rows.Next() {
		var sv SprintVelocity
		if err := rows.Scan(&sv.SprintID, &sv.SprintName, &sv.PointsDone, &sv.IssuesDone); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result.Trend = append(result.Trend, sv)
		totalPoints += sv.PointsDone
		result.SprintCount++
		if result.LastSprintID == nil {
			id := sv.SprintID
			result.LastSprintID = &id
			result.LastPoints = sv.IssuesDone
		}
	}
	if result.SprintCount > 0 {
		result.Average = totalPoints / float64(result.SprintCount)
	}
	return result, rows.Err()
}

// --- Lead Time ---

// LeadTimeResult 前置时间统计。
type LeadTimeResult struct {
	ProjectID  int64   `json:"project_id"`
	P50Days    float64 `json:"p50_days"`
	P85Days    float64 `json:"p85_days"`
	SampleSize int     `json:"sample_size"`
}

// GetLeadTime 计算需求前置时间（created_at → completed_at）。
func (s *Service) GetLeadTime(ctx context.Context, wsID, projectID int64, days int) (*LeadTimeResult, error) {
	if days <= 0 {
		days = 90
	}
	since := time.Now().AddDate(0, 0, -days)

	rows, err := s.db.Query(ctx, `
		SELECT EXTRACT(EPOCH FROM (i.completed_at - i.created_at)) / 86400.0 AS lead_days
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.workspace_id = $2
		  AND i.type_code = 'requirement'
		  AND st."group" = 'completed'
		  AND i.completed_at IS NOT NULL
		  AND i.created_at >= $3
		  AND i.deleted_at IS NULL
		ORDER BY lead_days ASC`,
		projectID, wsID, since)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var leadDays []float64
	for rows.Next() {
		var d float64
		if err := rows.Scan(&d); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		leadDays = append(leadDays, d)
	}

	result := &LeadTimeResult{ProjectID: projectID, SampleSize: len(leadDays)}
	if len(leadDays) > 0 {
		result.P50Days = percentile(leadDays, 50)
		result.P85Days = percentile(leadDays, 85)
	}
	return result, rows.Err()
}

// --- Quality Metrics ---

// QualityMetrics 质量指标聚合。
type QualityMetrics struct {
	ProjectID      int64   `json:"project_id"`
	TotalDefects   int     `json:"total_defects"`
	EscapedDefects int     `json:"escaped_defects"`
	EscapeRate     float64 `json:"escape_rate"`
	DefectDensity  float64 `json:"defect_density"`
	AvgDefectAge   float64 `json:"avg_defect_age_days"`
	ReopenedCount  int     `json:"reopened_count"`
	ReopenRate     float64 `json:"reopen_rate"`
}

// GetQualityMetrics 计算质量指标。
func (s *Service) GetQualityMetrics(ctx context.Context, wsID, projectID int64) (*QualityMetrics, error) {
	var m QualityMetrics
	m.ProjectID = projectID

	// 缺陷总数 + 逃逸数 + 龄
	err := s.db.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE i.deleted_at IS NULL),
			count(*) FILTER (WHERE i.deleted_at IS NULL AND d.found_phase IN ('production','customer')),
			coalesce(avg(EXTRACT(EPOCH FROM (now() - i.created_at)) / 86400) FILTER (WHERE st."group" NOT IN ('completed','cancelled')), 0),
			count(*) FILTER (WHERE d.reopened_at IS NOT NULL)
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
		JOIN states st ON st.id = i.state_id
		LEFT JOIN defect_extra d ON d.issue_id = i.id
		WHERE i.project_id = $1 AND i.workspace_id = $2 AND i.type_code = 'defect'`,
		projectID, wsID).Scan(&m.TotalDefects, &m.EscapedDefects, &m.AvgDefectAge, &m.ReopenedCount)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if m.TotalDefects > 0 {
		m.EscapeRate = float64(m.EscapedDefects) / float64(m.TotalDefects)
		m.ReopenRate = float64(m.ReopenedCount) / float64(m.TotalDefects)
	}

	// 缺陷密度 = 缺陷数 / 需求故事点
	var totalPoints sql.NullFloat64
	err = s.db.QueryRow(ctx, `
		SELECT coalesce(sum(estimate_points), 0)
		FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
		WHERE i.project_id = $1 AND i.workspace_id = $2 AND i.type_code = 'requirement' AND i.deleted_at IS NULL`,
		projectID, wsID).Scan(&totalPoints)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if totalPoints.Valid && totalPoints.Float64 > 0 {
		m.DefectDensity = float64(m.TotalDefects) / totalPoints.Float64
	}

	return &m, nil
}

// --- DORA ---

// DORA 指标结果。
type DORAResult struct {
	ProjectID           int64   `json:"project_id"`
	DeploymentFrequency float64 `json:"deployment_freq_per_day"`     // DF: 每日部署次数（30 天窗口）
	LeadTimeForChanges  float64 `json:"lead_time_for_changes_hours"` // LTC: median commit→deploy
	ChangeFailureRate   float64 `json:"change_failure_rate"`         // CFR: 失败部署占比
	MTTR                float64 `json:"mttr_hours"`                  // MTTR: 故障恢复中位时长
	Level               string  `json:"performance_level"`           // elite / high / medium / low
}

// GetDORA 计算 DORA 四指标（30 天窗口）。
func (s *Service) GetDORA(ctx context.Context, wsID, projectID int64) (*DORAResult, error) {
	r := &DORAResult{ProjectID: projectID}

	// DF: 过去 30 天成功部署次数 / 30
	var successCount int
	err := s.db.QueryRow(ctx, `
		SELECT count(*) FROM deployment_events
		WHERE project_id = $1 AND workspace_id = $2 AND status = 'success'
		  AND deployed_at >= now() - INTERVAL '30 days'`,
		projectID, wsID).Scan(&successCount)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errs.ErrInternal.Wrap(err)
	}
	r.DeploymentFrequency = float64(successCount) / 30.0

	// CFR: 失败部署 / 全部部署
	var totalDeployments, failedDeployments int
	err = s.db.QueryRow(ctx, `
		SELECT
			count(*),
			count(*) FILTER (WHERE status = 'failed')
		FROM deployment_events
		WHERE project_id = $1 AND workspace_id = $2 AND deployed_at >= now() - INTERVAL '30 days'`,
		projectID, wsID).Scan(&totalDeployments, &failedDeployments)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if totalDeployments > 0 {
		r.ChangeFailureRate = float64(failedDeployments) / float64(totalDeployments)
	}

	// LTC: commit→deploy 中位数（小时）
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (deployed_at - started_at)) / 3600), 0)
		FROM deployment_events
		WHERE project_id = $1 AND workspace_id = $2 AND status = 'success'
		  AND started_at IS NOT NULL AND deployed_at IS NOT NULL
		  AND deployed_at >= now() - INTERVAL '30 days'`,
		projectID, wsID).Scan(&r.LeadTimeForChanges)
	if err != nil && err != pgx.ErrNoRows {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// MTTR: 故障→恢复中位时长（通过 rolled_back 事件估算）
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (r2.deployed_at - r1.deployed_at)) / 3600), 0)
		FROM deployment_events r1
		JOIN deployment_events r2 ON r2.project_id = r1.project_id
			AND r2.deployed_at > r1.deployed_at AND r2.status = 'success'
		WHERE r1.project_id = $1 AND r1.workspace_id = $2 AND r1.status = 'failed'`,
		projectID, wsID).Scan(&r.MTTR)
	if err != nil && err != pgx.ErrNoRows {
		// MTTR 查询复杂，失败不阻塞整体
		r.MTTR = 0
	}

	r.Level = classifyDORALevel(r)
	return r, nil
}

// classifyDORALevel 根据 DORA 指标分级。
// 基准: https://dora.dev/research/2023-accelerate-state-of-devops/
func classifyDORALevel(r *DORAResult) string {
	score := 0
	switch {
	case r.DeploymentFrequency >= 1.0: // 每日 ≥1 次
		score += 2
	case r.DeploymentFrequency >= 1.0/7.0: // 每周 ≥1 次
		score += 1
	}
	switch {
	case r.LeadTimeForChanges <= 1.0: // < 1 小时
		score += 2
	case r.LeadTimeForChanges <= 24.0: // < 1 天
		score += 1
	}
	switch {
	case r.ChangeFailureRate < 0.05: // < 5%
		score += 2
	case r.ChangeFailureRate < 0.15: // < 15%
		score += 1
	}
	switch {
	case r.MTTR <= 1.0: // < 1 小时
		score += 2
	case r.MTTR <= 24.0: // < 1 天
		score += 1
	}

	switch {
	case score >= 7:
		return "elite"
	case score >= 5:
		return "high"
	case score >= 3:
		return "medium"
	default:
		return "low"
	}
}

// --- Daily Snapshot Aggregation ---

// AggregateDailySnapshots 执行每日快照聚合 Job。
// 幂等：snapshot_date + ON CONFLICT DO UPDATE。
func (s *Service) AggregateDailySnapshots(ctx context.Context, snapshotDate string) (int, error) {
	projects, err := s.db.Query(ctx,
		`SELECT id, workspace_id FROM projects WHERE deleted_at IS NULL AND status = 'active'`)
	if err != nil {
		return 0, fmt.Errorf("list projects: %w", err)
	}
	defer projects.Close()

	totalWritten := 0
	for projects.Next() {
		var projectID, wsID int64
		if err := projects.Scan(&projectID, &wsID); err != nil {
			continue
		}

		// 写入 WIP 快照（今日活跃 sprint WIP）
		var wipCount int
		_ = s.db.QueryRow(ctx,
			`SELECT count(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i JOIN sprint_issues si ON si.issue_id = i.id
			 JOIN sprints sp ON sp.id = si.sprint_id JOIN states st ON st.id = i.state_id
			 WHERE i.project_id = $1 AND sp.status = 'active' AND st."group" = 'started'
			 AND i.deleted_at IS NULL`,
			projectID).Scan(&wipCount)
		if err := s.writeSnapshot(ctx, wsID, projectID, MetricWIP, float64(wipCount), nil, snapshotDate); err == nil {
			totalWritten++
		}

		// Velocity（近 6 个完成迭代均值）
		if vel, err := s.GetVelocity(ctx, wsID, projectID, 6); err == nil && vel.SprintCount > 0 {
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricVelocity, vel.Average,
				map[string]any{"sprint_count": vel.SprintCount}, snapshotDate); err == nil {
				totalWritten++
			}
		}

		// Lead Time P50 / P85（过去 90 天完成的需求）
		if lt, err := s.GetLeadTime(ctx, wsID, projectID, 90); err == nil && lt.SampleSize > 0 {
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricLeadTimeP50, lt.P50Days,
				map[string]any{"sample_size": lt.SampleSize}, snapshotDate); err == nil {
				totalWritten++
			}
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricLeadTimeP85, lt.P85Days, nil, snapshotDate); err == nil {
				totalWritten++
			}
		}

		// Throughput（过去 30 天完成 issue 的日均值）
		var throughput float64
		if err := s.db.QueryRow(ctx, `
			SELECT COALESCE(count(*)::numeric / 30.0, 0)
			FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
			JOIN states st ON st.id = i.state_id
			WHERE i.project_id = $1 AND i.workspace_id = $2
			  AND st."group" = 'completed' AND i.completed_at IS NOT NULL
			  AND i.completed_at >= now() - INTERVAL '30 days' AND i.deleted_at IS NULL`,
			projectID, wsID).Scan(&throughput); err == nil && throughput > 0 {
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricThroughput, throughput, nil, snapshotDate); err == nil {
				totalWritten++
			}
		}

		// 质量指标：缺陷密度 / 逃逸率 / 返工率
		if qm, err := s.GetQualityMetrics(ctx, wsID, projectID); err == nil {
			if qm.TotalDefects > 0 {
				if err := s.writeSnapshot(ctx, wsID, projectID, MetricDefectDensity, qm.DefectDensity, nil, snapshotDate); err == nil {
					totalWritten++
				}
				if err := s.writeSnapshot(ctx, wsID, projectID, MetricEscapeRate, qm.EscapeRate, nil, snapshotDate); err == nil {
					totalWritten++
				}
				if err := s.writeSnapshot(ctx, wsID, projectID, MetricReworkRate, qm.ReopenRate, nil, snapshotDate); err == nil {
					totalWritten++
				}
			}
		}

		// DORA 四指标（基于 deployment_events 30 天窗口）
		if dora, err := s.GetDORA(ctx, wsID, projectID); err == nil {
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricDORADF, dora.DeploymentFrequency, nil, snapshotDate); err == nil {
				totalWritten++
			}
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricDORALT, dora.LeadTimeForChanges, nil, snapshotDate); err == nil {
				totalWritten++
			}
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricDORACFR, dora.ChangeFailureRate, nil, snapshotDate); err == nil {
				totalWritten++
			}
			if err := s.writeSnapshot(ctx, wsID, projectID, MetricDORAMTTR, dora.MTTR,
				map[string]any{"performance_level": dora.Level}, snapshotDate); err == nil {
				totalWritten++
			}
		}
	}

	return totalWritten, nil
}

// writeSnapshot 写一条 metric_snapshots 记录（幂等：ON CONFLICT DO UPDATE）。
func (s *Service) writeSnapshot(ctx context.Context, wsID, projectID int64, metric string, value float64, dims map[string]any, snapshotDate string) error {
	if dims == nil {
		dims = map[string]any{}
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO metric_snapshots (workspace_id, project_id, granularity, ref_id, metric, value, dimensions, snapshot_date)
		VALUES ($1, $2, 'daily', NULL, $3, $4, $5, $6)
		ON CONFLICT (workspace_id, project_id, granularity, ref_id, metric, snapshot_date)
		DO UPDATE SET value = $4, dimensions = $5`,
		wsID, projectID, metric, value, dims, snapshotDate)
	return err
}

// RecordDeploymentEvent 记录外部 CI/CD 推送的部署事件（DORA 数据源）。
func (s *Service) RecordDeploymentEvent(ctx context.Context, wsID, projectID int64, env, status, source, commitSha string, startedAt, deployedAt time.Time) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO deployment_events (workspace_id, project_id, env, status, source, commit_sha, started_at, deployed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (deployment_id, env, project_id) WHERE deployment_id IS NOT NULL
		DO UPDATE SET status = EXCLUDED.status, deployed_at = EXCLUDED.deployed_at
		RETURNING id`,
		wsID, projectID, env, status, source, commitSha, startedAt, deployedAt).Scan(&id)
	return id, err
}

// --- Helpers ---

// percentile 返回有序切片的第 p 百分位数（p ∈ [0,100]）。
func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := float64(p) / 100.0 * float64(len(sorted)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[lo]
	}
	t := idx - float64(lo)
	return sorted[lo]*(1-t) + sorted[hi]*t
}
