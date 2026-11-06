<script setup lang="ts">
/**
 * WorkspaceViewsView — 工作空间级保存视图管理页。
 *
 * 能力：
 *  - 卡片网格展示视图（布局徽标 / 筛选摘要 / 收藏星标 / 公开徽标）
 *  - 新建 / 编辑视图（名称、布局选择、筛选 JSON 编辑器、展示选项 JSON、排序字段与方向）
 *  - 收藏 / 取消收藏
 *  - 复制视图（duplicate）
 *  - 删除视图
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import {
  workspaceViewApi,
  type CreateWorkspaceViewInput,
  type UpdateWorkspaceViewInput,
  type WorkspaceView,
  type WorkspaceViewLayout,
} from "@/api/services/workspaceview";
import { AppBadge, AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer } from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const { t } = useTranslation();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

/** 布局选项（图标名与 DynamicIcon 注册表对齐；layout 值与后端 Layout 常量对齐） */
const LAYOUT_OPTIONS: { value: WorkspaceViewLayout; labelKey: string; icon: string }[] = [
  { value: "board", labelKey: "workspaceViews.layoutBoard", icon: "Kanban" },
  { value: "list", labelKey: "workspaceViews.layoutList", icon: "ListChecks" },
  { value: "gantt", labelKey: "workspaceViews.layoutGantt", icon: "ChartGantt" },
  { value: "calendar", labelKey: "workspaceViews.layoutCalendar", icon: "Calendar" },
  { value: "spreadsheet", labelKey: "workspaceViews.layoutSpreadsheet", icon: "Table2" },
];

/** 可选排序字段 */
const SORT_FIELD_OPTIONS = [
  "created_at",
  "updated_at",
  "target_date",
  "priority",
  "state",
  "name",
  "assignee",
];

function layoutMeta(layout: WorkspaceViewLayout) {
  return (
    LAYOUT_OPTIONS.find((o) => o.value === layout) ?? {
      value: layout,
      labelKey: "workspaceViews.layoutList",
      icon: "LayoutGrid",
    }
  );
}

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const loadError = ref("");
const views = ref<WorkspaceView[]>([]);

async function loadViews() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  loadError.value = "";
  try {
    const res = await workspaceViewApi.list(workspaceId.value);
    views.value = res.results ?? [];
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : t("workspaceViews.loadFailed");
  } finally {
    loading.value = false;
  }
}

onMounted(loadViews);

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

/**
 * 筛选摘要：把 filters JSON 压成可读短句。
 * 例：{ state_group: ["started"], assignee: [12] } → "状态组: started · 指派人: 12"
 */
function filterSummary(view: WorkspaceView): string {
  const filters = view.filters ?? {};
  const parts: string[] = [];
  for (const [key, raw] of Object.entries(filters)) {
    if (raw === undefined || raw === null || raw === "") continue;
    if (Array.isArray(raw)) {
      if (raw.length === 0) continue;
      parts.push(`${key}: ${raw.join(", ")}`);
      continue;
    }
    if (typeof raw === "object") {
      const nested = Object.entries(raw)
        .filter(([, v]) => v !== undefined && v !== null && v !== "")
        .map(([k, v]) => `${k}=${JSON.stringify(v)}`)
        .join(", ");
      if (nested) parts.push(`${key}: ${nested}`);
      continue;
    }
    // 原始类型（string / number / boolean）：JSON.stringify 可安全序列化
    parts.push(`${key}: ${JSON.stringify(raw)}`);
  }
  return parts.length ? parts.join(" · ") : t("workspaceViews.noFilters");
}

/**
 * 排序摘要。
 * 后端 sort 为自由 JSON，常见形态 { field, dir } / { by, order } / { column, asc }，
 * 这里做防御式解析，取不到时回退到键值摘要，避免直接字符串化对象。
 */
function sortSummary(view: WorkspaceView): string {
  const sort = view.sort ?? {};
  const entries = Object.entries(sort).filter(
    ([, v]) => v !== undefined && v !== null && v !== "",
  );
  if (entries.length === 0) return t("workspaceViews.noSort");

  const fieldKeys = ["field", "by", "column", "key", "sort_field"];
  const dirKeys = ["dir", "order", "direction", "asc"];
  const fieldEntry = entries.find(([k]) => fieldKeys.includes(k));
  const dirEntry = entries.find(([k]) => dirKeys.includes(k));

  if (fieldEntry) {
    const dirText = dirEntry
      ? ` ${String(dirEntry[1]).startsWith("a") || dirEntry[1] === true ? "↑" : "↓"}`
      : "";
    return `${String(fieldEntry[1])}${dirText}`;
  }
  return entries.map(([k, v]) => `${k}=${JSON.stringify(v)}`).join(", ");
}

