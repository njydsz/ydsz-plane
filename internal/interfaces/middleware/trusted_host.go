package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TrustedHost 防护 Host 头注入攻击。
//
// 校验 r.Host / X-Forwarded-Host 必须在 allowedHosts 列表中。
// allowedHosts 为空时禁用校验（向后兼容，适用于未配置 YDSZ_ALLOWED_HOSTS 的环境）。
// 健康检查路径（/healthz / /metrics）跳过校验，避免阻止 orchestrator probe。
// 不匹配时返回 400 Bad Request + 日志 warn。
func TrustedHost(allowedHosts []string, log *zap.Logger) gin.HandlerFunc {
	// 预规范化：去空、统一小写。
	hosts := make([]string, 0, len(allowedHosts))
	for _, h := range allowedHosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		hosts = append(hosts, strings.ToLower(h))
	}
	hostSet := make(map[string]struct{}, len(hosts))
	for _, h := range hosts {
		hostSet[h] = struct{}{}
	}
	hostCount := len(hostSet)

	return func(c *gin.Context) {
		// 配置未启用 → 跳过。
		if hostCount == 0 {
			c.Next()
			return
		}

		// 健康检查/orchestrator probe 跳过。
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/metrics" || path == "/readyz" {
			c.Next()
			return
		}

		if !isHostAllowed(c, hostSet) {
			log.Warn("Host header rejected",
				zap.String("host", c.Request.Host),
				zap.String("x_forwarded_host", c.GetHeader("X-Forwarded-Host")),
				zap.String("path", path),
				zap.String("request_id", c.GetString(CtxRequestID)),
				zap.String("ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":       "bad_request",
					"message":    "invalid host header",
					"request_id": c.GetString(CtxRequestID),
				},
			})
			return
		}
		c.Next()
	}
}

// isHostAllowed 校验请求的 Host 与 X-Forwarded-Host（如存在）是否在允许列表中。
func isHostAllowed(c *gin.Context, hostSet map[string]struct{}) bool {
	host := normalizeHost(c.Request.Host)
	if _, ok := hostSet[host]; ok {
		return true
	}
	if xfh := c.GetHeader("X-Forwarded-Host"); xfh != "" {
		// X-Forwarded-Host 可能含多个值（proxy 链），取第一个。
		if comma := strings.IndexByte(xfh, ','); comma != -1 {
			xfh = xfh[:comma]
		}
		xfh = normalizeHost(xfh)
		if _, ok := hostSet[xfh]; ok {
			return true
		}
	}
	return false
}

// normalizeHost 剥离端口后小写化，适配带端口或不带端口两种形式。
func normalizeHost(h string) string {
	h = strings.TrimSpace(strings.ToLower(h))
	// 兼容 IPv6（[::1]:8080）；否则按最后一个 : 剥离端口。
	if strings.HasPrefix(h, "[") {
		if idx := strings.LastIndex(h, "]"); idx >= 0 {
			return h[:idx+1]
		}
	}
	if idx := strings.LastIndex(h, ":"); idx >= 0 {
		return h[:idx]
	}
	return h
}
