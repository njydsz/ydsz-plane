// Package notification — 通知偏好设置 handlers。
package notification

import (
	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// RegisterPreferenceRoutes 注册通知偏好路由。
func (h *Handler) RegisterPreferenceRoutes(ws *gin.RouterGroup) {
	pref := ws.Group("/notification-preferences")
	{
		pref.GET("", h.GetPreference)
		pref.PUT("", h.UpdatePreference)
	}
}

// GetPreference 获取当前用户的通知偏好。
//
//	@Summary		获取通知偏好
//	@Tags			notification
//	@Produce		json
//	@Success		200	{object}	NotificationPreference
//	@Router			/notification-preferences [get]
func (h *Handler) GetPreference(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	p, err := h.svc.GetUserPreference(c.Request.Context(), wsID, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, p)
}

// UpdatePreference 更新当前用户的通知偏好。
//
//	@Summary		更新通知偏好
//	@Tags			notification
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	NotificationPreference
//	@Router			/notification-preferences [put]
func (h *Handler) UpdatePreference(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	var input PreferenceUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field:  "body", Reason: err.Error(),
		}))
		return
	}

	p, err := h.svc.UpdatePreference(c.Request.Context(), wsID, userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, p)
}
