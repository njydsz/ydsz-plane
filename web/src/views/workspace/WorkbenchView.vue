<script setup lang="ts">
/**
 * WorkbenchView — 个人工作台首屏。
 * 聚合：我的任务分桶（今日/逾期/进行中/即将/待规划）+ 迭代概览 + 最近访问 + 快捷操作。
 * 后端：GET /workspaces/:wsId/workbench/summary
 */

import { computed, onMounted, ref } from "vue";

import AppEmptyState from "@/components/AppEmptyState.vue";
import AppLoadingState from "@/components/AppLoadingState.vue";
import AppErrorState from "@/components/AppErrorState.vue";
import {
  workbenchApi,
  type WorkbenchSummary,
  type IssueDigest,
  type SprintOverview,
} from "@/api/services/workbench";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const props = defineProps<{
  workspaceId: number;
}>();

const ws = ref<Workspace | null>(null);
const summary = ref<WorkbenchSummary | null>(null);
const loading = ref(true);
const error = ref("");

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.get(props.workspaceId);
    summary.value = await workbenchApi.getSummary(ws.value.id);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载工作台失败";
  } finally {
    loading.value = false;
  }
}

function priorityLabel(p: string): string {
  return ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[p] ?? p;
}

function priorityClass(p: string): string {
  return `priority priority--${p}`;
}

function fmtDate(d?: string | null): string {
  if (!d) return "";
  return d.slice(5); // MM-DD
}

function daysLabel(days: number): string {
  if (days < 0) return `已结束 ${-days} 天`;
  if (days === 0) return "今天结束";
  if (days <= 3) return `仅剩 ${days} 天`;
  return `${days} 天后`;
}

