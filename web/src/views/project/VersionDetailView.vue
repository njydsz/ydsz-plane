<script setup lang="ts">
/**
 * 版本日详情页 — 展示进度、质量门禁、缺陷面板与迭代聚合。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version } from "@/api/services/version";
import { AppBadge, AppButton, ProgressBar, MiniGantt, AppEmptyState } from "@/components";
import type { GanttSprint } from "@/components/MiniGantt.vue";
import DefectPanel from "./DefectPanel.vue";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const versionId = computed(() => Number(route.params.versionId));

const activeTab = ref<"overview" | "sprints" | "defects" | "notes">("overview");

const version = ref<Version | null>(null);
const loading = ref(true);
const error = ref("");
const actionError = ref("");
const actionLoading = ref(""); // which action is in progress

// sprint management
const showAddSprint = ref(false);
const availableSprints = ref<{ id: number; name: string }[]>([]);
const removingSprintId = ref<number | null>(null);

// regenerate notes
const regeneratingNotes = ref(false);

let wsIdVal = 0;

/* ---------- status helpers ---------- */

const statusLabel: Record<string, string> = {
  planning: "规划中",
  active: "进行中",
  released: "已发布",
  archived: "已归档",
};

const statusBadgeVariant: Record<string, "warning" | "success" | "info" | "default"> = {
  planning: "warning",
  active: "success",
  released: "info",
  archived: "default",
};

const sprintStatusLabel: Record<string, string> = {
  planned: "未开始",
  active: "进行中",
  completed: "已完成",
};

/* ---------- computed ---------- */

const canEdit = computed(() => {
  const s = version.value?.status;
  return s === "planning" || s === "active";
});

const ganttSprints = computed<GanttSprint[]>(() => {
  if (!version.value?.sprints) return [];
  return version.value.sprints.map((s) => ({
    id: s.sprint_id,
    name: s.name,
    startDate: s.start_date,
    endDate: s.end_date,
    progress: s.progress
      ? s.progress.total_points > 0
        ? (s.progress.done_points / s.progress.total_points) * 100
        : 0
      : undefined,
    status: s.status,
  }));
});

const progressColor = computed(() => {
  const rate = version.value?.progress?.completion_rate ?? 0;
  if (rate >= 0.9) return "var(--success-500)";
  if (rate >= 0.6) return "var(--brand-500)";
  return "var(--warning-500)";
});

/* ---------- data ---------- */

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    version.value = await versionApi.getVersion(wsId, projectId.value, versionId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载版本详情失败";
  } finally {
    loading.value = false;
  }
}

/* ---------- actions ---------- */

async function transition(action: "activate" | "archive") {
  actionError.value = "";
  actionLoading.value = action;
  try {
    const wsId = await resolveWsId();
    if (action === "activate") {
      await versionApi.activateVersion(wsId, projectId.value, versionId.value);
    } else {
      if (!confirm(`确定要归档版本「${version.value?.name}」吗？`)) return;
      await versionApi.archiveVersion(wsId, projectId.value, versionId.value);
    }
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "操作失败";
  } finally {
    actionLoading.value = "";
  }
}

async function toggleChecklist(itemId: string) {
  if (!version.value || !canEdit.value) return;
  const v = version.value;
  const newList = (v.checklist ?? []).map((c) =>
    c.id === itemId ? { ...c, checked: !c.checked } : c,
  );
  try {
    const wsId = await resolveWsId();
    await versionApi.updateVersion(wsId, projectId.value, versionId.value, {
      checklist: newList,
      version: (v as any).version ?? 0,
    } as any);
    version.value = await versionApi.getVersion(wsId, projectId.value, versionId.value);
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "保存清单失败";
  }
}

async function refreshAvailableSprints() {
  try {
    const { sprintApi } = await import("@/api/services/sprint");
    const wsId = await resolveWsId();
    const res = await sprintApi.listSprints(wsId, projectId.value, { limit: 100 });
    const attachedIds = new Set((version.value?.sprints ?? []).map((s) => s.sprint_id));
    availableSprints.value = (res.results ?? [])
      .filter((s: any) => !attachedIds.has(s.id))
      .map((s: any) => ({ id: s.id, name: s.name }));
  } catch {
    availableSprints.value = [];
  }
}

