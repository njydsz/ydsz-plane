<script setup lang="ts">
/**
 * RecurringRulesView — 项目级「重复工作项」规则管理页。
 *
 * 能力：
 *  - 规则列表：标题模板 / 类型 / Cron 表达式 / 下次生成时间 / 已生成次数 / 启用状态
 *  - 新建 / 编辑规则（标题模板、类型、Cron 表达式、描述、指派人、标签、初始状态、启用开关）
 *  - 启用 / 停用 toggle
 *  - 立即执行一次（run now）验证规则配置
 *  - 删除规则
 *
 * Cron 表达式提供常用预设与格式说明，降低配置门槛。
 */
import { computed, onMounted, reactive, ref, watch } from "vue";

import {
  recurringApi,
  type CreateRecurringRuleInput,
  type RecurringRule,
  type RecurringWorkItemType,
  type UpdateRecurringRuleInput,
} from "@/api/services/recurring";
import { AppBadge, AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer } from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const { t } = useTranslation();
const { wsId, projectId, ready, error: ctxError } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

const TYPE_OPTIONS: { value: RecurringWorkItemType; labelKey: string }[] = [
  { value: "requirement", labelKey: "recurring.typeRequirement" },
  { value: "task", labelKey: "recurring.typeTask" },
  { value: "defect", labelKey: "recurring.typeDefect" },
];

/** Cron 常用预设（5 段：分 时 日 月 周） */
const CRON_PRESETS: { labelKey: string; cron: string }[] = [
  { labelKey: "recurring.cronEveryDay", cron: "0 9 * * *" },
  { labelKey: "recurring.cronEveryWeek", cron: "0 9 * * 1" },
  { labelKey: "recurring.cronEveryMonth", cron: "0 9 1 * *" },
  { labelKey: "recurring.cronEveryQuarter", cron: "0 9 1 */3 *" },
  { labelKey: "recurring.cronEveryYear", cron: "0 9 1 1 *" },
];

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const loadError = ref("");
const rules = ref<RecurringRule[]>([]);

