// Package recurring — 重复工作项应用服务。
//
// 提供规则的 CRUD / 启停 / 立即执行，以及调度器入口 EvaluateDue：
// 扫描所有 enabled 且 next_run_at <= now 的规则，按模板创建工作项并推进
// next_run_at 到下一个 cron 触发点。
package recurring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 重复工作项应用服务。
type Service struct {
	db      *pgxpool.Pool
	creator WorkItemCreator
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrRuleNotFound 规则不存在。
	ErrRuleNotFound = errors.New("recurring rule not found")
	// ErrInvalidCron cron 表达式非法。
	ErrInvalidCron = errors.New("invalid cron expression")
	// ErrInvalidWorkItemType 工作项类型非法。
	ErrInvalidWorkItemType = errors.New("invalid work item type")
)

// WorkItemCreator 创建工作项的最小依赖（由 issue.Service 实现）。
//
// 通过接口注入而非直接持有 *issue.Service，既避免潜在的包耦合，
// 也让调度器在单测中可用 fake 实现替换。
type WorkItemCreator interface {
	CreateWorkItem(ctx context.Context, in WorkItemTemplate) error
}

// WorkItemTemplate 生成工作项所需的字段投影。
//
// 该结构是 recurring 域的入参契约；装配层（cmd/api）负责把它翻译为
// issue.CreateIssueInput。使用导出类型保证跨包可实现。
type WorkItemTemplate struct {
	WorkspaceID     int64
	ProjectID       int64
	TypeCode        string
	Name            string
	DescriptionHTML string
	StateID         int64
	Priority        string
	Assignees       []int64
	Labels          []int64
	CreatedBy       int64
}

// WithWorkItemCreator 注入工作项创建器（链式可选）。
func (s *Service) WithWorkItemCreator(c WorkItemCreator) *Service {
	s.creator = c
	return s
}

// ruleColumns 与 recurring_rules 表列一一对应（List/Get 共用）。
const ruleColumns = `id, workspace_id, project_id, work_item_type, title_template,
	COALESCE(description_template,''), cron_expr, next_run_at, last_run_at,
	assignee_ids, label_ids, state_id, enabled, created_by, created_at, updated_at`

// CreateRule 创建重复规则。
//
// 校验：
//   - work_item_type 必须是 task/requirement/defect；
//   - title_template 不能为空；
//   - cron_expr 必须是合法的 5 段式表达式；
//   - next_run_at 由 cron 表达式基于当前时间计算（首次触发至少 1 分钟后）。
func (s *Service) CreateRule(ctx context.Context, in CreateRuleInput) (*RecurringRule, error) {
	if !isValidWorkItemType(in.WorkItemType) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "必须是 task、requirement 或 defect"})
	}
	if strings.TrimSpace(in.TitleTemplate) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title_template", Reason: "标题模板不能为空"})
	}
	if len([]rune(in.TitleTemplate)) > 500 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title_template", Reason: "标题模板不能超过 500 字符"})
	}
	schedule, err := parseCron(in.CronExpr)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "cron_expr", Reason: "无效的 cron 表达式（需为 5 段式：分 时 日 月 周）"})
	}

	enabled := in.Enabled
	nextRun := schedule.Next(time.Now().UTC()).Truncate(time.Second)

	var r RecurringRule
	err = scanRule(s.db.QueryRow(ctx, `
		INSERT INTO recurring_rules
			(id, workspace_id, project_id, work_item_type, title_template, description_template,
			 cron_expr, next_run_at, assignee_ids, label_ids, state_id, enabled, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING `+ruleColumns,
		genID(), in.WorkspaceID, in.ProjectID, in.WorkItemType, in.TitleTemplate,
		nullIfEmpty(in.DescriptionTemplate), in.CronExpr, nextRun,
		int64SliceOrEmpty(in.AssigneeIDs), int64SliceOrEmpty(in.LabelIDs),
		nullIfZero(in.StateID), enabled, in.CreatedBy,
	), &r)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("recurring.CreateRule: %w", err))
	}
	return &r, nil
}

