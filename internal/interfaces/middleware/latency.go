// Package middleware 提供 Gin 中间件链与可组合的安全/治理组件。
//
// latency.go 记录每个 HTTP 请求的耗时详情：
//   - 结构化 Debug 日志（method/path/status_code/duration_ms/user_agent/client_ip）
//   - Prometheus 指标 plane_http_request_duration_seconds（histogram）
//   - Prometheus 指标 plane_http_requests_total（counter）
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	// plane 指标命名空间前缀（遵循项目约束）。
	planeNamespace = "plane"
)

var (
	// planeHTTPRequestDuration 记录 HTTP 请求延迟分布。
	// Buckets 覆盖 10ms-10s 典型范围，满足 API 延迟可观测性需求。
	planeHTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: planeNamespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency distribution in seconds.",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// planeHTTPRequestTotal 按 method/path/status 统计请求总数。
	planeHTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: planeNamespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
)

// Latency 记录每个 HTTP 请求的耗时详情到 Debug 日志与 Prometheus 指标。
//
// 设计约定：
//   - 日志使用 Debug 级别，不污染 INFO 访问日志（AccessLog 中间件已有 INFO 级记录）。
//   - path 使用 route pattern（c.FullPath()），不暴露具体资源 ID，如
//     /api/v1/workspaces/:workspace_id/issues/:issue_id 始终归一到同一标签值。
//   - 状态码使用粗粒度分类（2xx/3xx/4xx/5xx），控制指标基数。
//
// 该中间件应在路由链中尽量靠前（SecurityHeaders/RequestID 之后、业务逻辑之前），
// 以捕获完整的请求处理时间。
func Latency(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := classStatus(c.Writer.Status())
		durationMs := float64(duration.Milliseconds())

		// Debug 级别结构化日志 —— 包含 user_agent 与 client_ip，便于问题排查。
		log.Debug("latency",
			zap.String("method", c.Request.Method),
			zap.String("path", route),
			zap.Int("status_code", c.Writer.Status()),
			zap.Float64("duration_ms", durationMs),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("client_ip", c.ClientIP()),
		)

		// Prometheus 指标埋点。
		planeHTTPRequestDuration.WithLabelValues(c.Request.Method, route, status).Observe(duration.Seconds())
		planeHTTPRequestTotal.WithLabelValues(c.Request.Method, route, status).Inc()
	}
}

// classStatus 将 HTTP 状态码映射为粗粒度标签 "2xx".."5xx"。
func classStatus(code int) string {
	switch {
	case code < 200:
		return "1xx"
	case code < 300:
		return "2xx"
	case code < 400:
		return "3xx"
	case code < 500:
		return "4xx"
	default:
		return "5xx"
	}
}
