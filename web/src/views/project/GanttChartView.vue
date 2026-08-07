<script setup lang="ts">
/**
 * GanttChartView — 只读项目甘特图视图。
 *
 * 纯前端渲染，数据由后端 gantt 接口一次性返回。
 * 功能：
 *  - 时间轴条块（按 start_date/target_date 定位宽度与偏移）
 *  - 依赖箭头（SVG 连线显示工作项依赖关系）
 *  - 状态/优先级颜色编码
 *  - 时间范围缩放（周/月/季/全部）
 *  - WBS 层级缩进
 *  - 点击行跳转详情
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import dayjs from "dayjs";

import { ganttApi } from "@/api/services/gantt";
import type { GanttIssue, GanttDependency } from "@/api/services/gantt";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppErrorState, AppEmptyState, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceId));
const wsId = computed(() => wsStore.current?.id ?? 0);

const loading = ref(true);
const error = ref("");
const issues = ref<GanttIssue[]>([]);
const dependencies = ref<GanttDependency[]>([]);

/** 时间范围 */
type TimeRange = 'week' | 'month' | 'quarter' | 'all';
const TIME_RANGES: { value: TimeRange; label: string }[] = [
  { value: 'week', label: '周' },
  { value: 'month', label: '月' },
  { value: 'quarter', label: '季' },
  { value: 'all', label: '全部' },
];
const selectedRange = ref<TimeRange>('quarter');

/** 计算时间范围日期 */
function calcDateRange(range: TimeRange): { from?: string; to?: string } {
  const today = dayjs();
  switch (range) {
    case 'week':
      return { from: today.startOf('week').format('YYYY-MM-DD'), to: today.endOf('week').add(7, 'day').format('YYYY-MM-DD') };
    case 'month':
      return { from: today.startOf('month').format('YYYY-MM-DD'), to: today.endOf('month').add(1, 'month').format('YYYY-MM-DD') };
    case 'quarter':
      return { from: today.subtract(1, 'month').startOf('month').format('YYYY-MM-DD'), to: today.add(3, 'month').endOf('month').format('YYYY-MM-DD') };
    case 'all':
      return {};
  }
}

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = '';
  try {
    const { from, to } = calcDateRange(selectedRange.value);
    const data = await ganttApi.getGanttData(wsId.value, projectId.value, { date_from: from, date_to: to });
    issues.value = data.issues;
    dependencies.value = data.dependencies;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败';
  } finally {
    loading.value = false;
  }
}

function setRange(range: TimeRange) {
  selectedRange.value = range;
  void load();
}

onMounted(() => {
  if (wsId.value) load();
});
watch([wsId], () => { if (wsId.value) load(); });

/** 跳转工作项详情 */
function goToIssue(issueId: number) {
  void router.push(`/${workspaceSlug.value}/projects/${projectId.value}/issues/${issueId}`);
}

// --- 时间轴计算 ---

/** 所有条目的日期范围 */
const timelineRange = computed(() => {
  let min = Infinity;
  let max = -Infinity;
  let hasDate = false;
  for (const issue of issues.value) {
    if (issue.start_date) {
      const t = dayjs(issue.start_date).valueOf();
      min = Math.min(min, t);
      hasDate = true;
    }
    if (issue.target_date) {
      const t = dayjs(issue.target_date).valueOf();
      max = Math.max(max, t);
      hasDate = true;
    }
  }
  if (!hasDate) {
    const now = Date.now();
    min = now;
    max = now + 30 * 86400000;
  }
  // 确保最少 7 天宽度
  const range = Math.max(max - min, 7 * 86400000);
  return { min, max: min + range };
});

/** 百分比定位 */
function percentFromDate(dateStr?: string): number {
  if (!dateStr) return 0;
  const t = dayjs(dateStr).valueOf();
  const { min, max } = timelineRange.value;
  return ((t - min) / (max - min)) * 100;
}

function widthFromDates(start?: string, end?: string): number {
  const s = start ? dayjs(start).valueOf() : timelineRange.value.min;
  const e = end ? dayjs(end).valueOf() : s + 7 * 86400000;
  const { min, max } = timelineRange.value;
  return Math.max(((e - s) / (max - min)) * 100, 1.5);
}

