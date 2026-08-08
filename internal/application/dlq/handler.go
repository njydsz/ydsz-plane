// Package dlq — DLQ 管理 HTTP 处理器。
package dlq

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler DLQ 管理 handler。
type Handler struct {
	svc *Service
}

// NewHandler 创建 DLQ handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册 DLQ 管理路由。
// 路由前缀：/api/v1/workspaces/:workspace_id/admin/dlq（需 audit:read 权限）。
func (h *Handler) Register(r *gin.RouterGroup) {
	dlq := r.Group("/admin/dlq")
	{
		dlq.GET("", h.List)
		dlq.POST("/:id/retry", h.Retry)
		dlq.DELETE("/:id", h.Remove)
		dlq.POST("/cleanup", h.Cleanup)
	}
}

// List godoc
//   - GET .../admin/dlq?offset=&limit=&unresolved_only=
func (h *Handler) List(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")

	opts := ListOptions{}
	if v := c.Query("limit"); v != "" {
		opts.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		opts.Offset, _ = strconv.Atoi(v)
	}
	if v := c.Query("unresolved_only"); v == "true" || v == "1" {
		opts.UnresolvedOnly = true
	}

	items, total, err := h.svc.List(c.Request.Context(), wsID, opts)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// Retry godoc
//   - POST .../admin/dlq/:id/retry
func (h *Handler) Retry(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, errs.Validation("DLQ.INVALID_ID", "无效的死信 ID"))
		return
	}
	if err := h.svc.Retry(c.Request.Context(), wsID, id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "已重新入队，等待出站重放"})
}

// Remove godoc
//   - DELETE .../admin/dlq/:id
func (h *Handler) Remove(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, errs.Validation("DLQ.INVALID_ID", "无效的死信 ID"))
		return
	}
	if err := h.svc.Remove(c.Request.Context(), wsID, id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Cleanup godoc
//   - POST .../admin/dlq/cleanup  body: {"event_ids": [...]} | {"resolved_all": true}
func (h *Handler) Cleanup(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")

	var req struct {
		EventIDs    []int64 `json:"event_ids"`
		ResolvedAll bool    `json:"resolved_all"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("DLQ.INVALID_REQUEST", "无效请求体"))
		return
	}
	if len(req.EventIDs) == 0 && !req.ResolvedAll {
		respondError(c, errs.Validation("DLQ.INVALID_REQUEST", "需提供 event_ids 或 resolved_all"))
		return
	}

	n, err := h.svc.Cleanup(c.Request.Context(), wsID, req.EventIDs, req.ResolvedAll)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "resolved": n})
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
