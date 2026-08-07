<script setup lang="ts">
/**
 * AppAvatar — 用户头像组件。
 *
 * Props:
 *   name:  显示名称（取首字符作占位）
 *   src?:  图片 URL，存在时显示图片头像
 *   size?: 'sm' | 'md' | 'lg'
 */
withDefaults(defineProps<{
  name: string;
  src?: string;
  size?: "sm" | "md" | "lg";
}>(), {
  size: "md",
});
</script>

<template>
  <span
    class="app-avatar"
    :class="`app-avatar--${size}`"
    :aria-label="name"
    role="img"
  >
    <img v-if="src" :src="src" :alt="name" class="app-avatar__img" />
    <span v-else class="app-avatar__fallback">{{ name.charAt(0).toUpperCase() }}</span>
  </span>
</template>

<style scoped>
.app-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 600;
  user-select: none;
}

.app-avatar--sm { width: 24px; height: 24px; font-size: 11px; }
.app-avatar--md { width: 32px; height: 32px; font-size: 13px; }
.app-avatar--lg { width: 40px; height: 40px; font-size: 16px; }

.app-avatar__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-avatar__fallback {
  line-height: 1;
}
</style>
