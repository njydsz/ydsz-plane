/**
 * 前端权限类型定义
 *
 * 与后端 auth.PermXxx 常量、db roles/role_permissions 表对齐。
 * 通过 GET /api/v1/workspaces/:id/role 接口加载当前用户权限集。
 */

/** 8 种工作空间角色（与后端 auth.WorkspaceRole 对齐） */
export type WorkspaceRoleSlug =
  | 'owner'
  | 'admin'
  | 'pm'
  | 'po'
  | 'techlead'
  | 'qalead'
  | 'dev'
  | 'guest'

/** 角色定义（来自 GET /roles） */
export interface RoleDefinition {
  slug: WorkspaceRoleSlug
  name: string
  description: string
  level: number
  is_system: boolean
  icon: string
}

/** 当前用户在某工作空间的角色 + 权限列表 */
export interface MyRoleResponse {
  role: RoleDefinition
  permissions: string[]
}

/* ------------------------------------------------------------------ */
/*  菜单定义（左侧导航）                                               */
/* ------------------------------------------------------------------ */

export type PermissionCode = string

export interface MenuItem {
  name: string
  path: string
  titleKey: string
  icon: string
  permissions?: PermissionCode[]
  minLevel?: number
  children?: MenuItem[]
}

/**
 * 工作空间左侧导航所需权限 / 角色级别参考表：
 *
 * | 菜单                    | 所需角色         | 所需权限                         |
 * |-------------------------|------------------|----------------------------------|
 * | 工作台 (Dashboard)      | guest+           | workspace:read                   |
 * | 我的工作项 (My Issues)  | guest+           | workspace:read                   |
 * | 项目 (Projects)         | guest+           | project:read                     |
 * | 迭代 (Sprints)          | dev+             | sprint:read                      |
 * | 版本 (Versions)         | pm+/techlead+/po | version:read                     |
 * | 报表 (Analytics)        | pm+/techlead     | analytics:read                   |
 * | 自动化 (Automation)     | admin+/pm        | automation:manage                |
 * | 审计日志 (Audit Logs)   | admin+/owner     | audit:read                       |
 * | 成员管理 (Members)      | admin+/owner     | member:change_role               |
 * | 工作空间设置 (Settings) | admin+/owner     | workspace:update                 |
 * | 收件箱 (Intake)         | admin+/owner     | intake:manage                    |
 * | Webhooks               | admin+/owner     | webhook:manage                   |
 */
export const WORKSPACE_MENU: MenuItem[] = [
  { name: 'workspace-dashboard', path: 'dashboard', titleKey: 'menu.dashboard', icon: 'LayoutDashboard' },
  { name: 'workspace-my-issues', path: 'my-issues', titleKey: 'menu.myIssues', icon: 'UserCircle' },
  { name: 'workspace-projects', path: 'projects', titleKey: 'menu.projects', icon: 'FolderKanban', permissions: ['project:read'] },
  { name: 'workspace-sprints', path: 'sprints', titleKey: 'menu.sprints', icon: 'Clock', permissions: ['sprint:read'], minLevel: 30 },
  { name: 'workspace-versions', path: 'versions', titleKey: 'menu.versions', icon: 'Tag', permissions: ['version:read'], minLevel: 50 },
  { name: 'workspace-analytics', path: 'analytics', titleKey: 'menu.analytics', icon: 'BarChart3', permissions: ['analytics:read'], minLevel: 50 },
  { name: 'workspace-automation', path: 'automation', titleKey: 'menu.automation', icon: 'Zap', permissions: ['automation:manage'], minLevel: 80 },
  { name: 'workspace-audit', path: 'audit', titleKey: 'menu.audit', icon: 'ScrollText', permissions: ['audit:read'], minLevel: 80 },
  { name: 'workspace-intake', path: 'intake', titleKey: 'menu.intake', icon: 'Inbox', permissions: ['intake:manage'], minLevel: 80 },
  { name: 'workspace-webhooks', path: 'webhooks', titleKey: 'menu.webhooks', icon: 'Link', permissions: ['webhook:manage'], minLevel: 80 },
  { name: 'workspace-members', path: 'members', titleKey: 'menu.members', icon: 'Users', permissions: ['member:change_role'], minLevel: 80 },
  { name: 'workspace-settings', path: 'settings', titleKey: 'menu.settings', icon: 'Settings', permissions: ['workspace:update'], minLevel: 80 },
]
