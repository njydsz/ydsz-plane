<script setup lang="ts">
/**
 * IssueFilter — 需求/任务/缺陷通用过滤器组件（看板/列表/日历/甘特共用）。
 *
 * 对标 Plane 的 IssueFilter：
 *   - 通过 viewType 参数适配不同视图
 *   - 使用 FilterState 强类型状态替代旧散落 ref
 *   - 使用 FilterAdapter 进行状态序列化/反序列化
 *   - 统一走 lib/prefs.ts 的 localStorage（替代旧的双重存储）
 *   - 服务端偏好按 view_type 自动选择
 */
import { onMounted, ref, watch } from "vue";

import type { IssueType, StateGroup } from "@/api/services/issue";
import { preferenceApi } from "@/api/services/preference";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import {
  type FilterState,
  type IssueTypeCode,
  safeParseFilters,
  hasActiveFilter,
  activeFilterCount,
} from "@/lib/filter-adapter";
import { prefs } from "@/lib/prefs";

// ---- Props ----
const props = defineProps<{
  projectId: number;
  workspaceId: number;
  /** 当前视图类型，决定了服务端偏好保存的 view_type 键 */
  viewType?: "kanban" | "list" | "calendar" | "gantt" | "spreadsheet";
}>();

const emit = defineEmits<{
  (e: "filter-change", filters: FilterState): void;
}>();

const viewType = props.viewType ?? "list";

// ---- 过滤状态（强类型 FilterState） ----
const filters = ref<FilterState>({});

// ---- 选项 ----
const stateGroupOptions: { value: StateGroup | ""; label: string }[] = [
  { value: "", label: "全部状态" },
  { value: "backlog", label: "待办" },
  { value: "started", label: "进行中" },
  { value: "completed", label: "已完成" },
  { value: "cancelled", label: "已取消" },
];

const typeOptions: { value: IssueType | ""; label: string }[] = [
  { value: "", label: "全部类型" },
  { value: "epic", label: "史诗" },
  { value: "requirement", label: "需求" },
  { value: "task", label: "任务" },
  { value: "defect", label: "缺陷" },
];

const priorityOptions: { value: string; label: string }[] = [
  { value: "", label: "全部优先级" },
  { value: "urgent", label: "紧急" },
  { value: "high", label: "高" },
  { value: "medium", label: "中" },
  { value: "low", label: "低" },
  { value: "none", label: "无" },
];

const severityOptions: { value: number | null; label: string }[] = [
  { value: null, label: "全部严重级别" },
  { value: 5, label: "≥S5 致命" },
  { value: 4, label: "≥S4 严重" },
  { value: 3, label: "≥S3 一般" },
  { value: 2, label: "≥S2 轻微" },
  { value: 1, label: "≥S1 建议" },
];

// ---- 派生辅助 ----
function hasActive(): boolean {
  return hasActiveFilter(filters.value);
}
function activeCount(): number {
  return activeFilterCount(filters.value);
}

// ---- 发射过滤变化 ----
function emitFilter() {
  // 清理空值，构建干净的 FilterState
  const cleaned: FilterState = {};
  for (const [k, v] of Object.entries(filters.value)) {
    if (v !== undefined && v !== null && v !== "" && !(typeof v === "number" && isNaN(v))) {
      (cleaned as Record<string, unknown>)[k] = v;
    }
  }
  filters.value = cleaned;
  emit("filter-change", cleaned);
}

// ---- localStorage 持久化（统一走 prefs.ts） ----
function storageKey(): string {
  return `filter:${viewType}:${props.projectId}`;
}
function saveToStorage(): void {
  prefs.set(storageKey(), filters.value);
}
function loadFromStorage(): FilterState {
  return prefs.get<FilterState>(storageKey(), {});
}

// ---- 服务端视图偏好 ----
const { wsId } = useWorkspaceContext();

async function saveServerPreference(): Promise<void> {
  const wsIdVal = wsId.value;
  if (!wsIdVal || !props.projectId) return;
  try {
    await preferenceApi.save(wsIdVal, props.projectId, viewType, {
      filters: filters.value,
    });
  } catch {
    /* 静默失败，不影响本地体验 */
  }
}

async function loadServerPreference(): Promise<FilterState> {
  const wsIdVal = wsId.value;
  if (!wsIdVal || !props.projectId) return {};
  try {
    const vp = await preferenceApi.get(wsIdVal, props.projectId, viewType);
    if (vp?.filters && typeof vp.filters === "object" && !Array.isArray(vp.filters)) {
      return safeParseFilters(vp.filters);
    }
  } catch {
    /* ignore */
  }
  return {};
}

