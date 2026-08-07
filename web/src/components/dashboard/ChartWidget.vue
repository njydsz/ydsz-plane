<script setup lang="ts">
import * as echarts from "echarts";
import type { EChartsCoreOption } from "echarts";
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from "vue";

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
const chartInstance = shallowRef<echarts.ECharts | null>(null);

const chartOption = computed(() => props.options ?? props.option);
const chartHeight = computed(() =>
  typeof props.height === "number" ? `${props.height}px` : (props.height ?? "240px"),
);

function render() {
  if (!chartEl.value) return;
  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartEl.value);
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