/** 月份刻度 */
const monthTicks = computed(() => {
  const { min, max } = timelineRange.value;
  const result: { label: string; left: number }[] = [];
  const cursor = dayjs(min).startOf('month');
  const end = dayjs(max);
  while (cursor.isBefore(end) || cursor.isSame(end, 'day')) {
    const ts = cursor.valueOf();
    if (ts >= min) {
      result.push({
        label: cursor.format('YYYY-MM'),
        left: ((ts - min) / (max - min)) * 100,
      });
    }
    cursor.add(1, 'month');
  }
  return result;
});

/** 颜色映射 */
function stateColor(group: string): string {
  switch (group) {
    case 'completed': return '#16a34a';
    case 'started': return '#2563eb';
    case 'cancelled': return '#6b7280';
    default: return '#d97706';
  }
}

function priorityColor(priority: string): string {
  switch (priority) {
    case 'urgent': return '#dc2626';
    case 'high': return '#ea580c';
    case 'medium': return '#d97706';
    case 'low': return '#65a30d';
    default: return '#6b7280';
  }
}

/** 依赖箭头 SVG 路径 - 从 source 右侧连到 target 左侧 */
interface ArrowPath {
  id: number;
  d: string;
}

const arrowPaths = computed<ArrowPath[]>(() => {
  const ROW_HEIGHT = 36;
  const HEADER_HEIGHT = 28;
  const issueMap = new Map<number, number>();
  issues.value.forEach((issue, idx) => {
    issueMap.set(issue.id, idx);
  });

  return dependencies.value
    .map((dep) => {
      const srcIdx = issueMap.get(dep.source_id);
      const tgtIdx = issueMap.get(dep.target_id);
      if (srcIdx === undefined || tgtIdx === undefined) return null;

      const srcY = HEADER_HEIGHT + srcIdx * ROW_HEIGHT + ROW_HEIGHT / 2;
      const tgtY = HEADER_HEIGHT + tgtIdx * ROW_HEIGHT + ROW_HEIGHT / 2;

      const srcIssue = issues.value[srcIdx];
      const tgtIssue = issues.value[tgtIdx];
      const srcRight = percentFromDate(srcIssue.start_date) + widthFromDates(srcIssue.start_date, srcIssue.target_date);
      const tgtLeft = percentFromDate(tgtIssue.start_date);

      const x1 = srcRight;
      const x2 = tgtLeft;
      const midX = (x1 + x2) / 2;

      return {
        id: dep.id,
        d: `M ${x1} ${srcY} C ${midX} ${srcY}, ${midX} ${tgtY}, ${x2} ${tgtY}`,
      };
    })
    .filter(Boolean) as ArrowPath[];
});

const ROW_HEIGHT = 36;
const HEADER_HEIGHT = 28;
const svgHeight = computed(() => HEADER_HEIGHT + issues.value.length * ROW_HEIGHT + 8);
</script>

