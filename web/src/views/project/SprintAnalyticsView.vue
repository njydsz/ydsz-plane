<script setup lang="ts">
/**
 * SprintAnalyticsView — 迭代效能分析综合面板
 *
 * 能力：
 *  - 燃尽图 / 燃起图切换
 *  - 迭代速率（近期迭代完成点数趋势）
 *  - 工作项状态分布（按状态组）
 *  - 汇总指标：完成率、速率、范围变化
 *
 * 数据由父组件（SprintDetailView）加载后传入，避免重复请求。
 */

import { computed, ref } from "vue";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { LineChart, BarChart, PieChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";

import type { BurndownPoint, Sprint, SprintVelocity } from "@/api/services/sprint";
import { AppCard } from "@/components";

// ---- ECharts 按需注册 ----
use([LineChart, BarChart, PieChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent, CanvasRenderer]);

/* ------------------------------------------------------------------ */
/*  Props                                                              */
/* ------------------------------------------------------------------ */
const props = defineProps<{
  sprint: Sprint | null;
  burndown: BurndownPoint[];
  velocity: SprintVelocity[];
  velocityAvg: number;
}>();

const chartMode = ref<"burndown" | "burnup">("burndown");

/* ------------------------------------------------------------------ */
/*  指标计算                                                           */
/* ------------------------------------------------------------------ */
const progress = computed(() => props.sprint?.progress);
const totalPoints = computed(() => progress.value?.total_points ?? 0);
const donePoints = computed(() => progress.value?.done_points ?? 0);
const completionRate = computed(() => {
  if (!totalPoints.value) return 0;
  return Math.round((donePoints.value / totalPoints.value) * 100);
});
const totalIssues = computed(() => progress.value?.total_issues ?? 0);
const doneIssues = computed(() => progress.value?.done_issues ?? 0);
/** 范围燃尽图理想终点 */
const scopeChange = computed(() => {
  if (props.burndown.length < 2) return 0;
  return props.burndown[props.burndown.length - 1].total_points - props.burndown[0].total_points;
});

/** 中途加入故事点（速率影响） */
const addedPoints = computed(() => props.sprint?.progress?.added_points ?? 0);
/** 承诺故事点 = 总点数 - 中途加入点数 */
const committedPoints = computed(() => Math.max(totalPoints.value - addedPoints.value, 0));
/** 承诺完成率（排除中途加入的影响） */
const committedCompletionRate = computed(() => {
  if (!committedPoints.value) return 0;
  return Math.round((donePoints.value / committedPoints.value) * 100);
});

/** 燃起图数据 */
const burnupPoints = computed(() => props.burndown.map((p) => ({
  date: p.date,
  done_points: p.done_points,
  total_points: p.total_points,
  ideal_line: p.total_points > 0
    ? Math.round(p.total_points * ((props.burndown.indexOf(p) + 1) / props.burndown.length))
    : 0,
})));

/** 状态组分布数据 */
const stateDistribution = computed(() => {
  const bg = progress.value?.by_state_group;
  if (!bg) return [];
  const labelMap: Record<string, string> = { backlog: "待处理", started: "进行中", completed: "已完成", cancelled: "已取消" };
  const colorMap: Record<string, string> = { backlog: "#6b7280", started: "#f59e0b", completed: "#10b981", cancelled: "#ef4444" };
  return Object.entries(bg).map(([k, v]) => ({
    name: labelMap[k] || k,
    value: v,
    itemStyle: { color: colorMap[k] || "#999" },
  }));
});

/* ------------------------------------------------------------------ */
/*  日期格式化                                                           */
/* ------------------------------------------------------------------ */
function fmtDate(d: string): string {
  if (!d) return "";
  const parts = d.split("-");
  return parts.length >= 3 ? `${parseInt(parts[1])}/${parseInt(parts[2])}` : d;
}

/* ------------------------------------------------------------------ */
/*  燃尽图配置                                                         */
/* ------------------------------------------------------------------ */
const burndownOption = computed(() => {
  if (props.burndown.length === 0) return {};
  const dates = props.burndown.map((p) => fmtDate(p.date));
  return {
    tooltip: { trigger: "axis" },
    legend: { data: ["理想线", "剩余点数", "已完成点数"], bottom: 0, textStyle: { fontSize: 11 } },
    grid: { left: "3%", right: "4%", top: "8%", bottom: "14%", containLabel: true },
    xAxis: { type: "category", data: dates, boundaryGap: false, axisLabel: { fontSize: 10, color: "#999" } },
    yAxis: { type: "value", name: "pt", axisLabel: { fontSize: 10, color: "#999" }, splitLine: { lineStyle: { type: "dashed", color: "#f0f0f0" } } },
    series: [
      { name: "理想线", type: "line", data: props.burndown.map((p) => p.ideal_line), lineStyle: { type: "dashed", color: "#faad14" }, symbol: "none" },
      { name: "剩余点数", type: "line", data: props.burndown.map((p) => p.remaining), lineStyle: { color: "#3b82f6", width: 2 }, itemStyle: { color: "#3b82f6" }, symbol: "circle", symbolSize: 5 },
      { name: "已完成点数", type: "line", data: props.burndown.map((p) => p.done_points), lineStyle: { color: "#10b981", width: 2 }, itemStyle: { color: "#10b981" }, symbol: "diamond", symbolSize: 5 },
    ],
  };
});

/* ------------------------------------------------------------------ */
/*  燃起图配置                                                         */
/* ------------------------------------------------------------------ */
const burnupOption = computed(() => {
  if (burnupPoints.value.length === 0) return {};
  const dates = burnupPoints.value.map((p) => fmtDate(p.date));
  const maxTotal = Math.max(...burnupPoints.value.map((p) => p.total_points), 1);
  return {
    tooltip: { trigger: "axis" },
    legend: { data: ["总量", "已完成", "理想线"], bottom: 0, textStyle: { fontSize: 11 } },
    grid: { left: "3%", right: "4%", top: "8%", bottom: "14%", containLabel: true },
    xAxis: { type: "category", data: dates, boundaryGap: false, axisLabel: { fontSize: 10, color: "#999" } },
    yAxis: { type: "value", name: "pt", max: Math.ceil(maxTotal * 1.1), axisLabel: { fontSize: 10, color: "#999" }, splitLine: { lineStyle: { type: "dashed", color: "#f0f0f0" } } },
    series: [
      { name: "总量", type: "line", data: burnupPoints.value.map((p) => p.total_points), lineStyle: { color: "#6b7280", width: 2 }, itemStyle: { color: "#6b7280" }, symbol: "none", areaStyle: { color: "rgba(107,114,128,0.05)" } },
      { name: "已完成", type: "line", data: burnupPoints.value.map((p) => p.done_points), lineStyle: { color: "#3b82f6", width: 2.5 }, itemStyle: { color: "#3b82f6" }, symbol: "circle", symbolSize: 5, areaStyle: { color: "rgba(59,130,246,0.12)" } },
      { name: "理想线", type: "line", data: burnupPoints.value.map((p) => p.ideal_line), lineStyle: { type: "dashed", color: "#faad14" }, symbol: "none" },
    ],
  };
});

/* ------------------------------------------------------------------ */
/*  速率图配置                                                         */
/* ------------------------------------------------------------------ */
const velocityOption = computed(() => {
  if (props.velocity.length === 0) return {};
  return {
    tooltip: { trigger: "axis" },
    grid: { left: "3%", right: "4%", top: "10%", bottom: "10%", containLabel: true },
    xAxis: {
      type: "category",
      data: props.velocity.map((v) => v.sprint_name),
      axisLabel: { fontSize: 10, color: "#999", rotate: 20 },
    },
    yAxis: { type: "value", name: "故事点", axisLabel: { fontSize: 10, color: "#999" }, splitLine: { lineStyle: { type: "dashed", color: "#f0f0f0" } } },
    series: [
      {
        type: "bar",
        data: props.velocity.map((v) => v.completed_points),
        barWidth: "50%",
        itemStyle: {
          color: { type: "linear", x: 0, y: 0, x2: 0, y2: 1, colorStops: [
            { offset: 0, color: "#3b82f6" },
            { offset: 1, color: "#93c5fd" },
          ] },
          borderRadius: [4, 4, 0, 0],
        },
        markLine: {
          data: [{ type: "average", name: "平均" }],
          lineStyle: { color: "#f59e0b", type: "dashed" },
          label: { formatter: "平均 {c}pt", position: "insideEndTop" },
        },
      },
    ],
  };
});

/* ------------------------------------------------------------------ */
/*  状态分布饼图                                                       */
/* ------------------------------------------------------------------ */
const pieOption = computed(() => {
  if (stateDistribution.value.length === 0) return {};
  return {
    tooltip: { trigger: "item", formatter: "{b}: {c} 项 ({d}%)" },
    legend: { bottom: 0, textStyle: { fontSize: 11 } },
    series: [
      {
        type: "pie",
        radius: ["45%", "70%"],
        center: ["50%", "45%"],
        avoidLabelOverlap: true,
        label: { show: true, formatter: "{b}\n{d}%", fontSize: 11 },
        labelLine: { length: 8, length2: 8 },
        data: stateDistribution.value,
      },
    ],
  };
});
</script>

<template>
  <div class="sprint-analytics">
    <!-- 指标卡片 -->
    <div class="metrics-grid">
      <div class="metric-card">
        <span class="metric-label">完成率</span>
        <span class="metric-value">{{ completionRate }}<small>%</small></span>
        <span class="metric-detail">{{ doneIssues }}/{{ totalIssues }} 项工作项</span>
      </div>
      <div class="metric-card">
        <span class="metric-label">已完成点数</span>
        <span class="metric-value">{{ donePoints }}<small>pt</small></span>
        <span class="metric-detail">总计 {{ totalPoints }} 故事点</span>
      </div>
      <div class="metric-card">
        <span class="metric-label">平均速率</span>
        <span class="metric-value">{{ velocityAvg }}<small>pt/迭代</small></span>
        <span class="metric-detail">{{ props.velocity.length }} 个近期迭代</span>
      </div>
      <div class="metric-card" :class="{ 'metric-card--warn': scopeChange > 0 }">
        <span class="metric-label">范围变化</span>
        <span class="metric-value">{{ scopeChange >= 0 ? '+' : '' }}{{ scopeChange }}<small>pt</small></span>
        <span class="metric-detail">{{ scopeChange > 0 ? '范围扩大' : scopeChange < 0 ? '范围缩小' : '无变化' }}</span>
      </div>
      <div v-if="addedPoints > 0" class="metric-card metric-card--midway">
        <span class="metric-label">中途加入 <small>(速率影响)</small></span>
        <span class="metric-value">+{{ addedPoints }}<small>pt</small></span>
        <span class="metric-detail">承诺 {{ committedPoints }}pt · 承诺完成率 {{ committedCompletionRate }}%</span>
      </div>
    </div>

    <!-- 燃尽/燃起图切换 -->
    <AppCard padding="md" shadow>
      <div class="chart-header">
        <h3 class="chart-title">进度趋势</h3>
        <div class="chart-toggle">
          <button
            class="toggle-btn"
            :class="{ 'toggle-btn--active': chartMode === 'burndown' }"
            @click="chartMode = 'burndown'"
          >
            燃尽图
          </button>
          <button
            class="toggle-btn"
            :class="{ 'toggle-btn--active': chartMode === 'burnup' }"
            @click="chartMode = 'burnup'"
          >
            燃起图
          </button>
        </div>
      </div>
      <div class="chart-container">
        <v-chart v-if="chartMode === 'burndown' && props.burndown.length > 0" :option="burndownOption" style="height: 320px; width: 100%;" autoresize />
        <v-chart v-else-if="props.burndown.length > 0" :option="burnupOption" style="height: 320px; width: 100%;" autoresize />
        <div v-else class="chart-empty">暂无快照数据</div>
      </div>
    </AppCard>

    <!-- 速率 + 状态分布 -->
    <div class="analytics-row">
      <AppCard padding="md" shadow class="chart-card--half">
        <h3 class="chart-title">迭代速率</h3>
        <div v-if="props.velocity.length > 0" class="chart-container">
          <v-chart :option="velocityOption" style="height: 280px; width: 100%;" autoresize />
        </div>
        <div v-else class="chart-empty">暂无速率数据</div>
      </AppCard>

      <AppCard padding="md" shadow class="chart-card--half">
        <h3 class="chart-title">状态分布</h3>
        <div v-if="stateDistribution.length > 0" class="chart-container">
          <v-chart :option="pieOption" style="height: 280px; width: 100%;" autoresize />
        </div>
        <div v-else class="chart-empty">暂无状态数据</div>
      </AppCard>
    </div>
  </div>
</template>

<style scoped>
.sprint-analytics {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ---- Metrics Grid ---- */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

@media (max-width: 960px) {
  .metrics-grid { grid-template-columns: repeat(2, 1fr); }
}

.metric-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 16px;
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-subtle, #f0f0f0);
  border-radius: var(--radius-md, 8px);
}

.metric-card--warn .metric-value { color: var(--warning-500, #f59e0b); }
.metric-card--midway { background: var(--warning-50, #fffbeb); border-color: var(--warning-200, #fde68a); }
.metric-card--midway .metric-value { color: var(--warning-600, #d97706); }
.metric-card--midway .metric-label small { font-weight: 400; opacity: 0.7; }

.metric-label {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary, #1f2937);
}

.metric-value small {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-tertiary, #9ca3af);
  margin-left: 2px;
}

.metric-detail {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

/* ---- Chart Header ---- */
.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.chart-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0;
}

.chart-toggle {
  display: flex;
  gap: 0;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: 6px;
  overflow: hidden;
}

.toggle-btn {
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 500;
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--text-secondary, #6b7280);
  font-family: inherit;
  transition: all 0.15s;
}

.toggle-btn--active {
  background: var(--brand-default, #3b82f6);
  color: #fff;
}

.toggle-btn:not(.toggle-btn--active):hover {
  background: var(--surface-2, #f9fafb);
}

/* ---- Chart Container ---- */
.chart-container {
  width: 100%;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-tertiary, #9ca3af);
  font-size: 13px;
}

/* ---- Row ---- */
.analytics-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 768px) {
  .analytics-row { grid-template-columns: 1fr; }
}

.chart-card--half {
  min-width: 0;
}
</style>
