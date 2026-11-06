<script setup lang="ts">
/**
 * IntegrationsView — 集成中心（Marketplace + 第三方连接器）。
 *
 * 功能：
 *   - Marketplace 网格：展示可用 provider（来自 GET /providers），含图标 / 名称 / 描述 / 连接状态；
 *   - 已连接集成列表：状态徽章、最近同步时间、最近错误、操作（测试 / 同步 / 暂停 / 恢复 / 事件日志 / 删除）；
 *   - 连接弹窗：provider 选择 → 按 provider 动态渲染凭证字段 + 配置字段；
 *   - 测试连接：结果 toast 反馈；
 *   - 立即同步：toast 反馈同步结果；
 *   - 事件日志抽屉：方向 / 状态徽章 + payload 预览。
 */
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import {
  integrationApi,
  type Integration,
  type IntegrationEvent,
  type IntegrationProvider,
  type ProviderMeta,
} from "@/api/services/integration";
import { useTranslation } from "@/composables/useTranslation";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
  AppSkeleton,
} from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const { t } = useTranslation();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

/* ------------------------------------------------------------------ */
/* 状态                                                                */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const providers = ref<ProviderMeta[]>([]);
const integrations = ref<Integration[]>([]);

/** provider → 已连接集成（一个 provider 可连多个） */
const connectedByProvider = computed<Record<string, Integration[]>>(() => {
  const map: Record<string, Integration[]> = {};
  for (const it of integrations.value) {
    (map[it.provider] ??= []).push(it);
  }
  return map;
});

