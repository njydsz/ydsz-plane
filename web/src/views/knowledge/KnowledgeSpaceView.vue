<script setup lang="ts">
/**
 * KnowledgeSpaceView — 知识库空间文档管理页。
 *
 * 布局：
 *   - 左侧：文档树（递归嵌套，支持展开/折叠、新建子文档、重命名、删除）
 *   - 右侧：Markdown 编辑 + 实时预览、保存（乐观锁自动版本递增）、
 *           版本历史（查看旧版本 / 回滚）、关联工作项、发布/归档切换
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  knowledgeApi,
  type KnowledgePage,
  type KnowledgePageNode,
  type KnowledgePageRelation,
  type KnowledgePageVersion,
  type KnowledgeSpace,
  type PageStatus,
} from "@/api/services/knowledge";
import { searchApi } from "@/api/services/search";
import { renderMarkdown } from "@/lib/markdown";
import { toast } from "@/lib/toast";
import KnowledgePageTreeNode from "./KnowledgePageTreeNode.vue";
import { AppEmptyState, AppErrorState } from "@/components";
import InlineEdit from "@/components/InlineEdit.vue";

const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const spaceId = computed(() => Number(route.params.spaceId ?? 0));

const loading = ref(true);
const error = ref("");

const space = ref<KnowledgeSpace | null>(null);
const tree = ref<KnowledgePageNode[]>([]);
const collapsed = ref<Set<number>>(new Set());
const selectedId = ref<number | null>(null);

/* ===== 编辑器状态 ===== */
const editorMode = ref<"edit" | "preview">("edit");
const editorMarkdown = ref("");
const changeSummary = ref("");
const saving = ref(false);
const dirty = ref(false);

/* ===== 版本历史 ===== */
const versions = ref<KnowledgePageVersion[]>([]);
const versionsLoading = ref(false);
const viewingVersion = ref<KnowledgePageVersion | null>(null);

/* ===== 关联工作项 ===== */
const relations = ref<KnowledgePageRelation[]>([]);
const relationsLoading = ref(false);
const relationInput = ref("");
const addingRelation = ref(false);

const statusLabels: Record<PageStatus, string> = {
  draft: "草稿",
  published: "已发布",
  archived: "已归档",
};

const statusClass: Record<PageStatus, string> = {
  draft: "status-badge--draft",
  published: "status-badge--published",
  archived: "status-badge--archived",
};

/* ===== 树工具函数 ===== */
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

