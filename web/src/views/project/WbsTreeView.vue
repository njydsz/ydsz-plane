<script setup lang="ts">
/**
 * WbsTreeView — WBS 树形视图页。
 *
 * 以递归树形结构展示工作项的三级 WBS 层级：
 *   - 需求：Epic → Feature → Story
 *   - 任务：主任务 → 子任务 → 子子任务
 *   - 缺陷：主缺陷 → 子缺陷
 *
 * 特性：
 *   - 按工作项类型分组展示（需求组/任务组/缺陷组）
 *   - 每类型组内以 parent_id 构建递归树
 *   - 展开/折叠、行内重命名、新建子工作项
 *   - 搜索过滤（按名称或编号）
 *   - 进度自动汇总（父 = sum(子完成故事点)/sum(子总故事点)）
 *   - 跳转到工作项详情
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { issueApi, type Issue, type IssueType } from "@/api/services/issue";
import { usePeekStore } from "@/stores/peek";
import { toast } from "@/lib/toast";
import { AppEmptyState, AppErrorState, AppLoadingState } from "@/components";
import WbsTreeNode from "./WbsTreeNode.vue";
import IssueCreateModal from "./IssueCreateModal.vue";

const route = useRoute();
const peek = usePeekStore();

const projectId = computed(() => Number(route.params.projectId));
const wsId = computed(() => Number(route.params.workspaceId));

// ---- 状态 ----
const loading = ref(true);
const error = ref("");
const issues = ref<Issue[]>([]);

// 折叠状态：collapsed Set 中的节点 ID 为"已折叠"
const collapsedIssues = ref<Set<number>>(new Set());

// 当前选中的工作项 ID（用于高亮）
const selectedIssueId = ref<number | null>(null);

// 搜索关键词
const searchQuery = ref("");
const groupIdFilter = ref<string | null>(null);

// 新建子工作项弹窗
const showCreateModal = ref(false);
const createParentId = ref<number | null>(null);

// ---- 数据获取 ----
async function load() {
  loading.value = true;
  error.value = "";
  try {
    // 拉取全量工作项（WBS 视图需完整树结构）
    // 用较大 limit 覆盖典型项目规模；超大型项目后续可优化为分页加载根 + 懒加载子
    const res = await issueApi.listIssues(wsId.value, projectId.value, {
      limit: 1000,
      sort: "sort_order",
    });
    issues.value = res.results || [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载工作项失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch([wsId, projectId], load);

// ---- 树构建 ----
interface TypeGroup {
  type: IssueType;
  label: string;
  icon: string;
  roots: WbsNodeGroup[];
}

interface WbsNodeGroup {
  issue: Issue;
  children: WbsNodeGroup[];
}

/**
 * 从扁平 Issue 列表构建树。
 * 算法：O(n) 哈希表 + 单次遍历。
 */
