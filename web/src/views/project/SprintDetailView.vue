<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { sprintApi, type BurndownPoint, type Sprint, type SprintIssueView } from "@/api/services/sprint";
import { workspaceApi } from "@/api/services/workspace";
import BurndownChart from "./BurndownChart.vue";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const sprintId = computed(() => Number(route.params.sprintId));

const sprint = ref<Sprint | null>(null);
const issues = ref<SprintIssueView[]>([]);
const burndown = ref<BurndownPoint[]>([]);
const loading = ref(true);
const error = ref("");
const busy = ref(false);

// complete modal
const showComplete = ref(false);
const strategy = ref<"backlog" | "next_sprint" | "keep">("backlog");
const completeError = ref("");

let wsIdVal = 0;

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const [spRes, issRes, bdRes] = await Promise.all([
      sprintApi.getSprint(wsId, projectId.value, sprintId.value),
      sprintApi.listSprintIssues(wsId, projectId.value, sprintId.value),
      sprintApi.burndown(wsId, projectId.value, sprintId.value).catch(() => ({ points: [] })),
    ]);
    sprint.value = spRes;
    issues.value = issRes.results;
    burndown.value = (bdRes as { points: BurndownPoint[] }).points;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function startSprint() {
  if (!sprint.value) return;
  busy.value = true;
  try {
    const wsId = await resolveWsId();
    sprint.value = await sprintApi.startSprint(wsId, projectId.value, sprint.value.id);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "启动失败";
  } finally {
    busy.value = false;
  }
}

async function submitComplete() {
  if (!sprint.value) return;
  busy.value = true;
  completeError.value = "";
  try {
    const wsId = await resolveWsId();
    sprint.value = await sprintApi.completeSprint(wsId, projectId.value, sprint.value.id, {
      strategy: strategy.value,
      // next_sprint_id 暂不传递（需 select 联动）
    });
    showComplete.value = false;
  } catch (e: unknown) {
    completeError.value = e instanceof Error ? e.message : "结束失败";
  } finally {
    busy.value = false;
  }
}

function statusColor(s: string): string {
  const map: Record<string, string> = {
    planned: "var(--text-tertiary)",
    active: "var(--success-500)",
    completed: "var(--brand-500)",
  };
  return map[s] ?? "var(--text-tertiary)";
}

function fmtDate(s?: string) {
  return s ? s.slice(0, 10) : "?";
}

onMounted(load);
</script>

