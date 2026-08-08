<script setup lang="ts">
/**
 * DLQ 死信消息监控页 — 展示工作空间 RabbitMQ 死信队列详情，
 * 支持重试 / 清理 / 批量清理。
 *
 * UI 参考：RabbitMQ Management DLQ、SQS Dead Letter Queue Console。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { dlqApi, type DLQItem } from "@/api/services/dlq";
import { AppBadge } from "@/components";
import AppButton from "@/components/AppButton.vue";
import AppLoadingState from "@/components/AppLoadingState.vue";
import AppEmptyState from "@/components/AppEmptyState.vue";
import AppErrorState from "@/components/AppErrorState.vue";
import { formatRelativeTime } from "@/lib/formatTime";

const route = useRoute();
const wsId = computed(() => Number(route.params.workspaceId));

const items = ref<DLQItem[]>([]);
const total = ref(0);
const loading = ref(true);
const error = ref("");
const page = ref(1);
const pageSize = 20;
const unresolvedOnly = ref(false);
const selectedIds = ref<Set<number>>(new Set());
const retryingIds = ref<Set<number>>(new Set());
const cleaningIds = ref<Set<number>>(new Set());

// --- 统计 ---
const unresolvedCount = computed(() => items.value.filter((i) => !i.resolved_at).length);
const resolvedCount = computed(() => items.value.filter((i) => !!i.resolved_at).length);
const latestTime = computed(() => {
  if (items.value.length === 0) return "-";
  const sorted = [...items.value].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
  return formatRelativeTime(sorted[0].created_at);
});

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));

const allSelected = computed(() => {
  const visible = items.value.filter((i) => !i.resolved_at);
  return visible.length > 0 && visible.every((i) => selectedIds.value.has(i.id));
});

function statusVariant(item: DLQItem): "danger" | "success" {
  return item.resolved_at ? "success" : "danger";
}

function statusLabel(item: DLQItem): string {
  return item.resolved_at ? "已解决" : "未解决";
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const result = await dlqApi.list(wsId.value, {
      offset: (page.value - 1) * pageSize,
      limit: pageSize,
      unresolved_only: unresolvedOnly.value || undefined,
    });
    items.value = result.items;
    total.value = result.total;
    selectedIds.value.clear();
  } catch (e: any) {
    error.value = e.message ?? "加载死信消息失败";
  } finally {
    loading.value = false;
  }
}

async function retryItem(item: DLQItem) {
  if (retryingIds.value.has(item.id)) return;
  retryingIds.value.add(item.id);
  try {
    await dlqApi.retry(wsId.value, item.id);
    await load();
  } catch (e: any) {
    error.value = `重试失败: ${e.message ?? "未知错误"}`;
  } finally {
    retryingIds.value.delete(item.id);
  }
}

async function cleanItem(item: DLQItem) {
  if (cleaningIds.value.has(item.id)) return;
  cleaningIds.value.add(item.id);
  try {
    await dlqApi.remove(wsId.value, item.id);
    await load();
  } catch (e: any) {
    error.value = `清理失败: ${e.message ?? "未知错误"}`;
  } finally {
    cleaningIds.value.delete(item.id);
  }
}

function toggleSelect(id: number) {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id);
  } else {
    selectedIds.value.add(id);
  }
}

function toggleAll() {
  const visible = items.value.filter((i) => !i.resolved_at);
  if (allSelected.value) {
    visible.forEach((i) => selectedIds.value.delete(i.id));
  } else {
    visible.forEach((i) => selectedIds.value.add(i.id));
  }
}

async function batchCleanup() {
  if (selectedIds.value.size === 0) return;
  const ids = [...selectedIds.value];
  try {
    await dlqApi.cleanup(wsId.value, { event_ids: ids });
    selectedIds.value.clear();
    await load();
  } catch (e: any) {
    error.value = `批量清理失败: ${e.message ?? "未知错误"}`;
  }
}

async function cleanupAllResolved() {
  try {
    await dlqApi.cleanup(wsId.value, { resolved_all: true });
    await load();
  } catch (e: any) {
    error.value = `全部清理失败: ${e.message ?? "未知错误"}`;
  }
}

function onPageChange(p: number) {
  page.value = p;
  load();
}

onMounted(load);
</script>

<template>
  <div class="dlq">
    <!-- 页头 -->
    <header class="dlq__header">
      <div>
        <h1>DLQ 死信监控</h1>
        <p class="hint">RabbitMQ 死信队列管理 — 仅 owner / admin 可见</p>
      </div>
      <div class="actions">
        <label class="filter-toggle">
          <input v-model="unresolvedOnly" type="checkbox" @change="page = 1; load()" />
          <span>仅未解决</span>
        </label>
        <AppButton variant="ghost" size="sm" @click="load">刷新</AppButton>
      </div>
    </header>

    <!-- 加载态 -->
    <AppLoadingState v-if="loading" text="加载死信消息..." />

    <!-- 错误态 -->
    <AppErrorState v-else-if="error && items.length === 0" :message="error" @retry="load" />

    <template v-else>
      <!-- 统计卡片 -->
      <div class="stats-grid">
        <div class="stat-card stat-card--danger">
          <div class="stat-card__label">未解决</div>
          <div class="stat-card__value">{{ unresolvedCount }}</div>
          <div class="stat-card__hint">条死信待处理</div>
        </div>
        <div class="stat-card stat-card--success">
          <div class="stat-card__label">已解决</div>
          <div class="stat-card__value">{{ resolvedCount }}</div>
          <div class="stat-card__hint">条已标记 resolved</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__label">最新事件</div>
          <div class="stat-card__value stat-card__value--sm">{{ latestTime }}</div>
          <div class="stat-card__hint">最近一条入队时间</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__label">本页合计</div>
          <div class="stat-card__value">{{ items.length }}</div>
          <div class="stat-card__hint">共 {{ total }} 条记录</div>
        </div>
      </div>

      <!-- 错误提示（非致命，数据已加载） -->
      <div v-if="error" class="inline-error">
        <span>{{ error }}</span>
        <button class="inline-error__close" @click="error = ''">×</button>
      </div>

      <!-- 操作栏 -->
      <div class="toolbar">
        <div class="toolbar__left">
          <label class="check-all">
            <input type="checkbox" :checked="allSelected" @change="toggleAll" />
            <span>全选未解决</span>
          </label>
          <span v-if="selectedIds.size > 0" class="selected-count">
            已选 {{ selectedIds.size }} 条
          </span>
        </div>
        <div class="toolbar__right">
          <AppButton
            variant="secondary"
            size="sm"
            :disabled="selectedIds.size === 0"
            @click="batchCleanup"
          >
            批量清理
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="cleanupAllResolved">
            清理全部已解决
          </AppButton>
        </div>
      </div>

      <!-- 表格 -->
      <div v-if="items.length > 0" class="table-wrapper">
        <table class="dlq-table">
          <thead>
            <tr>
              <th class="col-check"></th>
              <th class="col-id">ID</th>
              <th class="col-event">Event ID</th>
              <th class="col-queue">队列</th>
              <th class="col-routing">路由键</th>
              <th class="col-error">错误原因</th>
              <th class="col-time">创建时间</th>
              <th class="col-status">状态</th>
              <th class="col-action">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id" :class="{ resolved: !!item.resolved_at }">
              <td class="col-check">
                <input
                  v-if="!item.resolved_at"
                  type="checkbox"
                  :checked="selectedIds.has(item.id)"
                  @change="toggleSelect(item.id)"
                />
              </td>
              <td class="col-id">{{ item.id }}</td>
              <td class="col-event">
                <code>{{ item.event_id }}</code>
              </td>
              <td class="col-queue">
                <span class="mono">{{ item.queue }}</span>
              </td>
              <td class="col-routing">
                <span class="mono">{{ item.routing_key }}</span>
              </td>
              <td class="col-error">
                <span class="error-text" :title="item.error_reason">
                  {{ item.error_reason }}
                </span>
              </td>
              <td class="col-time">{{ formatRelativeTime(item.created_at) }}</td>
              <td class="col-status">
                <AppBadge :variant="statusVariant(item)" dot></AppBadge>
                <span class="status-label">{{ statusLabel(item) }}</span>
              </td>
              <td class="col-action">
                <div v-if="!item.resolved_at" class="action-btns">
                  <AppButton
                    variant="secondary"
                    size="sm"
                    :loading="retryingIds.has(item.id)"
                    :disabled="retryingIds.has(item.id)"
                    @click="retryItem(item)"
                  >
                    重试
                  </AppButton>
                  <AppButton
                    variant="ghost"
                    size="sm"
                    :disabled="cleaningIds.has(item.id)"
                    @click="cleanItem(item)"
                  >
                    清理
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 空状态 -->
      <AppEmptyState
        v-if="!loading && items.length === 0"
        icon="🛡️"
        title="没有死信消息"
        :description="unresolvedOnly ? '暂无未解决的死信事件' : '死信队列为空，所有消息均正常处理'"
      />

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="pagination">
        <button class="page-btn" :disabled="page <= 1" @click="onPageChange(page - 1)">
          上一页
        </button>
        <span class="page-info">第 {{ page }} / {{ totalPages }} 页</span>
        <button class="page-btn" :disabled="page >= totalPages" @click="onPageChange(page + 1)">
          下一页
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.dlq__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.dlq__header h1 {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 4px;
  color: var(--txt-primary);
}

.hint {
  color: var(--txt-tertiary);
  font-size: var(--text-13);
  margin: 0;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-surface-1);
  color: var(--txt-secondary);
  font-size: var(--text-13);
  cursor: pointer;
  user-select: none;
}

.filter-toggle:hover {
  background: var(--bg-surface-2);
}

.filter-toggle input[type="checkbox"] {
  accent-color: var(--brand-500);
  margin: 0;
}

/* --- 统计卡片 --- */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.stat-card {
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
}

