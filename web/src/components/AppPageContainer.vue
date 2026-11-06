<script setup lang="ts">
/**
 * AppPageContainer — 统一页面容器。
 *
 * 统一页面最大宽度与内边距，消除各视图 14 种不同 max-width 的混乱。
 *
 * Props:
 *   max: 最大宽度 (CSS 值，默认 1280px)
 *   pad: 是否应用 --padding-page 内边距 (默认 true)
 */
withDefaults(defineProps<{
  max?: string;
  pad?: boolean;
}>(), {
  max: "1280px",
  pad: true,
});
</script>

<template>
  <div
    class="app-page-container"
    :style="{ maxWidth: max }"
    :class="{ 'app-page-container--pad': pad }"
  >
    <slot />
  </div>
</template>

<style scoped>
.app-page-container {
  margin: 0 auto;
  width: 100%;
}

.app-page-container--pad {
  padding-left: var(--padding-page-x);
  padding-right: var(--padding-page-x);
  padding-top: var(--padding-page-y);
  padding-bottom: var(--padding-page-y);
}
</style>