// ---- 公开操作 ----
function clearFilters(): void {
  filters.value = {};
  emitFilter();
  saveToServerAndStorage();
}

function saveToServerAndStorage(): void {
  saveToStorage();
  void saveServerPreference();
}

// ---- 各字段 setter ----
function setSearch(v: string): void {
  filters.value = { ...filters.value, search: v || undefined };
}
function setGroup(v: StateGroup | ""): void {
  filters.value = { ...filters.value, group: v || undefined };
}
function setType(v: IssueType | ""): void {
  filters.value = { ...filters.value, type: (v || undefined) as IssueTypeCode | undefined };
}
function setPriority(v: string): void {
  filters.value = { ...filters.value, priority: (v || undefined) as FilterState["priority"] };
}
function setSeverity(v: number | null): void {
  filters.value = { ...filters.value, severity_from: v ?? undefined };
}

// ---- 防抖搜索 ----
let searchTimer: ReturnType<typeof setTimeout> | null = null;

function onSearchInput(): void {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    emitFilter();
    saveToStorage(); // 搜索只走 localStorage，不频繁打服务端
  }, 300);
}

// 选择变化立即触发 + 持久化
watch(
  () => [filters.value.group, filters.value.type, filters.value.priority, filters.value.severity_from],
  () => {
    emitFilter();
    saveToServerAndStorage();
  },
);

// ---- 初始化 ----
onMounted(async () => {
  // 优先级：服务端 > localStorage > 空
  const serverFilters = await loadServerPreference();
  const localFilters = loadFromStorage();
  filters.value = { ...serverFilters, ...localFilters };
  emitFilter();
});

// ---- 暴露接口（供父组件编程调用） ----
defineExpose({
  clearFilters,
  saveServerPreference,
  loadServerPreference,
  getFilters: () => filters.value,
  setFilters: (f: FilterState) => {
    filters.value = { ...f };
    emitFilter();
    saveToServerAndStorage();
  },
});
</script>

<template>
  <div class="filter-bar">
    <div class="filter-bar__row">
      <!-- 搜索 -->
      <div class="filter-field filter-field--search">
        <input
          :value="filters.search ?? ''"
          class="filter-input"
          placeholder="搜索需求/任务/缺陷..."
          @input="setSearch(($event.target as HTMLInputElement).value); onSearchInput()"
        />
      </div>

      <!-- 状态分组 -->
      <div class="filter-field">
        <select
          :value="filters.group ?? ''"
          class="filter-select"
          @change="setGroup(($event.target as HTMLSelectElement).value as StateGroup | '')"
        >
          <option v-for="opt in stateGroupOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <!-- 类型 -->
      <div class="filter-field">
        <select
          :value="filters.type ?? ''"
          class="filter-select"
          @change="setType(($event.target as HTMLSelectElement).value as IssueType | '')"
        >
          <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <!-- 优先级 -->
      <div class="filter-field">
        <select
          :value="filters.priority ?? ''"
          class="filter-select"
          @change="setPriority(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="opt in priorityOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <!-- 严重级别 -->
      <div class="filter-field">
        <select
          :value="filters.severity_from ?? null"
          class="filter-select"
          @change="setSeverity(Number(($event.target as HTMLSelectElement).value) || null)"
        >
          <option
            v-for="opt in severityOptions"
            :key="String(opt.value)"
            :value="opt.value ?? undefined"
          >
            {{ opt.label }}
          </option>
        </select>
      </div>

      <!-- 清除 -->
      <button v-if="hasActive()" class="filter-clear" @click="clearFilters">
        清除过滤
        <span v-if="activeCount() > 0" class="filter-count">{{ activeCount() }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  margin-bottom: 16px;
}

.filter-bar__row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-field {
  flex-shrink: 0;
}

.filter-field--search {
  flex: 1;
  min-width: 180px;
  max-width: 320px;
}

.filter-input,
.filter-select {
  padding: 6px 10px;
  font-size: 12px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  transition: border-color 0.15s;
}

.filter-input {
  width: 100%;
}

.filter-input:focus,
.filter-select:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.filter-select {
  cursor: pointer;
  min-width: 110px;
}

.filter-clear {
  padding: 6px 10px;
  font-size: 12px;
  color: var(--text-tertiary);
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.filter-clear:hover {
  color: var(--danger-500);
}

.filter-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  font-size: 10px;
  font-weight: 600;
  color: white;
  background: var(--brand-500);
  border-radius: 8px;
  line-height: 1;
}
</style>
