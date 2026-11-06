<script setup lang="ts">
/**
 * ImporterView — 项目级「数据导入」页。
 *
 * 能力：
 *  - 来源选择卡片（CSV / Jira / Linear / Asana），展示说明与凭证要求
 *  - CSV 流程：选择文件 → 上传 → 展示结果
 *  - 平台流程：填写凭证（base_url / email / api_token / project_key / team_id）→ 导入 → 结果
 *  - 导入任务列表：状态徽章 / 来源 / 计数（imported / skipped / failed）/ 创建时间
 *  - 结果详情弹窗：total / imported / skipped / failed + 错误明细
 *  - 下载示例 CSV 模板（客户端 Blob 生成）
 */
import { computed, onMounted, reactive, ref, watch } from "vue";

import {
  importerApi,
  isPlatformSource,
  type ImportError,
  type ImportJob,
  type ImportSource,
  type SourceInfo,
} from "@/api/services/importer";
import {
  AppBadge,
  AppButton,
  AppEmptyState,
  AppErrorState,
  AppLoadingState,
  AppModal,
  AppPageContainer,
} from "@/components";
import DynamicIcon from "@/components/DynamicIcon.vue";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import { useTranslation } from "@/composables/useTranslation";
import { toast } from "@/lib/toast";

const { t } = useTranslation();
const { wsId, projectId, ready, error: ctxError } = useWorkspaceContext();

/* ------------------------------------------------------------------ */
/* 常量                                                                */
/* ------------------------------------------------------------------ */

/** 来源展示顺序与本地图标（卡片用）。 */
const SOURCE_ICONS: Record<string, string> = {
  csv: "FileSpreadsheet",
  jira: "Globe",
  linear: "Zap",
  asana: "Blocks",
};

/** 示例 CSV 模板表头（与后端 CSVImporter.normalizeHeader 对齐）。 */
const SAMPLE_CSV_COLUMNS = [
  "external_id",
  "title",
  "description",
  "type",
  "priority",
  "state",
  "assignee",
  "creator",
  "labels",
  "due_date",
  "parent_external_id",
  "estimate",
];

/** CSV 类型可选值（default_type）。 */
const DEFAULT_TYPE_OPTIONS: { value: string; labelKey: string }[] = [
  { value: "task", labelKey: "importer.typeTask" },
  { value: "requirement", labelKey: "importer.typeRequirement" },
  { value: "defect", labelKey: "importer.typeDefect" },
];

/** 任务状态 → 徽章变体。 */
const STATUS_VARIANT: Record<string, "default" | "brand" | "success" | "warning" | "danger"> = {
  pending: "default",
  running: "brand",
  completed: "success",
  failed: "danger",
};

/** 平台来源需要的凭证字段 → i18n label key / input type。 */
interface CredentialFieldDef {
  key: "base_url" | "email" | "api_token" | "project_key" | "team_id";
  labelKey: string;
  type: "text" | "email" | "password" | "url";
  placeholder: string;
  required: boolean;
}

const CREDENTIAL_FIELDS: Record<string, CredentialFieldDef[]> = {
  jira: [
    { key: "base_url", labelKey: "importer.fieldBaseUrl", type: "url", placeholder: "https://your-domain.atlassian.net", required: true },
    { key: "email", labelKey: "importer.fieldEmail", type: "email", placeholder: "you@company.com", required: true },
    { key: "api_token", labelKey: "importer.fieldApiToken", type: "password", placeholder: "ATATT3xFfGF0...", required: true },
    { key: "project_key", labelKey: "importer.fieldProjectKey", type: "text", placeholder: "PROJ", required: true },
  ],
  linear: [
    { key: "api_token", labelKey: "importer.fieldApiToken", type: "password", placeholder: "lin_api_...", required: true },
    { key: "team_id", labelKey: "importer.fieldTeamId", type: "text", placeholder: "可选，留空导入全部团队", required: false },
  ],
  asana: [
    { key: "api_token", labelKey: "importer.fieldApiToken", type: "password", placeholder: "1/1234567890:abcdef...", required: true },
    { key: "team_id", labelKey: "importer.fieldTeamId", type: "text", placeholder: "可选，留空导入全部团队", required: false },
  ],
};

/* ------------------------------------------------------------------ */
/* 列表状态                                                            */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const loadError = ref("");
const sources = ref<SourceInfo[]>([]);
const jobs = ref<ImportJob[]>([]);
const refreshing = ref(false);

