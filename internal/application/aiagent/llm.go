// Package aiagent — LLM 集成与工具注册。
package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/llm"
)

// ToolProvider 工作空间工具执行器。
//
// 由应用层注入，负责把 LLM 的工具调用落到实际领域服务
// （工作项/迭代/模块/成员等），避免 aiagent 包直接依赖 issue 包。
type ToolProvider interface {
	// ExecuteTool 执行工具调用，返回 JSON 可序列化结果。
	ExecuteTool(ctx context.Context, workspaceID, userID int64, name string, args json.RawMessage) (any, error)
}

// LLMBridge 连接 aiagent 与 llm 网关的桥接层。
type LLMBridge struct {
	client *llm.Client
	db     *pgxpool.Pool
	tools  *llm.ToolRegistry
}

// NewLLMBridge 构造桥接层。
func NewLLMBridge(client *llm.Client, db *pgxpool.Pool, provider ToolProvider) *LLMBridge {
	b := &LLMBridge{client: client, db: db, tools: llm.NewToolRegistry()}

	defs := llm.WorkspaceTools()
	for _, def := range defs {
		d := def
		b.tools.Register(d, func(ctx context.Context, args json.RawMessage) (any, error) {
			return provider.ExecuteTool(ctx, workspaceIDFromContext(ctx), userIDFromContext(ctx), d.Name, args)
		})
	}
	return b
}

// IsEnabled 返回 LLM 是否可用。
func (b *LLMBridge) IsEnabled() bool { return b != nil && b.client.IsEnabled() }

// Status 返回 LLM 状态。
func (b *LLMBridge) Status() map[string]any {
	if b == nil {
		return map[string]any{"enabled": false}
	}
	return b.client.Status()
}

// ToolNames 返回已注册工具名。
func (b *LLMBridge) ToolNames() []string {
	defs := b.tools.Definitions()
	out := make([]string, 0, len(defs))
	for _, d := range defs {
		// llm.Tool 的 Name 嵌套在 Function 中（OpenAI function calling 格式）。
		out = append(out, d.Function.Name)
	}
	return out
}

// --- 模式对应的系统提示 ---

// SystemPromptFor 按模式与 Agent 配置生成系统提示。
func SystemPromptFor(agent *Agent, mode Mode, workspaceName string) string {
	var sb strings.Builder
	sb.WriteString("你是 Ydsz Plane 项目管理平台的 AI 助手。\n")
	if workspaceName != "" {
		sb.WriteString("当前工作空间：")
		sb.WriteString(workspaceName)
		sb.WriteString("\n")
	}
	if agent != nil && agent.Instructions != "" {
		sb.WriteString("\n【助手设定】\n")
		sb.WriteString(agent.Instructions)
		sb.WriteString("\n")
	}
	if agent != nil && agent.Name != "" {
		sb.WriteString("\n助手名称：")
		sb.WriteString(agent.Name)
		sb.WriteString("\n")
	}

	switch mode {
	case ModeAsk:
		sb.WriteString(`
【模式：Ask（只读问答）】
- 你可以查询工作项、迭代、模块、成员等信息来回答问题；
- 禁止执行任何写操作（创建/更新/删除/流转/指派）；
- 如需变更，请明确告知用户需要手动操作，或建议切换到 Build/Autopilot 模式。
`)
	case ModeBuild:
		sb.WriteString(`
【模式：Build（计划构建）】
- 你先生成一份行动计划（列出要做的变更）；
- 每一项变更对应一个动作，动作必须通过工具 create_work_item / update_work_item 等发起；
- 你发起的动作会进入「待确认」状态，由用户批准后才真正生效；
- 不要假设动作已执行，除非工具返回明确的成功结果。
`)
	case ModeAutopilot:
		sb.WriteString(`
【模式：Autopilot（自动执行）】
- 你可以直接执行写操作（创建/更新/删除/流转/指派/评论）；
- 执行前请确保参数正确；对不可逆操作（删除）务必谨慎，必要时先向用户确认；
- 执行后请汇报每个动作的结果。
`)
	}

	sb.WriteString(`
【输出要求】
- 使用中文回答，简洁清晰；
- 涉及工作项时给出 ID 与标题；
- 不确定时明确说明，不要编造数据。
`)
	return sb.String()
}

// Chat 执行一轮带工具的对话。
func (b *LLMBridge) Chat(ctx context.Context, agent *Agent, mode Mode, workspaceName string, history []llm.Message, userInput string, temperature float64) (*llm.AgentResult, error) {
	if !b.IsEnabled() {
		return nil, llm.ErrNotConfigured
	}
	if temperature <= 0 {
		if agent != nil && agent.Temperature > 0 {
			temperature = agent.Temperature
		} else {
			temperature = 0.7
		}
	}
	return b.client.RunAgent(ctx, llm.AgentConfig{
		SystemPrompt: SystemPromptFor(agent, mode, workspaceName),
		Messages:     history,
		UserInput:    userInput,
		Registry:     b.tools,
		Temperature:  temperature,
		MaxTokens:    4096,
		Mode:         string(mode),
	})
}

// --- context helpers ---

// contextKey 私有 context key 类型，避免冲突。
type contextKey string

const (
	ctxWorkspaceID contextKey = "aiagent_workspace_id"
	ctxUserID      contextKey = "aiagent_user_id"
)

// WithActor 将工作空间/用户信息写入 context。
func WithActor(ctx context.Context, workspaceID, userID int64) context.Context {
	ctx = context.WithValue(ctx, ctxWorkspaceID, workspaceID)
	return context.WithValue(ctx, ctxUserID, userID)
}

func workspaceIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctxWorkspaceID).(int64); ok {
		return v
	}
	return 0
}

func userIDFromContext(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctxUserID).(int64); ok {
		return v
	}
	return 0
}

var _ = fmt.Sprintf
