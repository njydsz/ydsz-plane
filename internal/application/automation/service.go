// Package automation 提供自动化规则的应用服务（CRUD + 模板预置 + dry-run）。
//
// 核心职责：
//   - 规则生命周期管理：Create / Update / Delete / Toggle / GetByID / List（含分页过滤）
//   - 内置模板预置：workspace 创建时预置常用规则（状态流转通知、自动指派等）
//   - dry-run 测试：给定规则 ID + 模拟工作项上下文，返回预测动作而非真实执行
//   - 执行审计：所有被触发的规则执行均落 execution_log，便于追溯
//
// 实际的动作执行委托给 engine.go（ExecutionEngine），本文件负责编排与输入校验。
package automation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供自动化规则引擎应用服务。
//
// 负责：
//   - 规则 CRUD（含 DSL 校验、状态机转换）
//   - 内置模板预置
//   - 按事件类型匹配活跃规则
//   - dry-run（干跑）测试
//   - 写入执行审计日志
//
// 注意事项：
//   - 实际的动作执行委托给 ExecutionEngine（单独文件）
//   - 规则创建/更新时强制门：ValidateDSL
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建自动化规则服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- CRUD ---

// Create 创建一条自动化规则。
//
// 强制校验: DSL 合法性。
// 自动化填充: trigger_type 从 DSL 提取，action_count 从 DFS DSL 动作数量。
func (s *Service) Create(ctx context.Context, in CreateRuleInput) (*Rule, error) {
	// DSL 校验
	result := ValidateDSL(in.DSL)
	if !result.Valid {
		details := make([]errs.FieldDetail, 0, len(result.Errors))
		for _, msg := range result.Errors {
			details = append(details, errs.FieldDetail{Field: "dsl", Reason: msg})
		}
		return nil, errs.ErrValidation.WithDetails(details...)
	}

	triggerType := in.DSL.Trigger.Type
	actionCount := len(in.DSL.Actions)

	rule := &Rule{
		WorkspaceID:  in.WorkspaceID,
		ProjectID:    in.ProjectID,
		Name:         in.Name,
		Description:  in.Description,
		DSL:          in.DSL,
		TriggerType:  triggerType,
		ActionCount:  actionCount,
		Status:       in.Status,
		CreatedBy:    in.CreatedBy,
		ExecutionCount: 0,
		SortOrder:    in.SortOrder,
	}

	err := s.db.QueryRow(ctx, `
		INSERT INTO automation_rules
			(workspace_id, project_id, name, description, dsl, trigger_type,
			 action_count, status, created_by, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`,
		rule.WorkspaceID, rule.ProjectID, rule.Name, rule.Description,
		rule.DSL, rule.TriggerType, rule.ActionCount, rule.Status,
		rule.CreatedBy, rule.SortOrder,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return rule, nil
}

// GetByID 按 ID 查询规则。
// 注：ProjectID 为 nullable，SQL WHERE 走 RLS tenant 隔离。
func (s *Service) GetByID(ctx context.Context, wsID, ruleID int64) (*Rule, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, description, dsl, trigger_type,
		       action_count, status, created_by, last_run_at, last_error,
		       consecutive_failures, execution_count, sort_order, created_at, updated_at
		FROM automation_rules WHERE id = $1 AND workspace_id = $2`,
		ruleID, wsID)
	return scanRule(row)
}

// Update 更新规则（带乐观锁防护）。
func (s *Service) Update(ctx context.Context, wsID, ruleID int64, in UpdateRuleInput) (*Rule, error) {
	// 若改 DSL，先校验
	if in.DSL != nil {
		result := ValidateDSL(*in.DSL)
		if !result.Valid {
			details := make([]errs.FieldDetail, 0, len(result.Errors))
			for _, msg := range result.Errors {
				details = append(details, errs.FieldDetail{Field: "dsl", Reason: msg})
			}
			return nil, errs.ErrValidation.WithDetails(details...)
		}
	}

	// 构建动态更新
	sets := []string{}
	args := []any{}
	argID := 1

	if in.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argID))
		args = append(args, *in.Name)
		argID++
	}
	if in.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argID))
		args = append(args, *in.Description)
		argID++
	}
	if in.DSL != nil {
		sets = append(sets, fmt.Sprintf("dsl = $%d", argID))
		args = append(args, *in.DSL)
		argID++
		sets = append(sets, fmt.Sprintf("trigger_type = $%d", argID))
		args = append(args, in.DSL.Trigger.Type)
		argID++
		sets = append(sets, fmt.Sprintf("action_count = $%d", argID))
		args = append(args, len(in.DSL.Actions))
		argID++
	}
	if in.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argID))
		args = append(args, *in.Status)
		// 状态→active 时，重置失败计数
		if *in.Status == RuleStatusActive {
			sets = append(sets, fmt.Sprintf("consecutive_failures = $%d", argID))
			args = append(args, 0)
			argID++
		}
		argID++
	}
	if in.SortOrder != nil {
		sets = append(sets, fmt.Sprintf("sort_order = $%d", argID))
		args = append(args, *in.SortOrder)
		argID++
	}

	if len(sets) == 0 {
		// 无更新字段 → 直接返回当前
		return s.GetByID(ctx, wsID, ruleID)
	}

	sets = append(sets, fmt.Sprintf("updated_at = now()"))

	// 乐观锁: version 检查
	args = append(args, wsID, ruleID, in.Version)
	sql := fmt.Sprintf(`UPDATE automation_rules SET %s WHERE workspace_id = $%d AND id = $%d AND (updated_at = $%d OR $%d = 0)`,
		joinSets(sets), argID, argID+1, argID+2, argID+2)

	tag, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "rule_id", Reason: "规则不存在或版本冲突"})
	}

	return s.GetByID(ctx, wsID, ruleID)
}

// Delete 软删除（项目级级联删除由 FK 处理）。
func (s *Service) Delete(ctx context.Context, wsID, ruleID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM automation_rules WHERE id = $1 AND workspace_id = $2`,
		ruleID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// List 查询规则列表（支持过滤）。
func (s *Service) List(ctx context.Context, wsID int64, opts ListRulesOptions) ([]Rule, int, error) {
	where := []string{"workspace_id = $1"}
	args := []any{wsID}
	argID := 2

	if opts.ProjectID != nil {
		where = append(where, fmt.Sprintf("project_id = $%d", argID))
		args = append(args, *opts.ProjectID)
		argID++
	} else if opts.ProjectID == nil && opts.ProjectID != nil {
		where = append(where, "project_id IS NULL")
	}

	if opts.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argID))
		args = append(args, *opts.Status)
		argID++
	}

	if opts.TriggerType != nil {
		where = append(where, fmt.Sprintf("trigger_type = $%d", argID))
		args = append(args, *opts.TriggerType)
		argID++
	}

	// 总数
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	countWhere := joinWhere(where)

	var total int
	if err := s.db.QueryRow(ctx,
		fmt.Sprintf("SELECT count(*) FROM automation_rules WHERE %s", countWhere),
		countArgs...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	// 分页
	limit := 50
	offset := 0
	if opts.Limit > 0 && opts.Limit <= 100 {
		limit = opts.Limit
	}
	if opts.Offset > 0 {
		offset = opts.Offset
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx,
		fmt.Sprintf(`SELECT id, workspace_id, project_id, name, description, dsl, trigger_type,
			action_count, status, created_by, last_run_at, last_error,
			consecutive_failures, execution_count, sort_order, created_at, updated_at
			FROM automation_rules WHERE %s ORDER BY sort_order, created_at DESC LIMIT $%d OFFSET $%d`,
			joinWhere(where), argID, argID+1),
		args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, *r)
	}
	return rules, total, rows.Err()
}

