/** 应用根组件：挂载路由视图，并全局注入命令面板与消息提示。 */
<script setup lang="ts">
import CommandPalette from '@/components/CommandPalette.vue'
import AppToast from '@/components/AppToast.vue'
</script>

<template>
  <router-view />
  <CommandPalette />
  <AppToast />
</template>
