// Package automation 提供领域自动化规则引擎的运行时调度与执行能力。
//
// 核心职责：
//   - 事件驱动：订阅领域事件（issue.created/updated、sprint.started、version.released 等），匹配活跃规则
//   - 根据事件优先级与规则权重排序执行；支持 dry-run（干跑）模式产出执行预览
//   - 通过 SPI（ExecutionContextProvider / ActionExecutor）解耦领域副作用，便于单元测试 fake
//
// 与 automation_dsl.go（DSL 定义）和 service.go（应用服务）协作，形成"定义-存储-执行"闭环。
package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// ExecutionContextProvider 是构建运行时上下文的 SPI 接口。
//
// 此接口解耦自动化引擎与具体域服务（Issue、Sprint 等）的依赖，
// 使得引擎可在单元测试中用 fake 实现隔离运行。
type ExecutionContextProvider interface {
	// BuildContext 从领域事件聚合根构建运行时上下文。
	BuildContext(ctx context.Context, event mq.EventEnvelope) (*ExecutionContext, error)
}

// ActionExecutor SPI：执行动作所需的外部副作用服务。
type ActionExecutor interface {
	// TransitionIssueStatus 执行工作项状态流转。
	TransitionIssueStatus(ctx context.Context, wsID, projectID, issueID int64, targetState string) error
	// AssignIssue 指派工作项给用户。
	AssignIssue(ctx context.Context, wsID, projectID, issueID int64, userID int64) error
	// UpdateIssueField 更新工作项字段。
	UpdateIssueField(ctx context.Context, wsID, projectID, issueID int64, field string, value any) error
	// SendNotification 发送通知。
	SendNotification(ctx context.Context, n NotificationRequest) error
	// CreateIssue 创建工作项。
	CreateIssue(ctx context.Context, in CreateIssueRequest) (int64, error)
}

// NotificationRequest 通知动作的请求。
type NotificationRequest struct {
	WorkspaceID int64
	ProjectID   int64
	RecipientID int64
	Title       string
	Body        string
	ActorID     *int64
	ActorName   string
}

// CreateIssueRequest 创建工作项的请求。
type CreateIssueRequest struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Description string
	TypeCode    string
	CreatedBy   int64
}

// Engine 是自动化规则执行引擎。
//
// 并发模型:
//   - 每个 Engine 实例是 goroutine-safe（无状态共享）
//   - 事件到达顺序处理，同一工作项在 Redis 锁中串行化
//   - 防循环：via_automation=true 且 depth >5 → 丢弃
type Engine struct {
	svc       *Service
	execDeps  ActionExecutor
	ctxProv   ExecutionContextProvider
	log       *zap.Logger
	// db 供引擎内部 DB 查询使用（assignees / tech_lead / least_loaded）。
	// 由 newEngine 注入；NewEngine 构造时可为 nil（不执行需 DB 的动作）。
	db *pgxpool.Pool
}

// NewEngine 创建执行引擎。
func NewEngine(svc *Service, deps ActionExecutor, prov ExecutionContextProvider, log *zap.Logger) *Engine {
	if log == nil {
		log = zap.NewNop()
	}
	return &Engine{
		svc:      svc,
		execDeps: deps,
		ctxProv:  prov,
		log:      log,
	}
}

