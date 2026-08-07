// Package workspace — 项目风险规则初始化。
//
// 在项目创建后注入 3 条默认风险检测规则，覆盖逾期、阻塞、高优先级未解决场景。
package workspace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultRiskRule 定义项目级别的单条默认风险检测规则。
type defaultRiskRule struct {
	ruleName       string
	ruleType       string
	conditionJSON  map[string]any
	notifyChannels []string
}

// defaultProjectRiskRules 项目默认风险规则清单。
//
// 包含三类：
//   - overdue_issue:    target_date 超过 1 天未完成的工作项
//   - blocked_count:    阻塞下游数量超过 3 条
//   - high_priority_open: urgent 优先级超过 7 天未解决的
var defaultProjectRiskRules = []defaultRiskRule{
	{
		ruleName:       "逾期工作项",
		ruleType:       "overdue_issue",
		conditionJSON:  map[string]any{"threshold_days": 1},
		notifyChannels: []string{"in_app"},
	},
	{
		ruleName:       "阻塞工作项累计超阈值",
		ruleType:       "blocked_count",
		conditionJSON:  map[string]any{"threshold_count": 3},
		notifyChannels: []string{"in_app"},
	},
	{
		ruleName:       "高优先级未解决",
		ruleType:       "high_priority_open",
		conditionJSON:  map[string]any{"priority": "urgent", "age_days": 7},
		notifyChannels: []string{"in_app"},
	},
}

// EnsureProjectDefaultRiskRules 为项目创建默认风险检测规则。
//
// 应在 ProjectService.Create 成功后调用。已存在同名规则时幂等跳过。
// 单条规则插入失败不影响其余规则（错误被静默吞掉，不影响主流程）。
func EnsureProjectDefaultRiskRules(ctx context.Context, db *pgxpool.Pool, wsID, projectID int64) {
	for _, r := range defaultProjectRiskRules {
		condJSON, _ := json.Marshal(r.conditionJSON)
		_, _ = db.Exec(ctx, `
			INSERT INTO risk_rules (workspace_id, project_id, rule_name, rule_type, condition_json, notify_channels, is_active)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6::text[], TRUE)
			ON CONFLICT DO NOTHING`,
			wsID, projectID, r.ruleName, r.ruleType, condJSON, pgTextArray(r.notifyChannels))
	}
}

// pgTextArray 将 Go 字符串切片转为 PostgreSQL text[] 字面量。
func pgTextArray(items []string) string {
	out := "{"
	for i, it := range items {
		if i > 0 {
			out += ","
		}
		out += fmt.Sprintf("%q", it)
	}
	out += "}"
	return out
}
