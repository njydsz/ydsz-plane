<script setup lang="ts">
/**
 * StateMachineConfigView — 项目状态机配置视图。
 *
 * 能力：
 *  - 状态列表展示（颜色徽章 / 分组 / 适用类型 / 排序）
*  - 创建 / 编辑状态（名称、分组、颜色、排序、适用需求/任务/缺陷类型）
*  - 删除状态（有需求/任务/缺陷使用时拒绝）
 *  - 流转规则列表（矩阵视图 + 列表视图切换）
 *  - 添加 / 删除流转规则
 *  - 必填字段配置（针对特定流转）
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import {
  stateMachineApi,
  STATE_GROUP_LABELS,
  STATE_GROUP_COLORS,
  TYPE_CODE_LABELS,
  type State,
  type StateGroup,
  type TransitionRule,
  type CreateStateRequest,
  type UpdateStateRequest,
  type AddTransitionRequest,
} from "@/api/services/stateMachine";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* Workspace context                                                   */
/* ------------------------------------------------------------------ */

const route = useRoute();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ------------------------------------------------------------------ */
/* Data state                                                          */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const states = ref<State[]>([]);
const transitions = ref<TransitionRule[]>([]);

/* ------------------------------------------------------------------ */
/* View mode                                                           */
/* ------------------------------------------------------------------ */

type ViewMode = 'states' | 'transitions';
const viewMode = ref<ViewMode>('states');

/* ------------------------------------------------------------------ */
/* State form                                                          */
/* ------------------------------------------------------------------ */

const showStateForm = ref(false);
const editingState = ref<State | null>(null);
const stateFormSaving = ref(false);

const stateForm = ref<CreateStateRequest & { name: string }>({
  name: '',
  group: 'backlog',
  color: '#8DA2C2',
  sequence: 65535,
  is_default: false,
  applicable_types: ['all'],
});

function resetStateForm() {
  stateForm.value = {
    name: '',
    group: 'backlog',
    color: '#8DA2C2',
    sequence: 65535,
    is_default: false,
    applicable_types: ['all'],
  };
  editingState.value = null;
}

function openCreateState() {
  resetStateForm();
  showStateForm.value = true;
}

function openEditState(st: State) {
  editingState.value = st;
  stateForm.value = {
    name: st.name,
    group: st.group,
    color: st.color,
    sequence: st.sequence,
    is_default: st.is_default,
    applicable_types: st.applicable_types,
  };
  showStateForm.value = true;
}

async function saveState() {
  if (!wsId.value || !stateForm.value.name.trim()) return;
  stateFormSaving.value = true;
  try {
    if (editingState.value) {
      await stateMachineApi.updateState(wsId.value, projectId.value, editingState.value.id, stateForm.value);
      toast.success('状态已更新');
    } else {
      await stateMachineApi.createState(wsId.value, projectId.value, stateForm.value);
      toast.success('状态已创建');
    }
    showStateForm.value = false;
    resetStateForm();
    await loadStates();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '保存失败';
    toast.error(msg);
  } finally {
    stateFormSaving.value = false;
  }
}

async function deleteState(st: State) {
  if (!wsId.value) return;
  if (!confirm(`确认删除状态「${st.name}」？此操作不可恢复。`)) return;
  try {
    await stateMachineApi.deleteState(wsId.value, projectId.value, st.id);
    toast.success('状态已删除');
    await loadStates();
    await loadTransitions();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '删除失败';
    toast.error(msg);
  }
}

/* ------------------------------------------------------------------ */
/* Transition form                                                     */
/* ------------------------------------------------------------------ */

const showTransitionForm = ref(false);
const transitionFormSaving = ref(false);
const transitionForm = ref<AddTransitionRequest>({
  from_state_id: 0,
  to_state_id: 0,
  type_code: 'all',
  required_fields: [],
});

function resetTransitionForm() {
  transitionForm.value = {
    from_state_id: 0,
    to_state_id: 0,
    type_code: 'all',
    required_fields: [],
  };
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
  if (!wsId.value) return;
  if (transitionForm.value.from_state_id === transitionForm.value.to_state_id) {
    toast.error('起止状态不能相同');
    return;
  }
  transitionFormSaving.value = true;
  try {
    await stateMachineApi.addTransition(wsId.value, projectId.value, transitionForm.value);
    toast.success('流转规则已添加');
    showTransitionForm.value = false;
    resetTransitionForm();
    await loadTransitions();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '添加失败';
    toast.error(msg);
  } finally {
    transitionFormSaving.value = false;
  }
}

