<script setup lang="ts">
/**
 * OAuthAppsView — 工作空间级 OAuth 应用（开发者应用）管理。
 *
 * 功能：
 *  - 列表展示已注册应用（名称/描述/Client ID/状态/创建时间）；
 *  - 注册应用（名称、描述、回调地址、Scopes、图标）；
 *  - 创建后一次性回显 Client ID + Client Secret；
 *  - 编辑应用信息；
 *  - 轮换 Client Secret（确认后一次性回显新密钥）；
 *  - 删除应用。
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import {
  oauthAppApi,
  type OAuthApp,
  type OAuthAppStatus,
} from "@/api/services/oauthapp";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
} from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useI18n } from "@/composables/useI18n";
import { useConfirm } from "@/composables/useConfirm";
import { toast } from "@/lib/toast";

const confirm = useConfirm();
const { t } = useI18n();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const apps = ref<OAuthApp[]>([]);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    apps.value = await oauthAppApi.list(workspaceId.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("oauthApp.loadFailed");
  } finally {
    loading.value = false;
  }
}

onMounted(load);

/* ------------------------------------------------------------------ */
/* 表单状态                                                            */
/* ------------------------------------------------------------------ */

interface AppForm {
  name: string;
  description: string;
  /** 多行文本，每行一个 URI */
  redirectUrisText: string;
  scopes: string[];
  iconUrl: string;
}

function emptyForm(): AppForm {
  return {
    name: "",
    description: "",
    redirectUrisText: "",
    scopes: [],
    iconUrl: "",
  };
}

/** 可选 Scopes（与后端权限常量对齐）。 */
const ALL_SCOPES: { value: string; label: string }[] = [
  { value: "workspace:read", label: "workspace:read" },
  { value: "workspace:update", label: "workspace:update" },
  { value: "project:read", label: "project:read" },
  { value: "project:update", label: "project:update" },
  { value: "issue:read", label: "issue:read" },
  { value: "issue:write", label: "issue:write" },
  { value: "comment:moderate", label: "comment:moderate" },
  { value: "sprint:read", label: "sprint:read" },
  { value: "version:read", label: "version:read" },
];

function parseRedirectUris(text: string): string[] {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
}

/** 将多行 URI 文本解析为数组并做基础校验；返回 null 表示校验失败。 */
function validateForm(form: AppForm): string[] | null {
  const uris = parseRedirectUris(form.redirectUrisText);
  if (!form.name.trim()) {
    toast.error(t("oauthApp.nameRequired"));
    return null;
  }
  if (uris.length === 0) {
    toast.error(t("oauthApp.redirectUriRequired"));
    return null;
  }
  const bad = uris.find(
    (uri) => !uri.startsWith("https://") && !uri.startsWith("http://localhost"),
  );
  if (bad) {
    toast.error(`${t("oauthApp.needHttps")}（${bad}）`);
    return null;
  }
  return uris;
}

function buildPayload(form: AppForm) {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    redirect_uris: parseRedirectUris(form.redirectUrisText),
    scopes: [...form.scopes],
    icon_url: form.iconUrl.trim(),
  };
}

/* ------------------------------------------------------------------ */
/* 创建 / 编辑                                                         */
/* ------------------------------------------------------------------ */

const showFormModal = ref(false);
const editingId = ref<number | null>(null);
const saving = ref(false);
const form = reactive<AppForm>(emptyForm());

const isEditing = computed(() => editingId.value !== null);

function openCreate() {
  editingId.value = null;
  Object.assign(form, emptyForm());
  showFormModal.value = true;
}

function openEdit(app: OAuthApp) {
  editingId.value = app.id;
  Object.assign(form, {
    name: app.name,
    description: app.description ?? "",
    redirectUrisText: (app.redirect_uris ?? []).join("\n"),
    scopes: [...(app.scopes ?? [])],
    iconUrl: app.icon_url ?? "",
  });
  showFormModal.value = true;
}

