<script setup lang="ts">
/**
 * 工作项列表页 — 表格视图展示工作项。
 * 支持: 服务端排序 / 分页 / 列过滤 / 批量选择与删除。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { type IssueType, type ListIssuesParams, type State, issueApi } from "@/api/services/issue";
import { workspaceApi } from "@/api/services/workspace";
import { useIssueStore } from "@/stores/issue";
import { usePeekStore } from "@/stores/peek";
import { prefs } from "@/lib/prefs";
import IssueFilter from "./IssueFilter.vue";
import { AppLoadingState, AppErrorState, AppEmptyState } from "@/components";

const route = useRoute();
const issueStore = useIssueStore();
const peek = usePeekStore();

// ---- 状态 ----
const projectId = computed(() => Number(route.params.projectId));
const wsId = ref(0);
const loading = ref(true);
const error = ref("");

// 排序（服务端）
const sortField = ref<string>("-updated_at"); // 默认按更新时间倒序

// 分页
const page = ref(1);
const perPage = ref(50);
const total = computed(() => issueStore.total);

// 当前过滤参数
const currentFilter = ref<ListIssuesParams>({});

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

const hasSelection = computed(() => selectedIds.value.size > 0);

function buildSortParam(): string {
  return sortField.value;
}

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

    const params: ListIssuesParams = {
      ...currentFilter.value,
      sort: buildSortParam(),
      limit: perPage.value,
      offset: (page.value - 1) * perPage.value,
    };

    await Promise.all([
      issueStore.fetchStates(wsIdVal, projectId.value),
      issueStore.fetchIssues(wsIdVal, projectId.value, params),
    ]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function onFilterChange(params: ListIssuesParams) {
  currentFilter.value = params;
  page.value = 1;  // 过滤条件变化时回到第一页
  load();
}

function toggleSort(field: string) {
  // 服务端排序：切换升序/降序或更改排序字段
  const isDesc = sortField.value.startsWith("-");
  const currentCol = isDesc ? sortField.value.slice(1) : sortField.value;

  if (currentCol === field) {
    // 切换方向
    sortField.value = isDesc ? field : `-${field}`;
  } else {
    // 新字段默认降序
    sortField.value = `-${field}`;
  }
  page.value = 1;
  load();
}

function sortIndicator(field: string): string {
  const isDesc = sortField.value.startsWith("-");
  const currentCol = isDesc ? sortField.value.slice(1) : sortField.value;
  if (currentCol !== field) return "";
  return isDesc ? " ↓" : " ↑";
}

function goToPage(p: number) {
  page.value = p;
  load();
}

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / perPage.value)));

function openIssue(issueId: number) {
  // 单击打开 Peek 预览抽屉；抽屉内有「打开详情」按钮执行路由跳转
  peek.open(String(route.params.workspaceSlug), projectId.value, issueId);
}

function toggleSelect(issueId: number) {
  const next = new Set(selectedIds.value);
  if (next.has(issueId)) next.delete(issueId);
  else next.add(issueId);
  selectedIds.value = next;
}

function toggleSelectAll() {
  if (selectedIds.value.size === issueStore.issues.length) {
    selectedIds.value = new Set();
  } else {
    selectedIds.value = new Set(issueStore.issues.map((i) => i.id));
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
    load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除失败";
  } finally {
    batchDeleting.value = false;
  }
}

async function batchTransition(toStateId: number) {
  try {
    const r = await issueApi.batch(wsId.value, projectId.value, {
      issue_ids: [...selectedIds.value],
      to_state_id: toStateId,
    });
    selectedIds.value = new Set();
    if (r.failed > 0) error.value = `${r.succeeded} 项成功，${r.failed} 项失败`;
    load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "批量流转失败";
  }
}

async function batchUpdatePriority(pri: string) {
  try {
    const r = await issueApi.batch(wsId.value, projectId.value, {
      issue_ids: [...selectedIds.value],
      priority: pri,
    });
    selectedIds.value = new Set();
    if (r.failed > 0) error.value = `${r.succeeded} 项成功，${r.failed} 项失败`;
    load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "批量更新失败";
  }
}

function severityText(s?: number | null): string {
  if (s == null) return "-";
  return `S${s}`;
}

const columns: { key: string; label: string; width?: string; sortable?: boolean }[] = [
  { key: "identifier", label: "编号", width: "120px" },
  { key: "name", label: "名称", sortable: true },
  { key: "type_code", label: "类型", width: "72px", sortable: true },
  { key: "priority", label: "优先级", width: "72px", sortable: true },
  { key: "state", label: "状态", width: "90px" },
  { key: "severity", label: "严重度", width: "72px", sortable: true },
  { key: "point", label: "点数", width: "60px", sortable: true },
  { key: "assignees", label: "指派人", width: "100px" },
  { key: "updated_at", label: "更新时间", width: "130px", sortable: true },
];

onMounted(() => {
  prefs.setLastView(projectId.value, "list");
  load();
});

const exportCsvUrl = computed(() =>
  issueApi.exportUrl(wsId.value, projectId.value, currentFilter.value, "csv"),
);

const exportXlsxUrl = computed(() =>
  issueApi.exportUrl(wsId.value, projectId.value, currentFilter.value, "xlsx"),
);

/** 导出格式下拉是否展开 */
const showExportDropdown = ref(false);

