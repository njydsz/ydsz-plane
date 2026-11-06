<script setup lang="ts">
/**
 * DlpScanView — 敏感数据扫描面板。
 *
 * 功能：
 *  - 创建扫描任务（项目/文档/工作项/全局）
 *  - 查看扫描进度与报告（按类别/严重级别统计）
 *  - 发现列表（脱敏值 + 置信度 + 严重级别）
 *  - 标记已处理
 *  - 写入前检查（粘贴/输入时实时检测）
 *  - 规则列表
 */
import { computed, onMounted, ref } from "vue";

import {
  dlpApi,
  type DlpFinding,
  type DlpJob,
  type DlpReport,
  type DlpRule,
  type DlpScope,
  type DlpCheckResult,
} from "@/api/services/dlp";
import { useWorkspaceStore } from "@/stores/workspace";
import { useTranslation } from "@/composables/useTranslation";
import {
  AppLoadingState,
  AppErrorState,
  AppEmptyState,
  AppBadge,
  AppModal,
} from "@/components";
import { toast } from "@/lib/toast";

const { t } = useTranslation();
const wsStore = useWorkspaceStore();
const wsId = computed(() => wsStore.current?.id ?? 0);

/* ---- 数据 ---- */

const jobs = ref<DlpJob[]>([]);
const findings = ref<DlpFinding[]>([]);
const report = ref<DlpReport | null>(null);
const rules = ref<DlpRule[]>([]);
const loading = ref(true);
const error = ref("");

/* ---- 筛选 ---- */
const severityFilter = ref<string>("");
const categoryFilter = ref<string>("");
const actionFilter = ref<string>("");

/* ---- 新建任务 ---- */
const showCreate = ref(false);
const creating = ref(false);
const newScope = ref<DlpScope>("project");
const newTargetId = ref<number>(0);

/* ---- 实时检查 ---- */
const checkText = ref("");
const checkResult = ref<DlpCheckResult | null>(null);
const checking = ref(false);

/* ---- 枚举 ---- */

const severityVariant: Record<string, "danger" | "warning" | "info" | "default"> = {
  critical: "danger",
  high: "warning",
  medium: "info",
  low: "default",
};

const categoryLabels: Record<string, string> = {
  id_card: "身份证",
  bank_card: "银行卡",
  phone: "手机号",
  email: "邮箱",
  password: "密码",
  api_key: "API 密钥",
  address: "地址",
  passport: "护照",
  license: "营业执照",
  custom: "自定义",
};

const scopeLabels: Record<string, string> = {
  project: "项目",
  page: "文档",
  issue: "工作项",
  global: "全局",
};

const statusLabels: Record<string, string> = {
  pending: "待执行",
  running: "执行中",
  completed: "已完成",
  failed: "失败",
};

/* ---- 加载 ---- */

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const [j, f, r] = await Promise.all([
      dlpApi.listJobs(wsId.value),
      dlpApi.listFindings(wsId.value, { limit: 200 }),
      dlpApi.listRules(wsId.value),
    ]);
    jobs.value = (j as { items?: DlpJob[] }).items ?? [];
    findings.value = (f as { items?: DlpFinding[] }).items ?? [];
    rules.value = (r as { items?: DlpRule[] }).items ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false
  }
}

onMounted(load);

/* ---- 任务操作 ---- */

