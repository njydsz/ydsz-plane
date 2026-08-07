<script setup lang="ts">
/**
 * 排期规划页 — Backlog 与迭代间拖拽分配工作项。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  sprintApi,
  type BacklogItem,
  type Sprint,
  type SprintIssueView,
} from "@/api/services/sprint";
import { workspaceApi } from "@/api/services/workspace";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));

const loading = ref(true);
const saving = ref(false);
const error = ref("");

let wsIdVal = 0;

const plannedSprints = ref<Sprint[]>([]);
const activeSprint = ref<Sprint | null>(null);
const backlog = ref<BacklogItem[]>([]);
const sprintIssues = ref<SprintIssueView[]>([]);
const selectedSprintId = ref<number | null>(null);
const capacityStats = ref<{ avg: number; p50: number } | null>(null);

const selectedSprint = computed(() =>
  plannedSprints.value.find((s) => s.id === selectedSprintId.value) ?? activeSprint.value,
);

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function loadAll() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const [sprintRes, backlogRes] = await Promise.all([
      sprintApi.listSprints(wsId, projectId.value),
      sprintApi.getBacklog(wsId, projectId.value),
    ]);
    plannedSprints.value = sprintRes.results.filter((s) => s.status === "planned");
    activeSprint.value = sprintRes.results.find((s) => s.status === "active") ?? null;
    backlog.value = backlogRes.results;

    // 默认选中 active，否则第一个 planned
    const target = activeSprint.value ?? plannedSprints.value[0];
    if (target) {
      selectedSprintId.value = target.id;
      await loadSprintIssues(target.id);
    }

    // 速率建议
    try {
      const stats = await sprintApi.suggestCapacity(wsId, projectId.value);
      capacityStats.value = { avg: Math.round(stats.avg_points * 10) / 10, p50: stats.p50 };
    } catch {
      capacityStats.value = null;
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadSprintIssues(sprintId: number) {
  const wsId = wsIdVal;
  const res = await sprintApi.listSprintIssues(wsId, projectId.value, sprintId);
  sprintIssues.value = res.results;
}

function totalPoints(items: SprintIssueView[]): number {
  return items.reduce((sum, i) => sum + (i.point ?? 0), 0);
}

function donePoints(): number {
  return sprintIssues.value
    .filter((i) => i.state_group === "completed")
    .reduce((sum, i) => sum + (i.point ?? 0), 0);
}

async function pickSprint(sp: Sprint) {
  selectedSprintId.value = sp.id;
  await loadSprintIssues(sp.id);
}

// drag state
const dragBacklogId = ref<number | null>(null);
const dragSprintIdx = ref<number | null>(null);

function onBacklogDragStart(issueId: number) {
  dragBacklogId.value = issueId;
}

function onSprintReorderStart(idx: number) {
  dragSprintIdx.value = idx;
}

function onDragOver(e: DragEvent) {
  e.preventDefault();
}

async function onDropToSprint() {
  if (dragBacklogId.value == null) return;
  if (!selectedSprint.value) {
    error.value = "请先选择或创建一个迭代";
    return;
  }
  const issue = backlog.value.find((b) => b.issue_id === dragBacklogId.value);
  if (!issue) return;

  saving.value = true;
  try {
    await sprintApi.addIssue(wsIdVal, projectId.value, selectedSprint.value.id, issue.issue_id);
    backlog.value = backlog.value.filter((b) => b.issue_id !== issue.issue_id);
    await loadSprintIssues(selectedSprint.value.id);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "添加失败";
  } finally {
    saving.value = false;
    dragBacklogId.value = null;
  }
}

async function onDropToBacklog() {
  if (dragSprintIdx.value == null) return;
  const target = sprintIssues.value[dragSprintIdx.value];
  if (!target) return;

  saving.value = true;
  try {
    await sprintApi.removeIssue(wsIdVal, projectId.value, selectedSprintId.value!, target.issue_id);
    sprintIssues.value = sprintIssues.value.filter((_, i) => i !== dragSprintIdx.value);
    await reloadBacklog();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "移除失败";
  } finally {
    saving.value = false;
    dragSprintIdx.value = null;
  }
}

async function reloadBacklog() {
  const res = await sprintApi.getBacklog(wsIdVal, projectId.value);
  backlog.value = res.results;
}

function openSprint(sp: Sprint) {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/${sp.id}`);
}

onMounted(loadAll);
</script>

<template>
  <div class="planning">
    <header class="header">
      <div>
        <h1>排期规划</h1>
        <p class="hint">从左侧 Backlog 拖拽工作项到右侧迭代</p>
      </div>
      <div class="actions">
        <span v-if="capacityStats" class="capacity-badge">
          速率: avg {{ capacityStats.avg }}pt / P50 {{ capacityStats.p50 }}pt
        </span>
        <button class="btn btn-secondary" @click="router.push(`/${workspaceSlug}/projects/${projectId}/sprints`)">
          返回列表
        </button>
      </div>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="content">
      <!-- 左：Backlog -->
      <section class="panel backlog-panel">
        <div class="panel-header">
          <h2>Backlog</h2>
          <span class="count">{{ backlog.length }} 项</span>
        </div>
        <div
          class="panel-body"
          @dragover="onDragOver"
          @drop="onDropToBacklog"
        >
          <div
            v-for="item in backlog"
            :key="item.issue_id"
            class="issue-item"
            draggable="true"
            @dragstart="onBacklogDragStart(item.issue_id)"
          >
            <div class="title-row">
              <span class="type-badge" :class="`type-${item.type_code}`">
                {{ ({ requirement: "需", task: "任", defect: "缺" } as Record<string, string>)[item.type_code] }}
              </span>
              <span class="name">{{ item.name }}</span>
            </div>
            <div class="meta">
              <span :style="{ color: item.state_color }">{{ item.state_name }}</span>
              <span v-if="item.point != null" class="point">{{ item.point }}pt</span>
              <span v-if="item.priority === 'urgent'" class="priority-dot urgent"></span>
            </div>
          </div>
          <div v-if="backlog.length === 0" class="empty">Backlog 为空</div>
        </div>
      </section>

      <!-- 右：Sprint -->
      <section class="panel sprint-panel">
        <div class="sprint-tabs">
          <button
            v-for="sp in plannedSprints"
            :key="sp.id"
            class="tab"
            :class="{ active: selectedSprintId === sp.id }"
            @click="pickSprint(sp)"
          >
            {{ sp.name }}
          </button>
          <span
            v-if="activeSprint"
            class="tab active"
            style="cursor: default;"
          >
            {{ activeSprint.name }} (active)
          </span>
        </div>

        <div class="panel-header" v-if="selectedSprint">
          <div>
            <h2>{{ selectedSprint.name }}</h2>
            <p v-if="selectedSprint.goal" class="goal">{{ selectedSprint.goal }}</p>
          </div>
          <div class="stats">
            <div class="stat">
              <span class="num">{{ donePoints() }}</span>
              <span class="label">已完成</span>
            </div>
            <div class="stat">
              <span class="num">{{ totalPoints(sprintIssues) }}</span>
              <span class="label">承诺</span>
            </div>
            <div class="stat" v-if="selectedSprint.capacity">
              <span class="num" :class="{ over: totalPoints(sprintIssues) > (selectedSprint.capacity ?? 0) }">
                {{ Math.round((totalPoints(sprintIssues) / selectedSprint.capacity) * 100) }}%
              </span>
              <span class="label">饱和度</span>
            </div>
          </div>
        </div>

        <div
          v-if="selectedSprint"
          class="panel-body"
          @dragover="onDragOver"
          @drop="onDropToSprint"
        >
          <div
            v-for="(iss, idx) in sprintIssues"
            :key="iss.issue_id"
            class="issue-item"
            draggable="true"
            @dragstart="onSprintReorderStart(idx)"
          >
            <div class="title-row">
              <span class="type-badge" :class="`type-${iss.type_code}`">
                {{ ({ requirement: "需", task: "任", defect: "缺" } as Record<string, string>)[iss.type_code] }}
              </span>
              <span class="name">{{ iss.name }}</span>
            </div>
            <div class="meta">
              <span :style="{ color: iss.state_color }">{{ iss.state_name }}</span>
              <span v-if="iss.point != null" class="point">{{ iss.point }}pt</span>
            </div>
          </div>
          <div v-if="sprintIssues.length === 0 && selectedSprint" class="empty">
            拖拽工作项到此迭代中
          </div>
          <div v-if="!selectedSprint" class="empty">请先选择或创建一个迭代</div>
        </div>

        <div class="panel-footer" v-if="selectedSprint">
          <button class="btn btn-primary" @click="openSprint(selectedSprint)">
            查看迭代详情
          </button>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 16px; gap: 12px; flex-wrap: wrap;
}
.header h1 { margin: 0; font-size: 20px; }
.hint { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }
.actions { display: flex; align-items: center; gap: 8px; }
.capacity-badge {
  font-size: 12px; padding: 3px 8px; border-radius: 10px;
  background: var(--brand-50); color: var(--brand-700);
}

.loading, .error { text-align: center; padding: 48px 0; }
.error { color: var(--danger-500); }

.content { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.panel {
  background: var(--surface-1); border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md); display: flex; flex-direction: column;
  min-height: 400px;
}

.panel-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  padding: 12px 14px; border-bottom: 1px solid var(--border-subtle);
}
.panel-header h2 { margin: 0; font-size: 14px; font-weight: 600; }
.goal { margin: 2px 0 0; font-size: 11px; color: var(--text-secondary); }
.count { font-size: 11px; color: var(--text-tertiary); padding: 2px 6px; border-radius: 10px; background: var(--surface-3); }

.panel-body { flex: 1; overflow-y: auto; padding: 8px; display: flex; flex-direction: column; gap: 6px; }
.panel-footer { padding: 10px 14px; border-top: 1px solid var(--border-subtle); display: flex; justify-content: flex-end; }

.empty { text-align: center; color: var(--text-tertiary); font-size: 12px; padding: 32px 0; }

.issue-item {
  padding: 8px 10px; background: var(--surface-1);
  border: 1px solid var(--border-subtle); border-radius: var(--radius-sm);
  cursor: grab; transition: box-shadow 0.15s;
}
.issue-item:hover { box-shadow: var(--shadow-card); }
.issue-item:active { cursor: grabbing; }

.title-row { display: flex; align-items: center; gap: 6px; margin-bottom: 4px; }
.name { font-size: 12px; font-weight: 500; color: var(--text-primary);
  display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }

.type-badge {
  font-size: 9px; padding: 1px 5px; border-radius: 3px;
  background: var(--surface-3); color: var(--text-secondary); font-weight: 600;
  min-width: 18px; text-align: center;
}
.type-badge.type-defect { background: var(--danger-50); color: var(--danger-600); }
.type-badge.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.type-badge.type-task { background: var(--success-50); color: var(--success-600); }

.meta { display: flex; align-items: center; gap: 8px; font-size: 11px; color: var(--text-tertiary); }
.point { font-family: var(--font-mono); }
.priority-dot { width: 6px; height: 6px; border-radius: 50%; }
.priority-dot.urgent { background: var(--danger-500); }

.sprint-tabs {
  display: flex; gap: 4px; padding: 8px 10px 0;
  border-bottom: 1px solid var(--border-subtle); flex-wrap: wrap;
}
.tab {
  background: none; border: none; padding: 6px 10px; font-size: 12px;
  color: var(--text-secondary); cursor: pointer; border-radius: 4px 4px 0 0;
  border-bottom: 2px solid transparent;
}
.tab.active { color: var(--text-primary); border-bottom-color: var(--brand-500); font-weight: 500; }

.stats { display: flex; gap: 12px; }
.stat { display: flex; flex-direction: column; align-items: center; }
.stat .num { font-size: 14px; font-weight: 600; font-family: var(--font-mono); }
.stat .num.over { color: var(--danger-500); }
.stat .label { font-size: 10px; color: var(--text-tertiary); }

.btn {
  font-size: 13px; font-weight: 500; padding: 6px 14px; border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle); cursor: pointer; transition: background 0.15s;
}
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover { background: var(--brand-600); }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
</style>
