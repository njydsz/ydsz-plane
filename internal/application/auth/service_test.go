// Package auth 测试：验证 RBAC 权限矩阵与 token 签发/解析行为。
package auth

import (
	"testing"
	"time"
)

// TestRolePermissionMatrix 验证权限矩阵中各角色被授予/拒绝的权限。
func TestRolePermissionMatrix(t *testing.T) {
	cases := []struct {
		role WorkspaceRole
		perm string
		want bool
	}{
		// Owner：全部权限
		{RoleOwner, PermWorkspaceRead, true},
		{RoleOwner, PermWorkspaceDelete, true},
		{RoleOwner, PermMemberChangeRole, true},
		{RoleOwner, PermProjectCreate, true},
		// Admin：不可删除工作空间、不可变更角色
		{RoleAdmin, PermWorkspaceRead, true},
		{RoleAdmin, PermMemberInvite, true},
		{RoleAdmin, PermProjectDelete, true},
		{RoleAdmin, PermWorkspaceDelete, false},
		{RoleAdmin, PermMemberChangeRole, false},
		// Member：只读 + 创建
		{RoleMember, PermWorkspaceRead, true},
		{RoleMember, PermProjectCreate, true},
		{RoleMember, PermIssueCreate, true},
		{RoleMember, PermMemberInvite, false},
		{RoleMember, PermProjectDelete, false},
		// Guest：仅只读
		{RoleGuest, PermWorkspaceRead, true},
		{RoleGuest, PermProjectCreate, false},
		{RoleGuest, PermIssueCreate, false},
	}
	for _, c := range cases {
		m := WorkspaceMembership{Role: c.role, WorkspaceID: 1, UserID: 1}
		got := m.HasPermission(c.perm)
		if got != c.want {
			t.Errorf("role=%s perm=%s: got %v want %v", c.role, c.perm, got, c.want)
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
		{RoleOwner, RoleAdmin, true},
		{RoleAdmin, RoleAdmin, true},
		{RoleMember, RoleAdmin, false},
		{RoleGuest, RoleMember, false},
	}
	for _, c := range cases {
		if got := c.role.IsAtLeast(c.min); got != c.want {
			t.Errorf("role=%s IsAtLeast(%s): got %v want %v", c.role, c.min, got, c.want)
		}
	}
}

// TestInvalidRole 验证未知角色枚举值被判定为非法。
func TestInvalidRole(t *testing.T) {
	if WorkspaceRole("hacker").IsValid() {
		t.Errorf("unknown role should be invalid")
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
