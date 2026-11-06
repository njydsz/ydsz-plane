<script setup lang="ts">
/**
 * 倡议详情页 — 展示倡议详情、进度汇总、关联项目/工作项与活动记录。
 *
 * 数据来源：initiativeApi.get / progress / addProject / removeProject /
 * addWorkItem / removeWorkItem / archive / restore。
 * 活动记录：由于后端倡议域暂无独立活动端点，这里以会话内事件流
 * （创建/更新/归档/恢复/关联变更）形式呈现，避免虚构 API。
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
import { formatRelativeTime } from "@/lib/formatTime";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";
import { AppBadge, AppButton, AppErrorState, AppLoadingState, AppModal, AppPageContainer, ProgressBar } from "@/components";

const { t } = useI18n();
const confirm = useConfirm();
const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const initiativeId = computed(() => Number(route.params.id));

/* ---------- 详情状态 ---------- */

const initiative = ref<Initiative | null>(null);
const loading = ref(true);
const error = ref("");
const actionError = ref("");
const actionLoading = ref("");

/* ---------- 关联数据 ---------- */

const members = ref<Member[]>([]);
const projects = ref<Project[]>([]);
/** 已关联项目 ID 集合（用于过滤可添加项目） */
const linkedProjectIds = ref<Set<number>>(new Set());

/* ---------- 活动记录（会话内事件流） ---------- */

interface ActivityEntry {
  id: number;
  verb: string;
  text: string;
  at: string;
}

const activities = ref<ActivityEntry[]>([]);
let nextActivityId = 1;

function pushActivity(verb: string, text: string) {
  activities.value.unshift({
    id: nextActivityId++,
    verb,
    text,
    at: new Date().toISOString(),
  });
}

/* ---------- 标签映射 ---------- */

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

/* ---------- 计算属性 ---------- */

const owner = computed(() => {
  const oid = initiative.value?.owner_id;
  if (!oid) return null;
  return members.value.find((m) => m.id === oid) ?? null;
});

const availableProjects = computed(() =>
  projects.value.filter((p) => !linkedProjectIds.value.has(p.id)),
);

const progressPercent = computed(() =>
  Math.round((initiative.value?.progress?.completion_rate ?? 0) * 100),
);

const progressColor = computed(() => {
  const r = initiative.value?.progress?.completion_rate ?? 0;
  if (r >= 0.8) return "var(--success-500)";
  if (r >= 0.5) return "var(--brand-500)";
  return "var(--warning-500)";
});

/* ---------- 数据加载 ---------- */

async function load() {
  if (!workspaceId.value || !initiativeId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    initiative.value = await initiativeApi.get(workspaceId.value, initiativeId.value);
    // 初始活动记录
    if (activities.value.length === 0) {
      pushActivity(
        "created",
        `${t("initiative.title")}: ${initiative.value.title}`,
      );
    }
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
    // 关联项目集合：优先取 progress 中的项目数，实际 ID 通过倡议详情里的项目关联端点
    // 后端未在 get 中返回 project_ids，这里以空集合起步，添加时增量维护
    linkedProjectIds.value = new Set();
  } catch {
    members.value = [];
    projects.value = [];
  }
}

/* ---------- 归档 / 恢复 ---------- */

async function archive() {
  if (!initiative.value) return;
  if (!await confirm({ title: "归档倡议", message: t("initiative.archiveConfirm", { title: initiative.value.title }), confirmText: "归档", variant: "danger" })) return;
  actionLoading.value = "archive";
  actionError.value = "";
  try {
    await initiativeApi.archive(workspaceId.value, initiativeId.value);
    toast.success(t("initiative.archiveSuccess"));
    pushActivity("archived", initiative.value.title);
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    actionLoading.value = "";
  }
}

async function restore() {
  if (!initiative.value) return;
  if (!await confirm({ title: "恢复倡议", message: t("initiative.restoreConfirm", { title: initiative.value.title }), confirmText: "恢复" })) return;
  actionLoading.value = "restore";
  actionError.value = "";
  try {
    await initiativeApi.restore(workspaceId.value, initiativeId.value);
    toast.success(t("initiative.restoreSuccess"));
    pushActivity("restored", initiative.value.title);
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    actionLoading.value = "";
  }
}

/* ---------- 关联项目 ---------- */

const showAddProject = ref(false);
const selectedProjectId = ref<number | null>(null);
const addingProject = ref(false);

