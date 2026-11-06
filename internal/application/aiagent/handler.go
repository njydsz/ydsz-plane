// Package aiagent — HTTP handlers。
package aiagent

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/llm"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *Service
	// LLM 可选；未配置时 Chat 退化为仅存储用户消息并返回提示。
	LLM *LLMBridge
	// RunSvc 可选；提供 Run 生命周期查询。
	RunSvc *RunService
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册路由（工作空间级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	// Agent 管理
	agents := r.Group("/ai-agents")
	agents.GET("", h.ListAgents)
	agents.POST("", h.CreateAgent)
	agents.GET("/:agent_id", h.GetAgent)
	agents.PATCH("/:agent_id", h.UpdateAgent)
	agents.DELETE("/:agent_id", h.DeleteAgent)

	// 记忆
	mem := r.Group("/ai-memory")
	mem.GET("", h.ListMemory)
	mem.POST("", h.SetMemory)
	mem.DELETE("/:memory_id", h.DeleteMemory)

	// 会话
	conv := r.Group("/ai/conversations")
	conv.GET("", h.ListConversations)
	conv.POST("", h.CreateConversation)
	conv.GET("/:conversation_id", h.GetConversation)
	conv.POST("/:conversation_id/messages", h.Chat)

	// 动作
	act := r.Group("/ai/actions")
	act.GET("", h.ListPendingActions)
	act.POST("/:action_id/approve", h.ApproveAction)
	act.POST("/:action_id/reject", h.RejectAction)
	act.POST("/:action_id/execute", h.ExecuteAction)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrAgentNotFound, ErrConversationNotFound, ErrActionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrActionNotPending:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrCreditExhausted:
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	default:
		if ec, ok := err.(interface{ HTTPStatus() int }); ok {
			c.JSON(ec.HTTPStatus(), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// ListAgents godoc
//
//	@Summary		AI Agent 列表
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai-agents [get]
func (h *Handler) ListAgents(c *gin.Context) {
	items, err := h.d.Svc.ListAgents(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateAgent godoc
//
//	@Summary		创建 AI Agent
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Agent
//	@Router			/ai-agents [post]
func (h *Handler) CreateAgent(c *gin.Context) {
	var body struct {
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		Instructions       string   `json:"instructions"`
		Triggers           []string `json:"triggers"`
		Skills             []string `json:"skills"`
		ScopeProjectIDs    []int64  `json:"scope_project_ids"`
		ScopeWorkItemTypes []string `json:"scope_work_item_types"`
		Model              string   `json:"model"`
		Temperature        float64  `json:"temperature"`
		DailyCreditLimit   int      `json:"daily_credit_limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent, err := h.d.Svc.CreateAgent(c.Request.Context(), AgentInput{
		WorkspaceID:        c.GetInt64(middleware.CtxWorkspaceID),
		Name:               body.Name,
		Description:        body.Description,
		Instructions:       body.Instructions,
		Triggers:           body.Triggers,
		Skills:             body.Skills,
		ScopeProjectIDs:    body.ScopeProjectIDs,
		ScopeWorkItemTypes: body.ScopeWorkItemTypes,
		Model:              body.Model,
		Temperature:        body.Temperature,
		DailyCreditLimit:   body.DailyCreditLimit,
		CreatedBy:          c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, agent)
}

// GetAgent godoc
//
//	@Summary		AI Agent 详情
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	Agent
//	@Router			/ai-agents/{agent_id} [get]
func (h *Handler) GetAgent(c *gin.Context) {
	agent, err := h.d.Svc.GetAgent(c.Request.Context(), paramInt(c, "agent_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, agent)
}

// UpdateAgent godoc
//
//	@Summary		更新 AI Agent
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Agent
//	@Router			/ai-agents/{agent_id} [patch]
func (h *Handler) UpdateAgent(c *gin.Context) {
	var body struct {
		Name               string   `json:"name"`
		Description        string   `json:"description"`
		Instructions       string   `json:"instructions"`
		Triggers           []string `json:"triggers"`
		Skills             []string `json:"skills"`
		ScopeProjectIDs    []int64  `json:"scope_project_ids"`
		ScopeWorkItemTypes []string `json:"scope_work_item_types"`
		Model              string   `json:"model"`
		Temperature        float64  `json:"temperature"`
		DailyCreditLimit   int      `json:"daily_credit_limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent, err := h.d.Svc.UpdateAgent(c.Request.Context(), paramInt(c, "agent_id"), AgentInput{
		WorkspaceID:        c.GetInt64(middleware.CtxWorkspaceID),
		Name:               body.Name,
		Description:        body.Description,
		Instructions:       body.Instructions,
		Triggers:           body.Triggers,
		Skills:             body.Skills,
		ScopeProjectIDs:    body.ScopeProjectIDs,
		ScopeWorkItemTypes: body.ScopeWorkItemTypes,
		Model:              body.Model,
		Temperature:        body.Temperature,
		DailyCreditLimit:   body.DailyCreditLimit,
		CreatedBy:          c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, agent)
}

// DeleteAgent godoc
//
//	@Summary		删除 AI Agent
//	@Tags			ai-agent
//	@Produce		json
//	@Success		204
//	@Router			/ai-agents/{agent_id} [delete]
func (h *Handler) DeleteAgent(c *gin.Context) {
	if err := h.d.Svc.DeleteAgent(c.Request.Context(), paramInt(c, "agent_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 记忆 ---

// ListMemory godoc
//
//	@Summary		AI 记忆列表
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai-memory [get]
func (h *Handler) ListMemory(c *gin.Context) {
	scope := MemoryScope(c.Query("scope"))
	var scopeID int64
	if v := c.Query("scope_id"); v != "" {
		scopeID, _ = strconv.ParseInt(v, 10, 64)
	}
	items, err := h.d.Svc.ListMemory(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), scope, scopeID)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// SetMemory godoc
//
//	@Summary		写入 AI 记忆
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	AIWorkspaceMemory
//	@Router			/ai-memory [post]
func (h *Handler) SetMemory(c *gin.Context) {
	var body struct {
		Scope   string `json:"scope"`
		ScopeID int64  `json:"scope_id"`
		Key     string `json:"key"`
		Content string `json:"content"`
		Source  string `json:"source"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.d.Svc.SetMemory(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID),
		MemoryScope(body.Scope), body.ScopeID, body.Key, body.Content,
		body.Source, c.GetInt64(middleware.CtxUserID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

// DeleteMemory godoc
//
//	@Summary		删除 AI 记忆
//	@Tags			ai-agent
//	@Produce		json
//	@Success		204
//	@Router			/ai-memory/{memory_id} [delete]
func (h *Handler) DeleteMemory(c *gin.Context) {
	if err := h.d.Svc.DeleteMemory(c.Request.Context(), paramInt(c, "memory_id"), c.GetInt64(middleware.CtxWorkspaceID)); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 会话 ---

// ListConversations godoc
//
//	@Summary		AI 会话列表
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai/conversations [get]
func (h *Handler) ListConversations(c *gin.Context) {
	items, err := h.d.Svc.ListConversations(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), c.GetInt64(middleware.CtxUserID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateConversation godoc
//
//	@Summary		创建 AI 会话
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Conversation
//	@Router			/ai/conversations [post]
func (h *Handler) CreateConversation(c *gin.Context) {
	var body struct {
		AgentID   int64  `json:"agent_id"`
		Mode      string `json:"mode"`
		Message   string `json:"message"`
		ProjectID int64  `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conv, err := h.d.Svc.CreateConversation(c.Request.Context(), ChatInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		UserID:      c.GetInt64(middleware.CtxUserID),
		AgentID:     body.AgentID,
		Mode:        Mode(body.Mode),
		Message:     body.Message,
		ProjectID:   body.ProjectID,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, conv)
}

// GetConversation godoc
//
//	@Summary		AI 会话详情（含消息）
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai/conversations/{conversation_id} [get]
func (h *Handler) GetConversation(c *gin.Context) {
	conv, msgs, err := h.d.Svc.GetConversation(c.Request.Context(), paramInt(c, "conversation_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"conversation": conv, "messages": msgs})
}

// Chat godoc
//
//	@Summary		发送 AI 消息
//	@Tags			ai-agent
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Message
//	@Router			/ai/conversations/{conversation_id}/messages [post]
func (h *Handler) Chat(c *gin.Context) {
	var body struct {
		Content   string `json:"content"`
		Mode      string `json:"mode"`
		ProjectID int64  `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := WithActor(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), c.GetInt64(middleware.CtxUserID))
	conversationID := paramInt(c, "conversation_id")

	// 1. 持久化用户消息
	userMsg, err := h.d.Svc.AppendMessage(ctx, conversationID, "user", body.Content, nil)
	if err != nil {
		h.fail(c, err)
		return
	}

	// 2. 未配置 LLM 时给出明确提示（不静默失败）
	if h.d.LLM == nil || !h.d.LLM.IsEnabled() {
		notice := "AI 模型未配置。请在环境变量中设置 YDSZ_LLM_API_BASE / YDSZ_LLM_API_KEY / YDSZ_LLM_MODEL 后重试。"
		assistantMsg, _ := h.d.Svc.AppendMessage(ctx, conversationID, "system", notice, nil)
		c.JSON(http.StatusCreated, gin.H{"user_message": userMsg, "assistant_message": assistantMsg, "llm_enabled": false})
		return
	}

	// 3. 读取会话与历史
	conv, history, err := h.d.Svc.GetConversation(ctx, conversationID)
	if err != nil {
		h.fail(c, err)
		return
	}
	agent, _ := h.d.Svc.GetAgent(ctx, conv.AgentID)

	// 3.1 创建并推进 Run（可观测生命周期）
	var runInfo map[string]any
	var runSvc *RunService
	if h.d.RunSvc != nil {
		runSvc = h.d.RunSvc
		run, rerr := runSvc.CreateRun(ctx, CreateRunInput{
			WorkspaceID:    c.GetInt64(middleware.CtxWorkspaceID),
			AgentID:        conv.AgentID,
			UserID:         c.GetInt64(middleware.CtxUserID),
			ConversationID: conversationID,
			Mode:           Mode(conv.Mode),
			Input:          body.Content,
			Trigger:        "user",
		})
		if rerr == nil {
			if _, serr := runSvc.AddStep(ctx, run.ID, "thought", "收到用户输入", body.Content); serr == nil {
				_ = serr
			}
			runInfo = map[string]any{"run_id": run.ID, "status": string(run.Status)}
			if _, terr := runSvc.Transition(ctx, run.ID, RunInProgress, "", ""); terr == nil {
				runInfo["status"] = string(RunInProgress)
			}
			// 记录 run_id 供后续完成时更新
			c.Set("ai_run_id", run.ID)
		}
	}

	// 4. 组装对话历史（转为 LLM 消息）
	llmHistory := make([]llm.Message, 0, len(history))
	for _, m := range history {
		role := llm.Role(m.Role)
		if role != llm.RoleUser && role != llm.RoleAssistant {
			continue
		}
		llmHistory = append(llmHistory, llm.Message{Role: role, Content: m.Content})
	}

	// 5. 调用 LLM（含工具调用）
	mode := Mode(conv.Mode)
	if body.Mode != "" && IsValidMode(body.Mode) {
		mode = Mode(body.Mode)
	}
	wsName := ""
	if v, ok := c.Get("workspace_name"); ok {
		if s, ok := v.(string); ok {
			wsName = s
		}
	}

	result, err := h.d.LLM.Chat(ctx, agent, mode, wsName, llmHistory, body.Content, 0)
	if err != nil {
		errMsg := "AI 调用失败：" + err.Error()
		assistantMsg, _ := h.d.Svc.AppendMessage(ctx, conversationID, "system", errMsg, nil)
		if runSvc != nil {
			if runID, ok := c.Get("ai_run_id"); ok {
				if _, terr := runSvc.Transition(ctx, runID.(int64), RunFailed, "", errMsg); terr == nil {
					runInfo["status"] = string(RunFailed)
				}
				runInfo["error"] = err.Error()
			}
		}
		c.JSON(http.StatusOK, gin.H{"user_message": userMsg, "assistant_message": assistantMsg, "llm_enabled": true, "error": err.Error(), "run": runInfo})
		return
	}

	// 6. 持久化 assistant 回复
	assistantMsg, err := h.d.Svc.AppendMessage(ctx, conversationID, "assistant", result.FinalText, nil)
	if err != nil {
		h.fail(c, err)
		return
	}

	// 7. 收尾 Run（记录步骤/用量/终态）
	if runSvc != nil {
		if runID, ok := c.Get("ai_run_id"); ok {
			id := runID.(int64)
			for _, tc := range result.ToolCalls {
				detail, _ := json.Marshal(tc)
				_, _ = runSvc.AddStep(ctx, id, "action", tc.Name, string(detail))
			}
			if result.FinalText != "" {
				_, _ = runSvc.AddStep(ctx, id, "response", "生成回复", result.FinalText)
			}
			_ = runSvc.RecordUsage(ctx, id, result.Usage.PromptTokens, result.Usage.CompletionTokens, len(result.ToolCalls))
			if _, terr := runSvc.Transition(ctx, id, RunCompleted, "", result.FinalText); terr == nil {
				runInfo["status"] = string(RunCompleted)
			}
			runInfo["prompt_tokens"] = result.Usage.PromptTokens
			runInfo["completion_tokens"] = result.Usage.CompletionTokens
			runInfo["tool_call_count"] = len(result.ToolCalls)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_message":      userMsg,
		"assistant_message": assistantMsg,
		"llm_enabled":       true,
		"tool_calls":        result.ToolCalls,
		"usage":             result.Usage,
		"iterations":        result.Iterations,
		"run":               runInfo,
	})
}

// --- 动作 ---

// ListPendingActions godoc
//
//	@Summary		待处理动作列表
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai/actions [get]
func (h *Handler) ListPendingActions(c *gin.Context) {
	items, err := h.d.Svc.ListPendingActions(c.Request.Context(), paramInt(c, "conversation_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ApproveAction godoc
//
//	@Summary		批准动作
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	AIAction
//	@Router			/ai/actions/{action_id}/approve [post]
func (h *Handler) ApproveAction(c *gin.Context) {
	a, err := h.d.Svc.ResolveAction(c.Request.Context(), paramInt(c, "action_id"), "approved", c.GetInt64(middleware.CtxUserID), "", "")
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

// RejectAction godoc
//
//	@Summary		拒绝动作
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	AIAction
//	@Router			/ai/actions/{action_id}/reject [post]
func (h *Handler) RejectAction(c *gin.Context) {
	a, err := h.d.Svc.ResolveAction(c.Request.Context(), paramInt(c, "action_id"), "rejected", c.GetInt64(middleware.CtxUserID), "", "")
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

// ExecuteAction godoc
//
//	@Summary		执行已批准的动作
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	AIAction
//	@Router			/ai/actions/{action_id}/execute [post]
func (h *Handler) ExecuteAction(c *gin.Context) {
	actions, err := h.d.Svc.ListPendingActions(c.Request.Context(), 0)
	if err == nil {
		for _, a := range actions {
			if a.ID == paramInt(c, "action_id") {
				res, eerr := h.d.Svc.ExecuteAction(c.Request.Context(), a, c.GetInt64(middleware.CtxUserID))
				if eerr != nil {
					h.fail(c, eerr)
					return
				}
				c.JSON(http.StatusOK, res)
				return
			}
		}
	}
	// 回退：直接按 ID 查询并执行
	a, err := h.d.Svc.getActionByID(c.Request.Context(), paramInt(c, "action_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	res, err := h.d.Svc.ExecuteAction(c.Request.Context(), a, c.GetInt64(middleware.CtxUserID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// --- Run 生命周期 ---

// ListRuns godoc
//
//	@Summary		Agent Run 列表
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai-runs [get]
func (h *Handler) ListRuns(c *gin.Context) {
	if h.d.RunSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "run service not configured"})
		return
	}
	agentID := int64Query(c, "agent_id")
	if agentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id required"})
		return
	}
	items, err := h.d.RunSvc.ListRuns(c.Request.Context(), agentID, intQuery(c, "limit", 50))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetRun godoc
//
//	@Summary		Agent Run 详情
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	AgentRun
//	@Router			/ai-runs/{run_id} [get]
func (h *Handler) GetRun(c *gin.Context) {
	if h.d.RunSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "run service not configured"})
		return
	}
	run, err := h.d.RunSvc.GetRun(c.Request.Context(), paramInt(c, "run_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

// ListRunSteps godoc
//
//	@Summary		Run 执行步骤
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/ai-runs/{run_id}/steps [get]
func (h *Handler) ListRunSteps(c *gin.Context) {
	if h.d.RunSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "run service not configured"})
		return
	}
	items, err := h.d.RunSvc.ListSteps(c.Request.Context(), paramInt(c, "run_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// StopRun godoc
//
//	@Summary		停止 Run
//	@Tags			ai-agent
//	@Produce		json
//	@Success		200	{object}	AgentRun
//	@Router			/ai-runs/{run_id}/stop [post]
func (h *Handler) StopRun(c *gin.Context) {
	if h.d.RunSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "run service not configured"})
		return
	}
	run, err := h.d.RunSvc.StopRun(c.Request.Context(), paramInt(c, "run_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

// int64Query 解析 query 中的 int64 参数（不存在返回 0）。
func int64Query(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return n
}
