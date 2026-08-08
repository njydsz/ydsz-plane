<script setup lang="ts">
/**
 * 看板视图 — 按状态分列展示工作项。
 * 支持: 列间拖拽流转 / 列内拖拽排序 / 视觉反馈 / 中值插入排序。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { issueApi, type Issue, type IssuePriority } from "@/api/services/issue";
import { preferenceApi } from "@/api/services/preference";
import { useIssueStore } from "@/stores/issue";
import { usePeekStore } from "@/stores/peek";
import { prefs } from "@/lib/prefs";
import { toast, promiseToast } from "@/lib/toast";
import IssueCreateModal from "./IssueCreateModal.vue";
import { AppErrorState, AppEmptyState, InlineEdit, InlineSelectEdit, AppSkeleton } from "@/components";

const route = useRoute();
const issueStore = useIssueStore();
const peek = usePeekStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = ref(0);
const loading = ref(true);
const error = ref("");
const showCreateModal = ref(false);

// --- 视图偏好（列宽可调 + 持久化）---
const columnWidth = ref(280);
const prefsReady = ref(false);
let savePrefsTimer: ReturnType<typeof setTimeout> | null = null;

/** 加载当前用户的看板偏好（列宽等） */
async function loadViewPreferences() {
  try {
    const wsIdVal = Number(route.params.workspaceId);
    const pref = await preferenceApi.get(wsIdVal, projectId.value, "kanban");
    if (pref?.extra && typeof pref.extra === "object") {
      const extra = pref.extra as Record<string, unknown>;
      const w = Number(extra.column_width);
      if (w >= 200 && w <= 480) {
        columnWidth.value = w;
      }
    }
  } catch {
    // 偏好加载失败不阻塞看板
  } finally {
    prefsReady.value = true;
  }
}

/** 防抖保存列宽到服务端（按用户 × 项目 × 看板） */
function persistColumnWidth() {
  if (!prefsReady.value || !wsId.value) return;
  if (savePrefsTimer) clearTimeout(savePrefsTimer);
  savePrefsTimer = setTimeout(async () => {
    try {
      await preferenceApi.save(wsId.value, projectId.value, "kanban", {
        extra: { column_width: columnWidth.value },
      });
    } catch {
      // 静默失败，下次继续尝试
    }
  }, 600);
}

/** 列宽拖拽调整 */
function startColumnResize(e: PointerEvent) {
  e.preventDefault();
  const startX = e.clientX;
  const startWidth = columnWidth.value;
  const onMove = (ev: PointerEvent) => {
    columnWidth.value = Math.min(480, Math.max(200, startWidth + (ev.clientX - startX)));
    persistColumnWidth();
  };
  const onUp = () => {
    persistColumnWidth();
    window.removeEventListener("pointermove", onMove);
    window.removeEventListener("pointerup", onUp);
  };
  window.addEventListener("pointermove", onMove);
  window.addEventListener("pointerup", onUp);
}

// --- 拖拽状态 ---
const dragIssue = ref<Issue | null>(null);
const dragOverColumn = ref<number | null>(null);
const dropIndex = ref<number | null>(null);
/** 正在执行 transition / reorder API 调用，阻止重复提交 */
const processingDrop = ref(false);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsIdVal = Number(route.params.workspaceId);
    wsId.value = wsIdVal;

    await Promise.all([
      issueStore.fetchStates(wsIdVal, projectId.value),
      issueStore.fetchIssues(wsIdVal, projectId.value),
    ]);
    void loadViewPreferences();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/** 获取某列内的工作项，按 sort_order 升序 */
function issuesInState(stateId: number): Issue[] {
  return issueStore.issues
    .filter((i) => i.state_id === stateId)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
}

// --- 拖拽事件 ---

function onDragStart(issue: Issue, event: DragEvent) {
  dragIssue.value = issue;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", String(issue.id));
    // 设置拖拽预览：一个半透明的小型标识卡（对标 Plane 的 drag ghost）
    const ghost = document.createElement("div");
    ghost.className = "kanban-ghost";
    ghost.textContent = `${issue.identifier} · ${issue.name}`;
    document.body.appendChild(ghost);
    event.dataTransfer.setDragImage(ghost, 12, 12);
    // 清理：setDragImage 后幽灵元素不再需要，但为避免累积在下一帧移除
    requestAnimationFrame(() => ghost.remove());
  }
}

function onDragEnd() {
  dragIssue.value = null;
  dragOverColumn.value = null;
  dropIndex.value = null;
}

function onColumnDragOver(stateId: number, event: DragEvent) {
  event.preventDefault();
  if (!processingDrop.value) {
    dragOverColumn.value = stateId;
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }
}

