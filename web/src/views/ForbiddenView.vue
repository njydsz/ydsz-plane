<script setup lang="ts">
/**
 * 403 无权限页面 — 用户试图访问缺少工作空间权限的路由时展示。
 */
import { useRoute } from "vue-router";

const route = useRoute();
const from = (route.query.from as string) ?? "/";
</script>

<template>
  <div class="forbidden">
    <div class="forbidden__code">403</div>
    <h2>没有权限</h2>
    <p class="forbidden__desc">您没有访问该页面的工作空间权限。如需访问，请联系空间管理员或系统管理员。</p>
    <router-link :to="from" class="forbidden__back">返回上一页</router-link>
  </div>
</template>

<style scoped>
.forbidden {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-secondary);
}
.forbidden__code {
  font-size: 64px;
  font-weight: 700;
  color: var(--warning-500);
  line-height: 1;
}
.forbidden h2 {
  margin: 0;
  font-size: 22px;
  color: var(--text-primary);
}
.forbidden__desc {
  max-width: 360px;
  text-align: center;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
}
.forbidden__back {
  margin-top: 8px;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  text-decoration: none;
  font-size: 14px;
}
.forbidden__back:hover {
  background: var(--brand-600);
}
</style>
