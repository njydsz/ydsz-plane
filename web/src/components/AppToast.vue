<script setup lang="ts">
/**
 * AppToast — 全局消息提示渲染器（挂在 App.vue 根部）。
 * 消费 lib/toast.ts 的消息队列，右上角堆叠展示。
 *
 * 支持 loading 状态（常驻不自动关闭，带 spinner 动画）。
 */
import { toasts, dismiss, type ToastItem } from "@/lib/toast";
import DynamicIcon from "./DynamicIcon.vue";

const icons: Record<ToastItem["type"], string> = {
  success: "Check",
  error: "X",
  info: "Info",
  warning: "AlertTriangle",
  loading: "",
};

function label(type: ToastItem["type"]): string {
  const map: Record<ToastItem["type"], string> = {
    success: "成功",
    error: "错误",
    info: "提示",
    warning: "警告",
    loading: "处理中",
  };
  return map[type] ?? "";
}
</script>

<template>
  <div class="toast-container" aria-live="polite" aria-label="通知">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="toast"
        :class="`toast--${t.type}`"
        role="status"
      >
        <!-- loading 类型使用 spinner -->
        <span v-if="t.type === 'loading'" class="toast__spinner" aria-hidden="true" />
        <!-- 普通类型使用图标 -->
        <span v-else class="toast__icon"><DynamicIcon :name="icons[t.type]" :size="16" /></span>

        <div class="toast__body">
          <span class="toast__label">{{ label(t.type) }}</span>
          <span class="toast__message">{{ t.message }}</span>
        </div>

        <!-- loading 类型不展示关闭按钮（常驻） -->
        <button
          v-if="t.type !== 'loading'"
          class="toast__close"
          aria-label="关闭"
          @click="dismiss(t.id)"
        >
          ×
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 16px;
  right: 16px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  pointer-events: none;
}

.toast {
  pointer-events: auto;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 260px;
  max-width: 380px;
  padding: 10px 12px;
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  box-shadow: var(--shadow-popover, 0 4px 16px rgba(0, 0, 0, 0.1));
}

.toast--success { border-left: 3px solid var(--success-500); }
.toast--error   { border-left: 3px solid var(--danger-500); }
.toast--info    { border-left: 3px solid var(--brand-500); }
.toast--warning { border-left: 3px solid var(--warning-500); }
.toast--loading { border-left: 3px solid var(--brand-default); }

.toast__spinner {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--brand-default);
  border-radius: 50%;
  animation: toast-spin 0.65s linear infinite;
  margin-top: 1px;
}

@keyframes toast-spin {
  to { transform: rotate(360deg); }
}

.toast__icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--text-on-color);
  margin-top: 1px;
}

.toast--success .toast__icon { background: var(--success-500); }
.toast--error   .toast__icon { background: var(--danger-500); }
.toast--info    .toast__icon { background: var(--brand-500); }
.toast--warning .toast__icon { background: var(--warning-500); }

.toast__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.toast__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.toast__message {
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.45;
  word-break: break-word;
}

.toast__close {
  background: none;
  border: none;
  padding: 0 2px;
  font-size: 15px;
  line-height: 1;
  color: var(--text-tertiary);
  cursor: pointer;
  flex-shrink: 0;
}

.toast__close:hover {
  color: var(--text-primary);
}

/* 过渡动画 */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(24px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(24px);
}
</style>
