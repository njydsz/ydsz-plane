/**
 * 工作空间域 API — axios 调用 + 类型化返回。
 */
import { http, type ApiResponse } from "../client";

/** 工作空间（含当前用户角色与成员数可选字段） */
export interface Workspace {
  id: number;
  name: string;
  slug: string;
  logo_url?: string;
  timezone: string;
  language: string;
  status: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
  role?: string;
  member_count?: number;
}

/** 工作空间成员 */
export interface Member {
  id: number;
  email: string;
  display_name: string;
  avatar_url?: string;
  role: string;
  joined_at: string;
}

/** 工作空间邀请记录 */
export interface Invitation {
  id: number;
  workspace_id: number;
  inviter_id: number;
  inviter_name?: string;
  email: string;
  role: string;
  status: string;
  message?: string;
  expires_at: string;
  accepted_at?: string;
  created_at: string;
}

/** 邀请预览（接受前展示，无鉴权） */
export interface InvitationPreview {
  workspace_id: number;
  workspace_name: string;
  inviter_name: string;
  email: string;
  role: string;
  expires_at: string;
  status: string;
}

/** 项目（工作空间下二级聚合根） */
export interface Project {
  id: number;
  workspace_id: number;
  name: string;
  slug: string;
  identifier: string;
  description?: string;
  network: string;
  icon?: string;
  color?: string;
  status: string;
  sort_order: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

// --- Workspaces ---
/** 工作空间域 API：空间 / 成员 / 邀请 / 项目 CRUD */
export const workspaceApi = {
  list: () => wrap<Workspace[]>(http.get("/workspaces")),
  get: (wsId: number) => wrap<Workspace>(http.get(`/workspaces/${wsId}`)),
  getBySlug: (slug: string) => wrap<Workspace>(http.get(`/workspaces/slug/${slug}`)),
  create: (input: { name: string; slug?: string; timezone?: string; language?: string }) =>
    wrap<Workspace>(http.post("/workspaces", input)),
  update: (wsId: number, input: { name?: string; timezone?: string; language?: string; logo_url?: string }) =>
    wrap<Workspace>(http.patch(`/workspaces/${wsId}`, input)),
  archive: (wsId: number) => wrap<void>(http.delete(`/workspaces/${wsId}`)),

  // --- Members ---
  listMembers: (wsId: number) => wrap<Member[]>(http.get(`/workspaces/${wsId}/members`)),
  changeRole: (wsId: number, userId: number, role: string) =>
    wrap<void>(http.patch(`/workspaces/${wsId}/members/${userId}`, { role })),
  removeMember: (wsId: number, userId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/members/${userId}`)),

  // --- Invitations ---
  sendInvitation: (wsId: number, input: { email: string; role: string; message?: string }) =>
    wrap<Invitation>(http.post(`/workspaces/${wsId}/invitations`, input)),
  listInvitations: (wsId: number, status?: string) =>
    wrap<Invitation[]>(http.get(`/workspaces/${wsId}/invitations`, { params: status ? { status } : {} })),
  revokeInvitation: (wsId: number, invId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/invitations/${invId}`)),
  previewInvitation: (token: string) =>
    wrap<InvitationPreview>(http.get(`/invitations/${token}`)),
  acceptInvitation: (token: string) =>
    wrap<Invitation>(http.post("/invitations/accept", { token })),

  // --- Projects ---
  listProjects: (wsId: number) => wrap<Project[]>(http.get(`/workspaces/${wsId}/projects`)),
  getProject: (wsId: number, projectId: number) =>
    wrap<Project>(http.get(`/workspaces/${wsId}/projects/${projectId}`)),
  createProject: (wsId: number, input: { name: string; slug?: string; identifier?: string; description?: string; network?: string; icon?: string; color?: string }) =>
    wrap<Project>(http.post(`/workspaces/${wsId}/projects`, input)),
  updateProject: (wsId: number, projectId: number, input: { name?: string; slug?: string; description?: string; network?: string; icon?: string; color?: string }) =>
    wrap<Project>(http.patch(`/workspaces/${wsId}/projects/${projectId}`, input)),
  archiveProject: (wsId: number, projectId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}`)),
};
