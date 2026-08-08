<script setup lang="ts">
/**
 * RelationPanel — 工作项关联关系面板。
 *
 * 展示当前工作项的关联/被关联列表，支持新增关联与删除。
 */
import { onMounted, ref } from "vue";

import { issueApi, type IssueRelation, type IssueDependency, type DependencyType, DEPENDENCY_TYPE_LABELS } from "@/api/services/issue";
import { AppLoadingState, AppErrorState } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
}>();

const relations = ref<IssueRelation[]>([]);
const loading = ref(false);
const showForm = ref(false);
const targetIssueId = ref<number | null>(null);
const relationType = ref("relates_to");
const loadError = ref("");
const error = ref("");
const submitting = ref(false);

const relationLabels: Record<string, string> = {
  duplicate: "重复",
  relates_to: "关联到",
  blocked_by: "被阻塞",
  start_before: "开始前于",
  finish_before: "完成前于",
  implemented_by: "被实现",
};

const typeOptions = Object.entries(relationLabels).map(([k, v]) => ({ value: k, label: v }));

// ---- 任务依赖 ----
const dependencies = ref<{ predecessors: IssueDependency[]; successors: IssueDependency[] }>({ predecessors: [], successors: [] });
const depsLoading = ref(false);
const showDepForm = ref(false);
const depPredecessorId = ref<number | null>(null);
const depSuccessorId = ref<number | null>(null);
const depType = ref<DependencyType>("FS");
const depLagDays = ref(0);
const depError = ref("");
const depSubmitting = ref(false);

const dependencyTypeLabels = DEPENDENCY_TYPE_LABELS;

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const res = await issueApi.listRelations(props.workspaceId, props.projectId, props.issueId);
    relations.value = res.results;
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : "加载关联关系失败";
  } finally {
    loading.value = false;
  }
}

async function createRel() {
  if (!targetIssueId.value || submitting.value) return;
  submitting.value = true;
  error.value = "";
  try {
    await issueApi.createRelation(props.workspaceId, props.projectId, props.issueId, {
      target_issue_id: targetIssueId.value,
      relation_type: relationType.value,
    });
    showForm.value = false;
    targetIssueId.value = null;
    await load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "添加关联失败";
  } finally {
    submitting.value = false;
  }
}

async function removeRel(relationId: number) {
  try {
    await issueApi.deleteRelation(props.workspaceId, props.projectId, props.issueId, relationId);
    await load();
  } catch {
    // 静默失败
  }
}

async function loadDeps() {
  depsLoading.value = true;
  try {
    dependencies.value = await issueApi.listDependencies(props.workspaceId, props.projectId, props.issueId);
  } catch {
    // 非关键模块
  } finally {
    depsLoading.value = false;
  }
}

async function createDep() {
  if (!depPredecessorId.value || !depSuccessorId.value || depSubmitting.value) return;
  depSubmitting.value = true;
  depError.value = "";
  try {
    await issueApi.createDependency(props.workspaceId, props.projectId, props.issueId, {
      predecessor_id: depPredecessorId.value,
      successor_id: depSuccessorId.value,
      dependency_type: depType.value,
      lag_days: depLagDays.value,
    });
    showDepForm.value = false;
    depPredecessorId.value = null;
    depSuccessorId.value = null;
    depLagDays.value = 0;
    await loadDeps();
  } catch (e: unknown) {
    depError.value = e instanceof Error ? e.message : "添加依赖失败";
  } finally {
    depSubmitting.value = false;
  }
}

async function removeDep(depId: number) {
  try {
    await issueApi.deleteDependency(props.workspaceId, props.projectId, props.issueId, depId);
    await loadDeps();
  } catch {
    // 静默失败
  }
}

onMounted(() => {
  load();
  loadDeps();
});
</script>

