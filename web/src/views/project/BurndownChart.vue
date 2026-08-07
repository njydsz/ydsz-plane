<script setup lang="ts">
/**
 * BurndownChart — 交互式燃尽图（基于 ECharts）。
 *
 * 能力：
 *  - 三条折线：理想线（虚线）、实际剩余线（实线）、已完成线（面积）
 *  - 悬停 Tooltip 展示当日明细
 *  - 图例切换显示/隐藏
 *  - 数据缩放（长迭代时启用）
 *  - 迭代起止日期标注线
 *  - 响应式尺寸（ResizeObserver）
 *  - 空数据 / 加载态 / 错误态
 *
 * 设计参考：ECharts 官方最佳实践 + 互联网大厂数据可视化规范（蚂蚁 G2/字节 VChart）
 */
import { computed, onMounted, onUnmounted, ref } from "vue";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { AppLoadingState, AppErrorState } from "@/components";
import { LineChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  MarkLineComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { EChartsOption } from "echarts";

/* ---- ECharts 按需注册（tree-shaking）---- */
use([
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  DataZoomComponent,
  MarkLineComponent,
  CanvasRenderer,
]);

/* ------------------------------------------------------------------ */
/* Types                                                               */
/* ------------------------------------------------------------------ */

/** 燃尽图单日数据点 */
export interface BPoint {
  date: string;
  done_points: number;
  remaining: number;
  ideal_line: number;
}

/* ------------------------------------------------------------------ */
/* Props                                                               */
/* ------------------------------------------------------------------ */

const props = defineProps<{
  points: BPoint[];
  startDate?: string;
  endDate?: string;
  loading?: boolean;
  error?: string | null;
}>();

/* ------------------------------------------------------------------ */
/* 响应式容器                                                           */
/* ------------------------------------------------------------------ */

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

/* ------------------------------------------------------------------ */
/* 日期格式化                                                           */
/* ------------------------------------------------------------------ */

function fmtDate(d: string): string {
  if (!d) return "";
  // ISO date → MM/DD
  const parts = d.split("-");
  if (parts.length >= 3) return `${parseInt(parts[1])}/${parseInt(parts[2])}`;
  return d;
}

function fmtTooltipDate(d: string): string {
  if (!d) return "";
  const parts = d.split("-");
  if (parts.length >= 3) return `${parts[0]}-${parts[1]}-${parts[2]}`;
  return d;
}

/* ------------------------------------------------------------------ */
/* ECharts 配置                                                         */
/* ------------------------------------------------------------------ */

const option = computed<EChartsOption>(() => {
  const hasData = props.points.length > 0;

  // X 轴数据（日期）
  const dates = hasData ? props.points.map((p) => fmtDate(p.date)) : [];

  // 三条线
  const idealData = hasData ? props.points.map((p) => p.ideal_line) : [];
  const remainingData = hasData ? props.points.map((p) => p.remaining) : [];
  const doneData = hasData ? props.points.map((p) => p.done_points) : [];

  // 空数据占位
  if (!hasData) {
    return {
      title: {
        text: "暂无燃尽图数据",
        subtext: "启动迭代后将自动生成每日快照",
        left: "center",
        top: "center",
        textStyle: { color: "#999", fontSize: 14, fontWeight: "normal" as const },
        subtextStyle: { color: "#bbb", fontSize: 12 },
      },
      grid: { show: false },
    };
  }

  // 标记线：迭代起止日期
  const markLines: EChartsOption["series"] = [];
  if (props.startDate) {
    const startIdx = dates.indexOf(fmtDate(props.startDate));
    if (startIdx >= 0) {
      markLines.push({
        type: "line",
        name: "迭代启动",
        markLine: {
          silent: true,
          symbol: "none",
          lineStyle: { type: "solid", color: "#52c41a", width: 1.5 },
          label: { show: true, formatter: "启动", color: "#52c41a", fontSize: 11 },
          data: [{ xAxis: startIdx }],
        },
      });
    }
  }
  if (props.endDate) {
    const endIdx = dates.lastIndexOf(fmtDate(props.endDate));
    if (endIdx >= 0) {
      markLines.push({
        type: "line",
        name: "迭代结束",
        markLine: {
          silent: true,
          symbol: "none",
          lineStyle: { type: "dashed", color: "#ff4d4f", width: 1.5 },
          label: { show: true, formatter: "截止", color: "#ff4d4f", fontSize: 11 },
          data: [{ xAxis: endIdx }],
        },
      });
    }
  }

  const enableZoom = props.points.length > 10;

  return {
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "cross", crossStyle: { color: "#999" } },
      backgroundColor: "rgba(255,255,255,0.95)",
      borderColor: "#e8e8e8",
      borderWidth: 1,
      textStyle: { color: "#333", fontSize: 12 },
      formatter: (params: unknown) => {
        const items = params as Array<{
          axisValue: string;
          seriesName: string;
          marker: string;
          value: number;
        }>;
        if (!items?.length) return "";
        const dateIdx = dates.indexOf(items[0].axisValue);
        const rawDate = dateIdx >= 0 ? fmtTooltipDate(props.points[dateIdx].date) : items[0].axisValue;
        let html = `<div style="font-weight:600;margin-bottom:4px">📅 ${rawDate}</div>`;
        for (const item of items) {
          html += `<div style="display:flex;align-items:center;gap:4px;">${item.marker} ${item.seriesName}: <b>${item.value} pt</b></div>`;
        }
        return html;
      },
    },
    legend: {
      data: ["理想线", "剩余点数", "已完成点数"],
      bottom: 0,
      textStyle: { fontSize: 12, color: "#666" },
      itemWidth: 18,
      itemHeight: 10,
    },
    grid: {
      left: "3%",
      right: "4%",
      top: "8%",
      bottom: enableZoom ? "16%" : "12%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      data: dates,
      boundaryGap: false,
      axisLabel: { fontSize: 10, color: "#999", rotate: props.points.length > 15 ? 45 : 0 },
      axisLine: { lineStyle: { color: "#e8e8e8" } },
      axisTick: { show: false },
    },
    yAxis: {
      type: "value",
      name: "故事点 (pt)",
      nameTextStyle: { fontSize: 11, color: "#999" },
      axisLabel: { fontSize: 10, color: "#999" },
      splitLine: { lineStyle: { color: "#f0f0f0", type: "dashed" } },
      minInterval: 1,
    },
    dataZoom: enableZoom
      ? [
          {
            type: "slider",
            start: 0,
            end: 100,
            height: 20,
            bottom: 30,
            borderColor: "transparent",
            backgroundColor: "#f5f5f5",
            fillerColor: "rgba(24,144,255,0.15)",
            handleStyle: { color: "#1890ff" },
            textStyle: { fontSize: 10 },
          },
        ]
      : undefined,
    series: [
      {
        name: "理想线",
        type: "line",
        data: idealData,
        lineStyle: { color: "#faad14", type: "dashed", width: 1.5 },
        itemStyle: { color: "#faad14" },
        symbol: "none",
        smooth: false,
      },
      {
        name: "剩余点数",
        type: "line",
        data: remainingData,
        lineStyle: { color: "#1890ff", width: 2.5 },
        itemStyle: { color: "#1890ff" },
        symbol: "circle",
        symbolSize: 5,
        smooth: false,
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(24,144,255,0.15)" },
              { offset: 1, color: "rgba(24,144,255,0.02)" },
            ],
          },
        },
      },
      {
        name: "已完成点数",
        type: "line",
        data: doneData,
        lineStyle: { color: "#52c41a", width: 2.5 },
        itemStyle: { color: "#52c41a" },
        symbol: "diamond",
        symbolSize: 5,
        smooth: false,
      },
      ...markLines,
    ],
  };
});

/* ------------------------------------------------------------------ */
/* ECharts 实例引用 — 供外部获取底层图表实例                              */
/* ------------------------------------------------------------------ */

defineExpose({ chartRef });
</script>

<template>
  <div ref="containerRef" class="burndown-chart" :style="{ minHeight: chartHeight }">
    <!-- 加载态 -->
    <AppLoadingState v-if="loading" text="加载燃尽图数据..." />

    <!-- 错误态 -->
    <AppErrorState
      v-else-if="error"
      :message="error"
      @retry="$emit('retry')"
    />

    <!-- 图表 -->
    <v-chart
      v-else
      ref="chartRef"
      class="chart-canvas"
      :option="option"
      :style="{ height: chartHeight }"
      autoresize
    />
  </div>
</template>

<style scoped>
.burndown-chart {
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
