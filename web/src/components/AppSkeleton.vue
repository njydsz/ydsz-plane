<script setup lang="ts">
/**
 * AppSkeleton — 骨架屏加载占位组件（对标 Plane 的 skeleton 体系）。
 *
 * 变体:
 *   card     — 卡片骨架（看板卡片）
 *   row      — 表格行骨架（列表行）
 *   table    — 完整表格骨架（表头 + 多行）
 *   board    — 看板列骨架（多列 + 卡片）
 *   text     — 纯文本行骨架（标题/段落）
 *   dashboard — 仪表盘块骨架
 *
 * Props:
 *   variant: 变体类型
 *   rows:    row/table 变体行数
 *   cols:    board 变体列数
 */
withDefaults(defineProps<{
  variant?: "card" | "row" | "table" | "board" | "text" | "dashboard";
  rows?: number;
  cols?: number;
}>(), {
  variant: "table",
  rows: 6,
  cols: 4,
});
</script>

<template>
  <div class="skeleton" :class="`skeleton--${variant}`" role="status" aria-label="加载中">
    <!-- 文本行 -->
    <template v-if="variant === 'text'">
      <div v-for="n in rows" :key="n" class="sk-text" :style="{ width: (85 - n * 5) + '%' }" />
    </template>

    <!-- 单卡片 -->
    <div v-else-if="variant === 'card'" class="sk-card">
      <div class="sk-card__line sk-card__line--sm" />
      <div class="sk-card__line" />
      <div class="sk-card__line sk-card__line--md" />
      <div class="sk-card__footer">
        <div class="sk-avatar" />
        <div class="sk-avatar" />
      </div>
    </div>

    <!-- 单行 -->
    <div v-else-if="variant === 'row'" class="sk-row">
      <div class="sk-row__cell" style="width: 10%" />
      <div class="sk-row__cell" style="width: 40%" />
      <div class="sk-row__cell" style="width: 12%" />
      <div class="sk-row__cell" style="width: 10%" />
      <div class="sk-row__cell" style="width: 14%" />
      <div class="sk-row__cell" style="width: 14%" />
    </div>

    <!-- 表格 -->
    <div v-else-if="variant === 'table'" class="sk-table">
      <div class="sk-table__head">
        <div v-for="n in 6" :key="n" class="sk-th" :style="{ width: n === 2 ? '38%' : '10%' }" />
      </div>
      <div v-for="n in rows" :key="n" class="sk-row">
        <div class="sk-row__cell" style="width: 10%" />
        <div class="sk-row__cell" style="width: 38%" />
        <div class="sk-row__cell" style="width: 12%" />
        <div class="sk-row__cell" style="width: 10%" />
        <div class="sk-row__cell" style="width: 14%" />
        <div class="sk-row__cell" style="width: 14%" />
      </div>
    </div>

    <!-- 看板列 -->
    <div v-else-if="variant === 'board'" class="sk-board">
      <div v-for="n in cols" :key="n" class="sk-column">
        <div class="sk-column__head">
          <div class="sk-pill" />
          <div class="sk-pill sk-pill--sm" />
        </div>
        <div v-for="m in 3" :key="m" class="sk-card">
          <div class="sk-card__line sk-card__line--sm" />
          <div class="sk-card__line" />
          <div class="sk-card__line sk-card__line--md" />
        </div>
      </div>
    </div>

    <!-- 仪表盘块 -->
    <div v-else-if="variant === 'dashboard'" class="sk-dash">
      <div v-for="n in Math.min(rows, 4)" :key="n" class="sk-dash__block">
        <div class="sk-dash__title" />
        <div class="sk-dash__metric" />
        <div class="sk-dash__bars">
          <div v-for="m in 6" :key="m" class="sk-dash__bar" :style="{ height: (30 + ((n * 17 + m * 13) % 50)) + '%' }" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.skeleton {
  --sk-base: var(--bg-layer-1);
  --sk-highlight: var(--bg-layer-1-hover);
  position: relative;
  overflow: hidden;
}

/* 基础 shimmer 动画 */
.sk-card,
.sk-row,
.sk-th,
.sk-text,
.sk-pill,
.sk-avatar,
.sk-dash__title,
.sk-dash__metric,
.sk-dash__bar {
  position: relative;
  background: var(--sk-base);
  border-radius: 4px;
}

.skeleton::after {
  content: "";
  position: absolute;
  inset: 0;
  transform: translateX(-100%);
  background: linear-gradient(90deg, transparent, var(--sk-highlight), transparent);
  animation: sk-shimmer 1.4s infinite;
}

@keyframes sk-shimmer {
  100% { transform: translateX(100%); }
}

/* 文本 */
.sk-text {
  height: 12px;
  margin-bottom: 10px;
}

/* 卡片 */
.sk-card {
  padding: 12px;
  background: var(--bg-surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  margin-bottom: 8px;
}

.sk-card__line {
  height: 12px;
  width: 90%;
  margin-bottom: 8px;
}

.sk-card__line--sm {
  width: 30%;
  height: 8px;
}

.sk-card__line--md {
  width: 60%;
}

.sk-card__footer {
  display: flex;
  gap: 6px;
  margin-top: 10px;
}

.sk-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
}

/* 行 */
.sk-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 10px;
  border-bottom: 1px solid var(--border-subtle);
}

.sk-row__cell {
  height: 11px;
  flex-shrink: 0;
}

/* 表格 */
.sk-table {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
  overflow: hidden;
}

.sk-table__head {
  display: flex;
  gap: 12px;
  padding: 12px 10px;
  background: var(--bg-layer-1);
  border-bottom: 1px solid var(--border-subtle);
}

.sk-th {
  height: 10px;
}

/* 看板 */
.sk-board {
  display: flex;
  gap: 16px;
  overflow: hidden;
}

.sk-column {
  flex: 0 0 280px;
}

.sk-column__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--border-subtle);
}

.sk-pill {
  height: 12px;
  width: 70px;
}

.sk-pill--sm {
  width: 24px;
  height: 14px;
  border-radius: 8px;
}

.sk-column .sk-card {
  margin: 8px;
}

/* 仪表盘 */
.sk-dash {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.sk-dash__block {
  background: var(--bg-surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 16px;
}

.sk-dash__title {
  height: 12px;
  width: 40%;
  margin-bottom: 12px;
}

.sk-dash__metric {
  height: 22px;
  width: 30%;
  margin-bottom: 16px;
}

.sk-dash__bars {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  height: 60px;
}

.sk-dash__bar {
  flex: 1;
  min-height: 10px;
}
</style>
