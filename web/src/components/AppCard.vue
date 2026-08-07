<script setup lang="ts">
/**
 * AppCard — 通用卡片容器。
 *
 * Props:
 *   hoverable: 悬停效果
 *   padding:   'none' | 'sm' | 'md' | 'lg'
 *   shadow:    是否显示阴影
 */
withDefaults(defineProps<{
  hoverable?: boolean;
  padding?: "none" | "sm" | "md" | "lg";
  shadow?: boolean;
}>(), {
  padding: "md",
  shadow: false,
});
</script>

<template>
  <div
    class="app-card"
    :class="[`app-card--${padding}`, { hoverable, shadow }]"
  >
    <slot />
  </div>
</template>

<style scoped>
.app-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--surface-1);
  transition: box-shadow 0.2s, border-color 0.2s;
}

.app-card--none  { padding: 0; }
.app-card--sm   { padding: 12px; }
.app-card--md   { padding: 20px; }
.app-card--lg   { padding: 28px; }

.app-card.hoverable:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow-card);
}

.app-card.shadow {
  box-shadow: var(--shadow-card);
}
</style>
