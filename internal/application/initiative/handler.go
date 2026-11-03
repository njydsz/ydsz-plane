// Package initiative — 倡议 HTTP handlers（REST API）。
//
// 提供倡议 CRUD、归档/恢复、项目/工作项关联管理、进度聚合等 REST 端点。
// 路由通过 RegisterInitiativeRoutes 注册到 Gin 引擎。
package initiative

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 倡议 handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler 是 Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// RegisterRead 注册只读路由（列表 / 详情 / 进度）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/initiatives")
	g.GET("", h.listInitiatives)

	it := g.Group("/:initiative_id")
	{
		it.GET("", h.getInitiative)
		it.GET("/detail", h.getInitiativeDetail)
		it.GET("/progress", h.getProgress)
	}
}

// RegisterWrite 注册写路由（创建 / 更新 / 删除 / 归档 / 恢复 / 关联管理）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/initiatives")
	g.POST("", h.createInitiative)

	it := g.Group("/:initiative_id")
	{
		it.PATCH("", h.updateInitiative)
		it.DELETE("", h.deleteInitiative)

		// 归档 / 恢复
		it.POST("/archive", h.archiveInitiative)
		it.POST("/restore", h.restoreInitiative)

		// 关联项目
		it.POST("/projects", h.addProject)
		it.DELETE("/projects/:project_id", h.removeProject)

		// 关联工作项
		it.POST("/work-items", h.addWorkItem)
		it.DELETE("/work-items/:item_id", h.removeWorkItem)
	}
}

// Register 注册全部路由（读 + 写），供单一权限层级场景使用。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- CRUD ---

// createInitiative 创建倡议。
//
//	@Summary		创建倡议
//	@Description	在工作空间下创建战略性倡议
//	@Tags			initiative
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Initiative
//	@Failure		422	{object}	errs.AppError
//	@Router			/initiatives [post]
func (h *Handler) createInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Title       string  `json:"title" binding:"required,min=1,max=200"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		Priority    string  `json:"priority"`
		StartDate   *string `json:"start_date"`
		TargetDate  *string `json:"target_date"`
		OwnerID     *int64  `json:"owner_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := CreateInitiativeInput{
		WorkspaceID: wsID,
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		CreatedBy:   userID,
	}
	if req.Status != "" {
		in.Status = InitiativeStatus(req.Status)
	}
	if req.Priority != "" {
		in.Priority = InitiativePriority(req.Priority)
	}
	if req.StartDate != nil {
		if d, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			in.StartDate = &d
		}
	}
	if req.TargetDate != nil {
		if d, err := time.Parse("2006-01-02", *req.TargetDate); err == nil {
			in.TargetDate = &d
		}
	}

	it, err := h.d.Svc.Create(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, it)
}

// listInitiatives 列出倡议。
//
//	@Summary		列出倡议
//	@Description	支持按状态/优先级/负责人过滤
//	@Tags			initiative
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/initiatives [get]
func (h *Handler) listInitiatives(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	opts := ListInitiativesOptions{
		WorkspaceID: wsID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if v := c.Query("status"); v != "" {
		s := InitiativeStatus(v)
		opts.Status = &s
	}
	if v := c.Query("priority"); v != "" {
		p := InitiativePriority(v)
		opts.Priority = &p
	}
	if v := c.Query("owner_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			opts.OwnerID = &id
		}
	}

	items, total, err := h.d.Svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": total})
}

