<script setup lang="ts">
/**
 * 排期规划页 — Backlog 与迭代间拖拽分配工作项。
 *
 * 交互能力（对齐互联网大厂产品标准）：
 *  - Backlog ↔ Sprint 双向拖拽（vue-draggable-plus）
 *  - Sprint 内拖拽重排序（幂等 upsert 持久化 sort_order）
 *  - 拖拽视觉反馈（ghost / chosen / drag 三类状态样式）
 *  - 键盘无障碍替代操作（「加入迭代」/「移出迭代」按钮）
 *  - 容量饱和度实时计算与超限警告
 *  - 加载 / 错误 / 空态三态处理
 */

import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { VueDraggable } from "vue-draggable-plus";

import {
  sprintApi,
  type BacklogItem,
  type Sprint,
  type SprintIssueView,
} from "@/api/services/sprint";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import { AppLoadingState, AppErrorState } from "@/components";

/* ------------------------------------------------------------------ */
/* 路由上下文                                                           */
/* ------------------------------------------------------------------ */

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const { wsId, ready } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* 状态                                                                 */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const saving = ref(false);
const error = ref("");

const plannedSprints = ref<Sprint[]>([]);
const activeSprint = ref<Sprint | null>(null);
const backlog = ref<BacklogItem[]>([]);
const sprintIssues = ref<SprintIssueView[]>([]);
const selectedSprintId = ref<number | null>(null);
const capacityStats = ref<{ avg: number; p50: number } | null>(null);

const selectedSprint = computed(() =>
  plannedSprints.value.find((s) => s.id === selectedSprintId.value) ?? activeSprint.value,
);

/** 已完成点数（用于饱和度展示） */
const donePoints = computed(() =>
  sprintIssues.value
    .filter((i) => i.state_group === "completed")
    .reduce((sum, i) => sum + (i.point ?? 0), 0),
);

/** 承诺总点数 */
const totalPoints = computed(() =>
  sprintIssues.value.reduce((sum, i) => sum + (i.point ?? 0), 0),
);

/** 容量饱和度（0~100+），超过 100 标红 */
const saturation = computed(() => {
  const cap = selectedSprint.value?.capacity;
  if (!cap || cap <= 0) return null;
  return Math.round((totalPoints.value / cap) * 100);
});

/* ------------------------------------------------------------------ */
/* 数据加载                                                             */
/* ------------------------------------------------------------------ */

