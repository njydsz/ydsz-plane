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
  if (to.name === "login" && auth.isAuthenticated) {
    return { name: "home" };
  }
  return true;
});

export default router;
