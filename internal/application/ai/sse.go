// Package ai — SSE 流式端点。
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/llm"
)

// SSE 事件类型。
const (
	// EventDelta 增量文本。
	EventDelta = "delta"
	// EventDone 完成。
	EventDone = "done"
	// EventError 错误。
	EventError = "error"
	// EventMeta 元信息（模型/用量）。
	EventMeta = "meta"
)

// StreamRequest 流式请求。
type StreamRequest struct {
	// 任务类型：assist（续写）/ rewrite（改写）/ grammar（语法）/ chat（对话）
	Task string `json:"task"`
	// 上下文
	Context  string `json:"context"`
	FullText string `json:"full_text"`
	// 改写风格
	Style string `json:"style"`
	// 语言
	Language string `json:"language"`
	// 最大 token
	MaxTokens int `json:"max_tokens"`
	// 对话消息（chat 任务）
	Messages []llm.Message `json:"messages"`
}

// sseWriter SSE 写入辅助。
type sseWriter struct {
	c       *gin.Context
	flusher http.Flusher
	written bool
}

func newSSEWriter(c *gin.Context) *sseWriter {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲
	c.Status(http.StatusOK)
	flusher, _ := c.Writer.(http.Flusher)
	return &sseWriter{c: c, flusher: flusher}
}

func (w *sseWriter) event(event string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	if !w.written {
		w.written = true
	}
	fmt.Fprintf(w.c.Writer, "event: %s\ndata: %s\n\n", event, payload)
	if w.flusher != nil {
		w.flusher.Flush()
	}
}

func (w *sseWriter) delta(text string) {
	w.event(EventDelta, gin.H{"text": text})
}

func (w *sseWriter) done(reason string, usage llm.Usage) {
	w.event(EventDone, gin.H{
		"finish_reason": reason,
		"usage":         usage,
		"at":            time.Now().Unix(),
	})
}

func (w *sseWriter) err(message string) {
	w.event(EventError, gin.H{"message": message})
}

// buildPrompt 根据任务类型构造 prompt。
func (s *Service) buildStreamPrompt(req StreamRequest) (system string, user string, temperature float64) {
	switch req.Task {
	case "assist":
		lang := "中文"
		if req.Language == "en" {
			lang = "English"
		}
		return fmt.Sprintf(
			"你是一位专业的技术写作助手。请基于上下文续写内容，使用%s输出，保持风格一致，不要重复已有内容。",
			lang,
		), req.FullText + "\n\n[需要续写的起点] " + req.Context, 0.7

	case "rewrite":
		styleDesc := map[string]string{
			"formal":  "更正式、规范的书面语",
			"concise": "更精简、去除冗余",
			"fluent":  "更流畅、自然",
			"expand":  "在保持原意基础上适当扩写补充细节",
		}
		desc, ok := styleDesc[req.Style]
		if !ok {
			desc = "优化表达"
		}
		lang := "中文"
		if req.Language == "en" {
			lang = "English"
		}
		return fmt.Sprintf(
			"你是技术写作专家。请将用户提供的文本改写为%s，使用%s输出。只输出改写后的文本，不要附加解释或引号。",
			desc, lang,
		), req.FullText, 0.5

	case "grammar":
		lang := "中文"
		if req.Language == "en" {
			lang = "English"
		}
		return fmt.Sprintf(
			"你是语法校对专家。请修正用户文本中的语法、拼写、标点错误，使用%s输出。只输出修正后的文本，不要附加解释。",
			lang,
		), req.FullText, 0.3

	case "chat":
		return "你是 Ydsz Plane 项目管理平台的 AI 助手。请使用中文简洁清晰地回答。", req.Context, 0.7
	}
	return "你是 AI 助手。", req.Context, 0.7
}

// StreamWriting godoc
//
//	@Summary		流式 AI 写作
//	@Tags			ai
//	@Accept			json
//	@Produce		text/event-stream
//	@Success		200	{string}	string
//	@Router			/ai/stream [post]
func (h *Handler) StreamWriting(c *gin.Context) {
	var req StreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 1024
	}

	// 未配置 LLM 时返回错误事件（保持 SSE 格式，前端统一处理）
	if h.d.AiSvc.llm == nil || !h.d.AiSvc.llm.IsEnabled() {
		w := newSSEWriter(c)
		w.err("AI 模型未配置。请在环境变量中设置 YDSZ_LLM_API_BASE / YDSZ_LLM_API_KEY / YDSZ_LLM_MODEL")
		return
	}

	system, user, temperature := h.d.AiSvc.buildStreamPrompt(req)

	llmReq := llm.Request{
		Messages:    []llm.Message{{Role: llm.RoleSystem, Content: system}, {Role: llm.RoleUser, Content: user}},
		Temperature: temperature,
		MaxTokens:   req.MaxTokens,
	}

	// chat 任务：使用客户端提供的消息
	if req.Task == "chat" && len(req.Messages) > 0 {
		llmReq.Messages = req.Messages
	}

	ctx := c.Request.Context()
	ch, err := h.d.AiSvc.llm.Stream(ctx, llmReq)
	if err != nil {
		w := newSSEWriter(c)
		w.err("AI 调用失败：" + err.Error())
		return
	}

	w := newSSEWriter(c)
	for ev := range ch {
		if ev.Error != nil {
			w.err(ev.Error.Error())
			return
		}
		if ev.Data.Delta != "" {
			w.delta(ev.Data.Delta)
		}
		if ev.Data.Done {
			w.done(ev.Data.FinishReason, ev.Data.Usage)
			return
		}
	}
}

// StreamAgentChat godoc
//
//	@Summary		流式 Agent 对话
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		text/event-stream
//	@Router			/ai/agents/{id}/stream [post]
func (h *Handler) StreamAgentChat(c *gin.Context) {
	// Agent 流式对话复用写作流式端点的 chat 任务
	var req StreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Task = "chat"
	// 复用 StreamWriting 逻辑
	h.StreamWriting(c)
}

// --- 辅助 ---
