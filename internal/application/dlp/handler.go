// Package dlp — HTTP handlers。
package dlp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册路由（工作空间级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/dlp")
	g.GET("/findings", h.ListFindings)
	g.GET("/jobs", h.ListJobs)
	g.GET("/jobs/:job_id", h.GetJob)
	g.GET("/jobs/:job_id/report", h.GetReport)
	g.POST("/jobs", h.CreateJob)
	g.POST("/jobs/:job_id/run", h.RunJob)
	g.POST("/findings/:id/resolve", h.ResolveFinding)
	g.POST("/check", h.Check)
	g.GET("/rules", h.ListRules)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrJobNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		if ec, ok := err.(interface{ HTTPStatus() int }); ok {
			c.JSON(ec.HTTPStatus(), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func intParam(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// ListFindings godoc
//
//	@Summary		敏感数据发现列表
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/dlp/findings [get]
func (h *Handler) ListFindings(c *gin.Context) {
	category := Category(c.Query("category"))
	severity := Severity(c.Query("severity"))
	action := c.Query("action")
	items, err := h.d.Svc.ListFindings(c.Request.Context(),
		c.GetInt64(middleware.CtxWorkspaceID), category, severity, action, intQuery(c, "limit", 100))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ListJobs godoc
//
//	@Summary		扫描任务列表
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/dlp/jobs [get]
func (h *Handler) ListJobs(c *gin.Context) {
	items := h.d.Svc.ListJobs(c.GetInt64(middleware.CtxWorkspaceID))
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetJob godoc
//
//	@Summary		扫描任务详情
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	Job
//	@Router			/dlp/jobs/{job_id} [get]
func (h *Handler) GetJob(c *gin.Context) {
	job, err := h.d.Svc.GetJob(intParam(c, "job_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

// GetReport godoc
//
//	@Summary		扫描报告
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	ScanReport
//	@Router			/dlp/jobs/{job_id}/report [get]
func (h *Handler) GetReport(c *gin.Context) {
	report, err := h.d.Svc.GetReport(c.Request.Context(), intParam(c, "job_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

// CreateJob godoc
//
//	@Summary		创建扫描任务
//	@Tags			dlp
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Job
//	@Router			/dlp/jobs [post]
func (h *Handler) CreateJob(c *gin.Context) {
	var body struct {
		Scope    string `json:"scope"`
		TargetID int64  `json:"target_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.d.Svc.CreateJob(c.GetInt64(middleware.CtxWorkspaceID), body.Scope,
		body.TargetID, c.GetInt64(middleware.CtxUserID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

// RunJob godoc
//
//	@Summary		执行扫描
//	@Tags			dlp
//	@Produce		json
//	@Success		202	{object}	Job
//	@Router			/dlp/jobs/{job_id}/run [post]
func (h *Handler) RunJob(c *gin.Context) {
	job, err := h.d.Svc.GetJob(intParam(c, "job_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	h.d.Svc.RunScan(c.Request.Context(), job.ID)
	c.JSON(http.StatusAccepted, job)
}

// ResolveFinding godoc
//
//	@Summary		标记发现已处理
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/dlp/findings/{id}/resolve [post]
func (h *Handler) ResolveFinding(c *gin.Context) {
	if err := h.d.Svc.ResolveFinding(c.Request.Context(),
		intParam(c, "id"), c.GetInt64(middleware.CtxUserID)); err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Check godoc
//
//	@Summary		写入前检查（拦截/告警）
//	@Tags			dlp
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/dlp/check [post]
func (h *Handler) Check(c *gin.Context) {
	var body struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	action, matches := h.d.Svc.Check(body.Text, DefaultPolicy)
	c.JSON(http.StatusOK, gin.H{
		"action":  action,
		"matches": matches,
		"count":   len(matches),
	})
}

// ListRules godoc
//
//	@Summary		检测规则列表
//	@Tags			dlp
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/dlp/rules [get]
func (h *Handler) ListRules(c *gin.Context) {
	rules := h.d.Svc.Detector().Rules()
	items := make([]map[string]any, 0, len(rules))
	for _, r := range rules {
		items = append(items, map[string]any{
			"name":       r.Name,
			"category":   r.Category,
			"severity":   r.Severity,
			"confidence": r.Confidence,
			"enabled":    r.Enabled,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
