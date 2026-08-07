<script setup lang="ts">
/**
 * DefectAnalyticsView — 项目缺陷分析报表页。
 *
 * 对标 Jira Issue Statistics / Plane Analytics / Linear Insights。
 * 提供多维度交叉分析：严重程度、发现阶段、模块、根因、缺陷龄、趋势。
 * 纯只读视图，所有数据由后端聚合接口一次性返回。
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import type { EChartsCoreOption } from "echarts";

import { defectAnalyticsApi, SEVERITY_LABELS, PHASE_LABELS } from "@/api/services/defectAnalytics";
import type { DefectAnalytics, SeverityCount, PhaseCount, ModuleCount, RootCauseCount, AgeBucket, TrendPoint } from "@/api/services/defectAnalytics";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";
import ChartWidget from "@/components/dashboard/ChartWidget.vue";
import { toast } from "@/lib/toast";

const route = useRoute();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = computed(() => wsStore.current?.id ?? 0);

const loading = ref(true);
const error = ref("");
const analytics = ref<DefectAnalytics | null>(null);

/** 时间范围快捷选项 */
const TIME_RANGES = [
  { value: "30d", label: "30 天" },
  { value: "90d", label: "90 天" },
  { value: "180d", label: "半年" },
  { value: "all", label: "全部" },
];

function getDateRange(range: string): { from?: string; to?: string } {
  if (range === "all") return {};
  const days = parseInt(range.replace("d", ""), 10);
  const to = new Date();
  const from = new Date(Date.now() - days * 86400000);
  return {
    from: from.toISOString().slice(0, 10),
    to: to.toISOString().slice(0, 10),
  };
}

const selectedRange = ref("90d");
const dateFrom = ref<string | undefined>();
const dateTo = ref<string | undefined>();

