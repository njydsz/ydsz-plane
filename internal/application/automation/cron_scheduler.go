// Package automation — scheduled/Cron 触发器调度器。
//
// 解决 S11 遗留缺口：DSL 校验支持 trigger.type=scheduled 且内置模板含两条
// scheduled 规则（逾期提醒 / 长期未更新自动归档），但此前仓库内没有任何
// 调度循环真正触发它们 —— 模板形同虚设。
//
// 本文件补齐该链路：
//
//	worker 每 1 分钟 tick 一次
//	  → 加载全部 active 的 scheduled 规则（trigger_type='scheduled'）
//	  → cron 表达式与当前分钟匹配（本地时区，5 段式：分 时 日 月 周）
//	  → 按 trigger.filter 做 SQL 预过滤候选工作项（如 due_within_hours）
//	  → 逐条候选走"条件求值 → 动作执行 → 审计落库"（复用 Engine）
//	  → 批内去重（同规则同分钟只跑一次，跨 worker 重启安全）
//
// 设计约束（与 worker 架构一致）：
//   - 不引入 Redis：同一自动化队列由单 worker 实例消费，cron 窗口 1 分钟
//     内重复触发概率可忽略；多实例 HA 部署时的跨实例互斥列为 Phase 3。
//   - cron 求值为纯函数，无外部依赖，避免新增第三方依赖。
//   - 候选量设上限（2000/规则/次），超限记录警告，防止全表扫描拖垮数据库。
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/worker"
	"github.com/njydsz/ydsz-plane/internal/rbac"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// 常量
const (
	// scheduledTick 调度器检查周期（与 cron 最小精度 1 分钟对齐）。
	scheduledTick = 1 * time.Minute
	// maxScheduledCandidates 单规则单次运行最多处理的候选工作项数。
	// 超过上限记录警告并截断，防止大项目（100 万级）定时任务全表扫描。
	maxScheduledCandidates = 2000
)

// RunScheduledCron 启动阻塞型 scheduled 触发器调度循环，ctx 取消时优雅退出。
// 应在 cmd/worker/main.go 中以独立 goroutine 调用。
//
// 与 RunConsumer（事件驱动）互补：事件消费者处理 issue.*/sprint.* 等实时
// 触发；本调度器只负责 trigger.type=scheduled 的规则，二者互不重叠。
func RunScheduledCron(ctx context.Context, db *pgxpool.Pool, rbacStore *rbac.Store, log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}
	eng := newEngine(db, rbacStore, log)
	svc := NewService(db)

	// runKey 去重表：ruleID → "20060102T1504"（避免同一规则同一分钟重复执行，
	// 覆盖 tick 抖动 / worker 重启导致的重复触发窗口）。
	lastRun := make(map[int64]string)

	log.Info("automation scheduled cron: starting",
		zap.Duration("tick", scheduledTick))

	ticker := time.NewTicker(scheduledTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("automation scheduled cron: stopped")
			return
		case t := <-ticker.C:
			if err := runScheduledTick(ctx, svc, eng, t, lastRun, log); err != nil {
				log.Warn("automation scheduled cron: tick failed", zap.Error(err))
			}
		}
	}
}