function accessedAt(s: string): string {
  const d = new Date(s);
  const now = Date.now();
  const diff = now - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 60) return `${mins || 1}分钟前`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}小时前`;
  return d.toLocaleDateString("zh-CN");
}

onMounted(load);

const hasIssues = computed(() => (summary.value?.my_issues.total ?? 0) > 0);
const hasOverviews = computed(() => (summary.value?.sprint_overviews.length ?? 0) > 0);
const hasRecent = computed(() => (summary.value?.recent_items.length ?? 0) > 0);

function issueRoute(issue: IssueDigest) {
  return `/${props.workspaceId}/projects/${issue.group_id}/issues/${issue.id}`;
}

function sprintRoute(sprint: SprintOverview) {
  return `/${props.workspaceId}/projects/${sprint.project_id}/sprints/${sprint.sprint_id}`;
}
</script>

<template>
  <div class="workbench">
    <!-- Header -->
    <header class="workbench__header">
      <div>
        <h1 class="workbench__title">工作台</h1>
        <p class="workbench__subtitle">{{ ws?.name ?? workspaceId }} · 个人工作概览</p>
      </div>
      <div class="workbench__actions">
        <span v-if="summary" class="workbench__stats">
          今日 {{ summary.my_issues.today.length }} · 进行中 {{ summary.my_issues.in_progress.length }}
          <span v-if="summary.overdue_count > 0" class="workbench__overdue">
            · 逾期 {{ summary.overdue_count }}
          </span>
        </span>
      </div>
    </header>

    <AppLoadingState v-if="loading" text="加载工作台..." />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <div v-else-if="summary" class="workbench__body">
      <div class="workbench__main">
        <!-- 我的任务 -->
        <section class="workbench__section">
          <div class="workbench__section-head">
            <h2>我的任务</h2>
            <span v-if="hasIssues" class="workbench__badge">{{ summary.my_issues.total }}</span>
          </div>

          <AppEmptyState
            v-if="!hasIssues"
            title="暂无待办任务"
            description="你没有被分配的工作项"
          />

          <div v-else class="workbench__buckets">
            <!-- 今日 -->
            <div v-if="summary.my_issues.today.length" class="bucket">
              <h3 class="bucket__title bucket__title--today">今日 ({{ summary.my_issues.today.length }})</h3>
              <div
                v-for="issue in summary.my_issues.today"
                :key="issue.id"
                class="bucket__item"
              >
                <router-link :to="issueRoute(issue)" class="issue-link">
                  <span class="issue-link__id">{{ issue.identifier }}</span>
                  <span class="issue-link__name">{{ issue.title }}</span>
                  <span class="issue-link__meta">
                    <span :class="priorityClass(issue.priority)">{{ priorityLabel(issue.priority) }}</span>
                    <span class="issue-link__state" :style="{ color: issue.state_color }">{{ issue.state_name }}</span>
                  </span>
                </router-link>
              </div>
            </div>

            <!-- 逾期 -->
            <div v-if="summary.my_issues.overdue.length" class="bucket">
              <h3 class="bucket__title bucket__title--overdue">逾期 ({{ summary.my_issues.overdue.length }})</h3>
              <div
                v-for="issue in summary.my_issues.overdue"
                :key="issue.id"
                class="bucket__item"
              >
                <router-link :to="issueRoute(issue)" class="issue-link">
                  <span class="issue-link__id">{{ issue.identifier }}</span>
                  <span class="issue-link__name">{{ issue.title }}</span>
                  <span class="issue-link__meta">
                    <span class="issue-link__target issue-link__target--overdue">截止 {{ fmtDate(issue.target_date) }}</span>
                  </span>
                </router-link>
              </div>
            </div>

            <!-- 进行中 -->
            <div v-if="summary.my_issues.in_progress.length" class="bucket">
              <h3 class="bucket__title bucket__title--progress">进行中 ({{ summary.my_issues.in_progress.length }})</h3>
              <div
                v-for="issue in summary.my_issues.in_progress"
                :key="issue.id"
                class="bucket__item"
              >
                <router-link :to="issueRoute(issue)" class="issue-link">
                  <span class="issue-link__id">{{ issue.identifier }}</span>
                  <span class="issue-link__name">{{ issue.title }}</span>
                  <span class="issue-link__meta">
                    <span :class="priorityClass(issue.priority)">{{ priorityLabel(issue.priority) }}</span>
                    <span v-if="issue.is_blocked" class="issue-link__blocked">⛔ 阻塞</span>
                  </span>
                </router-link>
              </div>
            </div>

            <!-- 即将 -->
            <div v-if="summary.my_issues.upcoming.length" class="bucket">
              <h3 class="bucket__title bucket__title--upcoming">即将 ({{ summary.my_issues.upcoming.length }})</h3>
              <div
                v-for="issue in summary.my_issues.upcoming"
                :key="issue.id"
                class="bucket__item"
              >
                <router-link :to="issueRoute(issue)" class="issue-link">
                  <span class="issue-link__id">{{ issue.identifier }}</span>
                  <span class="issue-link__name">{{ issue.title }}</span>
                  <span class="issue-link__meta">
                    <span class="issue-link__target">{{ fmtDate(issue.target_date) }}</span>
                  </span>
                </router-link>
              </div>
            </div>

            <!-- 待规划 -->
            <div v-if="summary.my_issues.backlog.length" class="bucket">
              <h3 class="bucket__title">待规划 ({{ summary.my_issues.backlog.length }})</h3>
              <div
                v-for="issue in summary.my_issues.backlog"
                :key="issue.id"
                class="bucket__item"
              >
                <router-link :to="issueRoute(issue)" class="issue-link">
                  <span class="issue-link__id">{{ issue.identifier }}</span>
                  <span class="issue-link__name">{{ issue.title }}</span>
                </router-link>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- 侧边栏 -->
      <aside v-if="hasOverviews || hasRecent" class="workbench__sidebar">
        <!-- 迭代概览 -->
        <section v-if="hasOverviews" class="workbench__section">
          <h2>迭代概览</h2>
          <div
            v-for="sprint in summary.sprint_overviews"
            :key="sprint.sprint_id"
            class="sprint-card"
          >
            <router-link :to="sprintRoute(sprint)" class="sprint-card__link">
              <div class="sprint-card__top">
                <span class="sprint-card__name">{{ sprint.sprint_name }}</span>
                <span
                  class="sprint-card__days"
                  :class="{ 'sprint-card__days--urgent': sprint.days_remaining <= 3 && sprint.days_remaining >= 0 }"
                >
                  {{ daysLabel(sprint.days_remaining) }}
                </span>
              </div>
              <div class="sprint-card__progress">
                <div class="progress-bar">
                  <div class="progress-bar__fill" :style="{ width: `${Math.round(sprint.progress * 100)}%` }"></div>
                </div>
                <span class="sprint-card__meta">
                  {{ Math.round(sprint.progress * 100) }}% · {{ sprint.my_issue_count }} 项
                </span>
              </div>
            </router-link>
          </div>
        </section>

        <!-- 最近访问 -->
        <section v-if="hasRecent" class="workbench__section">
          <h2>最近访问</h2>
          <div class="recent-list">
            <a
              v-for="item in summary.recent_items"
              :key="`${item.item_type}-${item.item_id}`"
              :href="item.url || '#'"
              class="recent-item"
            >
              <span class="recent-item__title">
                {{ item.identifier ? item.identifier + ' ' : '' }}{{ item.title }}
              </span>
              <span class="recent-item__time">{{ accessedAt(item.accessed_at) }}</span>
            </a>
          </div>
        </section>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.workbench {
  max-width: 1200px;
  margin: 0 auto;
}

.workbench__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.workbench__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0;
}

.workbench__subtitle {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
  margin: 4px 0 0;
}

.workbench__stats {
  font-size: 13px;
  color: var(--text-secondary, #4b5563);
}

.workbench__overdue {
  color: var(--danger-500, #ef4444);
  font-weight: 600;
}

.workbench__body {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 28px;
}

.workbench__section {
  margin-bottom: 28px;
}

.workbench__section:last-child {
  margin-bottom: 0;
}

.workbench__section-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.workbench__section-head h2 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0;
}

.workbench__badge {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #9ca3af);
  background: var(--surface-3, #f3f4f6);
  padding: 1px 8px;
  border-radius: 10px;
}

/* ---- Buckets ---- */
.workbench__buckets {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.bucket__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #4b5563);
  margin: 0 0 8px;
}

.bucket__title--today { color: var(--brand-600, #2563eb); }
.bucket__title--overdue { color: var(--danger-500, #ef4444); }
.bucket__title--progress { color: var(--warning-600, #d97706); }
.bucket__title--upcoming { color: var(--success-600, #059669); }

.bucket__item {
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.bucket__item:last-child {
  border-bottom: none;
}

.issue-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  text-decoration: none;
  color: inherit;
  font-size: 13px;
  transition: opacity 0.15s;
}

.issue-link:hover {
  opacity: 0.7;
}

.issue-link__id {
  font-family: var(--font-mono, 'Consolas', monospace);
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  flex-shrink: 0;
}

.issue-link__name {
  flex: 1;
  color: var(--text-primary, #1f2937);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.issue-link__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  font-size: 11px;
}

.issue-link__state {
  font-weight: 500;
}

.issue-link__target {
  color: var(--text-tertiary, #9ca3af);
}

.issue-link__target--overdue {
  color: var(--danger-500, #ef4444);
}

.issue-link__blocked {
  color: var(--warning-600, #d97706);
}

.priority {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 500;
}

.priority--urgent { background: var(--danger-50, rgba(239,68,68,0.1)); color: var(--danger-600, #dc2626); }
.priority--high { background: var(--warning-50, rgba(245,158,11,0.1)); color: var(--warning-600, #d97706); }
.priority--medium { background: var(--info-50, rgba(59,130,246,0.1)); color: var(--info-600, #2563eb); }
.priority--low { background: var(--surface-3, #f3f4f6); color: var(--text-tertiary, #9ca3af); }

/* ---- Sidebar Sprint cards ---- */
.sprint-card {
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  margin-bottom: 8px;
  transition: border-color 0.15s;
}

.sprint-card:hover {
  border-color: var(--brand-200, #bfdbfe);
}

.sprint-card__link {
  display: block;
  padding: 10px 12px;
  text-decoration: none;
  color: inherit;
}

.sprint-card__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.sprint-card__name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.sprint-card__days {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.sprint-card__days--urgent {
  color: var(--danger-500, #ef4444);
  font-weight: 600;
}

.sprint-card__progress {
  display: flex;
  align-items: center;
  gap: 8px;
}

.progress-bar {
  flex: 1;
  height: 4px;
  background: var(--surface-3, #f3f4f6);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar__fill {
  height: 100%;
  background: var(--brand-500, #3b82f6);
  border-radius: 2px;
  transition: width 0.3s;
}

.sprint-card__meta {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  white-space: nowrap;
}

/* ---- Recent items ---- */
.recent-list {
  display: flex;
  flex-direction: column;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  font-size: 12px;
  text-decoration: none;
  color: var(--text-secondary, #4b5563);
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.recent-item:last-child {
  border-bottom: none;
}

.recent-item:hover {
  color: var(--brand-500, #3b82f6);
}

.recent-item__title {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
}

.recent-item__time {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  flex-shrink: 0;
}

@media (max-width: 900px) {
  .workbench__body {
    grid-template-columns: 1fr;
  }
}
</style>