<template>
  <div class="relation-panel">
    <h3>关联关系</h3>

    <button v-if="!showForm" class="btn btn--outline" @click="showForm = true">
      ＋ 添加关联
    </button>

    <!-- 添加关联表单 -->
    <div v-if="showForm" class="relation-form">
      <div v-if="error" class="form-error">{{ error }}</div>
      <div class="relation-form__row">
        <input
          v-model.number="targetIssueId"
          type="number"
          class="relation-input"
          placeholder="目标工作项 ID"
          min="1"
        />
        <select v-model="relationType" class="relation-select">
          <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>
      <div class="relation-form__actions">
        <button class="btn btn--sm" :disabled="submitting" @click="showForm = false">取消</button>
        <button class="btn btn--sm btn--primary" :disabled="!targetIssueId || submitting" @click="createRel">
          {{ submitting ? "添加中..." : "确认" }}
        </button>
      </div>
    </div>

    <!-- 关联列表 -->
    <AppLoadingState v-if="loading" text="加载关联关系..." />
    <AppErrorState
      v-else-if="loadError"
      :message="loadError"
      @retry="load"
    />
    <div v-else-if="relations.length > 0" class="relation-list">
      <div v-for="rel in relations" :key="rel.id" class="relation-item">
        <div class="relation-item__info">
          <span class="relation-item__type">{{ relationLabels[rel.relation_type] ?? rel.relation_type }}</span>
          <span class="relation-item__target">
            #{{ rel.source_issue_id === props.issueId ? rel.target_issue_id : rel.source_issue_id }}
          </span>
        </div>
        <button class="relation-item__remove" title="移除关联" @click="removeRel(rel.id)">✕</button>
      </div>
    </div>
    <div v-else-if="!showForm" class="text-muted">暂无关联关系</div>

    <!-- 任务依赖 -->
    <h3 style="margin-top: 24px">任务依赖</h3>
    <button v-if="!showDepForm" class="btn btn--outline" @click="showDepForm = true">
      ＋ 添加依赖
    </button>

    <!-- 添加依赖表单 -->
    <div v-if="showDepForm" class="relation-form">
      <div v-if="depError" class="form-error">{{ depError }}</div>
      <div class="relation-form__row">
        <input
          v-model.number="depPredecessorId"
          type="number"
          class="relation-input"
          placeholder="前置任务 ID"
          min="1"
        />
        <select v-model="depType" class="relation-select">
          <option v-for="(label, key) in dependencyTypeLabels" :key="key" :value="key">
            {{ label }}
          </option>
        </select>
      </div>
      <div class="relation-form__row">
        <input
          v-model.number="depSuccessorId"
          type="number"
          class="relation-input"
          placeholder="后置任务 ID"
          min="1"
        />
        <input
          v-model.number="depLagDays"
          type="number"
          class="relation-input"
          placeholder="滞后天数"
          min="0"
          style="max-width: 100px"
        />
      </div>
      <div class="relation-form__actions">
        <button class="btn btn--sm" :disabled="depSubmitting" @click="showDepForm = false">取消</button>
        <button class="btn btn--sm btn--primary" :disabled="!depPredecessorId || !depSuccessorId || depSubmitting" @click="createDep">
          {{ depSubmitting ? "添加中..." : "确认" }}
        </button>
      </div>
    </div>

    <!-- 依赖列表 -->
    <AppLoadingState v-if="depsLoading" text="加载依赖关系..." />
    <div v-else-if="dependencies.predecessors.length > 0 || dependencies.successors.length > 0" class="relation-list">
      <template v-if="dependencies.predecessors.length > 0">
        <div class="dep-section-label">前置任务</div>
        <div v-for="dep in dependencies.predecessors" :key="'pre-' + dep.id" class="relation-item">
          <div class="relation-item__info">
            <span class="dep-type-badge" :title="dependencyTypeLabels[dep.dependency_type] ?? dep.dependency_type">{{ dep.dependency_type }}</span>
            <span class="relation-item__target">#{{ dep.predecessor_id }}</span>
            <span v-if="dep.lag_days > 0" class="dep-lag">+{{ dep.lag_days }}天</span>
          </div>
          <button class="relation-item__remove" title="移除依赖" @click="removeDep(dep.id)">✕</button>
        </div>
      </template>
      <template v-if="dependencies.successors.length > 0">
        <div class="dep-section-label">后置任务</div>
        <div v-for="dep in dependencies.successors" :key="'suc-' + dep.id" class="relation-item">
          <div class="relation-item__info">
            <span class="dep-type-badge" :title="dependencyTypeLabels[dep.dependency_type] ?? dep.dependency_type">{{ dep.dependency_type }}</span>
            <span class="relation-item__target">#{{ dep.successor_id }}</span>
            <span v-if="dep.lag_days > 0" class="dep-lag">+{{ dep.lag_days }}天</span>
          </div>
          <button class="relation-item__remove" title="移除依赖" @click="removeDep(dep.id)">✕</button>
        </div>
      </template>
    </div>
    <div v-else-if="!showDepForm" class="text-muted">暂无任务依赖</div>
  </div>
</template>

<style scoped>
h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px;
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 13px;
  margin-top: 8px;
}

.form-error {
  color: var(--danger-500);
  font-size: 12px;
  margin-bottom: 8px;
}

/* Button */
.btn--outline {
  border: 1px dashed var(--border-strong);
  background: none;
  color: var(--text-tertiary);
  font-size: 12px;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-family: inherit;
  width: 100%;
  text-align: center;
}

.btn--outline:hover {
  border-color: var(--brand-500);
  color: var(--brand-500);
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

.btn--primary {
  background: var(--brand-500);
  color: #fff;
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Form */
.relation-form {
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  margin-top: 8px;
}

.relation-form__row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.relation-input,
.relation-select {
  padding: 5px 8px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
}

.relation-input { flex: 1; }
.relation-select { flex: 1; cursor: pointer; }

.relation-input:focus,
.relation-select:focus {
  border-color: var(--brand-500);
}

.relation-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

/* List */
.relation-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.relation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.relation-item__info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.relation-item__type {
  color: var(--text-secondary);
  font-weight: 500;
}

.relation-item__target {
  font-family: var(--font-mono);
  color: var(--brand-500);
  font-weight: 500;
}

.relation-item__remove {
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  font-size: 12px;
  font-family: inherit;
}

.relation-item__remove:hover {
  color: var(--danger-500);
}

.dep-type-badge {
  display: inline-block;
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--brand-600);
  background: var(--brand-50);
  border: 1px solid var(--brand-200);
  border-radius: 3px;
  cursor: help;
  letter-spacing: 0.5px;
}

.dep-section-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  padding: 4px 0 2px;
}

.dep-lag {
  font-size: 11px;
  color: var(--warning-500);
  margin-left: 4px;
}
</style>
