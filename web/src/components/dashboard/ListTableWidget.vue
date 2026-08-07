<script setup lang="ts">
/**
 * ListTableWidget — 通用表格 widget，用于逾期/阻塞工作项。
 * 通过 config.kind 区分数据模式。
 */
import type { OverdueItem, BlockedItem } from "@/api/services/dashboard";

const props = defineProps<{
  data?: { total: number; items: (OverdueItem | BlockedItem)[] };
  config?: Record<string, any>;
}>();

function isOverdue(): boolean {
  const items = props.data?.items ?? [];
  return items.length > 0 && "overdue_days" in items[0];
}

function priorityColor(priority: string): string {
  const map: Record<string, string> = {
    urgent: "var(--priority-urgent)",
    high: "var(--priority-high)",
    medium: "var(--priority-medium)",
    low: "var(--priority-low)",
  };
  return map[priority] ?? "var(--text-tertiary)";
}
</script>

<template>
  <div class="list-table">
    <p class="list-table__summary">
      共 <strong>{{ data?.total ?? 0 }}</strong> 条记录
    </p>
    <div v-if="data?.items?.length" class="list-table__wrap">
      <table class="list-table__inner">
        <thead>
          <tr>
            <th class="col-id">ID</th>
            <th class="col-title">标题</th>
            <th v-if="isOverdue()" class="col-days">逾期天数</th>
            <th v-else class="col-count">阻塞数</th>
            <th class="col-extra">负责人 / 阻塞项</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data.items" :key="item.id">
            <td class="col-id">
              <span
                v-if="'priority' in item"
                class="col-id__dot"
                :style="{ background: priorityColor((item as OverdueItem).priority) }"
              />
              <span v-else class="col-id__dot" style="background: var(--text-tertiary)" />
              {{ item.identifier }}
            </td>
            <td class="col-title">{{ item.title }}</td>
            <td v-if="isOverdue()" class="col-days">
              <span class="overdue-tag">
                {{ (item as OverdueItem).overdue_days }}天
              </span>
            </td>
            <td v-else class="col-count">
              {{ (item as BlockedItem).blocked_count }}
            </td>
            <td class="col-extra">
              {{ isOverdue() ? (item as OverdueItem).assignee : (item as BlockedItem).blocker_names }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="list-table__empty">暂无记录</div>
  </div>
</template>

<style scoped>
.list-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.list-table__summary {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.list-table__summary strong {
  color: var(--text-primary, #1f2937);
}

.list-table__wrap {
  overflow-x: auto;
}

.list-table__inner {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.list-table__inner th {
  text-align: left;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  color: var(--text-tertiary, #9ca3af);
  font-weight: 500;
}

.list-table__inner td {
  padding: 8px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  color: var(--text-primary, #1f2937);
}

.list-table__inner tr:last-child td {
  border-bottom: none;
}

.col-id {
  white-space: nowrap;
  font-family: var(--font-mono, monospace);
  width: 80px;
}

.col-id__dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

.col-title {
  min-width: 120px;
}

.col-days,
.col-count,
.col-extra {
  white-space: nowrap;
  width: 70px;
}

.overdue-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
  font-size: 11px;
  font-weight: 500;
}

.list-table__empty {
  padding: 24px 0;
  text-align: center;
  color: var(--text-tertiary, #9ca3af);
  font-size: 13px;
}
</style>
