// Package webhook — Webhook HTTP 处理器：订阅 CRUD、日志查看、测试投递。
package webhook

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 是 webhook Handler 的依赖集。
type HandlerDeps struct {
	WebhookSvc     *Service
	WorkspaceStore *auth.WorkspaceMembershipStore
	Dispatcher     *Dispatcher
}

// Handler 是 webhook HTTP handler。
type Handler struct {
	svc        *Service
	wsStore    *auth.WorkspaceMembershipStore
	dispatcher *Dispatcher
}

// NewHandler 构造 webhook handler。
func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		svc:        deps.WebhookSvc,
		wsStore:    deps.WorkspaceStore,
		dispatcher: deps.Dispatcher,
	}
}

// Register 注册 webhook 路由。
// 路由前缀：/api/v1/workspaces/:workspace_id/webhooks
func (h *Handler) Register(r *gin.RouterGroup) {
	hooks := r.Group("/webhooks")
	{
		hooks.POST("", h.Create)
		hooks.GET("", h.List)
		hooks.GET("/:webhook_id", h.Get)
		hooks.PATCH("/:webhook_id", h.Update)
		hooks.DELETE("/:webhook_id", h.Delete)

		// 投递日志
		hooks.GET("/:webhook_id/logs", h.ListLogs)
		// 测试投递
		hooks.POST("/:webhook_id/test", h.TestPing)
		// 重投（手动）
		hooks.POST("/:webhook_id/logs/:log_id/retry", h.Retry)
		// 暂停 / 恢复（PATCH is_active 的语义化别名，管理页"暂停/恢复"按钮）
		hooks.POST("/:webhook_id/pause", h.Pause)
		hooks.POST("/:webhook_id/resume", h.Resume)
	}
}

// Create 创建 Webhook 订阅。
//
//	@Summary		创建 Webhook
//	@Description	在工作空间下新建一个 Webhook 订阅，创建时返回 secret（后续不再返回）
//	@Tags			webhook
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	map[string]any
//	@Failure		422		{object}	errs.AppError
//	@Router			/webhooks [post]
func (h *Handler) Create(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	userID := c.GetInt64("user_id")

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("WEBHOOK.INVALID_REQUEST", "无效请求"))
		return
	}

	// 生成 secret（32 字节 hex = 256 位熵）
	secret := req.Secret
	if secret == "" {
		secret = generateSecret()
	}

	input := CreateInput{
		WorkspaceID: wsID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		TargetURL:   req.TargetURL,
		Secret:      secret,
		Events:      req.Events,
		CreatedBy:   userID,
	}

	w, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}

	// 创建时返回 secret（后续 GET 不再返回）
	type createResponse struct {
		Webhook
		Secret string `json:"secret"`
	}
	c.JSON(http.StatusOK, createResponse{
		Webhook: *w,
		Secret:  secret,
	})
}

// List 列出 Webhook 订阅。
//
//	@Summary		列出 Webhook
//	@Description	列出工作空间下的全部 Webhook 订阅，支持分页与项目过滤
//	@Tags			webhook
//	@Produce		json
//	@Param			project_id	query		int	false	"项目 ID"
//	@Param			limit		query		int	false	"每页数 (1-100)"
//	@Param			offset		query		int	false	"偏移"
//	@Success		200			{object}	map[string]any
//	@Router			/webhooks [get]
func (h *Handler) List(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")

	input := ListInput{WorkspaceID: wsID}
	if pid := c.Query("project_id"); pid != "" {
		if id, err := strconv.ParseInt(pid, 10, 64); err == nil {
			input.ProjectID = &id
		}
	}
	if v := c.Query("limit"); v != "" {
		input.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		input.Offset, _ = strconv.Atoi(v)
	}

	result, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  input.Limit,
		"offset": input.Offset,
	})
}