async function createAndRun() {
  if (!wsId.value) return;
  creating.value = true;
  try {
    const job = await dlpApi.createJob(wsId.value, newScope.value, newTargetId.value || undefined);
    toast.success("扫描任务已创建");
    showCreate.value = false;
    await dlpApi.runJob(wsId.value, job.id);
    toast.success("扫描已启动");
    await load()
    pollJob(job.id)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建失败")
  } finally {
    creating.value = false
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

function pollJob(jobId: number) {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(async () => {
    try {
      const job = await dlpApi.getJob(wsId.value, jobId)
      const idx = jobs.value.findIndex((j) => j.id === jobId)
      if (idx >= 0) jobs.value[idx] = job
      if (job.status === "completed" || job.status === "failed") {
        if (pollTimer) clearInterval(pollTimer)
        pollTimer = null
        await loadReport(jobId)
        await load()
        if (job.status === "completed") toast.success("扫描完成")
        else toast.error("扫描失败：" + (job.error_message || "未知"))
      }
    } catch { /* */ }
  }, 1500)
}

async function loadReport(jobId: number) {
  try {
    report.value = await dlpApi.getReport(wsId.value, jobId)
  } catch { /* */ }
}

/* ---- 发现操作 ---- */

async function resolve(id: number) {
  try {
    await dlpApi.resolveFinding(wsId.value, id)
    toast.success("已标记为已处理")
    await load()
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "操作失败")
  }
}

/* ---- 实时检查 ---- */

async function runCheck() {
  if (!checkText.value.trim()) return
  checking.value = true
  try {
    checkResult.value = await dlpApi.check(wsId.value, checkText.value)
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "检查失败")
  } finally {
    checking.value = false
  }
}

/* ---- 过滤 ---- */

const filteredFindings = computed(() =>
  findings.value.filter((f) => {
    if (severityFilter.value && f.severity !== severityFilter.value) return false
    if (categoryFilter.value && f.category !== categoryFilter.value) return false
    if (actionFilter.value && f.action !== actionFilter.value) return false
    return true
  }),
)
</script>

