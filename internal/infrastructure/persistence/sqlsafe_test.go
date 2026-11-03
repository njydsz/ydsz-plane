// Package persistence 测试：SQL 注入防护层（sqlsafe.go）。
//
// 安全基座守护：验证白名单校验、标识符格式校验、SET 子句构建的
// 注入防护行为。任何绕过白名单的输入必须 fail-fast（panic）。
//
// 参考：OWASP A03:2021、docs/architecture/06-权限与安全设计.md。
package persistence

import (
	"strings"
	"testing"
)

func TestSafeTablePrefix_AllowsWhitelisted(t *testing.T) {
	t.Parallel()

	allowed := []string{"task", "requirement", "defect"}
	for _, prefix := range allowed {
		if got := SafeTablePrefix(prefix); got != prefix {
			t.Errorf("SafeTablePrefix(%q) = %q, want %q", prefix, got, prefix)
		}
	}
}

func TestSafeTablePrefix_PanicsOnInjection(t *testing.T) {
	t.Parallel()

	// 攻击者尝试通过表名前缀注入任意 SQL 片段。
	malicious := []string{
		"task; DROP TABLE users; --",
		"task' OR '1'='1",
		"task`",
		"task\"",
		"unknown_table",
		"",
		"TASK", // 大小写敏感，不在白名单
	}
	for _, prefix := range malicious {
		prefix := prefix
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("SafeTablePrefix(%q) should panic", prefix)
				}
			}()
			_ = SafeTablePrefix(prefix)
		})
	}
}

func TestSafeTableName_AllowsWhitelisted(t *testing.T) {
	t.Parallel()

	allowed := []string{
		"task", "requirement", "defect", "sprint_tasks", "sprint_requirements",
		"sprint_defects", "sprint_issues", "automation_rules", "automation_rule_executions",
		"knowledge_spaces", "knowledge_pages", "knowledge_page_comments",
		"page_templates", "page_shares", "issues", "states", "webhooks",
		"webhook_logs", "versions", "sprints", "users", "workspaces",
	}
	for _, name := range allowed {
		if got := SafeTableName(name); got != name {
			t.Errorf("SafeTableName(%q) = %q, want %q", name, got, name)
		}
	}
}

func TestSafeTableName_PanicsOnInjection(t *testing.T) {
	t.Parallel()

	malicious := []string{
		"users; DROP TABLE users; --",
		"users' OR '1'='1",
		"task\"); DELETE FROM task; --",
		"nonexistent_table",
		"",
	}
	for _, name := range malicious {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("SafeTableName(%q) should panic", name)
				}
			}()
			_ = SafeTableName(name)
		})
	}
}

func TestSafeIdentifier(t *testing.T) {
	t.Parallel()

	valid := []string{
		"id", "name", "created_at", "workspace_id", "a1_b2", "_leading",
	}
	for _, ident := range valid {
		if got := SafeIdentifier(ident); got != ident {
			t.Errorf("SafeIdentifier(%q) = %q, want %q", ident, got, ident)
		}
	}
}

func TestSafeIdentifier_PanicsOnInjection(t *testing.T) {
	t.Parallel()

	malicious := []string{
		"name; DROP TABLE users; --",
		"name' OR '1'='1",
		"name\"; DELETE FROM task; --",
		"1column",   // 数字开头
		"",          // 空
		"name-table", // 含连字符
		"name table", // 含空格
		"name.table", // 含点号（不作为标识符）
	}
	for _, ident := range malicious {
		ident := ident
		t.Run(ident, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("SafeIdentifier(%q) should panic", ident)
				}
			}()
			_ = SafeIdentifier(ident)
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ident string
		want  bool
	}{
		{"id", true},
		{"created_at", true},
		{"a1_b2", true},
		{"_leading", true},
		{"1column", false},
		{"name-table", false},
		{"name table", false},
		{"name;DROP", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsValidIdentifier(c.ident); got != c.want {
			t.Errorf("IsValidIdentifier(%q) = %v, want %v", c.ident, got, c.want)
		}
	}
}

func TestSetClause_BuildsParameterized(t *testing.T) {
	t.Parallel()

	values := map[string]any{"name": "foo", "priority": "high"}
	clause, nextIdx := SetClause(values, &PostgresDialect{}, 3)

	if clause == "" {
		t.Fatal("SetClause returned empty clause")
	}
	if nextIdx != 5 {
		t.Errorf("nextIdx = %d, want 5", nextIdx)
	}
	// 列名应被双引号包裹，占位符应使用 $N。
	for _, want := range []string{`"name"=$`, `"priority"=$`} {
		if !strings.Contains(clause, want) {
			t.Errorf("clause %q missing %q", clause, want)
		}
	}
	// 不应出现用户输入的原始值（防注入）。
	if strings.Contains(clause, "foo") || strings.Contains(clause, "high") {
		t.Errorf("clause leaked raw values: %q", clause)
	}
}

func TestSetClause_EmptyMap(t *testing.T) {
	t.Parallel()

	clause, nextIdx := SetClause(map[string]any{}, &PostgresDialect{}, 3)
	if clause != "" {
		t.Errorf("clause = %q, want empty", clause)
	}
	if nextIdx != 3 {
		t.Errorf("nextIdx = %d, want 3", nextIdx)
	}
}

func TestSetClause_PanicsOnBadColumn(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("SetClause with malicious column should panic")
		}
	}()
	clause, _ := SetClause(map[string]any{"name; DROP TABLE users; --": "x"}, &PostgresDialect{}, 1)
	_ = clause
}

func TestQuoteAndValidate(t *testing.T) {
	t.Parallel()

	got := QuoteAndValidate(&PostgresDialect{}, "my_column")
	if got != `"my_column"` {
		t.Errorf("QuoteAndValidate = %q, want %q", got, `"my_column"`)
	}
}