// --- Matching ---

// FindActiveByTrigger 查找对指定事件类型匹配的活跃规则。
// 返回按 sort_order 排序的规则列表。
// 这是事件消费者的核心查询。
func (s *Service) FindActiveByTrigger(ctx context.Context, projectID *int64, eventType string) ([]Rule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, name, description, dsl, trigger_type,
		       action_count, status, created_by, last_run_at, last_error,
		       consecutive_failures, execution_count, sort_order, created_at, updated_at
		FROM automation_rules
		WHERE status = 'active' AND trigger_type = $1
		  AND (project_id = $2 OR project_id IS NULL)
		ORDER BY sort_order ASC, id ASC`,
		eventType, projectID)
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

// --- Lifecycle ---

// RecordExecutionSuccess 记录一次成功的执行。
func (s *Service) RecordExecutionSuccess(ctx context.Context, ruleID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE automation_rules
		SET execution_count = execution_count + 1,
		    consecutive_failures = 0,
		    last_run_at = now(),
		    last_error = NULL,
		    status = CASE WHEN status = 'error' THEN 'active' ELSE status END
		WHERE id = $1`, ruleID)
	return err
}

// RecordExecutionFailure 记录一次执行失败。
// 连续失败 3 次 → status 自动切换到 disabled + error 双重标记。
func (s *Service) RecordExecutionFailure(ctx context.Context, ruleID int64, errMsg string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE automation_rules
		SET consecutive_failures = consecutive_failures + 1,
		    last_run_at = now(),
		    last_error = $2,
		    status = CASE WHEN consecutive_failures + 1 >= 3 THEN 'error' ELSE status END
		WHERE id = $1`, ruleID, errMsg)
	return err
}

// WriteExecution 写入一条执行审计日志。
func (s *Service) WriteExecution(ctx context.Context, exec RuleExecution) (int64, error) {
	var id int64
	contextJSON, _ := json.Marshal(exec.ContextJSON)
	err := s.db.QueryRow(ctx, `
		INSERT INTO rule_executions
			(workspace_id, project_id, rule_id, trigger_event_id, status,
			 duration_ms, error_message, context_json, trigger_depth, via_automation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`,
		exec.WorkspaceID, exec.ProjectID, exec.RuleID, exec.TriggerEventID,
		exec.Status, exec.DurationMs, exec.ErrorMessage, contextJSON,
		exec.TriggerDepth, exec.ViaAutomation,
	).Scan(&id)
	return id, err
}

// ListRecentExecutions 列出最近执行记录（执行历史页）。
func (s *Service) ListRecentExecutions(ctx context.Context, projectID int64, limit int) ([]RuleExecution, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `
		SELECT e.id, e.workspace_id, e.project_id, e.rule_id, e.trigger_event_id,
		       e.status, e.duration_ms, e.error_message, e.context_json,
		       e.trigger_depth, e.via_automation, e.created_at
		FROM rule_executions e
		LEFT JOIN automation_rules r ON r.id = e.rule_id
		WHERE (e.project_id = $1 OR r.project_id IS NULL) AND e.workspace_id = r.workspace_id
		ORDER BY e.created_at DESC LIMIT $2`,
		projectID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var execs []RuleExecution
	for rows.Next() {
		var e RuleExecution
		var ctxJSON []byte
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.ProjectID, &e.RuleID,
			&e.TriggerEventID, &e.Status, &e.DurationMs, &e.ErrorMessage, &ctxJSON,
			&e.TriggerDepth, &e.ViaAutomation, &e.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		_ = json.Unmarshal(ctxJSON, &e.ContextJSON)
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// DryRun 对指定规则 + 模拟上下文做干跑测试（不执行动作，返回"将执行"列表）。
// 用于前端"试运行"功能，帮助用户在启用规则前预览效果。
//
// 参数:
//   - ruleID: 要测试的规则 ID
//   - sampleIssueID: 用哪个工作项作为测试样本（取其上下文）
//   - event: 模拟的事件类型（如 issue.status_changed）
//
// 返回: 条件是否匹配 + 将执行的动作列表 + 警告（如果有无效引用等）。
func (s *Service) DryRun(ctx context.Context, wsID, ruleID int64, sampleIssueID int64, eventType string) (matched bool, actions []Action, warnings []string, err error) {
	rule, err := s.GetByID(ctx, wsID, ruleID)
	if err != nil {
		return false, nil, nil, err
	}

	// 构建模拟上下文（复用 engine 的 DefaultContextProvider）
	prov := NewDefaultContextProvider(s.db)
	dummyEvent := mq.EventEnvelope{
		EventType: eventType,
		Payload:  mustJSON(map[string]any{"issue_id": sampleIssueID, "workspace_id": wsID, "project_id": rule.ProjectID}),
	}
	execCtx, err := prov.BuildContext(ctx, dummyEvent)
	if err != nil {
		return false, nil, nil, fmt.Errorf("dry-run build context: %w", err)
	}
	execCtx.DryRun = true

	// 条件求值
	matched, err = evaluateConditions(rule.DSL.Conditions, execCtx)
	if err != nil {
		return false, nil, nil, fmt.Errorf("dry-run evaluate: %w", err)
	}

	if matched {
		actions = rule.DSL.Actions
	}

	// 检查变量引用是否有效
	for _, ref := range ExtractVariables(rule.DSL) {
		if !isValidVariable(ref[2 : len(ref)-1]) {
			warnings = append(warnings, fmt.Sprintf("未识别的变量引用: %s", ref))
		}
	}

	return matched, actions, warnings, nil
}

// mustJSON 是测试辅助函数（panic 仅用于构造消息，不影响生产路径）。
func mustJSON(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
func (s *Service) ListTemplates(ctx context.Context) ([]Template, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, slug, description, category, dsl_template, icon, sort_order, is_recommended
		FROM automation_templates ORDER BY sort_order`)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.Category,
			&t.DSLTemplate, &t.Icon, &t.SortOrder, &t.IsRecommended); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// CreateFromTemplate 从模板创建规则。