function onColumnDragLeave(stateId: number) {
  if (dragOverColumn.value === stateId) {
    dragOverColumn.value = null;
  }
}

/** 卡片间拖拽悬停 — 计算插入位置 */
function onCardDragOver(stateId: number, index: number, event: DragEvent) {
  event.preventDefault();
  event.stopPropagation();
  dragOverColumn.value = stateId;

  // 根据鼠标在卡片上的位置确定插入上方还是下方
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const mid = rect.top + rect.height / 2;
  dropIndex.value = event.clientY < mid ? index : index + 1;
}

/** 列容器拖拽悬停 — 检测拖拽到列表末尾空白区（追加到末尾） */
function onCardsDragOver(stateId: number, event: DragEvent) {
  event.preventDefault();
  event.stopPropagation();
  dragOverColumn.value = stateId;
  // 当拖拽位于最后一张卡片之下时，dropIndex 设为列表长度（即追加到末尾）
  const cards = (event.currentTarget as HTMLElement).querySelectorAll(".issue-card");
  if (cards.length === 0) {
    dropIndex.value = 0;
    return;
  }
  const lastCard = cards[cards.length - 1] as HTMLElement;
  const lastRect = lastCard.getBoundingClientRect();
  if (event.clientY > lastRect.bottom) {
    dropIndex.value = cards.length;
  }
}

async function onColumnDrop(stateId: number, event: DragEvent) {
  event.preventDefault();
  const dragged = dragIssue.value;
  if (!dragged || processingDrop.value) return;

  const targetIdx = dropIndex.value;
  dragIssue.value = null;
  dragOverColumn.value = null;
  dropIndex.value = null;

  // 同列且未移动位置 → 无操作
  if (dragged.state_id === stateId && targetIdx === null) return;

  processingDrop.value = true;
  try {
    let updatedIssue: Issue;

    if (dragged.state_id !== stateId) {
      // 跨列流转 — 使用 promiseToast 显示 loading / success / error 三态
      const targetName = issueStore.states.find((s) => s.id === stateId)?.name ?? "";
      updatedIssue = await promiseToast(
        issueApi.transition(wsId.value, projectId.value, dragged.id, stateId),
        {
          loading: "正在流转...",
          success: () => `已流转至「${targetName}」`,
          error: (err) => err instanceof Error ? err.message : "流转失败",
        },
      );
    } else {
      updatedIssue = dragged;
    }

    // 列内排序 — 过滤掉自身后计算中值插入位置
    const columnIssues = issuesInState(stateId).filter((i) => i.id !== dragged.id);
    const insertIdx = Math.min(targetIdx ?? columnIssues.length, columnIssues.length);

    const prevIssue = insertIdx > 0 ? columnIssues[insertIdx - 1] : null;
    const nextIssue = insertIdx < columnIssues.length ? columnIssues[insertIdx] : null;

    // 跨列或同列有位置变化 → 调用 reorder
    if (targetIdx !== null || dragged.state_id !== stateId) {
      updatedIssue = await promiseToast(
        issueApi.reorder(
          wsId.value,
          projectId.value,
          updatedIssue.id,
          prevIssue?.sort_order ?? null,
          nextIssue?.sort_order ?? null,
        ),
        {
          loading: "正在排序...",
          success: () => "排序已更新",
          error: (err) => err instanceof Error ? err.message : "排序失败",
        },
      );
    }

    // 乐观更新：在列表中直接更新 state_id + sort_order
    const idx = issueStore.issues.findIndex((i) => i.id === updatedIssue.id);
    if (idx >= 0) {
      issueStore.issues[idx] = { ...issueStore.issues[idx], ...updatedIssue };
    }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "操作失败";
    toast.error(msg);
    // 回滚：重新拉取以恢复服务端真实状态
    await issueStore.fetchIssues(wsId.value, projectId.value);
  } finally {
    processingDrop.value = false;
  }
}