/* ------------------------------------------------------------------ */
/* 数据加载                                                            */
/* ------------------------------------------------------------------ */

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [providerRes, listRes] = await Promise.all([
      integrationApi.listProviders(workspaceId.value),
      integrationApi.list(workspaceId.value, { limit: 100 }),
    ]);
    providers.value = providerRes.items ?? [];
    integrations.value = listRes.items ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t("integration.loadFailed");
  } finally {
    loading.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 状态 / 徽章映射                                                     */
/* ------------------------------------------------------------------ */

const statusMeta: Record<
  Integration["status"],
  { label: string; variant: "success" | "warning" | "danger" | "info" | "default" }
> = {
  active: { label: t("integration.status.active"), variant: "success" },
  paused: { label: t("integration.status.paused"), variant: "default" },
  error: { label: t("integration.status.error"), variant: "danger" },
  pending: { label: t("integration.status.pending"), variant: "info" },
};

const eventStatusVariant: Record<
  IntegrationEvent["status"],
  "success" | "warning" | "danger" | "info" | "default"
> = {
  success: "success",
  failed: "danger",
  ignored: "default",
};

/* ------------------------------------------------------------------ */
/* 时间格式化                                                          */
/* ------------------------------------------------------------------ */

function formatTime(value?: string | null): string {
  if (!value) return t("integration.neverSynced");
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toLocaleString();
}

/* ------------------------------------------------------------------ */
/* 连接弹窗                                                            */
/* ------------------------------------------------------------------ */

const showConnect = ref(false);
const connectProvider = ref<IntegrationProvider | "">("");
const saving = ref(false);

/** 当前选中 provider 的元数据 */
const activeProvider = computed<ProviderMeta | undefined>(() =>
  providers.value.find((p) => p.provider === connectProvider.value),
);

/** 凭证字段值（key → value） */
const credentialForm = reactive<Record<string, string>>({});
/** 配置字段值（key → value） */
const configForm = reactive<Record<string, string>>({});

/** 通用配置字段（repo / channel / project），按 provider 过滤掉已在凭证字段中的 key */
const configFieldDefs = computed(() => {
  const p = activeProvider.value;
  if (!p) return [];
  const credKeys = new Set(p.fields.map((f) => f.key));
  const all = [
    { key: "repo", label: t("integration.config.repo") },
    { key: "channel", label: t("integration.config.channel") },
    { key: "project", label: t("integration.config.project") },
  ];
  return all.filter((c) => !credKeys.has(c.key));
});

function openConnect(provider?: IntegrationProvider) {
  connectProvider.value = provider ?? providers.value[0]?.provider ?? "";
  // 重置表单
  for (const k of Object.keys(credentialForm)) delete credentialForm[k];
  for (const k of Object.keys(configForm)) delete configForm[k];
  showConnect.value = true;
}

function closeConnect() {
  showConnect.value = false;
  connectProvider.value = "";
}

/** 校验必填凭证字段 */
function validateConnectForm(): boolean {
  const p = activeProvider.value;
  if (!p) return false;
  const missing = p.fields.filter((f) => f.required && !credentialForm[f.key]?.trim());
  if (missing.length > 0) {
    toast.warning(t("integration.fillRequiredFields"));
    return false;
  }
  return true;
}

async function submitConnect() {
  const p = activeProvider.value;
  if (!p || saving.value) return;
  if (!validateConnectForm()) return;

  saving.value = true;
  try {
    const credentials: Record<string, string> = {};
    for (const f of p.fields) {
      const v = credentialForm[f.key]?.trim();
      if (v) credentials[f.key] = v;
    }
    const config: Record<string, string> = {};
    for (const c of configFieldDefs.value) {
      const v = configForm[c.key]?.trim();
      if (v) config[c.key] = v;
    }

    await integrationApi.create(workspaceId.value, {
      provider: p.provider,
      name: credentialForm.name?.trim() || p.name,
      credentials,
      config,
    });
    toast.success(t("integration.connectSuccess"));
    closeConnect();
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("integration.connectFailed"));
  } finally {
    saving.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 行内操作                                                            */
/* ------------------------------------------------------------------ */

const testingId = ref<number | null>(null);
const syncingId = ref<number | null>(null);

async function testConnection(it: Integration) {
  if (testingId.value) return;
  testingId.value = it.id;
  try {
    const r = await integrationApi.test(workspaceId.value, it.id);
    if (r.ok) toast.success(r.message || t("integration.testSuccess"));
    else toast.error(r.message || t("integration.testFailed"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("integration.testFailed"));
  } finally {
    testingId.value = null;
  }
}

async function syncNow(it: Integration) {
  if (syncingId.value) return;
  syncingId.value = it.id;
  try {
    const r = await integrationApi.sync(workspaceId.value, it.id);
    toast.success(r.message || t("integration.syncSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("integration.syncFailed"));
  } finally {
    syncingId.value = null;
  }
}

async function togglePause(it: Integration) {
  try {
    if (it.status === "paused") await integrationApi.resume(workspaceId.value, it.id);
    else await integrationApi.pause(workspaceId.value, it.id);
    toast.success(it.status === "paused" ? t("integration.resumeSuccess") : t("integration.pauseSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("common.failed"));
  }
}

async function disconnect(it: Integration) {
  if (!await confirm({ title: "断开集成", message: t("integration.disconnectConfirm", { name: it.name }), confirmText: "断开", variant: "danger" })) return;
  try {
    await integrationApi.disconnect(workspaceId.value, it.id);
    toast.success(t("integration.disconnectSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("common.failed"));
  }
}

async function remove(it: Integration) {
  if (!await confirm({ title: "删除集成", message: t("integration.deleteConfirm", { name: it.name }), confirmText: "删除", variant: "danger" })) return;
  try {
    await integrationApi.remove(workspaceId.value, it.id);
    toast.success(t("integration.deleteSuccess"));
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("common.failed"));
  }
}

/* ------------------------------------------------------------------ */
/* 事件日志抽屉                                                        */
/* ------------------------------------------------------------------ */

const showEvents = ref(false);
const eventsLoading = ref(false);
const eventsError = ref("");
const events = ref<IntegrationEvent[]>([]);
const activeEventIntegration = ref<Integration | null>(null);
const expandedEvent = ref<number | null>(null);

async function openEvents(it: Integration) {
  activeEventIntegration.value = it;
  showEvents.value = true;
  eventsLoading.value = true;
  eventsError.value = "";
  events.value = [];
  try {
    const r = await integrationApi.listEvents(workspaceId.value, it.id, { limit: 50 });
    events.value = r.items ?? [];
  } catch (err: unknown) {
    eventsError.value = err instanceof Error ? err.message : t("integration.eventsLoadFailed");
  } finally {
    eventsLoading.value = false;
  }
}

function togglePayload(ev: IntegrationEvent) {
  expandedEvent.value = expandedEvent.value === ev.id ? null : ev.id;
}

/** payload 预览（截断 + 尝试 JSON 格式化） */
function payloadPreview(ev: IntegrationEvent): string {
  if (!ev.payload) return "";
  try {
    const parsed = JSON.parse(ev.payload);
    const pretty = JSON.stringify(parsed, null, 2);
    return pretty.length > 800 ? `${pretty.slice(0, 800)}…` : pretty;
  } catch {
    return ev.payload.length > 800 ? `${ev.payload.slice(0, 800)}…` : ev.payload;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <!-- 头部 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">{{ t("integration.title") }}</h1>
        <p class="mt-1 text-sm text-[var(--text-tertiary)]">{{ t("integration.subtitle") }}</p>
      </div>
      <AppButton variant="primary" @click="openConnect()">
        {{ t("integration.addIntegration") }}
      </AppButton>
    </div>

    <!-- 加载 -->
    <div v-if="loading" class="space-y-3">
      <AppSkeleton class="h-24 w-full" />
      <AppSkeleton class="h-24 w-full" />
      <AppSkeleton class="h-24 w-full" />
    </div>

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- Marketplace 网格 -->
      <section>
        <h2 class="mb-3 text-base font-semibold text-[var(--text-primary)]">
          {{ t("integration.marketplace") }}
        </h2>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="p in providers"
            :key="p.provider"
            class="rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4 transition hover:border-[var(--border-default)]"
          >
            <div class="flex items-start justify-between">
              <div class="flex items-center gap-3">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-md bg-[var(--surface-2)] text-sm font-semibold text-[var(--text-primary)]"
                >
                  {{ p.name.slice(0, 1) }}
                </div>
                <div>
                  <div class="text-sm font-medium text-[var(--text-primary)]">{{ p.name }}</div>
                  <div class="mt-0.5 flex items-center gap-1.5">
                    <AppBadge
                      v-if="connectedByProvider[p.provider]?.length"
                      variant="success"
                    >
                      {{ t("integration.connected") }}
                    </AppBadge>
                    <AppBadge v-else variant="default">{{ t("integration.notConnected") }}</AppBadge>
                    <span
                      v-if="p.supports_sync"
                      class="text-xs text-[var(--text-tertiary)]"
                    >
                      · {{ t("integration.supportsSync") }}
                    </span>
                    <span
                      v-if="p.supports_notify"
                      class="text-xs text-[var(--text-tertiary)]"
                    >
                      · {{ t("integration.supportsNotify") }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-[var(--text-secondary)]">
              {{ p.description }}
            </p>
            <div class="mt-3 flex items-center justify-end gap-2">
              <AppButton
                v-if="connectedByProvider[p.provider]?.length"
                variant="secondary"
                size="sm"
                @click="openConnect(p.provider)"
              >
                {{ t("integration.addMore") }}
              </AppButton>
              <AppButton v-else variant="primary" size="sm" @click="openConnect(p.provider)">
                {{ t("integration.connect") }}
              </AppButton>
            </div>
          </div>
        </div>
      </section>

      <!-- 已连接集成 -->
      <section>
        <h2 class="mb-3 text-base font-semibold text-[var(--text-primary)]">
          {{ t("integration.connectedTitle") }}
        </h2>

        <AppEmptyState
          v-if="integrations.length === 0"
          :title="t('integration.empty.title')"
          :description="t('integration.empty.description')"
        />

        <div v-else class="space-y-3">
          <div
            v-for="it in integrations"
            :key="it.id"
            class="rounded-lg border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-[var(--text-primary)]">{{ it.name }}</span>
                  <AppBadge :variant="statusMeta[it.status]?.variant ?? 'default'">
                    {{ statusMeta[it.status]?.label ?? it.status }}
                  </AppBadge>
                </div>
                <div class="mt-1 text-xs text-[var(--text-tertiary)]">
                  {{ t("integration.provider") }}: {{ it.provider }} ·
                  {{ t("integration.lastSync") }}: {{ formatTime(it.last_sync_at) }}
                </div>
                <div
                  v-if="it.last_error"
                  class="mt-1 truncate text-xs text-[var(--danger-500)]"
                >
                  {{ t("integration.lastError") }}: {{ it.last_error }}
                </div>
              </div>

              <div class="flex flex-shrink-0 flex-wrap items-center gap-2">
                <AppButton
                  variant="ghost"
                  size="sm"
                  :loading="testingId === it.id"
                  @click="testConnection(it)"
                >
                  {{ t("integration.testConnection") }}
                </AppButton>
                <AppButton
                  variant="ghost"
                  size="sm"
                  :loading="syncingId === it.id"
                  @click="syncNow(it)"
                >
                  {{ t("integration.syncNow") }}
                </AppButton>
                <AppButton variant="ghost" size="sm" @click="openEvents(it)">
                  {{ t("integration.eventLog") }}
                </AppButton>
                <AppButton variant="ghost" size="sm" @click="togglePause(it)">
                  {{ it.status === "paused" ? t("integration.resume") : t("integration.pause") }}
                </AppButton>
                <AppButton variant="ghost" size="sm" @click="disconnect(it)">
                  {{ t("integration.disconnect") }}
                </AppButton>
                <AppButton variant="danger" size="sm" @click="remove(it)">
                  {{ t("common.delete") }}
                </AppButton>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>

    <!-- 连接弹窗 -->
    <AppModal
      :visible="showConnect"
      :title="t('integration.connectTitle')"
      width="560px"
      @close="closeConnect"
    >
      <div class="space-y-4">
        <!-- provider 选择 -->
        <div>
          <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
            {{ t("integration.provider") }}
          </label>
          <select
            v-model="connectProvider"
            class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
          >
            <option
              v-for="p in providers"
              :key="p.provider"
              :value="p.provider"
            >
              {{ p.name }}
            </option>
          </select>
        </div>

        <!-- 动态凭证字段 -->
        <template v-if="activeProvider">
          <div>
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ t("integration.name") }}
            </label>
            <input
              v-model="credentialForm.name"
              type="text"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
              :placeholder="activeProvider.name"
            />
          </div>

          <div v-for="f in activeProvider.fields" :key="f.key">
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ f.label }}
              <span v-if="f.required" class="text-[var(--danger-500)]">*</span>
            </label>
            <input
              v-model="credentialForm[f.key]"
              :type="f.secret ? 'password' : 'text'"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
              :placeholder="f.placeholder || ''"
              :autocomplete="f.secret ? 'new-password' : 'off'"
            />
          </div>

          <!-- 配置字段 -->
          <div v-for="c in configFieldDefs" :key="c.key">
            <label class="mb-1 block text-xs text-[var(--text-tertiary)]">
              {{ c.label }}
            </label>
            <input
              v-model="configForm[c.key]"
              type="text"
              class="w-full rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)]"
            />
          </div>

          <p class="text-xs text-[var(--text-tertiary)]">
            {{ t("integration.credentialHint") }}
          </p>
              </div>

      <template #footer>
        <AppButton variant="secondary" @click="closeConnect">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton variant="primary" :loading="saving" @click="submitConnect">
          {{ t("integration.connect") }}
        </AppButton>
      
    </AppModal>

    <!-- 事件日志抽屉 -->
    <AppModal
      :visible="showEvents"
      :title="t('integration.eventLog')"
      width="720px"
      @close="showEvents = false"
    >
      <div class="space-y-3">
        <div v-if="activeEventIntegration" class="text-xs text-[var(--text-tertiary)]">
          {{ activeEventIntegration.name }} · {{ activeEventIntegration.provider }}
        </div>

        <AppLoadingState v-if="eventsLoading" :text="t('common.loading')" />
        <AppErrorState v-else-if="eventsError" :message="eventsError" @retry="openEvents(activeEventIntegration!)" />
        <AppEmptyState
          v-else-if="events.length === 0"
          :title="t('integration.eventsEmpty')"
        />

        <div v-else class="space-y-2">
          <div
            v-for="ev in events"
            :key="ev.id"
            class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] p-3"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <AppBadge variant="info">
                  {{ ev.direction === "inbound"
                    ? t("integration.direction.inbound")
                    : t("integration.direction.outbound") }}
                </AppBadge>
                <span class="text-sm font-medium text-[var(--text-primary)]">{{ ev.event_type }}</span>
                <AppBadge :variant="eventStatusVariant[ev.status] ?? 'default'">
                  {{ ev.status }}
                </AppBadge>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs text-[var(--text-tertiary)]">{{ formatTime(ev.created_at) }}</span>
                <button
                  v-if="ev.payload"
                  class="text-xs text-[var(--brand-600)]"
                  @click="togglePayload(ev)"
                >
                  {{ expandedEvent === ev.id ? t("integration.hidePayload") : t("integration.viewPayload") }}
                </button>
              </div>
            </div>

            <div
              v-if="ev.error"
              class="mt-1 truncate text-xs text-[var(--danger-500)]"
            >
              {{ ev.error }}
            </div>

            <pre
              v-if="ev.payload && expandedEvent === ev.id"
              class="mt-2 max-h-48 overflow-auto rounded bg-[var(--surface-2)] p-2 text-xs text-[var(--text-secondary)]"
            >{{ payloadPreview(ev) }}</pre>
          </div>
        </div>
      </div>
    </AppModal>
  </div>

