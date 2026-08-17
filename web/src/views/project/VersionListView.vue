<script setup lang="ts">
/**
 * 版本列表页 — 展示版本列表，支持创建与状态流转。
 * 业务规则：一个版本聚合多个迭代（1:N），target_date 只是版本的一个属性。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version, type VersionStatus } from "@/api/services/version";
import { AppButton, AppBadge, AppEmptyState, AppErrorState, ProgressBar, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const versions = ref<Version[]>([]);
const loading = ref(true);
const error = ref("");
const total = ref(0);
const filterStatus = ref<VersionStatus | "">("");

// create
const showCreate = ref(false);
const form = ref({ name: "", semver: "", description: "", start_date: "", end_date: "", target_date: "" });
const creating = ref(false);
const createError = ref("");

// delete
const deletingId = ref<number | null>(null);

let wsIdVal = 0;

/* ---------- status helpers ---------- */

const STATUS_GROUPS = ["active", "planning", "released", "archived"] as const;

const statusLabel: Record<VersionStatus, string> = {
  planning: "规划中",
  active: "进行中",
  released: "已发布",
  archived: "已归档",
};

const statusBadgeVariant: Record<VersionStatus, "warning" | "success" | "info" | "default"> = {
  planning: "warning",
  active: "success",
  released: "info",
  archived: "default",
};

/* ---------- data ---------- */

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.get(workspaceId.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const res = await versionApi.listVersions(wsId, projectId.value, {
      status: filterStatus.value || undefined,
      limit: 100,
      offset: 0,
    });
    versions.value = res.results;
    total.value = res.total;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载版本列表失败";
  } finally {
    loading.value = false;
  }
}

async function createVersion() {
  if (!form.value.name.trim() || !form.value.semver.trim()) {
    createError.value = "版本名称和语义版本号不能为空";
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    const wsId = await resolveWsId();
    const v = await versionApi.createVersion(wsId, projectId.value, {
      name: form.value.name.trim(),
      semver: form.value.semver.trim(),
      description: form.value.description || undefined,
      start_date: form.value.start_date || undefined,
      end_date: form.value.end_date || undefined,
      target_date: form.value.target_date || undefined,
    });
    showCreate.value = false;
    form.value = { name: "", semver: "", description: "", start_date: "", end_date: "", target_date: "" };
    versions.value = [v, ...versions.value];
    total.value++;
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    creating.value = false;
  }
}

async function deleteVersion(v: Version) {
  if (!confirm(`确定删除版本「${v.name}」吗？此操作不可撤销。`)) return;
  deletingId.value = v.id;
  try {
    const wsId = await resolveWsId();
    await versionApi.deleteVersion(wsId, projectId.value, v.id);
    versions.value = versions.value.filter((x) => x.id !== v.id);
    total.value--;
  } catch (e: unknown) {
    alert(e instanceof Error ? e.message : "删除失败");
  } finally {
    deletingId.value = null;
  }
}

function openDetail(v: Version) {
  router.push({
    name: "version-detail",
    params: { versionId: v.id },
  });
}

function openDelivery(v: Version) {
  router.push({
    name: "version-delivery-report",
    params: { versionId: v.id },
  });
}

function setFilter(s: VersionStatus | "") {
  filterStatus.value = s;
  load();
}

/* ---------- lifecycle ---------- */

onMounted(load);
</script>