async function addSprint(sprintId: number) {
  if (!sprintId) return;
  try {
    const wsId = await resolveWsId();
    await versionApi.addSprint(wsId, projectId.value, versionId.value, { sprint_id: sprintId });
    showAddSprint.value = false;
    await load();
    await refreshAvailableSprints();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "挂入迭代失败";
  }
}

async function removeSprint(sprintId: number) {
  removingSprintId.value = sprintId;
  try {
    const wsId = await resolveWsId();
    await versionApi.removeSprint(wsId, projectId.value, versionId.value, sprintId);
    await load();
    await refreshAvailableSprints();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "解绑迭代失败";
  } finally {
    removingSprintId.value = null;
  }
}

async function regenerateNotes() {
  regeneratingNotes.value = true;
  try {
    const wsId = await resolveWsId();
    const res = await versionApi.regenerateNotes(wsId, projectId.value, versionId.value, false);
    if (version.value) {
      version.value = {
        ...version.value,
        release_notes: res.release_notes,
      };
    }
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "重新生成失败";
  } finally {
    regeneratingNotes.value = false;
  }
}

function goRelease() {
  router.push({ name: "version-release", params: { versionId: versionId.value } });
}

function goDeliveryReport() {
  router.push({ name: "version-delivery-report", params: { versionId: versionId.value } });
}

/* ---------- lifecycle ---------- */

onMounted(async () => {
  await load();
  await refreshAvailableSprints();
});
</script>

