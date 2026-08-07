<script setup lang="ts">
/**
 * AppButton — 统一按钮组件。
 *
 * Props:
 *   variant: 'primary' | 'secondary' | 'danger' | 'ghost'
 *   size:    'sm' | 'md' | 'lg'
 *   loading: 显示加载态
 *   disabled: 禁用
 *   block:   100% 宽度
 */
withDefaults(defineProps<{
  variant?: "primary" | "secondary" | "danger" | "ghost";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
  disabled?: boolean;
  block?: boolean;
  type?: "button" | "submit";
}>(), {
  variant: "primary",
  size: "md",
  type: "button",
});
</script>

<template>
  <button
    :type="type"
    class="app-btn"
    :class="[`app-btn--${variant}`, `app-btn--${size}`, { block, loading }]"
    :disabled="disabled || loading"
  >
    <span v-if="loading" class="app-btn__spinner" />
    <slot />
  </button>
</template>

<style scoped>
.app-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: var(--radius-sm);
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: background 0.15s, border-color 0.15s, opacity 0.15s;
  white-space: nowrap;
  font-family: inherit;
  line-height: 1;
}

/* --- variants --- */
.app-btn--primary {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-color: var(--brand-500);
}
.app-btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.app-btn--secondary {
  background: var(--surface-1);
  color: var(--text-secondary);
  border-color: var(--border-default);
}
.app-btn--secondary:hover:not(:disabled) {
  background: var(--surface-2);
}

.app-btn--danger {
  background: var(--danger-500);
  color: var(--text-on-brand);
  border-color: var(--danger-500);
}
.app-btn--danger:hover:not(:disabled) {
  opacity: 0.9;
}

.app-btn--ghost {
  background: transparent;
  color: var(--text-secondary);
}
.app-btn--ghost:hover:not(:disabled) {
  background: var(--surface-3);
}

/* --- sizes --- */
.app-btn--sm { height: 28px; padding: 0 10px; font-size: 12px; }
.app-btn--md { height: 36px; padding: 0 16px; font-size: 13px; }
.app-btn--lg { height: 44px; padding: 0 24px; font-size: 15px; }

/* --- states --- */
.app-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.app-btn.block {
  width: 100%;
}

.app-btn__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