async function loadRules() {
  if (!ready.value) return;
  loading.value = true;
  loadError.value = "";
  try {
    const res = await recurringApi.list(wsId.value, projectId.value);
    rules.value = res.results ?? [];
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : t("recurring.loadFailed");
  } finally {
    loading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function typeLabel(type: RecurringWorkItemType): string {
  return t(
    TYPE_OPTIONS.find((o) => o.value === type)?.labelKey ?? "recurring.typeTask",
  );
}

function formatTime(iso?: string | null): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "-";
  return d.toLocaleString("zh-CN", { hour12: false });
}

/** 标题模板预览：把 {date} 占位符替换为今天 */
function previewTitle(template: string): string {
  const today = new Date().toISOString().slice(0, 10);
  return template.includes("{date}") ? template.replace(/\{date\}/g, today) : template;
}

/* ------------------------------------------------------------------ */
/* 表单状态                                                            */
/* ------------------------------------------------------------------ */

interface RecurringForm {
  title_template: string;
  work_item_type: RecurringWorkItemType;
  cron_expr: string;
  description_template: string;
  /** 指派人：逗号分隔的成员 ID 文本 */
  assigneesText: string;
  /** 标签：逗号分隔的标签 ID 文本 */
  labelsText: string;
  /** 初始状态 ID（空串表示不指定） */
  stateIdText: string;
  enabled: boolean;
}

function emptyForm(): RecurringForm {
  return {
    title_template: "",
    work_item_type: "task",
    cron_expr: "",
    description_template: "",
    assigneesText: "",
    labelsText: "",
    stateIdText: "",
    enabled: true,
  };
}

const showForm = ref(false);
const editingRule = ref<RecurringRule | null>(null);
const saving = ref(false);
const formError = ref("");
const form = reactive<RecurringForm>(emptyForm());

const isEditing = computed(() => editingRule.value !== null);

/** 解析逗号/空格分隔的 ID 文本为数字数组 */
function parseIds(text: string): number[] {
  return text
    .split(/[,，\s]+/)
    .map((s) => s.trim())
    .filter(Boolean)
    .map((s) => Number(s))
    .filter((n) => Number.isInteger(n) && n > 0);
}

function buildPayload(): CreateRecurringRuleInput {
  const stateId = form.stateIdText.trim();
  return {
    work_item_type: form.work_item_type,
    title_template: form.title_template.trim(),
    description_template: form.description_template.trim() || undefined,
    cron_expr: form.cron_expr.trim(),
    assignee_ids: parseIds(form.assigneesText),
    label_ids: parseIds(form.labelsText),
    state_id: stateId ? Number(stateId) : 0,
    enabled: form.enabled,
  };
}

/** 基础校验：标题模板 + Cron 表达式必填，Cron 必须为 5 段 */
function validateForm(): string {
  if (!form.title_template.trim()) return t("recurring.titleRequired");
  if (!form.cron_expr.trim()) return t("recurring.cronRequired");
  if (form.cron_expr.trim().split(/\s+/).length !== 5) {
    return t("recurring.cronFormat");
  }
  return "";
}

function openCreate() {
  editingRule.value = null;
  Object.assign(form, emptyForm());
  formError.value = "";
  showForm.value = true;
}

function openEdit(rule: RecurringRule) {
  editingRule.value = rule;
  Object.assign(form, {
    title_template: rule.title_template,
    work_item_type: rule.work_item_type,
    cron_expr: rule.cron_expr,
    description_template: rule.description_template ?? "",
    assigneesText: (rule.assignee_ids ?? []).join(", "),
    labelsText: (rule.label_ids ?? []).join(", "),
    stateIdText: rule.state_id ? String(rule.state_id) : "",
    enabled: rule.enabled,
  });
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingRule.value = null;
  formError.value = "";
}

function applyPreset(cron: string) {
  form.cron_expr = cron;
}

async function saveForm() {
  const invalid = validateForm();
  if (invalid) {
    formError.value = invalid;
    return;
  }
  saving.value = true;
  formError.value = "";
  try {
    const payload: CreateRecurringRuleInput = buildPayload();
    if (editingRule.value) {
      const updatePayload: UpdateRecurringRuleInput = {
        title_template: payload.title_template,
        description_template: payload.description_template,
        cron_expr: payload.cron_expr,
        assignee_ids: payload.assignee_ids,
        label_ids: payload.label_ids,
        state_id: payload.state_id,
        enabled: payload.enabled,
      };
      await recurringApi.update(
        wsId.value,
        projectId.value,
        editingRule.value.id,
        updatePayload,
      );
      toast.success(t("recurring.updateSuccess"));
    } else {
      await recurringApi.create(wsId.value, projectId.value, payload);
      toast.success(t("recurring.createSuccess"));
    }
    closeForm();
    await loadRules();
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : t("recurring.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 列表操作                                                            */
/* ------------------------------------------------------------------ */

async function toggleRule(rule: RecurringRule) {
  const enable = !rule.enabled;
  try {
    const updated = await recurringApi.setEnabled(
      wsId.value,
      projectId.value,
      rule.id,
      enable,
    );
    rules.value = rules.value.map((r) => (r.id === rule.id ? updated : r));
    toast.success(enable ? t("recurring.enableSuccess") : t("recurring.disableSuccess"));
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("recurring.toggleFailed"));
  }
}

async function runNow(rule: RecurringRule) {
  try {
    await recurringApi.runNow(wsId.value, projectId.value, rule.id);
    toast.success(t("recurring.runNowSuccess"));
    await loadRules();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("recurring.runNowFailed"));
  }
}

async function removeRule(rule: RecurringRule) {
  if (!await confirm({ title: "删除规则", message: t("recurring.confirmDelete", { name: previewTitle(rule.title_template) }), confirmText: "删除", variant: "danger" })) {
    return;
  }
  try {
    await recurringApi.remove(wsId.value, projectId.value, rule.id);
    toast.success(t("recurring.deleteSuccess"));
    await loadRules();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("recurring.deleteFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

watch(ready, (v) => {
  if (v) void loadRules();
});

onMounted(() => {
  if (ready.value) void loadRules();
});
</script>

<template>
  
  <AppPageContainer max="980px">
    <div class="recurring">
    <!-- ===== 页头 ===== -->
    <header class="recurring__header">
      <div>
        <h1 class="recurring__title">{{ t("recurring.title") }}</h1>
        <p class="recurring__subtitle">{{ t("recurring.subtitle") }}</p>
      </div>
      <AppButton :disabled="!ready" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" />
        {{ t("recurring.newRule") }}
      </AppButton>
    </header>

    <!-- ===== 上下文未就绪 ===== -->
    <AppLoadingState
      v-if="!ready && !ctxError"
      :text="t('recurring.loadingContext')"
    />
    <AppErrorState
      v-else-if="ctxError || loadError"
      :message="ctxError || loadError"
      :retry="t('common.retry')"
      @retry="loadRules"
    />

    <template v-else>
      <!-- ===== 加载态 ===== -->
      <AppLoadingState v-if="loading" :text="t('recurring.loading')" />

      <!-- ===== 空态 ===== -->
      <AppEmptyState
        v-else-if="!rules.length"
        icon="🔁"
        :title="t('recurring.emptyTitle')"
        :description="t('recurring.emptyDesc')"
        :cta-text="t('recurring.newRule')"
        @cta-click="openCreate"
      />

      <!-- ===== 规则列表 ===== -->
      <div v-else class="recurring__list">
        <article v-for="rule in rules" :key="rule.id" class="recurring__card">
          <div class="recurring__card-main">
            <div class="recurring__card-head">
              <span class="recurring__card-title">
                {{ previewTitle(rule.title_template) }}
              </span>
              <AppBadge :variant="rule.enabled ? 'success' : 'default'">
                {{ rule.enabled ? t("recurring.enabled") : t("recurring.disabled") }}
              </AppBadge>
              <AppBadge variant="brand">{{ typeLabel(rule.work_item_type) }}</AppBadge>
            </div>

            <div class="recurring__card-meta">
              <span class="recurring__cron">
                <DynamicIcon name="Clock" :size="12" />
                <code>{{ rule.cron_expr }}</code>
              </span>
              <span class="recurring__meta-item">
                {{ t("recurring.nextRun") }}：{{ formatTime(rule.next_run_at) }}
              </span>
              <span class="recurring__meta-item">
                {{ t("recurring.lastRun") }}：{{ formatTime(rule.last_run_at) }}
              </span>
            </div>

            <p v-if="rule.description_template" class="recurring__card-desc">
              {{ rule.description_template }}
            </p>

            <div class="recurring__card-tags">
              <span v-if="rule.assignee_ids?.length" class="recurring__tag">
                {{ t("recurring.assignees") }}：{{ rule.assignee_ids.join(", ") }}
              </span>
              <span v-if="rule.label_ids?.length" class="recurring__tag">
                {{ t("recurring.labels") }}：{{ rule.label_ids.join(", ") }}
              </span>
              <span v-if="rule.state_id" class="recurring__tag">
                {{ t("recurring.state") }}：#{{ rule.state_id }}
              </span>
            </div>
          </div>

          <div class="recurring__card-actions">
            <AppButton size="sm" variant="secondary" @click="runNow(rule)">
              <DynamicIcon name="Play" :size="12" />
              {{ t("recurring.runNow") }}
            </AppButton>
            <AppButton size="sm" variant="ghost" @click="toggleRule(rule)">
              {{ rule.enabled ? t("common.disable") : t("common.enable") }}
            </AppButton>
            <AppButton size="sm" variant="ghost" @click="openEdit(rule)">
              <DynamicIcon name="Pencil" :size="12" />
              {{ t("common.edit") }}
            </AppButton>
            <AppButton size="sm" variant="danger" @click="removeRule(rule)">
              <DynamicIcon name="Trash" :size="12" />
              {{ t("common.delete") }}
            </AppButton>
          </div>
        </article>
      </div>
    </template>

    <!-- ===== 新建 / 编辑弹窗 ===== -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('recurring.editRule') : t('recurring.newRule')"
      width="600px"
      @close="closeForm"
    >
      <div class="recurring__form">
        <div class="recurring__field">
          <label class="recurring__label">
            {{ t("recurring.titleTemplate") }} <span class="recurring__required">*</span>
          </label>
          <input
            v-model="form.title_template"
            class="recurring__input"
            :placeholder="t('recurring.titleTemplatePlaceholder')"
            maxlength="200"
          />
          <p class="recurring__hint">{{ t("recurring.titleTemplateHint") }}</p>
        </div>

        <div class="recurring__grid">
          <div class="recurring__field">
            <label class="recurring__label">{{ t("recurring.type") }}</label>
          <select v-model="form.work_item_type" class="recurring__input">
            <option v-for="o in TYPE_OPTIONS" :key="o.value" :value="o.value">
              {{ t(o.labelKey) }}
            </option>
          </select>
          </div>

          <div class="recurring__field">
            <label class="recurring__label">{{ t("recurring.state") }}</label>
            <input
              v-model="form.stateIdText"
              class="recurring__input"
              :placeholder="t('recurring.statePlaceholder')"
              inputmode="numeric"
            />
          </div>
        </div>

        <div class="recurring__field">
          <label class="recurring__label">
            {{ t("recurring.cronExpression") }} <span class="recurring__required">*</span>
          </label>
          <input
            v-model="form.cron_expr"
            class="recurring__input recurring__input--mono"
            placeholder="0 9 * * 1"
            spellcheck="false"
          />
          <p class="recurring__hint">{{ t("recurring.cronHint") }}</p>
          <div class="recurring__presets">
            <button
              v-for="p in CRON_PRESETS"
              :key="p.cron"
              type="button"
              class="recurring__preset"
              @click="applyPreset(p.cron)"
            >
              {{ t(p.labelKey) }}
            </button>
          </div>
        </div>

        <div class="recurring__field">
          <label class="recurring__label">{{ t("recurring.description") }}</label>
          <textarea
            v-model="form.description_template"
            class="recurring__input recurring__textarea"
            :placeholder="t('recurring.descriptionPlaceholder')"
            rows="2"
          />
        </div>

        <div class="recurring__grid">
          <div class="recurring__field">
            <label class="recurring__label">{{ t("recurring.assignees") }}</label>
            <input
              v-model="form.assigneesText"
              class="recurring__input"
              :placeholder="t('recurring.idsPlaceholder')"
            />
          </div>
          <div class="recurring__field">
            <label class="recurring__label">{{ t("recurring.labels") }}</label>
            <input
              v-model="form.labelsText"
              class="recurring__input"
              :placeholder="t('recurring.idsPlaceholder')"
            />
          </div>
        </div>

        <label class="recurring__switch">
          <input v-model="form.enabled" type="checkbox" />
          <span>{{ t("recurring.enabledToggle") }}</span>
        </label>

        <p v-if="formError" class="recurring__error">{{ formError }}</p>
      </div>

      <template #footer>
        <AppButton variant="ghost" :disabled="saving" @click="closeForm">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton :loading="saving" @click="saveForm">
          {{ isEditing ? t("common.save") : t("common.create") }}
        </AppButton>
      </template>
    </AppModal>
    </div>
  </AppPageContainer>
</template>

<style scoped>
.recurring {
  display: flex;
  flex-direction: column;
  gap: 20px;
  
  padding: 24px;
  
}

/* ---- 页头 ---- */
.recurring__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.recurring__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.recurring__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* ---- 列表 ---- */
.recurring__list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recurring__card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
}

.recurring__card-main {
  flex: 1;
  min-width: 0;
}

.recurring__card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.recurring__card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-word;
}

.recurring__card-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.recurring__cron {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.recurring__cron code {
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--surface-3);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
  color: var(--text-secondary);
}

.recurring__card-desc {
  margin: 0 0 6px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.recurring__card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.recurring__tag {
  padding: 1px 8px;
  border-radius: 4px;
  background: var(--surface-2);
  font-size: 11px;
  color: var(--text-tertiary);
}

.recurring__card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  flex-shrink: 0;
}

/* ---- 表单 ---- */
.recurring__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.recurring__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.recurring__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.recurring__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.recurring__required {
  color: var(--danger-500);
}

.recurring__input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}

.recurring__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.recurring__input--mono {
  font-family: var(--font-mono, ui-monospace, monospace);
}

.recurring__textarea {
  resize: vertical;
  line-height: 1.6;
}

.recurring__hint {
  margin: 0;
  font-size: 11px;
  line-height: 1.5;
  color: var(--text-tertiary);
}

.recurring__presets {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
}

.recurring__preset {
  padding: 3px 10px;
  border: 1px solid var(--border-default);
  border-radius: 999px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}

.recurring__preset:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}

.recurring__switch {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

.recurring__switch input {
  accent-color: var(--brand-500);
}

.recurring__error {
  margin: 0;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--danger-50, rgba(239, 68, 68, 0.06));
  color: var(--danger-500);
  font-size: 12px;
}

@media (max-width: 720px) {
  .recurring {
    padding: 16px;
  }

  .recurring__card {
    flex-direction: column;
  }

  .recurring__card-actions {
    width: 100%;
  }

  .recurring__grid {
    grid-template-columns: 1fr;
  }
}
</style>
