/**
 * 效能度量 API — 速度、前置时间、质量指标、DORA 四指标、资源负载。
 *
 * 路由: /api/v1/workspaces/:ws/projects/:pid/metrics
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export interface SprintTrend {
  sprint_id: number;
  sprint_name: string;
  completed_count: number;
  committed_count: number;
  completion_rate: number;
  started_date?: string;
  ended_date?: string;
}

/** 速率结果 — 平均速率、迭代数、各迭代趋势。 */
export interface VelocityResult {
  project_id: number;
  average: number;
  sprint_count: number;
  trend: SprintTrend[];
}

/** 前置时间百分位数据（p50 / p85 / p95）。 */
export interface LeadTimePercentile {
  p50_hours: number;
  p85_hours: number;
  p95_hours: number;
}

/** 前置时间趋势点（周期 + p50 / p885 + 样本数）。 */
export interface LeadTimeTrend {
  period: string;
  p50_hours: number;
  p85_hours: number;
  count: number;
}

/** 前置时间分析结果（整体百分位 + 趋势序列）。 */
export interface LeadTimeResult {
  project_id: number;
  period_days: number;
  percentiles: LeadTimePercentile;
  trend: LeadTimeTrend[];
}

/** 质量指标 — 缺陷密度 / 逃逸率 / 重开率。 */
export interface QualityMetrics {
  project_id: number;
  defect_density: number;
  escape_rate: number;
  reopen_rate: number;
  total_defects: number;
  escaped_defects: number;
  period_days: number;
}

/** DORA 单指标（等级 + 值 + 单位）。 */
export interface DORAMetric {
  level: "elite" | "high" | "medium" | "low";
  value: number;
  unit: string;
}

/** DORA 四指标结果（部署频率 / 变更前置时间 / 故障恢复时间 / 变更失败率）。 */
export interface DORAResult {
  project_id: number;
  deployment_frequency: DORAMetric;
  lead_time_for_changes: DORAMetric;
  change_failure_rate: DORAMetric;
  mean_time_to_restore: DORAMetric;
}

/** 资源负载结果 — 当前 WIP / 已启动工作项数。 */
export interface ResourceLoadResult {
  project_id: number;
  active_wip: number;
  total_started_issues: number;
}

/** 指标快照（指标名 + 值 + 维度 + 快照时间）。 */
export interface MetricSnapshot {
  metric: string;
  value: number;
  dimensions: Record<string, any>;
  snapshot_date: string;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 效能度量域 API — 速度 / 前置时间 / 质量 / DORA / 资源负载 / 快照。 */
export const metricsApi = {
  getVelocity: (wsId: number, projectId: number, lastN = 6) =>
    wrap<VelocityResult>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/velocity`, {
        params: { last_n: lastN },
      }),
    ),
  getVelocityTrend: (wsId: number, projectId: number, lastN = 6) =>
    wrap<{ trend: SprintTrend[]; average: number; sprint_count: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/velocity/trend`, {
        params: { last_n: lastN },
      }),
    ),
  getLeadTime: (wsId: number, projectId: number, days = 90) =>
    wrap<LeadTimeResult>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/lead-time`, {
        params: { days },
      }),
    ),
  getQuality: (wsId: number, projectId: number) =>
    wrap<QualityMetrics>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/quality`),
    ),
  getDORA: (wsId: number, projectId: number) =>
    wrap<DORAResult>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/dora`),
    ),
  getResourceLoad: (wsId: number, projectId: number) =>
    wrap<ResourceLoadResult>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/resource-load`),
    ),
  listSnapshots: (wsId: number, projectId: number, metric?: string) =>
    wrap<MetricSnapshot[]>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/metrics/snapshots`, {
        params: metric ? { metric } : {},
      }),
    ),
};
