<script setup lang="ts">
/**
 * 缺陷列表页 — 仅展示 defect 类型的需求/任务/缺陷。
 * 复用 IssueListView 的表格能力，固定按 type_code = 'defect' 过滤。
 */
import IssueListView from "./IssueListView.vue";
</script>

<template>
  <IssueListView type-code="defect" />
</template>
