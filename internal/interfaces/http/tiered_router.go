// Package httpapi — 三层 API 路由物理分离（对标 Plane 的多路由层架构）。
//
// 三层路径前缀 + 差异化凭证解析策略：
//   - /_session/*  → 浏览器 SPA（JWT + Cookie）
//   - /_apikeys/*  → 程序化访问（API Key）
//   - /_public/*   → 公开端点（Webhook / 健康 / 邀请预览）
//
// 这些路由与现有 /api/v1/* 并行存在；客户端可按需选择接入层。
package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// RegisterTieredRoutes 注册三层差异化路由。
// 调用方需在 NewEngine 之后调用；可选择仅在 d.Cfg.EnableTieredRoutes 为 true 时启用。
func RegisterTieredRoutes(r *gin.Engine, d *Deps) {
	// 共享的项目级中间件链（不绑定 auth，各自层自行注入）。
	projectScoped := func(base *gin.RouterGroup) *gin.RouterGroup {
		g := base.Group("/workspaces/:workspace_id/projects/:project_id")
		g.Use(middleware.RequireWorkspaceParam(), middleware.RequireProjectParam())
		return g
	}

	// ---- Session 层（浏览器 SPA） ----
	// 注意：中间件顺序 — RequireWorkspaceParam 必须在 RequirePermissionFromDB 之前，
	// 否则权限检查时 CtxWorkspaceID 尚未注入，会直接 403。
	session := r.Group("/_session")
	session.Use(
		middleware.SessionAuth(d.Cfg, d.Auth, d.Redis),
		middleware.RequireWorkspaceParam(),
		middleware.RequirePermissionFromDB(d.RBACStore, auth.PermWorkspaceRead),
	)
	if d.IssueHandler != nil {
		// Session 层对接现有 handler（Issue / Sprint / Module / Version）。
		issueGroup := projectScoped(session)
		d.IssueHandler.Register(issueGroup, nil, nil)
	}
	{
		sprintGroup := projectScoped(session)
		if d.SprintHandler != nil {
			d.SprintHandler.Register(sprintGroup)
		}
	}
	{
		prefGroup := projectScoped(session)
		if d.PrefHandler != nil {
			d.PrefHandler.Register(prefGroup)
		}
	}

	// ---- API Key 层（程序化访问）----
	apikeys := r.Group("/_apikeys")
	apikeys.Use(
		middleware.APIKeyAuth(d.ApiTokenSvc, d.Auth, d.Redis),
		middleware.RequireAPIScope(auth.PermWorkspaceRead),
	)
	if d.IssueHandler != nil {
		// API Key 层同样可用 Issue Handler（Scope 替代 RBAC 鉴权）。
		issueGroup := projectScoped(apikeys)
		d.IssueHandler.Register(issueGroup, nil, nil)
	}
	{
		sprintGroup := projectScoped(apikeys)
		if d.SprintHandler != nil {
			d.SprintHandler.Register(sprintGroup)
		}
	}

	// ---- Public 层（完全公开，无鉴权）----
	public := r.Group("/_public")
	public.Use(middleware.AnonymousSession())
	{
		public.GET("/healthz", healthz())
		public.GET("/readyz", readyz(d))

		// 邀请预览（公开只读，无需登录）
		public.GET("/invitations/:token", getInvitationPreview(d))

		// Webhook 接收端点
		if d.IntakePublicHandler != nil {
			public.GET("/track", d.IntakePublicHandler.TrackIssue)
		}
	}
}