async function submit() {
  const uris = validateForm(form);
  if (!uris) return;

  saving.value = true;
  try {
    const payload = buildPayload(form);
    if (isEditing.value) {
      await oauthAppApi.update(workspaceId.value, editingId.value!, payload);
      toast.success(t("oauthApp.updateSuccess"));
      showFormModal.value = false;
    } else {
      const created = await oauthAppApi.create(workspaceId.value, payload);
      toast.success(t("oauthApp.createSuccess"));
      showFormModal.value = false;
      // 一次性回显 client_id + client_secret
      secretModal.value = {
        title: t("oauthApp.secretOnceTitle"),
        warning: t("oauthApp.secretOnceWarning"),
        clientId: created.client_id,
        clientSecret: created.client_secret ?? "",
        showClientId: true,
      };
      showSecretModal.value = true;
    }
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("oauthApp.saveFailed"));
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 一次性密钥回显弹窗                                                   */
/* ------------------------------------------------------------------ */

interface SecretModalState {
  title: string;
  warning: string;
  clientId: string;
  clientSecret: string;
  showClientId: boolean;
}

const showSecretModal = ref(false);
const secretModal = ref<SecretModalState>({
  title: "",
  warning: "",
  clientId: "",
  clientSecret: "",
  showClientId: true,
});

/* ------------------------------------------------------------------ */
/* 轮换密钥                                                            */
/* ------------------------------------------------------------------ */

const rotatingId = ref<number | null>(null);

async function rotateSecret(app: OAuthApp) {
  if (!await confirm({ title: "轮换密钥", message: t("oauthApp.confirmRotate", { name: app.name }), confirmText: "轮换", variant: "danger" })) return;
  rotatingId.value = app.id;
  try {
    const secret = await oauthAppApi.rotateSecret(workspaceId.value, app.id);
    toast.success(t("oauthApp.rotateSuccess"));
    secretModal.value = {
      title: t("oauthApp.rotateOnceTitle"),
      warning: t("oauthApp.rotateOnceWarning"),
      clientId: app.client_id,
      clientSecret: secret,
      showClientId: false,
    };
    showSecretModal.value = true;
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("oauthApp.saveFailed"));
  } finally {
    rotatingId.value = null;
  }
}

/* ------------------------------------------------------------------ */
/* 删除                                                                */
/* ------------------------------------------------------------------ */