<template>
  <div class="version-list">
    <!-- Header -->
    <header class="header">
      <div>
        <h1 class="header__title">版本</h1>
        <p class="header__hint">规划与追踪版本发布节奏</p>
      </div>
      <div class="header__actions">
        <AppButton
          variant="primary"
          size="sm"
          @click="showCreate = true"
        >
          + 新建版本
        </AppButton>
      </div>
    </header>

    <!-- Create panel -->
    <div v-if="showCreate" class="create-panel" @keydown.escape="showCreate = false">
      <div class="create-panel__inner">
        <h3 class="create-panel__title">新建版本</h3>
        <form class="create-form" @submit.prevent="createVersion">
          <div class="create-form__row">
            <label class="create-form__field">
              <span class="create-form__label">
                版本名称 <span class="create-form__required">*</span>
              </span>
              <input
                v-model="form.name"
                placeholder="例如：v1.0 正式版"
                maxlength="120"
                class="create-form__input"
                autofocus
              />
            </label>
            <label class="create-form__field">
              <span class="create-form__label">
                语义版本号 <span class="create-form__required">*</span>
              </span>
              <input
                v-model="form.semver"
                placeholder="例如：1.0.0"
                maxlength="50"
                class="create-form__input create-form__input--mono"
              />
            </label>
          </div>
          <label class="create-form__field">
            <span class="create-form__label">目标日期</span>
            <input
              v-model="form.target_date"
              type="date"
              class="create-form__input"
            />
          </label>
          <div class="create-form__row">
            <label class="create-form__field">
              <span class="create-form__label">开始时间</span>
              <input
                v-model="form.start_date"
                type="date"
                class="create-form__input"
              />
            </label>
            <label class="create-form__field">
              <span class="create-form__label">结束时间</span>
              <input
                v-model="form.end_date"
                type="date"
                class="create-form__input"
              />
            </label>
          </div>
          <label class="create-form__field">
            <span class="create-form__label">描述（可选）</span>
            <textarea
              v-model="form.description"
              placeholder="版本目标与范围简述"
              maxlength="2000"
              rows="2"
              class="create-form__input"
            ></textarea>
          </label>
          <div v-if="createError" class="create-form__error">{{ createError }}</div>
          <div class="create-form__footer">
            <AppButton variant="secondary" size="sm" @click="showCreate = false">
              取消
            </AppButton>
            <AppButton variant="primary" size="sm" type="submit" :loading="creating">
              创建
            </AppButton>
          </div>
        </form>
      </div>
    </div>

    <!-- Status tabs -->
    <nav class="tabs">
      <button
        class="tabs__tab"
        :class="{ 'tabs__tab--active': filterStatus === '' }"
        @click="setFilter('')"
      >
        全部
        <span class="tabs__count">{{ total }}</span>
      </button>
      <button
        v-for="s in STATUS_GROUPS"
        :key="s"
        class="tabs__tab"
        :class="{ 'tabs__tab--active': filterStatus === s }"
        @click="setFilter(s)"
      >
        {{ statusLabel[s] }}
      </button>
    </nav>

    <!-- Loading -->
    <AppSkeleton v-if="loading" variant="table" :rows="6" />

    <!-- Error -->
    <AppErrorState
      v-else-if="error"
      :message="error"
      @retry="load"
    />

    <!-- Empty -->
    <AppEmptyState
      v-else-if="versions.length === 0"
      icon="📦"
      :title="filterStatus ? `暂无「${statusLabel[filterStatus as VersionStatus]}」状态的版本` : '暂无版本'"
      description="点击「新建版本」开始规划第一个版本"
    >
      <AppButton variant="primary" size="sm" @click="showCreate = true">
        新建版本
      </AppButton>
    </AppEmptyState>

    <!-- Version cards -->
    <div v-else class="cards">
      <article
        v-for="v in versions"
        :key="v.id"
        class="card"
        tabindex="0"
        role="button"
        @click="openDetail(v)"
        @keydown.enter="openDetail(v)"
      >
        <div class="card__top">
          <div class="card__title-row">
            <h3 class="card__name">{{ v.name }}</h3>
            <span class="card__semver">{{ v.semver }}</span>
          </div>
          <AppBadge :variant="statusBadgeVariant[v.status]">
            {{ statusLabel[v.status] }}
          </AppBadge>
        </div>

        <!-- Progress -->
        <div v-if="v.progress" class="card__progress">
          <ProgressBar
            :percent="Math.round((v.progress.completion_rate ?? 0) * 100)"
            size="sm"
            :color="(v.progress.completion_rate ?? 0) >= 0.8 ? 'var(--success-500)' : 'var(--warning-500)'"
            :label="`${v.progress.done_points}/${v.progress.total_points} 故事点  ·  ${v.progress.done_issues}/${v.progress.total_issues} 工作项`"
          />
        </div>

        <!-- Meta row -->
        <div class="card__meta">
          <span v-if="v.start_date || v.end_date" class="card__date">
            📅 {{ v.start_date }} ~ {{ v.end_date }}
          </span>
          <span v-if="v.target_date" class="card__date">
            🎯 目标 {{ v.target_date }}
          </span>
          <span class="card__sprint-count">
            {{ v.sprints?.length ?? 0 }} 个迭代
          </span>
          <span
