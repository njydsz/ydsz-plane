<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from "vue";

import { init, use } from "echarts/core";
import { BarChart, LineChart, PieChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { ECharts, EChartsCoreOption } from "echarts/core";

/* ---- ECharts 按需注册（tree-shaking）----
 * ChartWidget 为通用封装，需覆盖所有 dashboard 子组件可能使用的系列与组件：
 *   BarChartWidget / LineChartWidget / PieChartWidget /
 *   StackedBarWidget / ModuleDistributionWidget /
 *   VersionBurndownWidget / ProjectCompareWidget / DefectAnalyticsView
 */
use([
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  CanvasRenderer,
]);

/**
 * ChartWidget - 通用 ECharts 封装。
 * 通过 ref 持有 echarts 实例，在 onBeforeUnmount 中调用 dispose() 释放资源，
 * 防止 HMR / 路由切换后内存泄漏。
 */
const props = defineProps<{
  /** 图表配置项（兼容 option / options 两种 prop 名） */
  options?: EChartsCoreOption;
  option?: EChartsCoreOption;
  /** 高度（数字按 px 处理） */
  height?: string | number;
}>();

const chartEl = ref<HTMLDivElement | null>(null);
const chartInstance = shallowRef<ECharts | null>(null);

const chartOption = computed(() => props.options ?? props.option);
const chartHeight = computed(() =>
  typeof props.height === "number" ? `${props.height}px` : (props.height ?? "240px"),
);

function render() {
  if (!chartEl.value || !chartOption.value) return;
  if (!chartInstance.value) {
    chartInstance.value = init(chartEl.value);
  }
  chartInstance.value.setOption(chartOption.value, true);
}

function handleResize() {
  chartInstance.value?.resize();
}

onMounted(() => {
  render();
  window.addEventListener("resize", handleResize);
});

watch(() => props.options ?? props.option, render, { deep: true });

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  chartInstance.value?.dispose();
  chartInstance.value = null;
});
</script>

<template>
  <div ref="chartEl" class="chart-widget" :style="{ height: chartHeight }" />
</template>

<style scoped>
.chart-widget {
  width: 100%;
  min-height: 0;
}
</style>
