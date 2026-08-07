// Package httpapi 装配 Gin 引擎、中间件链与路由表。
package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/application/attachment"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/application/dashboard"
	"github.com/njydsz/ydsz-plane/internal/application/issue"
	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/preference"
	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/version"
	"github.com/njydsz/ydsz-plane/internal/application/workbench"
	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Deps 携带各 handler 的依赖。
type Deps struct {
	Cfg      *config.Config
	Log      *zap.Logger
	DB       *pgxpool.Pool
	Redis    *redis.Client
	Auth     *auth.Service
	ResetSvc *auth.PasswordResetService
	// PrincipalParser 解析访问凭证（JWT 或 API Token）为认证主体。
	// 由 cmd/api/main.go 装配复合解析器；未设置时回退为 JWT-only（测试环境）。
	PrincipalParser func(token string) (auth.Principal, error)
	ApiTokenSvc     *apitoken.Service
	WorkspaceStore  *auth.WorkspaceMembershipStore
	WorkspaceSvc    *workspace.Service
	MemberSvc       *workspace.MemberService
	InvitationSvc   *workspace.InvitationService
	ProjectSvc      *workspace.ProjectService
	AuditSvc        *workspace.AuditService
	Mail            mail.EmailService
	// Issue 域
	IssueSvc            *issue.Service
	StateSvc            *issue.StateService
	ActivitySvc         *issue.ActivityService
	TimeLogSvc          *issue.TimeLogService
	IssueHandler        *issue.IssueHandler
	PrefHandler         *preference.Handler
	SearchHandler       *search.SearchHandler
	SprintHandler       *sprint.Handler
	VersionHandler      *version.Handler
	WorkbenchHandler    *workbench.WorkbenchHandler
	DashboardHandler    *dashboard.DashboardHandler
	NotificationHandler      *notif.Handler
	AttachmentHandler        *attachment.Handler
	DefectAnalyticsHandler   *issue.DefectAnalyticsHandler
	WSHub                    *ws.Hub
}

// RegisterIssueRoutes 注册工作项路由（在 NewEngine 之后调用）。
func RegisterIssueRoutes(r *gin.Engine, d *Deps) {
	if d.IssueHandler == nil {
		return
	}
	v1 := r.Group("/api/v1/workspaces/:workspace_id")
	v1.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
	)
	projects := v1.Group("/projects/:project_id")
	projects.Use(middleware.RequireProjectParam())
	// 读操作需要 workspace:read
	read := projects.Group("")
	read.Use(middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead))
	d.IssueHandler.Register(read, nil, nil)
}

// RegisterPreferenceRoutes 注册视图偏好路由（项目级）。
func RegisterPreferenceRoutes(r *gin.Engine, d *Deps) {
	if d.PrefHandler == nil {
		return
	}
	projects := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id")
	projects.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.PrefHandler.Register(projects)
}

// RegisterSprintRoutes 注册迭代路由（独立于 Issue 路由，可在 IssueHandler 未就绪时注册）。
func RegisterSprintRoutes(r *gin.Engine, d *Deps) {
	if d.SprintHandler == nil {
		return
	}
	v1 := r.Group("/api/v1/workspaces/:workspace_id")
	v1.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
	)
	projects := v1.Group("/projects/:project_id")
	projects.Use(middleware.RequireProjectParam())
	d.SprintHandler.Register(projects)
}

// RegisterVersionRoutes 注册版本路由（独立于 Issue 路由）。
func RegisterVersionRoutes(r *gin.Engine, d *Deps) {
	if d.VersionHandler == nil {
		return
	}
	v1 := r.Group("/api/v1/workspaces/:workspace_id")
	v1.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
	)
	projects := v1.Group("/projects/:project_id")
	projects.Use(middleware.RequireProjectParam())
	d.VersionHandler.Register(projects)
}

