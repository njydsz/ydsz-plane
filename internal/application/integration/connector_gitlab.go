// Package integration — GitLab 连接器。
//
// 能力：
//   - ValidateCredentials：校验 token + base_url；
//   - TestConnection：GET /user 验证 token；
//   - SyncWorkItems：当前版本返回 not-implemented（校验与测试已就绪）。
package integration

import (
	"context"
	"fmt"
	"strings"
)

// GitLabConnector 是 GitLab 的 IntegrationConnector 实现。
type GitLabConnector struct{}

// NewGitLabConnector 构造 GitLab 连接器。
func NewGitLabConnector() *GitLabConnector { return &GitLabConnector{} }

// Provider 返回 provider 标识。
func (c *GitLabConnector) Provider() string { return string(ProviderGitLab) }

// baseURL 返回 GitLab 实例的 API base（默认 gitlab.com）。
func (c *GitLabConnector) baseURL(creds map[string]string) string {
	if v := strings.TrimSpace(creds["base_url"]); v != "" {
		return strings.TrimRight(v, "/") + "/api/v4"
	}
	return "https://gitlab.com/api/v4"
}

// ValidateCredentials 校验凭证字段完整性。
func (c *GitLabConnector) ValidateCredentials(creds map[string]string) error {
	return requireFields(c.Provider(), creds, "token")
}

// TestConnection 通过 GET /user 验证 token 有效性。
func (c *GitLabConnector) TestConnection(creds map[string]string) error {
	if err := c.ValidateCredentials(creds); err != nil {
		return err
	}
	var out struct {
		Username string `json:"username"`
	}
	if err := doJSON(context.Background(), "GET",
		buildURL(c.baseURL(creds), "/user", nil),
		map[string]string{"PRIVATE-TOKEN": creds["token"]}, nil, &out); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return fmt.Errorf("%s: test connection failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: test connection failed: %w", c.Provider(), err)
	}
	if out.Username == "" {
		return fmt.Errorf("%s: test connection returned empty username", c.Provider())
	}
	return nil
}

// SyncWorkItems 暂未实现（凭证校验与连通性测试已就绪）。
func (c *GitLabConnector) SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error) {
	return SyncResult{}, fmt.Errorf("%s: SyncWorkItems not yet implemented", c.Provider())
}
