// Package datajob — 异步数据任务 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:wid/data-jobs，提供任务创建、进度查询、
// 状态流转（运行/成功/失败/取消）与列表监控端点。
package datajob

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 数据任务 handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// RegisterRead 注册只读路由（列表 / 详情）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/data-jobs")
	g.GET("", h.ListJobs)
	g.GET("/:job_id", h.GetJob)
}

// RegisterWrite 注册写路由（创建 / 进度 / 状态流转 / 删除）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/data-jobs")
	g.POST("", h.CreateJob)
	g.PATCH("/:job_id/progress", h.UpdateProgress)
	g.POST("/:job_id/run", h.MarkRunning)
	g.POST("/:job_id/success", h.MarkSuccess)
	g.POST("/:job_id/fail", h.MarkFailed)
	g.POST("/:job_id/cancel", h.CancelJob)
	g.DELETE("/:job_id", h.DeleteJob)
}

// Register 注册全部路由，供单一权限层级场景使用。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// ListJobs godoc
//
//	@Summary		数据任务列表
//	@Description	按工作空间 + 可选项目/类型/状态过滤查询异步任务
//	@Tags			datajob
//	@Produce		json
//	@Param			job_type	query	string	false	"任务类型"
//	@Param			status		query	string	false	"任务状态"
//	@Param			limit		query	int		false	"每页条数，默认 50"
//	@Param			offset		query	int		false	"偏移量，默认 0"
//	@Success		200	{object}	map[string]any
//	@Router			/data-jobs [get]
func (h *Handler) ListJobs(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	filter := ListJobsFilter{
		WorkspaceID: wsID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if projectID > 0 {
		filter.ProjectID = &projectID
	}
	if jt := c.Query("job_type"); jt != "" {
		filter.JobType = jt
	}
	if st := c.Query("status"); st != "" {
		filter.Status = JobStatus(st)
	}

	jobs, err := h.d.Svc.List(c.Request.Context(), filter)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": jobs})
}

// GetJob godoc
//
//	@Summary		数据任务详情
//	@Tags			datajob
//	@Produce		json
//	@Success		200	{object}	DataJob
//	@Router			/data-jobs/{job_id} [get]
func (h *Handler) GetJob(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("job_id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "job_id", Reason: "invalid"},
		))
		return
	}
	job, err := h.d.Svc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

// CreateJob godoc
//
//	@Summary		创建数据任务
//	@Description	默认 pending 状态；同工作空间同名任务不允许重复
//	@Tags			datajob
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	DataJob
//	@Router			/data-jobs [post]
func (h *Handler) CreateJob(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		JobType   string         `json:"job_type" binding:"required"`
		Name      string         `json:"name" binding:"required"`
		Parameters map[string]any `json:"parameters"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}
	uid := userID

	job, err := h.d.Svc.Create(c.Request.Context(), CreateJobInput{
		TenantID:    1,
		WorkspaceID: wsID,
		ProjectID:   projID,
		JobType:     req.JobType,
		Name:        req.Name,
		Parameters:  req.Parameters,
		TriggeredBy: &uid,
		CreatedBy:   userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

// UpdateProgress godoc
//
//	@Summary		更新任务进度
//	@Description	仅 pending/running 状态允许更新；progress 取值 0-100
//	@Tags			datajob
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/data-jobs/{job_id}/progress [patch]
func (h *Handler) UpdateProgress(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	var req struct {
		Progress int `json:"progress" binding:"min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	if err := h.d.Svc.UpdateProgress(c.Request.Context(), id, req.Progress); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkRunning godoc
//
//	@Summary		标记任务开始执行
//	@Tags			datajob
//	@Success		204
//	@Router			/data-jobs/{job_id}/run [post]
func (h *Handler) MarkRunning(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	if err := h.d.Svc.MarkRunning(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkSuccess godoc
//
//	@Summary		标记任务成功
//	@Tags			datajob
//	@Success		204
//	@Router			/data-jobs/{job_id}/success [post]
func (h *Handler) MarkSuccess(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	if err := h.d.Svc.MarkSuccess(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkFailed godoc
//
//	@Summary		标记任务失败
//	@Tags			datajob
//	@Accept			json
//	@Success		204
//	@Router			/data-jobs/{job_id}/fail [post]
func (h *Handler) MarkFailed(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	var req struct {
		ErrorMessage string `json:"error_message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	if err := h.d.Svc.MarkFailed(c.Request.Context(), id, req.ErrorMessage); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CancelJob godoc
//
//	@Summary		取消任务
//	@Description	仅 pending/running 状态可取消
//	@Tags			datajob
//	@Success		204
//	@Router			/data-jobs/{job_id}/cancel [post]
func (h *Handler) CancelJob(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)
	if err := h.d.Svc.Cancel(c.Request.Context(), id, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteJob godoc
//
//	@Summary		删除数据任务
//	@Description	软删除语义
//	@Tags			datajob
//	@Success		204
//	@Router			/data-jobs/{job_id} [delete]
func (h *Handler) DeleteJob(c *gin.Context) {
	id, err := parseJobID(c)
	if err != nil {
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)
	if err := h.d.Svc.Delete(c.Request.Context(), id, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Helpers ---

func parseJobID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("job_id"), 10, 64)
	if err != nil || id <= 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "job_id", Reason: "invalid"},
		))
		return 0, err
	}
	return id, nil
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
