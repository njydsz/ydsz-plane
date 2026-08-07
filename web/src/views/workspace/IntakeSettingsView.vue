<script setup lang="ts">
/**
 * IntakeSettingsView — 收件箱通道管理与工单审核。
 *
 * 能力：
 *  - 通道 CRUD（名称、Slug、默认类型、公开开关、限流等）
 *  - 工单审核队列（列表、审批/拒绝/归档、转正为需求/缺陷）
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import { intakeApi, type IntakeChannel, type IntakeIssue, type IssueType } from "@/api/services/intake";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { AppLoadingState, AppErrorState } from "@/components";

const route = useRoute();
const workspaceSlug = computed(() => String(route.params.workspaceSlug));

const loading = ref(true);
const error = ref("");

const wsId = ref(0);
const projects = ref<Project[]>([]);

/* ------------------------------------------------------------------ */
/* Tabs                                                               */
/* ------------------------------------------------------------------ */

type Tab = "channels" | "queue";
const tab = ref<Tab>("channels");

/* ------------------------------------------------------------------ */
/* Channels                                                           */
/* ------------------------------------------------------------------ */

const channels = ref<IntakeChannel[]>([]);
const showChannelForm = ref(false);
const editingChannel = ref<IntakeChannel | null>(null);
const channelSaving = ref(false);
const channelError = ref("");

const chForm = reactive({
  slug: "",
  name: "",
  description: "",
  is_public: true,
  default_issue_type: "bug" as IssueType,
  default_priority: 3,
  rate_limit_per_min: 10,
  project_id: null as number | null,
});

function openChannelCreate() {
  editingChannel.value = null;
  chForm.slug = "";
  chForm.name = "";
  chForm.description = "";
  chForm.is_public = true;
  chForm.default_issue_type = "bug";
  chForm.default_priority = 3;
  chForm.rate_limit_per_min = 10;
  chForm.project_id = null;
  channelError.value = "";
  showChannelForm.value = true;
}

function openChannelEdit(ch: IntakeChannel) {
  editingChannel.value = ch;
  chForm.slug = ch.slug;
  chForm.name = ch.name;
  chForm.description = ch.description;
  chForm.is_public = ch.is_public;
  chForm.default_issue_type = ch.default_issue_type;
  chForm.default_priority = ch.default_priority;
  chForm.rate_limit_per_min = ch.rate_limit_per_min;
  chForm.project_id = ch.project_id ?? null;
  channelError.value = "";
  showChannelForm.value = true;
}

function closeChannelForm() {
  showChannelForm.value = false;
  editingChannel.value = null;
}

async function saveChannel() {
  channelError.value = "";
  if (!chForm.slug.trim() || !chForm.name.trim()) {
    channelError.value = "Slug 和名称不能为空";
    return;
  }

  channelSaving.value = true;
  try {
    if (editingChannel.value) {
      await intakeApi.updateChannel(wsId.value, editingChannel.value.id, {
        name: chForm.name.trim(),
        description: chForm.description,
        is_public: chForm.is_public,
      });
    } else {
      await intakeApi.createChannel(wsId.value, {
        slug: chForm.slug.trim(),
        name: chForm.name.trim(),
        description: chForm.description || undefined,
        is_public: chForm.is_public,
        default_issue_type: chForm.default_issue_type,
        default_priority: chForm.default_priority,
        rate_limit_per_min: chForm.rate_limit_per_min,
        project_id: chForm.project_id ?? undefined,
      });
    }
    await loadChannels();
    closeChannelForm();
  } catch (e: any) {
    channelError.value = e instanceof ApiError ? e.message : "保存失败";
  } finally {
    channelSaving.value = false;
  }
}

async function removeChannel(ch: IntakeChannel) {
  if (!confirm(`确定删除通道「${ch.name}」吗？`)) return;
  try {
    await intakeApi.deleteChannel(wsId.value, ch.id);
    channels.value = channels.value.filter((c) => c.id !== ch.id);
  } catch (e: any) {
    alert(`删除失败：${e instanceof ApiError ? e.message : e}`);
  }
}

async function loadChannels() {
  try {
    const res = await intakeApi.listChannels(wsId.value, { limit: 100 });
    channels.value = res.items;
  } catch { /* handled by parent load */ }
}

/* ------------------------------------------------------------------ */
/* Review Queue                                                       */
/* ------------------------------------------------------------------ */

const queueItems = ref<IntakeIssue[]>([]);
const queueTotal = ref(0);
const queueLoading = ref(false);
const queueFilter = ref<"pending" | "reviewed" | "">("pending");

const reviewForm = reactive({
  action: "approve" as "approve" | "reject" | "archive",
  target_issue_type: "bug",
  target_project_id: 0,
  reason: "",
});
const reviewSaving = ref(false);
const reviewingId = ref(0);