/** 是否有任务处于 running 状态（用于自动轮询）。 */
const hasRunning = computed(() => jobs.value.some((j) => j.status === "running"));

async function loadAll() {
  if (!ready.value) return;
  loading.value = true;
  loadError.value = "";
  try {
    const [srcRes, jobsRes] = await Promise.all([
      importerApi.listSources(wsId.value, projectId.value),
      importerApi.listJobs(wsId.value, projectId.value),
    ]);
    sources.value = srcRes ?? [];
    jobs.value = jobsRes ?? [];
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : t("importer.loadFailed");
  } finally {
    loading.value = false;
  }
}

/** 刷新任务列表（不重置 loading，用于轮询）。 */
async function refreshJobs() {
  if (!ready.value) return;
  refreshing.value = true;
  try {
    jobs.value = (await importerApi.listJobs(wsId.value, projectId.value)) ?? [];
  } catch {
    // 轮询失败静默：不打断用户操作
  } finally {
    refreshing.value = false;
  }
}

/** 轮询：存在 running 任务时每 3s 拉取一次，直到全部结束。 */
watch(hasRunning, (running) => {
  if (!running) return;
  const timer = window.setInterval(() => {
    if (!hasRunning.value) {
      window.clearInterval(timer);
      return;
    }
    void refreshJobs();
  }, 3000);
});

/* ------------------------------------------------------------------ */
/* 展示辅助                                                            */
/* ------------------------------------------------------------------ */

function sourceName(source: string): string {
  return sources.value.find((s) => s.id === source)?.name ?? source.toUpperCase();
}

function formatTime(iso?: string | null): string {
  if (!iso) return "-";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "-";
  return d.toLocaleString("zh-CN", { hour12: false });
}

function statusLabel(status: string): string {
  return t(`importer.status.${status}`);
}

function countsOf(job: ImportJob): { imported: number; skipped: number; failed: number } {
  const r = job.result;
  if (!r) return { imported: 0, skipped: 0, failed: 0 };
  return { imported: r.imported, skipped: r.skipped, failed: r.failed };
}

/* ------------------------------------------------------------------ */
/* 新建导入：来源选择                                                  */
/* ------------------------------------------------------------------ */

const showCreate = ref(false);
const selectedSource = ref<ImportSource | "">("");
const creating = ref(false);
const createError = ref("");

const selectedInfo = computed<SourceInfo | undefined>(() =>
  sources.value.find((s) => s.id === selectedSource.value),
);

const credentialDefs = computed<CredentialFieldDef[]>(() =>
  selectedSource.value ? (CREDENTIAL_FIELDS[selectedSource.value] ?? []) : [],
);

/** 平台凭证表单（reactive）。 */
const credentials = reactive<Record<string, string>>({
  base_url: "",
  email: "",
  api_token: "",
  project_key: "",
  team_id: "",
});

/** CSV 流程状态。 */
const csvFile = ref<File | null>(null);
const csvDefaultType = ref("task");
const uploadingCsv = ref(false);
const csvError = ref("");

/** 平台流程状态。 */
const importing = ref(false);
const platformError = ref("");

const fileInput = ref<HTMLInputElement | null>(null);

function openCreate() {
  selectedSource.value = "";
  createError.value = "";
  csvFile.value = null;
  csvDefaultType.value = "task";
  csvError.value = "";
  platformError.value = "";
  Object.keys(credentials).forEach((k) => {
    credentials[k] = "";
  });
  showCreate.value = true;
}

function closeCreate() {
  if (creating.value || uploadingCsv.value || importing.value) return;
  showCreate.value = false;
}

