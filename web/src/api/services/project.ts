/**
 * 项目域 API — 对接后端 Project 域 REST 接口。
 */
import { http } from "../client";
import type { Project } from "./workspace";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 项目列表查询参数 */
export interface ListProjectsParams {
  status?: string;
  limit?: number;
  offset?: number;
  search?: string;
}

/** 创建项目入参 */
export interface CreateProjectInput {
  name: string;
  slug?: string;
  identifier?: string;
  description?: string;
  network?: string;
  icon?: string;
  color?: string;
  cover_image_url?: string;
  template?: string;
  modules?: {
    sprint: boolean;
    version: boolean;
    estimate: boolean;
  };
}

/** 更新项目入参 */
export interface UpdateProjectInput {
  name?: string;
  slug?: string;
  description?: string;
  network?: string;
  icon?: string;
  color?: string;
  cover_image_url?: string;
  modules?: {
    sprint: boolean;
    version: boolean;
    estimate: boolean;
  };
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 项目域 API：CRUD / 成员 / 列表 */
export const projectApi = {
  /** 分页查询工作空间下的项目列表 */
  listProjects: (wsId: number, params?: ListProjectsParams) =>
    wrap<{ results: Project[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects`, { params }),
    ),

  /** 获取单个项目详情 */
  getProject: (wsId: number, projectId: number) =>
    wrap<Project>(http.get(`/workspaces/${wsId}/projects/${projectId}`)),

  /** 创建项目 */
  createProject: (wsId: number, input: CreateProjectInput) =>
    wrap<Project>(http.post(`/workspaces/${wsId}/projects`, input)),

  /** 更新项目 */
  updateProject: (wsId: number, projectId: number, input: UpdateProjectInput) =>
    wrap<Project>(http.patch(`/workspaces/${wsId}/projects/${projectId}`, input)),

  /** 归档项目（软删除） */
  archiveProject: (wsId: number, projectId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}`)),

  // --- 项目成员 ---
  listProjectMembers: (wsId: number, projectId: number) =>
    wrap<Project[]>(http.get(`/workspaces/${wsId}/projects/${projectId}/members`)),

  addProjectMember: (wsId: number, projectId: number, input: { user_id: number; role: string }) =>
    wrap<void>(http.post(`/workspaces/${wsId}/projects/${projectId}/members`, input)),

  changeProjectMemberRole: (wsId: number, projectId: number, userId: number, role: string) =>
    wrap<void>(http.patch(`/workspaces/${wsId}/projects/${projectId}/members/${userId}`, { role })),

  removeProjectMember: (wsId: number, projectId: number, userId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/members/${userId}`)),
};
