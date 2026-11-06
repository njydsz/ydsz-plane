<script setup lang="ts">
/**
 * 工作空间级「倡议」列表页 — 展示倡议卡片，支持筛选、创建/编辑、归档/恢复。
 *
 * 倡议（Initiative）是工作空间级战略聚合：可关联多个项目与工作项，
 * 并通过 progress 汇总整体进度。点击卡片进入倡议详情页。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  initiativeApi,
  type Initiative,
  type InitiativePriority,
  type InitiativeStatus,
} from "@/api/services/initiative";
import { workspaceApi, type Member, type Project } from "@/api/services/workspace";
import { useI18n } from "@/composables/useI18n";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";
import { AppBadge, AppButton, AppEmptyState, AppErrorState, AppModal, AppPageContainer, AppSkeleton, ProgressBar } from "@/components";

const { t } = useI18n();
const confirm = useConfirm();
const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ---------- 列表状态 ---------- */

const initiatives = ref<Initiative[]>([]);
const loading = ref(true);
const error = ref("");
const total = ref(0);

const filterStatus = ref<InitiativeStatus | "">("");
const filterPriority = ref<InitiativePriority | "">("");

const STATUS_GROUPS: InitiativeStatus[] = ["on_track", "at_risk", "off_track", "completed"];
const PRIORITY_GROUPS: InitiativePriority[] = ["urgent", "high", "medium", "low", "none"];

const statusLabel: Record<InitiativeStatus, string> = {
  on_track: t("initiative.statusLabel.on_track"),
  at_risk: t("initiative.statusLabel.at_risk"),
  off_track: t("initiative.statusLabel.off_track"),
  completed: t("initiative.statusLabel.completed"),
};

const statusBadgeVariant: Record<InitiativeStatus, "success" | "warning" | "danger" | "info"> = {
  on_track: "success",
  at_risk: "warning",
  off_track: "danger",
  completed: "info",
};

const priorityLabel: Record<InitiativePriority, string> = {
  urgent: t("initiative.priorityLabel.urgent"),
  high: t("initiative.priorityLabel.high"),
  medium: t("initiative.priorityLabel.medium"),
  low: t("initiative.priorityLabel.low"),
  none: t("initiative.priorityLabel.none"),
};

const priorityBadgeVariant: Record<InitiativePriority, "danger" | "warning" | "info" | "default"> = {
  urgent: "danger",
  high: "warning",
  medium: "info",
  low: "default",
  none: "default",
};

/* ---------- 关联数据（负责人 / 项目） ---------- */

const members = ref<Member[]>([]);
const projects = ref<Project[]>([]);

function ownerName(ownerId: number): string {
  const m = members.value.find((x) => x.id === ownerId);
  return m ? m.display_name || m.email : `#${ownerId}`;
}

/* ---------- 创建 / 编辑 Modal ---------- */

const showForm = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const formError = ref("");

interface InitiativeForm {
  title: string;
  description: string;
  status: InitiativeStatus;
  priority: InitiativePriority;
  start_date: string;
  target_date: string;
  owner_id: number | null;
  project_ids: number[];
}

const emptyForm = (): InitiativeForm => ({
  title: "",
  description: "",
  status: "on_track",
  priority: "medium",
  start_date: "",
  target_date: "",
  owner_id: null,
  project_ids: [],
});

const form = ref<InitiativeForm>(emptyForm());

function openCreate() {
  editingId.value = null;
  form.value = emptyForm();
  formError.value = "";
  showForm.value = true;
}

function openEdit(item: Initiative) {
  editingId.value = item.id;
  form.value = {
    title: item.title,
    description: item.description ?? "",
    status: item.status,
    priority: item.priority,
    start_date: item.start_date ?? "",
    target_date: item.target_date ?? "",
    owner_id: item.owner_id ?? null,
    project_ids: [],
  };
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingId.value = null;
  formError.value = "";
}

async function submitForm() {
  if (!form.value.title.trim()) {
    formError.value = t("validation.required", { field: t("initiative.titleField") });
    return;
  }
  saving.value = true;
  formError.value = "";
  try {
    const payload = {
      title: form.value.title.trim(),
      description: form.value.description || undefined,
      status: form.value.status,
      priority: form.value.priority,
      start_date: form.value.start_date || undefined,
      target_date: form.value.target_date || undefined,
      owner_id: form.value.owner_id ?? undefined,
      project_ids: form.value.project_ids.length > 0 ? form.value.project_ids : undefined,
    };
    if (editingId.value) {
      await initiativeApi.update(workspaceId.value, editingId.value, payload);
      toast.success(t("initiative.updateSuccess"));
    } else {
      await initiativeApi.create(workspaceId.value, payload);
      toast.success(t("initiative.createSuccess"));
    }
    showForm.value = false;
    editingId.value = null;
    await load();
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    saving.value = false;
  }
}

