<script setup lang="ts">
/**
 * Focus Mode 沉浸视图（S8.7）
 *
 * 全屏专注模式：隐藏侧栏、顶栏等一切干扰元素，仅展示单个工作项的完整内容。
 * 适用场景：深度阅读/审阅工作项详情、文档评审、远程演示。
 *
 * 特性：
 *   - ESC 键退出沉浸模式
 *   - 显示工作项 ID 徽章、类型、状态、优先级
 *   - 渲染富文本描述（TipTap HTML 输出）
 *   - 展示关键字段（经办人、故事点、严重程度、发现版本、修复版本等）
 *   - 显示工时汇总与最近活动
 *   - 快速链接：编辑 / 返回详情页 / 看板
 */

import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { issueApi, type IssueActivity, type State, type TimeLog } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { versionApi, type Version } from "@/api/services/version";
import { AppBadge, AppCard } from "@/components";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
}>();

const router = useRouter();

// ---- 数据 ----
const issue = ref<any>(null);
const states = ref<State[]>([]);
const activities = ref<IssueActivity[]>([]);
const timeLogs = ref<TimeLog[]>([]);
const versions = ref<Version[]>([]);
const loading = ref(true);
const error = ref("");

// ---- 活动动词 → 中文标签（与 ActivityTimelineWidget 保持一致） ----
function verbLabel(verb: string): string {
  const map: Record<string, string> = {
    created: "创建了",
    updated: "更新了",
    transitioned: "流转了",
    commented: "评论了",
    assigned: "指派了",
    unassigned: "取消指派了",
  };
  return map[verb] ?? verb;
}

// ---- 派生 ----
const totalMinutes = computed(() =>
  timeLogs.value.reduce((s, t) => s + (t.duration_minutes ?? 0), 0),
);

const stateInfo = computed(() =>
  states.value.find((s) => s.id === issue.value?.state_id),
);

const typeLabel = computed(() => {
  const t = issue.value?.type;
  return ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[t] ?? t ?? "";
});

const priorityLabel = computed(() => {
  const p = issue.value?.priority;
  return ({
    urgent: "紧急",
    high: "高",
    medium: "中",
    low: "低",
    none: "无",
  } as Record<string, string>)[p] ?? p ?? "-";
});

const priorityVariant = computed<"default" | "success" | "info" | "warning" | "brand" | "danger">(() => {
  const p = issue.value?.priority;
  const map: Record<string, "default" | "success" | "info" | "warning" | "brand" | "danger"> = {
    urgent: "danger",
    high: "warning",
    medium: "info",
    low: "default",
    none: "default",
  };
  return map[p] ?? "default";
});

const severityLabel = computed(() => {
  const s = issue.value?.severity;
  if (!s) return "-";
  const map: Record<number, string> = {
    5: "S5 · 致命",
    4: "S4 · 严重",
    3: "S3 · 一般",
    2: "S2 · 轻微",
    1: "S1 · 建议",
  };
  return map[s] ?? `S${s}`;
});

