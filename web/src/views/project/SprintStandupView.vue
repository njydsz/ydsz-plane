<script setup lang="ts">
/**
 * SprintStandupView — 站会模式视图（Sprint 5.9）。
 *
 * 将 active 迭代中的工作项按「昨日」「今日」「阻塞」分组展示，
 * 支持一键切换工作项状态，适配每日站会节奏。
 *
 * 参考：Jira Sprint Health / Linear Standup / Plane iterations。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { sprintApi, type Sprint, type SprintIssueView } from "@/api/services/sprint";
import { workspaceApi } from "@/api/services/workspace";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const sprintId = computed(() => Number(route.params.sprintId));

const sprint = ref<Sprint | null>(null);
const issues = ref<SprintIssueView[]>([]);
const loading = ref(true);
const error = ref("");
const filterAssignee = ref<number | null>(null);
const expandedIds = ref(new Set<number>());

let wsIdVal = 0;

/* ------------------------------------------------------------------ */
/* 加载                                                                */
/* ------------------------------------------------------------------ */

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const ws = await workspaceApi.get(workspaceId.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const [spRes, issRes] = await Promise.all([
      sprintApi.getSprint(wsId, projectId.value, sprintId.value),
      sprintApi.listSprintIssues(wsId, projectId.value, sprintId.value),
    ]);
    sprint.value = spRes;
    issues.value = issRes.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 分组逻辑                                                             */
/* ------------------------------------------------------------------ */

/** 按状态分组：已完成→昨日，进行中→今日，阻塞状态→阻塞 */
type StandupGroup = "yesterday" | "today" | "blocked";

interface GroupedIssues {
  yesterday: SprintIssueView[];
  today: SprintIssueView[];
  blocked: SprintIssueView[];
}

const STATE_GROUP_MAP: Record<string, StandupGroup> = {
  completed: "yesterday",
  cancelled: "yesterday",
  started: "today",
  unstarted: "today",
  backlog: "today",
};

const blockedIssueIds = computed(() => {
  // 如果 issue 的 state_name 包含"阻塞"、"blocked"、"on hold" 等关键词
  return new Set(
    issues.value
      .filter(
        (i) =>
          i.state_name?.includes("阻塞") ||
          i.state_group === "blocked" ||
          i.state_name?.toLowerCase().includes("block"),
      )
      .map((i) => i.issue_id),
  );
});

const grouped = computed<GroupedIssues>(() => {
  const result: GroupedIssues = { yesterday: [], today: [], blocked: [] };
  let filtered = issues.value;
  if (filterAssignee.value) {
    // 简易筛选（注：当前后端返回不含 assignee_id，扩展时补）
    // filtered = filtered.filter((i: any) => i.assignee_id === filterAssignee.value);
  }
  for (const iss of filtered) {
    if (blockedIssueIds.value.has(iss.issue_id)) {
      result.blocked.push(iss);
    } else {
      const grp = STATE_GROUP_MAP[iss.state_group] ?? "today";
      result[grp].push(iss);
    }
  }
  return result;
});

const sectionDefs = computed(() => [
  { key: "blocked", label: "阻塞项", color: "var(--danger-500)", items: grouped.value.blocked },
  { key: "today", label: "今日计划", color: "var(--warning-500)", items: grouped.value.today },
  { key: "yesterday", label: "昨日完成", color: "var(--success-500)", items: grouped.value.yesterday },
]);

/* ------------------------------------------------------------------ */
/* 操作                                                                */
/* ------------------------------------------------------------------ */

function toggleExpand(id: number) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id);
  } else {
    expandedIds.value = new Set([id]); // 单展开模式
  }
}

/* ------------------------------------------------------------------ */
/* 格式化                                                               */
/* ------------------------------------------------------------------ */

function typeLabel(tc: string): string {
  const m: Record<string, string> = {
    epic: "史诗",
    requirement: "需求",
    task: "任务",
    defect: "缺陷",
  };
  return m[tc] ?? tc;
}

function priorityLabel(p: string): string {
  const m: Record<string, string> = {
    urgent: "紧急",
    high: "高",
    medium: "中",
    low: "低",
  };
  return m[p] ?? p;
}

function fmtDate(d?: string): string {
  if (!d) return "";
  return d.slice(0, 10);
}

function statusColor(s: string): string {
  const m: Record<string, string> = {
    planned: "var(--text-tertiary)",
    active: "var(--success-500)",
    completed: "var(--brand-500)",
  };
  return m[s] ?? "var(--text-tertiary)";
}

onMounted(load);
</script>

