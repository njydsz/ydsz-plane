<script setup lang="ts">
/**
 * IssueFilter — 工作项通用过滤器组件。
 *
 * 支持按状态分组、类型、优先级、搜索关键词进行过滤。
 * 过滤偏好自动保存到 localStorage，下次打开恢复。
 */
import { onMounted, ref, watch } from "vue";

import type { IssueType, ListIssuesParams, StateGroup } from "@/api/services/issue";

const STORAGE_KEY = "ydsz_issue_filter";

// ---- Props ----
const props = defineProps<{
  projectId: number;
  workspaceSlug: string;
}>();

const emit = defineEmits<{
  (e: "filter-change", params: ListIssuesParams): void;
}>();

// ---- 过滤状态 ----
const search = ref("");
const stateGroup = ref<StateGroup | "">("");
const type = ref<IssueType | "">("");
const priority = ref<string>("");

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

// ---- 是否有活跃过滤 ----
const hasFilter = () =>
  stateGroup.value !== "" || type.value !== "" || priority.value !== "" || search.value.trim() !== "";

// ---- 构建过滤参数并发射 ----
function emitFilter() {
  const params: ListIssuesParams = {};

  if (search.value.trim()) params.search = search.value.trim();
  if (stateGroup.value) params.group = stateGroup.value;
  if (type.value) params.type = type.value;
  if (priority.value) params.priority = priority.value as ListIssuesParams["priority"];

  emit("filter-change", params);
  saveToStorage();
}

// ---- localStorage 持久化 ----
function saveToStorage() {
  try {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        search: search.value,
        stateGroup: stateGroup.value,
        type: type.value,
        priority: priority.value,
      }),
    );
  } catch { /* ignore */ }
}

function loadFromStorage() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return;
    const saved = JSON.parse(raw);
    if (saved.search) search.value = saved.search;
    if (saved.stateGroup) stateGroup.value = saved.stateGroup;
    if (saved.type) type.value = saved.type;
    if (saved.priority) priority.value = saved.priority;
  } catch { /* ignore */ }
}

function clearFilters() {
  search.value = "";
  stateGroup.value = "";
  type.value = "";
  priority.value = "";
  emitFilter();
}

// 防抖搜索
let searchTimer: ReturnType<typeof setTimeout> | null = null;
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(emitFilter, 300);
}

onMounted(() => {
  loadFromStorage();
  emitFilter();
});

// 选择变化立即触发
watch([stateGroup, type, priority], () => emitFilter());
</script>

<template>
  <div class="filter-bar">
    <div class="filter-bar__row">
      <!-- 搜索 -->
      <div class="filter-field filter-field--search">
        <input
          v-model="search"
          class="filter-input"
          placeholder="搜索工作项..."
          @input="onSearchInput"
        />
      </div>

      <!-- 状态分组 -->
      <div class="filter-field">
        <select v-model="stateGroup" class="filter-select">
          <option
            v-for="opt in stateGroupOptions"
            :key="opt.value"
            :value="opt.value"
          >{{ opt.label }}</option>
        </select>
      </div>

      <!-- 类型 -->
      <div class="filter-field">
        <select v-model="type" class="filter-select">
          <option
            v-for="opt in typeOptions"
            :key="opt.value"
            :value="opt.value"
          >{{ opt.label }}</option>
        </select>
      </div>

      <!-- 优先级 -->
      <div class="filter-field">
        <select v-model="priority" class="filter-select">
          <option
            v-for="opt in priorityOptions"
            :key="opt.value"
            :value="opt.value"
          >{{ opt.label }}</option>
        </select>
      </div>

      <!-- 清除 -->
      <button v-if="hasFilter()" class="filter-clear" @click="clearFilters">
        清除过滤
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
}

.filter-clear:hover {
  color: var(--danger-500);
}
</style>