// getInitiative 获取倡议详情。
//
//	@Summary		获取倡议详情
//	@Tags			initiative
//	@Produce		json
//	@Success		200	{object}	Initiative
//	@Failure		404	{object}	errs.AppError
//	@Router			/initiatives/{initiative_id} [get]
func (h *Handler) getInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	it, err := h.d.Svc.GetByID(c.Request.Context(), wsID, initiativeID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// getInitiativeDetail 获取倡议详情（含关联项目与工作项）。
//
//	@Summary		获取倡议详情（含关联数据）
//	@Tags			initiative
//	@Produce		json
//	@Success		200	{object}	InitiativeDetail
//	@Router			/initiatives/{initiative_id}/detail [get]
func (h *Handler) getInitiativeDetail(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	detail, err := h.d.Svc.GetDetail(c.Request.Context(), wsID, initiativeID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// updateInitiative 更新倡议字段。
//
//	@Summary		更新倡议字段
//	@Tags			initiative
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Initiative
//	@Router			/initiatives/{initiative_id} [patch]
func (h *Handler) updateInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *string `json:"priority"`
		StartDate   *string `json:"start_date"`
		TargetDate  *string `json:"target_date"`
		OwnerID     *int64  `json:"owner_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := UpdateInitiativeInput{}
	if req.Title != nil {
		in.Title = req.Title
	}
	if req.Description != nil {
		in.Description = req.Description
	}
	if req.Status != nil {
		s := InitiativeStatus(*req.Status)
		in.Status = &s
	}
	if req.Priority != nil {
		p := InitiativePriority(*req.Priority)
		in.Priority = &p
	}
	if req.OwnerID != nil {
		in.OwnerID = req.OwnerID
	}
	if req.StartDate != nil {
		if d, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			in.StartDate = &d
		}
	}
	if req.TargetDate != nil {
		if d, err := time.Parse("2006-01-02", *req.TargetDate); err == nil {
			in.TargetDate = &d
		}
	}

	it, err := h.d.Svc.Update(c.Request.Context(), wsID, initiativeID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// deleteInitiative 删除倡议。
//
//	@Summary		删除倡议（硬删除，同时清理关联关系）
//	@Tags			initiative
//	@Success		204
//	@Router			/initiatives/{initiative_id} [delete]
func (h *Handler) deleteInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	if err := h.d.Svc.Delete(c.Request.Context(), wsID, initiativeID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Archive / Restore ---

// archiveInitiative 归档倡议。
//
//	@Summary		归档倡议（软删除）
//	@Tags			initiative
//	@Success		204
//	@Router			/initiatives/{initiative_id}/archive [post]
func (h *Handler) archiveInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	if err := h.d.Svc.Archive(c.Request.Context(), wsID, initiativeID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// restoreInitiative 恢复已归档倡议。
//
//	@Summary		恢复已归档倡议
//	@Tags			initiative
//	@Success		204
//	@Router			/initiatives/{initiative_id}/restore [post]
func (h *Handler) restoreInitiative(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	if err := h.d.Svc.Restore(c.Request.Context(), wsID, initiativeID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 关联管理：项目 ---

// addProject 将项目关联到倡议。
//
//	@Summary		关联项目到倡议
//	@Tags			initiative
//	@Accept			json
//	@Success		204
//	@Router			/initiatives/{initiative_id}/projects [post]
func (h *Handler) addProject(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	var req struct {
		ProjectID int64 `json:"project_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if err := h.d.Svc.AddProject(c.Request.Context(), wsID, initiativeID, req.ProjectID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// removeProject 从倡议移除项目关联。
//
//	@Summary		从倡议移除项目关联
//	@Tags			initiative
//	@Success		204
//	@Router			/initiatives/{initiative_id}/projects/{project_id} [delete]
func (h *Handler) removeProject(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")
	projectID := int64Param(c, "project_id")

	if err := h.d.Svc.RemoveProject(c.Request.Context(), wsID, initiativeID, projectID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 关联管理：工作项 ---

// addWorkItem 将工作项关联到倡议。
//
//	@Summary		关联工作项到倡议
//	@Tags			initiative
//	@Accept			json
//	@Success		204
//	@Router			/initiatives/{initiative_id}/work-items [post]
func (h *Handler) addWorkItem(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	var req struct {
		WorkItemID   int64  `json:"work_item_id" binding:"required"`
		WorkItemType string `json:"work_item_type" binding:"required,oneof=requirement task defect"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if err := h.d.Svc.AddWorkItem(c.Request.Context(), wsID, initiativeID, req.WorkItemID, req.WorkItemType); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// removeWorkItem 从倡议移除工作项关联。
//
//	@Summary		从倡议移除工作项关联
//	@Tags			initiative
//	@Success		204
//	@Router			/initiatives/{initiative_id}/work-items/{item_id} [delete]
func (h *Handler) removeWorkItem(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")
	itemID := int64Param(c, "item_id")

	if err := h.d.Svc.RemoveWorkItem(c.Request.Context(), wsID, initiativeID, itemID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 进度聚合 ---

// getProgress 获取倡议进度聚合。
//
//	@Summary		获取倡议进度聚合
//	@Description	聚合关联项目与工作项的完成情况
//	@Tags			initiative
//	@Produce		json
//	@Success		200	{object}	InitiativeProgress
//	@Router			/initiatives/{initiative_id}/progress [get]
func (h *Handler) getProgress(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	initiativeID := int64Param(c, "initiative_id")

	progress, err := h.d.Svc.GetProgress(c.Request.Context(), wsID, initiativeID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, progress)
}

// --- helpers ---

func int64Param(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return v
}

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
