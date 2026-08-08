<script setup lang="ts">
/**
 * MetricsView — 效能度量仪表盘。
 *
 * 聚合 DORA 四指标 + 速度 + 前置时间 + 质量 + 资源负载，
 * 使用 Card Grid 展示，每个卡片含：标题、当前值、等级标签、趋势指标。
 *
 * 大厂对标：Linear 的 Insights、GitLab 的 Vortex 度量页、
 * 字节飞书的项目效能面板。原则是「一屏看全局、重点指标突出、异常红色预警」。
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { BarChart, LineChart, ScatterChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { EChartsOption } from "echarts";

import {
  metricsApi,
  type CFDDataPoint,
  type ControlChartResult,
  type DORAResult,
  type LeadTimeResult,
  type QualityMetrics,
  type ResourceLoadResult,
  type VelocityResult,
  type WeeklyThroughput,
} from "@/api/services/metrics";
import { AppLoadingState, AppErrorState, AppCard, AppBadge } from "@/components";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";

/* ---- ECharts 按需注册（tree-shaking） ---- */
use([
  BarChart,
  LineChart,
  ScatterChart,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  TooltipComponent,
  CanvasRenderer,
]);

const route = useRoute();
const { wsId, ready } = useWorkspaceContext();

const projectId = computed(() => Number(route.params.projectId));

// -------- 状态 --------
const loading = ref(true);
const error = ref("");
const velocity = ref<VelocityResult | null>(null);
const leadTime = ref<LeadTimeResult | null>(null);
const quality = ref<QualityMetrics | null>(null);
const dora = ref<DORAResult | null>(null);
const resource = ref<ResourceLoadResult | null>(null);
const cfd = ref<CFDDataPoint[] | null>(null);
const controlChart = ref<ControlChartResult | null>(null);
const throughput = ref<WeeklyThroughput[] | null>(null);

// -------- DORA 等级 → variant 映射 --------
const levelVariant: Record<string, "success" | "info" | "warning" | "danger"> = {
  elite: "success",
  high: "info",
  medium: "warning",
  low: "danger",
};
const levelLabels: Record<string, string> = {
  elite: "精英",
  high: "高",
  medium: "中",
  low: "低",
};

// --------
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const ws = wsId.value;
    const pid = projectId.value;

    const [v, lt, q, d, r, c, cc, tp] = await Promise.all([
      metricsApi.getVelocity(ws, pid).catch(() => null),
      metricsApi.getLeadTime(ws, pid).catch(() => null),
      metricsApi.getQuality(ws, pid).catch(() => null),
      metricsApi.getDORA(ws, pid).catch(() => null),
      metricsApi.getResourceLoad(ws, pid).catch(() => null),
      metricsApi.getCFD(ws, pid).catch(() => null),
      metricsApi.getControlChart(ws, pid).catch(() => null),
      metricsApi.getWeeklyThroughput(ws, pid).catch(() => null),
    ]);
    velocity.value = v;
    leadTime.value = lt;
    quality.value = q;
    dora.value = d;
    resource.value = r;
    cfd.value = c;
    controlChart.value = cc;
    throughput.value = tp;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (ready.value) void load();
});
watch(ready, (r) => {
  if (r) void load();
});

/** 速度趋势方向：最近迭代 vs 平均 */
const velocityTrend = computed(() => {
  if (!velocity.value || velocity.value.trend.length === 0) return null;
  const latest = velocity.value.trend[velocity.value.trend.length - 1];
  const diff = latest.completed_count - velocity.value.average;
  return { latest: latest.completed_count, diff, average: velocity.value.average };
});

/** 前置时间天数（p50） */
const leadTimeDays = computed(() => {
  if (!leadTime.value) return null;
  return (leadTime.value.percentiles.p50_hours / 24).toFixed(1);
});

