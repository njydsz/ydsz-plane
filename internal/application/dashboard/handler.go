// Package dashboard — 仪表盘 HTTP handlers。
package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps dashboard handler 依赖。
type HandlerDeps struct {
	DashboardSvc *Service
}

// DashboardHandler Gin handler 集合。
type DashboardHandler struct {
	d *HandlerDeps
}

// NewDashboardHandler 构造 handler。
func NewDashboardHandler(d *HandlerDeps) *DashboardHandler {
	return &DashboardHandler{d: d}
}

// Register 注册仪表盘路由。
func (h *DashboardHandler) Register(r *gin.RouterGroup) {
	r.GET("", h.GetDashboard)
	r.GET("/widgets", h.ListWidgets)
	r.POST("/widgets", h.CreateWidget)
	r.PATCH("/widgets/:widget_id", h.UpdateWidget)
	r.DELETE("/widgets/:widget_id", h.DeleteWidget)
	r.GET("/alerts", h.ListAlerts)
	r.POST("/alerts/:alert_id/resolve", h.ResolveAlert)
	r.GET("/templates", h.ListTemplates)
}

// GetDashboard 获取项目仪表盘完整数据。
//
//	@Summary		仪表盘数据
//	@Tags			dashboard
//	@Produce		json
//	@Success		200	{object}	DashboardData
//	@Router			/dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	data, err := h.d.DashboardSvc.GetDashboard(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// ListWidgets 列出项目 widgets。
func (h *DashboardHandler) ListWidgets(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	widgets, err := h.d.DashboardSvc.getWidgets(c.Request.Context(), projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if widgets == nil {
		widgets = []DashboardWidget{}
	}
	c.JSON(http.StatusOK, gin.H{"results": widgets})
}

// CreateWidget 创建 widget。
func (h *DashboardHandler) CreateWidget(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		WidgetType WidgetType     `json:"widget_type" binding:"required"`
		Title      string         `json:"title" binding:"required,max=50"`
		GridX      int            `json:"grid_x"`
		GridY      int            `json:"grid_y"`
		GridW      int            `json:"grid_w"`
		GridH      int            `json:"grid_h"`
		Config     map[string]any `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	w, err := h.d.DashboardSvc.SaveWidget(c.Request.Context(), SaveWidgetInput{
		ProjectID:  projectID,
		UserID:     userID,
		WidgetType: req.WidgetType,
		Title:      req.Title,
		GridX:      req.GridX,
		GridY:      req.GridY,
		GridW:      req.GridW,
		GridH:      req.GridH,
		Config:     req.Config,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, w)
}

// UpdateWidget 更新 widget 的网格位置 / 尺寸 / 配置 / 标题。
func (h *DashboardHandler) UpdateWidget(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	widgetID := int64Param(c, "widget_id")

	var req struct {
		GridX *int            `json:"grid_x"`
		GridY *int            `json:"grid_y"`
		GridW *int            `json:"grid_w"`
		GridH *int            `json:"grid_h"`
		Title *string         `json:"title"`
		Config map[string]any `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	if req.GridX == nil && req.GridY == nil && req.GridW == nil && req.GridH == nil &&
		req.Title == nil && req.Config == nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "body", Reason: "至少提供一个可更新字段（grid_x/grid_y/grid_w/grid_h/title/config）",
		}))
		return
	}

	w, err := h.d.DashboardSvc.UpdateWidget(c.Request.Context(), UpdateWidgetInput{
		WidgetID:  widgetID,
		ProjectID: projectID,
		GridX:     req.GridX,
		GridY:     req.GridY,
		GridW:     req.GridW,
		GridH:     req.GridH,
		Title:     req.Title,
		Config:    req.Config,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

// DeleteWidget 删除 widget。
func (h *DashboardHandler) DeleteWidget(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	widgetID := int64Param(c, "widget_id")

	if err := h.d.DashboardSvc.DeleteWidget(c.Request.Context(), widgetID, projectID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListAlerts 列出未解决风险告警。
func (h *DashboardHandler) ListAlerts(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	alerts, err := h.d.DashboardSvc.getActiveAlerts(c.Request.Context(), projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if alerts == nil {
		alerts = []RiskAlert{}
	}
	// TODO: 待 worker 定时检测新风险命中后，通过 hub.Publish 广播告警给全 ws 用户。
	// 调用 dashboard.BroadcastRiskAlert(ctx, hub, wsID, newAlert) 推送 {"type":"risk_alert","data":...}
	c.JSON(http.StatusOK, gin.H{"results": alerts})
}

// ResolveAlert 解决告警。
func (h *DashboardHandler) ResolveAlert(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	alertID := int64Param(c, "alert_id")

	if err := h.d.DashboardSvc.ResolveAlert(c.Request.Context(), wsID, alertID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListTemplates 列出仪表盘模板。
func (h *DashboardHandler) ListTemplates(c *gin.Context) {
	category := c.Query("category")
	templates, err := h.d.DashboardSvc.ListTemplates(c.Request.Context(), category)
	if err != nil {
		writeErr(c, err)
		return
	}
	if templates == nil {
		templates = []DashboardTemplate{}
	}
	c.JSON(http.StatusOK, gin.H{"results": templates})
}

// GetProjectCompare 返回工作空间下所有项目的完成率 / 缺陷数对比数据。
func (h *DashboardHandler) GetProjectCompare(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	items, err := h.d.DashboardSvc.GetProjectCompare(c.Request.Context(), wsID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if items == nil {
		items = []ProjectCompareItem{}
	}
	c.JSON(http.StatusOK, gin.H{"results": items})
}

// --- Helpers ---

func int64Param(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(c.Param(key), 10, 64)
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