async function addProject() {
  if (!selectedProjectId.value) return;
  addingProject.value = true;
  actionError.value = "";
  try {
    await initiativeApi.addProject(workspaceId.value, initiativeId.value, selectedProjectId.value);
    const p = projects.value.find((x) => x.id === selectedProjectId.value);
    toast.success(t("initiative.addProjectSuccess"));
    pushActivity("linked", `${t("initiative.addProject")}: ${p?.name ?? selectedProjectId.value}`);
    linkedProjectIds.value.add(selectedProjectId.value);
    showAddProject.value = false;
    selectedProjectId.value = null;
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    addingProject.value = false;
  }
}

const removingProjectId = ref<number | null>(null);

async function removeProject(projectId: number) {
  removingProjectId.value = projectId;
  actionError.value = "";
  try {
    await initiativeApi.removeProject(workspaceId.value, initiativeId.value, projectId);
    toast.success(t("initiative.removeProjectSuccess"));
    pushActivity("unlinked", `${t("initiative.removeProject")}: #${projectId}`);
    linkedProjectIds.value.delete(projectId);
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    removingProjectId.value = null;
  }
}

/* ---------- 关联工作项 ---------- */

const showAddWorkItem = ref(false);
const workItemIdentifier = ref("");
const workItemType = ref<"requirement" | "task" | "defect">("task");
const addingWorkItem = ref(false);
/** 会话内已关联工作项（用于展示与移除；后端未提供列表端点） */
const linkedWorkItems = ref<{ id: number; identifier: string; type: string }[]>([]);

async function addWorkItem() {
  const identifier = workItemIdentifier.value.trim();
  if (!identifier) return;
  addingWorkItem.value = true;
  actionError.value = "";
  try {
    // 通过项目内搜索定位工作项 ID（取第一个项目作为查询上下文）
    const project = projects.value[0];
    if (!project) {
      actionError.value = t("initiative.loadFailed");
      return;
    }
    const { issueApi } = await import("@/api/services/issue");
    const res = await issueApi.listIssues(workspaceId.value, project.id, {
      search: identifier,
      limit: 5,
    });
    const match = (res.results ?? []).find(
      (i) => i.identifier.toLowerCase() === identifier.toLowerCase(),
    );
    if (!match) {
      actionError.value = t("initiative.selectWorkItem") + `: ${identifier}`;
      return;
    }
    await initiativeApi.addWorkItem(workspaceId.value, initiativeId.value, {
      work_item_id: match.id,
      work_item_type: workItemType.value,
    });
    toast.success(t("initiative.addWorkItemSuccess"));
    pushActivity("linked", `${t("initiative.addWorkItem")}: ${match.identifier}`);
    linkedWorkItems.value.push({
      id: match.id,
      identifier: match.identifier,
      type: workItemType.value,
    });
    showAddWorkItem.value = false;
    workItemIdentifier.value = "";
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    addingWorkItem.value = false;
  }
}

const removingWorkItemId = ref<number | null>(null);

async function removeWorkItem(itemId: number) {
  removingWorkItemId.value = itemId;
  actionError.value = "";
  try {
    await initiativeApi.removeWorkItem(workspaceId.value, initiativeId.value, itemId);
    toast.success(t("initiative.removeWorkItemSuccess"));
    pushActivity("unlinked", `${t("initiative.removeWorkItem")}: #${itemId}`);
    linkedWorkItems.value = linkedWorkItems.value.filter((x) => x.id !== itemId);
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : t("initiative.saveFailed");
  } finally {
    removingWorkItemId.value = null;
  }
}

/* ---------- 导航 ---------- */

function backToList() {
  router.push({ name: "workspace-initiatives" });
}

/* ---------- 生命周期 ---------- */

