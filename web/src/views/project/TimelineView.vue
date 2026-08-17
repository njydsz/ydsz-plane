<script setup lang="ts">
/**
 * TimelineView — ECharts 甘特图风格时间线视图。
 *
 * 功能：
 *   - 基于 ECharts bar 系列 + 时间 xAxis，每个 issue 一行横条
 *   - 横条从 start_date/created_at 延伸到 target_date
 *   - 颜色按 state.color 着色
 *   - Toolbar：时间缩放（周/月/季/全部）、标注开关、状态筛选、指派人筛选
 *   - 点击横条跳转需求/任务/缺陷详情
 *   - 键盘导航 ←→ ↑↓ 切换行
 *   - 空状态：没有 target_date 时提示
 *
 * Issue 接口已有 start_date/target_date 字段（issueApi.listIssues 已返回）。
 * 不涉及后端改动。
 */
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  shallowRef,
  watch,
} from "vue";
import { useRoute, useRouter } from "vue-router";
import * as echarts from "echarts";
import dayjs from "dayjs";

import { issueApi, type Issue, type State } from "@/api/services/issue";
import { ApiError } from "@/api/client";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppErrorState, AppEmptyState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

// --- Route ---
const route = useRoute();
const router = useRouter();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const workspaceId = computed(() => Number(route.params.workspaceId));
const wsId = computed(() => wsStore.current?.id ?? 0);

// --- State ---
const loading = ref(true);
const error = ref("");
const issues = ref<Issue[]>([]);
const states = ref<State[]>([]);

/** 时间缩放级别 */
type ZoomLevel = "week" | "month" | "quarter" | "all";
const ZOOM_LEVELS: { value: ZoomLevel; label: string }[] = [
  { value: "week", label: "周" },
  { value: "month", label: "月" },
  { value: "quarter", label: "季" },
  { value: "all", label: "全部" },
];
const zoom = ref<ZoomLevel>("month");

/** 筛选：全部状态 */
const filterState = ref<"all" | string>("all");
/** 筛选：指派人 */
const filterAssignee = ref<number | null>(null);
/** 显示标注 */
const showLabels = ref(true);

/** 键盘导航活跃行 */
const activeRow = ref(0);

// --- Refs ---
const chartEl = ref<HTMLDivElement | null>(null);
const chart = shallowRef<echarts.ECharts | null>(null);

// --- 数据加载 ---
async function loadData() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const [res, stRes] = await Promise.all([
      issueApi.listIssues(wsId.value, projectId.value, { limit: 200 }),
      issueApi.listStates(wsId.value, projectId.value).catch(() => [] as State[]),
    ]);
    issues.value = res.results;
    states.value = stRes;
  } catch (e: unknown) {
    error.value = e instanceof ApiError ? e.message : "加载时间线数据失败";
    toast.error(error.value);
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  if (wsId.value) loadData();
});
watch(wsId, () => {
  if (wsId.value) loadData();
});

// --- 派生：过滤后的 issues ---
const filteredIssues = computed(() => {
  let list = issues.value;
  if (filterState.value !== "all") {
    list = list.filter((i) => i.state?.group === filterState.value);
  }
  if (filterAssignee.value != null) {
    list = list.filter((i) => i.assignees.includes(filterAssignee.value!));
  }
  return list;
});

/** 有日期的 issues（用于 Gantt 展示） */
const datedIssues = computed(() =>
  filteredIssues.value.filter((i) => i.start_date || i.created_at),
);

/** 所有日期边界 */
const dateBounds = computed(() => {
  let min = Infinity;
  let max = -Infinity;
  for (const i of datedIssues.value) {
    const s = dayjs(i.start_date ?? i.created_at).valueOf();
    const e = i.target_date
      ? dayjs(i.target_date).valueOf()
      : s + 7 * 86400000;
    if (s < min) min = s;
    if (e > max) max = e;
  }
  if (!isFinite(min) || !isFinite(max)) {
    const now = Date.now();
    min = now;
    max = now + 30 * 86400000;
  }
  // 各留 5% 缓冲
  const span = max - min;
  return { min: Math.floor(min - span * 0.05), max: Math.ceil(max + span * 0.05) };
});

