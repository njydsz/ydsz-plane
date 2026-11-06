<script setup lang="ts">
/**
 * CalendarView — 项目日历视图。
 *
 * 展示两类数据：
 *   1. 需求/任务/缺陷在日历网格上的分布（按 target_date 排布）；
 *   2. 独立日历事件（calendar_events 表：会议/里程碑/发布日/截止日）。
 * 支持月/周/日视图切换，独立事件可通过顶部按钮创建。
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import dayjs from "dayjs";

import { calendarApi, type CalendarEvent, type CalendarEventType } from "@/api/services/calendar";
import { issueApi } from "@/api/services/issue";
import { useWorkspaceStore } from "@/stores/workspace";
import { AppButton, AppEmptyState, AppErrorState, AppPageContainer, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();

interface IssueCalendarEvent {
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
const workspaceId = computed(() => Number(route.params.workspaceId));
const wsId = computed(() => wsStore.current?.id ?? 0);

const loading = ref(true);
const error = ref('');
const events = ref<IssueCalendarEvent[]>([]); // 当前视图范围内的需求/任务/缺陷
/** 独立日历事件（calendar_events 表）。 */
const calendarEvents = ref<CalendarEvent[]>([]);
const currentDate = ref(dayjs());

/** 新建事件弹窗状态。 */
const showCreateModal = ref(false);
const createForm = ref({
  title: '',
  description: '',
  date: '',
  startTime: '09:00',
  endTime: '10:00',
  location: '',
  eventType: 'meeting' as CalendarEventType,
  isAllDay: false,
});
const saving = ref(false);

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
    default:
      return { start: base.startOf('month'), end: base.endOf('month') };
  }
});

/** 合并后的日事件列表：工作项 + 独立日历事件。 */
interface DayEvent {
  id: string;
  kind: 'issue' | 'calendar';
  name: string;
  meta: string;
  color: string;
  issueId?: number;
  calendarId?: number;
}

/** 日历网格（月视图） */
const calendarGrid = computed(() => {
  const { start, end } = viewRange.value;
  const days: { date: dayjs.Dayjs; isCurrentMonth: boolean; events: DayEvent[] }[] = [];
  let cursor = start;
  while (cursor.isBefore(end) || cursor.isSame(end, 'day')) {
    const dayEvents = dayEventsFor(cursor);
    days.push({
      date: cursor,
      isCurrentMonth: cursor.month() === currentDate.value.month(),
      events: dayEvents,
    });
    cursor = cursor.add(1, 'day');
  }
  return days;
});

/** 计算某一天的合并事件列表。 */
function dayEventsFor(day: dayjs.Dayjs): DayEvent[] {
  const out: DayEvent[] = [];
  for (const e of events.value) {
    if (!e.target_date) continue;
    if (dayjs(e.target_date).isSame(day, 'day')) {
      out.push({
        id: `issue-${e.id}`,
        kind: 'issue',
        name: e.name,
        meta: e.identifier,
        color: typeColor(e.type_code),
        issueId: e.id,
      });
    }
  }
  for (const e of calendarEvents.value) {
    const start = dayjs(e.start_time);
    const end = dayjs(e.end_time);
    // 事件与当天有交集即展示。
    if (start.isSame(day, 'day') || end.isSame(day, 'day') ||
        (start.isBefore(day, 'day') && end.isAfter(day, 'day'))) {
      out.push({
        id: `cal-${e.id}`,
        kind: 'calendar',
        name: e.title,
        meta: e.event_type,
        color: calendarColor(e.event_type),
        calendarId: e.id,
      });
    }
  }
  return out;
}

/** 周视图行 */
const weekDays = computed(() => {
  const { start } = viewRange.value;
  const days: { date: dayjs.Dayjs; events: DayEvent[] }[] = [];
  for (let i = 0; i < 7; i++) {
    const d = start.add(i, 'day');
    days.push({ date: d, events: dayEventsFor(d) });
  }
  return days;
});