<template>
  <div class="version-detail">
    <!-- Loading -->
    <div v-if="loading" class="skeleton-detail">
      <div class="skeleton-line" style="width:50%"></div>
      <div class="skeleton-line" style="width:30%"></div>
      <div class="skeleton-line" style="width:100%; height:80px;"></div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="center-message error">
      <p class="center-message__title">加载版本详情失败</p>
      <p class="center-message__detail">{{ error }}</p>
      <AppButton variant="secondary" size="sm" @click="load">重试</AppButton>
    </div>

    <template v-else-if="version">
      <!-- Header -->
      <header class="detail-header">
        <div class="detail-header__left">
          <button class="back-btn" @click="router.back()" title="返回列表">
            ← 返回
          </button>
          <div>
            <h1 class="detail-header__name">{{ version.name }}</h1>
            <div class="detail-header__meta">
              <span class="detail-header__semver">{{ version.semver }}</span>
              <AppBadge :variant="statusBadgeVariant[version.status]">
                {{ statusLabel[version.status] }}
              </AppBadge>
              <span v-if="version.target_date" class="detail-header__date">
                📅 目标 {{ version.target_date }}
              </span>
              <span v-if="version.delivered_at" class="detail-header__date">
                🚀 已发布于 {{ new Date(version.delivered_at).toLocaleDateString("zh-CN") }}
              </span>
            </div>
          </div>
        </div>
        <div class="detail-header__actions">
          <AppButton
            v-if="version.status === 'planning'"
            variant="warning"
            size="sm"
            :loading="actionLoading === 'activate'"
            @click="transition('activate')"
          >
            启动开发
          </AppButton>
          <AppButton
            v-if="version.status === 'active'"
            variant="success"
            size="sm"
            @click="goRelease"
          >
            发布版本
          </AppButton>
          <AppButton
            v-if="version.status === 'released' || version.status === 'archived'"
            variant="secondary"
            size="sm"
            @click="goDeliveryReport"
          >
            交付报告
          </AppButton>
          <AppButton
            v-if="version.status !== 'archived'"
            variant="secondary"
            size="sm"
            :loading="actionLoading === 'archive'"
            @click="transition('archive')"
          >
            归档
          </AppButton>
        </div>
      </header>

      <p v-if="actionError" class="action-error">{{ actionError }}</p>

      <!-- Tabs -->
      <nav class="tabs">
        <button
          v-for="t in (['overview', 'sprints', 'defects', 'notes'] as const)"
          :key="t"
          class="tabs__tab"
          :class="{ 'tabs__tab--active': activeTab === t }"
          @click="activeTab = t"
        >
          {{ { overview: "概览", sprints: "迭代", defects: "缺陷", notes: "Release Notes" }[t] }}
        </button>
      </nav>

      <!-- ========== Overview Tab ========== -->
      <div v-if="activeTab === 'overview'" class="tab-content">
        <!-- Description -->
        <div v-if="version.description" class="panel">
          <p class="panel__text">{{ version.description }}</p>
        </div>

        <!-- Progress -->
        <div class="panel">
          <h3 class="panel__title">版本进度</h3>
          <ProgressBar
            :percent="Math.round((version.progress?.completion_rate ?? 0) * 100)"
            size="lg"
            :color="progressColor"
            :striped="version.status === 'active'"
          />
          <div class="progress-stats">
            <div class="progress-stat">
              <span class="progress-stat__value">{{ version.progress?.done_points ?? 0 }} / {{ version.progress?.total_points ?? 0 }}</span>
              <span class="progress-stat__label">故事点</span>
            </div>
            <div class="progress-stat">
              <span class="progress-stat__value">{{ version.progress?.done_issues ?? 0 }} / {{ version.progress?.total_issues ?? 0 }}</span>
              <span class="progress-stat__label">工作项</span>
            </div>
            <div class="progress-stat">
              <span class="progress-stat__value">{{ version.progress?.sprint_count ?? 0 }}</span>
              <span class="progress-stat__label">关联迭代</span>
            </div>
          </div>
        </div>

        <!-- Quality Gate -->
        <div class="panel">
          <h3 class="panel__title">质量门禁</h3>
          <div class="quality-grid">
            <div class="quality-item" :class="{ 'quality-item--pass': (version.quality?.critical_bugs ?? 0) === 0 }">
              <span class="quality-item__value">{{ version.quality?.critical_bugs ?? 0 }}</span>
              <span class="quality-item__label">致命/严重未关闭</span>
            </div>
            <div class="quality-item">
              <span class="quality-item__value">{{ version.quality?.major_bugs ?? 0 }}</span>
              <span class="quality-item__label">一般未关闭</span>
            </div>
            <div class="quality-item">
              <span class="quality-item__value">{{ Math.round((version.quality?.fix_rate ?? 0) * 100) }}%</span>
              <span class="quality-item__label">修复率</span>
            </div>
            <div class="quality-item" :class="version.quality?.pass_quality_gate ? 'quality-item--pass' : 'quality-item--fail'">
              <span class="quality-item__value">{{ version.quality?.pass_quality_gate ? '✓ 通过' : '✗ 未通过' }}</span>
              <span class="quality-item__label">质量门禁</span>
            </div>
          </div>
        </div>

        <!-- Mini Gantt -->
        <div v-if="ganttSprints.length > 0" class="panel">
          <h3 class="panel__title">迭代时间线</h3>
          <MiniGantt
            :sprints="ganttSprints"
            :versionEnd="version.target_date"
          />
        </div>

        <!-- Checklist -->
        <div class="panel">
          <h3 class="panel__title">
            发布检查清单
            <span class="panel__hint">
              {{ (version.checklist ?? []).filter(c => c.checked).length }}/{{ version.checklist?.length ?? 0 }} 项完成
            </span>
          </h3>
          <ul v-if="version.checklist?.length" class="checklist">
            <li
              v-for="item in version.checklist"
              :key="item.id"
              class="checklist__item"
              :class="{ 'checklist__item--checked': item.checked }"
            >
              <input
                type="checkbox"
                :checked="item.checked"
                :disabled="!canEdit"
                :id="`chk-${item.id}`"
                class="checklist__checkbox"
                @change="toggleChecklist(item.id)"
              />
              <label :for="`chk-${item.id}`" class="checklist__label">
                {{ item.label }}
              </label>
              <span v-if="item.required" class="checklist__required">必做</span>
            </li>
          </ul>
          <AppEmptyState
            v-else
            title="暂无检查项"
            description="检查清单在创建或编辑版本时可配置"
          />
        </div>
      </div>

      <!-- ========== Sprints Tab ========== -->
      <div v-if="activeTab === 'sprints'" class="tab-content">
        <div class="sprint-toolbar">
          <span class="sprint-toolbar__count">
            {{ version.sprints?.length ?? 0 }} 个关联迭代
          </span>
          <AppButton
            v-if="canEdit"
            variant="secondary"
            size="sm"
            @click="showAddSprint = !showAddSprint; if (showAddSprint) refreshAvailableSprints()"
          >
            + 挂入迭代
          </AppButton>
        </div>

        <!-- Add sprint panel -->
        <div v-if="showAddSprint" class="add-sprint-panel">
          <select
            class="add-sprint-panel__select"
            @change="(e) => addSprint(Number((e.target as HTMLSelectElement).value))"
          >
            <option value="">选择迭代…</option>
            <option
              v-for="s in availableSprints"
              :key="s.id"
              :value="s.id"
            >{{ s.name }}</option>
          </select>
          <button class="add-sprint-panel__cancel" @click="showAddSprint = false">取消</button>
        </div>

        <!-- Sprint list -->
        <div v-if="version.sprints?.length" class="sprint-list">
          <div
            v-for="s in version.sprints"
            :key="s.sprint_id"
            class="sprint-row"
          >
            <div class="sprint-row__info">
              <span class="sprint-row__dot" :class="`sprint-row__dot--${s.status}`"></span>
              <span class="sprint-row__name">{{ s.name }}</span>
              <span class="sprint-row__status">{{ sprintStatusLabel[s.status] ?? s.status }}</span>
              <span v-if="s.start_date || s.end_date" class="sprint-row__date">
                {{ s.start_date ?? "?" }} → {{ s.end_date ?? "?" }}
              </span>
            </div>
            <div class="sprint-row__progress">
              <ProgressBar
                v-if="s.progress"
                :percent="s.progress.total_points > 0 ? Math.round((s.progress.done_points / s.progress.total_points) * 100) : 0"
                size="sm"
                :showLabel="false"
                :label="`${s.progress.done_points}/${s.progress.total_points} 点`"
              />
            </div>
            <AppButton
              v-if="canEdit"
              variant="ghost"
              size="sm"
              :loading="removingSprintId === s.sprint_id"
              @click="removeSprint(s.sprint_id)"
            >
              解绑
            </AppButton>
          </div>
        </div>

        <AppEmptyState
          v-else
          icon="🔗"
          title="暂无关联迭代"
          description="将迭代挂入版本日以追踪整体进度"
        >
          <AppButton
            v-if="canEdit"
            variant="primary"
            size="sm"
            @click="showAddSprint = true"
          >
            挂入迭代
          </AppButton>
        </AppEmptyState>
      </div>

      <!-- ========== Defects Tab ========== -->
      <div v-if="activeTab === 'defects'" class="tab-content">
        <DefectPanel
          :workspace-slug="workspaceSlug"
          :project-id="projectId"
          :version-id="versionId"
        />
      </div>

      <!-- ========== Release Notes Tab ========== -->
      <div v-if="activeTab === 'notes'" class="tab-content">
        <div class="notes-toolbar">
          <span class="notes-toolbar__label">Release Notes</span>
          <AppButton
            v-if="version.status === 'active' || version.status === 'released'"
            variant="secondary"
            size="sm"
            :loading="regeneratingNotes"
            @click="regenerateNotes"
          >
            重新生成
          </AppButton>
        </div>

        <div v-if="version.release_notes" class="notes-content">
          <pre class="notes-content__md">{{ version.release_notes }}</pre>
        </div>
        <AppEmptyState
          v-else
          icon="📝"
          title="尚未生成 Release Notes"
          description="发布版本时将自动生成，或点击「重新生成」手动触发"
        />
      </div>
    </template>
  </div>