function setRange(range: string) {
  selectedRange.value = range;
  const { from, to } = getDateRange(range);
  dateFrom.value = from;
  dateTo.value = to;
  void load();
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    analytics.value = await defectAnalyticsApi.getAnalytics(wsId.value, projectId.value, {
      date_from: dateFrom.value,
      date_to: dateTo.value,
    });
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function handleExport() {
  if (!wsId.value) return;
  try {
    const blob = await defectAnalyticsApi.exportDefects(wsId.value, projectId.value, {
      date_from: dateFrom.value,
      date_to: dateTo.value,
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `defect-analytics-${projectId.value}-${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  } catch {
    toast.error("导出失败");
  }
}

onMounted(() => {
  setRange("90d");
});

watch([wsId], () => {
  if (wsId.value) setRange(selectedRange.value);
});

// --- 图表构造函数 ---

const SEVERITY_COLORS = ["#dc2626", "#ea580c", "#d97706", "#ca8a04", "#65a30d"];

function severityChartOption(data: SeverityCount[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "item", formatter: "{b}: {c} ({d}%)" },
    legend: { orient: "horizontal", bottom: 0, textStyle: { color: "var(--text-secondary)", fontSize: 12 } },
    series: [{
      type: "pie",
      radius: ["35%", "65%"],
      center: ["50%", "42%"],
      itemStyle: { borderRadius: 4, borderColor: "var(--surface-1)", borderWidth: 2 },
      label: { show: false },
      data: data.map((d, i) => ({
        name: d.label || SEVERITY_LABELS[d.severity] || `L${d.severity}`,
        value: d.count,
        itemStyle: { color: SEVERITY_COLORS[i % SEVERITY_COLORS.length] },
      })),
    }],
  };
}

function phaseChartOption(data: PhaseCount[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    grid: { left: 90, right: 24, top: 16, bottom: 24 },
    xAxis: { type: "value", axisLabel: { color: "var(--text-secondary)" } },
    yAxis: {
      type: "category",
      data: data.map((d) => PHASE_LABELS[d.phase] || d.phase),
      axisLabel: { color: "var(--text-secondary)" },
    },
    series: [{
      type: "bar",
      data: data.map((d) => d.count),
      itemStyle: { color: "var(--chart-1)", borderRadius: [0, 4, 4, 0] },
      barWidth: 16,
    }],
  };
}

function moduleChartOption(data: ModuleCount[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    grid: { left: 90, right: 24, top: 16, bottom: 24 },
    xAxis: { type: "value", axisLabel: { color: "var(--text-secondary)" } },
    yAxis: {
      type: "category",
      data: data.map((d) => d.module_name || "未分类"),
      axisLabel: { color: "var(--text-secondary)" },
    },
    series: [{
      type: "bar",
      data: data.map((d) => d.count),
      itemStyle: { color: "var(--chart-2)", borderRadius: [0, 4, 4, 0] },
      barWidth: 16,
    }],
  };
}

function rootCauseChartOption(data: RootCauseCount[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "item", formatter: "{b}: {c} ({d}%)" },
    legend: { orient: "horizontal", bottom: 0, textStyle: { color: "var(--text-secondary)", fontSize: 12 } },
    series: [{
      type: "pie",
      radius: ["35%", "65%"],
      center: ["50%", "42%"],
      itemStyle: { borderRadius: 4, borderColor: "var(--surface-1)", borderWidth: 2 },
      label: { show: false },
      data: data.map((d, i) => ({
        name: d.root_cause || "未分类",
        value: d.count,
        itemStyle: { color: `var(--chart-${(i % 6) + 1})` },
      })),
    }],
  };
}

function ageChartOption(data: AgeBucket[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    grid: { left: 40, right: 24, top: 16, bottom: 32 },
    xAxis: {
      type: "category",
      data: data.map((d) => d.range),
      axisLabel: { color: "var(--text-secondary)", fontSize: 11 },
    },
    yAxis: { type: "value", axisLabel: { color: "var(--text-secondary)" } },
    series: [{
      type: "bar",
      data: data.map((d) => d.count),
      itemStyle: { color: "var(--chart-3)", borderRadius: [4, 4, 0, 0] },
      barWidth: "60%",
    }],
  };
}

function trendChartOption(data: TrendPoint[]): EChartsCoreOption {
  if (!data.length) return emptyChartOption;
  return {
    tooltip: { trigger: "axis" },
    legend: { data: ["新增", "解决"], top: 0, textStyle: { color: "var(--text-secondary)" } },
    grid: { left: 40, right: 24, top: 32, bottom: 32 },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: data.map((d) => d.date.slice(5)),
      axisLabel: { color: "var(--text-secondary)", fontSize: 11 },
    },
    yAxis: { type: "value", axisLabel: { color: "var(--text-secondary)" } },
    series: [
      { name: "新增", type: "line", smooth: true, data: data.map((d) => d.opened), itemStyle: { color: "#dc2626" }, areaStyle: { opacity: 0.1 } },
      { name: "解决", type: "line", smooth: true, data: data.map((d) => d.resolved), itemStyle: { color: "#16a34a" }, areaStyle: { opacity: 0.1 } },
    ],
  };
}

const emptyChartOption: EChartsCoreOption = {
  title: { text: "暂无数据", left: "center", top: "center", textStyle: { color: "var(--text-tertiary)", fontSize: 13 } },
};
</script>

<template>
  <div class="defect-analytics">
    <!-- 顶部栏 -->
    <header class="page-header">
      <div>
        <h1 class="page-title">缺陷分析</h1>
        <p class="page-subtitle">多维度缺陷统计与趋势分析</p>
      </div>
      <div class="header-actions">
        <!-- 时间范围选择 -->
        <div class="range-tabs">
          <button
            v-for="r in TIME_RANGES"
            :key="r.value"
            :class="['range-tab', { active: selectedRange === r.value }]"
            @click="setRange(r.value)"
          >
            {{ r.label }}
          </button>
        </div>
        <!-- 导出按钮 -->
        <button class="btn-secondary" :disabled="loading || !analytics" @click="handleExport">
          导出 CSV
        </button>
      </div>
    </header>

    <!-- 加载中 -->
    <AppLoadingState v-if="loading" />
    <!-- 加载错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <!-- 空数据 -->
    <AppEmptyState v-else-if="analytics && analytics.total_defects === 0" description="当前时间范围内无缺陷数据" />
    <!-- 图表内容 -->
    <template v-else-if="analytics">
      <!-- 关键指标卡片 -->
      <section class="kpi-grid">
        <div class="kpi-card">
          <span class="kpi-label">缺陷总数</span>
          <span class="kpi-value">{{ analytics.total_defects }}</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">未解决</span>
          <span class="kpi-value text-danger">{{ analytics.open_defects }}</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">已解决</span>
          <span class="kpi-value text-success">{{ analytics.resolved_defects }}</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">平均缺陷龄</span>
          <span class="kpi-value">{{ analytics.avg_age_days.toFixed(1) }}<small>天</small></span>
        </div>
      </section>

      <!-- 趋势折线图 -->
      <section class="chart-card chart-wide">
        <h3 class="chart-title">缺陷趋势（新增 vs 解决）</h3>
        <ChartWidget :option="trendChartOption(analytics.trend)" :height="280" />
      </section>

      <!-- 2×2 网格：严重程度、发现阶段、模块、根因 -->
      <section class="chart-grid">
        <div class="chart-card">
          <h3 class="chart-title">严重程度分布</h3>
          <ChartWidget :option="severityChartOption(analytics.severity_dist)" :height="260" />
        </div>
        <div class="chart-card">
          <h3 class="chart-title">发现阶段分布</h3>
          <ChartWidget :option="phaseChartOption(analytics.phase_dist)" :height="260" />
        </div>
        <div class="chart-card">
          <h3 class="chart-title">模块分布</h3>
          <ChartWidget :option="moduleChartOption(analytics.module_dist)" :height="260" />
        </div>
        <div class="chart-card">
          <h3 class="chart-title">根因分类</h3>
          <ChartWidget :option="rootCauseChartOption(analytics.root_cause_dist)" :height="260" />
        </div>
      </section>

      <!-- 缺陷龄分布 -->
      <section class="chart-card chart-wide">
        <h3 class="chart-title">缺陷龄分布（天）</h3>
        <ChartWidget :option="ageChartOption(analytics.age_buckets)" :height="260" />
      </section>
    </template>
  </div>
</template>

<style scoped>
.defect-analytics {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.range-tabs {
  display: flex;
  background: var(--surface-2);
  border-radius: 6px;
  padding: 2px;
}

.range-tab {
  padding: 5px 12px;
  font-size: 12px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.range-tab.active {
  background: var(--surface-1);
  color: var(--text-primary);
  font-weight: 500;
}

.btn-secondary {
  padding: 6px 14px;
  font-size: 13px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  color: var(--text-primary);
  border-radius: 6px;
  cursor: pointer;
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* KPI 卡片 */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.kpi-card {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.kpi-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.kpi-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.kpi-value small {
  font-size: 13px;
  font-weight: 400;
  margin-left: 2px;
}

.text-danger { color: #dc2626; }
.text-success { color: #16a34a; }

/* 图表卡片 */
.chart-card {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 16px;
}

.chart-wide {
  width: 100%;
}

.chart-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.chart-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0 0 8px;
}

/* 响应式 */
@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .chart-grid { grid-template-columns: 1fr; }
}
</style>
