// Package llm — 大语言模型网关。
//
// 目标：为全平台提供统一的 LLM 访问能力，屏蔽不同 provider 的差异。
//
// 能力：
//   - OpenAI Chat Completions 兼容协议（覆盖 OpenAI / DeepSeek / 通义 / 智谱 / 月之暗面 /
//     Ollama / vLLM / 以及任何 OpenAI 兼容网关）；
//   - Function Calling（工具调用）；
//   - 主/备 provider 故障自动切换 + 重试；
//   - 请求级超时；
//   - 用量统计（token 数）供 AI 额度扣减使用。
//
// 配置（环境变量）：
//
//	YDSZ_LLM_PROVIDER=deepseek          # 默认 provider 名
//	YDSZ_LLM_API_BASE=https://api.deepseek.com
//	YDSZ_LLM_API_KEY=sk-xxx
//	YDSZ_LLM_MODEL=deepseek-chat
//	YDSZ_LLM_FALLBACK_PROVIDER=
//	YDSZ_LLM_FALLBACK_API_BASE=
//	YDSZ_LLM_FALLBACK_API_KEY=
//	YDSZ_LLM_FALLBACK_MODEL=
//	YDSZ_LLM_TIMEOUT=30s
//	YDSZ_LLM_MAX_RETRIES=2
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Provider 单个 LLM provider 配置。
type Provider struct {
	Name     string
	APIBase  string
	APIKey   string
	Model    string
	Timeout  time.Duration
	MaxRetry int
}

// LoadFromEnv 从环境变量加载主/备 provider 配置。
func LoadFromEnv() (primary Provider, fallback Provider, ok bool) {
	timeout := 30 * time.Second
	if v := os.Getenv("YDSZ_LLM_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			timeout = d
		}
	}
	maxRetry := 2
	if v := os.Getenv("YDSZ_LLM_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxRetry = n
		}
	}

	newProvider := func(prefix string) Provider {
		p := Provider{
			Name:     os.Getenv(prefix + "PROVIDER"),
			APIBase:  strings.TrimRight(os.Getenv(prefix+"API_BASE"), "/"),
			APIKey:   os.Getenv(prefix + "API_KEY"),
			Model:    os.Getenv(prefix + "MODEL"),
			Timeout:  timeout,
			MaxRetry: maxRetry,
		}
		if p.APIBase == "" {
			p.APIBase = "https://api.openai.com/v1"
		}
		return p
	}

	primary = newProvider("YDSZ_LLM_")
	fallback = newProvider("YDSZ_LLM_FALLBACK_")

	if primary.APIKey != "" && primary.Model != "" {
		if primary.Name == "" {
			primary.Name = "primary"
		}
		ok = true
	}
	return primary, fallback, ok
}

// Client LLM 客户端。
type Client struct {
	primary  Provider
	fallback Provider
	enabled  bool
	http     *http.Client
}

// NewClient 构造客户端。
func NewClient(primary, fallback Provider, enabled bool) *Client {
	return &Client{
		primary:  primary,
		fallback: fallback,
		enabled:  enabled,
		http:     &http.Client{Timeout: primary.Timeout},
	}
}

// IsEnabled 返回是否已配置可用 provider。
func (c *Client) IsEnabled() bool { return c != nil && c.enabled }

// Status 返回配置状态（供 /ai/status 展示）。
func (c *Client) Status() map[string]any {
	if c == nil {
		return map[string]any{"enabled": false}
	}
	return map[string]any{
		"enabled":  c.enabled,
		"primary":  providerStatus(c.primary),
		"fallback": providerStatus(c.fallback),
	}
}

func providerStatus(p Provider) map[string]any {
	return map[string]any{
		"name":       p.Name,
		"api_base":   p.APIBase,
		"model":      p.Model,
		"configured": p.APIKey != "" && p.Model != "",
	}
}

// --- 消息与工具 ---

// Role 消息角色。
type Role string

const (
	// RoleSystem 系统提示。
	RoleSystem Role = "system"
	// RoleUser 用户输入。
	RoleUser Role = "user"
	// RoleAssistant 模型回复。
	RoleAssistant Role = "assistant"
	// RoleTool 工具结果。
	RoleTool Role = "tool"
)

// Message 对话消息。
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
	// ToolCallID 工具调用标识（role=tool 时使用）。
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// Tool 工具定义（OpenAI function calling 格式）。
type Tool struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef 函数定义。
type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolDefinition 工具定义（FunctionDef 的别名，语义更清晰）。
type ToolDefinition = FunctionDef

// ToolCall 模型发起的工具调用。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用详情。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Usage token 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// --- 请求 / 响应 ---

// Request 补全请求。
type Request struct {
	Messages    []Message
	Tools       []Tool
	ToolChoice  string
	Temperature float64
	MaxTokens   int
}

// Response 补全响应。
type Response struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
	Model     string
	Provider  string
}

// --- OpenAI 兼容协议内部结构 ---

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_choice,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// ErrNotConfigured 未配置可用 provider。
var ErrNotConfigured = fmt.Errorf("llm: no provider configured")

// Complete 执行非流式补全（含主/备 provider 自动切换与重试）。
func (c *Client) Complete(ctx context.Context, req Request) (*Response, error) {
	if !c.enabled {
		return nil, ErrNotConfigured
	}

	var lastErr error
	providers := []Provider{c.primary}
	if c.fallback.APIKey != "" && c.fallback.Model != "" {
		providers = append(providers, c.fallback)
	}

	for _, p := range providers {
		for attempt := 0; attempt <= p.MaxRetry; attempt++ {
			resp, err := c.completeOnce(ctx, p, req)
			if err == nil {
				return resp, nil
			}
			lastErr = err
			if isNonRetryable(err) {
				break
			}
			if attempt < p.MaxRetry {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(attempt+1) * 500 * time.Millisecond):
				}
			}
		}
	}
	if lastErr == nil {
		lastErr = ErrNotConfigured
	}
	return nil, lastErr
}

// completeOnce 单次补全请求。
func (c *Client) completeOnce(ctx context.Context, p Provider, req Request) (*Response, error) {
	body := chatRequest{
		Model:      p.Model,
		Messages:   req.Messages,
		Tools:      req.Tools,
		ToolChoice: req.ToolChoice,
		MaxTokens:  req.MaxTokens,
	}
	if req.Temperature > 0 {
		t := req.Temperature
		body.Temperature = &t
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := p.APIBase + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	httpClient := c.http
	if p.Timeout > 0 && p.Timeout != httpClient.Timeout {
		httpClient = &http.Client{Timeout: p.Timeout}
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
			Provider:   p.Name,
		}
	}

	var parsed chatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("llm: provider %s returned error: %s", p.Name, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("llm: provider %s returned no choices", p.Name)
	}

	msg := parsed.Choices[0].Message
	out := &Response{
		Content:   msg.Content,
		ToolCalls: msg.ToolCalls,
		Model:     parsed.Model,
		Provider:  p.Name,
	}
	if parsed.Usage != nil {
		out.Usage = *parsed.Usage
	}
	return out, nil
}

// APIError provider 返回的 API 错误。
type APIError struct {
	StatusCode int
	Body       string
	Provider   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("llm: provider %s returned HTTP %d: %s", e.Provider, e.StatusCode, e.Body)
}

// isNonRetryable 判断错误是否不可重试（4xx 客户端错误，429 除外）。
func isNonRetryable(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429
	}
	return false
}
