// Package notification — 通知域 HTTP handlers。
//
// 提供通知查询、标记已读、全部已读及通知偏好设置等 REST 端点。
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

// List 获取当前用户的通知列表。
//
//	@Summary		通知列表
//	@Description	分页获取当前工作空间内当前用户的通知，支持已读过滤与事件类型过滤
//	@Tags			notification
//	@Produce		json
//	@Param			limit		query	int		false	"每页数"	default(20)
//	@Param			offset		query	int		false	"偏移"	default(0)
//	@Param			is_read		query	boolean	false	"已读过滤"
//	@Param			event_type	query	string	false	"事件类型"
//	@Param			since		query	int		false	"起始时间戳 (毫秒)"
//	@Success		200			{object}	map[string]any
//	@Router			/notifications [get]
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

// UnreadCount 获取未读通知数量。
//
//	@Summary		未读通知数
//	@Tags			notification
//	@Produce		json
//	@Success		200	{object}	map[string]int
//	@Router			/notifications/unread-count [get]
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

// MarkRead 标记单条通知为已读。
//
//	@Summary		标记已读
//	@Tags			notification
//	@Param			id	path	int	true	"通知 ID"
//	@Success		200	{object}	map[string]bool
//	@Router			/notifications/{id}/read [put]
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

// MarkAllRead 标记当前用户全部通知为已读。
//
//	@Summary		全部已读
//	@Tags			notification
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/notifications/read-all [put]
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

// Archive 归档单条通知。
//
//	@Summary		归档通知
//	@Tags			notification
//	@Param			id	path	int	true	"通知 ID"
//	@Success		200	{object}	map[string]bool
//	@Router			/notifications/{id}/archive [put]
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
