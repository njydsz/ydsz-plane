<script setup lang="ts">
/**
 * AppToast — 全局消息提示渲染器（挂在 App.vue 根部）。
 * 消费 lib/toast.ts 的消息队列，右上角堆叠展示。
 */
import { toasts, dismiss, type ToastItem } from "@/lib/toast";

const icons: Record<ToastItem["type"], string> = {
  success: "✓",
  error: "✕",
  info: "ℹ",
  warning: "!",
};

function label(type: ToastItem["type"]): string {
  return ({ success: "成功", error: "错误", info: "提示", warning: "警告" } as Record<string, string>)[type] ?? "";
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
        <span class="toast__icon">{{ icons[t.type] }}</span>
        <div class="toast__body">
          <span class="toast__label">{{ label(t.type) }}</span>
          <span class="toast__message">{{ t.message }}</span>
        </div>
        <button class="toast__close" aria-label="关闭" @click="dismiss(t.id)">×</button>
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
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-default, #d1d5db);
  box-shadow: var(--shadow-popover, 0 4px 16px rgba(0, 0, 0, 0.1));
}

.toast--success { border-left: 3px solid var(--success-500, #16a34a); }
.toast--error   { border-left: 3px solid var(--danger-500, #ef4444); }
.toast--info    { border-left: 3px solid var(--brand-500, #3b82f6); }
.toast--warning { border-left: 3px solid var(--warning-500, #f59e0b); }

.toast__icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #fff;
  margin-top: 1px;
}

.toast--success .toast__icon { background: var(--success-500, #16a34a); }
.toast--error   .toast__icon { background: var(--danger-500, #ef4444); }
.toast--info    .toast__icon { background: var(--brand-500, #3b82f6); }
.toast--warning .toast__icon { background: var(--warning-500, #f59e0b); }

.toast__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toast__label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary, #9ca3af);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.toast__message {
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  line-height: 1.45;
  word-break: break-word;
}

.toast__close {
  background: none;
  border: none;
  padding: 0 2px;
  font-size: 15px;
  line-height: 1;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  flex-shrink: 0;
}

.toast__close:hover {
  color: var(--text-primary, #1f2937);
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