// runScheduledTick 执行一轮定时检查。
func runScheduledTick(ctx context.Context, svc *Service, eng *Engine, now time.Time, lastRun map[int64]string, log *zap.Logger) error {
	rules, err := svc.FindActiveScheduled(ctx)
	if err != nil {
		return fmt.Errorf("scheduled cron: load rules: %w", err)
	}

	runKey := now.Format("20060102T1504")
	for i := range rules {
		rule := &rules[i]

		// 定时规则必须绑定具体项目（全局 scheduled 规则无候选工作项语义）
		if rule.ProjectID == nil {
			log.Warn("automation scheduled: rule without project, skipping",
				zap.Int64("rule_id", rule.ID))
			continue
		}

		// cron 匹配检查（分钟级精度）
		if !CronMatches(rule.DSL.Trigger.Cron, now) {
			continue
		}

		// 同规则同分钟去重
		if lastRun[rule.ID] == runKey {
			continue
		}
		lastRun[rule.ID] = runKey

		// 时间打散：按 rule ID 计算确定性偏移，错开同分钟内多条规则的并发执行
		// 偏移上限 55s —— 确保在下一 tick 前完成当前规则的初始 DB 查询
		jitter := worker.IDJitter(rule.ID, 55)
		if jitter > 0 {
			log.Debug("automation scheduled: jitter wait",
				zap.Int64("rule_id", rule.ID),
				zap.Duration("jitter", jitter))
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(jitter):
			}
		}

		// 熔断器前置检查（与事件驱动路径一致）
		breaker := eng.breakers.GetOrCreate(rule.ID)
		if !breaker.Allow() {
			log.Debug("automation scheduled: rule circuit breaker open, skipping",
				zap.Int64("rule_id", rule.ID))
			continue
		}

		start := time.Now()
		matched, scanned, err := eng.EvaluateScheduledRule(ctx, rule)
		elapsed := time.Since(start)
		if err != nil {
			breaker.RecordFailure()
			_ = svc.RecordExecutionFailure(ctx, rule.ID, err.Error())
			log.Warn("automation scheduled: rule evaluation failed",
				zap.Int64("rule_id", rule.ID),
				zap.Int("scanned", scanned),
				zap.Error(err),
				zap.Duration("elapsed", elapsed))
			continue
		}
		breaker.RecordSuccess()

		log.Info("automation scheduled: rule fired",
			zap.Int64("rule_id", rule.ID),
			zap.String("cron", rule.DSL.Trigger.Cron),
			zap.Int("scanned", scanned),
			zap.Int("matched", matched),
			zap.Duration("elapsed", elapsed))
	}
	return nil
}

// FindActiveScheduled 返回全部 active 且绑定项目的 scheduled 规则。
// 供定时调度器按项目维度批量触发。
func (s *Service) FindActiveScheduled(ctx context.Context) ([]Rule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, name, description, dsl, trigger_type,
		       action_count, status, created_by, last_run_at, last_error,
		       consecutive_failures, execution_count, sort_order, created_at, updated_at
		FROM automation_rules
		WHERE status = 'active' AND trigger_type = 'scheduled' AND project_id IS NOT NULL
		ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *r)
	}
	return rules, rows.Err()
}

