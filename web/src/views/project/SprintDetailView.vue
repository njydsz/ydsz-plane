<script setup lang="ts">
/**
 * 迭代详情页 — 展示进度、工作项列表、燃尽图、复盘与生命周期操作。
 *
 * P1 增强：
 *  - 编辑迭代（仅 planned 状态，复用表单）
 *  - 结束迭代时支持「移至下一迭代」策略（next_sprint_id 联动选择）
 *  - Active 状态快捷入口：站会模式 / 排期规划
 *  - 燃尽图 loading / error 态透传
 */

import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  sprintApi,
  type BurndownPoint,
  type Sprint,
  type SprintIssueView,
  type SprintVelocity,
} from "@/api/services/sprint";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import SprintStatusBadge from "@/components/sprint/SprintStatusBadge.vue";
import SprintAnalyticsView from "./SprintAnalyticsView.vue";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

/* ------------------------------------------------------------------ */
/* 路由上下文                                                           */
/* ------------------------------------------------------------------ */

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const sprintId = computed(() => Number(route.params.sprintId));
const { wsId, ready } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* 状态                                                                 */
/* ------------------------------------------------------------------ */

const sprint = ref<Sprint | null>(null);
const issues = ref<SprintIssueView[]>([]);
const burndown = ref<BurndownPoint[]>([]);
const velocity = ref<SprintVelocity[]>([]);
const velocityAvg = ref(0);
const loading = ref(true);
const error = ref("");
const busy = ref(false);

/** planned 迭代列表（供 next_sprint 选择与编辑校验） */
const plannedSprints = ref<Sprint[]>([]);