/** 缩放范围 */
const zoomRange = computed((): [number, number] => {
  const all = dateBounds.value;
  if (zoom.value === "all") return [all.min, all.max];

  const now = dayjs().valueOf();
  const dayMs = 86400000;
  let halfSpan = 0;
  switch (zoom.value) {
    case "week":
      halfSpan = 4 * dayMs;
      break;
    case "month":
      halfSpan = 15 * dayMs;
      break;
    case "quarter":
      halfSpan = 45 * dayMs;
      break;
  }
  let lo = now - halfSpan;
  let hi = now + halfSpan;
  // 不超出实际数据边界
  lo = Math.max(lo, all.min);
  hi = Math.min(hi, all.max);
  return [lo, hi];
});

// --- 状态颜色 ---
const stateColorMap = computed(() => {
  const m: Record<number, string> = {};
  for (const s of states.value) m[s.id] = s.color;
  return m;
});

function barColor(issue: Issue): string {
  if (issue.state_id != null && stateColorMap.value[issue.state_id]) {
    return stateColorMap.value[issue.state_id];
  }
  // 回退：按优先级色
  switch (issue.priority) {
    case "urgent":
      return "#dc2626";
    case "high":
      return "#ea580c";
    case "medium":
      return "#d97706";
    case "low":
      return "#65a30d";
    default:
      return "#6b7280";
  }
}

// --- ECharts 选项 ---
function buildOption(): echarts.EChartsOption {
  const [zoomLo, zoomHi] = zoomRange.value;
  const dateFormat = (ms: number) => dayjs(ms).format("YYYY-MM-DD");

  // y 标签
  const yLabels = datedIssues.value.map(
    (i) => `${i.identifier}  ${i.name}`,
  );

  // 数据：每条 issue === [startOffset, endOffset]
  const seriesData = datedIssues.value.map((i) => {
    const s = dayjs(i.start_date ?? i.created_at).valueOf();
    const e = i.target_date
      ? dayjs(i.target_date).valueOf()
      : s + 7 * 86400000;
    return [
      { value: [s, e], itemStyle: { color: barColor(i) } },
    ];
  });

  return {
    grid: {
      left: 240,
      right: 40,
      top: 16,
      bottom: 40,
      containLabel: false,
    },
    tooltip: {
      trigger: "item" as const,
      formatter: (params: any) => {
        const idx = params.dataIndex;
        const issue = datedIssues.value[idx];
        if (!issue) return "";
        const start = dateFormat(dayjs(issue.start_date ?? issue.created_at).valueOf());
        const end = issue.target_date
          ? dateFormat(dayjs(issue.target_date).valueOf())
          : "未设定";
        return `<b>${issue.identifier}</b><br/>${issue.name}<br/>${start} → ${end}`;
      },
    },
    xAxis: {
      type: "time",
      min: zoomLo,
      max: zoomHi,
      axisLabel: {
        fontSize: 11,
        color: "#6b7280",
        formatter: (val: number) => dayjs(val).format("MM-DD"),
      },
      splitLine: {
        show: true,
        lineStyle: { color: "#f3f4f6", type: "dashed" },
      },
    },
    yAxis: {
      type: "category",
      data: yLabels,
      inverse: true,
      axisLabel: {
        width: 230,
        overflow: "truncate",
        fontSize: 12,
        color: "#374151",
        formatter: (v: string) => v.length > 32 ? v.slice(0, 32) + "…" : v,
      },
      axisTick: { show: false },
      axisLine: { show: false },
    },
    series: [
      {
        type: "custom",
        name: "timeline",
        label: showLabels.value
          ? {
              show: true,
              position: "inside",
              formatter: (p: any) => {
                const i = datedIssues.value[p.dataIndex];
                return i && i.name.length <= 12 ? i.name : "";
              },
              fontSize: 10,
              color: "#fff",
            }
          : undefined,
        renderItem(params: any, api: any) {
          const idx = params.dataIndex;
          const issue = datedIssues.value[idx];
          if (!issue) return { type: "group", children: [] };

          const startMs = dayjs(issue.start_date ?? issue.created_at).valueOf();
          const endMs = issue.target_date
            ? dayjs(issue.target_date).valueOf()
            : startMs + 7 * 86400000;

          const startPx = api.coord([startMs, idx]);
          const endPx = api.coord([endMs, idx]);
          const yPx = startPx[1];
          const barH = api.size([0, 1])[1] * 0.6;
          const x = startPx[0];
          const w = Math.max(endPx[0] - x, 4);

          return {
            type: "rect",
            shape: { x, y: yPx - barH / 2, width: w, height: barH, r: 4 },
            style: {
              fill: barColor(issue),
              opacity: 0.85,
            },
          };
        },
        encode: { x: [0, 1], y: 0 },
        data: seriesData,
      } as echarts.CustomSeriesOption,
    ],
    dataZoom: [
      { type: "inside", xAxisIndex: 0, filterMode: "none" },
    ],
  };
}

