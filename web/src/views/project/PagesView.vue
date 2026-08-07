<script setup lang="ts">
/**
 * PagesView — 项目文档模块（对标 Plane 的 Pages）。
 *
 * 布局：
 *   - 左侧：文档树（按 parent_id 组装，支持折叠/嵌套）
 *   - 右侧：选中文档的富文本编辑器（基于 TipTap）
 *
 * 交互：
 *   - 顶部「+ 新建文档」创建根文档；树节点 hover 可新建子文档/删除
 *   - 点击树节点加载文档内容到编辑器
 *   - 编辑器内容变化后自动保存（防抖），乐观锁 version 冲突时提示
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { pagesApi, type Page } from "@/api/services/pages";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import { toast } from "@/lib/toast";
import RichTextEditor from "@/components/RichTextEditor.vue";
import { AppErrorState, AppEmptyState, InlineEdit } from "@/components";

const route = useRoute();
const { wsId } = useWorkspaceContext();

const projectId = computed(() => Number(route.params.projectId));
const loading = ref(true);
const error = ref("");

const pages = ref<Page[]>([]);
const currentPageId = ref<number | null>(null);
const collapsed = ref<Set<number>>(new Set());

// 编辑器状态
const editorHtml = ref("");
const editorJson = ref("");
const saving = ref(false);
let saveTimer: ReturnType<typeof setTimeout> | null = null;

// ---- 文档树组装 ----
interface PageNode extends Page {
  children: PageNode[];
}

const rootNodes = computed<PageNode[]>(() => {
  const map = new Map<number, PageNode>();
  for (const p of pages.value) map.set(p.id, { ...p, children: [] });
  const roots: PageNode[] = [];
  for (const node of map.values()) {
    if (node.parent_id && map.has(node.parent_id)) {
      map.get(node.parent_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  }
  // 按 sort_order 排序
  const sort = (ns: PageNode[]) =>
    ns.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
  sort(roots);
  for (const n of map.values()) sort(n.children);
  return roots;
});

const currentPage = computed(() =>
  pages.value.find((p) => p.id === currentPageId.value) ?? null,
);

// ---- 加载 ----
async function load() {
  loading.value = true;
  error.value = "";
  try {
    pages.value = await pagesApi.list(wsId.value, projectId.value);
    if (!currentPageId.value && pages.value.length > 0) {
      currentPageId.value = pages.value[0].id;
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

// 选中文档 → 载入编辑器内容
watch(currentPageId, (id) => {
  if (saveTimer) { clearTimeout(saveTimer); saveTimer = null; }
  if (!id) { editorHtml.value = ""; editorJson.value = ""; return; }
  const page = pages.value.find((p) => p.id === id);
  editorHtml.value = page?.description_html ?? "";
  editorJson.value = page?.description_json ? JSON.stringify(page.description_json) : "";
});

// ---- 操作 ----
async function createPage(parentId?: number | null) {
  try {
    const page = await pagesApi.create(wsId.value, projectId.value, {
      name: "未命名文档",
      description_html: "",
      description_json: "",
      parent_id: parentId ?? null,
    });
    pages.value.push(page);
    currentPageId.value = page.id;
    if (parentId) collapsed.value.delete(parentId);
    toast.success("已创建文档");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建失败");
  }
}

async function renamePage(page: Page, name: string) {
  if (!name.trim() || name === page.name) return;
  try {
    const updated = await pagesApi.update(wsId.value, projectId.value, page.id, {
      name: name.trim(),
      version: page.version,
    });
    const idx = pages.value.findIndex((p) => p.id === page.id);
    if (idx >= 0) pages.value[idx] = updated;
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "重命名失败");
  }
}

async function deletePage(page: Page) {
  if (!window.confirm(`确定删除文档「${page.name}」？子文档也会一并删除。`)) return;
  try {
    await pagesApi.remove(wsId.value, projectId.value, page.id);
    // 级联移除本地子树
    const idsToRemove = new Set<number>();
    const collect = (p: Page) => {
      idsToRemove.add(p.id);
      for (const child of pages.value.filter((c) => c.parent_id === p.id)) collect(child);
    };
    collect(page);
    pages.value = pages.value.filter((p) => !idsToRemove.has(p.id));
    if (currentPageId.value === page.id) {
      currentPageId.value = pages.value.length > 0 ? pages.value[0].id : null;
    }
    toast.success("已删除文档");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

// ---- 自动保存（防抖 800ms） ----
function onEditorUpdate() {
  if (!currentPageId.value) return;
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(saveCurrent, 800);
}

async function saveCurrent() {
  const page = currentPage.value;
  if (!page) return;
  saving.value = true;
  try {
    const updated = await pagesApi.update(wsId.value, projectId.value, page.id, {
      description_html: editorHtml.value,
      description_json: editorJson.value,
      description_stripped: stripHtml(editorHtml.value),
      version: page.version,
    });
    const idx = pages.value.findIndex((p) => p.id === page.id);
    if (idx >= 0) pages.value[idx] = updated;
  } catch (e: unknown) {
    toast.error("保存失败：" + (e instanceof Error ? e.message : "未知错误"));
  } finally {
    saving.value = false;
  }
}

function stripHtml(html: string): string {
  const el = document.createElement("div");
  el.innerHTML = html;
  return el.textContent ?? "";
}

function toggleCollapse(id: number) {
  const next = new Set(collapsed.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  collapsed.value = next;
}

onMounted(load);
</script>

<template>
  <div class="pages">
    <!-- 头部 -->
    <header class="pages__header">
      <div>
        <h1>文档</h1>
        <p class="hint">项目文档与知识库 — 支持嵌套层级与富文本</p>
      </div>
      <div class="pages__header-right">
        <span v-if="saving" class="pages__saving">保存中...</span>
        <button class="btn btn--primary" @click="createPage(null)">+ 新建文档</button>
      </div>
    </header>

    <AppErrorState v-if="error" :message="error" @retry="load" />

    <div v-else class="pages__body">
      <!-- 左侧文档树 -->
      <aside class="pages__tree">
        <div class="pages__tree-title">文档目录</div>
        <div v-if="pages.length === 0 && !loading" class="pages__tree-empty">
          <AppEmptyState title="暂无文档" description="点击右上角「新建文档」开始记录" />
        </div>
        <ul v-else class="pages__tree-list">
          <li v-for="node in rootNodes" :key="node.id" class="tree-node">
            <div
              class="tree-row"
              :class="{ 'tree-row--active': currentPageId === node.id }"
            >
              <span
                v-if="node.children.length > 0"
                class="tree-caret"
                :class="{ 'tree-caret--open': !collapsed.has(node.id) }"
                @click="toggleCollapse(node.id)"
              >▶</span>
              <span v-else class="tree-caret tree-caret--placeholder" />
              <span class="tree-file" @click="currentPageId = node.id">📄</span>
              <InlineEdit
                class="tree-name"
                :model-value="node.name"
                placeholder="未命名"
                :max-length="120"
                @submit="(v) => renamePage(node, v)"
              />
              <span class="tree-actions">
                <button class="tree-action" title="新建子文档" @click.stop="createPage(node.id)">＋</button>
                <button class="tree-action tree-action--danger" title="删除" @click.stop="deletePage(node)">×</button>
              </span>
            </div>
            <ul v-if="node.children.length > 0 && !collapsed.has(node.id)" class="tree-children">
              <li v-for="child in node.children" :key="child.id" class="tree-node">
                <div
                  class="tree-row"
                  :class="{ 'tree-row--active': currentPageId === child.id }"
                >
                  <span
                    v-if="child.children.length > 0"
                    class="tree-caret"
                    :class="{ 'tree-caret--open': !collapsed.has(child.id) }"
                    @click="toggleCollapse(child.id)"
                  >▶</span>
                  <span v-else class="tree-caret tree-caret--placeholder" />
                  <span class="tree-file" @click="currentPageId = child.id">📄</span>
                  <InlineEdit
                    class="tree-name"
                    :model-value="child.name"
                    placeholder="未命名"
                    :max-length="120"
                    @submit="(v) => renamePage(child, v)"
                  />
                  <span class="tree-actions">
                    <button class="tree-action" title="新建子文档" @click.stop="createPage(child.id)">＋</button>
                    <button class="tree-action tree-action--danger" title="删除" @click.stop="deletePage(child)">×</button>
                  </span>
                </div>
                <ul v-if="child.children.length > 0 && !collapsed.has(child.id)" class="tree-children">
                  <li v-for="grand in child.children" :key="grand.id" class="tree-node">
                    <div
                      class="tree-row"
                      :class="{ 'tree-row--active': currentPageId === grand.id }"
                      @click="currentPageId = grand.id"
                    >
                      <span class="tree-caret tree-caret--placeholder" />
                      <span class="tree-file">📄</span>
                      <span class="tree-name">{{ grand.name }}</span>
                      <span class="tree-actions">
                        <button class="tree-action" title="新建子文档" @click.stop="createPage(grand.id)">＋</button>
                        <button class="tree-action tree-action--danger" title="删除" @click.stop="deletePage(grand)">×</button>
                      </span>
                    </div>
                  </li>
                </ul>
              </li>
            </ul>
          </li>
        </ul>
      </aside>

      <!-- 右侧编辑器 -->
      <main class="pages__editor">
        <AppEmptyState
          v-if="!currentPage"
          icon="📄"
          title="选择或创建文档"
          description="从左侧目录选择文档，或点击「新建文档」开始撰写"
        >
          <button class="btn btn--primary" @click="createPage(null)">+ 新建文档</button>
        </AppEmptyState>
        <div v-else class="pages__doc">
          <div class="pages__doc-title">
            <InlineEdit
              :model-value="currentPage!.name"
              placeholder="未命名文档"
              :max-length="120"
              @submit="(v) => renamePage(currentPage!, v)"
            />
          </div>
          <RichTextEditor
            v-model:content-html="editorHtml"
            v-model:content-json="editorJson"
            variant="full"
            :min-height="'60vh'"
            placeholder="开始撰写文档..."
            @update:content-html="onEditorUpdate"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.pages__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}
.pages__header h1 { font-size: 20px; margin: 0 0 4px; }
.hint { color: var(--text-tertiary); font-size: 13px; margin: 0; }
.pages__header-right { display: flex; align-items: center; gap: 12px; }
.pages__saving { font-size: 12px; color: var(--text-tertiary); }

.btn--primary {
  padding: 8px 16px;
  background: var(--brand-500);
  color: var(--text-on-brand);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
}
.btn--primary:hover { background: var(--brand-600); }

.pages__body {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 16px;
  align-items: start;
}

/* ---- 文档树 ---- */
.pages__tree {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  padding: 8px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
  position: sticky;
  top: 0;
}

