<script setup lang="ts">
/**
 * LineChartWidget — 燃尽折线图。
 * 根据 config.series 决定渲染 ideal / actual 两条线。
 */
import type { EChartsCoreOption } from "echarts";
import type { BurndownData } from "@/api/services/dashboard";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  data?: BurndownData;
  config?: Record<string, any>;
}>();

function buildOption(): EChartsCoreOption {
  const data = props.data;
  if (!data) return emptyOption;

  // 根据 remaining_days + total_points 模拟理想/实际曲线
  // 后端若返回逐日点列，可直接替换此逻辑
  const days = Math.max(data.remaining_days, 1);
  const totalPoints = data.total_points || 0;
  const burned = data.burned_points || 0;
  const ratio = totalPoints > 0 ? Math.min(burned / totalPoints, 1) : 0;

  const categories: string[] = [];
  const idealLine: number[] = [];
  const actualLine: number[] = [];

  for (let i = 0; i <= days; i++) {
    categories.push(`D${i}`);
    idealLine.push(Math.round(totalPoints * (1 - i / days)));
    if (i < days) {
      // 简化：均匀燃烧
      actualLine.push(Math.round(totalPoints - (totalPoints * i * ratio) / days));
    } else {
      actualLine.push(totalPoints - burned);
    }
  }

  return {
    tooltip: { trigger: "axis" },
    legend: {
      data: ["理想线", "实际线"],
      bottom: 0,
      textStyle: { color: "var(--text-secondary)", fontSize: 11 },
    },
    grid: { top: 20, left: 40, right: 16, bottom: 36 },
    xAxis: {
      type: "category",
      data: categories,
      axisLine: { lineStyle: { color: "var(--border-default)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
    },
    yAxis: {
      type: "value",
      splitLine: { lineStyle: { color: "var(--border-subtle)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        name: "理想线",
        type: "line",
        smooth: true,
        lineStyle: { type: "dashed", color: "var(--text-tertiary)", width: 1.5 },
        itemStyle: { color: "var(--text-tertiary)" },
        symbol: "none",
        data: idealLine,
      },
      {
        name: "实际线",
        type: "line",
        smooth: true,
        lineStyle: { color: "var(--brand-500)", width: 2 },
        itemStyle: { color: "var(--brand-500)" },
        areaStyle: { color: "rgba(63, 99, 241, 0.08)" },
        symbol: "circle",
        symbolSize: 5,
        data: actualLine,
      },
    ],
  };
}

const emptyOption: EChartsCoreOption = {
  title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
};
</script>

<template>
  <div class="line-chart">
    <div v-if="data" class="line-chart__meta">
      <span class="line-chart__meta-item">迭代: <strong>{{ data.sprint_name || data.sprint_id }}</strong></span>
      <span class="line-chart__meta-item">剩余天数: <strong>{{ data.remaining_days }}</strong></span>
    </div>
    <ChartWidget :options="buildOption()" />
  </div>
</template>

<style scoped>
.line-chart__meta {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.line-chart__meta-item strong {
  color: var(--text-secondary, #4b5563);
}
</style>
