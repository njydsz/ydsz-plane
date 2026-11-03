<script setup lang="ts">
/**
 * WorkflowConfigView — 项目工作流与审批配置视图。
 *
 * 能力：
 *  - 左侧工作流列表（点击切换选中）
 *  - 工作流 CRUD（新建 / 编辑 / 删除，支持设为默认）
 *  - 流转规则管理（添加 / 删除 / 校验）
 *  - 流转规则可视化（from → to 流程条）
 *  - 审批配置（requires_approval / approver_role / approver_ids）
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import {
  workflowApi,
  type Workflow,
  type WorkflowTransition,
  type AddTransitionInput,
} from "@/api/services/workflow";
import { stateMachineApi, type State } from "@/api/services/stateMachine";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { useWorkspaceStore } from "@/stores/workspace";
import { useTranslation } from "@/composables/useTranslation";
import {
  AppLoadingState,
  AppErrorState,
  AppEmptyState,
  AppModal,
  AppBadge,
} from "@/components";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* i18n                                                               */
/* ------------------------------------------------------------------ */

const { t } = useTranslation();

/* ------------------------------------------------------------------ */
/* Workspace context                                                  */
/* ------------------------------------------------------------------ */

const route = useRoute();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ------------------------------------------------------------------ */
/* Data state                                                         */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const workflows = ref<Workflow[]>([]);
const selectedId = ref<number | null>(null);
const states = ref<State[]>([]);
const members = ref<Member[]>([]);

/** 当前选中的工作流 */
const selectedWorkflow = computed<Workflow | null>(
  () => workflows.value.find((w) => w.id === selectedId.value) ?? null,
);

/** 当前工作流的流转规则 */
const selectedTransitions = computed<WorkflowTransition[]>(
  () => selectedWorkflow.value?.transitions ?? [],
);

/* ------------------------------------------------------------------ */
/* Workflow form                                                      */
/* ------------------------------------------------------------------ */

const showWorkflowForm = ref(false);
const workflowFormSaving = ref(false);
const editingWorkflow = ref<Workflow | null>(null);
const workflowFormError = ref("");

const workflowForm = ref<{
  name: string;
  description: string;
  is_default: boolean;
}>({
  name: "",
  description: "",
  is_default: false,
});

function resetWorkflowForm() {
  workflowForm.value = { name: "", description: "", is_default: false };
  editingWorkflow.value = null;
  workflowFormError.value = "";
}

function openCreateWorkflow() {
  resetWorkflowForm();
  showWorkflowForm.value = true;
}

function openEditWorkflow(wf: Workflow) {
  editingWorkflow.value = wf;
  workflowForm.value = {
    name: wf.name,
    description: wf.description ?? "",
    is_default: wf.is_default,
  };
  workflowFormError.value = "";
  showWorkflowForm.value = true;
}

async function saveWorkflow() {
  if (!wsId.value) return;
  if (!workflowForm.value.name.trim()) {
    workflowFormError.value = t("workflow.form.nameRequired");
    return;
  }
  workflowFormSaving.value = true;
  workflowFormError.value = "";
  try {
    const payload = {
      name: workflowForm.value.name.trim(),
      description: workflowForm.value.description.trim() || undefined,
      is_default: workflowForm.value.is_default,
    };
    if (editingWorkflow.value) {
      await workflowApi.update(wsId.value, projectId.value, editingWorkflow.value.id, payload);
      toast.success(t("workflow.workflow.updated"));
    } else {
      await workflowApi.create(wsId.value, projectId.value, payload);
      toast.success(t("workflow.workflow.created"));
    }
    showWorkflowForm.value = false;
    resetWorkflowForm();
    await loadWorkflows();
  } catch (e: unknown) {
    workflowFormError.value = e instanceof Error ? e.message : t("common.saveFailed");
  } finally {
    workflowFormSaving.value = false;
  }
}