onMounted(async () => {
  await loadRelations();
  await load();
});
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="initiative-detail">
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else-if="initiative">
      <!-- Header -->
      <header class="detail-header">
        <div class="detail-header__left">
          <button class="back-btn" @click="backToList">
            {{ t("initiative.backToList") }}
          </button>
          <div>
            <h1 class="detail-header__name">{{ initiative.title }}</h1>
            <div class="detail-header__meta">
              <AppBadge :variant="statusBadgeVariant[initiative.status]">
                {{ statusLabel[initiative.status] }}
              </AppBadge>
              <AppBadge :variant="priorityBadgeVariant[initiative.priority]">
                {{ t("initiative.filterPriority") }}: {{ priorityLabel[initiative.priority] }}
              </AppBadge>
              <span v-if="initiative.archived_at" class="detail-header__date">
                {{ t("initiative.archived") }}
              </span>
            </div>
          </div>
        </div>
        <div class="detail-header__actions">
          <AppButton
            v-if="!initiative.archived_at"
            variant="secondary"
            size="sm"
            :loading="actionLoading === 'archive'"
            @click="archive"
          >
            {{ t("initiative.archive") }}
          </AppButton>
          <AppButton
            v-else
            variant="primary"
            size="sm"
            :loading="actionLoading === 'restore'"
            @click="restore"
          >
            {{ t("initiative.restore") }}
          </AppButton>
        </div>
      </header>

      <p v-if="actionError" class="action-error">{{ actionError }}</p>

      <div class="layout">
        <!-- Main column -->
        <div class="main-col">
          <!-- Description -->
          <section v-if="initiative.description" class="panel">
            <h3 class="panel__title">{{ t("initiative.description") }}</h3>
            <p class="panel__text">{{ initiative.description }}</p>
          </section>

          <!-- Progress -->
          <section class="panel">
            <h3 class="panel__title">{{ t("initiative.progress") }}</h3>
            <ProgressBar
              :percent="progressPercent"
              size="lg"
              :color="progressColor"
              :striped="initiative.status === 'on_track'"
            />
            <div class="progress-stats">
              <div class="progress-stat">
                <span class="progress-stat__value">
                  {{ initiative.progress?.completed_work_items ?? 0 }}/{{
                    initiative.progress?.total_work_items ?? 0
                  }}
                </span>
                <span class="progress-stat__label">{{ t("initiative.workItems") }}</span>
              </div>
              <div class="progress-stat">
                <span class="progress-stat__value">
                  {{ initiative.progress?.total_projects ?? 0 }}
                </span>
                <span class="progress-stat__label">{{ t("initiative.projects") }}</span>
              </div>
              <div class="progress-stat">
                <span class="progress-stat__value">{{ progressPercent }}%</span>
                <span class="progress-stat__label">{{ t("initiative.completionRate") }}</span>
              </div>
            </div>
            <div class="progress-status-row">
              <span class="status-pill status-pill--success">
                {{ t("initiative.statusLabel.on_track") }}: {{ initiative.progress?.on_track_count ?? 0 }}
              </span>
              <span class="status-pill status-pill--warning">
                {{ t("initiative.statusLabel.at_risk") }}: {{ initiative.progress?.at_risk_count ?? 0 }}
              </span>
              <span class="status-pill status-pill--danger">
                {{ t("initiative.statusLabel.off_track") }}: {{ initiative.progress?.off_track_count ?? 0 }}
              </span>
            </div>
          </section>

          <!-- Linked projects -->
          <section class="panel">
            <div class="panel__header">
              <h3 class="panel__title">
                {{ t("initiative.linkedProjects") }}
                <span class="panel__hint">{{ linkedProjectIds.size || initiative.progress?.total_projects || 0 }}</span>
              </h3>
              <AppButton variant="secondary" size="sm" @click="showAddProject = true">
                {{ t("initiative.addProject") }}
              </AppButton>
            </div>
            <div v-if="linkedProjectIds.size === 0" class="panel__empty">
              —
            </div>
            <ul v-else class="link-list">
              <li v-for="pid in linkedProjectIds" :key="pid" class="link-list__item">
                <span class="link-list__name">
                  {{ projects.find((p) => p.id === pid)?.name ?? `#${pid}` }}
                </span>
                <button
                  class="link-list__remove"
                  :disabled="removingProjectId === pid"
                  @click="removeProject(pid)"
                >
                  {{ removingProjectId === pid ? t("common.loading") : t("initiative.removeProject") }}
                </button>
              </li>
            </ul>
          </section>

          <!-- Linked work items -->
          <section class="panel">
            <div class="panel__header">
              <h3 class="panel__title">
                {{ t("initiative.linkedWorkItems") }}
                <span class="panel__hint">{{ linkedWorkItems.length || initiative.progress?.total_work_items || 0 }}</span>
              </h3>
              <AppButton variant="secondary" size="sm" @click="showAddWorkItem = true">
                {{ t("initiative.addWorkItem") }}
              </AppButton>
            </div>
            <div v-if="linkedWorkItems.length === 0" class="panel__empty">
              —
            </div>
            <ul v-else class="link-list">
              <li v-for="item in linkedWorkItems" :key="item.id" class="link-list__item">
                <span class="link-list__name">{{ item.identifier }}</span>
                <AppBadge variant="default">{{ item.type }}</AppBadge>
                <button
                  class="link-list__remove"
                  :disabled="removingWorkItemId === item.id"
                  @click="removeWorkItem(item.id)"
                >
                  {{ removingWorkItemId === item.id ? t("common.loading") : t("initiative.removeWorkItem") }}
                </button>
              </li>
            </ul>
          </section>
        </div>

        <!-- Sidebar -->
        <aside class="side-col">
          <!-- Info -->
          <section class="panel">
            <h3 class="panel__title">{{ t("common.status") }}</h3>
            <dl class="info-list">
              <div class="info-list__row">
                <dt>{{ t("initiative.startDate") }}</dt>
                <dd>{{ initiative.start_date ?? "—" }}</dd>
              </div>
              <div class="info-list__row">
                <dt>{{ t("initiative.targetDate") }}</dt>
                <dd>{{ initiative.target_date ?? "—" }}</dd>
              </div>
              <div class="info-list__row">
                <dt>{{ t("initiative.owner") }}</dt>
                <dd>{{ owner ? owner.display_name || owner.email : `#${initiative.owner_id}` }}</dd>
              </div>
              <div class="info-list__row">
                <dt>{{ t("common.createdAt") }}</dt>
                <dd>{{ new Date(initiative.created_at).toLocaleString() }}</dd>
              </div>
              <div class="info-list__row">
                <dt>{{ t("common.updatedAt") }}</dt>
                <dd>{{ new Date(initiative.updated_at).toLocaleString() }}</dd>
              </div>
            </dl>
          </section>

          <!-- Activity -->
          <section class="panel">
            <h3 class="panel__title">{{ t("initiative.activity") }}</h3>
            <div v-if="activities.length === 0" class="panel__empty">
              {{ t("initiative.noActivity") }}
            </div>
            <div v-else class="activity-timeline">
              <div v-for="act in activities" :key="act.id" class="activity-item">
                <div class="activity-item__dot" :class="`verb-${act.verb}`"></div>
                <div class="activity-item__body">
                  <div class="activity-item__text">{{ act.text }}</div>
                  <div class="activity-item__time">{{ formatRelativeTime(act.at) }}</div>
                </div>
              </div>
            </div>
          </section>
        </aside>
      </div>
    </template>

    <!-- Add project modal -->
    <AppModal
      :visible="showAddProject"
      :title="t('initiative.addProject')"
      width="440px"
      @close="showAddProject = false"
    >
      <label class="form__field">
        <span class="form__label">{{ t("initiative.selectProject") }}</span>
        <select v-model="selectedProjectId" class="form__input">
          <option :value="null" disabled>{{ t("initiative.selectProject") }}</option>
          <option v-for="p in availableProjects" :key="p.id" :value="p.id">
            {{ p.name }} ({{ p.identifier }})
          </option>
        </select>
      </label>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showAddProject = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :loading="addingProject"
          :disabled="!selectedProjectId"
          @click="addProject"
        >
          {{ t("common.confirm") }}
        </AppButton>
          </AppModal>

    <!-- Add work item modal -->
    <AppModal
      :visible="showAddWorkItem"
      :title="t('initiative.addWorkItem')"
      width="440px"
      @close="showAddWorkItem = false"
    >
      <label class="form__field">
        <span class="form__label">{{ t("initiative.selectWorkItem") }}</span>
        <input
          v-model="workItemIdentifier"
          :placeholder="t('initiative.selectWorkItem')"
          class="form__input"
        />
      </label>
      <label class="form__field">
        <span class="form__label">{{ t("issue.type") }}</span>
        <select v-model="workItemType" class="form__input">
          <option value="requirement">{{ t("issue.typeRequirement") }}</option>
          <option value="task">{{ t("issue.typeTask") }}</option>
          <option value="defect">{{ t("issue.typeDefect") }}</option>
        </select>
      </label>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showAddWorkItem = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :loading="addingWorkItem"
          :disabled="!workItemIdentifier.trim()"
          @click="addWorkItem"
        >
          {{ t("common.confirm") }}
        </AppButton>
      
    </AppModal>
  
    </div>
  </AppPageContainer>


    <!-- Add project modal -->
    <AppModal
      :visible="showAddProject"
      :title="t('initiative.addProject')"
      width="440px"
      @close="showAddProject = false"
    >
      <label class="form__field">
        <span class="form__label">{{ t("initiative.selectProject") }}</span>
        <select v-model="selectedProjectId" class="form__input">
          <option :value="null" disabled>{{ t("initiative.selectProject") }}</option>
          <option v-for="p in availableProjects" :key="p.id" :value="p.id">
            {{ p.name }} ({{ p.identifier }})
          </option>
        </select>
      </label>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showAddProject = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :loading="addingProject"
          :disabled="!selectedProjectId"
          @click="addProject"
        >
          {{ t("common.confirm") }}
        </AppButton>
      
    </AppModal>

    <!-- Add work item modal -->
    <AppModal
      :visible="showAddWorkItem"
      :title="t('initiative.addWorkItem')"
      width="440px"
      @close="showAddWorkItem = false"
    >
      <label class="form__field">
        <span class="form__label">{{ t("initiative.selectWorkItem") }}</span>
        <input
          v-model="workItemIdentifier"
          :placeholder="t('initiative.selectWorkItem')"
          class="form__input"
        />
      </label>
      <label class="form__field">
        <span class="form__label">{{ t("issue.type") }}</span>
        <select v-model="workItemType" class="form__input">
          <option value="requirement">{{ t("issue.typeRequirement") }}</option>
          <option value="task">{{ t("issue.typeTask") }}</option>
          <option value="defect">{{ t("issue.typeDefect") }}</option>
        </select>
      </label>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showAddWorkItem = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :loading="addingWorkItem"
          :disabled="!workItemIdentifier.trim()"
          @click="addWorkItem"
        >
          {{ t("common.confirm") }}
        </AppButton>
      
    </AppModal>
  </div>


