// Package pql — SQL 编译器。
//
// 将 AST 编译为参数化 SQL 片段，杜绝字符串拼接注入。
package pql

import (
	"fmt"
	"strings"
)

// CompiledQuery 编译结果。
type CompiledQuery struct {
	// WhereSQL 形如 "col1 = $1 AND (col2 >= $2 OR col3 ~ $3)"（不含 WHERE 关键字）
	WhereSQL string
	// Args 与占位符顺序对应
	Args []any
	// SortColumn 允许排序的列名（白名单校验后）
	SortColumn string
	SortDir    string
	Limit      int
	Offset     int
}

// Compiler 编译器。
type Compiler struct {
	mappings map[string]FieldMapping
	args     []any
}

// NewCompiler 创建编译器。
func NewCompiler() *Compiler {
	return &Compiler{mappings: DefaultFieldMappings}
}

// Compile 将查询编译为 SQL。
//
// @userIDs 用于解析 "me" 与 "@username"；返回的 map 为 name → id。
func (c *Compiler) Compile(q *Query, meUserID int64, userIDs map[string]int64) (*CompiledQuery, error) {
	out := &CompiledQuery{
		SortDir: strings.ToLower(q.SortDir),
		Limit:   q.Limit,
		Offset:  q.Offset,
	}
	if out.SortDir == "" {
		out.SortDir = "asc"
	}

	if q.Where != nil {
		sql, err := c.compileExpr(q.Where, meUserID, userIDs)
		if err != nil {
			return nil, err
		}
		out.WhereSQL = sql
	}
	out.Args = c.args

	if q.SortBy != "" {
		if !AllowedSortFields[q.SortBy] {
			return nil, fmt.Errorf("pql: field %q is not allowed in sort", q.SortBy)
		}
		m, ok := c.mappings[q.SortBy]
		if !ok {
			return nil, fmt.Errorf("pql: unknown sort field %q", q.SortBy)
		}
		out.SortColumn = m.Column
	}
	return out, nil
}

func (c *Compiler) addArg(v any) string {
	c.args = append(c.args, v)
	return fmt.Sprintf("$%d", len(c.args))
}

func (c *Compiler) compileExpr(e Expr, meUserID int64, userIDs map[string]int64) (string, error) {
	switch n := e.(type) {
	case *LogicalExpr:
		left, err := c.compileExpr(n.Left, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		right, err := c.compileExpr(n.Right, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		op := "AND"
		if n.Op == TokenOr {
			op = "OR"
		}
		return fmt.Sprintf("(%s %s %s)", left, op, right), nil

	case *NotExpr:
		inner, err := c.compileExpr(n.Inner, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(NOT %s)", inner), nil

	case *ComparisonExpr:
		return c.compileComparison(n, meUserID, userIDs)
	}
	return "", fmt.Errorf("pql: unsupported expression node %T", e)
}

func (c *Compiler) compileComparison(n *ComparisonExpr, meUserID int64, userIDs map[string]int64) (string, error) {
	m, ok := c.mappings[n.Field]
	if !ok {
		return "", fmt.Errorf("pql: unknown field %q", n.Field)
	}
	col := m.Column

	// 枚举字段：校验取值合法性（避免非法值进入查询语义）
	if m.Kind == "enum" && n.Op != TokenIn && n.Value != "" && !n.IsMe && n.AtUser == "" {
		if !m.EnumSet[strings.ToLower(n.Value)] {
			return "", fmt.Errorf("pql: invalid value %q for field %q", n.Value, n.Field)
		}
	}

	switch n.Op {
	case TokenEq:
		arg, err := c.valueArg(n, m, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s = %s", col, c.addArg(arg)), nil

	case TokenNe:
		arg, err := c.valueArg(n, m, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s IS NULL OR %s <> %s)", col, col, c.addArg(arg)), nil

	case TokenGt, TokenGte, TokenLt, TokenLte:
		arg, err := c.valueArg(n, m, meUserID, userIDs)
		if err != nil {
			return "", err
		}
		op := map[TokenKind]string{
			TokenGt: ">", TokenGte: ">=", TokenLt: "<", TokenLte: "<=",
		}[n.Op]
		return fmt.Sprintf("%s %s %s", col, op, c.addArg(arg)), nil

	case TokenTilde:
		if n.Value == "" {
			return "", fmt.Errorf("pql: like operator requires a value")
		}
		return fmt.Sprintf("%s ILIKE %s", col, c.addArg("%"+n.Value+"%")), nil

	case TokenIn:
		if len(n.Values) == 0 {
			return "", fmt.Errorf("pql: in operator requires at least one value")
		}
		placeholders := make([]string, 0, len(n.Values))
		for _, v := range n.Values {
			placeholders = append(placeholders, c.addArg(v))
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ", ")), nil
	}
	return "", fmt.Errorf("pql: unsupported operator %s", n.Op)
}

// valueArg 将比较节点的值转换为 SQL 参数。
func (c *Compiler) valueArg(n *ComparisonExpr, m FieldMapping, meUserID int64, userIDs map[string]int64) (any, error) {
	switch {
	case n.IsMe:
		if m.Kind != "user" {
			return nil, fmt.Errorf("pql: 'me' is only valid for user fields")
		}
		return meUserID, nil
	case n.AtUser != "":
		if m.Kind != "user" {
			return nil, fmt.Errorf("pql: '@user' is only valid for user fields")
		}
		id, ok := userIDs[strings.ToLower(n.AtUser)]
		if !ok {
			return nil, fmt.Errorf("pql: unknown user @%s", n.AtUser)
		}
		return id, nil
	}

	switch m.Kind {
	case "number":
		var num float64
		if _, err := fmt.Sscanf(n.Value, "%f", &num); err != nil {
			return nil, fmt.Errorf("pql: field %q expects a number, got %q", n.Field, n.Value)
		}
		return num, nil
	case "date":
		return NormalizeDate(n.Value), nil
	default:
		return n.Value, nil
	}
}
