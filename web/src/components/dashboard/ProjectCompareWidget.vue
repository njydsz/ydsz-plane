<script setup lang="ts">
/**
 * ProjectCompareWidget — 多项目对比。
 *
 * 展示工作空间下所有项目的需求完成率 / 缺陷数对比条形图。
 * 数据来自 workspace 级聚合接口 dashboardApi.getProjectCompare。
 */
import { computed, onMounted, ref } from "vue";
import type { EChartsCoreOption } from "echarts";
import { dashboardApi, type ProjectCompareItem } from "@/api/services/dashboard";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  wsId?: number;
  projectId?: number;
  config?: Record<string, any>;
}>();

const loading = ref(true);
const error = ref("");
const items = ref<ProjectCompareItem[]>([]);

async function load() {
  if (!props.wsId) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    items.value = await dashboardApi.getProjectCompare(props.wsId);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

const chartOption = computed<EChartsCoreOption>(() => {
  const rows = items.value;
  if (rows.length === 0) {
    return {
      title: { text: "暂无项目数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
    };
  }
  const labels = rows.map((r) => r.project_name);
  return {
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      formatter: (params: any) => {
        const idx = params[0]?.dataIndex ?? 0;
        const row = rows[idx];
        if (!row) return "";
        const lines = params.map(
          (p: any) =>
            `${p.marker}${p.seriesName}：<b>${typeof p.value === "number" ? (p.seriesName.includes("完成率") ? (p.value * 100).toFixed(1) + "%" : p.value) : p.value}</b>`,
        );
        return `${row.project_name}<br/>${lines.join("<br/>")}<br/>需求 <b>${row.done_issues}/${row.total_issues}</b> · 活跃迭代 <b>${row.active_sprint_count}</b>`;
      },
    },
    legend: { bottom: 0, textStyle: { color: "var(--text-secondary)", fontSize: 11 } },
    grid: { top: 20, left: 44, right: 16, bottom: 36 },
    xAxis: {
      type: "category",
      data: labels,
      axisLine: { lineStyle: { color: "var(--border-default)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 11, rotate: labels.length > 4 ? 25 : 0 },
    },
    yAxis: [
      {
        type: "value",
        name: "完成率",
        max: 1,
        axisLabel: { color: "var(--text-tertiary)", fontSize: 11, formatter: (v: number) => `${Math.round(v * 100)}%` },
        splitLine: { lineStyle: { color: "var(--border-subtle)" } },
      },
      {
        type: "value",
        name: "缺陷数",
        minInterval: 1,
        splitLine: { show: false },
        axisLabel: { color: "var(--text-tertiary)", fontSize: 11 },
      },
    ],
    series: [
      {
        name: "需求完成率",
        type: "bar",
        barWidth: "30%",
        itemStyle: { color: "var(--success-500, #10b981)", borderRadius: [3, 3, 0, 0] },
        data: rows.map((r) => r.completion_rate),
      },
      {
        name: "缺陷数",
        type: "bar",
        yAxisIndex: 1,
        barWidth: "30%",
        itemStyle: { color: "var(--danger-400, #f87171)", borderRadius: [3, 3, 0, 0] },
        data: rows.map((r) => r.defect_count),
      },
    ],
  };
});

onMounted(load);
</script>

<template>
  <div class="project-compare">
    <div v-if="loading" class="project-compare__hint">加载中...</div>
    <div v-else-if="error" class="project-compare__hint project-compare__hint--error">{{ error }}</div>
    <ChartWidget v-else :option="chartOption" :height="200" />
  </div>
</template>

<style scoped>
.project-compare {
  height: 100%;
}

.project-compare__hint {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
}

.project-compare__hint--error {
  color: var(--danger-500, #dc2f2f);
}
</style>
