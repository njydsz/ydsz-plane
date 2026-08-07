// Package httpapi wires the Gin engine, middleware chain and route table.
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

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Deps carries handler dependencies.
type Deps struct {
	Cfg            *config.Config
	Log            *zap.Logger
	DB             *pgxpool.Pool
	Redis          *redis.Client
	Auth           *auth.Service
	WorkspaceStore *auth.WorkspaceMembershipStore
	Mail           mail.EmailService
}

// NewEngine builds the HTTP engine with the full middleware chain.
func NewEngine(d *Deps) *gin.Engine {
	if !d.Cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	origins := []string{"http://localhost:5173", "http://localhost:8080"}
	r.Use(
		middleware.SecurityHeaders(), // defense-in-depth HTTP headers
		middleware.RequestID(),
		middleware.Recovery(d.Log),
		middleware.CORS(origins),
		middleware.AccessLog(d.Log),
		telemetry.MetricsMiddleware(), // RED metrics (Rate/Errors/Duration)
	)

	r.GET("/healthz", healthz())
	r.GET("/readyz", readyz(d))
	r.GET("/metrics", gin.WrapH(promhttp.Handler())) // Prometheus scrape endpoint
	// Swagger UI (only in development)
	if d.Cfg.IsDev() {
		r.GET("/swagger/*Any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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
	}

	// authenticated routes
	authed := v1.Group("")
	authed.Use(middleware.RequireAuth(d.Auth.ParseAccess))
	authed.Use(middleware.RateLimit(d.Redis, 100, func(c *gin.Context) string {
		return "user:" + userKey(c)
	}))
	{
		authed.GET("/me", me(d))
		authed.POST("/auth/reset-password", resetPassword(d))
			// 工作空间作用域路由：由 RequireAuth 注入 UserID，RequireWorkspaceParam 注入 WorkspaceID，
			// RequirePermission 基于角色矩阵向下钻取到权限粒度。
			ws := authed.Group("/workspaces/:workspace_id")
			ws.Use(middleware.RequireWorkspaceParam(), middleware.RequirePermission(d.WorkspaceStore, auth.PermWorkspaceRead))
			{
				ws.GET("/members", listWorkspaceMembers(d))
			}
		}
	}

	r.NoRoute(middleware.NoRoute())
	return r
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

// --- handlers ---

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

// listWorkspaceMembers godoc
//
//	@Summary		列出工作空间成员
//	@Description	返回当前用户在 :workspace_id 工作空间里的成员列表
//	@Tags			workspace
//	@Produce		json
//	@Security		Bearer
//	@Param			workspace_id	path		int	true	"工作空间 ID"
//	@Success		200				{object}	[]workspaceMemberResponse
//	@Failure		403				{object}	errs.AppError
//	@Router			/workspaces/{workspace_id}/members [get]
func listWorkspaceMembers(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsIDVal, _ := c.Get(middleware.CtxWorkspaceID)
		wsID := wsIDVal.(int64)
		roleVal, _ := c.Get("workspace_role")
		_ = roleVal

		rows, err := d.DB.Query(c.Request.Context(), `
			SELECT u.id, u.email, u.display_name, wm.role, wm.joined_at::text
			FROM workspace_members wm
			JOIN users u ON u.id = wm.user_id
			WHERE wm.workspace_id = $1 AND u.is_active
			ORDER BY wm.role, wm.joined_at`, wsID)
		if err != nil {
			middleware.AbortWithError(c, errs.ErrInternal.Wrap(err))
			return
		}
		defer rows.Close()

		var out []workspaceMemberResponse
		for rows.Next() {
			var m workspaceMemberResponse
			if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.JoinedAt); err != nil {
				middleware.AbortWithError(c, errs.ErrInternal.Wrap(err))
				return
			}
			out = append(out, m)
		}
		c.JSON(http.StatusOK, out)
	}
}

type workspaceMemberResponse struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

// forgotPassword godoc
//
//	@Summary		请求密码重置
//	@Description	根据邮箱发送一次性重置链接（15 分钟有效）。无论邮箱是否存在均返回 202，避免枚举。
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	forgotPasswordRequest	true	"邮箱"
//	@Success		202		"已接受（不一定实际存在该邮箱）"
//	@Failure		422		{object}	errs.AppError
//	@Router			/auth/forgot-password [post]
func forgotPassword(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req forgotPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}

		// 模糊化：不管用户是否存在，都返回 202（防用户名枚举）。
		var (
			userID int64
			name   string
		)
		err := d.DB.QueryRow(c.Request.Context(),
			`SELECT id, COALESCE(display_name,'用户') FROM users WHERE email = $1 AND is_active`, req.Email).
			Scan(&userID, &name)
		if err == nil && d.Mail != nil {
			// 异步发送（避免接口延迟受 SMTP 影响）；错误仅记日志不回写响应。
			go func() {
				// 占位：实际实现需 token 写入 DB（hash）+ 邮件发一次原始 token。
				_ = userID
				_ = name
			}()
		}
		c.Status(http.StatusAccepted)
	}
}

// resetPassword godoc
//
//	@Summary		使用 token 重置密码
//	@Description	提交一次性 token + 新密码；成功后 token 标记为已使用。
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	resetPasswordRequest	true	"重置请求"
//	@Success		204		"重置成功"
//	@Failure		400		{object}	errs.AppError	"token 无效或已过期"
//	@Failure		422		{object}	errs.AppError
//	@Router			/auth/reset-password [post]
func resetPassword(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req resetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetails(err)...))
			return
		}
		// 校验 token、更新密码、标记 token used_at。MVP 接入点占位。
		middleware.AbortWithError(c, errs.ErrNotFound)
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

func fieldDetails(err error) []errs.FieldDetail {
	// Minimal mapping; a full validator-tag translator lands in S2.
	return []errs.FieldDetail{{Field: "body", Reason: err.Error()}}
}
