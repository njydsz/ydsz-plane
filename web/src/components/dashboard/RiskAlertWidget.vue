<script setup lang="ts">
/**
 * RiskAlertWidget - 风险告警卡片列表。
 * severity 颜色映射：critical -> red, high -> orange, medium -> yellow, low -> blue, info -> gray。
 * 右上角 resolve 按钮。
 */
import { computed } from "vue";
import type { RiskAlert } from "@/api/services/dashboard";
import { formatRelativeTime } from "@/lib/formatTime";

const props = defineProps<{
  /** 告警列表 */
  data?: RiskAlert[];
  /** 兼容外部传入 alerts 形式 */
  alerts?: RiskAlert[];
}>();

const emit = defineEmits<{
  resolve: [alertId: number];
}>();

const alertList = computed(() => props.data ?? props.alerts ?? []);

function severityClass(severity: RiskAlert["severity"]): string {
  return `alert-card--${severity}`;
}

function severityLabel(severity: RiskAlert["severity"]): string {
  const map: Record<string, string> = {
    critical: "紧急",
    high: "高",
    medium: "中",
    low: "低",
    info: "提示",
  };
  return map[severity] ?? severity;
}
</script>

<template>
  <div class="risk-alerts">
    <div v-if="alertList.length" class="risk-alerts__list">
      <div
        v-for="alert in alertList"
        :key="alert.id"
        class="alert-card"
        :class="[
          severityClass(alert.severity),
          { 'alert-card--resolved': alert.is_resolved },
        ]"
      >
        <div class="alert-card__header">
          <span
            class="alert-card__severity"
            :class="`severity-badge--${alert.severity}`"
          >
            {{ severityLabel(alert.severity) }}
          </span>
          <span class="alert-card__title">{{ alert.title }}</span>
          <button
            v-if="!alert.is_resolved"
            class="alert-card__resolve"
            type="button"
            title="标记为已处理"
            @click="emit('resolve', alert.id)"
          >
            解决
          </button>
          <span v-else class="alert-card__resolved-tag">已处理</span>
        </div>
        <p class="alert-card__desc">{{ alert.description }}</p>
        <span class="alert-card__time">{{ formatRelativeTime(alert.created_at) }}</span>
      </div>
    </div>
    <div v-else class="risk-alerts__empty">暂无告警</div>
  </div>
</template>

<style scoped>
.risk-alerts {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.risk-alerts__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-card {
  padding: 12px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border-subtle, #e5e7eb);
  background: var(--surface-1, #fff);
  transition: opacity 0.2s;
}

.alert-card--resolved {
  opacity: 0.55;
}

.alert-card--critical {
  border-left: 3px solid var(--danger-500, #dc2f2f);
}

.alert-card--high {
  border-left: 3px solid var(--warning-500, #f59e0b);
}

.alert-card--medium {
  border-left: 3px solid var(--priority-medium, #f7c948);
}

.alert-card--low {
  border-left: 3px solid var(--info-500, #17bee9);
}

.alert-card--info {
  border-left: 3px solid var(--text-tertiary, #9ca3af);
}

.alert-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.alert-card__severity {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 3px;
}

.severity-badge--critical {
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}

.severity-badge--high {
  background: var(--warning-50, #fffbeb);
  color: var(--warning-600, #d97706);
}

.severity-badge--medium {
  background: #fefce8;
  color: #ca8a04;
}

.severity-badge--low {
  background: #ecfeff;
  color: #0891b2;
}

.severity-badge--info {
  background: var(--surface-3, #eef0f2);
  color: var(--text-tertiary, #9ca3af);
}

.alert-card__title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  flex: 1;
}

.alert-card__resolve {
  flex-shrink: 0;
  background: none;
  border: 1px solid var(--border-default, #dfe2e6);
  border-radius: var(--radius-sm, 6px);
  padding: 2px 10px;
  font-size: 11px;
  color: var(--text-secondary, #4b5563);
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s, border-color 0.15s, color 0.15s;
}

.alert-card__resolve:hover {
  background: var(--success-50, #f0fdf4);
  border-color: var(--success-500, #0fc27b);
  color: var(--success-600, #15803d);
}

.alert-card__resolved-tag {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.alert-card__desc {
  margin: 4px 0;
  font-size: 12px;
  color: var(--text-secondary, #4b5563);
  line-height: 1.5;
}

.alert-card__time {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.risk-alerts__empty {
  padding: 24px 0;
  text-align: center;
  color: var(--text-tertiary, #9ca3af);
  font-size: 13px;
}
</style>