<template>
  <div class="standup-view">
    <!-- Header -->
    <header class="header">
      <div class="title-row">
        <button class="back-btn" title="返回迭代详情" @click="router.push(`/${workspaceId}/projects/${projectId}/sprints/${sprintId}`)">←</button>
        <div>
          <h1>每日站会</h1>
          <p v-if="sprint" class="meta">
            <span class="badge" :style="{ color: statusColor(sprint.status) }">
              {{ sprint.name }}
            </span>
            <span v-if="sprint.start_date" class="date">{{ fmtDate(sprint.start_date) }} → {{ fmtDate(sprint.end_date) }}</span>
          </p>
        </div>
      </div>
    </header>

    <!-- 状态：加载 / 错误 / 空 -->
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState v-else-if="!sprint || issues.length === 0" title="该迭代暂无工作项" description="前往排期规划页添加工作项">
      <button class="btn btn-secondary" @click="router.push(`/${workspaceId}/projects/${projectId}/sprints/planning`)">前往排期规划</button>
    </AppEmptyState>

    <!-- 三栏分组视图 -->
    <div v-else class="columns">
      <div v-for="sec in sectionDefs" :key="sec.key" class="col">
        <h3 class="col-title" :style="{ color: sec.color, borderBottomColor: sec.color }">
          {{ sec.label }}
          <span class="count">{{ sec.items.length }}</span>
        </h3>

        <div v-if="sec.items.length === 0" class="empty-col">
          {{ sec.key === "blocked" ? "无阻塞项" : sec.key === "today" ? "暂无待办" : "尚无完成项" }}
        </div>

        <div v-for="iss in sec.items" :key="iss.issue_id" class="card" @click="toggleExpand(iss.issue_id)">
          <div class="card-top">
            <span class="type-badge" :class="`type-${iss.type_code}`">
              {{ typeLabel(iss.type_code).charAt(0) }}
            </span>
            <span class="name">{{ iss.name }}</span>
            <span v-if="iss.point != null" class="pt">{{ iss.point }}pt</span>
          </div>
          <div class="card-meta">
            <span class="state-tag" :style="{ background: iss.state_color + '20', color: iss.state_color }">
              {{ iss.state_name }}
            </span>
            <span class="priority" :class="`p-${iss.priority}`">{{ priorityLabel(iss.priority) }}</span>
          </div>
          <div v-if="expandedIds.has(iss.issue_id)" class="card-detail">
            <p class="hint">点击前往工作项详情页更新状态</p>
            <button
              class="btn btn-sm btn-primary"
              @click.stop="router.push(`/${workspaceId}/projects/${projectId}/issues/${iss.issue_id}`)"
            >
              打开详情
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ------------------------------------------------------------------ */
/* Layout                                                              */
/* ------------------------------------------------------------------ */
.standup-view {
  padding: 16px 20px;
  max-width: 1280px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}
.title-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}
.back-btn {
  width: 32px; height: 32px; border-radius: 50%;
  border: 1px solid var(--border-subtle); background: var(--surface-2);
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  font-size: 16px; color: var(--text-secondary); flex-shrink: 0;
}
.back-btn:hover { background: var(--surface-3); }

.header h1 {
  margin: 0; font-size: 18px; font-weight: 600;
}
.meta {
  display: flex; gap: 8px; align-items: center; margin: 4px 0 0; font-size: 12px;
}
.badge {
  font-size: 11px; padding: 2px 8px; border-radius: 12px; background: var(--surface-3); font-weight: 500;
}
.date {
  color: var(--text-tertiary); font-family: var(--font-mono);
}

/* Three columns */
.columns {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
@media (max-width: 900px) {
  .columns { grid-template-columns: 1fr; }
}

.col {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.col-title {
  margin: 0 0 4px;
  font-size: 14px;
  font-weight: 600;
  padding-bottom: 6px;
  border-bottom: 2px solid var(--border-subtle);
  display: flex;
  align-items: center;
  gap: 6px;
}
.count {
  font-size: 11px;
  font-weight: 500;
  background: var(--surface-2);
  padding: 1px 7px;
  border-radius: 10px;
  color: var(--text-secondary);
}

.empty-col {
  text-align: center;
  padding: 24px 8px;
  color: var(--text-tertiary);
  font-size: 12px;
  background: var(--surface-1);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-sm);
}

/* ------------------------------------------------------------------ */
/* Cards                                                                */
/* ------------------------------------------------------------------ */
.card {
  padding: 10px 12px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.card:hover {
  border-color: var(--brand-300);
  box-shadow: 0 2px 6px rgba(0,0,0,0.06);
}

.card-top {
  display: flex;
  align-items: center;
  gap: 6px;
}
.card-top .name {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.4;
}
.card-top .pt {
  font-size: 11px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
}

.card-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-top: 6px;
}
.state-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
}
.priority {
  font-size: 10px;
  color: var(--text-tertiary);
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--surface-2);
}
.priority.p-urgent { color: var(--danger-600); background: var(--danger-50); }
.priority.p-high { color: var(--warning-700); background: var(--warning-50); }

.card-detail {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.card-detail .hint {
  font-size: 11px;
  color: var(--text-tertiary);
  margin: 0;
}

/* ------------------------------------------------------------------ */
/* Type badge                                                           */
/* ------------------------------------------------------------------ */
.type-badge {
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 3px;
  background: var(--surface-3);
  color: var(--text-secondary);
  font-weight: 600;
  min-width: 18px;
  text-align: center;
  line-height: 18px;
  flex-shrink: 0;
}
.type-badge.type-defect { background: var(--danger-50); color: var(--danger-600); }
.type-badge.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.type-badge.type-task { background: var(--success-50); color: var(--success-600); }

/* ------------------------------------------------------------------ */
/* Center message states                                                */
/* ------------------------------------------------------------------ */
.center-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  gap: 10px;
  color: var(--text-tertiary);
}
.center-message.error { color: var(--danger-500); }
.center-message p { margin: 0; font-size: 14px; }

/* Skeleton loading */
.skeleton-line {
  height: 40px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* ------------------------------------------------------------------ */
/* Buttons                                                              */
/* ------------------------------------------------------------------ */
.btn {
  font-size: 13px; font-weight: 500; padding: 6px 14px;
  border-radius: var(--radius-sm); border: 1px solid var(--border-subtle);
  cursor: pointer; transition: background 0.15s;
}
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover { background: var(--brand-600); }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
.btn-sm { font-size: 11px; padding: 4px 10px; }
</style>