.stat-card--danger {
  border-left: 3px solid var(--danger-500);
}

.stat-card--success {
  border-left: 3px solid var(--success-500);
}

.stat-card__label {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
  margin-bottom: 6px;
}

.stat-card__value {
  font-size: 28px;
  font-weight: 600;
  color: var(--txt-primary);
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.stat-card__value--sm {
  font-size: 18px;
}

.stat-card__hint {
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  margin-top: 4px;
}

/* --- 错误提示 --- */
.inline-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  margin-bottom: 12px;
  border-radius: var(--radius-sm);
  background: var(--bg-danger-subtle, rgba(220, 47, 47, 0.06));
  border: 1px solid var(--border-danger-subtle);
  color: var(--txt-danger-primary);
  font-size: var(--text-13);
}

.inline-error__close {
  background: none;
  border: none;
  font-size: 18px;
  color: inherit;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

/* --- 工具条 --- */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  margin-bottom: 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-surface-2);
  border: 1px solid var(--border-subtle);
}

.toolbar__left,
.toolbar__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.check-all {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: var(--text-13);
  color: var(--txt-secondary);
  cursor: pointer;
  user-select: none;
}

.check-all input[type="checkbox"] {
  accent-color: var(--brand-500);
}

.selected-count {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-accent-subtle);
  color: var(--txt-accent-primary);
}

