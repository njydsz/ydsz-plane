// Command perf — 性能回归测试：针对 100 万工作项基线，测量关键查询延迟。
//
// 参考互联网大厂标准（字节/美团 DBA 团队压测流程）：
//   1. 确认 1M 工作项已就位（先 go run ./scripts/seed --count 1000000）
//   2. 执行预定义 SQL 查询模板，采样 P50/P95/P99
//   3. 对比索引优化前后耗时
//
// 使用示例：
//   # 全部用例（默认 30 次采样）
//   go run ./scripts/perf --samples 30
//
//   # 仅测主列表查询（100 次采样）
//   go run ./scripts/perf --case list --samples 100
//
//   # 指定连接
//   go run ./scripts/perf --db-dsn "postgres://..."
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

var (
	flagSamples = flag.Int("samples", 30, "每查询采样次数")
	flagDSN     = flag.String("db-dsn", "", "数据库连接串")
	flagCase    = flag.String("case", "all", "指定用例: list|detail|filter_by_type|filter_by_priority|calendar_range|workbench_summary|dashboard_priority|overdue_count|search_stripped|activity_stream|all")
)

// perfCase 是单性能测试用例。
type perfCase struct {
	Name  string
	Query string
	Args  []interface{}
}

func main() {
	flag.Parse()

	dsn := *flagDSN
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "load config:", err)
			os.Exit(1)
		}
		dsn = cfg.Database.URL
	}

	ctx := context.Background()
	pool, err := persistence.NewPool(ctx, dsn, 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect db:", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 找出含最多工作项的项目
	projectID, wsID, err := findTargetProject(ctx, pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "find target project:", err)
		os.Exit(1)
	}
	fmt.Printf("[perf] target: workspace_id=%d project_id=%d samples=%d\n", wsID, projectID, *flagSamples)

	// 设置 RLS 上下文
	_, err = pool.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", wsID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "set tenant:", err)
		os.Exit(1)
	}

	cases := buildCases(projectID, wsID)
	for _, c := range cases {
		if *flagCase != "all" && *flagCase != c.Name {
			continue
		}
		runCase(c, pool, *flagSamples)
	}

	fmt.Println("\n[perf] 测试完成。对比索引优化前后建议执行:")
	fmt.Println("  EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) <query>;")
}

// findTargetProject 找到含最多工作项的项目（作为压测目标）。
func findTargetProject(ctx context.Context, pool *persistence.Pool) (projectID, wsID int64, err error) {
	err = pool.QueryRow(ctx, `
		SELECT p.id, p.workspace_id
		FROM projects p
		WHERE p.deleted_at IS NULL
		ORDER BY (SELECT count(*) FROM issues i WHERE i.project_id = p.id AND i.deleted_at IS NULL) DESC
		LIMIT 1`).Scan(&projectID, &wsID)
	return
}

