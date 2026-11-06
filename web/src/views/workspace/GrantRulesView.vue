<script setup lang="ts">
/**
 * GrantRulesView — 工作空间级条件授权（ABAC）规则管理。
 *
 * 功能：
 *  - 列表展示规则（名称/描述/角色/权限/范围/启用状态/优先级）；
 *  - 创建 / 编辑规则（含可视化条件构造器：and / or / not 嵌套 + 字段/操作符/值行）；
 *  - 试算：给定上下文属性，调用后端 evaluate 校验条件是否命中；
 *  - 启用 / 停用 / 删除规则。
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import {
  grantApi,
  type GrantContext,
  type GrantExpr,
  type GrantField,
  type GrantRule,
} from "@/api/services/grant";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
} from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import GrantConditionBuilder, {
  type ConditionNode,
} from "@/components/GrantConditionBuilder.vue";
import { useI18n } from "@/composables/useI18n";
import { useConfirm } from "@/composables/useConfirm";
import { toast } from "@/lib/toast";

const { t } = useI18n();
const confirm = useConfirm();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

const ROLE_OPTIONS: { value: string; label: string }[] = [
  { value: "owner", label: "owner" },
  { value: "admin", label: "admin" },
  { value: "pm", label: "pm" },
  { value: "po", label: "po" },
  { value: "techlead", label: "techlead" },
  { value: "qalead", label: "qalead" },
  { value: "dev", label: "dev" },
  { value: "guest", label: "guest" },
];

const PERMISSION_OPTIONS: { value: string; label: string }[] = [
  { value: "issue:read", label: "issue:read" },
  { value: "issue:edit-own", label: "issue:edit-own" },
  { value: "issue:edit-all", label: "issue:edit-all" },
  { value: "issue:delete", label: "issue:delete" },
  { value: "issue:transition", label: "issue:transition" },
  { value: "issue:reassign", label: "issue:reassign" },
  { value: "issue:change-priority", label: "issue:change-priority" },
  { value: "sprint:create", label: "sprint:create" },
  { value: "sprint:update", label: "sprint:update" },
  { value: "module:create", label: "module:create" },
  { value: "module:update", label: "module:update" },
  { value: "page:create", label: "page:create" },
  { value: "page:update", label: "page:update" },
  { value: "member:invite", label: "member:invite" },
  { value: "member:remove", label: "member:remove" },
  { value: "project:update", label: "project:update" },
];

/* ------------------------------------------------------------------ */
/* 条件构造器实例引用（用于序列化校验 / 反序列化）                     */
/* ------------------------------------------------------------------ */

const builderRef = ref<InstanceType<typeof GrantConditionBuilder> | null>(null);

function nodeToExpr(node: ConditionNode): GrantExpr | null {
  return builderRef.value?.nodeToExpr(node) ?? null;
}

function nodeIsValid(node: ConditionNode): boolean {
  return builderRef.value?.nodeIsValid(node) ?? false;
}

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const rules = ref<GrantRule[]>([]);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    rules.value = await grantApi.list(workspaceId.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("grantRule.loadFailed");
  } finally {
    loading.value = false;
  }
}

onMounted(load);

/* ------------------------------------------------------------------ */
/* 表单状态                                                            */
/* ------------------------------------------------------------------ */

interface RuleForm {
  name: string;
  description: string;
  scope: string;
  role_slugs: string[];
  permissions: string[];
  priority: number;
  root: ConditionNode;
}

function emptyForm(): RuleForm {
  return {
    name: "",
    description: "",
    scope: "",
    role_slugs: [],
    permissions: [],
    priority: 100,
    root: {
      key: `root_${Date.now()}`,
      op: "and",
      field: "",
      value: "",
      valuesText: "",
      children: [
        {
          key: `c_${Date.now()}`,
          op: "eq",
          field: "projectId",
          value: "",
          valuesText: "",
          children: [],
          inner: null,
        },
      ],
      inner: null,
    },
  };
}

const showFormModal = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const form = reactive<RuleForm>(emptyForm());

const isEditing = computed(() => editingId.value !== null);

function toggleRole(slug: string) {
  const i = form.role_slugs.indexOf(slug);
  if (i >= 0) form.role_slugs.splice(i, 1);
  else form.role_slugs.push(slug);
}

function togglePermission(perm: string) {
  const i = form.permissions.indexOf(perm);
  if (i >= 0) form.permissions.splice(i, 1);
  else form.permissions.push(perm);
}

/* ---------------- 打开 / 提交 ---------------- */

