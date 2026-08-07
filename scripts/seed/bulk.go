package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
)

/* ================================================================== */
/* 模式 2：批量工作项造数                                                */
/* ================================================================== */

// runBulk 负责性能压测造数的全流程：
//  1. 解析 DB 连接串
//  2. 创建连接池
//  3. 保证基础表数据存在（至少 1 个 workspace / 1 个 project / 若干 state）
//  4. 批量 INSERT 工作项
//  5. 打印进度与最终统计
func runBulk(count, batchSize int, dsn string) error {
	if count <= 0 {
		return fmt.Errorf("count 必须 > 0")
	}
	if batchSize <= 0 {
		batchSize = 1000
	}

	// 如果未通过 flag 指定 DSN，回退读取环境变量或默认值。
	if dsn == "" {
		cfg, err := loadDSN()
		if err != nil {
			return err
		}
		dsn = cfg
	}

	ctx := context.Background()

	fmt.Printf("[bulk] 连接数据库: %s\n", maskDSN(dsn))

	pool, err := persistence.NewPool(ctx, dsn, 50)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer pool.Close()

	fmt.Printf("[bulk] 开始造数: count=%d, batch-size=%d\n", count, batchSize)

	// 准备外键（workspace / project / states / member）
	fk, err := prepareForeignKeyEntities(ctx, pool)
	if err != nil {
		return fmt.Errorf("准备外键实体失败: %w", err)
	}

	// 批量插入工作项
	start := time.Now()
	inserted, err := bulkInsertIssues(ctx, pool, fk, count, batchSize)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf("  造数完成: %d / %d\n", inserted, count)
	fmt.Printf("  耗时: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  速率: %.0f 行/秒\n", float64(inserted)/elapsed.Seconds())
	fmt.Printf("  workspace_id=%d project_id=%d\n", fk.workspaceID, fk.projectID)
	fmt.Println("═══════════════════════════════════════════════")
	return nil
}

// ----- 外键实体准备 -----

// fkEntities 描述造数工作项所需的外键引用。
type fkEntities struct {
	workspaceID int64
	projectID   int64
	projectSlug string
	states      []int64   // 项目已有状态的 ID 列表
	memberIDs   []int64   // 工作空间成员用户 ID
	now         time.Time
}

// prepareForeignKeyEntities 确保：
//   - 至少 1 个工作空间（如不存在则创建 "perf-test-ws"）
//   - 至少 1 个项目（如不存在则创建 "perf-test-project"）
//   - 至少 4 个状态（如不存在则创建 4 个标准状态）
//   - 工作空间有至少 1 个成员
func prepareForeignKeyEntities(ctx context.Context, pool *persistence.Pool) (*fkEntities, error) {
	fk := &fkEntities{now: time.Now()}

	var err error
	fk.workspaceID, err = ensureWorkspace(ctx, pool, "perf-test-ws", "性能测试空间", 1)
	if err != nil {
		return nil, err
	}

	// 取工作空间前 5 个成员（作为 assignee / created_by 等）
	fk.memberIDs, err = fetchMemberIDs(ctx, pool, fk.workspaceID, 5)
	if err != nil {
		return nil, err
	}
	if len(fk.memberIDs) == 0 {
		// 兜底：工作空间 owner
		var ownerID int64
		err = pool.QueryRow(ctx, `SELECT owner_id FROM workspaces WHERE id = $1`, fk.workspaceID).Scan(&ownerID)
		if err != nil {
			return nil, fmt.Errorf("fetch workspace owner: %w", err)
		}
		fk.memberIDs = []int64{ownerID}
	}

	fk.projectID, fk.projectSlug, err = ensureProject(ctx, pool, fk.workspaceID, fk.memberIDs[0])
	if err != nil {
		return nil, err
	}

	// 保证至少 4 个状态
	fk.states, err = ensureStates(ctx, pool, fk.workspaceID, fk.projectID)
	if err != nil {
		return nil, err
	}

	return fk, nil
}

func ensureWorkspace(ctx context.Context, pool *persistence.Pool, slug, name string, ownerID int64) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) WHERE status <> 'archived'
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, name, slug, ownerID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err.Error() == "no rows in result set" {
		return id, pool.QueryRow(ctx,
			`SELECT id FROM workspaces WHERE slug = $1 AND status = 'active'`, slug).Scan(&id)
	}
	// 兜底：尝试用 ownerID=1 创建
	err = pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) WHERE status <> 'archived'
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, name, slug, 1).Scan(&id)
	if err == nil {
		return id, nil
	}
	return id, pool.QueryRow(ctx,
		`SELECT id FROM workspaces WHERE slug = $1 AND status = 'active'`, slug).Scan(&id)
}

