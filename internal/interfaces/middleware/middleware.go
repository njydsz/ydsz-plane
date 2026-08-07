// Package middleware contains the Gin middleware chain:
// request_id → recovery → cors → ratelimit → auth → tenant → rbac → audit.
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

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// CtxKey enumerates keys stored in gin.Context.
const (
	CtxRequestID   = "request_id"
	CtxUserID      = "user_id"
	CtxWorkspaceID = "workspace_id"
	CtxProjectID   = "project_id"
)

// SecurityHeaders adds defense-in-depth HTTP headers. These complement the
// reverse-proxy nginx config and harden the app even when the edge terminates
// TLS. Reference: OWASP Secure Header Project, Google gts/security guide.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Clickjacking protection. SAMEORIGIN allows same-site iframe embedding
		// for potential future internal dashboards.
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// Browser XSS filter — kept at 0 (disabled) because it introduces
		// filter-bypass vectors in older browsers; rely on CSP instead.
		c.Header("X-XSS-Protection", "0")

		// Limit referrer leakage on cross-origin navigations
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Control which browser features the document/context can use
		c.Header("Permissions-Policy",
			"accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")

		// Cross-origin isolation (process-level siloing).
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Content-Security-Policy — baseline restrictive policy.
		// For the SPA (separate Vite dev server + nginx-served prod bundle),
		// the API endpoints only need a minimal policy; the SPA's HTML
		// should set its own CSP via nginx meta tag / response header.
		c.Header("Content-Security-Policy",
			"default-src 'none'; frame-ancestors 'self'; base-uri 'none'; form-action 'self'")

		// HSTS is set only by the TLS-terminating edge (nginx/CDN). We
		// intentionally omit it here so the API can run behind any proxy
		// without conflicting with the proxy's HSTS max-age.

		c.Next()
	}
}

// RequestID assigns a unique request id (or propagates X-Request-ID).
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

// Recovery converts panics into the unified 500 envelope and logs a stack.
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

// CORS returns a restrictive CORS policy (configurable allowed origins).
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

// AccessLog emits one structured line per request (skip /healthz noise).
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

// RateLimit is a Redis token-bucket limiter. keyFn extracts the limit key
// (user id or IP). On exceed it returns 429 with Retry-After.
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
			// fail-open on limiter outage, but log it
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

// respondError renders the unified error envelope.
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

// AbortWithError lets handlers abort with an AppError using the envelope.
func AbortWithError(c *gin.Context, e *errs.AppError) {
	respondError(c, e)
	c.Abort()
}

// RequireAuth validates the access token (see internal/application/auth) and
// injects the user id. Implemented as a factory to avoid an import cycle.
func RequireAuth(parseToken func(token string) (userID int64, err error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			respondError(c, errs.ErrUnauthorized)
			c.Abort()
			return
		}
		uid, err := parseToken(token)
		if err != nil {
			respondError(c, errs.ErrTokenExpired)
			c.Abort()
			return
		}
		c.Set(CtxUserID, uid)
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if len(h) > len(prefix) && h[:len(prefix)] == prefix {
		return h[len(prefix):]
	}
	// cookie-based session for the SPA
	if ck, err := c.Cookie("ydsz_access"); err == nil {
		return ck
	}
	return ""
}

// NoRoute returns the JSON 404 for unknown API paths.
func NoRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		respondError(c, errs.ErrNotFound)
		c.AbortWithStatus(http.StatusNotFound)
	}
}
