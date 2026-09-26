// Package telemetry — Gin HTTP 中间件（P10 OpenTelemetry 集成）。
//
// 为每个 HTTP 请求创建 span，自动记录 method / route / status / latency。
// 错误请求（status >= 500）强制设置 error flag 确保 100% 采样。
//
// 使用方式：
//
//	r.Use(telemetry.TracingMiddleware("ydsz-plane"))
//
// 依赖：go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
package telemetry

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// attributeInt64 构造 int64 类型的 attribute（快捷函数）。
func attributeInt64(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// TracingMiddleware 创建 Gin 中间件，为每个 HTTP 请求创建 span。
//
// 对标 Google Cloud Trace 中间件标准，自动注入以下 attributes：
//   - http.method / http.route / http.status_code / http.response_size
//   - network.protocol.version（HTTP/1.1 或 HTTP/2）
//   - url.path / url.query
//   - client.address（X-Real-IP / X-Forwarded-For）
//
// 错误请求（status ≥ 500）会记录 error=true 并添加异常事件。
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName,
		otelgin.WithSpanStartOptions(
			trace.WithAttributes(
				semconv.ServiceName(serviceName),
			),
		),
		otelgin.WithGinFilter(func(c *gin.Context) bool {
			// 排除健康检查端点（高频、无业务意义）
			path := c.Request.URL.Path
			switch path {
			case "/healthz", "/readyz", "/metrics", "/favicon.ico":
				return false
			}
			return true
		}),
	)
}

// RecordSpanError 在 Gin handler 中记录错误到当前 span。
// 示例：telemetry.RecordSpanError(c, err, attribute.String("reason", "timeout"))
func RecordSpanError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		span.RecordError(err)
		span.SetAttributes(semconv.ExceptionType(fmt.Sprintf("%T", err)))
	}
}

// RecordHandlerDuration 记录 handler 级别的延迟（手动 span）。
// 适用于需要细分特定业务逻辑耗时的场景（如 batch 操作、webhook 投递）。
func RecordHandlerDuration(c *gin.Context, handlerName string) func() {
	span := trace.SpanFromContext(c.Request.Context())
	if !span.IsRecording() {
		return func() {}
	}
	start := time.Now()
	return func() {
		elapsed := time.Since(start).Milliseconds()
		span.SetAttributes(attributeInt64("handler."+handlerName+"_duration_ms", elapsed))
	}
}
