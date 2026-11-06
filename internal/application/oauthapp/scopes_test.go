// Package oauthapp — 纯逻辑单元测试。
//
// 覆盖：scope 判定、令牌哈希与识别、scope 解析。
// 不依赖数据库与网络。
package oauthapp

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHasScope(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		scopes   []string
		required string
		want     bool
	}{
		{"nil claims", nil, "read:issue", false},
		{"exact match", []string{"read:issue"}, "read:issue", true},
		{"wildcard", []string{"*"}, "read:issue", true},
		{"wildcard covers write", []string{"*"}, "write:issue", true},
		{"write implies read", []string{"write:issue"}, "read:issue", true},
		{"read does not imply write", []string{"read:issue"}, "write:issue", false},
		{"unrelated scope", []string{"read:issue"}, "read:member", false},
		{"empty scopes", []string{}, "read:issue", false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var claims *TokenClaims
			if c.scopes != nil {
				claims = &TokenClaims{Scopes: c.scopes}
			}
			if got := HasScope(claims, c.required); got != c.want {
				t.Errorf("HasScope(%v, %q) = %v, want %v", c.scopes, c.required, got, c.want)
			}
		})
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	t.Parallel()

	const raw = "raw-token-value"
	want := sha256.Sum256([]byte(raw))
	wantHex := hex.EncodeToString(want[:])

	if got := HashToken(raw); got != wantHex {
		t.Errorf("HashToken(%q) = %q, want %q", raw, got, wantHex)
	}
	// 相同输入结果一致。
	if HashToken(raw) != HashToken(raw) {
		t.Error("HashToken is not deterministic")
	}
	// 不同输入结果不同。
	if HashToken(raw) == HashToken(raw+"x") {
		t.Error("HashToken collision for different inputs")
	}
}

func TestLooksLikeOAuthToken(t *testing.T) {
	t.Parallel()

	if !LooksLikeOAuthToken(TokenPrefix + "abcdef") {
		t.Errorf("token with prefix %q should be recognized", TokenPrefix)
	}
	if LooksLikeOAuthToken("some-other-token") {
		t.Error("token without prefix should not be recognized")
	}
	if LooksLikeOAuthToken("") {
		t.Error("empty string should not be recognized")
	}
}

func TestSplitScopes(t *testing.T) {
	t.Parallel()

	if got := splitScopes(""); got != nil {
		t.Errorf("splitScopes(\"\") = %v, want nil", got)
	}
	if got := splitScopes("   "); got != nil {
		t.Errorf("splitScopes(whitespace) = %v, want nil", got)
	}
	got := splitScopes("read:issue write:issue")
	if len(got) != 2 || got[0] != "read:issue" || got[1] != "write:issue" {
		t.Errorf("splitScopes = %v, want [read:issue write:issue]", got)
	}
}

func TestSubsetScopes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		owned    []string
		required []string
		want     bool
	}{
		{[]string{"read:issue", "write:issue"}, []string{"read:issue"}, true},
		{[]string{"read:issue"}, []string{"read:issue", "write:issue"}, false},
		{[]string{"*"}, []string{"read:issue"}, false}, // 子集判定不处理通配符
		{[]string{}, nil, true},
		{[]string{"a", "b"}, nil, true},
	}
	for _, c := range cases {
		if got := subsetScopes(c.owned, c.required); got != c.want {
			t.Errorf("subsetScopes(%v, %v) = %v, want %v", c.owned, c.required, got, c.want)
		}
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		tok, err := GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken: %v", err)
		}
		if !LooksLikeOAuthToken(tok) {
			t.Errorf("GenerateToken() = %q, missing prefix %q", tok, TokenPrefix)
		}
		if _, dup := seen[tok]; dup {
			t.Fatalf("GenerateToken collision at iteration %d", i)
		}
		seen[tok] = struct{}{}
	}
}