function fmtDuration(mins: number): string {
  if (mins < 60) return `${mins}分钟`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`;
}

const foundVersion = computed(() =>
  versions.value.find((v) => v.id === issue.value?.found_version_id),
);
const fixVersion = computed(() =>
  versions.value.find((v) => v.id === issue.value?.fix_version_id),
);

// ---- 加载 ----
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.get(props.workspaceId);
    const [iss, st, acts] = await Promise.all([
      issueApi.getIssue(ws.id, props.projectId, props.issueId),
      issueApi.listStates(ws.id, props.projectId),
      issueApi.listActivities(ws.id, props.projectId, props.issueId),
    ]);
    issue.value = iss;
    states.value = st;
    activities.value = acts.results ?? [];

    // 并行加载工时与版本（非关键，失败静默）
    const [tl, vs] = await Promise.allSettled([
      issueApi.listTimeLogs(ws.id, props.projectId, props.issueId),
      versionApi.listVersions(ws.id, props.projectId),
    ]);
    if (tl.status === "fulfilled") timeLogs.value = tl.value.results ?? [];
    if (vs.status === "fulfilled") {
      const vData = vs.value;
      versions.value = Array.isArray(vData) ? vData : (vData as any).results ?? [];
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/** 是否有需要显示的属性字段 */
const hasAttributes = computed(() => {
  if (!issue.value) return false;
  return (
    issue.value.assignee_name ||
    (issue.value.point != null && issue.value.point >= 0) ||
    issue.value.found_phase ||
    foundVersion.value ||
    fixVersion.value ||
    issue.value.expected_date
  );
});

// ---- 导航 ----
function exitFocus() {
  router.push(`/${props.workspaceId}/projects/${props.projectId}/issues/${props.issueId}`);
}
function goToBoard() {
  router.push(`/${props.workspaceId}/projects/${props.projectId}/board`);
}
function goToEdit() {
  // 直接在详情页中带编辑查询参数（预留）
  router.push(`/${props.workspaceId}/projects/${props.projectId}/issues/${props.issueId}`);
}

// ---- 键盘快捷键 ----
function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") exitFocus();
}

onMounted(() => {
  void load();
  document.addEventListener("keydown", handleKeydown);
  // 沉浸模式下给 body 加 class，隐藏全局滚动条
  document.body.classList.add("focus-mode-active");
});
onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.classList.remove("focus-mode-active");
});

// ---- 时间格式化 ----
function fmtTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "刚刚";
  if (diffMin < 60) return `${diffMin}分钟前`;
  const diffH = Math.floor(diffMin / 60);
  if (diffH < 24) return `${diffH}小时前`;
  const diffDay = Math.floor(diffH / 24);
  if (diffDay < 30) return `${diffDay}天前`;
  return d.toLocaleDateString("zh-CN");
}
</script>

<template>
  <!-- eslint-disable vue/no-v-html -- description_html 由服务端 bluemonday 白名单清洗后落库，属受信输出 -->
  <div class="focus-mode">
    <!-- 顶部工具栏 -->
    <header class="focus-mode__toolbar">
      <button class="focus-mode__back" title="退出专注模式 (ESC)" @click="exitFocus">
        ← 退出
      </button>
      <div class="focus-mode__title">
        <span class="focus-mode__identifier">{{ issue?.identifier ?? "..." }}</span>
        <span class="focus-mode__name">{{ issue?.name ?? "加载中..." }}</span>
      </div>
      <div class="focus-mode__actions">
        <button class="focus-mode__action-btn" @click="goToBoard">看板</button>
        <button class="focus-mode__action-btn" @click="goToEdit">编辑</button>
      </div>
    </header>

    <!-- 加载 / 错误 -->
    <div v-if="loading" class="focus-mode__loading">
      <span class="focus-mode__spinner"></span>
      <p>加载工作项...</p>
    </div>
    <div v-else-if="error" class="focus-mode__error">
      <p>{{ error }}</p>
      <button @click="load">重试</button>
    </div>

    <!-- 主体内容 -->
    <main v-else class="focus-mode__body">
      <div class="focus-mode__content">
        <!-- 头部：徽章区域 -->
        <section class="focus-mode__header">
          <div class="focus-mode__badges">
            <AppBadge variant="brand">{{ typeLabel }}</AppBadge>
            <AppBadge :variant="issue?.type === 'defect' ? 'danger' : 'success'">
              {{ stateInfo?.name ?? "未知" }}
            </AppBadge>
            <AppBadge :variant="priorityVariant">{{ priorityLabel }}优先级</AppBadge>
            <AppBadge v-if="issue?.severity" variant="warning">{{ severityLabel }}</AppBadge>
          </div>
        </section>

        <!-- 描述 -->
        <section class="focus-mode__description">
          <div
            v-if="issue?.description_html"
            class="focus-mode__prose"
            v-html="issue.description_html"
          />
          <p v-else class="focus-mode__empty">暂无描述</p>
        </section>

        <!-- 属性网格 -->
        <section class="focus-mode__attributes">
          <AppCard v-if="hasAttributes" padding="sm" shadow>
            <div class="attr-grid">
              <div v-if="issue?.assignee_name" class="attr-item">
                <span class="attr-label">经办人</span>
                <span class="attr-value">{{ issue.assignee_name }}</span>
              </div>
              <div v-if="issue?.point != null && issue?.point >= 0" class="attr-item">
                <span class="attr-label">故事点</span>
                <span class="attr-value">{{ issue.point }}</span>
              </div>
              <div v-if="issue?.found_phase" class="attr-item">
                <span class="attr-label">发现阶段</span>
                <span class="attr-value">{{ issue.found_phase }}</span>
              </div>
              <div v-if="foundVersion" class="attr-item">
                <span class="attr-label">发现版本</span>
                <span class="attr-value">{{ foundVersion.name }}</span>
              </div>
              <div v-if="fixVersion" class="attr-item">
                <span class="attr-label">修复版本</span>
                <span class="attr-value">{{ fixVersion.name }}</span>
              </div>
              <div v-if="issue?.expected_date" class="attr-item">
                <span class="attr-label">预计完成</span>
                <span class="attr-value">{{ issue.expected_date }}</span>
              </div>
            </div>
          </AppCard>
        </section>

        <!-- 工时汇总 -->
        <section v-if="timeLogs.length > 0" class="focus-mode__timesummary">
          <h3 class="focus-mode__section-title">工时记录 ({{ fmtDuration(totalMinutes) }})</h3>
          <AppCard padding="sm">
            <ul class="time-list">
              <li v-for="tl in timeLogs.slice(0, 10)" :key="tl.id" class="time-item">
                <span class="time-duration">{{ fmtDuration(tl.duration_minutes) }}</span>
                <span class="time-desc">{{ tl.description || "—" }}</span>
                <span class="time-date">{{ tl.spent_date }}</span>
              </li>
            </ul>
          </AppCard>
        </section>

        <!-- 最近活动 -->
        <section v-if="activities.length > 0" class="focus-mode__activity">
          <h3 class="focus-mode__section-title">最近活动</h3>
          <AppCard padding="sm">
            <ul class="activity-list">
              <li v-for="act in activities.slice(0, 8)" :key="act.id" class="activity-item">
                <span class="activity-action">{{ verbLabel(act.verb) }}</span>
                <span class="activity-meta">
                  {{ act.actor_name ?? "系统" }} · {{ fmtTime(act.created_at) }}
                </span>
              </li>
            </ul>
          </AppCard>
        </section>
      </div>
    </main>

    <!-- 底部提示 -->
    <footer class="focus-mode__footer">
      <kbd>ESC</kbd> 退出专注模式
    </footer>
  </div>
</template>

<style scoped>
.focus-mode {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: var(--bg-base, #f9fafb);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ---- 工具栏 ---- */
.focus-mode__toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 24px;
  background: var(--surface-1, #fff);
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  flex-shrink: 0;
}

.focus-mode__back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary, #4b5563);
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}
.focus-mode__back:hover {
  background: var(--surface-3, #f3f4f6);
  color: var(--text-primary, #1f2937);
}

.focus-mode__title {
  flex: 1;
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
}

.focus-mode__identifier {
  font-size: 14px;
  font-weight: 600;
  color: var(--brand-500, #3b82f6);
  flex-shrink: 0;
}

.focus-mode__name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.focus-mode__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.focus-mode__action-btn {
  padding: 5px 12px;
  font-size: 12px;
  color: var(--text-secondary, #4b5563);
  background: none;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.focus-mode__action-btn:hover {
  background: var(--surface-2, #f9fafb);
  border-color: var(--border-default, #d1d5db);
}

/* ---- 加载 / 错误 ---- */
.focus-mode__loading,
.focus-mode__error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-tertiary, #9ca3af);
  font-size: 14px;
}
.focus-mode__spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-subtle, #e5e7eb);
  border-top-color: var(--brand-500, #3b82f6);
  border-radius: 50%;
  animation: focus-spin 0.8s linear infinite;
}
@keyframes focus-spin {
  to { transform: rotate(360deg); }
}
.focus-mode__error {
  color: var(--danger-500, #ef4444);
}

/* ---- 主体内容 ---- */
.focus-mode__body {
  flex: 1;
  overflow-y: auto;
  padding: 24px 16px;
}

.focus-mode__content {
  max-width: 760px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ---- 头部徽章 ---- */
.focus-mode__header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.focus-mode__badges {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

/* ---- 描述（富文本渲染） ---- */
.focus-mode__description {
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: 8px;
  padding: 20px 24px;
}

.focus-mode__empty {
  color: var(--text-tertiary, #9ca3af);
  font-size: 14px;
  font-style: italic;
  margin: 0;
  text-align: center;
  padding: 20px 0;
}

/* Prose 样式（v-html 注入的 TipTap 输出） */
.focus-mode__prose :deep(h2) {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 8px;
}
.focus-mode__prose :deep(h3) {
  font-size: 16px;
  font-weight: 600;
  margin: 12px 0 6px;
}
.focus-mode__prose :deep(p) {
  margin: 0 0 10px;
  font-size: 14px;
  line-height: 1.7;
}
.focus-mode__prose :deep(ul), .focus-mode__prose :deep(ol) {
  padding-left: 24px;
  margin: 6px 0 12px;
}
.focus-mode__prose :deep(li) {
  margin-bottom: 4px;
}
.focus-mode__prose :deep(blockquote) {
  border-left: 3px solid var(--brand-200, #bfdbfe);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--text-secondary, #4b5563);
}
.focus-mode__prose :deep(code) {
  background: var(--surface-3, #f3f4f6);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}
.focus-mode__prose :deep(pre) {
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 12px;
  margin: 8px 0;
  overflow-x: auto;
}
.focus-mode__prose :deep(pre code) {
  background: none;
  padding: 0;
}
.focus-mode__prose :deep(a) {
  color: var(--brand-500, #3b82f6);
  text-decoration: underline;
}
.focus-mode__prose :deep(img) {
  max-width: 100%;
  border-radius: 6px;
  margin: 8px 0;
}
.focus-mode__prose :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
}
.focus-mode__prose :deep(th), .focus-mode__prose :deep(td) {
  border: 1px solid var(--border-subtle);
  padding: 8px 12px;
}
.focus-mode__prose :deep(th) {
  background: var(--surface-2, #f9fafb);
}

/* Callout */
.focus-mode__prose :deep(div[data-type="callout"]) {
  border-radius: 6px;
  padding: 12px 16px;
  margin: 8px 0;
  border-left: 3px solid;
}
.focus-mode__prose :deep(div[data-type="callout"][data-callout-type="info"]) {
  background: var(--brand-50, #eef2fe);
  border-color: var(--brand-default, #3b82f6);
}
.focus-mode__prose :deep(div[data-type="callout"][data-callout-type="warning"]) {
  background: var(--warning-50, #fffbeb);
  border-color: var(--warning-500, #f59e0b);
}
.focus-mode__prose :deep(div[data-type="callout"][data-callout-type="error"]) {
  background: var(--danger-50, #fef2f2);
  border-color: var(--danger-500, #ef4444);
}
.focus-mode__prose :deep(div[data-type="callout"][data-callout-type="success"]) {
  background: var(--success-50, #ecfdf5);
  border-color: var(--success-500, #10b981);
}

/* Task list */
.focus-mode__prose :deep(ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;
}
.focus-mode__prose :deep(ul[data-type="taskList"] li) {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-bottom: 4px;
}

/* ---- 属性网格 ---- */
.attr-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 12px;
}
.attr-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.attr-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary, #9ca3af);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.attr-value {
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  word-break: break-all;
}

/* ---- Section title ---- */
.focus-mode__section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #4b5563);
  margin: 0 0 8px;
}

/* ---- 工时列表 ---- */
.time-list,
.activity-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.time-item {
  display: grid;
  grid-template-columns: 100px 1fr auto;
  gap: 10px;
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border-subtle, #f3f4f6);
}
.time-duration {
  font-weight: 600;
  color: var(--brand-500, #3b82f6);
}
.time-desc {
  color: var(--text-primary, #1f2937);
}
.time-date {
  color: var(--text-tertiary, #9ca3af);
  font-size: 12px;
}

/* ---- 活动列表 ---- */
.activity-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border-subtle, #f3f4f6);
  font-size: 13px;
}
.activity-action {
  color: var(--text-primary, #1f2937);
}
.activity-meta {
  color: var(--text-tertiary, #9ca3af);
  font-size: 12px;
  flex-shrink: 0;
}

/* ---- 底部 ---- */
.focus-mode__footer {
  padding: 8px;
  text-align: center;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  background: var(--surface-1, #fff);
  border-top: 1px solid var(--border-subtle, #e5e7eb);
  flex-shrink: 0;
}
.focus-mode__footer kbd {
  display: inline-block;
  padding: 2px 6px;
  font-family: var(--font-mono, monospace);
  font-size: 10px;
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-subtle);
  border-radius: 3px;
  margin: 0 2px;
}

/* ---- 全局：沉浸模式下隐藏根元素 ---- */
body.focus-mode-active #app > *:not(.focus-mode) {
  display: none !important;
}
</style>