/* ---------- 归档 / 恢复 ---------- */

const archivingId = ref<number | null>(null);

async function archiveItem(item: Initiative) {
  if (!await confirm({ title: "归档倡议", message: t("initiative.archiveConfirm", { title: item.title }), confirmText: "归档", variant: "danger" })) return;
  archivingId.value = item.id;
  try {
    await initiativeApi.archive(workspaceId.value, item.id);
    toast.success(t("initiative.archiveSuccess"));
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("initiative.saveFailed"));
  } finally {
    archivingId.value = null;
  }
}

async function restoreItem(item: Initiative) {
  if (!await confirm({ title: "恢复倡议", message: t("initiative.restoreConfirm", { title: item.title }), confirmText: "恢复" })) return;
  archivingId.value = item.id;
  try {
    await initiativeApi.restore(workspaceId.value, item.id);
    toast.success(t("initiative.restoreSuccess"));
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("initiative.saveFailed"));
  } finally {
    archivingId.value = null;
  }
}

/* ---------- 进度百分比 ---------- */

function percent(rate: number | undefined): number {
  if (rate === undefined || Number.isNaN(rate)) return 0;
  return Math.round(Math.max(0, Math.min(1, rate)) * 100);
}

function progressColor(rate: number | undefined): string {
  const r = rate ?? 0;
  if (r >= 0.8) return "var(--success-500)";
  if (r >= 0.5) return "var(--brand-500)";
  return "var(--warning-500)";
}

/* ---------- 导航 ---------- */

function openDetail(item: Initiative) {
  router.push({
    name: "workspace-initiative-detail",
    params: { id: item.id },
  });
}

/* ---------- 数据加载 ---------- */

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const res = await initiativeApi.list(workspaceId.value, {
      status: filterStatus.value || undefined,
      priority: filterPriority.value || undefined,
      page: 1,
      pageSize: 100,
    });
    initiatives.value = res.results ?? [];
    total.value = res.total ?? 0;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t("initiative.loadFailed");
  } finally {
    loading.value = false;
  }
}

async function loadRelations() {
  try {
    const [m, p] = await Promise.all([
      workspaceApi.listMembers(workspaceId.value).catch(() => [] as Member[]),
      workspaceApi.listProjects(workspaceId.value).catch(() => [] as Project[]),
    ]);
    members.value = m ?? [];
    projects.value = p ?? [];
  } catch {
    members.value = [];
    projects.value = [];
  }
}

function setFilterStatus(s: InitiativeStatus | "") {
  filterStatus.value = s;
  load();
}

function setFilterPriority(p: InitiativePriority | "") {
  filterPriority.value = p;
  load();
}

onMounted(async () => {
  await Promise.all([load(), loadRelations()]);
});
</script>