// --- 渲染 ---
function render() {
  if (!chartEl.value) return;
  if (!chart.value) {
    chart.value = echarts.init(chartEl.value);
    chart.value.on("click", (params: any) => {
      const idx = params.dataIndex;
      const issue = datedIssues.value[idx];
      if (issue) goToIssue(issue.id);
    });
  }
  chart.value.setOption(buildOption(), true);
}

function handleResize() {
  chart.value?.resize();
}

// --- 监听 ---
watch(
  [datedIssues, zoom, showLabels],
  () => {
    void nextTick(() => render());
  },
  { deep: true },
);

// --- 导航 ---
function goToIssue(id: number) {
  router.push(`/${workspaceId.value}/projects/${projectId.value}/issues/${id}`);
}

// --- 键盘导航 ---
function handleKeydown(e: KeyboardEvent) {
  if (!datedIssues.value.length) return;
  const group = e.target as HTMLElement;
  if (group?.tagName === "INPUT" || group?.tagName === "SELECT") return;

  switch (e.key) {
    case "ArrowDown":
      e.preventDefault();
      activeRow.value = Math.min(activeRow.value + 1, datedIssues.value.length - 1);
      focusActiveRow();
      break;
    case "ArrowUp":
      e.preventDefault();
      activeRow.value = Math.max(activeRow.value - 1, 0);
      focusActiveRow();
      break;
    case "Enter":
      e.preventDefault();
      goToIssue(datedIssues.value[activeRow.value].id);
      break;
  }
}

function focusActiveRow() {
  if (!chart.value) return;
  chart.value.dispatchAction({ type: "downplay" });
  chart.value.dispatchAction({ type: "highlight", seriesIndex: 0, dataIndex: activeRow.value });
}

