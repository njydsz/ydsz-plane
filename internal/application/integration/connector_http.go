// Package integration — 连接器共享 HTTP 工具。
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpClient 是连接器使用的共享 HTTP 客户端。
//
// 超时设为 15s：既能覆盖跨洋 API 往返，又避免凭证测试在第三方故障时挂起。
var httpClient = &http.Client{Timeout: 15 * time.Second}

// apiError 是第三方 API 返回错误的归一化表示。
type apiError struct {
	// Provider 为产生错误的 provider。
	Provider string
	// StatusCode 为 HTTP 状态码。
	StatusCode int
	// Message 为错误摘要。
	Message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s: %s (HTTP %d)", e.Provider, e.Message, e.StatusCode)
}

// newAPIError 构造归一化的第三方 API 错误。
func newAPIError(provider string, status int, body string) *apiError {
	msg := strings.TrimSpace(body)
	if len(msg) > 500 {
		msg = msg[:500]
	}
	if msg == "" {
		msg = http.StatusText(status)
	}
	return &apiError{Provider: provider, StatusCode: status, Message: msg}
}

// doJSON 发起一次 HTTP 请求，将 body 序列化为 JSON，并把响应解析到 out。
//
// 请求头由 headers 提供（如 Authorization）。当响应状态码非 2xx 时，
// 返回 *apiError；调用方据此区分凭证无效（401/403）与其它故障。
func doJSON(ctx context.Context, method, rawURL string, headers map[string]string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = strings.NewReader(string(buf))
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB 上限，防响应炸弹
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newAPIError("", resp.StatusCode, string(respBody))
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshal response body: %w", err)
		}
	}
	return nil
}

// requireFields 校验凭证 map 中必填字段是否齐全。
// 返回第一个缺失字段的错误。
func requireFields(provider string, creds map[string]string, fields ...string) error {
	for _, f := range fields {
		if strings.TrimSpace(creds[f]) == "" {
			return fmt.Errorf("%s: missing required credential field %q", provider, f)
		}
	}
	return nil
}

// buildURL 将 base 与 path 拼接，并追加 query 参数。
func buildURL(base, path string, query url.Values) string {
	u := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

// nowUTC 返回当前 UTC 时间（供各连接器填充 SyncResult.FetchedAt）。
func nowUTC() time.Time { return time.Now().UTC() }

// parseRFC3339 解析 RFC3339 时间字符串；失败时返回零值 time.Time。
func parseRFC3339(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