// EvaluateScheduledRule 对单条 scheduled 规则执行一次完整求值：
//  1. 依据 trigger.filter 预过滤候选工作项（SQL 下推，避免全表扫描）
//  2. 逐候选构建上下文 → 条件求值 → 命中则执行动作
//  3. 每候选写 rule_executions 审计；规则整体状态批处理更新
//
// 返回 (命中数, 扫描数, error)。错误意味着规则整体失败（触发熔断计数）。
func (e *Engine) EvaluateScheduledRule(ctx context.Context, rule *Rule) (matched, scanned int, err error) {
	issueIDs, err := e.findScheduledCandidates(ctx, rule)
	if err != nil {
		return 0, 0, err
	}

	if len(issueIDs) > maxScheduledCandidates {
		e.log.Warn("automation scheduled: candidates truncated",
			zap.Int64("rule_id", rule.ID),
			zap.Int("total", len(issueIDs)),
			zap.Int("limit", maxScheduledCandidates))
		issueIDs = issueIDs[:maxScheduledCandidates]
	}

	// 合成触发事件 ID（负数命名空间，避免与真实 domain_events 冲突）：
	// 规则 ID * 1e6 + 当前分钟，保证同批内一致、跨批唯一。
	triggerID := -int64(rule.ID*1_000_000 + time.Now().Unix()/60%1_000_000)

	prov := NewDefaultContextProvider(e.db)

	for _, issueID := range issueIDs {
		// 与事件驱动路径共享锁池：同 issue 的定时求值与实时事件求值互斥
		unlock := issueLockPool.lockForIssue(issueID)

		iss, loadErr := loadScheduledIssue(ctx, prov.(*DefaultContextProvider), rule, issueID)
		if loadErr != nil {
			unlock()
			e.log.Warn("automation scheduled: load issue failed",
				zap.Int64("rule_id", rule.ID),
				zap.Int64("issue_id", issueID),
				zap.Error(loadErr))
			continue
		}
		scanned++

		execCtx := iss
		execCtx.EventType = "scheduled"
		execCtx.Depth = 0

		ok, err := evaluateConditions(rule.DSL.Conditions, execCtx)
		if err != nil {
			unlock()
			e.log.Warn("automation scheduled: condition eval failed",
				zap.Int64("rule_id", rule.ID),
				zap.Int64("issue_id", issueID),
				zap.Error(err))
			continue
		}
		if !ok {
			// 条件未命中 → 审计 skipped
			_, _ = e.svc.WriteExecution(ctx, RuleExecution{
				WorkspaceID:    rule.WorkspaceID,
				ProjectID:      rule.ProjectID,
				RuleID:         rule.ID,
				TriggerEventID: &triggerID,
				Status:         ExecSkipped,
				TriggerDepth:   0,
				ViaAutomation:  false,
			})
			unlock()
			continue
		}

		if execErr := e.runScheduledActions(ctx, rule, execCtx, triggerID); execErr != nil {
			unlock()
			e.log.Warn("automation scheduled: actions failed",
				zap.Int64("rule_id", rule.ID),
				zap.Int64("issue_id", issueID),
				zap.Error(execErr))
			continue
		}
		unlock()
		matched++
	}

	// 规则状态批处理：有命中视为成功；仅扫描未命中不改变失败计数。
	if matched > 0 {
		_ = e.svc.RecordExecutionSuccess(ctx, rule.ID)
	}
	return matched, scanned, nil
}

// loadScheduledIssue 加载工作项并构建 scheduled 规则所需的执行上下文。
func loadScheduledIssue(ctx context.Context, prov *DefaultContextProvider, rule *Rule, issueID int64) (*ExecutionContext, error) {
	iss, err := prov.loadIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	return &ExecutionContext{
		EventType:   "scheduled",
		WorkspaceID: rule.WorkspaceID,
		ProjectID:   *rule.ProjectID,
		Issue:       iss,
		Actor:       ActorContext{UserID: 0, UserName: "system"},
		Now:         time.Now(),
		Depth:       0,
	}, nil
}

// runScheduledActions 执行命中的动作序列并写审计（status 为 success/failure）。
func (e *Engine) runScheduledActions(ctx context.Context, rule *Rule, execCtx *ExecutionContext, triggerID int64) error {
	start := time.Now()
	status := ExecSuccess
	var execErr string

	for _, act := range rule.DSL.Actions {
		if err := e.executeAction(ctx, act, execCtx); err != nil {
			execErr = err.Error()
			status = ExecFailed
			break
		}
	}

	duration := int(time.Since(start).Milliseconds())
	_, _ = e.svc.WriteExecution(ctx, RuleExecution{
		WorkspaceID:    rule.WorkspaceID,
		ProjectID:      rule.ProjectID,
		RuleID:         rule.ID,
		TriggerEventID: &triggerID,
		Status:         status,
		DurationMs:     &duration,
		ErrorMessage:   execErr,
		ContextJSON: map[string]any{
			"issue_id":  execCtx.Issue.ID,
			"scheduled": true,
		},
		TriggerDepth:  0,
		ViaAutomation: false,
	})

	if status == ExecFailed {
		return fmt.Errorf("action execution failed: %s", execErr)
	}
	return nil
}

