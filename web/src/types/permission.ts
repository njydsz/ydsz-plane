/**
 * 前端权限类型定义
 *
 * 与后端 auth.PermXxx 常量、db roles/role_permissions 表对齐。
 * 通过 GET /api/v1/workspaces/:id/role 接口加载当前用户权限集。
 *
 * 角色层级：admin(L5/100) > owner(L4/80) > pm/po/techlead/qa(L3/50) > dev(L2/30) > guest(L1/10)
 */

/** 8 种工作空间角色（与后端 auth.WorkspaceRole 对齐） */
export type WorkspaceRoleSlug =
  | 'admin'
  | 'owner'
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
/*  需求/任务/缺陷权限常量                                              */
/* ------------------------------------------------------------------ */

/** 需求/任务/缺陷操作权限，与后端 auth.PermIssueXxx 常量对齐 */
export const IssuePermission = {
  Read: 'issue:read',
  Create: 'issue:create',
  EditOwn: 'issue:edit_own',
  EditAll: 'issue:edit_all',
  Delete: 'issue:delete',
  Transition: 'issue:transition',
  Reassign: 'issue:reassign',
  ChangePriority: 'issue:change_priority',
  ManageSprint: 'issue:manage_sprint',
} as const

/** 角色级别常量 */
export const ROLE_LEVEL = {
  admin: 100, // 系统管理员（平台级）
  owner: 80, // 空间管理员（空间级最高）
  pm: 50, // 项目经理
  po: 50, // 产品经理
  techlead: 50, // 技术经理
  qalead: 50, // 测试经理
  dev: 30, // 开发
  guest: 10, // 访客
} as const

/* ------------------------------------------------------------------ */
/*  菜单定义（左侧导航）                                               */
/* ------------------------------------------------------------------ */

export type PermissionCode = string

export type ProjectModuleKey = 'sprint' | 'version' | 'estimate'

export interface MenuItem {
  name: string
  path: string
  titleKey: string
  icon: string
  permissions?: PermissionCode[]
  /** 最低角色级别（≥ 该值可见），admin(100) 和 owner(80) 都能看到所有管理菜单 */
  minLevel?: number
  children?: MenuItem[]
  /** 关联的功能模块 key：启用/禁用受项目模块开关控制 */
  moduleKey?: ProjectModuleKey
}

/**
 * 工作空间左侧导航顺序与所需权限 / 角色级别参考表（按侧栏从上到下）：
 *
 * | 顺序 | 菜单                    | 所需角色         | 所需权限                         |
 * |-----|-------------------------|------------------|----------------------------------|
 * | 1   | 工作台 (Dashboard)      | guest+           | workspace:read                   |
 * | 2   | 项目 (Projects)         | guest+           | project:read                     |
 * | 3   | 版本 (Versions)         | pm+/techlead+/po | version:read                     |
 * | 4   | 迭代 (Sprints)          | dev+ (30)        | sprint:read                      |
 * | 5   | 报表 (Analytics)        | pm+/techlead     | analytics:read                   |
 * | 6   | 知识库 (Knowledge)      | guest+           | workspace:read                   |
 * | 7   | 自动化 (Automation)     | admin+/owner     | automation:manage                |
 * | 8   | 空间成员 (Members)      | admin+/owner     | member:change_role               |
 * | 9   | 空间设置 (Settings)     | admin+/owner     | workspace:update                 |
 */
export const WORKSPACE_MENU: MenuItem[] = [
  { name: 'workspace-dashboard', path: 'dashboard', titleKey: 'menu.dashboard', icon: 'LayoutDashboard' },
  { name: 'workspace-projects', path: 'projects', titleKey: 'menu.projects', icon: 'FolderKanban', permissions: ['project:read'] },
  { name: 'workspace-versions', path: 'versions', titleKey: 'menu.versions', icon: 'Tag', permissions: ['version:read'], minLevel: 50, moduleKey: 'version' },
  { name: 'workspace-sprints', path: 'sprints', titleKey: 'menu.sprints', icon: 'Clock', permissions: ['sprint:read'], minLevel: 30, moduleKey: 'sprint' },
  { name: 'workspace-initiatives', path: 'initiatives', titleKey: 'menu.initiatives', icon: 'Target', minLevel: 30 },
  { name: 'workspace-teamspaces', path: 'teamspaces', titleKey: 'menu.teamspaces', icon: 'Users', minLevel: 30 },
  { name: 'workspace-projects-as-code', path: 'projects-as-code', titleKey: 'menu.projectsAsCode', icon: 'FileCode2', minLevel: 30 },
  { name: 'workspace-pql', path: 'pql', titleKey: 'menu.pql', icon: 'SearchCode', minLevel: 30 },
  { name: 'workspace-views', path: 'views', titleKey: 'menu.views', icon: 'LayoutGrid', permissions: ['project:update'], minLevel: 30 },
  { name: 'workspace-customers', path: 'customers', titleKey: 'menu.customers', icon: 'Building2', minLevel: 30 },
  { name: 'workspace-customer-requests', path: 'customer-requests', titleKey: 'menu.customerRequests', icon: 'Inbox', minLevel: 30 },
  { name: 'workspace-analytics', path: 'analytics', titleKey: 'menu.analytics', icon: 'BarChart3', permissions: ['analytics:read'], minLevel: 50 },
  { name: 'workspace-knowledge', path: 'knowledge', titleKey: 'menu.knowledge', icon: 'BookOpen' },
  { name: 'workspace-automation', path: 'automation', titleKey: 'menu.automation', icon: 'Zap', permissions: ['automation:manage'], minLevel: 80 },
  { name: 'workspace-integrations', path: 'integrations', titleKey: 'menu.integrations', icon: 'Blocks', minLevel: 30 },
  { name: 'workspace-ai-agents', path: 'ai-agents', titleKey: 'menu.aiAgents', icon: 'Sparkles', minLevel: 30 },
  { name: 'workspace-ai-memory', path: 'ai-memory', titleKey: 'menu.aiMemory', icon: 'Brain', minLevel: 30 },
  { name: 'workspace-oauth-apps', path: 'oauth-apps', titleKey: 'menu.oauthApps', icon: 'KeyRound', minLevel: 40 },
  { name: 'workspace-grant-rules', path: 'grant-rules', titleKey: 'menu.grantRules', icon: 'ShieldCheck', minLevel: 40 },
  { name: 'workspace-members', path: 'members', titleKey: 'menu.members', icon: 'Users', permissions: ['member:change_role'], minLevel: 80 },
  { name: 'workspace-settings', path: 'settings', titleKey: 'menu.settings', icon: 'Settings', permissions: ['workspace:update'], minLevel: 80 },
]

