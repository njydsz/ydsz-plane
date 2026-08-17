/**
 * 甘特图 API — 获取项目需求/任务/缺陷时间线数据与依赖关系。
 *
 * v0.3+ 使用 issue_dependencies 表获取真实依赖数据（FS/SS/FF/SF）。
 * 路径：复用 /api/v1/workspaces/:ws/projects/:pid/issues（带日期过滤）
 */
import { issueApi, type IssueType, type IssueDependency } from './issue';

/** 甘特图需求/任务/缺陷条目（精简，仅含时间线必要字段） */
export interface GanttIssue {
  id: number;
  identifier: string;
  name: string;
  type_code: IssueType;
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

/** 甘特图域 API — 获取项目需求/任务/缺陷时间线数据与依赖关系（v0.3+）。 */
export const ganttApi = {
  /**
   * 获取项目甘特图数据（需求/任务/缺陷时间线 + 真实依赖关系）。
   * v0.3+ 使用 issue_dependencies 表获取 FS/SS/FF/SF 四种依赖类型。
   */
  async getGanttData(wsId: number, projectId: number, query: GanttQuery = {}): Promise<GanttData> {
    // 并行加载工作项和依赖关系
    const [result, depResult] = await Promise.all([
      issueApi.listIssues(wsId, projectId, {
        start_date_from: query.date_from,
        target_date_to: query.date_to,
        limit: 500,
        sprint_id: query.sprint_id,
      }),
      issueApi.listProjectDependencies(wsId, projectId).catch(() => ({ results: [] as IssueDependency[] })),
    ]);

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

    // 使用真实依赖数据（FS/SS/FF/SF）
    const issueIds = new Set(issues.map((i) => i.id));
    const dependencies: GanttDependency[] = depResult.results
      .filter((d) => issueIds.has(d.issue_id) && issueIds.has(d.depends_on_id))
      .map((d, idx) => ({
        id: idx + 1,
        source_id: d.depends_on_id,
        target_id: d.issue_id,
        type: d.dependency_type,
      }));

    return { issues, dependencies };
  },
};