<template>
  
  <AppPageContainer max="1100px">
    <div class="initiative-list">
    <!-- Header -->
    <header class="header">
      <div>
        <h1 class="header__title">{{ t("initiative.title") }}</h1>
        <p class="header__hint">{{ t("initiative.emptyDesc") }}</p>
      </div>
      <div class="header__actions">
        <AppButton variant="primary" size="sm" @click="openCreate">
          + {{ t("initiative.create") }}
        </AppButton>
      </div>
    </header>

    <!-- Filter bar -->
    <div class="filter-bar">
      <div class="filter-bar__group">
        <span class="filter-bar__label">{{ t("initiative.filterStatus") }}:</span>
        <button
          class="filter-chip"
          :class="{ 'filter-chip--active': filterStatus === '' }"
          @click="setFilterStatus('')"
        >
          {{ t("initiative.allStatus") }}
        </button>
        <button
          v-for="s in STATUS_GROUPS"
          :key="s"
          class="filter-chip"
          :class="{ 'filter-chip--active': filterStatus === s }"
          @click="setFilterStatus(s)"
        >
          {{ statusLabel[s] }}
        </button>
      </div>
      <div class="filter-bar__group">
        <span class="filter-bar__label">{{ t("initiative.filterPriority") }}:</span>
        <button
          class="filter-chip"
          :class="{ 'filter-chip--active': filterPriority === '' }"
          @click="setFilterPriority('')"
        >
          {{ t("initiative.allPriority") }}
        </button>
        <button
          v-for="p in PRIORITY_GROUPS"
          :key="p"
          class="filter-chip"
          :class="{ 'filter-chip--active': filterPriority === p }"
          @click="setFilterPriority(p)"
        >
          {{ priorityLabel[p] }}
        </button>
      </div>
    </div>

    <!-- Loading -->
    <AppSkeleton v-if="loading" variant="card" />

    <!-- Error -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- Empty -->
    <AppEmptyState
      v-else-if="initiatives.length === 0"
      icon="🎯"
      :title="t('initiative.empty')"
      :description="t('initiative.emptyDesc')"
    >
      <AppButton variant="primary" size="sm" @click="openCreate">
        {{ t("initiative.create") }}
      </AppButton>
    </AppEmptyState>

    <!-- Cards -->
    <div v-else class="cards">
      <article
        v-for="item in initiatives"
        :key="item.id"
        class="card"
        :class="{ 'card--archived': !!item.archived_at }"
        tabindex="0"
        role="button"
        @click="openDetail(item)"
        @keydown.enter="openDetail(item)"
      >
        <div class="card__top">
          <div class="card__title-row">
            <h3 class="card__name">{{ item.title }}</h3>
            <AppBadge v-if="item.archived_at" variant="default">
              {{ t("initiative.archived") }}
            </AppBadge>
          </div>
          <div class="card__badges">
            <AppBadge :variant="statusBadgeVariant[item.status]">
              {{ statusLabel[item.status] }}
            </AppBadge>
            <AppBadge :variant="priorityBadgeVariant[item.priority]">
              {{ priorityLabel[item.priority] }}
            </AppBadge>
          </div>
        </div>

        <p v-if="item.description" class="card__desc">{{ item.description }}</p>

        <!-- Progress -->
        <div v-if="item.progress" class="card__progress">
          <ProgressBar
            :percent="percent(item.progress.completion_rate)"
            size="sm"
            :color="progressColor(item.progress.completion_rate)"
            :label="`${t('initiative.completionRate')} ${percent(item.progress.completion_rate)}%`"
          />
        </div>

        <!-- Meta -->
        <div class="card__meta">
          <span class="card__meta-item">
            {{ t("initiative.projects") }}: {{ item.progress?.total_projects ?? 0 }}
          </span>
          <span class="card__meta-item">
            {{ t("initiative.workItems") }}: {{ item.progress?.completed_work_items ?? 0 }}/{{
              item.progress?.total_work_items ?? 0
            }}
          </span>
          <span v-if="item.start_date || item.target_date" class="card__meta-item">
            📅 {{ item.start_date ?? "—" }} ~ {{ item.target_date ?? "—" }}
          </span>
          <span class="card__meta-item">
            {{ t("initiative.owner") }}: {{ ownerName(item.owner_id) }}
          </span>
        </div>

        <!-- Actions -->
        <div class="card__actions" @click.stop>
          <button class="card__action" @click="openEdit(item)">
            {{ t("common.edit") }}
          </button>
          <button
            v-if="!item.archived_at"
            class="card__action card__action--danger"
            :disabled="archivingId === item.id"
            @click="archiveItem(item)"
          >
            {{ archivingId === item.id ? t("common.loading") : t("initiative.archive") }}
          </button>
          <button
            v-else
            class="card__action"
            :disabled="archivingId === item.id"
            @click="restoreItem(item)"
          >
            {{ archivingId === item.id ? t("common.loading") : t("initiative.restore") }}
          </button>
        </div>
      </article>
    </div>

    <!-- Create / Edit Modal -->
    <AppModal
      :visible="showForm"
      :title="editingId ? t('initiative.edit') : t('initiative.create')"
      width="560px"
      @close="closeForm"
    >
      <form @submit.prevent="submitForm">
        <label class="form__field">
          <span class="form__label">
            {{ t("initiative.titleField") }} <span class="form__required">*</span>
          </span>
          <input
            v-model="form.title"
            :placeholder="t('initiative.titlePlaceholder')"
            maxlength="200"
            class="form__input"
            autofocus
          />
        </label>

        <label class="form__field">
          <span class="form__label">{{ t("initiative.description") }}</span>
          <textarea
            v-model="form.description"
            :placeholder="t('initiative.descriptionPlaceholder')"
            maxlength="2000"
            rows="3"
            class="form__input"
          ></textarea>
        </label>

        <div class="form__row">
          <label class="form__field">
            <span class="form__label">{{ t("initiative.filterStatus") }}</span>
            <select v-model="form.status" class="form__input">
              <option v-for="s in STATUS_GROUPS" :key="s" :value="s">
                {{ statusLabel[s] }}
              </option>
            </select>
          </label>
          <label class="form__field">
            <span class="form__label">{{ t("initiative.filterPriority") }}</span>
            <select v-model="form.priority" class="form__input">
              <option v-for="p in PRIORITY_GROUPS" :key="p" :value="p">
                {{ priorityLabel[p] }}
              </option>
            </select>
          </label>
        </div>

        <div class="form__row">
          <label class="form__field">
            <span class="form__label">{{ t("initiative.startDate") }}</span>
            <input v-model="form.start_date" type="date" class="form__input" />
          </label>
          <label class="form__field">
            <span class="form__label">{{ t("initiative.targetDate") }}</span>
            <input v-model="form.target_date" type="date" class="form__input" />
          </label>
        </div>

        <label class="form__field">
          <span class="form__label">{{ t("initiative.owner") }}</span>
          <select v-model="form.owner_id" class="form__input">
            <option :value="null">{{ t("initiative.ownerPlaceholder") }}</option>
            <option v-for="m in members" :key="m.id" :value="m.id">
              {{ m.display_name || m.email }}
            </option>
          </select>
        </label>

        <label class="form__field">
          <span class="form__label">{{ t("initiative.linkedProjects") }}</span>
          <select v-model="form.project_ids" multiple class="form__input form__input--multi">
            <option v-for="p in projects" :key="p.id" :value="p.id">
              {{ p.name }} ({{ p.identifier }})
            </option>
          </select>
        </label>

        <div v-if="formError" class="form__error">{{ formError }}</div>
      </form>

      <template #footer>
        <AppButton variant="secondary" size="sm" @click="closeForm">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton variant="primary" size="sm" :loading="saving" @click="submitForm">
          {{ t("common.save") }}
        </AppButton>
      </template>
    </AppModal>
    </div>
  </AppPageContainer>
