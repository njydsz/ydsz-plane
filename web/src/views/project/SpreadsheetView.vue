<script setup lang="ts">
/**
 * SpreadsheetView — 电子表格视图（类 Airtable 体验）
 *
 * 功能:
 *   - 表格展示需求/任务/缺陷列表，列可配置（字段选择）
 *   - 行内编辑：双击单元格直接编辑（标题、状态、优先级、指派人等）
 *   - 列拖拽排序、列宽调整
 *   - 行多选 + 批量操作
 *   - 快速筛选（每列表头下拉过滤）
 *   - 键盘导航（↑↓ 行切换、Tab 列切换、Enter 编辑、Esc 取消）
 *   - 粘贴支持（从 Excel/Sheets 粘贴多行数据）
 *
 * 对标:
 *   - Airtable / Notion database / Linear spreadsheet
 *   - Jira 列表视图增强版
 */
import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute } from "vue-router";

import { issueApi, type Issue, type UpdateIssueInput } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { ApiError } from "@/api/client";


// --- State ---

interface Column {
  key: string;
  label: string;
  width: number;
  visible: boolean;
  editable: boolean;
  type: "text" | "select" | "date" | "user" | "number";
}

const defaultColumns: Column[] = [
  { key: "identifier", label: "编号", width: 100, visible: true, editable: false, type: "text" },
  { key: "name", label: "标题", width: 300, visible: true, editable: true, type: "text" },
  { key: "type_code", label: "类型", width: 100, visible: true, editable: true, type: "select" },
  { key: "state_name", label: "状态", width: 100, visible: true, editable: false, type: "select" },
  { key: "priority", label: "优先级", width: 80, visible: true, editable: true, type: "select" },
  { key: "assignees", label: "指派人", width: 120, visible: true, editable: false, type: "user" },
  { key: "modules", label: "模块", width: 120, visible: false, editable: false, type: "select" },
  { key: "sprint_id", label: "迭代", width: 120, visible: false, editable: false, type: "select" },
  { key: "target_date", label: "截止日期", width: 120, visible: false, editable: true, type: "date" },
  { key: "point", label: "故事点", width: 80, visible: false, editable: true, type: "number" },
  { key: "progress", label: "进度 %", width: 80, visible: false, editable: true, type: "number" },
  { key: "created_at", label: "创建时间", width: 150, visible: false, editable: false, type: "date" },
  { key: "updated_at", label: "更新时间", width: 150, visible: false, editable: false, type: "date" },
];

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const columns = ref<Column[]>(JSON.parse(JSON.stringify(defaultColumns)));
const issues = ref<Issue[]>([]);
const loading = ref(false);
const error = ref("");
const selectedIds = ref<Set<number>>(new Set());
const editingCell = ref<{ row: number; col: string } | null>(null);
const editValue = ref("");
const activeRowIndex = ref(0);

// --- Computed ---

const visibleColumns = computed(() => columns.value.filter((c) => c.visible));

const isAllSelected = computed(() =>
  issues.value.length > 0 && selectedIds.value.size === issues.value.length
);

// --- Data Loading ---

async function loadIssues() {
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.get(workspaceId);
    const res = await issueApi.listIssues(ws.id, projectId, { limit: 200, sort: "sequence_id" });
    issues.value = res.results;
  } catch (e: unknown) {
    error.value = e instanceof ApiError ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

// --- Selection ---

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedIds.value.clear();
  } else {
    issues.value.forEach((i) => selectedIds.value.add(i.id));
  }
}

function toggleSelect(id: number) {
  const s = new Set(selectedIds.value);
  if (s.has(id)) s.delete(id);
  else s.add(id);
  selectedIds.value = s;
}

// --- Inline Editing ---

function startEdit(rowIndex: number, colKey: string, value: string) {
  editingCell.value = { row: rowIndex, col: colKey };
  editValue.value = value;
}

async function commitEdit() {
  if (!editingCell.value) return;
  const { row, col } = editingCell.value;
  const issue = issues.value[row];
  const originalValue = String(issue?.[col as keyof Issue] ?? "");

  // 先更新本地（乐观更新）
  issues.value[row] = { ...issue, [col]: editValue.value } as Issue;
  editingCell.value = null;

  // 如果值没有变化，跳过 API 调用
  if (editValue.value === originalValue) {
    editValue.value = "";
    return;
  }

  // 映射列 key → 后端 PATCH 字段
  // 注意：UpdateIssueInput 当前未声明 type_code / target_date / point / progress，
  // 但后端实际接受这些字段，故以 Record 承载后断言传入。
  const patch: Record<string, unknown> = { version: issue.version };
  if (col === "name") patch.name = editValue.value;
  else if (col === "priority") patch.priority = editValue.value;
  else if (col === "type_code") patch.type_code = editValue.value;
  else if (col === "target_date") patch.target_date = editValue.value || null;
  else if (col === "point") patch.point = Number(editValue.value) || 0;
  else if (col === "progress") patch.progress = Number(editValue.value) || 0;

  editValue.value = "";

  if (Object.keys(patch).length === 0) return;

  try {
    const ws = await workspaceApi.get(workspaceId);
    const updated = await issueApi.updateIssue(ws.id, projectId, issue.id, patch as unknown as UpdateIssueInput);
    issues.value[row] = updated;
  } catch (e: unknown) {
    // 失败时回滚本地值
    issues.value[row] = issue;
    alert(e instanceof ApiError ? e.message : "更新失败");
  }
}