async function loadAll() {
  loading.value = true;
  error.value = "";
  try {
    const wsIdVal = wsId.value;
    const [sprintRes, backlogRes] = await Promise.all([
      sprintApi.listSprints(wsIdVal, projectId.value),
      sprintApi.getBacklog(wsIdVal, projectId.value),
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

    // 速率建议（失败不阻塞页面）
    try {
      const stats = await sprintApi.suggestCapacity(wsIdVal, projectId.value);
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
  const res = await sprintApi.listSprintIssues(wsId.value, projectId.value, sprintId);
  sprintIssues.value = res.results;
}

async function reloadBacklog() {
  const res = await sprintApi.getBacklog(wsId.value, projectId.value);
  backlog.value = res.results;
}

async function pickSprint(sp: Sprint) {
  selectedSprintId.value = sp.id;
  await loadSprintIssues(sp.id);
}

// 上下文就绪后自动加载（路由参数变化时重试）
watch(
  ready,
  (r) => {
    if (r) void loadAll();
  },
  { immediate: true },
);

onMounted(() => {
  if (ready.value) void loadAll();
});

/* ------------------------------------------------------------------ */
/* 拖拽逻辑（vue-draggable-plus）                                       */
/* ------------------------------------------------------------------ */

/** 通过拖拽元素上的 data-list 标记判断来源/目标面板 */
const LIST_BACKLOG = "backlog";
const LIST_SPRINT = "sprint";

const backlogListId = ref(LIST_BACKLOG);
const sprintListId = ref(LIST_SPRINT);

function listIdOf(el: HTMLElement | null | undefined): string | null {
  return el?.getAttribute?.("data-list") ?? null;
}

/** 幂等同步迭代内工作项排序（ON CONFLICT upsert sort_order） */
async function syncSortOrder() {
  if (!selectedSprint.value) return;
  const sprintId = selectedSprint.value.id;
  for (let i = 0; i < sprintIssues.value.length; i++) {
    await sprintApi.addIssue(wsId.value, projectId.value, sprintId, sprintIssues.value[i].issue_id, (i + 1) * 1000);
  }
}

/** 拖拽结束统一处理 */
async function onDragEnd(evt: {
  from: HTMLElement;
  to: HTMLElement;
  oldIndex?: number;
  newIndex?: number;
  item: HTMLElement;
}) {
  const fromList = listIdOf(evt.from);
  const toList = listIdOf(evt.to);
  if (!fromList || !toList) return;

  saving.value = true;
  error.value = "";
  try {
    if (fromList === LIST_BACKLOG && toList === LIST_SPRINT) {
      // Backlog → Sprint：加入迭代（v-model 已移动数组，此处持久化 + 刷新结构）
      await syncSortOrder();
      await loadSprintIssues(selectedSprint.value!.id);
      await reloadBacklog();
    } else if (fromList === LIST_SPRINT && toList === LIST_BACKLOG) {
      // Sprint → Backlog：移出迭代
      const removed = sprintIssues.value[evt.oldIndex ?? 0];
      if (removed) {
        await sprintApi.removeIssue(wsId.value, projectId.value, selectedSprintId.value!, removed.issue_id);
        sprintIssues.value = sprintIssues.value.filter((i) => i.issue_id !== removed.issue_id);
        await reloadBacklog();
      }
    } else if (fromList === LIST_SPRINT && toList === LIST_SPRINT) {
      // Sprint 内重排
      await syncSortOrder();
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "拖拽操作失败";
    // 失败时回滚：重新加载两端数据
    await Promise.allSettled([
      selectedSprint.value ? loadSprintIssues(selectedSprint.value.id) : Promise.resolve(),
      reloadBacklog(),
    ]);
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 键盘无障碍替代操作                                                    */
/* ------------------------------------------------------------------ */

async function moveToSprint(item: BacklogItem) {
  if (!selectedSprint.value) {
    error.value = "请先选择或创建一个迭代";
    return;
  }
  saving.value = true;
  try {
    await sprintApi.addIssue(wsId.value, projectId.value, selectedSprint.value.id, item.issue_id);
    backlog.value = backlog.value.filter((b) => b.issue_id !== item.issue_id);
    await loadSprintIssues(selectedSprint.value.id);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "添加失败";
  } finally {
    saving.value = false;
  }
}

async function moveToBacklog(item: SprintIssueView) {
  if (!selectedSprintId.value) return;
  saving.value = true;
  try {
    await sprintApi.removeIssue(wsId.value, projectId.value, selectedSprintId.value, item.issue_id);
    sprintIssues.value = sprintIssues.value.filter((i) => i.issue_id !== item.issue_id);
    await reloadBacklog();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "移除失败";
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 工具                                                                 */
/* ------------------------------------------------------------------ */

const typeLabel: Record<string, string> = { requirement: "需", task: "任", defect: "缺" };

function openSprint(sp: Sprint) {
  router.push(`/${workspaceSlug.value}/projects/${projectId.value}/sprints/${sp.id}`);
}
</script>

<template>
  <div class="planning">
    <header class="header">
      <div>
        <h1>排期规划</h1>
        <p class="hint">从左侧 Backlog 拖拽工作项到右侧迭代，迭代内可拖拽排序</p>
      </div>
      <div class="actions">
        <span v-if="capacityStats" class="capacity-badge" title="基于历史迭代完成点数的速率建议">
          速率: avg {{ capacityStats.avg }}pt / P50 {{ capacityStats.p50 }}pt
        </span>
        <button class="btn btn-secondary" @click="router.push(`/${workspaceSlug}/projects/${projectId}/sprints`)">
          返回列表
        </button>
      </div>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadAll" />

    <div v-else class="content">
      <!-- 左：Backlog -->
      <section class="panel backlog-panel" aria-label="Backlog 工作项">
        <div class="panel-header">
          <h2>Backlog</h2>
          <span class="count">{{ backlog.length }} 项</span>
        </div>

        <VueDraggable
          v-model="backlog"
          class="panel-body panel-drop"
          group="sprint-planning"
          :animation="150"
          :ghost-class="'ghost'"
          :chosen-class="'chosen'"
          :drag-class="'dragging'"
          :data-list="backlogListId"
          :disabled="saving"
          @end="onDragEnd"
        >
          <div
            v-for="item in backlog"
            :key="item.issue_id"
            class="issue-item"
            :aria-grabbed="false"
            :tabindex="0"
            @keydown.enter="moveToSprint(item)"
          >
            <div class="title-row">
              <span class="type-badge" :class="`type-${item.type_code}`">
                {{ typeLabel[item.type_code] ?? "?" }}
              </span>
              <span class="name">{{ item.name }}</span>
              <button
                class="icon-btn"
                title="加入选中的迭代"
                :aria-label="`将 ${item.name} 加入迭代`"
                :disabled="saving || !selectedSprint"
                @click.stop="moveToSprint(item)"
              >
                ➕
              </button>
            </div>
            <div class="meta">
              <span :style="{ color: item.state_color }">{{ item.state_name }}</span>
              <span v-if="item.point != null" class="point">{{ item.point }}pt</span>
              <span v-if="item.priority === 'urgent'" class="priority-dot urgent" title="紧急"></span>
            </div>
          </div>

          <template #footer>
            <div v-if="backlog.length === 0" class="empty">Backlog 为空</div>
          </template>
        </VueDraggable>
      </section>

      <!-- 右：Sprint -->
      <section class="panel sprint-panel" aria-label="迭代工作项">
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
          <span v-if="activeSprint" class="tab active" style="cursor: default;">
            {{ activeSprint.name }} (active)
          </span>
        </div>

        <div v-if="selectedSprint" class="panel-header">
          <div>
            <h2>{{ selectedSprint.name }}</h2>
            <p v-if="selectedSprint.goal" class="goal">{{ selectedSprint.goal }}</p>
          </div>
          <div class="stats">
            <div class="stat">
              <span class="num">{{ donePoints }}</span>
              <span class="label">已完成</span>
            </div>
            <div class="stat">
              <span class="num">{{ totalPoints }}</span>
              <span class="label">承诺</span>
            </div>
            <div v-if="saturation !== null" class="stat">
              <span class="num" :class="{ over: saturation > 100 }" :title="saturation > 100 ? '已超出容量，请减少承诺点数' : ''">
                {{ saturation }}%
              </span>
              <span class="label">饱和度</span>
            </div>
          </div>
        </div>

        <VueDraggable
          v-if="selectedSprint"
          v-model="sprintIssues"
          class="panel-body panel-drop"
          group="sprint-planning"
          :animation="150"
          :ghost-class="'ghost'"
          :chosen-class="'chosen'"
          :drag-class="'dragging'"
          :data-list="sprintListId"
          :disabled="saving"
          @end="onDragEnd"
        >
          <div
            v-for="iss in sprintIssues"
            :key="iss.issue_id"
            class="issue-item"
            :aria-grabbed="false"
            :tabindex="0"
            @keydown.enter="moveToBacklog(iss)"
          >
            <div class="title-row">
              <span class="type-badge" :class="`type-${iss.type_code}`">
                {{ typeLabel[iss.type_code] ?? "?" }}
              </span>
              <span class="name">{{ iss.name }}</span>
              <button
                class="icon-btn"
                title="移出迭代到 Backlog"
                :aria-label="`将 ${iss.name} 移出迭代`"
                :disabled="saving"
                @click.stop="moveToBacklog(iss)"
              >
                ➖
              </button>
            </div>
            <div class="meta">
              <span :style="{ color: iss.state_color }">{{ iss.state_name }}</span>
              <span v-if="iss.point != null" class="point">{{ iss.point }}pt</span>
            </div>
          </div>

          <template #footer>
            <div v-if="sprintIssues.length === 0" class="empty">
              拖拽工作项到此迭代中
            </div>
          </template>
        </VueDraggable>

        <div v-if="!selectedSprint" class="panel-body">
          <div class="empty">请先选择或创建一个迭代</div>
        </div>

        <div v-if="selectedSprint" class="panel-footer">
          <button class="btn btn-primary" :disabled="saving" @click="openSprint(selectedSprint)">
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

.state { text-align: center; padding: 48px 0; color: var(--text-tertiary); }
.state-error { color: var(--danger-500); }

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

/* ---- 拖拽视觉反馈 ---- */
.ghost {
  opacity: 0.4;
  background: var(--brand-50) !important;
  border: 1px dashed var(--brand-500) !important;
}
.chosen { box-shadow: var(--shadow-card, 0 2px 8px rgba(0,0,0,0.12)); }
.dragging { opacity: 0.9; transform: scale(1.01); }

.issue-item {
  padding: 8px 10px; background: var(--surface-1);
  border: 1px solid var(--border-subtle); border-radius: var(--radius-sm);
  cursor: grab; transition: box-shadow 0.15s;
  display: flex; flex-direction: column; gap: 4px;
}
.issue-item:hover { box-shadow: var(--shadow-card); }
.issue-item:active { cursor: grabbing; }
.issue-item:focus-visible { outline: 2px solid var(--brand-500); outline-offset: 1px; }

.title-row { display: flex; align-items: center; gap: 6px; }
.name { font-size: 12px; font-weight: 500; color: var(--text-primary); flex: 1;
  display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }

.icon-btn {
  border: none; background: none; font-size: 12px; cursor: pointer;
  opacity: 0.45; padding: 2px 4px; border-radius: 4px;
  transition: opacity 0.15s, background 0.15s;
}
.icon-btn:hover { opacity: 1; background: var(--surface-3); }
.icon-btn:disabled { opacity: 0.25; cursor: not-allowed; }

.type-badge {
  font-size: 9px; padding: 1px 5px; border-radius: 3px;
  background: var(--surface-3); color: var(--text-secondary); font-weight: 600;
  min-width: 18px; text-align: center; flex-shrink: 0;
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
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover:not(:disabled) { background: var(--brand-600); }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
</style>
