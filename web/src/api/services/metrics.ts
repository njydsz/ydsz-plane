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

export interface VelocityResult {
  project_id: number;
  average: number;
  sprint_count: number;
  trend: SprintTrend[];
}

export interface LeadTimePercentile {
  p50_hours: number;
  p85_hours: number;
  p95_hours: number;
}

export interface LeadTimeTrend {
  period: string;
  p50_hours: number;
  p85_hours: number;
  count: number;
}

export interface LeadTimeResult {
  project_id: number;
  period_days: number;
  percentiles: LeadTimePercentile;
  trend: LeadTimeTrend[];
}

export interface QualityMetrics {
  project_id: number;
  defect_density: number;
  escape_rate: number;
  reopen_rate: number;
  total_defects: number;
  escaped_defects: number;
  period_days: number;
}

export interface DORAMetric {
  level: "elite" | "high" | "medium" | "low";
  value: number;
  unit: string;
}

export interface DORAResult {
  project_id: number;
  deployment_frequency: DORAMetric;
  lead_time_for_changes: DORAMetric;
  change_failure_rate: DORAMetric;
  mean_time_to_restore: DORAMetric;
}

export interface ResourceLoadResult {
  project_id: number;
  active_wip: number;
  total_started_issues: number;
}

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
