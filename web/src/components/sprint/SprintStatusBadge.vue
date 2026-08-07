<script setup lang="ts">
/**
 * SprintStatusBadge — 迭代状态徽章（统一状态 → 颜色/文案映射）。
 *
 * 集中管理 Sprint 状态展示约定，避免各视图重复实现：
 *   planned   → default「未开始」
 *   active    → success「进行中」
 *   completed → brand「已完成」
 */
import { computed } from "vue";
import AppBadge from "@/components/AppBadge.vue";
import type { SprintStatus } from "@/api/services/sprint";

const props = withDefaults(
  defineProps<{
    status: SprintStatus;
    /** 仅显示状态圆点（无文字） */
    dot?: boolean;
    /** 显示状态圆点 + 文案 */
    withDot?: boolean;
  }>(),
  { dot: false, withDot: false },
);

const config = computed<{ label: string; variant: "default" | "success" | "brand" }>(() => {
  switch (props.status) {
    case "active":
      return { label: "进行中", variant: "success" };
    case "completed":
      return { label: "已完成", variant: "brand" };
    default:
      return { label: "未开始", variant: "default" };
  }
});

const dotColor = computed(() => {
  switch (props.status) {
    case "active":
      return "var(--success-500)";
    case "completed":
      return "var(--brand-500)";
    default:
      return "var(--text-tertiary)";
  }
});
</script>

<template>
  <!-- 仅圆点 -->
  <span v-if="dot" class="sprint-status-dot" :style="{ background: dotColor }" :title="config.label" />

  <!-- 圆点 + 文案 -->
  <span v-else-if="withDot" class="sprint-status-inline">
    <span class="sprint-status-dot" :style="{ background: dotColor }" />
    <AppBadge :variant="config.variant">{{ config.label }}</AppBadge>
  </span>

  <!-- 仅徽章 -->
  <AppBadge v-else :variant="config.variant">{{ config.label }}</AppBadge>
</template>

<style scoped>
.sprint-status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sprint-status-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
</style>
