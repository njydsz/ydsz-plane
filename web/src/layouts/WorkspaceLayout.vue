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
import DynamicIcon from "@/components/DynamicIcon.vue";
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

// ===== S13 P1: 移动响应式 =====
const MOBILE_BREAKPOINT = 768;
const isMobile = ref(window.innerWidth < MOBILE_BREAKPOINT);
const mobileSidebarOpen = ref(false);
function handleResize() {
  const wasMobile = isMobile.value;
  isMobile.value = window.innerWidth < MOBILE_BREAKPOINT;
  // 从移动端切换到桌面端时关闭抽屉
  if (wasMobile && !isMobile.value) mobileSidebarOpen.value = false;
}
function toggleMobileSidebar() { mobileSidebarOpen.value = !mobileSidebarOpen.value; }
function closeMobileSidebar() { mobileSidebarOpen.value = false; }

// S13 P1: 移动侧栏打开时锁定 body 滚动，防止背景内容滚动
watch(mobileSidebarOpen, (open) => {
  if (isMobile.value) {
    document.body.style.overflow = open ? "hidden" : "";
  }
});

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const workspaceList = computed(() => wsStore.list);
const currentWs = computed(() => wsStore.current);
/** 是否处于首页（工作空间列表/创建页）。首页下侧栏只显示「工作台」入口，
 *  顶部快速创建按钮也切换为「新建空间」，符合「登录后落点 = 空间管理」的语义。 */
const isHome = computed(() => route.path === "/");
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

