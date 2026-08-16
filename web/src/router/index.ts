/**
 * 路由表 — 定义全部页面路由与全局鉴权守卫。
 *
 * 约定：
 *  - meta.public=true 表示公开路由（无需登录，如登录/注册/邀请预览）；
 *  - 其余路由要求已登录，未登录跳转 /login 并携带 redirect 回跳参数；
 *  - 已登录用户访问认证页（login/register 等）会被重定向到首页；
 *  - meta.permission 表示该路由需要的工作空间权限，缺失时重定向到 403 页。
 */
import { createRouter, createWebHistory } from "vue-router";

import { useAuthStore } from "@/stores/auth";
import { useWorkspaceStore } from "@/stores/workspace";

/** 需要工作空间级鉴权的路由（:workspaceId 前缀）及其最低权限。
 *  留空 string 表示"仅 owner / admin 可访问"（跳过 permission 集合判断，直接校验角色）。 */
const WORKSPACE_PERMISSIOND_ROUTES: Record<string, string> = {
  "workspace-settings": "workspace:update",
  "workspace-members": "member:change_role",
  "workspace-versions": "version:read",
  "workspace-sprints": "sprint:read",
  "workspace-analytics": "analytics:read",
  "workspace-automation": "automation:manage",
  "webhook-settings": "webhook:manage",
  "audit-logs": "audit:read",
  "workspace-dlq": "",
  "workspace-rbac": "",
};

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/auth/LoginView.vue"),
      meta: { public: true },
    },
    {
      path: "/register",
      name: "register",
      component: () => import("@/views/auth/RegisterView.vue"),
      meta: { public: true },
    },
    {
      // S13 OIDC SSO 回调页 — 后端 /api/v1/auth/oidc/callback 重定向至此
      path: "/sso/callback",
      name: "sso-callback",
      component: () => import("@/views/auth/SSOCallbackView.vue"),
      meta: { public: true },
    },
    {
      path: "/forgot-password",
      name: "forgot-password",
      component: () => import("@/views/auth/ForgotPasswordView.vue"),
      meta: { public: true },
    },
    {
      path: "/reset-password",
      name: "reset-password",
      component: () => import("@/views/auth/ResetPasswordView.vue"),
      meta: { public: true },
    },
    // 文档公开分享只读视图（免登录）
    {
      path: "/public/page/:token",
      name: "public-page",
      component: () => import("@/views/project/PublicPageView.vue"),
      meta: { public: true },
      props: (route) => ({ token: String(route.params.token) }),
    },
    {
      path: "/settings/api-tokens",
      name: "api-tokens",
      component: () => import("@/views/settings/ApiTokensView.vue"),
    },
    {
      path: "/settings/profile",
      name: "profile",
      component: () => import("@/views/settings/ProfileView.vue"),
    },
    {
      path: "/",
      component: () => import("@/layouts/WorkspaceLayout.vue"),
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/views/workspace/WorkspaceListView.vue"),
        },
        {
          path: ":workspaceId(\\d+)/workbench",
          name: "workbench",
          component: () => import("@/views/workspace/WorkbenchView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        {
          path: ":workspaceId(\\d+)/search",
          name: "search",
          component: () => import("@/views/workspace/SearchView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 通知中心全屏页（从铃铛下拉「查看全部通知」进入）
        {
          path: ":workspaceId(\\d+)/notifications",
          name: "notifications",
          component: () => import("@/views/workspace/NotificationsView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 我的工作项：当前用户被分配的工作项（跨项目聚合）
        {
          path: ":workspaceId(\\d+)/my-issues",
          name: "my-issues",
          component: () => import("@/views/workspace/MyIssuesView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 邀请接受链接：公开可读，但 POST accept 需鉴权
        {
          path: "invite/:token",
          name: "invite-preview",
          component: () => import("@/views/workspace/InvitePreview.vue"),
          meta: { public: true },
        },
        {
          path: ":workspaceId(\\d+)/settings",
          name: "workspace-settings",
          component: () => import("@/views/workspace/WorkspaceSettingsView.vue"),
          props: (route) => ({ workspace_id: Number(route.params.workspaceId) }),
        },
        // S10 Webhook 管理
        {
          path: ":workspaceId(\\d+)/settings/webhooks",
          name: "webhook-settings",
          component: () => import("@/views/workspace/WebhookSettingsView.vue"),
        },
        {
          path: ":workspaceId(\\d+)/settings/notifications",
          name: "notification-preferences",
          component: () => import("@/views/workspace/NotificationPreferencesView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        {
          path: ":workspaceId(\\d+)/audit-logs",
          name: "audit-logs",
          component: () => import("@/views/workspace/AuditReportView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // S13 DLQ 死信监控（仅 owner / admin）
        {
          path: ":workspaceId(\\d+)/admin/dlq",
          name: "workspace-dlq",
          component: () => import("@/views/workspace/DLQMonitoringView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // S13 角色权限管理（仅 owner / admin）
        {
          path: ":workspaceId(\\d+)/admin/rbac",
          name: "workspace-rbac",
          component: () => import("@/views/workspace/RolesPermissionsView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        {
          path: ":workspaceId(\\d+)/dashboard",
          name: "workspace-dashboard",
          component: () => import("@/views/workspace/WorkspaceDashboardView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // S13 工作空间级成员管理（不严控 — 普通成员可看列表，操作列在视图内按角色控制）
        {
          path: ":workspaceId(\\d+)/members",
          name: "workspace-members",
          component: () => import("@/views/workspace/WorkspaceMembersView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 工作空间级版本跳板页（项目选择器 + 跨项目摘要）
        {
          path: ":workspaceId(\\d+)/versions",
          name: "workspace-versions",
          component: () => import("@/views/workspace/WorkspaceVersionsView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 工作空间级迭代跳板页
        {
          path: ":workspaceId(\\d+)/sprints",
          name: "workspace-sprints",
          component: () => import("@/views/workspace/WorkspaceSprintsView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 工作空间级报表跳板页（跨项目完成率/缺陷/迭代对比）
        {
          path: ":workspaceId(\\d+)/analytics",
          name: "workspace-analytics",
          component: () => import("@/views/workspace/WorkspaceAnalyticsView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 工作空间级自动化跳板页（项目选择器）
        {
          path: ":workspaceId(\\d+)/automation",
          name: "workspace-automation",
          component: () => import("@/views/workspace/WorkspaceAutomationView.vue"),
          props: (route) => ({ workspaceId: Number(route.params.workspaceId) }),
        },
        // 知识库：空间列表
        {
          path: ":workspaceId(\\d+)/knowledge",
          name: "knowledge-list",
          component: () => import("@/views/knowledge/KnowledgeListView.vue"),
        },
        // 知识库：空间内文档管理
        {
          path: ":workspaceId(\\d+)/knowledge/:spaceId",
          name: "knowledge-space",
          component: () => import("@/views/knowledge/KnowledgeSpaceView.vue"),
        },
        {
          path: ":workspaceId(\\d+)/projects",
          name: "projects",
          component: () => import("@/views/project/ProjectListView.vue"),
        },
        // 项目仪表盘
        {
          path: ":workspaceId(\\d+)/projects/:projectId/dashboard",
          name: "project-dashboard",
          component: () => import("@/views/project/DashboardView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 项目看板视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/board",
          name: "project-board",
          component: () => import("@/views/project/KanbanBoardView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // S11 自动化规则
        {
          path: ":workspaceId(\\d+)/projects/:projectId/automation",
          name: "project-automation",
          component: () => import("@/views/project/AutomationView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // S11 效能度量
        {
          path: ":workspaceId(\\d+)/projects/:projectId/metrics",
          name: "project-metrics",
          component: () => import("@/views/project/MetricsView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 缺陷分析报表
        {
          path: ":workspaceId(\\d+)/projects/:projectId/analytics",
          name: "defect-analytics",
          component: () => import("@/views/project/DefectAnalyticsView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 文档页面 (Pages)
        {
          path: ":workspaceId(\\d+)/projects/:projectId/pages",
          name: "project-pages",
          component: () => import("@/views/project/PagesView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 甘特图视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/gantt",
          name: "gantt-chart",
          component: () => import("@/views/project/GanttChartView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 日历视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/calendar",
          name: "calendar-view",
          component: () => import("@/views/project/CalendarView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 电子表格视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/spreadsheet",
          name: "spreadsheet-view",
          component: () => import("@/views/project/SpreadsheetView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 需求列表视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/requirements",
          name: "requirement-list",
          component: () => import("@/views/project/RequirementListView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 任务列表视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/tasks",
          name: "task-list",
          component: () => import("@/views/project/TaskListView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 缺陷列表视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/defects",
          name: "defect-list",
          component: () => import("@/views/project/DefectListView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 时间线视图（ECharts Gantt）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/timeline",
          name: "project-timeline",
          component: () => import("@/views/project/TimelineView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // WBS 树形视图
        {
          path: ":workspaceId(\\d+)/projects/:projectId/wbs",
          name: "wbs-tree",
          component: () => import("@/views/project/WbsTreeView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 项目设置
        {
          path: ":workspaceId(\\d+)/projects/:projectId/settings",
          name: "project-settings",
          component: () => import("@/views/project/ProjectSettingsView.vue"),
        },
        // 模块管理（项目设置子页）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/settings/modules",
          name: "project-modules",
          component: () => import("@/views/project/ModuleSettingsView.vue"),
        },
        // 模块管理独立页（路由名 project-modules-list 避免与 settings 子页 project-modules 冲突）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/modules",
          name: "project-modules-list",
          component: () => import("@/views/project/ModulesView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 成员管理（项目设置子页）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/settings/members",
          name: "project-members",
          component: () => import("@/views/project/ProjectMembersView.vue"),
        },
        // 回收站
        {
          path: ":workspaceId(\\d+)/projects/:projectId/trash",
          name: "project-trash",
          component: () => import("@/views/project/TrashView.vue"),
        },
        // 迭代列表
        {
          path: ":workspaceId(\\d+)/projects/:projectId/sprints",
          name: "sprint-list",
          component: () => import("@/views/project/SprintListView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 排期规划（Backlog ↔ Sprint 拖拽）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/sprints/planning",
          name: "sprint-planning",
          component: () => import("@/views/project/SprintPlanningView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        // 迭代详情
        {
          path: ":workspaceId(\\d+)/projects/:projectId/sprints/:sprintId",
          name: "sprint-detail",
          component: () => import("@/views/project/SprintDetailView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            sprintId: Number(route.params.sprintId),
          }),
        },
        // 站会模式
        {
          path: ":workspaceId(\\d+)/projects/:projectId/sprints/:sprintId/standup",
          name: "sprint-standup",
          component: () => import("@/views/project/SprintStandupView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            sprintId: Number(route.params.sprintId),
          }),
        },
        // 版本聚合
        {
          path: ":workspaceId(\\d+)/projects/:projectId/versions",
          name: "version-list",
          component: () => import("@/views/project/VersionListView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
          }),
        },
        {
          path: ":workspaceId(\\d+)/projects/:projectId/versions/:versionId",
          name: "version-detail",
          component: () => import("@/views/project/VersionDetailView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        {
          path: ":workspaceId(\\d+)/projects/:projectId/versions/:versionId/release",
          name: "version-release",
          component: () => import("@/views/project/VersionReleaseView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        {
          path: ":workspaceId(\\d+)/projects/:projectId/versions/:versionId/delivery-report",
          name: "delivery-report",
          component: () => import("@/views/project/DeliveryReportView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        // 工作项沉浸模式（Focus Mode）
        {
          path: ":workspaceId(\\d+)/projects/:projectId/issues/:issueId/focus",
          name: "issue-focus",
          component: () => import("@/views/project/FocusModeView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            issueId: Number(route.params.issueId),
          }),
        },
        // 工作项详情页
        {
          path: ":workspaceId(\\d+)/projects/:projectId/issues/:issueId",
          name: "issue-detail",
          component: () => import("@/views/project/IssueDetailView.vue"),
          props: (route) => ({
            workspaceId: Number(route.params.workspaceId),
            projectId: Number(route.params.projectId),
            issueId: Number(route.params.issueId),
          }),
        },
      ],
    },
    {
      path: "/:pathMatch(.*)*",
      name: "not-found",
      component: () => import("@/views/NotFoundView.vue"),
      meta: { public: true },
    },
    // 403 工作空间权限不足（放在通配符兜底之前）
    {
      path: "/forbidden",
      name: "forbidden",
      component: () => import("@/views/ForbiddenView.vue"),
    },
  ],
});

/** 全局前置守卫：恢复会话、强制登录、登录态下禁止访问认证页、工作空间级权限检查 */
router.beforeEach(async (to) => {
  const auth = useAuthStore();
  const wsStore = useWorkspaceStore();

  // 仅在需要认证时才恢复会话：避免公开路由（如 /login）上
  // 因 fetchMe 失败触发拦截器 redirect 导致循环刷新
  const authPages = ["login", "register", "forgot-password", "reset-password"];
  const isAuthPage = authPages.includes(String(to.name));

  if (!auth.loaded && !isAuthPage && !to.meta.public) {
    await auth.fetchMe();
  }

  // 确保状态已标记（跳过 fetchMe 时仍需标记 loaded）
  if (!auth.loaded) {
    auth.loaded = true;
  }

  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  if (isAuthPage && auth.isAuthenticated) {
    return { name: "home" };
  }

  // 工作空间级权限守卫：进入含 :workspaceId 的路由时确保权限已加载，
  // 若目标路由有 meta.permission 要求，则校验。
  const wsId = Number(to.params.workspaceId ?? 0);
  if (wsId > 0 && auth.isAuthenticated) {
    // 确保工作空间已加载（含权限集合）
    if ((!wsStore.current || wsStore.current.id !== wsId) || wsStore.permissions.size === 0) {
      await wsStore.resolveById(wsId);
    }
    const required = WORKSPACE_PERMISSIOND_ROUTES[String(to.name ?? "")];
    if (required && !wsStore.hasPermission(required)) {
      return { name: "forbidden" };
    }
    // 空串表示仅 owner / admin 可访问（跳过 permission 集合判断，直接校验角色）
    if (required === "" && !wsStore.canManage) {
      return { name: "forbidden" };
    }
  }

  return true;
});

export default router;
