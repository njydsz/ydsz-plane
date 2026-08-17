<script setup lang="ts">
/**
 * TrashView — 回收站页面。
 * 展示已删除（软删除）的需求/任务/缺陷，支持恢复和彻底删除。
 */
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { issueApi, type TrashItem } from "@/api/services/issue";
import { toast } from "@/lib/toast";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

const route = useRoute();
const router = useRouter();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const items = ref<TrashItem[]>([]);
const loading = ref(true);
const error = ref("");
const restoring = ref<Set<number>>(new Set());
const deleting = ref<Set<number>>(new Set());

async function load() {
  loading.value = true;
  error.value = "";
  try {
    items.value = await issueApi.listTrash(workspaceId, projectId);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载回收站失败";
  } finally {
    loading.value = false;
  }
}

async function restoreItem(id: number) {
  if (restoring.value.has(id)) return;
  restoring.value.add(id);
  try {
    await issueApi.restoreIssue(workspaceId, projectId, id);
    toast.success("需求/任务/缺陷已恢复");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "恢复失败");
  } finally {
    restoring.value.delete(id);
  }
}

async function deletePermanently(id: number) {
  if (!confirm("此操作不可撤销，确定要彻底删除该需求/任务/缺陷吗？")) return;
  if (deleting.value.has(id)) return;
  deleting.value.add(id);
  try {
    await issueApi.permanentDelete(workspaceId, projectId, id);
    toast.success("已彻底删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  } finally {
    deleting.value.delete(id);
  }
}

function typeLabel(type: string): string {
  return ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

onMounted(load);
</script>

<template>
  <div class="trash-view">
    <header class="trash-view__header">
      <div>
        <h1>回收站</h1>
        <p class="meta">被归档的需求/任务/缺陷会保留 30 天后自动清理，你可以在此恢复。</p>
      </div>
      <button class="btn btn--ghost" @click="router.back()">← 返回</button>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="items.length === 0"
      title="回收站为空"
      description="暂无已删除的需求/任务/缺陷"
    />

    <div v-else class="trash-table">
      <div class="trash-table__header">
        <span class="col-type">类型</span>
        <span class="col-name">名称</span>
        <span class="col-priority">优先级</span>
        <span class="col-date">删除时间</span>
        <span class="col-actions">操作</span>
      </div>
      <div v-for="item in items" :key="item.id" class="trash-row">
        <span class="col-type">
          <span class="badge" :class="`badge-${item.type_code}`">{{ typeLabel(item.type_code) }}</span>
        </span>
        <span class="col-name">{{ item.name }}</span>
        <span class="col-priority">{{ item.priority }}</span>
        <span class="col-date">{{ new Date(item.deleted_at).toLocaleString() }}</span>
        <span class="col-actions">
          <button
            class="btn btn--sm btn--outline"
            :disabled="restoring.has(item.id)"
            @click="restoreItem(item.id)"
          >
            {{ restoring.has(item.id) ? "恢复中..." : "恢复" }}
          </button>
          <button
            class="btn btn--sm btn--danger"
            :disabled="deleting.has(item.id)"
            @click="deletePermanently(item.id)"
          >
            {{ deleting.has(item.id) ? "删除中..." : "彻底删除" }}
          </button>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.trash-view {
  max-width: 900px;
}

.trash-view__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.trash-view__header h1 {
  font-size: 20px;
  margin: 0;
  color: var(--text-primary);
}

.meta {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

.trash-table__header {
  display: grid;
  grid-template-columns: 80px 1fr 80px 180px 160px;
  gap: 12px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  border-bottom: 1px solid var(--border-subtle);
}

.trash-row {
  display: grid;
  grid-template-columns: 80px 1fr 80px 180px 160px;
  gap: 12px;
  padding: 12px;
  align-items: center;
  border-bottom: 1px solid var(--border-subtle);
  font-size: 13px;
  color: var(--text-primary);
  transition: background 0.1s;
}

.trash-row:hover {
  background: var(--surface-2);
}

.col-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-priority, .col-date {
  font-size: 12px;
  color: var(--text-secondary);
}

.col-actions {
  display: flex;
  gap: 8px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: 500;
}

.badge-requirement { background: var(--brand-50); color: var(--brand-600); }
.badge-task { background: var(--success-50); color: var(--success-600); }
.badge-defect { background: var(--danger-50); color: var(--danger-600); }
.badge-epic { background: var(--warning-50); color: var(--warning-600); }

.btn {
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  border: 1px solid transparent;
  font-family: inherit;
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--brand-500);
  padding: 4px 0;
}

.btn--outline {
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--outline:hover:not(:disabled) {
  border-color: var(--brand-500);
  color: var(--brand-500);
}

.btn--danger {
  border-color: var(--danger-200);
  background: var(--danger-50);
  color: var(--danger-600);
}

.btn--danger:hover:not(:disabled) {
  background: var(--danger-100);
}

.btn:disabled, .btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
