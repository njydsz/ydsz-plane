// Package sprint — Sprint HTTP handlers（REST API）。
package sprint

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler 是 Gin handler 集合。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册 Sprint 路由（使用已校验 auth / workspace / project 的父路由组）。
func (h *Handler) Register(r *gin.RouterGroup) {
	// 集合
	r.GET("/sprints", h.listSprints)
	r.POST("/sprints", h.createSprint)
	r.GET("/sprints/backlog", h.getBacklog)
	r.GET("/sprints/suggest-capacity", h.suggestCapacity)

	// 单资源
	sprint := r.Group("/sprints/:sprint_id")
	{
		sprint.GET("", h.getSprint)
		sprint.PATCH("", h.updateSprint)
		sprint.DELETE("", h.deleteSprint)
		sprint.POST("/start", h.startSprint)
		sprint.POST("/complete", h.completeSprint)

		// 进度/规划
		sprint.GET("/progress", h.getSprintProgress)
		sprint.GET("/issues", h.listSprintIssues)
		sprint.POST("/issues", h.addIssue)
		sprint.DELETE("/issues/:issue_id", h.removeIssue)

		// 燃尽图 / 复盘
		sprint.GET("/burndown", h.burndown)
		sprint.GET("/review", h.getReview)
	}
}

// --- Sprint CRUD ---