async function loadQueue() {
  queueLoading.value = true;
  try {
    const params: any = { limit: 50 };
    if (queueFilter.value) params.status = queueFilter.value;
    const res = await intakeApi.listIssues(wsId.value, params);
    queueItems.value = res.items;
    queueTotal.value = res.total;
  } catch (e: any) {
    /* silent */
  } finally {
    queueLoading.value = false;
  }
}

function openReview(issue: IntakeIssue) {
  reviewingId.value = issue.id;
  reviewForm.action = "approve";
  reviewForm.target_issue_type = issue.issue_type;
  reviewForm.target_project_id = 0;
  reviewForm.reason = "";
}

async function doReview() {
  reviewSaving.value = true;
  try {
    if (reviewForm.action === "approve") {
      // 先审核再转正
      await intakeApi.reviewIssue(wsId.value, reviewingId.value, {
        action: "approve",
        target_issue_type: reviewForm.target_issue_type,
        target_project_id: reviewForm.target_project_id || undefined,
        reason: reviewForm.reason || undefined,
      });
      // 转正
      if (reviewForm.target_project_id) {
        await intakeApi.convertIssue(wsId.value, reviewingId.value, {
          target_project_id: reviewForm.target_project_id,
          target_issue_type: reviewForm.target_issue_type,
        });
      }
    } else {
      await intakeApi.reviewIssue(wsId.value, reviewingId.value, {
        action: reviewForm.action,
        reason: reviewForm.reason || undefined,
      });
    }
    reviewingId.value = 0;
    await loadQueue();
  } catch (e: any) {
    alert(`操作失败：${e instanceof ApiError ? e.message : e}`);
  } finally {
    reviewSaving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

function statusLabel(s: string): string {
  const map: Record<string, string> = {
    pending: "待审核", reviewed: "已审核", converted: "已转正",
    rejected: "已拒绝", archived: "已归档",
  };
  return map[s] ?? s;
}

function statusClass(s: string): string {
  const map: Record<string, string> = {
    pending: "pending", reviewed: "active", converted: "active",
    rejected: "unhealthy", archived: "paused",
  };
  return map[s] ?? "";
}

function publicUrl(ch: IntakeChannel): string {
  return `${window.location.origin}/api/v1/public/intake/${wsId.value}/${ch.slug}`;
}

/* ------------------------------------------------------------------ */
/* Load                                                               */
/* ------------------------------------------------------------------ */

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.getBySlug(workspaceSlug.value);
    wsId.value = ws.id;
    const [chRes, projRes] = await Promise.all([
      intakeApi.listChannels(ws.id, { limit: 100 }),
      workspaceApi.listProjects(ws.id),
    ]);
    channels.value = chRes.items;
    projects.value = projRes;
  } catch (e: any) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <AppLoadingState v-if="loading" text="正在加载收件箱配置..." />
  <AppErrorState v-else-if="error" :message="error" @retry="load" />

  <div v-else class="in-view">
    <header class="in-header">
      <div>
        <h1>收件箱</h1>
        <p class="in-sub">管理外部反馈通道，审核并转正提交的工单</p>
      </div>
    </header>

    <nav class="tabs">
      <button :class="{ active: tab === 'channels' }" @click="tab = 'channels'">
        通道 ({{ channels.length }})
      </button>
      <button :class="{ active: tab === 'queue' }" @click="tab = 'queue'; loadQueue()">
        审核队列
      </button>
    </nav>

    <!-- ===== 通道管理 ===== -->
    <section v-if="tab === 'channels'">
      <div style="margin-bottom:16px">
        <button class="btn btn--primary" @click="openChannelCreate">＋ 新建通道</button>
      </div>

      <div v-if="channels.length" class="ch-list">
        <div v-for="ch in channels" :key="ch.id" class="ch-card">
          <div class="ch-card__main">
            <div class="ch-card__head">
              <strong>{{ ch.name }}</strong>
              <code class="mono">/{{ ch.slug }}</code>
              <span class="wh-badge" :class="ch.is_active ? 'active' : 'paused'">
                {{ ch.is_active ? "启用" : "禁用" }}
              </span>
              <span class="wh-badge" :class="ch.is_public ? 'active' : 'paused'">
                {{ ch.is_public ? "公开" : "内部" }}
              </span>
            </div>
            <p v-if="ch.description" class="ch-desc">{{ ch.description }}</p>
            <div class="ch-meta">
              <span>默认类型：{{ ch.default_issue_type === "bug" ? "缺陷" : ch.default_issue_type === "requirement" ? "需求" : "任务" }}</span>
              <span>限流：{{ ch.rate_limit_per_min }} 次/分钟</span>
              <span v-if="ch.is_public" class="ch-url mono">{{ publicUrl(ch) }}</span>
            </div>
          </div>
          <div class="ch-card__actions">
            <button class="btn-sm" @click="openChannelEdit(ch)">编辑</button>
            <button class="btn-sm btn-sm--danger" @click="removeChannel(ch)">删除</button>
          </div>
        </div>
      </div>
      <p v-else class="muted">暂无收件通道</p>
    </section>

    <!-- ===== 审核队列 ===== -->
    <section v-if="tab === 'queue'">
      <div class="queue-filters">
        <select v-model="queueFilter" @change="loadQueue" class="wh-input" style="width:160px">
          <option value="pending">待审核</option>
          <option value="reviewed">已审核</option>
          <option value="">全部</option>
        </select>
        <span class="meta">共 {{ queueTotal }} 条</span>
      </div>

      <table v-if="queueItems.length" class="member-table">
        <thead>
          <tr>
            <th>标题</th>
            <th>提交者</th>
            <th>类型</th>
            <th>状态</th>
            <th>时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="iss in queueItems" :key="iss.id">
            <td>{{ iss.title }}</td>
            <td>{{ iss.submitter_name }}<br><span class="meta">{{ iss.submitter_email }}</span></td>
            <td>
              <span class="ev-tag">{{ iss.issue_type === "bug" ? "缺陷" : iss.issue_type === "requirement" ? "需求" : "任务" }}</span>
            </td>
            <td><span class="wh-badge" :class="statusClass(iss.status)">{{ statusLabel(iss.status) }}</span></td>
            <td class="meta">{{ iss.created_at.slice(0, 10) }}</td>
            <td>
              <button
                v-if="iss.status === 'pending'"
                class="btn-sm"
                @click="openReview(iss)"
              >
                审核
              </button>
              <span v-else class="meta">-</span>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else-if="!queueLoading" class="muted">暂无工单</p>
    </section>

    <!-- ===== Channel Form Modal ===== -->
    <div v-if="showChannelForm" class="wh-overlay" @click.self="closeChannelForm">
      <div class="wh-modal">
        <h2>{{ editingChannel ? "编辑通道" : "新建通道" }}</h2>

        <label class="wh-field">
          <span class="wh-label">Slug <i>*</i></span>
          <input v-model="chForm.slug" class="wh-input" maxlength="64" placeholder="如：bug-report" :disabled="!!editingChannel" />
        </label>

        <label class="wh-field">
          <span class="wh-label">名称 <i>*</i></span>
          <input v-model="chForm.name" class="wh-input" maxlength="128" placeholder="如：Bug 反馈通道" />
        </label>

        <label class="wh-field">
          <span class="wh-label">描述</span>
          <textarea v-model="chForm.description" class="wh-input" rows="2" placeholder="可选" />
        </label>

        <label class="wh-field wh-field--row">
          <span class="wh-label">公开访问</span>
          <input type="checkbox" v-model="chForm.is_public" />
          <span class="meta">公开通道允许免登录提交</span>
        </label>

        <div class="wh-field-row">
          <label class="wh-field wh-field--half">
            <span class="wh-label">默认类型</span>
            <select v-model="chForm.default_issue_type" class="wh-input">
              <option value="bug">缺陷</option>
              <option value="requirement">需求</option>
              <option value="task">任务</option>
            </select>
          </label>

          <label class="wh-field wh-field--half">
            <span class="wh-label">默认优先级</span>
            <select v-model.number="chForm.default_priority" class="wh-input">
              <option :value="5">S5 致命</option>
              <option :value="4">S4 严重</option>
              <option :value="3">S3 一般</option>
              <option :value="2">S2 轻微</option>
              <option :value="1">S1 建议</option>
            </select>
          </label>
        </div>

        <label class="wh-field">
          <span class="wh-label">限流（次/分钟）</span>
          <input v-model.number="chForm.rate_limit_per_min" type="number" class="wh-input" min="1" max="300" />
        </label>

        <label class="wh-field">
          <span class="wh-label">关联项目</span>
          <select v-model.number="chForm.project_id" class="wh-input">
            <option :value="null">不关联</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </label>

        <p v-if="channelError" class="wh-err">{{ channelError }}</p>

        <div class="wh-modal-actions">
          <button class="btn btn--primary" :disabled="channelSaving" @click="saveChannel">
            {{ channelSaving ? "保存中..." : "保存" }}
          </button>
          <button class="btn" @click="closeChannelForm">取消</button>
        </div>
      </div>
    </div>

    <!-- ===== Review Modal ===== -->
    <div v-if="reviewingId" class="wh-overlay" @click.self="reviewingId = 0">
      <div class="wh-modal">
        <h2>审核工单</h2>

        <label class="wh-field">
          <span class="wh-label">操作</span>
          <select v-model="reviewForm.action" class="wh-input">
            <option value="approve">通过并转正</option>
            <option value="reject">拒绝</option>
            <option value="archive">归档</option>
          </select>
        </label>

        <template v-if="reviewForm.action === 'approve'">
          <label class="wh-field">
            <span class="wh-label">目标项目</span>
            <select v-model.number="reviewForm.target_project_id" class="wh-input">
              <option :value="0" disabled>请选择项目</option>
              <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
          </label>

          <label class="wh-field">
            <span class="wh-label">目标类型</span>
            <select v-model="reviewForm.target_issue_type" class="wh-input">
              <option value="bug">缺陷</option>
              <option value="requirement">需求</option>
              <option value="task">任务</option>
            </select>
          </label>
        </template>

        <label class="wh-field">
          <span class="wh-label">备注</span>
          <textarea v-model="reviewForm.reason" class="wh-input" rows="2" placeholder="审核意见（可选）" />
        </label>

        <div class="wh-modal-actions">
          <button class="btn btn--primary" :disabled="reviewSaving" @click="doReview">
            {{ reviewSaving ? "处理中..." : "确认" }}
          </button>
          <button class="btn" @click="reviewingId = 0">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.in-view { max-width: 880px; }

.in-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.in-header h1 { margin: 0; font-size: 20px; }
.in-sub { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }

.tabs {
  display: flex;
  gap: 2px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 24px;
}

.tabs button {
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-tertiary);
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
}

