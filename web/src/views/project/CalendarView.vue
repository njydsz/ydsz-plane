<script setup lang="ts">
/**
 * CalendarView — 项目日历视图。
 *
 * 展示工作项在日历网格上的分布（按 target_date 排布）。
 * 支持月/周/日视图切换。
 * 纯前端渲染，数据由项目列表接口聚合。
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import dayjs from "dayjs";

import { issueApi } from "@/api/services/issue";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppErrorState, AppEmptyState, AppSkeleton } from "@/components";

interface CalendarEvent {
  id: number;
  identifier: string;
  name: string;
  type_code: string;
  state: { name: string; group: string; color: string };
  priority: string;
  target_date?: string;
  start_date?: string;
}

const route = useRoute();
const router = useRouter();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug));
const wsId = computed(() => wsStore.current?.id ?? 0);

const loading = ref(true);
const error = ref('');
const events = ref<CalendarEvent[]>([]); // 当前视图范围内的事件
const currentDate = ref(dayjs());

type ViewMode = 'month' | 'week' | 'day';
const VIEW_MODES: { value: ViewMode; label: string }[] = [
  { value: 'month', label: '月' },
  { value: 'week', label: '周' },
  { value: 'day', label: '日' },
];
const viewMode = ref<ViewMode>('month');

/** 视图范围 */
const viewRange = computed(() => {
  const base = currentDate.value;
  switch (viewMode.value) {
    case 'month':
      return { start: base.startOf('month').startOf('week'), end: base.endOf('month').endOf('week') };
    case 'week':
      return { start: base.startOf('week'), end: base.endOf('week') };
    case 'day':
      return { start: base.startOf('day'), end: base.endOf('day') };
  }
});

/** 日历网格（月视图） */
const calendarGrid = computed(() => {
  const { start, end } = viewRange.value;
  const days: { date: dayjs.Dayjs; isCurrentMonth: boolean; events: CalendarEvent[] }[] = [];
  let cursor = start;
  while (cursor.isBefore(end) || cursor.isSame(end, 'day')) {
    const dayEvents = events.value.filter((e) => {
      if (!e.target_date) return false;
      return dayjs(e.target_date).isSame(cursor, 'day');
    });
    days.push({
      date: cursor,
      isCurrentMonth: cursor.month() === currentDate.value.month(),
      events: dayEvents,
    });
    cursor = cursor.add(1, 'day');
  }
  return days;
});

/** 周视图行 */
const weekDays = computed(() => {
  const { start } = viewRange.value;
  const days: { date: dayjs.Dayjs; events: CalendarEvent[] }[] = [];
  for (let i = 0; i < 7; i++) {
    const d = start.add(i, 'day');
    days.push({
      date: d,
      events: events.value.filter((e) => e.target_date && dayjs(e.target_date).isSame(d, 'day')),
    });
  }
  return days;
});

/** 日视图小时行 */
const dayHours = computed(() => {
  const hours: { hour: number; events: CalendarEvent[] }[] = [];
  for (let h = 0; h < 24; h++) {
    hours.push({
      hour: h,
      events: [], // 工作项不具备小时级精度，日视图按全天展示
    });
  }
  // 日视图下所有当天事件都放入第一格
  if (viewMode.value === 'day' && events.value.length > 0) {
    const dayEvents = events.value.filter((e) => e.target_date && dayjs(e.target_date).isSame(currentDate.value, 'day'));
    hours[0].events = dayEvents;
  }
  return hours;
});

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = '';
  try {
    // 扩大范围加载（前后各一个月确保覆盖视图）
    const start = currentDate.value.subtract(2, 'month').format('YYYY-MM-DD');
    const end = currentDate.value.add(2, 'month').format('YYYY-MM-DD');

    // 使用 start_date_from / target_date_to 复合过滤获取时间窗口内的工作项
    const result = await issueApi.listIssues(wsId.value, projectId.value, {
      start_date_from: start,
      target_date_to: end,
      limit: 500,
    });
    events.value = result.results
      .filter((item) => item.target_date) // 日历视图仅展示有目标日期的工作项
      .map((item) => ({
        id: item.id,
        identifier: item.identifier,
        name: item.name,
        type_code: item.type_code,
        state: item.state ?? { name: '', group: 'backlog', color: '#6b7280' },
        priority: item.priority,
        target_date: item.target_date,
        start_date: item.start_date,
      }));
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败';
  } finally {
    loading.value = false;
  }
}

