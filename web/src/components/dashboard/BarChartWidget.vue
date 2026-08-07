<script setup lang="ts">
/**
 * BarChartWidget — 通用柱状图。
 * 根据 config.kind 区分 "state"（状态分布 / 分组柱）或 "velocity"（速率图）。
 */
import type { EChartsCoreOption } from "echarts";
import type { StateDistributionData, VelocityData } from "@/api/services/dashboard";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  data?: StateDistributionData | VelocityData;
  config?: Record<string, any>;
}>();

function asStateData(d?: StateDistributionData | VelocityData): StateDistributionData | null {
  return d && "by_state" in d ? (d as StateDistributionData) : null;
}

function asVelocityData(d?: StateDistributionData | VelocityData): VelocityData | null {
  return d && "sprints" in d ? (d as VelocityData) : null;
}

function buildOption(): EChartsCoreOption {
  const stateData = asStateData(props.data);
  if (stateData && stateData.by_state.length > 0) {
    const byState = stateData.by_state;
    return {
      tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
      grid: { top: 20, left: 40, right: 16, bottom: 28 },
      xAxis: {
        type: "category",
        data: byState.map((s: { state_name: string }) => s.state_name),
        axisLine: { lineStyle: { color: "var(--border-default)" } },
        axisLabel: { color: "var(--text-tertiary)", fontSize: 11, rotate: byState.length > 5 ? 30 : 0 },
      },
      yAxis: {
        type: "value",
        splitLine: { lineStyle: { color: "var(--border-subtle)" } },
        axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
      },
      series: [
        {
          type: "bar",
          barWidth: "55%",
          data: byState.map((s: { count: number; color: string }) => ({
            value: s.count,
            itemStyle: { color: s.color, borderRadius: [3, 3, 0, 0] },
          })),
        },
      ],
    };
  }

  // velocity 柱状图
  const velData = asVelocityData(props.data);
  const sprints = velData?.sprints ?? [];
  if (sprints.length === 0) return emptyOption;
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    legend: {
      data: ["完成", "承诺"],
      bottom: 0,
      textStyle: { color: "var(--text-secondary)", fontSize: 11 },
    },
    grid: { top: 16, left: 40, right: 16, bottom: 36 },
    xAxis: {
      type: "category",
      data: sprints.map((s) => s.sprint_name),
      axisLine: { lineStyle: { color: "var(--border-default)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11, rotate: sprints.length > 4 ? 30 : 0 },
    },
    yAxis: {
      type: "value",
      splitLine: { lineStyle: { color: "var(--border-subtle)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        name: "完成",
        type: "bar",
        barWidth: "30%",
        itemStyle: { color: "var(--success-500)", borderRadius: [3, 3, 0, 0] },
        data: sprints.map((s) => s.completed_count),
      },
      {
        name: "承诺",
        type: "bar",
        barWidth: "30%",
        itemStyle: { color: "var(--brand-500)", borderRadius: [3, 3, 0, 0] },
        data: sprints.map((s) => s.committed_count),
      },
    ],
  };
}

const emptyOption: EChartsCoreOption = {
  title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
};

function isStateData(d?: StateDistributionData | VelocityData): d is StateDistributionData {
  return !!d && "by_state" in d;
}
</script>

<template>
  <ChartWidget :options="buildOption()" />
</template>
