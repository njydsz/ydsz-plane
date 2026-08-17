<script setup lang="ts">
/**
 * 创建需求/任务/缺陷弹窗 — 类型/标题/优先级/经办人/标签等表单。
 */

import { computed, ref, watch } from "vue";

import { aiApi, type ClassifyResult, type DuplicateCandidate } from "@/api/services/ai";
import { type CreateIssueInput, type IssueType } from "@/api/services/issue";
import { versionApi, type Version } from "@/api/services/version";
import { useIssueStore } from "@/stores/issue";
import { RichTextEditor } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  visible: boolean;
  /** 预选类型（可选，不传则由用户选择） */
  presetType?: IssueType;
  /** 从哪个需求/任务/缺陷提缺陷（可选，用于一键从需求提缺陷） */
  sourceIssueId?: number;
  /** 父需求/任务/缺陷 ID（可选，用于 WBS 子需求/任务/缺陷创建） */
  parentId?: number;
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
const priorityRef = ref<string>("medium");
const point = ref<number | null>(null);
const parentIdInput = ref<number | null>(null);
const isDraft = ref(false);

// 缺陷专属字段
const severity = ref<number>(3);
const foundPhase = ref("");
const reproduceSteps = ref("");
const reproduceExpected = ref("");
const reproduceActual = ref("");
const environment = ref("");
const foundVersionId = ref<number | null>(null);
const fixVersionId = ref<number | null>(null);
const rootCauseCategory = ref("");
const verifierId = ref("");
const versions = ref<Version[]>([]);

// ---- AI 智能辅助 ----
const aiEnabled = ref(false);
const classifyResult = ref<ClassifyResult | null>(null);
const duplicates = ref<DuplicateCandidate[]>([]);
const aiLoading = ref(false);
const aiError = ref<string | null>(null);
let aiAbort: AbortController | null = null;
let aiTimer: ReturnType<typeof setTimeout> | null = null;

/** 检测 AI 功能是否可用 */
async function checkAiStatus() {
  try {
    const status = await aiApi.getStatus(props.workspaceId, props.projectId);
    aiEnabled.value = status.enabled;
    aiError.value = null;
  } catch {
    aiEnabled.value = false;
    aiError.value = "AI 服务端暂不可用，请稍后重试或在系统设置中检查 AI 集成（如未配置大模型 Provider，推理功能将不可用）。";
  }
}

/** 防抖调用 AI 分类 + 重复检测 */
function scheduleAiCheck() {
  if (!aiEnabled.value || currentStep.value !== "form") return;
  if (aiTimer) clearTimeout(aiTimer);
  aiTimer = setTimeout(runAiCheck, 800);
}

async function runAiCheck() {
  const title = name.value.trim();
  if (title.length < 4) return;

  aiAbort?.abort();
  aiAbort = new AbortController();
  aiLoading.value = true;
  aiError.value = null;

  // 使用 allSettled，单条推理失败也保留另一条结果，并向用户给出可见提示
  const [classifyRes, dupRes] = await Promise.allSettled([
    aiApi.smartClassify(props.workspaceId, props.projectId, title, description.value),
    aiApi.detectDuplicates(props.workspaceId, props.projectId, title, description.value),
  ]);

  classifyResult.value = classifyRes.status === "fulfilled" ? classifyRes.value : null;
  duplicates.value = dupRes.status === "fulfilled" ? dupRes.value : [];

  const anyFailed = classifyRes.status === "rejected" || dupRes.status === "rejected";
  if (anyFailed) {
    aiError.value = "部分 AI 推理未能完成（可能是 AI 服务未启用或 Provider 未配置）。分类/查重结果可能不完整。";
  }
  aiLoading.value = false;
}

/** 采纳 AI 推荐类型 */
function applyAiType(type: string) {
  if (["requirement", "task", "defect"].includes(type)) {
    selectedType.value = type as IssueType;
    classifyResult.value = null;
  }
}

/** 采纳 AI 推荐优先级 */
function applyAiPriority(priority: string) {
  if (["critical", "high", "medium", "low"].includes(priority)) {
    priorityRef.value = priority;
    classifyResult.value = null;
  }
}

// ---- 派生 ----
const requiresExtraFields = computed(() => selectedType.value === "defect");

const canSubmit = computed(() => {
  if (!name.value.trim()) return false;
  return true;
});

// 加载版本列表（缺陷表单用）
async function loadVersions() {
  try {
    const res = await versionApi.listVersions(props.workspaceId, props.projectId);
    versions.value = Array.isArray(res) ? res : (res as { results: Version[] }).results ?? [];
  } catch { /* 版本列表不可用时静默忽略 */ }
}