function openCreate() {
  editingId.value = null;
  Object.assign(form, emptyForm());
  showFormModal.value = true;
}

function openEdit(rule: GrantRule) {
  editingId.value = rule.id;
  const root = builderRef.value?.exprToNode(rule.condition);
  Object.assign(form, {
    name: rule.name,
    description: rule.description ?? "",
    scope: rule.scope ?? "",
    role_slugs: [...(rule.role_slugs ?? [])],
    permissions: [...(rule.permissions ?? [])],
    priority: rule.priority ?? 100,
    root: root ?? emptyForm().root,
  });
  showFormModal.value = true;
}

function validateForm(): boolean {
  if (!form.name.trim()) {
    toast.error(t("grantRule.nameRequired"));
    return false;
  }
  if (form.permissions.length === 0) {
    toast.error(t("grantRule.permissionRequired"));
    return false;
  }
  if (!nodeIsValid(form.root)) {
    toast.error(t("grantRule.conditionInvalid"));
    return false;
  }
  return true;
}

async function submit() {
  if (!validateForm()) return;
  const condition = nodeToExpr(form.root);
  if (!condition) return;

  saving.value = true;
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description.trim(),
      scope: form.scope.trim(),
      role_slugs: [...form.role_slugs],
      permissions: [...form.permissions],
      priority: Number(form.priority) || 100,
      condition,
    };
    if (isEditing.value) {
      await grantApi.update(workspaceId.value, editingId.value!, payload);
      toast.success(t("grantRule.updateSuccess"));
      showFormModal.value = false;
    } else {
      await grantApi.create(workspaceId.value, payload);
      toast.success(t("grantRule.createSuccess"));
      showFormModal.value = false;
    }
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("grantRule.saveFailed"));
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 启用 / 停用 / 删除                                                  */
/* ------------------------------------------------------------------ */

async function toggleEnabled(rule: GrantRule) {
  try {
    if (rule.enabled) {
      await grantApi.disable(workspaceId.value, rule.id);
      toast.success(t("grantRule.disableSuccess"));
    } else {
      await grantApi.enable(workspaceId.value, rule.id);
      toast.success(t("grantRule.enableSuccess"));
    }
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("grantRule.actionFailed"));
  }
}

