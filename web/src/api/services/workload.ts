/**
 * 工作量热力图 API — 获取项目成员 × 日期的工时分布数据。
 *
 * 路径：GET /api/v1/workspaces/:ws/projects/:pid/workload-heatmap
 */
import { apiClient } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 单成员单日工时汇总 */
export interface WorkloadHeatmapEntry {
  user_id: number;
  spent_date: string;
  total_minutes: number;
  total_hours: number;
  issue_count: number;
  log_count: number;
}

/** 成员工时汇总 */
export interface WorkloadMember {
  user_id: number;
  total_hours: number;
  day_count: number;
}

/** 项目工时统计概览 */
export interface WorkloadSummary {
  total_hours: number;
  total_members: number;
  total_days: number;
  daily_average_hours: number;
}

/** 热力图全量数据 */
export interface WorkloadHeatmapData {
  entries: WorkloadHeatmapEntry[];
  members: WorkloadMember[];
  summary: WorkloadSummary;
  date_from: string;
  date_to: string;
}

/** 查询参数 */
export interface WorkloadHeatmapQuery {
  date_from?: string;
  date_to?: string;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 工作量热力图 API */
export const workloadApi = {
  /**
   * 获取项目工时热力图数据（成员 × 日期）。
   *
   * @param wsId 工作空间 ID
   * @param projectId 项目 ID
   * @param query 可选日期范围（默认最近 30 天）
   */
  async getHeatmap(
    wsId: number,
    projectId: number,
    query?: WorkloadHeatmapQuery,
  ): Promise<WorkloadHeatmapData> {
    const qs = new URLSearchParams();
    if (query?.date_from) qs.set("date_from", query.date_from);
    if (query?.date_to) qs.set("date_to", query.date_to);
    const q = qs.toString();
    const url = `/api/v1/workspaces/${wsId}/projects/${projectId}/workload-heatmap${q ? `?${q}` : ""}`;
    const { data } = await apiClient.get<WorkloadHeatmapData>(url);
    return data;
  },
};
