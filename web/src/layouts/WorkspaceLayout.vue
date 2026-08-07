<script setup lang="ts">
/**
 * 工作空间布局组件 — 登录后应用的主框架。
 *
 * 职责：
 *  - 左侧边栏：工作空间切换器（支持搜索过滤）、导航菜单、用户区；
 *  - 顶部栏：页面标题/操作区；
 *  - 主内容区：路由出口（项目列表/看板/迭代/版本等子页面）。
 *  - 挂载时按 URL slug 解析并切换当前工作空间。
 */
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import type { Workspace } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import { useWorkspaceStore } from "@/stores/workspace";
import { useNotificationStore } from "@/stores/notification";
import { wsClient } from "@/lib/ws-client";
import NotificationBell from "@/components/NotificationBell.vue";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const wsStore = useWorkspaceStore();
const notifStore = useNotificationStore();

const collapsed = ref(false);
const showSwitcher = ref(false);
const switcherFilter = ref("");

const slug = computed(() => String(route.params.workspaceSlug ?? ""));
const workspaceList = computed(() => wsStore.list);
const currentWs = computed(() => wsStore.current);
const filteredWorkspaces = computed(() => {
  const q = switcherFilter.value.trim().toLowerCase();
  if (!q) return workspaceList.value;
  return workspaceList.value.filter(
    (w) => w.name.toLowerCase().includes(q) || w.slug.toLowerCase().includes(q),
  );
});

/** 角色 → 中文展示名映射 */
function roleLabel(role?: string): string {
  const map: Record<string, string> = { owner: "所有者", admin: "管理员", member: "成员", guest: "访客" };
  return map[role ?? ""] ?? "";
}

/** 初始化：加载空间列表并按 URL slug 切换当前空间 */
async function bootstrap() {
  await wsStore.load();
  if (slug.value) {
    await wsStore.resolveBySlug(slug.value);
  }
  // 注意：在根路由（工作空间列表页）不主动 redirect，由用户自主点击
}

/* ===== WebSocket 实时通知 ===== */

/** 当前空间 ID（number 时建立连接） */
const currentWsId = computed(() => currentWs.value?.id ?? null);

/** 新通知到达时刷新未读数 */
function handleNotification() {
  if (!currentWsId.value) return;
  void notifStore.fetchUnreadCount(currentWsId.value);
  // 若铃铛面板已打开（items 非空），同步列表
  if (notifStore.items.length > 0) {
    void notifStore.fetchList(currentWsId.value, { limit: 20 });
  }
}

/** 工作项/迭代/版本变更也会刷新未读数（可能伴随新通知） */
function handleAnyChange() {
  if (!currentWsId.value) return;
  void notifStore.fetchUnreadCount(currentWsId.value);
}

watch(
  currentWsId,
  (id) => {
    // 切换空间：断开旧连接，建立新连接
    wsClient.disconnect();
    if (id != null) {
      wsClient.connect(id, auth.user?.id);
      wsClient.on("notification.created", handleNotification);
      wsClient.on("notification.updated", handleNotification);
      wsClient.on("issue.updated", handleAnyChange);
      wsClient.on("issue.created", handleAnyChange);
      wsClient.on("comment.created", handleAnyChange);
      wsClient.on("sprint.started", handleAnyChange);
      wsClient.on("version.released", handleAnyChange);
      // 初始拉取未读数
      void notifStore.fetchUnreadCount(id);
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  wsClient.disconnect();
});

/** 选中工作空间：关闭切换器并跳转到该项目列表 */
function selectWs(ws: Workspace) {
  showSwitcher.value = false;
  router.push(`/${ws.slug}/projects`);
}

/** 跳转创建空间页（工作空间列表页） */
function gotoCreate() {
  showSwitcher.value = false;
  router.push("/");
}

/** 跳转工作空间列表页 */
function gotoList() {
  showSwitcher.value = false;
  router.push("/");
}

onMounted(bootstrap);
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
        <router-link
          :to="`/${wsStore.currentSlug}/workbench`"
          class="nav-item"
          active-class="is-active"
        >
          <span class="nav-icon">📊</span>
          <span v-if="!collapsed">工作台</span>
        </router-link>
        <router-link
          :to="`/${wsStore.currentSlug}/projects`"
          class="nav-item"
          active-class="is-active"
        >
          <span class="nav-icon">▦</span>
          <span v-if="!collapsed">项目</span>
        </router-link>
        <router-link
          :to="`/${wsStore.currentSlug}/settings`"
          class="nav-item"
          active-class="is-active"
        >
          <span class="nav-icon">⚙</span>
          <span v-if="!collapsed">设置</span>
        </router-link>
        <router-link
          :to="`/${wsStore.currentSlug}/settings/notifications`"
          class="nav-item"
          active-class="is-active"
        >
          <span class="nav-icon">🔔</span>
          <span v-if="!collapsed">通知设置</span>
        </router-link>
      </nav>

      <button class="sidebar__collapse" @click="collapsed = !collapsed">
        {{ collapsed ? "»" : "«" }}
      </button>
    </aside>

    <div class="main">
      <header class="header">
        <div class="header__breadcrumb">
          <slot name="breadcrumb">
            <span v-if="currentWs" class="crumb">{{ currentWs.name }}</span>
          </slot>
        </div>
        <div class="header__actions">
          <kbd class="cmdk-hint" title="搜索 (Ctrl+K)" @click="$emit('search')">⌘ K</kbd>
          <NotificationBell />
          <span class="user">{{ auth.user?.display_name ?? "" }}</span>
          <button class="logout" @click="auth.logout()">退出</button>
        </div>
      </header>
      <main class="content">
        <router-view />
      </main>
    </div>
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

/* ===== Nav ===== */
.sidebar__nav {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
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

.nav-icon {
  width: 18px;
  text-align: center;
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

.user {
  color: var(--text-secondary);
  font-size: 13px;
}

.logout {
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
}

.logout:hover {
  color: var(--danger-500);
}

.content {
  flex: 1;
  overflow: auto;
  padding: 24px;
}
</style>
