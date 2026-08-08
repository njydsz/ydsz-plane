<script setup lang="ts">
/**
 * DoraWidget — 研发效能（DORA）四指标。
 *
 * 数据由组件自行从 metricsApi.getDORA 拉取（与 MetricsView 共用同一后端接口），
 * 展示部署频率 / 变更前置时间 / 变更失败率 / 故障恢复时间 四个指标数字与效能等级。
 */
import { onMounted, ref } from "vue";
import { metricsApi, type DORAResult } from "@/api/services/metrics";

const props = defineProps<{
  wsId?: number;
  projectId?: number;
  config?: Record<string, any>;
}>();

const loading = ref(true);
const error = ref("");
const dora = ref<DORAResult | null>(null);

const levelVariant: Record<string, string> = {
  elite: "dora-level--elite",
  high: "dora-level--high",
  medium: "dora-level--medium",
  low: "dora-level--low",
};
const levelLabels: Record<string, string> = {
  elite: "精英",
  high: "高",
  medium: "中",
  low: "低",
};

async function load() {
  if (!props.wsId || !props.projectId) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    dora.value = await metricsApi.getDORA(props.wsId, props.projectId);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function fmt(n: number, digits = 2): string {
  return Number.isFinite(n) ? n.toFixed(digits) : "0";
}

onMounted(load);
</script>

<template>
  <div class="dora-widget">
    <div v-if="loading" class="dora-widget__hint">加载中...</div>
    <div v-else-if="error" class="dora-widget__hint dora-widget__hint--error">{{ error }}</div>
    <template v-else-if="dora">
      <div class="dora-widget__level">
        <span class="dora-widget__level-label">效能等级</span>
        <span class="dora-level" :class="levelVariant[dora.performance_level] ?? levelVariant.low">
          {{ levelLabels[dora.performance_level] ?? dora.performance_level }}
        </span>
      </div>
      <div class="dora-widget__grid">
        <div class="dora-stat">
          <span class="dora-stat__label">部署频率</span>
          <span class="dora-stat__value">{{ fmt(dora.deployment_freq_per_day) }}<small>次/天</small></span>
        </div>
        <div class="dora-stat">
          <span class="dora-stat__label">变更前置时间</span>
          <span class="dora-stat__value">{{ fmt(dora.lead_time_for_changes_hours) }}<small>h</small></span>
        </div>
        <div class="dora-stat">
          <span class="dora-stat__label">变更失败率</span>
          <span class="dora-stat__value">{{ (dora.change_failure_rate * 100).toFixed(1) }}<small>%</small></span>
        </div>
        <div class="dora-stat">
          <span class="dora-stat__label">故障恢复时间</span>
          <span class="dora-stat__value">{{ fmt(dora.mttr_hours) }}<small>h</small></span>
        </div>
      </div>
    </template>
    <div v-else class="dora-widget__hint">暂无 DORA 数据</div>
  </div>
</template>

<style scoped>
.dora-widget {
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
}

.dora-widget__hint {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
}

.dora-widget__hint--error {
  color: var(--danger-500, #dc2f2f);
}

.dora-widget__level {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.dora-widget__level-label {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.dora-level {
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

.dora-level--elite {
  background: var(--success-50, #ecfdf5);
  color: var(--success-600, #059669);
}

.dora-level--high {
  background: var(--brand-50, #eef2fe);
  color: var(--brand-600, #2f4fd0);
}

.dora-level--medium {
  background: var(--warning-50, #fffbeb);
  color: var(--warning-600, #d97706);
}

.dora-level--low {
  background: var(--danger-50, #fef2f2);
  color: var(--danger-600, #dc2626);
}

.dora-widget__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.dora-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px;
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-2, #f7f8f9);
}

.dora-stat__label {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.dora-stat__value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.dora-stat__value small {
  font-size: 11px;
  font-weight: 400;
  color: var(--text-tertiary, #9ca3af);
  margin-left: 2px;
}
</style>