// GetRule 查询单条规则。
func (s *Service) GetRule(ctx context.Context, ruleID int64) (*RecurringRule, error) {
	var r RecurringRule
	if err := scanRule(s.db.QueryRow(ctx, `
		SELECT `+ruleColumns+` FROM recurring_rules WHERE id = $1`, ruleID), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ListRules 列出项目下的规则。
func (s *Service) ListRules(ctx context.Context, opts ListRulesOptions) ([]RecurringRule, error) {
	where := []string{"workspace_id = $1", "project_id = $2"}
	args := []any{opts.WorkspaceID, opts.ProjectID}
	idx := 3
	if opts.Enabled != nil {
		where = append(where, fmt.Sprintf("enabled = $%d", idx))
		args = append(args, *opts.Enabled)
		idx++
	}

	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)

	query := `SELECT ` + ruleColumns + ` FROM recurring_rules
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("recurring.ListRules: %w", err))
	}
	defer rows.Close()

	out := make([]RecurringRule, 0)
	for rows.Next() {
		var r RecurringRule
		if err := scanRule(rows, &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateRule 更新规则（部分更新）。
//
// 若更新 cron_expr，则重新计算 next_run_at（避免旧时间表继续生效）。
func (s *Service) UpdateRule(ctx context.Context, ruleID int64, in UpdateRuleInput) (*RecurringRule, error) {
	existing, err := s.GetRule(ctx, ruleID)
	if err != nil {
		return nil, err
	}

	// 动态拼装 SET 子句。
	setCols := make([]string, 0, 7)
	args := make([]any, 0, 8)
	idx := 1

	if in.TitleTemplate != nil {
		if strings.TrimSpace(*in.TitleTemplate) == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title_template", Reason: "标题模板不能为空"})
		}
		setCols = append(setCols, fmt.Sprintf("title_template = $%d", idx))
		args = append(args, *in.TitleTemplate)
		idx++
	}
	if in.DescriptionTemplate != nil {
		setCols = append(setCols, fmt.Sprintf("description_template = $%d", idx))
		args = append(args, nullIfEmpty(*in.DescriptionTemplate))
		idx++
	}
	if in.CronExpr != nil {
		if _, err := parseCron(*in.CronExpr); err != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "cron_expr", Reason: "无效的 cron 表达式"})
		}
		setCols = append(setCols, fmt.Sprintf("cron_expr = $%d", idx))
		args = append(args, *in.CronExpr)
		idx++
	}
	if in.AssigneeIDs != nil {
		setCols = append(setCols, fmt.Sprintf("assignee_ids = $%d", idx))
		args = append(args, int64SliceOrEmpty(in.AssigneeIDs))
		idx++
	}
	if in.LabelIDs != nil {
		setCols = append(setCols, fmt.Sprintf("label_ids = $%d", idx))
		args = append(args, int64SliceOrEmpty(in.LabelIDs))
		idx++
	}
	if in.StateID != nil {
		setCols = append(setCols, fmt.Sprintf("state_id = $%d", idx))
		args = append(args, nullIfZero(*in.StateID))
		idx++
	}
	if in.Enabled != nil {
		setCols = append(setCols, fmt.Sprintf("enabled = $%d", idx))
		args = append(args, *in.Enabled)
		idx++
	}

	// cron 变更或规则被重新启用时，重算 next_run_at。
	needReschedule := false
	if in.CronExpr != nil {
		existing.CronExpr = *in.CronExpr
		needReschedule = true
	}
	if in.Enabled != nil && *in.Enabled && !existing.Enabled {
		needReschedule = true
	}
	if needReschedule {
		if sched, err := parseCron(existing.CronExpr); err == nil {
			next := sched.Next(time.Now().UTC()).Truncate(time.Second)
			setCols = append(setCols, fmt.Sprintf("next_run_at = $%d", idx))
			args = append(args, next)
			idx++
		}
	}

	if len(setCols) == 0 {
		return existing, nil
	}

	setCols = append(setCols, "updated_at = now()")
	args = append(args, ruleID)
	query := fmt.Sprintf(`UPDATE recurring_rules SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setCols, ", "), idx, ruleColumns)

	var r RecurringRule
	if err := scanRule(s.db.QueryRow(ctx, query, args...), &r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRuleNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("recurring.UpdateRule: %w", err))
	}
	return &r, nil
}