/** 质量综合评分（简化的启发式：0-100） */
const qualityScore = computed(() => {
  if (!quality.value) return null;
  // 缺陷密度越低越好（上限 5 个/百点），逃逸率越低越好，重开率越低越好
  const densityScore = Math.max(0, 100 - quality.value.defect_density * 20);
  const escapeScore = Math.max(0, 100 - quality.value.escape_rate * 100);
  const reopenScore = Math.max(0, 100 - quality.value.reopen_rate * 100);
  return Math.round((densityScore + escapeScore + reopenScore) / 3);
});

/* ================= 分析图（CFD / 控制图 / 周吞吐量） ================= */

/** CFD 累积流图：按状态组堆叠面积，直观反映在制品积压与交付趋势。 */
const cfdOption = computed<EChartsOption>(() => {
  const points = cfd.value ?? [];
  const dates = points.map((p) => p.date);
  const series = [
    { name: "已完成", key: "done" as const, color: "#10B981" },
    { name: "进行中", key: "in_progress" as const, color: "#F59E0B" },
    { name: "待办", key: "todo" as const, color: "#A2B8D8" },
    { name: "积压", key: "backlog" as const, color: "#8DA2C2" },
    { name: "已取消", key: "cancelled" as const, color: "#9CA3AF" },
  ];
  return {
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" },
    legend: { bottom: 0 },
    grid: { left: 8, right: 16, top: 24, bottom: 40, containLabel: true },
    xAxis: { type: "category", data: dates, boundaryGap: false },
    yAxis: { type: "value", minInterval: 1 },
    series: series.map((s) => ({
      name: s.name,
      type: "line",
      stack: "cfd",
      smooth: true,
      showSymbol: false,
      areaStyle: { opacity: 0.75 },
      lineStyle: { width: 1.5 },
      itemStyle: { color: s.color },
      emphasis: { focus: "series" },
      data: points.map((p) => p[s.key]),
    })),
  };
});

/** 前置时间控制图：散点（单个工作项）+ P50/P85/P95/UCL 参考线 + 7 点移动均线。 */
const controlChartOption = computed<EChartsOption>(() => {
  const cc = controlChart.value;
  if (!cc) return {};
  const days = cc.points.map((p) => p.date);
  const ma = cc.moving_avg_7d.map((m) => m.value);
  const lines = [
    { label: `P50 ${cc.p50.toFixed(1)}d`, value: cc.p50, color: "#3B82F6" },
    { label: `P85 ${cc.p85.toFixed(1)}d`, value: cc.p85, color: "#8B5CF6" },
    { label: `P95 ${cc.p95.toFixed(1)}d`, value: cc.p95, color: "#EF4444" },
    { label: `UCL ${cc.upper_control_limit.toFixed(1)}d`, value: cc.upper_control_limit, color: "#F97316" },
  ];
  return {
    backgroundColor: "transparent",
    tooltip: {
      trigger: "item",
      formatter: (p: any) =>
        `${p.marker}${p.name}：<b>${Number(p.value[1]).toFixed(1)} 天</b>`,
    },
    legend: { bottom: 0 },
    grid: { left: 8, right: 16, top: 24, bottom: 40, containLabel: true },
    xAxis: { type: "category", data: days },
    yAxis: { type: "value", name: "前置时间(天)" },
    series: [
      {
        name: "工作项前置时间",
        type: "scatter",
        symbolSize: 7,
        itemStyle: { color: "#94A3B8", opacity: 0.75 },
        data: cc.points.map((p, i) => [days[i], p.lead_days]),
      },
      {
        name: "7 点移动均线",
        type: "line",
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, color: "#F59E0B" },
        data: ma,
        z: 3,
      },
      ...lines.map((l) => ({
        name: l.label,
        type: "line" as const,
        markLine: {
          silent: true,
          symbol: "none" as const,
          label: { formatter: l.label, position: "insideEndTop" as const },
          lineStyle: { type: "dashed" as const, color: l.color },
          data: [{ yAxis: l.value }],
        },
        data: [] as number[],
      })),
    ],
  };
});

