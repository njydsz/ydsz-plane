<script setup lang="ts">
/**
 * DashboardView - 项目仪表盘主页面。
 *
 * CSS Grid 12 列响应式布局 + 顶部栏 + Widget 网格 + 浮动操作按钮。
 * 数据由后端 overview 接口一次性返回（widgets + snapshots + alerts）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { dashboardApi } from "@/api/services/dashboard";
import type { DashboardData, RiskAlert, WidgetType } from "@/api/services/dashboard";
import { useWorkspaceStore } from "@/stores/workspace";
import AppModal from "@/components/AppModal.vue";
import { AppErrorState, AppSkeleton } from "@/components";
import { WIDGET_COMPONENTS, WIDGET_NAME_MAP } from "@/components/dashboard/widgetRegistry";
import DashWidgetCard from "@/components/dashboard/DashWidgetCard.vue";
import EmptyStateWidget from "@/components/dashboard/EmptyStateWidget.vue";
import { toast } from "@/lib/toast";

const route = useRoute();
const wsStore = useWorkspaceStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = computed(() => wsStore.current?.id ?? 0);

const loading = ref(true);
const error = ref("");
const dashboardData = ref<DashboardData | null>(null);
const isFullscreen = ref(false);

// --- 拖拽重排 ---
const draggingWidgetId = ref<number | null>(null);
const dragOverWidgetId = ref<number | null>(null);
let saveTimer: ReturnType<typeof setTimeout> | null = null;
const savingIds = ref<Set<number>>(new Set());

function onDragStart(e: DragEvent, widgetId: number) {
  draggingWidgetId.value = widgetId;
  // 需要一个 data 才能触发 drag/drop 在部分浏览器上
  e.dataTransfer?.setData("text/plain", String(widgetId));
  e.dataTransfer!.effectAllowed = "move";
}

function onDragOver(e: DragEvent, widgetId: number) {
  e.preventDefault(); // 允许 drop
  if (draggingWidgetId.value === null || draggingWidgetId.value === widgetId) return;
  dragOverWidgetId.value = widgetId;
  e.dataTransfer!.dropEffect = "move";
}

function onDragLeave(_e: DragEvent, widgetId: number) {
  if (dragOverWidgetId.value === widgetId) {
    dragOverWidgetId.value = null;
  }
}

function onDrop(e: DragEvent, targetId: number) {
  e.preventDefault();
  const sourceId = draggingWidgetId.value;
  draggingWidgetId.value = null;
  dragOverWidgetId.value = null;

  if (sourceId === null || sourceId === targetId) return;

  const widgets = dashboardData.value?.widgets;
  if (!widgets) return;

  const source = widgets.find((w) => w.id === sourceId);
  const target = widgets.find((w) => w.id === targetId);
  if (!source || !target) return;

  // 根据拖拽释放位置（相对 target 左/右半区）决定插入到 target 左侧或右侧
  const targetEl = (e.currentTarget as HTMLElement);
  const rect = targetEl.getBoundingClientRect();
  const isLeftHalf = (e.clientX - rect.left) < rect.width / 2;

  // 计算目标 grid 坐标
  let newX: number;
  let newY: number;
  if (isLeftHalf) {
    // 放置到 target 左侧
    newX = Math.max(0, target.grid_x - source.grid_w);
    newY = target.grid_y;
    // 如果左侧空间不足，尝试贴到 target 左边缘并下移一行
    if (target.grid_x - source.grid_w < 0) {
      newX = 0;
      newY = target.grid_y + target.grid_h;
    }
  } else {
    // 放置到 target 右侧
    newX = target.grid_x + target.grid_w;
    newY = target.grid_y;
    // 如果超出 12 列，换行
    if (newX + source.grid_w > 12) {
      newX = 0;
      newY = target.grid_y + target.grid_h;
    }
  }

  // 本地更新
  source.grid_x = newX;
  source.grid_y = newY;

  // 触发响应式更新
  dashboardData.value = { ...dashboardData.value };

  // 防抖保存到后端
  scheduleSave(sourceId, { grid_x: newX, grid_y: newY, grid_w: source.grid_w, grid_h: source.grid_h });
}

function onDragEnd() {
  draggingWidgetId.value = null;
  dragOverWidgetId.value = null;
}

/** 防抖保存：300ms 内多次操作只发一次请求 */
function scheduleSave(widgetId: number, payload: { grid_x: number; grid_y: number; grid_w: number; grid_h: number }) {
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    void doSave(widgetId, payload);
  }, 300);
}