// EvaluateEvent 消费一条领域事件并触发匹配的自动化规则。
//
// 处理流程:
//  1. BuildContext: 解析事件 → 运行时上下文
//  2. FindActiveByTrigger: 查找匹配的活跃规则
//  3. 对每条规则:
//     a. 防循环检查（depth ≤5）
//     b. 条件求值（纯函数）
//     c. 匹配 → 执行动作
//     d. 写入审计日志
//     e. 更新规则的成功/失败状态
//
// 错误隔离：一条规则失败不影响其他规则。
func (e *Engine) EvaluateEvent(ctx context.Context, event mq.EventEnvelope) error {
	// 构建运行时上下文
	execCtx, err := e.ctxProv.BuildContext(ctx, event)
	if err != nil {
		return fmt.Errorf("automation: build context: %w", err)
	}

	// 防循环深度检查
	if execCtx.Depth > 5 {
		e.log.Warn("automation: execution depth exceeded, skipping",
			zap.String("event", event.EventType),
			zap.Int("depth", execCtx.Depth))
		return nil
	}

	// 解析 project context from event
	projectID := e.extractProjectID(event)

	// 查找活跃规则
	rules, err := e.svc.FindActiveByTrigger(ctx, projectID, event.EventType)
	if err != nil {
		return fmt.Errorf("automation: find rules: %w", err)
	}
	if len(rules) == 0 {
		return nil // 没有匹配规则，正常返回
	}

	e.log.Debug("automation: evaluating rules for event",
		zap.String("event", event.EventType),
		zap.Int("rule_count", len(rules)))

	// 逐条规则评估
	for _, rule := range rules {
		if err := e.evaluateRule(ctx, &rule, execCtx, event); err != nil {
			e.log.Warn("automation: rule evaluation failed",
				zap.Int64("rule", rule.ID),
				zap.Error(err))
			// 写失败状态 + 审计
			_ = e.svc.RecordExecutionFailure(ctx, rule.ID, err.Error())
			_, _ = e.svc.WriteExecution(ctx, RuleExecution{
				WorkspaceID:    rule.WorkspaceID,
				ProjectID:      rule.ProjectID,
				RuleID:         rule.ID,
				TriggerEventID: &event.EventID,
				Status:         ExecFailed,
				ErrorMessage:   err.Error(),
				TriggerDepth:   execCtx.Depth,
				ViaAutomation:  execCtx.ViaAutomation,
			})
		}
	}

	return nil
}

// evaluateRule: 对单条规则做"条件求值 → 动作执行"。
func (e *Engine) evaluateRule(ctx context.Context, rule *Rule, execCtx *ExecutionContext, event mq.EventEnvelope) error {
	start := time.Now()

	// 1. 条件求值（纯函数，无 IO）
	matched, err := evaluateConditions(rule.DSL.Conditions, execCtx)
	if err != nil {
		return err
	}

	if !matched {
		// 条件不满足 → 记录 skipped + 返回
		_, _ = e.svc.WriteExecution(ctx, RuleExecution{
			WorkspaceID:    rule.WorkspaceID,
			ProjectID:      rule.ProjectID,
			RuleID:         rule.ID,
			TriggerEventID: &event.EventID,
			Status:         ExecSkipped,
			TriggerDepth:   execCtx.Depth,
			ViaAutomation:  execCtx.ViaAutomation,
		})
		return nil
	}

	// 2. 执行动作
	dryRun := execCtx.DryRun
	status := ExecSuccess
	var execErr string
	var actionsTaken []Action

	if !dryRun {
		for _, act := range rule.DSL.Actions {
			if err := e.executeAction(ctx, act, execCtx); err != nil {
				execErr = err.Error()
				status = ExecFailed
				break
			}
			actionsTaken = append(actionsTaken, act)
		}
	} else {
		actionsTaken = rule.DSL.Actions
		status = ExecDryRun
	}

	duration := int(time.Since(start).Milliseconds())

	// 3. 写审计日志
	contextJSON := map[string]any{
		"conditions_matched": matched,
		"actions_taken":      len(actionsTaken),
		"dry_run":            dryRun,
	}
	_, _ = e.svc.WriteExecution(ctx, RuleExecution{
		WorkspaceID:    rule.WorkspaceID,
		ProjectID:      rule.ProjectID,
		RuleID:         rule.ID,
		TriggerEventID: &event.EventID,
		Status:         status,
		DurationMs:     &duration,
		ErrorMessage:   execErr,
		ContextJSON:    contextJSON,
		TriggerDepth:   execCtx.Depth,
		ViaAutomation:  execCtx.ViaAutomation,
	})

	// 4. 更新规则状态
	if status == ExecSuccess {
		_ = e.svc.RecordExecutionSuccess(ctx, rule.ID)
	} else if status == ExecFailed {
		_ = e.svc.RecordExecutionFailure(ctx, rule.ID, execErr)
	}

	if execErr != "" {
		return fmt.Errorf("action execution failed: %s", execErr)
	}
	return nil
}

