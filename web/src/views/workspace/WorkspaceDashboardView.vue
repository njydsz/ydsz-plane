<script setup lang="ts">
/**
 * 工作空间仪表盘 — 跨项目汇总：项目完成率对比 + 风险告警 + 仪表盘模板。
 * 数据来源：dashboardApi（工作空间级方法）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  dashboardApi,
  type ProjectCompareItem,
  type RiskAlert,
} from "@/api/services/dashboard";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const compare = ref<ProjectCompareItem[]>([]);
const alerts = ref<RiskAlert[]>([]);

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    const [c, a] = await Promise.all([
      dashboardApi.getProjectCompare(workspaceId.value),
      dashboardApi.listWorkspaceAlerts(workspaceId.value),
    ]);
    compare.value = c ?? [];
    alerts.value = a ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function severityLabel(s: RiskAlert["severity"]): string {
  const map: Record<string, string> = {
    info: "提示", low: "低", medium: "中", high: "高", critical: "严重",
  };
  return map[s] ?? s;
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">仪表盘</h1>
      <button class="text-sm text-[var(--brand-600)] hover:underline" @click="load">刷新</button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 4" :key="i" class="h-20 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 项目对比 -->
      <section>
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">项目完成率</h2>
        <AppEmptyState
          v-if="compare.length === 0"
          title="暂无项目"
          description="创建项目后，各项目的完成率与缺陷统计将在此汇总。"
        />
        <div v-else class="space-y-3">
          <div
            v-for="p in compare"
            :key="p.project_id"
            class="rounded-md border border-[var(--border-subtle)] p-4"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="font-mono text-xs text-[var(--brand-600)]">{{ p.identifier }}</span>
                <span class="text-sm font-medium text-[var(--text-primary)]">{{ p.project_name }}</span>
              </div>
              <span class="text-sm font-medium text-[var(--text-primary)]">
                {{ Math.round(p.completion_rate * 100) }}%
              </span>
            </div>
            <div class="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-[var(--surface-2)]">
              <div
                class="h-full rounded-full bg-[var(--brand-600)]"
                :style="{ width: `${Math.min(100, p.completion_rate * 100)}%` }"
              />
            </div>
            <div class="mt-1 text-xs text-[var(--text-tertiary)]">
              {{ p.done_issues }}/{{ p.total_issues }} 完成 · 缺陷 {{ p.defect_count }} · 进行中迭代 {{ p.active_sprint_count }}
            </div>
          </div>
        </div>
      </section>

      <!-- 风险告警 -->
      <section>
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">风险告警</h2>
        <AppEmptyState
          v-if="alerts.length === 0"
          title="暂无告警"
          description="当前没有需要关注的风险。"
        />
        <div v-else class="space-y-2">
          <div
            v-for="a in alerts"
            :key="a.id"
            class="flex items-start gap-3 rounded-md border border-[var(--border-subtle)] px-4 py-3"
          >
            <span
              class="mt-1 h-2 w-2 shrink-0 rounded-full"
              :class="a.severity === 'critical' || a.severity === 'high'
                ? 'bg-[var(--danger, #ef4444)]'
                : a.severity === 'medium' ? 'bg-[var(--warning, #f59e0b)]' : 'bg-[var(--text-tertiary)]'"
            />
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-[var(--text-primary)]">{{ a.title }}</div>
              <div class="mt-0.5 text-xs text-[var(--text-secondary)]">{{ a.description }}</div>
            </div>
            <span class="text-xs text-[var(--text-tertiary)]">{{ severityLabel(a.severity) }}</span>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
