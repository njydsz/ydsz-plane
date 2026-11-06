// Package marketplace — HTTP handlers。
package marketplace

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	g := r.Group("/marketplace")
	// 目录（只读）
	g.GET("/categories", h.ListCategories)
	g.GET("/apps", h.ListApps)
	g.GET("/apps/:slug", h.GetApp)
	g.GET("/apps/:id/reviews", h.ListReviews)
	// 安装（写）
	g.GET("/installed", h.ListInstalled)
	g.POST("/apps/:id/install", h.Install)
	g.DELETE("/apps/:id/install", h.Uninstall)
	g.POST("/apps/:id/reviews", h.CreateReview)
	// 发布（实例管理员或官方）
	g.POST("/apps", h.CreateApp)
	g.PATCH("/apps/:slug", h.UpdateApp)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrAppNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrAppSlugTaken:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrInvalidRating:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// ListCategories godoc
//
//	@Summary		应用分类
//	@Tags			marketplace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/marketplace/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.d.Svc.ListCategories(c.Request.Context())
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ListApps godoc
//
//	@Summary		应用目录
//	@Tags			marketplace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/marketplace/apps [get]
func (h *Handler) ListApps(c *gin.Context) {
	cat := Category(c.Query("category"))
	if cat != "" && !IsValidCategory(string(cat)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category"})
		return
	}
	items, err := h.d.Svc.List(c.Request.Context(), cat, c.Query("q"),
		intQuery(c, "limit", 50), intQuery(c, "offset", 0))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetApp godoc
//
//	@Summary		应用详情
//	@Tags			marketplace
//	@Produce		json
//	@Success		200	{object}	App
//	@Router			/marketplace/apps/{slug} [get]
func (h *Handler) GetApp(c *gin.Context) {
	app, err := h.d.Svc.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

// CreateApp godoc
//
//	@Summary		发布应用
//	@Tags			marketplace
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	App
//	@Router			/marketplace/apps [post]
func (h *Handler) CreateApp(c *gin.Context) {
	var body struct {
		Slug          string         `json:"slug"`
		Name          string         `json:"name"`
		Description   string         `json:"description"`
		Category      string         `json:"category"`
		Publisher     string         `json:"publisher"`
		PublisherURL  string         `json:"publisher_url"`
		IconURL       string         `json:"icon_url"`
		Screenshots   []string       `json:"screenshots"`
		Tags          []string       `json:"tags"`
		Pricing       string         `json:"pricing"`
		InstallType   string         `json:"install_type"`
		InstallConfig map[string]any `json:"install_config"`
		DocsURL       string         `json:"docs_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.d.Svc.Create(c.Request.Context(), CreateAppInput{
		Slug:          body.Slug,
		Name:          body.Name,
		Description:   body.Description,
		Category:      Category(body.Category),
		Publisher:     body.Publisher,
		PublisherURL:  body.PublisherURL,
		IconURL:       body.IconURL,
		Screenshots:   body.Screenshots,
		Tags:          body.Tags,
		Pricing:       Pricing(body.Pricing),
		InstallType:   InstallType(body.InstallType),
		InstallConfig: body.InstallConfig,
		DocsURL:       body.DocsURL,
		CreatedBy:     c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, app)
}

// UpdateApp godoc
//
//	@Summary		更新应用
//	@Tags			marketplace
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	App
//	@Router			/marketplace/apps/{slug} [patch]
func (h *Handler) UpdateApp(c *gin.Context) {
	var body struct {
		Name          *string        `json:"name"`
		Description   *string        `json:"description"`
		PublisherURL  *string        `json:"publisher_url"`
		IconURL       *string        `json:"icon_url"`
		Screenshots   []string       `json:"screenshots"`
		Tags          []string       `json:"tags"`
		Pricing       *string        `json:"pricing"`
		InstallConfig map[string]any `json:"install_config"`
		DocsURL       *string        `json:"docs_url"`
		Status        *string        `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var pricing *Pricing
	if body.Pricing != nil {
		p := Pricing(*body.Pricing)
		pricing = &p
	}
	var status *Status
	if body.Status != nil {
		s := Status(*body.Status)
		status = &s
	}
	app, err := h.d.Svc.Update(c.Request.Context(), c.Param("slug"), UpdateAppInput{
		Name:          body.Name,
		Description:   body.Description,
		PublisherURL:  body.PublisherURL,
		IconURL:       body.IconURL,
		Screenshots:   body.Screenshots,
		Tags:          body.Tags,
		Pricing:       pricing,
		InstallConfig: body.InstallConfig,
		DocsURL:       body.DocsURL,
		Status:        status,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

// Install godoc
//
//	@Summary		安装应用
//	@Tags			marketplace
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Install
//	@Router			/marketplace/apps/{id}/install [post]
func (h *Handler) Install(c *gin.Context) {
	var body struct {
		Config map[string]any `json:"config"`
	}
	_ = c.ShouldBindJSON(&body)
	inst, err := h.d.Svc.Install(c.Request.Context(), InstallInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		AppID:       intParam(c, "id"),
		InstalledBy: c.GetInt64(middleware.CtxUserID),
		Config:      body.Config,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, inst)
}

// Uninstall godoc
//
//	@Summary		卸载应用
//	@Tags			marketplace
//	@Produce		json
//	@Success		204
//	@Router			/marketplace/apps/{id}/install [delete]
func (h *Handler) Uninstall(c *gin.Context) {
	if err := h.d.Svc.Uninstall(c.Request.Context(),
		c.GetInt64(middleware.CtxWorkspaceID), intParam(c, "id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListInstalled godoc
//
//	@Summary		已安装应用
//	@Tags			marketplace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/marketplace/installed [get]
func (h *Handler) ListInstalled(c *gin.Context) {
	items, err := h.d.Svc.ListInstalled(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateReview godoc
//
//	@Summary		提交评分
//	@Tags			marketplace
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Review
//	@Router			/marketplace/apps/{id}/reviews [post]
func (h *Handler) CreateReview(c *gin.Context) {
	var body struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.d.Svc.CreateReview(c.Request.Context(), ReviewInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		AppID:       intParam(c, "id"),
		UserID:      c.GetInt64(middleware.CtxUserID),
		Rating:      body.Rating,
		Comment:     body.Comment,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

// ListReviews godoc
//
//	@Summary		应用评分列表
//	@Tags			marketplace
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/marketplace/apps/{id}/reviews [get]
func (h *Handler) ListReviews(c *gin.Context) {
	items, err := h.d.Svc.ListReviews(c.Request.Context(), intParam(c, "id"), intQuery(c, "limit", 20))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// --- 内置种子 ---

// SeedBuiltins 返回内置应用目录（首次启动时写入）。
func SeedBuiltins() []CreateAppInput {
	return []CreateAppInput{
		{Slug: "github-integration", Name: "GitHub 集成", Category: CategoryIntegrations,
			Description: "连接 GitHub 仓库，同步 Issue 与 PR 状态", InstallType: InstallOAuth,
			Tags: []string{"github", "vcs", "sync"}, IconURL: "github"},
		{Slug: "gitlab-integration", Name: "GitLab 集成", Category: CategoryIntegrations,
			Description: "连接 GitLab 项目，同步 Issue 与 MR", InstallType: InstallOAuth,
			Tags: []string{"gitlab", "vcs", "sync"}, IconURL: "gitlab"},
		{Slug: "slack-notify", Name: "Slack 通知", Category: CategoryIntegrations,
			Description: "工作项变更推送到 Slack 频道", InstallType: InstallWebhook,
			Tags: []string{"slack", "notification"}, IconURL: "slack"},
		{Slug: "sentry-tracker", Name: "Sentry 错误跟踪", Category: CategoryIntegrations,
			Description: "将 Sentry Issue 关联为缺陷工作项", InstallType: InstallOAuth,
			Tags: []string{"sentry", "error"}, IconURL: "sentry"},
		{Slug: "jira-importer", Name: "Jira 导入器", Category: CategoryImporters,
			Description: "从 Jira 迁移项目、工作项、附件与评论", InstallType: InstallCLI,
			Tags: []string{"jira", "migration"}, IconURL: "download"},
		{Slug: "linear-importer", Name: "Linear 导入器", Category: CategoryImporters,
			Description: "从 Linear 迁移 Issue 与 Cycle", InstallType: InstallCLI,
			Tags: []string{"linear", "migration"}, IconURL: "download"},
		{Slug: "asana-importer", Name: "Asana 导入器", Category: CategoryImporters,
			Description: "从 Asana 迁移任务与项目", InstallType: InstallCLI,
			Tags: []string{"asana", "migration"}, IconURL: "download"},
		{Slug: "csv-importer", Name: "CSV 导入器", Category: CategoryImporters,
			Description: "通用 CSV/Excel 导入，支持自定义列映射", InstallType: InstallBuiltin,
			Tags: []string{"csv", "excel", "migration"}, IconURL: "upload"},
		{Slug: "scrum-template", Name: "Scrum 项目模板", Category: CategoryTemplates,
			Description: "含 Sprint 流程、状态机与估算点的项目模板", InstallType: InstallBuiltin,
			Tags: []string{"scrum", "agile"}, IconURL: "layout-template"},
		{Slug: "kanban-template", Name: "看板项目模板", Category: CategoryTemplates,
			Description: "轻量级看板流程，适合运维与支持团队", InstallType: InstallBuiltin,
			Tags: []string{"kanban", "ops"}, IconURL: "layout-template"},
		{Slug: "defect-tracking", Name: "缺陷跟踪套件", Category: CategoryApps,
			Description: "缺陷分析、严重度流转与质量门禁", InstallType: InstallBuiltin,
			Tags: []string{"defect", "quality"}, IconURL: "bug"},
		{Slug: "dora-metrics", Name: "DORA 效能度量", Category: CategoryApps,
			Description: "部署频率、变更前置时间、MTTR 与变更失败率", InstallType: InstallBuiltin,
			Tags: []string{"dora", "metrics"}, IconURL: "chart-line"},
		{Slug: "ai-assistant", Name: "AI 助手", Category: CategoryApps,
			Description: "Ask/Build/Autopilot 三种模式的 AI 工作助手", InstallType: InstallBuiltin,
			Tags: []string{"ai", "agent"}, IconURL: "bot"},
		{Slug: "pql-console", Name: "PQL 查询控制台", Category: CategoryApps,
			Description: "声明式查询语言与可视化构建器", InstallType: InstallBuiltin,
			Tags: []string{"pql", "search"}, IconURL: "search-code"},
	}
}

// SeedIfEmpty 在目录为空时写入内置应用。
func (s *Service) SeedIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM marketplace_apps`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, in := range SeedBuiltins() {
		if _, err := s.Create(ctx, in); err != nil && err != ErrAppSlugTaken {
			return err
		}
	}
	return nil
}

var _ = json.Marshal
var _ = strings.TrimSpace
