/**
 * Sprint 域 API — 对接后端 Sprint 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type SprintStatus = "planned" | "active" | "completed";
/** 迭代结束时未完成需求/任务/缺陷的处理策略 */
export type UnfinishedStrategy = "backlog" | "next_sprint" | "keep";

/** 迭代聚合（含可选进度与复盘快照） */
export interface Sprint {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  description?: string;
  goal?: string;
  status: SprintStatus;
  start_date?: string;
  end_date?: string;
  capacity?: number;
  owner_id?: number;
  viewport?: Record<string, unknown>;
  progress?: SprintProgress;
  review_snapshot?: ReviewSnapshot;
  started_at?: string;
  completed_at?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 迭代实时进度汇总（点数与需求/任务/缺陷数） */
export interface SprintProgress {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
  by_state_group?: Record<string, number>;
  saturation?: number;
  added_points: number;
  removed_points: number;
}

/** 迭代复盘快照（承诺/完成/加入/移除统计） */
export interface ReviewSnapshot {
  committed_points: number;
  completed_points: number;
  joined_points: number;
  removed_points: number;
  committed_issues: number;
  completed_issues: number;
  joined_issues: number;
  removed_issues: number;
  completion_rate: number;
}

/** 迭代每日快照记录（燃尽图数据源） */
export interface SprintSnapshot {
  id: number;
  workspace_id: number;
  project_id: number;
  sprint_id: number;
  snapshot_date: string;
  data: SnapshotData;
  created_at: string;
}

/** 单日快照内容 */
export interface SnapshotData {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
  by_state_group?: Record<string, number>;
  added_points: number;
  removed_points: number;
}

/** 燃尽图单日数据点（含理想线） */
export interface BurndownPoint {
  date: string;
  total_points: number;
  done_points: number;
  remaining: number;
  ideal_line: number;
}

/** 迭代速率统计（平均值/中位数/近期迭代） */
export interface VelocityStats {
  avg_points: number;
  avg_issues: number;
  p50: number;
  recent_sprints: SprintVelocity[];
  count: number;
}

/** 单个迭代的速率数据 */
export interface SprintVelocity {
  sprint_id: number;
  sprint_name: string;
  completed_points: number;
  completed_issues: number;
  end_date: string;
}

/** 迭代内需求/任务/缺陷的视图投影 */
export interface SprintIssueView {
  issue_id: number;
  sort_order: number;
  name: string;
  type_code: string;
  priority: string;
  state_id: number;
  state_name: string;
  state_color: string;
  state_group: string;
  added_midway: boolean;
  point?: number;
  severity?: number;
  created_at: string;
}

/** Backlog 需求/任务/缺陷条目（未规划进 active 迭代） */
export interface BacklogItem {
  issue_id: number;
  name: string;
  type_code: string;
  priority: string;
  state_id: number;
  state_name: string;
  state_color: string;
  state_group: string;
  has_sprint: boolean;
  sprint_id?: number;
  sprint_name?: string;
  point?: number;
}

/** 创建迭代入参 */
export interface CreateSprintInput {
  name: string;
  description?: string;
  goal?: string;
  start_date?: string;
  end_date?: string;
  capacity?: number;
  owner_id?: number;
}

/** 更新迭代入参（仅 planned 状态可编辑） */
export interface UpdateSprintInput {
  name?: string;
  description?: string;
  goal?: string;
  start_date?: string;
  end_date?: string;
  capacity?: number;
  owner_id?: number;
  version: number;
}

/** 结束迭代入参（未完成需求/任务/缺陷处理策略） */
export interface CompleteSprintInput {
  strategy: UnfinishedStrategy;
  next_sprint_id?: number;
}

/** 迭代列表查询参数 */
export interface ListSprintsParams {
  status?: SprintStatus;
  limit?: number;
  offset?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 迭代域 API：CRUD / 生命周期 / 进度 / 规划 / 燃尽图 / 复盘 / 速率建议 */
export const sprintApi = {
  // --- 迭代 CRUD ---
  listSprints: (wsId: number, projectId: number, params?: ListSprintsParams) =>
    wrap<{ results: Sprint[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints`, { params }),
    ),
  getSprint: (wsId: number, projectId: number, sprintId: number) =>
    wrap<Sprint>(http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}`)),
  createSprint: (wsId: number, projectId: number, input: CreateSprintInput) =>
    wrap<Sprint>(http.post(`/workspaces/${wsId}/projects/${projectId}/sprints`, input)),
  updateSprint: (wsId: number, projectId: number, sprintId: number, input: UpdateSprintInput) =>
    wrap<Sprint>(http.patch(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}`, input)),
  deleteSprint: (wsId: number, projectId: number, sprintId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}`)),

  // --- 迭代生命周期 ---
  startSprint: (wsId: number, projectId: number, sprintId: number) =>
    wrap<Sprint>(http.post(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/start`)),
  completeSprint: (wsId: number, projectId: number, sprintId: number, input: CompleteSprintInput) =>
    wrap<Sprint>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/complete`, input),
    ),

  // --- 进度 ---
  getSprintProgress: (wsId: number, projectId: number, sprintId: number) =>
    wrap<SprintProgress>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/progress`),
    ),

  // --- 规划：Backlog 与 Sprint issues ---
  listSprintIssues: (wsId: number, projectId: number, sprintId: number, limit = 50, offset = 0) =>
    wrap<{ results: SprintIssueView[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/issues`, {
        params: { limit, offset },
      }),
    ),
  addIssue: (wsId: number, projectId: number, sprintId: number, issueId: number, sortOrder = 65535) =>
    wrap<void>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/issues`, {
        issue_id: issueId,
        sort_order: sortOrder,
      }),
    ),
  removeIssue: (wsId: number, projectId: number, sprintId: number, issueId: number) =>
    wrap<void>(
      http.delete(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/issues/${issueId}`),
    ),

  // --- Backlog 视图 ---
  getBacklog: (wsId: number, projectId: number, limit = 50, offset = 0) =>
    wrap<{ results: BacklogItem[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/backlog`, {
        params: { limit, offset },
      }),
    ),

  // --- 分析 ---
  burndown: (wsId: number, projectId: number, sprintId: number) =>
    wrap<{ sprint: Sprint; points: BurndownPoint[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/burndown`),
    ),
  getReview: (wsId: number, projectId: number, sprintId: number) =>
    wrap<ReviewSnapshot>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/${sprintId}/review`),
    ),
  suggestCapacity: (wsId: number, projectId: number) =>
    wrap<VelocityStats>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/sprints/suggest-capacity`),
    ),
};