<template>
  <div class="dlp">
    <!-- 头部 -->
    <div class="dlp__header">
      <div>
        <h1 class="dlp__title">{{ t("dlp.title", "敏感数据扫描") }}</h1>
        <p class="dlp__subtitle">
          {{ t("dlp.subtitle", "识别身份证、银行卡、手机号等敏感信息，防止数据泄漏") }}
        </p>
      </div>
      <button class="dlp__primary" @click="showCreate = true">
        ＋ {{ t("dlp.newScan", "新建扫描") }}
      </button>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 实时检查 -->
      <section class="dlp__card">
        <h2 class="dlp__card-title">{{ t("dlp.check", "写入前检查") }}</h2>
        <textarea
          v-model="checkText"
          class="dlp__textarea"
          rows="3"
          :placeholder="t('dlp.checkPlaceholder', '粘贴文本以检测敏感信息...')"
        />
        <div class="dlp__check-actions">
          <button class="dlp__btn" :disabled="checking || !checkText.trim()" @click="runCheck">
            {{ checking ? t("dlp.checking", "检查中...") : t("dlp.runCheck", "检查") }}
          </button>
        </div>
        <div v-if="checkResult" class="dlp__check-result">
          <AppBadge
            :variant="checkResult.action === 'block' ? 'danger' : checkResult.action === 'warn' ? 'warning' : 'success'"
          >
            {{ checkResult.action === "block" ? t("dlp.action.block", "阻断")
               : checkResult.action === "warn" ? t("dlp.action.warn", "告警")
               : t("dlp.action.allow", "允许") }}
            ({{ checkResult.count }})
          </AppBadge>
          <ul class="dlp__match-list">
            <li v-for="(m, i) in checkResult.matches" :key="i" class="dlp__match">
              <AppBadge :variant="severityVariant[m.severity]">{{ m.severity }}</AppBadge>
              <span class="dlp__match-cat">{{ categoryLabels[m.category] ?? m.category }}</span>
              <code class="dlp__match-val">{{ m.value }}</code>
              <span class="dlp__match-conf">{{ Math.round(m.confidence * 100) }}%</span>
            </li>
          </ul>
        </div>
      </section>

      <!-- 扫描任务 -->
      <section class="dlp__card">
        <h2 class="dlp__card-title">{{ t("dlp.jobs", "扫描任务") }}</h2>
        <div v-if="!jobs.length" class="dlp__empty">{{ t("dlp.noJobs", "暂无扫描任务") }}</div>
        <div v-else class="dlp__job-list">
          <div v-for="j in jobs" :key="j.id" class="dlp__job">
            <div class="dlp__job-main">
              <span class="dlp__job-scope">{{ scopeLabels[j.scope] ?? j.scope }}</span>
              <AppBadge
                :variant="j.status === 'completed' ? 'success' : j.status === 'running' ? 'info' : j.status === 'failed' ? 'danger' : 'default'"
              >
                {{ statusLabels[j.status] ?? j.status }}
              </AppBadge>
              <span class="dlp__job-counts">
                {{ j.scanned_items }}/{{ j.total_items }} · {{ j.findings }} {{ t("dlp.findingsUnit", "发现") }}
              </span>
            </div>
            <button class="dlp__link" @click="loadReport(j.id)">
              {{ t("dlp.report", "报告") }}
            </button>
          </div>
        </div>
      </section>

      <!-- 报告 -->
      <section v-if="report" class="dlp__card">
        <h2 class="dlp__card-title">{{ t("dlp.report", "扫描报告") }}</h2>
        <div class="dlp__stats">
          <div class="dlp__stat">
            <span class="dlp__stat-value">{{ report.total_items }}</span>
            <span class="dlp__stat-label">{{ t("dlp.totalItems", "扫描项") }}</span>
          </div>
          <div class="dlp__stat">
            <span class="dlp__stat-value dlp__stat-value--danger">{{ report.findings }}</span>
            <span class="dlp__stat-label">{{ t("dlp.findings", "发现") }}</span>
          </div>
        </div>
        <div class="dlp__breakdown">
          <div v-if="Object.keys(report.by_severity).length">
            <h3 class="dlp__sub-title">{{ t("dlp.bySeverity", "按严重级别") }}</h3>
            <div class="dlp__bars">
              <div v-for="(n, sev) in report.by_severity" :key="sev" class="dlp__bar-row">
                <span class="dlp__bar-label">{{ sev }}</span>
                <div class="dlp__bar-track">
                  <div
                    class="dlp__bar-fill"
                    :class="`dlp__bar-fill--${sev}`"
                    :style="{ width: Math.min(100, (n / Math.max(1, report.findings)) * 100) + '%' }"
                  />
                </div>
                <span class="dlp__bar-value">{{ n }}</span>
              </div>
            </div>
          </div>
          <div v-if="Object.keys(report.by_category).length">
            <h3 class="dlp__sub-title">{{ t("dlp.byCategory", "按类别") }}</h3>
            <div class="dlp__bars">
              <div v-for="(n, cat) in report.by_category" :key="cat" class="dlp__bar-row">
                <span class="dlp__bar-label">{{ categoryLabels[cat] ?? cat }}</span>
                <div class="dlp__bar-track">
                  <div
                    class="dlp__bar-fill dlp__bar-fill--category"
                    :style="{ width: Math.min(100, (n / Math.max(1, report.findings)) * 100) + '%' }"
                  />
                </div>
                <span class="dlp__bar-value">{{ n }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 发现列表 -->
      <section class="dlp__card">
        <div class="dlp__card-head">
          <h2 class="dlp__card-title">{{ t("dlp.findings", "发现") }}</h2>
          <div class="dlp__filters">
            <select v-model="severityFilter" class="dlp__select">
              <option value="">{{ t("dlp.allSeverities", "全部级别") }}</option>
              <option value="critical">critical</option>
              <option value="high">high</option>
              <option value="medium">medium</option>
              <option value="low">low</option>
            </select>
            <select v-model="categoryFilter" class="dlp__select">
              <option value="">{{ t("dlp.allCategories", "全部类别") }}</option>
              <option v-for="(label, cat) in categoryLabels" :key="cat" :value="cat">
                {{ label }}
              </option>
            </select>
            <select v-model="actionFilter" class="dlp__select">
              <option value="">{{ t("dlp.allActions", "全部状态") }}</option>
              <option value="flagged">flagged</option>
              <option value="blocked">blocked</option>
              <option value="resolved">resolved</option>
            </select>
          </div>
        </div>

        <AppEmptyState
          v-if="!filteredFindings.length"
          :title="t('dlp.noFindings', '暂无发现')"
          :description="t('dlp.noFindingsDesc', '未检测到敏感数据，或创建扫描任务后查看结果')"
        />

        <div v-else class="dlp__table">
          <div class="dlp__table-head">
            <span>{{ t("dlp.colValue", "值（脱敏）") }}</span>
            <span>{{ t("dlp.colCategory", "类别") }}</span>
            <span>{{ t("dlp.colSeverity", "级别") }}</span>
            <span>{{ t("dlp.colTarget", "位置") }}</span>
            <span>{{ t("dlp.colConfidence", "置信度") }}</span>
            <span>{{ t("dlp.colAction", "操作") }}</span>
          </div>
          <div v-for="f in filteredFindings" :key="f.id" class="dlp__row">
            <code class="dlp__value">{{ f.masked_value }}</code>
            <span>{{ categoryLabels[f.category] ?? f.category }}</span>
            <span><AppBadge :variant="severityVariant[f.severity]">{{ f.severity }}</AppBadge></span>
            <span class="dlp__muted">{{ f.target_title || `#${f.target_id}` }}</span>
            <span>{{ Math.round(f.confidence * 100) }}%</span>
            <span>
              <button
                v-if="f.action === 'flagged'"
                class="dlp__link"
                @click="resolve(f.id)"
              >
                {{ t("dlp.resolve", "标记已处理") }}
              </button>
              <span v-else class="dlp__muted">{{ f.action }}</span>
            </span>
          </div>
        </div>
      </section>

      <!-- 规则 -->
      <section class="dlp__card">
        <h2 class="dlp__card-title">{{ t("dlp.rules", "检测规则") }}</h2>
        <div class="dlp__rule-list">
          <div v-for="r in rules" :key="r.name" class="dlp__rule">
            <span class="dlp__rule-name">{{ r.name }}</span>
            <span class="dlp__muted">{{ categoryLabels[r.category] ?? r.category }}</span>
            <AppBadge :variant="severityVariant[r.severity]">{{ r.severity }}</AppBadge>
            <span class="dlp__muted">{{ Math.round(r.confidence * 100) }}%</span>
          </div>
        </div>
      </section>
    </template>

    <!-- 新建任务 -->
    <AppModal
      :visible="showCreate"
      :title="t('dlp.newScan', '新建扫描')"
      @close="showCreate = false"
    >
      <div class="dlp__form">
        <label class="dlp__label">{{ t("dlp.scope", "扫描范围") }}</label>
        <select v-model="newScope" class="dlp__select">
          <option value="project">{{ t("dlp.scopeProject", "项目") }}</option>
          <option value="page">{{ t("dlp.scopePage", "文档") }}</option>
          <option value="issue">{{ t("dlp.scopeIssue", "工作项") }}</option>
          <option value="global">{{ t("dlp.scopeGlobal", "全局") }}</option>
        </select>
        <label v-if="newScope !== 'global'" class="dlp__label">
          {{ t("dlp.targetId", "目标 ID") }}
        </label>
        <input
          v-if="newScope !== 'global'"
          v-model.number="newTargetId"
          type="number"
          class="dlp__input"
          placeholder="ID"
        />
      </div>
      <template #footer>
        <button class="dlp__ghost" @click="showCreate = false">{{ t("common.cancel", "取消") }}</button>
        <button class="dlp__primary" :disabled="creating" @click="createAndRun">
          {{ creating ? t("dlp.creating", "创建中...") : t("dlp.createAndRun", "创建并扫描") }}
        </button>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
.dlp {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.dlp__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.dlp__title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--txt-primary, #1e293b);
}

.dlp__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary, #64748b);
}

