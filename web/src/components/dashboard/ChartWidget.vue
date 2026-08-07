<script setup lang="ts">
import * as echarts from "echarts";
import type { EChartsCoreOption } from "echarts";
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from "vue";

/**
 * ChartWidget - 通用 ECharts 封装。
 * 通过 ref 持有 echarts 实例，在 onBeforeUnmount 中调用 dispose() 释放资源，
 * 防止 HMR / 路由切换后内存泄漏。
 */
const props = defineProps<{
  options: EChartsCoreOption;
  height?: string;
}>();

const chartEl = ref<HTMLDivElement | null>(null);
const chartInstance = shallowRef<echarts.ECharts | null>(null);

function render() {
  if (!chartEl.value) return;
  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartEl.value);
  }
  chartInstance.value.setOption(props.options, true);
}

function handleResize() {
  chartInstance.value?.resize();
}

onMounted(() => {
  render();
  window.addEventListener("resize", handleResize);
});

watch(() => props.options, render, { deep: true });

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  chartInstance.value?.dispose();
  chartInstance.value = null;
});
</script>

<template>
  <div ref="chartEl" class="chart-widget" :style="{ height: height ?? '240px' }" />
</template>

<style scoped>
.chart-widget {
  width: 100%;
  min-height: 0;
}
</style>
