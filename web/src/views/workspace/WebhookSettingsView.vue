<script setup lang="ts">
/**
 * WebhookSettingsView — 工作空间级 Webhook 订阅管理。
 *
 * 能力：
 *  - 列表展示（分页、搜索、状态筛选）
 *  - 创建（含 HMAC Secret 一次性展示）
 *  - 编辑、删除
 *  - 暂停/恢复/测试推送
 *  - 投递日志查看（最近 50 条）
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import {
  webhookApi,
  type Webhook,
  type WebhookLog,
  type WebhookEvent,
  type CreateWebhookInput,
} from "@/api/services/webhook";
import { workspaceApi } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { AppLoadingState, AppErrorState } from "@/components";

/* ------------------------------------------------------------------ */
/* Constants                                                          */
/* ------------------------------------------------------------------ */

const EVENT_OPTIONS: { value: WebhookEvent; label: string }[] = [
  { value: "issue.created", label: "工作项创建" },
  { value: "issue.updated", label: "工作项更新" },
  { value: "issue.deleted", label: "工作项删除" },
  { value: "issue.status_changed", label: "工作项状态变更" },
  { value: "issue.commented", label: "工作项评论" },
  { value: "issue.assigned", label: "工作项分配" },
  { value: "sprint.started", label: "迭代开始" },
  { value: "sprint.completed", label: "迭代完成" },
  { value: "version.released", label: "版本发布" },
  { value: "version.created", label: "版本创建" },
  { value: "project.created", label: "项目创建" },
  { value: "member.joined", label: "成员加入" },
  { value: "member.removed", label: "成员移除" },
  { value: "comment.created", label: "评论创建" },
  { value: "attachment.uploaded", label: "附件上传" },
  { value: "intake.submitted", label: "收件提交" },
  { value: "automation.triggered", label: "自动化触发" },
];

/* ------------------------------------------------------------------ */
/* State                                                              */
/* ------------------------------------------------------------------ */

const route = useRoute();
const workspaceSlug = computed(() => String(route.params.workspaceSlug));

const loading = ref(true);
const error = ref("");

const wsId = ref(0);
const hooks = ref<Webhook[]>([]);
const total = ref(0);

/* ------------------------------------------------------------------ */
/* Create / Edit form                                                 */
/* ------------------------------------------------------------------ */

const showForm = ref(false);
const editingHook = ref<Webhook | null>(null);
const formSaving = ref(false);
const formError = ref("");
const form = reactive({
  name: "",
  url: "",
  events: [] as WebhookEvent[],
  description: "",
  project_id: null as number | null,
});

/** 创建后一次性展示的 secret */
const createdSecret = ref("");
const secretCopied = ref(false);

/* ------------------------------------------------------------------ */
/* Log viewer                                                         */
/* ------------------------------------------------------------------ */

const logHookId = ref(0);
const logs = ref<WebhookLog[]>([]);
const logsLoading = ref(false);

/* ------------------------------------------------------------------ */
/* Actions                                                            */
/* ------------------------------------------------------------------ */

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.getBySlug(workspaceSlug.value);
    wsId.value = ws.id;
    const res = await webhookApi.list(ws.id, { limit: 100 });
    hooks.value = res.items;
    total.value = res.total;
  } catch (e: any) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingHook.value = null;
  form.name = "";
  form.url = "";
  form.events = ["issue.created", "issue.updated", "issue.status_changed"];
  form.description = "";
  form.project_id = null;
  createdSecret.value = "";
  secretCopied.value = false;
  formError.value = "";
  showForm.value = true;
}

function openEdit(hook: Webhook) {
  editingHook.value = hook;
  form.name = hook.name;
  form.url = hook.url;
  form.events = [...hook.events];
  form.description = hook.description ?? "";
  form.project_id = hook.project_id ?? null;
  createdSecret.value = "";
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingHook.value = null;
}

function toggleEvent(ev: WebhookEvent) {
  const idx = form.events.indexOf(ev);
  if (idx >= 0) form.events.splice(idx, 1);
  else form.events.push(ev);
}

