/**
 * 团队空间 API — 封装 Teamspace（团队 = 人 + 项目集合）的 REST 接口。
 *
 * 路由: /api/v1/workspaces/:ws/teamspaces
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 团队空间 */
export interface Teamspace {
  id: number;
  workspace_id: number;
  name: string;
  description: string;
  color: string;
  lead_id: number;
  created_by: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

/** 团队成员 */
export interface TeamspaceMember {
  teamspace_id: number;
  user_id: number;
  role: "lead" | "member";
  created_at: string;
}

/** 团队进度汇总 */
export interface TeamspaceProgress {
  teamspace_id: number;
  member_count: number;
  total_work_items: number;
  completed_work_items: number;
  overdue_work_items: number;
  in_progress_work_items: number;
  project_count: number;
  active_cycles: number;
  completion_rate: number;
}

/** 成员工作负载 */
export interface TeamspaceMemberWorkload {
  user_id: number;
  total_work_items: number;
  in_progress_items: number;
  overdue_items: number;
  completion_rate: number;
}

/** 创建团队空间入参 */
export interface CreateTeamspaceInput {
  name: string;
  description?: string;
  color?: string;
  lead_id?: number;
  member_ids?: number[];
  project_ids?: number[];
}

/** 更新团队空间入参 */
export interface UpdateTeamspaceInput {
  name?: string;
  description?: string;
  color?: string;
  lead_id?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 团队空间域 API */
export const teamspaceApi = {
  /** 列出工作空间下全部团队空间 */
  list: (workspaceId: number | string) =>
    wrap<Teamspace[]>(http.get(`/workspaces/${workspaceId}/teamspaces`)),

  /** 获取团队空间详情 */
  get: (workspaceId: number | string, id: number | string) =>
    wrap<Teamspace>(http.get(`/workspaces/${workspaceId}/teamspaces/${id}`)),

  /** 创建团队空间 */
  create: (workspaceId: number | string, data: CreateTeamspaceInput) =>
    wrap<Teamspace>(http.post(`/workspaces/${workspaceId}/teamspaces`, data)),

  /** 更新团队空间 */
  update: (
    workspaceId: number | string,
    id: number | string,
    data: UpdateTeamspaceInput,
  ) =>
    wrap<Teamspace>(
      http.patch(`/workspaces/${workspaceId}/teamspaces/${id}`, data),
    ),

  /** 删除团队空间 */
  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/teamspaces/${id}`)),

  /** 归档 */
  archive: (workspaceId: number | string, id: number | string) =>
    wrap<Teamspace>(
      http.post(`/workspaces/${workspaceId}/teamspaces/${id}/archive`),
    ),

  /** 恢复 */
  restore: (workspaceId: number | string, id: number | string) =>
    wrap<Teamspace>(
      http.post(`/workspaces/${workspaceId}/teamspaces/${id}/restore`),
    ),

  /* --- 成员 --- */

  listMembers: (workspaceId: number | string, id: number | string) =>
    wrap<{ items: TeamspaceMember[] }>(
      http.get(`/workspaces/${workspaceId}/teamspaces/${id}/members`),
    ),

  addMember: (
    workspaceId: number | string,
    id: number | string,
    userId: number,
    role: "lead" | "member" = "member",
  ) =>
    wrap<void>(
      http.post(`/workspaces/${workspaceId}/teamspaces/${id}/members`, {
        user_id: userId,
        role,
      }),
    ),

  removeMember: (
    workspaceId: number | string,
    id: number | string,
    userId: number,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/teamspaces/${id}/members/${userId}`,
      ),
    ),

  /* --- 项目 --- */

  listProjects: (workspaceId: number | string, id: number | string) =>
    wrap<{ items: number[] }>(
      http.get(`/workspaces/${workspaceId}/teamspaces/${id}/projects`),
    ),

  addProject: (
    workspaceId: number | string,
    id: number | string,
    projectId: number,
  ) =>
    wrap<void>(
      http.post(`/workspaces/${workspaceId}/teamspaces/${id}/projects`, {
        project_id: projectId,
      }),
    ),

  removeProject: (
    workspaceId: number | string,
    id: number | string,
    projectId: number,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/teamspaces/${id}/projects/${projectId}`,
      ),
    ),

  /* --- 进度 --- */

  progress: (workspaceId: number | string, id: number | string) =>
    wrap<TeamspaceProgress>(
      http.get(`/workspaces/${workspaceId}/teamspaces/${id}/progress`),
    ),

  workloads: (workspaceId: number | string, id: number | string) =>
    wrap<{ items: TeamspaceMemberWorkload[] }>(
      http.get(`/workspaces/${workspaceId}/teamspaces/${id}/workloads`),
    ),
};