async function doSave(widgetId: number, payload: { grid_x: number; grid_y: number; grid_w: number; grid_h: number }) {
  if (!wsId.value) return;
  savingIds.value.add(widgetId);
  savingIds.value = new Set(savingIds.value); // 触发响应式
  try {
    await dashboardApi.updateWidget(wsId.value, projectId.value, widgetId, payload);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "布局保存失败");
    // 失败时回滚：重新加载
    await load();
  } finally {
    savingIds.value.delete(widgetId);
    savingIds.value = new Set(savingIds.value);
  }
}

// --- 全屏模式 ---
function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value;
  if (isFullscreen.value) {
    document.documentElement.requestFullscreen?.();
  } else {
    document.exitFullscreen?.();
  }
}
document.addEventListener("fullscreenchange", () => {
  if (!document.fullscreenElement) isFullscreen.value = false;
});

// --- 时间范围（localStorage 持久化） ---
const TIME_RANGES = [
  { value: "7d", label: "7 天" },
  { value: "30d", label: "30 天" },
  { value: "90d", label: "90 天" },
  { value: "all", label: "全部" },
];
const STORAGE_KEY = "dashboard_time_range";
function loadTimeRange(): string {
  return localStorage.getItem(STORAGE_KEY) ?? "30d";
}
const timeRange = ref(loadTimeRange());
function setTimeRange(v: string) {
  timeRange.value = v;
  localStorage.setItem(STORAGE_KEY, v);
  void load();
}

// --- Widget 类型选择 modal ---
const showAddModal = ref(false);
const selectedWidgetType = ref<WidgetType | "">("");
const customTitle = ref("");

// --- 模板加载 ---
const templatesLoading = ref(false);
const templatesDropdownOpen = ref(false);

const visibleWidgets = computed(() =>
  (dashboardData.value?.widgets ?? []).filter((w) => w.is_visible).sort((a, b) => a.sort_order - b.sort_order),
);

