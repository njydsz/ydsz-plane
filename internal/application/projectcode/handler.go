// Package projectcode — Projects-as-Code HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:workspace_id/project-code。
//
// 端点：
//   - GET    /configs              列出已保存配置
//   - POST   /configs              保存 YAML 配置
//   - GET    /configs/:id          配置详情
//   - DELETE /configs/:id          删除配置
//   - POST   /parse                校验并返回解析后的规范（不保存）
//   - POST   /diff                 规范 vs 既有项目 diff
//   - POST   /apply                应用规范（创建/更新项目）
//   - GET    /export/:project_id   导出既有项目为 YAML
package projectcode

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
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
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/project-code")
	g.GET("/configs", h.ListConfigs)
	g.POST("/configs", h.SaveConfig)
	g.GET("/configs/:id", h.GetConfig)
	g.DELETE("/configs/:id", h.DeleteConfig)
	g.POST("/parse", h.Parse)
	g.POST("/diff", h.Diff)
	g.POST("/apply", h.Apply)
	g.GET("/export/:project_id", h.Export)
}

// --- helpers ---

// fail 统一错误响应：AppError 携带 HTTP 状态码，其余按 500。
func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrConfigNotFound), errors.Is(err, ErrProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidYAML):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		var appErr *errs.AppError
		if errs.As(err, &appErr) {
			c.JSON(appErr.HTTP, gin.H{"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Details,
			}})
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

func int64Query(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return n
}

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// --- handlers ---

// ListConfigs godoc
//
//	@Summary		列出项目配置
//	@Tags			project-code
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/project-code/configs [get]
func (h *Handler) ListConfigs(c *gin.Context) {
	items, err := h.d.Svc.ListConfigs(c.Request.Context(), ListConfigsOptions{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		ProjectID:   int64Query(c, "project_id"),
		Limit:       intQuery(c, "limit", 100),
		Offset:      intQuery(c, "offset", 0),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetConfig godoc
//
//	@Summary		配置详情
//	@Tags			project-code
//	@Produce		json
//	@Success		200	{object}	Config
//	@Router			/project-code/configs/{id} [get]
func (h *Handler) GetConfig(c *gin.Context) {
	cfg, err := h.d.Svc.GetConfig(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), paramInt(c, "id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// SaveConfig godoc
//
//	@Summary		保存 YAML 配置
//	@Tags			project-code
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Config
//	@Router			/project-code/configs [post]
func (h *Handler) SaveConfig(c *gin.Context) {
	var body struct {
		Name      string `json:"name"`
		ProjectID int64  `json:"project_id"`
		YAML      string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.d.Svc.SaveConfig(c.Request.Context(), SaveConfigInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		ProjectID:   body.ProjectID,
		Name:        body.Name,
		YAML:        body.YAML,
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
		UpdatedBy:   c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, cfg)
}

// DeleteConfig godoc
//
//	@Summary		删除配置
//	@Tags			project-code
//	@Produce		json
//	@Success		204
//	@Router			/project-code/configs/{id} [delete]
func (h *Handler) DeleteConfig(c *gin.Context) {
	if err := h.d.Svc.DeleteConfig(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), paramInt(c, "id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Parse godoc
//
//	@Summary		校验并解析 YAML（不保存）
//	@Tags			project-code
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ProjectSpec
//	@Router			/project-code/parse [post]
func (h *Handler) Parse(c *gin.Context) {
	var body struct {
		YAML string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	spec, err := ParseSpec(body.YAML)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, spec)
}

// Diff godoc
//
//	@Summary		规范 vs 既有项目 diff
//	@Tags			project-code
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	DiffResult
//	@Router			/project-code/diff [post]
func (h *Handler) Diff(c *gin.Context) {
	var body struct {
		ProjectID int64  `json:"project_id"`
		YAML      string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ProjectID <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "project_id 不能为空"})
		return
	}
	spec, err := ParseSpec(body.YAML)
	if err != nil {
		h.fail(c, err)
		return
	}
	res, err := h.d.Svc.DiffConfig(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), body.ProjectID, spec)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Apply godoc
//
//	@Summary		应用规范（创建/更新项目）
//	@Tags			project-code
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ApplyResult
//	@Router			/project-code/apply [post]
func (h *Handler) Apply(c *gin.Context) {
	var body struct {
		ProjectID int64  `json:"project_id"`
		YAML      string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	spec, err := ParseSpec(body.YAML)
	if err != nil {
		h.fail(c, err)
		return
	}
	res, err := h.d.Svc.ApplySpec(c.Request.Context(), ApplySpecInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		ProjectID:   body.ProjectID,
		Spec:        spec,
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
		UpdatedBy:   c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Export godoc
//
//	@Summary		导出既有项目为 YAML 规范
//	@Tags			project-code
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/project-code/export/{project_id} [get]
func (h *Handler) Export(c *gin.Context) {
	spec, err := h.d.Svc.ExportProject(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), paramInt(c, "project_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	out, err := MarshalYAML(spec)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"spec": spec,
		"yaml": string(out),
	})
}