async function removeApp(app: OAuthApp) {
  if (!await confirm({ title: "删除应用", message: t("oauthApp.confirmDelete", { name: app.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await oauthAppApi.remove(workspaceId.value, app.id);
    apps.value = apps.value.filter((a) => a.id !== app.id);
    toast.success(t("oauthApp.deleteSuccess"));
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("oauthApp.saveFailed"));
  }
}

/* ------------------------------------------------------------------ */
/* 复制                                                                */
/* ------------------------------------------------------------------ */

const copiedKey = ref<string>("");

async function copyText(key: string, text: string) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      // 兼容旧浏览器：execCommand 兜底
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.style.position = "fixed";
      ta.style.opacity = "0";
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
    }
    copiedKey.value = key;
    toast.success(t("oauthApp.copied"));
    window.setTimeout(() => {
      if (copiedKey.value === key) copiedKey.value = "";
    }, 1500);
  } catch {
    toast.error(t("oauthApp.copy"));
  }
}

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function statusVariant(status: OAuthAppStatus): "success" | "danger" {
  return status === "active" ? "success" : "danger";
}

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
          {{ t("oauthApp.title") }}
        </h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          {{ t("oauthApp.subtitle") }}
        </p>
      </div>
      <AppButton @click="openCreate">
        <DynamicIcon name="KeyRound" :size="16" />
        {{ t("oauthApp.register") }}
      </AppButton>
    </header>

    <!-- 加载 -->
    <AppLoadingState v-if="loading" :text="t('common.loading')" />

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" :retry="t('common.retry')" @retry="load" />

    <!-- 空状态 -->
    <AppEmptyState
      v-else-if="apps.length === 0"
      scenario="api-token"
      :title="t('oauthApp.emptyTitle')"
      :description="t('oauthApp.emptyDesc')"
      :cta-text="t('oauthApp.register')"
      @cta-click="openCreate"
    />

    <!-- 应用网格 -->
    <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
      <div
        v-for="app in apps"
        :key="app.id"
        class="flex flex-col rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4 transition hover:border-[var(--border-default)] hover:shadow-[var(--shadow-card)]"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="flex min-w-0 items-center gap-2.5">
            <div
              v-if="app.icon_url"
              class="h-9 w-9 shrink-0 overflow-hidden rounded-lg border border-[var(--border-subtle)]"
            >
              <img :src="app.icon_url" :alt="app.name" class="h-full w-full object-cover" />
            </div>
            <div
              v-else
              class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--brand-50)] text-[var(--brand-600)]"
            >
              <DynamicIcon name="AppWindow" :size="18" />
            </div>
            <div class="min-w-0">
              <div class="truncate text-sm font-semibold text-[var(--text-primary)]">
                {{ app.name }}
              </div>
              <div class="text-xs text-[var(--text-tertiary)]">
                {{ t("oauthApp.createdAt") }} {{ formatDate(app.created_at) }}
              </div>
            </div>
          </div>
          <AppBadge :variant="statusVariant(app.status)">
            {{
              app.status === "active"
                ? t("oauthApp.statusActive")
                : t("oauthApp.statusRevoked")
            }}
          </AppBadge>
        </div>

        <p
          v-if="app.description"
          class="mt-3 line-clamp-2 text-sm text-[var(--text-secondary)]"
        >
          {{ app.description }}
        </p>

        <!-- Client ID -->
        <div class="mt-3 flex items-center gap-2 rounded-md bg-[var(--surface-2)] px-2.5 py-1.5">
          <span class="shrink-0 text-xs text-[var(--text-tertiary)]">
            {{ t("oauthApp.clientId") }}
          </span>
          <code class="min-w-0 flex-1 truncate font-mono text-xs text-[var(--text-primary)]">
            {{ app.client_id }}
          </code>
          <button
            class="shrink-0 text-xs text-[var(--brand-600)] hover:underline"
            :aria-label="t('oauthApp.copy')"
            @click="copyText('card-' + app.id, app.client_id)"
          >
            {{ copiedKey === 'card-' + app.id ? t("oauthApp.copied") : t("oauthApp.copy") }}
          </button>
        </div>

        <!-- scopes -->
        <div v-if="app.scopes && app.scopes.length" class="mt-3 flex flex-wrap gap-1">
          <span
            v-for="scope in app.scopes"
            :key="scope"
            class="rounded bg-[var(--surface-3)] px-1.5 py-0.5 font-mono text-[11px] text-[var(--text-secondary)]"
          >
            {{ scope }}
          </span>
        </div>

        <!-- 操作 -->
        <div class="mt-4 flex items-center gap-3 border-t border-[var(--border-subtle)] pt-3">
          <button
            class="text-xs text-[var(--brand-600)] hover:underline disabled:opacity-50"
            :disabled="rotatingId === app.id"
            @click="rotateSecret(app)"
          >
            {{ rotatingId === app.id ? t("oauthApp.rotating") : t("oauthApp.rotateSecret") }}
          </button>
          <button class="text-xs text-[var(--brand-600)] hover:underline" @click="openEdit(app)">
            {{ t("oauthApp.edit") }}
          </button>
          <button
            class="ml-auto text-xs text-[var(--danger-500)] hover:underline"
            @click="removeApp(app)"
          >
            {{ t("oauthApp.delete") }}
          </button>
        </div>
      </div>
    </div>

    <!-- 创建 / 编辑弹窗 -->
    <AppModal
      :visible="showFormModal"
      :title="isEditing ? t('oauthApp.editTitle') : t('oauthApp.createTitle')"
      width="560px"
      @close="showFormModal = false"
    >
      <div class="space-y-4">
        <!-- 名称 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("oauthApp.name") }} <span class="text-[var(--danger-500)]">*</span>
          </span>
          <input
            v-model="form.name"
            type="text"
            :placeholder="t('oauthApp.namePlaceholder')"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </label>

        <!-- 描述 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("oauthApp.description") }}
          </span>
          <textarea
            v-model="form.description"
            rows="2"
            :placeholder="t('oauthApp.descriptionPlaceholder')"
            class="w-full resize-none rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </label>

        <!-- 回调地址 -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("oauthApp.redirectUris") }}
            <span class="text-[var(--danger-500)]">*</span>
          </span>
          <textarea
            v-model="form.redirectUrisText"
            rows="3"
            :placeholder="t('oauthApp.redirectUrisPlaceholder')"
            class="w-full resize-y rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 font-mono text-xs text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
          <span class="mt-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("oauthApp.redirectUrisHint") }}
          </span>
        </label>

        <!-- Scopes -->
        <div>
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("oauthApp.scopes") }}
          </span>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
            <label
              v-for="scope in ALL_SCOPES"
              :key="scope.value"
              class="flex cursor-pointer items-center gap-2 rounded-md border border-[var(--border-subtle)] px-2.5 py-1.5 text-xs text-[var(--text-secondary)] transition hover:bg-[var(--surface-2)]"
            >
              <input
                v-model="form.scopes"
                type="checkbox"
                :value="scope.value"
                class="h-3.5 w-3.5 accent-[var(--brand-500)]"
              />
              <span class="font-mono">{{ scope.label }}</span>
            </label>
          </div>
        </div>

        <!-- 图标 URL -->
        <label class="block">
          <span class="mb-1 block text-sm text-[var(--text-secondary)]">
            {{ t("oauthApp.iconUrl") }}
          </span>
          <input
            v-model="form.iconUrl"
            type="url"
            :placeholder="t('oauthApp.iconUrlPlaceholder')"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
          />
        </label>
      </div>

      <template #footer>
        <AppButton variant="secondary" @click="showFormModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton :loading="saving" @click="submit">
          {{ saving ? t("oauthApp.saving") : t("oauthApp.save") }}
        </AppButton>
      </template>
    </AppModal>

    <!-- 一次性密钥回显弹窗 -->
    <AppModal
      :visible="showSecretModal"
      :title="secretModal.title"
      width="520px"
      @close="showSecretModal = false"
    >
      <div
        class="rounded-md border border-[var(--warning-300)] bg-[rgba(245,158,11,0.08)] px-3 py-2.5 text-sm text-[var(--text-primary)]"
      >
        {{ secretModal.warning }}
      </div>

      <div v-if="secretModal.showClientId" class="mt-4">
        <div class="mb-1 text-xs text-[var(--text-tertiary)]">{{ t("oauthApp.clientId") }}</div>
        <div class="flex items-center gap-2 rounded-md bg-[var(--surface-2)] px-2.5 py-2">
          <code class="min-w-0 flex-1 break-all font-mono text-xs text-[var(--text-primary)]">
            {{ secretModal.clientId }}
          </code>
          <button
            class="shrink-0 text-xs text-[var(--brand-600)] hover:underline"
            @click="copyText('secret-client-id', secretModal.clientId)"
          >
            {{ copiedKey === "secret-client-id" ? t("oauthApp.copied") : t("oauthApp.copy") }}
          </button>
        </div>
      </div>

      <div class="mt-4">
        <div class="mb-1 text-xs text-[var(--text-tertiary)]">
          {{ t("oauthApp.clientSecret") }}
        </div>
        <div class="flex items-center gap-2 rounded-md bg-[var(--surface-2)] px-2.5 py-2">
          <code class="min-w-0 flex-1 break-all font-mono text-xs text-[var(--text-primary)]">
            {{ secretModal.clientSecret || "—" }}
          </code>
          <button
            class="shrink-0 text-xs text-[var(--brand-600)] hover:underline"
            @click="copyText('secret-client-secret', secretModal.clientSecret)"
          >
            {{
              copiedKey === "secret-client-secret"
                ? t("oauthApp.copied")
                : t("oauthApp.copy")
            }}
          </button>
        </div>
      </div>

      <template #footer>
        <AppButton @click="showSecretModal = false">
          {{ t("oauthApp.close") }}
        </AppButton>
          </AppModal>
  </div>

