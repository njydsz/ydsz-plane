// Package llm — 流式补全（SSE）。
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StreamChunk 流式响应片段。
type StreamChunk struct {
	// 增量文本
	Delta string `json:"delta"`
	// 是否为结束片段
	Done bool `json:"done"`
	// 完成原因（stop / length / tool_calls / error）
	FinishReason string `json:"finish_reason,omitempty"`
	// 工具调用（若模型返回）
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// 累计用量
	Usage Usage `json:"usage,omitempty"`
}

// StreamEvent SSE 事件类型。
type StreamEvent struct {
	Data  StreamChunk
	Error error
}

// Stream 执行流式补全。
//
// 返回一个只读 channel：调用方通过 range 消费增量片段；
// channel 关闭表示流结束（无论成功或失败）。
//
// 用法：
//
//	ch, err := client.Stream(ctx, req)
//	if err != nil { return err }
//	for ev := range ch {
//	    if ev.Error != nil { /* 处理错误 */ }
//	    fmt.Print(ev.Data.Delta)
//	}
func (c *Client) Stream(ctx context.Context, req Request) (<-chan StreamEvent, error) {
	if !c.enabled {
		return nil, ErrNotConfigured
	}

	providers := []Provider{c.primary}
	if c.fallback.APIKey != "" && c.fallback.Model != "" {
		providers = append(providers, c.fallback)
	}

	var lastErr error
	for _, p := range providers {
		ch, err := c.streamOnce(ctx, p, req)
		if err == nil {
			return ch, nil
		}
		lastErr = err
		if isNonRetryable(err) {
			break
		}
	}
	if lastErr == nil {
		lastErr = ErrNotConfigured
	}
	return nil, lastErr
}

// streamOnce 单次流式请求。
func (c *Client) streamOnce(ctx context.Context, p Provider, req Request) (<-chan StreamEvent, error) {
	body := chatRequest{
		Model:      p.Model,
		Messages:   req.Messages,
		Tools:      req.Tools,
		ToolChoice: req.ToolChoice,
		MaxTokens:  req.MaxTokens,
		Stream:     true,
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
	httpReq.Header.Set("Accept", "text/event-stream")

	httpClient := c.http
	if p.Timeout > 0 && p.Timeout != httpClient.Timeout {
		// 流式请求不使用短超时（由 ctx 控制）
		httpClient = &http.Client{Timeout: 5 * timeMinute}
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
			Provider:   p.Name,
		}
	}

	ch := make(chan StreamEvent, 16)
	go parseSSEStream(resp.Body, ch)
	return ch, nil
}

// timeMinute 辅助常量。
const timeMinute = 60 // 秒

// parseSSEStream 解析 OpenAI 兼容的 SSE 流。
func parseSSEStream(r io.ReadCloser, ch chan<- StreamEvent) {
	defer close(ch)
	defer r.Close()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var usage Usage
	var accumulated strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

		// 结束标记
		if payload == "[DONE]" {
			ch <- StreamEvent{Data: StreamChunk{
				Delta:        "",
				Done:         true,
				FinishReason: "stop",
				Usage:        usage,
			}}
			return
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		if chunk.Error != nil && chunk.Error.Message != "" {
			ch <- StreamEvent{Error: fmt.Errorf("llm: %s", chunk.Error.Message)}
			return
		}

		if len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]
		delta := choice.Delta.Content
		if delta != "" {
			accumulated.WriteString(delta)
			ch <- StreamEvent{Data: StreamChunk{
				Delta:        delta,
				ToolCalls:    choice.Delta.ToolCalls,
				FinishReason: choice.FinishReason,
			}}
		}
		if choice.FinishReason != "" {
			ch <- StreamEvent{Data: StreamChunk{
				Done:         true,
				FinishReason: choice.FinishReason,
			}}
			return
		}
		if chunk.Usage != nil {
			usage = *chunk.Usage
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- StreamEvent{Error: err}
	}
}

// streamChunk OpenAI 流式响应结构。
type streamChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