v-if="v.quality" class="card__quality" :class="{
            'card__quality--pass': v.quality.pass_quality_gate,
            'card__quality--fail': !v.quality.pass_quality_gate,
          }"
>
            {{ v.quality.pass_quality_gate ? '✓ 质量通过' : '✗ 质量未通过' }}
          </span>
        </div>

        <!-- Actions -->
        <div class="card__actions" @click.stop>
          <button
            v-if="v.status === 'released' || v.status === 'archived'"
            class="card__action"
            @click="openDelivery(v)"
          >
            交付报告
          </button>
          <button
            v-if="v.status === 'planning' || v.status === 'archived'"
            class="card__action card__action--danger"
            :disabled="deletingId === v.id"
            @click="deleteVersion(v)"
          >
            {{ deletingId === v.id ? '删除中...' : '删除' }}
          </button>
        </div>
      </article>
    </div>
  </div>
</template>

<style scoped>
.version-list {
  max-width: 960px;
}

/* ---- header ---- */
.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 12px;
  flex-wrap: wrap;
}
.header__title { margin: 0; font-size: 20px; font-weight: 600; }
.header__hint { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }
.header__actions { display: flex; gap: 8px; }

/* ---- create panel ---- */
.create-panel {
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  box-shadow: var(--shadow-popover);
}
.create-panel__inner { padding: 16px; }
.create-panel__title { margin: 0 0 12px; font-size: 14px; font-weight: 600; }

.create-form { display: flex; flex-direction: column; gap: 10px; }
.create-form__row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.create-form__field { display: flex; flex-direction: column; gap: 4px; }
.create-form__label { font-size: 12px; color: var(--text-secondary); font-weight: 500; }
.create-form__required { color: var(--danger-500); }
.create-form__input {
  font-size: 13px;
  padding: 7px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-family: inherit;
  transition: border-color 0.15s;
}
.create-form__input:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}
.create-form__input--mono { font-family: var(--font-mono); }
textarea.create-form__input { resize: vertical; min-height: 56px; }
.create-form__error { font-size: 12px; color: var(--danger-500); }
.create-form__footer { display: flex; justify-content: flex-end; gap: 8px; }

/* ---- tabs ---- */
.tabs {
  display: flex;
  gap: 0;
  margin-bottom: 16px;
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
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.tabs__tab:hover { color: var(--text-primary); }
.tabs__tab--active {
  color: var(--brand-500);
  border-bottom-color: var(--brand-500);
}
.tabs__count {
  font-size: 11px;
  background: var(--surface-3);
  color: var(--text-tertiary);
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 500;
}

/* ---- cards ---- */
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}

.card {
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: box-shadow 0.15s, transform 0.1s, border-color 0.15s;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.card:hover {
  box-shadow: var(--shadow-card);
  border-color: var(--border-default);
  transform: translateY(-1px);
}
.card:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 2px;
}

.card__top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}
.card__title-row { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.card__name { margin: 0; font-size: 14px; font-weight: 600; line-height: 1.3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card__semver {
  font-size: 12px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

.card__progress { min-height: 24px; }

.card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 11px;
  color: var(--text-tertiary);
}
.card__date {
  font-family: var(--font-mono);
}
.card__quality--pass { color: var(--success-500); }
.card__quality--fail { color: var(--danger-500); }

.card__actions {
  display: flex;
  gap: 8px;
  border-top: 1px solid var(--border-subtle);
  padding-top: 8px;
  margin-top: 2px;
}
.card__action {
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.15s;
}
.card__action:hover { background: var(--surface-2); }
.card__action--danger { color: var(--danger-500); }
.card__action--danger:hover { background: rgba(220, 47, 47, 0.06); }
.card__action:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
