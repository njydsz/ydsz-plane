<script setup lang="ts">
/**
 * TrashView — 回收站页面。
 * 展示已删除（软删除）的需求/任务/缺陷，支持恢复和彻底删除。
 */
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { issueApi, type TrashItem } from "@/api/services/issue";
import { toast } from "@/lib/toast";
import { AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppPageContainer } from "@/components";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
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
  if (!await confirm({ title: "彻底删除", message: "此操作不可撤销，确定要彻底删除该需求/任务/缺陷吗？", confirmText: "删除", variant: "danger" })) return;
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
  
  <AppPageContainer max="900px">
    <div class="trash-view">
    <header class="trash-view__header">
      <div>
        <h1>回收站</h1>
        <p class="meta">被归档的需求/任务/缺陷会保留 30 天后自动清理，你可以在此恢复。</p>
      </div>
      <AppButton variant="ghost" @click="router.back()">← 返回</AppButton>
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
          <AppButton
            size="sm"
            variant="secondary"
            :loading="restoring.has(item.id)"
            :disabled="restoring.has(item.id)"
            @click="restoreItem(item.id)"
          >
            {{ restoring.has(item.id) ? "恢复中..." : "恢复" }}
          </AppButton>
          <AppButton
            size="sm"
            variant="danger"
            :loading="deleting.has(item.id)"
            :disabled="deleting.has(item.id)"
            @click="deletePermanently(item.id)"
          >
            {{ deleting.has(item.id) ? "删除中..." : "彻底删除" }}
          </AppButton>
        </span>
      </div>
    </div>
  
    </div>
  </AppPageContainer>
</template>

<style scoped>
.trash-view {
  
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

</style>
