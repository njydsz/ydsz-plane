// Package metrics — 效能度量 HTTP Handler。
//
// 暴露项目级效能指标查询端点：速度、前置时间、质量、DORA 四指标。
// 前端可通过这些接口渲染仪表盘卡片（趋势图、DORA 雷达图、缺陷逃逸漏斗等）。
package metrics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps metrics handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// MetricsHandler Gin handler 集合。
type MetricsHandler struct {
	d *HandlerDeps
}

// NewMetricsHandler 构造 handler。
func NewMetricsHandler(d *HandlerDeps) *MetricsHandler {
	return &MetricsHandler{d: d}
}

// Register 注册效能度量路由（全部为只读，挂在 project 层级下）。
func (h *MetricsHandler) Register(r *gin.RouterGroup) {
	r.GET("/velocity", h.GetVelocity)
	r.GET("/velocity/trend", h.GetVelocityTrend)
	r.GET("/lead-time", h.GetLeadTime)
	r.GET("/quality", h.GetQualityMetrics)
	r.GET("/dora", h.GetDORA)
	r.GET("/resource-load", h.GetResourceLoad)
	r.POST("/deployments", h.RecordDeployment)
	r.GET("/snapshots", h.ListSnapshots)
}

// GetVelocity 查询项目速率统计。
//
//	@Summary		迭代速率
//	@Description	返回最近 N 个迭代完成的平均故事点数与趋势
//	@Tags			metrics
//	@Produce		json
//	@Param			last_n	query	int	false	"迭代数（默认 6，最大 20）"
//	@Success		200		{object}	VelocityResult
//	@Router			/metrics/velocity [get]
func (h *MetricsHandler) GetVelocity(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	lastN, _ := strconv.Atoi(c.DefaultQuery("last_n", "6"))

	result, err := h.d.Svc.GetVelocity(c.Request.Context(), wsID, projectID, lastN)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetVelocityTrend 查询迭代速率趋势（与 GetVelocity 相同渲染，仅作为独立 API 端点便于前端卡片绑定）。
func (h *MetricsHandler) GetVelocityTrend(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	lastN, _ := strconv.Atoi(c.DefaultQuery("last_n", "6"))

	result, err := h.d.Svc.GetVelocity(c.Request.Context(), wsID, projectID, lastN)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"project_id":   result.ProjectID,
		"trend":        result.Trend,
		"average":      result.Average,
		"sprint_count": result.SprintCount,
	})
}

// GetLeadTime 查询需求前置时间（P50/P85）。
//
//	@Summary		前置时间
//	@Tags			metrics
//	@Produce		json
//	@Param			days	query	int	false	"窗口天数（默认 90）"
//	@Success		200		{object}	LeadTimeResult
//	@Router			/metrics/lead-time [get]
func (h *MetricsHandler) GetLeadTime(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))

	result, err := h.d.Svc.GetLeadTime(c.Request.Context(), wsID, projectID, days)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetQualityMetrics 查询质量指标（缺陷密度、逃逸率、重开率）。
