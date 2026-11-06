// Package integration — Slack 连接器。
//
// 能力：
//   - ValidateCredentials：校验 bot_token；
//   - TestConnection：POST auth.test 验证 bot token；
//   - Notify：向指定 channel 发送消息（供出站通知使用）；
//   - SyncWorkItems：当前版本返回 not-implemented（Slack 为通知型集成）。
package integration

import (
	"context"
	"fmt"
	"strings"
)

// SlackConnector 是 Slack 的 IntegrationConnector 实现。
type SlackConnector struct{}

// NewSlackConnector 构造 Slack 连接器。
func NewSlackConnector() *SlackConnector { return &SlackConnector{} }

// Provider 返回 provider 标识。
func (c *SlackConnector) Provider() string { return string(ProviderSlack) }

// ValidateCredentials 校验凭证字段完整性。
func (c *SlackConnector) ValidateCredentials(creds map[string]string) error {
	return requireFields(c.Provider(), creds, "bot_token")
}

// TestConnection 通过 POST /api/auth.test 验证 bot token。
func (c *SlackConnector) TestConnection(creds map[string]string) error {
	if err := c.ValidateCredentials(creds); err != nil {
		return err
	}
	var out struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := doJSON(context.Background(), "POST",
		"https://slack.com/api/auth.test",
		map[string]string{"Authorization": "Bearer " + creds["bot_token"]},
		map[string]any{}, &out); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return fmt.Errorf("%s: test connection failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: test connection failed: %w", c.Provider(), err)
	}
	if !out.OK {
		return fmt.Errorf("%s: test connection failed: %s", c.Provider(), out.Error)
	}
	return nil
}

// Notify 向指定 channel 发送一条文本消息。
//
// config 约定：channel（必填，如 "#general" 或 channel ID）。
func (c *SlackConnector) Notify(creds map[string]string, config map[string]string, text string) error {
	channel := strings.TrimSpace(config["channel"])
	if channel == "" {
		return fmt.Errorf("%s: config field %q is required", c.Provider(), "channel")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("%s: message text is required", c.Provider())
	}
	var out struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	body := map[string]any{"channel": channel, "text": text}
	if err := doJSON(context.Background(), "POST",
		"https://slack.com/api/chat.postMessage",
		map[string]string{"Authorization": "Bearer " + creds["bot_token"]},
		body, &out); err != nil {
		if apiErr, ok := err.(*apiError); ok {
			return fmt.Errorf("%s: notify failed: %s", c.Provider(), apiErr.Message)
		}
		return fmt.Errorf("%s: notify failed: %w", c.Provider(), err)
	}
	if !out.OK {
		return fmt.Errorf("%s: notify failed: %s", c.Provider(), out.Error)
	}
	return nil
}

// SyncWorkItems 暂未实现（Slack 为通知型集成，无工作项同步语义）。
func (c *SlackConnector) SyncWorkItems(ctx context.Context, creds map[string]string, config map[string]string) (SyncResult, error) {
	return SyncResult{}, fmt.Errorf("%s: SyncWorkItems not yet implemented", c.Provider())
}