async function removeTransition(rule: TransitionRule) {
  if (!wsId.value) return;
  if (!confirm(`确认删除流转规则「${rule.from_state_name || rule.from_state_id} → ${rule.to_state_name || rule.to_state_id}」？`)) return;
  try {
    await stateMachineApi.removeTransition(wsId.value, projectId.value, rule.id);
    toast.success('流转规则已删除');
    await loadTransitions();
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '删除失败';
    toast.error(msg);
  }
}

/* ------------------------------------------------------------------ */
/* Matrix view helpers                                                 */
/* ------------------------------------------------------------------ */

/** 流转矩阵：from_id → to_id → rule */
const transitionMatrix = computed(() => {
  const matrix = new Map<number, Map<number, TransitionRule>>();
  for (const rule of transitions.value) {
    if (!matrix.has(rule.from_state_id)) {
      matrix.set(rule.from_state_id, new Map());
    }
    matrix.get(rule.from_state_id)!.set(rule.to_state_id, rule);
  }
  return matrix;
});

function hasTransition(fromId: number, toId: number): boolean {
  return transitionMatrix.value.get(fromId)?.has(toId) ?? false;
}

/* ------------------------------------------------------------------ */
/* Load data                                                           */
/* ------------------------------------------------------------------ */

async function loadStates() {
  if (!wsId.value) return;
  const data = await stateMachineApi.getStates(wsId.value, projectId.value);
  states.value = data;
}

