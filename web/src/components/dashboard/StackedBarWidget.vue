/**
 * StackedBarWidget — 仪表盘小组件：堆叠柱状图（团队成员工作量分布）。
 *
 * 基于 ECharts 渲染，横向堆叠展示每位成员的待办 / 进行中 / 已完成需求/任务/缺陷数，
 * 用于快速识别团队负载均衡情况。空数据时展示空态占位。
 */
<script setup lang="ts">
import type { EChartsCoreOption } from "echarts";
import type { TeamWorkloadWidgetData } from "@/api/services/dashboard";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  data?: TeamWorkloadWidgetData;
}>();

function buildOption(): EChartsCoreOption {
  const members = props.data?.members ?? [];
  if (members.length === 0) return emptyOption;

  const names = members.map((m) => m.user_name);
  const todoData = members.map((m) => m.todo);
  const inProgressData = members.map((m) => m.in_progress);
  const doneData = members.map((m) => m.done);

  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    legend: {
      data: ["待办", "进行中", "已完成"],
      bottom: 0,
      textStyle: { color: "var(--text-secondary)", fontSize: 11 },
    },
    grid: { top: 16, left: 80, right: 16, bottom: 36 },
    xAxis: {
      type: "value",
      splitLine: { lineStyle: { color: "var(--border-subtle)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
    },
    yAxis: {
      type: "category",
      data: names,
      axisLine: { lineStyle: { color: "var(--border-default)" } },
      axisLabel: { color: "var(--text-secondary)", fontSize: 12 },
    },
    series: [
      {
        name: "待办",
        type: "bar",
        stack: "total",
        barWidth: "50%",
        itemStyle: { color: "var(--text-tertiary)", borderRadius: [0, 0, 0, 0] },
        data: todoData,
      },
      {
        name: "进行中",
        type: "bar",
        stack: "total",
        itemStyle: { color: "var(--brand-500)", borderRadius: [0, 0, 0, 0] },
        data: inProgressData,
      },
      {
        name: "已完成",
        type: "bar",
        stack: "total",
        itemStyle: { color: "var(--success-500)", borderRadius: [0, 3, 3, 0] },
        data: doneData,
      },
    ],
  };
}

const emptyOption: EChartsCoreOption = {
  title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
};
</script>

<template>
  <ChartWidget :options="buildOption()" :height="'280px'" />
</template>
