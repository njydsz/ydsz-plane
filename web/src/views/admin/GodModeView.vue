<script setup lang="ts">
/**
 * GodModeView — 实例管理面板（God Mode）。
 *
 * 聚合展示并管理实例级配置：
 *  - 系统概览（版本 / 版本类型 / 运行时长 / 环境 / 功能开关 / 组件健康）
 *  - 许可证（edition / plan / 席位 / 到期 / 功能清单）
 *  - 用量统计（空间 / 用户 / 项目 / 工作项 / 评论 / 存储 / API 调用）
 *  - 实例设置（SMTP / 认证 / 安全策略 / 遥测 / 界面）
 *
 * 鉴权：后端要求实例管理员（X-Instance-Admin-Key 或 YDSZ_INSTANCE_ADMINS 白名单）。
 * 若部署配置了 YDSZ_INSTANCE_ADMIN_KEY，可设置下方 ADMIN_KEY 常量，
 * 由 godmode.ts 客户端随每个请求注入 X-Instance-Admin-Key 头。
 *
 * 加载策略：概览类数据（status / license / usage）并行拉取；
 * 设置单独拉取，避免设置请求失败时整页不可用。
 */
import { computed, onMounted, ref } from "vue";

import {
  createGodModeApi,
  type InstanceSettings,
  type LicenseInfo,
  type SystemStatus,
  type UsageStats,
} from "@/api/services/godmode";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
  DynamicIcon,
} from "@/components";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* 客户端                                                              */
/* ------------------------------------------------------------------ */

/** 实例管理员密钥（部署时按需填写；留空则走后端白名单鉴权） */
const INLINE_ADMIN_KEY = "";

/** 实例管理员密钥解析：
 *  1) 代码内联常量 INLINE_ADMIN_KEY（部署时确定）；
 *  2) Vite 环境变量 VITE_INSTANCE_ADMIN_KEY（.env.local / 构建注入）；
 *  均未配置时留空，依赖后端 YDSZ_INSTANCE_ADMINS 用户白名单鉴权。
 *  注意：不要将真实密钥提交到仓库，请通过环境变量注入。 */
const ADMIN_KEY: string | undefined =
  INLINE_ADMIN_KEY || import.meta.env.VITE_INSTANCE_ADMIN_KEY || undefined;

const api = createGodModeApi(ADMIN_KEY);

const { t } = useTranslation();

/* ------------------------------------------------------------------ */
/* 状态                                                                */
/* ------------------------------------------------------------------ */

/** 概览类加载态 */
const loadingOverview = ref(true);
const overviewError = ref("");

/** 设置类加载/保存态 */
const loadingSettings = ref(true);
const settingsError = ref("");
const saving = ref(false);
const resetting = ref(false);

/** 测试 SMTP */
const showSmtpModal = ref(false);
const smtpTo = ref("");
const testingSmtp = ref(false);

/** 重置确认 */
const showResetModal = ref(false);

/** 数据 */
const status = ref<SystemStatus | null>(null);
const license = ref<LicenseInfo | null>(null);
const usage = ref<UsageStats | null>(null);
const settings = ref<InstanceSettings | null>(null);

/** 表单草稿（与已加载设置解耦，避免编辑时被刷新覆盖） */
const form = ref<InstanceSettings | null>(null);

/** 标签页 */
const activeTab = ref<"overview" | "settings">("overview");

/* ------------------------------------------------------------------ */
/* 派生                                                                */
/* ------------------------------------------------------------------ */

const tabs = computed(() => [
  { key: "overview" as const, label: t("godMode.tab.overview"), icon: "Activity" },
  { key: "settings" as const, label: t("godMode.tab.settings"), icon: "Settings" },
]);

const hasSettings = computed(() => form.value !== null);

/** 是否被修改（与已加载设置做 JSON 比较） */
const dirty = computed(() => {
  if (!settings.value || !form.value) return false;
  // 密码不参与比较：后端永不回显，输入密码不应被视为“未保存”的唯一原因
  const a = JSON.stringify({ ...settings.value, smtp: { ...settings.value.smtp, password: "" } });
  const b = JSON.stringify({ ...form.value, smtp: { ...form.value.smtp, password: "" } });
  return a !== b;
});

