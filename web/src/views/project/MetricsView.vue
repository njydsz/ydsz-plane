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

import {
  metricsApi,
  type DORAResult,
  type LeadTimeResult,
  type QualityMetrics,
  type ResourceLoadResult,
  type VelocityResult,
} from "@/api/services/metrics";
import { AppLoadingState, AppErrorState, AppCard, AppBadge } from "@/components";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";

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

    const [v, lt, q, d, r] = await Promise.all([
      metricsApi.getVelocity(ws, pid).catch(() => null),
      metricsApi.getLeadTime(ws, pid).catch(() => null),
      metricsApi.getQuality(ws, pid).catch(() => null),
      metricsApi.getDORA(ws, pid).catch(() => null),
      metricsApi.getResourceLoad(ws, pid).catch(() => null),
    ]);
    velocity.value = v;
    leadTime.value = lt;
    quality.value = q;
    dora.value = d;
    resource.value = r;
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

.empty {
  text-align: center;
  padding: 64px 0;
  color: var(--text-tertiary);
  font-size: 14px;
}
</style>