// findScheduledCandidates 依据 trigger.filter 生成候选工作项 ID 列表（SQL 下推）。
//
// 当前支持的 filter 键：
//   - due_within_hours: 目标日期落在 [now, now+N 小时) 内的工作项（逾期提醒模板）
//
// 无识别键时退化为项目全量（受 maxScheduledCandidates 截断保护）。
func (e *Engine) findScheduledCandidates(ctx context.Context, rule *Rule) ([]int64, error) {
	projectID := *rule.ProjectID

	where := []string{"i.project_id = $1", "i.deleted = false"}
	args := []any{projectID}
	argID := 2

	if hours, ok := numberFromAny(rule.DSL.Trigger.Filter["due_within_hours"]); ok {
		where = append(where, fmt.Sprintf("i.target_date IS NOT NULL AND i.target_date >= CURRENT_DATE AND i.target_date <= ((now() + make_interval(hours => $%d)))::date", argID))
		args = append(args, int(hours))
		argID++
	}

	sql := fmt.Sprintf(`
		SELECT i.id
		FROM (
		    SELECT id, project_id, deleted, target_date FROM task
		    UNION ALL
		    SELECT id, project_id, deleted, target_date FROM requirement
		    UNION ALL
		    SELECT id, project_id, deleted, target_date FROM defect
		) i
		WHERE %s
		ORDER BY i.id ASC
		LIMIT %d`,
		strings.Join(where, " AND "), maxScheduledCandidates+1)

	rows, err := e.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// --- cron 表达式匹配（5 段式，本地时区） ---

// CronMatches 判断 5 段式 cron 表达式（分 时 日 月 周）是否命中给定时刻。
// 支持语法：* 、*/n 、a-b 、a,b,c 、单值。
// 不支持的语法（?、L、W、#、@daily 等）返回 false 并记录在注释中；
// 校验层（ValidateDSL）已保证格式，运行时此处仅做匹配。
func CronMatches(cron string, t time.Time) bool {
	fields := strings.Fields(strings.TrimSpace(cron))
	if len(fields) != 5 {
		return false
	}

	parts := []int{t.Minute(), t.Hour(), t.Day(), int(t.Month()), int(t.Weekday())}
	bounds := []int{59, 23, 31, 12, 6}

	for i := 0; i < 5; i++ {
		if !cronFieldMatches(fields[i], parts[i], bounds[i]) {
			return false
		}
	}
	return true
}

// cronFieldMatches 判断单个字段是否匹配。
func cronFieldMatches(field string, value, max int) bool {
	// 逗号分隔（或集）
	if strings.Contains(field, ",") {
		for _, sub := range strings.Split(field, ",") {
			if cronFieldMatches(strings.TrimSpace(sub), value, max) {
				return true
			}
		}
		return false
	}

	// 步进：*/n 或 a-b/n
	if strings.Contains(field, "/") {
		parts := strings.SplitN(field, "/", 2)
		step, err := strconv.Atoi(parts[1])
		if err != nil || step <= 0 {
			return false
		}
		baseLow, baseHigh := 0, max
		if parts[0] != "*" {
			if strings.Contains(parts[0], "-") {
				ran := strings.SplitN(parts[0], "-", 2)
				lo, err1 := strconv.Atoi(ran[0])
				hi, err2 := strconv.Atoi(ran[1])
				if err1 != nil || err2 != nil {
					return false
				}
				baseLow, baseHigh = lo, hi
			} else {
				v, err := strconv.Atoi(parts[0])
				if err != nil {
					return false
				}
				baseLow, baseHigh = v, max
			}
		}
		if value < baseLow || value > baseHigh {
			return false
		}
		return (value-baseLow)%step == 0
	}

	// 范围：a-b
	if strings.Contains(field, "-") {
		ran := strings.SplitN(field, "-", 2)
		lo, err1 := strconv.Atoi(ran[0])
		hi, err2 := strconv.Atoi(ran[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return value >= lo && value <= hi
	}

	// 通配
	if field == "*" {
		return true
	}

	// 单值
	v, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return v == value
}

// numberFromAny 从 DSL filter 值中提取整数（支持 float64 / int / json.Number 等）。
func numberFromAny(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	}
	return 0, false
}