.dlp__primary {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  color: #fff;
  background: var(--accent, #3b82f6);
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.dlp__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dlp__ghost {
  padding: 8px 14px;
  font-size: 13px;
  color: var(--txt-secondary, #64748b);
  background: none;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
  cursor: pointer;
}

.dlp__card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 12px;
  background: var(--bg-surface, #fff);
}

.dlp__card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.dlp__card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary, #1e293b);
}

.dlp__sub-title {
  margin: 8px 0 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--txt-secondary, #64748b);
}

.dlp__textarea,
.dlp__input,
.dlp__select {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
  background: var(--bg-surface, #fff);
  color: var(--txt-primary, #1e293b);
}

.dlp__textarea {
  resize: vertical;
  font-family: ui-monospace, monospace;
}

.dlp__check-actions {
  display: flex;
  justify-content: flex-end;
}

.dlp__btn {
  padding: 6px 14px;
  font-size: 13px;
  color: #fff;
  background: var(--accent, #3b82f6);
  border: none;
  border-radius: 6px;
  cursor: pointer;
}

.dlp__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dlp__check-result {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  background: var(--bg-surface-2, #f8fafc);
}

.dlp__match-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dlp__match {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.dlp__match-cat {
  color: var(--txt-secondary, #64748b);
}

.dlp__match-val {
  font-family: ui-monospace, monospace;
  color: var(--txt-primary, #1e293b);
}

.dlp__match-conf {
  margin-left: auto;
  color: var(--txt-tertiary, #94a3b8);
}

.dlp__job-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dlp__job {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
}

.dlp__job-main {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.dlp__job-scope {
  font-weight: 500;
  color: var(--txt-primary, #1e293b);
}

.dlp__job-counts {
  color: var(--txt-tertiary, #94a3b8);
}

.dlp__link {
  font-size: 12px;
  color: var(--accent, #3b82f6);
  background: none;
  border: none;
  cursor: pointer;
}

.dlp__stats {
  display: flex;
  gap: 12px;
}

.dlp__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 12px 20px;
  border-radius: 8px;
  background: var(--bg-surface-2, #f8fafc);
}

.dlp__stat-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary, #1e293b);
}

.dlp__stat-value--danger {
  color: var(--danger, #ef4444);
}

.dlp__stat-label {
  font-size: 11px;
  color: var(--txt-tertiary, #94a3b8);
}

.dlp__bars {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dlp__bar-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.dlp__bar-label {
  width: 90px;
  flex-shrink: 0;
  color: var(--txt-secondary, #64748b);
}

.dlp__bar-track {
  flex: 1;
  height: 8px;
  border-radius: 4px;
  background: var(--bg-surface-2, #f8fafc);
  overflow: hidden;
}

.dlp__bar-fill {
  height: 100%;
  border-radius: 4px;
  background: var(--accent, #3b82f6);
}

.dlp__bar-fill--critical {
  background: var(--danger, #ef4444);
}

.dlp__bar-fill--high {
  background: var(--warning, #f59e0b);
}

.dlp__bar-fill--medium {
  background: var(--info, #3b82f6);
}

.dlp__bar-fill--low {
  background: var(--txt-tertiary, #94a3b8);
}

.dlp__bar-value {
  width: 36px;
  text-align: right;
  color: var(--txt-primary, #1e293b);
}

.dlp__filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.dlp__table {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
  overflow: hidden;
}

.dlp__table-head,
.dlp__row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 0.8fr 1.5fr 0.7fr 1fr;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  font-size: 12px;
}

.dlp__table-head {
  font-weight: 600;
  color: var(--txt-tertiary, #94a3b8);
  background: var(--bg-surface-2, #f8fafc);
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}

.dlp__row {
  border-bottom: 1px solid var(--border-subtle, #e2e8f0);
}

.dlp__row:last-child {
  border-bottom: none;
}

.dlp__value {
  font-family: ui-monospace, monospace;
  color: var(--txt-primary, #1e293b);
}

.dlp__muted {
  color: var(--txt-tertiary, #94a3b8);
}

.dlp__rule-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dlp__rule {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 6px;
  font-size: 12px;
}

.dlp__rule:hover {
  background: var(--bg-surface-2, #f8fafc);
}

.dlp__rule-name {
  font-family: ui-monospace, monospace;
  color: var(--txt-primary, #1e293b);
}

.dlp__empty {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--txt-tertiary, #94a3b8);
}

.dlp__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dlp__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary, #64748b);
}
</style>