func fetchMemberIDs(ctx context.Context, pool *persistence.Pool, wsID int64, limit int) ([]int64, error) {
	rows, err := pool.Query(ctx,
		`SELECT user_id FROM workspace_members WHERE workspace_id = $1 LIMIT $2`,
		wsID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		ids = append(ids, uid)
	}
	return ids, rows.Err()
}

func ensureProject(ctx context.Context, pool *persistence.Pool, wsID, ownerID int64) (int64, string, error) {
	slug := "perf-test-project"
	identifier := "PERF"
	name := "性能测试项目"

	var id int64
	err := pool.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, public_id, name, slug, identifier, status, created_by)
		VALUES ($1, gen_random_uuid(), $2, $3, $3, 'active', $4)
		ON CONFLICT (workspace_id, slug) WHERE deleted_at IS NULL
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, wsID, name, identifier, ownerID).Scan(&id)
	if err == nil {
		return id, identifier, nil
	}
	if err.Error() == "no rows in result set" {
		err = pool.QueryRow(ctx,
			`SELECT id FROM projects WHERE workspace_id = $1 AND slug = $2 AND deleted_at IS NULL`,
			wsID, slug).Scan(&id)
		return id, identifier, err
	}
	// 冲突在 identifier 上 → 用 (workspace_id, slug) 查找
	err = pool.QueryRow(ctx,
		`SELECT id FROM projects WHERE workspace_id = $1 AND slug = $2 AND deleted_at IS NULL`,
		wsID, slug).Scan(&id)
	return id, identifier, err
}

