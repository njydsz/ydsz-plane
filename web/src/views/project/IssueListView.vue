<script setup lang="ts">
/**
 * 工作项列表页 — 表格视图展示工作项。
 * 支持: 服务端排序 / 分页 / 列过滤 / 批量选择与删除 / 跨页勾选 / 列配置持久化 / 批量指派+标签。
 */
import { computed, nextTick, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { type IssueType, type IssuePriority, type ListIssuesParams, type State, issueApi } from "@/api/services/issue";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { preferenceApi } from "@/api/services/preference";
import type { SavedView } from "@/api/services/preference";
import { useIssueStore } from "@/stores/issue";
import { usePeekStore } from "@/stores/peek";
import { prefs } from "@/lib/prefs";
import { toast } from "@/lib/toast";
import IssueFilter from "./IssueFilter.vue";
import ViewsManager from "./ViewsManager.vue";
import { AppErrorState, AppEmptyState, InlineEdit, InlineSelectEdit, AppSkeleton } from "@/components";
import { type FilterState, filterToListParams } from "@/lib/filter-adapter";
import ImportDialog from "@/components/ImportDialog.vue";

const props = withDefaults(defineProps<{
  /** 固定类型过滤 — 传入后列表仅展示该类型（requirement/task/defect），用户无法覆盖 */
  typeCode?: IssueType;
}>(), {
  typeCode: undefined,
});

const route = useRoute();
const issueStore = useIssueStore();
const peek = usePeekStore();

/** P1-3: 列表视图字段裁剪 — 默认仅取关键字段，展开/详情时再加载全量 */
const LIST_VIEW_FIELDS = ["id", "identifier", "name", "state_id", "priority", "type_code", "severity", "point", "assignees", "updated_at"];

/** P1-3: 虚拟滚动常量 */
const ROW_HEIGHT = 48; // px
const BUFFER_ROWS = 5; // 上下各缓冲行数

// ---- 状态 ----
const projectId = computed(() => Number(route.params.projectId));
const wsId = ref(0);
const loading = ref(true);
const error = ref("");

// 排序（服务端）
const sortField = ref<string>("-updated_at"); // 默认按更新时间倒序

// ========== 命名视图 ==========
const activeViewId = ref<number | null>(null);
const showViewsDropdown = ref(false);
const viewConfig = ref<Record<string, unknown>>({});

/** 更新当前视图配置快照 */
function captureViewConfig() {
  viewConfig.value = {
    filters: currentFilter.value,
    sort: sortField.value,
    columns: columnConfigs.value,
  };
}

/** 加载命名视图的配置 */
function onLoadView(view: SavedView) {
  const cfg = view.config || {};
  activeViewId.value = view.id;

  // 应用过滤
  if (cfg.filters) {
    currentFilter.value = cfg.filters as FilterState;
  }
  // 应用排序
  if (cfg.sort) {
    sortField.value = cfg.sort as string;
  }
  // 应用列配置
  if (cfg.columns && Array.isArray(cfg.columns)) {
    columnConfigs.value = defaultColumns.map((def) => {
      const savedCol = (cfg.columns as ColumnConfig[]).find((s) => s.key === def.key);
      return savedCol ? { ...def, ...savedCol } : def;
    });
  }

  page.value = 1;
  load();
  showViewsDropdown.value = false;
}

/** 保存当前视图前捕获配置 */
function onSaveView() {
  captureViewConfig();
}

// 分页
const page = ref(1);
const perPage = ref(50);
const total = computed(() => issueStore.total);

// 当前过滤参数（强类型 FilterState，对标 Plane 的多视图 FilterAdapter）
const currentFilter = ref<FilterState>({});

// ========== P1-2: 跨页勾选 ==========
// selectedIds 跨页保持；翻页不清零
const selectedIds = ref<Set<number>>(new Set());
// 跨页选中的数量：总数减去当前页选中的数量
const crossPageSelectedCount = computed(() => {
  if (selectedIds.value.size === 0) return 0;
  const currentPageIds = new Set(issueStore.issues.map((i) => i.id));
  let count = 0;
  for (const id of selectedIds.value) {
    if (!currentPageIds.has(id)) count++;
  }
  return count;
});
const totalSelectedCount = computed(() => selectedIds.value.size);

// ========== P1-2: 列配置 ==========
interface ColumnConfig {
  key: string;
  label: string;
  width?: string;
  sortable?: boolean;
  visible: boolean;
  pinned?: boolean;
}

const defaultColumns: ColumnConfig[] = [
  { key: "identifier", label: "编号", width: "120px", visible: true },
  { key: "name", label: "名称", sortable: true, visible: true },
  { key: "type_code", label: "类型", width: "72px", sortable: true, visible: true },
  { key: "priority", label: "优先级", width: "72px", sortable: true, visible: true },
  { key: "state", label: "状态", width: "90px", visible: true },
  { key: "severity", label: "严重度", width: "72px", sortable: true, visible: true },
  { key: "point", label: "点数", width: "60px", sortable: true, visible: true },
  { key: "assignees", label: "指派人", width: "100px", visible: true },
  { key: "updated_at", label: "更新时间", width: "130px", sortable: true, visible: true },
];

const columnConfigs = ref<ColumnConfig[]>(JSON.parse(JSON.stringify(defaultColumns)));
const showColumnConfigModal = ref(false);
const columnConfigDraft = ref<ColumnConfig[]>([]);
const savingColumnConfig = ref(false);

/** 可见列 */
const visibleColumns = computed(() => columnConfigs.value.filter((c) => c.visible));

/** 从偏好加载列配置 */
async function loadColumnConfig() {
  try {
    const pref = await preferenceApi.get(wsId.value, projectId.value, "list");
    if (pref?.columns && Array.isArray(pref.columns) && pref.columns.length > 0) {
      const saved = pref.columns as ColumnConfig[];
      // 合并：保留 saved 中已有的列的新增/变更，但确保列定义与 defaults 对齐
      columnConfigs.value = defaultColumns.map((def) => {
        const savedCol = saved.find((s) => s.key === def.key);
        return savedCol ? { ...def, ...savedCol } : def;
      });
    }
  } catch {
    /* 无偏好时使用默认列 */
  }
}

/** 保存列配置到偏好 */
async function saveColumnConfig(configs: ColumnConfig[]) {
  if (!wsId.value) return;
  savingColumnConfig.value = true;
  try {
    await preferenceApi.save(wsId.value, projectId.value, "list", {
      columns: configs,
    });
    columnConfigs.value = JSON.parse(JSON.stringify(configs));
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存列配置失败");
  } finally {
    savingColumnConfig.value = false;
  }
}

/** 打开列配置弹窗 */
function openColumnConfigModal() {
  columnConfigDraft.value = JSON.parse(JSON.stringify(columnConfigs.value));
  showColumnConfigModal.value = true;
}

/** 提交列配置 */
function applyColumnConfig() {
  saveColumnConfig(columnConfigDraft.value);
  showColumnConfigModal.value = false;
}

// ========== P1-2: Batch assign/tags ==========
const members = ref<Member[]>([]);
/** 标签列表 — 通过 project-level labels 获取（当前版本无独立标签 API，使用空列表占位，UI 保留扩展点） */
const labels = ref<{ id: number; name: string }[]>([]);
const showBatchAssign = ref(false);
const showBatchLabel = ref(false);
const batchAssignId = ref<number | null>(null);
const batchLabelId = ref<number | null>(null);
const batchOperating = ref(false);

async function loadMembers() {
  if (!wsId.value) return;
  try {
    members.value = await workspaceApi.listMembers(wsId.value);
  } catch {
    /* 静默失败 */
  }
}

async function loadLabels() {
  /* 未来对接标签 API 时在此加载 */
  labels.value = [];
}

/** 批量指派 */
async function batchAssign() {
  if (batchAssignId.value == null || selectedIds.value.size === 0) return;
  batchOperating.value = true;
  try {
    const r = await issueApi.batch(wsId.value, projectId.value, {
      issue_ids: [...selectedIds.value],
      assignee_id: batchAssignId.value,
    });
    if (r.failed > 0) error.value = `${r.succeeded} 项成功，${r.failed} 项失败`;
    else toast.success(`已指派 ${r.succeeded} 项`);
    showBatchAssign.value = false;
    batchAssignId.value = null;
    selectedIds.value = new Set();
    load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "批量指派失败";
  } finally {
    batchOperating.value = false;
  }
}

/** 批量添加标签 */
async function batchAddLabel() {
  if (batchLabelId.value == null || selectedIds.value.size === 0) return;
  batchOperating.value = true;
  try {
    // 批量标签: 使用 batch API 的 labels 字段（labels: number[]）
    const r = await issueApi.batch(wsId.value, projectId.value, {
      issue_ids: [...selectedIds.value],
      labels: [batchLabelId.value],
    } as any);
    if (r.failed > 0) error.value = `${r.succeeded} 项成功，${r.failed} 项失败`;
    else toast.success(`已添加标签到 ${r.succeeded} 项`);
    showBatchLabel.value = false;
    batchLabelId.value = null;
    selectedIds.value = new Set();
    load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "批量添加标签失败";
  } finally {
    batchOperating.value = false;
  }
}

// ---- 批量删除 ----
const showDeleteConfirm = ref(false);
const batchDeleting = ref(false);

// ---- 派生 ----
const typeLabel = (t: IssueType) =>
  ({ epic: "史诗", requirement: "需求", task: "任务", defect: "缺陷" } as Record<string, string>)[t] ?? t;

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
    const wsIdVal = Number(route.params.workspaceId);

    // FilterState → ListParams 转换（统一在 API 边界）
    const filterParams = filterToListParams(currentFilter.value);
    // 如果指定了 typeCode prop，强制覆盖过滤类型（需求/任务/缺陷独立视图）
    if (props.typeCode) {
      filterParams.type = props.typeCode;
    }
    const params: ListIssuesParams = {
      ...filterParams,
      sort: buildSortParam(),
      limit: perPage.value,
      offset: (page.value - 1) * perPage.value,
    };

    // 使用 issueApi.listIssues 直接调用以支持字段裁剪（stores 层不支持 fields 参数）
    const fetchIssuesWithFields = async () => {
      issueStore.loading = true;
      issueStore.error = null;
      try {
        const res = await issueApi.listIssues(wsIdVal, projectId.value, params, LIST_VIEW_FIELDS);
        issueStore.issues = res.results;
        issueStore.total = res.total;
      } catch (e: unknown) {
        issueStore.error = e instanceof Error ? e.message : "加载失败";
        throw e;
      } finally {
        issueStore.loading = false;
      }
    };

    await Promise.all([
      issueStore.fetchStates(wsIdVal, projectId.value),
      fetchIssuesWithFields(),
      loadColumnConfig(),
    ]);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function onFilterChange(f: FilterState) {
  currentFilter.value = f;
  page.value = 1; // 过滤条件变化时回到第一页
  load();
  // 服务端持久化由 IssueFilter 内部处理，此处不再重复
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
  // 传入当前列表可见的 issue ID 序列，支持抽屉内上一个/下一个连续预览
  peek.openWithContext(
    Number(route.params.workspaceId),
    projectId.value,
    issueId,
    issueStore.issues.map((i) => i.id),
  );
}

function toggleSelect(issueId: number) {
  const next = new Set(selectedIds.value);
  if (next.has(issueId)) next.delete(issueId);
  else next.add(issueId);
  selectedIds.value = next;
}

function toggleSelectAll() {
  const currentPageIds = issueStore.issues.map((i) => i.id);
  if (currentPageIds.every((id) => selectedIds.value.has(id))) {
    // 当前页全部选中 → 取消当前页
    const next = new Set(selectedIds.value);
    for (const id of currentPageIds) next.delete(id);
    selectedIds.value = next;
  } else {
    // 选中当前页全部
    const next = new Set(selectedIds.value);
    for (const id of currentPageIds) next.add(id);
    selectedIds.value = next;
  }
}

/** 清空所有选择 */
function clearSelection() {
  selectedIds.value = new Set();
}

/** 选择全部匹配项（通过后端的 listIssues 取全量 ID） */
async function selectAllMatching() {
  try {
    loading.value = true;
    const filterParams = filterToListParams(currentFilter.value);
    // 用一个较大的 limit 取全量 ID（上限提示）
    const SAFE_LIMIT = 5000;
    const params: ListIssuesParams = {
      ...filterParams,
      sort: buildSortParam(),
      limit: SAFE_LIMIT,
      offset: 0,
    };
    const result = await issueApi.listIssues(wsId.value, projectId.value, params);
    if (result.total > SAFE_LIMIT) {
      toast.warning(`匹配项超过 ${SAFE_LIMIT} 条，仅选择前 ${SAFE_LIMIT} 项。建议缩小过滤范围。`);
    }
    selectedIds.value = new Set(result.results.map((i) => i.id));
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "选择全部失败";
  } finally {
    loading.value = false;
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

/** 内联更新：更新单个字段并同步本地状态（乐观更新 + 失败回滚） */
async function inlineUpdate(iss: any, patch: Record<string, unknown>) {
  const prev = { ...iss };
  // 乐观更新
  const idx = issueStore.issues.findIndex((i) => i.id === iss.id);
  if (idx >= 0) issueStore.issues[idx] = { ...issueStore.issues[idx], ...patch };
  try {
    const updated = await issueApi.updateIssue(wsId.value, projectId.value, iss.id, {
      ...patch,
      version: iss.version,
    } as Parameters<typeof issueApi.updateIssue>[3]);
    if (idx >= 0) issueStore.issues[idx] = updated;
  } catch (e: unknown) {
    if (idx >= 0) issueStore.issues[idx] = prev;
    const msg = e instanceof Error ? e.message : "更新失败";
    toast.error(msg);
  }
}

/**
 * 处理工作项状态变更，调用专用状态流转接口，适配后端状态机规则
 */
async function handleStateChange(issueId: number, toStateId: number) {
  try {
    const updated = await issueApi.transition(wsId.value, projectId.value, issueId, toStateId);
    // 更新列表中的工作项状态
    const idx = issueStore.issues.findIndex((i) => i.id === issueId);
    if (idx >= 0) issueStore.issues[idx] = updated;
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "状态流转失败";
    toast.error(msg);
    // 流转失败时刷新列表，恢复正确状态
    await issueStore.fetchIssues(wsId.value, projectId.value);
  }
}

const priorityOptions: { value: IssuePriority; label: string; color: string; icon: string }[] = [
  { value: "urgent", label: "紧急", color: "var(--danger-500)", icon: "🔴" },
  { value: "high", label: "高", color: "var(--warning-500)", icon: "🟠" },
  { value: "medium", label: "中", color: "var(--brand-500)", icon: "🔵" },
  { value: "low", label: "低", color: "var(--text-tertiary)", icon: "⚪" },
  { value: "none", label: "无", color: "var(--text-tertiary)", icon: "⬜" },
];

// 列头右键菜单
const showColumnContextMenu = ref(false);
const contextMenuColKey = ref("");

function onColumnHeaderContext(colKey: string, event: MouseEvent) {
  event.preventDefault();
  contextMenuColKey.value = colKey;
  showColumnContextMenu.value = true;
}

function closeContextMenu() {
  showColumnContextMenu.value = false;
}

onMounted(() => {
  prefs.setLastView(projectId.value, "list");
  load();
  loadMembers();
  loadLabels();
  nextTick(() => measureContainer());
});

const showImportDialog = ref(false);

function onImported() {
  load();
}

const exportCsvUrl = computed(() =>
  issueApi.exportUrl(wsId.value, projectId.value, filterToListParams(currentFilter.value), "csv"),
);

const exportXlsxUrl = computed(() =>
  issueApi.exportUrl(wsId.value, projectId.value, filterToListParams(currentFilter.value), "xlsx"),
);

/** 导出格式下拉是否展开 */
const showExportDropdown = ref(false);

// ========== P1-3: 虚拟滚动 ==========
const scrollContainerRef = ref<HTMLElement | null>(null);
const scrollTop = ref(0);
const containerHeight = ref(600);

/** 可见行范围 */
const visibleRange = computed(() => {
  const total = issueStore.issues.length;
  if (total === 0) return { start: 0, end: 0 };
  const start = Math.max(0, Math.floor(scrollTop.value / ROW_HEIGHT) - BUFFER_ROWS);
  const visibleCount = Math.ceil(containerHeight.value / ROW_HEIGHT);
  const end = Math.min(total, start + visibleCount + BUFFER_ROWS * 2);
  return { start, end };
});

/** 当前可视行 */
const visibleIssues = computed(() => {
  return issueStore.issues.slice(visibleRange.value.start, visibleRange.value.end);
});

/** 滚动占位总高度 */
const scrollSpacerHeight = computed(() => issueStore.issues.length * ROW_HEIGHT);

/** 滚动偏移 */
const scrollOffsetTop = computed(() => visibleRange.value.start * ROW_HEIGHT);

function onTableScroll(e: Event) {
  scrollTop.value = (e.target as HTMLElement).scrollTop;
}

/** 测量容器高度 */
function measureContainer() {
  if (scrollContainerRef.value) {
    containerHeight.value = scrollContainerRef.value.clientHeight;
  }
}

/** 当前页 checkbox 全选状态 */
const isCurrentPageAllSelected = computed(() => {
  if (issueStore.issues.length === 0) return false;
  return issueStore.issues.every((i) => selectedIds.value.has(i.id));
});

</script>

<template>
  <div class="list-view">
    <header class="list-view__header">
      <div>
        <h1>列表</h1>
        <p class="hint">共 {{ total }} 个工作项</p>
      </div>
      <div class="list-view__header-right">
        <div class="view-dropdown">
          <button
            class="btn btn--sm btn--view"
            @click="showViewsDropdown = !showViewsDropdown"
          >
            视图
          </button>
          <div v-if="showViewsDropdown" class="view-dropdown__panel">
            <ViewsManager
              :workspace-id="Number(route.params.workspaceId)"
              :project-id="projectId"
              view-type="list"
              :current-config="viewConfig"
              :active-view-id="activeViewId"
              @load-view="onLoadView"
              @save-view="onSaveView"
            />
          </div>
        </div>
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
        <button
          class="btn btn--sm btn--import"
          @click="showImportDialog = true"
        >
          导入
        </button>
        <div class="view-switcher">
          <router-link
            :to="`/${route.params.workspaceId}/projects/${projectId}/board`"
            class="view-tab"
          >
看板
</router-link>
          <router-link
            :to="`/${route.params.workspaceId}/projects/${projectId}/list`"
            class="view-tab is-active"
          >
列表
</router-link>
        </div>
      </div>
    </header>

    <!-- 过滤器（使用 FilterAdapter，view_type="list"） -->
    <IssueFilter
      :project-id="projectId"
      :workspace-id="Number(route.params.workspaceId)"
      view-type="list"
      @filter-change="onFilterChange"
    />

    <!-- P1-2: 跨页批量操作工具栏 -->
    <div v-if="hasSelection" class="batch-bar">
      <span class="batch-bar__info">
        已选 <strong>{{ totalSelectedCount }}</strong> 项
        <template v-if="crossPageSelectedCount > 0">（含跨页 {{ crossPageSelectedCount }} 项）</template>
      </span>
      <select class="batch-select" @change="(e: Event) => { const v = Number((e.target as HTMLSelectElement).value); if (v) batchTransition(v); (e.target as HTMLSelectElement).value = '' }">
        <option value="">批量流转...</option>
        <option v-for="st in issueStore.states" :key="st.id" :value="st.id">{{ st.name }}</option>
      </select>
      <select class="batch-select" @change="(e: Event) => { const v = (e.target as HTMLSelectElement).value; if (v) batchUpdatePriority(v); (e.target as HTMLSelectElement).value = '' }">
        <option value="">批量优先级...</option>
        <option value="urgent">紧急</option>
        <option value="high">高</option>
        <option value="medium">中</option>
        <option value="low">低</option>
        <option value="none">无</option>
      </select>

      <!-- 批量指派 -->
      <div class="batch-inline-dropdown">
        <button class="btn btn--sm" :class="{ 'btn--active': showBatchAssign }" @click="showBatchAssign = !showBatchAssign">
          批量指派
        </button>
        <div v-if="showBatchAssign" class="batch-inline-panel">
          <select v-model="batchAssignId" class="batch-select">
            <option :value="null" disabled>选择成员...</option>
            <option v-for="m in members" :key="m.id" :value="m.id">{{ m.display_name || m.email }}</option>
          </select>
          <button class="btn btn--sm btn--primary" :disabled="batchAssignId == null || batchOperating" @click="batchAssign">
            {{ batchOperating ? '处理中...' : '确认' }}
          </button>
        </div>
      </div>

      <!-- 批量标签 -->
      <div class="batch-inline-dropdown">
        <button class="btn btn--sm" :class="{ 'btn--active': showBatchLabel }" @click="showBatchLabel = !showBatchLabel">
          批量标签
        </button>
        <div v-if="showBatchLabel" class="batch-inline-panel">
          <select v-model="batchLabelId" class="batch-select">
            <option :value="null" disabled>选择标签...</option>
            <option v-for="l in labels" :key="l.id" :value="l.id">{{ l.name }}</option>
          </select>
          <button class="btn btn--sm btn--primary" :disabled="batchLabelId == null || batchOperating" @click="batchAddLabel">
            {{ batchOperating ? '处理中...' : '确认' }}
          </button>
        </div>
      </div>

      <button class="btn btn--sm btn--danger" @click="showDeleteConfirm = true">批量删除</button>
      <button class="btn btn--sm btn--ghost" @click="clearSelection">清空选择</button>
      <button class="btn btn--sm btn--ghost" @click="selectAllMatching">选择全部匹配项</button>
    </div>

    <!-- 列头右键菜单 -->
    <div v-if="showColumnContextMenu" class="context-menu-overlay" @click="closeContextMenu">
      <div class="context-menu" :style="{ left: '50%', top: '120px' }" @click.stop>
        <button class="context-menu__item" @click="openColumnConfigModal(); closeContextMenu()">
          配置列...
        </button>
      </div>
    </div>

    <AppSkeleton v-if="loading" variant="table" :rows="8" />
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
                :checked="isCurrentPageAllSelected"
                @change="toggleSelectAll"
              />
            </th>
            <th
              v-for="col in visibleColumns"
              :key="col.key"
              :style="col.width ? { width: col.width, minWidth: col.width } : {}"
              :class="{ 'th--sortable': col.sortable }"
              @click="col.sortable && toggleSort(col.key)"
              @contextmenu="onColumnHeaderContext(col.key, $event)"
            >
              {{ col.label }}<span class="sort-indicator">{{ sortIndicator(col.key) }}</span>
            </th>
          </tr>
        </thead>
      </table>
      <!-- 虚拟滚动容器 -->
      <div
        ref="scrollContainerRef"
        class="virtual-scroll-container"
        @scroll="onTableScroll"
      >
        <div class="virtual-scroll-spacer" :style="{ height: scrollSpacerHeight + 'px' }">
          <table class="table table--virtual">
            <colgroup>
              <col style="width: 40px" />
              <col
                v-for="col in visibleColumns"
                :key="col.key"
                :style="col.width ? { width: col.width, minWidth: col.width } : {}"
              />
            </colgroup>
            <tbody :style="{ transform: 'translateY(' + scrollOffsetTop + 'px)' }">
              <tr
                v-for="iss in visibleIssues"
                :key="iss.id"
                class="row"
                :class="{ 'row--selected': selectedIds.has(iss.id) }"
                :style="{ height: ROW_HEIGHT + 'px' }"
              >
            <td class="td-check" @click.stop>
              <input
                type="checkbox"
                :checked="selectedIds.has(iss.id)"
                @change="toggleSelect(iss.id)"
              />
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'identifier')" class="td-identifier">
              <span class="identifier-link" @click="openIssue(iss.id)">
                {{ iss.identifier }}
              </span>
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'name')" class="td-name">
              <span class="name-link" @click="openIssue(iss.id)">
                <span class="type-dot" :class="`dot-${iss.type_code}`"></span>
              </span>
              <InlineEdit
                :model-value="iss.name"
                placeholder="未命名"
                :max-length="200"
                @submit="(v) => inlineUpdate(iss, { name: v })"
              />
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'type_code')">
              <span class="badge-sm" :class="`type-${iss.type_code}`">
                {{ typeLabel(iss.type_code) }}
              </span>
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'priority')">
              <InlineSelectEdit
                :model-value="iss.priority"
                :options="priorityOptions"
                placeholder="无"
                @submit="(v) => inlineUpdate(iss, { priority: v as IssuePriority })"
              >
                <template #trigger>
                  <span class="priority-text" :class="`pri-${iss.priority}`">
                    {{ priorityLabel(iss.priority) }}
                  </span>
                </template>
              </InlineSelectEdit>
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'state')">
              <InlineSelectEdit
                :model-value="iss.state_id"
                :options="issueStore.states.map((s) => ({ value: s.id, label: s.name, color: s.color }))"
                placeholder="未设置状态"
                @submit="(v) => handleStateChange(iss.id, Number(v))"
              >
                <template #trigger>
                  <span
                    class="state-badge"
                    :style="{ backgroundColor: stateMap[iss.state_id]?.color ?? '#ccc' }"
                  >
                    {{ stateMap[iss.state_id]?.name ?? iss.state_id }}
                  </span>
                </template>
              </InlineSelectEdit>
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'severity')" class="td-severity">
              {{ severityText(iss.severity) }}
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'point')" class="td-num">
              {{ iss.point != null ? iss.point + "pt" : "-" }}
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'assignees')" class="td-assignees">
              <span v-if="iss.assignees?.length > 0">
                <span v-for="uid in iss.assignees.slice(0, 3)" :key="uid" class="avatar-placeholder">
                  U{{ uid }}
                </span>
                <span v-if="iss.assignees.length > 3" class="avatar-more">+{{ iss.assignees.length - 3 }}</span>
              </span>
              <span v-else class="text-muted">-</span>
            </td>
            <td v-if="visibleColumns.some(c => c.key === 'updated_at')" class="td-date">
              {{ new Date(iss.updated_at).toLocaleDateString("zh-CN") }}
            </td>
          </tr>
            <tr v-if="visibleIssues.length === 0">
              <td :colspan="visibleColumns.length + 1" class="empty-cell">
                暂无工作项
              </td>
            </tr>
          </tbody>
          </table>
        </div>
      </div>

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

    <!-- P1-2: 列配置弹窗 -->
    <div v-if="showColumnConfigModal" class="modal-overlay" @click.self="showColumnConfigModal = false">
      <div class="column-config-modal">
        <div class="column-config-modal__header">
          <h3>配置列</h3>
          <button class="btn btn--ghost btn--sm" @click="showColumnConfigModal = false">×</button>
        </div>
        <p class="column-config-modal__hint">拖拽排序 / 勾选显示 / 输入宽度</p>
        <ul class="column-config-list">
          <li
            v-for="col in columnConfigDraft"
            :key="col.key"
            class="column-config-item"
          >
            <span class="column-config-item__drag">⋮⋮</span>
            <label class="column-config-item__check">
              <input v-model="col.visible" type="checkbox" />
              <span>{{ col.label }}</span>
            </label>
            <input
              v-if="col.width !== undefined"
              class="column-config-item__width"
              type="text"
              placeholder="auto"
              :value="col.width"
              @input="(e: Event) => { col.width = (e.target as HTMLInputElement).value }"
            />
            <span v-else class="column-config-item__width column-config-item__width--auto">auto</span>
          </li>
        </ul>
        <div class="column-config-modal__actions">
          <button class="btn btn--ghost" @click="showColumnConfigModal = false">取消</button>
          <button class="btn btn--primary" :disabled="savingColumnConfig" @click="applyColumnConfig">
            {{ savingColumnConfig ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </div>

    <!-- CSV/XLSX 导入弹窗 -->
    <ImportDialog
      :visible="showImportDialog"
      :ws-id="Number(route.params.workspaceId)"
      :project-id="projectId"
      @close="showImportDialog = false"
      @imported="onImported"
    />
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
.view-tab.is-active { background: var(--brand-500); color: var(--text-on-brand); }

.loading, .error { text-align: center; padding: 48px 0; color: var(--text-tertiary); }
.error { color: var(--danger-500); }

/* 批量工具栏 */
.batch-bar {
  display: flex; flex-wrap: wrap; align-items: center; gap: 10px;
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

/* 行内下拉批量操作 */
.batch-inline-dropdown {
  position: relative;
}

.batch-inline-panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-popover);
}

.btn--sm { padding: 4px 10px; font-size: 12px; font-family: inherit; border-radius: var(--radius-sm); cursor: pointer; border: 1px solid var(--border-default); background: var(--surface-1); color: var(--text-primary); }
.btn--danger { background: var(--danger-500); color: var(--text-on-brand); border: none; }
.btn--danger:hover { background: var(--danger-600); }
.btn--ghost { background: none; color: var(--text-secondary); border-color: transparent; }
.btn--ghost:hover { background: var(--surface-3); }
.btn--primary { background: var(--brand-500); color: var(--text-on-brand); border-color: var(--brand-500); }
.btn--primary:hover:not(:disabled) { background: var(--brand-600); }
.btn--primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn--active { background: var(--brand-100); color: var(--brand-600); border-color: var(--brand-200); }
.btn--export { background: var(--success-500); color: var(--text-on-brand); text-decoration: none; border: none; font-size: 12px; }
.btn--export:hover { background: var(--success-600); }
.btn--import { background: var(--surface-1); color: var(--text-primary); border: 1px solid var(--border-default); font-size: 12px; }
.btn--import:hover { background: var(--surface-3); }
.btn--view { background: var(--brand-500); color: var(--text-on-brand); text-decoration: none; border: none; font-size: 12px; }
.btn--view:hover { background: var(--brand-600); }

/* 视图下拉面板 */
.view-dropdown {
  position: relative;
}
.view-dropdown__panel {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 300px;
  max-height: 480px;
  overflow-y: auto;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
  padding: 12px;
  z-index: 200;
}

/* 导出下拉菜单 */
.export-dropdown { position: relative; }
.export-dropdown__menu {
  position: absolute; top: 100%; right: 0; margin-top: 4px;
  background: var(--surface-1); border: 1px solid var(--border-default);
  border-radius: var(--radius-sm); box-shadow: var(--shadow-popover);
  min-width: 160px; z-index: 100;
}
.export-dropdown__item {
  display: block; padding: 8px 14px; font-size: 13px;
  color: var(--text-primary); text-decoration: none;
  transition: background 0.1s;
}
.export-dropdown__item:hover { background: var(--surface-2); }
.export-dropdown__item + .export-dropdown__item { border-top: 1px solid var(--border-subtle); }

/* 表格 */
.table-wrap { overflow-x: auto; display: flex; flex-direction: column; }
.table { width: 100%; border-collapse: collapse; font-size: 13px; }
.table--virtual { table-layout: fixed; }

/* P1-3: 虚拟滚动 */
.virtual-scroll-container {
  overflow-y: auto;
  flex: 1;
  position: relative;
  max-height: calc(100vh - 280px); /* 视口高度 - 顶部工具栏高度估算 */
}
.virtual-scroll-spacer {
  position: relative;
  width: 100%;
}
.table--virtual {
  position: absolute;
  top: 0;
  left: 0;
}
.table--virtual tbody {
  position: absolute;
  width: 100%;
}
.table--virtual .row {
  height: 48px;
}
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

.state-badge { font-size: 11px; padding: 1px 8px; border-radius: 10px; color: var(--text-on-brand); font-weight: 500; }

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
.page-btn--active { background: var(--brand-500); color: var(--text-on-brand); border-color: var(--brand-500); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-ellipsis { font-size: 12px; color: var(--text-tertiary); padding: 4px 4px; }
.page-info { font-size: 12px; color: var(--text-tertiary); margin-left: 12px; }

/* 弹窗 */
.modal-overlay {
  position: fixed; inset: 0; background: var(--bg-backdrop);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.modal-box { background: var(--surface-1); padding: 24px; border-radius: var(--radius-md); max-width: 400px; width: 90%; }
.modal-box h3 { margin: 0 0 8px; }
.modal-box p { font-size: 14px; color: var(--text-secondary); margin: 0 0 16px; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; }

/* 上下文菜单 */
.context-menu-overlay {
  position: fixed; inset: 0; z-index: 999;
}
.context-menu {
  position: fixed;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-popover);
  min-width: 150px;
  z-index: 1000;
  padding: 4px 0;
}
.context-menu__item {
  display: block;
  width: 100%;
  padding: 8px 14px;
  text-align: left;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: none;
  border: none;
  cursor: pointer;
}
.context-menu__item:hover { background: var(--surface-2); }

/* 列配置弹窗 */
.column-config-modal {
  background: var(--surface-1);
  border-radius: var(--radius-md);
  width: 480px;
  max-width: 90vw;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  padding: 24px;
}
.column-config-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.column-config-modal__header h3 { margin: 0; font-size: 16px; }
.column-config-modal__hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}
.column-config-list {
  list-style: none;
  margin: 0;
  padding: 0;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.column-config-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  cursor: move;
}
.column-config-item__drag {
  color: var(--text-tertiary);
  font-size: 14px;
  flex-shrink: 0;
  cursor: grab;
}
.column-config-item__check {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  font-size: 13px;
  cursor: pointer;
}
.column-config-item__check input { accent-color: var(--brand-500); }
.column-config-item__width {
  width: 60px;
  padding: 3px 6px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  text-align: center;
}
.column-config-item__width:focus { border-color: var(--brand-500); }
.column-config-item__width--auto {
  width: 60px;
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary);
}
.column-config-modal__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

/* ===== S13 P1: 移动端响应式 ===== */
@media (max-width: 768px) {
  .table-wrap {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }
  .table {
    min-width: 100%;
    font-size: 13px;
  }
  .table th,
  .table td {
    padding: 8px 6px;
    min-height: 44px;
  }
  /* 小屏隐藏非核心列 */
  .table th:nth-child(n+5),
  .table td:nth-child(n+5) {
    display: none;
  }
  .issue-key {
    font-size: 11px;
  }
  /* 触控友好 */
  .table button,
  .table .clickable {
    min-width: 44px;
    min-height: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  /* 列表操作条 */
  .list-toolbar,
  .bulk-actions {
    flex-wrap: wrap;
    gap: 6px;
    padding: 8px;
  }
}

</style>
