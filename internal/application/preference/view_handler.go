// Package preference — 命名视图 HTTP handlers。
package preference

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ViewHandler 命名视图 handler。
type ViewHandler struct {
	svc *ViewService
}

// NewViewHandler 创建视图 handler。
func NewViewHandler(svc *ViewService) *ViewHandler {
	return &ViewHandler{svc: svc}
}

// Register 注册视图路由。
func (h *ViewHandler) Register(r *gin.RouterGroup) {
	views := r.Group("/views")
	{
		views.GET("", h.listViews)
		views.POST("", h.createView)
		views.GET("/:view_id", h.getView)
		views.PATCH("/:view_id", h.updateView)
		views.DELETE("/:view_id", h.deleteView)
		views.POST("/:view_id/default", h.setDefaultView)
	}
}

// listViews 列出当前项目下的视图。
//
// Query: scope=personal|team|default（默认 team，列出所有可用视图）。
func (h *ViewHandler) listViews(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)
	scope := SavedViewScope(c.DefaultQuery("scope", "team"))

	list, err := h.svc.List(c.Request.Context(), wsID, projectID, userID, scope)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []SavedView{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// createView 创建新视图。
func (h *ViewHandler) createView(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var input CreateViewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	v, err := h.svc.Create(c.Request.Context(), wsID, projectID, userID, &input)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}

// getView 获取视图详情。
func (h *ViewHandler) getView(c *gin.Context) {
	viewID, err := parseViewID(c)
	if err != nil {
		return
	}

	v, err := h.svc.Get(c.Request.Context(), viewID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// updateView 更新视图（仅 owner）。
func (h *ViewHandler) updateView(c *gin.Context) {
	viewID, err := parseViewID(c)
	if err != nil {
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	// 先解析原始 JSON 为 map，做部分更新兼容
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	input := &UpdateViewInput{}
	if v, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err == nil {
			input.Name = &name
		}
	}
	if v, ok := raw["type"]; ok {
		var vt ViewType
		if err := json.Unmarshal(v, &vt); err == nil {
			input.Type = &vt
		}
	}
	if v, ok := raw["scope"]; ok {
		var sc SavedViewScope
		if err := json.Unmarshal(v, &sc); err == nil {
			input.Scope = &sc
		}
	}
	if v, ok := raw["config"]; ok {
		input.Config = &v
	}
	if v, ok := raw["is_shared"]; ok {
		var shared bool
		if err := json.Unmarshal(v, &shared); err == nil {
			input.IsShared = &shared
		}
	}

	sv, err := h.svc.Update(c.Request.Context(), viewID, userID, input)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sv)
}

// deleteView 删除视图（仅 owner）。
func (h *ViewHandler) deleteView(c *gin.Context) {
	viewID, err := parseViewID(c)
	if err != nil {
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.svc.Delete(c.Request.Context(), viewID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// setDefaultView 设置团队默认视图（管理员操作）。
func (h *ViewHandler) setDefaultView(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	viewID, err := parseViewID(c)
	if err != nil {
		return
	}

	if err := h.svc.SetDefault(c.Request.Context(), wsID, projectID, viewID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// --- helpers ---

func parseViewID(c *gin.Context) (int64, error) {
	var uri struct {
		ViewID int64 `uri:"view_id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "无效的视图 ID"}))
		return 0, err
	}
	return uri.ViewID, nil
}
