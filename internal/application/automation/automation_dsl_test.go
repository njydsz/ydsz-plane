// Package automation — DSL 校验纯函数单元测试。
//
// 覆盖 ValidateDSL / ValidateDSLBytes / ExtractVariables / CanonicalTriggerKey 四条核心路径，
// 以及 trigger、condition、action 三大子模块的边界条件。
//
// 纯函数、无 IO、毫秒级执行 — 符合单元测试门禁标准。
package automation

import (
	"encoding/json"
	"strings"
	"testing"
)

// ==========================================================================
// ValidateDSL：合法 / 非法场景
// ==========================================================================

func TestValidateDSL_ValidMinimal(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{
			"template": "welcome",
			"channel":  "email",
		}}},
	}
	r := ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("最小合法 DSL 应通过, got errors: %v", r.Errors)
	}
}

func TestValidateDSL_ValidScheduledWithCron(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "scheduled", Cron: "0 9 * * 1"},
		Actions: []Action{{Type: ActionTransition, Value: "completed"}},
	}
	r := ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("scheduled + cron 应通过, got errors: %v", r.Errors)
	}
}

func TestValidateDSL_EmptyTrigger(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("空 trigger.type 应报错")
	}
	if !containsString(r.Errors, "trigger.type 不能为空") {
		t.Errorf("期望错误提示 'trigger.type 不能为空', got: %v", r.Errors)
	}
}

func TestValidateDSL_InvalidTriggerType(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "unknown.event"},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("非法 trigger.type 应报错")
	}
	hasTriggerErr := false
	for _, e := range r.Errors {
		if strings.Contains(e, "trigger.type 不支持") {
			hasTriggerErr = true
		}
	}
	if !hasTriggerErr {
		t.Errorf("期望 trigger.type 错误提示, got: %v", r.Errors)
	}
}

func TestValidateDSL_ScheduledWithoutCron(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "scheduled"},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("scheduled 必须 cron，缺 cron 应报错")
	}
	hasCronErr := false
	for _, e := range r.Errors {
		if strings.Contains(e, "trigger.type=scheduled 时必须提供 trigger.cron") {
			hasCronErr = true
		}
	}
	if !hasCronErr {
		t.Errorf("期望 cron 缺失错误提示, got: %v", r.Errors)
	}
}

func TestValidateDSL_EmptyActions(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("空 actions 应报错")
	}
	if !containsString(r.Errors, "actions 不能为空") {
		t.Errorf("期望 'actions 不能为空', got: %v", r.Errors)
	}
}

func TestValidateDSL_TooManyActions(t *testing.T) {
	actions := make([]Action, 11)
	for i := range actions {
		actions[i] = Action{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}
	}
	dsl := RuleDSL{Trigger: Trigger{Type: "issue.created"}, Actions: actions}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("11 个 actions 应报错超限")
	}
	hasLimitErr := false
	for _, e := range r.Errors {
		if strings.Contains(e, "超过上限 10") {
			hasLimitErr = true
		}
	}
	if !hasLimitErr {
		t.Errorf("期望数量超限错误提示, got: %v", r.Errors)
	}
}

// ==========================================================================
// Action 类型校验
// ==========================================================================

func TestValidateDSL_InvalidActionType(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{{Type: "delete_everything"}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("非法 action.type 应报错")
	}
}

func TestValidateDSL_TransitionRequiresValue(t *testing.T) {
	// transition 缺少 value
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.status_changed"},
		Actions: []Action{{Type: ActionTransition}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("transition 缺少 value 应报错")
	}
	hasErr := false
	for _, e := range r.Errors {
		if strings.HasSuffix(e, "不能为空（需指定目标状态）") {
			hasErr = true
		}
	}
	if !hasErr {
		t.Errorf("期望 transition value 校验错误, got: %v", r.Errors)
	}
}

func TestValidateDSL_TransitionWithValue_OK(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.status_changed"},
		Actions: []Action{{Type: ActionTransition, Value: "in_progress"}},
	}
	r := ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("带 value 的 transition 应通过, got errors: %v", r.Errors)
	}
}

func TestValidateDSL_AssignRequiresValueOrConfig(t *testing.T) {
	// assign 没有 value 也没有 config → 报错
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{{Type: ActionAssign}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("assign 缺少 value 和 config 应报错")
	}

	// 仅提供 value → 通过
	dsl.Actions = []Action{{Type: ActionAssign, Value: 42}}
	r = ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("assign + value 应通过, got: %v", r.Errors)
	}

	// 仅提供 config.strategy → 通过
	dsl.Actions = []Action{{Type: ActionAssign, Config: map[string]any{"strategy": "round_robin"}}}
	r = ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("assign + config.strategy 应通过, got: %v", r.Errors)
	}
}

func TestValidateDSL_NotifyRequiresTemplateAndChannel(t *testing.T) {
	// 缺 template + channel
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"extra": "ignored"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("notify 缺 template+channel 应报错")
	}

	// 齐全
	dsl.Actions = []Action{{Type: ActionNotify, Config: map[string]any{
		"template": "resolved",
		"channel":  "slack",
	}}}
	r = ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("notify + template+channel 应通过, got: %v", r.Errors)
	}
}

func TestValidateDSL_UpdateFieldRequiresField(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.updated"},
		Actions: []Action{{Type: ActionUpdateField, Value: "urgent"}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("update_field 缺 field 应报错")
	}

	dsl.Actions = []Action{{Type: ActionUpdateField, Field: "priority", Value: "urgent"}}
	r = ValidateDSL(dsl)
	if !r.Valid {
		t.Errorf("update_field + field 应通过, got: %v", r.Errors)
	}
}

// ==========================================================================
// Condition 校验
// ==========================================================================

