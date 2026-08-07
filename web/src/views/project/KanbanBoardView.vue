<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi } from "@/api/services/workspace";
import { useIssueStore } from "@/stores/issue";

const route = useRoute();
const router = useRouter();
const issueStore = useIssueStore();

const projectId = computed(() => Number(route.params.projectId));
const loading = ref(true);
const error = ref("");
const dragIssueId = ref<number | null>(null);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsSlug = String(route.params.workspaceSlug ?? "");
    let wsIdVal: number;
    if (wsSlug) {
      const ws = await workspaceApi.getBySlug(wsSlug);
      wsIdVal = ws.id;
    } else {
      wsIdVal = Number(route.params.wsId);
    }
    await Promise.all([
      issueStore.fetchStates(wsIdVal, projectId.value),
      issueStore.fetchIssues(wsIdVal, projectId.value),
    ]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function issuesInState(stateId: number) {
  return issueStore.issues.filter((i) => i.state_id === stateId);
}

function onDragStart(issueId: number) {
  dragIssueId.value = issueId;
}

function onDragOver(e: DragEvent) {
  e.preventDefault();
}

async function onDrop(stateId: number) {
  if (dragIssueId.value == null) return;
  const issueId = dragIssueId.value;
  dragIssueId.value = null;

  const issue = issueStore.issues.find((i) => i.id === issueId);
  if (!issue || issue.state_id === stateId) return;

  const wsIdVal = Number(route.params.wsId ?? 0);
  try {
    await issueStore.transitionIssue(wsIdVal, projectId.value, issueId, stateId);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "流转失败";
  }
}

function openIssue(issueId: number) {
  router.push(`/${route.params.workspaceSlug}/projects/${projectId.value}/issues/${issueId}`);
}

function priorityColor(priority: string): string {
  const map: Record<string, string> = {
    urgent: "var(--danger-500)",
    high: "var(--warning-500)",
    medium: "var(--brand-500)",
    low: "var(--text-tertiary)",
  };
  return map[priority] ?? "var(--text-tertiary)";
}

onMounted(load);
</script>

<template>
  <div class="kanban">
    <header class="kanban__header">
      <h1>看板</h1>
      <p class="hint">拖拽工作项到不同状态列进行流转</p>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="kanban__board">
      <div
        v-for="state in issueStore.states"
        :key="state.id"
        class="kanban__column"
        @dragover="onDragOver"
        @drop="onDrop(state.id)"
      >
        <div class="kanban__column-header" :style="{ borderTopColor: state.color }">
          <span class="kanban__column-name">{{ state.name }}</span>
          <span class="kanban__column-count">{{ issuesInState(state.id).length }}</span>
        </div>

        <div class="kanban__cards">
          <div
            v-for="iss in issuesInState(state.id)"
            :key="iss.id"
            class="issue-card"
            draggable="true"
            @dragstart="onDragStart(iss.id)"
            @click="openIssue(iss.id)"
          >
            <div class="issue-card__header">
              <span class="issue-card__identifier" :style="{ color: priorityColor(iss.priority) }">
                {{ iss.identifier }}
              </span>
              <span
                class="issue-card__priority"
                :style="{ backgroundColor: priorityColor(iss.priority) }"
              ></span>
            </div>
            <div class="issue-card__name">{{ iss.name }}</div>
            <div class="issue-card__meta">
              <span class="issue-card__type-badge" :class="`type-${iss.type_code}`">
                {{ ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[iss.type_code] }}
              </span>
              <span v-if="iss.severity" class="issue-card__severity">S{{ iss.severity }}</span>
              <span v-if="iss.point != null" class="issue-card__point">{{ iss.point }}pt</span>
            </div>
            <div class="issue-card__footer">
              <div class="issue-card__assignees">
                <span v-for="uid in iss.assignees.slice(0, 3)" :key="uid" class="avatar-placeholder">
                  U{{ uid }}
                </span>
                <span v-if="iss.assignees.length > 3" class="avatar-more">
                  +{{ iss.assignees.length - 3 }}
                </span>
              </div>
              <div v-if="iss.progress > 0 && iss.progress < 100" class="issue-card__progress">
                <div class="progress-bar" :style="{ width: iss.progress + '%' }"></div>
              </div>
            </div>
          </div>

          <div v-if="issuesInState(state.id).length === 0" class="kanban__empty">
            暂无工作项
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.kanban__header {
  margin-bottom: 16px;
}
.kanban__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}
.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }

.loading, .error {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error { color: var(--danger-500); }

.kanban__board {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 16px;
  min-height: 400px;
}

.kanban__column {
  flex: 0 0 280px;
  background: var(--surface-2);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 220px);
}

.kanban__column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-top: 3px solid;
  border-bottom: 1px solid var(--border-subtle);
}

.kanban__column-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.kanban__column-count {
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 1px 8px;
  border-radius: 10px;
  background: var(--surface-3);
}

.kanban__cards {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kanban__empty {
  text-align: center;
  color: var(--text-tertiary);
  font-size: 12px;
  padding: 24px 0;
}

.issue-card {
  padding: 12px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: grab;
  transition: box-shadow 0.15s;
}
.issue-card:hover {
  box-shadow: var(--shadow-card);
}
.issue-card:active {
  cursor: grabbing;
}

.issue-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.issue-card__identifier {
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.issue-card__priority {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.issue-card__name {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.4;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.issue-card__meta {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.issue-card__type-badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--surface-3);
  color: var(--text-secondary);
  font-weight: 500;
}
.issue-card__type-badge.type-defect { background: var(--danger-50); color: var(--danger-600); }
.issue-card__type-badge.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.issue-card__type-badge.type-task { background: var(--success-50); color: var(--success-600); }

.issue-card__severity, .issue-card__point {
  font-size: 10px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.issue-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.issue-card__assignees {
  display: flex;
  align-items: center;
}

.avatar-placeholder {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--brand-100);
  color: var(--brand-600);
  font-size: 9px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--surface-1);
  margin-left: -4px;
}
.avatar-placeholder:first-child { margin-left: 0; }

.avatar-more {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--surface-3);
  color: var(--text-tertiary);
  font-size: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--surface-1);
  margin-left: -4px;
}

.issue-card__progress {
  width: 40px;
  height: 4px;
  background: var(--surface-3);
  border-radius: 2px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: var(--success-500);
  transition: width 0.3s;
}
</style>