function openIssue(issueId: number) {
  // 单击打开 Peek 预览抽屉；抽屉内有「打开详情」按钮执行路由跳转
  // 传入当前看板可见的 issue ID 序列，支持抽屉内上一个/下一个连续预览
  const contextIds = Object.values(issueStore.issuesByState)
    .flat()
    .map((i) => i.id);
  peek.openWithContext(Number(route.params.workspaceId), projectId.value, issueId, contextIds);
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

/** 内联更新：更新单个字段并同步本地状态 */
async function inlineUpdate(iss: Issue, patch: Partial<Pick<Issue, "name" | "priority" | "state_id">>) {
  try {
    const updated = await issueApi.updateIssue(wsId.value, projectId.value, iss.id, {
      ...patch,
      version: iss.version,
    } as Parameters<typeof issueApi.updateIssue>[3]);
    const idx = issueStore.issues.findIndex((i) => i.id === iss.id);
    if (idx >= 0) issueStore.issues[idx] = updated;
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "更新失败";
    toast.error(msg);
    // 回滚：重新拉取以恢复服务端真实状态
    await issueStore.fetchIssues(wsId.value, projectId.value);
  }
}

const priorityOptions = [
  { value: "urgent" as IssuePriority, label: "紧急", color: "var(--danger-500)", icon: "🔴" },
  { value: "high" as IssuePriority, label: "高", color: "var(--warning-500)", icon: "🟠" },
  { value: "medium" as IssuePriority, label: "中", color: "var(--brand-500)", icon: "🔵" },
  { value: "low" as IssuePriority, label: "低", color: "var(--text-tertiary)", icon: "⚪" },
  { value: "none" as IssuePriority, label: "无", color: "var(--text-tertiary)", icon: "⬜" },
];

onMounted(() => {
  prefs.setLastView(projectId.value, "board");
  load();
});
</script>

<template>
  <div class="kanban">
    <header class="kanban__header">
      <div>
        <h1>看板</h1>
        <p class="hint">拖拽工作项到不同列进行流转，列内拖拽调整排序</p>
      </div>
      <div class="header-right">
        <div class="view-switcher">
          <router-link
            :to="`/${route.params.workspaceId}/projects/${projectId}/board`"
            class="view-tab is-active"
          >
看板
</router-link>
          <router-link
            :to="`/${route.params.workspaceId}/projects/${projectId}/list`"
            class="view-tab"
          >
列表
</router-link>
        </div>
        <button class="btn btn--primary" @click="showCreateModal = true">+ 创建工作项</button>
      </div>
    </header>

    <AppSkeleton v-if="loading" variant="board" :cols="issueStore.states.length || 4" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState v-else-if="issueStore.issues.length === 0" icon="📋" title="暂无工作项" description="创建或拖拽工作项到此看板">
      <button class="btn btn--primary" @click="showCreateModal = true">+ 创建工作项</button>
    </AppEmptyState>

    <div v-else class="kanban__board" :style="{ '--column-width': columnWidth + 'px' }">
      <div
        v-for="state in issueStore.states"
        :key="state.id"
        class="kanban__column"
        :class="{
          'kanban__column--over': dragOverColumn === state.id && dragIssue?.state_id !== state.id,
          'kanban__column--reorder': dragOverColumn === state.id && dragIssue?.state_id === state.id,
        }"
        @dragover="onColumnDragOver(state.id, $event)"
        @dragleave="onColumnDragLeave(state.id)"
        @drop="onColumnDrop(state.id, $event)"
      >
        <div class="kanban__column-header" :style="{ borderTopColor: state.color }">
          <span class="kanban__column-name">{{ state.name }}</span>
          <span class="kanban__column-count">{{ issuesInState(state.id).length }}</span>
          <!-- 列宽拖拽把手 -->
          <span
            class="kanban__col-resize"
            title="拖拽调整列宽"
            @pointerdown="startColumnResize"
          >⋮</span>
        </div>

        <div
          class="kanban__cards"
          @dragover="onCardsDragOver(state.id, $event)"
        >
          <div
            v-for="(iss, idx) in issuesInState(state.id)"
            :key="iss.id"
            class="issue-card"
            :class="{
              'issue-card--dragging': dragIssue?.id === iss.id,
              'drop-above': dragIssue?.id !== iss.id && dropIndex === idx && dragOverColumn === state.id,
              'drop-below': dragIssue?.id !== iss.id && dropIndex === issuesInState(state.id).length && dragOverColumn === state.id && idx === issuesInState(state.id).length - 1,
            }"
            draggable="true"
            @dragstart="onDragStart(iss, $event)"
            @dragend="onDragEnd"
            @dragover="onCardDragOver(state.id, idx, $event)"
            @click="openIssue(iss.id)"
          >
            <div class="issue-card__header">
              <span class="issue-card__identifier" :style="{ color: priorityColor(iss.priority) }">
                {{ iss.identifier }}
              </span>
              <InlineSelectEdit
                class="issue-card__priority-edit"
                :model-value="iss.priority"
                :options="priorityOptions"
                placeholder="优先级"
                align="right"
                @submit="(v) => inlineUpdate(iss, { priority: v as IssuePriority })"
              />
            </div>
            <div class="issue-card__name" @click.stop>
              <InlineEdit
                :model-value="iss.name"
                trigger="dblclick"
                placeholder="双击编辑名称..."
                @submit="(v) => inlineUpdate(iss, { name: v })"
              />
            </div>
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

          <!-- 拖拽到空白列/列表末尾的插入指示器 -->
          <div
            v-if="issuesInState(state.id).length === 0 && dragOverColumn === state.id"
            class="kanban__drop-empty"
          >
            释放以移动到此列
          </div>

          <div v-if="issuesInState(state.id).length === 0 && dragOverColumn !== state.id" class="kanban__empty">
            暂无工作项
          </div>
        </div>
      </div>
    </div>

    <IssueCreateModal
      :workspace-id="wsId"
      :project-id="projectId"
      :visible="showCreateModal"
      @close="showCreateModal = false"
      @created="load"
    />
  </div>