.pages__tree-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-tertiary);
  padding: 4px 8px 10px;
}

.pages__tree-empty { padding: 8px; }

.pages__tree-list, .tree-children {
  list-style: none;
  margin: 0;
  padding: 0;
}

.tree-children { padding-left: 16px; }

.tree-node { margin: 1px 0; }

.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 6px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.1s;
}

.tree-row:hover { background: var(--surface-3); }
.tree-row--active { background: var(--brand-50); }

.tree-caret {
  font-size: 9px;
  color: var(--text-tertiary);
  width: 14px;
  text-align: center;
  flex-shrink: 0;
  transition: transform 0.12s;
}
.tree-caret--open { transform: rotate(90deg); }
.tree-caret--placeholder { visibility: hidden; }

.tree-file { font-size: 12px; flex-shrink: 0; }

.tree-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-primary);
}

.tree-actions {
  display: none;
  gap: 2px;
  flex-shrink: 0;
}

.tree-row:hover .tree-actions { display: inline-flex; }

.tree-action {
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 3px;
  background: none;
  color: var(--text-tertiary);
  font-size: 12px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
}
.tree-action:hover { background: var(--surface-3); color: var(--text-primary); }
.tree-action--danger:hover { background: var(--danger-50); color: var(--danger-500); }

/* ---- 编辑器 ---- */
.pages__editor {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-1);
  min-height: 60vh;
  padding: 16px 24px;
}

.pages__doc-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}
</style>
