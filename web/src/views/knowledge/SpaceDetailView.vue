<script setup lang="ts">
/**
 * SpaceDetailView — 知识库空间详情页。
 *
 * 布局：
 *   - 左栏 240px：文档树（递归嵌套，使用 KnowledgePageTreeNode 递归组件）
 *   - 主栏 1fr：面包屑 + 内容（新建根文档 / 文档详情）
 *
 * 交互：
 *   - 树节点选中 → selectedPage 设置 → 主栏加载页面详情
 *   - ＋ 浮动按钮创建根文档；hover 树上节点出现 ＋ 子页面
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  knowledgeApi,
  type KnowledgePage,
  type KnowledgePageNode,
  type KnowledgeSpace,
} from "@/api/services/knowledge";
import { toast } from "@/lib/toast";
import KnowledgePageTreeNode from "./KnowledgePageTreeNode.vue";
import KnowledgePageDetailView from "./KnowledgePageDetailView.vue";
import { AppEmptyState, AppErrorState } from "@/components";

const route = useRoute();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const spaceId = computed(() => Number(route.params.spaceId ?? 0));

const loading = ref(true);
const error = ref("");

const space = ref<KnowledgeSpace | null>(null);
const tree = ref<KnowledgePageNode[]>([]);
const collapsed = ref<Set<number>>(new Set());
const selectedId = ref<number | null>(null);

/* ===== 树工具 ===== */
function findNode(nodes: KnowledgePageNode[], id: number): KnowledgePageNode | null {
  for (const n of nodes) {
    if (n.id === id) return n;
    const found = findNode(n.children ?? [], id);
    if (found) return found;
  }
  return null;
}

function replaceNode(nodes: KnowledgePageNode[], updated: KnowledgePageNode): KnowledgePageNode[] {
  return nodes.map((n) => {
    if (n.id === updated.id) return { ...updated, children: n.children };
    if (n.children?.length) return { ...n, children: replaceNode(n.children, updated) };
    return n;
  });
}

function removeNodeById(nodes: KnowledgePageNode[], id: number): KnowledgePageNode[] {
  const result: KnowledgePageNode[] = [];
  for (const n of nodes) {
    if (n.id === id) continue;
    const children = n.children?.length ? removeNodeById(n.children, id) : n.children;
    result.push({ ...n, children });
  }
  return result;
}

function toNode(page: KnowledgePage, prev?: KnowledgePageNode | null): KnowledgePageNode {
  return { ...page, children: prev?.children ?? [] };
}

const currentPage = computed<KnowledgePage | null>(() => {
  if (!selectedId.value) return null;
  return findNode(tree.value, selectedId.value) ?? null;
});