/** 日视图事件列表（工作项 + 独立日历事件）。 */
const dayViewEvents = computed(() => dayEventsFor(currentDate.value));

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = '';
  try {
    // 扩大范围加载（前后各一个月确保覆盖视图）
    const start = currentDate.value.subtract(2, 'month').format('YYYY-MM-DD');
    const end = currentDate.value.add(2, 'month').format('YYYY-MM-DD');

    // 并行拉取工作项 + 独立日历事件。
    const [result, calEvents] = await Promise.all([
      issueApi.listIssues(wsId.value, projectId.value, {
        start_date_from: start,
        target_date_to: end,
        limit: 500,
      }),
      calendarApi.list(wsId.value, {
        start: `${start}T00:00:00Z`,
        end: `${end}T23:59:59Z`,
        limit: 200,
      }).catch(() => [] as CalendarEvent[]),
    ]);

    events.value = result.results
      .filter((item) => item.target_date) // 日历视图仅展示有目标日期的需求/任务/缺陷
      .map((item) => ({
        id: item.id,
        identifier: item.identifier,
        name: item.name,
        type_code: item.type_code,
        state: item.state ?? { name: '', group: 'backlog', color: 'var(--text-secondary)' },
        priority: item.priority,
        target_date: item.target_date,
        start_date: item.start_date,
      }));
    calendarEvents.value = calEvents;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败';
  } finally {
    loading.value = false;
  }
}

/** 打开新建事件弹窗，默认填充当前视图日期。 */
function openCreateModal() {
  createForm.value.date = currentDate.value.format('YYYY-MM-DD');
  showCreateModal.value = true;
}

/** 提交新建事件。 */
async function submitCreate() {
  const f = createForm.value;
  if (!f.title.trim()) {
    toast.error('请填写事件标题');
    return;
  }
  if (!f.date) {
    toast.error('请选择日期');
    return;
  }

  const start = f.isAllDay
    ? dayjs(`${f.date}T00:00:00`).toISOString()
    : dayjs(`${f.date}T${f.startTime}:00`).toISOString();
  const end = f.isAllDay
    ? dayjs(`${f.date}T23:59:59`).toISOString()
    : dayjs(`${f.date}T${f.endTime}:00`).toISOString();

  saving.value = true;
  try {
    await calendarApi.create(wsId.value, {
      title: f.title,
      description: f.description || undefined,
      start_time: start,
      end_time: end,
      is_all_day: f.isAllDay,
      location: f.location || undefined,
      event_type: f.eventType,
    });
    showCreateModal.value = false;
    createForm.value = {
      title: '',
      description: '',
      date: '',
      startTime: '09:00',
      endTime: '10:00',
      location: '',
      eventType: 'meeting',
      isAllDay: false,
    };
    toast.success('事件已创建');
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : '创建失败');
  } finally {
    saving.value = false;
  }
}

/** 归档独立日历事件。 */
async function archiveEvent(eventId: number) {
  if (!await confirm({ title: "归档事件", message: "确定归档该事件？", confirmText: "归档", variant: "danger" })) return;
  try {
    await calendarApi.archive(wsId.value, eventId);
    calendarEvents.value = calendarEvents.value.filter((e) => e.id !== eventId);
    toast.success('事件已归档');
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : '归档失败');
  }
}

