/**
 * 仪表盘 API — 封装仪表盘总览、Widget CRUD、告警、模板等接口。
 *
 * 所有路径挂载在 /api/v1/workspaces/:ws/projects/:pid/dashboard 下。
 */
import { apiClient } from '../client';

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type WidgetType =
  | "progress_overview"
  | "burndown"
  | "velocity"
  | "priority_split"
  | "state_distribution"
  | "overdue_list"
  | "blocked_list"
  | "risk_alert"
  | "recent_activity"
  | "team_workload"
  | "version_burndown"
  | "module_distribution";

/** 仪表盘 Widget 实体（位置、尺寸、配置、显隐状态）。 */
export interface DashboardWidget {
  id: number;
  project_id: number;
  widget_type: WidgetType;
  title: string;
  grid_x: number;
  grid_y: number;
  grid_w: number;
  grid_h: number;
  config: Record<string, any>;
  is_visible: boolean;
  sort_order: number;
}

/** 进度总览数据 — 各状态工作项计数与完成率。 */
export interface ProgressOverviewData {
  total_issues: number;
  done_issues: number;
  in_progress: number;
  overdue_issues: number;
  blocked_issues: number;
  completion_rate: number;
  active_sprints: number;
}

/** 燃尽图迭代级数据（点数与工作项数进度）。 */
export interface BurndownData {
  sprint_id: number;
  sprint_name: string;
  total_points: number;
  burned_points: number;
  total_issues: number;
  burned_issues: number;
  remaining_days: number;
  is_active: boolean;
}

/** 优先级分布数据 — total + 各优先级计数。 */
export interface PrioritySplitData {
  total: number;
  by_priority: Record<string, number>;
}

/** 状态分布数据 — 各状态（含分组 / 颜色）的工作项计数。 */
export interface StateDistributionData {
  total: number;
  by_state: Array<{
    state_id: number;
    state_name: string;
    group_name: string;
    color: string;
    count: number;
  }>;
}

/** 速率数据 — 多迭代的完成 / 承诺计数与完成率。 */
export interface VelocityData {
  average: number;
  sprints: Array<{
    sprint_id: number;
    sprint_name: string;
    completed_count: number;
    committed_count: number;
    completion_rate: number;
  }>;
}

/** 逾期工作项条目（逾期天数 + 优先级 + 指派人）。 */
export interface OverdueItem {
  id: number;
  identifier: string;
  title: string;
  priority: string;
  overdue_days: number;
  assignee: string;
}

/** 逾期工作项列表（total + items）。 */
export interface OverdueListData {
  total: number;
  items: OverdueItem[];
}

/** 阻塞工作项条目（阻塞方 + 阻塞引用数）。 */
export interface BlockedItem {
  id: number;
  identifier: string;
  title: string;
  blocked_count: number;
  blocker_names: string;
}

/** 阻塞工作项列表（total + items）。 */
export interface BlockedListData {
  total: number;
  items: BlockedItem[];
}

/** 团队成员工作量（各状态工作项计数）。 */
export interface TeamMemberWorkload {
  user_id: number;
  user_name: string;
  avatar?: string;
  todo: number;
  in_progress: number;
  done: number;
  total: number;
}

/** 团队负载 widget 数据 — 成员工作量汇总集合。 */
export interface TeamWorkloadWidgetData {
  members: TeamMemberWorkload[];
}

/** 活动日志条目（领域事件 actor + verb + target）。 */
export interface ActivityItem {
  id: number;
  issue_id: number;
  issue_identifier: string;
  actor_id: number;
  actor_name: string;
  actor_avatar?: string;
  verb: string;
  target_state?: string;
  created_at: string;
}

/** 最近活动数据 — 活动条目集合。 */
export interface RecentActivityData {
  items: ActivityItem[];
}

/** 风险告警实体 — 触发规则 + 严重级 + 描述 + 解决状态。 */
export interface RiskAlert {
  id: number;
  project_id?: number;
  rule_id: number;
  severity: "info" | "low" | "medium" | "high" | "critical";
  title: string;
  description: string;
  metadata: Record<string, any>;
  is_resolved: boolean;
  created_at: string;
}