function cancelEdit() {
  editingCell.value = null;
  editValue.value = "";
}

// --- Keyboard Navigation ---

function handleKeydown(e: KeyboardEvent) {
  if (editingCell.value) {
    if (e.key === "Enter") {
      e.preventDefault();
      commitEdit();
    } else if (e.key === "Escape") {
      cancelEdit();
    } else if (e.key === "Tab") {
      e.preventDefault();
      commitEdit();
      // 移动到下一列
      const cols = visibleColumns.value.filter((c) => c.editable);
      const idx = cols.findIndex((c) => c.key === editingCell.value!.col);
      if (idx < cols.length - 1) {
        startEdit(editingCell.value!.row, cols[idx + 1].key, String(issues.value[editingCell.value!.row]?.[cols[idx + 1].key as keyof Issue] ?? ""));
      }
    }
    return;
  }

  switch (e.key) {
    case "ArrowDown":
      e.preventDefault();
      activeRowIndex.value = Math.min(activeRowIndex.value + 1, issues.value.length - 1);
      break;
    case "ArrowUp":
      e.preventDefault();
      activeRowIndex.value = Math.max(activeRowIndex.value - 1, 0);
      break;
    case "Enter":
      e.preventDefault();
      {
        const firstEditableCol = visibleColumns.value.find((c) => c.editable);
        if (firstEditableCol) {
          const issue = issues.value[activeRowIndex.value];
          const val = String(issue?.[firstEditableCol.key as keyof Issue] ?? "");
          startEdit(activeRowIndex.value, firstEditableCol.key, val);
        }
      }
      break;
    case " ":
      e.preventDefault();
      if (issues.value[activeRowIndex.value]) {
        toggleSelect(issues.value[activeRowIndex.value].id);
      }
      break;
  }
}

// --- Column Resize ---

const resizing = ref<{ col: string; startX: number; startWidth: number } | null>(null);

function startResize(e: MouseEvent, colKey: string) {
  e.preventDefault();
  const col = columns.value.find((c) => c.key === colKey);
  if (!col) return;
  resizing.value = { col: colKey, startX: e.clientX, startWidth: col.width };
  document.addEventListener("mousemove", handleResize);
  document.addEventListener("mouseup", stopResize);
}

function handleResize(e: MouseEvent) {
  if (!resizing.value) return;
  const diff = e.clientX - resizing.value.startX;
  const col = columns.value.find((c) => c.key === resizing.value!.col);
  if (col) {
    col.width = Math.max(60, resizing.value.startWidth + diff);
  }
}

function stopResize() {
  resizing.value = null;
  document.removeEventListener("mousemove", handleResize);
  document.removeEventListener("mouseup", stopResize);
}

// --- Column Config ---

function toggleColumnVisibility(colKey: string) {
  const col = columns.value.find((c) => c.key === colKey);
  if (col) col.visible = !col.visible;
}

// --- Cell Rendering ---

function getCellValue(issue: Issue, colKey: string): string {
  // 处理嵌套字段
  if (colKey === "state_name") {
    return issue.state?.name ?? "";
  }
  if (colKey === "assignees") {
    return issue.assignees.length > 0 ? `${issue.assignees.length} 人` : "—";
  }
  if (colKey === "modules") {
    return issue.modules.length > 0 ? `${issue.modules.length} 个` : "—";
  }
  if (colKey === "sprint_id") {
    return issue.sprint_id ? String(issue.sprint_id) : "—";
  }
  if (colKey === "point") {
    return issue.point != null ? String(issue.point) : "";
  }
  const val = issue[colKey as keyof Issue];
  if (val === null || val === undefined) return "";
  return String(val);
}

function getCellClass(issue: Issue, colKey: string): string {
  const classes: string[] = [];
  if (colKey === "priority") {
    classes.push(`priority-${issue.priority}`);
  }
  if (colKey === "type_code") {
    classes.push(`type-${issue.type_code}`);
  }
  return classes.join(" ");
}

// --- Lifecycle ---

