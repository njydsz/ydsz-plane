// Package grant — 条件授权（属性基授权，ABAC）。
//
// 对标 Plane Enterprise 的 Conditional Grants。
//
// 在角色权限（RBAC）之上叠加基于属性的约束：
//   - 字段比较：eq / ne / gt / lt / in / contains
//   - 逻辑组合：and / or / not
//   - 受支持字段（白名单）：projectId / type / stateGroup / priority /
//     creatorId / assigneeId / labelId / workspaceId / moduleId / sprintId
//
// 安全：条件表达式为受限 JSON，求值器只识别白名单操作符与字段，
// 不执行任意表达式，防止注入。
package grant

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"time"
)

// Rule 条件授权规则。
type Rule struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	RoleSlugs   []string        `json:"role_slugs"`
	Permissions []string        `json:"permissions"`
	Scope       string          `json:"scope"`
	Condition   json.RawMessage `json:"condition"`
	Enabled     bool            `json:"enabled"`
	Priority    int             `json:"priority"`
	CreatedBy   int64           `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Op 操作符。
type Op string

const (
	// OpAnd 逻辑与。
	OpAnd Op = "and"
	// OpOr 逻辑或。
	OpOr Op = "or"
	// OpNot 逻辑非。
	OpNot Op = "not"
	// OpEq 相等。
	OpEq Op = "eq"
	// OpNe 不等。
	OpNe Op = "ne"
	// OpIn 包含于。
	OpIn Op = "in"
	// OpContains 数组包含。
	OpContains Op = "contains"
	// OpGt 大于。
	OpGt Op = "gt"
	// OpLt 小于。
	OpLt Op = "lt"
)

// Expr 条件表达式节点。
type Expr struct {
	Op     Op              `json:"op"`
	Field  string          `json:"field,omitempty"`
	Value  json.RawMessage `json:"value,omitempty"`
	Values []any           `json:"values,omitempty"`
	Inner  *Expr           `json:"inner,omitempty"`
	All    []*Expr         `json:"all,omitempty"`
	Any    []*Expr         `json:"any,omitempty"`
}

// AllowedFields 允许参与条件判断的字段白名单。
var AllowedFields = map[string]bool{
	"projectId":   true,
	"type":        true,
	"stateGroup":  true,
	"priority":    true,
	"creatorId":   true,
	"assigneeId":  true,
	"labelId":     true,
	"workspaceId": true,
	"moduleId":    true,
	"sprintId":    true,
}

// Context 求值上下文（属性集合）。
type Context map[string]any

// CreateRuleInput 创建规则入参。
type CreateRuleInput struct {
	WorkspaceID int64
	Name        string
	Description string
	RoleSlugs   []string
	Permissions []string
	Scope       string
	Condition   json.RawMessage
	Priority    int
	CreatedBy   int64
}

// UpdateRuleInput 更新规则入参。
type UpdateRuleInput struct {
	Name        *string
	Description *string
	RoleSlugs   []string
	Permissions []string
	Scope       *string
	Condition   json.RawMessage
	Enabled     *bool
	Priority    *int
}

func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
