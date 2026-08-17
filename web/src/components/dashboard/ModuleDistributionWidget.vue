<script setup lang="ts">
/**
 * ModuleDistributionWidget — 模块分布图。
 *
 * 展示各模块的需求/任务/缺陷数量与进度。
 * 数据由组件自行拉取模块列表并统计 issues 数量。
 */
import { computed, onMounted, ref } from "vue";
import type { EChartsCoreOption } from "echarts";
import { moduleApi } from "@/api/services/module";
import { issueApi } from "@/api/services/issue";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  wsId?: number;
  projectId?: number;
  config?: Record<string, any>;
}>();

interface ModuleStat {
  name: string;
  issueCount: number;
}

const loading = ref(true);
const error = ref("");
const stats = ref<ModuleStat[]>([]);

async function loadStats() {
  if (!props.wsId || !props.projectId) return;
  loading.value = true;
  error.value = "";
  try {
    const modules = await moduleApi.list(props.wsId, props.projectId);

    if (modules.length === 0) {
      stats.value = [];
      loading.value = false;
      return;
    }

    // 拉取全量 issues 并按 modules 字段统计
    const res = await issueApi.listIssues(props.wsId, props.projectId, { limit: 200 });

    const countMap = new Map<number, number>();
    for (const issue of res.results) {
      for (const modId of issue.modules) {
        countMap.set(modId, (countMap.get(modId) ?? 0) + 1);
      }
    }

    // 取 TOP 8 模块 + 其他
    const allStats = modules
      .map((m) => ({ name: m.name, issueCount: countMap.get(m.id) ?? 0 }))
      .sort((a, b) => b.issueCount - a.issueCount);

    const top = allStats.slice(0, 8);
    const rest = allStats.slice(8);
    if (rest.length > 0) {
      const otherCount = rest.reduce((sum, r) => sum + r.issueCount, 0);
      top.push({ name: `其他 (${rest.length})`, issueCount: otherCount });
    }
    stats.value = top;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

const COLORS = [
  "#3f63f1", "#10b981", "#f59e0b", "#ef4444",
  "#8b5cf6", "#06b6d4", "#ec4899", "#84cc16",
  "#6b7280",
];

const chartOption = computed<EChartsCoreOption>(() => {
  if (stats.value.length === 0) {
    return {
      title: { text: "暂无模块数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
    };
  }
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    grid: { top: 10, left: 8, right: 16, bottom: 8, containLabel: true },
    xAxis: {
      type: "value",
      splitLine: { lineStyle: { color: "var(--border-subtle)" } },
      axisLabel: { color: "var(--text-tertiary)", fontSize: 10 },
    },
    yAxis: {
      type: "category",
      data: stats.value.map((s) => s.name).reverse(),
      axisLine: { lineStyle: { color: "var(--border-default)" } },
      axisLabel: { color: "var(--text-secondary)", fontSize: 11 },
    },
    series: [
      {
        type: "bar",
        data: stats.value.map((s, i) => ({
          value: s.issueCount,
          itemStyle: { color: COLORS[i % COLORS.length], borderRadius: [0, 4, 4, 0] },
        })).reverse(),
        barWidth: "60%",
        label: { show: true, position: "right", color: "var(--text-tertiary)", fontSize: 11 },
      },
    ],
  };
});

onMounted(loadStats);
</script>

<template>
  <div class="module-dist">
    <ChartWidget
      v-if="!loading && !error"
      :option="chartOption"
      :height="180"
    />
    <div v-else-if="loading" class="loading-hint">加载中...</div>
    <div v-else class="error-hint">{{ error }}</div>
  </div>
</template>

<style scoped>
.module-dist {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.loading-hint,
.error-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
}

.error-hint {
  color: var(--danger-500, #dc2f2f);
}
</style>
