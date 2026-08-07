<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { versionApi, type DeliveryReport, type Version, type SprintRef } from "@/api/services/version";
import { AppBadge, AppButton, ProgressBar } from "@/components";

const route = useRoute();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const versionId = computed(() => Number(route.params.versionId));

const version = ref<Version | null>(null);
const report = ref<DeliveryReport | null>(null);
const sprints = ref<SprintRef[]>([]);
const loading = ref(true);
const error = ref("");
const printMode = ref(false);

let wsIdVal = 0;

/* ---------- computed ---------- */

const passRatePercent = computed(() => Math.round((report.value?.pass_rate ?? 0) * 100));
const completionPercent = computed(() => Math.round((version.value?.progress?.completion_rate ?? 0) * 100));
const fixRatePercent = computed(() => Math.round((version.value?.quality?.fix_rate ?? 0) * 100));

const qualityGatePassed = computed(() => (version.value?.quality?.critical_bugs ?? 0) === 0);

const statusLabel: Record<string, string> = {
  planning: "规划中",
  active: "进行中",
  released: "已发布",
  archived: "已归档",
};

const sprintStatusLabel: Record<string, string> = {
  planned: "未开始",
  active: "进行中",
  completed: "已完成",
};

/* ---------- data ---------- */

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const [v, r, s] = await Promise.all([
      versionApi.getVersion(wsId, projectId.value, versionId.value),
      versionApi.getDeliveryReport(wsId, projectId.value, versionId.value),
      versionApi.listVersionSprints(wsId, projectId.value, versionId.value),
    ]);
    version.value = v;
    report.value = r;
    sprints.value = s.results ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载交付报告失败";
  } finally {
    loading.value = false;
  }
}

function handlePrint() {
  printMode.value = true;
  setTimeout(() => {
    window.print();
    printMode.value = false;
  }, 100);
}

onMounted(load);
</script>

