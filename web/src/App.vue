<!--
  应用根组件：挂载路由视图，并全局注入搜索面板与消息提示。
  这两个全局 UI 通过 Teleport/固定定位浮层呈现，与具体页面解耦，
  因此在应用根部一次性引入。
-->
<script setup lang="ts">
import GlobalSearch from '@/components/GlobalSearch.vue'
import AppToast from '@/components/AppToast.vue'
</script>

<template>
  <router-view />
  <GlobalSearch />
  <AppToast />
</template>
