<script setup lang="ts">
/**
 * 工作空间布局组件 — 登录后应用的主框架。
 *
 * 职责：
 *  - 左侧边栏：工作空间切换器（支持搜索过滤）、收藏区、导航菜单（分组折叠）、用户区
 *  - 顶部栏：页面标题/操作区
 *  - 主内容区：路由出口（项目列表/看板/迭代/版本等子页面）
 *  - 挂载时按 URL slug 解析并切换当前工作空间
 *
 * P2 增强：
 *  - 收藏区（Favorites）：可按 ★ 收藏项目/侧栏快捷入口
 *  - 分组折叠：设置分组可折叠，减少视觉噪音
 *  - 空间切换器：最近访问 + 收藏置顶
 */
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type Workspace } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import { useSearchStore } from "@/stores/search";
import { useWorkspaceStore } from "@/stores/workspace";
import { useFavoritesStore } from "@/stores/favorites";
import { useNotificationStore } from "@/stores/notification";
import { wsClient } from "@/lib/ws-client";
import NotificationBell from "@/components/NotificationBell.vue";
import { IssuePeekOverview } from "@/components";
import ThemeToggle from "@/components/ThemeToggle.vue";
import SidebarUserMenu from "@/components/SidebarUserMenu.vue";
import { WORKSPACE_MENU } from "@/types/permission";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const wsStore = useWorkspaceStore();
const favoritesStore = useFavoritesStore();
const notifStore = useNotificationStore();
const searchStore = useSearchStore();

const collapsed = ref(false);
const showSwitcher = ref(false);
const switcherFilter = ref("");
/** 设置分组是否折叠 */
const settingsGroupCollapsed = ref(true);

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const workspaceList = computed(() => wsStore.list);
const currentWs = computed(() => wsStore.current);
const currentProjectId = computed(() => {
  const pid = route.params.projectId;
  return pid ? Number(pid) : null;
});
const filteredWorkspaces = computed(() => {
  const q = switcherFilter.value.trim().toLowerCase();
  if (!q) return workspaceList.value;
  return workspaceList.value.filter(
    (w) => w.name.toLowerCase().includes(q) || w.slug.toLowerCase().includes(q),
  );
});

/** 当前侧边栏渲染的 fav 条目（仅展示前 8 条） */
const sidebarFavorites = computed(() => favoritesStore.favorites.slice(0, 8));


function roleLabel(role?: string): string {
  const map: Record<string, string> = {
    admin: "系统管理员",
    owner: "空间管理员",
    pm: "项目经理",
    po: "产品经理",
    techlead: "技术经理",
    qalead: "测试经理",
    dev: "开发",
    guest: "访客",
    member: "成员",
  };
  return map[role ?? ""] ?? "";
}

/** 过滤后的侧边栏主菜单：权限 + 项目模块开关过滤 */
const visibleMainMenu = computed(() =>
  WORKSPACE_MENU.filter((item) => {
    if (!wsStore.canSeeMenu(item)) return false;
    if (item.moduleKey && !wsStore.isProjectModuleEnabled(item.moduleKey)) return false;
    return true;
  }),
);

/* ===== 初始化与 WebSocket ===== */

async function bootstrap() {
  await wsStore.load();
  if (workspaceId.value) {
    await wsStore.resolveById(workspaceId.value);
  }
}

function handleNotification() {
  if (!wsStore.currentId) return;
  void notifStore.fetchUnreadCount(wsStore.currentId);
  if (notifStore.items.length > 0) {
    void notifStore.fetchList(wsStore.currentId, { limit: 20 });
  }
}

function handleAnyChange() {
  if (!wsStore.currentId) return;
  void notifStore.fetchUnreadCount(wsStore.currentId);
}

function handleReconnectCompensation() {
  if (!wsStore.currentId) return;
  const since = wsClient.lastDisconnectTimestamp;
  void notifStore.fetchUnreadCount(wsStore.currentId);
  void notifStore.fetchList(wsStore.currentId, { limit: 20, since: since || undefined });
}

