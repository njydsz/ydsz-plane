<script setup lang="ts">
/**
 * MiniGantt — 紧凑型甘特图，展示版本关联迭代时间线。
 *
 * Props:
 *   sprints:    关联的迭代列表（含日期和进度）
 *   versionStart: 可选，标记版本开始日期
 *   versionEnd:   可选，标记版本目标日期
 */
import { computed } from "vue";

/** 甘特图迭代条目（含日期范围与进度百分比） */
export interface GanttSprint {
  id: number;
  name: string;
  startDate?: string;
  endDate?: string;
  progress?: number; // 0-100
  status?: string;
}

const props = withDefaults(defineProps<{
  sprints: GanttSprint[];
  versionStart?: string;
  versionEnd?: string;
}>(), {
  versionStart: "",
  versionEnd: "",
});

interface TimelineEntry {
  sprint: GanttSprint;
  left: number;
  width: number;
}

/**
 * 计算每条迭代在时间轴上的定位（left）与宽度（width），均为百分比。
 *
 * 算法说明：
 *  - 时间轴范围 = [min, max]，取所有迭代/版本 target_date 属性的最小与最大值；
 *    若范围不足 7 天则强制拉长为 7 天，避免日期间隔过近时条块重叠。
 *  - left   = (迭代开始 - min) / range * 100，即相对起始点的偏移百分比。
 *  - width  = (迭代结束 - 开始) / range * 100，最小宽度钳制为 2%，保证可见。
 *  - 缺失开始/结束日期的迭代分别回退到 min / max，从而铺满可视范围。
 */
const timeline = computed<TimelineEntry[]>(() => {
  const items = props.sprints.filter((s) => s.startDate || s.endDate);
  if (items.length === 0) return [];

  // 汇总所有日期点（迭代起止 + 版本起止），用于确定时间轴整体范围
  const allDates: number[] = [];
  items.forEach((s) => {
    if (s.startDate) allDates.push(new Date(s.startDate).getTime());
    if (s.endDate) allDates.push(new Date(s.endDate).getTime());
  });
  if (props.versionStart) allDates.push(new Date(props.versionStart).getTime());
  if (props.versionEnd) allDates.push(new Date(props.versionEnd).getTime());

  // 时间轴最小/最大值，并确保至少有 7 天的可视宽度
  const min = Math.min(...allDates);
  const max = Math.max(...allDates);
  const range = Math.max(max - min, 7 * 86400000); // min 7 days

  return items.map((s) => {
    const start = s.startDate ? new Date(s.startDate).getTime() : min;
    const end = s.endDate ? new Date(s.endDate).getTime() : max;
    return {
      sprint: s,
      left: ((start - min) / range) * 100,
      width: Math.max(((end - start) / range) * 100, 2),
    };
  });
});

const hasData = computed(() => timeline.value.length > 0);

/**
 * 生成月份刻度标签。
 * 从时间轴起始月份逐月递增到结束月份，仅保留落在 [min, max] 区间内的月份，
 * 并计算每个月份首日在时间轴上的百分比位置，供横轴刻度渲染。
 */
const months = computed(() => {
  if (timeline.value.length === 0) return [];
  // 复用与 timeline 相同的时间范围计算
  const items = props.sprints.filter((s) => s.startDate || s.endDate);
  const allDates: number[] = [];
  items.forEach((s) => {
    if (s.startDate) allDates.push(new Date(s.startDate).getTime());
    if (s.endDate) allDates.push(new Date(s.endDate).getTime());
  });
  const min = Math.min(...allDates);
  const max = Math.max(...allDates);
  const range = Math.max(max - min, 7 * 86400000);

  const result: { label: string; left: number }[] = [];
  const start = new Date(min);
  const end = new Date(max);
  // 从起始月份的首日作为游标，逐月推进
  const cursor = new Date(start.getFullYear(), start.getMonth(), 1);

  while (cursor <= end) {
    const ts = cursor.getTime();
    if (ts >= min) {
      const left = ((ts - min) / range) * 100;
      result.push({
        label: `${cursor.getFullYear()}-${String(cursor.getMonth() + 1).padStart(2, "0")}`,
        left,
      });
    }
    cursor.setMonth(cursor.getMonth() + 1);
  }
  return result;
});
</script>