// executeAction: 执行单个动作。
func (e *Engine) executeAction(ctx context.Context, act Action, execCtx *ExecutionContext) error {
	if execCtx.Issue == nil {
		return fmt.Errorf("action %s requires issue context", act.Type)
	}

	wsID := execCtx.WorkspaceID
	projectID := execCtx.ProjectID
	issueID := execCtx.Issue.ID

	switch act.Type {
	case ActionTransition:
		targetState, _ := act.Value.(string)
		if targetState == "" {
			return fmt.Errorf("transition: missing target state")
		}
		return e.execDeps.TransitionIssueStatus(ctx, wsID, projectID, issueID, targetState)

	case ActionAssign:
		userID, err := e.resolveAssignTarget(ctx, act, execCtx)
		if err != nil {
			return fmt.Errorf("assign: %w", err)
		}
		return e.execDeps.AssignIssue(ctx, wsID, projectID, issueID, userID)

	case ActionUpdateField:
		return e.execDeps.UpdateIssueField(ctx, wsID, projectID, issueID, act.Field, act.Value)

	case ActionCopyField:
		// 把 act.Field 对应的目标字段设为 act.Value（先 resolveTemplate 再转换）。
		resolved := resolveTemplate(toString(act.Value), execCtx)
		return e.execDeps.UpdateIssueField(ctx, wsID, projectID, issueID, act.Field, resolved)

	case ActionNotify:
		return e.executeNotifyAction(ctx, act, execCtx)

	case ActionCreateIssue:
		req, err := buildCreateIssueRequest(act, execCtx)
		if err != nil {
			return fmt.Errorf("create_issue: %w", err)
		}
		_, err = e.execDeps.CreateIssue(ctx, req)
		return err

	case ActionWebhookCall:
		// Webhook 调用走外部 API，异步到 TaskExchange
		return fmt.Errorf("webhook_call: not implemented in v0.2")

	default:
		return fmt.Errorf("unknown action type: %s", act.Type)
	}
}

// executeNotifyAction: 执行通知动作。
func (e *Engine) executeNotifyAction(ctx context.Context, act Action, execCtx *ExecutionContext) error {
	config := act.Config
	channel, _ := config["channel"].(string)
	template, _ := config["template"].(string)
	target, _ := config["target"].(string)

	// MVP: 仅支持 in_app channel
	if channel == "" {
		channel = "in_app"
	}

	// 解析目标用户
	var recipients []int64
	switch target {
	case "${issue.assignees}":
		// 从 issue context 获取
		// (实际实现需查数据库，这里简化)
		recipients = e.getIssueAssigneeIDs(ctx, execCtx.Issue.ID)
	case "${issue.created_by}":
		recipients = []int64{execCtx.Issue.CreatedBy}
	case "${project.tech_lead}":
		tlID, err := e.getProjectTechLead(ctx, execCtx.ProjectID)
		if err != nil {
			return err
		}
		recipients = []int64{tlID}
	default:
		// 尝试解析为用户 ID
		if uid, err := strconv.ParseInt(target, 10, 64); err == nil {
			recipients = []int64{uid}
		}
	}

	for _, uid := range recipients {
		if uid == execCtx.Actor.UserID {
			continue // 不通知操作者自己
		}
		if err := e.execDeps.SendNotification(ctx, NotificationRequest{
			WorkspaceID: execCtx.WorkspaceID,
			ProjectID:   execCtx.ProjectID,
			RecipientID: uid,
			Title:       resolveTemplate(template, execCtx),
			Body:        execCtx.Issue.Name,
			ActorID:     &execCtx.Actor.UserID,
			ActorName:   execCtx.Actor.UserName,
		}); err != nil {
			e.log.Warn("automation: send notification failed",
				zap.Int64("recipient", uid), zap.Error(err))
		}
	}
	return nil
}

// --- Condition Evaluation (纯函数) ---

// evaluateConditions: 评估条件列表（AND 语义）。
func evaluateConditions(conditions []Condition, ctx *ExecutionContext) (bool, error) {
	for _, cond := range conditions {
		matched, err := evaluateCondition(cond, ctx)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil // AND 语义：一条不即全部不
		}
	}
	return true, nil
}

// evaluateCondition: 评估单条条件。
func evaluateCondition(cond Condition, ctx *ExecutionContext) (bool, error) {
	sampleValue := resolveField(cond.Field, ctx)

	switch cond.Op {
	case "is_empty":
		return isEmpty(sampleValue), nil
	case "is_not_empty":
		return !isEmpty(sampleValue), nil
	case "changed":
		// поле 是否在当前事件中变更（对于实时事件恒为 true）
		return true, nil
	}

	// 其他 op 需要 cond.Value
	if cond.Value == nil {
		return false, fmt.Errorf("condition %s requires value for op=%s", cond.Field, cond.Op)
	}

	return compare(sampleValue, cond.Op, cond.Value)
}

