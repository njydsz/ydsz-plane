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

// Handler 是 MetricsHandler 的类型别名，供 CachedHandler 嵌入使用。
type Handler = MetricsHandler

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
	r.GET("/resource-load/detail", h.GetResourceLoadDetail)
	r.POST("/deployments", h.RecordDeployment)
	r.GET("/snapshots", h.ListSnapshots)

	// P1: 高级效能指标（CFD/控制图/周吞吐量）
	r.GET("/cfd", h.GetCFD)
	r.GET("/control-chart", h.GetControlChart)
	r.GET("/throughput", h.GetWeeklyThroughput)
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
	ProjectID      int64 `json:"project_id"`
	ActiveWIP      int   `json:"active_wip"`
	TotalStarted   int   `json:"total_started_issues"`
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
		SELECT count(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
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
		SELECT count(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM requirement WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM task WHERE deleted_at IS NULL UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, external_id, created_by, created_at, updated_at, deleted_at FROM defect WHERE deleted_at IS NULL) AS w i
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
	Env        string    `json:"env" binding:"required"`
	Status     string    `json:"status" binding:"required,oneof=success failed rollback"`
	Source     string    `json:"source" binding:"required"`
	CommitSHA  string    `json:"commit_sha"`
	StartedAt  time.Time `json:"started_at"`
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
		writeErr(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "body", Reason: err.Error()},
		))
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

	query := `SELECT metric, value, dimensions, snapshot_date FROM metric_snapshots
			  WHERE workspace_id = $1 AND project_id = $2`
	args := []any{wsID, projectID}
	if metric != "" {
		query += ` AND metric = $3`
		args = append(args, metric)
	}
	query += ` ORDER BY snapshot_date DESC LIMIT 90`

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
	c.JSON(http.StatusInternalServerError, errs.ErrInternal.Wrap(err))
}

// --- P1: Advanced Metrics Handlers ---

// cfdQuery CFD 查询参数。
type cfdQuery struct {
	Days int `form:"days" binding:"max=365"`
}

// GetCFD 查询项目累积流图数据。
//
//	@Summary		累积流图（CFD）
//	@Description	按日期分桶统计各状态组工作项数量，用于绘制堆叠面积图
//	@Tags			metrics
//	@Produce		json
//	@Param			days	query	int	false	"天数（默认 30，最大 365）"
//	@Success		200		{object}	[]CFDDataPoint
//	@Router			/metrics/cfd [get]
func (h *MetricsHandler) GetCFD(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	var q cfdQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "days", Reason: "必须在 1-365 之间"},
		))
		return
	}
	if q.Days <= 0 {
		q.Days = 30
	}

	fromDate := time.Now().AddDate(0, 0, -q.Days)
	toDate := time.Now()

	calc := NewCFDCalculator(h.d.Svc.db)
	points, err := calc.Calculate(c.Request.Context(), wsID, projectID, fromDate, toDate)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, points)
}

// GetControlChart 查询前置时间控制图数据。
//
//	@Summary		前置时间控制图
//	@Description	按完成日期排序的前置时间散点图 + P50/P85/P95 控制线
//	@Tags			metrics
//	@Produce		json
//	@Param			days	query	int	false	"天数（默认 90）"
//	@Success		200		{object}	ControlChartResult
//	@Router			/metrics/control-chart [get]
func (h *MetricsHandler) GetControlChart(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	if days > 365 {
		days = 365
	}

	calc := NewControlChartCalculator(h.d.Svc.db)
	result, err := calc.Calculate(c.Request.Context(), wsID, projectID, days)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetWeeklyThroughput 查询周吞吐量。
//
//	@Summary		周吞吐量
//	@Description	近 N 周每周完成的需求数与故事点数
//	@Tags			metrics
//	@Produce		json
//	@Param			weeks	query	int	false	"周数（默认 12）"
//	@Success		200		{object}	[]WeeklyThroughput
//	@Router			/metrics/throughput [get]
func (h *MetricsHandler) GetWeeklyThroughput(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	weeks, _ := strconv.Atoi(c.DefaultQuery("weeks", "12"))

	result, err := h.d.Svc.GetWeeklyThroughput(c.Request.Context(), wsID, projectID, weeks)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetResourceLoadDetail 查询项目成员级资源负载明细。
//
//	@Summary		资源负载明细
//	@Description	每位成员当前 WIP 数量与故事点负载
//	@Tags			metrics
//	@Produce		json
//	@Success		200	{object}	ResourceLoadDetail
//	@Router			/metrics/resource-load/detail [get]
func (h *MetricsHandler) GetResourceLoadDetail(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	result, err := h.d.Svc.GetResourceLoadDetail(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
