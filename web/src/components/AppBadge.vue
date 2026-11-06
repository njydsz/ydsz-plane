<script setup lang="ts">
/**
 * AppBadge — 轻量标签/徽章组件。
 *
 * Props:
 *   variant: 'default' | 'brand' | 'success' | 'warning' | 'danger' | 'info'
 *   dot:     仅显示圆点
 */
withDefaults(defineProps<{
  variant?: "default" | "brand" | "success" | "warning" | "danger" | "info";
  dot?: boolean;
}>(), {
  variant: "default",
});
</script>

<template>
  <span class="app-badge" :class="[`app-badge--${variant}`, { dot }]">
    <slot />
  </span>
</template>

<style scoped>
.app-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
}

.app-badge.dot {
  width: 8px;
  height: 8px;
  padding: 0;
  border-radius: 50%;
}

.app-badge--default { background: var(--surface-3); color: var(--text-secondary); }
.app-badge--brand   { background: var(--brand-50); color: var(--brand-600); }
.app-badge--success { background: var(--success-50); color: var(--success-500); }
.app-badge--warning { background: var(--warning-50); color: var(--warning-500); }
.app-badge--danger  { background: var(--danger-50); color: var(--danger-500); }
.app-badge--info    { background: var(--info-50); color: var(--info-500); }
</style>