/* ===== 加载 ===== */
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const [sp, pageTree] = await Promise.all([
      knowledgeApi.getSpace(workspaceId.value, spaceId.value),
      knowledgeApi.getPageTree(workspaceId.value, spaceId.value),
    ]);
    space.value = sp;
    tree.value = pageTree;
    if (!selectedId.value && pageTree.length > 0) {
      selectedId.value = pageTree[0].id;
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/* ===== 折叠切换 ===== */
function toggleCollapse(id: number) {
  const next = new Set(collapsed.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  collapsed.value = next;
}

/* ===== 文档操作 ===== */
async function createPage(parentId?: number | null) {
  try {
    const page = await knowledgeApi.createPage(workspaceId.value, spaceId.value, {
      title: "未命名文档",
      content_md: "",
      parent_id: parentId ?? null,
      status: "draft",
    });
    // 重新拉取树以获取正确的层级关系
    tree.value = await knowledgeApi.getPageTree(workspaceId.value, spaceId.value);
    if (parentId) collapsed.value.delete(parentId);
    selectedId.value = page.id;
    toast.success("已创建文档");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建文档失败");
  }
}

async function renamePage(node: KnowledgePageNode, title: string) {
  const name = title.trim();
  if (!name || name === node.title) return;
  try {
    const updated = await knowledgeApi.updatePage(
      workspaceId.value, spaceId.value, node.id,
      { title: name, version: node.version },
    );
    const prev = findNode(tree.value, node.id);
    tree.value = replaceNode(tree.value, toNode(updated, prev));
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "重命名失败");
  }
}

async function deletePage(node: KnowledgePageNode) {
  if (!window.confirm(`确定删除文档「${node.title}」？其子文档也会一并删除。`)) return;
  try {
    await knowledgeApi.deletePage(workspaceId.value, spaceId.value, node.id);
    tree.value = removeNodeById(tree.value, node.id);
    if (selectedId.value === node.id) {
      selectedId.value = tree.value.length > 0 ? tree.value[0].id : null;
    }
    toast.success("已删除文档");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除文档失败");
  }
}

/* 页面删除后（在子组件中发生删除时）刷新树 */
async function onPageDeleted(deletedId: number) {
  tree.value = await knowledgeApi.getPageTree(workspaceId.value, spaceId.value);
  if (selectedId.value === deletedId) {
    selectedId.value = tree.value.length > 0 ? tree.value[0].id : null;
  }
}

/* 页面保存后刷新树 */
async function onPageSaved(page: KnowledgePage) {
  const prev = findNode(tree.value, page.id);
  tree.value = replaceNode(tree.value, toNode(page, prev));
}

onMounted(load);
</script>

<template>
  <div class="space-detail">
    <AppErrorState v-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 空间头部 -->
      <header class="space-detail__header">
        <div class="space-detail__header-left">
          <router-link
            :to="`/${workspaceId}/knowledge`"
            class="space-detail__back"
          >← 返回空间列表</router-link>
          <h1 v-if="space">{{ space.name }}</h1>
          <h1 v-else>...</h1>
          <p v-if="space?.description" class="hint">{{ space.description }}</p>
          <p v-else class="hint">/{{ space?.slug ?? "" }}</p>
        </div>
        <div class="space-detail__header-right">
          <button class="btn btn--primary" @click="createPage(null)">＋ 新建根文档</button>
        </div>
      </header>

      <div class="space-detail__body">
        <!-- 左侧文档树 -->
        <aside class="space-detail__tree">
          <div class="tree-header">
            <span class="tree-header__title">文档目录</span>
            <button class="tree-header__add" title="新建根文档" @click="createPage(null)">＋</button>
          </div>

          <div v-if="tree.length === 0 && !loading" class="tree-empty">
            <AppEmptyState
              icon="📄"
              title="暂无文档"
              description="点击顶部或右上＋按钮创建第一篇文档"
            />
          </div>

          <ul v-else class="tree-list">
            <KnowledgePageTreeNode
              v-for="node in tree"
              :key="node.id"
              :node="node"
              :collapsed="collapsed"
              :selected-id="selectedId"
              @select="selectedId = $event"
              @toggle="toggleCollapse"
              @create-child="createPage"
              @rename="renamePage"
              @delete="deletePage"
            />
          </ul>
        </aside>

        <!-- 右侧主栏 -->
        <main class="space-detail__main">
          <AppEmptyState
            v-if="!currentPage"
            icon="📄"
            title="选择或创建文档"
            description="从左侧目录选择文档，或点击「新建根文档」开始撰写"
          />
          <KnowledgePageDetailView
            v-else-if="currentPage"
            :key="currentPage.id"
            :page="currentPage"
            :workspace-id="workspaceId"
            :space-id="spaceId"
            @deleted="onPageDeleted"
            @saved="onPageSaved"
          />
        </main>
      </div>
    </template>
  </div>
</template>

<style scoped>
.space-detail__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 16px;
}
.space-detail__header-left {
  flex: 1;
  min-width: 0;
}
.space-detail__header h1 {
  font-size: 20px;
  margin: 4px 0;
}
.space-detail__back {
  font-size: 12px;
  color: var(--text-tertiary);
  text-decoration: none;
}
.space-detail__back:hover {
  color: var(--brand-500);
}
.hint {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}
.space-detail__header-right {
  flex-shrink: 0;
}

.space-detail__body {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 16px;
  align-items: start;
}

/* ---- 左侧文档树 ---- */
.space-detail__tree {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  padding: 8px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  position: sticky;
  top: 0;
}
.tree-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px 10px;
}
.tree-header__title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-tertiary);
}
.tree-header__add {
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 4px;
  background: none;
  color: var(--text-tertiary);
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
}
.tree-header__add:hover {
  background: var(--surface-3);
  color: var(--brand-500);
}
.tree-empty {
  padding: 8px;
}
.tree-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

/* ---- 右侧主栏 ---- */
.space-detail__main {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  min-height: 60vh;
  padding: 16px 24px;
}

/* ---- 按钮 ---- */
.btn {
  padding: 8px 16px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.btn:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 500;
}
.btn--primary:hover {
  background: var(--brand-600);
  color: var(--text-on-brand);
}
</style>
