// Package auth 测试：验证 RBAC 角色枚举与 token 签发/解析行为。
//
// 注意：RBAC 权限矩阵已迁移至 DB-backed internal/rbac（internal/rbac/store.go），
// 本文件不再维护硬编码权限矩阵，仅覆盖角色枚举语义与 token 往返。
package auth

import (
	"testing"
	"time"
)

// TestRoleValidity 验证角色枚举值合法性。
func TestRoleValidity(t *testing.T) {
	valid := []WorkspaceRole{RoleOwner, RoleAdmin, RolePM, RolePO, RoleTechLead, RoleQALead, RoleDev, RoleGuest}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("role %q should be valid", r)
		}
	}
	if WorkspaceRole("hacker").IsValid() {
		t.Errorf("unknown role should be invalid")
	}
}

// TestRoleLevel 验证角色层级单调递增。
func TestRoleLevel(t *testing.T) {
	// 层级语义：admin=100（平台级）> owner=80（空间级）> pm/po/techlead=50 > dev=30 > guest=10
	if RoleAdmin.Level() <= RoleOwner.Level() {
		t.Errorf("admin level must exceed owner level")
	}
	if RoleOwner.Level() <= RoleGuest.Level() {
		t.Errorf("owner level must exceed guest level")
	}
	// 所有角色层级非负
	for _, r := range []WorkspaceRole{RoleOwner, RoleAdmin, RolePM, RolePO, RoleTechLead, RoleQALead, RoleDev, RoleGuest} {
		if r.Level() < 0 {
			t.Errorf("role %q has negative level", r)
		}
	}
}

// TestRoleIsAtLeast 验证角色级别比较逻辑。
func TestRoleIsAtLeast(t *testing.T) {
	cases := []struct {
		role WorkspaceRole
		min  WorkspaceRole
		want bool
	}{
		{RoleAdmin, RoleOwner, true}, // admin(100) >= owner(80)
		{RoleOwner, RoleOwner, true},
		{RoleDev, RoleAdmin, false},
		{RoleGuest, RoleDev, false},
		{RoleAdmin, RoleGuest, true},
		{RoleOwner, RolePM, true}, // owner(80) >= pm(50)
	}
	for _, c := range cases {
		if got := c.role.IsAtLeast(c.min); got != c.want {
			t.Errorf("role=%s IsAtLeast(%s): got %v want %v", c.role, c.min, got, c.want)
		}
	}
}

// 令牌签发/解析往返（不依赖 DB）
func TestTokenRoundTrip(t *testing.T) {
	svc := NewService(nil, "test-secret", "ydsz-plane", 15*time.Minute, 720*time.Hour, 4, true)

	pair, err := svc.issuePair(42, "a@b.c", "Tester", "")
	if err != nil {
		t.Fatalf("issuePair: %v", err)
	}
	if pair.TokenType != "Bearer" || pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("incomplete pair: %+v", pair)
	}

	uid, err := svc.ParseAccess(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if uid != 42 {
		t.Fatalf("uid = %d, want 42", uid)
	}
}

// TestParseAccessRejectsRefreshKind 验证 refresh 令牌不能当作 access 令牌使用。
func TestParseAccessRejectsRefreshKind(t *testing.T) {
	svc := NewService(nil, "test-secret", "ydsz-plane", 15*time.Minute, 720*time.Hour, 4, true)
	pair, err := svc.issuePair(1, "a@b.c", "T", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ParseAccess(pair.RefreshToken); err == nil {
		t.Fatal("refresh token must not be accepted as access token")
	}
}

// TestParseAccessRejectsWrongSecret 验证使用其他密钥签发的令牌会被拒绝。
func TestParseAccessRejectsWrongSecret(t *testing.T) {
	a := NewService(nil, "secret-a", "ydsz-plane", time.Minute, time.Hour, 4, true)
	b := NewService(nil, "secret-b", "ydsz-plane", time.Minute, time.Hour, 4, true)
	pair, err := a.issuePair(1, "a@b.c", "T", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.ParseAccess(pair.AccessToken); err == nil {
		t.Fatal("token signed by another secret must be rejected")
	}
}

// TestHashPasswordVerifiable 验证生成的 bcrypt 散列与明文不同且长度合规。
func TestHashPasswordVerifiable(t *testing.T) {
	svc := NewService(nil, "s", "i", time.Minute, time.Hour, 4, true)
	hash, err := svc.HashPassword("Admin@123")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "Admin@123" || len(hash) < 20 {
		t.Fatalf("bad hash: %q", hash)
	}
}
