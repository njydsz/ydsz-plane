// Package notification 通知域 HTTP 处理器：查询/标记已读/全部已读，
// 以及通知偏好（preferences）的设置与读取。
package notification

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 通知 Handler 依赖。
type HandlerDeps struct {
	NotificationSvc *Service
	WorkspaceStore  *auth.WorkspaceMembershipStore
}

// Handler 通知 HTTP handler。
type Handler struct {
	svc     *Service
	wsStore *auth.WorkspaceMembershipStore
}

// NewHandler 创建通知 handler。
func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{svc: deps.NotificationSvc, wsStore: deps.WorkspaceStore}
}

// respondError 将 AppError 序列化为统一错误响应。
func respondError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		c.JSON(appErr.HTTP, gin.H{"error": appErr})
		return
	}
	log.Printf("notification handler: unexpected error: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL.ERROR", "message": "服务内部错误"},
	})
}

// RegisterRoutes 注册通知路由。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	notif := r.Group("/notifications")
	{
		notif.GET("", h.List)
		notif.GET("/unread-count", h.UnreadCount)
		notif.PUT("/:id/read", h.MarkRead)
		notif.PUT("/read-all", h.MarkAllRead)
		notif.PUT("/:id/archive", h.Archive)
	}
}

// List GET /api/v1/notifications
func (h *Handler) List(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	userID := c.GetInt64("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	input := ListInput{
		WorkspaceID: wsID,
		RecipientID: userID,
		Limit:       limit,
		Offset:      offset,
	}
	if v := c.Query("is_read"); v == "true" || v == "false" {
		b := v == "true"
		input.IsRead = &b
	}
	if v := c.Query("event_type"); v != "" {
		input.EventType = &v
	}
	if v := c.Query("since"); v != "" {
		if sinceMs, err := strconv.ParseInt(v, 10, 64); err == nil {
			input.Since = &sinceMs
		}
	}

	result, err := h.svc.List(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": result.Items, "total": result.Total,
		"limit": limit, "offset": offset,
	})
}

// UnreadCount GET /api/v1/notifications/unread-count
func (h *Handler) UnreadCount(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	userID := c.GetInt64("user_id")

	count, err := h.svc.UnreadCount(c.Request.Context(), wsID, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkRead PUT /api/v1/notifications/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, errs.Validation("NOTIFICATION.INVALID_ID", "无效的通知 ID"))
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// MarkAllRead PUT /api/v1/notifications/read-all
func (h *Handler) MarkAllRead(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	userID := c.GetInt64("user_id")

	count, err := h.svc.MarkAllRead(c.Request.Context(), wsID, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "count": count})
}

// Archive PUT /api/v1/notifications/:id/archive
func (h *Handler) Archive(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, errs.Validation("NOTIFICATION.INVALID_ID", "无效的通知 ID"))
		return
	}
	if err := h.svc.Archive(c.Request.Context(), id, userID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
