<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { type Issue, type IssueType, type State } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { useIssueStore } from "@/stores/issue";

const route = useRoute();
const router = useRouter();
const issueStore = useIssueStore();

// ---- 状态 ----
const projectId = computed(() => Number(route.params.projectId));
const wsId = ref(0);
const loading = ref(true);
const error = ref("");

// 排序
const sortField = ref<string>("identifier");
const sortDir = ref<"asc" | "desc">("asc");

// 搜索
const searchQuery = ref("");

// 批量操作
const selectedIds = ref<Set<number>>(new Set());
const showDeleteConfirm = ref(false);
const batchDeleting = ref(false);

// ---- 派生 ----
const typeLabel = (t: IssueType) =>
  ({ requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[t] ?? t;

const priorityLabel = (p: string) =>
  ({ urgent: "紧急", high: "高", medium: "中", low: "低", none: "无" } as Record<string, string>)[p] ?? p;

const stateMap = computed(() => {
  const m: Record<number, State> = {};
  for (const s of issueStore.states) m[s.id] = s;
  return m;
});

// 过滤后的数据
const filteredIssues = computed(() => {
  let list = [...issueStore.issues];
  const q = searchQuery.value.trim().toLowerCase();
  if (q) {
    list = list.filter(
      (i) =>
        i.name.toLowerCase().includes(q) ||
        i.identifier.toLowerCase().includes(q),
    );
  }
  return list;
});

// 排序后的数据
const sortedIssues = computed(() => {
  const list = [...filteredIssues.value];
  const field = sortField.value;
  const dir = sortDir.value;

  list.sort((a, b) => {
    let va: unknown, vb: unknown;
    switch (field) {
      case "identifier":
        va = a.identifier;
        vb = b.identifier;
        break;
      case "name":
        va = a.name;
        vb = b.name;
        break;
      case "type_code":
        va = a.type_code;
        vb = b.type_code;
        break;
      case "priority": {
        const po: Record<string, number> = { urgent: 0, high: 1, medium: 2, low: 3, none: 4 };
        va = po[a.priority] ?? 99;
        vb = po[b.priority] ?? 99;
        break;
      }
      case "state":
        va = stateMap.value[a.state_id]?.name ?? "";
        vb = stateMap.value[b.state_id]?.name ?? "";
        break;
      case "severity":
        va = a.severity ?? 0;
        vb = b.severity ?? 0;
        break;
      case "point":
        va = a.point ?? 0;
        vb = b.point ?? 0;
        break;
      case "updated_at":
        va = a.updated_at;
        vb = b.updated_at;
        break;
      case "created_at":
        va = a.created_at;
        vb = b.created_at;
        break;
      default:
        return 0;
    }
    if (va == null) return dir === "asc" ? -1 : 1;
    if (vb == null) return dir === "asc" ? 1 : -1;
    if (va < vb) return dir === "asc" ? -1 : 1;
    if (va > vb) return dir === "asc" ? 1 : -1;
    return 0;
  });
  return list;
});

const hasSelection = computed(() => selectedIds.value.size > 0);

// ---- 方法 ----
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsSlug = String(route.params.workspaceSlug ?? "");
    let wsIdVal: number;
    if (wsSlug) {
      const ws = await workspaceApi.getBySlug(wsSlug);
      wsIdVal = ws.id;
    } else {
      wsIdVal = Number(route.params.wsId);
    }
    wsId.value = wsIdVal;
    await Promise.all([
      issueStore.fetchStates(wsIdVal, projectId.value),
      issueStore.fetchIssues(wsIdVal, projectId.value),
    ]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function toggleSort(field: string) {
  if (sortField.value === field) {
    sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
  } else {
    sortField.value = field;
    sortDir.value = "asc";
  }
}

function sortIndicator(field: string): string {
  if (sortField.value !== field) return "";
  return sortDir.value === "asc" ? " ↑" : " ↓";
}

function openIssue(issueId: number) {
  router.push(`/${route.params.workspaceSlug}/projects/${projectId.value}/issues/${issueId}`);
}

function toggleSelect(issueId: number) {
  const next = new Set(selectedIds.value);
  if (next.has(issueId)) next.delete(issueId);
  else next.add(issueId);
  selectedIds.value = next;
}

function toggleSelectAll() {
  if (selectedIds.value.size === sortedIssues.value.length) {
    selectedIds.value = new Set();
  } else {
    selectedIds.value = new Set(sortedIssues.value.map((i) => i.id));
  }
}

async function batchDelete() {
  if (selectedIds.value.size === 0) return;
  batchDeleting.value = true;
  try {
    for (const id of selectedIds.value) {
      await issueStore.deleteIssue(wsId.value, projectId.value, id);
    }
    selectedIds.value = new Set();
    showDeleteConfirm.value = false;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除失败";
  } finally {
    batchDeleting.value = false;
  }
}

const columns: { key: string; label: string; width?: string; sortable?: boolean }[] = [
  { key: "identifier", label: "编号", width: "120px", sortable: true },
  { key: "name", label: "名称", sortable: true },
  { key: "type_code", label: "类型", width: "72px", sortable: true },
  { key: "priority", label: "优先级", width: "72px", sortable: true },
  { key: "state", label: "状态", width: "90px", sortable: true },
  { key: "severity", label: "严重度", width: "72px", sortable: true },
  { key: "point", label: "点数", width: "60px", sortable: true },
  { key: "assignees", label: "指派人", width: "100px" },
  { key: "updated_at", label: "更新时间", width: "130px", sortable: true },
];

onMounted(load);
</script>

<template>
  <div class="list-view">
    <header class="list-view__header">
      <div>
        <h1>列表</h1>
        <p class="hint">共 {{ issueStore.total }} 个工作项</p>
      </div>
      <div class="list-view__header-right">
        <input
          v-model="searchQuery"
          class="search-input"
          placeholder="搜索工作项名称或编号..."
        />
      </div>
    </header>

    <!-- 批量操作工具栏 -->
    <div v-if="hasSelection" class="batch-bar">
      <span class="batch-bar__info">已选 {{ selectedIds.size }} 项</span>
      <button class="btn btn--sm btn--danger" @click="showDeleteConfirm = true">批量删除</button>
      <button class="btn btn--sm btn--ghost" @click="selectedIds = new Set()">取消选择</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>

    <!-- 表格 -->
    <div v-else class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th class="th-check">
              <input
                type="checkbox"
                :checked="selectedIds.size === sortedIssues.length && sortedIssues.length > 0"
                @change="toggleSelectAll"
              />
            </th>
            <th
              v-for="col in columns"
              :key="col.key"
              :style="col.width ? { width: col.width, minWidth: col.width } : {}"
              :class="{ 'th--sortable': col.sortable }"
              @click="col.sortable && toggleSort(col.key)"
            >
              {{ col.label }}<span class="sort-indicator">{{ sortIndicator(col.key) }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="sortedIssues.length === 0">
            <td :colspan="columns.length + 1" class="empty-cell">
              暂无工作项
            </td>
          </tr>
          <tr
            v-for="iss in sortedIssues"
            :key="iss.id"
            :class="{ 'row--selected': selectedIds.has(iss.id) }"
            class="row"
          >
            <td class="td-check" @click.stop>
              <input
                type="checkbox"
                :checked="selectedIds.has(iss.id)"
                @change="toggleSelect(iss.id)"
              />
            </td>
            <td class="td-identifier">
              <span class="identifier-link" @click="openIssue(iss.id)">
                {{ iss.identifier }}
              </span>
            </td>
            <td class="td-name">
              <span class="name-link" @click="openIssue(iss.id)">
                <span class="type-dot" :class="`dot-${iss.type_code}`"></span>
                {{ iss.name }}
              </span>
            </td>
            <td>
              <span class="badge-sm" :class="`type-${iss.type_code}`">
                {{ typeLabel(iss.type_code) }}
              </span>
            </td>
            <td>
              <span class="priority-tag" :class="`pri-${iss.priority}`">
                {{ priorityLabel(iss.priority) }}
              </span>
            </td>
            <td>
              <span
                class="state-pill"
                :style="{ backgroundColor: stateMap[iss.state_id]?.color ?? '#8DA2C2' }"
              >
                {{ stateMap[iss.state_id]?.name ?? "未知" }}
              </span>
            </td>
            <td class="td-num">
              {{ iss.severity ? "S" + iss.severity : "—" }}
            </td>
            <td class="td-num">
              {{ iss.point != null ? iss.point + "pt" : "—" }}
            </td>
            <td class="td-num">
              {{ iss.assignees.length || "—" }}
            </td>
            <td class="td-date">
              {{ new Date(iss.updated_at).toLocaleDateString() }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 批量删除确认弹窗 -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
        <div class="modal-confirm" @click.stop>
          <h3>确认删除</h3>
          <p>确定要归档选中的 {{ selectedIds.size }} 个工作项吗？</p>
          <div class="modal-confirm__actions">
            <button class="btn btn--secondary" @click="showDeleteConfirm = false" :disabled="batchDeleting">取消</button>
            <button class="btn btn--danger" @click="batchDelete" :disabled="batchDeleting">
              {{ batchDeleting ? "删除中..." : "确认删除" }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.list-view__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 12px;
}

.list-view__header h1 {
  font-size: 20px;
  margin: 0 0 4px;
}

.list-view__header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hint {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}

.search-input {
  padding: 6px 12px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
  width: 260px;
  transition: border-color 0.15s;
}

.search-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

/* Batch bar */
.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--brand-50);
  border: 1px solid var(--brand-500);
  border-radius: var(--radius-sm);
  margin-bottom: 12px;
}

.batch-bar__info {
  font-size: 13px;
  color: var(--brand-600);
  font-weight: 500;
  flex: 1;
}

/* Table */
.table-wrap {
  overflow-x: auto;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  font-size: 12px;
  color: var(--text-tertiary);
  background: var(--surface-2);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
  user-select: none;
}

.th--sortable {
  cursor: pointer;
}

.th--sortable:hover {
  color: var(--text-primary);
}

.sort-indicator {
  color: var(--brand-500);
  font-weight: 600;
}

.th-check {
  width: 40px;
  padding: 10px 8px !important;
}

.th-check input {
  width: 14px;
  height: 14px;
  accent-color: var(--brand-500);
}

.table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  vertical-align: middle;
}

.td-check {
  width: 40px;
  padding: 10px 8px !important;
}

.td-check input {
  width: 14px;
  height: 14px;
  accent-color: var(--brand-500);
}

.td-identifier {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
}

.td-name {
  min-width: 240px;
}

.td-num {
  text-align: center;
}

.td-date {
  font-size: 12px;
  white-space: nowrap;
}

.row:hover {
  background: var(--surface-2);
}

.row--selected {
  background: var(--brand-50);
}

.row--selected:hover {
  background: var(--brand-50);
}

.empty-cell {
  text-align: center;
  color: var(--text-tertiary);
  padding: 48px 0 !important;
}

.identifier-link,
.name-link {
  cursor: pointer;
  color: var(--text-primary);
}

.identifier-link:hover {
  color: var(--brand-500);
}

.name-link {
  display: flex;
  align-items: center;
  gap: 6px;
}

.name-link:hover {
  color: var(--brand-500);
}

/* Type dot */
.type-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-requirement { background: var(--brand-500); }
.dot-task { background: var(--success-500); }
.dot-defect { background: var(--danger-500); }

/* Badge */
.badge-sm {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
  white-space: nowrap;
}

.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.type-task { background: var(--success-50); color: var(--success-600); }
.type-defect { background: var(--danger-50); color: var(--danger-600); }

/* Priority */
.priority-tag {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.pri-urgent { color: var(--priority-urgent); }
.pri-high { color: var(--priority-high); }
.pri-medium { color: var(--priority-medium); }
.pri-low { color: var(--priority-low); }
.pri-none { color: var(--priority-none); }

/* State pill */
.state-pill {
  display: inline-block;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  color: #fff;
  font-weight: 500;
  white-space: nowrap;
}

/* Buttons */
.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  font-family: inherit;
}

.btn--sm { padding: 4px 10px; font-size: 12px; }

.btn--danger {
  background: var(--danger-500);
  color: #fff;
  border-color: var(--danger-500);
}

.btn--danger:hover:not(:disabled) {
  background: #c62828;
}

.btn--secondary {
  background: var(--surface-1);
  color: var(--text-secondary);
  border-color: var(--border-default);
}

.btn--secondary:hover:not(:disabled) {
  background: var(--surface-3);
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--text-tertiary);
}

.btn--ghost:hover:not(:disabled) {
  color: var(--text-primary);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Confirm modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
}

.modal-confirm {
  width: 400px;
  padding: 24px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
}

.modal-confirm h3 {
  margin: 0 0 8px;
  font-size: 16px;
}

.modal-confirm p {
  margin: 0 0 20px;
  font-size: 13px;
  color: var(--text-secondary);
}

.modal-confirm__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.loading, .error {
  text-align: center;
  padding: 48px 0;
  color: var(--text-tertiary);
}

.error { color: var(--danger-500); }
</style>