</script>

<template>
  <div class="list-view">
    <header class="list-view__header">
      <div>
        <h1>列表</h1>
        <p class="hint">共 {{ total }} 个工作项</p>
      </div>
      <div class="list-view__header-right">
        <div class="export-dropdown" @mouseleave="showExportDropdown = false">
          <button
            class="btn btn--sm btn--export"
            @mouseenter="showExportDropdown = true"
          >
导出
</button>
          <div v-if="showExportDropdown" class="export-dropdown__menu">
            <a :href="exportCsvUrl" class="export-dropdown__item" download>导出 CSV</a>
            <a :href="exportXlsxUrl" class="export-dropdown__item" download>导出 Excel (.xlsx)</a>
          </div>
        </div>
        <div class="view-switcher">
          <router-link
            :to="`/${route.params.workspaceSlug}/projects/${projectId}/board`"
            class="view-tab"
          >
看板
</router-link>
          <router-link
            :to="`/${route.params.workspaceSlug}/projects/${projectId}/list`"
            class="view-tab is-active"
          >
列表
</router-link>
        </div>
      </div>
    </header>

    <!-- 过滤器 -->
    <IssueFilter
      :project-id="projectId"
      :workspace-slug="String(route.params.workspaceSlug)"
      @filter-change="onFilterChange"
    />

    <!-- 批量操作工具栏 -->
    <div v-if="hasSelection" class="batch-bar">
      <span class="batch-bar__info">已选 {{ selectedIds.size }} 项</span>
      <select class="batch-select" @change="(e: Event) => { const v = Number((e.target as HTMLSelectElement).value); if (v) batchTransition(v) }">
        <option value="">批量流转...</option>
        <option v-for="st in issueStore.states" :key="st.id" :value="st.id">{{ st.name }}</option>
      </select>
      <select class="batch-select" @change="(e: Event) => { const v = (e.target as HTMLSelectElement).value; if (v) batchUpdatePriority(v) }">
        <option value="">批量优先级...</option>
        <option value="urgent">紧急</option>
        <option value="high">高</option>
        <option value="medium">中</option>
        <option value="low">低</option>
        <option value="none">无</option>
      </select>
      <button class="btn btn--sm btn--danger" @click="showDeleteConfirm = true">批量删除</button>
      <button class="btn btn--sm btn--ghost" @click="selectedIds = new Set()">取消选择</button>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <!-- 表格 -->
    <AppEmptyState
      v-else-if="!loading && !error && issueStore.issues.length === 0"
      title="暂无工作项"
      description="当前过滤条件下没有工作项"
    />
    <div v-else class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th class="th-check">
              <input
                type="checkbox"
                :checked="selectedIds.size === issueStore.issues.length && issueStore.issues.length > 0"
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
          <tr v-if="issueStore.issues.length === 0">
            <td :colspan="columns.length + 1" class="empty-cell">
              暂无工作项
            </td>
          </tr>
          <tr
            v-for="iss in issueStore.issues"
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
              <span class="priority-text" :class="`pri-${iss.priority}`">
                {{ priorityLabel(iss.priority) }}
              </span>
            </td>
            <td>
              <span
                class="state-badge"
                :style="{ backgroundColor: stateMap[iss.state_id]?.color ?? '#ccc' }"
              >
                {{ stateMap[iss.state_id]?.name ?? iss.state_id }}
              </span>
            </td>
            <td class="td-severity">
              {{ severityText(iss.severity) }}
            </td>
            <td class="td-num">
              {{ iss.point != null ? iss.point + "pt" : "-" }}
            </td>
            <td class="td-assignees">
              <span v-if="iss.assignees?.length > 0">
                <span v-for="uid in iss.assignees.slice(0, 3)" :key="uid" class="avatar-placeholder">
                  U{{ uid }}
                </span>
                <span v-if="iss.assignees.length > 3" class="avatar-more">+{{ iss.assignees.length - 3 }}</span>
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td class="td-date">
              {{ new Date(iss.updated_at).toLocaleDateString("zh-CN") }}
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页器 -->
      <div v-if="totalPages > 1" class="pagination">
        <button
          class="page-btn"
          :disabled="page <= 1"
          @click="goToPage(page - 1)"
        >