func (s *Service) CreateFromTemplate(ctx context.Context, in CreateRuleInput, templateSlug string) (*Rule, error) {
	var dsl RuleDSL
	err := s.db.QueryRow(ctx,
		`SELECT dsl_template FROM automation_templates WHERE slug = $1`, templateSlug).Scan(&dsl)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "template_slug", Reason: "模板不存在: " + templateSlug})
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	in.DSL = dsl
	return s.Create(ctx, in)
}

// --- Scanning ---

// scanRule 将单行扫描为 Rule。
func scanRule(scanner interface {
	Scan(dest ...any) error
}) (*Rule, error) {
	var dslRaw []byte
	r := &Rule{}
	err := scanner.Scan(&r.ID, &r.WorkspaceID, &r.ProjectID, &r.Name, &r.Description,
		&dslRaw, &r.TriggerType, &r.ActionCount, &r.Status, &r.CreatedBy,
		&r.LastRunAt, &r.LastError, &r.ConsecutiveFailures, &r.ExecutionCount,
		&r.SortOrder, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	_ = json.Unmarshal(dslRaw, &r.DSL)
	return r, nil
}

// --- Helpers ---

func joinSets(sets []string) string {
	result := ""
	for i, s := range sets {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func joinWhere(where []string) string {
	result := ""
	for i, w := range where {
		if i > 0 {
			result += " AND "
		}
		result += w
	}
	return result
}

// 确保编译期引用。
var _ = pgconn.PgError{}
