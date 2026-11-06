// Package integration — Jira 连接器（最小实现）。
//
// 能力：
//   - ValidateCredentials：校验 token + base_url；
//   - TestConnection：GET /rest/api/3/myself 验证 token；
//   - SyncWorkItems：当前版本返回 not-implemented。
//
// 说明：Jira 为 Provider 枚举中的预留项，完整同步能力在后续迭代补齐。
package integration

import (
	"context"
	"fmt"
	"strings"
)

// JiraConnector 是 Jira 的 IntegrationConnector 实现。
type JiraConnector struct{}

// NewJiraConnector 构造 Jira 连接器。
func NewJiraConnector() *JiraConnector { return &JiraConnector{} }

// Provider 返回 provider 标识。
func (c *JiraConnector) Provider() string { return string(ProviderJira) }

// ValidateCredentials 校验凭证字段完整性。
func (c *JiraConnector) ValidateCredentials(creds map[string]string) error {
	return requireFields(c.Provider(), creds, "token", "base_url")
}

// TestConnection 通过 GET /rest/api/3/myself 验证 token。
func (c *JiraConnector) TestConnection(creds map[string]string) error {
	if err := c.ValidateCredentials(creds); err != nil {
		return err
	}
	var out struct {
		AccountID string `json:"accountId"`
	}
	if err := doJSON(context.Background(), "GET",
		buildURL(strings.TrimRight(creds["base_url"], "/"), "/rest/api/3/myself", nil),
		map[string]string{"Authorization": "Bearer " + creds["token"]}, nil, &out); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return fmt.Errorf("%s: test connection failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: test connection failed: %w", c.Provider(), err)
	}
	return nil
}

// SyncWorkItems 暂未实现。
func (c *JiraConnector) SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error) {
	return SyncResult{}, fmt.Errorf("%s: SyncWorkItems not yet implemented", c.Provider())
}