</template>

<style scoped>
.kanban__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}
.kanban__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}
.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.view-switcher {
  display: flex;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.view-tab {
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
  text-decoration: none;
  background: var(--surface-2);
  transition: background 0.1s;
}

.view-tab + .view-tab {
  border-left: 1px solid var(--border-default);
}

.view-tab:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.view-tab.is-active {
  background: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary {
  padding: 8px 16px;
  background: var(--brand-500);
  color: var(--text-on-brand);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
}
.btn--primary:hover {
  background: var(--brand-600);
}

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
  flex: 0 0 var(--column-width, 280px);
  background: var(--surface-2);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 220px);
  transition: background 0.15s, box-shadow 0.15s;
}

/* 跨列拖拽 - 目标列高亮 */
.kanban__column--over {
  background: var(--brand-50);
  box-shadow: inset 0 0 0 2px var(--brand-400);
  transition: background 0.15s ease, box-shadow 0.15s ease;
}

/* 列内排序 - 轻微高亮 */
.kanban__column--reorder {
  box-shadow: inset 0 0 0 1px var(--brand-300);
  transition: background 0.15s ease, box-shadow 0.15s ease;
}

.kanban__column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-top: 3px solid;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
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

/* 列宽拖拽把手 */
.kanban__col-resize {
  margin-left: 6px;
  padding: 0 4px;
  color: var(--text-tertiary);
  font-size: 12px;
  letter-spacing: 1px;
  cursor: ew-resize;
  border-radius: var(--radius-sm);
  user-select: none;
  transition: background 0.15s, color 0.15s;
}
.kanban__col-resize:hover {
  background: var(--surface-3);
  color: var(--brand-500);
}

.kanban__cards {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.kanban__empty {
  text-align: center;
  color: var(--text-tertiary);
  font-size: 12px;
  padding: 24px 0;
}

.kanban__drop-empty {
  text-align: center;
  color: var(--brand-500);
  font-size: 13px;
  font-weight: 500;
  padding: 24px 8px;
  border: 2px dashed var(--brand-300);
  border-radius: var(--radius-sm);
}

.issue-card {
  padding: 12px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: grab;
  transition: box-shadow 0.15s, opacity 0.15s, transform 0.15s;
  margin-bottom: 2px;
  position: relative;
}
.issue-card:hover {
  box-shadow: var(--shadow-card);
}
.issue-card:active {
  cursor: grabbing;
}

/* 正在拖拽的卡片 */
.issue-card--dragging {
  opacity: 0.4;
  transform: scale(0.97);
}

/* 插入指示线 — 插入到某张卡片上方 */
.drop-above::before {
  content: "";
  position: absolute;
  top: -3px;
  left: 8px;
  right: 8px;
  height: 3px;
  background: var(--brand-500);
  border-radius: 2px;
  z-index: 1;
  box-shadow: 0 0 6px var(--brand-400);
}

/* 插入指示线 — 追加到列表末尾（最后一张卡片下方） */
.drop-below::after {
  content: "";
  position: absolute;
  bottom: -3px;
  left: 8px;
  right: 8px;
  height: 3px;
  background: var(--brand-500);
  border-radius: 2px;
  z-index: 1;
  box-shadow: 0 0 6px var(--brand-400);
}

/* 拖拽幽灵卡片（setDragImage 使用，全局样式因 element 挂在 body 下） */
:global(.kanban-ghost) {
  position: fixed;
  top: -1000px;
  left: -1000px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-primary);
  background: var(--bg-surface-1);
  border: 1px solid var(--border-accent-strong);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-popover);
  white-space: nowrap;
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  pointer-events: none;
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
  display: none;
}

.issue-card__name {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.4;
  margin-bottom: 8px;
}

.issue-card__name :deep(.inline-edit) {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.issue-card__priority-edit {
  font-size: 11px;
  color: var(--text-tertiary);
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
