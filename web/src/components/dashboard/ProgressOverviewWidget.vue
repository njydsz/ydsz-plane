<script setup lang="ts">
/**
 * ProgressOverviewWidget — 4 个统计卡片（总数/已完成/进行中/完成率）。
 * 底部展示 overdue/blocked/active 警示徽章。
 */
import type { ProgressOverviewData } from "@/api/services/dashboard";

defineProps<{
  data?: ProgressOverviewData;
}>();

function formatRate(n: number): string {
  return (n * 100).toFixed(1) + "%";
}
</script>

<template>
  <div class="progress-overview">
    <div class="progress-overview__grid">
      <div class="stat-card stat-card--total">
        <span class="stat-card__label">总数</span>
        <span class="stat-card__value">{{ data?.total_issues ?? 0 }}</span>
      </div>
      <div class="stat-card stat-card--done">
        <span class="stat-card__label">已完成</span>
        <span class="stat-card__value">{{ data?.done_issues ?? 0 }}</span>
      </div>
      <div class="stat-card stat-card--in-progress">
        <span class="stat-card__label">进行中</span>
        <span class="stat-card__value">{{ data?.in_progress ?? 0 }}</span>
      </div>
      <div class="stat-card stat-card--rate">
        <span class="stat-card__label">完成率</span>
        <span class="stat-card__value">{{ data ? formatRate(data.completion_rate) : "0%" }}</span>
      </div>
    </div>

    <div v-if="data" class="progress-overview__badges">
      <span v-if="data.overdue_issues > 0" class="badge badge--danger">
        逾期 {{ data.overdue_issues }}
      </span>
      <span v-if="data.blocked_issues > 0" class="badge badge--warning">
        阻塞 {{ data.blocked_issues }}
      </span>
      <span v-if="data.active_sprints > 0" class="badge badge--info">
        活跃迭代 {{ data.active_sprints }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.progress-overview {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.progress-overview__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-2, #f7f8f9);
}

.stat-card__label {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.stat-card__value {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.stat-card--done .stat-card__value {
  color: var(--success-500, #0fc27b);
}

.stat-card--in-progress .stat-card__value {
  color: var(--brand-500, #3f63f1);
}

.progress-overview__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

.badge--danger {
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}

.badge--warning {
  background: var(--warning-50, #fffbeb);
  color: var(--warning-600, #d97706);
}

.badge--info {
  background: var(--brand-50, #eef2fe);
  color: var(--brand-600, #2f4fd0);
}
</style>
