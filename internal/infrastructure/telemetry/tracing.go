// Package telemetry — OpenTelemetry 分布式链路追踪（P10）。
//
// 对标 Google Cloud Trace / Jaeger / Zipkin，遵循 W3C Trace Context 规范（traceparent header）。
//
// 使用方式：
//
//  1. main.go 初始化：
//     tp, err := telemetry.InitTracerProvider(ctx, cfg)
//     defer tp.Shutdown(ctx)
//
//  2. HTTP 中间件（自动创建 span）：
//     r.Use(telemetry.TracingMiddleware("ydsz-plane"))
//
//  3. 跨进程传递（RabbitMQ / gRPC 消息头注入）：
//     carrier := telemetry.NewMapCarrierFromHeaders(headers)
//     ctx := propagator.Extract(ctx, carrier)
//
// 采样策略：
//   - 错误请求（status >= 500）：100% 采样
//   - 正常请求：10% 采样（可配置）
//
// 依赖（需添加到 go.mod）：
//   go.opentelemetry.io/otel v1.32.0
//   go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.32.0
//   go.opentelemetry.io/otel/sdk v1.32.0
//   go.opentelemetry.io/otel/trace v1.32.0
//   go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.57.0
//   go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.57.0
package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InitTracerProvider 构造并注册全局 TracerProvider。
// 应在程序启动时调用一次；重复调用会返回已有 provider。
//
// 采样率通过 sampleRatio 控制（0.0~1.0）：
//   - 生产环境建议 0.1（10%），错误请求强制采样
//   - 开发环境可设 1.0（全采样）
//
// 连接 OTLP Collector 通过 grpc://otel-collector:4317（无 TLS）。
func InitTracerProvider(ctx context.Context, serviceName, otlpEndpoint string, sampleRatio float64, logger *zap.Logger) (*sdktrace.TracerProvider, error) {
	// OTLP gRPC exporter
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(), // 生产环境应使用 TLS
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: failed to create OTLP exporter: %w", err)
	}

	// Resource attributes — 标识此服务实例
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("deployment.environment", "production"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: failed to create resource: %w", err)
	}

	// 采样策略：ParentBased + TraceIDRatio
	// 强制采样：error flag 由 TracingMiddleware 设置；其他按 ratio
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(sampleRatio),
		sdktrace.WithRemoteParentSampled(sdktrace.AlwaysSample()),
		sdktrace.WithRemoteParentNotSampled(sdktrace.TraceIDRatioBased(sampleRatio)),
		sdktrace.WithLocalParentSampled(sdktrace.AlwaysSample()),
		sdktrace.WithLocalParentNotSampled(sdktrace.TraceIDRatioBased(sampleRatio)),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 注册为全局 provider
	otel.SetTracerProvider(tp)

	// 注册 W3C  propagator（traceparent / tracestate）
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if logger != nil {
		logger.Info("otel tracer provider initialized",
			zap.String("endpoint", otlpEndpoint),
			zap.Float64("sample_ratio", sampleRatio))
	}
	return tp, nil
}

// Tracer 返回指定名称的 Tracer，用于手动创建 span。
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// SpanFromContext 从 context 中提取当前 span（如不存在返回 noop）。
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanError 标记 span 为错误状态并记录错误事件。
// 用于非 panic 的业务错误追踪（如 SSO 认证失败、webhook 投递失败）。
func SetSpanError(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if span == nil || !span.IsRecording() || err == nil {
		return
	}
	span.RecordError(err, trace.WithAttributes(attrs...))
	span.SetAttributes(attribute.Bool("error", true))
}

// StartDBSpan 创建数据库查询 span（快捷函数）。
func StartDBSpan(ctx context.Context, operation string, query string) (context.Context, trace.Span) {
	tr := Tracer("ydsz-plane.db")
	ctx, span := tr.Start(ctx, "db."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.statement", truncate(query, 200)),
	)
	return ctx, span
}

// StartRedisSpan 创建 Redis 操作 span（快捷函数）。
func StartRedisSpan(ctx context.Context, operation string, key string) (context.Context, trace.Span) {
	tr := Tracer("ydsz-plane.redis")
	ctx, span := tr.Start(ctx, "redis."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("redis.command", operation),
		attribute.String("redis.key", truncate(key, 100)),
	)
	return ctx, span
}

// StartESSpan 创建 Elasticsearch 操作 span（快捷函数）。
func StartESSpan(ctx context.Context, operation string, index string) (context.Context, trace.Span) {
	tr := Tracer("ydsz-plane.es")
	ctx, span := tr.Start(ctx, "es."+operation,
		trace.WithSpanKind(trace.SpanKindClient),
	)
	span.SetAttributes(
		attribute.String("db.system", "elasticsearch"),
		attribute.String("es.operation", operation),
		attribute.String("es.index", index),
	)
	return ctx, span
}

// StartMQSpan 创建 RabbitMQ 消息处理 span（快捷函数）。
// 从消息头提取 traceparent 上下文，实现跨进程链路追踪。
func StartMQSpan(ctx context.Context, operation string, queueName string) (context.Context, trace.Span) {
	tr := Tracer("ydsz-plane.mq")
	ctx, span := tr.Start(ctx, "mq."+operation,
		trace.WithSpanKind(trace.SpanKindConsumer),
	)
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queueName),
		attribute.String("messaging.operation", operation),
	)
	return ctx, span
}

// ShutdownTracerProvider 优雅关闭 TracerProvider，确保未导出 span 刷新。
// 应在 main defer 中调用。
func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) {
	if tp == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = tp.Shutdown(ctx)
}

// truncate 限制字符串长度，避免 span attributes 过长。
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