/* ------------------------------------------------------------------ */
/* 数据加载                                                             */
/* ------------------------------------------------------------------ */

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [spRes, issRes, bdRes, velRes] = await Promise.all([
      sprintApi.getSprint(wsId.value, projectId.value, sprintId.value),
      sprintApi.listSprintIssues(wsId.value, projectId.value, sprintId.value),
      sprintApi.burndown(wsId.value, projectId.value, sprintId.value).catch(() => ({ points: [] as BurndownPoint[] })),
      sprintApi.suggestCapacity(wsId.value, projectId.value).catch(() => ({ avg_points: 0, avg_issues: 0, p50: 0, recent_sprints: [] as SprintVelocity[], count: 0 })),
    ]);
    sprint.value = spRes;
    issues.value = issRes.results;
    burndown.value = bdRes.points;
    velocity.value = velRes.recent_sprints || [];
    velocityAvg.value = velRes.avg_points || 0;

    // 加载 planned 迭代（用于 next_sprint 联动）
    const listRes = await sprintApi.listSprints(wsId.value, projectId.value, { status: "planned" });
    plannedSprints.value = listRes.results.filter((s) => s.id !== sprintId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期：启动                                                       */
/* ------------------------------------------------------------------ */

async function startSprint() {
  if (!sprint.value) return;
  busy.value = true;
  error.value = "";
  try {
    sprint.value = await sprintApi.startSprint(wsId.value, projectId.value, sprint.value.id);
    await load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "启动失败";
  } finally {
    busy.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期：结束（含 next_sprint 联动）                                 */
/* ------------------------------------------------------------------ */

const showComplete = ref(false);
const strategy = ref<"backlog" | "next_sprint" | "keep">("backlog");
const nextSprintId = ref<number | null>(null);
const completeError = ref("");

function openComplete() {
  strategy.value = "backlog";
  nextSprintId.value = plannedSprints.value[0]?.id ?? null;
  completeError.value = "";
  showComplete.value = true;
}

async function submitComplete() {
  if (!sprint.value) return;
  if (strategy.value === "next_sprint" && nextSprintId.value == null) {
    completeError.value = "请选择目标迭代";
    return;
  }
  busy.value = true;
  completeError.value = "";
  try {
    sprint.value = await sprintApi.completeSprint(wsId.value, projectId.value, sprint.value.id, {
      strategy: strategy.value,
      next_sprint_id: strategy.value === "next_sprint" ? nextSprintId.value ?? undefined : undefined,
    });
    showComplete.value = false;
    await load();
  } catch (e: unknown) {
    completeError.value = e instanceof Error ? e.message : "结束失败";
  } finally {
    busy.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 编辑迭代                                                             */
/* ------------------------------------------------------------------ */

const showEdit = ref(false);
const editError = ref("");
const editForm = ref({
  name: "",
  description: "",
  goal: "",
  start_date: "",
  end_date: "",
  capacity: undefined as number | undefined,
});

function openEdit() {
  if (!sprint.value) return;
  editError.value = "";
  editForm.value = {
    name: sprint.value.name,
    description: sprint.value.description ?? "",
    goal: sprint.value.goal ?? "",
    start_date: sprint.value.start_date ?? "",
    end_date: sprint.value.end_date ?? "",
    capacity: sprint.value.capacity,
  };
  showEdit.value = true;
}

async function saveEdit() {
  if (!sprint.value) return;
  if (!editForm.value.name.trim()) {
    editError.value = "迭代名称不能为空";
    return;
  }
  busy.value = true;
  editError.value = "";
  try {
    await sprintApi.updateSprint(wsId.value, projectId.value, sprint.value.id, {
      name: editForm.value.name,
      description: editForm.value.description || undefined,
      goal: editForm.value.goal || undefined,
      start_date: editForm.value.start_date || undefined,
      end_date: editForm.value.end_date || undefined,
      capacity: editForm.value.capacity,
      version: 0,
    });
    showEdit.value = false;
    await load();
  } catch (e: unknown) {
    editError.value = e instanceof Error ? e.message : "保存失败";
  } finally {
    busy.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 工具                                                                 */
/* ------------------------------------------------------------------ */

const typeLabel: Record<string, string> = { epic: "史", requirement: "需", task: "任", defect: "缺" };

function fmtDate(s?: string) {
  return s ? s.slice(0, 10) : "?";
}

function goPlanning() {
  router.push(`/${workspaceId.value}/projects/${projectId.value}/sprints/planning`);
}

onMounted(() => {
  if (ready.value) void load();
});
watch(ready, (r) => {
  if (r) void load();
});
</script>

<template>
  <div class="sprint-detail">
    <header class="header">
      <div class="title-row">
        <button class="back-btn" @click="router.back()">←</button>
        <div>
          <h1 v-if="sprint">{{ sprint.name }}</h1>
          <p v-if="sprint" class="meta">
            <SprintStatusBadge :status="sprint.status" />
            <span v-if="sprint.start_date" class="date">{{ fmtDate(sprint.start_date) }} → {{ fmtDate(sprint.end_date) }}</span>
          </p>
        </div>
      </div>
      <div class="actions">
        <template v-if="sprint">
          <!-- planned：编辑 + 启动 -->
          <template v-if="sprint.status === 'planned'">
            <button class="btn btn-secondary" :disabled="busy" @click="openEdit">编辑</button>
            <button class="btn btn-primary" :disabled="busy" @click="startSprint">
              {{ busy ? "处理中..." : "启动迭代" }}
            </button>
          </template>
          <!-- active：站会 / 排期规划 / 结束 -->
          <template v-else-if="sprint.status === 'active'">
            <button class="btn btn-secondary" @click="router.push(`/${workspaceId}/projects/${projectId}/sprints/${sprint.id}/standup`)">
              站会模式
            </button>
            <button class="btn btn-secondary" @click="goPlanning">排期规划</button>
            <button class="btn btn-danger" :disabled="busy" @click="openComplete">结束迭代</button>
          </template>
          <button v-else class="btn btn-secondary" disabled>已结束</button>
        </template>
      </div>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!sprint"
      title="迭代不存在或已被删除"
      description="请检查迭代 ID 是否正确"
    >
      <button class="btn btn-secondary" @click="router.back()">返回</button>
    </AppEmptyState>

    <div v-else class="content">
      <!-- 迭代效能分析面板（指标 + 燃尽/燃起 + 速率 + 状态分布） -->
      <section class="analytics-section">
        <SprintAnalyticsView
          :sprint="sprint"
          :burndown="burndown"
          :velocity="velocity"
          :velocity-avg="velocityAvg"
        />
      </section>

      <!-- 工作项列表 -->
      <section class="issues-card">
        <h2>工作项 ({{ issues.length }})</h2>
        <div class="issues-list">
          <div v-for="iss in issues" :key="iss.issue_id" class="issue-row">
            <span class="type-badge" :class="`type-${iss.type_code}`">
              {{ typeLabel[iss.type_code] ?? "?" }}
            </span>
            <span class="name">{{ iss.name }}</span>
            <span :style="{ color: iss.state_color }" class="state">{{ iss.state_name }}</span>
            <span v-if="iss.point != null" class="point">{{ iss.point }}pt</span>
          </div>
          <div v-if="issues.length === 0" class="empty">暂无工作项。前往 <a @click="goPlanning">排期规划</a></div>
        </div>
      </section>

      <!-- 复盘数据 -->
      <section v-if="sprint.review_snapshot" class="review-card">
        <h2>复盘数据</h2>
        <div class="review-stats">
          <div class="stat"><span class="num">{{ (sprint.review_snapshot.completion_rate * 100).toFixed(0) }}%</span><span class="label">完成率</span></div>
          <div class="stat"><span class="num">{{ sprint.review_snapshot.completed_points }}/{{ sprint.review_snapshot.committed_points }}</span><span class="label">完成/承诺 故事点</span></div>
          <div class="stat"><span class="num">{{ sprint.review_snapshot.completed_issues }}/{{ sprint.review_snapshot.committed_issues }}</span><span class="label">完成/承诺 工作项</span></div>
          <div class="stat"><span class="num">+{{ sprint.review_snapshot.joined_issues }}</span><span class="label">中途加入</span></div>
        </div>
      </section>
    </div>

    <!-- 编辑迭代 modal -->
    <div v-if="showEdit" class="modal-overlay" @click.self="showEdit = false">
      <div class="modal">
        <header>
          <h2>编辑迭代</h2>
          <button class="close" @click="showEdit = false">×</button>
        </header>
        <form @submit.prevent="saveEdit">
          <label>名称 <span class="req">*</span>
            <input v-model="editForm.name" maxlength="80" />
          </label>
          <label>目标
            <input v-model="editForm.goal" maxlength="500" />
          </label>
          <div class="row">
            <label>开始日期 <input v-model="editForm.start_date" type="date" /></label>
            <label>结束日期 <input v-model="editForm.end_date" type="date" /></label>
          </div>
          <label>容量（故事点）
            <input v-model.number="editForm.capacity" type="number" min="0" step="1" />
          </label>
          <div v-if="editError" class="error">{{ editError }}</div>
          <footer>
            <button type="button" class="btn btn-secondary" @click="showEdit = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="busy">
              {{ busy ? "保存中..." : "保存" }}
            </button>
          </footer>
        </form>
      </div>
    </div>

    <!-- 结束迭代 modal -->
    <div v-if="showComplete" class="modal-overlay" @click.self="showComplete = false">
      <div class="modal">
        <header>
          <h2>结束迭代</h2>
          <button class="close" @click="showComplete = false">×</button>
        </header>
        <form @submit.prevent="submitComplete">
          <p class="hint">选择未完成任务的处理策略：</p>
          <label class="radio">
            <input v-model="strategy" type="radio" value="backlog" />
            <div>
              <strong>退回 Backlog</strong>
              <span>未完成任务从当前迭代移除，回到 Backlog 池</span>
            </div>
          </label>
          <label class="radio">
            <input v-model="strategy" type="radio" value="next_sprint" />
            <div>
              <strong>移至下一迭代</strong>
              <span>未完成任务结转至指定的计划中迭代</span>
            </div>
          </label>
          <label v-if="strategy === 'next_sprint'" class="next-select">
            目标迭代
            <select v-model.number="nextSprintId">
              <option v-for="sp in plannedSprints" :key="sp.id" :value="sp.id">{{ sp.name }}</option>
              <option v-if="plannedSprints.length === 0" value="" disabled>暂无计划中的迭代，请先创建</option>
            </select>
          </label>
          <label class="radio">
            <input v-model="strategy" type="radio" value="keep" />
            <div>
              <strong>仅归档</strong>
              <span>保留在原迭代但关闭关联</span>
            </div>
          </label>

          <div v-if="completeError" class="error">{{ completeError }}</div>
          <footer>
            <button type="button" class="btn btn-secondary" @click="showComplete = false">取消</button>
            <button type="submit" class="btn btn-danger" :disabled="busy">
              {{ busy ? "处理中..." : "确认结束" }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 16px; gap: 12px; flex-wrap: wrap;
}
.title-row { display: flex; gap: 10px; }
.back-btn {
  width: 32px; height: 32px; border-radius: 50%;
  border: 1px solid var(--border-subtle); background: var(--surface-2);
  cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 16px;
}
.back-btn:hover { background: var(--surface-3); }
.header h1 { margin: 0; font-size: 20px; }
.meta { display: flex; gap: 8px; align-items: center; margin: 4px 0 0; }
.date { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }

.loading, .error, .empty { text-align: center; padding: 24px 0; color: var(--text-tertiary); }
.error { color: var(--danger-500); }
.empty a { color: var(--brand-500); cursor: pointer; }

.center-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  gap: 12px;
  color: var(--text-tertiary);
}
.center-message.error { color: var(--danger-500); }
.center-message.empty { color: var(--text-secondary); }
.center-message p { margin: 0; font-size: 14px; }

.skeleton-card {
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
}
.skeleton-line {
  height: 14px;
  background: var(--surface-2);
  border-radius: 4px;
  animation: pulse 1.5s infinite;
}
.skeleton-line:last-child { height: 30px; }
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

.content { display: flex; flex-direction: column; gap: 14px; }

.progress-card, .burndown-card, .issues-card, .review-card {
  padding: 14px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

section h2 { margin: 0 0 12px; font-size: 14px; font-weight: 600; }

.progress-stats, .review-stats {
  display: grid; grid-template-columns: repeat(auto-fit, minmax(100px, 1fr)); gap: 12px;
}
.stat { display: flex; flex-direction: column; align-items: center; padding: 10px; background: var(--surface-2); border-radius: var(--radius-sm); }
.stat .num { font-size: 16px; font-weight: 600; font-family: var(--font-mono); }
.stat .num.over { color: var(--danger-500); }
.stat .label { font-size: 10px; color: var(--text-tertiary); margin-top: 2px; }

.state-bars { margin-top: 12px; display: flex; flex-direction: column; gap: 6px; }
.bar-row { display: flex; align-items: center; gap: 8px; font-size: 11px; }
.grp { width: 60px; color: var(--text-secondary); text-align: right; }
.bar { flex: 1; height: 8px; background: var(--surface-3); border-radius: 4px; overflow: hidden; }
.fill { height: 100%; border-radius: 4px; transition: width 0.4s; }
.fill-backlog { background: var(--brand-300); }
.fill-started { background: var(--warning-500); }
.fill-completed { background: var(--success-500); }
.fill-cancelled { background: var(--text-tertiary); }
.val { width: 40px; font-family: var(--font-mono); color: var(--text-tertiary); }

.issues-list { display: flex; flex-direction: column; gap: 4px; }
.issue-row {
  display: flex; align-items: center; gap: 8px; padding: 6px 8px;
  background: var(--surface-2); border-radius: var(--radius-sm); font-size: 12px;
}
.issue-row .name { flex: 1; color: var(--text-primary); }
.issue-row .state { color: var(--text-secondary); font-size: 11px; font-weight: 500; }
.issue-row .point { font-family: var(--font-mono); color: var(--text-tertiary); font-size: 11px; }

.type-badge {
  font-size: 9px; padding: 1px 5px; border-radius: 3px;
  background: var(--surface-3); color: var(--text-secondary); font-weight: 600;
  min-width: 18px; text-align: center;
}
.type-badge.type-defect { background: var(--danger-50); color: var(--danger-600); }
.type-badge.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.type-badge.type-task { background: var(--success-50); color: var(--success-600); }

.btn {
  font-size: 13px; font-weight: 500; padding: 6px 14px; border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle); cursor: pointer; transition: background 0.15s;
}
.btn-primary { background: var(--brand-500); color: var(--text-on-brand); border-color: var(--brand-500); }
.btn-primary:hover:not(:disabled) { background: var(--brand-600); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
.btn-danger { background: var(--danger-500); color: var(--text-on-brand); border-color: var(--danger-500); }
.btn-danger:hover:not(:disabled) { background: var(--danger-600); }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-overlay {
  position: fixed; inset: 0; background: var(--bg-backdrop);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: var(--surface-1); padding: 20px; border-radius: var(--radius-md);
  width: 460px; max-width: 90vw; box-shadow: var(--shadow-elevated);
}
.modal header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.modal h2 { margin: 0; font-size: 16px; }
.close { background: none; border: none; font-size: 22px; cursor: pointer; line-height: 1; color: var(--text-tertiary); }

.hint { font-size: 12px; color: var(--text-tertiary); }
.radio {
  display: flex; align-items: flex-start; gap: 8px; padding: 8px 10px;
  background: var(--surface-2); border-radius: var(--radius-sm); margin-bottom: 8px; cursor: pointer;
}
.radio div { display: flex; flex-direction: column; gap: 2px; font-size: 12px; }
.radio span { color: var(--text-tertiary); font-size: 11px; }

.next-select {
  display: flex; flex-direction: column; gap: 4px; font-size: 12px;
  color: var(--text-secondary); margin: -4px 0 8px 26px;
}
.next-select select {
  font-size: 13px; padding: 6px 10px; border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm); background: var(--surface-2); color: var(--text-primary);
}

form { display: flex; flex-direction: column; gap: 10px; }
label { font-size: 12px; color: var(--text-secondary); display: flex; flex-direction: column; gap: 4px; }
.req { color: var(--danger-500); }
input {
  font-size: 13px; padding: 6px 10px; border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm); background: var(--surface-2); color: var(--text-primary);
  font-family: inherit;
}
input:focus { outline: none; border-color: var(--brand-500); }

.row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }

form footer { display: flex; justify-content: flex-end; gap: 8px; margin-top: 12px; }
</style>