<template>
  <div class="gantt-view">
    <!-- 顶部栏 -->
    <header class="page-header">
      <div>
        <h1 class="page-title">甘特图</h1>
        <p class="page-subtitle">工作时间线与依赖关系（只读）</p>
      </div>
      <div class="header-actions">
        <div class="range-tabs">
          <button
            v-for="r in TIME_RANGES"
            :key="r.value"
            :class="['range-tab', { active: selectedRange === r.value }]"
            @click="setRange(r.value)"
          >
            {{ r.label }}
          </button>
        </div>
      </div>
    </header>

    <!-- 状态 -->
    <AppSkeleton v-if="loading" variant="board" :cols="2" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- 主内容 -->
    <div v-else class="gantt-container">
      <!-- 列标题 -->
      <div class="gantt-header-row">
        <div class="issue-col">工作项</div>
        <div class="timeline-col">
          <!-- 月份刻度 -->
          <div class="month-ticks">
            <span
              v-for="tick in monthTicks"
              :key="tick.left"
              class="month-tick"
              :style="{ left: `${tick.left}%` }"
            >{{ tick.label }}</span>
          </div>
        </div>
      </div>

      <!-- 空态 -->
      <AppEmptyState v-if="issues.length === 0" description="该项目暂无工作项的时间数据" />

      <!-- 甘特内容 -->
      <div v-else class="gantt-body">
        <div class="gantt-scroll">
          <div
            v-for="issue in issues"
            :key="issue.id"
            class="gantt-row"
            @click="goToIssue(issue.id)"
          >
            <!-- 左侧：工作项标识 -->
            <div class="issue-col">
              <span class="issue-badge" :style="{ borderColor: priorityColor(issue.priority) }">
                {{ issue.identifier }}
              </span>
              <span class="issue-name" :title="issue.name">{{ issue.name }}</span>
            </div>

            <!-- 右侧：时间轴 -->
            <div class="timeline-col">
              <div class="timeline-track">
                <div
                  v-if="issue.start_date || issue.target_date"
                  class="gantt-bar"
                  :style="{
                    left: `${percentFromDate(issue.start_date)}%`,
                    width: `${Math.max(widthFromDates(issue.start_date, issue.target_date), 1.5)}%`,
                    backgroundColor: stateColor(issue.state?.group ?? 'backlog'),
                  }"
                >
                  <span class="bar-progress" :style="{ width: `${issue.progress}%` }"></span>
                </div>
              </div>
            </div>
          </div>

          <!-- 依赖箭头 SVG 层 -->
          <svg
            v-if="arrowPaths.length > 0"
            class="arrows-layer"
            :viewBox="`0 0 100 ${svgHeight}`"
            preserveAspectRatio="none"
            :height="svgHeight"
          >
            <path
              v-for="arrow in arrowPaths"
              :key="arrow.id"
              :d="arrow.d"
              class="arrow-path"
              fill="none"
              markerEnd="url(#arrowhead)"
            />
            <defs>
              <marker id="arrowhead" markerWidth="8" markerHeight="6" refX="7" refY="3" orient="auto">
                <polygon points="0 0, 8 3, 0 6" fill="#6b7280" />
              </marker>
            </defs>
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.gantt-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

.range-tabs {
  display: flex;
  background: var(--surface-2);
  border-radius: 6px;
  padding: 2px;
}

.range-tab {
  padding: 5px 14px;
  font-size: 13px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.range-tab.active {
  background: var(--surface-1);
  color: var(--text-primary);
  font-weight: 500;
}

/* 甘特图容器 */
.gantt-container {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.gantt-header-row {
  display: flex;
  align-items: center;
  height: 28px;
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
}

.issue-col {
  width: 260px;
  flex-shrink: 0;
  padding: 0 12px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.timeline-col {
  flex: 1;
  position: relative;
}

/* 月份刻度 */
.month-ticks {
  position: relative;
  height: 28px;
}

.month-tick {
  position: absolute;
  transform: translateX(-50%);
  font-size: 10px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
}

/* 行 */
.gantt-body {
  position: relative;
}

.gantt-scroll {
  position: relative;
  overflow-x: auto;
}

.gantt-row {
  display: flex;
  align-items: center;
  height: 36px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background 0.1s;
}

.gantt-row:hover {
  background: var(--surface-2);
}

.gantt-row .issue-col {
  font-size: 12px;
  color: var(--text-primary);
}

.issue-badge {
  font-size: 11px;
  font-family: var(--font-mono);
  padding: 1px 6px;
  border: 1px solid;
  border-radius: 4px;
  flex-shrink: 0;
  background: var(--surface-1);
}

.issue-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 时间轴条 */
.timeline-track {
  position: relative;
  height: 100%;
  width: 100%;
}

.gantt-bar {
  position: absolute;
  top: 8px;
  height: 18px;
  border-radius: 4px;
  min-width: 6px;
  display: flex;
  align-items: center;
  overflow: hidden;
}

.bar-progress {
  height: 100%;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 4px 0 0 4px;
}

/* 箭头 SVG 层 */
.arrows-layer {
  position: absolute;
  top: 0;
  left: 260px;
  right: 0;
  pointer-events: none;
}

.arrow-path {
  stroke: #6b7280;
  stroke-width: 1.2;
  stroke-dasharray: 4 2;
  opacity: 0.5;
}
</style>