// RegisterSearchRoutes 注册全局搜索路由（独立于 Issue/Sprint/Version 路由）。
func RegisterSearchRoutes(r *gin.Engine, d *Deps) {
	if d.SearchHandler == nil {
		return
	}

	// ----- 项目级搜索（含过滤） -----
	project := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id/search")
	project.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.SearchHandler.Register(project)

	// ----- 工作空间级全局搜索（跨项目） -----
	wsSearch := r.Group("/api/v1/workspaces/:workspace_id/search")
	wsSearch.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	wsSearch.GET("", d.SearchHandler.Search)
	wsSearch.GET("/history", d.SearchHandler.ListHistory)
	wsSearch.DELETE("/history", d.SearchHandler.ClearHistory)
	wsSearch.DELETE("/history/:history_id", d.SearchHandler.DeleteHistory)
	wsSearch.GET("/bookmarks", d.SearchHandler.ListBookmarks)
	wsSearch.POST("/bookmarks", d.SearchHandler.CreateBookmark)
	wsSearch.PATCH("/bookmarks/:bookmark_id", d.SearchHandler.UpdateBookmark)
	wsSearch.DELETE("/bookmarks/:bookmark_id", d.SearchHandler.DeleteBookmark)

	// ----- 工作台（项目级） -----
	projectWb := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id/workbench")
	projectWb.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.WorkbenchHandler.Register(projectWb)

	// ----- 工作台（工作空间级，跨项目汇总） -----
	wsWb := r.Group("/api/v1/workspaces/:workspace_id/workbench")
	wsWb.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	wsWb.GET("/summary", d.WorkbenchHandler.GetSummary)
	wsWb.GET("/config", d.WorkbenchHandler.GetConfig)
	wsWb.PUT("/config", d.WorkbenchHandler.SaveConfig)
	wsWb.GET("/recent", d.WorkbenchHandler.ListRecent)
	wsWb.POST("/recent", d.WorkbenchHandler.RecordRecent)
	wsWb.GET("/templates", d.WorkbenchHandler.ListTemplates)
	wsWb.POST("/templates/apply", d.WorkbenchHandler.ApplyTemplate)
}

// RegisterWorkbenchRoutes 保留签名兼容（路由已在 RegisterSearchRoutes 中注册）。
func RegisterWorkbenchRoutes(_ *gin.Engine, _ *Deps) {}

// RegisterDashboardRoutes 注册项目仪表盘路由。
func RegisterDashboardRoutes(r *gin.Engine, d *Deps) {
	if d.DashboardHandler == nil {
		return
	}

	// 项目级仪表盘
	project := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id/dashboard")
	project.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.DashboardHandler.Register(project)

	// 工作空间级仪表盘（跨项目汇总）
	ws := r.Group("/api/v1/workspaces/:workspace_id/dashboard")
	ws.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	ws.GET("/alerts", d.DashboardHandler.ListAlerts)
	ws.POST("/alerts/:alert_id/resolve", d.DashboardHandler.ResolveAlert)
	ws.GET("/templates", d.DashboardHandler.ListTemplates)
}

// RegisterNotificationRoutes 注册通知路由。
func RegisterNotificationRoutes(r *gin.Engine, d *Deps) {
	if d.NotificationHandler == nil {
		return
	}
	ws := r.Group("/api/v1/workspaces/:workspace_id")
	ws.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.NotificationHandler.RegisterRoutes(ws)
	d.NotificationHandler.RegisterPreferenceRoutes(ws)
}

// RegisterAttachmentRoutes 注册附件路由（项目级）。
func RegisterAttachmentRoutes(r *gin.Engine, d *Deps) {
	if d.AttachmentHandler == nil {
		return
	}
	projects := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id")
	projects.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.AttachmentHandler.Register(projects)
}

// RegisterDefectAnalyticsRoutes 注册缺陷分析报表路由（项目级）。
func RegisterDefectAnalyticsRoutes(r *gin.Engine, d *Deps) {
	if d.DefectAnalyticsHandler == nil {
		return
	}
	projects := r.Group("/api/v1/workspaces/:workspace_id/projects/:project_id")
	projects.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
		middleware.RequireProjectParam(),
		middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead),
	)
	d.DefectAnalyticsHandler.Register(projects)
}

// RegisterWSRoutes 注册 WebSocket 实时推送路由。
// 客户端通过 wss://host/ws/:workspace_id 连接，认证走 Access Token（Cookie 或 Bearer）。
func RegisterWSRoutes(r *gin.Engine, d *Deps) {
	if d.WSHub == nil {
		return
	}
	wsGroup := r.Group("/ws/:workspace_id")
	wsGroup.Use(
		middleware.RequireAuth(d.principalParser()),
		middleware.RequireWorkspaceParam(),
	)
	wsGroup.GET("", func(c *gin.Context) {
		userID := c.GetInt64(middleware.CtxUserID)
		workspaceID := c.GetInt64(middleware.CtxWorkspaceID)
		d.WSHub.HandleWebSocket(c.Writer, c.Request, userID, workspaceID)
	})
}