/* --- 表格 --- */
.table-wrapper {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
  overflow-x: auto;
  max-height: 600px;
  overflow-y: auto;
}

.dlq-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--text-13);
}

.dlq-table thead th {
  position: sticky;
  top: 0;
  background: var(--bg-surface-2);
  border-bottom: 1px solid var(--border-subtle);
  padding: 10px 12px;
  text-align: left;
  font-weight: 500;
  color: var(--txt-tertiary);
  font-size: var(--text-12);
  white-space: nowrap;
}

.dlq-table tbody td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: middle;
}

.dlq-table tbody tr:hover {
  background: var(--bg-surface-2);
}

.dlq-table tbody tr.resolved {
  opacity: 0.6;
}

.col-check { width: 36px; text-align: center; }
.col-id { width: 60px; color: var(--txt-secondary); font-variant-numeric: tabular-nums; }
.col-event { width: 90px; }
.col-event code {
  font-family: var(--font-mono);
  font-size: var(--text-12);
  background: var(--bg-surface-3);
  padding: 1px 6px;
  border-radius: 3px;
  color: var(--txt-secondary);
}
.col-queue { min-width: 120px; }
.col-routing { min-width: 120px; width: 160px; }
.col-error { min-width: 180px; max-width: 280px; }
.col-time { width: 100px; color: var(--txt-tertiary); }
.col-status { width: 90px; }
.col-action { width: 140px; }

.mono {
  font-family: var(--font-mono);
  font-size: var(--text-12);
  color: var(--txt-secondary);
}

.error-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--txt-danger-secondary);
  font-size: var(--text-12);
}

.status-label {
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  margin-left: 4px;
}

.action-btns {
  display: flex;
  gap: 6px;
}

/* --- 分页 --- */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 16px;
  padding: 12px 0;
}

.page-btn {
  padding: 6px 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-surface-1);
  color: var(--txt-secondary);
  font-size: var(--text-13);
  cursor: pointer;
  font-family: inherit;
}

.page-btn:hover:not(:disabled) {
  background: var(--bg-surface-2);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  font-size: var(--text-13);
  color: var(--txt-tertiary);
  font-variant-numeric: tabular-nums;
}
</style>