/** 第一步：创建导入任务。 */
async function createJobForSource() {
  if (!selectedSource.value) {
    createError.value = t("importer.sourceRequired");
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    const job = await importerApi.createJob(
      wsId.value,
      projectId.value,
      selectedSource.value,
    );
    await loadAll();
    // 直接推进到第二步
    if (isPlatformSource(job.source)) {
      void importFromPlatform(job);
    } else {
      void uploadCsvFile(job);
    }
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : t("importer.createFailed");
  } finally {
    creating.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* CSV 流程                                                            */
/* ------------------------------------------------------------------ */

function onFilePicked(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0] ?? null;
  csvFile.value = file;
  csvError.value = "";
  if (!file) return;
  if (!/\.csv$/i.test(file.name) && file.type !== "text/csv") {
    csvError.value = t("importer.csvTypeMismatch");
  }
}

function clearFile() {
  csvFile.value = null;
  csvError.value = "";
  if (fileInput.value) fileInput.value.value = "";
}

async function uploadCsvFile(job?: ImportJob) {
  if (!job) {
    // 用户尚未点「创建任务」：先建任务
    if (!selectedSource.value) {
      createError.value = t("importer.sourceRequired");
      return;
    }
    creating.value = true;
    createError.value = "";
    try {
      job = await importerApi.createJob(wsId.value, projectId.value, selectedSource.value);
    } catch (e: unknown) {
      createError.value = e instanceof Error ? e.message : t("importer.createFailed");
      return;
    } finally {
      creating.value = false;
    }
  }

  if (!csvFile.value) {
    csvError.value = t("importer.csvFileRequired");
    return;
  }

  uploadingCsv.value = true;
  csvError.value = "";
  try {
    await importerApi.uploadCsv(
      wsId.value,
      projectId.value,
      job.id,
      csvFile.value,
      csvDefaultType.value || undefined,
    );
    toast.success(t("importer.csvImportSuccess"));
    showCreate.value = false;
    await loadAll();
    // 打开结果详情
    await openJobDetail(job.id);
  } catch (e: unknown) {
    csvError.value = e instanceof Error ? e.message : t("importer.csvImportFailed");
  } finally {
    uploadingCsv.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 平台流程                                                            */
/* ------------------------------------------------------------------ */

/** 平台导入：按当前凭证表单提交。 */
async function importFromPlatform(job?: ImportJob) {
  if (!selectedSource.value) {
    createError.value = t("importer.sourceRequired");
    return;
  }

  if (!job) {
    creating.value = true;
    createError.value = "";
    try {
      job = await importerApi.createJob(wsId.value, projectId.value, selectedSource.value);
    } catch (e: unknown) {
      createError.value = e instanceof Error ? e.message : t("importer.createFailed");
      return;
    } finally {
      creating.value = false;
    }
  }

  const missing = credentialDefs.value
    .filter((f) => f.required && !credentials[f.key]?.trim())
    .map((f) => t(f.labelKey));
  if (missing.length > 0) {
    platformError.value = t("importer.credentialsRequired", { fields: missing.join("、") });
    return;
  }

  importing.value = true;
  platformError.value = "";
  try {
    await importerApi.importPlatform(
      wsId.value,
      projectId.value,
      job.id,
      {
        base_url: credentials.base_url?.trim() || undefined,
        email: credentials.email?.trim() || undefined,
        api_token: credentials.api_token?.trim() || undefined,
        project_key: credentials.project_key?.trim() || undefined,
        team_id: credentials.team_id?.trim() || undefined,
      },
      { default_type: csvDefaultType.value || undefined },
    );
    toast.success(t("importer.platformImportSuccess"));
    showCreate.value = false;
    await loadAll();
    await openJobDetail(job.id);
  } catch (e: unknown) {
    platformError.value = e instanceof Error ? e.message : t("importer.platformImportFailed");
  } finally {
    importing.value = false;
  }
}

/* ------------------------------------------------------------------ */
/* 结果详情                                                            */
/* ------------------------------------------------------------------ */

const showResult = ref(false);
const resultJob = ref<ImportJob | null>(null);
const resultLoading = ref(false);
const resultError = ref("");

async function openJobDetail(jobId: number) {
  showResult.value = true;
  resultJob.value = null;
  resultError.value = "";
  resultLoading.value = true;
  try {
    resultJob.value = await importerApi.getJob(wsId.value, projectId.value, jobId);
  } catch (e: unknown) {
    resultError.value = e instanceof Error ? e.message : t("importer.loadFailed");
  } finally {
    resultLoading.value = false;
  }
}

function closeResult() {
  showResult.value = false;
  resultJob.value = null;
}

const resultCounts = computed(() => {
  const r = resultJob.value?.result;
  if (!r) return null;
  return {
    total: r.total_records,
    imported: r.imported,
    skipped: r.skipped,
    failed: r.failed,
  };
});

const resultErrors = computed<ImportError[]>(() => resultJob.value?.result?.errors ?? []);

/* ------------------------------------------------------------------ */
/* 示例 CSV 模板下载（客户端 Blob 生成）                                */
/* ------------------------------------------------------------------ */

function downloadSample() {
  const header = SAMPLE_CSV_COLUMNS.join(",");
  const sample = [
    "EXT-1001,登录页支持手机号验证码登录,用户在登录页输入手机号并获取验证码完成登录,requirement,high,进行中,zhangsan@company.com,lisi@company.com,登录,认证,2025-06-30,,5",
    "EXT-1002,修复支付回调重复扣款缺陷,支付回调幂等处理缺陷,task,urgent,待处理,wangwu@company.com,zhangsan@company.com,支付,缺陷,,EXT-1001,3",
  ].join("\n");
  const csv = `${header}\n${sample}\n`;
  // 加 BOM，保证 Excel 打开中文不乱码
  const blob = new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "import-sample.csv";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

/* ------------------------------------------------------------------ */
/* 生命周期                                                            */
/* ------------------------------------------------------------------ */

watch(ready, (v) => {
  if (v) void loadAll();
});

onMounted(() => {
  if (ready.value) void loadAll();
});
</script>

<template>
  <AppPageContainer max="1080px">
    <div class="importer">
      <!-- ===== 页头 ===== -->
      <header class="importer__header">
        <div>
          <h1 class="importer__title">{{ t("importer.title") }}</h1>
          <p class="importer__subtitle">{{ t("importer.subtitle") }}</p>
        </div>
        <div class="importer__header-actions">
          <AppButton variant="secondary" @click="downloadSample">
            <DynamicIcon name="Download" :size="14" />
            {{ t("importer.downloadSample") }}
          </AppButton>
          <AppButton :disabled="!ready" @click="openCreate">
            <DynamicIcon name="Plus" :size="14" />
            {{ t("importer.newImport") }}
          </AppButton>
        </div>
      </header>

      <!-- ===== 上下文未就绪 ===== -->
      <AppLoadingState
        v-if="!ready && !ctxError"
        :text="t('importer.loadingContext')"
      />
      <AppErrorState
        v-else-if="ctxError || loadError"
        :message="ctxError || loadError"
        :retry="t('common.retry')"
        @retry="loadAll"
      />

      <template v-else>
        <!-- ===== 加载态 ===== -->
        <AppLoadingState v-if="loading" :text="t('importer.loading')" />

        <!-- ===== 空态 ===== -->
        <AppEmptyState
          v-else-if="!jobs.length"
          icon="📥"
          :title="t('importer.emptyTitle')"
          :description="t('importer.emptyDesc')"
          :cta-text="t('importer.newImport')"
          @cta-click="openCreate"
        />

        <!-- ===== 任务列表 ===== -->
        <div v-else class="importer__list">
          <div class="importer__list-head">
            <span class="importer__list-title">{{ t("importer.jobList") }}</span>
            <AppButton
              v-if="hasRunning"
              size="sm"
              variant="ghost"
              :loading="refreshing"
              @click="refreshJobs"
            >
              {{ t("importer.refreshing") }}
            </AppButton>
          </div>

          <article v-for="job in jobs" :key="job.id" class="importer__job">
            <div class="importer__job-main">
              <div class="importer__job-head">
                <span class="importer__job-source">
                  <DynamicIcon :name="SOURCE_ICONS[job.source] ?? 'Upload'" :size="14" />
                  {{ sourceName(job.source) }}
                </span>
                <AppBadge :variant="STATUS_VARIANT[job.status] ?? 'default'">
                  {{ statusLabel(job.status) }}
                </AppBadge>
              </div>

              <div class="importer__job-meta">
                <span class="importer__meta-item">
                  <DynamicIcon name="Hash" :size="12" />
                  #{{ job.id }}
                </span>
                <span class="importer__meta-item">
                  {{ t("importer.createdAt") }}：{{ formatTime(job.created_at) }}
                </span>
                <span
                  v-if="job.result"
                  class="importer__meta-item importer__meta-item--counts"
                >
                  {{ t("importer.imported") }} <b>{{ countsOf(job).imported }}</b>
                  · {{ t("importer.skipped") }} <b>{{ countsOf(job).skipped }}</b>
                  · {{ t("importer.failed") }} <b>{{ countsOf(job).failed }}</b>
                </span>
                <span
                  v-else-if="job.status === 'running'"
                  class="importer__meta-item"
                >
                  {{ t("importer.runningHint") }}
                </span>
              </div>
            </div>

            <div class="importer__job-actions">
              <AppButton
                size="sm"
                variant="secondary"
                :disabled="job.status === 'running'"
                @click="openJobDetail(job.id)"
              >
                <DynamicIcon name="Eye" :size="12" />
                {{ t("importer.viewResult") }}
              </AppButton>
            </div>
          </article>
        </div>
      </template>

      <!-- ===== 新建导入弹窗：第一步 选择来源 ===== -->
      <AppModal
        :visible="showCreate && !selectedSource"
        :title="t('importer.newImport')"
        width="720px"
        @close="closeCreate"
      >
        <p class="importer__step-hint">{{ t("importer.selectSourceHint") }}</p>

        <div class="importer__sources">
          <button
            v-for="src in sources"
            :key="src.id"
            type="button"
            class="importer__source-card"
            :class="{ 'is-selected': selectedSource === src.id }"
            @click="selectedSource = src.id as ImportSource"
          >
            <span class="importer__source-icon">
              <DynamicIcon :name="SOURCE_ICONS[src.id] ?? 'Upload'" :size="20" />
            </span>
            <span class="importer__source-body">
              <span class="importer__source-name">{{ src.name }}</span>
              <span class="importer__source-desc">{{ src.description }}</span>
              <span class="importer__source-req">
                <AppBadge v-if="src.needs_credentials" variant="warning" dot />
                <AppBadge v-else variant="success" dot />
                {{
                  src.needs_credentials
                    ? t("importer.needsCredentials")
                    : t("importer.noCredentials")
                }}
              </span>
            </span>
          </button>

          <p v-if="!sources.length" class="importer__error">
            {{ t("importer.sourcesEmpty") }}
          </p>
        </div>

        <p v-if="createError" class="importer__error">{{ createError }}</p>

        <template #footer>
          <AppButton variant="ghost" @click="closeCreate">
            {{ t("common.cancel") }}
          </AppButton>
          <AppButton :disabled="!selectedSource" :loading="creating" @click="createJobForSource">
            {{ t("importer.nextStep") }}
          </AppButton>
        </template>
      </AppModal>

      <!-- ===== 新建导入弹窗：第二步 CSV 上传 ===== -->
      <AppModal
        :visible="showCreate && selectedSource === 'csv'"
        :title="t('importer.csvTitle')"
        width="560px"
        @close="closeCreate"
      >
        <div class="importer__form">
          <div class="importer__field">
            <label class="importer__label">
              {{ t("importer.csvFile") }} <span class="importer__required">*</span>
            </label>
            <div
              class="importer__dropzone"
              :class="{ 'is-dragging': false }"
              @click="fileInput?.click()"
            >
              <DynamicIcon name="Upload" :size="22" />
              <p class="importer__dropzone-text">
                {{ t("importer.dragAndDrop") }}
              </p>
              <p class="importer__dropzone-hint">
                {{ t("importer.csvFileHint") }}
              </p>
              <input
                ref="fileInput"
                type="file"
                accept=".csv,text/csv"
                class="importer__file-input"
                @change="onFilePicked"
              />
            </div>

            <div v-if="csvFile" class="importer__file-chip">
              <DynamicIcon name="FileSpreadsheet" :size="14" />
              <span class="importer__file-name">{{ csvFile.name }}</span>
              <span class="importer__file-size">
                {{ (csvFile.size / 1024).toFixed(1) }} KB
              </span>
              <button
                type="button"
                class="importer__file-remove"
                :aria-label="t('common.clear')"
                @click="clearFile"
              >
                <DynamicIcon name="X" :size="12" />
              </button>
            </div>
          </div>

          <div class="importer__field">
            <label class="importer__label">{{ t("importer.defaultType") }}</label>
            <select v-model="csvDefaultType" class="importer__input">
              <option v-for="o in DEFAULT_TYPE_OPTIONS" :key="o.value" :value="o.value">
                {{ t(o.labelKey) }}
              </option>
            </select>
            <p class="importer__hint">{{ t("importer.defaultTypeHint") }}</p>
          </div>

          <div class="importer__field">
            <button type="button" class="importer__link" @click="downloadSample">
              <DynamicIcon name="Download" :size="12" />
              {{ t("importer.downloadSample") }}
            </button>
          </div>

          <p v-if="csvError" class="importer__error">{{ csvError }}</p>
        </div>

        <template #footer>
          <AppButton variant="ghost" :disabled="uploadingCsv" @click="selectedSource = ''">
            {{ t("common.back") }}
          </AppButton>
          <AppButton :loading="uploadingCsv" :disabled="!csvFile" @click="uploadCsvFile()">
            <DynamicIcon name="Upload" :size="14" />
            {{ t("importer.startImport") }}
          </AppButton>
        </template>
      </AppModal>

      <!-- ===== 新建导入弹窗：第二步 平台凭证 ===== -->
      <AppModal
        :visible="showCreate && !!selectedSource && isPlatformSource(selectedSource)"
        :title="t('importer.platformTitle', { source: selectedInfo?.name ?? selectedSource })"
        width="560px"
        @close="closeCreate"
      >
        <div class="importer__form">
          <p class="importer__hint importer__hint--lead">
            {{ t("importer.platformHint", { source: selectedInfo?.name ?? selectedSource }) }}
          </p>

          <div v-for="f in credentialDefs" :key="f.key" class="importer__field">
            <label class="importer__label">
              {{ t(f.labelKey) }}
              <span v-if="f.required" class="importer__required">*</span>
            </label>
            <input
              v-model="credentials[f.key]"
              class="importer__input"
              :type="f.type"
              :placeholder="f.placeholder"
              :autocomplete="f.key === 'api_token' ? 'off' : undefined"
              spellcheck="false"
            />
          </div>

          <p class="importer__hint">
            {{ t("importer.credentialsHint") }}
          </p>

          <p v-if="platformError" class="importer__error">{{ platformError }}</p>
        </div>

        <template #footer>
          <AppButton variant="ghost" :disabled="importing" @click="selectedSource = ''">
            {{ t("common.back") }}
          </AppButton>
          <AppButton :loading="importing" @click="importFromPlatform()">
            <DynamicIcon name="Upload" :size="14" />
            {{ t("importer.startImport") }}
          </AppButton>
        </template>
      </AppModal>

      <!-- ===== 结果详情弹窗 ===== -->
      <AppModal
        :visible="showResult"
        :title="t('importer.resultTitle')"
        width="680px"
        @close="closeResult"
      >
        <AppLoadingState v-if="resultLoading" :text="t('importer.loadingResult')" />
        <AppErrorState
          v-else-if="resultError"
          :message="resultError"
          :retry="t('common.retry')"
          @retry="closeResult"
        />

        <template v-else-if="resultJob">
          <div class="importer__result-head">
            <span class="importer__job-source">
              <DynamicIcon
                :name="SOURCE_ICONS[resultJob.source] ?? 'Upload'"
                :size="14"
              />
              {{ sourceName(resultJob.source) }}
            </span>
            <AppBadge :variant="STATUS_VARIANT[resultJob.status] ?? 'default'">
              {{ statusLabel(resultJob.status) }}
            </AppBadge>
          </div>

          <div v-if="resultCounts" class="importer__stats">
            <div class="importer__stat importer__stat--total">
              <span class="importer__stat-value">{{ resultCounts.total }}</span>
              <span class="importer__stat-label">{{ t("importer.total") }}</span>
            </div>
            <div class="importer__stat importer__stat--success">
              <span class="importer__stat-value">{{ resultCounts.imported }}</span>
              <span class="importer__stat-label">{{ t("importer.imported") }}</span>
            </div>
            <div class="importer__stat importer__stat--warning">
              <span class="importer__stat-value">{{ resultCounts.skipped }}</span>
              <span class="importer__stat-label">{{ t("importer.skipped") }}</span>
            </div>
            <div class="importer__stat importer__stat--danger">
              <span class="importer__stat-value">{{ resultCounts.failed }}</span>
              <span class="importer__stat-label">{{ t("importer.failed") }}</span>
            </div>
          </div>

          <p v-else class="importer__hint">
            {{ t("importer.noResultYet") }}
          </p>

          <div v-if="resultErrors.length" class="importer__errors">
            <p class="importer__errors-title">
              {{ t("importer.errorList") }}（{{ resultErrors.length }}）
            </p>
            <ul class="importer__error-list">
              <li v-for="(err, i) in resultErrors" :key="i" class="importer__error-item">
                <div class="importer__error-meta">
                  <code v-if="err.external_id" class="importer__error-id">
                    {{ err.external_id }}
                  </code>
                  <span v-if="err.title" class="importer__error-title">
                    {{ err.title }}
                  </span>
                </div>
                <p class="importer__error-reason">{{ err.reason }}</p>
              </li>
            </ul>
          </div>
          <p v-else-if="resultCounts" class="importer__hint">
            {{ t("importer.noErrors") }}
          </p>
        </template>

        <template #footer>
          <AppButton variant="ghost" @click="closeResult">
            {{ t("common.close") }}
          </AppButton>
        </template>
      </AppModal>
    </div>
  </AppPageContainer>
</template>

<style scoped>
.importer {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ---- 页头 ---- */
.importer__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.importer__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.importer__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
}

.importer__header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

/* ---- 任务列表 ---- */
.importer__list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.importer__list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.importer__list-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.importer__job {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--surface-1);
}

.importer__job-main {
  flex: 1;
  min-width: 0;
}

.importer__job-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.importer__job-source {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.importer__job-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px 16px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.importer__meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.importer__meta-item--counts b {
  color: var(--text-secondary);
  font-weight: 600;
}

.importer__job-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* ---- 来源卡片 ---- */
.importer__step-hint {
  margin: 0 0 14px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.importer__sources {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.importer__source-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--border-default);
  border-radius: 10px;
  background: var(--surface-1);
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  transition: border-color 0.15s, background 0.15s;
}

.importer__source-card:hover {
  border-color: var(--brand-500);
}

.importer__source-card.is-selected {
  border-color: var(--brand-500);
  background: var(--brand-50, rgba(59, 130, 246, 0.06));
}

.importer__source-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 8px;
  background: var(--surface-2);
  color: var(--brand-600);
}

.importer__source-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.importer__source-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.importer__source-desc {
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-tertiary);
}

.importer__source-req {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ---- 表单 ---- */
.importer__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.importer__field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.importer__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.importer__required {
  color: var(--danger-500);
}

.importer__input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
}

.importer__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.importer__hint {
  margin: 0;
  font-size: 11px;
  line-height: 1.5;
  color: var(--text-tertiary);
}

.importer__hint--lead {
  font-size: 12px;
}

.importer__link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  border: none;
  background: none;
  color: var(--brand-600);
  font-family: inherit;
  font-size: 12px;
  cursor: pointer;
}

.importer__link:hover {
  text-decoration: underline;
}

.importer__error {
  margin: 0;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--danger-50, rgba(239, 68, 68, 0.06));
  color: var(--danger-500);
  font-size: 12px;
}