func TestValidateDSL_ConditionEmptyOp(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.updated"},
		Conditions: []Condition{
			{Field: "priority", Op: "", Value: "high"},
		},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("condition.op 为空应报错")
	}
}

func TestValidateDSL_ConditionInvalidOp(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.updated"},
		Conditions: []Condition{
			{Field: "priority", Op: "LIKE", Value: "high"},
		},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("非法 op 应报错")
	}
}

func TestValidateDSL_ConditionValueRequired(t *testing.T) {
	// eq 必须有 value
	dsl := RuleDSL{
		Trigger: Trigger{Type: "issue.updated"},
		Conditions: []Condition{
			{Field: "severity", Op: "gt", Value: nil},
		},
		Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("gt + nil value 应报错")
	}
}

func TestValidateDSL_ConditionNoValueRequiredForSpecialOps(t *testing.T) {
	// is_empty / is_not_empty / changed 不需要 value
	for _, op := range []string{"is_empty", "is_not_empty", "changed"} {
		dsl := RuleDSL{
			Trigger: Trigger{Type: "issue.updated"},
			Conditions: []Condition{
				{Field: "description_json", Op: op, Value: nil},
			},
			Actions: []Action{{Type: ActionNotify, Config: map[string]any{"template": "x", "channel": "email"}}},
		}
		r := ValidateDSL(dsl)
		if !r.Valid {
			t.Errorf("%s 不应要求 value, got errors: %v", op, r.Errors)
		}
	}
}

// ==========================================================================
// ValidateDSLBytes：JSON 字节流入口
// ==========================================================================

func TestValidateDSLBytes_Empty(t *testing.T) {
	_, r := ValidateDSLBytes(json.RawMessage{})
	if r.Valid {
		t.Error("空 JSON 应返回 invalid")
	}
}

func TestValidateDSLBytes_InvalidJSON(t *testing.T) {
	_, r := ValidateDSLBytes(json.RawMessage(`{"trigger":`))
	if r.Valid {
		t.Error("损坏的 JSON 应返回 invalid")
	}
	if len(r.Errors) == 0 || !strings.HasPrefix(r.Errors[0], "JSON 解析失败:") {
		t.Errorf("期望 JSON 解析错误, got: %v", r.Errors)
	}
}

func TestValidateDSLBytes_ValidRoundTrip(t *testing.T) {
	original := RuleDSL{
		Trigger: Trigger{Type: "issue.created"},
		Actions: []Action{{
			Type:   ActionTransition,
			Value:  "done",
			Config: map[string]any{},
		}},
	}
	raw, _ := json.Marshal(original)
	parsed, r := ValidateDSLBytes(raw)
	if !r.Valid {
		t.Errorf("合法 JSON 应通过, got errors: %v", r.Errors)
	}
	if parsed.Trigger.Type != "issue.created" {
		t.Errorf("trigger.type 未正确解析, got: %s", parsed.Trigger.Type)
	}
}

// ==========================================================================
// ExtractVariables / CanonicalTriggerKey
// ==========================================================================

func TestExtractVariables_Basic(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{
			Type: "issue.updated",
			Filter: map[string]any{
				"to_group": "${issue.state.group}",
			},
		},
		Actions: []Action{{
			Type:  ActionNotify,
			Value: "${actor.name} changed ${issue.name}",
		}},
	}
	vars := ExtractVariables(dsl)
	expected := map[string]bool{
		"${issue.state.group}": true,
		"${actor.name}":        true,
		"${issue.name}":        true,
	}
	if len(vars) != len(expected) {
		t.Errorf("期望 %d 个变量, got %d: %v", len(expected), len(vars), vars)
	}
	for _, v := range vars {
		if !expected[v] {
			t.Errorf("unexpected variable: %s", v)
		}
	}
}

func TestExtractVariables_Deduplicates(t *testing.T) {
	// 同一变量出现多次
	raw := json.RawMessage(`{
		"trigger": {"type": "issue.updated"},
		"actions": [
			{"type": "notify", "value": "${issue.name} by ${issue.name}"}
		]
	}`)
	dsl, _ := ValidateDSLBytes(raw)
	vars := ExtractVariables(dsl)
	if len(vars) != 1 || vars[0] != "${issue.name}" {
		t.Errorf("重复变量应去重, got: %v", vars)
	}
}

func TestCanonicalTriggerKey_WithProjectID(t *testing.T) {
	pid := int64(123)
	key := CanonicalTriggerKey(&pid, "issue.created")
	want := "project:123:issue.created"
	if key != want {
		t.Errorf("got %q, want %q", key, want)
	}
}

func TestCanonicalTriggerKey_Global(t *testing.T) {
	key := CanonicalTriggerKey(nil, "scheduled")
	want := "global:scheduled"
	if key != want {
		t.Errorf("got %q, want %q", key, want)
	}
}

// ==========================================================================
// 综合场景：多重错误累积
// ==========================================================================

func TestValidateDSL_MultipleErrorsCumulative(t *testing.T) {
	dsl := RuleDSL{
		Trigger: Trigger{}, // 空 trigger
		Actions: []Action{
			{Type: ActionTransition}, // 缺 value
			{Type: "invalid_action"}, // 非法 type
		},
		Conditions: []Condition{
			{Field: "", Op: ""}, // 全空
		},
	}
	r := ValidateDSL(dsl)
	if r.Valid {
		t.Error("多重错误场景应返回 invalid")
	}
	if len(r.Errors) < 3 {
		t.Errorf("应累积至少 3 条错误, got %d: %v", len(r.Errors), r.Errors)
	}
	t.Logf("累积错误数: %d → %v", len(r.Errors), r.Errors)
}

// ==========================================================================
// helpers
// ==========================================================================

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}