//
//	@Summary		质量指标
//	@Tags			metrics
//	@Produce		json
//	@Success		200	{object}	QualityMetrics
//	@Router			/metrics/quality [get]
func (h *MetricsHandler) GetQualityMetrics(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	result, err := h.d.Svc.GetQualityMetrics(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetDORA 查询 DORA 四指标。
//
//	@Summary		DORA 效能指标
//	@Description	返回部署频率（DF）、变更前置时间（LTC）、变更失败率（CFR）、平均恢复时间（MTTR）
//	@Tags			metrics
//	@Produce		json
//	@Success		200	{object}	DORAResult
//	@Router			/metrics/dora [get]
func (h *MetricsHandler) GetDORA(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	result, err := h.d.Svc.GetDORA(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// resourceLoadResponse 资源负载视图。
type resourceLoadResponse struct {
	ProjectID     int64 `json:"project_id"`
	ActiveWIP     int   `json:"active_wip"`
	TotalStarted  int   `json:"total_started_issues"`
	SampleSprintID int64 `json:"sample_sprint_id,omitempty"`
}

// GetResourceLoad 查询资源负载（进行中工作项 WIP）。
//
//	@Summary		资源负载
//	@Tags			metrics
//	@Produce		json
//	@Success		200	{object}	resourceLoadResponse
//	@Router			/metrics/resource-load [get]
func (h *MetricsHandler) GetResourceLoad(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	var wipCount int
	err := h.d.Svc.db.QueryRow(c.Request.Context(), `
		SELECT count(*) FROM issues i
		JOIN sprint_issues si ON si.issue_id = i.id
		JOIN sprints sp ON sp.id = si.sprint_id
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.workspace_id = $2 AND sp.status = 'active' AND st."group" = 'started'
			AND i.deleted_at IS NULL`,
		projectID, wsID).Scan(&wipCount)
	if err != nil {
		writeErr(c, errs.ErrInternal.Wrap(err))
		return
	}

	var startedCount int
	err = h.d.Svc.db.QueryRow(c.Request.Context(), `
		SELECT count(*) FROM issues i
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $1 AND i.workspace_id = $2 AND st."group" = 'started' AND i.deleted_at IS NULL`,
		projectID, wsID).Scan(&startedCount)
	if err != nil {
		writeErr(c, errs.ErrInternal.Wrap(err))
		return
	}

	c.JSON(http.StatusOK, resourceLoadResponse{
		ProjectID:    projectID,
		ActiveWIP:    wipCount,
		TotalStarted: startedCount,
	})
}

// deploymentRequest 部署事件上报请求。
type deploymentRequest struct {
	Env       string    `json:"env" binding:"required"`
	Status    string    `json:"status" binding:"required,oneof=success failed rollback"`
	Source    string    `json:"source" binding:"required"`
	CommitSHA string    `json:"commit_sha"`
	StartedAt time.Time `json:"started_at"`
	DeployedAt time.Time `json:"deployed_at"`
}

// RecordDeployment 接收 CI/CD 推送的部署事件（DORA 数据源）。
//
//	@Summary		记录部署事件
//	@Description	CI/CD 系统在部署完成后调用此接口记录事件（DORA 数据源）
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			body	body		deploymentRequest	true	"部署事件"
//	@Success		201		{object}	map[string]int64
//	@Router			/metrics/deployments [post]
func (h *MetricsHandler) RecordDeployment(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	var req deploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails([]errs.FieldDetail{
			{Field: "body", Reason: err.Error()},
		}))
		return
	}

	if req.DeployedAt.IsZero() {
		req.DeployedAt = time.Now()
	}

	id, err := h.d.Svc.RecordDeploymentEvent(c.Request.Context(), wsID, projectID,
		req.Env, req.Status, req.Source, req.CommitSHA, req.StartedAt, req.DeployedAt)
	if err != nil {
		writeErr(c, errs.ErrInternal.Wrap(err))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ListSnapshots 查询指标快照历史。
//
//	@Summary		指标快照列表
//	@Tags			metrics
//	@Produce		json
//	@Param			metric	query	string	false	"指标名"
//	@Success		200		{object}	[]map[string]any
//	@Router			/metrics/snapshots [get]
func (h *MetricsHandler) ListSnapshots(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	metric := c.Query("metric")

	var rows interface{}
	var err error
	_ = rows

	query := `SELECT metric, value, dimensions, snapshot_date FROM metric_snapshots
			  WHERE workspace_id = $1 AND project_id = $2`
	args := []any{wsID, projectID}
	if metric != "" {
		query += ` AND metric = $3`
		args = append(args, metric)
	}
	query += ` ORDER BY snapshot_date DESC LIMIT 90`

	// 通用的 map 扫描（简化）
	type snapshotRow struct {
		Metric   string  `json:"metric"`
		Value    float64 `json:"value"`
		Date     string  `json:"snapshot_date"`
	}
	_ = snapshotRow{}

	r, err := h.d.Svc.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		writeErr(c, errs.ErrInternal.Wrap(err))
		return
	}
	defer r.Close()

	var results []map[string]any
	for r.Next() {
		var m string
		var v float64
		var dims []byte
		var d string
		if err := r.Scan(&m, &v, &dims, &d); err != nil {
			continue
		}
		results = append(results, map[string]any{
			"metric":        m,
			"value":         v,
			"dimensions":    json.RawMessage(dims),
			"snapshot_date": d,
		})
	}
	c.JSON(http.StatusOK, results)
}

// writeErr 是 handler 内部统一的错误序列化 helper。
func writeErr(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		c.JSON(appErr.HTTP, appErr)
		return
	}
	c.JSON(http.StatusInternalServerError, errs.ErrInternal.WithMessage(err.Error()))
}
