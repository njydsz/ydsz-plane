/**
 * 工作空间域 API — axios 调用 + 类型化返回。
 */
import { http } from "../client";

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

/** 角色定义 */
export interface RoleDefinition {
  slug: string;
  name: string;
  description: string;
  level: number;
  is_system: boolean;
  icon: string;
}

/** 当前用户在某工作空间的角色 + 权限列表 */
export interface MyRoleResponse {
  role: RoleDefinition;
  permissions: string[];
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
  cover_image_url?: string;
  status: string;
  sort_order: number;
  created_by: number;
  created_at: string;
  updated_at: string;
  template?: string;
  modules?: ProjectModuleToggles;
}

export interface ProjectModuleToggles {
  intake: boolean;
  sprint: boolean;
  version: boolean;
  estimate: boolean;
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
  /** 上传工作空间 Logo（multipart/form-data，字段名 'file'），返回 { logo_url } */
  uploadLogo: (wsId: number, formData: FormData) =>
    wrap<{ logo_url: string }>(http.post(`/workspaces/${wsId}/logo`, formData, {
      headers: { "Content-Type": "multipart/form-data" },
    })),
  /** 移除工作空间 Logo */
  removeLogo: (wsId: number) => wrap<void>(http.delete(`/workspaces/${wsId}/logo`)),

  // --- Members ---
  listMembers: (wsId: number) => wrap<Member[]>(http.get(`/workspaces/${wsId}/members`)),
  /** 变更成员角色（PATCH 兼容旧版） */
  changeRole: (wsId: number, userId: number, role: string) =>
    wrap<void>(http.patch(`/workspaces/${wsId}/members/${userId}`, { role })),
  /** 变更成员角色（PUT 新版，PUT /workspaces/:wsId/members/:userId/role） */
  updateMemberRole: (wsId: number, userId: number, role: string) =>
    wrap<void>(http.put(`/workspaces/${wsId}/members/${userId}/role`, { role })),
  removeMember: (wsId: number, userId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/members/${userId}`)),

  // --- RBAC ---
  /** 获取当前用户在该工作空间的角色 + 权限列表 */
  getMyRole: (wsId: number) => wrap<MyRoleResponse>(http.get(`/workspaces/${wsId}/role`)),
  /** 获取所有角色定义列表 */
  listRoles: (wsId: number) =>
    wrap<{ items: RoleDefinition[] }>(http.get(`/workspaces/${wsId}/roles`)).then((r) => r.items),

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
  createProject: (wsId: number, input: { name: string; slug?: string; identifier?: string; description?: string; network?: string; icon?: string; color?: string; cover_image_url?: string; template?: string; modules?: ProjectModuleToggles }) =>
    wrap<Project>(http.post(`/workspaces/${wsId}/projects`, input)),
  updateProject: (wsId: number, projectId: number, input: { name?: string; slug?: string; description?: string; network?: string; icon?: string; color?: string; cover_image_url?: string; modules?: ProjectModuleToggles }) =>
    wrap<Project>(http.patch(`/workspaces/${wsId}/projects/${projectId}`, input)),
  archiveProject: (wsId: number, projectId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}`)),

  // --- Project Members ---
  /** 项目成员 */
  listProjectMembers: (wsId: number, projectId: number) =>
    wrap<Member[]>(http.get(`/workspaces/${wsId}/projects/${projectId}/members`)),
  addProjectMember: (wsId: number, projectId: number, input: { user_id: number; role: string }) =>
    wrap<void>(http.post(`/workspaces/${wsId}/projects/${projectId}/members`, input)),
  changeProjectMemberRole: (wsId: number, projectId: number, userId: number, role: string) =>
    wrap<void>(http.patch(`/workspaces/${wsId}/projects/${projectId}/members/${userId}`, { role })),
  removeProjectMember: (wsId: number, projectId: number, userId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/members/${userId}`)),
};
