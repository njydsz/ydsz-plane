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
	WebhookSvc   *Service
	WorkspaceStore *auth.WorkspaceMembershipStore
	Dispatcher   *Dispatcher
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
	}
}

// Create godoc
//   - POST .../webhooks
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

// List godoc
//   - GET .../webhooks?project_id=&limit=&offset=
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

// Get godoc
//   - GET .../webhooks/:webhook_id
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

// Update godoc
//   - PATCH .../webhooks/:webhook_id
func (h *Handler) Update(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("WEBHOOK.INVALID_REQUEST", "无效请求"))
		return
	}

	input := UpdateInput{
		Name:      req.Name,
		TargetURL: req.TargetURL,
		Events:    req.Events,
		IsActive:  req.IsActive,
	}

	w, err := h.svc.Update(c.Request.Context(), wsID, webhookID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// Delete godoc
//   - DELETE .../webhooks/:webhook_id
func (h *Handler) Delete(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	webhookID := c.GetInt64("webhook_id")

	if err := h.svc.Delete(c.Request.Context(), wsID, webhookID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ListLogs godoc
//   - GET .../webhooks/:webhook_id/logs?status=&event_type=&limit=&offset=
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

// TestPing godoc
//   - POST .../webhooks/:webhook_id/test
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

// Retry godoc
//   - POST .../webhooks/:webhook_id/logs/:log_id/retry
func (h *Handler) Retry(c *gin.Context) {
	// MVP: 仅返回占位；完整实现需要重新读取 log 并执行投递
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "手动重投已接受"})
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