onMounted(() => { if (wsId.value) void load(); });
watch([wsId, projectId], () => { if (wsId.value) void load(); });

function setView(mode: ViewMode) {
  viewMode.value = mode;
}

function prev() {
  switch (viewMode.value) {
    case 'month': currentDate.value = currentDate.value.subtract(1, 'month'); break;
    case 'week': currentDate.value = currentDate.value.subtract(1, 'week'); break;
    case 'day': currentDate.value = currentDate.value.subtract(1, 'day'); break;
  }
}

function next() {
  switch (viewMode.value) {
    case 'month': currentDate.value = currentDate.value.add(1, 'month'); break;
    case 'week': currentDate.value = currentDate.value.add(1, 'week'); break;
    case 'day': currentDate.value = currentDate.value.add(1, 'day'); break;
  }
}

function goToday() {
  currentDate.value = dayjs();
}

function goToIssue(issueId: number) {
  void router.push(`/${workspaceSlug.value}/projects/${projectId.value}/issues/${issueId}`);
}

/** 视图标题 */
const viewTitle = computed(() => {
  switch (viewMode.value) {
    case 'month': return currentDate.value.format('YYYY 年 MM 月');
    case 'week': return `${viewRange.value.start.format('MM/DD')} - ${viewRange.value.end.format('MM/DD')}`;
    case 'day': return currentDate.value.format('YYYY 年 MM 月 DD 日');
  }
});

function typeIcon(type: string): string {
  switch (type) {
    case 'requirement': return 'RQ';
    case 'task': return 'TK';
    case 'defect': return 'BG';
    default: return '--';
  }
}

function typeColor(type: string): string {
  switch (type) {
    case 'requirement': return '#2563eb';
    case 'task': return '#d97706';
    case 'defect': return '#dc2626';
    default: return '#6b7280';
  }
}
</script>

