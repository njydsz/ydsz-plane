<script setup lang="ts">
/**
 * MilestonesView — 项目级里程碑管理页。
 *
 * 能力：
 *  - 时间线视图：按目标日期排序的里程碑节点 + 进度条 + 关联工作项数
 *  - 新建 / 编辑里程碑（名称、描述、目标日期、状态）
 *  - 关联 / 解除工作项（按标识符搜索并挂载到里程碑）
 *  - 删除里程碑
 */
import { computed, onMounted, reactive, ref, watch } from "vue";

import {
  milestoneApi,
  type CreateMilestoneInput,
  type Milestone,
  type MilestoneProgress,
  type MilestoneStatus,
  type MilestoneWorkItem,
  type UpdateMilestoneInput,
} from "@/api/services/milestone";
import { issueApi, type Issue } from "@/api/services/issue";
import { AppBadge, AppButton, AppEmptyState, AppErrorState, AppLoadingState, AppModal, AppPageContainer, ProgressBar } from "@/components";
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

const STATUS_OPTIONS: { value: MilestoneStatus; labelKey: string; variant: "default" | "brand" | "success" | "danger" }[] = [
  { value: "not_started", labelKey: "milestones.statusNotStarted", variant: "default" },
  { value: "in_progress", labelKey: "milestones.statusInProgress", variant: "brand" },
  { value: "completed", labelKey: "milestones.statusCompleted", variant: "success" },
  { value: "missed", labelKey: "milestones.statusMissed", variant: "danger" },
];

function statusMeta(status: MilestoneStatus) {
  return (
    STATUS_OPTIONS.find((o) => o.value === status) ?? {
      value: "not_started" as const,
      labelKey: "milestones.statusNotStarted",
      variant: "default" as const,
    }
  );
}

/** 进度条颜色：完成 → 绿， missed → 红，进行中 → 品牌色，未开始 → 中性 */
function progressColor(m: Milestone): string {
  if (m.status === "completed") return "var(--success-500)";
  if (m.status === "missed") return "var(--danger-500)";
  if (m.status === "in_progress") return "var(--brand-500)";
  return "var(--neutral-400)";
}

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const loadError = ref("");
const milestones = ref<Milestone[]>([]);
/** 里程碑 ID -> 进度聚合（列表接口不返回进度，需逐条拉取） */
const progressMap = ref<Record<number, MilestoneProgress>>({});

/** 按目标日期升序（无日期的排最后） */
const sortedMilestones = computed(() =>
  [...milestones.value].sort((a, b) => {
    if (!a.target_date) return 1;
    if (!b.target_date) return -1;
    return a.target_date.localeCompare(b.target_date);
  }),
);

async function loadMilestones() {
  if (!ready.value) return;
  loading.value = true;
  loadError.value = "";
  try {
    const res = await milestoneApi.list(wsId.value, projectId.value);
    milestones.value = res.results ?? [];
    // 逐条拉取进度（列表接口不内嵌进度），失败不影响主列表展示
    await Promise.all(
      milestones.value.map(async (m) => {
        try {
          const p = await milestoneApi.getProgress(wsId.value, projectId.value, m.id);
          progressMap.value = { ...progressMap.value, [m.id]: p };
        } catch {
          /* 忽略单条进度失败 */
        }
      }),
    );
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : t("milestones.loadFailed");
  } finally {
    loading.value = false;
  }
}