const featureEntries = computed(() => {
  const map = status.value?.features ?? {};
  return Object.entries(map).map(([name, enabled]) => ({ name, enabled }));
});

const seatsText = computed(() => {
  const lic = license.value;
  if (!lic) return "-";
  if (!lic.seats) return `${lic.used_seats} / ∞`;
  return `${lic.used_seats} / ${lic.seats}`;
});

const seatsRatio = computed(() => {
  const lic = license.value;
  if (!lic || !lic.seats) return 0;
  return Math.min(100, Math.round((lic.used_seats / lic.seats) * 100));
});

const usageCards = computed(() => {
  const u = usage.value;
  if (!u) return [];
  return [
    { label: t("godMode.usage.workspaces"), value: u.total_workspaces, icon: "Building2" },
    { label: t("godMode.usage.users"), value: u.total_users, icon: "Users" },
    { label: t("godMode.usage.activeUsers30d"), value: u.active_users_30d, icon: "UserCircle" },
    { label: t("godMode.usage.projects"), value: u.total_projects, icon: "FolderKanban" },
    { label: t("godMode.usage.workItems"), value: u.total_work_items, icon: "ClipboardList" },
    { label: t("godMode.usage.comments"), value: u.total_comments, icon: "MessagesSquare" },
    { label: t("godMode.usage.storage"), value: `${u.storage_used_gb.toFixed(2)} GB`, icon: "Database" },
    { label: t("godMode.usage.apiCalls30d"), value: u.api_calls_30d, icon: "BarChart3" },
  ];
});

/** allowed_email_domains: string[] ↔ 逗号分隔文本 的双向映射 */
const domainsText = computed<string>({
  get: () => (form.value?.auth.allowed_email_domains ?? []).join(", "),
  set: (v: string) => {
    if (!form.value) return;
    form.value.auth.allowed_email_domains = v
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
  },
});

/* ------------------------------------------------------------------ */
/* 格式化                                                              */
/* ------------------------------------------------------------------ */

