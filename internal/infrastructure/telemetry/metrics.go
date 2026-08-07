// Package telemetry 暴露 Prometheus 指标（服务端），遵循 RED 方法：
// Rate（每秒请求数）、Errors（每秒失败请求数）、Duration（延迟直方图，
// P50/P95/P99 分桶）。导入 "github.com/prometheus/client_golang/prometheus/promhttp"
// 可暴露 /metrics 端点。
//
// 参考：Google SRE Book 第 6 章《监控分布式系统》。
package telemetry

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// namespace 是所有指标的前缀，遵循 Prometheus 命名规范。
const (
	namespace = "ydsz"
	subsystem = "http"
)

var (
	// RequestTotal 按 method / route / status 统计请求总数。
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_total",
			Help:      "Total HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	// RequestDurationMs 记录请求延迟分布。
	RequestDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
		Subsystem: subsystem,
			Name:      "request_duration_ms",
			Help:      "HTTP request latency distribution.",
			Buckets:   prometheus.DefBuckets, // .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		},
		[]string{"method", "route"},
	)

	// DBDurationMs 记录数据库查询延迟（由服务层手动调用埋点）。
	DBDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_ms",
			Help:      "Database query latency distribution.",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 5000},
		},
		[]string{"operation"},
	)

	// RedisPublished 统计转发到 Redis Streams 的 outbox 事件（成功/失败）。
	RedisPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "redis_events_published_total",
			Help:      "Outbox events relayed to Redis Streams.",
		},
		[]string{"status"},
	)

	// AuthOperations 统计注册/登录/刷新结果。
	AuthOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_operations_total",
			Help:      "Authentication operations.",
		},
		[]string{"operation", "status"},
	)

	// VersionOperations 统计版本 CRUD 与生命周期操作（S6 大厂加固）。
	// labels: operation=create|update|delete|active|released|archived, status=success|error
	VersionOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "version_operations_total",
			Help:      "Version CRUD and lifecycle operations.",
		},
		[]string{"operation", "status"},
	)
)

// MetricsMiddleware 为每个请求记录 request_total 与 request_duration_ms。
// 它提取匹配的路由模式（c.FullPath()），使形如
// /api/v1/workspaces/{slug}/issues/123 的路径坍缩为同一组标签序列。
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := float64(time.Since(start).Milliseconds())

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := classStatus(c.Writer.Status())
		RequestTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		RequestDurationMs.WithLabelValues(c.Request.Method, route).Observe(latency)
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