// resolveField: 从运行时上下文解析字段值。
func resolveField(field string, ctx *ExecutionContext) any {
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return nil
	}

	section := parts[0]
	switch section {
	case "issue":
		if ctx.Issue == nil {
			return nil
		}
		return resolveIssueField(ctx.Issue, parts[1:])
	case "sprint":
		if ctx.Sprint == nil {
			return nil
		}
		return resolveSprintField(ctx.Sprint, parts[1:])
	case "version":
		if ctx.Version == nil {
			return nil
		}
		return resolveVersionField(ctx.Version, parts[1:])
	case "actor":
		return resolveActorField(&ctx.Actor, parts[1:])
	case "project":
		// project fields need DB query - use simple lookup
		return nil
	case "now":
		return ctx.Now
	default:
		return nil
	}
}

func resolveIssueField(issue *IssueContext, path []string) any {
	if len(path) == 0 {
		return issue
	}
	key := path[0]
	switch key {
	case "id":
		return issue.ID
	case "identifier":
		return issue.Identifier
	case "name":
		return issue.Name
	case "type_code", "type":
		return issue.TypeCode
	case "state_id":
		return issue.StateID
	case "state_name":
		return issue.StateName
	case "state_group", "state.group":
		return issue.StateGroup
	case "priority":
		return issue.Priority
	case "severity":
		if issue.Severity != nil {
			return *issue.Severity
		}
		return nil
	case "estimate_points":
		if issue.EstimatePoints != nil {
			return *issue.EstimatePoints
		}
		return nil
	case "created_by":
		return issue.CreatedBy
	case "parent", "parent_id":
		if issue.ParentID != nil {
			return *issue.ParentID
		}
		return nil
	case "project_id":
		return issue.ProjectID
	case "started_at":
		return issue.StartedAt
	case "completed_at":
		return issue.CompletedAt
	}
	return nil
}

func resolveSprintField(sprint *SprintContext, path []string) any {
	if len(path) == 0 {
		return sprint
	}
	switch path[0] {
	case "id":
		return sprint.ID
	case "name":
		return sprint.Name
	case "status":
		return sprint.Status
	case "project_id":
		return sprint.ProjectID
	}
	return nil
}

func resolveVersionField(version *VersionContext, path []string) any {
	if len(path) == 0 {
		return version
	}
	switch path[0] {
	case "id":
		return version.ID
	case "name":
		return version.Name
	case "status":
		return version.Status
	}
	return nil
}

func resolveActorField(actor *ActorContext, path []string) any {
	if len(path) == 0 {
		return actor
	}
	switch path[0] {
	case "user_id", "id":
		return actor.UserID
	case "user_name", "name":
		return actor.UserName
	}
	return nil
}

// compare: 比较实际值与期望值。
func compare(actual any, op string, expected any) (bool, error) {
	if actual == nil {
		// null 下的 eq/ne 语义
		switch op {
		case "eq":
			return expected == nil, nil
		case "ne":
			return expected != nil, nil
		}
		return false, nil
	}

	switch op {
	case "eq":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected), nil
	case "ne":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected), nil
	case "contains":
		actualStr := fmt.Sprintf("%v", actual)
		expectedStr := fmt.Sprintf("%v", expected)
		return strings.Contains(actualStr, expectedStr), nil
	case "in":
		// expected 是数组，actual 在其内即返回 true
		arr, ok := expected.([]any)
		if ok {
			for _, item := range arr {
				if fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", item) {
					return true, nil
				}
			}
		}
		return false, nil
	}

	// 数值比较 (gt/gte/lt/lte)
	actualNum, aOk := toFloat64(actual)
	expectedNum, eOk := toFloat64(expected)
	if !aOk || !eOk {
		return false, fmt.Errorf("comparison %s requires numeric values: actual=%v expected=%v", op, actual, expected)
	}

	switch op {
	case "gt":
		return actualNum > expectedNum, nil
	case "gte":
		return actualNum >= expectedNum, nil
	case "lt":
		return actualNum < expectedNum, nil
	case "lte":
		return actualNum <= expectedNum, nil
	}

	return false, fmt.Errorf("unknown op: %s", op)
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case *string:
		return val == nil || *val == ""
	case *int:
		return val == nil
	case *int64:
		return val == nil
	case *time.Time:
		return val == nil
	}
	return false
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}

// --- Template Variable Resolution ---