/** 周吞吐量：柱（完成需求数）+ 线（完成故事点）双轴组合。 */
const throughputOption = computed<EChartsOption>(() => {
  const rows = throughput.value ?? [];
  const labels = rows.map((r) => r.week_start.slice(5));
  return {
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" },
    legend: { bottom: 0 },
    grid: { left: 8, right: 8, top: 24, bottom: 40, containLabel: true },
    xAxis: { type: "category", data: labels },
    yAxis: [
      { type: "value", name: "需求数", minInterval: 1 },
      { type: "value", name: "故事点", splitLine: { show: false } },
    ],
    series: [
      {
        name: "完成需求数",
        type: "bar",
        barMaxWidth: 28,
        itemStyle: { color: "#3B82F6", borderRadius: [4, 4, 0, 0] },
        data: rows.map((r) => r.completed),
      },
      {
        name: "完成故事点",
        type: "line",
        yAxisIndex: 1,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, color: "#F59E0B" },
        itemStyle: { color: "#F59E0B" },
        data: rows.map((r) => r.points),
      },
    ],
  };
});
</script>

<template>
  <AppLoadingState v-if="loading" text="加载效能指标..." />
  <AppErrorState v-else-if="error" :message="error" @retry="load" />

  <div v-else class="metrics-dashboard">
    <header class="metrics-dashboard__header">
      <h1>效能度量</h1>
      <p class="subtitle">项目 DORA 指标与效能全景（近 90 天窗口）</p>
    </header>

    <!-- 无数据空态 -->
    <AppEmptyState
      v-if="!dora && !velocity && !leadTime && !quality && !resource"
      scenario="analytics"
      :cta-text="''"
      class="metrics-empty"
    />

    <!-- DORA 四指标 -->
    <template v-else>
      <section v-if="dora" class="card-grid card-grid--dora">
      <AppCard
        v-for="(metric, key) in {
          deployment_frequency: dora.deployment_frequency,
          lead_time_for_changes: dora.lead_time_for_changes,
          mean_time_to_restore: dora.mean_time_to_restore,
          change_failure_rate: dora.change_failure_rate,
        }"
        :key="key"
        class="metric-card"
        padding="sm"
      >
        <div class="metric-card__header">
          <span class="metric-card__title">
            {{
              {
                deployment_frequency: "部署频率",
                lead_time_for_changes: "变更前置时间",
                mean_time_to_restore: "故障恢复时间",
                change_failure_rate: "变更失败率",
              }[key]
            }}
          </span>
          <AppBadge :variant="levelVariant[metric.level]">
            {{ levelLabels[metric.level] }}
          </AppBadge>
        </div>
        <div class="metric-card__value">
          {{ metric.value }}
          <span class="metric-card__unit">{{ metric.unit }}</span>
        </div>
      </AppCard>
    </section>

    <!-- 速度 + 前置时间 + 质量 -->
    <section class="card-grid card-grid--secondary">
      <!-- 平均速度 -->
      <AppCard v-if="velocityTrend" class="metric-card" padding="sm">
        <div class="metric-card__header">
          <span class="metric-card__title">迭代速度</span>
          <AppBadge v-if="velocityTrend.diff > 0" variant="success">↑ 高于均值</AppBadge>
          <AppBadge v-else-if="velocityTrend.diff < 0" variant="warning">↓ 低于均值</AppBadge>
        </div>
        <div class="metric-card__value">
          {{ velocityTrend.latest }}
          <span class="metric-card__unit">点/迭代</span>
        </div>
        <div class="metric-card__sub">
          均值 {{ velocityTrend.average.toFixed(1) }} 点 · {{ velocity?.sprint_count ?? 0 }} 个迭代
        </div>
      </AppCard>

      <!-- 前置时间 -->
      <AppCard v-if="leadTimeDays" class="metric-card" padding="sm">
        <div class="metric-card__header">
          <span class="metric-card__title">前置时间（P50）</span>
        </div>
        <div class="metric-card__value">
          {{ leadTimeDays }}
          <span class="metric-card__unit">天</span>
        </div>
        <div class="metric-card__sub">
          P85 {{ (leadTime!.percentiles.p85_hours / 24).toFixed(1) }} 天 ·
          P95 {{ (leadTime!.percentiles.p95_hours / 24).toFixed(1) }} 天
        </div>
      </AppCard>

      <!-- 质量综合 -->
      <AppCard v-if="qualityScore !== null" class="metric-card" padding="sm">
        <div class="metric-card__header">
          <span class="metric-card__title">质量评分</span>
          <AppBadge :variant="qualityScore >= 80 ? 'success' : qualityScore >= 60 ? 'warning' : 'danger'">
            {{ qualityScore >= 80 ? "良好" : qualityScore >= 60 ? "一般" : "需改进" }}
          </AppBadge>
        </div>
        <div class="metric-card__value">
          {{ qualityScore }}
          <span class="metric-card__unit">/ 100</span>
        </div>
        <div class="metric-card__sub">
          缺陷密度 {{ quality!.defect_density }} · 逃逸率 {{ (quality!.escape_rate * 100).toFixed(0) }}%
        </div>
      </AppCard>

      <!-- 资源负载 -->
      <AppCard v-if="resource" class="metric-card" padding="sm">
        <div class="metric-card__header">
          <span class="metric-card__title">在制品（WIP）</span>
        </div>
        <div class="metric-card__value">
          {{ resource.active_wip }}
          <span class="metric-card__unit">项进行中</span>
        </div>
        <div class="metric-card__sub">
          总计 {{ resource.total_started_issues }} 项已启动
        </div>
      </AppCard>
    </section>

    <!-- ===== 分析图：CFD / 控制图 / 周吞吐量 ===== -->
    <section v-if="cfd" class="chart-block">
      <AppCard padding="sm">
        <div class="chart-block__header">
          <span>累积流图（CFD）</span>
          <span class="chart-block__hint">按状态组堆叠 · 悬停查看明细</span>
        </div>
        <VChart class="analytics-chart" :option="cfdOption" autoresize />
      </AppCard>
    </section>

    <section v-if="controlChart" class="chart-block">
      <AppCard padding="sm">
        <div class="chart-block__header">
          <span>前置时间控制图</span>
          <span class="chart-block__hint">P50/P85/P95 + UCL 控制线 · 超 UCL 触发预警</span>
        </div>
        <VChart class="analytics-chart" :option="controlChartOption" autoresize />
      </AppCard>
    </section>

    <section v-if="throughput?.length" class="chart-block">
      <AppCard padding="sm">
        <div class="chart-block__header">
          <span>周吞吐量</span>
          <span class="chart-block__hint">近 12 周 · 需求数与故事点双轴</span>
        </div>
        <VChart class="analytics-chart" :option="throughputOption" autoresize />
      </AppCard>
    </section>
    </template>
  </div>
</template>

<style scoped>
.metrics-dashboard {
  max-width: 1100px;
  margin: 0 auto;
}

.metrics-dashboard__header {
  margin-bottom: 24px;
}

.metrics-dashboard__header h1 {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
}

.subtitle {
  color: var(--text-tertiary);
  font-size: 13px;
  margin-top: 4px;
}

/* ===== Card Grid ===== */
.card-grid {
  display: grid;
  gap: 16px;
  margin-bottom: 24px;
}

.card-grid--dora {
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.card-grid--secondary {
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

/* ===== Metric Card ===== */
.metric-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.metric-card__title {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.metric-card__value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.metric-card__unit {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-tertiary);
  margin-left: 4px;
}

.metric-card__sub {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 8px;
}

/* ===== 分析图区块 ===== */
.chart-block {
  margin-bottom: 24px;
}

.chart-block__header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-block__hint {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary);
}

.analytics-chart {
  width: 100%;
  height: 320px;
  margin-top: 8px;
}

.empty {
  text-align: center;
  padding: 64px 0;
  color: var(--text-tertiary);
  font-size: 14px;
}
</style>
