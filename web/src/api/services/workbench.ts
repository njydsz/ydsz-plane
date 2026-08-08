/**
 * 工作台服务 — 对接后端 Workbench 域 REST 接口。
 * 后端：internal/application/workbench
 */
import { apiClient } from "../client";

/* ------------------------------------------------------------------ */
/* Types (mirror backend models)                                      */
/* ------------------------------------------------------------------ */

/** 工作台首屏聚合数据 */
export interface WorkbenchSummary {
  my_issues: MyIssuesBucket;
  sprint_overviews: SprintOverview[];
  recent_items: RecentItem[];
  overdue_count: number;
  blocked_count: number;
  quick_actions: QuickActionSet;
}

/** 我的工作项分桶视图 */
export interface MyIssuesBucket {
  total: number;
  today: IssueDigest[];
  upcoming: IssueDigest[];
  overdue: IssueDigest[];
  in_progress: IssueDigest[];
  backlog: IssueDigest[];
}

/** 工作项工作台摘要 */
export interface IssueDigest {
  id: number;
  identifier: string;
  title: string;
  type_code: string;
  priority: string;
  state_id: number;
  state_name: string;
  state_color: string;
  group_id: number;
  project_name: string;
  sprint_id?: number | null;
  sprint_name?: string;
  target_date?: string | null;
  is_blocked: boolean;
}

/** 迭代工作台概览 */
export interface SprintOverview {
  sprint_id: number;
  sprint_name: string;
  project_id: number;
  project_name: string;
  status: string;
  progress: number;
  my_issue_count: number;
  days_remaining: number;
  goal: string;
}

/** 最近访问条目 */
export interface RecentItem {
  item_type: string;
  item_id: number;
  project_id: number;
  title: string;
  identifier?: string;
  accessed_at: string;
  url: string;
}

/** 快捷操作入口 */
export interface QuickActionSet {
  can_create_issue: boolean;
  can_start_sprint: boolean;
  active_issue_count: number;
}

/** 工作台布局配置 */
export interface WorkbenchConfig {
  layout: LayoutConfig;
  widget_states: Record<string, unknown>;
  focus_enabled: boolean;
}

/** 拖拽布局配置 */
export interface LayoutConfig {
  widgets: LayoutWidget[];
}

/** 单个 Widget 布局 */
export interface LayoutWidget {
  type: string;
  w: number;
  h: number;
  x: number;
  y: number;
}

/** 周度趋势数据点 */
export interface WeeklyTrend {
  week: string;
  count: number;
  points: number;
}

/** 个人效率报告 */
export interface EfficiencyReport {
  week_points: number;
  week_issues: number;
  week_hours: number;
  overdue_count: number;
  weekly_trend: WeeklyTrend[];
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 工作台域 API — 聚合汇总首屏数据 + 工作台配置读写 + 布局保存。 */
export const workbenchApi = {
  /** 获取工作空间级工作台汇总（首屏数据） */
  getSummary: (wsId: number | string) =>
    wrap<WorkbenchSummary>(
      apiClient.get(`/workspaces/${wsId}/workbench/summary`),
    ),

  /** 获取工作台配置 */
  getConfig: (wsId: number | string) =>
    wrap<WorkbenchConfig>(
      apiClient.get(`/workspaces/${wsId}/workbench/config`),
    ),

  /** 保存工作台配置 */
  saveConfig: (wsId: number | string, config: Partial<WorkbenchConfig>): Promise<void> =>
    apiClient.put(`/workspaces/${wsId}/workbench/config`, config).then(() => {}),

  /** 获取最近访问 */
  getRecent: (wsId: number | string) =>
    wrap<RecentItem[]>(
      apiClient.get(`/workspaces/${wsId}/workbench/recent`),
    ),

  /** 记录访问 */
  recordRecent: (
    wsId: number | string,
    item: { item_type: string; item_id: number; project_id: number; title: string; identifier?: string },
  ): Promise<void> =>
    apiClient.post(`/workspaces/${wsId}/workbench/recent`, item).then(() => {}),

  /** 获取个人效率报告（工作空间级） */
  getEfficiency: (wsId: number | string) =>
    wrap<EfficiencyReport>(
      apiClient.get(`/workspaces/${wsId}/workbench/efficiency`),
    ),
};