async function saveForm() {
  formError.value = "";
  if (!form.name.trim()) { formError.value = "名称不能为空"; return; }
  if (!form.url.trim()) { formError.value = "URL 不能为空"; return; }
  if (form.events.length === 0) { formError.value = "至少选择一个事件"; return; }

  formSaving.value = true;
  try {
    if (editingHook.value) {
      // Update
      await webhookApi.update(wsId.value, editingHook.value.id, {
        name: form.name.trim(),
        url: form.url.trim(),
        events: form.events,
        description: form.description || undefined,
      });
    } else {
      // Create
      const input: CreateWebhookInput = {
        name: form.name.trim(),
        url: form.url.trim(),
        events: form.events,
        description: form.description || undefined,
        project_id: form.project_id ?? undefined,
      };
      const created = await webhookApi.create(wsId.value, input);
      createdSecret.value = created.secret;
    }
    await load();
    if (!editingHook.value) {
      // 保持表单展示 secret
      editingHook.value = null;
    } else {
      closeForm();
    }
  } catch (e: any) {
    formError.value = e instanceof ApiError ? e.message : "保存失败";
  } finally {
    formSaving.value = false;
  }
}

async function copySecret() {
  try {
    await navigator.clipboard.writeText(createdSecret.value);
    secretCopied.value = true;
    setTimeout(() => (secretCopied.value = false), 2000);
  } catch { /* noop */ }
}

async function removeHook(hook: Webhook) {
  if (!confirm(`确定删除 Webhook「${hook.name}」吗？`)) return;
  try {
    await webhookApi.delete(wsId.value, hook.id);
    hooks.value = hooks.value.filter((h) => h.id !== hook.id);
  } catch (e: any) {
    alert(`删除失败：${e instanceof ApiError ? e.message : e}`);
  }
}

async function toggleHook(hook: Webhook) {
  try {
    if (hook.is_active) {
      await webhookApi.pause(wsId.value, hook.id);
    } else {
      await webhookApi.resume(wsId.value, hook.id);
    }
    await load();
  } catch (e: any) {
    alert(`操作失败：${e instanceof ApiError ? e.message : e}`);
  }
}

async function testHook(hook: Webhook) {
  try {
    const res = await webhookApi.test(wsId.value, hook.id);
    alert(res.success ? "测试推送成功" : `测试失败：${res.error ?? "未知错误"}`);
  } catch (e: any) {
    alert(`测试失败：${e instanceof ApiError ? e.message : e}`);
  }
}

