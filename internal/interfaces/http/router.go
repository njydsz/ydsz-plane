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
	"go.uber.org/zap"

	"github.com/ydszopen/ydsz-plane/internal/application/auth"
	"github.com/ydszopen/ydsz-plane/internal/config"
	"github.com/ydszopen/ydsz-plane/internal/interfaces/middleware"
	"github.com/ydszopen/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/ydszopen/ydsz-plane/pkg/errs"
)

// Deps carries handler dependencies.
type Deps struct {
	Cfg   *config.Config
	Log   *zap.Logger
	DB    *pgxpool.Pool
	Redis *redis.Client
	Auth  *auth.Service
}

// NewEngine builds the HTTP engine with the full middleware chain.
func NewEngine(d *Deps) *gin.Engine {
	if !d.Cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	origins := []string{"http://localhost:5173", "http://localhost:8080"}
	r.Use(
		middleware.RequestID(),
		middleware.Recovery(d.Log),
		middleware.CORS(origins),
		middleware.AccessLog(d.Log),
		telemetry.MetricsMiddleware(), // RED metrics (Rate/Errors/Duration)
	)

	r.GET("/healthz", healthz())
	r.GET("/readyz", readyz(d))
	r.GET("/metrics", gin.WrapH(promhttp.Handler())) // Prometheus scrape endpoint

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
		}

		// authenticated routes
		authed := v1.Group("")
		authed.Use(middleware.RequireAuth(d.Auth.ParseAccess))
		authed.Use(middleware.RateLimit(d.Redis, 100, func(c *gin.Context) string {
			return "user:" + userKey(c)
		}))
		{
			authed.GET("/me", me(d))
			// S2+ routes mount here:
			// workspaces, projects, issues, sprints, versions, ...
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

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
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=64"`
}

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