<template>
  <div class="calendar-view">
    <!-- 顶部栏 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">日历</h1>
        <div class="nav-controls">
          <button class="icon-btn" @click="prev" aria-label="上一个">‹</button>
          <span class="view-title">{{ viewTitle }}</span>
          <button class="icon-btn" @click="next" aria-label="下一个">›</button>
          <button class="btn-today" @click="goToday">今天</button>
        </div>
      </div>
      <div class="header-actions">
        <div class="view-tabs">
          <button
            v-for="v in VIEW_MODES"
            :key="v.value"
            :class="['view-tab', { active: viewMode === v.value }]"
            @click="setView(v.value)"
          >{{ v.label }}</button>
        </div>
      </div>
    </header>

    <!-- 加载中 -->
    <AppSkeleton v-if="loading" variant="dashboard" :rows="2" />
    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" :retry="load" />

    <!-- 月视图 -->
    <div v-else-if="viewMode === 'month'" class="calendar-month">
      <!-- 星期头 -->
      <div class="weekday-header">
        <span v-for="d in ['日', '一', '二', '三', '四', '五', '六']" :key="d" class="weekday-cell">{{ d }}</span>
      </div>
      <!-- 日期网格 -->
      <div class="month-grid">
        <div
          v-for="(day, idx) in calendarGrid"
          :key="idx"
          :class="['day-cell', { 'not-current-month': !day.isCurrentMonth, 'is-today': day.date.isSame(dayjs(), 'day') }]"
        >
          <span class="day-number">{{ day.date.date() }}</span>
          <div class="day-events">
            <div
              v-for="ev in day.events.slice(0, 3)"
              :key="ev.id"
              class="cal-event"
              :style="{ borderLeftColor: typeColor(ev.type_code) }"
              @click.stop="goToIssue(ev.id)"
            >
              <span class="event-type">{{ typeIcon(ev.type_code) }}</span>
              <span class="event-name">{{ ev.name }}</span>
            </div>
            <span v-if="day.events.length > 3" class="more-events">+{{ day.events.length - 3 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 周视图 -->
    <div v-else-if="viewMode === 'week'" class="calendar-week">
      <div class="week-header">
        <div
          v-for="day in weekDays"
          :key="day.date.format('YYYY-MM-DD')"
          :class="['week-day-header', { 'is-today': day.date.isSame(dayjs(), 'day') }]"
        >
          <span class="weekday-name">{{ day.date.format('dd') }}</span>
          <span class="weekday-date">{{ day.date.date() }}</span>
        </div>
      </div>
      <div class="week-body">
        <div
          v-for="day in weekDays"
          :key="day.date.format('YYYY-MM-DD')"
          class="week-day-col"
        >
          <div
            v-for="ev in day.events"
            :key="ev.id"
            class="cal-event cal-event--block"
            :style="{ borderLeftColor: typeColor(ev.type_code) }"
            @click="goToIssue(ev.id)"
          >
            <span class="event-identifier">{{ ev.identifier }}</span>
            <span class="event-name">{{ ev.name }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 日视图 -->
    <div v-else class="calendar-day">
      <AppEmptyState v-if="dayHours[0]?.events.length === 0" description="当日无工作项" />
      <div v-else class="day-events">
        <div
          v-for="ev in dayHours[0].events"
          :key="ev.id"
          class="cal-event cal-event--block"
          :style="{ borderLeftColor: typeColor(ev.type_code) }"
          @click="goToIssue(ev.id)"
        >
          <span class="event-identifier">{{ ev.identifier }}</span>
          <span class="event-name">{{ ev.name }}</span>
          <span class="event-state" :style="{ backgroundColor: ev.state.color }">{{ ev.state.name }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.calendar-view {
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
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: var(--text-primary);
}

.nav-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  border-radius: 6px;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  color: var(--text-primary);
}

.view-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  min-width: 140px;
  text-align: center;
}

.btn-today {
  padding: 4px 12px;
  font-size: 12px;
  border: 1px solid var(--border);
  background: var(--surface-1);
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-primary);
}

.view-tabs {
  display: flex;
  background: var(--surface-2);
  border-radius: 6px;
  padding: 2px;
}

.view-tab {
  padding: 5px 14px;
  font-size: 13px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
}

.view-tab.active {
  background: var(--surface-1);
  color: var(--text-primary);
  font-weight: 500;
}

/* --- 月视图 --- */
.calendar-month {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.weekday-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  height: 32px;
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
}

.weekday-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary);
}

.month-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
}

.day-cell {
  min-height: 110px;
  border-right: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
  padding: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.day-cell:nth-child(7n) { border-right: none; }

.not-current-month {
  background: var(--surface-2);
  opacity: 0.6;
}

.is-today .day-number {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-radius: 50%;
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.day-number {
  font-size: 12px;
  color: var(--text-secondary);
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2px;
}

.day-events {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  overflow: hidden;
}

.cal-event {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  font-size: 11px;
  background: var(--surface-2);
  border-left: 3px solid;
  border-radius: 3px;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: background 0.1s;
}

.cal-event:hover {
  background: var(--surface-3);
}

.cal-event--block {
  padding: 4px 8px;
  margin-bottom: 4px;
}

.event-type {
  font-family: var(--font-mono);
  font-weight: 600;
  font-size: 10px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.event-identifier {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.event-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.event-state {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  color: var(--text-on-brand);
  flex-shrink: 0;
  margin-left: auto;
}

.more-events {
  font-size: 10px;
  color: var(--text-tertiary);
  padding: 0 6px;
}

/* --- 周视图 --- */
.calendar-week {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.week-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  height: 48px;
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
}

.week-day-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-right: 1px solid var(--border-subtle);
}

.week-day-header.is-today .weekday-date {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-radius: 50%;
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.weekday-name {
  font-size: 11px;
  color: var(--text-tertiary);
}

.weekday-date {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.week-body {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  min-height: 500px;
}

.week-day-col {
  border-right: 1px solid var(--border-subtle);
  padding: 6px;
}

.week-day-col:nth-child(7n) { border-right: none; }

/* --- 日视图 --- */
.calendar-day {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 24px;
  min-height: 500px;
}

.day-events {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
