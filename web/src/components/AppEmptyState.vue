<script setup lang="ts">
/**
 * AppEmptyState — 空状态占位。
 *
 * Props:
 *   title:       主文案
 *   description: 副文案
 *   icon?:       图标 (emoji 或字符)
 */
withDefaults(defineProps<{
  title?: string;
  description?: string;
  icon?: string;
}>(), {
  title: "暂无数据",
});
</script>

<template>
  <div class="app-empty">
    <span v-if="icon" class="app-empty__icon">{{ icon }}</span>
    <p class="app-empty__title">{{ title }}</p>
    <p v-if="description" class="app-empty__desc">{{ description }}</p>
    <slot />
  </div>
</template>

<style scoped>
.app-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.app-empty__icon {
  font-size: 36px;
  margin-bottom: 12px;
}

.app-empty__title {
  margin: 0;
  font-size: 14px;
  color: var(--text-secondary);
  font-weight: 500;
}

.app-empty__desc {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
  max-width: 300px;
}
</style>