async function load() {
  if (!wsId.value) return;
  loading.value = true;
  error.value = "";
  try {
    dashboardData.value = await dashboardApi.getOverview(wsId.value, projectId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function getWidgetComponent(type: WidgetType) {
  return WIDGET_COMPONENTS[type] ?? null;
}

function gridStyle(widget: { grid_x: number; grid_y: number; grid_w: number; grid_h: number }) {
  return {
    gridColumn: `${widget.grid_x + 1} / span ${widget.grid_w}`,
    gridRow: `${widget.grid_y + 1} / span ${widget.grid_h}`,
  };
}

async function handleAddWidget() {
  if (!selectedWidgetType.value || !wsId.value) return;
  const widgetType = selectedWidgetType.value;
  const title = customTitle.value.trim() || WIDGET_NAME_MAP[widgetType] || widgetType;
  try {
    await dashboardApi.createWidget(wsId.value, projectId.value, {
      widget_type: widgetType,
      title,
      grid_x: 0,
      grid_y: 999,
      grid_w: 6,
      grid_h: 2,
    });
    showAddModal.value = false;
    selectedWidgetType.value = "";
    customTitle.value = "";
    toast.success("Widget 已添加");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "添加失败");
  }
}

async function handleRemove(widgetId: number) {
  if (!wsId.value) return;
  if (!confirm("确定要删除这个 widget 吗？")) return;
  try {
    await dashboardApi.deleteWidget(wsId.value, projectId.value, widgetId);
    toast.success("Widget 已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

async function loadTemplates() {
  if (!wsId.value) return;
  templatesLoading.value = true;
  try {
    const templates = await dashboardApi.listTemplates(wsId.value, projectId.value);
    if (templates.length > 0) {
      applyTemplate(templates[0]);
    } else {
      toast.error("暂无可用模板");
    }
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "加载模板失败");
  } finally {
    templatesLoading.value = false;
    templatesDropdownOpen.value = false;
  }
}

function applyTemplate(template: { name?: string; layout: Record<string, any> }) {
  const layoutWidgets = template.layout?.widgets;
  if (!Array.isArray(layoutWidgets)) return;
  toast.success(`已应用模板: ${template.name ?? "未命名模板"}`);
}

async function handleResolve(alertId: number) {
  if (!wsId.value) return;
  try {
    await dashboardApi.resolveAlert(wsId.value, projectId.value, alertId);
    toast.success("告警已处理");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "处理失败");
  }
}

const alerts = computed<RiskAlert[]>(() => dashboardData.value?.alerts ?? []);
const hasWidgets = computed(() => visibleWidgets.value.length > 0);

onMounted(load);
</script>

<template>
  <div class="dashboard" :class="{ 'dashboard--fullscreen': isFullscreen }">
    <!-- ===== 顶部栏 ===== -->
    <header class="dashboard__header">
      <div class="dashboard__header-left">
        <h1 class="dashboard__title">项目仪表盘</h1>
        <div class="dashboard__time-range">
          <button
            v-for="opt in TIME_RANGES"
            :key="opt.value"
            class="time-btn"
            :class="{ 'time-btn--active': timeRange === opt.value }"
            @click="setTimeRange(opt.value)"
          >
            {{ opt.label }}
          </button>
        </div>
      </div>
      <div class="dashboard__header-right">
        <button class="action-btn action-btn--ghost" @click="toggleFullscreen">
          {{ isFullscreen ? "⤓ 退出全屏" : "⤢ 全屏" }}
        </button>
        <button class="action-btn" @click="showAddModal = true">+ 添加 Widget</button>
        <div class="template-dropdown">
          <button class="action-btn action-btn--ghost" @click="templatesDropdownOpen = !templatesDropdownOpen">
            布局模板
          </button>
          <div v-if="templatesDropdownOpen" class="template-dropdown__menu">
            <button
              class="template-dropdown__item"
              :disabled="templatesLoading"
              @click="loadTemplates"
            >
              {{ templatesLoading ? "加载中..." : "加载推荐模板" }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- 加载 / 错误 -->
    <AppSkeleton v-if="loading" variant="dashboard" :rows="4" />
    <AppErrorState
      v-else-if="error"
      :message="error"
      @retry="load"
    />

    <!-- ===== 保存状态指示器 ===== -->
    <transition name="fade">
      <span v-if="savingIds.size > 0" class="dashboard__save-indicator">
        保存中…
      </span>
    </transition>

    <!-- ===== Widget 网格 ===== -->
    <div
      v-else-if="hasWidgets"
      class="dashboard__grid"
      :class="{ 'dashboard__grid--dragging': draggingWidgetId !== null }"
    >
      <div
        v-for="w in visibleWidgets"
        :key="w.id"
        class="grid-cell"
        :class="{
          'grid-cell--dragging': draggingWidgetId === w.id,
          'grid-cell--drag-over': dragOverWidgetId === w.id,
          'grid-cell--saving': savingIds.has(w.id),
        }"
        :style="gridStyle(w)"
        draggable="true"
        @dragstart="onDragStart($event, w.id)"
        @dragover="onDragOver($event, w.id)"
        @dragleave="onDragLeave($event, w.id)"
        @drop="onDrop($event, w.id)"
        @dragend="onDragEnd"
      >
        <DashWidgetCard :title="w.title" :is-saving="savingIds.has(w.id)" @remove="handleRemove(w.id)">
          <component
            :is="getWidgetComponent(w.widget_type)"
            v-if="getWidgetComponent(w.widget_type)"
            :data="dashboardData!.snapshots[w.widget_type]"
            :config="w.config"
            :ws-id="wsId"
            :project-id="projectId"
            :alerts="w.widget_type === 'risk_alert' ? alerts : undefined"
            @resolve="handleResolve"
          />
        </DashWidgetCard>
      </div>
    </div>

    <!-- ===== 空态 ===== -->
    <EmptyStateWidget v-else class="dashboard__empty" @add="showAddModal = true" />

    <!-- ===== 浮动操作按钮 ===== -->
    <button class="fab" title="快速添加 Widget" @click="showAddModal = true">
      +
    </button>

    <!-- ===== 添加 Widget Modal ===== -->
    <AppModal :visible="showAddModal" title="添加 Widget" @close="showAddModal = false">
      <div class="add-widget-modal">
        <p class="add-widget-modal__label">选择 Widget 类型</p>
        <div class="add-widget-modal__grid">
          <button
            v-for="(label, type) in WIDGET_NAME_MAP"
            :key="type"
            class="widget-type-card"
            :class="{ 'widget-type-card--active': selectedWidgetType === type }"
            @click="selectedWidgetType = type"
          >
            <span class="widget-type-card__label">{{ label }}</span>
          </button>
        </div>
        <label class="add-widget-modal__custom-title">
          <span class="add-widget-modal__label">自定义标题（可选）</span>
          <input
            v-model="customTitle"
            type="text"
            class="add-widget-modal__input"
            placeholder="留空则使用默认名称"
          />
        </label>
      </div>
      <template #footer>
        <button class="action-btn action-btn--ghost" @click="showAddModal = false">取消</button>
        <button
          class="action-btn"
          :disabled="!selectedWidgetType"
          @click="handleAddWidget"
        >
          确认添加
        </button>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
  padding-bottom: 80px;
}

/* ===== 全屏模式 ===== */
.dashboard--fullscreen {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: var(--surface-1, #fff);
  padding: 16px 24px 80px;
  overflow-y: auto;
}

/* ===== Top bar ===== */
.dashboard__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.dashboard__header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.dashboard__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0;
}

.dashboard__time-range {
  display: flex;
  border: 1px solid var(--border-default, #dfe2e6);
  border-radius: var(--radius-sm, 6px);
  overflow: hidden;
}

.time-btn {
  padding: 5px 12px;
  font-size: 12px;
  font-family: inherit;
  background: var(--surface-1, #fff);
  border: none;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.time-btn:hover {
  background: var(--surface-2, #f7f8f9);
  color: var(--text-primary, #1f2937);
}

.time-btn--active {
  background: var(--brand-500, #3f63f1);
  color: var(--text-on-brand);
}

.dashboard__header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  padding: 7px 14px;
  font-size: 13px;
  font-family: inherit;
  background: var(--brand-500, #3f63f1);
  color: var(--text-on-brand);
  border: 1px solid var(--brand-500, #3f63f1);
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: background 0.15s;
  white-space: nowrap;
}

.action-btn:hover {
  background: var(--brand-600, #2f4fd0);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn--ghost {
  background: var(--surface-1, #fff);
  color: var(--text-secondary, #4b5563);
  border-color: var(--border-default, #dfe2e6);
}

.action-btn--ghost:hover {
  background: var(--surface-2, #f7f8f9);
}

/* Template dropdown */
.template-dropdown {
  position: relative;
}

.template-dropdown__menu {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-default, #dfe2e6);
  border-radius: var(--radius-sm, 6px);
  box-shadow: var(--shadow-popover);
  z-index: 50;
  min-width: 160px;
  overflow: hidden;
}

.template-dropdown__item {
  display: block;
  width: 100%;
  padding: 8px 12px;
  font-size: 12px;
  font-family: inherit;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  color: var(--text-secondary, #4b5563);
}

.template-dropdown__item:hover:not(:disabled) {
  background: var(--surface-2, #f7f8f9);
}

.template-dropdown__item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== 保存状态指示器 ===== */
.dashboard__save-indicator {
  position: absolute;
  top: 8px;
  right: 24px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--brand-500, #3f63f1);
  background: var(--surface-1, #fff);
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  padding: 5px 12px;
  z-index: 60;
  box-shadow: var(--shadow-popover, 0 2px 8px rgba(0, 0, 0, 0.08));
}

.dashboard__save-indicator::before {
  content: "";
  width: 10px;
  height: 10px;
  border: 2px solid var(--brand-500, #3f63f1);
  border-top-color: transparent;
  border-radius: 50%;
  animation: dashboard-spin 0.7s linear infinite;
}

@keyframes dashboard-spin {
  to {
    transform: rotate(360deg);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* ===== Grid ===== */
.dashboard__grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 12px;
  align-items: stretch;
  position: relative;
}

.dashboard__grid--dragging {
  /* 拖拽期间给网格区域加一个淡色背景作为可放置区域提示 */
  background-image: linear-gradient(
    to right,
    transparent calc(100% / 12 - 1px),
    var(--border-subtle, #e5e7eb) calc(100% / 12 - 1px),
    var(--border-subtle, #e5e7eb) calc(100% / 12),
    transparent calc(100% / 12)
  );
  background-size: calc(100% + 12px) 100%;
  background-position: -6px 0;
}

.grid-cell {
  min-height: 120px;
  transition: transform 0.15s, box-shadow 0.15s, opacity 0.15s, outline-offset 0.15s;
  outline-offset: 0;
}

.grid-cell--dragging {
  opacity: 0.4;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  transform: scale(0.97);
  cursor: grabbing;
}

.grid-cell--drag-over {
  outline: 2px dashed var(--brand-500, #3f63f1);
  outline-offset: -2px;
  background-color: color-mix(in srgb, var(--brand-500, #3f63f1) 6%, transparent);
  border-radius: var(--radius-md, 8px);
}

.grid-cell--saving {
  opacity: 0.7;
}

/* ===== Empty state ===== */
.dashboard__empty {
  border: 1px dashed var(--border-default, #dfe2e6);
  border-radius: var(--radius-lg, 12px);
}

/* ===== FAB ===== */
.fab {
  position: fixed;
  right: 32px;
  bottom: 32px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--brand-500, #3f63f1);
  color: var(--text-on-brand);
  border: none;
  font-size: 24px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(63, 99, 241, 0.4);
  transition: transform 0.15s, box-shadow 0.15s;
  z-index: 90;
}

.fab:hover {
  transform: scale(1.08);
  box-shadow: 0 6px 16px rgba(63, 99, 241, 0.5);
}

/* ===== Add widget modal ===== */
.add-widget-modal {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.add-widget-modal__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #4b5563);
  margin: 0;
}

.add-widget-modal__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.widget-type-card {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-1, #fff);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.widget-type-card:hover {
  border-color: var(--brand-500, #3f63f1);
  background: var(--brand-50, #eef2fe);
}

.widget-type-card--active {
  border-color: var(--brand-500, #3f63f1);
  background: var(--brand-50, #eef2fe);
  box-shadow: inset 0 0 0 2px var(--brand-500, #3f63f1);
}

.widget-type-card__label {
  font-size: 13px;
  color: var(--text-primary, #1f2937);
}

.add-widget-modal__custom-title {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.add-widget-modal__input {
  padding: 8px 10px;
  border: 1px solid var(--border-default, #dfe2e6);
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary, #1f2937);
  background: var(--surface-1, #fff);
  outline: none;
}

.add-widget-modal__input:focus {
  border-color: var(--brand-500, #3f63f1);
  box-shadow: 0 0 0 3px var(--brand-50, #eef2fe);
}

/* ===== 响应式 ===== */
@media (max-width: 1200px) {
  .dashboard__grid {
    grid-template-columns: repeat(8, 1fr);
  }
}

@media (max-width: 768px) {
  .dashboard__grid {
    grid-template-columns: repeat(4, 1fr);
  }
  .dashboard__header {
    flex-direction: column;
    align-items: flex-start;
  }
  .add-widget-modal__grid {
    grid-template-columns: 1fr;
  }
}
</style>
