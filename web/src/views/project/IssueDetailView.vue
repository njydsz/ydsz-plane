<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { issueApi, type Issue, type IssueActivity, type State } from "@/api/services/issue";
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

onMounted(load);
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
        <div class="text-muted">工时记录功能即将上线</div>
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
</style>
