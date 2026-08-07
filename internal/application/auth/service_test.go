package auth

import (
	"testing"
	"time"
)

func TestRolePermissionMatrix(t *testing.T) {
	cases := []struct {
		role WorkspaceRole
		perm string
		want bool
	}{
		// Owner: all
		{RoleOwner, PermWorkspaceRead, true},
		{RoleOwner, PermWorkspaceDelete, true},
		{RoleOwner, PermMemberChangeRole, true},
		{RoleOwner, PermProjectCreate, true},
		// Admin: can't delete workspace, can't change role
		{RoleAdmin, PermWorkspaceRead, true},
		{RoleAdmin, PermMemberInvite, true},
		{RoleAdmin, PermProjectDelete, true},
		{RoleAdmin, PermWorkspaceDelete, false},
		{RoleAdmin, PermMemberChangeRole, false},
		// Member: read + create
		{RoleMember, PermWorkspaceRead, true},
		{RoleMember, PermProjectCreate, true},
		{RoleMember, PermIssueCreate, true},
		{RoleMember, PermMemberInvite, false},
		{RoleMember, PermProjectDelete, false},
		// Guest: read only
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