const typeOptions: { value: IssueType; label: string; desc: string }[] = [
  { value: "epic", label: "史诗", desc: "顶层容器，包含多个需求/任务/缺陷，对标 Plane Epic" },
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
      priorityRef.value = "medium";
      point.value = null;
      parentIdInput.value = props.parentId ?? null;
      isDraft.value = false;
      severity.value = 3;
      foundPhase.value = "";
      reproduceSteps.value = "";
      reproduceExpected.value = "";
      reproduceActual.value = "";
      environment.value = "";
      foundVersionId.value = null;
      fixVersionId.value = null;
      rootCauseCategory.value = "";
      verifierId.value = "";
      errorMsg.value = "";
      classifyResult.value = null;
      duplicates.value = [];
      aiError.value = null;
      loadVersions();
      void checkAiStatus();
    }
  },
);

// ---- AI：标题/描述变化时触发防抖检测 ----
watch([name, description], () => {
  scheduleAiCheck();
});

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
    priority: priorityRef.value as CreateIssueInput["priority"],
    is_draft: isDraft.value,
  };

  if (description.value.trim()) {
    input.description_html = description.value.trim();
  }
  if (point.value != null) {
    input.point = point.value;
  }
  if (parentIdInput.value != null) {
    input.parent_id = parentIdInput.value;
  }

  // 缺陷专属字段
  if (selectedType.value === "defect") {
    input.severity = severity.value;
    if (foundPhase.value.trim()) {
      input.found_phase = foundPhase.value.trim();
    }
    // reproduce_steps 拆分为三个字段
    const rsSteps = reproduceSteps.value.trim();
    const rsExpected = reproduceExpected.value.trim();
    const rsActual = reproduceActual.value.trim();
    if (rsSteps || rsExpected || rsActual) {
      input.reproduce_steps = {
        steps: rsSteps,
        expected: rsExpected,
        actual: rsActual,
      };
    }
    if (environment.value.trim()) {
      input.environment = { value: environment.value.trim() };
    }
    if (foundVersionId.value != null) {
      input.found_version_id = foundVersionId.value;
    }
    if (fixVersionId.value != null) {
      input.fix_version_id = fixVersionId.value;
    }
    if (rootCauseCategory.value) {
      input.root_cause_category = rootCauseCategory.value;
    }
    if (verifierId.value.trim()) {
      const vid = Number(verifierId.value.trim());
      if (!isNaN(vid) && vid > 0) {
        input.verifier_id = vid;
      }
    }
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
            {{ currentStep === "type" ? "创建需求/任务/缺陷" : "新建" + ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[selectedType] }}
          </h2>
          <button class="modal__close" :disabled="submitting" @click="cancel">✕</button>
        </div>

        <!-- Step 1: 选择类型 -->
        <div v-if="currentStep === 'type'" class="modal__body">
          <p class="modal__hint">请选择要创建的需求/任务/缺陷类型</p>
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
              placeholder="输入需求/任务/缺陷名称"
              maxlength="500"
              autofocus
              @keydown.enter.prevent="submit"
            />
            <span class="form-hint">{{ name.length }}/500</span>
          </div>

          <!-- AI 智能建议条 -->
          <div v-if="aiEnabled && currentStep === 'form' && (classifyResult || duplicates.length > 0 || aiLoading || aiError)" class="ai-suggestions">
            <div v-if="aiLoading" class="ai-suggestions__loading">
              <span class="ai-spinner"></span> AI 正在分析...
            </div>
            <div v-else-if="aiError" class="ai-suggestions__error">
              <span class="ai-tag ai-tag--warn">AI</span>
              <span class="ai-error-text">{{ aiError }}</span>
            </div>
            <template v-else>
              <div v-if="classifyResult" class="ai-suggestions__classify">
                <span class="ai-tag">AI 推荐</span>
                <span class="ai-label">类型</span>
                <button class="ai-pill" :class="{ 'ai-pill--active': selectedType === classifyResult.type_code }" @click="applyAiType(classifyResult.type_code)">
                  {{ ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[classifyResult.type_code] }}
                </button>
                <span class="ai-label">优先级</span>
                <button class="ai-pill" :class="{ 'ai-pill--active': priorityRef === classifyResult.priority }" @click="applyAiPriority(classifyResult.priority)">
                  {{ ({ critical: "紧急", high: "高", medium: "中", low: "低" } as Record<string, string>)[classifyResult.priority] }}
                </button>
                <span class="ai-confidence">{{ Math.round(classifyResult.confidence * 100) }}%</span>
              </div>
              <div v-if="duplicates.length > 0" class="ai-suggestions__dups">
                <span class="ai-tag ai-tag--warn">可能重复</span>
                <span v-for="dup in duplicates.slice(0, 3)" :key="dup.issue_id" class="ai-dup-item">
                  {{ dup.identifier }} · {{ dup.title }} ({{ Math.round(dup.similarity * 100) }}%)
                </span>
              </div>
            </template>
          </div>

          <div class="form-group">
            <label class="form-label">描述</label>
            <RichTextEditor
              v-model:content-html="description"
              variant="comment"
              placeholder="输入需求/任务/缺陷描述..."
              :min-height="'80px'"
              :workspace-id="workspaceId"
              :project-id="projectId"
            />
          </div>

          <div class="form-row">
            <div class="form-group form-group--inline">
              <label class="form-label">优先级</label>
              <select v-model="priorityRef" class="form-select">
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
                  <option value="unit">单元测试</option>
                  <option value="integration">集成测试</option>
                  <option value="uat">验收测试 (UAT)</option>
                  <option value="production">生产环境</option>
                  <option value="customer">用户反馈</option>
                </select>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group form-group--inline">
                <label class="form-label">发现版本</label>
                <select v-model="foundVersionId" class="form-select">
                  <option :value="null">-- 请选择 --</option>
                  <option v-for="v in versions" :key="v.id" :value="v.id">{{ v.name }}</option>
                </select>
              </div>
              <div class="form-group form-group--inline">
                <label class="form-label">修复版本</label>
                <select v-model="fixVersionId" class="form-select">
                  <option :value="null">-- 请选择 --</option>
                  <option v-for="v in versions" :key="v.id" :value="v.id">{{ v.name }}</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">复现步骤</label>
              <textarea
                v-model="reproduceSteps"
                class="form-input form-input--textarea"
                rows="3"
                placeholder="描述缺陷的复现操作路径（1. 步骤A 2. 步骤B 3. ...）"
              ></textarea>
            </div>

            <div class="form-group">
              <label class="form-label">期望结果</label>
              <textarea
                v-model="reproduceExpected"
                class="form-input form-input--textarea"
                rows="2"
                placeholder="描述正常情况下应该出现的结果"
              ></textarea>
            </div>

            <div class="form-group">
              <label class="form-label">实际结果</label>
              <textarea
                v-model="reproduceActual"
                class="form-input form-input--textarea"
                rows="2"
                placeholder="描述实际出现的异常/错误结果"
              ></textarea>
            </div>

            <div class="form-row">
              <div class="form-group form-group--inline">
                <label class="form-label">根因分类</label>
                <select v-model="rootCauseCategory" class="form-select">
                  <option value="">-- 请选择 --</option>
                  <option value="requirement">需求问题</option>
                  <option value="technical">技术问题</option>
                  <option value="environment">环境问题</option>
                  <option value="data">数据问题</option>
                  <option value="other">其他</option>
                </select>
              </div>
              <div class="form-group form-group--inline">
                <label class="form-label">验证人</label>
                <input
                  v-model="verifierId"
                  type="text"
                  class="form-input"
                  placeholder="输入验证人用户ID"
                />
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">环境</label>
              <input
                v-model="environment"
                type="text"
                class="form-input"
                placeholder="如：Chrome 120 / Windows 11 / 测试环境"
              />
            </div>
          </div>

          <!-- 高级选项 -->
          <details class="form-advanced">
            <summary class="form-advanced__toggle">高级选项</summary>
            <div class="form-advanced__body">
              <div class="form-group form-group--inline">
                <label class="form-label">父需求/任务/缺陷 ID</label>
                <input
                  v-model.number="parentIdInput"
                  type="number"
                  class="form-input form-input--sm"
                  placeholder="父级ID"
                  min="1"
                />
              </div>
              <div class="form-check">
                <input id="chk-draft" v-model="isDraft" type="checkbox" />
                <label for="chk-draft">创建为草稿</label>
              </div>
            </div>
          </details>

          <!-- 操作栏 -->
          <div class="modal__actions">
            <button v-if="!props.presetType" class="btn btn--ghost" :disabled="submitting" @click="backToType">
              ← 返回选择类型
            </button>
            <div class="spacer"></div>
            <button class="btn btn--secondary" :disabled="submitting" @click="cancel">取消</button>
            <button class="btn btn--primary" :disabled="!canSubmit || submitting" @click="submit">
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
  color: var(--text-on-brand);
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

/* ===== AI Suggestions ===== */
.ai-suggestions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  padding: 10px 12px;
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.ai-suggestions__loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-tertiary);
}

.ai-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: ai-spin 0.6s linear infinite;
}

@keyframes ai-spin {
  to { transform: rotate(360deg); }
}

.ai-suggestions__classify,
.ai-suggestions__dups {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.ai-tag {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 3px;
  background: var(--brand-50);
  color: var(--brand-600);
}

.ai-tag--warn {
  background: var(--warning-50);
  color: var(--warning-600);
}

.ai-label {
  color: var(--text-tertiary);
  font-size: 11px;
}

.ai-pill {
  padding: 2px 8px;
  font-size: 11px;
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.ai-pill:hover {
  border-color: var(--brand-500);
  color: var(--brand-500);
}

.ai-pill--active {
  border-color: var(--brand-500);
  background: var(--brand-50);
  color: var(--brand-600);
}

.ai-confidence {
  color: var(--text-tertiary);
  font-size: 10px;
}

.ai-dup-item {
  color: var(--text-secondary);
  font-size: 11px;
}
</style>
