<script setup lang="ts">
/**
 * 创建工作项弹窗 — 类型/标题/优先级/经办人/标签等表单。
 */

import { computed, ref, watch } from "vue";

import { issueApi, type CreateIssueInput, type IssueType } from "@/api/services/issue";
import { useIssueStore } from "@/stores/issue";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  visible: boolean;
  /** 预选类型（可选，不传则由用户选择） */
  presetType?: IssueType;
  /** 从哪个工作项提缺陷（可选，用于一键从需求提缺陷） */
  sourceIssueId?: number;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "created", issueId: number): void;
}>();

const issueStore = useIssueStore();

// ---- 表单状态 ----
const currentStep = ref<"type" | "form">(props.presetType ? "form" : "type");
const selectedType = ref<IssueType>(props.presetType ?? "task");
const submitting = ref(false);
const errorMsg = ref("");

// 基础字段
const name = ref("");
const description = ref("");
const priority = ref<string>("medium");
const point = ref<number | null>(null);
const parentId = ref<number | null>(null);
const isDraft = ref(false);

// 缺陷专属字段
const severity = ref<number>(3);
const foundPhase = ref("");

// ---- 派生 ----
const requiresExtraFields = computed(() => selectedType.value === "defect");

const canSubmit = computed(() => {
  if (!name.value.trim()) return false;
  return true;
});

const typeOptions: { value: IssueType; label: string; desc: string }[] = [
  { value: "requirement", label: "需求", desc: "产品需求或用户故事，可分解为子需求" },
  { value: "task", label: "任务", desc: "开发任务，可分解为子任务，支持工时记录" },
  { value: "defect", label: "缺陷", desc: "Bug 或质量问题，包含严重程度与复现步骤" },
];

// ---- 监听 visible 重置表单 ----
watch(
  () => props.visible,
  (v) => {
    if (v) {
      currentStep.value = props.presetType ? "form" : "type";
      selectedType.value = props.presetType ?? "task";
      name.value = "";
      description.value = "";
      priority.value = "medium";
      point.value = null;
      parentId.value = null;
      isDraft.value = false;
      severity.value = 3;
      foundPhase.value = "";
      errorMsg.value = "";
    }
  },
);

// ---- Actions ----
function selectType(type: IssueType) {
  selectedType.value = type;
  currentStep.value = "form";
}

function backToType() {
  if (props.presetType) return;
  currentStep.value = "type";
}

async function submit() {
  if (!canSubmit.value || submitting.value) return;
  submitting.value = true;
  errorMsg.value = "";

  const input: CreateIssueInput = {
    type: selectedType.value,
    name: name.value.trim(),
    priority: priority.value as CreateIssueInput["priority"],
    is_draft: isDraft.value,
  };

  if (description.value.trim()) {
    input.description_html = description.value.trim();
  }
  if (point.value != null) {
    input.point = point.value;
  }
  if (parentId.value != null) {
    input.parent_id = parentId.value;
  }

  // 缺陷专属字段
  if (selectedType.value === "defect") {
    input.severity = severity.value;
    if (foundPhase.value.trim()) {
      input.found_phase = foundPhase.value.trim();
    }
  }

  // 使用默认状态（store 中已有的第一个默认状态）
  const defaultState = issueStore.states.find((s) => s.is_default);
  if (defaultState) {
    input.state_id = defaultState.id;
  }

  try {
    const iss = await issueStore.createIssue(props.workspaceId, props.projectId, input);
    emit("created", iss.id);
    emit("close");
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : "创建失败，请重试";
  } finally {
    submitting.value = false;
  }
}

