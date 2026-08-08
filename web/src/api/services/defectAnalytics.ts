/**
 * 缺陷分析 API — 封装缺陷聚合统计、趋势、交叉分析等接口。
 *
 * 所有路径挂载在 /api/v1/workspaces/:ws/projects/:pid/analytics 下。
 */
import { apiClient } from '../client';

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 单维度统计基线 */
export interface SeverityCount {
  severity: number;
  label: string;
  count: number;
}

/** 缺陷发现阶段分布（phase → count）。 */
export interface PhaseCount {
  phase: string;
  count: number;
}

/** 缺陷模块分布（module_id / module_name → count）。 */
export interface ModuleCount {
  module_id?: number;
  module_name?: string;
  count: number;
}

/** 缺陷根因分类分布（root_cause → count）。 */
export interface RootCauseCount {
  root_cause?: string;
  count: number;
}

/** 缺陷龄桶 */
export interface AgeBucket {
  range: string;
  min_days: number;
  max_days: number;
  count: number;
}

/** 趋势数据点 */
export interface TrendPoint {
  date: string;
  opened: number;
  resolved: number;
}

/** 缺陷分析聚合结果 */
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

/** 查询参数 */
export interface AnalyticsQuery {
  date_from?: string;
  date_to?: string;
  severity_from?: number;
  severity_to?: number;
  phase?: string;
  module_id?: number;
}

/** 严重程度标签映射（对齐后端约定） */
export const SEVERITY_LABELS: Record<number, string> = {
  1: '致命',
  2: '严重',
  3: '一般',
  4: '轻微',
  5: '建议',
};

/** 发现阶段标签映射 */
export const PHASE_LABELS: Record<string, string> = {
  requirements: '需求阶段',
  design: '设计阶段',
  coding: '编码阶段',
  testing: '测试阶段',
  staging: '预发布',
  production: '生产环境',
  uat: 'UAT',
  other: '其他',
};

// ---------- 缺陷龄分析类型 ----------

/** 单状态缺陷龄统计 */
export interface DefectAgeStat {
  state_name: string;
  state_group: string;
  count: number;
  min_days: number;
  max_days: number;
  avg_days: number;
  median_days: number;
}

/** 超阈值缺陷明细 */
export interface DefectAgeItem {
  id: number;
  identifier: string;
  name: string;
  severity?: number;
  severity_label: string;
  state_name: string;
  age_days: number;
}

/** 缺陷龄分析聚合结果 */
export interface DefectAgeAnalysis {
  state_stats: DefectAgeStat[];
  overdue_count: number;
  overdue_items: DefectAgeItem[];
}

// ---------- 根因分布类型 ----------

/** 根因分类统计 */
export interface RootCauseStat {
  root_cause: string;
  count: number;
  percentage: number;
}

/** 根因分布分析结果 */
export interface RootCauseAnalysis {
  items: RootCauseStat[];
  total_count: number;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

export const defectAnalyticsApi = {
  /**
   * 获取缺陷分析聚合数据。
   */
  async getAnalytics(wsId: number, projectId: number, query: AnalyticsQuery = {}): Promise<DefectAnalytics> {
    const params = new URLSearchParams();
    if (query.date_from) params.set('date_from', query.date_from);
    if (query.date_to) params.set('date_to', query.date_to);
    if (query.severity_from != null) params.set('severity_from', String(query.severity_from));
    if (query.severity_to != null) params.set('severity_to', String(query.severity_to));
    if (query.phase) params.set('phase', query.phase);
    if (query.module_id != null) params.set('module_id', String(query.module_id));

    const qs = params.toString();
    const { data } = await apiClient.get<DefectAnalytics>(
      `/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/defects${qs ? '?' + qs : ''}`,
    );
    return data;
  },

  /**
   * 获取缺陷龄分析数据 — 按状态分组统计未关闭缺陷的滞留时长。
   */
  async getDefectAge(wsId: number, projectId: number, query: AnalyticsQuery = {}): Promise<DefectAgeAnalysis> {
    const params = new URLSearchParams();
    if (query.date_from) params.set('date_from', query.date_from);
    if (query.date_to) params.set('date_to', query.date_to);
    if (query.severity_from != null) params.set('severity_from', String(query.severity_from));
    if (query.severity_to != null) params.set('severity_to', String(query.severity_to));

    const qs = params.toString();
    const { data } = await apiClient.get<DefectAgeAnalysis>(
      `/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/defect-age${qs ? '?' + qs : ''}`,
    );
    return data;
  },

  /**
   * 获取根因分布统计 — 按 root_cause_category 分组统计缺陷数量和占比。
   */
  async getRootCause(wsId: number, projectId: number, query: AnalyticsQuery = {}): Promise<RootCauseAnalysis> {
    const params = new URLSearchParams();
    if (query.date_from) params.set('date_from', query.date_from);
    if (query.date_to) params.set('date_to', query.date_to);

    const qs = params.toString();
    const { data } = await apiClient.get<RootCauseAnalysis>(
      `/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/root-cause${qs ? '?' + qs : ''}`,
    );
    return data;
  },

  /**
   * 导出缺陷明细（CSV）。
   */
  async exportDefects(wsId: number, projectId: number, query: AnalyticsQuery = {}, format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> {
    const params = new URLSearchParams();
    if (query.date_from) params.set('date_from', query.date_from);
    if (query.date_to) params.set('date_to', query.date_to);
    if (query.severity_from != null) params.set('severity_from', String(query.severity_from));
    if (query.severity_to != null) params.set('severity_to', String(query.severity_to));
    params.set('format', format);

    const { data } = await apiClient.get<Blob>(
      `/api/v1/workspaces/${wsId}/projects/${projectId}/analytics/defects/export?${params.toString()}`,
      { responseType: 'blob' },
    );
    return data;
  },
};
