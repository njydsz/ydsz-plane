/**
 * RBAC 权限矩阵 API —— 前端硬编码权限常量 + 角色成员管理调用。
 *
 * 后端已有 workspaceApi.listMembers / workspaceApi.changeRole，
 * 本模块补充：
 *  - 权限矩阵常量（resourceGroups / roles）
 *  - 按角色统计成员数的辅助函数
 */
import { apiClient } from "../client";

/* ---------------------------------------------------------------------------
 * 权限矩阵常量（与 rbac.go 对齐，前端硬编码）
 * ------------------------------------------------------------------------- */

export type RoleKey = "owner" | "admin" | "member" | "guest";

export interface PermissionDef {
  key: string;
  label: string;
  guest: boolean;
  member: boolean;
  admin: boolean;
  owner: boolean;
}

export interface ResourceGroup {
  name: string;
  icon: string;
  permissions: PermissionDef[];
}

export interface RoleDef {
  key: RoleKey;
  label: string;
  icon: string;
  level: number;
  description: string;
}

export interface RBACData {
  resourceGroups: ResourceGroup[];
  roles: RoleDef[];
}

/** 完整权限矩阵数据 */
export const RBAC_MATRIX: RBACData = {
  resourceGroups: [
    {
      name: "工作项管理",
      icon: "📋",
      permissions: [
        { key: "issue:read", label: "查看工作项", guest: true, member: true, admin: true, owner: true },
        { key: "issue:create", label: "创建工作项", guest: false, member: true, admin: true, owner: true },
        { key: "issue:update", label: "编辑工作项", guest: false, member: true, admin: true, owner: true },
        { key: "issue:delete", label: "删除工作项", guest: false, member: false, admin: true, owner: true },
        { key: "issue:assign", label: "指派工作项", guest: false, member: true, admin: true, owner: true },
      ],
    },
    {
      name: "项目管理",
      icon: "🎯",
      permissions: [
        { key: "project:read", label: "查看项目", guest: true, member: true, admin: true, owner: true },
        { key: "project:create", label: "创建项目", guest: false, member: false, admin: true, owner: true },
        { key: "project:update", label: "编辑项目", guest: false, member: false, admin: true, owner: true },
        { key: "project:delete", label: "删除项目", guest: false, member: false, admin: false, owner: true },
      ],
    },
    {
      name: "迭代与版本",
      icon: "🔄",
      permissions: [
        { key: "sprint:manage", label: "管理迭代", guest: false, member: true, admin: true, owner: true },
        { key: "version:release", label: "发布版本", guest: false, member: false, admin: true, owner: true },
      ],
    },
    {
      name: "成员管理",
      icon: "👥",
      permissions: [
        { key: "member:invite", label: "邀请成员", guest: false, member: false, admin: true, owner: true },
        { key: "member:remove", label: "移除成员", guest: false, member: false, admin: true, owner: true },
        { key: "member:change_role", label: "修改角色", guest: false, member: false, admin: false, owner: true },
      ],
    },
    {
      name: "自动化与集成",
      icon: "⚡",
      permissions: [
        { key: "automation:manage", label: "管理自动化规则", guest: false, member: true, admin: true, owner: true },
        { key: "webhook:manage", label: "管理 Webhook", guest: false, member: false, admin: true, owner: true },
        { key: "api_token:manage", label: "管理 API Token", guest: false, member: false, admin: true, owner: true },
      ],
    },
    {
      name: "审计与设置",
      icon: "🔒",
      permissions: [
        { key: "audit:read", label: "查看审计日志", guest: false, member: false, admin: true, owner: true },
        { key: "workspace:settings", label: "修改工作空间设置", guest: false, member: false, admin: true, owner: true },
      ],
    },
  ],
  roles: [
    { key: "owner", label: "所有者", icon: "👑", level: 80, description: "拥有工作空间全部权限" },
    { key: "admin", label: "管理员", icon: "🛡️", level: 60, description: "管理项目、成员与设置，但不能删除工作空间" },
    { key: "member", label: "成员", icon: "👤", level: 30, description: "参与日常工作项和迭代" },
    { key: "guest", label: "访客", icon: "👁️", level: 10, description: "仅查看权限" },
  ],
};

/** RBAC 域 API 调用 */
export const rbacApi = {
  /**
   * 变更成员角色 — PUT /workspaces/:wsId/members/:userId/role
   */
  async updateMemberRole(wsId: number, userId: number, role: string): Promise<void> {
    await apiClient.put(`/workspaces/${wsId}/members/${userId}/role`, { role });
  },
};

/* ---------------------------------------------------------------------------
 * 辅助函数
 * ------------------------------------------------------------------------- */

/** 获取某角色在所有成员中的计数 */
export function countMembersByRole<T extends { role: string }>(
  members: T[],
  roleKey: RoleKey,
): number {
  return members.filter((m) => m.role === roleKey).length;
}

/** 检查某角色在某权限上是否允许 */
export function isPermissionAllowed(
  perm: PermissionDef,
  roleKey: RoleKey,
): boolean {
  return perm[roleKey];
}

/** 获取角色定义的便捷函数 */
export function getRoleDef(key: RoleKey): RoleDef | undefined {
  return RBAC_MATRIX.roles.find((r) => r.key === key);
}