// S13 P1: 监听窗口 resize → 自动切换到移动/桌面布局
onMounted(() => {
  window.addEventListener("resize", handleResize);
});
onUnmounted(() => {
  wsClient.disconnect();
  window.removeEventListener("resize", handleResize);
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

/** 侧栏菜单项：
 *  - 首页下只显示「工作台」
 *  - 项目上下文（存在 currentProjectId）下隐藏工作空间主菜单，
 *    仅由「项目子菜单」承载导航；
 *  - 其他路由按权限全量展示。 */
const sidebarMenu = computed(() => {
  if (isHome.value) {
    return [
      {
        name: "workbench-home",
        to: { name: "home" },
        icon: "LayoutDashboard",
        labelKey: "__workbench__",
        isRaw: true,
      },
    ];
  }
  // 进入项目内：屏蔽工作空间级菜单（项目/版本/迭代/知识库/自动化等），
  // 避免和下方项目子菜单重复，也避免给用户"我在哪儿"的歧义。
  if (currentProjectId.value) {
    return [];
  }
  return visibleMainMenu.value.map((item) => ({
    name: item.name,
    to: `/${wsStore.currentId}/${item.path}`,
    icon: item.icon,
    labelKey: item.titleKey,
    isRaw: false,
  }));
});

/** 快速创建：在项目上下文中新建需求/任务/缺陷，否则新建项目到空间根；
 *  在首页（`/`）下，行为切换为新建工作空间，并保留 URL 上的 action 标记
 *  以让 WorkspaceListView 自动展开新建表单。 */
function handleQuickCreate() {
  if (isHome.value) {
    router.push({ path: "/", query: { action: "create" } });
    return;
  }
  if (currentProjectId.value && workspaceId.value) {
    // 注意：直接跳转到项目需求/任务/缺陷列表页，用户可在该页新建；
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
  <div class="ws-layout" :class="{ 'is-mobile': isMobile, 'mobile-sidebar-open': mobileSidebarOpen }">
    <!-- S13 P1: 移动抽屉遮罩层 -->
    <div
      v-if="isMobile && mobileSidebarOpen"
      class="mobile-overlay"
      @click="closeMobileSidebar"
    ></div>

    <aside
      class="sidebar"
      :class="{ collapsed, 'mobile-drawer': isMobile, 'drawer-open': mobileSidebarOpen }"
    >
      <!-- S13 P1: 移动侧栏关闭按钮 -->
      <button
        v-if="isMobile"
        class="mobile-close-btn"
        aria-label="关闭菜单"
        @click="closeMobileSidebar"
      >
        ✕
      </button>

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

        <!--
          切换器下拉：必须放在 .ws-switcher 内部，
          这样 position: absolute 才会以 ws-switcher (relative) 为参照，
          否则会跳过 sibling 直接定位到 sidebar，跑到底部。
        -->
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
      </div>

      <!-- 快速创建按钮（collapsed 时也显示为纯图标） -->
      <button
        class="quick-create"
        :class="{ 'quick-create--collapsed': collapsed }"
        :title="
          isHome
            ? '新建空间'
            : currentProjectId
            ? '新建需求/任务/缺陷'
            : '新建项目'
        "
        @click="handleQuickCreate"
      >
        <span v-if="!collapsed">
          ＋{{ isHome ? "新建空间" : currentProjectId ? "新建需求/任务/缺陷" : "新建项目" }}
        </span>
        <span v-else>＋</span>
      </button>

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

        <!-- 核心导航（权限驱动）：首页下只渲染「工作台」 -->
        <div v-if="sidebarMenu.length > 0" class="nav-group">
          <router-link
            v-for="item in sidebarMenu"
            :key="item.name"
            :to="item.to"
            class="nav-item"
            active-class="is-active"
            exact-active-class="is-active"
          >
            <DynamicIcon :name="item.icon" class="nav-icon" />
            <span v-if="!collapsed">
              {{ item.isRaw ? '工作台' : $t(item.labelKey) }}
            </span>
          </router-link>
        </div>

        <!-- 项目子导航 -->
        <template v-if="currentProjectId">
          <div class="nav-group">
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/board`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="KanbanSquare" class="nav-icon" />
              <span v-if="!collapsed">看板</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/requirements`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="ClipboardList" class="nav-icon" />
              <span v-if="!collapsed">需求</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/tasks`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="SquareCheckBig" class="nav-icon" />
              <span v-if="!collapsed">任务</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/defects`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="Bug" class="nav-icon" />
              <span v-if="!collapsed">缺陷</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/pages`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="FileText" class="nav-icon" />
              <span v-if="!collapsed">文档</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/calendar`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="Calendar" class="nav-icon" />
              <span v-if="!collapsed">日历</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/metrics`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="Gauge" class="nav-icon" />
              <span v-if="!collapsed">效能</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/analytics`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="LineChart" class="nav-icon" />
              <span v-if="!collapsed">分析</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/gantt`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="GanttChart" class="nav-icon" />
              <span v-if="!collapsed">甘特图</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/timeline`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="History" class="nav-icon" />
              <span v-if="!collapsed">时间线</span>
            </router-link>
            <router-link
              :to="`/${wsStore.currentId}/projects/${currentProjectId}/dashboard`"
              class="nav-item"
              active-class="is-active"
              exact-active-class="is-active"
            >
              <DynamicIcon name="LayoutDashboard" class="nav-icon" />
              <span v-if="!collapsed">仪表盘</span>
            </router-link>
          </div>
        </template>
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
        <!-- S13 P1: 移动汉堡菜单按钮 -->
        <button
          v-if="isMobile"
          class="hamburger-btn"
          aria-label="打开菜单"
          @click="toggleMobileSidebar"
        >
          <span class="hamburger-icon"><i></i><i></i><i></i></span>
        </button>
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
  /* 紧贴 sidebar 顶部：去除外侧 padding，靠自身 min-height 撑开 */
  padding: 0 14px;
  min-height: var(--header-height);
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  user-select: none;
  position: relative;
  box-sizing: border-box;
  flex-shrink: 0;
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
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: var(--bg-accent-subtle);
  color: var(--txt-accent-primary);
  font-weight: 600;
  font-size: 13px;
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

/* ===== S13 P1: 移动响应式 ===== */

/* 汉堡按钮 */
.hamburger-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  margin-right: 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
}
.hamburger-icon {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 22px;
  height: 16px;
}
.hamburger-icon i {
  display: block;
  height: 2px;
  width: 100%;
  background: var(--text-primary);
  border-radius: 2px;
}

/* 移动遮罩 */
.mobile-overlay {
  position: fixed;
  inset: 0;
  z-index: 99;
  background: rgba(0, 0, 0, 0.45);
  animation: fadeIn 0.2s ease;
}
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

/* 移动侧栏抽屉 */
.sidebar.mobile-drawer {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 100;
  height: 100%;
  width: 280px !important;
  transform: translateX(-100%);
  transition: transform 0.25s ease;
  box-shadow: var(--shadow-drawer, 4px 0 24px rgba(0, 0, 0, 0.25));
}
.sidebar.mobile-drawer.drawer-open {
  transform: translateX(0);
}

/* 移动侧栏关闭按钮 */
.mobile-close-btn {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--surface-2);
  color: var(--text-primary);
  font-size: 16px;
  cursor: pointer;
}

/* 移动主内容区域：避免底部内容被遮挡 */
.is-mobile .main {
  width: 100%;
  overflow-x: hidden;
}

/* 触摸友好: 增大导航项触控面积 */
.is-mobile .nav-item {
  min-height: 44px;
}
.is-mobile .ws-switcher {
  min-height: 48px;
}

</style>