<template>
  <div class="sprint-detail">
    <header class="header">
      <div class="title-row">
        <button class="back-btn" @click="router.back()">←</button>
        <div>
          <h1 v-if="sprint">{{ sprint.name }}</h1>
          <p v-if="sprint" class="meta">
            <span class="badge" :style="{ color: statusColor(sprint.status) }">
              {{ ({ planned: "未开始", active: "进行中", completed: "已完成" } as Record<string, string>)[sprint.status] }}
            </span>
            <span v-if="sprint.start_date" class="date">{{ fmtDate(sprint.start_date) }} → {{ fmtDate(sprint.end_date) }}</span>
          </p>
        </div>
      </div>
      <div class="actions">
        <template v-if="sprint">
          <button v-if="sprint.status === 'active'" class="btn btn-secondary" @click="router.push(`/${workspaceSlug}/projects/${projectId}/sprints/${sprint.id}/standup`)">
            站会模式
          </button>
          <button v-if="sprint.status === 'planned'" class="btn btn-primary" :disabled="busy" @click="startSprint">
            启动迭代
          </button>
          <button v-if="sprint.status === 'active'" class="btn btn-danger" :disabled="busy" @click="showComplete = true">
            结束迭代
          </button>
          <button v-if="sprint.status === 'completed'" class="btn btn-secondary" disabled>
            已结束
          </button>
        </template>
      </div>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else-if="sprint" class="content">
      <!-- 进度卡片 -->
      <section class="progress-card">
        <h2>进度</h2>
        <div v-if="sprint.progress" class="progress-stats">
          <div class="stat">
            <span class="num">{{ sprint.progress.done_issues }}/{{ sprint.progress.total_issues }}</span>
            <span class="label">已完成工作项</span>
          </div>
          <div class="stat">
            <span class="num">{{ sprint.progress.done_points }}/{{ sprint.progress.total_points }}</span>
            <span class="label">已完成故事点</span>
          </div>
          <div class="stat" v-if="sprint.progress.saturation != null">
            <span class="num" :class="{ over: sprint.progress.saturation > 1 }">{{ Math.round(sprint.progress.saturation * 100) }}%</span>
            <span class="label">饱和度</span>
          </div>
        </div>
        <div v-if="sprint.progress?.by_state_group" class="state-bars">
          <div class="bar-row" v-for="(pts, group) in sprint.progress.by_state_group" :key="group">
            <span class="grp">{{ ({ backlog: "待办", started: "进行中", completed: "已完成", cancelled: "取消" } as Record<string, string>)[group] ?? group }}</span>
            <div class="bar">
              <div class="fill" :class="`fill-${group}`" :style="{ width: sprint.progress!.total_points > 0 ? (pts / sprint.progress!.total_points) * 100 + '%' : '0%' }"></div>
            </div>
            <span class="val">{{ pts }}pt</span>
          </div>
        </div>
      </section>

      <!-- 燃尽图 -->
      <section class="burndown-card">
        <h2>燃尽图</h2>
        <div v-if="burndown.length === 0" class="empty">暂无快照数据</div>
        <BurndownChart v-else :points="burndown" :start-date="sprint.start_date" :end-date="sprint.end_date" />
      </section>

      <!-- 工作项列表 -->
      <section class="issues-card">
        <h2>工作项 ({{ issues.length }})</h2>
        <div class="issues-list">
          <div v-for="iss in issues" :key="iss.issue_id" class="issue-row">
            <span class="type-badge" :class="`type-${iss.type_code}`">
              {{ ({ requirement: "需", task: "任", defect: "缺" } as Record<string, string>)[iss.type_code] }}
            </span>
            <span class="name">{{ iss.name }}</span>
            <span :style="{ color: iss.state_color }" class="state">{{ iss.state_name }}</span>
            <span v-if="iss.point != null" class="point">{{ iss.point }}pt</span>
          </div>
          <div v-if="issues.length === 0" class="empty">暂无工作项。前往 <a @click="router.push(`/${workspaceSlug}/projects/${projectId}/sprints/planning`)">排期规划</a></div>
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
            <input type="radio" v-model="strategy" value="backlog" />
            <div>
              <strong>退回 Backlog</strong>
              <span>未完成任务从当前迭代移除，回到 Backlog 池</span>
            </div>
          </label>
          <label class="radio">
            <input type="radio" v-model="strategy" value="next_sprint" />
            <div>
              <strong>移至下一迭代</strong>
              <span>保留在计划中但需要在排期时重新规划（本版本简化为取消 sprint_id）</span>
            </div>
          </label>
          <label class="radio">
            <input type="radio" v-model="strategy" value="keep" />
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
.badge { font-size: 12px; padding: 2px 8px; border-radius: 12px; background: var(--surface-3); font-weight: 500; }
.date { font-size: 11px; color: var(--text-tertiary); font-family: var(--font-mono); }
.actions { display: flex; gap: 8px; }

.loading, .error, .empty { text-align: center; padding: 24px 0; color: var(--text-tertiary); }
.error { color: var(--danger-500); }
.empty a { color: var(--brand-500); cursor: pointer; }

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

.burndown-svg { background: var(--surface-2); border-radius: var(--radius-sm); }
.burndown-svg .grid { stroke: var(--border-subtle); stroke-width: 1; }
.burndown-svg .axis { stroke: var(--text-tertiary); stroke-width: 1; }
.burndown-svg .axis-label { font-size: 10px; fill: var(--text-tertiary); }
.burndown-svg .ideal { stroke: var(--warning-500); stroke-width: 1.5; stroke-dasharray: 4 3; fill: none; }
.burndown-svg .remaining { stroke: var(--brand-500); stroke-width: 2; fill: none; }
.burndown-svg .done { stroke: var(--success-500); stroke-width: 2; fill: none; }
.burndown-svg .dot { fill: var(--brand-500); }

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
.btn-primary { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.btn-primary:hover:not(:disabled) { background: var(--brand-600); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--surface-2); color: var(--text-primary); }
.btn-secondary:hover { background: var(--surface-3); }
.btn-danger { background: var(--danger-500); color: #fff; border-color: var(--danger-500); }
.btn-danger:hover:not(:disabled) { background: var(--danger-600); }
.btn-danger:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
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

form footer { display: flex; justify-content: flex-end; gap: 8px; margin-top: 12px; }
</style>
