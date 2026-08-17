<script setup lang="ts">
/**
 * PieChartWidget — 优先级环形图。
 * 展示各优先级的需求/任务/缺陷数量占比。
 */
import type { EChartsCoreOption } from "echarts";
import type { PrioritySplitData } from "@/api/services/dashboard";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  data?: PrioritySplitData;
}>();

const COLORS = [
  "var(--priority-urgent)",
  "var(--priority-high)",
  "var(--priority-medium)",
  "var(--priority-low)",
  "var(--priority-none)",
];

function buildOption(): EChartsCoreOption {
  const byPriority = props.data?.by_priority ?? {};
  const entries = Object.entries(byPriority);
  if (entries.length === 0) {
    return { title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } } };
  }
  return {
    tooltip: { trigger: "item", formatter: "{b}: {c} ({d}%)" },
    legend: {
      orient: "horizontal",
      bottom: 0,
      textStyle: { color: "var(--text-secondary)", fontSize: 12 },
    },
    series: [
      {
        type: "pie",
        radius: ["40%", "68%"],
        center: ["50%", "44%"],
        avoidLabelOverlap: true,
        itemStyle: { borderRadius: 4, borderColor: "var(--surface-1)", borderWidth: 2 },
        label: { show: false },
        labelLine: { show: false },
        data: entries.map(([name, value], i) => ({
          name,
          value,
          itemStyle: { color: COLORS[i % COLORS.length] },
        })),
      },
    ],
  };
}
</script>

<template>
  <div class="pie-chart">
    <ChartWidget :options="buildOption()" />
    <div v-if="data" class="pie-chart__total">
      <span class="pie-chart__total-num">{{ data.total }}</span>
      <span class="pie-chart__total-label">总计</span>
    </div>
  </div>
</template>

<style scoped>
.pie-chart {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.pie-chart__total {
  position: absolute;
  top: 38%;
  display: flex;
  flex-direction: column;
  align-items: center;
  pointer-events: none;
}

.pie-chart__total-num {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.pie-chart__total-label {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}
</style>
