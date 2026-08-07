<script setup lang="ts">
/**
 * AutomationView — 项目级自动化规则配置视图。
 *
 * 能力：
 *  - 规则列表展示（状态徽章 / 触发器类型 / 动作数 / 失败次数）
 *  - 创建 / 编辑表单（触发器 + 条件列表 + 动作列表）
 *  - 模板快速创建入口
 *  - Dry-Run 干跑测试
 *  - 规则执行历史查看
 *  - 启用 / 禁用 toggle
 */
import { computed, onMounted, reactive, ref } from "vue";

import {
  automationApi,
  type AutomationRule,
  type AutomationTemplate,
  type DryRunResult,
  type RuleDSL,
  type RuleDSLAction,
  type RuleExecution,
  type TriggerType,
} from "@/api/services/automation";
import { ApiError } from "@/api/client";
import { toast } from "@/lib/toast";
import { AppLoadingState, AppErrorState } from "@/components";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";

/* ------------------------------------------------------------------ */
/* Constants                                                          */
/* ------------------------------------------------------------------ */

const TRIGGER_OPTIONS: { value: TriggerType; label: string }[] = [
  { value: "issue.created", label: "工作项创建" },
  { value: "issue.updated", label: "工作项更新" },
  { value: "issue.status_changed", label: "状态变更" },
  { value: "issue.assigned", label: "已分配" },
  { value: "issue.commented", label: "已评论" },
  { value: "sprint.started", label: "迭代开始" },
  { value: "sprint.completed", label: "迭代完成" },
  { value: "version.released", label: "版本发布" },
  { value: "cron", label: "定时触发" },
];

const OPERATOR_OPTIONS = [
  { value: "eq", label: "等于 (=)" },
  { value: "neq", label: "不等于 (≠)" },
  { value: "gt", label: "大于 (>)" },
  { value: "lt", label: "小于 (<)" },
  { value: "contains", label: "包含 (contains)" },
];

const ACTION_OPTIONS = [
  { value: "transition_status", label: "变更状态" },
  { value: "assign", label: "分配负责人" },
  { value: "update_field", label: "更新字段" },
  { value: "send_notification", label: "发送通知" },
  { value: "create_issue", label: "创建工作项" },
  { value: "copy_field", label: "复制字段" },
];

const STATUS_BADGE: Record<string, { text: string; cls: string }> = {
  draft: { text: "草稿", cls: "au-badge--draft" },
  active: { text: "运行中", cls: "au-badge--active" },
  disabled: { text: "已停用", cls: "au-badge--disabled" },
  error: { text: "异常", cls: "au-badge--error" },
};

/* ------------------------------------------------------------------ */
/* Execution status map                                                */
/* ------------------------------------------------------------------ */
const EXECUTION_STATUS: Record<string, { text: string; cls: string }> = {
  success: { text: "成功", cls: "au-badge--active" },
  failed: { text: "失败", cls: "au-badge--error" },
  skipped: { text: "跳过", cls: "au-badge--draft" },
};

/* ------------------------------------------------------------------ */
/* Workspace context                                                   */
/* ------------------------------------------------------------------ */

const { wsId, projectId, ready, error: ctxError } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* List state                                                          */
/* ------------------------------------------------------------------ */

const loading = ref(false);
const loadError = ref("");
const rules = ref<AutomationRule[]>([]);
const templates = ref<AutomationTemplate[]>([]);

/* ------------------------------------------------------------------ */
/* Form state                                                          */
/* ------------------------------------------------------------------ */

const showForm = ref(false);
const editingRule = ref<AutomationRule | null>(null);
const formSaving = ref(false);
const formError = ref("");

interface FormData {
  name: string;
  description: string;
  triggerType: TriggerType;
  conditions: Array<{ field: string; operator: string; value: string }>;
  actions: Array<{ type: string; params: Record<string, string> }>;
}

