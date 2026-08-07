<script setup lang="ts">
/**
 * 首页 — 简单欢迎页（预留为工作台入口）。
 */

import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
</script>

<template>
  <div class="home">
    <h1>工作台</h1>
    <p class="hello">你好，{{ auth.user?.display_name }}。个人工作台将在 Sprint 8 交付（见开发计划）。</p>
    <div class="placeholder-grid">
      <div class="card">我的待办<span class="tag">S8</span></div>
      <div class="card">今日日程<span class="tag">S8</span></div>
      <div class="card">迭代概览<span class="tag">S8</span></div>
      <div class="card">关注动态<span class="tag">S9</span></div>
    </div>
  </div>
</template>

<style scoped>
.home h1 {
  margin: 0 0 8px;
  font-size: 20px;
}

.hello {
  color: var(--text-secondary);
  margin: 0 0 24px;
}

.placeholder-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.card {
  padding: 20px;
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tag {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 999px;
  background: var(--brand-50);
  color: var(--brand-600);
}
</style>