async function deleteWorkflow(wf: Workflow) {
  if (!wsId.value) return;
  if (!confirm(t("workflow.workflow.deleteConfirm", { name: wf.name }))) return;
  try {
    await workflowApi.remove(wsId.value, projectId.value, wf.id);
    toast.success(t("workflow.workflow.deleted"));
    if (selectedId.value === wf.id) selectedId.value = null;
    await loadWorkflows();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("common.deleteFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* Transition form                                                    */
/* ------------------------------------------------------------------ */

const showTransitionForm = ref(false);
const transitionFormSaving = ref(false);
const transitionFormError = ref("");

const transitionForm = ref<AddTransitionInput>({
  from_state_id: 0,
  to_state_id: 0,
  name: "",
  requires_approval: false,
  approver_role: "",
  approver_ids: [],
});

function resetTransitionForm() {
  transitionForm.value = {
    from_state_id: 0,
    to_state_id: 0,
    name: "",
    requires_approval: false,
    approver_role: "",
    approver_ids: [],
  };
  transitionFormError.value = "";
}

function openAddTransition() {
  resetTransitionForm();
  if (states.value.length >= 2) {
    transitionForm.value.from_state_id = states.value[0].id;
    transitionForm.value.to_state_id = states.value[1].id;
  }
  showTransitionForm.value = true;
}

async function saveTransition() {
  if (!wsId.value || !selectedWorkflow.value) return;
  if (transitionForm.value.from_state_id === transitionForm.value.to_state_id) {
    transitionFormError.value = t("workflow.transition.sameStateError");
    return;
  }
  if (!transitionForm.value.name.trim()) {
    transitionFormError.value = t("workflow.form.nameRequired");
    return;
  }
  transitionFormSaving.value = true;
  transitionFormError.value = "";
  try {
    await workflowApi.addTransition(
      wsId.value,
      projectId.value,
      selectedWorkflow.value.id,
      {
        ...transitionForm.value,
        name: transitionForm.value.name.trim(),
        approver_role: transitionForm.value.approver_role || undefined,
        approver_ids: transitionForm.value.requires_approval
          ? transitionForm.value.approver_ids
          : [],
      },
    );
    toast.success(t("workflow.transition.added"));
    showTransitionForm.value = false;
    resetTransitionForm();
    await loadWorkflows();
  } catch (e: unknown) {
    transitionFormError.value = e instanceof Error ? e.message : t("common.saveFailed");
  } finally {
    transitionFormSaving.value = false;
  }
}

async function removeTransition(tr: WorkflowTransition) {
  if (!wsId.value || !selectedWorkflow.value) return;
  if (!confirm(t("workflow.transition.deleteConfirm"))) return;
  try {
    await workflowApi.removeTransition(
      wsId.value,
      projectId.value,
      selectedWorkflow.value.id,
      tr.id,
    );
    toast.success(t("workflow.transition.deleted"));
    await loadWorkflows();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("common.deleteFailed"));
  }
}

async function validateTransition(tr: WorkflowTransition, workItemId: number | string) {
  if (!wsId.value || !selectedWorkflow.value) return;
  try {
    const res = await workflowApi.validateTransition(
      wsId.value,
      projectId.value,
      selectedWorkflow.value.id,
      tr.id,
      workItemId,
    );
    if (res.valid) {
      toast.success(t("workflow.transition.validateOk"));
    } else {
      toast.error((res.errors ?? []).join("；") || t("workflow.transition.validateFail"));
    }
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workflow.transition.validateFail"));
  }
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

function stateName(id: number): string {
  return states.value.find((s) => s.id === id)?.name ?? `#${id}`;
}

function stateColor(id: number): string {
  return states.value.find((s) => s.id === id)?.color ?? "#9CA3AF";
}

function memberName(id: number): string {
  const m = members.value.find((m) => m.id === id);
  return m?.display_name || m?.email || `用户${id}`;
}

const APPROVER_ROLE_OPTIONS = [
  { value: "", label: "— 选择角色 —" },
  { value: "owner", label: "空间管理员" },
  { value: "admin", label: "系统管理员" },
  { value: "pm", label: "项目经理" },
  { value: "po", label: "产品经理" },
  { value: "techlead", label: "技术经理" },
  { value: "qalead", label: "测试经理" },
  { value: "dev", label: "开发" },
];

/* ------------------------------------------------------------------ */
/* Load                                                               */
/* ------------------------------------------------------------------ */

async function loadWorkflows() {
  if (!wsId.value) return;
  const data = await workflowApi.list(wsId.value, projectId.value);
  workflows.value = data;
  if (!selectedId.value && data.length > 0) {
    selectedId.value = data[0].id;
  }
}

async function loadStates() {
  if (!wsId.value) return;
  states.value = await stateMachineApi.getStates(wsId.value, projectId.value);
}

async function loadMembers() {
  if (!wsId.value) return;
  members.value = await workspaceApi.listMembers(wsId.value);
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    await Promise.all([loadWorkflows(), loadStates(), loadMembers()]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t("common.loadFailed");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (wsId.value) void load();
});

watch(wsId, (id) => {
  if (id) void load();
});
</script>

<template>
  <div class="wf-view">
    <!-- 页头 -->
    <header class="wf-view__header">
      <div>
        <h1 class="wf-view__title">{{ t("workflow.title") }}</h1>
        <p class="wf-view__subtitle">{{ t("workflow.subtitle") }}</p>
      </div>
      <button class="wf-btn wf-btn--primary" @click="openCreateWorkflow">
        {{ t("workflow.workflow.add") }}
      </button>
    </header>

    <!-- 加载 / 错误 -->
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <div class="wf-layout">
        <!-- 左侧：工作流列表 -->
        <aside class="wf-sidebar">
          <div class="wf-sidebar__title">{{ t("workflow.workflow.list") }}</div>
          <AppEmptyState
            v-if="!workflows.length"
            :title="t('workflow.workflow.empty')"
            :description="t('workflow.workflow.emptyDesc')"
            illustration-size="sm"
            compact
          />
          <button
            v-for="wf in workflows"
            :key="wf.id"
            class="wf-sidebar__item"
            :class="{ 'wf-sidebar__item--active': wf.id === selectedId }"
            @click="selectedId = wf.id"
          >
            <span class="wf-sidebar__name">{{ wf.name }}</span>
            <AppBadge v-if="wf.is_default" variant="brand">{{ t("workflow.workflow.default") }}</AppBadge>
            <span class="wf-sidebar__count">{{ wf.transitions?.length ?? 0 }}</span>
          </button>
        </aside>

        <!-- 右侧：选中工作流详情 -->
        <main class="wf-main">
          <AppEmptyState
            v-if="!selectedWorkflow"
            :title="t('workflow.transition.empty')"
            :description="t('workflow.workflow.selectPlaceholder')"
          />

          <template v-else>
            <!-- 工作流元信息 -->
            <div class="wf-main__head">
              <div class="wf-main__meta">
                <h2 class="wf-main__name">{{ selectedWorkflow.name }}</h2>
                <p v-if="selectedWorkflow.description" class="wf-main__desc">
                  {{ selectedWorkflow.description }}
                </p>
                <div class="wf-main__tags">
                  <AppBadge v-if="selectedWorkflow.is_default" variant="brand">
                    {{ t("workflow.workflow.default") }}
                  </AppBadge>
                  <span class="wf-main__tag">
                    {{ t("workflow.transition.count", { count: selectedTransitions.length }) }}
                  </span>
                </div>
              </div>
              <div class="wf-main__actions">
                <button class="wf-btn wf-btn--ghost" @click="openEditWorkflow(selectedWorkflow)">
                  {{ t("common.edit") }}
                </button>
                <button class="wf-btn wf-btn--danger" @click="deleteWorkflow(selectedWorkflow)">
                  {{ t("common.delete") }}
                </button>
              </div>
            </div>

            <!-- 流转规则 -->
            <div class="wf-section">
              <div class="wf-section__head">
                <h3 class="wf-section__title">{{ t("workflow.transition.list") }}</h3>
                <button class="wf-btn wf-btn--primary wf-btn--sm" @click="openAddTransition">
                  {{ t("workflow.transition.add") }}
                </button>
              </div>

              <AppEmptyState
                v-if="!selectedTransitions.length"
                :title="t('workflow.transition.empty')"
                :description="t('workflow.transition.emptyDesc')"
                illustration-size="sm"
                compact
              />

              <!-- 可视化流转流 -->
              <div v-else class="wf-flow">
                <div
                  v-for="tr in selectedTransitions"
                  :key="tr.id"
                  class="wf-flow__row"
                >
                  <span
                    class="wf-flow__node"
                    :style="{ borderColor: stateColor(tr.from_state_id) }"
                  >
                    {{ stateName(tr.from_state_id) }}
                  </span>
                  <span class="wf-flow__arrow">
                    <span class="wf-flow__arrow-line" />
                    <span class="wf-flow__arrow-head" />
                  </span>
                  <span class="wf-flow__label">
                    {{ tr.name }}
                    <span v-if="tr.requires_approval" class="wf-flow__approval">
                      {{ t("workflow.transition.requiresApproval") }}
                    </span>
                  </span>
                  <span class="wf-flow__arrow">
                    <span class="wf-flow__arrow-line" />
                    <span class="wf-flow__arrow-head" />
                  </span>
                  <span
                    class="wf-flow__node"
                    :style="{ borderColor: stateColor(tr.to_state_id) }"
                  >
                    {{ stateName(tr.to_state_id) }}
                  </span>
                </div>
              </div>

              <!-- 流转规则表格 -->
              <div v-if="selectedTransitions.length" class="wf-table-wrap">
                <table class="wf-table">
                  <thead>
                    <tr>
                      <th>{{ t("workflow.transition.from") }}</th>
                      <th>{{ t("workflow.transition.to") }}</th>
                      <th>{{ t("workflow.transition.name") }}</th>
                      <th>{{ t("workflow.transition.approval") }}</th>
                      <th>{{ t("workflow.transition.approver") }}</th>
                      <th>{{ t("common.actions") }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="tr in selectedTransitions" :key="tr.id">
                      <td>
                        <span
                          class="wf-dot"
                          :style="{ backgroundColor: stateColor(tr.from_state_id) }"
                        />
                        {{ stateName(tr.from_state_id) }}
                      </td>
                      <td>
                        <span
                          class="wf-dot"
                          :style="{ backgroundColor: stateColor(tr.to_state_id) }"
                        />
                        {{ stateName(tr.to_state_id) }}
                      </td>
                      <td>{{ tr.name }}</td>
                      <td>
                        <AppBadge v-if="tr.requires_approval" variant="warning">
                          {{ t("workflow.transition.requiresApproval") }}
                        </AppBadge>
                        <span v-else class="wf-muted">—</span>
                      </td>
                      <td class="wf-cell--approver">
                        <span v-if="tr.approver_role" class="wf-role">
                          {{ APPROVER_ROLE_OPTIONS.find((o) => o.value === tr.approver_role)?.label || tr.approver_role }}
                        </span>
                        <span v-if="tr.approver_ids?.length" class="wf-approvers">
                          {{ tr.approver_ids.map(memberName).join("、") }}
                        </span>
                        <span v-if="!tr.approver_role && !tr.approver_ids?.length" class="wf-muted">
                          —
                        </span>
                      </td>
                      <td class="wf-cell--actions">
                        <button
                          class="wf-btn wf-btn--ghost wf-btn--xs"
                          @click="validateTransition(tr, 0)"
                        >
                          {{ t("workflow.transition.validate") }}
                        </button>
                        <button
                          class="wf-btn wf-btn--danger wf-btn--xs"
                          @click="removeTransition(tr)"
                        >
                          {{ t("common.delete") }}
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </template>
        </main>
      </div>
    </template>

    <!-- 工作流表单弹窗 -->
    <AppModal
      :visible="showWorkflowForm"
      :title="editingWorkflow ? t('workflow.workflow.edit') : t('workflow.workflow.add')"
      width="480px"
      @close="showWorkflowForm = false"
    >
      <form class="wf-form" @submit.prevent="saveWorkflow">
        <label class="wf-form__field">
          <span class="wf-form__label">{{ t("workflow.workflow.name") }} *</span>
          <input
            v-model="workflowForm.name"
            class="wf-input"
            :placeholder="t('workflow.workflow.namePlaceholder')"
            required
          />
        </label>
        <label class="wf-form__field">
          <span class="wf-form__label">{{ t("workflow.workflow.description") }}</span>
          <textarea
            v-model="workflowForm.description"
            class="wf-input wf-textarea"
            :placeholder="t('workflow.workflow.descriptionPlaceholder')"
            rows="3"
          />
        </label>
        <label class="wf-form__field wf-form__field--checkbox">
          <input v-model="workflowForm.is_default" type="checkbox" />
          <span>{{ t("workflow.workflow.setAsDefault") }}</span>
        </label>
        <p v-if="workflowFormError" class="wf-form__error">{{ workflowFormError }}</p>
        <div class="wf-form__actions">
          <button
            type="button"
            class="wf-btn wf-btn--ghost"
            @click="showWorkflowForm = false"
          >
            {{ t("common.cancel") }}
          </button>
          <button type="submit" class="wf-btn wf-btn--primary" :disabled="workflowFormSaving">
            {{ workflowFormSaving ? t("common.saving") : t("common.save") }}
          </button>
        </div>
      </form>
    </AppModal>

    <!-- 流转规则表单弹窗 -->
    <AppModal
      :visible="showTransitionForm"
      :title="t('workflow.transition.add')"
      width="520px"
      @close="showTransitionForm = false"
    >
      <form class="wf-form" @submit.prevent="saveTransition">
        <div class="wf-form__row">
          <label class="wf-form__field">
            <span class="wf-form__label">{{ t("workflow.transition.from") }} *</span>
            <select v-model.number="transitionForm.from_state_id" class="wf-select" required>
              <option value="0" disabled>— {{ t("common.select") }} —</option>
              <option v-for="st in states" :key="st.id" :value="st.id">{{ st.name }}</option>
            </select>
          </label>
          <label class="wf-form__field">
            <span class="wf-form__label">{{ t("workflow.transition.to") }} *</span>
            <select v-model.number="transitionForm.to_state_id" class="wf-select" required>
              <option value="0" disabled>— {{ t("common.select") }} —</option>
              <option v-for="st in states" :key="st.id" :value="st.id">{{ st.name }}</option>
            </select>
          </label>
        </div>
        <label class="wf-form__field">
          <span class="wf-form__label">{{ t("workflow.transition.name") }} *</span>
          <input
            v-model="transitionForm.name"
            class="wf-input"
            :placeholder="t('workflow.transition.namePlaceholder')"
            required
          />
        </label>
        <label class="wf-form__field wf-form__field--checkbox">
          <input v-model="transitionForm.requires_approval" type="checkbox" />
          <span>{{ t("workflow.transition.requiresApproval") }}</span>
        </label>
        <template v-if="transitionForm.requires_approval">
          <label class="wf-form__field">
            <span class="wf-form__label">{{ t("workflow.transition.approverRole") }}</span>
            <select v-model="transitionForm.approver_role" class="wf-select">
              <option v-for="opt in APPROVER_ROLE_OPTIONS" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </label>
          <label class="wf-form__field">
            <span class="wf-form__label">{{ t("workflow.transition.approverIds") }}</span>
            <select v-model="transitionForm.approver_ids" class="wf-select" multiple>
              <option v-for="m in members" :key="m.id" :value="m.id">
                {{ m.display_name || m.email }}
              </option>
            </select>
          </label>
        </template>
        <p v-if="transitionFormError" class="wf-form__error">{{ transitionFormError }}</p>
        <div class="wf-form__actions">
          <button
            type="button"
            class="wf-btn wf-btn--ghost"
            @click="showTransitionForm = false"
          >
            {{ t("common.cancel") }}
          </button>
          <button type="submit" class="wf-btn wf-btn--primary" :disabled="transitionFormSaving">
            {{ transitionFormSaving ? t("common.saving") : t("common.add") }}
          </button>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<style scoped>
.wf-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.wf-view__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.wf-view__title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.wf-view__subtitle {
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

/* Layout */
.wf-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 20px;
  align-items: start;
}

/* Sidebar */
.wf-sidebar {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  position: sticky;
  top: 16px;
}

.wf-sidebar__title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 6px 10px 8px;
}

.wf-sidebar__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: none;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  color: var(--text-secondary);
  text-align: left;
  transition: all 0.15s;
}

.wf-sidebar__item:hover {
  background: var(--surface-2);
  color: var(--text-primary);
}

.wf-sidebar__item--active {
  background: var(--brand-50);
  border-color: var(--brand-100);
  color: var(--brand-700);
  font-weight: 500;
}

.wf-sidebar__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-sidebar__count {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 99px;
  background: var(--surface-3);
  color: var(--text-tertiary);
}

/* Main */
.wf-main {
  min-width: 0;
}

.wf-main__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.wf-main__name {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.wf-main__desc {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 4px 0 8px;
}

.wf-main__tags {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wf-main__tag {
  font-size: 11px;
  color: var(--text-tertiary);
}

.wf-main__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

/* Section */
.wf-section {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 16px;
}

.wf-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.wf-section__title {
  font-size: 15px;
  font-weight: 600;
  margin: 0;
}

/* Flow visualization */
.wf-flow {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.wf-flow__row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.wf-flow__node {
  padding: 4px 12px;
  border: 1px solid;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 500;
  background: var(--surface-2);
  white-space: nowrap;
}

.wf-flow__arrow {
  display: inline-flex;
  align-items: center;
  color: var(--text-tertiary);
}

.wf-flow__arrow-line {
  display: inline-block;
  width: 20px;
  height: 2px;
  background: currentColor;
}

.wf-flow__arrow-head {
  display: inline-block;
  width: 0;
  height: 0;
  border-top: 4px solid transparent;
  border-bottom: 4px solid transparent;
  border-left: 6px solid currentColor;
}

.wf-flow__label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.wf-flow__approval {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  background: rgba(245, 158, 11, 0.12);
  color: var(--warning-500, #d97706);
}

/* Table */
.wf-table-wrap {
  overflow-x: auto;
}

.wf-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.wf-table th,
.wf-table td {
  padding: 10px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
}

.wf-table th {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
  background: var(--surface-2);
}

.wf-table tbody tr:hover {
  background: var(--surface-2);
}

.wf-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}

.wf-cell--approver {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.wf-role {
  font-size: 12px;
  color: var(--text-primary);
}

.wf-approvers {
  font-size: 11px;
  color: var(--text-tertiary);
}

.wf-cell--actions {
  display: flex;
  gap: 6px;
}

.wf-muted {
  color: var(--text-tertiary);
}

/* Buttons */
.wf-btn {
  padding: 7px 14px;
  border-radius: var(--radius-sm, 6px);
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-secondary);
  transition: all 0.15s;
}

.wf-btn--primary {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-color: var(--brand-500);
}

.wf-btn--ghost {
  background: transparent;
}

.wf-btn--danger {
  color: var(--danger-500, #dc2f2f);
  border-color: var(--danger-200, rgba(220, 47, 47, 0.2));
  background: transparent;
}

.wf-btn--sm {
  padding: 5px 10px;
  font-size: 12px;
}

.wf-btn--xs {
  padding: 3px 8px;
  font-size: 11px;
}

.wf-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Form */
.wf-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.wf-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.wf-form__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.wf-form__field--checkbox {
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

.wf-form__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.wf-input,
.wf-select {
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-family: inherit;
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
}

.wf-input:focus,
.wf-select:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.wf-textarea {
  resize: vertical;
  min-height: 60px;
}

.wf-select[multiple] {
  min-height: 90px;
}

.wf-form__error {
  color: var(--danger-500, #dc2f2f);
  font-size: 12px;
  margin: 0;
}

.wf-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
}

@media (max-width: 900px) {
  .wf-layout {
    grid-template-columns: 1fr;
  }

  .wf-sidebar {
    position: static;
  }
}

@media (max-width: 600px) {
  .wf-form__row {
    grid-template-columns: 1fr;
  }

  .wf-main__head {
    flex-direction: column;
  }
}
</style>