/** 从自由 JSON 的 sort 中提取字段名（兼容 field / by / column / key） */
function sortFieldOf(sort: Record<string, unknown>): string | undefined {
  for (const k of ["field", "by", "column", "key", "sort_field"]) {
    const v = sort[k];
    if (typeof v === "string" && v) return v;
  }
  return undefined;
}

/** 从自由 JSON 的 sort 中提取方向（兼容 dir / order / direction / asc 布尔） */
function sortDirOf(sort: Record<string, unknown>): "asc" | "desc" | undefined {
  const raw = sort.dir ?? sort.order ?? sort.direction;
  if (typeof raw === "string") {
    return raw.toLowerCase().startsWith("a") ? "asc" : "desc";
  }
  if (typeof sort.asc === "boolean") return sort.asc ? "asc" : "desc";
  return undefined;
}

/* ------------------------------------------------------------------ */
/* 表单状态                                                            */
/* ------------------------------------------------------------------ */

interface ViewForm {
  name: string;
  description: string;
  layout: WorkspaceViewLayout;
  /** 筛选条件 JSON 文本 */
  filtersText: string;
  /** 展示选项 JSON 文本（后端字段名 display_options） */
  displayOptionsText: string;
  sortField: string;
  sortDir: "asc" | "desc";
  is_public: boolean;
}

function prettyJson(value: Record<string, unknown>): string {
  return JSON.stringify(value ?? {}, null, 2);
}

function emptyForm(): ViewForm {
  return {
    name: "",
    description: "",
    layout: "list",
    filtersText: "{}",
    displayOptionsText: "{}",
    sortField: "updated_at",
    sortDir: "desc",
    is_public: false,
  };
}

const showForm = ref(false);
const editingView = ref<WorkspaceView | null>(null);
const saving = ref(false);
const formError = ref("");
const form = reactive<ViewForm>(emptyForm());

const isEditing = computed(() => editingView.value !== null);

/** 解析 JSON 文本；非法时返回错误信息 */
function parseJsonField(text: string, label: string): Record<string, unknown> | null {
  const trimmed = text.trim();
  if (!trimmed) return {};
  try {
    const parsed = JSON.parse(trimmed);
    if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
      formError.value = t("workspaceViews.jsonMustBeObject", { field: label });
      return null;
    }
    return parsed as Record<string, unknown>;
  } catch {
    formError.value = t("workspaceViews.invalidJson", { field: label });
    return null;
  }
}

function openCreate() {
  editingView.value = null;
  Object.assign(form, emptyForm());
  formError.value = "";
  showForm.value = true;
}

function openEdit(view: WorkspaceView) {
  editingView.value = view;
  Object.assign(form, {
    name: view.name,
    description: view.description ?? "",
    layout: view.layout,
    filtersText: prettyJson(view.filters ?? {}),
    displayOptionsText: prettyJson(view.display_options ?? {}),
    sortField: sortFieldOf(view.sort) ?? "updated_at",
    sortDir: sortDirOf(view.sort) === "asc" ? "asc" : "desc",
    is_public: view.is_public,
  });
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingView.value = null;
  formError.value = "";
}