.tabs button.active {
  color: var(--brand-600);
  border-bottom-color: var(--brand-500);
  font-weight: 500;
}

/* Channel cards */
.ch-list { display: flex; flex-direction: column; gap: 12px; }

.ch-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 14px 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.ch-card__main { flex: 1; min-width: 0; }
.ch-card__head { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.ch-card__head code { font-size: 12px; color: var(--text-tertiary); }
.ch-desc { font-size: 13px; color: var(--text-secondary); margin: 0 0 6px; }
.ch-meta { display: flex; gap: 16px; font-size: 11px; color: var(--text-tertiary); }
.ch-url { font-size: 11px; word-break: break-all; }
.ch-card__actions { display: flex; gap: 4px; margin-left: 12px; flex-shrink: 0; }

/* Queue */
.queue-filters { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }

/* Shared styles from WebhookSettingsView */
.wh-overlay {
  position: fixed;
  inset: 0;
  z-index: 300;
  background: rgba(0,0,0,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
}

.wh-modal {
  background: var(--surface-1);
  border-radius: var(--radius-md);
  padding: 24px;
  width: 560px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-popover);
}

.wh-modal h2 { margin: 0 0 16px; font-size: 16px; }

.wh-field { display: block; margin-bottom: 14px; }
.wh-field--row { display: flex; align-items: center; gap: 8px; }
.wh-field--half { flex: 1; }
.wh-field-row { display: flex; gap: 12px; }

.wh-label { font-size: 13px; color: var(--text-secondary); display: block; margin-bottom: 4px; }
.wh-label i { color: var(--danger-500); font-style: normal; }

.wh-input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}

