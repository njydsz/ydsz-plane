<script setup lang="ts">
/**
 * DashWidgetCard — 仪表盘 Widget 通用外壳。
 * 标题栏（含拖拽手柄）+ 右上角删除按钮（×），slot 渲染具体内容。
 *
 * 拖拽由父级 (.grid-cell) 上的 draggable="true" 驱动；本组件仅
 * 在标题区提供视觉手柄，并拦截其上的点击事件以免触发 widget 内容交互。
 */
defineProps<{
  title: string;
  isSaving?: boolean;
}>();

const emit = defineEmits<{ remove: [] }>();
</script>

<template>
  <div class="dash-card" :class="{ 'dash-card--saving': isSaving }">
    <div class="dash-card__header">
      <span
        class="dash-card__handle"
        aria-label="拖拽手柄"
        title="拖拽以重新排列"
      >
        <span class="dash-card__handle-grip"></span>
      </span>
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
  transition: opacity 0.15s, box-shadow 0.15s;
}

.dash-card--saving {
  opacity: 0.75;
}

.dash-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  flex-shrink: 0;
}

/* ---- 拖拽手柄 ---- */
.dash-card__handle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  border-radius: var(--radius-sm, 4px);
  cursor: grab;
  color: var(--text-tertiary, #9ca3af);
  transition: color 0.15s, background 0.15s;
}

.dash-card__handle:hover {
  color: var(--text-secondary, #4b5563);
  background: var(--surface-2, #f7f8f9);
}

.dash-card__handle:active {
  cursor: grabbing;
}

.dash-card__handle-grip {
  width: 10px;
  height: 10px;
  background-image: radial-gradient(circle, currentColor 1.2px, transparent 1.2px);
  background-size: 4px 4px;
  background-repeat: repeat;
  opacity: 0.7;
}

.dash-card__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
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