/** 仪表盘模板 — 预置 widget 布局与配置（agile/waterfall/generic 等）。 */
export interface DashboardTemplate {
  id: number;
  name: string;
  slug: string;
  description: string;
  layout: Record<string, any>;
  icon: string;
  category: string;
  is_default: boolean;
  sort_order: number;
}

/** 仪表盘全量数据聚合（widgets + snapshots + alerts）。 */
export interface DashboardData {
  widgets: DashboardWidget[];
  snapshots: Record<string, any>;
  alerts: RiskAlert[];
}

/** 创建 widget 入参 — 类型、标题、网格位置与配置。 */
export interface CreateWidgetInput {
  widget_type: WidgetType;
  title: string;
  grid_x?: number;
  grid_y?: number;
  grid_w?: number;
  grid_h?: number;
  config?: Record<string, any>;
}

/** 更新 Widget 入参 — 仅网格位置、尺寸与配置可编辑。 */
export interface UpdateWidgetInput {
  grid_x?: number;
  grid_y?: number;
  grid_w?: number;
  grid_h?: number;
  config?: Record<string, any>;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 仪表盘域 API — 总览、widget CRUD、告警、模板、快照等。 */
export const dashboardApi = {
  /** 获取仪表盘总览（widgets + 快照数据 + 告警） */
  getOverview: (wsId: number | string, projectId: number) =>
    wrap<DashboardData>(
      apiClient.get(`/workspaces/${wsId}/projects/${projectId}/dashboard`),
    ),

  /** 列出当前项目的全部 widgets */
  listWidgets: (wsId: number | string, projectId: number) =>
    wrap<DashboardWidget[]>(
      apiClient.get(`/workspaces/${wsId}/projects/${projectId}/dashboard/widgets`),
    ),

  /** 创建 widget */
  createWidget: (wsId: number | string, projectId: number, input: CreateWidgetInput) =>
    wrap<DashboardWidget>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/dashboard/widgets`, input),
    ),

  /** 删除 widget */
  deleteWidget: (wsId: number | string, projectId: number, widgetId: number) =>
    wrap<void>(
      apiClient.delete(`/workspaces/${wsId}/projects/${projectId}/dashboard/widgets/${widgetId}`),
    ),

  /** 更新 widget 网格位置 / 尺寸 / 配置 */
  updateWidget: (wsId: number | string, projectId: number, widgetId: number, input: UpdateWidgetInput) =>
    wrap<DashboardWidget>(
      apiClient.patch(`/workspaces/${wsId}/projects/${projectId}/dashboard/widgets/${widgetId}`, input),
    ),

  /** 列出项目告警 */
  listAlerts: (wsId: number | string, projectId: number) =>
    wrap<RiskAlert[]>(
      apiClient.get(`/workspaces/${wsId}/projects/${projectId}/dashboard/alerts`),
    ),

  /** 解决告警 */
  resolveAlert: (wsId: number | string, projectId: number, alertId: number) =>
    wrap<void>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/dashboard/alerts/${alertId}/resolve`),
    ),

  /** 列出仪表盘布局模板 */
  listTemplates: (wsId: number | string, projectId: number) =>
    wrap<DashboardTemplate[]>(
      apiClient.get(`/workspaces/${wsId}/projects/${projectId}/dashboard/templates`),
    ),

  // --- 工作空间级仪表盘（跨项目汇总） ---

  /** 列出工作空间级风险告警（跨项目） */
  listWorkspaceAlerts: (wsId: number | string) =>
    wrap<RiskAlert[]>(
      apiClient.get(`/workspaces/${wsId}/dashboard/alerts`),
    ),

  /** 解决工作空间级告警 */
  resolveWorkspaceAlert: (wsId: number | string, alertId: number) =>
    wrap<void>(
      apiClient.post(`/workspaces/${wsId}/dashboard/alerts/${alertId}/resolve`),
    ),

  /** 列出工作空间级仪表盘模板 */
  listWorkspaceTemplates: (wsId: number | string) =>
    wrap<DashboardTemplate[]>(
      apiClient.get(`/workspaces/${wsId}/dashboard/templates`),
    ),
};