const form = reactive<FormData>({
  name: "",
  description: "",
  triggerType: "issue.created",
  conditions: [],
  actions: [],
});

/* ------------------------------------------------------------------ */
/* Template dropdown                                                   */
/* ------------------------------------------------------------------ */

const showTemplateMenu = ref(false);
const templateLoading = ref(false);

/* ------------------------------------------------------------------ */
/* Execution history drawer                                            */
/* ------------------------------------------------------------------ */

const showHistory = ref(false);
const executions = ref<RuleExecution[]>([]);
const historyLoading = ref(false);
const historyRuleName = ref("");

/* ------------------------------------------------------------------ */
/* Dry-Run panel                                                       */
/* ------------------------------------------------------------------ */

const showDryRun = ref(false);
const dryRunLoading = ref(false);
const dryRunResult = ref<DryRunResult | null>(null);

/* ------------------------------------------------------------------ */
/* List actions                                                        */
/* ------------------------------------------------------------------ */

async function loadRules() {
  if (!ready.value) return;
  loading.value = true;
  loadError.value = "";
  try {
    const res = await automationApi.list(wsId.value, projectId.value, { limit: 100 });
    rules.value = res.items;
  } catch (e: any) {
    loadError.value = e instanceof ApiError ? e.message : "加载规则失败";
  } finally {
    loading.value = false;
  }
}

async function loadTemplates() {
  if (!ready.value) return;
  try {
    templates.value = await automationApi.listTemplates(wsId.value, projectId.value);
  } catch {
    templates.value = [];
  }
}

async function loadAll() {
  await Promise.all([loadRules(), loadTemplates()]);
}

async function removeRule(rule: AutomationRule) {
  if (!confirm(`确定删除规则「${rule.name}」吗？`)) return;
  try {
    await automationApi.delete(wsId.value, projectId.value, rule.id);
    rules.value = rules.value.filter((r) => r.id !== rule.id);
    toast.success("规则已删除");
  } catch (e: any) {
    toast.error(`删除失败：${e instanceof ApiError ? e.message : e}`);
  }
}

async function toggleRule(rule: AutomationRule) {
  const enable = rule.status === "disabled" || rule.status === "draft";
  try {
    const updated = await automationApi.toggle(wsId.value, projectId.value, rule.id, enable, rule.version);
    rules.value = rules.value.map((r) => (r.id === rule.id ? updated : r));
    toast.success(enable ? "规则已启用" : "规则已停用");
  } catch (e: any) {
    toast.error(`操作失败：${e instanceof ApiError ? e.message : e}`);
  }
}

/* ------------------------------------------------------------------ */
/* Form helpers                                                        */
/* ------------------------------------------------------------------ */

function triggerLabel(t: string): string {
  return TRIGGER_OPTIONS.find((o) => o.value === t)?.label ?? t;
}

function statusBadge(rule: AutomationRule): { text: string; cls: string } {
  return STATUS_BADGE[rule.status] ?? { text: rule.status, cls: "" };
}

function formatTime(s?: string | null): string {
  return s ? s.replace("T", " ").slice(0, 19) : "-";
}

function emptyDSL(): RuleDSL {
  return {
    trigger: { type: "issue.created" },
    conditions: [],
    actions: [],
  };
}

function buildDSL(): RuleDSL {
  const actions: RuleDSLAction[] = form.actions.map((a) => ({
    type: a.type,
    params: a.params,
  }));
  const conditions = form.conditions
    .filter((c) => c.field.trim())
    .map((c) => ({ field: c.field.trim(), operator: c.operator, value: c.value }));
  return {
    trigger: { type: form.triggerType },
    conditions: conditions.length > 0 ? conditions : undefined,
    actions,
  };
}

function openCreate() {
  editingRule.value = null;
  form.name = "";
  form.description = "";
  form.triggerType = "issue.created";
  form.conditions = [];
  form.actions = [];
  formError.value = "";
  showTemplateMenu.value = false;
  showForm.value = true;
}