<style scoped>
.initiative-detail {
  
}

/* ---- header ---- */
.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.detail-header__left {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}
.back-btn {
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
  flex-shrink: 0;
}
.back-btn:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.detail-header__name {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}
.detail-header__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.detail-header__date {
  font-size: 12px;
  color: var(--text-tertiary);
}
.detail-header__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.action-error {
  color: var(--danger-500);
  font-size: 13px;
  margin: 0 0 12px;
}

/* ---- layout ---- */
.layout {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 16px;
  align-items: start;
}
@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }
}
.main-col,
.side-col {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

/* ---- panels ---- */
.panel {
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}
.panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;
}
.panel__title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.panel__header .panel__title {
  margin: 0;
}
.panel__hint {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary);
  margin-left: 4px;
}
.panel__text {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
  white-space: pre-wrap;
}
.panel__empty {
  font-size: 13px;
  color: var(--text-tertiary);
}

/* ---- progress ---- */
.progress-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 14px;
}
.progress-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.progress-stat__value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: var(--font-mono);
}
.progress-stat__label {
  font-size: 12px;
  color: var(--text-tertiary);
}
.progress-status-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}
.status-pill {
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 500;
}
.status-pill--success {
  background: var(--success-50);
  color: var(--success-500);
}
.status-pill--warning {
  background: var(--warning-50);
  color: var(--warning-500);
}
.status-pill--danger {
  background: var(--danger-50);
  color: var(--danger-500);
}

