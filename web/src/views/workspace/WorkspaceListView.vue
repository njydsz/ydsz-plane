<script setup lang="ts">
/**
 * 工作空间列表页 — 展示/创建工作空间，进入项目列表。
 */

import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { workspaceApi, type Workspace } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import CreateWorkspaceModal from "./CreateWorkspaceModal.vue";

const router = useRouter();
const auth = useAuthStore();
const workspaces = ref<Workspace[]>([]);
const loading = ref(true);
const error = ref("");
const showCreate = ref(false);

function roleLabel(role?: string): string {
  const map: Record<string, string> = { owner: "所有者", admin: "管理员", member: "成员", guest: "访客" };
  return map[role ?? ""] ?? "";
}

async function load() {
  loading.value = true;
  try {
    workspaces.value = await workspaceApi.list();
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

function openWs(ws: Workspace) {
  router.push(`/${ws.slug}/projects`);
}

function onCreateCreated(ws: Workspace) {
  showCreate.value = false;
  workspaces.value.unshift(ws);
  router.push(`/${ws.slug}/projects`);
}

onMounted(load);
defineExpose({ roleLabel });
</script>

<template>
  <div class="ws-list">
    <header class="ws-list__header">
      <div>
        <h1>工作空间</h1>
        <p class="sub">Hi {{ auth.user?.display_name }}，选择一个工作空间开始工作</p>
      </div>
      <button class="btn btn--primary" @click="showCreate = true">创建工作空间</button>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="workspaces.length === 0" class="empty">
      <p>您还没有加入任何工作空间</p>
      <button class="btn btn--primary" @click="showCreate = true">创建工作空间</button>
    </div>
    <div v-else class="ws-grid">
      <div
        v-for="ws in workspaces"
        :key="ws.id"
        class="ws-card"
        @click="openWs(ws)"
      >
        <div class="ws-card__avatar">
          <span class="ws-card__initials">{{ ws.name.charAt(0) }}</span>
        </div>
        <div class="ws-card__body">
          <div class="ws-card__name">{{ ws.name }}</div>
          <div class="ws-card__meta">
            <span class="ws-card__role" :data-role="ws.role">{{ roleLabel(ws.role) }}</span>
            <span>· {{ ws.member_count ?? 0 }} 成员</span>
          </div>
        </div>
      </div>
      <div class="ws-card ws-card--add" @click="showCreate = true">
        <span class="ws-card__plus">+</span>
        <span>创建新工作空间</span>
      </div>
    </div>

    <CreateWorkspaceModal
      v-if="showCreate"
      @close="showCreate = false"
      @created="onCreateCreated"
    />
  </div>
</template>

<style scoped>
.ws-list__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.ws-list__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}

.sub {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}

.loading,
.error,
.empty {
  text-align: center;
  color: var(--text-tertiary);
  padding: 48px 0;
}

.error {
  color: var(--danger-500);
}

.ws-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 14px;
}

.ws-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.ws-card:hover {
  border-color: var(--brand-500);
  box-shadow: var(--shadow-card);
}

.ws-card__avatar {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  background: var(--brand-50);
  color: var(--brand-600);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 16px;
  overflow: hidden;
}

.ws-card__body {
  min-width: 0;
  flex: 1;
}

.ws-card__name {
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ws-card__meta {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.ws-card__role {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  background: var(--surface-3);
  color: var(--text-secondary);
}

.ws-card__role[data-role="owner"] {
  background: var(--brand-50);
  color: var(--brand-600);
}

.ws-card--add {
  border-style: dashed;
  color: var(--text-tertiary);
  justify-content: center;
  font-size: 13px;
}

.ws-card--add:hover {
  color: var(--brand-600);
}

.ws-card__plus {
  font-size: 20px;
  font-weight: 600;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:hover {
  background: var(--brand-600);
}
</style>