// Get 获取 Webhook 详情。
//
//	@Summary		获取 Webhook 详情
//	@Description	返回单个 Webhook 的完整配置（不含 secret）
//	@Tags			webhook
//	@Produce		json
//	@Param			webhook_id	path		int	true	"Webhook ID"
//	@Success		200			{object}	Webhook
//	@Failure		404			{object}	errs.AppError
//	@Router			/webhooks/{webhook_id} [get]
func (h *Handler) Get(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	w, err := h.svc.GetByID(c.Request.Context(), wsID, webhookID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// Update 更新 Webhook 配置。
//
//	@Summary		更新 Webhook
//	@Description	PATCH 更新 Webhook 的 URL / 事件 / 名称 / 启用状态
//	@Tags			webhook
//	@Accept			json
//	@Produce		json
//	@Param			webhook_id	path		int				true	"Webhook ID"
//	@Success		200			{object}	Webhook
//	@Failure		404			{object}	errs.AppError
//	@Router			/webhooks/{webhook_id} [patch]
func (h *Handler) Update(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("WEBHOOK.INVALID_REQUEST", "无效请求"))
		return
	}

	input := UpdateInput(req)

	w, err := h.svc.Update(c.Request.Context(), wsID, webhookID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// Delete 删除 Webhook 订阅。
//
//	@Summary		删除 Webhook
//	@Description	删除指定 Webhook 订阅
//	@Tags			webhook
//	@Param			webhook_id	path	int	true	"Webhook ID"
//	@Success		200
//	@Router			/webhooks/{webhook_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	if err := h.svc.Delete(c.Request.Context(), wsID, webhookID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListLogs 列出 Webhook 投递日志。
//
//	@Summary		列出投递日志
//	@Description	分页查询指定 Webhook 的投递日志，支持状态/事件类型过滤
//	@Tags			webhook
//	@Produce		json
//	@Param			webhook_id	path		int		true	"Webhook ID"
//	@Param			status			query		string	false	"投递状态"
//	@Param			event_type		query		string	false	"事件类型"
//	@Param			limit			query		int		false	"每页数"
//	@Param			offset			query		int		false	"偏移"
//	@Success		200				{object}	map[string]any
//	@Router			/webhooks/{webhook_id}/logs [get]
func (h *Handler) ListLogs(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	input := ListLogsInput{WorkspaceID: wsID, WebhookID: &webhookID}
	if v := c.Query("status"); v != "" {
		input.Status = &v
	}
	if v := c.Query("event_type"); v != "" {
		input.EventType = &v
	}
	if v := c.Query("limit"); v != "" {
		input.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		input.Offset, _ = strconv.Atoi(v)
	}

	result, err := h.svc.ListLogs(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  input.Limit,
		"offset": input.Offset,
	})
}

// TestPing 发送测试事件到 Webhook。
//
//	@Summary		测试投递
//	@Description	向 Webhook 目标 URL 发送一条测试事件（ping）
//	@Tags			webhook
//	@Param			webhook_id	path	int	true	"Webhook ID"
//	@Success		200			{object}	map[string]any
//	@Router			/webhooks/{webhook_id}/test [post]
func (h *Handler) TestPing(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	w, err := h.svc.GetByID(c.Request.Context(), wsID, webhookID)
	if err != nil {
		respondError(c, err)
		return
	}

	if h.dispatcher == nil {
		respondError(c, errs.ErrInternal)
		return
	}

	if err := h.dispatcher.ExecuteTestPing(c.Request.Context(), w); err != nil {
		respondError(c, pingError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "测试事件已投递"})
}

// Retry 手动重投指定投递日志。
//
//	@Summary		重投日志
//	@Description	手动重投指定投递日志（回查 domain_events 重建原始事件并同步重投）
//	@Tags			webhook
//	@Param			webhook_id	path	int	true	"Webhook ID"
//	@Param			log_id		path	int	true	"日志 ID"
//	@Success		200			{object}	map[string]any
//	@Router			/webhooks/{webhook_id}/logs/{log_id}/retry [post]
func (h *Handler) Retry(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")
	logID := c.GetInt64("log_id")

	if h.dispatcher == nil {
		respondError(c, errs.ErrInternal)
		return
	}

	if err := h.dispatcher.RetryLog(c.Request.Context(), wsID, webhookID, logID); err != nil {
		respondError(c, errs.Validation("WEBHOOK.RETRY_FAILED", "重投失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "重投成功"})
}

// Pause 暂停 Webhook 投递。
//
//	@Summary		暂停投递
//	@Description	暂停 Webhook 投递（等价于 PATCH { is_active: false }）
//	@Tags			webhook
//	@Produce		json
//	@Param			webhook_id	path		int	true	"Webhook ID"
//	@Success		200			{object}	Webhook
//	@Router			/webhooks/{webhook_id}/pause [post]
func (h *Handler) Pause(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	active := false
	w, err := h.svc.Update(c.Request.Context(), wsID, webhookID, UpdateInput{IsActive: &active})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// Resume 恢复 Webhook 投递。
//
//	@Summary		恢复投递
//	@Description	恢复 Webhook 投递（等价于 PATCH { is_active: true }）
//	@Tags			webhook
//	@Produce		json
//	@Param			webhook_id	path		int	true	"Webhook ID"
//	@Success		200			{object}	Webhook
//	@Router			/webhooks/{webhook_id}/resume [post]
func (h *Handler) Resume(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	active := true
	w, err := h.svc.Update(c.Request.Context(), wsID, webhookID, UpdateInput{IsActive: &active})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// --- 请求 DTO ---

type createRequest struct {
	Name      string   `json:"name" binding:"required"`
	TargetURL string   `json:"target_url" binding:"required"`
	Events    []string `json:"events"`
	Secret    string   `json:"secret,omitempty"`
	ProjectID *int64   `json:"project_id,omitempty"`
}

type updateRequest struct {
	Name      *string  `json:"name,omitempty"`
	TargetURL *string  `json:"target_url,omitempty"`
	Events    []string `json:"events,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}

// --- 辅助 ---

func generateSecret() string {
	var b [32]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func pingError(err error) *errs.AppError {
	return errs.Validation("WEBHOOK.TEST_FAILED", "测试投递失败: "+err.Error())
}

func respondError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		c.JSON(appErr.HTTP, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL.ERROR", "message": "服务内部错误"},
	})
}