function buildTree(issues: Issue[]): WbsNodeGroup[] {
  const nodeMap = new Map<number, WbsNodeGroup>();
  const roots: WbsNodeGroup[] = [];

  // 第一遍：创建所有节点
  for (const iss of issues) {
    nodeMap.set(iss.id, { issue: iss, children: [] });
  }
  // 第二遍：挂父子
  for (const iss of issues) {
    const node = nodeMap.get(iss.id)!;
    if (iss.parent_id && nodeMap.has(iss.parent_id)) {
      nodeMap.get(iss.parent_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  }
  // 按 sort_order 稳定排序
  const sortNodes = (nodes: WbsNodeGroup[]) => {
    nodes.sort((a, b) => (a.issue.sort_order ?? a.issue.id) - (b.issue.sort_order ?? b.issue.id));
    nodes.forEach((n) => sortNodes(n.children));
  };
  sortNodes(roots);
  return roots;
}

/** 按 type_code 分组后各自构建树 */
const typeGroups = computed<TypeGroup[]>(() => {
  const query = searchQuery.value.trim().toLowerCase();
  const filtered = issues.value.filter((iss) => {
    const matchGroup = !groupIdFilter.value || iss.type_code === groupIdFilter.value;
    const matchSearch = !query ||
      iss.name.toLowerCase().includes(query) ||
      iss.identifier.toLowerCase().includes(query);
    return matchGroup && matchSearch;
  });

  const typeOrder: IssueType[] = ["epic", "requirement", "task", "defect"];
  const typeMeta: Record<IssueType, { label: string; icon: string }> = {
    epic: { label: "史诗需求", icon: "🎯" },
    requirement: { label: "需求", icon: "📋" },
    task: { label: "任务", icon: "✅" },
    defect: { label: "缺陷", icon: "🐛" },
  };

  return typeOrder
    .filter((t) => filtered.some((iss) => iss.type_code === t))
    .map((t) => ({
      type: t,
      label: typeMeta[t].label,
      icon: typeMeta[t].icon,
      roots: buildTree(filtered.filter((iss) => iss.type_code === t)),
    }));
});

// ---- 折叠/展开 ----
function toggleCollapse(issueId: number) {
  const s = collapsedIssues.value;
  if (s.has(issueId)) {
    s.delete(issueId);
  } else {
    s.add(issueId);
  }
  // 触发响应（Set 不是 deep reactive）
  collapsedIssues.value = new Set(s);
}

function expandAll() {
  collapsedIssues.value = new Set();
}

function collapseAll() {
  const all = new Set<number>();
  issues.value.forEach((iss) => {
    // 只有有子项的节点才需要折叠
    if (issues.value.some((c) => c.parent_id === iss.id)) {
      all.add(iss.id);
    }
  });
  collapsedIssues.value = all;
}

// ---- 操作 ----
function onSelectIssue(issueId: number) {
  selectedIssueId.value = issueId;
  peek.open(wsId.value, projectId.value, issueId);
}

function onCreateChild(parentId: number) {
  createParentId.value = parentId;
  showCreateModal.value = true;
}

async function onRename(issue: Issue, name: string) {
  try {
    await issueApi.updateIssue(wsId.value, projectId.value, issue.id, {
      name,
      version: issue.version,
    });
    issue.name = name;
    toast.success("重命名成功");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "重命名失败");
  }
}

async function onDelete(issue: Issue) {
  const hasChildren = issues.value.some((c) => c.parent_id === issue.id);
  const msg = hasChildren
    ? `确认删除 "${issue.identifier}" 及其所有子工作项？`
    : `确认删除 "${issue.identifier}"？`;
  if (!confirm(msg)) return;
  try {
    await issueApi.batch(wsId.value, projectId.value, {
      issue_ids: [issue.id],
      delete: true,
    });
    // 本地移除
    issues.value = issues.value.filter((i) => i.id !== issue.id);
    toast.success("删除成功");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

function onCreated(issueId: number) {
  // 新创建的工作项加入列表后重新加载以刷新树
  void issueId;
  showCreateModal.value = false;
  load();
}

// ---- 统计 ----
const totalCount = computed(() => issues.value.length);
const rootCount = computed(() => issues.value.filter((i) => !i.parent_id).length);
</script>

<template>
  <div class="wbs-view">
    <!-- 工具栏 -->
    <div class="wbs-toolbar">
      <div class="wbs-toolbar__left">
        <h2 class="wbs-title">WBS 树形视图</h2>
        <span class="wbs-meta">{{ totalCount }} 个工作项 · {{ rootCount }} 个根节点</span>
      </div>
      <div class="wbs-toolbar__right">
        <!-- 搜索 -->
        <input
          v-model="searchQuery"
          class="wbs-search"
          type="text"
          placeholder="🔍 搜索名称或编号..."
        />
        <!-- 类型过滤 -->
        <div class="wbs-filter-group">
          <button
            class="wbs-filter-btn"
            :class="{ 'wbs-filter-btn--active': !groupIdFilter }"
            @click="groupIdFilter = null"
          >
全部
</button>
          <button
            class="wbs-filter-btn"
            :class="{ 'wbs-filter-btn--active': groupIdFilter === 'epic' }"
            @click="groupIdFilter = 'epic'"
          >
🎯 史诗
</button>
          <button
            class="wbs-filter-btn"
            :class="{ 'wbs-filter-btn--active': groupIdFilter === 'requirement' }"
            @click="groupIdFilter = 'requirement'"
          >
📋 需求
</button>
          <button
            class="wbs-filter-btn"
            :class="{ 'wbs-filter-btn--active': groupIdFilter === 'task' }"
            @click="groupIdFilter = 'task'"
          >
✅ 任务
</button>
          <button
            class="wbs-filter-btn"
            :class="{ 'wbs-filter-btn--active': groupIdFilter === 'defect' }"
            @click="groupIdFilter = 'defect'"
          >
🐛 缺陷
</button>
        </div>
        <!-- 展开/折叠 -->
        <button class="wbs-btn" title="全部展开" @click="expandAll">⊞</button>
        <button class="wbs-btn" title="全部折叠" @click="collapseAll">⊟</button>
      </div>
    </div>

    <!-- 内容 -->
    <div class="wbs-content">
      <!-- 加载态 -->
      <AppLoadingState v-if="loading" message="正在加载工作项..." />

      <!-- 错误态 -->
      <AppErrorState v-else-if="error" :message="error" @retry="load" />

      <!-- 空态 -->
      <AppEmptyState
        v-else-if="typeGroups.length === 0"
        icon="🌳"
        title="暂无工作项"
        description="在列表中创建工作项，即可在此查看 WBS 树形视图。"
      />

      <!-- 类型分组 -->
      <div v-else class="wbs-groups">
        <section
          v-for="group in typeGroups"
          :key="group.type"
          class="wbs-group"
        >
          <header class="wbs-group-header">
            <span>{{ group.icon }} {{ group.label }}</span>
            <span class="wbs-group-count">{{ group.roots.length }} 个根</span>
          </header>

          <ul v-if="group.roots.length" class="wbs-tree">
            <WbsTreeNode
              v-for="node in group.roots"
              :key="node.issue.id"
              :node="node"
              :collapsed="collapsedIssues"
              :selected-id="selectedIssueId"
              @select="onSelectIssue"
              @toggle="toggleCollapse"
              @create-child="onCreateChild"
              @rename="onRename"
              @delete="onDelete"
            />
          </ul>

          <p v-else class="wbs-group-empty">该分类下暂无工作项</p>
        </section>
      </div>
    </div>

    <!-- 新建子工作项弹窗 -->
    <IssueCreateModal
      v-if="showCreateModal"
      :visible="showCreateModal"
      :workspace-id="wsId"
      :project-id="projectId"
      :parent-id="createParentId ?? undefined"
      @close="showCreateModal = false"
      @created="onCreated"
    />
  </div>
</template>

<style scoped>
.wbs-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.wbs-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  gap: 12px;
  flex-shrink: 0;
}

.wbs-toolbar__left {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.wbs-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.wbs-meta {
  font-size: 12px;
  color: var(--text-tertiary);
}

.wbs-toolbar__right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wbs-search {
  width: 200px;
  padding: 5px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--surface-1);
  color: var(--text-primary);
}

.wbs-search:focus {
  outline: none;
  border-color: var(--brand-400);
  box-shadow: 0 0 0 2px var(--brand-100);
}

.wbs-filter-group {
  display: flex;
  gap: 2px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  padding: 2px;
}

.wbs-filter-btn {
  border: none;
  background: none;
  padding: 3px 8px;
  font-size: 12px;
  border-radius: 3px;
  color: var(--text-secondary);
  cursor: pointer;
  white-space: nowrap;
}

.wbs-filter-btn:hover {
  background: var(--surface-3);
}

.wbs-filter-btn--active {
  background: var(--brand-500);
  color: white;
}

.wbs-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.wbs-btn:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.wbs-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 16px 16px;
}

.wbs-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.wbs-group {
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.wbs-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--surface-2);
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}

.wbs-group-count {
  font-size: 11px;
  color: var(--text-tertiary);
}

.wbs-tree {
  list-style: none;
  margin: 0;
  padding: 4px 0;
}

.wbs-group-empty {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0;
}
</style>
