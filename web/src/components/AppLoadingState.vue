<script setup lang="ts">
/**
 * AppLoadingState — 加载中占位。
 *
 * Props:
 *   text: 加载提示文案
 */
withDefaults(defineProps<{
  text?: string;
}>(), {
  text: "加载中...",
});
</script>

<template>
  <div class="app-loading">
    <span class="app-loading__spinner" />
    <p class="app-loading__text">{{ text }}</p>
  </div>
</template>

<style scoped>
.app-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 12px;
}

.app-loading__spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-default);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.app-loading__text {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