/* ---- 文件选择 ---- */
.importer__dropzone {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 24px 16px;
  border: 1px dashed var(--border-default);
  border-radius: 10px;
  background: var(--surface-2);
  color: var(--text-tertiary);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.importer__dropzone:hover,
.importer__dropzone.is-dragging {
  border-color: var(--brand-500);
  background: var(--brand-50, rgba(59, 130, 246, 0.06));
}

.importer__dropzone-text {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.importer__dropzone-hint {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
}

.importer__file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.importer__file-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--surface-1);
  font-size: 12px;
  color: var(--text-secondary);
}

.importer__file-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  color: var(--text-primary);
}

.importer__file-size {
  color: var(--text-tertiary);
}

.importer__file-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: var(--surface-3);
  color: var(--text-tertiary);
  cursor: pointer;
}

.importer__file-remove:hover {
  color: var(--danger-500);
}

/* ---- 结果统计 ---- */
.importer__result-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.importer__stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 14px;
}

.importer__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 12px 8px;
  border-radius: 8px;
  background: var(--surface-2);
}

.importer__stat-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.importer__stat-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.importer__stat--success .importer__stat-value { color: var(--success-500); }
.importer__stat--warning .importer__stat-value { color: var(--warning-500); }
.importer__stat--danger .importer__stat-value { color: var(--danger-500); }

/* ---- 错误明细 ---- */
.importer__errors-title {
  margin: 0 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.importer__error-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: 320px;
  overflow-y: auto;
}

.importer__error-item {
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  background: var(--surface-2);
}

.importer__error-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.importer__error-id {
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--surface-3);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
  color: var(--text-secondary);
}

.importer__error-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-word;
}

.importer__error-reason {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--danger-500);
  word-break: break-word;
}

/* ---- 窄屏 ---- */
@media (max-width: 720px) {
  .importer__header {
    flex-direction: column;
  }

  .importer__header-actions {
    width: 100%;
  }

  .importer__sources {
    grid-template-columns: 1fr;
  }

  .importer__job {
    flex-direction: column;
    align-items: stretch;
  }

  .importer__job-actions {
    justify-content: flex-end;
  }

  .importer__stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