async function loadTransitions() {
  if (!wsId.value) return;
  const data = await stateMachineApi.getTransitions(wsId.value, projectId.value);
  transitions.value = data;
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = '';
  try {
    await Promise.all([loadStates(), loadTransitions()]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败';
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (wsId.value) void load();
});

watch([wsId], () => {
  if (wsId.value) void load();
});

/* ------------------------------------------------------------------ */
/* Options                                                             */
/* ------------------------------------------------------------------ */

const groupOptions: { value: StateGroup; label: string }[] = [
  { value: 'backlog', label: '待办' },
  { value: 'started', label: '进行中' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' },
];

const typeCodeOptions: { value: string; label: string }[] = [
  { value: 'all', label: '全部类型' },
  { value: 'requirement', label: '需求' },
  { value: 'task', label: '任务' },
  { value: 'defect', label: '缺陷' },
  { value: 'epic', label: '史诗' },
];

const requiredFieldOptions = [
  { value: 'root_cause_category', label: '根因分类' },
  { value: 'fix_version_id', label: '修复版本' },
  { value: 'verifier_id', label: '验证人' },
  { value: 'severity', label: '严重程度' },
];
</script>

<template>
  <div class="sm-config">
    <!-- 页头 -->
    <header class="sm-config__header">
      <div>
        <h1 class="sm-config__title">状态机配置</h1>
        <p class="sm-config__subtitle">管理项目需求/任务/缺陷的状态与流转规则</p>
      </div>
    </header>

    <!-- 加载状态 -->
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 视图切换 Tab -->
      <div class="sm-config__tabs">
        <button
          class="sm-config__tab"
          :class="{ 'sm-config__tab--active': viewMode === 'states' }"
          @click="viewMode = 'states'"
        >
          状态管理
          <span class="sm-config__badge">{{ states.length }}</span>
        </button>
        <button
          class="sm-config__tab"
          :class="{ 'sm-config__tab--active': viewMode === 'transitions' }"
          @click="viewMode = 'transitions'"
        >
          流转规则
          <span class="sm-config__badge">{{ transitions.length }}</span>
        </button>
      </div>

      <!-- 状态管理视图 -->
      <section v-if="viewMode === 'states'" class="sm-config__section">
        <div class="sm-config__toolbar">
          <button class="sm-btn sm-btn--primary" @click="openCreateState">
            + 新建状态
          </button>
        </div>

        <AppEmptyState v-if="!states.length" description="暂无状态，请创建" />

        <div v-else class="sm-state-grid">
          <div
            v-for="st in states"
            :key="st.id"
            class="sm-state-card"
            :style="{ borderLeftColor: st.color }"
          >
            <div class="sm-state-card__header">
              <span class="sm-state-card__color" :style="{ backgroundColor: st.color }" />
              <span class="sm-state-card__name">{{ st.name }}</span>
              <span v-if="st.is_default" class="sm-state-card__default">默认</span>
            </div>
            <div class="sm-state-card__meta">
              <span class="sm-tag sm-tag--group">{{ STATE_GROUP_LABELS[st.group] || st.group }}</span>
              <span v-for="t in st.applicable_types" :key="t" class="sm-tag">
                {{ TYPE_CODE_LABELS[t] || t }}
              </span>
            </div>
            <div class="sm-state-card__actions">
              <button class="sm-btn sm-btn--ghost" @click="openEditState(st)">编辑</button>
              <button class="sm-btn sm-btn--danger" @click="deleteState(st)">删除</button>
            </div>
          </div>
        </div>
      </section>

      <!-- 流转规则视图 -->
      <section v-else class="sm-config__section">
        <div class="sm-config__toolbar">
          <button class="sm-btn sm-btn--primary" @click="openAddTransition">
            + 添加流转规则
          </button>
        </div>

        <AppEmptyState v-if="!transitions.length" description="暂无流转规则" />

        <!-- 流转矩阵 -->
        <div v-else class="sm-matrix-wrapper">
          <table class="sm-matrix">
            <thead>
              <tr>
                <th class="sm-matrix__corner">从 \ 到</th>
                <th v-for="toState in states" :key="toState.id" class="sm-matrix__col-head">
                  <span class="sm-matrix__dot" :style="{ backgroundColor: toState.color }" />
                  {{ toState.name }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="fromState in states" :key="fromState.id">
                <th class="sm-matrix__row-head">
                  <span class="sm-matrix__dot" :style="{ backgroundColor: fromState.color }" />
                  {{ fromState.name }}
                </th>
                <td
                  v-for="toState in states"
                  :key="toState.id"
                  class="sm-matrix__cell"
                  :class="{ 'sm-matrix__cell--has': hasTransition(fromState.id, toState.id) }"
                >
                  <span v-if="fromState.id === toState.id" class="sm-matrix__self">—</span>
                  <span v-else-if="hasTransition(fromState.id, toState.id)" class="sm-matrix__check">✓</span>
                  <span v-else class="sm-matrix__empty">·</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 流转规则列表 -->
        <div class="sm-transition-list">
          <div
            v-for="rule in transitions"
            :key="rule.id"
            class="sm-transition-item"
          >
            <span class="sm-transition-item__from">{{ rule.from_state_name || rule.from_state_id }}</span>
            <span class="sm-transition-item__arrow">→</span>
            <span class="sm-transition-item__to">{{ rule.to_state_name || rule.to_state_id }}</span>
            <span v-if="rule.type_code !== 'all'" class="sm-tag">{{ TYPE_CODE_LABELS[rule.type_code] || rule.type_code }}</span>
            <span v-if="rule.required_fields?.length" class="sm-transition-item__required">
              必填: {{ rule.required_fields.join(', ') }}
            </span>
            <button class="sm-btn sm-btn--danger sm-btn--xs" @click="removeTransition(rule)">删除</button>
          </div>
        </div>
      </section>
    </template>

    <!-- 状态表单弹窗 -->
    <Teleport to="body">
      <div v-if="showStateForm" class="sm-modal" @click.self="showStateForm = false">
        <div class="sm-modal__panel">
          <h2 class="sm-modal__title">{{ editingState ? '编辑状态' : '新建状态' }}</h2>
          <form class="sm-form" @submit.prevent="saveState">
            <label class="sm-form__field">
              <span class="sm-form__label">状态名称 *</span>
              <input v-model="stateForm.name" class="sm-input" placeholder="如：In Review" required />
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">状态分组</span>
              <select v-model="stateForm.group" class="sm-select">
                <option v-for="opt in groupOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">颜色</span>
              <input v-model="stateForm.color" type="color" class="sm-input sm-input--color" />
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">排序值</span>
              <input v-model.number="stateForm.sequence" type="number" class="sm-input" step="1000" />
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">适用需求/任务/缺陷类型</span>
              <select v-model="stateForm.applicable_types" class="sm-select" multiple>
                <option v-for="opt in typeCodeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </label>
            <label class="sm-form__field sm-form__field--checkbox">
              <input v-model="stateForm.is_default" type="checkbox" />
              <span>设为默认状态</span>
            </label>
            <div class="sm-modal__actions">
              <button type="button" class="sm-btn sm-btn--ghost" @click="showStateForm = false">取消</button>
              <button type="submit" class="sm-btn sm-btn--primary" :disabled="stateFormSaving">
                {{ stateFormSaving ? '保存中...' : '保存' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- 流转规则表单弹窗 -->
    <Teleport to="body">
      <div v-if="showTransitionForm" class="sm-modal" @click.self="showTransitionForm = false">
        <div class="sm-modal__panel">
          <h2 class="sm-modal__title">添加流转规则</h2>
          <form class="sm-form" @submit.prevent="saveTransition">
            <label class="sm-form__field">
              <span class="sm-form__label">起始状态 *</span>
              <select v-model.number="transitionForm.from_state_id" class="sm-select">
                <option v-for="st in states" :key="st.id" :value="st.id">{{ st.name }}</option>
              </select>
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">目标状态 *</span>
              <select v-model.number="transitionForm.to_state_id" class="sm-select">
                <option v-for="st in states" :key="st.id" :value="st.id">{{ st.name }}</option>
              </select>
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">适用需求/任务/缺陷类型</span>
              <select v-model="transitionForm.type_code" class="sm-select">
                <option v-for="opt in typeCodeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </label>
            <label class="sm-form__field">
              <span class="sm-form__label">必填字段（可选）</span>
              <select v-model="transitionForm.required_fields" class="sm-select" multiple>
                <option v-for="opt in requiredFieldOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </label>
            <div class="sm-modal__actions">
              <button type="button" class="sm-btn sm-btn--ghost" @click="showTransitionForm = false">取消</button>
              <button type="submit" class="sm-btn sm-btn--primary" :disabled="transitionFormSaving">
                {{ transitionFormSaving ? '添加中...' : '添加' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.sm-config {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.sm-config__header {
  margin-bottom: 24px;
}

.sm-config__title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.sm-config__subtitle {
  color: var(--text-secondary);
  margin: 4px 0 0;
}

.sm-config__tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
}

.sm-config__tab {
  padding: 10px 16px;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}

.sm-config__tab--active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  font-weight: 500;
}

.sm-config__badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--surface-2);
  font-size: 12px;
  margin-left: 4px;
}

.sm-config__toolbar {
  margin-bottom: 16px;
}

.sm-state-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.sm-state-card {
  background: var(--surface-1);
  border-radius: 8px;
  padding: 16px;
  border-left: 4px solid;
  border: 1px solid var(--border-color);
  border-left-width: 4px;
}

.sm-state-card__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.sm-state-card__color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.sm-state-card__name {
  font-weight: 500;
  flex: 1;
}

.sm-state-card__default {
  font-size: 11px;
  padding: 2px 6px;
  background: var(--primary-light);
  color: var(--primary);
  border-radius: 4px;
}

.sm-state-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 12px;
}

.sm-tag {
  font-size: 11px;
  padding: 2px 8px;
  background: var(--surface-2);
  border-radius: 4px;
  color: var(--text-secondary);
}

.sm-tag--group {
  background: var(--primary-light);
  color: var(--primary);
}

.sm-state-card__actions {
  display: flex;
  gap: 8px;
}

/* Buttons */
.sm-btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--surface-1);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s;
}

.sm-btn--primary {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.sm-btn--ghost {
  background: transparent;
}

.sm-btn--danger {
  color: var(--danger);
  border-color: var(--danger);
  background: transparent;
}

.sm-btn--xs {
  padding: 2px 8px;
  font-size: 12px;
}

.sm-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Matrix */
.sm-matrix-wrapper {
  overflow-x: auto;
  margin-bottom: 24px;
}

.sm-matrix {
  border-collapse: collapse;
  width: 100%;
  font-size: 13px;
}

.sm-matrix th,
.sm-matrix td {
  padding: 8px 12px;
  text-align: center;
  border: 1px solid var(--border-color);
}

.sm-matrix__corner {
  background: var(--surface-2);
}

.sm-matrix__col-head,
.sm-matrix__row-head {
  background: var(--surface-2);
  font-weight: 500;
  white-space: nowrap;
}

.sm-matrix__dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
}

.sm-matrix__cell {
  color: var(--text-tertiary);
}

.sm-matrix__cell--has {
  background: var(--primary-light);
  color: var(--primary);
  font-weight: 600;
}

.sm-matrix__self {
  color: var(--text-tertiary);
}

.sm-matrix__check {
  font-weight: 700;
}

/* Transition list */
.sm-transition-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sm-transition-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--surface-1);
  border-radius: 6px;
  border: 1px solid var(--border-color);
  font-size: 13px;
}

.sm-transition-item__from,
.sm-transition-item__to {
  font-weight: 500;
}

.sm-transition-item__arrow {
  color: var(--text-tertiary);
}

.sm-transition-item__required {
  font-size: 11px;
  color: var(--text-secondary);
  margin-left: auto;
}

/* Modal */
.sm-modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.sm-modal__panel {
  background: var(--surface-1);
  border-radius: 12px;
  padding: 24px;
  width: 480px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
}

.sm-modal__title {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 20px;
}

.sm-modal__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
}

/* Form */
.sm-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sm-form__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sm-form__field--checkbox {
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

.sm-form__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.sm-input,
.sm-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 14px;
  background: var(--surface-1);
}

.sm-input--color {
  width: 60px;
  height: 36px;
  padding: 2px;
  cursor: pointer;
}

.sm-select[multiple] {
  min-height: 100px;
}
</style>