/* ---- link list ---- */
.link-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.link-list__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}
.link-list__name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.link-list__remove {
  padding: 3px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  flex-shrink: 0;
}
.link-list__remove:hover:not(:disabled) {
  border-color: var(--danger-500);
  color: var(--danger-500);
}
.link-list__remove:disabled { opacity: 0.5; cursor: not-allowed; }

/* ---- info list ---- */
.info-list {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.info-list__row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}
.info-list__row dt {
  color: var(--text-tertiary);
  flex-shrink: 0;
}
.info-list__row dd {
  margin: 0;
  color: var(--text-primary);
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---- activity ---- */
.activity-timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.activity-item {
  display: flex;
  gap: 10px;
}
.activity-item__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-top: 5px;
  flex-shrink: 0;
  background: var(--neutral-400);
}
.activity-item__dot.verb-created { background: var(--success-500); }
.activity-item__dot.verb-archived { background: var(--danger-500); }
.activity-item__dot.verb-restored { background: var(--brand-500); }
.activity-item__dot.verb-linked { background: var(--success-500); }
.activity-item__dot.verb-unlinked { background: var(--warning-500); }
.activity-item__body {
  min-width: 0;
}
.activity-item__text {
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.4;
}
.activity-item__time {
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

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
.form__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.form__input::placeholder { color: var(--text-tertiary); }
</style>
