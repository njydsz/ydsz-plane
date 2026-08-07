<script setup lang="ts">
/**
 * SprintProgressBar — 迭代进度条（故事点完成率）。
 *
 * Props:
 *   donePoints   已完成故事点
 *   totalPoints  承诺总故事点
 *   showText     是否显示「done/total pt」文案（默认 true）
 *
 * 进度条在 total=0 时显示 0%，超 100% 时截断为 100%。
 */
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    donePoints: number;
    totalPoints: number;
    showText?: boolean;
  }>(),
  { showText: true },
);

const percent = computed(() => {
  if (!props.totalPoints || props.totalPoints <= 0) return 0;
  return Math.min(100, Math.round((props.donePoints / props.totalPoints) * 100));
});

const isOver = computed(() => props.donePoints > props.totalPoints && props.totalPoints > 0);
</script>

<template>
  <div class="sprint-progress">
    <div v-if="showText" class="sprint-progress__text">
      <span>{{ donePoints }}/{{ totalPoints }} 故事点</span>
      <span class="pct" :class="{ over: isOver }">{{ percent }}%</span>
    </div>
    <div
      class="sprint-progress__bar"
      role="progressbar"
      :aria-valuenow="percent"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="`迭代进度 ${percent}%`"
    >
      <div
        class="sprint-progress__fill"
        :class="{ over: isOver }"
        :style="{ width: percent + '%' }"
      />
    </div>
  </div>
</template>

<style scoped>
.sprint-progress__text {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  margin-bottom: 4px;
}
.sprint-progress__text .pct { color: var(--text-secondary); }
.sprint-progress__text .pct.over { color: var(--danger-500); }

.sprint-progress__bar {
  height: 4px;
  background: var(--surface-3);
  border-radius: 2px;
  overflow: hidden;
}
.sprint-progress__fill {
  height: 100%;
  background: var(--success-500);
  border-radius: 2px;
  transition: width 0.3s;
}
.sprint-progress__fill.over { background: var(--danger-500); }
</style>