</template>

<style scoped>
.version-detail {
  max-width: 960px;
}

/* ---- skeleton ---- */
.skeleton-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.skeleton-line {
  height: 14px;
  background: var(--surface-2);
  border-radius: 4px;
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

/* ---- center message ---- */
.center-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  gap: 10px;
}
.center-message.error { color: var(--danger-500); }
.center-message__title { margin: 0; font-size: 14px; font-weight: 500; }
.center-message__detail { margin: 0; font-size: 12px; opacity: 0.8; max-width: 400px; word-break: break-word; }

/* ---- header ---- */
.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}
.detail-header__left { display: flex; align-items: flex-start; gap: 12px; }
.back-btn {
  font-size: 13px;
  color: var(--text-tertiary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 0;
  white-space: nowrap;
  margin-top: 1px;
}
.back-btn:hover { color: var(--brand-500); }
.detail-header__name { margin: 0 0 4px; font-size: 20px; font-weight: 600; }
.detail-header__meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.detail-header__semver {
  font-size: 13px;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
}
.detail-header__date {
  font-size: 12px;
  color: var(--text-tertiary);
}
.detail-header__actions { display: flex; gap: 8px; }

.action-error {
  font-size: 12px;
  color: var(--danger-500);
  margin: 0 0 12px;
  padding: 8px 12px;
  background: rgba(220, 47, 47, 0.06);
  border-radius: var(--radius-sm);
}

/* ---- tabs ---- */
.tabs {
  display: flex;
  gap: 0;
  margin-bottom: 20px;
  border-bottom: 2px solid var(--border-subtle);
}
.tabs__tab {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-tertiary);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}
.tabs__tab:hover { color: var(--text-primary); }
.tabs__tab--active {
  color: var(--brand-500);
  border-bottom-color: var(--brand-500);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ---- panel ---- */
.panel {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 16px;
}
.panel__title {
  margin: 0 0 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}
.panel__hint {
  font-weight: 400;
  font-size: 12px;
  color: var(--text-tertiary);
}
.panel__text {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}

/* ---- progress stats ---- */
.progress-stats {
  display: flex;
  gap: 24px;
  margin-top: 12px;
}
.progress-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.progress-stat__value {
  font-size: 15px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.progress-stat__label {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ---- quality grid ---- */
.quality-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}
.quality-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 8px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  text-align: center;
}
.quality-item__value {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.quality-item__label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.quality-item--pass .quality-item__value { color: var(--success-500); }
.quality-item--fail .quality-item__value { color: var(--danger-500); }

/* ---- checklist ---- */
.checklist {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.checklist__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  transition: background 0.1s;
}
.checklist__item:hover { background: var(--surface-2); }
.checklist__item--checked .checklist__label {
  color: var(--text-tertiary);
  text-decoration: line-through;
}
.checklist__checkbox {
  width: 16px;
  height: 16px;
  accent-color: var(--brand-500);
  cursor: pointer;
  flex-shrink: 0;
}
.checklist__checkbox:disabled { cursor: not-allowed; }
.checklist__label {
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  flex: 1;
}
.checklist__required {
  font-size: 10px;
  font-weight: 600;
  color: var(--danger-500);
  background: rgba(220, 47, 47, 0.08);
  padding: 1px 6px;
  border-radius: 3px;
}

/* ---- sprint tab ---- */
.sprint-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.sprint-toolbar__count {
  font-size: 12px;
  color: var(--text-tertiary);
}

.add-sprint-panel {
  display: flex;
  gap: 8px;
  padding: 10px 12px;
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  align-items: center;
}
.add-sprint-panel__select {
  flex: 1;
  font-size: 13px;
  padding: 6px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
}
.add-sprint-panel__cancel {
  font-size: 12px;
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
}

.sprint-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.sprint-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
}
.sprint-row__info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.sprint-row__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.sprint-row__dot--planned { background: var(--text-tertiary); }
.sprint-row__dot--active { background: var(--success-500); }
.sprint-row__dot--completed { background: var(--brand-500); }
.sprint-row__name { font-size: 13px; font-weight: 500; }
.sprint-row__status { font-size: 11px; color: var(--text-tertiary); }
.sprint-row__date {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
.sprint-row__progress { width: 120px; flex-shrink: 0; }

/* ---- notes tab ---- */
.notes-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.notes-toolbar__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}
.notes-content {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 16px;
}
.notes-content__md {
  margin: 0;
  font-size: 13px;
  font-family: var(--font-mono);
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

/* ---- responsive ---- */
@media (max-width: 640px) {
  .quality-grid { grid-template-columns: repeat(2, 1fr); }
  .progress-stats { flex-wrap: wrap; gap: 16px; }
}
</style>
