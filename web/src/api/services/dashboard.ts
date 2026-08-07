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
  | "team_workload";

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

export interface ProgressOverviewData {
  total_issues: number;
  done_issues: number;
  in_progress: number;
  overdue_issues: number;
  blocked_issues: number;
  completion_rate: number;
  active_sprints: number;
}

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

export interface PrioritySplitData {
  total: number;
  by_priority: Record<string, number>;
}

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

export interface OverdueItem {
  id: number;
  identifier: string;
  title: string;
  priority: string;
  overdue_days: number;
  assignee: string;
}

export interface OverdueListData {
  total: number;
  items: OverdueItem[];
}

export interface BlockedItem {
  id: number;
  identifier: string;
  title: string;
  blocked_count: number;
  blocker_names: string;
}

export interface BlockedListData {
  total: number;
  items: BlockedItem[];
}

export interface TeamMemberWorkload {
  user_id: number;
  user_name: string;
  avatar?: string;
  todo: number;
  in_progress: number;
  done: number;
  total: number;
}

export interface TeamWorkloadWidgetData {
  members: TeamMemberWorkload[];
}

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

export interface RecentActivityData {
  items: ActivityItem[];
}

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

export interface DashboardData {
  widgets: DashboardWidget[];
  snapshots: Record<string, any>;
  alerts: RiskAlert[];
}

export interface CreateWidgetInput {
  widget_type: WidgetType;
  title: string;
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
};
