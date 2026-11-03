/**
 * Initiative 域 API — 对接后端 Initiative 域 REST 接口。
 *
 * 倡议（Initiative）是工作空间级的战略聚合：一个倡议可关联多个项目
 * 与多个需求/任务/缺陷，并通过 progress 汇总整体进度。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type InitiativeStatus = "on_track" | "at_risk" | "off_track" | "completed";
export type InitiativePriority = "urgent" | "high" | "medium" | "low" | "none";

/** 倡议实时进度汇总（跨关联项目/工作项） */
export interface InitiativeProgress {
  total_projects: number;
  total_work_items: number;
  completed_work_items: number;
  completion_rate: number;
  on_track_count: number;
  at_risk_count: number;
  off_track_count: number;
  status: "on_track" | "at_risk" | "off_track";
}

/** 倡议聚合根 */
export interface Initiative {
  id: number;
  workspace_id: number;
  title: string;
  description: string;
  status: InitiativeStatus;
  priority: InitiativePriority;
  start_date: string | null;
  target_date: string | null;
  owner_id: number;
  created_by: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
  progress?: InitiativeProgress;
}

/** 创建倡议入参 */
export interface InitiativeCreate {
  title: string;
  description?: string;
  status?: string;
  priority?: string;
  start_date?: string;
  target_date?: string;
  owner_id?: number;
  project_ids?: number[];
}

/** 倡议列表查询参数 */
export interface ListInitiativesParams {
  status?: string;
  priority?: string;
  page?: number;
  pageSize?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 倡议域 API：CRUD / 归档恢复 / 关联项目 / 关联工作项 / 进度 */
export const initiativeApi = {
  list: (workspaceId: number | string, params?: ListInitiativesParams) =>
    wrap<{ results: Initiative[]; total: number }>(
      http.get(`/workspaces/${workspaceId}/initiatives`, { params }),
    ),
  get: (workspaceId: number | string, id: number | string) =>
    wrap<Initiative>(http.get(`/workspaces/${workspaceId}/initiatives/${id}`)),
  create: (workspaceId: number | string, data: InitiativeCreate) =>
    wrap<Initiative>(http.post(`/workspaces/${workspaceId}/initiatives`, data)),
  update: (workspaceId: number | string, id: number | string, data: Partial<InitiativeCreate>) =>
    wrap<Initiative>(http.patch(`/workspaces/${workspaceId}/initiatives/${id}`, data)),
  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/initiatives/${id}`)),
  archive: (workspaceId: number | string, id: number | string) =>
    wrap<Initiative>(http.post(`/workspaces/${workspaceId}/initiatives/${id}/archive`)),
  restore: (workspaceId: number | string, id: number | string) =>
    wrap<Initiative>(http.post(`/workspaces/${workspaceId}/initiatives/${id}/restore`)),

  // --- 关联项目 ---
  addProject: (workspaceId: number | string, id: number | string, projectId: number | string) =>
    wrap<void>(
      http.post(`/workspaces/${workspaceId}/initiatives/${id}/projects`, {
        project_id: projectId,
      }),
    ),
  removeProject: (workspaceId: number | string, id: number | string, projectId: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/initiatives/${id}/projects/${projectId}`),
    ),

  // --- 关联工作项（需求/任务/缺陷） ---
  addWorkItem: (
    workspaceId: number | string,
    id: number | string,
    data: { work_item_id: number | string; work_item_type: string },
  ) =>
    wrap<void>(http.post(`/workspaces/${workspaceId}/initiatives/${id}/work-items`, data)),
  removeWorkItem: (workspaceId: number | string, id: number | string, itemId: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/initiatives/${id}/work-items/${itemId}`),
    ),

  // --- 进度汇总 ---
  progress: (workspaceId: number | string, id: number | string) =>
    wrap<InitiativeProgress>(
      http.get(`/workspaces/${workspaceId}/initiatives/${id}/progress`),
    ),
};