</template>


<style scoped>
.initiative-list {
  
}

/* ---- header ---- */
.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
  flex-wrap: wrap;
}
.header__title { margin: 0; font-size: 20px; font-weight: 600; }
.header__hint { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }
.header__actions { display: flex; gap: 8px; }

/* ---- filter bar ---- */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding: 12px 14px;
  margin-bottom: 16px;
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}
.filter-bar__group {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.filter-bar__label {
  font-size: 12px;
  color: var(--text-tertiary);
  font-weight: 500;
}
.filter-chip {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: 999px;
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}
.filter-chip:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.filter-chip--active {
  background: var(--brand-50);
  border-color: var(--brand-500);
  color: var(--brand-600);
  font-weight: 500;
}

/* ---- cards ---- */
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 14px;
}
.card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.card:hover {
  border-color: var(--brand-500);
  box-shadow: var(--shadow-popover);
}
.card--archived {
  opacity: 0.7;
}
.card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.card__title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.card__name {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.card__badges {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
.card__desc {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card__progress { margin-top: 2px; }
.card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 12px;
  color: var(--text-tertiary);
}
.card__meta-item { white-space: nowrap; }
.card__actions {
  display: flex;
  gap: 8px;
  padding-top: 10px;
  border-top: 1px solid var(--border-subtle);
}
.card__action {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}
.card__action:hover:not(:disabled) {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.card__action--danger:hover:not(:disabled) {
  border-color: var(--danger-500);
  color: var(--danger-500);
}
.card__action:disabled { opacity: 0.5; cursor: not-allowed; }

/* ---- form ---- */
.form__field {
  display: block;
  margin-bottom: 14px;
}
.form__label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-secondary);
}
.form__required { color: var(--danger-500); }
.form__input {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.15s;
  box-sizing: border-box;
}
textarea.form__input {
  height: auto;
  padding: 8px 12px;
  resize: vertical;
}
.form__input--multi {
  height: auto;
  min-height: 80px;
  padding: 6px;
}
.form__input--multi option {
  padding: 6px 8px;
  border-radius: var(--radius-sm);
}
.form__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.form__input::placeholder { color: var(--text-tertiary); }
.form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.form__error {
  color: var(--danger-500);
  font-size: 12px;
  margin-bottom: 8px;
}
</style>