async function viewLogs(hook: Webhook) {
  logHookId.value = hook.id;
  logsLoading.value = true;
  try {
    const res = await webhookApi.listLogs(wsId.value, { webhook_id: hook.id, limit: 50 });
    logs.value = res.items;
  } catch (e: any) {
    alert(`加载日志失败：${e instanceof ApiError ? e.message : e}`);
  } finally {
    logsLoading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

function eventLabel(ev: string): string {
  return EVENT_OPTIONS.find((o) => o.value === ev)?.label ?? ev;
}

function statusBadge(hook: Webhook): { text: string; cls: string } {
  if (!hook.is_active) return { text: "已暂停", cls: "paused" };
  if (hook.status === "unhealthy") return { text: "异常", cls: "unhealthy" };
  return { text: "运行中", cls: "active" };
}

function formatTime(s?: string | null): string {
  return s ? s.replace("T", " ").slice(0, 19) : "-";
}

onMounted(load);
</script>

<template>
  <AppLoadingState v-if="loading" text="正在加载 Webhook 配置..." />
  <AppErrorState v-else-if="error" :message="error" @retry="load" />

  <div v-else class="wh-view">
    <header class="wh-header">
      <div>
        <h1>Webhook 管理</h1>
        <p class="wh-sub">配置事件推送，将平台事件实时同步到外部系统</p>
      </div>
      <button class="btn btn--primary" @click="openCreate">＋ 新建 Webhook</button>
    </header>

    <!-- ===== 列表 ===== -->
    <div v-if="hooks.length" class="wh-list">
      <div v-for="h in hooks" :key="h.id" class="wh-card">
        <div class="wh-card__main">
          <div class="wh-card__head">
            <strong class="wh-card__name">{{ h.name }}</strong>
            <span class="wh-badge" :class="statusBadge(h).cls">
              {{ statusBadge(h).text }}
            </span>
          </div>
          <p class="wh-card__url mono">{{ h.url }}</p>
          <div class="wh-card__events">
            <span v-for="ev in h.events.slice(0, 4)" :key="ev" class="ev-tag">
              {{ eventLabel(ev) }}
            </span>
            <span v-if="h.events.length > 4" class="ev-tag ev-tag--more">
              +{{ h.events.length - 4 }}
            </span>
          </div>
          <div class="wh-card__meta">
            <span>失败 {{ h.failure_count }} 次</span>
            <span v-if="h.last_triggered_at">上次触发 {{ formatTime(h.last_triggered_at) }}</span>
          </div>
        </div>
        <div class="wh-card__actions">
          <button class="btn-sm" @click="toggleHook(h)">
            {{ h.is_active ? "暂停" : "恢复" }}
          </button>
          <button class="btn-sm" @click="testHook(h)" title="发送测试事件">测试</button>
          <button class="btn-sm" @click="viewLogs(h)">日志</button>
          <button class="btn-sm" @click="openEdit(h)">编辑</button>
          <button class="btn-sm btn-sm--danger" @click="removeHook(h)">删除</button>
        </div>
      </div>
    </div>
    <p v-else class="wh-empty">暂无 Webhook 订阅。点击「新建 Webhook」创建第一个。</p>

    <!-- ===== 创建/编辑表单弹层 ===== -->
    <div v-if="showForm" class="wh-overlay" @click.self="closeForm">
      <div class="wh-modal">
        <h2>{{ editingHook ? "编辑 Webhook" : "新建 Webhook" }}</h2>

        <!-- 一次性 secret 展示 -->
        <div v-if="createdSecret" class="wh-secret-reveal">
          <p class="wh-secret-warn">
            ⚠️ 密钥仅显示一次，请立即复制保存。后续将无法再次查看。
          </p>
          <div class="wh-secret-row">
            <code class="mono wh-secret-val">{{ createdSecret }}</code>
            <button class="btn btn--primary" @click="copySecret">
              {{ secretCopied ? "已复制 ✓" : "复制" }}
            </button>
          </div>
          <button class="btn" style="margin-top:10px" @click="closeForm">关闭</button>
        </div>

        <template v-else>
          <label class="wh-field">
            <span class="wh-label">名称 <i>*</i></span>
            <input v-model="form.name" class="wh-input" maxlength="128" placeholder="如：CI 构建通知" />
          </label>

          <label class="wh-field">
            <span class="wh-label">目标 URL <i>*</i></span>
            <input v-model="form.url" class="wh-input" placeholder="https://example.com/webhook" />
          </label>

          <label class="wh-field">
            <span class="wh-label">描述</span>
            <input v-model="form.description" class="wh-input" maxlength="500" placeholder="可选" />
          </label>

          <fieldset class="wh-field">
            <legend class="wh-label">订阅事件 <i>*</i>（{{ form.events.length }} 个）</legend>
            <div class="wh-event-grid">
              <label
                v-for="o in EVENT_OPTIONS"
                :key="o.value"
                class="wh-event-opt"
                :class="{ checked: form.events.includes(o.value) }"
              >
                <input
                  type="checkbox"
                  :checked="form.events.includes(o.value)"
                  @change="toggleEvent(o.value)"
                />
                {{ o.label }}
              </label>
            </div>
          </fieldset>

          <p v-if="formError" class="wh-err">{{ formError }}</p>

          <div class="wh-modal-actions">
            <button class="btn btn--primary" :disabled="formSaving" @click="saveForm">
              {{ formSaving ? "保存中..." : editingHook ? "保存" : "创建" }}
            </button>
            <button class="btn" @click="closeForm">取消</button>
          </div>
        </template>
      </div>
    </div>

    <!-- ===== 日志查看器弹层 ===== -->
    <div v-if="logHookId" class="wh-overlay" @click.self="logHookId = 0">
      <div class="wh-modal wh-modal--lg">
        <h2>投递日志</h2>
        <AppLoadingState v-if="logsLoading" text="加载日志..." />
        <table v-else-if="logs.length" class="wh-log-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>事件</th>
              <th>状态码</th>
              <th>结果</th>
              <th>耗时</th>
              <th>错误</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="l in logs" :key="l.id">
              <td class="meta">{{ formatTime(l.created_at) }}</td>
              <td><span class="ev-tag">{{ l.event }}</span></td>
              <td>{{ l.status_code ?? "-" }}</td>
              <td>
                <span class="wh-badge" :class="l.success ? 'active' : 'unhealthy'">
                  {{ l.success ? "成功" : "失败" }}
                </span>
              </td>
              <td class="meta">{{ l.duration_ms ? `${l.duration_ms}ms` : "-" }}</td>
              <td class="wh-log-err">{{ l.error_message ?? "-" }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="wh-empty">暂无投递日志</p>
        <button class="btn" style="margin-top:12px" @click="logHookId = 0">关闭</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wh-view { max-width: 880px; }

.wh-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.wh-header h1 { margin: 0; font-size: 20px; }
.wh-sub { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }

.wh-list { display: flex; flex-direction: column; gap: 12px; }

.wh-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.wh-card__main { flex: 1; min-width: 0; }
.wh-card__head { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.wh-card__name { font-size: 14px; }

.wh-card__url {
  font-size: 12px;
  color: var(--text-tertiary);
  word-break: break-all;
  margin: 0 0 8px;
}

.wh-card__events { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 8px; }
.wh-card__meta {
  font-size: 11px;
  color: var(--text-tertiary);
  display: flex;
  gap: 16px;
}

.wh-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-left: 16px;
  flex-shrink: 0;
}

/* Badge */
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

/* Event tag */
.ev-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  background: var(--surface-3);
  color: var(--text-secondary);
}

.ev-tag--more { background: none; color: var(--text-tertiary); }

.wh-empty { color: var(--text-tertiary); font-size: 13px; padding: 32px 0; }

/* Overlay / Modal */
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

.wh-modal--lg { width: 720px; }
.wh-modal h2 { margin: 0 0 16px; font-size: 16px; }

/* Fields */
.wh-field { display: block; margin-bottom: 14px; }
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

/* Event grid */
.wh-event-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 4px 12px;
  max-height: 200px;
  overflow-y: auto;
  padding: 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.wh-event-opt {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 3px 4px;
  border-radius: 3px;
}

.wh-event-opt:hover { background: var(--surface-3); }
.wh-event-opt.checked { color: var(--brand-600); font-weight: 500; }
.wh-event-opt input { accent-color: var(--brand-500); }

.wh-err { color: var(--danger-500); font-size: 12px; margin: 8px 0 0; }

.wh-modal-actions { display: flex; gap: 8px; margin-top: 16px; }

/* Secret reveal */
.wh-secret-reveal {
  padding: 14px;
  border: 1px solid var(--warning-500,#f59e0b);
  border-radius: var(--radius-md);
  background: rgba(245,158,11,0.08);
  margin-bottom: 12px;
}

.wh-secret-warn {
  font-size: 12px;
  color: var(--warning-500,#f59e0b);
  font-weight: 500;
  margin: 0 0 10px;
}

.wh-secret-row { display: flex; gap: 8px; align-items: center; }
.wh-secret-val {
  flex: 1;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
  font-size: 12px;
  word-break: break-all;
  color: var(--text-primary);
}

/* Log table */
.wh-log-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.wh-log-table th { text-align: left; padding: 6px 8px; color: var(--text-tertiary); font-weight: 500; border-bottom: 1px solid var(--border-subtle); }
.wh-log-table td { padding: 6px 8px; border-bottom: 1px solid var(--border-subtle); color: var(--text-primary); }
.wh-log-err { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--danger-500); }

/* Buttons */
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

.mono { font-family: var(--font-mono); }
.meta { color: var(--text-tertiary); font-size: 12px; }
</style>