上一页
</button>

        <template v-for="p in totalPages" :key="p">
          <button
            v-if="p <= 3 || p > totalPages - 3 || Math.abs(p - page) <= 1"
            class="page-btn"
            :class="{ 'page-btn--active': p === page }"
            @click="goToPage(p)"
          >
{{ p }}
</button>
          <span v-else-if="p === 4 || p === totalPages - 3" class="page-ellipsis">...</span>
        </template>

        <button
          class="page-btn"
          :disabled="page >= totalPages"
          @click="goToPage(page + 1)"
        >
下一页
</button>

        <span class="page-info">第 {{ page }} / {{ totalPages }} 页</span>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal-box">
        <h3>确认删除</h3>
        <p>确定要批量删除 {{ selectedIds.size }} 个工作项吗？此操作不可撤销。</p>
        <div class="modal-actions">
          <button class="btn btn--ghost" @click="showDeleteConfirm = false">取消</button>
          <button
            class="btn btn--danger"
            :disabled="batchDeleting"
            @click="batchDelete"
          >
{{ batchDeleting ? "删除中..." : "确认删除" }}
</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-view__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}
.list-view__header h1 { font-size: 20px; margin: 0 0 4px; }
.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }

.list-view__header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.view-switcher {
  display: flex;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  overflow: hidden;
}
.view-tab {
  padding: 5px 12px; font-size: 12px; font-weight: 500;
  color: var(--text-tertiary); text-decoration: none;
  background: var(--surface-2); transition: background 0.1s;
}
.view-tab + .view-tab { border-left: 1px solid var(--border-default); }
.view-tab:hover { background: var(--surface-3); color: var(--text-primary); }
.view-tab.is-active { background: var(--brand-500); color: #fff; }

.loading, .error { text-align: center; padding: 48px 0; color: var(--text-tertiary); }
.error { color: var(--danger-500); }

/* 批量工具栏 */
.batch-bar {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 12px; margin-bottom: 12px;
  background: var(--brand-50); border: 1px solid var(--brand-200);
  border-radius: var(--radius-sm);
}
.batch-bar__info { font-size: 13px; font-weight: 500; color: var(--brand-600); }

.batch-select {
  padding: 4px 8px; font-size: 12px; font-family: inherit;
  background: var(--surface-1); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); color: var(--text-primary);
  cursor: pointer;
}
.batch-select:focus { border-color: var(--brand-500); outline: none; }

