// Package llm — Agent 执行引擎（工具调用循环）。
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MaxToolIterations 单次对话允许的最大工具调用轮次（防失控循环）。
const MaxToolIterations = 8

// AgentResult Agent 执行结果。
type AgentResult struct {
	// FinalText 面向用户的最终回复
	FinalText string
	// ToolCalls 本次执行中模型发起的工具调用（按顺序）
	ToolCalls []ExecutedTool
	// Usage 累计 token 用量
	Usage Usage
	// Iterations 实际工具循环轮次
	Iterations int
	// PendingActions 需要人工确认的动作（Build 模式）
	PendingActions []PendingAction
}

// ExecutedTool 已执行的工具调用。
type ExecutedTool struct {
	Name   string `json:"name"`
	Args   any    `json:"args,omitempty"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// PendingAction 待确认动作（Build 模式的动作卡片）。
type PendingAction struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

// AgentConfig Agent 运行配置。
type AgentConfig struct {
	// SystemPrompt 系统提示
	SystemPrompt string
	// Messages 对话历史（不含本轮用户输入）
	Messages []Message
	// UserInput 本轮用户输入
	UserInput string
	// Registry 工具注册表
	Registry *ToolRegistry
	// Temperature 温度
	Temperature float64
	// MaxTokens 单次补全上限
	MaxTokens int
	// Mode 模式：ask / build / autopilot
	Mode string
}

// RunAgent 执行一次带工具调用的对话。
//
// 流程：
//  1. 组装 messages（system + history + user）；
//  2. 调用模型；
//  3. 若返回 tool_calls，执行工具并把结果作为 role=tool 消息追加，回到 2；
//  4. 若返回纯文本，作为最终回复返回；
//  5. 超过 MaxToolIterations 仍未收敛时，要求模型用现有信息直接作答。
func (c *Client) RunAgent(ctx context.Context, cfg AgentConfig) (*AgentResult, error) {
	if !c.enabled {
		return nil, ErrNotConfigured
	}
	if cfg.Registry == nil || len(cfg.Registry.Definitions()) == 0 {
		// 无工具时退化为普通补全
		return c.simpleChat(ctx, cfg)
	}

	messages := make([]Message, 0, len(cfg.Messages)+2)
	if cfg.SystemPrompt != "" {
		messages = append(messages, Message{Role: RoleSystem, Content: cfg.SystemPrompt})
	}
	messages = append(messages, cfg.Messages...)
	messages = append(messages, Message{Role: RoleUser, Content: cfg.UserInput})

	result := &AgentResult{}
	tools := cfg.Registry.Definitions()

	for i := 0; i < MaxToolIterations; i++ {
		resp, err := c.Complete(ctx, Request{
			Messages:    messages,
			Tools:       tools,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
		if err != nil {
			return nil, err
		}
		result.Usage.PromptTokens += resp.Usage.PromptTokens
		result.Usage.CompletionTokens += resp.Usage.CompletionTokens
		result.Usage.TotalTokens += resp.Usage.TotalTokens

		// 无工具调用 → 收敛
		if len(resp.ToolCalls) == 0 {
			result.FinalText = resp.Content
			result.Iterations = i
			return result, nil
		}

		// 执行工具调用
		for _, tc := range resp.ToolCalls {
			args := json.RawMessage(tc.Function.Arguments)
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			out, execErr := cfg.Registry.Execute(ctx, tc.Function.Name, args)

			executed := ExecutedTool{Name: tc.Function.Name, Args: rawToAny(args)}
			if execErr != nil {
				executed.Error = execErr.Error()
			} else {
				executed.Result = out
			}
			result.ToolCalls = append(result.ToolCalls, executed)

			// 工具结果作为 role=tool 消息回传
			payload := ""
			if execErr != nil {
				payload = `{"error":` + quote(execErr.Error()) + `}`
			} else {
				b, _ := json.Marshal(out)
				payload = string(b)
			}
			messages = append(messages, Message{
				Role:       RoleTool,
				Content:    payload,
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
			})
		}
		result.Iterations = i + 1
	}

	// 达到上限仍未收敛：要求模型基于现有信息直接作答
	messages = append(messages, Message{
		Role:    RoleSystem,
		Content: "已达到工具调用次数上限。请基于已获取的信息直接给出最终答案，不要再调用工具。",
	})
	resp, err := c.Complete(ctx, Request{
		Messages:    messages,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	result.FinalText = resp.Content
	result.Usage.PromptTokens += resp.Usage.PromptTokens
	result.Usage.CompletionTokens += resp.Usage.CompletionTokens
	result.Usage.TotalTokens += resp.Usage.TotalTokens
	return result, nil
}

// simpleChat 无工具时的普通补全。
func (c *Client) simpleChat(ctx context.Context, cfg AgentConfig) (*AgentResult, error) {
	messages := make([]Message, 0, len(cfg.Messages)+2)
	if cfg.SystemPrompt != "" {
		messages = append(messages, Message{Role: RoleSystem, Content: cfg.SystemPrompt})
	}
	messages = append(messages, cfg.Messages...)
	messages = append(messages, Message{Role: RoleUser, Content: cfg.UserInput})

	resp, err := c.Complete(ctx, Request{
		Messages:    messages,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	return &AgentResult{
		FinalText: resp.Content,
		Usage:     resp.Usage,
	}, nil
}

// --- helpers ---

func rawToAny(raw json.RawMessage) any {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

var _ = strings.TrimSpace
var _ = fmt.Sprintf
