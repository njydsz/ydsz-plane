// Package telemetry exposes Prometheus metrics (server side) following the
// RED method: Rate (requests/sec), Errors (failed req/sec), Duration
// (latency histogram with P50/P95/P99 buckets). Import
// "github.com/prometheus/client_golang/prometheus/promhttp" to expose /metrics.
//
// Reference: Google SRE Book Chapter 6 — Monitoring Distributed Systems.
package telemetry

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Namespace prefix for all metrics, following Prometheus naming conventions.
const (
	namespace = "ydsz"
	subsystem = "http"
)

var (
	// RequestTotal counts requests by method, route, and status class.
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_total",
			Help:      "Total HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	// RequestDurationMs records request latency distribution.
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

	// DBDurationMs records database query latency (called manually from services).
	DBDurationMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_ms",
			Help:      "Database query latency distribution.",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 5000},
		},
		[]string{"operation"},
	)

	// NATSPublished counts outbox events relayed to NATS (success/fail).
	NATSPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "nats_events_published_total",
			Help:      "Outbox events relayed to NATS.",
		},
		[]string{"status"},
	)

	// AuthOperations counts registration/login/refresh outcomes.
	AuthOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_operations_total",
			Help:      "Authentication operations.",
		},
		[]string{"operation", "status"},
	)
)

// MetricsMiddleware records request_total + request_duration_ms per request.
// It extracts the matched route pattern (c.FullPath()) so that parameters
// like /api/v1/workspaces/{slug}/issues/123 collapse into one label series.
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

// classStatus maps HTTP status code to the coarse label "2xx".."5xx".
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
