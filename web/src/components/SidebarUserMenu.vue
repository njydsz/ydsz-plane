<script setup lang="ts">
/**
 * SidebarUserMenu — 侧栏底部用户菜单。
 *
 * 展示当前用户头像 + 名称；点击后弹出下拉菜单，包含：
 *  - 访问令牌（链接到工作空间设置 → 访问令牌 Tab）
 *  - 主题切换
 *  - 退出登录
 *
 * 大厂对标：Linear 的 sidebar footer、Notion 的 workspace menu、
 * Asana 的 profile dropdown。核心原则是将「用户级资源」
 * （API Token、首选项、退出）从头部移到侧栏底部，释放头部给
 * 页面级操作（搜索、视图切换）。
 */
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { useAuthStore } from "@/stores/auth";
import { useWorkspaceStore } from "@/stores/workspace";
import ThemeToggle from "@/components/ThemeToggle.vue";

const auth = useAuthStore();
const wsStore = useWorkspaceStore();
const router = useRouter();

function goToProfile() {
  close();
  router.push({ name: "profile" });
}

const open = ref(false);
const triggerRef = ref<HTMLElement | null>(null);

const displayName = computed(() => auth.user?.display_name ?? "");
const initial = computed(() => (displayName.value.charAt(0) || "?").toUpperCase());
const wsId = computed(() => wsStore.currentId);

function toggle() {
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function goToApiTokens() {
  close();
  // 访问令牌是用户级资源，但在工作空间设置页的 Tab 中管理。
  // 跳转到当前工作空间的设置 → 访问令牌 Tab。
  router.push(`/${wsId.value}/settings?tab=api-tokens`);
}

function handleLogout() {
  close();
  auth.logout();
}

/** 点击菜单外部关闭 */
function onClickOutside(e: MouseEvent) {
  if (!triggerRef.value) return;
  if (!triggerRef.value.contains(e.target as Node)) {
    close();
  }
}

onMounted(() => document.addEventListener("mousedown", onClickOutside));
onBeforeUnmount(() => document.removeEventListener("mousedown", onClickOutside));
</script>

<template>
  <div ref="triggerRef" class="user-menu">
    <button class="user-menu__trigger" @click="toggle">
      <span class="user-menu__avatar">{{ initial }}</span>
      <span class="user-menu__meta">
        <span class="user-menu__name">{{ displayName }}</span>
        <span class="user-menu__email">{{ auth.user?.email ?? "" }}</span>
      </span>
      <span class="user-menu__caret" :class="{ 'is-open': open }">▾</span>
    </button>

    <div v-if="open" class="user-menu__dropdown" role="menu">
      <div class="user-menu__header">
        <span class="user-menu__avatar user-menu__avatar--lg">{{ initial }}</span>
        <div>
          <div class="user-menu__name">{{ displayName }}</div>
          <div class="user-menu__email">{{ auth.user?.email ?? "" }}</div>
        </div>
      </div>
      <div class="user-menu__divider" />
      <button class="user-menu__item" role="menuitem" @click="goToProfile">
        <span class="user-menu__item-icon">👤</span>
        <span class="user-menu__item-text">
          <span class="user-menu__item-label">个人设置</span>
          <span class="user-menu__item-desc">编辑个人资料与偏好</span>
        </span>
      </button>
      <button class="user-menu__item" role="menuitem" @click="goToApiTokens">
        <span class="user-menu__item-icon">🔑</span>
        <span class="user-menu__item-text">
          <span class="user-menu__item-label">访问令牌</span>
          <span class="user-menu__item-desc">管理个人 API Token</span>
        </span>
      </button>
      <div class="user-menu__divider" />
      <div class="user-menu__item user-menu__item--row">
        <span class="user-menu__item-icon">🎨</span>
        <span class="user-menu__item-label">外观主题</span>
        <ThemeToggle />
      </div>
      <div class="user-menu__divider" />
      <button class="user-menu__item user-menu__item--danger" role="menuitem" @click="handleLogout">
        <span class="user-menu__item-icon">⎋</span>
        <span class="user-menu__item-label">退出登录</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.user-menu {
  position: relative;
  border-top: 1px solid var(--border-subtle);
  padding: 8px;
}

.user-menu__trigger {
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
  transition: background 0.1s ease;
}

.user-menu__trigger:hover {
  background: var(--surface-3);
}

.user-menu__avatar {
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
  flex-shrink: 0;
}

.user-menu__avatar--lg {
  width: 36px;
  height: 36px;
  font-size: 14px;
}

.user-menu__meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.user-menu__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-menu__email {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-menu__caret {
  font-size: 10px;
  color: var(--text-tertiary);
  transition: transform 0.15s ease;
}

.user-menu__caret.is-open {
  transform: rotate(180deg);
}

/* ===== Dropdown ===== */
.user-menu__dropdown {
  position: absolute;
  bottom: 100%;
  left: 8px;
  right: 8px;
  margin-bottom: 4px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
  z-index: 300;
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.user-menu__header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
}

.user-menu__divider {
  height: 1px;
  background: var(--border-subtle);
  margin: 2px 0;
}

.user-menu__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  font-size: 13px;
  color: var(--text-secondary);
  width: 100%;
  transition: background 0.1s ease;
}

.user-menu__item:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.user-menu__item--row {
  justify-content: space-between;
}

.user-menu__item-icon {
  width: 18px;
  text-align: center;
  flex-shrink: 0;
}

.user-menu__item-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-menu__item-label {
  font-size: 13px;
  color: var(--text-primary);
}

.user-menu__item-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

.user-menu__item--danger:hover {
  background: var(--danger-50);
  color: var(--danger-600);
}

.user-menu__item--danger:hover .user-menu__item-label {
  color: var(--danger-600);
}
</style>
