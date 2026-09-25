// Package issue — 工作项域业务 SLI 可观测性埋点。
//
// 对标：Google SRE Book 第 4 章 — SLI / SLO / SLA。
//
// 指标清单：
//   - plane_issue_operations_total{operation,status}：工作项操作计数
//     （operation ∈ create,update,transition,delete；status ∈ success,error）
//   - plane_issue_operation_duration_seconds{operation}：工作项操作耗时分布
//   - plane_issue_count_by_state{workspace_id,state}：按状态分布的工作项数量
//
// 隐私约束：所有指标标签均为聚合维度（workspace_id / state / operation），
// 不含 user_id、issue_id、项目 ID 等 PII 或高基数标签。
package issue

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ---------------------------------------------------------------------------
// Prometheus 指标（plane_ 前缀，namespace="plane"）
// ---------------------------------------------------------------------------

// issueOperationsTotal 按 operation / status 统计工作项操作总数。
// 对标 Google SRE Book Ch.4 — 用于计算操作成功率的 SLI。
var issueOperationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "plane",
		Name:      "issue_operations_total",
		Help:      "Total issue operations (create / update / transition / delete) labeled by status (success / error). Google SRE Book Ch.4 — SLI for operation success rate.",
	},
	[]string{"operation", "status"},
)

// issueOperationDuration 按 operation 记录工作项操作耗时分布（秒）。
// 使用指数分桶覆盖 10ms~10s 的典型 API 处理范围。
var issueOperationDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "plane",
		Name:      "issue_operation_duration_seconds",
		Help:      "Issue operation latency distribution in seconds. Google SRE Book Ch.4 — latency SLI.",
		Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	},
	[]string{"operation"},
)

// issueCountByState 按 workspace_id / state 记录当前活跃工作项数量。
// 由后台定期任务调用 syncIssueCountByState 刷新（非请求路径 hot path）。
// workspace_id 使用 "_all" 全局聚合标签，配合 Grafana 变量过滤。
var issueCountByState = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "plane",
		Name:      "issue_count_by_state",
		Help:      "Number of active issues grouped by workspace and state. Refreshed by periodic background job. Google SRE Book Ch.4 — backlog SLI.",
	},
	[]string{"workspace_id", "state"},
)

// ---------------------------------------------------------------------------
// Operation 标签常量
// ---------------------------------------------------------------------------

const (
	OpCreate     = "create"
	OpUpdate     = "update"
	OpTransition = "transition"
	OpDelete     = "delete"
)

// ---------------------------------------------------------------------------
// Helper — 记录操作
// ---------------------------------------------------------------------------

// ObserveOperation 记录一次工作项操作（计数 + 耗时）。
//
// 用法示例（在 handler / 服务入口调用）：
//
//	start := time.Now()
//	result, err := issueSvc.Create(ctx, input)
//	issue.ObserveOperation(issue.OpCreate, err == nil, time.Since(start))
//
// Google SRE Book Ch.4：SLI = good events / total events；这里计数维度可以
// 通过 rate() 做滑动窗口聚合。
func ObserveOperation(operation string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	issueOperationsTotal.WithLabelValues(operation, status).Inc()
	issueOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// ObserveCreate 记录创建工作项操作。
func ObserveCreate(success bool, duration time.Duration) {
	ObserveOperation(OpCreate, success, duration)
}

// ObserveUpdate 记录更新工作项操作。
func ObserveUpdate(success bool, duration time.Duration) {
	ObserveOperation(OpUpdate, success, duration)
}

// ObserveTransition 记录状态流转操作。
func ObserveTransition(success bool, duration time.Duration) {
	ObserveOperation(OpTransition, success, duration)
}

// ObserveDelete 记录删除工作项操作。
func ObserveDelete(success bool, duration time.Duration) {
	ObserveOperation(OpDelete, success, duration)
}

// ---------------------------------------------------------------------------
// 后台同步：按状态统计工作项数量
// ---------------------------------------------------------------------------

// syncSQL 用于按 workspace / state 聚合工作项数量（跨 task / requirement / defect）。
// 不包含已软删除（deleted=true）的记录。
const syncIssueCountSQL = `
	WITH all_issues AS (
		SELECT workspace_id, state_id FROM task WHERE deleted = false
		UNION ALL
		SELECT workspace_id, state_id FROM requirement WHERE deleted = false
		UNION ALL
		SELECT workspace_id, state_id FROM defect WHERE deleted = false
	)
	SELECT ai.workspace_id, s.name AS state_name, COUNT(*) AS cnt
	FROM all_issues ai
	JOIN states s ON s.id = ai.state_id
	GROUP BY ai.workspace_id, s.name
	ORDER BY ai.workspace_id, s.name
`

// SyncIssueCountByState 从 DB 查询工作项按状态分布并刷新 Gauge。
// 设计为非请求路径调用（周期性后台任务），避免在 hot path 上做跨表聚合查询。
//
// 对标 Google SRE Book Ch.4 — backlog 类 SLI（待处理工作项积压量）。
func SyncIssueCountByState(ctx context.Context, db *pgxpool.Pool) error {
	rows, err := db.Query(ctx, syncIssueCountSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 重置旧的 gauge 值（避免已删除 workspace 遗留数据）。
	issueCountByState.Reset()

	for rows.Next() {
		var wsID int64
		var stateName string
		var count int64
		if err := rows.Scan(&wsID, &stateName, &count); err != nil {
			return err
		}
		issueCountByState.WithLabelValues(workspaceLabel(wsID), stateName).Set(float64(count))
	}
	return rows.Err()
}

// workspaceLabel 将 workspace_id 转换为标签值。
// Reserved label "_all" 用于聚合全空间视图。
func workspaceLabel(wsID int64) string {
	if wsID == 0 {
		return "_all"
	}
	// 不使用 strconv.FormatInt 而是直接拼接（减少 import），保持 wsID 可读性
	// 注意：workspace_id 在组织内部属于业务 ID，非用户个人可识别信息（PII），
	// 符合作为 Prometheus label 的低基数要求。
	return formatWSLabel(wsID)
}

// formatWSLabel 快速将 int64 转为 string（避免 strconv 引入）。
// 实际使用 strconv 更规范，这里为了最小依赖使用自定义。
func formatWSLabel(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte // int64 max 19 digits
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