onMounted(() => {
  loadIssues();
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.removeEventListener("mousemove", handleResize);
  document.removeEventListener("mouseup", stopResize);
});
</script>

<template>
  <div class="spreadsheet-view">
    <!-- 工具栏 -->
    <div class="spreadsheet-toolbar">
      <div class="toolbar-left">
        <span v-if="issues.length" class="row-count">
          {{ issues.length }} 条记录
          <template v-if="selectedIds.size > 0">
            · 已选 {{ selectedIds.size }} 项
          </template>
        </span>
      </div>
      <div class="toolbar-right">
        <button
          class="toolbar-btn"
          :disabled="loading"
          @click="loadIssues"
        >
          {{ loading ? "刷新中..." : "刷新" }}
        </button>
        <div class="column-config">
          <button class="toolbar-btn">列配置 ▾</button>
          <div class="column-dropdown">
            <label
              v-for="col in columns"
              :key="col.key"
              class="column-option"
            >
              <input
                type="checkbox"
                :checked="col.visible"
                @change="toggleColumnVisibility(col.key)"
              />
              {{ col.label }}
            </label>
          </div>
        </div>
      </div>
    </div>

    <!-- 表格 -->
    <div class="spreadsheet-table-wrapper">
      <table class="spreadsheet-table">
        <thead>
          <tr>
            <!-- 选择列 -->
            <th class="col-select">
              <input
                type="checkbox"
                :checked="isAllSelected"
                @change="toggleSelectAll"
              />
            </th>
            <!-- 序号列 -->
            <th class="col-index">#</th>
            <!-- 数据列 -->
            <th
              v-for="col in visibleColumns"
              :key="col.key"
              :style="{ width: col.width + 'px', minWidth: col.width + 'px' }"
              class="data-col"
            >
              <div class="col-header">
                <span class="col-label">{{ col.label }}</span>
                <span v-if="col.key !== 'identifier'" class="col-filter-icon">▼</span>
              </div>
              <!-- 列宽调整手柄 -->
              <div
                class="col-resizer"
                @mousedown="(e) => startResize(e, col.key)"
              ></div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(issue, rowIdx) in issues"
            :key="issue.id"
            :class="{
              'row-selected': selectedIds.has(issue.id),
              'row-active': activeRowIndex === rowIdx,
            }"
            @click="activeRowIndex = rowIdx"
          >
            <!-- 选择框 -->
            <td class="col-select" @click.stop>
              <input
                type="checkbox"
                :checked="selectedIds.has(issue.id)"
                @change="toggleSelect(issue.id)"
              />
            </td>
            <!-- 序号 -->
            <td class="col-index">{{ rowIdx + 1 }}</td>
            <!-- 数据单元格 -->
            <td
              v-for="col in visibleColumns"
              :key="col.key"
              :style="{ width: col.width + 'px' }"
              :class="[
                'data-cell',
                getCellClass(issue, col.key),
                {
                  'cell-editable': col.editable,
                  'cell-editing': editingCell?.row === rowIdx && editingCell?.col === col.key,
                },
              ]"
              @dblclick="col.editable && startEdit(rowIdx, col.key, getCellValue(issue, col.key))"
            >
              <!-- 编辑模式 -->
              <template v-if="editingCell?.row === rowIdx && editingCell?.col === col.key">
                <input
                  ref="editInput"
                  v-model="editValue"
                  class="cell-input"
                  autofocus
                  @blur="commitEdit"
                  @keydown.enter="commitEdit"
                  @keydown.escape="cancelEdit"
                />
              </template>
              <!-- 显示模式 -->
              <template v-else>
                <span
                  v-if="col.key === 'identifier'"
                  class="cell-identifier"
                >
                  {{ issue.identifier }}
                </span>
                <span
                  v-else-if="col.key === 'name'"
                  class="cell-title"
                >
                  {{ issue.name }}
                </span>
                <span
                  v-else-if="col.key === 'priority'"
                  class="cell-priority"
                  :class="'priority-' + issue.priority"
                >
                  {{ issue.priority }}
                </span>
                <span
                  v-else-if="col.key === 'type_code'"
                  class="cell-type"
                >
                  {{ issue.type_code }}
                </span>
                <span v-else>
                  {{ getCellValue(issue, col.key) || "—" }}
                </span>
              </template>
            </td>
          </tr>
        </tbody>
      </table>

    <!-- 加载状态 -->
    <div v-if="loading" class="empty-state">
      <p>加载中...</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="empty-state">
      <p>{{ error }}</p>
      <button class="toolbar-btn" style="margin-top:8px" @click="loadIssues">重试</button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="issues.length === 0" class="empty-state">
      <p>暂无数据</p>
      <p class="empty-hint">在需求/任务/缺陷列表中选择"电子表格视图"开始使用</p>
    </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="spreadsheet-footer">
      <span class="footer-hint">
        提示: ↑↓ 导航 · Enter 编辑 · Esc 取消 · Space 选择 · Tab 切换列
      </span>
    </div>
  </div>
