<script setup lang="ts">
import { ref } from "vue";

import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const collapsed = ref(false);
</script>

<template>
  <div class="ws-layout">
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sidebar__logo">
        <span class="logo-mark">YD</span>
        <span v-if="!collapsed" class="logo-text">Ydsz Plane</span>
      </div>
      <nav class="sidebar__nav">
        <router-link to="/" class="nav-item" exact-active-class="is-active">
          <span class="nav-icon">⌂</span>
          <span v-if="!collapsed">工作台</span>
        </router-link>
        <router-link to="/projects" class="nav-item" active-class="is-active">
          <span class="nav-icon">▦</span>
          <span v-if="!collapsed">项目</span>
        </router-link>
      </nav>
      <button class="sidebar__collapse" @click="collapsed = !collapsed">
        {{ collapsed ? "»" : "«" }}
      </button>
    </aside>

    <div class="main">
      <header class="header">
        <div class="header__breadcrumb"><slot name="breadcrumb" /></div>
        <div class="header__actions">
          <kbd class="cmdk-hint">Ctrl K</kbd>
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

.sidebar__logo {
  display: flex;
  align-items: center;
  gap: 10px;
  height: var(--header-height);
  padding: 0 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.logo-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 700;
  font-size: 12px;
}

.logo-text {
  font-weight: 600;
  color: var(--text-primary);
}

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