<template>
  <div class="delivery-report" :class="{ 'delivery-report--print': printMode }">
    <!-- Loading -->
    <div v-if="loading" class="delivery-report__loading">
      <div class="skeleton-line" style="width:60%"></div>
      <div class="skeleton-line" style="width:40%"></div>
      <div class="skeleton-line" style="width:80%"></div>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="delivery-report__error">
      <p>{{ error }}</p>
      <AppButton variant="secondary" size="sm" @click="load">重试</AppButton>
    </div>

    <template v-else-if="version && report">
      <!-- Header -->
      <header class="report-header" :class="{ 'no-print': !printMode }">
        <div class="report-header__left">
          <button class="report-header__back" @click="$router.back()">← 返回</button>
          <div>
            <h1 class="report-header__title">交付报告</h1>
            <p class="report-header__subtitle">{{ version.name }} · {{ version.semver }}</p>
          </div>
        </div>
        <div class="report-header__actions no-print">
          <AppButton variant="secondary" size="sm" @click="handlePrint">
            打印 / 导出
          </AppButton>
        </div>
      </header>

      <!-- Overall Status Banner -->
      <div class="status-banner" :class="report.eligible_to_release ? 'status-banner--pass' : 'status-banner--fail'">
        <span class="status-banner__icon">{{ report.eligible_to_release ? '✅' : '⚠️' }}</span>
        <span class="status-banner__text">
          {{ report.eligible_to_release
            ? '该版本满足发布准出条件'
            : '该版本尚未满足准出条件（通过率 < 80% 或存在严重致命未关闭缺陷）'
          }}
        </span>
      </div>

      <!-- Key Metrics Grid -->
      <section class="metrics-grid">
        <div class="metric-card">
          <div class="metric-card__icon">📊</div>
          <div class="metric-card__value">{{ completionPercent }}%</div>
          <div class="metric-card__label">完成进度</div>
          <ProgressBar
            :percent="completionPercent"
            size="sm"
            :color="completionPercent >= 80 ? 'var(--success-500)' : 'var(--warning-500)'"
            :showLabel="false"
          />
        </div>

        <div class="metric-card">
          <div class="metric-card__icon">🐛</div>
          <div class="metric-card__value">{{ report.fixed_bug_count }} / {{ report.bug_count }}</div>
          <div class="metric-card__label">缺陷修复率</div>
          <ProgressBar
            :percent="fixRatePercent"
            size="sm"
            :color="fixRatePercent >= 80 ? 'var(--success-500)' : 'var(--warning-500)'"
            :showLabel="false"
          />
        </div>

        <div class="metric-card">
          <div class="metric-card__icon">🎯</div>
          <div class="metric-card__value">{{ passRatePercent }}%</div>
          <div class="metric-card__label">综合通过率</div>
          <ProgressBar
            :percent="passRatePercent"
            size="sm"
            :color="passRatePercent >= 80 ? 'var(--success-500)' : 'var(--danger-500)'"
            :showLabel="false"
          />
        </div>

        <div class="metric-card" :class="qualityGatePassed ? 'metric-card--pass' : 'metric-card--fail'">
          <div class="metric-card__icon">{{ qualityGatePassed ? '✅' : '❌' }}</div>
          <div class="metric-card__value">{{ qualityGatePassed ? '通过' : '未通过' }}</div>
          <div class="metric-card__label">质量门禁</div>
          <div class="metric-card__detail">
            致命/严重未关闭: {{ version.quality?.critical_bugs ?? 0 }}
          </div>
        </div>
      </section>

      <!-- Detail Section: Overview -->
      <section class="report-section">
        <h2 class="report-section__title">概览数据</h2>
        <div class="stat-grid">
          <div class="stat-item">
            <span class="stat-item__label">关联迭代数</span>
            <span class="stat-item__value">{{ report.sprint_count }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">总工作项</span>
            <span class="stat-item__value">{{ report.total_issues }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">已完成工作项</span>
            <span class="stat-item__value">{{ report.completed_issues }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">总故事点</span>
            <span class="stat-item__value">{{ Math.round(report.total_points) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">已完成故事点</span>
            <span class="stat-item__value">{{ Math.round(report.completed_points) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">缺陷总数</span>
            <span class="stat-item__value">{{ report.bug_count }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">已修复缺陷</span>
            <span class="stat-item__value">{{ report.fixed_bug_count }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">综合通过率</span>
            <span class="stat-item__value">{{ passRatePercent }}%</span>
          </div>
          <div class="stat-item">
            <span class="stat-item__label">生成时间</span>
            <span class="stat-item__value stat-item__value--date">
              {{ new Date(report.generated_at).toLocaleString("zh-CN") }}
            </span>
          </div>
        </div>
      </section>

      <!-- Detail Section: Sprint Breakdown -->
      <section v-if="sprints.length > 0" class="report-section">
        <h2 class="report-section__title">迭代明细</h2>
        <div class="sprint-table-wrap">
          <table class="sprint-table">
            <thead>
              <tr>
                <th>迭代名称</th>
                <th>状态</th>
                <th>时间</th>
                <th>承诺点数</th>
                <th>完成点数</th>
                <th>达成率</th>
                <th>工作项</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in sprints" :key="s.sprint_id">
                <td class="sprint-table__name">{{ s.name }}</td>
                <td>
                  <AppBadge :variant="s.status === 'completed' ? 'success' : s.status === 'active' ? 'warning' : 'default'">
                    {{ sprintStatusLabel[s.status] ?? s.status }}
                  </AppBadge>
                </td>
                <td class="sprint-table__date">
                  {{ s.start_date ?? '?' }} → {{ s.end_date ?? '?' }}
                </td>
                <td class="sprint-table__num">{{ s.progress?.total_points ?? 0 }}</td>
                <td class="sprint-table__num">{{ s.progress?.done_points ?? 0 }}</td>
                <td class="sprint-table__num">
                  <span :class="(s.progress && s.progress.total_points > 0 && (s.progress.done_points / s.progress.total_points) >= 0.8) ? 'text-success' : ''">
                    {{ s.progress && s.progress.total_points > 0
                      ? Math.round((s.progress.done_points / s.progress.total_points) * 100)
                      : 0 }}%
                  </span>
                </td>
                <td class="sprint-table__num">
                  {{ s.progress?.done_issues ?? 0 }} / {{ s.progress?.total_issues ?? 0 }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Release Notes Preview -->
      <section v-if="version.release_notes" class="report-section">
        <h2 class="report-section__title">Release Notes</h2>
        <div class="notes-block">
          <pre class="notes-block__md">{{ version.release_notes }}</pre>
        </div>
      </section>

      <!-- Footer -->
      <footer class="report-footer no-print">
        <p class="report-footer__text">
          交付报告由系统在 {{ new Date(report.generated_at).toLocaleString("zh-CN") }} 自动生成
        </p>
      </footer>
    </template>
  </div>
</template>

<style scoped>
.delivery-report {
  max-width: 900px;
}

/* ---- loading / error ---- */
.delivery-report__loading {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 48px 0;
}
.skeleton-line {
  height: 14px;
  background: var(--surface-2);
  border-radius: 4px;
  animation: pulse 1.5s infinite;
}
@keyframes pulse {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 0.8; }
}

.delivery-report__error {
  text-align: center;
  padding: 48px 0;
  color: var(--danger-500);
}
.delivery-report__error p { margin: 0 0 12px; }

/* ---- header ---- */
.report-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}
.report-header__left { display: flex; align-items: flex-start; gap: 12px; }
.report-header__back {
  font-size: 13px;
  color: var(--text-tertiary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 0;
  white-space: nowrap;
  margin-top: 1px;
}
.report-header__back:hover { color: var(--brand-500); }
.report-header__title { margin: 0; font-size: 20px; font-weight: 600; }
.report-header__subtitle {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
.report-header__actions { display: flex; gap: 8px; }

/* ---- banner ---- */
.status-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-radius: var(--radius-md);
  margin-bottom: 20px;
  font-size: 14px;
  font-weight: 500;
}
.status-banner--pass {
  background: rgba(15, 194, 123, 0.08);
  color: var(--success-500);
  border: 1px solid rgba(15, 194, 123, 0.2);
}
.status-banner--fail {
  background: rgba(245, 158, 11, 0.08);
  color: var(--warning-500);
  border: 1px solid rgba(245, 158, 11, 0.2);
}
.status-banner__icon { font-size: 18px; }

/* ---- metrics grid ---- */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

.metric-card {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.metric-card--pass { border-color: rgba(15, 194, 123, 0.3); }
.metric-card--fail { border-color: rgba(220, 47, 47, 0.3); }
.metric-card__icon { font-size: 20px; }
.metric-card__value {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.metric-card__label {
  font-size: 12px;
  color: var(--text-tertiary);
}
.metric-card__detail {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* ---- section ---- */
.report-section {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 20px;
  margin-bottom: 16px;
}
.report-section__title {
  margin: 0 0 14px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

/* ---- stat grid ---- */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 12px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
}
.stat-item__label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.stat-item__value {
  font-size: 15px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.stat-item__value--date {
  font-size: 11px;
  font-weight: 400;
  color: var(--text-tertiary);
}

/* ---- sprint table ---- */
.sprint-table-wrap {
  overflow-x: auto;
}
.sprint-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}
.sprint-table th {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-align: left;
  padding: 8px 10px;
  border-bottom: 2px solid var(--border-subtle);
  white-space: nowrap;
}
.sprint-table td {
  padding: 9px 10px;
  border-bottom: 1px solid var(--border-subtle);
}
.sprint-table__name { font-weight: 500; }
.sprint-table__date {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}
.sprint-table__num {
  font-family: var(--font-mono);
  font-size: 12px;
}

/* ---- notes block ---- */
.notes-block {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 16px;
}
.notes-block__md {
  margin: 0;
  font-size: 13px;
  font-family: var(--font-mono);
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

/* ---- footer ---- */
.report-footer {
  text-align: center;
  padding: 16px 0;
}
.report-footer__text {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
}

/* ---- print ---- */
@media print {
  .no-print { display: none !important; }
  .delivery-report--print {
    max-width: 100%;
    padding: 20px;
  }
  .report-section {
    break-inside: avoid;
  }
}

.text-success { color: var(--success-500); }

@media (max-width: 768px) {
  .metrics-grid { grid-template-columns: repeat(2, 1fr); }
  .stat-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