// DeleteRule 删除规则。
func (s *Service) DeleteRule(ctx context.Context, ruleID int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM recurring_rules WHERE id = $1`, ruleID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("recurring.DeleteRule: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}
	return nil
}

// EnableRule 启用规则（并重算 next_run_at）。
func (s *Service) EnableRule(ctx context.Context, ruleID int64) (*RecurringRule, error) {
	rule, err := s.GetRule(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	sched, err := parseCron(rule.CronExpr)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "cron_expr", Reason: "规则绑定的 cron 表达式已失效，无法启用"})
	}
	next := sched.Next(time.Now().UTC()).Truncate(time.Second)

	var r RecurringRule
	if err := scanRule(s.db.QueryRow(ctx, `
		UPDATE recurring_rules SET enabled = TRUE, next_run_at = $2, updated_at = now()
		WHERE id = $1 RETURNING `+ruleColumns, ruleID, next), &r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRuleNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("recurring.EnableRule: %w", err))
	}
	return &r, nil
}

// DisableRule 禁用规则。
func (s *Service) DisableRule(ctx context.Context, ruleID int64) (*RecurringRule, error) {
	var r RecurringRule
	if err := scanRule(s.db.QueryRow(ctx, `
		UPDATE recurring_rules SET enabled = FALSE, updated_at = now()
		WHERE id = $1 RETURNING `+ruleColumns, ruleID), &r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRuleNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("recurring.DisableRule: %w", err))
	}
	return &r, nil
}

// RunNow 立即执行一次规则（生成一个工作项），不改变调度状态。
//
// 用于用户手动「立即生成一次」场景，避免等待 cron 窗口。
func (s *Service) RunNow(ctx context.Context, ruleID int64) error {
	rule, err := s.GetRule(ctx, ruleID)
	if err != nil {
		return err
	}
	return s.generateWorkItem(ctx, rule)
}

// EvaluateDue 扫描所有到期规则并逐条生成工作项，随后推进 next_run_at。
//
// 返回本次处理（命中）的规则数。单个规则失败不阻塞其它规则。
// 调用方（worker / API 定时任务）应保证同一时刻只有一个实例在执行。
func (s *Service) EvaluateDue(ctx context.Context) (int, error) {
	if s.creator == nil {
		return 0, errs.ErrInternal.Wrap(errors.New("recurring: work item creator not configured"))
	}
	now := time.Now().UTC()
	rows, err := s.db.Query(ctx, `
		SELECT `+ruleColumns+`
		FROM recurring_rules
		WHERE enabled = TRUE AND next_run_at <= $1
		ORDER BY next_run_at ASC
		LIMIT 500`, now)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(fmt.Errorf("recurring.EvaluateDue: %w", err))
	}

	dueRules := make([]RecurringRule, 0)
	for rows.Next() {
		var r RecurringRule
		if err := scanRule(rows, &r); err != nil {
			rows.Close()
			return 0, err
		}
		dueRules = append(dueRules, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, errs.ErrInternal.Wrap(fmt.Errorf("recurring.EvaluateDue rows: %w", err))
	}

	processed := 0
	for i := range dueRules {
		rule := dueRules[i]
		if err := s.generateWorkItem(ctx, &rule); err != nil {
			// 单条失败不影响其余规则；记录后继续。
			continue
		}
		// 推进 next_run_at：从 max(now, next_run_at) 起算下一次，避免补跑历史导致死循环。
		base := rule.NextRunAt
		if base.Before(now) {
			base = now
		}
		sched, perr := parseCron(rule.CronExpr)
		if perr != nil {
			continue
		}
		next := sched.Next(base).Truncate(time.Second)
		if _, uerr := s.db.Exec(ctx, `
			UPDATE recurring_rules
			SET last_run_at = $2, next_run_at = $3, updated_at = now()
			WHERE id = $1`, rule.ID, now, next); uerr != nil {
			continue
		}
		processed++
	}
	return processed, nil
}

// generateWorkItem 按规则模板创建一个工作项。
func (s *Service) generateWorkItem(ctx context.Context, rule *RecurringRule) error {
	title := renderTemplate(rule.TitleTemplate, time.Now())
	desc := renderTemplate(rule.DescriptionTemplate, time.Now())

	return s.creator.CreateWorkItem(ctx, WorkItemTemplate{
		WorkspaceID:     rule.WorkspaceID,
		ProjectID:       rule.ProjectID,
		TypeCode:        rule.WorkItemType,
		Name:            title,
		DescriptionHTML: desc,
		StateID:         rule.StateID,
		Priority:        "none",
		Assignees:       rule.AssigneeIDs,
		Labels:          rule.LabelIDs,
		CreatedBy:       rule.CreatedBy,
	})
}

// --- helpers ---

func isValidWorkItemType(t string) bool {
	switch t {
	case "task", "requirement", "defect":
		return true
	default:
		return false
	}
}

// parseCron 解析 5 段式 cron 表达式（分 时 日 月 周）。
func parseCron(expr string) (cron.Schedule, error) {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	return p.Parse(strings.TrimSpace(expr))
}

// renderTemplate 渲染标题/描述模板中的占位符。
//
// 目前支持 {{date}} → YYYY-MM-DD，后续可扩展 {{year}}/{{month}}/{{week}} 等。
func renderTemplate(tmpl string, t time.Time) string {
	if tmpl == "" {
		return ""
	}
	return strings.ReplaceAll(tmpl, "{{date}}", t.Format("2006-01-02"))
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullIfZero(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

func int64SliceOrEmpty(in []int64) []int64 {
	if in == nil {
		return []int64{}
	}
	return in
}

// scanRule 从 pgx 行读取一条 RecurringRule。
func scanRule(row pgx.Row, r *RecurringRule) error {
	var lastRun sql.NullTime
	err := row.Scan(
		&r.ID, &r.WorkspaceID, &r.ProjectID, &r.WorkItemType, &r.TitleTemplate,
		&r.DescriptionTemplate, &r.CronExpr, &r.NextRunAt, &lastRun,
		&r.AssigneeIDs, &r.LabelIDs, &r.StateID, &r.Enabled, &r.CreatedBy,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRuleNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if lastRun.Valid {
		t := lastRun.Time
		r.LastRunAt = &t
	}
	return nil
}
