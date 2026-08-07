/**
 * 甘特图 API — 获取项目工作项时间线数据与依赖关系。
 *
 * 注：v0.2 使用 issue list 接口获取时间线数据（dates + basic fields），
 * 依赖箭头由 relation 数据在后续迭代补充。
 * 路径：复用 /api/v1/workspaces/:ws/projects/:pid/issues（带日期过滤）
 */
import { apiClient } from '../client';
import { issueApi } from './issue';

/** 甘特图工作项条目（精简，仅含时间线必要字段） */
export interface GanttIssue {
  id: number;
  identifier: string;
  name: string;
  type_code: 'requirement' | 'task' | 'defect';
  state?: {
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

/** 依赖箭头（预留，v0.3 启用） */
export interface GanttDependency {
  id: number;
  source_id: number;
  target_id: number;
  type: 'fs' | 'ss' | 'ff' | 'sf';
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
   * 获取项目甘特图数据（工作项时间线，无日期的工作项也会返回）。
   * v0.2 版本依赖关系暂不使用。
   */
  async getGanttData(wsId: number, projectId: number, query: GanttQuery = {}): Promise<GanttData> {
    // 加载全量工作项（前端按日期过滤渲染）
    const result = await issueApi.listIssues(wsId, projectId, {
      start_date_from: query.date_from,
      target_date_to: query.date_to,
      limit: 500,
      sprint_id: query.sprint_id,
    });

    const issues: GanttIssue[] = result.results.map((item) => ({
      id: item.id,
      identifier: item.identifier,
      name: item.name,
      type_code: item.type_code,
      state: item.state ?? undefined,
      priority: item.priority,
      progress: item.progress,
      start_date: item.start_date ?? undefined,
      target_date: item.target_date ?? undefined,
      sprint_id: item.sprint_id ?? undefined,
      parent_id: item.parent_id ?? undefined,
    }));

    // 通过父工作项派生 FS(完成→开始)依赖箭头,与甘特图渲染对齐。
    const byId = new Map(issues.map((i) => [i.id, i]));
    const dependencies: GanttDependency[] = [];
    let depSeq = 1;
    for (const child of issues) {
      if (child.parent_id != null && byId.has(child.parent_id)) {
        dependencies.push({
          id: depSeq++,
          source_id: child.parent_id,
          target_id: child.id,
          type: 'fs',
        });
      }
    }

    return { issues, dependencies };
  },
};
