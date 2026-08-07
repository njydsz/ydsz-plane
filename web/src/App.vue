<script setup lang="ts">
import GlobalSearch from '@/components/GlobalSearch.vue'
import AppToast from '@/components/AppToast.vue'
</script>

<template>
  <router-view />
  <GlobalSearch />
  <AppToast />
</template>