// ensureStates 确保项目至少有 4 个状态；不足时补齐。
func ensureStates(ctx context.Context, pool *persistence.Pool, wsID, projectID int64) ([]int64, error) {
	// 查已有状态
	rows, err := pool.Query(ctx,
		`SELECT id FROM states WHERE project_id = $1 AND deleted_at IS NULL ORDER BY sequence`,
		projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []int64
	for rows.Next() {
		var sid int64
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		states = append(states, sid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(states) >= 4 {
		return states, nil
	}

	// 补齐为 4 个标准状态
	type stateDef struct {
		name  string
		group string
	}
	defs := []stateDef{
		{"待处理", "backlog"},
		{"进行中", "started"},
		{"已完成", "completed"},
		{"已取消", "cancelled"},
	}
	for i, d := range defs {
		// 避免重复创建（根据 project_id + name）
		var sid int64
		checkErr := pool.QueryRow(ctx,
			`SELECT id FROM states WHERE project_id = $1 AND name = $2 AND deleted_at IS NULL`,
			projectID, d.name).Scan(&sid)
		if checkErr == nil {
			states = append(states, sid)
			continue
		}
		seq := float64((i + 1) * 65535)
		color := "#8DA2C2"
		err := pool.QueryRow(ctx, `
			INSERT INTO states (workspace_id, project_id, name, "group", color, sequence, is_default)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			wsID, projectID, d.name, d.group, color, seq, i == 0).Scan(&sid)
		if err != nil {
			return nil, fmt.Errorf("insert state %s: %w", d.name, err)
		}
		states = append(states, sid)
	}
	return states, nil
}

// ----- 批量插入 -----

// bulkInsertIssues 将 N 条工作项分批次插入。
//
// 性能策略：
//   - 按批次开启事务，在事务内用 pgx Batch 一次性发送多条 INSERT
//   - sequence_id 通过 SELECT ... FOR UPDATE 在 project_sequences 表中原子递增分配
//   - 每完成一批 COMMIT，避免长事务
//   - created_at / updated_at 在过去 90 天内随机分布
func bulkInsertIssues(ctx context.Context, pool *persistence.Pool, fk *fkEntities, total, batchSize int) (int, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	inserted := 0
	progressStep := total / 10
	if progressStep < 1000 {
		progressStep = 1000
	}

	for inserted < total {
		remaining := total - inserted
		if remaining > batchSize {
			remaining = batchSize
		}

		err := insertBatch(ctx, pool, fk, remaining, rng)
		if err != nil {
			return inserted, fmt.Errorf("在已插入 %d 条时失败: %w", inserted, err)
		}
		inserted += remaining

		// 进度打印
		if inserted%progressStep == 0 || inserted == total {
			pct := float64(inserted) * 100 / float64(total)
			fmt.Printf("  已插入 %d / %d (%.0f%%)\n", inserted, total, pct)
		}
	}
	return inserted, nil
}

// insertBatch 在单一批事务中插入 batchCount 条工作项。
func insertBatch(ctx context.Context, pool *persistence.Pool, fk *fkEntities, batchCount int, rng *rand.Rand) error {
	// 开启事务，设置租户上下文以绕过 RLS 策略
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 设置 RLS 租户（绕过）
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", fk.workspaceID); err != nil {
		return fmt.Errorf("set tenant: %w", err)
	}

	// 原子分配 sequence_id 区间
	startSeq, err := allocateSequenceRange(ctx, tx, fk.projectID, batchCount)
	if err != nil {
		return err
	}

	// 构造批量 INSERT
	type rowData struct {
		projectID     int64
		sequenceID    int64
		typeCode      string
		name          string
		description   string
		stateID       int64
		priority      string
		createdBy     int64
		createdAt     time.Time
		updatedAt     time.Time
		targetDate    *time.Time
		point         *int
		progress      int
	}

	rows := make([]rawIssueRow, batchCount)
	now := fk.now
	for i := 0; i < batchCount; i++ {
		rows[i] = generateIssueRow(rng, fk, startSeq+int64(i), now)
	}

	// 使用 pgx Batch 发送
	batch := &pgx.Batch{}
	for _, r := range rows {
		batch.Queue(`
			INSERT INTO issues (
				workspace_id, project_id, sequence_id, type_code, name,
				description_json, description_html, description_stripped,
				state_id, priority, created_by, progress,
				created_at, updated_at, target_date, point, version, depth, sort_order
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11, $12,
				$13, $14, $15, $16, 1, 1, $17
			)`,
			fk.workspaceID, r.projectID, r.sequenceID, r.typeCode, r.name,
			fmt.Sprintf(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"%s"}]}]}`, r.description),
			fmt.Sprintf("<p>%s</p>", r.description),
			r.description,
			r.stateID, r.priority, r.createdBy, r.progress,
			r.createdAt, r.updatedAt, r.targetDate, r.point,
			float64(rng.Intn(1000000)),
		)
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batchCount; i++ {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("insert row: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return fmt.Errorf("close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// allocateSequenceRange 从 project_sequences 原子递增分配 seqCount 个序号。
// 如果不存在记录，初始化为 1。
func allocateSequenceRange(ctx context.Context, tx pgx.Tx, projectID int64, seqCount int) (int64, error) {
	var startSeq int64
	err := tx.QueryRow(ctx, `
		INSERT INTO project_sequences (project_id, next_value)
		VALUES ($1, $2)
		ON CONFLICT (project_id) DO UPDATE SET next_value = project_sequences.next_value + $3
		RETURNING next_value - $4`,
		projectID, 1+int64(seqCount), int64(seqCount), int64(seqCount)).Scan(&startSeq)
	if err != nil {
		return 0, fmt.Errorf("allocate sequence: %w", err)
	}
	return startSeq, nil
}

// rawIssueRow 是单条工作项插入前的中间表示。
type rawIssueRow struct {
	projectID   int64
	sequenceID  int64
	typeCode    string
	name        string
	description string
	stateID     int64
	priority    string
	createdBy   int64
	createdAt   time.Time
	updatedAt   time.Time
	targetDate  *time.Time
	point       *int
	progress    int
}

// generateIssueRow 生成单条工作项记录的伪随机数据。
func generateIssueRow(rng *rand.Rand, fk *fkEntities, sequenceID int64, now time.Time) rawIssueRow {
	var row rawIssueRow
	row.projectID = fk.projectID
	row.sequenceID = sequenceID

	// type_code
	types := []string{"requirement", "task", "defect"}
	row.typeCode = types[rng.Intn(len(types))]

	// name：中文+英文混合
	row.name = generateIssueName(rng, row.typeCode)

	// description
	row.description = generateDescription(rng)

	// state_id：随机
	row.stateID = fk.states[rng.Intn(len(fk.states))]

	// priority
	priorities := []string{"urgent", "high", "medium", "low", "none"}
	row.priority = priorities[rng.Intn(len(priorities))]

	// created_by：工作空间成员
	row.createdBy = fk.memberIDs[rng.Intn(len(fk.memberIDs))]

	// created_at / updated_at：过去 90 天内随机
	daysAgo := rng.Intn(90)
	hoursOffset := rng.Intn(24)
	minsOffset := rng.Intn(60)
	created := now.AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursOffset) * time.Hour).Add(-time.Duration(minsOffset) * time.Minute)
	row.createdAt = created
	// updated_at 在 created_at 之后 0~24 小时
	row.updatedAt = created.Add(time.Duration(rng.Intn(24)) * time.Hour)

	// target_date：50% 概率有，范围在 created_at 后 1~60 天
	if rng.Float64() < 0.5 {
		td := created.AddDate(0, 0, 1+rng.Intn(60))
		row.targetDate = &td
	}

	// point：30% 概率有，范围 0~12
	if rng.Float64() < 0.3 {
		pt := rng.Intn(13)
		row.point = &pt
	}

	// progress：跟 state 相关（简化：随机 0~100）
	row.progress = rng.Intn(101)

	return row
}

// 用于生成标题的前缀/后缀
var (
	cnPrefixes = []string{"优化", "实现", "修复", "重构", "设计", "排查", "集成", "部署", "测试", "审查",
		"升级", "降级", "迁移", "审核", "发布", "配置", "清理", "合并", "拆分", "归档"}
	cnSubjects = []string{"用户登录模块", "支付接口", "数据同步机制", "缓存策略", "权限管理",
		"搜索功能", "消息推送", "文件上传", "报表导出", "工作流引擎",
		"API 网关", "监控面板", "定时任务", "邮件通知", "数据统计",
		"前端组件", "数据库索引", "日志收集", "CI/CD 流水线", "性能调优"}
	cnSuffixes = []string{"的功能", "的异常", "的覆盖率", "的性能", "的兼容性",
		"的安全性", "的可维护性", "的测试用例", "的文档", "的依赖"}
	enPrefixes = []string{"Add", "Fix", "Refactor", "Implement", "Optimize", "Design", "Debug", "Integrate", "Deploy", "Review",
		"Upgrade", "Migrate", "Configure", "Clean up", "Merge", "Split", "Archive", "Investigate", "Test", "Monitor"}
	enSubjects = []string{"user login module", "payment gateway", "data sync", "cache strategy", "RBAC system",
		"search feature", "push notification", "file upload", "report export", "workflow engine",
		"API gateway", "monitoring dashboard", "cron scheduler", "email service", "data analytics",
		"frontend components", "DB indexing", "log collection", "CI/CD pipeline", "performance tuning"}
	enSuffixes  = []string{"", " (P0)", " (Sprint #)", " v2.0", " - hotfix", " - Phase 1", " - Review", " - v3", " (experimental)", ""}
)

func generateIssueName(rng *rand.Rand, typeCode string) string {
	useCN := rng.Float64() < 0.5 // 50% 中文
	var sb strings.Builder

	if useCN {
		sb.WriteString(cnPrefixes[rng.Intn(len(cnPrefixes))])
		sb.WriteString(cnSubjects[rng.Intn(len(cnSubjects))])
		sb.WriteString(cnSuffixes[rng.Intn(len(cnSuffixes))])
		if rng.Float64() < 0.2 {
			sb.WriteString(" (")
			sb.WriteString(enSubjects[rng.Intn(len(enSubjects))])
			sb.WriteString(")")
		}
	} else {
		sb.WriteString(enPrefixes[rng.Intn(len(enPrefixes))])
		sb.WriteString(" ")
		sb.WriteString(enSubjects[rng.Intn(len(enSubjects))])
		sb.WriteString(enSuffixes[rng.Intn(len(enSuffixes))])
		if rng.Float64() < 0.2 {
			sb.WriteString(" (")
			sb.WriteString(cnSubjects[rng.Intn(len(cnSubjects))])
			sb.WriteString(")")
		}
	}

	return sb.String()
}

func generateDescription(rng *rand.Rand) string {
	templates := []string{
		"本工作项关联到 Sprint %d，需要在 %d 个工作日内完成。核心风险点：第三方依赖兼容性。",
		"来自 %s 反馈的问题，影响约 %d 个用户。需要联合前后端评估修复方案。",
		"技术债：当前模块存在 %d 处 TODO 标记，圈复杂度 %d，需要重构以提升可测试性。",
		"数据迁移步骤：1. 备份快照 2. 运行迁移脚本 3. 验证数据一致性 4. 切换读流量。",
		"安全审计发现该端点存在潜在的 %s 风险，OWASP 分类 %s，请在发布前修复。",
		"性能目标：P95 响应时间 < 200ms，查询计划需覆盖 %d 个索引扫描路径。",
		"需求变更：产品侧调整了交互流程，详见 Figma 设计稿链接（v%d），后端需适配新字段。",
	}
	t := templates[rng.Intn(len(templates))]
	switch rng.Intn(7) {
	case 0:
		return fmt.Sprintf(t, rng.Intn(20)+1, rng.Intn(30)+1)
	case 1:
		return fmt.Sprintf(t, []string{"客户支持", "QA 团队", "产品团队", "线上监控"}[rng.Intn(4)], rng.Intn(500)+1)
	case 2:
		return fmt.Sprintf(t, rng.Intn(15)+1, rng.Intn(20)+5)
	case 3:
		return t
	case 4:
		return fmt.Sprintf(t, []string{"XSS", "SQL 注入", "越权访问", "CSRF"}[rng.Intn(4)],
			[]string{"A03:2021", "A01:2021", "A05:2021", "A07:2021"}[rng.Intn(4)])
	case 5:
		return fmt.Sprintf(t, rng.Intn(8)+1)
	case 6:
		return fmt.Sprintf(t, rng.Intn(5)+1)
	}
	return fmt.Sprintf("自动生成的工作项 #%d", rng.Intn(1000000))
}
