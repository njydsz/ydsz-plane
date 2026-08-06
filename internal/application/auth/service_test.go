package auth

import (
	"testing"
	"time"
)

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
