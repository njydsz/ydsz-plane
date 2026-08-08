<script setup lang="ts">
/**
 * WorkspaceDashboardView — 工作空间级仪表盘（跨项目汇总）。
 *
 * 展示跨项目风险告警、项目健康度总览。
 * 后端路由已在 RegisterDashboardRoutes 中注册为 workspace 级。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { dashboardApi, type RiskAlert } from "@/api/services/dashboard";
import { workspaceApi } from "@/api/services/workspace";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId));

const loading = ref(true);
const error = ref("");
const alerts = ref<RiskAlert[]>([]);
const resolving = ref<number | null>(null);

const wsName = ref("");

const unresolvedAlerts = computed(() => alerts.value.filter((a) => !a.is_resolved));
const resolvedAlerts = computed(() => alerts.value.filter((a) => a.is_resolved));

const severityClass: Record<string, string> = {
  info: "alert-badge--info",
  low: "alert-badge--low",
  medium: "alert-badge--medium",
  high: "alert-badge--high",
  critical: "alert-badge--critical",
};

const severityLabel: Record<string, string> = {
  info: "信息",
  low: "低",
  medium: "中",
  high: "高",
  critical: "紧急",
};

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [alertsRes, ws] = await Promise.all([
      dashboardApi.listWorkspaceAlerts(workspaceId.value),
      workspaceApi.get(workspaceId.value).catch(() => null),
    ]);
    alerts.value = alertsRes;
    wsName.value = ws?.name ?? "";
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载仪表盘失败";
  } finally {
    loading.value = false;
  }
}

async function resolveAlert(alertId: number) {
  resolving.value = alertId;
  try {
    await dashboardApi.resolveWorkspaceAlert(workspaceId.value, alertId);
    toast.success("告警已解决");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "解决告警失败");
  } finally {
    resolving.value = null;
  }
}

function fmtDate(s: string) {
  return s ? s.slice(0, 10) + " " + s.slice(11, 16) : "—";
}

onMounted(() => void load());
</script>

<template>
  <div class="ws-dashboard">
    <header class="header">
      <h1>{{ wsName || "工作空间" }} 仪表盘</h1>
      <p class="subtitle">跨项目风险告警与健康度总览</p>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <div v-else class="content">
      <!-- 统计卡片 -->
      <div class="stats-row">
        <div class="stat-card">
          <span class="stat-value stat-value--warn">{{ unresolvedAlerts.length }}</span>
          <span class="stat-label">未解决告警</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ alerts.length }}</span>
          <span class="stat-label">告警总数</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ resolvedAlerts.length }}</span>
          <span class="stat-label">已解决</span>
        </div>
      </div>

      <!-- 未解决告警 -->
      <section class="alerts-section">
        <h2>活跃告警 <span class="count">({{ unresolvedAlerts.length }})</span></h2>

        <AppEmptyState
          v-if="unresolvedAlerts.length === 0"
          title="暂无活跃告警"
          description="工作空间运行良好"
        />

        <div v-else class="alert-list">
          <div v-for="alert in unresolvedAlerts" :key="alert.id" class="alert-card">
            <div class="alert-card__header">
              <span class="alert-badge" :class="severityClass[alert.severity]">
                {{ severityLabel[alert.severity] || alert.severity }}
              </span>
              <span class="alert-title">{{ alert.title }}</span>
            </div>
            <p v-if="alert.description" class="alert-desc">{{ alert.description }}</p>
            <div class="alert-card__footer">
              <span class="alert-time">{{ fmtDate(alert.created_at) }}</span>
              <button
                class="resolve-btn"
                :disabled="resolving === alert.id"
                @click="resolveAlert(alert.id)"
              >
                {{ resolving === alert.id ? "处理中..." : "标记已解决" }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- 已解决告警 -->
      <section v-if="resolvedAlerts.length > 0" class="alerts-section">
        <h2>已解决 <span class="count">({{ resolvedAlerts.length }})</span></h2>
        <div class="alert-list">
          <div v-for="alert in resolvedAlerts" :key="alert.id" class="alert-card alert-card--resolved">
            <div class="alert-card__header">
              <span class="alert-badge alert-badge--resolved">已解决</span>
              <span class="alert-title">{{ alert.title }}</span>
            </div>
            <div class="alert-card__footer">
              <span class="alert-time">{{ fmtDate(alert.created_at) }}</span>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.ws-dashboard {
  max-width: 900px;
  margin: 0 auto;
}

.header {
  margin-bottom: 24px;
}

.header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Stats */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-value--warn {
  color: var(--warning-600, #d97706);
}

.stat-label {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

/* Alerts */
.alerts-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alerts-section h2 {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}

.count {
  color: var(--text-tertiary);
  font-weight: 400;
}

.alert-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-card {
  padding: 12px 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.alert-card--resolved {
  opacity: 0.7;
}

.alert-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.alert-title {
  font-weight: 500;
  font-size: 13px;
}

.alert-desc {
  margin: 4px 0 8px;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.alert-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
}

.alert-time {
  color: var(--text-tertiary);
}

.resolve-btn {
  padding: 2px 10px;
  font-size: 11px;
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--text-primary);
  font-family: inherit;
}

.resolve-btn:hover:not(:disabled) {
  border-color: var(--success-500, #10b981);
  color: var(--success-600, #059669);
}

.resolve-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Alert badges */
.alert-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 3px;
  flex-shrink: 0;
}

.alert-badge--info { background: #eff6ff; color: #3b82f6; }
.alert-badge--low { background: #f0fdf4; color: #16a34a; }
.alert-badge--medium { background: #fffbeb; color: #d97706; }
.alert-badge--high { background: #fef2f2; color: #dc2626; }
.alert-badge--critical { background: #dc2626; color: #fff; }
.alert-badge--resolved { background: var(--surface-2); color: var(--text-tertiary); }
</style>
