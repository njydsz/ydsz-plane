// Package importer — HTTP handlers。
package importer

import (
	"encoding/json"
	"io"
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

// Register 注册路由（项目级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/imports")
	g.GET("/jobs", h.ListJobs)
	g.POST("/jobs", h.CreateJob)
	g.GET("/jobs/:job_id", h.GetJob)
	g.POST("/jobs/:job_id/csv", h.ImportCSV)
	g.POST("/jobs/:job_id/platform", h.ImportPlatform)
	g.GET("/sources", h.ListSources)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrJobNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUnsupportedSource:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
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

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// ListSources godoc
//
//	@Summary		支持的导入来源
//	@Tags			importer
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/imports/sources [get]
func (h *Handler) ListSources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sources": []map[string]any{
			{"id": "csv", "name": "CSV", "description": "通用 CSV 导入，支持列映射", "needs_credentials": false},
			{"id": "jira", "name": "Jira", "description": "Atlassian Jira Cloud / Data Center", "needs_credentials": true,
				"fields": []string{"base_url", "email", "api_token", "project_key"}},
			{"id": "linear", "name": "Linear", "description": "Linear GraphQL API", "needs_credentials": true,
				"fields": []string{"api_token", "team_id"}},
			{"id": "asana", "name": "Asana", "description": "Asana REST API", "needs_credentials": true,
				"fields": []string{"api_token", "team_id"}},
		},
	})
}

// ListJobs godoc
//
//	@Summary		导入任务列表
//	@Tags			importer
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/imports/jobs [get]
func (h *Handler) ListJobs(c *gin.Context) {
	items := h.d.Svc.ListJobs(c.GetInt64(middleware.CtxWorkspaceID), c.GetInt64(middleware.CtxProjectID))
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateJob godoc
//
//	@Summary		创建导入任务
//	@Tags			importer
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Job
//	@Router			/imports/jobs [post]
func (h *Handler) CreateJob(c *gin.Context) {
	var body struct {
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := h.d.Svc.CreateJob(
		c.GetInt64(middleware.CtxWorkspaceID), c.GetInt64(middleware.CtxProjectID),
		c.GetInt64(middleware.CtxUserID), Source(body.Source))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

// GetJob godoc
//
//	@Summary		导入任务详情
//	@Tags			importer
//	@Produce		json
//	@Success		200	{object}	Job
//	@Router			/imports/jobs/{job_id} [get]
func (h *Handler) GetJob(c *gin.Context) {
	job, err := h.d.Svc.GetJob(paramInt(c, "job_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

// ImportCSV godoc
//
//	@Summary		CSV 导入
//	@Tags			importer
//	@Accept			application/octet-stream
//	@Produce		json
//	@Success		200	{object}	ImportResult
//	@Router			/imports/jobs/{job_id}/csv [post]
func (h *Handler) ImportCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传 CSV 文件"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	data, err := io.ReadAll(io.LimitReader(src, 32<<20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	opts := ImportOptions{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		ProjectID:   c.GetInt64(middleware.CtxProjectID),
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
		DefaultType: c.PostForm("default_type"),
		BatchSize:   intQuery(c, "batch_size", 100),
	}
	result, err := h.d.Svc.RunCSV(c.Request.Context(), paramInt(c, "job_id"), data, opts)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ImportPlatform godoc
//
//	@Summary		平台导入（Jira/Linear/Asana）
//	@Tags			importer
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ImportResult
//	@Router			/imports/jobs/{job_id}/platform [post]
func (h *Handler) ImportPlatform(c *gin.Context) {
	var body struct {
		Credentials Credentials   `json:"credentials"`
		Options     ImportOptions `json:"options"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 覆盖工作空间/项目（以路径为准，防止越权）
	body.Options.WorkspaceID = c.GetInt64(middleware.CtxWorkspaceID)
	body.Options.ProjectID = c.GetInt64(middleware.CtxProjectID)
	body.Options.CreatedBy = c.GetInt64(middleware.CtxUserID)
	if body.Options.BatchSize <= 0 {
		body.Options.BatchSize = 100
	}

	result, err := h.d.Svc.RunPlatform(c.Request.Context(), paramInt(c, "job_id"), body.Credentials, body.Options)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

var _ = json.Marshal
