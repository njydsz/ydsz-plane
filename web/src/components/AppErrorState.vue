<script setup lang="ts">
/**
 * AppErrorState — 错误状态占位。
 *
 * Props:
 *   message:  错误信息
 *   retry?:   重试文案（设为空则不显示重试按钮）
 *
 * Events:
 *   @retry: 点击重试
 */
withDefaults(defineProps<{
  message: string;
  retry?: string;
}>(), {
  retry: "重试",
});

const emit = defineEmits<{
  retry: [];
}>();
</script>

<template>
  <div class="app-error">
    <span class="app-error__icon">!</span>
    <p class="app-error__text">{{ message }}</p>
    <button v-if="retry" class="app-error__btn" @click="emit('retry')">
      {{ retry }}
    </button>
    <slot />
  </div>
</template>

<style scoped>
.app-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
  gap: 10px;
}

.app-error__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(220, 47, 47, 0.1);
  color: var(--danger-500);
  font-size: 20px;
  font-weight: 700;
}

.app-error__text {
  margin: 0;
  font-size: 14px;
  color: var(--danger-500);
  max-width: 360px;
}

.app-error__btn {
  margin-top: 4px;
  padding: 6px 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
}

.app-error__btn:hover {
  background: var(--surface-2);
}
</style>