.btn--sm { padding: 4px 10px; font-size: 12px; font-family: inherit; border-radius: var(--radius-sm); cursor: pointer; border: 1px solid var(--border-default); }
.btn--danger { background: var(--danger-500); color: #fff; border: none; }
.btn--danger:hover { background: var(--danger-600); }
.btn--ghost { background: none; color: var(--text-secondary); }
.btn--export { background: var(--success-500); color: #fff; text-decoration: none; border: none; font-size: 12px; }
.btn--export:hover { background: var(--success-600); }

/* 导出下拉菜单 */
.export-dropdown { position: relative; }
.export-dropdown__menu {
  position: absolute; top: 100%; right: 0; margin-top: 4px;
  background: var(--surface-1); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  min-width: 160px; z-index: 100;
}
.export-dropdown__item {
  display: block; padding: 8px 14px; font-size: 13px;
  color: var(--text-primary); text-decoration: none;
  transition: background 0.1s;
}
.export-dropdown__item:hover { background: var(--surface-2); }
.export-dropdown__item + .export-dropdown__item { border-top: 1px solid var(--border-subtle); }
.btn--ghost:hover { background: var(--surface-3); }

/* 表格 */
.table-wrap { overflow-x: auto; }
.table { width: 100%; border-collapse: collapse; font-size: 13px; }
.table th, .table td { padding: 8px 10px; text-align: left; border-bottom: 1px solid var(--border-subtle); white-space: nowrap; }
.table th { font-size: 11px; font-weight: 600; color: var(--text-tertiary); text-transform: uppercase; }
.th--sortable { cursor: pointer; user-select: none; }
.th--sortable:hover { color: var(--text-primary); }
.sort-indicator { margin-left: 4px; }
.th-check { width: 40px; text-align: center; }
.td-check { width: 40px; text-align: center; }
.td-identifier { font-family: var(--font-mono); font-size: 12px; }
.td-name { min-width: 200px; word-break: break-all; white-space: normal; }
.td-severity { font-family: var(--font-mono); font-size: 11px; }
.td-num { font-family: var(--font-mono); font-size: 11px; color: var(--text-tertiary); }
.td-date { font-size: 11px; color: var(--text-tertiary); }
.td-assignees { display: flex; align-items: center; }
.text-muted { color: var(--text-tertiary); }

.row:hover { background: var(--surface-2); }
.row--selected { background: var(--brand-50); }

.row input[type="checkbox"] { cursor: pointer; }

.identifier-link, .name-link { cursor: pointer; }
.identifier-link:hover, .name-link:hover { color: var(--brand-600); }

.type-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; vertical-align: middle; }
.dot-requirement { background: var(--brand-500); }
.dot-task { background: var(--success-500); }
.dot-defect { background: var(--danger-500); }

.badge-sm { font-size: 11px; padding: 1px 6px; border-radius: 3px; font-weight: 500; }
.badge-sm.type-requirement { background: var(--brand-50); color: var(--brand-600); }
.badge-sm.type-task { background: var(--success-50); color: var(--success-600); }
.badge-sm.type-defect { background: var(--danger-50); color: var(--danger-600); }

.priority-text { font-size: 12px; }
.priority-text.pri-urgent { color: var(--danger-500); font-weight: 600; }
.priority-text.pri-high { color: var(--warning-500); }
.priority-text.pri-medium { color: var(--brand-500); }
.priority-text.pri-low, .priority-text.pri-none { color: var(--text-tertiary); }

.state-badge { font-size: 11px; padding: 1px 8px; border-radius: 10px; color: #fff; font-weight: 500; }

.empty-cell { text-align: center; padding: 32px 0; color: var(--text-tertiary); }

.avatar-placeholder {
  width: 22px; height: 22px; border-radius: 50%;
  background: var(--brand-100); color: var(--brand-600);
  font-size: 9px; font-weight: 600;
  display: inline-flex; align-items: center; justify-content: center;
  border: 2px solid var(--surface-1);
  margin-left: -4px;
}
.avatar-placeholder:first-child { margin-left: 0; }
.avatar-more {
  font-size: 9px; color: var(--text-tertiary); margin-left: 2px;
}

/* 分页器 */
.pagination {
  display: flex; align-items: center; justify-content: center; gap: 4px;
  padding: 16px 0;
}
.page-btn {
  padding: 4px 10px; font-size: 12px; font-family: inherit;
  color: var(--text-secondary); background: var(--surface-2);
  border: 1px solid var(--border-default); border-radius: var(--radius-sm);
  cursor: pointer; transition: all 0.1s;
}
.page-btn:hover:not(:disabled) { background: var(--surface-3); color: var(--text-primary); }
.page-btn--active { background: var(--brand-500); color: #fff; border-color: var(--brand-500); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-ellipsis { font-size: 12px; color: var(--text-tertiary); padding: 4px 4px; }
.page-info { font-size: 12px; color: var(--text-tertiary); margin-left: 12px; }

/* 弹窗 */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.3);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.modal-box { background: var(--surface-1); padding: 24px; border-radius: var(--radius-md); max-width: 400px; width: 90%; }
.modal-box h3 { margin: 0 0 8px; }
.modal-box p { font-size: 14px; color: var(--text-secondary); margin: 0 0 16px; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; }
</style>
