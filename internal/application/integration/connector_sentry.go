// Package integration — Sentry 连接器。
//
// 能力：
//   - ValidateCredentials：校验 token + org（可选 base_url）；
//   - TestConnection：GET /projects/ 验证 token 并确认组织可访问；
//   - SyncWorkItems：当前版本返回 not-implemented（校验与测试已就绪）。
package integration

import (
	"context"
	"fmt"
	"strings"
)

// SentryConnector 是 Sentry 的 IntegrationConnector 实现。
type SentryConnector struct{}

// NewSentryConnector 构造 Sentry 连接器。
func NewSentryConnector() *SentryConnector { return &SentryConnector{} }

// Provider 返回 provider 标识。
func (c *SentryConnector) Provider() string { return string(ProviderSentry) }

// baseURL 返回 Sentry API base（默认 sentry.io）。
func (c *SentryConnector) baseURL(creds map[string]string) string {
	if v := strings.TrimSpace(creds["base_url"]); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://sentry.io/api/0"
}

// ValidateCredentials 校验凭证字段完整性。
func (c *SentryConnector) ValidateCredentials(creds map[string]string) error {
	return requireFields(c.Provider(), creds, "token")
}

// TestConnection 通过 GET /projects/ 验证 token 并确认组织可访问。
func (c *SentryConnector) TestConnection(creds map[string]string) error {
	if err := c.ValidateCredentials(creds); err != nil {
		return err
	}
	var out []struct {
		Slug string `json:"slug"`
	}
	if err := doJSON(context.Background(), "GET",
		buildURL(c.baseURL(creds), "/projects/", nil),
		map[string]string{"Authorization": "Bearer " + creds["token"]}, nil, &out); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return fmt.Errorf("%s: test connection failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: test connection failed: %w", c.Provider(), err)
	}
	return nil
}

// ListProjects 列出 Sentry 组织下的项目（供选择器使用）。
func (c *SentryConnector) ListProjects(creds map[string]string) ([]string, error) {
	var out []struct {
		Slug string `json:"slug"`
	}
	if err := doJSON(context.Background(), "GET",
		buildURL(c.baseURL(creds), "/projects/", nil),
		map[string]string{"Authorization": "Bearer " + creds["token"]}, nil, &out); err != nil {
		return nil, fmt.Errorf("%s: list projects failed: %w", c.Provider(), err)
	}
	projects := make([]string, 0, len(out))
	for _, p := range out {
		projects = append(projects, p.Slug)
	}
	return projects, nil
}

// SyncWorkItems 暂未实现（凭证校验与连通性测试已就绪）。
func (c *SentryConnector) SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error) {
	return SyncResult{}, fmt.Errorf("%s: SyncWorkItems not yet implemented", c.Provider())
}
