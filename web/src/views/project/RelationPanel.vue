<script setup lang="ts">
import { onMounted, ref } from "vue";

import { issueApi, type IssueRelation } from "@/api/services/issue";

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

async function load() {
  loading.value = true;
  try {
    const res = await issueApi.listRelations(props.workspaceId, props.projectId, props.issueId);
    relations.value = res.results;
  } catch {
    // 非关键模块
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

onMounted(load);
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
        <button class="btn btn--sm" @click="showForm = false" :disabled="submitting">取消</button>
        <button class="btn btn--sm btn--primary" @click="createRel" :disabled="!targetIssueId || submitting">
          {{ submitting ? "添加中..." : "确认" }}
        </button>
      </div>
    </div>

    <!-- 关联列表 -->
    <div v-if="loading" class="text-muted" style="margin-top: 8px">加载中...</div>
    <div v-else-if="relations.length > 0" class="relation-list">
      <div v-for="rel in relations" :key="rel.id" class="relation-item">
        <div class="relation-item__info">
          <span class="relation-item__type">{{ relationLabels[rel.relation_type] ?? rel.relation_type }}</span>
          <span class="relation-item__target">
            #{{ rel.source_issue_id === props.issueId ? rel.target_issue_id : rel.source_issue_id }}
          </span>
        </div>
        <button class="relation-item__remove" @click="removeRel(rel.id)" title="移除关联">✕</button>
      </div>
    </div>
    <div v-else-if="!showForm" class="text-muted">暂无关联关系</div>
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
</style>
