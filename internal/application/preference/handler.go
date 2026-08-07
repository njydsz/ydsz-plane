// Package preference — 视图偏好 HTTP handlers。
package preference

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler 视图偏好 handler。
type Handler struct {
	svc *Service
}

// NewHandler 创建 handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册视图偏好路由（挂在项目子路由组下）。
func (h *Handler) Register(r *gin.RouterGroup) {
	pref := r.Group("/preferences")
	{
		pref.GET("/:view_type", h.getPreference)
		pref.PUT("/:view_type", h.savePreference)
		pref.GET("", h.listPreferences)
	}
}

// savePreference 保存视图偏好。
//
//	@Summary		保存视图偏好
//	@Tags			preference
//	@Accept			json
//	@Produce		json
//	@Param			body	body	upsertRequest	true	"偏好内容"
//	@Success		200		{object}	ViewPreference
//	@Router			/preferences/{view_type} [put]
func (h *Handler) savePreference(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)
	viewType := ViewType(c.Param("view_type"))

	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	vp, err := h.svc.Save(c.Request.Context(), wsID, projectID, userID, &ViewPreference{
		ViewType: viewType,
		Layout:   req.Layout,
		Columns:  req.Columns,
		Filters:  req.Filters,
		Sort:     req.Sort,
		Extra:    req.Extra,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, vp)
}

// getPreference 获取视图偏好。
//
//	@Summary		获取视图偏好
//	@Tags			preference
//	@Produce		json
//	@Success		200	{object}	ViewPreference
//	@Router			/preferences/{view_type} [get]
func (h *Handler) getPreference(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)
	viewType := ViewType(c.Param("view_type"))

	vp, err := h.svc.Get(c.Request.Context(), wsID, projectID, userID, viewType)
	if err != nil {
		writeErr(c, err)
		return
	}
	if vp == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, vp)
}

// listPreferences 列出全部视图偏好。
func (h *Handler) listPreferences(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	list, err := h.svc.ListByUser(c.Request.Context(), wsID, projectID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []ViewPreference{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

type upsertRequest struct {
	Layout  string          `json:"layout"`
	Columns json.RawMessage `json:"columns"`
	Filters json.RawMessage `json:"filters"`
	Sort    json.RawMessage `json:"sort"`
	Extra   json.RawMessage `json:"extra"`
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
