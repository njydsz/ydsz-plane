<script setup lang="ts">
/**
 * 需求列表页 — 仅展示 requirement 类型的需求/任务/缺陷。
 * 复用 IssueListView 的表格能力，固定按 type_code = 'requirement' 过滤。
 */
import IssueListView from "./IssueListView.vue";
</script>

<template>
  <IssueListView type-code="requirement" />
</template>
