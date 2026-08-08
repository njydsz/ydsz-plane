<script setup lang="ts">
/**
 * BurnupChart — 交互式燃起图（基于 ECharts）。
 *
 * 能力：
 *  - 两条实线：已完成线（底部上升）、总量线（顶部波动）
 *  - 理想参考线（虚线）
 *  - 可直观看出范围蔓延（总量线的上升）
 *  - 悬停 Tooltip、图例切换、响应式
 */

import { computed, onMounted, onUnmounted, ref } from "vue";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { LineChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { EChartsOption } from "echarts";
import { AppLoadingState, AppErrorState } from "@/components";

use([
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  CanvasRenderer,
]);

/** 燃起图单日数据点 */
export interface BurnupPoint {
  date: string;
  done_points: number;
  total_points: number;
  ideal_line: number;
}

const props = defineProps<{
  points: BurnupPoint[];
  loading?: boolean;
  error?: string | null;
}>();

defineEmits<{ retry: [] }>();

const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const containerRef = ref<HTMLElement | null>(null);
const chartHeight = ref("320px");

let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
  if (containerRef.value) {
    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const w = entry.contentRect.width;
        chartHeight.value = w < 480 ? "240px" : w < 768 ? "280px" : "340px";
      }
    });
    resizeObserver.observe(containerRef.value);
  }
});

onUnmounted(() => {
  resizeObserver?.disconnect();
});

function fmtDate(d: string): string {
  if (!d) return "";
  const parts = d.split("-");
  if (parts.length >= 3) return `${parseInt(parts[1])}/${parseInt(parts[2])}`;
  return d;
}

const option = computed<EChartsOption>(() => {
  const hasData = props.points.length > 0;
  const dates = hasData ? props.points.map((p) => fmtDate(p.date)) : [];
  const idealData = hasData ? props.points.map((p) => p.ideal_line) : [];
  const totalData = hasData ? props.points.map((p) => p.total_points) : [];
  const doneData = hasData ? props.points.map((p) => p.done_points) : [];

  if (!hasData) {
    return {
      title: {
        text: "暂无燃起图数据",
        left: "center",
        top: "center",
        textStyle: { color: "#999", fontSize: 14, fontWeight: "normal" as const },
      },
      grid: { show: false },
    };
  }

  const enableZoom = props.points.length > 10;

  return {
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "cross" },
      formatter: (params: unknown) => {
        const items = params as Array<{ axisValue: string; seriesName: string; marker: string; value: number }>;
        if (!items?.length) return "";
        let html = `<div style="font-weight:600;margin-bottom:4px">${items[0].axisValue}</div>`;
        for (const item of items) {
          html += `<div>${item.marker} ${item.seriesName}: <b>${item.value} pt</b></div>`;
        }
        return html;
      },
    },
    legend: {
      data: ["总量线", "已完成", "理想线"],
      bottom: 0,
      textStyle: { fontSize: 12 },
    },
    grid: {
      left: "3%", right: "4%", top: "8%",
      bottom: enableZoom ? "16%" : "12%",
      containLabel: true,
    },
    xAxis: {
      type: "category", data: dates, boundaryGap: false,
      axisLabel: { fontSize: 10, color: "#999" },
    },
    yAxis: {
      type: "value", name: "故事点 (pt)",
      nameTextStyle: { fontSize: 11, color: "#999" },
      axisLabel: { fontSize: 10, color: "#999" },
      splitLine: { lineStyle: { color: "#f0f0f0", type: "dashed" } },
      minInterval: 1,
    },
    dataZoom: enableZoom ? [{
      type: "slider", start: 0, end: 100, height: 20, bottom: 30,
      borderColor: "transparent", backgroundColor: "#f5f5f5",
      fillerColor: "rgba(16,185,129,0.1)",
    }] : undefined,
    series: [
      {
        name: "总量线",
        type: "line", data: totalData,
        lineStyle: { color: "#faad14", width: 2 },
        itemStyle: { color: "#faad14" },
        symbol: "circle", symbolSize: 4,
        areaStyle: {
          color: { type: "linear", x: 0, y: 0, x2: 0, y2: 1, colorStops: [
            { offset: 0, color: "rgba(250,173,20,0.15)" },
            { offset: 1, color: "rgba(250,173,20,0.02)" },
          ] },
        },
      },
      {
        name: "已完成",
        type: "line", data: doneData,
        lineStyle: { color: "#10b981", width: 2.5 },
        itemStyle: { color: "#10b981" },
        symbol: "diamond", symbolSize: 5,
        areaStyle: {
          color: { type: "linear", x: 0, y: 0, x2: 0, y2: 1, colorStops: [
            { offset: 0, color: "rgba(16,185,129,0.2)" },
            { offset: 1, color: "rgba(16,185,129,0.02)" },
          ] },
        },
      },
      {
        name: "理想线",
        type: "line", data: idealData,
        lineStyle: { color: "#d1d5db", type: "dashed", width: 1.5 },
        itemStyle: { color: "#d1d5db" },
        symbol: "none",
      },
    ],
  };
});

defineExpose({ chartRef });
</script>

<template>
  <div ref="containerRef" class="burnup-chart" :style="{ minHeight: chartHeight }">
    <AppLoadingState v-if="loading" text="加载燃起图数据..." />
    <AppErrorState v-else-if="error" :message="error" @retry="$emit('retry')" />
    <v-chart v-else ref="chartRef" class="chart-canvas" :option="option" :style="{ height: chartHeight }" autoresize />
  </div>
</template>

<style scoped>
.burnup-chart {
  position: relative;
  background: var(--surface-1, #fff);
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--border-subtle, #f0f0f0);
  overflow: hidden;
}
.chart-canvas {
  width: 100%;
}
</style>
