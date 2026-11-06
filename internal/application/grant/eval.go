// Package grant — 条件表达式求值器。
package grant

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Evaluate 对表达式求值。
//
// 仅识别白名单操作符；未知操作符或字段返回错误（拒绝而非放行）。
func Evaluate(expr *Expr, ctx Context) (bool, error) {
	if expr == nil {
		return true, nil
	}

	switch expr.Op {
	case OpAnd:
		if len(expr.All) == 0 {
			return false, fmt.Errorf("grant: and requires 'all'")
		}
		for _, sub := range expr.All {
			ok, err := Evaluate(sub, ctx)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil

	case OpOr:
		if len(expr.Any) == 0 {
			return false, fmt.Errorf("grant: or requires 'any'")
		}
		for _, sub := range expr.Any {
			ok, err := Evaluate(sub, ctx)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil

	case OpNot:
		if expr.Inner == nil {
			return false, fmt.Errorf("grant: not requires 'inner'")
		}
		ok, err := Evaluate(expr.Inner, ctx)
		if err != nil {
			return false, err
		}
		return !ok, nil

	case OpEq, OpNe, OpIn, OpContains, OpGt, OpLt:
		return evalComparison(expr, ctx)
	}
	return false, fmt.Errorf("grant: unsupported op %q", expr.Op)
}

func evalComparison(expr *Expr, ctx Context) (bool, error) {
	if !AllowedFields[expr.Field] {
		return false, fmt.Errorf("grant: field %q is not allowed", expr.Field)
	}
	actual, present := ctx[expr.Field]
	if !present {
		// 字段缺失：比较类操作一律不匹配（保守策略）
		return false, nil
	}

	switch expr.Op {
	case OpEq:
		return compareEqual(actual, expr.Value)
	case OpNe:
		eq, err := compareEqual(actual, expr.Value)
		return !eq, err
	case OpIn:
		// actual 在 expr.Values 中
		for _, v := range expr.Values {
			if eq, err := compareEqual(actual, mustJSON(v)); err == nil && eq {
				return true, nil
			}
		}
		return false, nil
	case OpContains:
		// actual 是数组且包含 expr.Value
		return containsValue(actual, expr.Value)
	case OpGt, OpLt:
		return compareOrder(actual, expr.Value, expr.Op == OpGt)
	}
	return false, fmt.Errorf("grant: unsupported comparison %q", expr.Op)
}

func compareEqual(actual any, expected json.RawMessage) (bool, error) {
	var want any
	if err := json.Unmarshal(expected, &want); err != nil {
		return false, err
	}
	return looseEqual(actual, want), nil
}

// looseEqual 宽松相等比较（数字统一按 float64，字符串忽略大小写）。
func looseEqual(a, b any) bool {
	// 数字统一
	if af, ok := toFloat(a); ok {
		if bf, ok := toFloat(b); ok {
			return af == bf
		}
	}
	// 字符串忽略大小写
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return equalFold(as, bs)
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func containsValue(actual any, expected json.RawMessage) (bool, error) {
	arr, ok := actual.([]any)
	if !ok {
		// 单值退化为相等
		return compareEqual(actual, expected)
	}
	var want any
	if err := json.Unmarshal(expected, &want); err != nil {
		return false, err
	}
	for _, v := range arr {
		if looseEqual(v, want) {
			return true, nil
		}
	}
	return false, nil
}

func compareOrder(actual any, expected json.RawMessage, greater bool) (bool, error) {
	af, aok := toFloat(actual)
	var want float64
	if err := json.Unmarshal(expected, &want); err != nil {
		return false, err
	}
	if !aok {
		return false, fmt.Errorf("grant: field is not numeric")
	}
	if greater {
		return af > want, nil
	}
	return af < want, nil
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
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

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return b
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// Validate 校验表达式结构合法性（写入前调用）。
func Validate(raw json.RawMessage) error {
	if len(raw) == 0 {
		return fmt.Errorf("grant: condition is required")
	}
	var expr Expr
	if err := json.Unmarshal(raw, &expr); err != nil {
		return fmt.Errorf("grant: invalid condition JSON: %w", err)
	}
	return validateExpr(&expr)
}

func validateExpr(e *Expr) error {
	if e == nil {
		return nil
	}
	switch e.Op {
	case OpAnd:
		if len(e.All) == 0 {
			return fmt.Errorf("grant: and requires at least one sub-expression")
		}
		for _, sub := range e.All {
			if err := validateExpr(sub); err != nil {
				return err
			}
		}
	case OpOr:
		if len(e.Any) == 0 {
			return fmt.Errorf("grant: or requires at least one sub-expression")
		}
		for _, sub := range e.Any {
			if err := validateExpr(sub); err != nil {
				return err
			}
		}
	case OpNot:
		if e.Inner == nil {
			return fmt.Errorf("grant: not requires inner")
		}
		return validateExpr(e.Inner)
	case OpEq, OpNe, OpIn, OpContains, OpGt, OpLt:
		if !AllowedFields[e.Field] {
			return fmt.Errorf("grant: field %q is not allowed", e.Field)
		}
	default:
		return fmt.Errorf("grant: unsupported op %q", e.Op)
	}
	return nil
}