function fmtDateTime(input?: string | null): string {
  if (!input) return "-";
  const d = new Date(input);
  if (Number.isNaN(d.getTime())) return String(input);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours(),
  )}:${pad(d.getMinutes())}`;
}

function componentVariant(s: string): "success" | "warning" | "danger" | "default" {
  const v = s.toLowerCase();
  if (["ok", "healthy", "up", "pass", "passing", "green"].includes(v)) return "success";
  if (["warn", "warning", "degraded", "yellow", "slow"].includes(v)) return "warning";
  if (["error", "fail", "failed", "down", "critical", "red"].includes(v)) return "danger";
  return "default";
}

function licenseVariant(s: string): "success" | "warning" | "danger" | "info" | "default" {
  const v = s.toLowerCase();
  if (["active", "valid", "ok"].includes(v)) return "success";
  if (["expired", "revoked", "invalid", "suspended"].includes(v)) return "danger";
  if (["trial", "grace", "grace_period"].includes(v)) return "info";
  if (["warning", "expiring"].includes(v)) return "warning";
  return "default";
}

/* ------------------------------------------------------------------ */
/* 数据加载                                                            */
/* ------------------------------------------------------------------ */

async function loadOverview() {
  loadingOverview.value = true;
  overviewError.value = "";
  try {
    const [s, l, u] = await Promise.all([
      api.getStatus(),
      api.getLicense(),
      api.getUsage(),
    ]);
    status.value = s;
    license.value = l;
    usage.value = u;
  } catch (err: unknown) {
    if (
      err instanceof Error &&
      (err as unknown as { statusCode?: number }).statusCode === 403
    ) {
      overviewError.value = t("godMode.error.forbidden");
    } else {
      overviewError.value = err instanceof Error ? err.message : t("godMode.error.loadFailed");
    }
  } finally {
    loadingOverview.value = false;
  }
}

async function loadSettings() {
  loadingSettings.value = true;
  settingsError.value = "";
  try {
    const data = await api.getSettings();
    settings.value = data;
    // 深拷贝为草稿，后续编辑只改草稿
    form.value = JSON.parse(JSON.stringify(data)) as InstanceSettings;
  } catch (err: unknown) {
    settingsError.value = err instanceof Error ? err.message : t("godMode.error.loadFailed");
  } finally {
    loadingSettings.value = false;
  }
}

async function loadAll() {
  await Promise.all([loadOverview(), loadSettings()]);
}

/* ------------------------------------------------------------------ */
/* 保存 / 重置                                                         */
/* ------------------------------------------------------------------ */

async function save() {
  if (!form.value || saving.value) return;
  saving.value = true;
  try {
    const updated = await api.updateSettings({
      smtp: { ...form.value.smtp },
      auth: { ...form.value.auth, allowed_email_domains: [...form.value.auth.allowed_email_domains] },
      security: { ...form.value.security },
      telemetry: { ...form.value.telemetry },
      ui: { ...form.value.ui },
    });
    settings.value = updated;
    // 密码后端不回显（json:"-"），保留用户刚输入的值到本地草稿
    const password = form.value.smtp.password;
    form.value = JSON.parse(JSON.stringify(updated)) as InstanceSettings;
    form.value.smtp.password = password;
    toast.success(t("godMode.message.saveSuccess"));
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("godMode.message.saveFailed"));
  } finally {
    saving.value = false;
  }
}

function askReset() {
  showResetModal.value = true;
}

async function confirmReset() {
  if (resetting.value) return;
  resetting.value = true;
  try {
    const reset = await api.resetSettings();
    settings.value = reset;
    form.value = JSON.parse(JSON.stringify(reset)) as InstanceSettings;
    showResetModal.value = false;
    toast.success(t("godMode.message.resetSuccess"));
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("godMode.message.resetFailed"));
  } finally {
    resetting.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* SMTP 测试                                                           */
/* ------------------------------------------------------------------ */

function openSmtpModal() {
  smtpTo.value = "";
  showSmtpModal.value = true;
}

async function submitSmtpTest() {
  const to = smtpTo.value.trim();
  if (!to) {
    toast.error(t("godMode.smtp.toRequired"));
    return;
  }
  testingSmtp.value = true;
  try {
    const res = await api.testSmtp(to);
    if (res.ok) {
      toast.success(res.message || t("godMode.smtp.testSuccess"));
    } else {
      toast.error(res.message || t("godMode.smtp.testFailed"));
    }
    showSmtpModal.value = false;
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : t("godMode.smtp.testFailed"));
  } finally {
    testingSmtp.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

onMounted(loadAll);
</script>

<template>
  <div class="gm">
    <!-- ===== 页头 ===== -->
    <header class="gm__header">
      <div class="gm__header-main">
        <div class="gm__title-wrap">
          <DynamicIcon name="ShieldCheck" :size="22" class="gm__title-icon" />
          <div>
            <h1 class="gm__title">{{ t("godMode.title") }}</h1>
            <p class="gm__subtitle">{{ t("godMode.subtitle") }}</p>
          </div>
        </div>
        <div class="gm__header-actions">
          <AppButton
            variant="secondary"
            size="sm"
            :loading="loadingOverview"
            @click="loadAll"
          >
            {{ t("godMode.action.refresh") }}
          </AppButton>
          <template v-if="activeTab === 'settings' && hasSettings">
            <AppButton
              variant="secondary"
              size="sm"
              :disabled="saving || resetting"
              @click="askReset"
            >
              {{ t("godMode.action.reset") }}
            </AppButton>
            <AppButton
              variant="primary"
              size="sm"
              :loading="saving"
              :disabled="!dirty"
              @click="save"
            >
              {{ t("godMode.action.save") }}
            </AppButton>
          </template>
        </div>
      </div>
    </header>

      <!-- 标签页 -->
      <nav class="gm__tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="gm__tab"
          :class="{ 'is-active': activeTab === tab.key }"
          role="tab"
          :aria-selected="activeTab === tab.key"
          @click="activeTab = tab.key"
        >
          <DynamicIcon :name="tab.icon" :size="15" />
          {{ tab.label }}
        </button>
      </nav>

    <!-- ================= 概览 ================= -->
    <div v-if="activeTab === 'overview'" class="gm__body">
      <AppLoadingState v-if="loadingOverview" :text="t('godMode.loading.overview')" />
      <AppErrorState v-else-if="overviewError" :message="overviewError" @retry="loadOverview" />
      <AppEmptyState
        v-else-if="!status && !license && !usage"
        :title="t('godMode.empty.title')"
        :description="t('godMode.empty.description')"
      />

      <template v-else>
        <!-- ===== 系统概览 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="Activity" :size="16" />
            {{ t("godMode.status.title") }}
          </h2>

          <div class="gm__grid">
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.status.version") }}</span>
              <span class="gm__metric-value">{{ status?.version || "-" }}</span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.status.edition") }}</span>
              <span class="gm__metric-value">
                {{ status?.edition || "-" }}
                <AppBadge v-if="status?.edition" variant="brand">{{ status.edition }}</AppBadge>
              </span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.status.uptime") }}</span>
              <span class="gm__metric-value">{{ status?.uptime || "-" }}</span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.status.environment") }}</span>
              <span class="gm__metric-value">
                <AppBadge v-if="status?.environment" variant="info">{{ status.environment }}</AppBadge>
                <span v-else>-</span>
              </span>
            </div>
          </div>

          <dl class="gm__info">
            <div class="gm__info-row">
              <dt>{{ t("godMode.status.startedAt") }}</dt>
              <dd>{{ fmtDateTime(status?.started_at) }}</dd>
            </div>
            <div class="gm__info-row">
              <dt>{{ t("godMode.status.uptimeSeconds") }}</dt>
              <dd>{{ status?.uptime_seconds ?? "-" }}</dd>
            </div>
          </dl>

          <!-- 功能开关 -->
          <div class="gm__subhead">{{ t("godMode.status.features") }}</div>
          <div v-if="featureEntries.length" class="gm__chips">
            <span
              v-for="f in featureEntries"
              :key="f.name"
              class="gm__chip"
              :class="f.enabled ? 'is-on' : 'is-off'"
            >
              <span class="gm__chip-dot" />
              {{ f.name }}
            </span>
          </div>
          <p v-else class="gm__muted">{{ t("godMode.empty.features") }}</p>

          <!-- 组件健康 -->
          <div class="gm__subhead">{{ t("godMode.status.components") }}</div>
          <div v-if="status?.components?.length" class="gm__components">
            <div
              v-for="c in status.components"
              :key="c.name"
              class="gm__component"
            >
              <div class="gm__component-head">
                <span class="gm__component-name">{{ c.name }}</span>
                <AppBadge :variant="componentVariant(c.status)">{{ c.status }}</AppBadge>
              </div>
              <div class="gm__component-meta">
                <span v-if="c.latency">
                  {{ t("godMode.status.latency") }}: {{ c.latency }}
                </span>
                <span v-if="c.message" class="gm__component-msg">{{ c.message }}</span>
              </div>
            </div>
          </div>
          <p v-else class="gm__muted">{{ t("godMode.empty.components") }}</p>
        </section>

        <!-- ===== 许可证 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="KeyRound" :size="16" />
            {{ t("godMode.license.title") }}
          </h2>

          <div class="gm__grid">
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.license.edition") }}</span>
              <span class="gm__metric-value">{{ license?.edition || "-" }}</span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.license.plan") }}</span>
              <span class="gm__metric-value">{{ license?.plan || "-" }}</span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.license.status") }}</span>
              <span class="gm__metric-value">
                <AppBadge v-if="license?.status" :variant="licenseVariant(license.status)">
                  {{ license.status }}
                </AppBadge>
                <span v-else>-</span>
              </span>
            </div>
            <div class="gm__metric">
              <span class="gm__metric-label">{{ t("godMode.license.seats") }}</span>
              <span class="gm__metric-value">{{ seatsText }}</span>
            </div>
          </div>

          <div v-if="license?.seats" class="gm__progress">
            <div class="gm__progress-bar">
              <div
                class="gm__progress-fill"
                :class="{ 'is-warn': seatsRatio >= 80 }"
                :style="{ width: `${seatsRatio}%` }"
              />
            </div>
            <span class="gm__progress-text">{{ seatsRatio }}%</span>
          </div>

          <dl class="gm__info">
            <div class="gm__info-row">
              <dt>{{ t("godMode.license.issuedAt") }}</dt>
              <dd>{{ fmtDateTime(license?.issued_at) }}</dd>
            </div>
            <div class="gm__info-row">
              <dt>{{ t("godMode.license.expiresAt") }}</dt>
              <dd>{{ fmtDateTime(license?.expires_at) }}</dd>
            </div>
          </dl>

          <div class="gm__subhead">{{ t("godMode.license.features") }}</div>
          <div v-if="license?.features?.length" class="gm__chips">
            <span v-for="f in license.features" :key="f" class="gm__chip is-on">
              <span class="gm__chip-dot" />{{ f }}
            </span>
          </div>
          <p v-else class="gm__muted">{{ t("godMode.empty.features") }}</p>
        </section>

        <!-- ===== 用量统计 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="BarChart3" :size="16" />
            {{ t("godMode.usage.title") }}
          </h2>

          <div class="gm__grid">
            <div
              v-for="card in usageCards"
              :key="card.label"
              class="gm__usage-card"
            >
              <div class="gm__usage-icon">
                <DynamicIcon :name="card.icon" :size="18" />
              </div>
              <div class="gm__usage-body">
                <span class="gm__usage-value">{{ card.value }}</span>
                <span class="gm__usage-label">{{ card.label }}</span>
              </div>
            </div>
          </div>
        </section>
      </template>
    </div>

    <!-- ================= 设置 ================= -->
    <div v-else class="gm__body">
      <AppLoadingState v-if="loadingSettings" :text="t('godMode.loading.settings')" />
      <AppErrorState v-else-if="settingsError" :message="settingsError" @retry="loadSettings" />
      <AppEmptyState
        v-else-if="!hasSettings"
        :title="t('godMode.empty.title')"
        :description="t('godMode.empty.description')"
      />

      <template v-else-if="form">
        <p class="gm__hint">
          {{ t("godMode.settings.hint") }}
          <span v-if="form.updated_at">
            · {{ t("godMode.settings.updatedAt") }}: {{ fmtDateTime(form.updated_at) }}
            <span v-if="form.updated_by">
              · {{ t("godMode.settings.updatedBy") }}: #{{ form.updated_by }}
            </span>
          </span>
          <span v-if="dirty" class="gm__dirty"> · {{ t("godMode.settings.unsaved") }}</span>
        </p>

        <!-- ===== SMTP ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="Mail" :size="16" />
            {{ t("godMode.smtp.title") }}
            <AppBadge :variant="form.smtp.enabled ? 'success' : 'default'">
              {{ form.smtp.enabled ? t("godMode.value.enabled") : t("godMode.value.disabled") }}
            </AppBadge>
          </h2>

          <div class="gm__form">
            <label class="gm__toggle">
              <span class="gm__toggle-label">{{ t("godMode.smtp.enabled") }}</span>
              <input v-model="form.smtp.enabled" type="checkbox" class="gm__switch" />
            </label>

            <div class="gm__form-row">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.smtp.host") }}</span>
                <input
                  v-model="form.smtp.host"
                  type="text"
                  class="gm__input"
                  placeholder="smtp.example.com"
                />
              </label>
              <label class="gm__field gm__field--narrow">
                <span class="gm__field-label">{{ t("godMode.smtp.port") }}</span>
                <input
                  v-model.number="form.smtp.port"
                  type="number"
                  class="gm__input"
                  min="1"
                  max="65535"
                />
              </label>
            </div>

            <div class="gm__form-row">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.smtp.username") }}</span>
                <input
                  v-model="form.smtp.username"
                  type="text"
                  class="gm__input"
                  autocomplete="off"
                />
              </label>
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.smtp.password") }}</span>
                <input
                  v-model="form.smtp.password"
                  type="password"
                  class="gm__input"
                  :placeholder="t('godMode.smtp.passwordPlaceholder')"
                  autocomplete="new-password"
                />
              </label>
            </div>

            <div class="gm__form-row">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.smtp.fromEmail") }}</span>
                <input
                  v-model="form.smtp.from_email"
                  type="email"
                  class="gm__input"
                  placeholder="no-reply@example.com"
                />
              </label>
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.smtp.fromName") }}</span>
                <input v-model="form.smtp.from_name" type="text" class="gm__input" />
              </label>
            </div>

            <div class="gm__form-actions">
              <label class="gm__toggle gm__toggle--inline">
                <input v-model="form.smtp.use_tls" type="checkbox" class="gm__switch" />
                <span class="gm__toggle-label">{{ t("godMode.smtp.useTls") }}</span>
              </label>
              <AppButton
                variant="secondary"
                size="sm"
                :disabled="!form.smtp.enabled"
                @click="openSmtpModal"
              >
                {{ t("godMode.smtp.test") }}
              </AppButton>
            </div>
          </div>
        </section>

        <!-- ===== 认证设置 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="UserCircle" :size="16" />
            {{ t("godMode.auth.title") }}
          </h2>

          <div class="gm__form">
            <div class="gm__toggle-grid">
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.allowEmailPassword") }}</span>
                <input v-model="form.auth.allow_email_password" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.allowOidc") }}</span>
                <input v-model="form.auth.allow_oidc" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.allowSaml") }}</span>
                <input v-model="form.auth.allow_saml" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.allowLdap") }}</span>
                <input v-model="form.auth.allow_ldap" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.forceSso") }}</span>
                <input v-model="form.auth.force_sso" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.auth.allowSignup") }}</span>
                <input v-model="form.auth.allow_signup" type="checkbox" class="gm__switch" />
              </label>
            </div>

            <label class="gm__field">
              <span class="gm__field-label">{{ t("godMode.auth.allowedEmailDomains") }}</span>
              <input
                v-model="domainsText"
                type="text"
                class="gm__input"
                :placeholder="t('godMode.auth.allowedEmailDomainsPlaceholder')"
              />
            </label>

            <label class="gm__field gm__field--narrow">
              <span class="gm__field-label">{{ t("godMode.auth.sessionMinutes") }}</span>
              <input
                v-model.number="form.auth.session_minutes"
                type="number"
                class="gm__input"
                min="5"
              />
            </label>
          </div>
        </section>

        <!-- ===== 安全策略 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="ShieldCheck" :size="16" />
            {{ t("godMode.security.title") }}
          </h2>

          <div class="gm__form">
            <div class="gm__form-row gm__form-row--3">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.security.minPasswordLength") }}</span>
                <input
                  v-model.number="form.security.min_password_length"
                  type="number"
                  class="gm__input"
                  min="1"
                />
              </label>
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.security.maxLoginAttempts") }}</span>
                <input
                  v-model.number="form.security.max_login_attempts"
                  type="number"
                  class="gm__input"
                  min="1"
                />
              </label>
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.security.lockoutMinutes") }}</span>
                <input
                  v-model.number="form.security.lockout_minutes"
                  type="number"
                  class="gm__input"
                  min="1"
                />
              </label>
            </div>

            <div class="gm__form-row gm__form-row--3">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.security.auditRetentionDays") }}</span>
                <input
                  v-model.number="form.security.audit_retention_days"
                  type="number"
                  class="gm__input"
                  min="1"
                />
              </label>
              <label class="gm__toggle gm__toggle--inline">
                <input v-model="form.security.force_mfa" type="checkbox" class="gm__switch" />
                <span class="gm__toggle-label">{{ t("godMode.security.forceMfa") }}</span>
              </label>
            </div>
          </div>
        </section>

        <!-- ===== 遥测 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="Radio" :size="16" />
            {{ t("godMode.telemetry.title") }}
          </h2>

          <div class="gm__form">
            <label class="gm__toggle">
              <span class="gm__toggle-label">{{ t("godMode.telemetry.enabled") }}</span>
              <input v-model="form.telemetry.enabled" type="checkbox" class="gm__switch" />
            </label>
            <label class="gm__field">
              <span class="gm__field-label">{{ t("godMode.telemetry.endpoint") }}</span>
              <input
                v-model="form.telemetry.endpoint"
                type="text"
                class="gm__input"
                placeholder="https://telemetry.example.com/collect"
              />
            </label>
            <label class="gm__field">
              <span class="gm__field-label">{{ t("godMode.telemetry.instanceId") }}</span>
              <input
                v-model="form.telemetry.instance_id"
                type="text"
                class="gm__input gm__input--mono"
              />
            </label>
          </div>
        </section>

        <!-- ===== 界面 ===== -->
        <section class="gm__section">
          <h2 class="gm__section-title">
            <DynamicIcon name="Palette" :size="16" />
            {{ t("godMode.ui.title") }}
          </h2>

          <div class="gm__form">
            <div class="gm__toggle-grid">
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.ui.showVersion") }}</span>
                <input v-model="form.ui.show_version" type="checkbox" class="gm__switch" />
              </label>
              <label class="gm__toggle">
                <span class="gm__toggle-label">{{ t("godMode.ui.maintenanceMode") }}</span>
                <input v-model="form.ui.maintenance_mode" type="checkbox" class="gm__switch" />
              </label>
            </div>

            <div class="gm__form-row">
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.ui.brandName") }}</span>
                <input v-model="form.ui.brand_name" type="text" class="gm__input" />
              </label>
              <label class="gm__field">
                <span class="gm__field-label">{{ t("godMode.ui.logoUrl") }}</span>
                <input
                  v-model="form.ui.logo_url"
                  type="text"
                  class="gm__input"
                  placeholder="https://..."
                />
              </label>
            </div>

            <label class="gm__field">
              <span class="gm__field-label">{{ t("godMode.ui.maintenanceMessage") }}</span>
              <textarea
                v-model="form.ui.maintenance_message"
                class="gm__input gm__textarea"
                rows="3"
                :placeholder="t('godMode.ui.maintenanceMessagePlaceholder')"
              />
            </label>
          </div>
        </section>
      </template>
    </div>

    <!-- ===== SMTP 测试弹窗 ===== -->
    <AppModal
      :visible="showSmtpModal"
      :title="t('godMode.smtp.testTitle')"
      width="460px"
      @close="showSmtpModal = false"
    >
      <p class="gm__muted">{{ t("godMode.smtp.testHint") }}</p>
      <label class="gm__field">
        <span class="gm__field-label">{{ t("godMode.smtp.to") }}</span>
        <input
          v-model="smtpTo"
          type="email"
          class="gm__input"
          placeholder="admin@example.com"
          @keyup.enter="submitSmtpTest"
        />
      </label>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showSmtpModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton variant="primary" size="sm" :loading="testingSmtp" @click="submitSmtpTest">
          {{ t("godMode.smtp.test") }}
        </AppButton>
      </template>
    </AppModal>

    <!-- ===== 重置确认弹窗 ===== -->
    <AppModal
      :visible="showResetModal"
      :title="t('godMode.reset.title')"
      width="440px"
      @close="showResetModal = false"
    >
      <p class="gm__muted">{{ t("godMode.reset.confirm") }}</p>
      <template #footer>
        <AppButton variant="secondary" size="sm" @click="showResetModal = false">
          {{ t("common.cancel") }}
        </AppButton>
        <AppButton variant="danger" size="sm" :loading="resetting" @click="confirmReset">
          {{ t("godMode.action.reset") }}
        </AppButton>
      
      </template>
    </AppModal>
  </div>


</template>

<style scoped>
.gm {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ===== Header ===== */
.gm__header {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.gm__header-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.gm__title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gm__title-icon {
  color: var(--brand-600);
  flex-shrink: 0;
}

.gm__title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.gm__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.gm__header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

/* ===== Tabs ===== */
.gm__tabs {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border-subtle);
}

.gm__tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  border-bottom: 2px solid transparent;
  background: none;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}

.gm__tab:hover {
  color: var(--text-primary);
}

.gm__tab.is-active {
  color: var(--brand-600);
  border-bottom-color: var(--brand-500);
}

/* ===== Body ===== */
.gm__body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.gm__hint {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
}

.gm__dirty {
  color: var(--warning-500);
  font-weight: 500;
}

/* ===== Sections ===== */
.gm__section {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--surface-1);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.gm__section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.gm__subhead {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.gm__muted {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

/* ===== Metrics ===== */
.gm__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.gm__metric {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.gm__metric-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.gm__metric-value {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-all;
}

/* ===== Info list ===== */
.gm__info {
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.gm__info-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  font-size: 13px;
}

.gm__info-row dt {
  color: var(--text-tertiary);
  min-width: 120px;
  flex-shrink: 0;
}

.gm__info-row dd {
  margin: 0;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

/* ===== Chips ===== */
.gm__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.gm__chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 12px;
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  background: var(--surface-2);
}

.gm__chip-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  opacity: 0.5;
}

.gm__chip.is-on {
  color: var(--success-600);
  border-color: rgba(22, 163, 74, 0.3);
  background: rgba(22, 163, 74, 0.08);
}

.gm__chip.is-on .gm__chip-dot {
  opacity: 1;
}

.gm__chip.is-off {
  color: var(--text-tertiary);
}

/* ===== Components ===== */
.gm__components {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.gm__component {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
}

.gm__component-head {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.gm__component-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  font-family: var(--font-mono, monospace);
}

.gm__component-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: var(--text-tertiary);
  flex-wrap: wrap;
  justify-content: flex-end;
}

.gm__component-msg {
  color: var(--warning-500);
}

/* ===== License progress ===== */
.gm__progress {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gm__progress-bar {
  flex: 1;
  height: 6px;
  border-radius: 999px;
  background: var(--surface-3);
  overflow: hidden;
}

.gm__progress-fill {
  height: 100%;
  background: var(--brand-500);
  border-radius: 999px;
  transition: width 0.2s ease;
}

.gm__progress-fill.is-warn {
  background: var(--warning-500);
}

.gm__progress-text {
  font-size: 12px;
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
  min-width: 36px;
  text-align: right;
}

/* ===== Usage cards ===== */
.gm__usage-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.gm__usage-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border-radius: var(--radius-sm);
  background: var(--brand-50);
  color: var(--brand-600);
  flex-shrink: 0;
}

.gm__usage-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.gm__usage-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  word-break: break-all;
}

.gm__usage-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* ===== Form ===== */
.gm__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.gm__form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.gm__form-row--3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.gm__form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 10px;
  border-top: 1px dashed var(--border-subtle);
}

.gm__toggle-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.gm__toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-2);
  cursor: pointer;
  transition: border-color 0.15s;
}

.gm__toggle:hover {
  border-color: var(--border-default);
}

.gm__toggle--inline {
  border: none;
  background: none;
  padding: 0;
  justify-content: flex-start;
}

.gm__toggle-label {
  font-size: 13px;
  color: var(--text-primary);
}

.gm__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.gm__field--narrow {
  max-width: 200px;
}

.gm__field-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.gm__input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  box-sizing: border-box;
}

.gm__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.gm__input--mono {
  font-family: var(--font-mono, monospace);
}

.gm__textarea {
  resize: vertical;
  min-height: 68px;
}

/* Switch */
.gm__switch {
  position: relative;
  width: 42px;
  height: 22px;
  appearance: none;
  background: var(--surface-3);
  border: 1px solid var(--border-default);
  border-radius: 11px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  flex-shrink: 0;
  margin: 0;
}

.gm__switch:checked {
  background: var(--brand-500);
  border-color: var(--brand-500);
}

.gm__switch::before {
  content: "";
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--text-on-brand);
  transition: transform 0.2s;
}

.gm__switch:checked::before {
  transform: translateX(20px);
}

.gm__switch:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== Responsive ===== */
@media (max-width: 1024px) {
  .gm__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .gm__form-row--3 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .gm__grid,
  .gm__form-row,
  .gm__form-row--3,
  .gm__toggle-grid {
    grid-template-columns: 1fr;
  }
  .gm__info-row dt {
    min-width: 96px;
  }
  .gm__component {
    flex-direction: column;
    align-items: flex-start;
  }
  .gm__component-meta {
    justify-content: flex-start;
  }
  .gm__field--narrow {
    max-width: none;
  }
}
</style>