// resolveTemplate: 替换模板中的 ${变量}。
// 未解析的变量原样保留（不 panic）。
func resolveTemplate(tpl string, ctx *ExecutionContext) string {
	if ctx == nil || ctx.Issue == nil {
		return tpl
	}
	result := tpl
	result = strings.ReplaceAll(result, "${issue.identifier}", ctx.Issue.Identifier)
	result = strings.ReplaceAll(result, "${issue.name}", ctx.Issue.Name)
	result = strings.ReplaceAll(result, "${issue.assignees}", "")
	result = strings.ReplaceAll(result, "${project.tech_lead}", "")
	result = strings.ReplaceAll(result, "${project.id}", fmt.Sprintf("%d", ctx.ProjectID))
	result = strings.ReplaceAll(result, "${actor.name}", ctx.Actor.UserName)
	if ctx.Issue.EstimatePoints != nil {
		result = strings.ReplaceAll(result, "${parent.estimate_points}", fmt.Sprintf("%d", *ctx.Issue.EstimatePoints))
	} else {
		result = strings.ReplaceAll(result, "${parent.estimate_points}", "0")
	}
	if ctx.Sprint != nil {
		result = strings.ReplaceAll(result, "${sprint.name}", ctx.Sprint.Name)
	}
	if ctx.Version != nil {
		result = strings.ReplaceAll(result, "${version.name}", ctx.Version.Name)
		result = strings.ReplaceAll(result, "${version.id}", fmt.Sprintf("%d", ctx.Version.ID))
	}
	result = strings.ReplaceAll(result, "${now}", ctx.Now.Format("2006-01-02"))
	return result
}

// resolveAssignTarget: 解析 assign 动作的目标用户。
func (e *Engine) resolveAssignTarget(ctx context.Context, act Action, ctxExec *ExecutionContext) (int64, error) {
	if act.Value != nil {
		if uid, ok := toFloat64(act.Value); ok {
			return int64(uid), nil
		}
	}
	// config.strategy=least_loaded 查询最闲的成员
	if strategy, _ := act.Config["strategy"].(string); strategy == "least_loaded" {
		if e.db == nil {
			return 0, fmt.Errorf("least_loaded strategy requires DB query (db not configured)")
		}
		role, _ := act.Config["role"].(string)
		if role == "" {
			role = "member"
		}
		userID, err := e.resolveLeastLoaded(ctx, ctxExec.WorkspaceID, ctxExec.ProjectID, role)
		if err != nil {
			return 0, fmt.Errorf("least_loaded: %w", err)
		}
		return userID, nil
	}
	return 0, fmt.Errorf("assign: unable to resolve target user")
}