function openEdit(rule: AutomationRule) {
  editingRule.value = rule;
  form.name = rule.name;
  form.description = rule.description ?? "";
  form.triggerType = rule.dsl.trigger.type;
  form.conditions = (rule.dsl.conditions ?? []).map((c) => ({
    field: c.field,
    operator: c.operator,
    value: String(c.value ?? ""),
  }));
  form.actions = rule.dsl.actions.map((a) => ({
    type: a.type,
    params: { ...(a.params ?? {}) } as Record<string, string>,
  }));
  formError.value = "";
  showForm.value = true;
}

function closeForm() {
  showForm.value = false;
  editingRule.value = null;
}

function addCondition() {
  form.conditions.push({ field: "", operator: "eq", value: "" });
}

function removeCondition(idx: number) {
  form.conditions.splice(idx, 1);
}

function addAction() {
  form.actions.push({ type: "transition_status", params: { target_state: "" } });
}

function removeAction(idx: number) {
  form.actions.splice(idx, 1);
}

function onActionTypeChange(idx: number) {
  const type = form.actions[idx].type;
  const defaults: Record<string, Record<string, string>> = {
    transition_status: { target_state: "" },
    assign: { assignee: "" },
    update_field: { field: "", value: "" },
    send_notification: { recipient: "", channel: "", message: "" },
    create_issue: { project_id: "", type: "" },
    copy_field: { source: "", target: "" },
  };
  form.actions[idx].params = defaults[type] ?? {};
}

/* ------------------------------------------------------------------ */
/* Dry-Run                                                             */
/* ------------------------------------------------------------------ */