function toggleCollapse(id: number) {
  const next = new Set(collapsed.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  collapsed.value = next;
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

async function loadVersions() {
  if (!selectedId.value) {
    versions.value = [];
    return;
  }
  versionsLoading.value = true;
  try {
    versions.value = await knowledgeApi.listVersions(
      workspaceId.value, spaceId.value, selectedId.value,
    );
  } catch {
    versions.value = [];
  } finally {
    versionsLoading.value = false;
  }
}

async function loadRelations() {
  if (!selectedId.value) {
    relations.value = [];
    return;
  }
  relationsLoading.value = true;
  try {
    relations.value = await knowledgeApi.listRelations(
      workspaceId.value, spaceId.value, selectedId.value,
    );
  } catch {
    relations.value = [];
  } finally {
    relationsLoading.value = false;
  }
}

/* 切换选中文档 → 载入编辑内容 */
watch(selectedId, (id) => {
  viewingVersion.value = null;
  editorMode.value = "edit";
  dirty.value = false;
  changeSummary.value = "";
  if (!id) {
    editorMarkdown.value = "";
    versions.value = [];
    relations.value = [];
    return;
  }
  const page = findNode(tree.value, id);
  editorMarkdown.value = page?.content_md ?? "";
  loadVersions();
  loadRelations();
});

/* ===== 文档操作 ===== */
async function createPage(parentId?: number | null) {
  try {
    const page = await knowledgeApi.createPage(workspaceId.value, spaceId.value, {
      title: "未命名文档",
      content_md: "",
      parent_id: parentId ?? null,
      status: "draft",
    });
    // 重新拉取树以获取正确的 lft/rgt/depth 与层级关系
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

/* ===== 保存 ===== */
async function savePage() {
  const page = currentPage.value;
  if (!page) return;
  saving.value = true;
  try {
    const updated = await knowledgeApi.updatePage(
      workspaceId.value, spaceId.value, page.id,
      {
        title: page.title,
        content_md: editorMarkdown.value,
        version: page.version,
        change_summary: changeSummary.value.trim() || undefined,
      },
    );
    const prev = findNode(tree.value, page.id);
    tree.value = replaceNode(tree.value, toNode(updated, prev));
    dirty.value = false;
    changeSummary.value = "";
    toast.success("保存成功，版本 v" + updated.version);
    await loadVersions();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function changeStatus(status: PageStatus) {
  const page = currentPage.value;
  if (!page || page.status === status) return;
  const confirms: Partial<Record<PageStatus, string>> = {
    published: `确定发布文档「${page.title}」？`,
    archived: `确定归档文档「${page.title}」？归档后仅可浏览。`,
    draft: `确定将文档「${page.title}」转为草稿？`,
  };
  if (confirms[status] && !window.confirm(confirms[status])) return;
  try {
    const updated = await knowledgeApi.updatePage(
      workspaceId.value, spaceId.value, page.id,
      { status, version: page.version },
    );
    const prev = findNode(tree.value, page.id);
    tree.value = replaceNode(tree.value, toNode(updated, prev));
    toast.success("状态已更新为「" + statusLabels[status] + "」");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "状态更新失败");
  }
}

/* ===== 版本历史 ===== */
function viewVersion(v: KnowledgePageVersion) {
  viewingVersion.value = v;
  editorMode.value = "preview";
}

function closeVersionPreview() {
  viewingVersion.value = null;
  editorMode.value = "edit";
}

async function revertToVersion(v: KnowledgePageVersion) {
  const page = currentPage.value;
  if (!page) return;
  if (!window.confirm(`确定回滚到版本 v${v.version}？当前未保存的更改将丢失。`)) return;
  try {
    const updated = await knowledgeApi.revertVersion(
      workspaceId.value, spaceId.value, page.id, v.version,
    );
    const prev = findNode(tree.value, page.id);
    tree.value = replaceNode(tree.value, toNode(updated, prev));
    editorMarkdown.value = updated.content_md ?? "";
    dirty.value = false;
    closeVersionPreview();
    toast.success(`已回滚到版本 v${v.version}`);
    await loadVersions();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "回滚失败");
  }
}

/* ===== 关联工作项 ===== */
async function addRelation() {
  const page = currentPage.value;
  if (!page || !relationInput.value.trim()) return;
  const issueId = Number(relationInput.value.trim());
  if (Number.isNaN(issueId) || issueId <= 0) {
    toast.error("请输入有效的工作项 ID");
    return;
  }
  addingRelation.value = true;
  try {
    await knowledgeApi.addRelation(workspaceId.value, spaceId.value, page.id, issueId);
    relationInput.value = "";
    toast.success("关联已添加");
    await loadRelations();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "添加关联失败");
  } finally {
    addingRelation.value = false;
  }
}

async function removeRelation(rel: KnowledgePageRelation) {
  const page = currentPage.value;
  if (!page || !window.confirm("确定删除该关联？")) return;
  try {
    await knowledgeApi.removeRelation(workspaceId.value, spaceId.value, page.id, rel.id);
    relations.value = relations.value.filter((r) => r.id !== rel.id);
    toast.success("关联已删除");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除关联失败");
  }
}

/** 通过工作空间搜索解析关联工作项所属项目，再跳转详情页 */
async function openIssue(issueId: number) {
  try {
    const res = await searchApi.searchWorkspace(workspaceId.value, {
      q: String(issueId),
      types: "issue",
      limit: 10,
    });
    const hit = res.results.issues?.find((i) => i.id === issueId && i.project_id);
    if (hit?.project_id) {
      router.push(`/${workspaceId.value}/projects/${hit.project_id}/issues/${issueId}`);
      return;
    }
    toast.info(`未能定位工作项 #${issueId} 所属项目`);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "解析工作项失败");
  }
}

/* ===== 其他 ===== */
function fmtTime(iso: string): string {
  const d = new Date(iso);
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, "0")}/${String(
    d.getDate(),
  ).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(
    d.getMinutes(),
  ).padStart(2, "0")}`;
}

const statusActions = computed(() => {
  const st = currentPage.value?.status;
  if (!st) return [];
  if (st === "draft") return [{ status: "published" as PageStatus, label: "发布", icon: "🚀" }];
  if (st === "published") {
    return [
      { status: "draft" as PageStatus, label: "转为草稿", icon: "✎" },
      { status: "archived" as PageStatus, label: "归档", icon: "🗄" },
    ];
  }
  return [
    { status: "published" as PageStatus, label: "重新发布", icon: "🚀" },
    { status: "draft" as PageStatus, label: "转为草稿", icon: "✎" },
  ];
});

onMounted(load);
</script>

<template>
  <div class="knowledge-space">
    <AppErrorState v-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 空间头部 -->
      <header class="knowledge-space__header">
        <div>
          <router-link
            :to="`/${workspaceId}/knowledge`"
            class="knowledge-space__back"
          >← {{ $t("knowledge.backToSpace") }}</router-link>
          <h1>{{ space?.name ?? "..." }}</h1>
          <p v-if="space?.description" class="hint">{{ space.description }}</p>
          <p v-else class="hint">/{{ space?.slug ?? "" }} · {{ space?.default_permission ?? "" }}</p>
        </div>
        <button class="btn btn--primary" @click="createPage(null)">＋ {{ $t("knowledge.newDoc") }}</button>
      </header>

      <div class="knowledge-space__body">
        <!-- 左侧文档树 -->
        <aside class="knowledge-space__tree">
          <div class="tree-header">
            <span class="tree-header__title">{{ $t("knowledge.docTree") }}</span>
            <button class="tree-header__add" title="新建根文档" @click="createPage(null)">＋</button>
          </div>
          <div v-if="tree.length === 0 && !loading" class="tree-empty">
            <AppEmptyState
              icon="📄"
              :title="$t('knowledge.docEmpty')"
              :description="$t('knowledge.docEmptyDesc')"
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

        <!-- 右侧内容区 -->
        <main class="knowledge-space__editor">
          <AppEmptyState
            v-if="!currentPage"
            icon="📄"
            :title="$t('knowledge.selectDoc')"
            :description="$t('knowledge.selectDocDesc')"
          >
            <button class="btn btn--primary" @click="createPage(null)">＋ {{ $t("knowledge.newDoc") }}</button>
          </AppEmptyState>

          <div v-else class="doc">
            <!-- 文档标题栏 -->
            <div class="doc__header">
              <InlineEdit
                class="doc__title"
                :model-value="currentPage.title"
                placeholder="未命名文档"
                :max-length="512"
                @submit="(v) => renamePage(currentPage, v)"
              />
              <div class="doc__meta">
                <span
                  class="status-badge"
                  :class="statusClass[currentPage.status]"
                >{{ statusLabels[currentPage.status] }}</span>
                <span class="doc__version">v{{ currentPage.version }}</span>
              </div>
            </div>

            <!-- 查看历史版本横幅 -->
            <div v-if="viewingVersion" class="version-banner">
              <span>
                正在查看版本 v{{ viewingVersion.version }}
                · {{ fmtTime(viewingVersion.created_at) }}
                <template v-if="viewingVersion.change_summary"> · {{ viewingVersion.change_summary }}</template>
              </span>
              <button class="btn btn--sm" @click="closeVersionPreview">返回编辑</button>
            </div>

            <!-- 编辑 / 预览切换 -->
            <div class="doc__toolbar">
              <div class="mode-switch">
                <button
                  class="mode-switch__btn"
                  :class="{ 'mode-switch__btn--active': editorMode === 'edit' }"
                  @click="editorMode = 'edit'"
                >{{ $t("knowledge.edit") }}</button>
                <button
                  class="mode-switch__btn"
                  :class="{ 'mode-switch__btn--active': editorMode === 'preview' }"
                  @click="editorMode = 'preview'"
                >{{ $t("knowledge.preview") }}</button>
              </div>
              <div class="doc__toolbar-right">
                <input
                  v-model="changeSummary"
                  class="form-input doc__summary"
                  placeholder="变更摘要（可选，将记录到版本历史）"
                  :disabled="!!viewingVersion"
                />
                <button
                  class="btn btn--primary btn--sm"
                  :disabled="saving || !!viewingVersion"
                  @click="savePage"
                >{{ saving ? $t("knowledge.saving") : $t("knowledge.save") }}</button>
                <button
                  v-for="act in statusActions"
                  :key="act.status"
                  class="btn btn--sm"
                  :disabled="!!viewingVersion"
                  @click="changeStatus(act.status)"
                >{{ act.icon }} {{ act.label }}</button>
              </div>
            </div>

            <!-- 编辑器 / 预览 -->
            <div class="doc__content">
              <textarea
                v-if="editorMode === 'edit' && !viewingVersion"
                v-model="editorMarkdown"
                class="doc__textarea"
                placeholder="使用 Markdown 撰写文档...（## 标题、**粗体**、`代码`、```代码块```）"
                spellcheck="false"
              />
              <div v-else class="doc__preview" v-html="renderMarkdown(viewingVersion?.content_md ?? editorMarkdown)" />
            </div>

            <!-- 版本历史 -->
            <div class="doc__section">
              <h4>{{ $t("knowledge.versionHistory") }}</h4>
              <div v-if="versionsLoading" class="text-muted">加载中...</div>
              <div v-else-if="versions.length === 0" class="text-muted">{{ $t("knowledge.versionEmpty") }}</div>
              <div v-else class="versions-list">
                <div
                  v-for="v in versions"
                  :key="v.id"
                  class="version-item"
                  :class="{ 'version-item--current': v.version === currentPage.version }"
                >
                  <span class="version-item__number">v{{ v.version }}</span>
                  <span class="version-item__time">{{ fmtTime(v.created_at) }}</span>
                  <span class="version-item__summary">{{ v.change_summary || "—" }}</span>
                  <span v-if="v.version === currentPage.version" class="version-item__current-tag">当前</span>
                  <button class="btn btn--sm btn--ghost" @click="viewVersion(v)">{{ $t("knowledge.view") }}</button>
                  <button
                    class="btn btn--sm btn--ghost"
                    :disabled="v.version === currentPage.version"
                    @click="revertToVersion(v)"
                  >{{ $t("knowledge.revert") }}</button>
                </div>
              </div>
            </div>

            <!-- 关联工作项 -->
            <div class="doc__section">
              <h4>{{ $t("knowledge.linkIssue") }}</h4>
              <div class="links-add">
                <input
                  v-model="relationInput"
                  type="text"
                  class="form-input"
                  :placeholder="$t('knowledge.linkIssuePlaceholder')"
                  style="width: 180px"
                  @keydown.enter="addRelation"
                />
                <button class="btn btn--sm btn--primary" :disabled="addingRelation" @click="addRelation">
                  {{ $t("knowledge.addLink") }}
                </button>
              </div>
              <div v-if="relationsLoading" class="text-muted links-state">加载中...</div>
              <div v-else-if="relations.length === 0" class="text-muted links-state">
                {{ $t("knowledge.linkEmpty") }}
              </div>
              <div v-else class="links-list">
                <div v-for="rel in relations" :key="rel.id" class="link-item">
                  <span class="link-item__type">工作项</span>
                  <a
                    class="link-item__id"
                    :title="`跳转到工作项 #${rel.issue_id}`"
                    @click="openIssue(rel.issue_id)"
                  >#{{ rel.issue_id }}</a>
                  <span class="link-item__time">{{ fmtTime(rel.created_at) }}</span>
                  <button class="btn btn--sm btn--ghost link-item__remove" @click="removeRelation(rel)">✕</button>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </template>
  </div>
</template>

<style scoped>
.knowledge-space__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.knowledge-space__header h1 {
  font-size: 20px;
  margin: 4px 0;
}

.knowledge-space__back {
  font-size: 12px;
  color: var(--text-tertiary);
  text-decoration: none;
}

.knowledge-space__back:hover {
  color: var(--brand-500);
}

.hint {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}

.knowledge-space__body {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 16px;
  align-items: start;
}

/* ---- 文档树 ---- */
.knowledge-space__tree {
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

/* ---- 编辑器 ---- */
.knowledge-space__editor {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  min-height: 60vh;
  padding: 16px 24px;
}

.doc__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.doc__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
}

.doc__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.status-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 500;
}