// createSprint godoc
//
//	@Summary		创建迭代
//	@Description	在项目下创建 planned 迭代
//	@Tags			sprint
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	Sprint
//	@Failure		422		{object}	errs.AppError
//	@Router			/sprints [post]
func (h *Handler) createSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Name        string  `json:"name" binding:"required,min=1,max=80"`
		Description string  `json:"description" binding:"max=500"`
		Goal        string  `json:"goal" binding:"max=500"`
		StartDate   *string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
		EndDate     *string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
		Capacity    *float64 `json:"capacity" binding:"omitempty,min=0,max=99999"`
		OwnerID     *int64  `json:"owner_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := CreateSprintInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Goal:        req.Goal,
		Capacity:    req.Capacity,
		OwnerID:     req.OwnerID,
		CreatedBy:   userID,
	}
	if req.StartDate != nil {
		if d, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			in.StartDate = &d
		}
	}
	if req.EndDate != nil {
		if d, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			in.EndDate = &d
		}
	}

	sp, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, sp)
}

// listSprints godoc
//
//	@Summary		列出迭代
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	sprintListResponse
//	@Router			/sprints [get]
func (h *Handler) listSprints(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListSprintsOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if v := c.Query("status"); v != "" {
		s := SprintStatusCode(v)
		opts.Status = &s
	}

	sprints, total, err := h.svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": sprints, "total": total})
}

// getSprint godoc
//
//	@Summary		获取迭代详情
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	Sprint
//	@Failure		404		{object}	errs.AppError
//	@Router			/sprints/{sprint_id} [get]
func (h *Handler) getSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	sp, err := h.svc.GetByID(c.Request.Context(), wsID, sprintID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// updateSprint godoc
//
//	@Summary		更新迭代字段（仅 planned 状态可编辑）
//	@Tags			sprint
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	Sprint
//	@Router			/sprints/{sprint_id} [patch]
func (h *Handler) updateSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Goal        *string  `json:"goal"`
		StartDate   *string  `json:"start_date"`
		EndDate     *string  `json:"end_date"`
		Capacity    *float64 `json:"capacity"`
		OwnerID     *int64   `json:"owner_id"`
		Version     int      `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := UpdateSprintInput{Version: req.Version}
	if req.Name != nil {
		in.Name = req.Name
	}
	if req.Description != nil {
		in.Description = req.Description
	}
	if req.Goal != nil {
		in.Goal = req.Goal
	}
	if req.Capacity != nil {
		in.Capacity = req.Capacity
	}
	if req.OwnerID != nil {
		in.OwnerID = req.OwnerID
	}
	if req.StartDate != nil {
		if d, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			in.StartDate = &d
		}
	}
	if req.EndDate != nil {
		if d, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			in.EndDate = &d
		}
	}

	sp, err := h.svc.Update(c.Request.Context(), wsID, sprintID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// deleteSprint godoc
//
//	@Summary		归档迭代（仅 planned / completed 可删除）
//	@Tags			sprint
//	@Success		204
//	@Router			/sprints/{sprint_id} [delete]
func (h *Handler) deleteSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	if err := h.svc.SoftDelete(c.Request.Context(), wsID, sprintID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Lifecycle ---

// startSprint godoc
//
//	@Summary		启动迭代
//	@Description	将迭代从 planned 切换到 active；校验唯一 active 约束
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	Sprint
//	@Failure		409		{object}	errs.AppError	"项目已有 active 迭代"
//	@Router			/sprints/{sprint_id}:start [post]
func (h *Handler) startSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	sp, err := h.svc.Start(c.Request.Context(), wsID, sprintID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// completeSprint godoc
//
//	@Summary		结束迭代
//	@Description	将迭代从 active 切换到 completed；处理未完成任务
//	@Tags			sprint
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	Sprint
//	@Router			/sprints/{sprint_id}:complete [post]
func (h *Handler) completeSprint(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	var req struct {
		Strategy     string `json:"strategy" binding:"required,oneof=backlog next_sprint keep"`
		NextSprintID *int64 `json:"next_sprint_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := CompleteSprintInput{
		Strategy:     UnfinishedStrategy(req.Strategy),
		NextSprintID: req.NextSprintID,
	}
	sp, err := h.svc.Complete(c.Request.Context(), wsID, sprintID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// getSprintProgress godoc
//
//	@Summary		获取迭代实时进度
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	SprintProgress
//	@Router			/sprints/{sprint_id}/progress [get]
func (h *Handler) getSprintProgress(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	sp, err := h.svc.GetByID(c.Request.Context(), wsID, sprintID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp.Progress)
}

// --- Planning ---

// listSprintIssues godoc
//
//	@Summary		列出迭代内工作项
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	sprintListIssuesResponse
//	@Router			/sprints/{sprint_id}/issues [get]
func (h *Handler) listSprintIssues(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	limit := intQuery(c, "limit", 50)
	offset := intQuery(c, "offset", 0)

	views, total, err := h.svc.ListSprintIssues(c.Request.Context(), wsID, sprintID, limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": views, "total": total})
}

// addIssue godoc
//
//	@Summary		将工作项加入迭代（拖拽规划 / 中途加项）
//	@Tags			sprint
//	@Accept			json
//	@Success		204
//	@Router			/sprints/{sprint_id}/issues [post]
func (h *Handler) addIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		IssueID   int64   `json:"issue_id" binding:"required"`
		SortOrder float64 `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if err := h.svc.AddIssue(c.Request.Context(), wsID, AddIssueInput{
		SprintID:  sprintID,
		IssueID:   req.IssueID,
		SortOrder: req.SortOrder,
		AddedBy:   userID,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// removeIssue godoc
//
//	@Summary		从迭代移除工作项
//	@Tags			sprint
//	@Success		204
//	@Router			/sprints/{sprint_id}/issues/{issue_id} [delete]
func (h *Handler) removeIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")
	issueID := int64Param(c, "issue_id")

	if err := h.svc.RemoveIssue(c.Request.Context(), wsID, RemoveIssueInput{
		SprintID: sprintID,
		IssueID:  issueID,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// getBacklog godoc
//
//	@Summary		获取 Backlog 工作项列表（未规划进 active 迭代的未完成工作项）
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	sprintListResponse
//	@Router			/sprints/backlog [get]
func (h *Handler) getBacklog(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	limit := intQuery(c, "limit", 50)
	offset := intQuery(c, "offset", 0)

	items, total, err := h.svc.GetBacklog(c.Request.Context(), wsID, projectID, limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": total})
}

// --- Analytics ---

// burndown godoc
//
//	@Summary		获取迭代燃尽图数据
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	sprintBurndownResponse
//	@Router			/sprints/{sprint_id}/burndown [get]
func (h *Handler) burndown(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	sp, points, err := h.svc.BurndownData(c.Request.Context(), wsID, sprintID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sprint": sp, "points": points})
}

// getReview godoc
//
//	@Summary		获取迭代复盘数据
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	ReviewSnapshot
//	@Router			/sprints/{sprint_id}/review [get]
func (h *Handler) getReview(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sprintID := int64Param(c, "sprint_id")

	sp, err := h.svc.GetByID(c.Request.Context(), wsID, sprintID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if sp.ReviewSnapshot == nil {
		middleware.AbortWithError(c, errs.New("SPRINT.NO_REVIEW", "迭代尚未完成，无复盘数据", 422))
		return
	}
	c.JSON(http.StatusOK, sp.ReviewSnapshot)
}

// suggestCapacity godoc
//
//	@Summary		速率建议（推荐容量）
//	@Description	基于近 N 期 completed 迭代的完成故事点统计
//	@Tags			sprint
//	@Produce		json
//	@Success		200		{object}	VelocityStats
//	@Router			/sprints/suggest-capacity [get]
func (h *Handler) suggestCapacity(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	stats, err := h.svc.SuggestCapacity(c.Request.Context(), wsID, projectID, []int{3, 6})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
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

// Ensure imports are used (for future-proofing, prevent compile errors).
var _ = strings.Join