async function runDryRun() {
  dryRunLoading.value = true;
  dryRunResult.value = null;
  showDryRun.value = true;
  try {
    const dsl = buildDSL();
    dryRunResult.value = await automationApi.dryRun(wsId.value, projectId.value, dsl);
    if (dryRunResult.value.valid) {
      toast.success("干跑校验通过");
    }
  } catch (e: any) {
    toast.error(`干跑失败：${e instanceof ApiError ? e.message : e}`);
  } finally {
    dryRunLoading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Save                                                                */
/* ------------------------------------------------------------------ */

async function saveForm(activate: boolean) {
  formError.value = "";
  if (!form.name.trim()) {
    formError.value = "规则名称不能为空";
    return;
  }
  if (form.actions.length === 0) {
    formError.value = "至少需要一个动作";
    return;
  }

  formSaving.value = true;
  try {
    const dsl = buildDSL();
    if (editingRule.value) {
      await automationApi.update(wsId.value, projectId.value, editingRule.value.id, {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        dsl,
        status: activate ? "active" : undefined,
        version: editingRule.value.version,
      });
      toast.success(activate ? "规则已更新并激活" : "规则已更新");
    } else {
      await automationApi.create(wsId.value, projectId.value, {
        name: form.name.trim(),
        description: form.description.trim() || undefined,
        dsl,
        status: activate ? "active" : "draft",
      });
      toast.success(activate ? "规则已创建并激活" : "草稿已保存");
    }
    closeForm();
    await loadRules();
  } catch (e: any) {
    formError.value = e instanceof ApiError ? e.message : "保存失败";
  } finally {
    formSaving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Templates                                                           */
/* ------------------------------------------------------------------ */

async function createFromTemplate(tpl: AutomationTemplate) {
  templateLoading.value = true;
  try {
    const rule = await automationApi.createFromTemplate(wsId.value, projectId.value, tpl.slug);
    toast.success("已从模板创建规则");
    showTemplateMenu.value = false;
    await loadRules();
    openEdit(rule);
  } catch (e: any) {
    toast.error(`创建失败：${e instanceof ApiError ? e.message : e}`);
  } finally {
    templateLoading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* Execution history                                                   */
/* ------------------------------------------------------------------ */

async function viewHistory(rule: AutomationRule) {
  historyRuleName.value = rule.name;
  showHistory.value = true;
  historyLoading.value = true;
  try {
    executions.value = await automationApi.listExecutions(wsId.value, projectId.value, 50);
  } catch (e: any) {
    toast.error(`加载历史失败：${e instanceof ApiError ? e.message : e}`);
    executions.value = [];
  } finally {
    historyLoading.value = false;
  }
}

function closeHistory() {
  showHistory.value = false;
  executions.value = [];
}

function execStatus(s: string): { text: string; cls: string } {
  return EXECUTION_STATUS[s] ?? { text: s, cls: "" };
}

function actionParamsSummary(type: string, params?: Record<string, any>): string {
  if (!params) return type;
  switch (type) {
    case "transition_status":
      return `→ ${params.target_state ?? "?"}`;
    case "assign":
      return `→ ${params.assignee ?? "?"}`;
    case "update_field":
      return `${params.field ?? "?"} = ${params.value ?? "?"}`;
    case "copy_field":
      return `${params.source ?? "?"} → ${params.target ?? "?"}`;
    default:
      return type;
  }
}

/* ------------------------------------------------------------------ */
/* Lifecycle                                                           */
/* ------------------------------------------------------------------ */

onMounted(() => {
  // useWorkspaceContext already resolves via watchEffect;
  // we just need to react to `ready` becoming true.
});

// Re-load when workspace context becomes ready
import { watch } from "vue";
watch(ready, (v) => {
  if (v) loadAll();
});
</script>

<template>
  <AppLoadingState v-if="!ready && !ctxError" text="正在加载工作空间上下文..." />
  <AppErrorState v-else-if="ctxError || loadError" :message="ctxError || loadError" @retry="loadAll" />

  <div v-else class="au-view">
    <!-- ===== Header ===== -->
    <header class="au-header">
      <div>
        <h1>自动化规则</h1>
        <p class="au-sub">配置事件驱动的工作项自动化，减少重复操作</p>
      </div>
      <div class="au-header__actions">
        <!-- Template dropdown -->
        <div class="au-dropdown-wrapper">
          <button class="btn btn--primary" @click="openCreate">
            ＋ 创建规则
          </button>
          <button
            class="btn btn--primary au-dropdown-toggle"
            :disabled="templateLoading"
            @click="showTemplateMenu = !showTemplateMenu"
            title="从模板创建"
          >
            <span class="au-caret" :class="{ open: showTemplateMenu }">▾</span>
          </button>
          <div v-if="showTemplateMenu" class="au-dropdown-menu" @mouseleave="showTemplateMenu = false">
            <div v-if="templates.length === 0" class="au-dropdown-empty">暂无模板</div>
            <button
              v-for="tpl in templates"
              :key="tpl.slug"
              class="au-dropdown-item"
              :disabled="templateLoading"
              @click="createFromTemplate(tpl)"
            >
              <span class="au-dropdown-item__icon">{{ tpl.icon ?? "📋" }}</span>
              <div class="au-dropdown-item__body">
                <strong>{{ tpl.name }}</strong>
                <span>{{ tpl.description }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- ===== Rule list ===== -->
    <AppLoadingState v-if="loading" text="正在加载规则..." />

    <div v-else-if="rules.length" class="au-list">
      <div v-for="r in rules" :key="r.id" class="au-card">
        <div class="au-card__main">
          <div class="au-card__head">
            <strong class="au-card__name">{{ r.name }}</strong>
            <span class="wh-badge" :class="statusBadge(r).cls">
              {{ statusBadge(r).text }}
              <span v-if="r.failure_count > 0" class="au-badge-count">{{ r.failure_count }}</span>
            </span>
          </div>
          <p v-if="r.description" class="au-card__desc">{{ r.description }}</p>
          <div class="au-card__tags">
            <span class="ev-tag">{{ triggerLabel(r.dsl.trigger.type) }}</span>
            <span class="au-au-count">
              {{ r.dsl.actions.length }} 个动作
            </span>
            <span v-if="r.last_triggered_at" class="au-meta">
              上次触发 {{ formatTime(r.last_triggered_at) }}
            </span>
            <span v-if="r.failure_count > 0" class="au-meta au-meta--err">
              失败 {{ r.failure_count }} 次
            </span>
          </div>
        </div>
        <div class="au-card__actions">
          <button class="btn-sm" @click="toggleRule(r)">
            {{ r.status === "disabled" || r.status === "draft" ? "启用" : "停用" }}
          </button>
          <button class="btn-sm" @click="viewHistory(r)">历史</button>
          <button class="btn-sm" @click="openEdit(r)">编辑</button>
          <button class="btn-sm btn-sm--danger" @click="removeRule(r)">删除</button>
        </div>
      </div>
    </div>

    <p v-else class="au-empty">暂无自动化规则。点击「创建规则」或从模板创建一个。</p>

    <!-- ===== Create / Edit form overlay ===== -->
    <div v-if="showForm" class="wh-overlay" @click.self="closeForm">
      <div class="wh-modal wh-modal--lg">
        <h2>{{ editingRule ? "编辑规则" : "创建规则" }}</h2>

        <label class="wh-field">
          <span class="wh-label">规则名称 <i>*</i></span>
          <input v-model="form.name" class="wh-input" maxlength="128" placeholder="如：分配负责人后自动变更状态" />
        </label>

        <label class="wh-field">
          <span class="wh-label">描述</span>
          <textarea
            v-model="form.description"
            class="wh-input au-textarea"
            maxlength="500"
            placeholder="可选，描述规则的用途与场景"
          />
        </label>

        <label class="wh-field">
          <span class="wh-label">触发器类型 <i>*</i></span>
          <select v-model="form.triggerType" class="wh-input">
            <option v-for="o in TRIGGER_OPTIONS" :key="o.value" :value="o.value">
              {{ o.label }}
            </option>
          </select>
        </label>

        <!-- Version (edit mode, read-only) -->
        <label v-if="editingRule" class="wh-field">
          <span class="wh-label">版本</span>
          <input :value="editingRule.version" class="wh-input" disabled />
        </label>

        <!-- ===== Conditions ===== -->
        <fieldset class="wh-field au-fieldset">
          <legend class="wh-label">条件（可选）</legend>
          <div v-if="form.conditions.length" class="au-cond-list">
            <div v-for="(c, idx) in form.conditions" :key="idx" class="au-cond-row">
              <input v-model="c.field" class="wh-input au-cond-field" placeholder="字段名 (如: state)" />
              <select v-model="c.operator" class="wh-input au-cond-op">
                <option v-for="o in OPERATOR_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
              </select>
              <input v-model="c.value" class="wh-input au-cond-val" placeholder="值" />
              <button class="btn-sm btn-sm--danger" @click="removeCondition(idx)" title="删除条件">×</button>
            </div>
          </div>
          <button class="btn-sm au-add-btn" @click="addCondition">＋ 添加条件</button>
        </fieldset>

        <!-- ===== Actions ===== -->
        <fieldset class="wh-field au-fieldset">
          <legend class="wh-label">动作 <i>*</i>（{{ form.actions.length }} 个）</legend>
          <div v-if="form.actions.length" class="au-act-list">
            <div v-for="(a, idx) in form.actions" :key="idx" class="au-act-row">
              <div class="au-act-row__head">
                <span class="au-act-row__num">{{ idx + 1 }}</span>
                <select v-model="a.type" class="wh-input au-act-type" @change="onActionTypeChange(idx)">
                  <option v-for="o in ACTION_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
                </select>
                <button class="btn-sm btn-sm--danger" @click="removeAction(idx)" title="删除动作">×</button>
              </div>
              <!-- transition_status params -->
              <template v-if="a.type === 'transition_status'">
                <input v-model="a.params.target_state" class="wh-input au-act-param" placeholder="目标状态 (如: in_progress)" />
              </template>
              <!-- assign params -->
              <template v-else-if="a.type === 'assign'">
                <input v-model="a.params.assignee" class="wh-input au-act-param" placeholder="负责人 ID 或邮箱" />
              </template>
              <!-- update_field params -->
              <template v-else-if="a.type === 'update_field'">
                <input v-model="a.params.field" class="wh-input au-act-param" placeholder="字段名" />
                <input v-model="a.params.value" class="wh-input au-act-param" placeholder="新值" />
              </template>
              <!-- send_notification params -->
              <template v-else-if="a.type === 'send_notification'">
                <input v-model="a.params.recipient" class="wh-input au-act-param" placeholder="收件人" />
                <input v-model="a.params.channel" class="wh-input au-act-param" placeholder="渠道 (如: email)" />
                <input v-model="a.params.message" class="wh-input au-act-param au-act-param--wide" placeholder="消息内容" />
              </template>
              <!-- create_issue params -->
              <template v-else-if="a.type === 'create_issue'">
                <input v-model="a.params.project_id" class="wh-input au-act-param" placeholder="目标项目 ID" />
                <input v-model="a.params.type" class="wh-input au-act-param" placeholder="工作项类型 (如: task)" />
              </template>
              <!-- copy_field params -->
              <template v-else-if="a.type === 'copy_field'">
                <input v-model="a.params.source" class="wh-input au-act-param" placeholder="源字段" />
                <input v-model="a.params.target" class="wh-input au-act-param" placeholder="目标字段" />
              </template>
            </div>
          </div>
          <button class="btn-sm au-add-btn" @click="addAction">＋ 添加动作</button>
        </fieldset>

        <!-- Dry-Run result panel -->
        <div v-if="showDryRun || dryRunResult" class="au-dryrun">
          <div class="au-dryrun__head">
            <strong>干跑测试结果</strong>
            <button class="btn-sm" @click="showDryRun = false; dryRunResult = null">×</button>
          </div>
          <div v-if="dryRunLoading" class="au-dryrun__loading">正在校验...</div>
          <div v-else-if="dryRunResult" class="au-dryrun__body">
            <div class="au-dryrun__status">
              <span
                class="wh-badge"
                :class="dryRunResult.valid ? 'active' : 'unhealthy'"
              >
                {{ dryRunResult.valid ? "校验通过" : "校验失败" }}
              </span>
              <span class="au-meta">动作数：{{ dryRunResult.actions_count }}</span>
              <span class="au-meta">触发器：{{ triggerLabel(dryRunResult.trigger_type) }}</span>
            </div>
            <ul v-if="dryRunResult.errors.length" class="au-dryrun__errors">
              <li v-for="(err, i) in dryRunResult.errors" :key="i">{{ err }}</li>
            </ul>
            <ul v-if="dryRunResult.warnings.length" class="au-dryrun__warnings">
              <li v-for="(w, i) in dryRunResult.warnings" :key="i">{{ w }}</li>
            </ul>
          </div>
        </div>

        <p v-if="formError" class="wh-err">{{ formError }}</p>

        <div class="wh-modal-actions">
          <button class="btn btn--primary" :disabled="formSaving" @click="saveForm(true)">
            {{ formSaving ? "保存中..." : editingRule ? "保存并激活" : "创建并激活" }}
          </button>
          <button class="btn" :disabled="formSaving" @click="saveForm(false)">
            {{ editingRule ? "保存" : "保存草稿" }}
          </button>
          <button class="btn" :disabled="dryRunLoading" @click="runDryRun">
            {{ dryRunLoading ? "校验中..." : "Dry Run" }}
          </button>
          <button class="btn" @click="closeForm">取消</button>
        </div>
      </div>
    </div>

    <!-- ===== Execution history drawer ===== -->
    <div v-if="showHistory" class="wh-overlay" @click.self="closeHistory">
      <div class="wh-modal wh-modal--lg">
        <h2>执行历史 — {{ historyRuleName }}</h2>
        <AppLoadingState v-if="historyLoading" text="加载执行记录..." />
        <table v-else-if="executions.length" class="wh-log-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>触发事件</th>
              <th>状态</th>
              <th>执行动作</th>
              <th>耗时</th>
              <th>错误</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="e in executions"
              :key="e.id"
              :class="{ 'au-log-row--failed': e.status === 'failed' }"
            >
              <td class="meta">{{ formatTime(e.created_at) }}</td>
              <td><span class="ev-tag">{{ triggerLabel(e.trigger_event) || e.trigger_event }}</span></td>
              <td>
                <span class="wh-badge" :class="execStatus(e.status).cls">
                  {{ execStatus(e.status).text }}
                </span>
              </td>
              <td class="meta au-log-actions">
                <span v-for="(a, i) in e.actions_executed" :key="i" class="au-log-action">{{ a }}</span>
              </td>
              <td class="meta">{{ e.duration_ms ? `${e.duration_ms}ms` : "-" }}</td>
              <td class="wh-log-err">{{ e.error_message ?? "-" }}</td>
            </tr>
          </tbody>
        </table>
        <p v-else class="au-empty">暂无执行记录</p>
        <button class="btn" style="margin-top:12px" @click="closeHistory">关闭</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.au-view { max-width: 880px; }

/* Header */
.au-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.au-header h1 { margin: 0; font-size: 20px; }
.au-sub { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }

.au-header__actions {
  display: flex;
  position: relative;
}

.au-dropdown-wrapper {
  display: flex;
  position: relative;
}

.au-dropdown-toggle {
  border-top-left-radius: 0 !important;
  border-bottom-left-radius: 0 !important;
  padding: 8px 8px !important;
  display: flex;
  align-items: center;
}

.au-caret {
  display: inline-block;
  transition: transform 0.15s;
  font-size: 10px;
}
.au-caret.open { transform: rotate(180deg); }

.au-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 400;
  min-width: 280px;
  max-height: 320px;
  overflow-y: auto;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
  margin-top: 4px;
}

.au-dropdown-empty {
  padding: 12px;
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: center;
}

.au-dropdown-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  width: 100%;
  background: none;
  border: none;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  color: var(--text-primary);
}
.au-dropdown-item:last-child { border-bottom: none; }
.au-dropdown-item:hover { background: var(--surface-2); }
.au-dropdown-item:disabled { opacity: 0.5; cursor: not-allowed; }

.au-dropdown-item__icon { font-size: 18px; flex-shrink: 0; }
.au-dropdown-item__body { display: flex; flex-direction: column; gap: 2px; }
.au-dropdown-item__body strong { font-size: 13px; font-weight: 500; }
.au-dropdown-item__body span { font-size: 11px; color: var(--text-tertiary); }

/* List */
.au-list { display: flex; flex-direction: column; gap: 12px; }

.au-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-1);
}

.au-card__main { flex: 1; min-width: 0; }
.au-card__head { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.au-card__name { font-size: 14px; }

.au-card__desc {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 0 8px;
}

.au-card__tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}

.au-au-count {
  font-size: 11px;
  color: var(--text-tertiary);
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--surface-3);
}

.au-meta { color: var(--text-tertiary); }
.au-meta--err { color: var(--danger-500); font-weight: 500; }

.au-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-left: 16px;
  flex-shrink: 0;
}

/* Badge count */
.au-badge-count {
  margin-left: 4px;
  font-size: 10px;
  background: rgba(255, 255, 255, 0.25);
  padding: 0 4px;
  border-radius: 6px;
}

.au-empty { color: var(--text-tertiary); font-size: 13px; padding: 32px 0; }

/* Textarea */
.au-textarea {
  min-height: 60px;
  resize: vertical;
  font-family: inherit;
}

/* Fieldset */
.au-fieldset {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  margin-bottom: 14px;
}

/* Conditions / Actions */
.au-cond-list, .au-act-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 8px;
}