<template>
  <div class="gantt">
    <div v-if="!hasData" class="gantt__empty">
      暂无关联迭代日期数据，无法展示甘特图
    </div>
    <div v-else class="gantt__chart">
      <!-- 月份刻度 -->
      <div class="gantt__months">
        <span
          v-for="(m, i) in months"
          :key="i"
          class="gantt__month-label"
          :style="{ left: `${m.left}%` }"
        >{{ m.label }}</span>
      </div>

      <!-- 迭代条 -->
      <div
        v-for="entry in timeline"
        :key="entry.sprint.id"
        class="gantt__row"
      >
        <span class="gantt__name">{{ entry.sprint.name }}</span>
        <div class="gantt__track">
          <div
            class="gantt__bar"
            :class="{ 'gantt__bar--completed': entry.sprint.status === 'completed' }"
            :style="{
              left: `${entry.left}%`,
              width: `${entry.width}%`,
            }"
          >
            <span class="gantt__bar-label">
              {{ entry.sprint.startDate ?? "?" }} ~ {{ entry.sprint.endDate ?? "?" }}
              <template v-if="entry.sprint.progress != null">
                · {{ Math.round(entry.sprint.progress) }}%
              </template>
            </span>
          </div>
        </div>
      </div>

      <!-- 版本目标日期标记 -->
      <div
v-if="versionEnd" class="gantt__marker" :style="{
        left: `${(() => {
          const items = sprints.filter(s => s.startDate || s.endDate);
          const allDates: number[] = [];
          items.forEach(s => { if (s.startDate) allDates.push(new Date(s.startDate).getTime()); if (s.endDate) allDates.push(new Date(s.endDate).getTime()); });
          if (versionStart) allDates.push(new Date(versionStart).getTime());
          if (versionEnd) allDates.push(new Date(versionEnd).getTime());
          const min = Math.min(...allDates);
          const max = Math.max(...allDates);
          const range = Math.max(max - min, 7 * 86400000);
          return ((new Date(versionEnd).getTime() - min) / range) * 100;
        })()}%`,
      }"
>
        <span class="gantt__marker-label">目标 {{ versionEnd }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.gantt {
  width: 100%;
  font-size: 12px;
}

.gantt__empty {
  padding: 24px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

.gantt__chart {
  position: relative;
}

.gantt__months {
  position: relative;
  height: 20px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--border-subtle);
}

.gantt__month-label {
  position: absolute;
  transform: translateX(-50%);
  font-size: 10px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  white-space: nowrap;
}

.gantt__row {
  display: flex;
  align-items: center;
  height: 36px;
  border-bottom: 1px solid var(--border-subtle);
}

.gantt__name {
  width: 100px;
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
  padding-right: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gantt__track {
  flex: 1;
  position: relative;
  height: 24px;
  background: var(--surface-2);
  border-radius: 4px;
  overflow: visible;
}

.gantt__bar {
  position: absolute;
  top: 0;
  height: 100%;
  background: var(--brand-500);
  border-radius: 4px;
  min-width: 4px;
  display: flex;
  align-items: center;
  padding: 0 6px;
  transition: opacity 0.15s;
}

.gantt__bar:hover {
  opacity: 0.85;
}

.gantt__bar--completed {
  background: var(--success-500);
}

.gantt__bar-label {
  font-size: 10px;
  color: var(--text-on-brand);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: var(--font-mono);
}

.gantt__marker {
  position: absolute;
  top: 20px;
  bottom: 0;
  width: 2px;
  background: var(--danger-500);
  z-index: 1;
}

.gantt__marker-label {
  position: absolute;
  top: -16px;
  left: 4px;
  font-size: 10px;
  color: var(--danger-500);
  white-space: nowrap;
  font-weight: 500;
}
</style>
