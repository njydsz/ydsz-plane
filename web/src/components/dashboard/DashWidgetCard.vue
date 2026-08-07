<script setup lang="ts">
/**
 * DashWidgetCard — 仪表盘 Widget 通用外壳。
 * 标题栏 + 右上角删除按钮（×），slot 渲染具体内容。
 */
defineProps<{ title: string }>();

const emit = defineEmits<{ remove: [] }>();
</script>

<template>
  <div class="dash-card">
    <div class="dash-card__header">
      <span class="dash-card__title">{{ title }}</span>
      <button
        class="dash-card__remove"
        type="button"
        aria-label="删除 widget"
        title="删除 widget"
        @click="emit('remove')"
      >
        &times;
      </button>
    </div>
    <div class="dash-card__body">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.dash-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}

.dash-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  flex-shrink: 0;
}

.dash-card__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dash-card__remove {
  flex-shrink: 0;
  background: none;
  border: none;
  font-size: 18px;
  line-height: 1;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  padding: 0;
  width: 22px;
  height: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm, 6px);
  transition: color 0.15s, background 0.15s;
}

.dash-card__remove:hover {
  color: var(--danger-500, #dc2f2f);
  background: var(--danger-50, #fef2f2);
}

.dash-card__body {
  flex: 1;
  min-height: 0;
  padding: 12px 14px;
  overflow: auto;
}
</style>