async function removeRule(rule: GrantRule) {
  if (
    !(await confirm({
      title: t("grantRule.deleteTitle"),
      message: t("grantRule.confirmDelete", { name: rule.name }),
      confirmText: t("common.delete"),
      variant: "danger",
    }))
  ) {
    return;
  }
  try {
    await grantApi.remove(workspaceId.value, rule.id);
    rules.value = rules.value.filter((r) => r.id !== rule.id);
    toast.success(t("grantRule.deleteSuccess"));
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("grantRule.actionFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 试算                                                                */
/* ------------------------------------------------------------------ */

const showEvalModal = ref(false);
const evalCondition = ref<GrantExpr | null>(null);
const evalContextText = ref("");
const evalLoading = ref(false);
const evalResult = ref<{ valid: boolean; result: boolean; error?: string } | null>(null);
const evalError = ref("");

function collectFields(expr: GrantExpr | null | undefined): GrantField[] {
  const out = new Set<GrantField>();
  const walk = (e: GrantExpr | null | undefined) => {
    if (!e) return;
    if (e.field) out.add(e.field);
    if (e.inner) walk(e.inner);
    for (const s of e.all ?? []) walk(s);
    for (const s of e.any ?? []) walk(s);
  };
  walk(expr);
  return Array.from(out);
}

function openEvaluate(rule: GrantRule) {
  evalCondition.value = rule.condition;
  const fields = collectFields(rule.condition);
  const prefill: Record<string, string> = {};
  for (const f of fields) prefill[f] = "";
  evalContextText.value = JSON.stringify(prefill, null, 2);
  evalResult.value = null;
  evalError.value = "";
  showEvalModal.value = true;
}

async function runEvaluate() {
  if (!evalCondition.value) return;
  let ctx: GrantContext;
  try {
    const parsed = JSON.parse(evalContextText.value || "{}");
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
      throw new Error("context must be an object");
    }
    ctx = parsed as GrantContext;
  } catch {
    evalError.value = t("grantRule.contextInvalid");
    evalResult.value = null;
    return;
  }

  evalLoading.value = true;
  evalError.value = "";
  evalResult.value = null;
  try {
    const r = await grantApi.evaluate(workspaceId.value, evalCondition.value, ctx);
    evalResult.value = r;
  } catch (err: unknown) {
    evalError.value = err instanceof Error ? err.message : t("grantRule.actionFailed");
  } finally {
    evalLoading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}/${m}/${day}`;
}
</script>

<template>
  <div class="space-y-6">
    <!-- 头部 -->
    <header class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
          {{ t("grantRule.title") }}
        </h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          {{ t("grantRule.subtitle") }}
        </p>
      </div>
      <AppButton @click="openCreate">
        <DynamicIcon name="ShieldCheck" :size="16" />
        {{ t("grantRule.create") }}
      </AppButton>
    </header>

    <!-- 加载 -->
    <AppLoadingState v-if="loading" :text="t('common.loading')" />

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" :retry="t('common.retry')" @retry="load" />

    <!-- 空状态 -->
    <AppEmptyState
      v-else-if="rules.length === 0"
      scenario="automation"
      :title="t('grantRule.emptyTitle')"
      :description="t('grantRule.emptyDesc')"
      :cta-text="t('grantRule.create')"
      @cta-click="openCreate"
    />

    <!-- 规则列表 -->
    <div v-else class="space-y-3">
      <div
        v-for="rule in rules"
        :key="rule.id"
        class="rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4 transition hover:border-[var(--border-default)] hover:shadow-[var(--shadow-card)]"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-semibold text-[var(--text-primary)]">
                {{ rule.name }}
              </span>
              <AppBadge :variant="rule.enabled ? 'success' : 'default'">
                {{ rule.enabled ? t("grantRule.enabled") : t("grantRule.disabled") }}
              </AppBadge>
              <AppBadge variant="info">P{{ rule.priority }}</AppBadge>
              <span
                v-if="rule.scope"
                class="rounded bg-[var(--surface-3)] px-1.5 py-0.5 font-mono text-[11px] text-[var(--text-secondary)]"
              >
                {{ rule.scope }}
              </span>
            </div>
            <p
              v-if="rule.description"
              class="mt-1 line-clamp-2 text-sm text-[var(--text-secondary)]"
            >
              {{ rule.description }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-3">
            <button
              class="text-xs text-[var(--brand-600)] hover:underline"
              @click="openEvaluate(rule)"
            >
              {{ t("grantRule.evaluate") }}
            </button>
            <button
              class="text-xs text-[var(--brand-600)] hover:underline"
              @click="openEdit(rule)"
            >
              {{ t("common.edit") }}
            </button>
            <button
              class="text-xs text-[var(--brand-600)] hover:underline"
              @click="toggleEnabled(rule)"
            >
              {{ rule.enabled ? t("grantRule.disable") : t("grantRule.enable") }}
            </button>
            <button
              class="text-xs text-[var(--danger-500)] hover:underline"
              @click="removeRule(rule)"
            >
              {{ t("common.delete") }}
            </button>
          </div>
        </div>

        <div class="mt-3 flex flex-wrap gap-1.5">
          <span
            v-for="role in rule.role_slugs"
            :key="`role-${role}`"
            class="rounded bg-[var(--brand-50)] px-1.5 py-0.5 font-mono text-[11px] text-[var(--brand-600)]"
          >
            {{ role }}
          </span>
          <span
            v-for="perm in rule.permissions"
            :key="`perm-${perm}`"
            class="rounded bg-[var(--surface-3)] px-1.5 py-0.5 font-mono text-[11px] text-[var(--text-secondary)]"
          >
            {{ perm }}
          </span>
        </div>

        <div class="mt-2 text-xs text-[var(--text-tertiary)]">
          {{ t("grantRule.updatedAt") }} {{ formatDate(rule.updated_at) }}
        </div>
      </div>
    </div>

    <!-- 创建 / 编辑弹窗 -->
    <AppModal
      :visible="showFormModal"
      :title="isEditing ? t('grantRule.editTitle') : t('grantRule.createTitle')"
      width="720px"
      @close="showFormModal = false"
    >
      <div class="space-y-4">
        <!-- 名称 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.name") }} <span class="text-[var(--danger-500)]">*</span>
          </span>
          <input
            v-model="form.name"
            type="text"
            :placeholder="t('grantRule.namePlaceholder')"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </label>

        <!-- 描述 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.description") }}
          </span>
          <textarea
            v-model="form.description"
            rows="2"
            :placeholder="t('grantRule.descriptionPlaceholder')"
            class="w-full resize-none rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </label>

        <!-- 范围 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.scope") }}
          </span>
          <input
            v-model="form.scope"
            type="text"
            :placeholder="t('grantRule.scopePlaceholder')"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.scopeHint") }}
          </span>
        </label>

        <!-- 角色 -->
        <div>
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.roles") }}
          </span>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <label
              v-for="role in ROLE_OPTIONS"
              :key="role.value"
              class="flex cursor-pointer items-center gap-2 rounded-md border border-[var(--border-subtle)] px-2.5 py-1.5 text-xs text-[var(--text-secondary)] transition hover:bg-[var(--surface-2)]"
            >
              <input
                type="checkbox"
                :checked="form.role_slugs.includes(role.value)"
                class="h-3.5 w-3.5 accent-[var(--brand-500)]"
                @change="toggleRole(role.value)"
              />
              <span class="font-mono">{{ role.label }}</span>
            </label>
          </div>
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.rolesHint") }}
          </span>
        </div>

        <!-- 权限 -->
        <div>
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.permissions") }}
            <span class="text-[var(--danger-500)]">*</span>
          </span>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
            <label
              v-for="perm in PERMISSION_OPTIONS"
              :key="perm.value"
              class="flex cursor-pointer items-center gap-2 rounded-md border border-[var(--border-subtle)] px-2.5 py-1.5 text-xs text-[var(--text-secondary)] transition hover:bg-[var(--surface-2)]"
            >
              <input
                type="checkbox"
                :checked="form.permissions.includes(perm.value)"
                class="h-3.5 w-3.5 accent-[var(--brand-500)]"
                @change="togglePermission(perm.value)"
              />
              <span class="font-mono">{{ perm.label }}</span>
            </label>
          </div>
        </div>

        <!-- 优先级 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.priority") }}
          </span>
          <input
            v-model.number="form.priority"
            type="number"
            min="1"
            class="w-32 rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.priorityHint") }}
          </span>
        </label>

        <!-- 条件构造器 -->
        <div>
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.condition") }}
            <span class="text-[var(--danger-500)]">*</span>
          </span>
          <div class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-2)] p-3">
            <GrantConditionBuilder ref="builderRef" v-model="form.root" />
          </div>
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.conditionHint") }}
          </span>
        </div>
      </div>

      <template #footer>
        <AppButton variant="secondary" @click="showFormModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton :loading="saving" @click="submit">
          {{ saving ? t("common.saving") : t("common.save") }}
        </AppButton>
      </template>
    </AppModal>

    <!-- 试算弹窗 -->
    <AppModal
      :visible="showEvalModal"
      :title="t('grantRule.evaluateTitle')"
      width="560px"
      @close="showEvalModal = false"
    >
      <div class="space-y-4">
        <div>
          <div class="mb-1 text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.conditionJson") }}
          </div>
          <pre
            class="max-h-40 overflow-auto rounded-md bg-[var(--surface-2)] p-2 font-mono text-xs text-[var(--text-secondary)]"
          >{{ JSON.stringify(evalCondition, null, 2) }}</pre>
        </div>

        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("grantRule.context") }}
          </span>
          <textarea
            v-model="evalContextText"
            rows="6"
            class="w-full resize-y rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 font-mono text-xs text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("grantRule.contextHint") }}
          </span>
        </label>

        <AppButton :loading="evalLoading" @click="runEvaluate">
          {{ t("grantRule.runEvaluate") }}
        </AppButton>

        <!-- 结果 -->
        <div
          v-if="evalResult"
          class="rounded-md border px-3 py-2.5 text-sm"
          :class="
            evalResult.valid
              ? evalResult.result
                ? 'border-[var(--success-300)] bg-[var(--success-50)] text-[var(--success-600)]'
                : 'border-[var(--border-default)] bg-[var(--surface-2)] text-[var(--text-secondary)]'
              : 'border-[var(--danger-300)] bg-[var(--danger-50)] text-[var(--danger-600)]'
          "
        >
          <div class="font-medium">
            {{
              !evalResult.valid
                ? t("grantRule.evalInvalid")
                : evalResult.result
                  ? t("grantRule.evalMatch")
                  : t("grantRule.evalNoMatch")
            }}
          </div>
          <div v-if="evalResult.error" class="mt-1 text-xs opacity-80">
            {{ evalResult.error }}
          </div>
        </div>

        <div
          v-else-if="evalError"
          class="rounded-md border border-[var(--danger-300)] bg-[var(--danger-50)] px-3 py-2.5 text-sm text-[var(--danger-600)]"
        >
          {{ evalError }}
        </div>
      </div>
    </AppModal>
  </div>
</template>