.wh-input:focus { border-color: var(--brand-500); box-shadow: 0 0 0 2px var(--brand-50); }
textarea.wh-input { resize: vertical; }

.wh-err { color: var(--danger-500); font-size: 12px; margin: 8px 0 0; }
.wh-modal-actions { display: flex; gap: 8px; margin-top: 16px; }

.wh-badge {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 500;
}

.wh-badge.active { background: rgba(15,194,123,0.12); color: var(--success-500,#0fc27b); }
.wh-badge.paused { background: var(--surface-3); color: var(--text-tertiary); }
.wh-badge.unhealthy { background: rgba(220,47,47,0.12); color: var(--danger-500,#dc2f2f); }
.wh-badge.pending { background: var(--brand-50); color: var(--brand-600); }

.ev-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  background: var(--surface-3);
  color: var(--text-secondary);
}

.btn {
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-sm {
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
}

.btn-sm:hover { background: var(--surface-3); }
.btn-sm--danger { color: var(--danger-500); }
.btn-sm--danger:hover { background: rgba(220,47,47,0.08); }

.member-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.member-table th {
  text-align: left;
  padding: 8px 12px;
  color: var(--text-tertiary);
  font-weight: 500;
  font-size: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.member-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.mono { font-family: var(--font-mono); }
.meta { color: var(--text-tertiary); font-size: 12px; }
.muted { color: var(--text-tertiary); font-size: 13px; padding: 24px 0; }
</style>
