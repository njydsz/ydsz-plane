// Package apitoken — API Token 纯函数单元测试。
//
// 测试范围：token 生成/哈希、scope 校验、scope 覆盖关系与权限映射。
// 数据库依赖逻辑由 E2E 覆盖（见 web/e2e/api-tokens.spec.ts）。
package apitoken

import (
	"strings"
	"testing"
)

// TestGenerateTokenFormat 验证 token 前缀、长度与字符集。
func TestGenerateTokenFormat(t *testing.T) {
	raw, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if !strings.HasPrefix(raw, TokenPrefix) {
		t.Errorf("token %q missing prefix %q", raw, TokenPrefix)
	}
	// ydz_ (4) + 32 字节 base64url 无填充 (43)
	if len(raw) != len(TokenPrefix)+43 {
		t.Errorf("token length = %d, want %d", len(raw), len(TokenPrefix)+43)
	}
}

// TestGenerateTokenUniqueness 验证连续生成的 token 不重复。
func TestGenerateTokenUniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 64)
	for i := 0; i < 64; i++ {
		raw, err := GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken #%d: %v", i, err)
		}
		if _, dup := seen[raw]; dup {
			t.Fatalf("duplicate token generated: %q", raw)
		}
		seen[raw] = struct{}{}
	}
}

// TestHashTokenDeterministic 验证同一 token 哈希稳定且长度正确（64 位 hex）。
func TestHashTokenDeterministic(t *testing.T) {
	raw, _ := GenerateToken()
	h1 := HashToken(raw)
	h2 := HashToken(raw)
	if h1 != h2 {
		t.Errorf("hash not deterministic: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64", len(h1))
	}
	// 存储的 hash 不应包含原始 token
	if strings.Contains(h1, raw) {
		t.Errorf("hash leaks raw token")
	}
}

// TestLooksLikeAPIToken 验证前缀分流逻辑。
func TestLooksLikeAPIToken(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"ydz_" + strings.Repeat("a", 43), true},
		{"ydz_short", true},
		{"not-a-token", false},
		{"", false},
		{"ydz", false}, // 前缀不完整
	}
	for _, tc := range cases {
		if got := LooksLikeAPIToken(tc.in); got != tc.want {
			t.Errorf("LooksLikeAPIToken(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestValidateScopes 覆盖合法/空/非法/重复四类输入。
func TestValidateScopes(t *testing.T) {
	cases := []struct {
		name   string
		scopes []string
		wantOK bool
	}{
		{"default read", []string{ScopeWorkspaceRead}, true},
		{"full write", []string{ScopeWorkspaceWrite, ScopeIssuesWrite}, true},
		{"wildcard", []string{ScopeAll}, true},
		{"all scopes", AllScopes, true},
		{"empty", nil, false},
		{"unknown scope", []string{"read:secrets"}, false},
		{"duplicate", []string{ScopeIssuesRead, ScopeIssuesRead}, false},
		{"empty string", []string{""}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := ValidateScopes(tc.scopes)
			if ok != tc.wantOK {
				t.Errorf("ValidateScopes(%v) ok = %v, want %v", tc.scopes, ok, tc.wantOK)
			}
		})
	}
}

// TestScopeCovers 覆盖 scope 覆盖关系矩阵（含 write 隐含 read 与通配符）。
func TestScopeCovers(t *testing.T) {
	cases := []struct {
		name     string
		owned    []string
		required string
		want     bool
	}{
		{"exact match", []string{ScopeIssuesRead}, ScopeIssuesRead, true},
		{"write covers read", []string{ScopeIssuesWrite}, ScopeIssuesRead, true},
		{"read does not cover write", []string{ScopeIssuesRead}, ScopeIssuesWrite, false},
		{"different domain", []string{ScopeIssuesWrite}, ScopeSprintsWrite, false},
		{"wildcard covers all", []string{ScopeAll}, ScopeVersionsWrite, true},
		{"wildcard covers audit", []string{ScopeAll}, ScopeAuditRead, true},
		{"empty owned", nil, ScopeWorkspaceRead, false},
		{"workspace write covers read", []string{ScopeWorkspaceWrite}, ScopeWorkspaceRead, true},
		{"audit needs explicit", []string{ScopeWorkspaceWrite}, ScopeAuditRead, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ScopeCovers(tc.owned, tc.required); got != tc.want {
				t.Errorf("ScopeCovers(%v, %q) = %v, want %v", tc.owned, tc.required, got, tc.want)
			}
		})
	}
}

// TestPermissionScope 验证 RBAC 权限 → scope 映射的完备性：
// 所有当前权限常量都必须有映射（防止未来新增权限被静默放行）。
func TestPermissionScope(t *testing.T) {
	perms := []string{
		"workspace:read",
		"workspace:update", "workspace:delete",
		"member:invite", "member:remove", "member:change_role",
		"project:create", "project:delete",
		"audit:read",
		"issue:create", "issue:delete",
		"version:create", "version:release", "version:delete", "version:update",
	}
	for _, p := range perms {
		if _, ok := PermissionScope(p); !ok {
			t.Errorf("PermissionScope(%q) missing mapping — 新增权限必须显式声明所需 scope", p)
		}
	}
}

// TestPermissionScopeWriteRequiresWrite 验证写权限映射到 write scope（而非 read）。
func TestPermissionScopeWriteRequiresWrite(t *testing.T) {
	writePerms := []string{"workspace:update", "member:invite", "project:create", "issue:create", "version:release"}
	for _, p := range writePerms {
		scope, _ := PermissionScope(p)
		if !strings.HasPrefix(scope, "write:") {
			t.Errorf("PermissionScope(%q) = %q, want write scope", p, scope)
		}
		// 只持有对应 read scope 的令牌必须被拒绝
		if ScopeCovers([]string{strings.Replace(scope, "write:", "read:", 1)}, scope) {
			t.Errorf("read scope must not cover %q", scope)
		}
	}
}
