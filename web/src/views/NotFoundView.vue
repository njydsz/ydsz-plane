<script setup lang="ts">
/**
 * 404 页面 — 未知路由兜底展示。
 */
</script>

<template>
  <div class="not-found">
    <h1>404</h1>
    <p>页面不存在</p>
    <router-link to="/">返回工作台</router-link>
  </div>
</template>

<style scoped>
.not-found {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-secondary);
}

.not-found h1 {
  font-size: 48px;
  margin: 0;
  color: var(--text-tertiary);
}
</style>
