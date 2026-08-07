import { createRouter, createWebHistory } from "vue-router";

import { useAuthStore } from "@/stores/auth";

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
    {
      path: "/",
      component: () => import("@/layouts/WorkspaceLayout.vue"),
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/views/workspace/WorkspaceListView.vue"),
        },
        // 邀请接受链接：公开可读，但 POST accept 需鉴权
        {
          path: "invite/:token",
          name: "invite-preview",
          component: () => import("@/views/workspace/InvitePreview.vue"),
          meta: { public: true },
        },
        {
          path: ":workspaceSlug/settings",
          name: "workspace-settings",
          component: () => import("@/views/workspace/WorkspaceSettingsView.vue"),
          props: (route) => ({ workspace_id: route.params.workspaceSlug }),
        },
        {
          path: ":workspaceSlug/projects",
          name: "projects",
          component: () => import("@/views/project/ProjectListView.vue"),
        },
        // 项目看板视图
        {
          path: ":workspaceSlug/projects/:projectId/board",
          name: "project-board",
          component: () => import("@/views/project/KanbanBoardView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
          }),
        },
        // 项目列表视图
        {
          path: ":workspaceSlug/projects/:projectId/list",
          name: "project-list",
          component: () => import("@/views/project/IssueListView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
          }),
        },
        // 项目设置
        {
          path: ":workspaceSlug/projects/:projectId/settings",
          name: "project-settings",
          component: () => import("@/views/project/ProjectSettingsView.vue"),
        },
        // 迭代列表
        {
          path: ":workspaceSlug/projects/:projectId/sprints",
          name: "sprint-list",
          component: () => import("@/views/project/SprintListView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
          }),
        },
        // 排期规划（Backlog ↔ Sprint 拖拽）
        {
          path: ":workspaceSlug/projects/:projectId/sprints/planning",
          name: "sprint-planning",
          component: () => import("@/views/project/SprintPlanningView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
          }),
        },
        // 迭代详情
        {
          path: ":workspaceSlug/projects/:projectId/sprints/:sprintId",
          name: "sprint-detail",
          component: () => import("@/views/project/SprintDetailView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
            sprintId: Number(route.params.sprintId),
          }),
        },
        // 版本日聚合
        {
          path: ":workspaceSlug/projects/:projectId/versions",
          name: "version-list",
          component: () => import("@/views/project/VersionListView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
          }),
        },
        {
          path: ":workspaceSlug/projects/:projectId/versions/:versionId",
          name: "version-detail",
          component: () => import("@/views/project/VersionDetailView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        {
          path: ":workspaceSlug/projects/:projectId/versions/:versionId/release",
          name: "version-release",
          component: () => import("@/views/project/VersionReleaseView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        {
          path: ":workspaceSlug/projects/:projectId/versions/:versionId/delivery-report",
          name: "delivery-report",
          component: () => import("@/views/project/DeliveryReportView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
            projectId: Number(route.params.projectId),
            versionId: Number(route.params.versionId),
          }),
        },
        // 工作项详情页
        {
          path: ":workspaceSlug/projects/:projectId/issues/:issueId",
          name: "issue-detail",
          component: () => import("@/views/project/IssueDetailView.vue"),
          props: (route) => ({
            workspaceSlug: route.params.workspaceSlug,
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
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.loaded) {
    await auth.fetchMe();
  }
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  const authPages = ["login", "register", "forgot-password", "reset-password"];
  if (authPages.includes(String(to.name)) && auth.isAuthenticated) {
    return { name: "home" };
  }
  return true;
});

export default router;