watch(
  () => wsStore.currentId,
  (id) => {
    wsClient.disconnect();
    favoritesStore.setWorkspace(id ?? null);
    if (id) {
      wsClient.connect(id, auth.user?.id);
      wsClient.on("notification.created", handleNotification);
      wsClient.on("notification.updated", handleNotification);
      wsClient.on("issue.updated", handleAnyChange);
      wsClient.on("issue.created", handleAnyChange);
      wsClient.on("comment.created", handleAnyChange);
      wsClient.on("sprint.started", handleAnyChange);
      wsClient.on("version.released", handleAnyChange);
      wsClient.onReconnect(handleReconnectCompensation);
      void notifStore.fetchUnreadCount(id);
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  wsClient.disconnect();
});

/* ===== 空间切换器 ===== */

function selectWs(ws: Workspace) {
  showSwitcher.value = false;
  router.push(`/${ws.id}/projects`);
}

function gotoCreate() {
  showSwitcher.value = false;
  router.push("/");
}

function gotoList() {
  showSwitcher.value = false;
  router.push("/");
}

/** 快速创建：在项目上下文中新建工作项，否则新建项目到空间根 */
function handleQuickCreate() {
  if (currentProjectId.value && workspaceId.value) {
    // 注意：直接跳转到项目工作项列表页，用户可在该页新建；
    // 后续可升级为内联创建弹层（P2）。
    router.push(`/${workspaceId.value}/projects/${currentProjectId.value}/list?action=create`);
  } else if (workspaceId.value) {
    router.push(`/${workspaceId.value}/projects?action=create`);
  }
}

onMounted(bootstrap);

/** 项目切换时加载功能模块开关到 store，驱动侧边栏按项目模块过滤 */
watch(
  () => currentProjectId.value,
  async (pid) => {
    if (!pid || !wsStore.currentId) {
      wsStore.setProjectModules(null);
      return;
    }
    try {
      const p = await workspaceApi.getProject(wsStore.currentId, pid);
      wsStore.setProjectModules(p.modules ?? null);
    } catch {
      wsStore.setProjectModules(null);
    }
  },
  { immediate: true },
);

</script>

<template>
  <div class="ws-layout">
    <aside class="sidebar" :class="{ collapsed }">
      <!-- ===== 工作空间切换器 ===== -->
      <div class="ws-switcher" @click="showSwitcher = !showSwitcher">
        <div class="ws-switcher__avatar">
          <span v-if="currentWs">{{ currentWs.name.charAt(0) }}</span>
          <span v-else>?</span>
        </div>
        <div v-if="!collapsed" class="ws-switcher__meta">
          <span class="ws-switcher__name">{{ currentWs?.name ?? "选择空间" }}</span>
          <span class="ws-switcher__role">{{ roleLabel(currentWs?.role) }}</span>
        </div>
        <span v-if="!collapsed" class="ws-switcher__caret">▾</span>
      </div>

      <!-- 快速创建按钮（collapsed 时也显示为纯图标） -->
      <button
        class="quick-create"
        :class="{ 'quick-create--collapsed': collapsed }"
        :title="currentProjectId ? '新建工作项' : '新建项目'"
        @click="handleQuickCreate"
      >
        <span v-if="!collapsed">＋{{ currentProjectId ? "新建工作项" : "新建项目" }}</span>
        <span v-else>＋</span>
      </button>

      <!-- 切换器下拉 -->
      <div v-if="showSwitcher" class="ws-switcher__dropdown" @click.stop>
        <input
          v-model="switcherFilter"
          class="ws-switcher__search"
          placeholder="搜索工作空间..."
          autofocus
        />
        <div class="ws-switcher__list">
          <button
            v-for="ws in filteredWorkspaces"
            :key="ws.id"
            class="ws-switcher__item"
            :class="{ active: ws.id === currentWs?.id }"
            @click="selectWs(ws)"
          >
            <span class="ws-item__avatar">{{ ws.name.charAt(0) }}</span>
            <span class="ws-item__info">
              <span class="ws-item__name">{{ ws.name }}</span>
              <span class="ws-item__sub">{{ ws.member_count }} 成员 · {{ roleLabel(ws.role) }}</span>
            </span>
          </button>
        </div>
        <div class="ws-switcher__actions">
          <button class="ws-switcher__action" @click="gotoList">📋 查看所有</button>
          <button class="ws-switcher__action" @click="gotoCreate">＋ 创建新空间</button>
        </div>
      </div>

      <!-- ===== 侧边导航 ===== -->
      <nav class="sidebar__nav">
        <!-- 收藏区 -->
        <template v-if="!collapsed && sidebarFavorites.length > 0">
          <div class="nav-group">
            <span class="nav-group__label">收藏</span>
            <router-link
              v-for="fav in sidebarFavorites"
              :key="fav.id"
              :to="fav.path"
              class="nav-item nav-item--fav"
              active-class="is-active"
              :title="fav.label"
            >
              <span class="nav-icon">{{ fav.icon }}</span>
              <span class="nav-label">{{ fav.label }}</span>
            </router-link>
          </div>
        </template>

        <!-- 核心导航（权限驱动） -->
        <div class="nav-group">
          <router-link
            v-for="item in visibleMainMenu"
            :key="item.name"
            :to="`/${wsStore.currentId}/${item.path}`"
            class="nav-item"
            active-class="is-active"
            exact-active-class="is-active"
          >
            <span class="nav-icon">▸</span>
            <span v-if="!collapsed">{{ item.titleKey }}</span>
          </router-link>
        </div>

        <!-- 项目子导航 -->
        <template v-if="currentProjectId">
          <div class="nav-group">
            <span v-if="!collapsed" class="nav-group__label">项目</span>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/dashboard`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📈</span>
              <span v-if="!collapsed">仪表盘</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/board`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">▥</span>
              <span v-if="!collapsed">看板</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/list`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">☰</span>
              <span v-if="!collapsed">列表</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/pages`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📄</span>
              <span v-if="!collapsed">文档</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/metrics`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📉</span>
              <span v-if="!collapsed">效能</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/analytics`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📋</span>
              <span v-if="!collapsed">分析</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/gantt`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📅</span>
              <span v-if="!collapsed">甘特图</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/calendar`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">🗓</span>
              <span v-if="!collapsed">日历</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/automation`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">⚡</span>
              <span v-if="!collapsed">自动化</span>
            </router-link>
          </div>
        </template>

        <!-- 设置分组（可折叠） -->
        <div v-if="wsStore.canManage" class="nav-group nav-group--collapsible">
          <button
            class="nav-group__toggle"
            @click="settingsGroupCollapsed = !settingsGroupCollapsed"
          >
            <span class="nav-icon">⚙</span>
            <span v-if="!collapsed" class="nav-group__label">管理</span>
            <span v-if="!collapsed" class="nav-group__chevron">
              {{ settingsGroupCollapsed ? "▸" : "▾" }}
            </span>
          </button>
          <div v-if="!settingsGroupCollapsed" class="nav-group__items">
            <router-link
              v-if="wsStore.hasPermission('workspace:update')"
              :to="`/${wsStore.currentId}/settings`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">⚙</span>
              <span v-if="!collapsed">工作空间设置</span>
            </router-link>
            <router-link
              v-if="wsStore.hasPermission('member:change_role')"
              :to="`/${wsStore.currentId}/members`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">👥</span>
              <span v-if="!collapsed">成员管理</span>
            </router-link>
            <router-link
              v-if="wsStore.hasPermission('intake:manage')"
              :to="`/${wsStore.currentId}/settings/intake`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📥</span>
              <span v-if="!collapsed">收件箱</span>
            </router-link>
            <router-link
              v-if="wsStore.hasPermission('webhook:manage')"
              :to="`/${wsStore.currentId}/settings/webhooks`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">🔗</span>
              <span v-if="!collapsed">Webhook</span>
            </router-link>
            <router-link
              v-if="wsStore.hasPermission('audit:read')"
              :to="`/${wsStore.currentId}/audit-logs`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">📜</span>
              <span v-if="!collapsed">审计日志</span>
            </router-link>
            <router-link
              v-if="wsStore.hasPermission('automation:manage')"
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/automation`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">⚡</span>
              <span v-if="!collapsed">自动化</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/settings/notifications`"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">🔔</span>
              <span v-if="!collapsed">通知设置</span>
            </router-link>
            <router-link
              to="/settings/profile"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">👤</span>
              <span v-if="!collapsed">个人资料</span>
            </router-link>
            <router-link
              to="/settings/api-tokens"
              class="nav-item nav-item--sub"
              active-class="is-active"
            >
              <span class="nav-icon">🔑</span>
              <span v-if="!collapsed">API Tokens</span>
            </router-link>
          </div>
        </div>
      </nav>

      <!-- 折叠按钮 -->
      <button class="sidebar__collapse" @click="collapsed = !collapsed">
        {{ collapsed ? "»" : "«" }}
      </button>

      <!-- 侧栏底部用户菜单（collapsed 时仅显示头像）-->
      <SidebarUserMenu v-if="!collapsed" class="sidebar__user" />
      <span v-else class="sidebar__user sidebar__user--collapsed">
        <span class="sidebar__user-avatar">{{ auth.user?.display_name?.charAt(0)?.toUpperCase() ?? "?" }}</span>
      </span>
    </aside>

    <div class="main">
      <header class="header">
        <div class="header__breadcrumb">
          <slot name="breadcrumb">
            <span v-if="currentWs" class="crumb">{{ currentWs.name }}</span>
          </slot>
        </div>
        <div class="header__actions">
          <kbd class="cmdk-hint" title="搜索 (Ctrl+K)" @click="searchStore.toggle()">⌘ K</kbd>
          <ThemeToggle />
          <NotificationBell />
        </div>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>

    <!-- 全局 Issue Peek Overview 抽屉 -->
    <IssuePeekOverview />
  </div>
</template>

<style scoped>
.ws-layout {
  display: flex;
  height: 100%;
  background: var(--surface-1);
}

.sidebar {
  width: var(--sidebar-width);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-subtle);
  background: var(--surface-2);
  transition: width 0.15s ease;
  position: relative;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

/* ===== Workspace Switcher ===== */
.ws-switcher {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  user-select: none;
  position: relative;
}

.ws-switcher:hover {
  background: var(--surface-3);
}

.ws-switcher__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 700;
  font-size: 13px;
  flex-shrink: 0;
}

.ws-switcher__meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.ws-switcher__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ws-switcher__role {
  font-size: 11px;
  color: var(--text-tertiary);
}

.ws-switcher__caret {
  color: var(--text-tertiary);
  font-size: 12px;
}

.ws-switcher__dropdown {
  position: absolute;
  top: 100%;
  left: 8px;
  right: 8px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
  z-index: 200;
  max-height: 340px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ws-switcher__search {
  margin: 8px;
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
  color: var(--text-primary);
  font-size: 12px;
  font-family: inherit;
  outline: none;
}

.ws-switcher__list {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
}

.ws-switcher__item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 6px 8px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
}

.ws-switcher__item:hover,
.ws-switcher__item.active {
  background: var(--surface-3);
}

.ws-item__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.ws-item__info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.ws-item__name {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ws-item__sub {
  font-size: 11px;
  color: var(--text-tertiary);
}

.ws-switcher__actions {
  border-top: 1px solid var(--border-subtle);
  padding: 4px;
  display: flex;
  flex-direction: column;
}

.ws-switcher__action {
  padding: 6px 8px;
  border: none;
  background: none;
  text-align: left;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  font-family: inherit;
}

.ws-switcher__action:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

/* ===== Quick Create  ===== */
.quick-create {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin: 6px 8px 4px;
  padding: 7px 10px;
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-sm);
  background: none;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  color: var(--brand-600);
  font-weight: 500;
  transition: all 0.15s ease;
}

.quick-create:hover {
  border-color: var(--brand-500);
  background: var(--brand-50);
  color: var(--brand-700);
}

.quick-create--collapsed {
  font-size: 18px;
  padding: 6px 0;
  margin: 6px 4px 4px;
  letter-spacing: 0;
}

/* ===== Nav ===== */
.sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-group {
  display: flex;
  flex-direction: column;
  gap: 1px;
  margin-bottom: 6px;
}

.nav-group__label {
  padding: 6px 10px 4px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Collapsible group */
.nav-group--collapsible {
  margin-bottom: 0;
}

.nav-group__toggle {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 10px;
  background: none;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  color: var(--text-secondary);
  text-align: left;
}

.nav-group__toggle:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.nav-group__chevron {
  margin-left: auto;
  font-size: 10px;
  color: var(--text-tertiary);
}

.nav-group__items {
  display: flex;
  flex-direction: column;
  gap: 1px;
  overflow: hidden;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 13px;
  text-decoration: none;
}

.nav-item:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.nav-item.is-active {
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.nav-item--sub {
  padding-left: 28px;
  font-size: 12px;
}

.nav-item--fav {
  position: relative;
}

.nav-item--fav .nav-icon {
  color: var(--warning-500);
}

.nav-icon {
  width: 18px;
  text-align: center;
}

.nav-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar__collapse {
  position: absolute;
  right: -12px;
  top: 50%;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-tertiary);
  cursor: pointer;
}

/* ===== Sidebar User (底部用户菜单) ===== */
.sidebar__user {
  flex-shrink: 0;
}

.sidebar__user--collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 0;
}

.sidebar__user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 700;
  font-size: 12px;
}

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.header {
  height: var(--header-height);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--surface-1);
}

.crumb {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.header__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cmdk-hint {
  padding: 2px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  color: var(--text-tertiary);
  font-size: 12px;
  font-family: var(--font-mono);
  cursor: pointer;
}

.content {
  flex: 1;
  overflow: auto;
  padding: 24px;
}
</style>
