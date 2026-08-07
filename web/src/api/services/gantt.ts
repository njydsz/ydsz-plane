/**
 * 甘特图 API — 获取项目工作项时间线数据与依赖关系。
 *
 * 路径：/api/v1/workspaces/:ws/projects/:pid/gantt
 */
import { apiClient } from '../client';

/** 甘特图工作项条目 */
export interface GanttIssue {
  id: number;
  identifier: string;
  name: string;
  type_code: 'requirement' | 'task' | 'defect';
  state: {
    id: number;
    name: string;
    group: string;
    color: string;
  };
  priority: string;
  progress: number;
  start_date?: string;
  target_date?: string;
  sprint_id?: number;
  sprint_name?: string;
  parent_id?: number;
}

/** 依赖箭头（前置 → 后继） */
export interface GanttDependency {
  id: number;
  source_id: number;  // 前置工作项
  target_id: number;  // 后继工作项
  type: 'fs' | 'ss' | 'ff' | 'sf'; // finish-to-start 等
}

/** 甘特图全量数据 */
export interface GanttData {
  issues: GanttIssue[];
  dependencies: GanttDependency[];
}

/** 查询参数 */
export interface GanttQuery {
  date_from?: string;
  date_to?: string;
  sprint_id?: number;
}

export const ganttApi = {
  /**
   * 获取项目甘特图数据（工作项 + 依赖）。
   */
  async getGanttData(wsId: number, projectId: number, query: GanttQuery = {}): Promise<GanttData> {
    const params = new URLSearchParams();
    if (query.date_from) params.set('date_from', query.date_from);
    if (query.date_to) params.set('date_to', query.date_to);
    if (query.sprint_id != null) params.set('sprint_id', String(query.sprint_id));

    const qs = params.toString();
    const { data } = await apiClient.get<GanttData>(
      `/api/v1/workspaces/${wsId}/projects/${projectId}/gantt${qs ? '?' + qs : ''}`,
    );
    return data;
  },
};