.au-cond-row {
  display: flex;
  gap: 6px;
  align-items: center;
}

.au-cond-field { flex: 2; }
.au-cond-op { flex: 1; min-width: 100px; }
.au-cond-val { flex: 2; }

.au-act-row {
  padding: 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.au-act-row__head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.au-act-row__num {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  width: 18px;
  text-align: center;
  flex-shrink: 0;
}

.au-act-type { flex: 1; }

.au-act-param {
  margin-top: 4px;
}

.au-act-param--wide { width: 100%; }

.au-add-btn {
  align-self: flex-start;
}

/* Dry run panel */
.au-dryrun {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
  overflow: hidden;
}

.au-dryrun__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--surface-2);
  font-size: 13px;
}

.au-dryrun__loading {
  padding: 12px;
  font-size: 12px;
  color: var(--text-tertiary);
  text-align: center;
}

.au-dryrun__body { padding: 10px 12px; }

.au-dryrun__status {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.au-dryrun__errors {
  list-style: none;
  padding: 0;
  margin: 0 0 6px;
  font-size: 12px;
  color: var(--danger-500);
}

.au-dryrun__errors li::before { content: "✕ "; }

.au-dryrun__warnings {
  list-style: none;
  padding: 0;
  margin: 0;
  font-size: 12px;
  color: var(--warning-500, #f59e0b);
}

.au-dryrun__warnings li::before { content: "⚠ "; }

/* Log table */
.au-log-row--failed { background: rgba(220, 47, 47, 0.04); }

.au-log-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}

.au-log-action {
  font-size: 11px;
  padding: 1px 5px;
  border-radius: 3px;
  background: var(--surface-3);
  color: var(--text-secondary);
}

/* Reuse styles from WebhookSettingsView */
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
.btn-sm--danger:hover { background: rgba(220, 47, 47, 0.08); }

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
.wh-badge.draft { background: var(--surface-3); color: var(--text-tertiary); }
.wh-badge.disabled { background: rgba(245,158,11,0.12); color: var(--warning-500,#f59e0b); }
.wh-badge.error { background: rgba(220,47,47,0.12); color: var(--danger-500,#dc2f2f); }

/* au-prefixed badge variants (used by AutomationView status maps) */
.au-badge--active { background: rgba(15,194,123,0.12); color: var(--success-500,#0fc27b); }
.au-badge--draft { background: var(--surface-3); color: var(--text-tertiary); }
.au-badge--disabled { background: rgba(245,158,11,0.12); color: var(--warning-500,#f59e0b); }
.au-badge--error { background: rgba(220,47,47,0.12); color: var(--danger-500,#dc2f2f); }

/* The wh-badge base class must also be applied when using au-badge--* variants
   since they share the same visual base. We rely on the template composing both classes. */

.ev-tag {
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  background: var(--surface-3);
  color: var(--text-secondary);
}

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

.wh-err { color: var(--danger-500); font-size: 12px; margin: 8px 0 0; }

.wh-modal-actions { display: flex; gap: 8px; margin-top: 16px; flex-wrap: wrap; }

.wh-log-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.wh-log-table th { text-align: left; padding: 6px 8px; color: var(--text-tertiary); font-weight: 500; border-bottom: 1px solid var(--border-subtle); }
.wh-log-table td { padding: 6px 8px; border-bottom: 1px solid var(--border-subtle); color: var(--text-primary); }
.wh-log-err { max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--danger-500); }

.mono { font-family: var(--font-mono); }
.meta { color: var(--text-tertiary); font-size: 12px; }
</style>