onMounted(() => {
  window.addEventListener("resize", handleResize);
  document.addEventListener("keydown", handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  document.removeEventListener("keydown", handleKeydown);
  chart.value?.dispose();
  chart.value = null;
});

// --- 状态分组选项（用于筛选下拉）---
const stateGroups = computed(() => {
  const groups = new Set<string>();
  for (const i of issues.value) {
    if (i.state?.group) groups.add(i.state.group);
  }
  return Array.from(groups);
});

const groupLabel = (g: string) =>
  (
    {
      backlog: "待办",
      started: "进行中",
      completed: "已完成",
      cancelled: "已取消",
    } as Record<string, string>
  )[g] ?? g;

// --- 切换缩放 ---
function setZoom(z: ZoomLevel) {
  zoom.value = z;
}

// --- 指派人列表（去重） ---
const uniqueAssignees = computed(() => {
  const map = new Map<number, string>();
  for (const i of issues.value) {
    for (const aid of i.assignees) {
      if (!map.has(aid)) map.set(aid, `U${aid}`);
    }
  }
  return Array.from(map.entries()).map(([id, label]) => ({ id, label }));
});
</script>

<template>
  <div class="timeline-view">
    <!-- 顶部栏 -->
    <header class="page-header">
      <div>
        <h1 class="page-title">时间线</h1>
        <p class="page-subtitle">
          甘特图式时间线 · 共 {{ datedIssues.length }} 项需求/任务/缺陷有日期
        </p>
      </div>

      <div class="header-actions">
        <div class="range-tabs">
          <button
            v-for="z in ZOOM_LEVELS"
            :key="z.value"
            :class="['range-tab', { active: zoom === z.value }]"
            @click="setZoom(z.value)"
          >
            {{ z.label }}
          </button>
        </div>
      </div>
    </header>

    <!-- 工具栏筛选 -->
    <div class="toolbar">
      <div class="filter-group">
        <label class="filter-label">状态</label>
        <select v-model="filterState" class="filter-select">
          <option value="all">全部</option>
          <option v-for="g in stateGroups" :key="g" :value="g">
            {{ groupLabel(g) }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <label class="filter-label">指派人</label>
        <select v-model="filterAssignee" class="filter-select">
          <option :value="null">全部</option>
          <option v-for="a in uniqueAssignees" :key="a.id" :value="a.id">
            {{ a.label }}
          </option>
        </select>
      </div>

      <label class="toggle-label">
        <input v-model="showLabels" type="checkbox" />
        标注
      </label>

      <button class="btn-refresh" :disabled="loading" @click="loadData">
        {{ loading ? "加载中…" : "刷新" }}
      </button>
    </div>

    <!-- 加载 / 错误 / 空 -->
    <AppSkeleton v-if="loading" variant="board" :cols="1" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadData" />

    <template v-else>
      <!-- 没有日期数据 -->
      <AppEmptyState
        v-if="datedIssues.length === 0"
        title="暂无时间安排"
        description="为需求/任务/缺陷设置开始日期和截止日期后，这里将展示时间线。"
      />

      <!-- Gantt 图表 -->
      <div v-else class="chart-wrapper">
        <div ref="chartEl" class="gantt-chart"></div>
      </div>
    </template>

    <div class="footer-hint">
      提示: ↑↓ 导航 · Enter 打开详情 · 滚轮缩放 · 拖拽平移
    </div>
  </div>
</template>

<style scoped>
.timeline-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--surface-1, #fff);
}

/* --- Header --- */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px 24px 0;
  flex-shrink: 0;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary, #111827);
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-tertiary, #6b7280);
  margin: 4px 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.range-tabs {
  display: flex;
  background: var(--surface-2, #f3f4f6);
  border-radius: 6px;
  padding: 2px;
}

.range-tab {
  padding: 5px 14px;
  font-size: 13px;
  border: none;
  background: transparent;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.range-tab.active {
  background: var(--surface-1, #fff);
  color: var(--text-primary, #111827);
  font-weight: 500;
}

/* --- Toolbar --- */
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 24px;
  border-bottom: 1px solid var(--border, #e5e7eb);
  background: var(--surface-2, #f9fafb);
  flex-shrink: 0;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
}

.filter-select {
  padding: 4px 8px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border, #e5e7eb);
  border-radius: 5px;
  background: var(--surface-1, #fff);
  color: var(--text-primary);
  cursor: pointer;
}

.filter-select:focus {
  border-color: var(--brand-500, #3b82f6);
  outline: none;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
}

.toggle-label input {
  accent-color: var(--brand-500, #3b82f6);
}

.btn-refresh {
  margin-left: auto;
  padding: 4px 12px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border, #e5e7eb);
  border-radius: 5px;
  background: var(--surface-1, #fff);
  color: var(--text-primary);
  cursor: pointer;
  transition: background 0.1s;
}

.btn-refresh:hover:not(:disabled) {
  background: var(--surface-2, #f3f4f6);
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* --- Chart --- */
.chart-wrapper {
  flex: 1;
  padding: 0 24px 0;
  min-height: 0;
}

.gantt-chart {
  width: 100%;
  height: 100%;
  min-height: 400px;
}

/* --- Footer --- */
.footer-hint {
  padding: 6px 24px;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  border-top: 1px solid var(--border, #e5e7eb);
  background: var(--surface-2, #f9fafb);
  opacity: 0.7;
  flex-shrink: 0;
}
</style>
