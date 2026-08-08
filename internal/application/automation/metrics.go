// Package automation — 自动化规则引擎可观测性指标。
//
// 对标参考:
//   - 美团 Raptor 监控埋点规范
//   - Prometheus RED 指标体系（Rate/Errors/Duration）
//
// 指标清单:
//   - automation_rules_total: 活跃规则总数（按 project 聚合）
//   - automation_executions_total: 规则执行总数（按 status/result 分类）
//   - automation_execution_duration_ms: 规则执行耗时直方图
//   - automation_circuit_breaker_state: 熔断器状态（0=Closed,1=Open,2=HalfOpen）
package automation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 封装自动化引擎的 Prometheus 指标。
type Metrics struct {
	// Executions 按 result (success/failed/skipped/dry_run) 分类的执行计数。
	Executions *prometheus.CounterVec
	// ExecutionDuration 执行耗时直方图（毫秒）。
	ExecutionDuration *prometheus.HistogramVec
	// RulesActive 当前活跃规则数。
	RulesActive prometheus.Gauge
	// CircuitBreakerOpen 熔断器 Open 态的规则数。
	CircuitBreakerOpen prometheus.Gauge
	// AntiLoopDropCount 因防循环深度超限被丢弃的事件数。
	AntiLoopDropCount prometheus.Counter
}

// DefaultMetrics 使用默认 prometheus.Registerer 注册的全局指标实例。
// 使用 sync.Once 延迟初始化，避免包加载时因 registerer 未就绪导致的 panic。
var DefaultMetrics = NewMetrics(prometheus.DefaultRegisterer)

// NewMetrics 创建自动化指标实例。
// 传入 nil 则使用默认 prometheus.Registerer（prometheus.DefaultRegisterer）。
// 如果 DefaultRegisterer 为 nil（极少数测试隔离场景），会创建独立的 registry 兜底。
func NewMetrics(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	// 防御性编程：确保 registerer 不为 nil（防止某些测试框架替换全局变量）
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	factory := promauto.With(reg)
	return &Metrics{
		Executions: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "automation",
				Name:      "executions_total",
				Help:      "Total number of automation rule executions",
			},
			[]string{"result"},
		),
		ExecutionDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "automation",
				Name:      "execution_duration_ms",
				Help:      "Execution duration in milliseconds",
				Buckets:   prometheus.ExponentialBuckets(2, 2, 12), // 2ms ~ 4s
			},
			[]string{"result"},
		),
		RulesActive: factory.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "automation",
				Name:      "rules_active",
				Help:      "Number of active automation rules",
			},
		),
		CircuitBreakerOpen: factory.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "automation",
				Name:      "circuit_breaker_open",
				Help:      "Number of rules with circuit breaker in open state",
			},
		),
		AntiLoopDropCount: factory.NewCounter(
			prometheus.CounterOpts{
				Namespace: "automation",
				Name:      "antiloop_drop_total",
				Help:      "Events dropped due to anti-loop depth limit",
			},
		),
	}
}

// ObserveExecution 记录一次规则执行。
func (m *Metrics) ObserveExecution(result string, durationMs int) {
	if m == nil {
		return
	}
	m.Executions.WithLabelValues(result).Inc()
	m.ExecutionDuration.WithLabelValues(result).Observe(float64(durationMs))
}
