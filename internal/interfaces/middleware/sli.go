// Package middleware — HTTP API SLI 指标（业务层）。
//
// 对标：Google SRE Book 第 4 章 — SLI / SLO / SLA。
//
// 指标清单：
//   - plane_api_errors_total{method,path,error_code}：业务错误总数（从 ctx 读取 ErrorCode）
//   - plane_api_sli_available_budget：可用错误预算（1 - 错误率），可由外部定时器通过
//     SetSLIBudget 写入；Grafana 面板同时提供 recomputation 版本
//
// 隐私约束：error_code 仅使用 errs.AppError.Code，不含任何 PII（如 user_id）。
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ---------------------------------------------------------------------------
// Prometheus 指标（plane_ 前缀，namespace="plane"）
// ---------------------------------------------------------------------------

// apiErrorsTotal 按 method / path / error_code 维度统计业务错误总数。
// error_code 来自 errs.AppError.Code（如 ISSUE.VERSION_CONFLICT），不含 PII。
var apiErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: planeNamespace,
		Name:      "api_errors_total",
		Help:      "Total business-validation errors, labeled by method, path, and error code. Used for SLI/SLO error-rate computation.",
	},
	[]string{"method", "path", "error_code"},
)

// apiSLIAvailableBudget 是可用错误预算 Gauge（1 - 实际错误率）。
// 理想情况下由外部定时任务根据 SLI 查询结果写入（如 30s 周期）；
// Grafana 面板直接使用 rate-based query 版本作为兜底显示。
//
// 参考：Google SRE Book Ch.4 — 错误预算 = 1 - SLO 目标可用性（如 99.9%）。
var apiSLIAvailableBudget = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: planeNamespace,
		Name:      "api_sli_available_budget",
		Help:      "Available error budget (1 - error rate). Target: 99.9% availability (budget >= 0). Updated by external SLI recorder or via SetSLIBudget.",
	},
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// RecordAPISLI 在 HTTP 处理链末端记录 SLI 指标。
//
// 通常在 MetricsMiddleware 或 Latency 之后、业务 handler 之外的中间件调用，
// 以从 gin.Context 中读取 handler 写入的 ErrorCode（由 respondError 注入）。
//
// 安全约束：error_code 维度仅来自 errs.AppError.Code、不含 user_id 等 PII 标签。
func RecordAPISLI(c *gin.Context) {
	if c.Writer.Status() < 400 {
		return
	}
	route := c.FullPath()
	if route == "" {
		route = "unmatched"
	}
	method := c.Request.Method

	// 从 ctx 读取 ErrorCode；handler 层通过 SetErrorCode 注入（默认 "internal"）。
	errorCode := "internal"
	if raw, exists := c.Get(CtxErrorCode); exists {
		if code, ok := raw.(string); ok && code != "" {
			errorCode = code
		}
	}

	apiErrorsTotal.WithLabelValues(method, route, errorCode).Inc()
}

// SetSLIBudget 更新可用错误预算 Gauge（由外部定时任务写入）。
// budget 取值范围 (-∞, 1.0]；通常 1 - error_rate；负数表示已突破 SLO 目标。
func SetSLIBudget(budget float64) {
	apiSLIAvailableBudget.Set(budget)
}

// GetAPISLIErrorCounter 暴露底层 CounterVec，供外部 recorder 直接使用。
// 用于在独立 goroutine 中计算 budget 时读取增量。
func GetAPISLIErrorCounter() *prometheus.CounterVec {
	return apiErrorsTotal
}

// GetAPIBudgetGauge 暴露底层 Gauge，供 recorder 写入 budget。
func GetAPIBudgetGauge() prometheus.Gauge {
	return apiSLIAvailableBudget
}

// MetricsSLIMiddleware 在每个 HTTP 请求结束后记录 SLI 指标。
// 它只应在 Latency 中间件之后（或替换 Latency）使用，便于读取 ctx 中的 ErrorCode。
//
// 典型装配方式（router.go）：
//
//	engine.Use(
//	  middleware.Latency(log),
//	  middleware.MetricsSLIMiddleware(),
//	)
func MetricsSLIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		RecordAPISLI(c)
	}
}
