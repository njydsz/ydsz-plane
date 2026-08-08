// Package issue — Module HTTP handlers（REST API）。
//
// 对齐项目现有路由模式：
//   - 列表 + 创建 挂在 /modules
//   - 单资源 挂在 /modules/:module_id
//   - 工作项分配 挂在 /modules/:module_id/issues
package issue

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ModuleHandler Gin handler 集合（模块 CRUD + 工作项分配）。
type ModuleHandler struct {
	svc *ModuleService
}

// NewModuleHandler 构造模块 handler。
func NewModuleHandler(svc *ModuleService) *ModuleHandler {
	return &ModuleHandler{svc: svc}
}

// Register 注册模块路由（在项目子路由组下）。
func (h *ModuleHandler) Register(r *gin.RouterGroup) {
	mods := r.Group("/modules")
	{
		mods.GET("", h.listModules)
		mods.POST("", h.createModule)
	}
	mod := mods.Group("/:module_id")
	{
		mod.GET("", h.getModule)
		mod.PATCH("", h.updateModule)
		mod.DELETE("", h.deleteModule)

		mod.GET("/issues", h.listModuleIssues)
		mod.POST("/issues", h.assignIssues)
		mod.DELETE("/issues/:issue_id", h.unassignIssue)
	}
}

// ---- request/response ----

type createModuleRequest struct {
	Name        string  `json:"name" binding:"required,max=120"`
	Description string  `json:"description"`
	LeadID      *int64  `json:"lead_id"`
	StartDate   *string `json:"start_date"`
	TargetDate  *string `json:"target_date"`
	SortOrder   float64 `json:"sort_order"`
}

type updateModuleRequest struct {
	Name        *string  `json:"name" binding:"omitempty,max=120"`
	Description *string  `json:"description"`
	LeadID      *int64   `json:"lead_id"`
	Status      *string  `json:"status" binding:"omitempty,oneof=active completed archived"`
	StartDate   *string  `json:"start_date"`
	TargetDate  *string  `json:"target_date"`
	SortOrder   *float64 `json:"sort_order"`
}

type assignIssuesRequest struct {
	IssueIDs []int64 `json:"issue_ids" binding:"required,min=1,dive,gt=0"`
}

// ---- handlers ----

// listModules GET /api/v1/workspaces/:ws_id/projects/:project_id/modules
// @Summary		列出项目模块
// @Tags			module
// @Produce		json
// @Param			status		query		string	false	"状态筛选 (active|completed|archived)"
// @Success		200			{array}		Module
// @Router			/modules [get]
func (h *ModuleHandler) listModules(c *gin.Context) {
	wsID, projectID := extractProjectParams(c)
	status := c.Query("status")

	modules, err := h.svc.ListModules(c.Request.Context(), ListModulesFilter{
		WorkspaceID: wsID, ProjectID: projectID, Status: status,
	})
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, modules)
}

// createModule POST /modules
// @Summary		创建模块
// @Tags			module
// @Accept			json
// @Produce		json
// @Param			body		body		createModuleRequest	true	"模块信息"
// @Success		201			{object}	Module
// @Router			/modules [post]
func (h *ModuleHandler) createModule(c *gin.Context) {
	var req createModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "body", Reason: "请求体无效: " + err.Error(),
		})})
		return
	}

	userID := extractActorID(c)
	in := CreateModuleInput{
		WorkspaceID: extractWsID(c), ProjectID: extractProjectID(c),
		Name: req.Name, Description: req.Description, LeadID: req.LeadID,
		SortOrder: req.SortOrder, CreatedBy: userID,
	}
	if req.StartDate != nil {
		if d := parseDate(*req.StartDate); !d.IsZero() {
			in.StartDate = &d
		}
	}
	if req.TargetDate != nil {
		if d := parseDate(*req.TargetDate); !d.IsZero() {
			in.TargetDate = &d
		}
	}

	m, err := h.svc.CreateModule(c.Request.Context(), in)
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

// getModule GET /modules/:module_id
func (h *ModuleHandler) getModule(c *gin.Context) {
	wsID := extractWsID(c)
	moduleID := extractModuleID(c)

	m, err := h.svc.GetModule(c.Request.Context(), moduleID, wsID)
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

// updateModule PATCH /modules/:module_id
func (h *ModuleHandler) updateModule(c *gin.Context) {
	var req updateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "body", Reason: "请求体无效: " + err.Error(),
		})})
		return
	}

	in := UpdateModuleInput{
		ID: extractModuleID(c), WorkspaceID: extractWsID(c),
		ProjectID: extractProjectID(c),
		Name: req.Name, Description: req.Description, LeadID: req.LeadID,
		Status: req.Status, SortOrder: req.SortOrder,
	}
	if req.StartDate != nil {
		if d := parseDate(*req.StartDate); !d.IsZero() {
			in.StartDate = &d
		}
	}
	if req.TargetDate != nil {
		if d := parseDate(*req.TargetDate); !d.IsZero() {
			in.TargetDate = &d
		}
	}

	m, err := h.svc.UpdateModule(c.Request.Context(), in)
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

// deleteModule DELETE /modules/:module_id
func (h *ModuleHandler) deleteModule(c *gin.Context) {
	wsID := extractWsID(c)
	projectID := extractProjectID(c)
	moduleID := extractModuleID(c)

	if err := h.svc.DeleteModule(c.Request.Context(), moduleID, wsID, projectID); err != nil {
		handleModuleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// listModuleIssues GET /modules/:module_id/issues
func (h *ModuleHandler) listModuleIssues(c *gin.Context) {
	moduleID := extractModuleID(c)

	ids, err := h.svc.ListModuleIssues(c.Request.Context(), moduleID)
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"issue_ids": ids})
}

// assignIssues POST /modules/:module_id/issues
func (h *ModuleHandler) assignIssues(c *gin.Context) {
	var req assignIssuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "body", Reason: "无效的 issue_ids: " + err.Error(),
		})})
		return
	}

	err := h.svc.AssignIssues(c.Request.Context(), AssignIssuesInput{
		ModuleID: extractModuleID(c), IssueIDs: req.IssueIDs,
		CreatedBy: extractActorID(c),
	})
	if err != nil {
		handleModuleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// unassignIssue DELETE /modules/:module_id/issues/:issue_id
func (h *ModuleHandler) unassignIssue(c *gin.Context) {
	moduleID := extractModuleID(c)
	issueID, err := strconv.ParseInt(c.Param("issue_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 issue_id"})
		return
	}

	if err := h.svc.UnassignIssue(c.Request.Context(), moduleID, issueID); err != nil {
		handleModuleErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- helpers ----

func extractModuleID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("module_id"), 10, 64)
	return id
}

func extractWsID(c *gin.Context) int64 {
	return c.GetInt64(middleware.CtxWorkspaceID)
}

func extractProjectID(c *gin.Context) int64 {
	return c.GetInt64(middleware.CtxProjectID)
}

func extractProjectParams(c *gin.Context) (int64, int64) {
	return extractWsID(c), extractProjectID(c)
}

func extractActorID(c *gin.Context) int64 {
	id := c.GetInt64(middleware.CtxUserID)
	return id
}

func handleModuleErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