/** 取某里程碑的进度（缺省 0） */
function progressOf(m: Milestone): MilestoneProgress {
  return (
    progressMap.value[m.id] ?? {
      total: 0,
      completed: 0,
      in_progress: 0,
      not_started: 0,
      percent: 0,
      computed_at: "",
    }
  );
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function formatDate(date?: string | null): string {
  return date || "-";
}

/** 是否已过期（目标日期早于今天且未完成） */
function isOverdue(m: Milestone): boolean {
  if (!m.target_date) return false;
  const today = new Date().toISOString().slice(0, 10);
  return m.target_date < today && m.status !== "completed";
}

/* ------------------------------------------------------------------ */
/* 表单状态                                                            */
/* ------------------------------------------------------------------ */

interface MilestoneForm {
  name: string;
  description: string;
  target_date: string;
  status: MilestoneStatus;
}

function emptyForm(): MilestoneForm {
  return {
    name: "",
    description: "",
    target_date: "",
    status: "not_started",
  };
}

const showForm = ref(false);
const editingMilestone = ref<Milestone | null>(null);
const saving = ref(false);
const formError = ref("");
const form = reactive<MilestoneForm>(emptyForm());

const isEditing = computed(() => editingMilestone.value !== null);

function openCreate() {
  editingMilestone.value = null;
  Object.assign(form, emptyForm());
  formError.value = "";
  showForm.value = true;
}

function openEdit(m: Milestone) {
  editingMilestone.value = m;
  Object.assign(form, {
    name: m.name,
    description: m.description ?? "",
    target_date: m.target_date ?? "",
    status: m.status,
  });
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingMilestone.value = null;
  formError.value = "";
}

async function saveForm() {
  if (!form.name.trim()) {
    formError.value = t("milestones.nameRequired");
    return;
  }
  saving.value = true;
  formError.value = "";
  try {
    const payload: CreateMilestoneInput = {
      name: form.name.trim(),
      description: form.description.trim() || undefined,
      target_date: form.target_date || null,
      status: form.status,
    };
    if (editingMilestone.value) {
      const updatePayload: UpdateMilestoneInput = { ...payload };
      await milestoneApi.update(
        wsId.value,
        projectId.value,
        editingMilestone.value.id,
        updatePayload,
      );
      toast.success(t("milestones.updateSuccess"));
    } else {
      await milestoneApi.create(wsId.value, projectId.value, payload);
      toast.success(t("milestones.createSuccess"));
    }
    closeForm();
    await loadMilestones();
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : t("milestones.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 关联工作项                                                          */
/* ------------------------------------------------------------------ */

/** 已展开「关联工作项」面板的里程碑 ID */
const showLinkPanelFor = ref<number | null>(null);
/** 搜索关键字 */
const linkSearch = ref("");
const linkSearching = ref(false);
const linkResults = ref<Issue[]>([]);
/** 里程碑 ID -> 已关联工作项（展示用，含标识符与状态） */
const linkedIssuesMap = ref<Record<number, MilestoneWorkItem[]>>({});

let linkSearchTimer: ReturnType<typeof setTimeout> | null = null;

function toggleLinkPanel(m: Milestone) {
  showLinkPanelFor.value = showLinkPanelFor.value === m.id ? null : m.id;
  linkSearch.value = "";
  linkResults.value = [];
  if (showLinkPanelFor.value !== null) {
    void loadLinkedIssues(m);
  }
}

/** 拉取该里程碑已关联的工作项（后端返回轻量列表，含 ID / 类型） */
async function loadLinkedIssues(m: Milestone) {
  try {
    const res = await milestoneApi.listWorkItems(wsId.value, projectId.value, m.id);
    linkedIssuesMap.value = {
      ...linkedIssuesMap.value,
      [m.id]: res.results ?? [],
    };
  } catch {
    linkedIssuesMap.value = { ...linkedIssuesMap.value, [m.id]: [] };
  }
}

/** 搜索可关联的工作项（按标识符 / 名称模糊匹配，本地过滤，防抖） */
function onLinkSearchInput() {
  if (linkSearchTimer) clearTimeout(linkSearchTimer);
  linkSearchTimer = setTimeout(() => void runLinkSearch(), 300);
}

async function runLinkSearch() {
  const keyword = linkSearch.value.trim();
  if (!keyword) {
    linkResults.value = [];
    return;
  }
  linkSearching.value = true;
  try {
    const res = await issueApi.listIssues(wsId.value, projectId.value, {
      search: keyword,
      limit: 20,
    });
    linkResults.value = res.results ?? [];
  } catch {
    linkResults.value = [];
  } finally {
    linkSearching.value = false;
  }
}

async function linkIssue(m: Milestone, issue: Issue) {
  try {
    await milestoneApi.linkWorkItem(wsId.value, projectId.value, m.id, {
      work_item_id: issue.id,
      work_item_type: issue.type_code,
    });
    toast.success(t("milestones.linkSuccess", { identifier: issue.identifier }));
    linkResults.value = linkResults.value.filter((i) => i.id !== issue.id);
    await loadLinkedIssues(m);
    await loadMilestones();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("milestones.linkFailed"));
  }
}

async function unlinkIssue(m: Milestone, item: MilestoneWorkItem) {
  try {
    await milestoneApi.unlinkWorkItem(
      wsId.value,
      projectId.value,
      m.id,
      item.work_item_id,
    );
    toast.success(t("milestones.unlinkSuccess", { identifier: String(item.work_item_id) }));
    await loadLinkedIssues(m);
    await loadMilestones();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("milestones.unlinkFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 删除                                                                */
/* ------------------------------------------------------------------ */

async function removeMilestone(m: Milestone) {
  if (!await confirm({ title: "删除里程碑", message: t("milestones.confirmDelete", { name: m.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await milestoneApi.remove(wsId.value, projectId.value, m.id);
    toast.success(t("milestones.deleteSuccess"));
    await loadMilestones();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("milestones.deleteFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

watch(ready, (v) => {
  if (v) void loadMilestones();
});

onMounted(() => {
  if (ready.value) void loadMilestones();
});
</script>

<template>
  
  <AppPageContainer max="980px">
    <div class="milestones">
    <!-- ===== 页头 ===== -->
    <header class="milestones__header">
      <div>
        <h1 class="milestones__title">{{ t("milestones.title") }}</h1>
        <p class="milestones__subtitle">{{ t("milestones.subtitle") }}</p>
      </div>
      <AppButton :disabled="!ready" @click="openCreate">
        <DynamicIcon name="Plus" :size="14" />
        {{ t("milestones.newMilestone") }}
      </AppButton>
    </header>

    <!-- ===== 上下文未就绪 ===== -->
    <AppLoadingState
      v-if="!ready && !ctxError"
      :text="t('milestones.loadingContext')"
    />
    <AppErrorState
      v-else-if="ctxError || loadError"
      :message="ctxError || loadError"
      :retry="t('common.retry')"
      @retry="loadMilestones"
    />

    <template v-else>
      <!-- ===== 加载态 ===== -->
      <AppLoadingState v-if="loading" :text="t('milestones.loading')" />

      <!-- ===== 空态 ===== -->
      <AppEmptyState
        v-else-if="!milestones.length"
        icon="🚩"
        :title="t('milestones.emptyTitle')"
        :description="t('milestones.emptyDesc')"
        :cta-text="t('milestones.newMilestone')"
        @cta-click="openCreate"
      />

      <!-- ===== 时间线 ===== -->
      <ol v-else class="milestones__timeline">
        <li
          v-for="m in sortedMilestones"
          :key="m.id"
          class="milestones__node"
          :class="{ 'milestones__node--overdue': isOverdue(m) }"
        >
          <span class="milestones__dot" :style="{ background: progressColor(m) }" />

          <div class="milestones__card">
            <div class="milestones__card-head">
              <div class="milestones__card-title">
                <span class="milestones__name">{{ m.name }}</span>
                <AppBadge :variant="statusMeta(m.status).variant">
                  {{ t(statusMeta(m.status).labelKey) }}
                </AppBadge>
                <AppBadge v-if="isOverdue(m)" variant="danger">
                  {{ t("milestones.overdue") }}
                </AppBadge>
              </div>
              <time class="milestones__date">
                <DynamicIcon name="CalendarPlus" :size="12" />
                {{ formatDate(m.target_date) }}
              </time>
            </div>

            <p v-if="m.description" class="milestones__desc">{{ m.description }}</p>

            <div class="milestones__progress">
              <ProgressBar
                :percent="progressOf(m).percent"
                size="sm"
                :color="progressColor(m)"
                :label="t('milestones.progressLabel', {
                  done: progressOf(m).completed,
                  total: progressOf(m).total,
                })"
              />
            </div>

            <div class="milestones__card-actions">
              <AppButton size="sm" variant="ghost" @click="toggleLinkPanel(m)">
                <DynamicIcon name="Link" :size="12" />
                {{ t("milestones.linkIssues") }}
                <span class="milestones__count">{{ progressOf(m).total }}</span>
              </AppButton>
              <AppButton size="sm" variant="ghost" @click="openEdit(m)">
                <DynamicIcon name="Pencil" :size="12" />
                {{ t("common.edit") }}
              </AppButton>
              <AppButton size="sm" variant="danger" @click="removeMilestone(m)">
                <DynamicIcon name="Trash" :size="12" />
                {{ t("common.delete") }}
              </AppButton>
            </div>

            <!-- ===== 关联面板 ===== -->
            <div v-if="showLinkPanelFor === m.id" class="milestones__link-panel">
              <div class="milestones__link-search">
                <input
                  v-model="linkSearch"
                  class="milestones__input"
                  :placeholder="t('milestones.linkSearchPlaceholder')"
                  @input="onLinkSearchInput"
                />
                <span v-if="linkSearching" class="milestones__link-hint">
                  {{ t("milestones.searching") }}
                </span>
              </div>

              <ul v-if="linkResults.length" class="milestones__link-results">
                <li
                  v-for="issue in linkResults"
                  :key="issue.id"
                  class="milestones__link-result"
                >
                  <span class="milestones__issue-id">{{ issue.identifier }}</span>
                  <span class="milestones__issue-name">{{ issue.name }}</span>
                  <AppButton size="sm" variant="secondary" @click="linkIssue(m, issue)">
                    {{ t("milestones.link") }}
                  </AppButton>
                </li>
              </ul>
              <p v-else-if="linkSearch && !linkSearching" class="milestones__link-hint">
                {{ t("milestones.noMatchingIssues") }}
              </p>

              <div class="milestones__linked">
                <p class="milestones__linked-title">{{ t("milestones.linkedIssues") }}</p>
                <p
                  v-if="!linkedIssuesMap[m.id] || !linkedIssuesMap[m.id].length"
                  class="milestones__link-hint"
                >
                  {{ t("milestones.noLinkedIssues") }}
                </p>
                <ul v-else class="milestones__linked-list">
                  <li
                    v-for="item in linkedIssuesMap[m.id]"
                    :key="item.work_item_id"
                    class="milestones__linked-item"
                  >
                    <span class="milestones__issue-id">{{ item.work_item_id }}</span>
                    <span class="milestones__issue-name">{{ item.work_item_type }}</span>
                    <button
                      class="milestones__unlink"
                      :title="t('milestones.unlink')"
                      @click="unlinkIssue(m, item)"
                    >
                      <DynamicIcon name="Unlink" :size="12" />
                    </button>
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </li>
      </ol>
    </template>

    <!-- ===== 新建 / 编辑弹窗 ===== -->
    <AppModal
      :visible="showForm"
      :title="isEditing ? t('milestones.editMilestone') : t('milestones.newMilestone')"
      width="520px"
      @close="closeForm"
    >
      <div class="milestones__form">
        <div class="milestones__field">
          <label class="milestones__label">
            {{ t("milestones.name") }} <span class="milestones__required">*</span>
          </label>
          <input
            v-model="form.name"
            class="milestones__input"
            :placeholder="t('milestones.namePlaceholder')"
            maxlength="120"
          />
        </div>

        <div class="milestones__field">
          <label class="milestones__label">{{ t("milestones.description") }}</label>
          <textarea
            v-model="form.description"
            class="milestones__input milestones__textarea"
            :placeholder="t('milestones.descriptionPlaceholder')"
            rows="3"
          />
        </div>

        <div class="milestones__grid">
          <div class="milestones__field">
            <label class="milestones__label">{{ t("milestones.targetDate") }}</label>
            <input v-model="form.target_date" type="date" class="milestones__input" />
          </div>
          <div class="milestones__field">
            <label class="milestones__label">{{ t("milestones.status") }}</label>
            <select v-model="form.status" class="milestones__input">
              <option v-for="o in STATUS_OPTIONS" :key="o.value" :value="o.value">
                {{ t(o.labelKey) }}
              </option>
            </select>
          </div>
        </div>

        <p v-if="formError" class="milestones__error">{{ formError }}</p>
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
.milestones {
  display: flex;
  flex-direction: column;
  gap: 20px;
  
  padding: 24px;
  
}

/* ---- 页头 ---- */
.milestones__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.milestones__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.milestones__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

/* ---- 时间线 ---- */
.milestones__timeline {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin: 0;
  padding: 0 0 0 8px;
  list-style: none;
}

.milestones__node {
  position: relative;
  padding-left: 28px;
}

/* 竖线 */
.milestones__node::before {
  content: "";
  position: absolute;
  left: 8px;
  top: 20px;
  bottom: -16px;
  width: 2px;
  background: var(--border-subtle);
}

.milestones__node:last-child::before {
  display: none;
}

.milestones__dot {
  position: absolute;
  left: 0;
  top: 16px;
  width: 18px;
  height: 18px;
  border: 3px solid var(--surface-1);
  border-radius: 50%;
  box-shadow: 0 0 0 2px var(--border-subtle);
}

.milestones__node--overdue .milestones__dot {
  box-shadow: 0 0 0 2px var(--danger-500);
}

/* ---- 卡片 ---- */
.milestones__card {
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
}

.milestones__card-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}

.milestones__card-title {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.milestones__name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.milestones__date {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.milestones__desc {
  margin: 0 0 10px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.milestones__progress {
  margin-bottom: 10px;
}

.milestones__card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.milestones__count {
  padding: 0 5px;
  border-radius: 999px;
  background: var(--surface-3);
  font-size: 11px;
}

/* ---- 关联面板 ---- */
.milestones__link-panel {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--border-subtle);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.milestones__link-search {
  display: flex;
  align-items: center;
  gap: 8px;
}

.milestones__input {
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

.milestones__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.milestones__link-hint {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
}

.milestones__link-results,
.milestones__linked-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.milestones__link-result,
.milestones__linked-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--surface-2);
}

.milestones__linked-item--done .milestones__issue-name {
  text-decoration: line-through;
  color: var(--text-tertiary);
}

.milestones__issue-id {
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--surface-3);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.milestones__issue-name {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.milestones__unlink {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
}

.milestones__unlink:hover {
  background: var(--danger-50, rgba(239, 68, 68, 0.08));
  color: var(--danger-500);
}

.milestones__linked-title {
  margin: 0 0 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
}

/* ---- 表单 ---- */
.milestones__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.milestones__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.milestones__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.milestones__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.milestones__required {
  color: var(--danger-500);
}

.milestones__textarea {
  resize: vertical;
  line-height: 1.6;
}

.milestones__error {
  margin: 0;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--danger-50, rgba(239, 68, 68, 0.06));
  color: var(--danger-500);
  font-size: 12px;
}

@media (max-width: 720px) {
  .milestones {
    padding: 16px;
  }

  .milestones__grid {
    grid-template-columns: 1fr;
  }
}
</style>
