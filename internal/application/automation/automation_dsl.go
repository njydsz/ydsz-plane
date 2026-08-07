// Package automation 提供领域自动化规则引擎的 DSL 定义、解析与校验能力。
//
// 核心职责：
//   - 定义 Automation DSL（Trigger / Condition / Action）并实现 JSON Schema 校验
//   - 提供内置动作常量（transition/assign/update_field/notify/create_issue）与触发器类型枚举
//   - 支持条件表达式解析（eq/ne/gt/gte/lt/lte/contains/in/is_empty/is_not_empty/changed）
//
// DSL 数据流：前端 DSL JSON → BuildRule → ValidateDSL → DB 存储 → 运行时引擎匹配执行。
package automation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// 支持的 trigger.type 值。
var validTriggerTypes = map[string]bool{
	"issue.created":         true,
	"issue.updated":         true,
	"issue.status_changed":  true,
	"issue.commented":       true,
	"issue.deleted":         true,
	"sprint.created":        true,
	"sprint.started":        true,
	"sprint.completed":      true,
	"version.released":      true,
	"version.created":       true,
	"version.updated":       true,
	"member.added":          true,
	"scheduled":             true,
}

// 支持的 condition op。
var validConditionOps = map[string]bool{
	"eq":           true,
	"ne":           true,
	"gt":           true,
	"gte":          true,
	"lt":           true,
	"lte":          true,
	"contains":     true,
	"in":           true,
	"is_empty":     true,
	"is_not_empty": true,
	"changed":      true,
}

// 支持的 action type。
var validActionTypes = map[string]bool{
	ActionTransition:  true,
	ActionAssign:      true,
	ActionUpdateField: true,
	ActionNotify:      true,
	ActionCreateIssue: true,
	ActionCopyField:   true,
	ActionWebhookCall: true,
}

// 变量引用正则: ${issue.name}, ${version.id}, ${now}, ${issue.assignees} 等。
var varRefRegex = regexp.MustCompile(`\$\{([a-zA-Z_.][a-zA-Z0-9_.]*)\}`)

// ValidateDSLResult 校验结果。
type ValidateDSLResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// ValidateDSL 校验规则 DSL 的合法性。
// 这是创建/更新规则前的强制门：
//  1. JSON 解析通过
//  2. trigger.type 合法
//  3. conditions 字段路径语法 + op 合法
//  4. actions 类型 + 必需字段齐全
//  5. actions 数量 ≤10
//  6. scheduled 触发器必须含 cron
func ValidateDSL(dsl RuleDSL) ValidateDSLResult {
	result := ValidateDSLResult{Valid: true}

	// 1. Trigger 校验
	if dsl.Trigger.Type == "" {
		result.addError("trigger.type 不能为空")
	} else if !validTriggerTypes[dsl.Trigger.Type] {
		result.addError(fmt.Sprintf("trigger.type 不支持: %s", dsl.Trigger.Type))
	}
	if dsl.Trigger.Type == "scheduled" && dsl.Trigger.Cron == "" {
		result.addError("trigger.type=scheduled 时必须提供 trigger.cron")
	}

	// 2. Actions 校验（必需）
	if len(dsl.Actions) == 0 {
		result.addError("actions 不能为空")
	} else if len(dsl.Actions) > 10 {
		result.addError(fmt.Sprintf("actions 数量 %d 超过上限 10", len(dsl.Actions)))
	}

	for i, act := range dsl.Actions {
		prefix := fmt.Sprintf("actions[%d]", i)
		if !validActionTypes[act.Type] {
			result.addError(fmt.Sprintf("%s.type 不支持: %s", prefix, act.Type))
		}
		// 动作类型特定校验
		switch act.Type {
		case ActionTransition:
			if act.Value == nil {
				result.addError(prefix + ".value 不能为空（需指定目标状态）")
			}
		case ActionAssign:
			// value 或 config.strategy 二选一
			if act.Value == nil && len(act.Config) == 0 {
				result.addError(prefix + " 必须指定 value（用户 ID）或 config.strategy")
			}
		case ActionNotify:
			if act.Config["template"] == nil && act.Config["channel"] == nil {
				result.addError(prefix + ".config 必须包含 template 和 channel")
			}
		case ActionUpdateField:
			if act.Field == "" {
				result.addError(prefix + ".field 不能为空")
			}
		}
	}

	// 3. Conditions 校验
	for i, cond := range dsl.Conditions {
		prefix := fmt.Sprintf("conditions[%d]", i)
		if cond.Field == "" {
			result.addError(prefix + ".field 不能为空")
		}
		if !validConditionOps[cond.Op] {
			result.addError(fmt.Sprintf("%s.op 不支持: %s", prefix, cond.Op))
		}
		// is_empty/is_not_empty/changed 不需要 value
		if cond.Op != "is_empty" && cond.Op != "is_not_empty" && cond.Op != "changed" {
			if cond.Value == nil {
				result.addError(prefix + ".value 不能为空")
			}
		}
	}

	// 4. 变量引用校验（仅格式）
	allJSON, _ := json.Marshal(dsl)
	refs := varRefRegex.FindAllStringSubmatch(string(allJSON), -1)
	for _, ref := range refs {
		if !isValidVariable(ref[1]) {
			result.addWarning(fmt.Sprintf("变量引用 ${%s} 可能无效（未识别的变量路径）", ref[1]))
		}
	}

	return result
}

// ValidateDSLBytes 从 JSON 字节流校验。
func ValidateDSLBytes(raw json.RawMessage) (RuleDSL, ValidateDSLResult) {
	var dsl RuleDSL
	if len(raw) == 0 {
		return dsl, ValidateDSLResult{Valid: false, Errors: []string{"DSL 不能为空"}}
	}
	if err := json.Unmarshal(raw, &dsl); err != nil {
		return dsl, ValidateDSLResult{Valid: false, Errors: []string{fmt.Sprintf("JSON 解析失败: %v", err)}}
	}
	result := ValidateDSL(dsl)
	return dsl, result
}

// isValidVariable 检查变量引用是否合法（用于条件求值）。
func isValidVariable(ref string) bool {
	// 已知合法前缀
	prefixes := []string{
		"issue.", "sprint.", "version.", "project.", "actor.", "now", "parent.",
	}
	for _, p := range prefixes {
		if ref == p || strings.HasPrefix(ref, p) {
			return true
		}
	}
	return false
}

// ExtractVariables 从 DSL 中提取所有变量引用（用于文档/调试）。
func ExtractVariables(dsl RuleDSL) []string {
	seen := map[string]bool{}
	raw, _ := json.Marshal(dsl)
	refs := varRefRegex.FindAllStringSubmatch(string(raw), -1)
	for _, r := range refs {
		seen[r[1]] = true
	}
	result := make([]string, 0, len(seen))
	for k := range seen {
		result = append(result, "${"+k+"}")
	}
	return result
}

// CanonicalTriggerKey 返回 trigger 的索引键（用于规则查找）。
// 格式: <project|global>:<trigger_type>
func CanonicalTriggerKey(projectID *int64, triggerType string) string {
	if projectID != nil {
		return fmt.Sprintf("project:%d:%s", *projectID, triggerType)
	}
	return fmt.Sprintf("global:%s", triggerType)
}

func (r *ValidateDSLResult) addError(msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, msg)
}

func (r *ValidateDSLResult) addWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}
