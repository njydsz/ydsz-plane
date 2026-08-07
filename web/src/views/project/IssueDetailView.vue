<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { issueApi, type Issue, type IssueActivity, type State, type TimeLog } from "@/api/services/issue";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const props = defineProps<{
  workspaceSlug: string;
  projectId: number;
  issueId: number;
}>();

const route = useRoute();
const router = useRouter();

const ws = ref<Workspace | null>(null);
const issue = ref<Issue | null>(null);
const states = ref<State[]>([]);
const activities = ref<IssueActivity[]>([]);
const loading = ref(true);
const error = ref("");
const transitionError = ref("");
const showTransitionMenu = ref(false);

// --- 工时 ---
const timeLogs = ref<TimeLog[]>([]);
const totalMinutes = ref(0);
const showTimeLogForm = ref(false);
const newSpentDate = ref(new Date().toISOString().slice(0, 10));
const newDurationHours = ref(1);
const newDurationMinutes = ref(0);
const newTimeDesc = ref("");
const timeLogError = ref("");
const timeLogSubmitting = ref(false);
const timeLogsLoading = ref(false);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.getBySlug(props.workspaceSlug);
    const [iss, st, acts] = await Promise.all([
      issueApi.getIssue(ws.value.id, props.projectId, props.issueId),
      issueApi.listStates(ws.value.id, props.projectId),
      issueApi.listActivities(ws.value.id, props.projectId, props.issueId),
    ]);
    issue.value = iss;
    states.value = st;
    activities.value = acts.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function doTransition(toStateId: number) {
  if (!ws.value || !issue.value) return;
  transitionError.value = "";
  showTransitionMenu.value = false;
  try {
    issue.value = await issueApi.transition(ws.value.id, props.projectId, props.issueId, toStateId);
  } catch (e: unknown) {
    transitionError.value = e instanceof Error ? e.message : "流转失败";
  }
}

async function doDelete() {
  if (!ws.value || !confirm("确定要归档该工作项吗？")) return;
  try {
    await issueApi.deleteIssue(ws.value.id, props.projectId, props.issueId);
    router.push(`/${props.workspaceSlug}/projects/${props.projectId}/board`);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除失败";
  }
}

function goBack() {
  router.push(`/${props.workspaceSlug}/projects/${props.projectId}/board`);
}

function stateName(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.name ?? "未知";
}

function stateColor(stateId: number): string {
  return states.value.find((s) => s.id === stateId)?.color ?? "#8DA2C2";
}

function typeLabel(type: string): string {
  return ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[type] ?? type;
}

const availableTransitions = computed(() => {
  if (!issue.value) return [];
  return states.value.filter((s) => s.id !== issue.value!.state_id);
});

async function loadTimeLogs() {
  if (!ws.value) return;
  timeLogsLoading.value = true;
  try {
    const res = await issueApi.listTimeLogs(ws.value.id, props.projectId, props.issueId);
    timeLogs.value = res.results;
    totalMinutes.value = res.results.reduce((sum, tl) => sum + tl.duration_minutes, 0);
  } catch {
    // 非关键模块，静默失败
  } finally {
    timeLogsLoading.value = false;
  }
}

async function submitTimeLog() {
  if (!ws.value || timeLogSubmitting.value) return;
  const totalMins = newDurationHours.value * 60 + newDurationMinutes.value;
  if (totalMins <= 0 || totalMins > 1440) {
    timeLogError.value = "请填写有效的工时（1分钟-24小时）";
    return;
  }
  timeLogSubmitting.value = true;
  timeLogError.value = "";
  try {
    await issueApi.createTimeLog(ws.value.id, props.projectId, props.issueId, {
      spent_date: newSpentDate.value,
      duration_minutes: totalMins,
      description: newTimeDesc.value.trim() || undefined,
    });
    showTimeLogForm.value = false;
    newTimeDesc.value = "";
    newDurationHours.value = 1;
    newDurationMinutes.value = 0;
    await loadTimeLogs();
  } catch (e: unknown) {
    timeLogError.value = e instanceof Error ? e.message : "记录失败";
  } finally {
    timeLogSubmitting.value = false;
  }
}