// resolveLeastLoaded 查询项目下 open issue 数量最少的 role 成员。
// 负载 = 该成员当前负责的未完成（state.group != completed）工作项数量。
func (e *Engine) resolveLeastLoaded(ctx context.Context, wsID, projectID int64, role string) (int64, error) {
	var userID int64
	err := e.db.QueryRow(ctx, `
		SELECT wm.user_id
		FROM workspace_members wm
		JOIN projects p ON p.id = $1 AND p.workspace_id = wm.workspace_id
		WHERE wm.workspace_id = $2 AND wm.role = $3
		ORDER BY (
			SELECT count(*)
			FROM issue_assignees ia
			JOIN issues i ON i.id = ia.issue_id
			JOIN states st ON st.id = i.state_id
			WHERE ia.user_id = wm.user_id AND i.project_id = $1 AND i.deleted_at IS NULL
			  AND st."group" != 'completed'
		) ASC, wm.user_id ASC
		LIMIT 1`, projectID, wsID, role).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func buildCreateIssueRequest(act Action, ctx *ExecutionContext) (CreateIssueRequest, error) {
	req := CreateIssueRequest{
		WorkspaceID: ctx.WorkspaceID,
		ProjectID:   ctx.ProjectID,
		CreatedBy:   ctx.Actor.UserID,
	}
	if name, _ := act.Config["name"].(string); name != "" {
		req.Name = name
	} else {
		req.Name = "Auto: " + ctx.Issue.Name
	}
	req.Description, _ = act.Config["description"].(string)
	req.TypeCode, _ = act.Config["type_code"].(string)
	if req.TypeCode == "" {
		req.TypeCode = "requirement"
	}
	return req, nil
}

// --- Helpers (DB queries, stubs) ---

func (e *Engine) extractProjectID(event mq.EventEnvelope) *int64 {
	var payload struct {
		ProjectID int64 `json:"project_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err == nil && payload.ProjectID > 0 {
		return &payload.ProjectID
	}
	return nil
}

func (e *Engine) getIssueAssigneeIDs(ctx context.Context, issueID int64) []int64 {
	if e.db == nil {
		return nil
	}
	rows, err := e.db.Query(ctx, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, issueID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err == nil {
			ids = append(ids, uid)
		}
	}
	return ids
}

// getProjectTechLead 返回项目技术负责人 ID。
// 注：projects 表无 tech_lead 列（SQL 0003 已确认），故以项目创建者 created_by 作为负责人代理。
func (e *Engine) getProjectTechLead(ctx context.Context, projectID int64) (int64, error) {
	if e.db == nil {
		return 0, fmt.Errorf("getProjectTechLead: db not configured")
	}
	var leadID int64
	err := e.db.QueryRow(ctx,
		`SELECT created_by FROM projects WHERE id = $1 AND deleted_at IS NULL`, projectID).Scan(&leadID)
	if err != nil {
		return 0, fmt.Errorf("getProjectTechLead: %w", err)
	}
	return leadID, nil
}

// --- Context Provider Implementation (占位，实际需查 DB) ---

// DefaultContextProvider 默认上下文提供器。
// 仅作骨架，实际生产环境中需查询 DB 获取完整上下文。
type DefaultContextProvider struct {
	db *pgxpool.Pool
}

// NewDefaultContextProvider 创建默认上下文提供器。
func NewDefaultContextProvider(db *pgxpool.Pool) ExecutionContextProvider {
	return &DefaultContextProvider{db: db}
}

// BuildContext 从领域事件聚合根构建运行时上下文。
func (p *DefaultContextProvider) BuildContext(ctx context.Context, event mq.EventEnvelope) (*ExecutionContext, error) {
	execCtx := &ExecutionContext{
		EventType: event.EventType,
		Now:       time.Now(),
		Depth:     0,
	}

	// 解析 event payload 获取 IDs
	var payload struct {
		WorkspaceID int64 `json:"workspace_id"`
		ProjectID   int64 `json:"project_id"`
		IssueID     int64 `json:"issue_id"`
		SprintID    int64 `json:"sprint_id"`
		VersionID   int64 `json:"version_id"`
		ActorID     int64 `json:"actor_id"`
		ActorName   string `json:"actor_name"`
	}
	_ = json.Unmarshal(event.Payload, &payload)

	execCtx.WorkspaceID = payload.WorkspaceID
	execCtx.Actor = ActorContext{UserID: payload.ActorID, UserName: payload.ActorName}

	if payload.ProjectID > 0 {
		execCtx.ProjectID = payload.ProjectID
	}

	// 根据事件类型加载 issue 上下文
	if payload.IssueID > 0 && strings.HasPrefix(event.EventType, "issue.") {
		iss, err := p.loadIssue(ctx, payload.IssueID)
		if err != nil && err != pgx.ErrNoRows {
			return nil, err
		}
		execCtx.Issue = iss
		if iss != nil {
			execCtx.ProjectID = iss.ProjectID
		}
	}

	return execCtx, nil
}

func (p *DefaultContextProvider) loadIssue(ctx context.Context, issueID int64) (*IssueContext, error) {
	var iss IssueContext
	var startedAt, completedAt *time.Time
	err := p.db.QueryRow(ctx, `
		SELECT i.id, i.identifier, i.name, i.type_code, i.state_id,
		       st.name, st."group", i.priority, i.severity, i.estimate_points,
		       i.created_by, i.parent_id, i.project_id, i.created_at, i.updated_at,
		       i.started_at, i.completed_at
		FROM issues i
		JOIN states st ON st.id = i.state_id
		WHERE i.id = $1 AND i.deleted_at IS NULL`, issueID).Scan(
		&iss.ID, &iss.Identifier, &iss.Name, &iss.TypeCode, &iss.StateID,
		&iss.StateName, &iss.StateGroup, &iss.Priority, &iss.Severity, &iss.EstimatePoints,
		&iss.CreatedBy, &iss.ParentID, &iss.ProjectID, &iss.CreatedAt, &iss.UpdatedAt,
		&startedAt, &completedAt)
	if err != nil {
		return nil, err
	}
	iss.StartedAt = startedAt
	iss.CompletedAt = completedAt
	return &iss, nil
}