async function saveForm() {
  formError.value = "";
  if (!form.name.trim()) {
    formError.value = t("workspaceViews.nameRequired");
    return;
  }
  const filters = parseJsonField(form.filtersText, t("workspaceViews.filters"));
  if (filters === null) return;
  const displayOptions = parseJsonField(
    form.displayOptionsText,
    t("workspaceViews.display"),
  );
  if (displayOptions === null) return;

  saving.value = true;
  try {
    const payload: CreateWorkspaceViewInput = {
      name: form.name.trim(),
      description: form.description.trim() || undefined,
      layout: form.layout,
      filters,
      display_options: displayOptions,
      sort: { field: form.sortField, dir: form.sortDir },
      is_public: form.is_public,
    };
    if (editingView.value) {
      const updatePayload: UpdateWorkspaceViewInput = { ...payload };
      await workspaceViewApi.update(
        workspaceId.value,
        editingView.value.id,
        updatePayload,
      );
      toast.success(t("workspaceViews.updateSuccess"));
    } else {
      await workspaceViewApi.create(workspaceId.value, payload);
      toast.success(t("workspaceViews.createSuccess"));
    }
    closeForm();
    await loadViews();
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : t("workspaceViews.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 列表操作                                                            */
/* ------------------------------------------------------------------ */

async function toggleFavorite(view: WorkspaceView) {
  try {
    // 后端为 Toggle 语义（不带请求体），返回切换后的 is_favorite
    const res = await workspaceViewApi.toggleFavorite(workspaceId.value, view.id);
    views.value = views.value.map((v) =>
      v.id === view.id ? { ...v, is_favorite: res.is_favorite } : v,
    );
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workspaceViews.favoriteFailed"));
  }
}

async function duplicateView(view: WorkspaceView) {
  try {
    await workspaceViewApi.duplicate(workspaceId.value, view.id);
    toast.success(t("workspaceViews.duplicateSuccess"));
    await loadViews();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workspaceViews.duplicateFailed"));
  }
}

async function removeView(view: WorkspaceView) {
  if (!await confirm({ title: "删除视图", message: t("workspaceViews.confirmDelete", { name: view.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await workspaceViewApi.remove(workspaceId.value, view.id);
    toast.success(t("workspaceViews.deleteSuccess"));
    await loadViews();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("workspaceViews.deleteFailed"));
  }
}
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="wviews">
    <!-- ===== 页头 ===== -->
    <header class="wviews__header">
      <div>
        <h1 class="wviews__title">{{ t("workspaceViews.title") }}</h1>
        <p class="wviews__subtitle">{{ t("workspaceViews.subtitle") }}</p>
      </div>
      <AppButton :disabled="!workspaceId" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" />
        {{ t("workspaceViews.newView") }}
      </AppButton>
    </header>

    <!-- ===== 错误态 ===== -->
    <AppErrorState
      v-if="loadError"
      :message="loadError"
      :retry="t('common.retry')"
      @retry="loadViews"
    />

    <!-- ===== 加载态 ===== -->
    <AppLoadingState v-else-if="loading" :text="t('workspaceViews.loading')" />

    <!-- ===== 空态 ===== -->
    <AppEmptyState
      v-else-if="!views.length"
      scenario="views"
      :title="t('workspaceViews.emptyTitle')"
      :description="t('workspaceViews.emptyDesc')"
      :cta-text="t('workspaceViews.newView')"
      @cta-click="openCreate"
    />

    <!-- ===== 卡片网格 ===== -->
    <div v-else class="wviews__grid">
      <article v-for="view in views" :key="view.id" class="wviews__card">
        <div class="wviews__card-head">
          <span class="wviews__layout">
            <DynamicIcon :name="layoutMeta(view.layout).icon" :size="14" />
            {{ t(layoutMeta(view.layout).labelKey) }}
          </span>
          <div class="wviews__badges">
            <AppBadge v-if="view.is_public" variant="info">
              <DynamicIcon name="Globe" :size="10" />
              {{ t("workspaceViews.public") }}
            </AppBadge>
            <button
              class="wviews__star"
              :class="{ 'wviews__star--active': view.is_favorite }"
              :title="view.is_favorite ? t('workspaceViews.unfavorite') : t('workspaceViews.favorite')"
              @click="toggleFavorite(view)"
            >
              <DynamicIcon name="Star" :size="14" />
            </button>
          </div>
        </div>

        <h3 class="wviews__name">{{ view.name }}</h3>
        <p v-if="view.description" class="wviews__desc">{{ view.description }}</p>

        <div class="wviews__meta">
          <p class="wviews__meta-row">
            <span class="wviews__meta-label">{{ t("workspaceViews.filters") }}</span>
            <span class="wviews__meta-value">{{ filterSummary(view) }}</span>
          </p>
          <p class="wviews__meta-row">
            <span class="wviews__meta-label">{{ t("workspaceViews.sort") }}</span>
            <span class="wviews__meta-value">{{ sortSummary(view) }}</span>
          </p>
        </div>

        <div class="wviews__card-actions">
          <AppButton size="sm" variant="secondary" @click="openEdit(view)">
            <DynamicIcon name="Pencil" :size="12" />
            {{ t("common.edit") }}
          </AppButton>
          <AppButton size="sm" variant="ghost" @click="duplicateView(view)">
            <DynamicIcon name="Copy" :size="12" />
            {{ t("workspaceViews.duplicate") }}
          </AppButton>
          <AppButton size="sm" variant="danger" @click="removeView(view)">
            <DynamicIcon name="Trash" :size="12" />
            {{ t("common.delete") }}
          </AppButton>
        </div>
      </article>
    </div>

    <!-- ===== 新建 / 编辑弹窗 ===== -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('workspaceViews.editView') : t('workspaceViews.newView')"
      width="640px"
      @close="closeForm"
    >
      <div class="wviews__form">
        <div class="wviews__field">
          <label class="wviews__label">
            {{ t("workspaceViews.name") }} <span class="wviews__required">*</span>
          </label>
          <input
            v-model="form.name"
            class="wviews__input"
            :placeholder="t('workspaceViews.namePlaceholder')"
            maxlength="120"
          />
        </div>

        <div class="wviews__field">
          <label class="wviews__label">{{ t("workspaceViews.description") }}</label>
          <input
            v-model="form.description"
            class="wviews__input"
            :placeholder="t('workspaceViews.descriptionPlaceholder')"
            maxlength="240"
          />
        </div>

        <div class="wviews__field">
          <label class="wviews__label">{{ t("workspaceViews.layout") }}</label>
          <div class="wviews__layouts">
            <button
              v-for="o in LAYOUT_OPTIONS"
              :key="o.value"
              type="button"
              class="wviews__layout-option"
              :class="{ 'wviews__layout-option--active': form.layout === o.value }"
              @click="form.layout = o.value"
            >
              <DynamicIcon :name="o.icon" :size="16" />
              <span>{{ t(o.labelKey) }}</span>
            </button>
          </div>
        </div>

        <div class="wviews__grid">
          <div class="wviews__field">
            <label class="wviews__label">{{ t("workspaceViews.filters") }}</label>
            <textarea
              v-model="form.filtersText"
              class="wviews__input wviews__textarea wviews__textarea--mono"
              rows="6"
              spellcheck="false"
              :placeholder="t('workspaceViews.filtersPlaceholder')"
            />
          </div>
          <div class="wviews__field">
            <label class="wviews__label">{{ t("workspaceViews.display") }}</label>
            <textarea
              v-model="form.displayOptionsText"
              class="wviews__input wviews__textarea wviews__textarea--mono"
              rows="6"
              spellcheck="false"
              :placeholder="t('workspaceViews.displayPlaceholder')"
            />
          </div>
        </div>

        <div class="wviews__grid">
          <div class="wviews__field">
            <label class="wviews__label">{{ t("workspaceViews.sortField") }}</label>
            <select v-model="form.sortField" class="wviews__input">
              <option v-for="f in SORT_FIELD_OPTIONS" :key="f" :value="f">{{ f }}</option>
            </select>
          </div>
          <div class="wviews__field">
            <label class="wviews__label">{{ t("workspaceViews.sortDir") }}</label>
            <select v-model="form.sortDir" class="wviews__input">
              <option value="asc">asc ↑</option>
              <option value="desc">desc ↓</option>
            </select>
          </div>
        </div>

        <label class="wviews__switch">
          <input v-model="form.is_public" type="checkbox" />
          <span>{{ t("workspaceViews.isPublic") }}</span>
        </label>

        <p v-if="formError" class="wviews__error">{{ formError }}</p>
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
    </AppModal>
  </div>


<style scoped>
.wviews {
  display: flex;
  flex-direction: column;
  gap: 20px;
  
  padding: 24px;
  
}

/* ---- 页头 ---- */
.wviews__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.wviews__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.wviews__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* ---- 网格 ---- */
.wviews__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.wviews__card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
}

.wviews__card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.wviews__layout {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--surface-2);
  font-size: 11px;
  color: var(--text-secondary);
}

.wviews__badges {
  display: flex;
  align-items: center;
  gap: 6px;
}

.wviews__star {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all 0.15s;
}

.wviews__star:hover {
  background: var(--surface-2);
  color: var(--warning-500);
}

.wviews__star--active {
  color: var(--warning-500);
}

.wviews__name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-word;
}

.wviews__desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.wviews__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--surface-2);
}

.wviews__meta-row {
  display: flex;
  gap: 8px;
  margin: 0;
  font-size: 11px;
  line-height: 1.5;
}

.wviews__meta-label {
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.wviews__meta-value {
  flex: 1;
  min-width: 0;
  color: var(--text-secondary);
  word-break: break-word;
}

.wviews__card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding-top: 4px;
  border-top: 1px dashed var(--border-subtle);
}

/* ---- 表单 ---- */
.wviews__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.wviews__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.wviews__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.wviews__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.wviews__required {
  color: var(--danger-500);
}

.wviews__input {
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

.wviews__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.wviews__textarea {
  resize: vertical;
  line-height: 1.6;
}

.wviews__textarea--mono {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
}

.wviews__layouts {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(108px, 1fr));
  gap: 8px;
}

.wviews__layout-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 10px 6px;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}

.wviews__layout-option:hover {
  border-color: var(--brand-300);
}

.wviews__layout-option--active {
  border-color: var(--brand-500);
  background: var(--brand-50, rgba(59, 130, 246, 0.08));
  color: var(--brand-600);
}

.wviews__switch {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}

.wviews__switch input {
  accent-color: var(--brand-500);
}

.wviews__error {
  margin: 0;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--danger-50, rgba(239, 68, 68, 0.06));
  color: var(--danger-500);
  font-size: 12px;
}

@media (max-width: 720px) {
  .wviews {
    padding: 16px;
  }

  .wviews__grid {
    grid-template-columns: 1fr;
  }
}
</style>
