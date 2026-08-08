<script setup lang="ts">
/**
 * VersionBurndownWidget — 版本燃尽图。
 *
 * 展示各版本的计划 vs 实际进度（按版本分组的 issues 数量与完成率）。
 * 数据由组件自行从 issueApi 拉取分组统计。
 */
import { computed, onMounted, ref } from "vue";
import type { EChartsCoreOption } from "echarts";
import { issueApi } from "@/api/services/issue";
import ChartWidget from "./ChartWidget.vue";

const props = defineProps<{
  wsId?: number;
  projectId?: number;
  config?: Record<string, any>;
}>();

interface VersionStat {
  name: string;
  total: number;
  done: number;
  remaining: number;
}

const loading = ref(true);
const error = ref("");
const versionStats = ref<VersionStat[]>([]);

async function loadStats() {
  if (!props.wsId || !props.projectId) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await issueApi.listIssues(props.wsId, props.projectId, { limit: 200 });
    const statsMap = new Map<string, { total: number; done: number }>();
    for (const issue of res.results) {
      // 按 type_code 分组统计（简化版：因为没有版本域，用类型代替）
      // 等版本域上线后改为按 version 分组
      const key = issue.type_code;
      const s = statsMap.get(key) ?? { total: 0, done: 0 };
      s.total++;
      if (issue.progress >= 100) s.done++;
      statsMap.set(key, s);
    }
    versionStats.value = Array.from(statsMap.entries()).map(([name, s]) => ({
      name: typeName(name),
      total: s.total,
      done: s.done,
      remaining: s.total - s.done,
    }));
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function typeName(code: string): string {
  const map: Record<string, string> = {
    epic: "史诗",
    requirement: "需求",
    task: "任务",
    defect: "缺陷",
  };
  return map[code] ?? code;
}

const chartOption = computed<EChartsCoreOption>(() => {
  if (versionStats.value.length === 0) {
    return {
      title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
    };
  }
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    legend: {
      data: ["已完成", "剩余"],
      bottom: 0,
      textStyle: { color: "var(--text-secondary)", fontSize: 11 },
    },
    grid: { top: 20, left: 40, right: 16, bottom: 36 },
    xAxis: {
      type: "category",
      data: versionStats.value.map((v) => v.name),
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
        name: "已完成",
        type: "bar",
        stack: "total",
        data: versionStats.value.map((v) => v.done),
        itemStyle: { color: "var(--success-500, #10b981)", borderRadius: [0, 0, 0, 0] },
        barWidth: "40%",
      },
      {
        name: "剩余",
        type: "bar",
        stack: "total",
        data: versionStats.value.map((v) => v.remaining),
        itemStyle: { color: "var(--brand-200, #c7d2fe)", borderRadius: [4, 4, 0, 0] },
        barWidth: "40%",
      },
    ],
  };
});

onMounted(loadStats);
</script>

<template>
  <div class="version-burndown">
    <ChartWidget
      v-if="!loading && !error"
      :option="chartOption"
      :height="180"
    />
    <div v-else-if="loading" class="loading-hint">加载中...</div>
    <div v-else class="error-hint">{{ error }}</div>
    <!-- 进度总览 -->
    <div v-if="versionStats.length" class="version-stats">
      <div v-for="v in versionStats" :key="v.name" class="stat-row">
        <span class="stat-name">{{ v.name }}</span>
        <span class="stat-bar">
          <span
            class="stat-bar__fill"
            :style="{ width: v.total > 0 ? (v.done * 100 / v.total) + '%' : '0%' }"
          ></span>
        </span>
        <span class="stat-pct">{{ v.total > 0 ? Math.round(v.done * 100 / v.total) : 0 }}%</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.version-burndown {
  display: flex;
  flex-direction: column;
  gap: 10px;
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

.version-stats {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 4px;
  border-top: 1px solid var(--border-subtle, #e5e7eb);
}

.stat-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.stat-name {
  width: 40px;
  flex-shrink: 0;
  color: var(--text-secondary, #4b5563);
}

.stat-bar {
  flex: 1;
  height: 6px;
  background: var(--surface-2, #f3f4f6);
  border-radius: 3px;
  overflow: hidden;
}

.stat-bar__fill {
  display: block;
  height: 100%;
  background: var(--success-500, #10b981);
  border-radius: 3px;
  transition: width 0.3s;
}

.stat-pct {
  width: 32px;
  text-align: right;
  flex-shrink: 0;
  color: var(--text-tertiary, #9ca3af);
}
</style>