/**
 * 项目级左侧导航顺序与所需权限 / 角色级别参考表（项目上下文内，按侧栏从上到下）：
 *
 * | 顺序 | 菜单                          | 所需角色         | 所需权限          |
 * |-----|-------------------------------|------------------|-------------------|
 * | 1   | 看板 (Board)                  | guest+           | issue:read        |
 * | 2   | 需求 (Requirements)           | guest+           | issue:read        |
 * | 3   | 任务 (Tasks)                  | guest+           | issue:read        |
 * | 4   | 缺陷 (Defects)                | guest+           | issue:read        |
 * | 5   | 文档 (Pages)                  | guest+           | page:read         |
 * | 6   | 日历 (Calendar)               | guest+           | issue:read        |
 * | 7   | 效能 (Metrics)                | pm+/techlead     | analytics:read    |
 * | 8   | 分析 (Analytics)              | pm+/techlead     | analytics:read    |
 * | 9   | 甘特图 (Gantt)                | guest+           | issue:read        |
 * | 10  | 时间线 (Timeline)             | guest+           | issue:read        |
 * | 11  | 仪表盘 (Dashboard)            | guest+           | project:read      |
 * | 12  | 工作项类型 (Work Item Types)  | admin            | project:update    |
 */
export const PROJECT_MENU: MenuItem[] = [
  { name: 'project-board', path: 'board', titleKey: 'menu.board', icon: 'KanbanSquare', permissions: ['issue:read'] },
  { name: 'project-requirements', path: 'requirements', titleKey: 'menu.requirements', icon: 'ClipboardList', permissions: ['issue:read'] },
  { name: 'project-tasks', path: 'tasks', titleKey: 'menu.tasks', icon: 'SquareCheckBig', permissions: ['issue:read'] },
  { name: 'project-defects', path: 'defects', titleKey: 'menu.defects', icon: 'Bug', permissions: ['issue:read'] },
  { name: 'project-pages', path: 'pages', titleKey: 'menu.pages', icon: 'FileText', permissions: ['page:read'] },
  { name: 'project-calendar', path: 'calendar', titleKey: 'menu.calendar', icon: 'Calendar', permissions: ['issue:read'] },
  { name: 'project-metrics', path: 'metrics', titleKey: 'menu.metrics', icon: 'Gauge', permissions: ['analytics:read'], minLevel: 50 },
  { name: 'project-analytics', path: 'analytics', titleKey: 'menu.analytics', icon: 'LineChart', permissions: ['analytics:read'], minLevel: 50 },
  { name: 'project-gantt', path: 'gantt', titleKey: 'menu.gantt', icon: 'GanttChart', permissions: ['issue:read'] },
  { name: 'project-timeline', path: 'timeline', titleKey: 'menu.timeline', icon: 'History', permissions: ['issue:read'] },
  { name: 'project-dashboard', path: 'dashboard', titleKey: 'menu.dashboard', icon: 'LayoutDashboard', permissions: ['project:read'] },
  { name: 'project-work-item-types', path: 'work-item-types', titleKey: 'menu.workItemTypes', icon: 'Shapes', permissions: ['project:update'], minLevel: 40 },
  { name: 'project-recurring-rules', path: 'recurring-rules', titleKey: 'menu.recurringRules', icon: 'Repeat', permissions: ['project:update'], minLevel: 40 },
  { name: 'project-milestones', path: 'milestones', titleKey: 'menu.milestones', icon: 'Milestone', permissions: ['project:update'], minLevel: 40 },
  { name: 'project-imports', path: 'imports', titleKey: 'menu.imports', icon: 'Upload', permissions: ['project:update'], minLevel: 40 },
]
