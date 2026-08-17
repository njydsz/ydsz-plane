<script setup lang="ts">
/**
 * 工作台 — 个人效率首页。
 * 聚合「我的需求/任务/缺陷」分桶 + 迭代概览 + 最近访问 + 周效率趋势。
 * 数据来源：workbenchApi.getSummary / getEfficiency（后端 workbench 域）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  workbenchApi,
  type WorkbenchSummary,
  type EfficiencyReport,
} from "@/api/services/workbench";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const summary = ref<WorkbenchSummary | null>(null);
const efficiency = ref<EfficiencyReport | null>(null);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [s, e] = await Promise.all([
      workbenchApi.getSummary(workspaceId.value),
      workbenchApi.getEfficiency(workspaceId.value).catch(() => null),
    ]);
    summary.value = s;
    efficiency.value = e;
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function openIssue() {
  // IssueDigest 不含 project_id，无法定位到具体项目内详情；跳到我的需求/任务/缺陷列表兜底
  router.push(`/${workspaceId.value}/my-issues`);
}

function typeLabel(code?: string): string {
  const map: Record<string, string> = {
    epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷",
  };
  return map[code ?? ""] ?? code ?? "";
}

/** 分桶渲染配置：key → 标题（排除 total 计数） */
type BucketKey = Exclude<keyof WorkbenchSummary["my_issues"], "total">;
const BUCKETS: { key: BucketKey; label: string }[] = [
  { key: "overdue", label: "已逾期" },
  { key: "today", label: "今天到期" },
  { key: "upcoming", label: "即将到期" },
  { key: "in_progress", label: "进行中" },
  { key: "backlog", label: "待办" },
];

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">工作台</h1>
      <button
        v-if="summary"
        class="text-sm text-[var(--brand-600)] hover:underline"
        @click="load"
      >
        刷新
      </button>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 5" :key="i" class="h-16 w-full" />
    </div>

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- 空态 -->
    <AppEmptyState
      v-else-if="!workspaceId || !summary"
      title="暂无数据"
      description="请选择工作空间后查看工作台。"
    />

    <template v-else>
      <!-- 顶部统计卡片 -->
      <div class="grid grid-cols-2 gap-4 md:grid-cols-4">
        <div class="rounded-lg border border-[var(--border-subtle)] p-4">
          <div class="text-xs text-[var(--text-tertiary)]">我的需求/任务/缺陷</div>
          <div class="mt-1 text-2xl font-bold text-[var(--text-primary)]">
            {{ summary.my_issues.total }}
          </div>
        </div>
        <div class="rounded-lg border border-[var(--border-subtle)] p-4">
          <div class="text-xs text-[var(--text-tertiary)]">已逾期</div>
          <div class="mt-1 text-2xl font-bold text-[var(--danger, #ef4444)]">
            {{ summary.overdue_count }}
          </div>
        </div>
        <div class="rounded-lg border border-[var(--border-subtle)] p-4">
          <div class="text-xs text-[var(--text-tertiary)]">被阻塞</div>
          <div class="mt-1 text-2xl font-bold text-[var(--warning, #f59e0b)]">
            {{ summary.blocked_count }}
          </div>
        </div>
        <div class="rounded-lg border border-[var(--border-subtle)] p-4">
          <div class="text-xs text-[var(--text-tertiary)]">本周完成点数</div>
          <div class="mt-1 text-2xl font-bold text-[var(--brand-600)]">
            {{ efficiency?.week_points ?? "-" }}
          </div>
        </div>
      </div>

      <!-- 我的需求/任务/缺陷分桶 -->
      <section>
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">我的需求/任务/缺陷</h2>
        <div class="space-y-4">
          <div
            v-for="bucket in BUCKETS"
            :key="bucket.key"
            class="rounded-lg border border-[var(--border-subtle)] overflow-hidden"
          >
            <div class="flex items-center justify-between bg-[var(--surface-2)] px-4 py-2">
              <span class="text-sm font-medium text-[var(--text-primary)]">{{ bucket.label }}</span>
              <span class="text-xs text-[var(--text-tertiary)]">
                {{ summary.my_issues[bucket.key]?.length ?? 0 }} 项
              </span>
            </div>
            <div v-if="summary.my_issues[bucket.key]?.length" class="divide-y divide-[var(--border-subtle)]">
              <div
                v-for="d in summary.my_issues[bucket.key]"
                :key="d.id"
                class="flex cursor-pointer items-center gap-3 px-4 py-2.5 hover:bg-[var(--surface-2)] transition-colors"
                @click="openIssue()"
              >
                <span class="font-mono text-xs text-[var(--brand-600)]">{{ d.identifier }}</span>
                <span class="flex-1 truncate text-sm text-[var(--text-primary)]">{{ d.title }}</span>
                <span class="text-xs text-[var(--text-tertiary)]">{{ typeLabel(d.type_code) }}</span>
                <span
                  v-if="d.target_date"
                  class="text-xs text-[var(--text-tertiary)]"
                >{{ d.target_date.slice(0, 10) }}</span>
              </div>
            </div>
            <div v-else class="px-4 py-2 text-xs text-[var(--text-tertiary)]">暂无</div>
          </div>
        </div>
      </section>

      <!-- 迭代概览 -->
      <section v-if="summary.sprint_overviews.length">
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">我的迭代</h2>
        <div class="space-y-3">
          <div
            v-for="sp in summary.sprint_overviews"
            :key="sp.sprint_id"
            class="rounded-lg border border-[var(--border-subtle)] p-4"
          >
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-[var(--text-primary)]">{{ sp.sprint_name }}</span>
              <span class="text-xs text-[var(--text-tertiary)]">
                {{ sp.project_name }} · 剩 {{ sp.days_remaining }} 天
              </span>
            </div>
            <div class="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-[var(--surface-2)]">
              <div
                class="h-full rounded-full bg-[var(--brand-600)]"
                :style="{ width: `${Math.min(100, sp.progress)}%` }"
              />
            </div>
            <div class="mt-1 text-xs text-[var(--text-tertiary)]">
              {{ sp.progress }}% · 我的 {{ sp.my_issue_count }} 项
            </div>
          </div>
        </div>
      </section>

      <!-- 最近访问 -->
      <section v-if="summary.recent_items.length">
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">最近访问</h2>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="r in summary.recent_items"
            :key="`${r.item_type}-${r.item_id}`"
            class="rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-sm text-[var(--text-secondary)] hover:bg-[var(--surface-2)] transition-colors"
          >
            {{ r.title }}
          </button>
        </div>
      </section>
    </template>
  </div>
</template>
