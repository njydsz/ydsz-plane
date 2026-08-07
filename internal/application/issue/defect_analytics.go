// Package issue — 缺陷分析报表聚合查询服务。
//
// 为 v0.2 S12 缺陷分析页提供后端数据接口，按模块/严重程度/发现阶段/根因分类聚合，
// 并产出缺陷龄分布与趋势数据（类 Jira Issue Statistics / Linear Insights）。
package issue

import (
	"context"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// DefectAnalytics 缺陷分析聚合结果（对应前端 ECharts 数据集）。
type DefectAnalytics struct {
	TotalDefects    int64            `json:"total_defects"`
	OpenDefects     int64            `json:"open_defects"`
	ResolvedDefects int64            `json:"resolved_defects"`
	AvgAgeDays      float64          `json:"avg_age_days"`
	SeverityDist    []SeverityCount  `json:"severity_dist"`
	PhaseDist       []PhaseCount     `json:"phase_dist"`
	ModuleDist      []ModuleCount    `json:"module_dist"`
	RootCauseDist   []RootCauseCount `json:"root_cause_dist"`
	AgeBuckets      []AgeBucket      `json:"age_buckets"`
	Trend           []TrendPoint     `json:"trend"`
}

// SeverityCount 按严重程度聚合。
type SeverityCount struct {
	Severity int    `json:"severity"`
	Label    string `json:"label"`
	Count    int64  `json:"count"`
}

// PhaseCount 按发现阶段聚合。
type PhaseCount struct {
	Phase string `json:"phase"`
	Count int64  `json:"count"`
}

// ModuleCount 按模块聚合（含模块名）。
type ModuleCount struct {
	ModuleID   *int64  `json:"module_id,omitempty"`
	ModuleName *string `json:"module_name,omitempty"`
	Count      int64   `json:"count"`
}

// RootCauseCount 按根因分类聚合。
type RootCauseCount struct {
	RootCause *string `json:"root_cause,omitempty"`
	Count     int64   `json:"count"`
}

// AgeBucket 缺陷龄分布桶。
type AgeBucket struct {
	Range     string `json:"range"`      // e.g. "0-3d"
	MinDays   int    `json:"min_days"`
	MaxDays   int    `json:"max_days"`   // -1 表示无穷
	Count     int64  `json:"count"`
}

// TrendPoint 缺陷创建/修复趋势点（按周聚合）。
type TrendPoint struct {
	Week      string `json:"week"`       // "YYYY-Wxx"
	Created   int64  `json:"created"`
	Resolved  int64  `json:"resolved"`
}

// AnalyticsQuery 缺陷分析过滤参数。
type AnalyticsQuery struct {
	WorkspaceID    int64
	ProjectID      int64
	DateFrom       *time.Time
	DateTo         *time.Time
	SeverityFrom   *int
	SeverityTo     *int
	ModuleID       *int64
	VersionID      *int64 // 版本过滤（found_version_id OR fix_version_id）
	IncludeDeleted bool
}

// DefectExportRow 缺陷导出明细行。
type DefectExportRow struct {
	Identifier  string
	Name        string
	Severity    *int
	StateName   string
	FoundPhase  *string
	RootCause   *string
	ModuleNames string
	CreatedAt   time.Time
	CompletedAt *time.Time
	AgeDays     float64
}

// DefectAnalyticsService 缺陷聚合分析服务。
type DefectAnalyticsService struct {
	db *pgxpool.Pool
}

// NewDefectAnalyticsService 创建分析服务。
func NewDefectAnalyticsService(db *pgxpool.Pool) *DefectAnalyticsService {
	return &DefectAnalyticsService{db: db}
}

// severityLabels 严重程度中文标签映射（1-5 级）。
var severityLabels = map[int]string{
	1: "致命",
	2: "严重",
	3: "一般",
	4: "轻微",
	5: "建议",
}

// GetAnalytics 聚合缺陷全量统计数据。
func (s *DefectAnalyticsService) GetAnalytics(ctx context.Context, q AnalyticsQuery) (*DefectAnalytics, error) {
	baseWhere, args := s.buildBaseWhere(q)
	analytics := &DefectAnalytics{}

	// 并行执行各维度聚合
	var err error
	if err = s.aggregateBase(ctx, baseWhere, args, analytics); err != nil {
		return nil, err
	}
	if analytics.SeverityDist, err = s.severityDist(ctx, baseWhere, args); err != nil {
		return nil, err
	}
	if analytics.PhaseDist, err = s.phaseDist(ctx, baseWhere, args); err != nil {
		return nil, err
	}
	if analytics.ModuleDist, err = s.moduleDist(ctx, baseWhere, args); err != nil {
		return nil, err
	}
	if analytics.RootCauseDist, err = s.rootCauseDist(ctx, baseWhere, args); err != nil {
		return nil, err
	}
	if analytics.AgeBuckets, err = s.ageBuckets(ctx, baseWhere, args); err != nil {
		return nil, err
	}
	if analytics.Trend, err = s.trend(ctx, baseWhere, args); err != nil {
		return nil, err
	}

	return analytics, nil
}

// buildBaseWhere 构建基础 where 条件（限定缺陷类型 + 项目 + 过滤参数）。
// 返回 SQL 片段与绑定参数，供各维度聚合与明细查询复用。
func (s *DefectAnalyticsService) buildBaseWhere(q AnalyticsQuery) (string, []interface{}) {
	// 基础 where 条件（限定缺陷类型 + 项目）
	baseWhere := `WHERE i.type_code = 'defect' AND i.workspace_id = $1 AND i.project_id = $2`
	args := []interface{}{q.WorkspaceID, q.ProjectID}
	argIdx := 3

	if !q.IncludeDeleted {
		baseWhere += ` AND i.deleted_at IS NULL`
	}
	if q.DateFrom != nil {
		baseWhere += ` AND i.created_at >= $` + strconv.Itoa(argIdx)
		args = append(args, *q.DateFrom)
		argIdx++
	}
	if q.DateTo != nil {
		baseWhere += ` AND i.created_at <= $` + strconv.Itoa(argIdx)
		args = append(args, *q.DateTo)
		argIdx++
	}
	if q.SeverityFrom != nil {
		baseWhere += ` AND i.severity >= $` + strconv.Itoa(argIdx)
		args = append(args, *q.SeverityFrom)
		argIdx++
	}
	if q.SeverityTo != nil {
		baseWhere += ` AND i.severity <= $` + strconv.Itoa(argIdx)
		args = append(args, *q.SeverityTo)
		argIdx++
	}
	if q.ModuleID != nil {
		baseWhere += ` AND EXISTS(SELECT 1 FROM issue_modules im WHERE im.issue_id = i.id AND im.module_id = $` + strconv.Itoa(argIdx) + `)`
		args = append(args, *q.ModuleID)
		argIdx++
	}
	if q.VersionID != nil {
		baseWhere += ` AND (i.found_version_id = $` + strconv.Itoa(argIdx) + ` OR i.fix_version_id = $` + strconv.Itoa(argIdx) + `)`
		args = append(args, *q.VersionID)
		argIdx++
	}
	return baseWhere, args
}

// ExportDefects 查询缺陷明细（供 CSV/XLSX 导出）。
// limit 控制最大导出条数，避免一次性导出海量数据拖垮连接。
func (s *DefectAnalyticsService) ExportDefects(ctx context.Context, q AnalyticsQuery, limit int) ([]DefectExportRow, error) {
	if limit <= 0 || limit > 5000 {
		limit = 5000 // 导出上限保护
	}
	baseWhere, args := s.buildBaseWhere(q)
	args = append(args, limit)

	rows, err := s.db.Query(ctx, `
		SELECT
			i.identifier,
			i.name,
			i.severity,
			COALESCE(st.name, '') AS state_name,
			i.found_phase,
			i.root_cause_category,
			i.created_at,
			i.completed_at,
			COALESCE(EXTRACT(EPOCH FROM (COALESCE(i.completed_at, now()) - i.created_at)) / 86400, 0) AS age_days,
			COALESCE((
				SELECT string_agg(m.name, '、' ORDER BY m.name)
				FROM issue_modules im JOIN modules m ON m.id = im.module_id
				WHERE im.issue_id = i.id
			), '') AS module_names
		FROM issues i
		LEFT JOIN states st ON st.id = i.state_id
		`+baseWhere+`
		ORDER BY i.created_at DESC
		LIMIT $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []DefectExportRow
	for rows.Next() {
		var r DefectExportRow
		var sev *int
		var foundPhase, rootCause, moduleNames *string
		var completedAt *time.Time
		if err := rows.Scan(
			&r.Identifier, &r.Name, &sev, &r.StateName,
			&foundPhase, &rootCause, &r.CreatedAt, &completedAt, &r.AgeDays, &moduleNames,
		); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		r.Severity = sev
		r.FoundPhase = foundPhase
		r.RootCause = rootCause
		r.CompletedAt = completedAt
		if moduleNames != nil {
			r.ModuleNames = *moduleNames
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// aggregateBase 统计基础指标：总数/未解决/已解决/平均龄。
func (s *DefectAnalyticsService) aggregateBase(ctx context.Context, where string, args []interface{}, a *DefectAnalytics) error {
	row := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE i.completed_at IS NULL) AS open_cnt,
			COUNT(*) FILTER (WHERE i.completed_at IS NOT NULL) AS resolved_cnt,
			COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(i.completed_at, now()) - i.created_at)) / 86400), 0) AS avg_age_days
		FROM issues i `+where, args...)
	if err := row.Scan(&a.TotalDefects, &a.OpenDefects, &a.ResolvedDefects, &a.AvgAgeDays); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// severityDist 按严重程度（1-5）聚合缺陷数。
func (s *DefectAnalyticsService) severityDist(ctx context.Context, where string, args []interface{}) ([]SeverityCount, error) {
	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(i.severity, 0) AS severity, COUNT(*) AS cnt
		FROM issues i `+where+`
		GROUP BY i.severity ORDER BY i.severity`, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []SeverityCount
	for rows.Next() {
		var sev, cnt int64
		if err := rows.Scan(&sev, &cnt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		label := "未知"
		if l, ok := severityLabels[int(sev)]; ok {
			label = l
		}
		result = append(result, SeverityCount{Severity: int(sev), Label: label, Count: cnt})
	}
	return result, rows.Err()
}

// phaseDist 按发现阶段聚合（found_phase 枚举值）。
func (s *DefectAnalyticsService) phaseDist(ctx context.Context, where string, args []interface{}) ([]PhaseCount, error) {
	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(i.found_phase, 'unknown') AS phase, COUNT(*) AS cnt
		FROM issues i `+where+`
		GROUP BY i.found_phase ORDER BY cnt DESC`, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []PhaseCount
	for rows.Next() {
		var phase string
		var cnt int64
		if err := rows.Scan(&phase, &cnt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, PhaseCount{Phase: phase, Count: cnt})
	}
	return result, rows.Err()
}

// moduleDist 按模块聚合缺陷数（含未分配模块的缺陷）。
func (s *DefectAnalyticsService) moduleDist(ctx context.Context, where string, args []interface{}) ([]ModuleCount, error) {
	// UNION：有模块关联 + 无模块关联
	query := `
		SELECT m.id AS module_id, m.name AS module_name, COUNT(i.id) AS cnt
		FROM issues i
		JOIN issue_modules im ON im.issue_id = i.id
		JOIN modules m ON m.id = im.module_id `+where+`
		GROUP BY m.id, m.name
		UNION ALL
		SELECT NULL::bigint, '未分配' AS module_name, COUNT(i.id) AS cnt
		FROM issues i `+where+`
		AND NOT EXISTS(SELECT 1 FROM issue_modules im WHERE im.issue_id = i.id)
		ORDER BY cnt DESC`
	// 注意：args 需要重复使用（两次 where）
	repeatedArgs := append(append([]interface{}{}, args...), args...)

	rows, err := s.db.Query(ctx, query, repeatedArgs...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []ModuleCount
	for rows.Next() {
		var mid *int64
		var mname *string
		var cnt int64
		if err := rows.Scan(&mid, &mname, &cnt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, ModuleCount{ModuleID: mid, ModuleName: mname, Count: cnt})
	}
	return result, rows.Err()
}

// rootCauseDist 按根因分类聚合。
func (s *DefectAnalyticsService) rootCauseDist(ctx context.Context, where string, args []interface{}) ([]RootCauseCount, error) {
	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(i.root_cause_category, 'unfilled') AS rc, COUNT(*) AS cnt
		FROM issues i `+where+`
		GROUP BY i.root_cause_category ORDER BY cnt DESC`, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []RootCauseCount
	for rows.Next() {
		var rc string
		var cnt int64
		if err := rows.Scan(&rc, &cnt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, RootCauseCount{RootCause: &rc, Count: cnt})
	}
	return result, rows.Err()
}

// ageBuckets 缺陷龄分布桶（0-3d / 4-7d / 8-14d / 15-30d / 31-90d / 90d+）。
func (s *DefectAnalyticsService) ageBuckets(ctx context.Context, where string, args []interface{}) ([]AgeBucket, error) {
	buckets := []AgeBucket{
		{Range: "0-3天", MinDays: 0, MaxDays: 3},
		{Range: "4-7天", MinDays: 4, MaxDays: 7},
		{Range: "8-14天", MinDays: 8, MaxDays: 14},
		{Range: "15-30天", MinDays: 15, MaxDays: 30},
		{Range: "31-90天", MinDays: 31, MaxDays: 90},
		{Range: "90天+", MinDays: 91, MaxDays: -1},
	}
	// copy where 补充 age 过滤
	ageExpr := `EXTRACT(EPOCH FROM (COALESCE(i.completed_at, now()) - i.created_at)) / 86400`
	for i, b := range buckets {
		bucketWhere := where + ` AND ` + ageExpr + ` >= ` + strconv.Itoa(b.MinDays)
		if b.MaxDays >= 0 {
			bucketWhere += ` AND ` + ageExpr + ` <= ` + strconv.Itoa(b.MaxDays)
		}
		var cnt int64
		err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM issues i `+bucketWhere, args...).Scan(&cnt)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		buckets[i].Count = cnt
	}
	return buckets, nil
}

// trend 缺陷创建与修复周趋势（倒推 12 周）。
func (s *DefectAnalyticsService) trend(ctx context.Context, where string, args []interface{}) ([]TrendPoint, error) {
	// 生成 12 周桶（数据库端 created 趋势 + 单独 resolved 查询）
	rows, err := s.db.Query(ctx, `
		SELECT to_char(dateweek, 'YYYY-"W"IW') AS week, COALESCE(c.cnt, 0) AS created, COALESCE(r.cnt, 0) AS resolved
		FROM generate_series(
			date_trunc('week', now()) - interval '11 weeks',
			date_trunc('week', now()),
			interval '1 week'
		) AS dateweek
		LEFT JOIN (
			SELECT date_trunc('week', created_at) AS wk, COUNT(*) AS cnt
			FROM issues i `+where+`
			GROUP BY wk
		) c ON c.wk = date_trunc('week', dateweek)
		LEFT JOIN (
			SELECT date_trunc('week', completed_at) AS wk, COUNT(*) AS cnt
			FROM issues i `+where+` AND i.completed_at IS NOT NULL
			GROUP BY wk
		) r ON r.wk = date_trunc('week', dateweek)
		ORDER BY dateweek`, args...)
	if err != nil {
		// 如果 generate_series 时间范围语法不支持，降级为仅 created 单线
		return s.trendFallback(ctx, where, args)
	}
	defer rows.Close()

	var result []TrendPoint
	for rows.Next() {
		var week string
		var created, resolved int64
		if err := rows.Scan(&week, &created, &resolved); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, TrendPoint{Week: week, Created: created, Resolved: resolved})
	}
	if len(result) == 0 {
		return s.trendFallback(ctx, where, args)
	}
	return result, rows.Err()
}

// trendFallback 降级趋势（仅 created，避免较旧 PG 版本 generate_series 兼容问题）。
func (s *DefectAnalyticsService) trendFallback(ctx context.Context, where string, args []interface{}) ([]TrendPoint, error) {
	rows, err := s.db.Query(ctx, `
		SELECT to_char(date_trunc('week', i.created_at), 'YYYY-"W"IW') AS week,
		       COUNT(*) AS created
		FROM issues i `+where+`
		GROUP BY week ORDER BY week DESC LIMIT 12`, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var result []TrendPoint
	for rows.Next() {
		var week string
		var created int64
		if err := rows.Scan(&week, &created); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		result = append(result, TrendPoint{Week: week, Created: created})
	}
	// 正序输出
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, rows.Err()
}
