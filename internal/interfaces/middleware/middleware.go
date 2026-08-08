// Package middleware 包含 Gin 中间件链：
// request_id → recovery → cors → ratelimit → auth → tenant → rbac → audit。
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// CtxKey 枚举存储在 gin.Context 中的键。
const (
	CtxRequestID   = "request_id"
	CtxUserID      = "user_id"
	CtxWorkspaceID = "workspace_id"
	CtxProjectID   = "project_id"
	CtxAuthKind    = "auth_kind"
	CtxAuthScopes  = "auth_scopes"
)

// SecurityHeaders 添加纵深防御型 HTTP 响应头。这些头补充反向代理 nginx
// 的配置，即使边缘节点终止 TLS 也能加固应用。
// 参考：OWASP Secure Header Project、Google gts/security 指南。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 禁止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 点击劫持防护。SAMEORIGIN 允许同站 iframe 内嵌，
		// 为将来可能的内部看板保留空间。
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// 浏览器 XSS 过滤器 —— 保持禁用（0），因为它在老浏览器中
		// 引入过滤器绕过向量；应依赖 CSP 而非此头。
		c.Header("X-XSS-Protection", "0")

		// 限制跨源导航时的 Referrer 泄露
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 控制文档/上下文可用的浏览器特性
		c.Header("Permissions-Policy",
			"accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")

		// 跨源隔离（进程级沙箱）。
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Content-Security-Policy —— 基线限制性策略。
		// 对 SPA（独立 Vite dev server + nginx 提供的生产包）而言，
		// API 端点只需要最小策略；SPA 的 HTML 应通过 nginx meta 标签/
		// 响应头设置自己的 CSP。
		c.Header("Content-Security-Policy",
			"default-src 'none'; frame-ancestors 'self'; base-uri 'none'; form-action 'self'")

		// HSTS 仅由 TLS 终结边缘（nginx/CDN）设置。此处有意省略，
		// 使 API 可以运行在任何代理之后，不会与代理的 HSTS max-age 冲突。

		c.Next()
	}
}

// RequestID 分配唯一请求 ID（或透传上游 X-Request-ID）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			var b [16]byte
			_, _ = rand.Read(b[:])
			id = hex.EncodeToString(b[:])
		}
		c.Set(CtxRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Recovery 将 panic 转换为统一 500 信封并记录堆栈。
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("request_id", c.GetString(CtxRequestID)),
					zap.String("path", c.Request.URL.Path),
					zap.Stack("stack"),
				)
				respondError(c, errs.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS 返回限制性 CORS 策略（可配置允许来源）。
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Api-Key", "X-CSRF-Token", "Idempotency-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// AccessLog 每个请求输出一行结构化日志（跳过 /healthz 噪音）。
func AccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if c.Request.URL.Path == "/healthz" {
			return
		}
		log.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", c.GetString(CtxRequestID)),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// RateLimit 是基于 Redis 的令牌桶限流器。keyFn 提取限流键
// （用户 ID 或 IP）。超限时返回 429 并附带 Retry-After。
func RateLimit(rdb *redis.Client, limitPerMin int, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:" + keyFn(c)
		ctx := c.Request.Context()

		now := time.Now().Unix()
		member := float64(now)

		pipe := rdb.TxPipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", itoa(now-60))
		addCmd := pipe.ZAdd(ctx, key, redis.Z{Score: member, Member: memberWithRand(now)})
		cardCmd := pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, 2*time.Minute)
		if _, err := pipe.Exec(ctx); err != nil {
			// 限流器故障时放行（fail-open），但记录日志
			c.Next()
			return
		}
		_ = addCmd
		remaining := int64(limitPerMin) - cardCmd.Val()
		c.Header("X-RateLimit-Limit", itoa(int64(limitPerMin)))
		c.Header("X-RateLimit-Reset", itoa(now+60))
		if remaining < 0 {
			c.Header("Retry-After", "60")
			respondError(c, errs.ErrRateLimited)
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Remaining", itoa(remaining))
		c.Next()
	}
}