// buildCases 构造全部性能测试 SQL。
func buildCases(projectID, wsID int64) []perfCase {
	return []perfCase{
		{
			Name: "list",
			Query: `SELECT i.id, i.public_id, i.workspace_id, i.project_id, i.sequence_id,
				i.type_code, i.parent_id, i.depth, i.name,
				i.state_id, s.name, s.color, s."group",
				i.priority, i.severity, i.category, i.point,
				i.start_date, i.target_date, i.progress, i.version,
				i.created_by, i.created_at, i.updated_at
			FROM issues i
			JOIN states s ON s.id = i.state_id
			WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2
			ORDER BY i.updated_at DESC LIMIT 50`,
			Args: []interface{}{projectID, wsID},
		},
		{
			Name: "detail",
			Query: `SELECT i.id, i.sequence_id, i.name, i.type_code, i.priority,
				s.name, s.color, s."group"
			FROM issues i
			JOIN states s ON s.id = i.state_id
			WHERE i.project_id = $1 AND i.workspace_id = $2
			ORDER BY i.updated_at DESC LIMIT 1`,
			Args: []interface{}{projectID, wsID},
		},
		{
			Name: "filter_by_type",
			Query: `SELECT i.id, i.sequence_id, i.name, i.state_id, i.priority, i.updated_at
			FROM issues i
			WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2 AND i.type_code = 'defect'
			ORDER BY i.created_at DESC LIMIT 50`,
			Args: []interface{}{projectID, wsID},
		},
		{
			Name: "filter_by_priority",
			Query: `SELECT i.id, i.sequence_id, i.name, i.state_id, i.type_code, i.updated_at
			FROM issues i
			WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2 AND i.priority IN ('urgent','high')
			ORDER BY i.updated_at DESC LIMIT 20`,
			Args: []interface{}{projectID, wsID},
		},
		{
			Name: "calendar_range",
			Query: `SELECT i.id, i.sequence_id, i.name, i.state_id, i.type_code, i.progress
			FROM issues i
			WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2
			  AND i.target_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
			ORDER BY i.target_date
			LIMIT 100`,
			Args: []interface{}{projectID, wsID},
		},
		{
			Name: "workbench_summary",
			Query: `SELECT st."group", count(*)
			FROM issues i
			JOIN states st ON st.id = i.state_id
			WHERE i.project_id = $1 AND i.deleted_at IS NULL
			GROUP BY st."group"`,
			Args: []interface{}{projectID},
		},
		{
			Name: "dashboard_priority",
			Query: `SELECT i.priority, count(*)
			FROM issues i
			WHERE i.project_id = $1 AND i.deleted_at IS NULL
			GROUP BY i.priority ORDER BY i.priority`,
			Args: []interface{}{projectID},
		},
		{
			Name: "overdue_count",
			Query: `SELECT count(*)
			FROM issues i
			JOIN states st ON st.id = i.state_id
			WHERE i.project_id = $1 AND i.deleted_at IS NULL
			  AND i.target_date < CURRENT_DATE AND st."group" NOT IN ('completed','cancelled')`,
			Args: []interface{}{projectID},
		},
		{
			Name: "search_stripped",
			Query: `SELECT i.id, i.sequence_id, i.name, i.priority, i.updated_at
			FROM issues i
			WHERE i.project_id = $1 AND i.deleted_at IS NULL
			  AND i.description_stripped LIKE '%性能%'
			ORDER BY i.updated_at DESC LIMIT 20`,
			Args: []interface{}{projectID},
		},
		{
			Name: "activity_stream",
			Query: `SELECT a.verb, a.field, a.old_value, a.new_value, a.actor_email, a.actor_name, a.created_at
			FROM issue_activities a
			WHERE a.issue_id IN (
				SELECT i.id FROM issues i WHERE i.project_id = $1 AND i.deleted_at IS NULL
				ORDER BY i.updated_at DESC LIMIT 20
			)
			ORDER BY a.created_at DESC
			LIMIT 100`,
			Args: []interface{}{projectID},
		},
	}
}

// runCase 对单用例采样 N 次并输出统计。
func runCase(c perfCase, pool *persistence.Pool, samples int) {
	ctx := context.Background()
	times := make([]time.Duration, 0, samples)

	for i := 0; i < samples; i++ {
		start := time.Now()
		rows, err := pool.Query(ctx, c.Query, c.Args...)
		if err != nil {
			fmt.Printf("  [%s] ERROR: %v\n", c.Name, err)
			return
		}
		for rows.Next() {
		}
		rows.Close()
		times = append(times, time.Since(start))
	}

	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var total time.Duration
	for _, t := range times {
		total += t
	}
	avg := total / time.Duration(samples)
	p50 := sorted[len(sorted)*50/100]
	p95 := sorted[len(sorted)*95/100]
	p99 := sorted[len(sorted)*99/100]

	fmt.Printf("  %-22s  avg=%-10s P50=%-10s P95=%-10s P99=%-10s\n",
		c.Name,
		avg.Round(time.Millisecond),
		p50.Round(time.Millisecond),
		p95.Round(time.Millisecond),
		p99.Round(time.Millisecond),
	)
}