.status-badge--draft {
  background: var(--surface-3);
  color: var(--text-tertiary);
}

.status-badge--published {
  background: var(--success-50, #e6f4ea);
  color: var(--success-600, #1e8e3e);
}

.status-badge--archived {
  background: var(--warning-50, #fef7e0);
  color: var(--warning-600, #b06000);
}

.doc__version {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-tertiary);
}

/* 版本浏览横幅 */
.version-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 12px;
  border: 1px solid var(--warning-300, #f0c36d);
  border-radius: var(--radius-sm);
  background: var(--warning-50, #fef7e0);
  color: var(--warning-700, #7a4f01);
  font-size: 12px;
}

/* 工具栏 */
.doc__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.doc__toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.mode-switch {
  display: inline-flex;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.mode-switch__btn {
  padding: 5px 14px;
  font-size: 12px;
  font-family: inherit;
  border: none;
  background: none;
  color: var(--text-tertiary);
  cursor: pointer;
}

.mode-switch__btn--active {
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.doc__summary {
  width: 220px;
  font-size: 12px;
  padding: 6px 10px;
}

.doc__content {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.doc__textarea {
  display: block;
  width: 100%;
  min-height: 42vh;
  border: none;
  outline: none;
  resize: vertical;
  padding: 14px 16px;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
  background: var(--surface-1);
}

.doc__preview {
  min-height: 42vh;
  max-height: 60vh;
  overflow-y: auto;
  padding: 14px 16px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
  word-break: break-word;
}

/* Markdown 预览样式（作用于 v-html 生成内容） */
.doc__preview :deep(h1),
.doc__preview :deep(h2),
.doc__preview :deep(h3),
.doc__preview :deep(h4) {
  color: var(--text-primary);
  margin: 1.2em 0 0.5em;
  line-height: 1.35;
}

.doc__preview :deep(h1) { font-size: 22px; border-bottom: 1px solid var(--border-subtle); padding-bottom: 8px; }
.doc__preview :deep(h2) { font-size: 18px; }
.doc__preview :deep(h3) { font-size: 16px; }
.doc__preview :deep(p) { margin: 0.6em 0; }
.doc__preview :deep(a) { color: var(--brand-500); }
.doc__preview :deep(code) {
  font-family: var(--font-mono);
  font-size: 12px;
  background: var(--surface-3);
  padding: 1px 5px;
  border-radius: 4px;
}
.doc__preview :deep(pre) {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 12px;
  overflow-x: auto;
}
.doc__preview :deep(pre code) {
  background: none;
  padding: 0;
  font-size: 12px;
  line-height: 1.6;
}
.doc__preview :deep(blockquote) {
  margin: 0.8em 0;
  padding: 4px 14px;
  border-left: 3px solid var(--brand-300, #9ec5ff);
  color: var(--text-tertiary);
  background: var(--surface-2);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
}
.doc__preview :deep(ul),
.doc__preview :deep(ol) {
  padding-left: 24px;
  margin: 0.6em 0;
}
.doc__preview :deep(table) {
  border-collapse: collapse;
  margin: 0.8em 0;
  width: 100%;
  font-size: 13px;
}
.doc__preview :deep(th),
.doc__preview :deep(td) {
  border: 1px solid var(--border-default);
  padding: 6px 10px;
  text-align: left;
}
.doc__preview :deep(th) {
  background: var(--surface-2);
  font-weight: 600;
  color: var(--text-primary);
}
.doc__preview :deep(img) {
  max-width: 100%;
  border-radius: var(--radius-sm);
}
.doc__preview :deep(hr) {
  border: none;
  border-top: 1px solid var(--border-subtle);
  margin: 1.2em 0;
}
.doc__preview :deep(del) { color: var(--text-tertiary); }

/* ---- 区块（版本历史 / 关联） ---- */
.doc__section {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.doc__section h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 12px;
}

.versions-list,
.links-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.version-item:hover {
  background: var(--surface-2);
}

.version-item--current {
  background: var(--brand-50);
}

.version-item__number {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--brand-500);
  min-width: 36px;
}

.version-item__time {
  color: var(--text-tertiary);
  font-size: 11px;
  white-space: nowrap;
}

.version-item__summary {
  flex: 1;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.version-item__current-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 999px;
  background: var(--success-50, #e6f4ea);
  color: var(--success-600, #1e8e3e);
}

.links-add {
  display: flex;
  gap: 8px;
  align-items: center;
}

.links-state {
  margin-top: 8px;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.link-item:hover {
  background: var(--surface-2);
}

.link-item__type {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.link-item__id {
  color: var(--brand-500);
  cursor: pointer;
  font-family: var(--font-mono);
  font-weight: 500;
  text-decoration: none;
}

.link-item__id:hover {
  text-decoration: underline;
}

.link-item__time {
  flex: 1;
  color: var(--text-tertiary);
  font-size: 11px;
}

.link-item__remove {
  opacity: 0;
  transition: opacity 0.15s;
}

.link-item:hover .link-item__remove {
  opacity: 1;
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

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--brand-500);
  padding: 2px 4px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-input {
  padding: 6px 10px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
}

.form-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}
</style>