// NewEngine 构建带完整中间件链的 HTTP 引擎。
func NewEngine(d *Deps) *gin.Engine {
	if !d.Cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	origins := []string{"http://localhost:5173", "http://localhost:8080"}
	r.Use(
		middleware.SecurityHeaders(),
		middleware.RequestID(),
		middleware.Recovery(d.Log),
		middleware.CORS(origins),
		middleware.AccessLog(d.Log),
		telemetry.MetricsMiddleware(),
	)

	r.GET("/healthz", healthz())
	r.GET("/readyz", readyz(d))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// Swagger UI（仅开发环境暴露）
	if d.Cfg.IsDev() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		rl := middleware.RateLimit(d.Redis, d.Cfg.Auth.LoginRateLimitPer, func(c *gin.Context) string {
			return "auth:" + c.ClientIP()
		})
		{
			authGroup.POST("/login", rl, login(d))
			authGroup.POST("/refresh", rl, refresh(d))
			authGroup.POST("/register", rl, register(d))
			authGroup.POST("/forgot-password", rl, forgotPassword(d))
			authGroup.POST("/reset-password", rl, resetPassword(d))
		}

	// 已认证路由（需要有效 access token + 用户级限流）
	authed := v1.Group("")
	authed.Use(middleware.RequireAuth(d.principalParser()))
		authed.Use(middleware.RateLimit(d.Redis, 100, func(c *gin.Context) string {
			return "user:" + userKey(c)
		}))
		{
			authed.GET("/me", me(d))

			// ----- 个人 API Token 管理（用户级，与工作空间无关） -----
			authed.GET("/me/api-tokens", listMyApiTokens(d))
			authed.POST("/me/api-tokens", createApiToken(d))
			authed.DELETE("/me/api-tokens/:token_id", revokeApiToken(d))

			// ----- 工作空间集合路由（无需 :workspace_id） -----
			authed.GET("/workspaces", listWorkspaces(d))
			authed.POST("/workspaces", createWorkspace(d))

			// 邀请接受链接（public-ish，需登录但无需 workspace 成员；在 RequireAuth 之后）
			authed.POST("/invitations/accept", acceptInvitation(d))
			authed.GET("/invitations/:token", getInvitationPreview(d))

			// 通过 slug 查 ID（前端路由使用 slug）
			authed.GET("/workspaces/slug/:slug", getWorkspaceBySlug(d))

			// ----- 工作空间作用域路由 -----
			ws := authed.Group("/workspaces/:workspace_id")
			ws.Use(middleware.RequireWorkspaceParam(), middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead))
			{
				ws.GET("", getWorkspace(d))
				ws.PATCH("", requireWsPermission(d, auth.PermWorkspaceUpdate), updateWorkspace(d))
				ws.DELETE("", requireWsPermission(d, auth.PermWorkspaceDelete), archiveWorkspace(d))

				// 成员
				ws.GET("/members", listMembers(d))
				ws.PATCH("/members/:user_id", requireWsPermission(d, auth.PermMemberChangeRole), changeMemberRole(d))
				ws.DELETE("/members/:user_id", requireWsPermission(d, auth.PermMemberRemove), removeMember(d))

				// 邀请
				ws.POST("/invitations", requireWsPermission(d, auth.PermMemberInvite), sendInvitation(d))
				ws.GET("/invitations", listInvitations(d))
				ws.DELETE("/invitations/:invitation_id", requireWsPermission(d, auth.PermMemberInvite), revokeInvitation(d))

				// 审计（owner/admin only）
				ws.GET("/audit-logs", requireWsPermission(d, auth.PermAuditRead), listAuditLogs(d))

				// 项目
				ws.GET("/projects", requireWsPermission(d, auth.PermWorkspaceRead), listProjects(d))
				ws.POST("/projects", requireWsPermission(d, auth.PermProjectCreate), createProject(d))
				ws.GET("/projects/:project_id", requireWsPermission(d, auth.PermWorkspaceRead), getProject(d))
				ws.PATCH("/projects/:project_id", requireWsPermission(d, auth.PermProjectCreate), updateProject(d))
				ws.DELETE("/projects/:project_id", requireWsPermission(d, auth.PermProjectDelete), archiveProject(d))
			}
		}
	}

	r.NoRoute(middleware.NoRoute())
	return r
}

// requireWsPermission 是组合中间件的语法糖。Gin 不支持链式式中文传递权限常量。
func requireWsPermission(d *Deps, perm string) gin.HandlerFunc {
	return middleware.RequirePermission(d.WorkspaceStore, perm)
}

// principalParser 返回认证主体解析器。
//
// 生产环境由 cmd/api/main.go 注入复合解析器（JWT → API Token 双通道）；
// 测试环境（stubDeps 未注入）回退为 JWT-only：Auth 为空时一律 401，
// 保证冒烟测试对"未认证请求返回 401"的断言仍然成立。
func (d *Deps) principalParser() func(token string) (auth.Principal, error) {
	if d.PrincipalParser != nil {
		return d.PrincipalParser
	}
	return func(token string) (auth.Principal, error) {
		if d.Auth == nil {
			return auth.Principal{}, errs.ErrUnauthorized
		}
		uid, err := d.Auth.ParseAccess(token)
		if err != nil {
			return auth.Principal{}, err
		}
		return auth.Principal{UserID: uid, Kind: auth.PrincipalJWT}, nil
	}
}