function cancel() {
  emit("close");
}
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="cancel">
      <div class="modal" @click.stop>
        <!-- 标题栏 -->
        <div class="modal__header">
          <h2 class="modal__title">
            {{ currentStep === "type" ? "创建工作项" : "新建" + ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[selectedType] }}
          </h2>
          <button class="modal__close" @click="cancel" :disabled="submitting">✕</button>
        </div>

        <!-- Step 1: 选择类型 -->
        <div v-if="currentStep === 'type'" class="modal__body">
          <p class="modal__hint">请选择要创建的工作项类型</p>
          <div class="type-grid">
            <button
              v-for="opt in typeOptions"
              :key="opt.value"
              class="type-card"
              @click="selectType(opt.value)"
            >
              <span class="type-card__icon">
                {{ opt.value === "requirement" ? "📋" : opt.value === "task" ? "✅" : "🐛" }}
              </span>
              <span class="type-card__label">{{ opt.label }}</span>
              <span class="type-card__desc">{{ opt.desc }}</span>
            </button>
          </div>
        </div>

        <!-- Step 2: 填写表单 -->
        <div v-else class="modal__body">
          <div v-if="errorMsg" class="form-error">{{ errorMsg }}</div>

          <div class="form-group">
            <label class="form-label">
              名称 <span class="required">*</span>
            </label>
            <input
              v-model="name"
              class="form-input"
              placeholder="输入工作项名称"
              maxlength="500"
              autofocus
              @keydown.enter.prevent="submit"
            />
            <span class="form-hint">{{ name.length }}/500</span>
          </div>

          <div class="form-group">
            <label class="form-label">描述</label>
            <textarea
              v-model="description"
              class="form-textarea"
              placeholder="输入工作项描述（支持 Markdown）"
              rows="4"
            ></textarea>
          </div>

          <div class="form-row">
            <div class="form-group form-group--inline">
              <label class="form-label">优先级</label>
              <select v-model="priority" class="form-select">
                <option value="urgent">紧急</option>
                <option value="high">高</option>
                <option value="medium">中</option>
                <option value="low">低</option>
                <option value="none">无</option>
              </select>
            </div>

            <div v-if="selectedType !== 'defect'" class="form-group form-group--inline">
              <label class="form-label">故事点</label>
              <input v-model.number="point" type="number" class="form-input form-input--sm" min="0" max="100" placeholder="--" />
            </div>
          </div>

          <!-- 缺陷专属字段 -->
          <div v-if="requiresExtraFields" class="defect-fields">
            <div class="defect-fields__title">缺陷信息</div>
            <div class="form-row">
              <div class="form-group form-group--inline">
                <label class="form-label">
                  严重程度 <span class="required">*</span>
                </label>
                <select v-model="severity" class="form-select">
                  <option :value="5">S5 · 致命</option>
                  <option :value="4">S4 · 严重</option>
                  <option :value="3">S3 · 一般</option>
                  <option :value="2">S2 · 轻微</option>
                  <option :value="1">S1 · 建议</option>
                </select>
              </div>
              <div class="form-group form-group--inline">
                <label class="form-label">发现阶段</label>
                <select v-model="foundPhase" class="form-select">
                  <option value="">-- 请选择 --</option>
                  <option value="开发自测">开发自测</option>
                  <option value="Code Review">Code Review</option>
                  <option value="功能测试">功能测试</option>
                  <option value="集成测试">集成测试</option>
                  <option value="验收测试">验收测试</option>
                  <option value="线上监控">线上监控</option>
                  <option value="用户反馈">用户反馈</option>
                </select>
              </div>
            </div>
          </div>

          <!-- 高级选项 -->
          <details class="form-advanced">
            <summary class="form-advanced__toggle">高级选项</summary>
            <div class="form-advanced__body">
              <div class="form-group form-group--inline">
                <label class="form-label">父工作项 ID</label>
                <input
                  v-model.number="parentId"
                  type="number"
                  class="form-input form-input--sm"
                  placeholder="父级ID"
                  min="1"
                />
              </div>
              <div class="form-check">
                <input v-model="isDraft" type="checkbox" id="chk-draft" />
                <label for="chk-draft">创建为草稿</label>
              </div>
            </div>
          </details>

          <!-- 操作栏 -->
          <div class="modal__actions">
            <button v-if="!props.presetType" class="btn btn--ghost" @click="backToType" :disabled="submitting">
              ← 返回选择类型
            </button>
            <div class="spacer"></div>
            <button class="btn btn--secondary" @click="cancel" :disabled="submitting">取消</button>
            <button class="btn btn--primary" @click="submit" :disabled="!canSubmit || submitting">
              {{ submitting ? "创建中..." : "创建" }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* ===== Overlay ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

.modal {
  width: 560px;
  max-height: 90vh;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ===== Header ===== */
.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.modal__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.modal__close {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-family: inherit;
}

.modal__close:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

/* ===== Body ===== */
.modal__body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.modal__hint {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}

/* ===== Type Grid ===== */
.type-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 20px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
  text-align: center;
  font-family: inherit;
}

.type-card:hover {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.type-card__icon {
  font-size: 28px;
}

.type-card__label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.type-card__desc {
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.4;
}

/* ===== Form ===== */
.form-group {
  margin-bottom: 16px;
}

.form-group--inline {
  flex: 1;
  min-width: 0;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.required {
  color: var(--danger-500);
}

.form-input,
.form-textarea,
.form-select {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  transition: border-color 0.15s;
}

.form-input:focus,
.form-textarea:focus,
.form-select:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-input--sm {
  max-width: 120px;
}

.form-select {
  cursor: pointer;
}

.form-hint {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.form-error {
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: var(--danger-50);
  color: var(--danger-600);
  font-size: 12px;
  margin-bottom: 16px;
}

.form-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 12px;
}

.form-check input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: var(--brand-500);
}

/* ===== Defect Fields ===== */
.defect-fields {
  margin-top: 8px;
  padding: 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.defect-fields__title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

/* ===== Advanced ===== */
.form-advanced {
  margin-top: 8px;
}

.form-advanced__toggle {
  font-size: 12px;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: 4px 0;
}

.form-advanced__toggle:hover {
  color: var(--brand-500);
}

.form-advanced__body {
  padding: 12px 0 0;
}

/* ===== Actions ===== */
.modal__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.spacer {
  flex: 1;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  font-family: inherit;
  transition: background 0.15s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--primary {
  background: var(--brand-500);
  color: #fff;
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn--secondary {
  background: var(--surface-1);
  color: var(--text-secondary);
  border-color: var(--border-default);
}

.btn--secondary:hover:not(:disabled) {
  background: var(--surface-3);
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--text-tertiary);
  font-size: 12px;
  padding: 4px 0;
}

.btn--ghost:hover:not(:disabled) {
  color: var(--brand-500);
}
</style>
