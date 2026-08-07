// Package workbench — 工作台 HTTP handlers。
package workbench

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps workbench handler 依赖。
type HandlerDeps struct {
	WorkbenchSvc *Service
}

// WorkbenchHandler Gin handler 集合。
type WorkbenchHandler struct {
	d *HandlerDeps
}

// NewWorkbenchHandler 构造 handler。
func NewWorkbenchHandler(d *HandlerDeps) *WorkbenchHandler {
	return &WorkbenchHandler{d: d}
}

// Register 注册工作台路由。
func (h *WorkbenchHandler) Register(r *gin.RouterGroup) {
	r.GET("/summary", h.GetSummary)
	r.GET("/config", h.GetConfig)
	r.PUT("/config", h.SaveConfig)
	r.GET("/recent", h.ListRecent)
	r.POST("/recent", h.RecordRecent)
	r.GET("/templates", h.ListTemplates)
	r.POST("/templates/apply", h.ApplyTemplate)
}

// GetSummary godoc
//
//	@Summary		工作台首屏聚合
//	@Description	一次调用获取我的任务分桶、迭代概览、最近访问、快捷操作
//	@Tags			workbench
//	@Produce		json
//	@Success		200	{object}	WorkbenchSummary
//	@Router			/workbench [get]
func (h *WorkbenchHandler) GetSummary(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	summary, err := h.d.WorkbenchSvc.GetSummary(c.Request.Context(), wsID, userID, projID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetConfig godoc
//
//	@Summary		获取工作台布局配置
//	@Tags			workbench
//	@Produce		json
//	@Success		200	{object}	WorkbenchConfig
//	@Router			/workbench/config [get]
func (h *WorkbenchHandler) GetConfig(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	cfg, err := h.d.WorkbenchSvc.GetConfig(c.Request.Context(), wsID, userID, projID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// SaveConfig godoc
//
//	@Summary		保存工作台布局配置
//	@Tags			workbench
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkbenchConfig
//	@Router			/workbench/config [put]
func (h *WorkbenchHandler) SaveConfig(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Layout       LayoutConfig   `json:"layout"`
		WidgetStates map[string]any `json:"widget_states"`
		FocusEnabled bool           `json:"focus_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	cfg, err := h.d.WorkbenchSvc.SaveConfig(c.Request.Context(), SaveLayoutInput{
		WorkspaceID:  wsID,
		ProjectID:    projID,
		UserID:       userID,
		Layout:       req.Layout,
		WidgetStates: req.WidgetStates,
		FocusEnabled: req.FocusEnabled,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// ListRecent godoc
//
//	@Summary		最近访问列表
//	@Tags			workbench
//	@Produce		json
//	@Success		200	{array}		RecentItem
//	@Router			/workbench/recent [get]
func (h *WorkbenchHandler) ListRecent(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	limit := intQuery(c, "limit", 10)

	items, err := h.d.WorkbenchSvc.ListRecentItems(c.Request.Context(), wsID, userID, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	if items == nil {
		items = []RecentItem{}
	}
	c.JSON(http.StatusOK, gin.H{"results": items})
}

// RecordRecent godoc
//
//	@Summary		记录最近访问
//	@Tags			workbench
//	@Accept			json
//	@Success		204
//	@Router			/workbench/recent [post]
func (h *WorkbenchHandler) RecordRecent(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		ItemType   string `json:"item_type" binding:"required,oneof=project issue sprint version"`
		ItemID     int64  `json:"item_id" binding:"required"`
		ProjectID  int64  `json:"project_id"`
		Title      string `json:"title"`
		Identifier string `json:"identifier"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if err := h.d.WorkbenchSvc.RecordRecent(c.Request.Context(), RecordRecentInput{
		WorkspaceID: wsID,
		UserID:      userID,
		ItemType:    req.ItemType,
		ItemID:      req.ItemID,
		ProjectID:   &req.ProjectID,
		Title:       req.Title,
		Identifier:  req.Identifier,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListTemplates godoc
//
//	@Summary		工作台模板列表
//	@Tags			workbench
//	@Produce		json
//	@Success		200	{array}		WorkbenchTemplate
//	@Router			/workbench/templates [get]
func (h *WorkbenchHandler) ListTemplates(c *gin.Context) {
	templates, err := h.d.WorkbenchSvc.ListTemplates(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": templates})
}

// ApplyTemplate godoc
//
//	@Summary		应用模板到工作台
//	@Tags			workbench
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkbenchConfig
//	@Router			/workbench/templates/apply [post]
func (h *WorkbenchHandler) ApplyTemplate(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		TemplateSlug string `json:"template_slug" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	cfg, err := h.d.WorkbenchSvc.ApplyTemplate(c.Request.Context(), ApplyTemplateInput{
		WorkspaceID:  wsID,
		ProjectID:    projID,
		UserID:       userID,
		TemplateSlug: req.TemplateSlug,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// --- Helpers ---

func intQuery(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return def
	}
	return v
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
