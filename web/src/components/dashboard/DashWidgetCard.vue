<script setup lang="ts">
/**
 * DashWidgetCard — 仪表盘 Widget 通用外壳。
 * 标题栏（含拖拽手柄）+ 右上角删除按钮（×），slot 渲染具体内容。
 *
 * 拖拽由父级 (.grid-cell) 上的 draggable="true" 驱动；本组件仅
 * 在标题区提供视觉手柄，并拦截其上的点击事件以免触发 widget 内容交互。
 *
 * 编辑模式（editMode）下额外显示右下角 resize 手柄，通过 resizeStart 事件
 * 通知父级进入尺寸调整流程。
 */
defineProps<{
  title: string;
  isSaving?: boolean;
  editMode?: boolean;
  isResizing?: boolean;
}>();

const emit = defineEmits<{
  remove: [];
  resizeStart: [e: PointerEvent];
}>();
</script>

<template>
  <div
    class="dash-card"
    :class="{ 'dash-card--saving': isSaving, 'dash-card--edit': editMode, 'dash-card--resizing': isResizing }"
  >
    <div class="dash-card__header">
      <span
        class="dash-card__handle"
        :class="{ 'dash-card__handle--hidden': !editMode }"
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
    <!-- 编辑模式：右下角缩放手柄 -->
    <span
      v-if="editMode"
      class="dash-card__resize-handle"
      role="button"
      aria-label="调整大小"
      title="拖拽右下角以调整大小"
      @pointerdown.prevent.stop="emit('resizeStart', $event)"
    ></span>
  </div>
</template>

<style scoped>
.dash-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
  transition: opacity 0.15s, box-shadow 0.15s, border-color 0.15s;
}

.dash-card--saving {
  opacity: 0.75;
}

/* 编辑模式：蓝色描边高亮 */
.dash-card--edit {
  border-color: var(--brand-400, #7c9aff);
  box-shadow: 0 0 0 1px var(--brand-400, #7c9aff);
}

.dash-card--resizing {
  opacity: 0.85;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.18);
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

.dash-card__handle--hidden {
  visibility: hidden;
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

/* ---- 编辑模式 resize 手柄 ---- */
.dash-card__resize-handle {
  position: absolute;
  right: 3px;
  bottom: 3px;
  width: 16px;
  height: 16px;
  cursor: nwse-resize;
  border-radius: var(--radius-sm, 4px);
  background-image: linear-gradient(
      135deg,
      transparent 0 40%,
      var(--brand-500, #3f63f1) 40% 55%,
      transparent 55% 100%
    );
  opacity: 0.55;
  transition: opacity 0.15s, transform 0.15s;
  z-index: 3;
}

.dash-card__resize-handle:hover,
.dash-card__resize-handle:active {
  opacity: 1;
  transform: scale(1.15);
}
</style>
