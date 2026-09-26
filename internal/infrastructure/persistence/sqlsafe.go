// Package persistence — SQL 注入防护层（P0-安全加固）。
//
// 对标 OWASP A03:2021、阿里/美团安全编码规范强制参数化原则。
// 所有动态表名、列名构建 SQL 前必须经本包白名单校验。
//
// 设计要点:
//   - 白名单映射：合法表名前缀 → 校验函数，拒绝非预期值
//   - 快速失败：invalid input 直接 panic（与 dialect.go 的 JSONB key 校验模式一致）
//   - CI 门禁：配合 Makefile 中的 security-scan 目标，fmt.Sprintf 拼接 SQL 即阻断
package persistence

import (
	"fmt"
	"regexp"
	"strings"
)

// --- 白名单定义 ---

// validTablePrefixes 是合法的表名前缀白名单。
// 所有工作项关联表（xxx_assignees / xxx_labels 等）只允许此前缀。
var validTablePrefixes = map[string]struct{}{
	"task":        {},
	"requirement": {},
	"defect":      {},
	// sprint_xxx 系列使用的 "sprint_tasks/defects/requirements" 归一为 sprint_issues 后不再新增
}

// validTableNames 是完整表名白名单（供 triggerProgressRollup 等直接使用 tableName 的场景）。
// 只读引用：实际表名在 database/sql 初始化时确定，不含运行时输入。
var validTableNames = map[string]struct{}{
	"task":                        {},
	"requirement":                 {},
	"defect":                      {},
	"sprint_tasks":                {},
	"sprint_requirements":         {},
	"sprint_defects":              {},
	"sprint_issues":               {},
	"automation_rules":            {},
	"automation_rule_executions":  {},
	"knowledge_spaces":            {},
	"knowledge_pages":             {},
	"knowledge_page_comments":     {},
	"page_templates":              {},
	"page_shares":                 {},
	"issues":                      {},
	"states":                      {},
	"webhooks":                    {},
	"webhook_logs":                {},
	"versions":                    {},
	"sprints":                     {},
	"users":                       {},
	"workspaces":                  {},
}

// identifierPattern 匹配合法的 PostgreSQL 标识符：字母/数字/下划线，不以数字开头。
var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// --- 安全函数 ---

// SafeTablePrefix 校验表名前缀在白名单中。
// 不在白名单 → panic（fail-fast，与 dialect.go 的 isValidJSONBKey 风格一致）。
// 调用方应为 switch 已穷尽的枚举值；此函数为防御性兜底。
func SafeTablePrefix(prefix string) string {
	if _, ok := validTablePrefixes[prefix]; !ok {
		panic(fmt.Sprintf("persistence: invalid table prefix %q (disallowed by SQL injection protection)", prefix))
	}
	return prefix
}

// SafeTableName 校验完整表名是否在白名单中。
// 用于需要直接传入 tableName 的场景（如 triggerProgressRollup）。
// 不在白名单 → panic。
func SafeTableName(tableName string) string {
	if _, ok := validTableNames[tableName]; !ok {
		panic(fmt.Sprintf("persistence: invalid table name %q (disallowed by SQL injection protection)", tableName))
	}
	return tableName
}

// SafeIdentifier 校验标识符格式（字母/数字/下划线，不以数字开头）。
// 用于动态列名、别名等场景。不检查白名单，仅防注入。
func SafeIdentifier(ident string) string {
	if !identifierPattern.MatchString(ident) {
		panic(fmt.Sprintf("persistence: invalid SQL identifier %q (仅允许 [a-zA-Z_][a-zA-Z0-9_])", ident))
	}
	return ident
}

// IsValidIdentifier 不 panic，返回 bool。适用于需要错误处理的场景。
func IsValidIdentifier(ident string) bool {
	return identifierPattern.MatchString(ident)
}

// QuoteAndValidate 先校验标识符合法性，再用 QuoteIdentifier 转义。
func QuoteAndValidate(d Dialect, ident string) string {
	SafeIdentifier(ident)
	return d.QuoteIdentifier(ident)
}

// SetClause 安全构建 SET 子句。将 map 中的 key 校验后拼接为 "col1"=$1,"col2"=$2。
// 返回 setClause 和参数 index 起始值。
// 用于替代 fmt.Sprintf(`UPDATE x SET %s WHERE ...`, joinSets(sets)) 模式。
// 使用示例：
//
//	sets := map[string]any{"name": "foo", "priority": "high"}
//	clause, nextIdx := persistence.SetClause(sets, dialect, 3)
//	// clause = "name"=$3,"priority"=$4  nextIdx = 5
func SetClause(values map[string]any, d Dialect, startIdx int) (string, int) {
	if len(values) == 0 {
		return "", startIdx
	}
	var parts []string
	idx := startIdx
	for col := range values {
		SafeIdentifier(col) // 防御性校验列名
		parts = append(parts, fmt.Sprintf("%s=%s", d.QuoteIdentifier(col), d.Placeholder(idx)))
		idx++
	}
	return strings.Join(parts, ","), idx
}