func memberWithRand(now int64) any {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return itoa(now) + "-" + hex.EncodeToString(b[:])
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

// respondError 渲染统一错误信封。
func respondError(c *gin.Context, e *errs.AppError) {
	c.JSON(e.HTTP, gin.H{
		"error": gin.H{
			"code":       e.Code,
			"message":    e.Message,
			"details":    e.Details,
			"request_id": c.GetString(CtxRequestID),
		},
	})
}

// AbortWithError 让 handler 以信封格式携带 AppError 中止请求。
func AbortWithError(c *gin.Context, e *errs.AppError) {
	respondError(c, e)
	c.Abort()
}

// RequireAuth 校验访问凭证（JWT 或 API Token）并注入认证主体。
//
// parse 由上层装配（见 cmd/api/main.go 的复合解析器）：先尝试 JWT
// ParseAccess，失败后交给 API Token 服务查表校验（含 scope 与过期时间）。
// 校验通过后向 ctx 注入：
//
//	user_id     — 主体用户 ID
//	auth_kind   — 凭证类型（"jwt" | "api_token"）
//	auth_scopes — API Token 携带的 scope 白名单（仅 api_token 非空）
//
// 任何凭证无效场景统一返回 401，不区分具体原因（防探测）。
func RequireAuth(parse func(token string) (auth.Principal, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		p, err := parse(token)
		if err != nil {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserID, p.UserID)
		c.Set(CtxAuthKind, string(p.Kind))
		if len(p.Scopes) > 0 {
			c.Set(CtxAuthScopes, p.Scopes)
		}
		c.Next()
	}
}

// bearerToken 从请求中提取访问凭证，优先级：
//
//  1. Authorization: Bearer <token>
//  2. X-Api-Key: <token>（API Token 规范头，见 docs/architecture/05）
//  3. Cookie: ydsz_access=<token>（SPA 会话）
func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && h[:len(prefix)] == prefix {
		return h[len(prefix):]
	}
	// 脚本/集成使用 X-Api-Key 头携带 API Token
	if k := c.GetHeader("X-Api-Key"); k != "" {
		return k
	}
	// SPA 的 cookie 会话
	if ck, err := c.Cookie("ydsz_access"); err == nil {
		return ck
	}
	return ""
}

// NoRoute 为未知 API 路径返回 JSON 404。
func NoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		respondError(c, errs.ErrNotFound)
		c.AbortWithStatus(http.StatusNotFound)
	}
}

// ============================================================
// 三层 API 路由鉴权中间件（P2-2）
// 对标 Plane 的多路由层：Session / API Key / Public
// ============================================================

// SessionAuth 浏览器 SPA 专用鉴权：JWT（Bearer + Cookie），最严格 CORS。
func SessionAuth(cfg *config.Config, authSvc *auth.Service, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		if len(token) > 5 && token[:5] == "reks_" {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		uid, err := authSvc.ParseAccess(token)
		if err != nil {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserID, uid)
		c.Set(CtxAuthKind, string(auth.PrincipalJWT))
		c.Next()
	}
}

// APIKeyAuth 程序化访问专用鉴权：X-API-Key → API Token 服务查表校验。
func APIKeyAuth(apiTokenSvc *apitoken.Service, authSvc *auth.Service, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Api-Key")
		if key == "" {
			key = bearerToken(c)
		}
		if key == "" {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		if len(key) < 5 || key[:5] != "reks_" {
			if uid, err := authSvc.ParseAccess(key); err == nil {
				c.Set(CtxUserID, uid)
				c.Set(CtxAuthKind, string(auth.PrincipalJWT))
				c.Next()
				return
			}
		}
		p, err := apiTokenSvc.ResolvePrincipal(c.Request.Context(), key)
		if err != nil {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserID, p.UserID)
		c.Set(CtxAuthKind, string(auth.PrincipalAPIToken))
		if len(p.Scopes) > 0 {
			c.Set(CtxAuthScopes, p.Scopes)
		}
		c.Next()
	}
}

// AnonymousSession Public 层入口：不鉴权，仅注入 auth_kind=anonymous。
func AnonymousSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CtxAuthKind, "anonymous")
		c.Next()
	}
}

// RequireAPIScope API Key 层专用权限中间件：校验 scope 白名单覆盖 required 权限。
func RequireAPIScope(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		kind, _ := c.Get(CtxAuthKind)
		if kind == string(auth.PrincipalJWT) || kind == string(auth.PrincipalAPIToken) {
			c.Next()
			return
		}
		rawScopes, _ := c.Get(CtxAuthScopes)
		scopes, _ := rawScopes.([]string)
		if !apitoken.ScopeCovers(scopes, required) {
			respondError(c, errs.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