/** 点击事件：工作项跳转详情；独立事件弹出归档确认。 */
function handleEventClick(ev: DayEvent) {
  if (ev.kind === 'issue' && ev.issueId) {
    goToIssue(ev.issueId);
    return;
  }
  if (ev.kind === 'calendar' && ev.calendarId) {
    void archiveEvent(ev.calendarId);
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
  void router.push(`/${workspaceId.value}/projects/${projectId.value}/issues/${issueId}`);
}

/** 视图标题 */
const viewTitle = computed(() => {
  switch (viewMode.value) {
    case 'month': return currentDate.value.format('YYYY 年 MM 月');
    case 'week': return `${viewRange.value.start.format('MM/DD')} - ${viewRange.value.end.format('MM/DD')}`;
    case 'day': return currentDate.value.format('YYYY 年 MM 月 DD 日');
    default: return currentDate.value.format('YYYY 年 MM 月');
  }
});

function typeIconFromMeta(meta: string): string {
  // meta 为工作项 identifier（如 REQ-123 / TSK-456 / BUG-789）。
  const prefix = meta.split('-')[0] ?? '';
  return prefix || '--';
}

function typeColor(type: string): string {
  switch (type) {
    case 'requirement': return 'var(--brand-600)';
    case 'task': return 'var(--warning-500)';
    case 'defect': return 'var(--danger-500)';
    default: return 'var(--text-secondary)';
  }
}

/** 独立日历事件的颜色映射。 */
function calendarColor(type: CalendarEventType): string {
  switch (type) {
    case 'meeting': return 'var(--brand-500)';
    case 'milestone': return 'var(--info-500)';
    case 'release': return 'var(--success-500)';
    case 'deadline': return 'var(--warning-500)';
    case 'other': return 'var(--text-secondary)';
    default: return 'var(--text-secondary)';
  }
}

/** 事件类型中文标签。 */
const EVENT_TYPE_LABELS: Record<CalendarEventType, string> = {
  meeting: '会议',
  milestone: '里程碑',
  release: '发布',
  deadline: '截止',
  other: '其他',
};
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="calendar-view">
    <!-- 顶部栏 -->
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">日历</h1>
        <div class="nav-controls">
          <button class="icon-btn" aria-label="上一个" @click="prev">‹</button>
          <span class="view-title">{{ viewTitle }}</span>
          <button class="icon-btn" aria-label="下一个" @click="next">›</button>
          <AppButton size="sm" variant="secondary" @click="goToday">今天</AppButton>
        </div>
      </div>
      <div class="header-actions">
        <AppButton @click="openCreateModal">+ 新建事件</AppButton>
        <div class="view-tabs">
          <button
            v-for="v in VIEW_MODES"
            :key="v.value"
            :class="['view-tab', { active: viewMode === v.value }]"
            @click="setView(v.value)"
          >
{{ v.label }}
</button>
        </div>
      </div>
    </header>

    <!-- 加载中 -->
    <AppSkeleton v-if="loading" variant="dashboard" :rows="2" />
    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

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
              :style="{ borderLeftColor: ev.color }"
              @click.stop="handleEventClick(ev)"
            >
              <span class="event-type">{{ ev.kind === 'calendar' ? 'CAL' : typeIconFromMeta(ev.meta) }}</span>
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
            :style="{ borderLeftColor: ev.color }"
            @click="handleEventClick(ev)"
          >
            <span class="event-identifier">{{ ev.meta }}</span>
            <span class="event-name">{{ ev.name }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 日视图 -->
    <div v-else class="calendar-day">
      <AppEmptyState v-if="dayViewEvents.length === 0" description="当日无需求/任务/缺陷或日历事件" />
      <div v-else class="day-events">
        <div
          v-for="ev in dayViewEvents"
          :key="ev.id"
          class="cal-event cal-event--block"
          :style="{ borderLeftColor: ev.color }"
          @click="handleEventClick(ev)"
        >
          <span class="event-identifier">{{ ev.meta }}</span>
          <span class="event-name">{{ ev.name }}</span>
        </div>
      </div>
    </div>

    <!-- 新建事件弹窗 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-card">
        <h2 class="modal-title">新建日历事件</h2>
        <div class="modal-body">
          <label class="form-field">
            <span>标题 *</span>
            <input v-model="createForm.title" type="text" placeholder="如：需求评审会" />
          </label>
          <label class="form-field">
            <span>类型</span>
            <select v-model="createForm.eventType">
              <option v-for="(label, key) in EVENT_TYPE_LABELS" :key="key" :value="key">{{ label }}</option>
            </select>
          </label>
          <label class="form-field">
            <span>日期 *</span>
            <input v-model="createForm.date" type="date" />
          </label>
          <label v-if="!createForm.isAllDay" class="form-field form-field--inline">
            <span>开始时间</span>
            <input v-model="createForm.startTime" type="time" />
            <span>结束时间</span>
            <input v-model="createForm.endTime" type="time" />
          </label>
          <label class="form-field form-field--checkbox">
            <input v-model="createForm.isAllDay" type="checkbox" />
            <span>全天事件</span>
          </label>
          <label class="form-field">
            <span>地点</span>
            <input v-model="createForm.location" type="text" placeholder="可选" />
          </label>
          <label class="form-field">
            <span>描述</span>
            <textarea v-model="createForm.description" rows="2" placeholder="可选"></textarea>
          </label>
        </div>
        <div class="modal-footer">
          <AppButton variant="secondary" @click="showCreateModal = false">取消</AppButton>
          <AppButton :loading="saving" :disabled="saving" @click="submitCreate">
            {{ saving ? '创建中…' : '创建' }}
          </AppButton>
        </div>
      </div>
    </div>
  
    </div>
  </AppPageContainer>
</template>

<style scoped>
.calendar-view {
  
  
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

/* --- 弹窗 --- */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 24px;
}

.modal-card {
  background: var(--surface-1);
  border: 1px solid var(--border);
  border-radius: 10px;
  width: 100%;
  max-width: 440px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.modal-title {
  margin: 0;
  padding: 16px 20px;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.modal-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 13px;
  color: var(--text-secondary);
}

.form-field input,
.form-field select,
.form-field textarea {
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--surface-1);
  color: var(--text-primary);
  font-family: inherit;
}

.form-field--inline {
  flex-direction: row;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.form-field--inline input {
  flex: 1;
  min-width: 100px;
}

.form-field--checkbox {
  flex-direction: row;
  align-items: center;
  gap: 6px;
}

.form-field--checkbox input {
  width: 14px;
  height: 14px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--border-subtle);
}
</style>
