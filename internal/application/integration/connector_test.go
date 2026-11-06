// Package integration — 连接器纯逻辑单元测试。
//
// 覆盖：凭据字段校验、URL 拼接。
// 不依赖网络（ValidateCredentials 仅做字段完整性检查，不发起请求）。
package integration

import (
	"strings"
	"testing"
)

func TestGitHubConnector_ValidateCredentials(t *testing.T) {
	t.Parallel()

	c := &GitHubConnector{}

	if err := c.ValidateCredentials(map[string]string{"token": "ghp_xxx"}); err != nil {
		t.Errorf("valid token should pass: %v", err)
	}
	if err := c.ValidateCredentials(map[string]string{}); err == nil {
		t.Error("missing token should fail")
	}
	if err := c.ValidateCredentials(map[string]string{"token": "   "}); err == nil {
		t.Error("whitespace-only token should fail")
	}
}

func TestGitLabConnector_ValidateCredentials(t *testing.T) {
	t.Parallel()

	c := &GitLabConnector{}

	if err := c.ValidateCredentials(map[string]string{"token": "glpat-xxx"}); err != nil {
		t.Errorf("valid token should pass: %v", err)
	}
	if err := c.ValidateCredentials(map[string]string{"token": ""}); err == nil {
		t.Error("empty token should fail")
	}
}

func TestJiraConnector_ValidateCredentials(t *testing.T) {
	t.Parallel()

	c := &JiraConnector{}

	valid := map[string]string{"token": "jira-token", "base_url": "https://example.atlassian.net"}
	if err := c.ValidateCredentials(valid); err != nil {
		t.Errorf("valid creds should pass: %v", err)
	}

	// 缺 base_url 应失败。
	if err := c.ValidateCredentials(map[string]string{"token": "t"}); err == nil {
		t.Error("missing base_url should fail")
	}
	// 缺 token 应失败。
	if err := c.ValidateCredentials(map[string]string{"base_url": "https://x"}); err == nil {
		t.Error("missing token should fail")
	}
}

func TestRequireFields_ErrorMentionsProviderAndField(t *testing.T) {
	t.Parallel()

	err := requireFields("github", map[string]string{}, "token", "secret")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "github") {
		t.Errorf("error should mention provider, got %q", msg)
	}
	if !strings.Contains(msg, "token") {
		t.Errorf("error should mention missing field, got %q", msg)
	}
}

func TestBuildURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		base  string
		path  string
		query map[string]string
		want  string
	}{
		{
			name: "simple join",
			base: "https://api.example.com",
			path: "repos/1",
			want: "https://api.example.com/repos/1",
		},
		{
			name: "trailing slash on base",
			base: "https://api.example.com/",
			path: "repos/1",
			want: "https://api.example.com/repos/1",
		},
		{
			name: "leading slash on path",
			base: "https://api.example.com",
			path: "/repos/1",
			want: "https://api.example.com/repos/1",
		},
		{
			name:  "with query",
			base:  "https://api.example.com",
			path:  "search",
			query: map[string]string{"q": "bug", "limit": "10"},
			want:  "https://api.example.com/search?limit=10&q=bug",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			q := map[string][]string{}
			for k, v := range c.query {
				q[k] = []string{v}
			}
			got := buildURL(c.base, c.path, q)
			if got != c.want {
				t.Errorf("buildURL(%q, %q) = %q, want %q", c.base, c.path, got, c.want)
			}
		})
	}
}
