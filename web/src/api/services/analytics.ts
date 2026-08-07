/**
 * 缺陷分析 API — 缺陷聚合统计与明细导出。
 *
 * 对应后端：GET /api/v1/workspaces/:wsId/projects/:projectId/analytics/defects
 *           GET /api/v1/workspaces/:wsId/projects/:projectId/analytics/defects/export
 */

/** 缺陷分析过滤参数（与后端 AnalyticsQuery 对齐）。 */
export interface DefectAnalyticsParams {
  date_from?: string;
  date_to?: string;
  severity_from?: number;
  severity_to?: number;
  module_id?: number;
  version_id?: number;
}

/** 缺陷严重程度分布。 */
export interface SeverityCount {
  severity: number;
  label: string;
  count: number;
}

/** 发现阶段分布。 */
export interface PhaseCount {
  phase: string;
  count: number;
}

/** 模块分布。 */
export interface ModuleCount {
  module_id?: number;
  module_name?: string;
  count: number;
}

/** 根因分类分布。 */
export interface RootCauseCount {
  root_cause?: string;
  count: number;
}

/** 缺陷龄分布桶。 */
export interface AgeBucket {
  range: string;
  min_days: number;
  max_days: number;
  count: number;
}

/** 周趋势点。 */
export interface TrendPoint {
  week: string;
  created: number;
  resolved: number;
}

/** 缺陷分析聚合结果。 */
export interface DefectAnalytics {
  total_defects: number;
  open_defects: number;
  resolved_defects: number;
  avg_age_days: number;
  severity_dist: SeverityCount[];
  phase_dist: PhaseCount[];
  module_dist: ModuleCount[];
  root_cause_dist: RootCauseCount[];
  age_buckets: AgeBucket[];
  trend: TrendPoint[];
}

/** 将过滤参数序列化为查询串。 */
function toQuery(params?: DefectAnalyticsParams): string {
  if (!params) return "";
  const qs = new URLSearchParams();
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);
  if (params.severity_from != null) qs.set("severity_from", String(params.severity_from));
  if (params.severity_to != null) qs.set("severity_to", String(params.severity_to));
  if (params.module_id != null) qs.set("module_id", String(params.module_id));
  if (params.version_id != null) qs.set("version_id", String(params.version_id));
  const q = qs.toString();
  return q ? `?${q}` : "";
}

/** 缺陷分析 API — 缺陷聚合统计与导出（对接 /analytics/defects 系列端点）。 */
export const analyticsApi = {
  /**
   * 拉取缺陷聚合分析数据。
   */
  getDefects: (wsId: number, projectId: number, params?: DefectAnalyticsParams) =>
    fetch(`/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/defects${toQuery(params)}`, {
      credentials: "include",
    }).then(async (res) => {
      if (!res.ok) throw new Error(`加载缺陷分析失败 (${res.status})`);
      return (await res.json()) as DefectAnalytics;
    }),

  /**
   * 生成缺陷明细导出下载地址（配合 <a download> 使用）。
   *
   * @param format 导出格式：csv（默认）| xlsx
   */
  exportUrl: (wsId: number, projectId: number, format: string, params?: DefectAnalyticsParams) => {
    const qs = new URLSearchParams();
    if (params?.date_from) qs.set("date_from", params.date_from);
    if (params?.date_to) qs.set("date_to", params.date_to);
    if (params?.severity_from != null) qs.set("severity_from", String(params.severity_from));
    if (params?.severity_to != null) qs.set("severity_to", String(params.severity_to));
    if (params?.module_id != null) qs.set("module_id", String(params.module_id));
    if (params?.version_id != null) qs.set("version_id", String(params.version_id));
    if (format) qs.set("format", format);
    const q = qs.toString();
    return `/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/defects/export${q ? `?${q}` : ""}`;
  },
};