func userKey(c *gin.Context) string {
	if uid, ok := c.Get(middleware.CtxUserID); ok {
		if id, ok := uid.(int64); ok {
			return "u" + itoa(id)
		}
	}
	return c.ClientIP()
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

// --- handlers (platform-level) ---

func healthz() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func readyz(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		checks := gin.H{}
		healthy := true

		if err := d.DB.Ping(ctx); err != nil {
			checks["postgres"] = "down"
			healthy = false
		} else {
			checks["postgres"] = "up"
		}
		if err := d.Redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "down"
			healthy = false
		} else {
			checks["redis"] = "up"
		}

		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"status": map[bool]string{true: "ready", false: "degraded"}[healthy], "checks": checks})
	}
}

type loginRequest struct {
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"your-password"`
}

// login godoc
//
//	@Summary		登录
//	@Description	使用邮箱和密码获取访问 / 刷新令牌对
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest		true	"登录凭据"
//	@Success		200		{object}	auth.TokenPair
//	@Failure		401		{object}	errs.AppError
//	@Failure		422		{object}	errs.AppError
//	@Router			/auth/login [post]
func login(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		pair, err := d.Auth.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			writeError(c, err)
			return
		}
		setAuthCookies(c, d, pair)
		c.JSON(http.StatusOK, pair)
	}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// refresh godoc
//
//	@Summary		刷新访问令牌
//	@Description	使用刷新令牌（body 或 Cookie）获取新令牌对
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		refreshRequest	true	"刷新令牌"
//	@Success		200		{object}	auth.TokenPair
//	@Failure		401		{object}	errs.AppError
//	@Router			/auth/refresh [post]
func refresh(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req refreshRequest
		_ = c.ShouldBindJSON(&req)
		token := req.RefreshToken
		if token == "" {
			token, _ = c.Cookie("ydsz_refresh")
		}
		if token == "" {
			middleware.AbortWithError(c, errs.ErrUnauthorized)
			return
		}
		pair, err := d.Auth.Refresh(c.Request.Context(), token)
		if err != nil {
			writeError(c, err)
			return
		}
		setAuthCookies(c, d, pair)
		c.JSON(http.StatusOK, pair)
	}
}

// register godoc
//
//	@Summary		注册
//	@Description	创建新用户并自动签发令牌
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		registerRequest	true	"注册信息"
//	@Success		201		{object}	auth.TokenPair
//	@Failure		409		{object}	errs.AppError	"邮箱已被注册"
//	@Failure		422		{object}	errs.AppError
//	@Router			/auth/register [post]
func register(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		pair, err := d.Auth.Register(c.Request.Context(), auth.RegisterInput{
			Email:       req.Email,
			Password:    req.Password,
			DisplayName: req.DisplayName,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		setAuthCookies(c, d, pair)
		c.JSON(http.StatusCreated, pair)
	}
}

type registerRequest struct {
	Email       string `json:"email" example:"alice@example.com"`
	Password    string `json:"password" example:"new-password"`
	DisplayName string `json:"display_name" example:"Alice"`
}

// me godoc
//
//	@Summary		获取当前用户
//	@Description	返回调用者的用户简介
//	@Tags			user
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	auth.UserBrief
//	@Failure		401	{object}	errs.AppError
//	@Router			/me [get]
func me(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetInt64(middleware.CtxUserID)
		var u auth.UserBrief
		err := d.DB.QueryRow(c.Request.Context(),
			`SELECT id, email, display_name, coalesce(avatar_url,'') FROM users WHERE id = $1 AND is_active`, uid).
			Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL)
		if err != nil {
			middleware.AbortWithError(c, errs.ErrNotFound)
			return
		}
		c.JSON(http.StatusOK, u)
	}
}

// forgotPassword 发起密码重置请求：校验通过后向邮箱发送重置链接，
// 无论邮箱是否存在都返回 202 以避免用户枚举。
func forgotPassword(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req forgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		_ = d.ResetSvc.RequestReset(c.Request.Context(), req.Email)
		// 模糊化返回 202（防枚举）
		c.Status(http.StatusAccepted)
	}
}

// resetPassword 使用重置令牌设置新密码，成功后返回 204。
func resetPassword(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req resetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		if err := d.ResetSvc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// --- helpers ---

func setAuthCookies(c *gin.Context, d *Deps, pair *auth.TokenPair) {
	secure := !d.Cfg.IsDev()
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "ydsz_access", Value: pair.AccessToken, Path: "/",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
		MaxAge: int(d.Cfg.Auth.AccessTokenTTL.Seconds()),
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "ydsz_refresh", Value: pair.RefreshToken, Path: "/api/v1/auth",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
		MaxAge: int(d.Cfg.Auth.RefreshTokenTTL.Seconds()),
	})
}

func writeError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}

// fieldDetails —— minimal mapping；完整 validator-tag 翻译在后续迭代落地。
func fieldDetails(err error) []errs.FieldDetail {
	return []errs.FieldDetail{{Field: "body", Reason: err.Error()}}
}