function fmtDuration(mins: number): string {
  if (mins < 60) return `${mins}分钟`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`;
}

onMounted(() => {
  load();
  loadTimeLogs();
});
</script>

<template>
  <div class="issue-detail">
    <header class="issue-detail__header">
      <button class="btn btn--ghost" @click="goBack">← 返回看板</button>
      <div class="issue-detail__actions">
        <button class="btn btn--danger" @click="doDelete">归档</button>
      </div>
    </header>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else-if="issue" class="issue-detail__body">
      <div class="issue-detail__main">
        <div class="issue-detail__identifier">{{ issue.identifier }}</div>
        <h1 class="issue-detail__name">{{ issue.name }}</h1>

        <div class="issue-detail__meta-row">
          <span class="badge" :class="`badge-${issue.type_code}`">{{ typeLabel(issue.type_code) }}</span>
          <span
            class="issue-detail__state-badge"
            :style="{ backgroundColor: stateColor(issue.state_id) }"
          >
            {{ stateName(issue.state_id) }}
          </span>
          <span class="issue-detail__priority">
            优先级: {{ ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[issue.priority] ?? issue.priority }}
          </span>
          <span v-if="issue.severity" class="issue-detail__field">严重度: S{{ issue.severity }}</span>
          <span v-if="issue.found_phase" class="issue-detail__field">发现阶段: {{ issue.found_phase }}</span>
          <span v-if="issue.point != null" class="issue-detail__field">点数: {{ issue.point }}</span>
        </div>

        <div class="issue-detail__section">
          <h3>描述</h3>
          <div v-if="issue.description_html" class="issue-detail__desc" v-html="issue.description_html"></div>
          <p v-else class="text-muted">暂无描述</p>
        </div>

        <div v-if="issue.type_code === 'defect'" class="issue-detail__section">
          <h3>缺陷信息</h3>
          <div class="issue-detail__fields">
            <div v-if="issue.reproduce_steps" class="field-row">
              <span class="field-label">复现步骤:</span>
              <span class="field-value">{{ JSON.stringify(issue.reproduce_steps) }}</span>
            </div>
            <div v-if="issue.environment" class="field-row">
              <span class="field-label">环境:</span>
              <span class="field-value">{{ JSON.stringify(issue.environment) }}</span>
            </div>
            <div v-if="issue.root_cause_category" class="field-row">
              <span class="field-label">根因分类:</span>
              <span class="field-value">{{ issue.root_cause_category }}</span>
            </div>
          </div>
        </div>

        <!-- 流转操作 -->
        <div class="issue-detail__section">
          <h3>状态流转</h3>
          <div v-if="transitionError" class="form-error">{{ transitionError }}</div>
          <div class="issue-detail__transitions">
            <button
              v-for="st in availableTransitions"
              :key="st.id"
              class="btn btn--state"
              :style="{ borderColor: st.color, color: st.color }"
              @click="doTransition(st.id)"
            >
              → {{ st.name }}
            </button>
          </div>
        </div>
      </div>

      <!-- 侧边栏：活动日志 -->
      <aside class="issue-detail__sidebar">
        <h3>活动日志</h3>
        <div v-if="activities.length === 0" class="text-muted">暂无活动记录</div>
        <div v-else class="activity-timeline">
          <div v-for="act in activities" :key="act.id" class="activity-item">
            <div class="activity-item__icon" :class="`verb-${act.verb}`"></div>
            <div class="activity-item__body">
              <div class="activity-item__text">
                <strong>{{ act.actor_name || "系统" }}</strong>
                {{ act.verb === "created" ? "创建了工作项" : act.verb === "transitioned" ? `流转状态: ${act.old_value} → ${act.new_value}` : `${act.field}: ${act.old_value} → ${act.new_value}` }}
              </div>
              <div class="activity-item__time">{{ new Date(act.created_at).toLocaleString() }}</div>
            </div>
          </div>
        </div>

        <!-- 工时 -->
        <h3 style="margin-top: 24px">工时</h3>
        <div class="timelog-summary" v-if="totalMinutes > 0">
          累计 {{ fmtDuration(totalMinutes) }}
          <span v-if="issue.actual_effort != null" class="timelog-effort">
            · 实耗 {{ fmtDuration(issue.actual_effort) }}
          </span>
          <span v-if="issue.remaining_effort != null" class="timelog-effort">
            · 剩余 {{ fmtDuration(issue.remaining_effort) }}
          </span>
        </div>

        <button v-if="!showTimeLogForm" class="btn btn--sm btn--outline" @click="showTimeLogForm = true">
          ＋ 记录工时
        </button>

        <!-- 工时记录表单 -->
        <div v-if="showTimeLogForm" class="timelog-form">
          <div v-if="timeLogError" class="form-error">{{ timeLogError }}</div>
          <div class="timelog-form__row">
            <input
              v-model="newSpentDate"
              type="date"
              class="timelog-input"
              :max="new Date().toISOString().slice(0, 10)"
            />
          </div>
          <div class="timelog-form__row timelog-form__duration">
            <input v-model.number="newDurationHours" type="number" class="timelog-input timelog-input--sm" min="0" max="24" />
            <span class="timelog-label">小时</span>
            <input v-model.number="newDurationMinutes" type="number" class="timelog-input timelog-input--sm" min="0" max="59" step="15" />
            <span class="timelog-label">分钟</span>
          </div>
          <textarea
            v-model="newTimeDesc"
            class="timelog-textarea"
            placeholder="工时描述（可选）"
            rows="2"
          ></textarea>
          <div class="timelog-form__actions">
            <button class="btn btn--sm" @click="showTimeLogForm = false" :disabled="timeLogSubmitting">取消</button>
            <button class="btn btn--sm btn--primary" @click="submitTimeLog" :disabled="timeLogSubmitting">
              {{ timeLogSubmitting ? "保存中..." : "保存" }}
            </button>
          </div>
        </div>

        <!-- 工时列表 -->
        <div v-if="timeLogsLoading" class="text-muted" style="margin-top:8px">加载中...</div>
        <div v-else-if="timeLogs.length > 0" class="timelog-list">
          <div v-for="tl in timeLogs.slice(0, 10)" :key="tl.id" class="timelog-item">
            <div class="timelog-item__meta">
              <span class="timelog-item__date">{{ tl.spent_date.slice(0, 10) }}</span>
              <span class="timelog-item__duration">{{ fmtDuration(tl.duration_minutes) }}</span>
            </div>
            <div v-if="tl.description" class="timelog-item__desc">{{ tl.description }}</div>
          </div>
        </div>
        <div v-else-if="!showTimeLogForm" class="text-muted">暂无工时记录</div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.issue-detail__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.issue-detail__body {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 32px;
}

.issue-detail__identifier {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
  font-weight: 600;
}

.issue-detail__name {
  font-size: 22px;
  font-weight: 600;
  margin: 4px 0 12px;
  color: var(--text-primary);
}

.issue-detail__meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-weight: 500;
}
.badge-requirement { background: var(--brand-50); color: var(--brand-600); }
.badge-task { background: var(--success-50); color: var(--success-600); }
.badge-defect { background: var(--danger-50); color: var(--danger-600); }

.issue-detail__state-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  color: #fff;
  font-weight: 500;
}

.issue-detail__priority, .issue-detail__field {
  font-size: 12px;
  color: var(--text-secondary);
}

.issue-detail__section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--border-subtle);
}

.issue-detail__section h3 {
  font-size: 14px;
  font-weight: 600;
  margin: 0 0 12px;
  color: var(--text-primary);
}

.issue-detail__desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.issue-detail__fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-row {
  display: flex;
  gap: 12px;
  font-size: 13px;
}

.field-label {
  color: var(--text-tertiary);
  min-width: 80px;
}

.field-value {
  color: var(--text-secondary);
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 13px;
}

.issue-detail__transitions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.form-error { color: var(--danger-500); font-size: 12px; margin-bottom: 8px; }

.btn {
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--brand-500);
  padding: 4px 0;
}

.btn--danger {
  background: var(--danger-50);
  border-color: var(--danger-200);
  color: var(--danger-600);
}

.btn--state {
  background: var(--surface-1);
  font-weight: 500;
}

.activity-timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.activity-item {
  display: flex;
  gap: 10px;
}

.activity-item__icon {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--brand-300);
  margin-top: 6px;
  flex-shrink: 0;
}
.activity-item__icon.verb-created { background: var(--success-500); }
.activity-item__icon.verb-transitioned { background: var(--brand-500); }

.activity-item__text {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.activity-item__time {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.loading, .error {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}
.error { color: var(--danger-500); }

/* ===== Time Log ===== */
.timelog-summary {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 500;
}

.timelog-effort {
  color: var(--text-tertiary);
  font-weight: 400;
}

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
}

.btn--sm:hover:not(:disabled) {
  background: var(--surface-3);
}

.btn--outline {
  border: 1px dashed var(--border-strong);
  background: none;
  color: var(--text-tertiary);
  width: 100%;
  text-align: center;
}

.btn--outline:hover {
  border-color: var(--brand-500);
  color: var(--brand-500);
}

.btn--primary {
  background: var(--brand-500);
  color: #fff;
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn:disabled,
.btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.timelog-form {
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  margin-top: 8px;
}

.timelog-form__row {
  margin-bottom: 8px;
}

.timelog-form__duration {
  display: flex;
  align-items: center;
  gap: 4px;
}

.timelog-label {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 4px;
}

.timelog-input {
  padding: 5px 8px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  width: 100%;
}

.timelog-input:focus {
  border-color: var(--brand-500);
}

.timelog-input--sm {
  width: 72px;
}

.timelog-textarea {
  width: 100%;
  padding: 6px 8px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  resize: vertical;
  margin-bottom: 8px;
}

.timelog-textarea:focus {
  border-color: var(--brand-500);
}

.timelog-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

.timelog-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.timelog-item {
  padding: 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.timelog-item__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.timelog-item__date {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.timelog-item__duration {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.timelog-item__desc {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}
</style>