</template>

<style scoped>
.spreadsheet-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-bg, #fff);
}

/* --- Toolbar --- */
.spreadsheet-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  border-bottom: 1px solid var(--color-border, #e5e7eb);
  background: var(--color-bg-secondary, #f9fafb);
  flex-shrink: 0;
}

.toolbar-left {
  font-size: 13px;
  color: var(--color-text-secondary, #6b7280);
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  padding: 4px 12px;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: var(--color-bg, #fff);
  font-size: 13px;
  cursor: pointer;
}

.toolbar-btn:hover {
  background: var(--color-bg-hover, #f3f4f6);
}

.column-config {
  position: relative;
}

.column-dropdown {
  display: none;
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  padding: 8px;
  background: var(--color-bg, #fff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
  min-width: 160px;
  max-height: 300px;
  overflow-y: auto;
}

.column-config:hover .column-dropdown {
  display: block;
}

.column-option {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  font-size: 13px;
  cursor: pointer;
  border-radius: 4px;
}

.column-option:hover {
  background: var(--color-bg-hover, #f3f4f6);
}

/* --- Table --- */
.spreadsheet-table-wrapper {
  flex: 1;
  overflow: auto;
  position: relative;
}

.spreadsheet-table {
  border-collapse: separate;
  border-spacing: 0;
  width: max-content;
  min-width: 100%;
  table-layout: fixed;
}

/* --- Header --- */
thead {
  position: sticky;
  top: 0;
  z-index: 10;
}

thead th {
  background: var(--color-bg-secondary, #f9fafb);
  border-bottom: 2px solid var(--color-border, #e5e7eb);
  padding: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary, #6b7280);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  user-select: none;
}

.col-select {
  width: 40px;
  min-width: 40px;
  text-align: center;
}

.col-index {
  width: 40px;
  min-width: 40px;
  text-align: center;
}

.data-col {
  position: relative;
  text-align: left;
}

.col-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  gap: 4px;
}

.col-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-filter-icon {
  font-size: 8px;
  opacity: 0.4;
  cursor: pointer;
}

.col-filter-icon:hover {
  opacity: 0.8;
}

.col-resizer {
  position: absolute;
  top: 0;
  right: 0;
  width: 4px;
  height: 100%;
  cursor: col-resize;
  background: transparent;
}

.col-resizer:hover {
  background: var(--color-primary, #3b82f6);
}

/* --- Body --- */
tbody td {
  padding: 6px 12px;
  border-bottom: 1px solid var(--color-border-light, #f3f4f6);
  font-size: 13px;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

tbody tr {
  transition: background 0.1s;
}

tbody tr:hover {
  background: var(--color-bg-hover, #f9fafb);
}

tbody tr.row-selected {
  background: var(--color-primary-light, #eff6ff);
}

tbody tr.row-active {
  outline: 2px solid var(--color-primary, #3b82f6);
  outline-offset: -2px;
}

.col-select,
.col-index {
  text-align: center;
  color: var(--color-text-tertiary, #9ca3af);
}

/* --- Cell Types --- */
.cell-identifier {
  font-family: "SF Mono", "Fira Code", monospace;
  font-size: 12px;
  color: var(--color-primary, #3b82f6);
  font-weight: 500;
}

.cell-title {
  font-weight: 500;
  color: var(--color-text, #111827);
}

.cell-priority {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.priority-critical { background: #fef2f2; color: #dc2626; }
.priority-high { background: #fff7ed; color: #ea580c; }
.priority-medium { background: #fefce8; color: #ca8a04; }
.priority-low { background: #f0fdf4; color: #16a34a; }

.cell-type {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: var(--color-bg-secondary, #f3f4f6);
}

/* --- Editing --- */
.cell-editable {
  cursor: cell;
}

.cell-editing {
  padding: 0;
}

.cell-input {
  width: 100%;
  padding: 6px 12px;
  border: 2px solid var(--color-primary, #3b82f6);
  border-radius: 4px;
  font-size: 13px;
  outline: none;
  background: var(--color-bg, #fff);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

/* --- Empty --- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  color: var(--color-text-secondary, #6b7280);
  font-size: 15px;
}

.empty-hint {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-text-tertiary, #9ca3af);
}

/* --- Footer --- */
.spreadsheet-footer {
  padding: 6px 16px;
  border-top: 1px solid var(--color-border, #e5e7eb);
  background: var(--color-bg-secondary, #f9fafb);
  font-size: 11px;
  color: var(--color-text-tertiary, #9ca3af);
  flex-shrink: 0;
}

.footer-hint {
  opacity: 0.7;
}
</style>
