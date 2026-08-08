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

import {
  pagesApi, versionsApi, linksApi, templatesApi, sharesApi,
  type Page, type DocumentVersion, type DocumentLink,
  type PageTemplate, type PageShare,
} from "@/api/services/pages";
import { useWorkspaceContext } from "@/composables/useWorkspaceContext";
import { toast } from "@/lib/toast";
import RichTextEditor from "@/components/RichTextEditor.vue";
import { AppErrorState, AppEmptyState, AppModal, InlineEdit } from "@/components";

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

// 分类
const editCategory = ref<string | null>(null);

// 版本历史
const versions = ref<DocumentVersion[]>([]);
const versionsLoading = ref(false);
const viewingVersion = ref<number | null>(null);
const viewingVersionContent = ref<DocumentVersion | null>(null);

// 关联工作项
const links = ref<DocumentLink[]>([]);
const linksLoading = ref(false);
const linkInputId = ref("");

// 模板
const templates = ref<PageTemplate[]>([]);
const templatesLoading = ref(false);
const showTemplateModal = ref(false);
const creatingFromTemplate = ref(false);

// 分享链接
const shares = ref<PageShare[]>([]);
const sharesLoading = ref(false);
const showShareModal = ref(false);
const sharePasswordInput = ref("");
const shareExpiryInput = ref(""); // ISO string or ""
const copiedShareIndex = ref<number | null>(null);

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
  viewingVersion.value = null;
  viewingVersionContent.value = null;
  if (!id) { editorHtml.value = ""; editorJson.value = ""; versions.value = []; links.value = []; return; }
  const page = pages.value.find((p) => p.id === id);
  editorHtml.value = page?.description_html ?? "";
  editorJson.value = page?.description_json ? JSON.stringify(page.description_json) : "";
  editCategory.value = null;
  loadVersions();
  loadLinks();
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

// ---- 分类 ----
const categoryOptions = [
  { value: "PRD", label: "PRD" },
  { value: "design", label: "技术方案" },
  { value: "api", label: "接口文档" },
  { value: "test", label: "测试报告" },
  { value: "checklist", label: "交付清单" },
];

function categoryLabel(cat?: string | null): string {
  if (!cat) return "—";
  return categoryOptions.find((c) => c.value === cat)?.label ?? cat;
}

async function saveCategory(page: Page, category: string) {
  if (category === (page.category ?? "")) return;
  try {
    const updated = await pagesApi.update(wsId.value, projectId.value, page.id, {
      category: category || "",
      version: page.version,
    });
    const idx = pages.value.findIndex((p) => p.id === page.id);
    if (idx >= 0) pages.value[idx] = updated;
    editCategory.value = null;
    toast.success("分类已更新");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "分类更新失败");
  }
}

// ---- 版本历史 ----
async function loadVersions() {
  if (!currentPageId.value) return;
  versionsLoading.value = true;
  try {
    versions.value = await versionsApi.list(wsId.value, projectId.value, currentPageId.value);
  } catch {
    // 非关键模块静默忽略
  } finally {
    versionsLoading.value = false;
  }
}

async function viewVersion(versionNumber: number) {
  if (!currentPageId.value) return;
  try {
    viewingVersionContent.value = await versionsApi.get(
      wsId.value, projectId.value, currentPageId.value, versionNumber,
    );
    viewingVersion.value = versionNumber;
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "加载版本失败");
  }
}

function closeViewVersion() {
  viewingVersion.value = null;
  viewingVersionContent.value = null;
}

async function rollbackToVersion(versionNumber: number) {
  if (!currentPageId.value) return;
  if (!window.confirm(`确定回滚到版本 v${versionNumber}？当前未保存的更改将丢失。`)) return;
  try {
    const page = await versionsApi.rollback(
      wsId.value, projectId.value, currentPageId.value, versionNumber,
    );
    const idx = pages.value.findIndex((p) => p.id === page.id);
    if (idx >= 0) pages.value[idx] = page;
    editorHtml.value = page.description_html ?? "";
    editorJson.value = page.description_json ? JSON.stringify(page.description_json) : "";
    viewingVersion.value = null;
    viewingVersionContent.value = null;
    toast.success("已回滚到版本 v" + versionNumber);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "回滚失败");
  }
}

// ---- 关联工作项 ----
async function loadLinks() {
  if (!currentPageId.value) return;
  linksLoading.value = true;
  try {
    links.value = await linksApi.list(wsId.value, projectId.value, currentPageId.value);
  } catch {
    // 非关键模块静默忽略
  } finally {
    linksLoading.value = false;
  }
}

async function addLink() {
  if (!currentPageId.value || !linkInputId.value.trim()) return;
  const issueId = Number(linkInputId.value.trim());
  if (isNaN(issueId) || issueId <= 0) {
    toast.error("请输入有效的工作项 ID");
    return;
  }
  try {
    await linksApi.create(wsId.value, projectId.value, currentPageId.value, {
      linkable_type: "issue",
      linkable_id: issueId,
    });
    linkInputId.value = "";
    toast.success("关联已添加");
    await loadLinks();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "关联失败");
  }
}

async function removeLink(linkId: number) {
  if (!currentPageId.value || !window.confirm("确定删除该关联？")) return;
  try {
    await linksApi.remove(wsId.value, projectId.value, currentPageId.value, linkId);
    toast.success("关联已删除");
    await loadLinks();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除关联失败");
  }
}

function linkableLabel(type: string): string {
  const map: Record<string, string> = { issue: "工作项", sprint: "迭代", version: "版本" };
  return map[type] ?? type;
}

// ---- 模板 ----
async function openTemplateModal() {
  showTemplateModal.value = true;
  templatesLoading.value = true;
  try {
    templates.value = await templatesApi.list(wsId.value, projectId.value);
  } catch {
    // 非关键模块静默忽略
  } finally {
    templatesLoading.value = false;
  }
}

function closeTemplateModal() {
  showTemplateModal.value = false;
}

async function createFromTemplate(tmpl: PageTemplate) {
  if (!tmpl.content_html) {
    toast.error("该模板无内容，无法创建");
    return;
  }
  creatingFromTemplate.value = true;
  try {
    const page = await pagesApi.create(wsId.value, projectId.value, {
      name: tmpl.name,
      description_html: tmpl.content_html,
      description_json: "",
      description_stripped: stripHtml(tmpl.content_html ?? ""),
    });
    pages.value.push(page);
    currentPageId.value = page.id;
    closeTemplateModal();
    toast.success(`已从模板「${tmpl.name}」创建文档`);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "从模板创建失败");
  } finally {
    creatingFromTemplate.value = false;
  }
}

const templatesByCategory = computed(() => {
  const groups: Record<string, PageTemplate[]> = {};
  for (const t of templates.value) {
    const cat = t.category || "其他";
    if (!groups[cat]) groups[cat] = [];
    groups[cat].push(t);
  }
  return groups;
});

function templateCategoryLabel(cat: string): string {
  const map: Record<string, string> = {
    requirement: "需求",
    technical: "技术",
    meeting: "会议",
    review: "复盘",
  };
  return map[cat] || cat || "其他";
}

// ---- 分享链接 ----
async function openShareModal() {
  if (!currentPageId.value) return;
  showShareModal.value = true;
  sharesLoading.value = true;
  try {
    shares.value = await sharesApi.list(wsId.value, projectId.value, currentPageId.value);
  } catch {
    // 非关键模块静默忽略
  } finally {
    sharesLoading.value = false;
  }
}

function closeShareModal() {
  showShareModal.value = false;
  sharePasswordInput.value = "";
  shareExpiryInput.value = "";
  copiedShareIndex.value = null;
}

async function createShare() {
  if (!currentPageId.value) return;
  try {
    const input: { password?: string; expires_at?: string } = {};
    if (sharePasswordInput.value.trim()) input.password = sharePasswordInput.value.trim();
    if (shareExpiryInput.value) input.expires_at = new Date(shareExpiryInput.value).toISOString();
    const share = await sharesApi.create(wsId.value, projectId.value, currentPageId.value, input);
    shares.value = [share, ...shares.value];
    sharePasswordInput.value = "";
    shareExpiryInput.value = "";
    toast.success("分享链接已生成");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建分享链接失败");
  }
}

async function toggleShareActive(share: PageShare) {
  try {
    const updated = await sharesApi.update(
      wsId.value, projectId.value, share.page_id, share.id, { is_active: !share.is_active },
    );
    const idx = shares.value.findIndex((s) => s.id === share.id);
    if (idx >= 0) shares.value[idx] = updated;
    toast.success(updated.is_active ? "链接已激活" : "链接已停用");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "操作失败");
  }
}

async function revokeShare(share: PageShare) {
  if (!window.confirm("确定吊销该分享链接？此操作不可恢复。")) return;
  try {
    await sharesApi.revoke(wsId.value, projectId.value, share.page_id, share.id);
    shares.value = shares.value.filter((s) => s.id !== share.id);
    toast.success("分享链接已吊销");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "吊销失败");
  }
}

function shareUrl(token: string): string {
  return `${window.location.origin}/public/page/${token}`;
}

async function copyShareUrl(token: string, index: number) {
  try {
    await navigator.clipboard.writeText(shareUrl(token));
    copiedShareIndex.value = index;
    toast.success("链接已复制到剪贴板");
    setTimeout(() => { if (copiedShareIndex.value === index) copiedShareIndex.value = null; }, 2000);
  } catch {
    toast.error("复制失败，请手动复制");
  }
}

function shareStatusLabel(share: PageShare): string {
  if (!share.is_active) return "已停用";
  if (share.expires_at && new Date(share.expires_at) < new Date()) return "已过期";
  return "生效中";
}

function shareStatusClass(share: PageShare): string {
  if (!share.is_active) return "share-row__status--inactive";
  if (share.expires_at && new Date(share.expires_at) < new Date()) return "share-row__status--expired";
  return "share-row__status--active";
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
        <button class="btn" @click="openTemplateModal">从模板创建</button>
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
            <div class="pages__doc-category">
              <span class="pages__doc-category-label">分类：</span>
              <span v-if="editCategory === null" class="pages__doc-category-value">
                <span class="pages__doc-category-badge" :class="currentPage?.category ? '' : 'pages__doc-category-badge--empty'">
                  {{ categoryLabel(currentPage?.category) }}
                </span>
                <button class="btn btn--sm btn--ghost" @click="editCategory = currentPage?.category ?? ''">✎</button>
              </span>
              <span v-else class="pages__doc-category-edit">
                <select v-model="editCategory" class="edit-select">
                  <option value="">-- 请选择 --</option>
                  <option v-for="opt in categoryOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
                <button class="btn btn--sm btn--primary" @click="saveCategory(currentPage!, editCategory)">保存</button>
                <button class="btn btn--sm" @click="editCategory = null">取消</button>
              </span>
            </div>
            <button class="btn btn--sm btn--ghost pages__share-btn" @click="openShareModal">🔗 分享</button>
          </div>
          <RichTextEditor
            v-model:content-html="editorHtml"
            v-model:content-json="editorJson"
            variant="full"
            :min-height="'40vh'"
            placeholder="开始撰写文档..."
            @update:content-html="onEditorUpdate"
          />

          <!-- 版本历史 -->
          <div class="pages__versions">
            <h4>版本历史</h4>
            <div v-if="versionsLoading" class="text-muted">加载中...</div>
            <div v-else-if="versions.length === 0" class="text-muted">暂无历史版本</div>
            <div v-else class="versions-list">
              <div
                v-for="dv in versions"
                :key="dv.id"
                class="version-item"
                :class="{ 'version-item--active': viewingVersion === dv.version_number }"
              >
                <span class="version-item__number">v{{ dv.version_number }}</span>
                <span class="version-item__time">{{ new Date(dv.created_at).toLocaleString() }}</span>
                <button class="btn btn--sm btn--ghost" @click="viewVersion(dv.version_number)">查看</button>
                <button class="btn btn--sm btn--ghost" @click="rollbackToVersion(dv.version_number)">回滚</button>
              </div>
            </div>

            <!-- 查看版本内容浮层 -->
            <div v-if="viewingVersion && viewingVersionContent" class="version-preview">
              <div class="version-preview__header">
                <span>版本 v{{ viewingVersion }} {{ new Date(viewingVersionContent.created_at).toLocaleString() }}</span>
                <button class="btn btn--sm" @click="closeViewVersion">关闭</button>
              </div>
              <div class="version-preview__content" v-html="viewingVersionContent.content_html || '<p>（空内容）</p>'"></div>
            </div>
          </div>

          <!-- 关联工作项 -->
          <div class="pages__links">
            <h4>关联工作项</h4>
            <div class="links-add">
              <input
                v-model="linkInputId"
                type="text"
                class="form-input"
                placeholder="输入工作项 ID"
                style="width: 160px"
                @keydown.enter="addLink"
              />
              <button class="btn btn--sm btn--primary" @click="addLink">添加关联</button>
            </div>
            <div v-if="linksLoading" class="text-muted" style="margin-top: 8px">加载中...</div>
            <div v-else-if="links.length === 0" class="text-muted" style="margin-top: 8px">暂无关联</div>
            <div v-else class="links-list">
              <div v-for="link in links" :key="link.id" class="link-item">
                <span class="link-item__type">{{ linkableLabel(link.linkable_type) }}</span>
                <router-link
                  :to="`/${wsId}/projects/${projectId}/issues/${link.linkable_id}`"
                  class="link-item__id"
                >
                  #{{ link.linkable_id }}
                </router-link>
                <button class="btn btn--sm btn--ghost link-item__remove" @click="removeLink(link.id)">✕</button>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>

  <!-- 从模板创建 模态框 -->
  <AppModal :visible="showTemplateModal" title="选择文档模板" width="640px" @close="closeTemplateModal">
    <div v-if="templatesLoading" class="text-muted">加载中...</div>
    <div v-else-if="templates.length === 0" class="text-muted">暂无可用模板</div>
    <div v-else class="template-list">
      <div v-for="(tmpls, cat) in templatesByCategory" :key="cat" class="template-list__group">
        <div class="template-list__cat">{{ templateCategoryLabel(cat as string) }}</div>
        <div
          v-for="tmpl in tmpls"
          :key="tmpl.id"
          class="template-item"
        >
          <div class="template-item__info">
            <span class="template-item__name">{{ tmpl.name }}</span>
            <span class="template-item__desc">{{ tmpl.description || '—' }}</span>
          </div>
          <button
            class="btn btn--sm btn--primary"
            :disabled="creatingFromTemplate || !tmpl.content_html"
            @click="createFromTemplate(tmpl)"
          >使用此模板</button>
        </div>
      </div>
    </div>
    <template #footer>
      <button class="btn" @click="closeTemplateModal">关闭</button>
    </template>
  </AppModal>

  <!-- 分享链接管理 模态框 -->
  <AppModal :visible="showShareModal" title="分享链接管理" width="640px" @close="closeShareModal">
    <div class="shares-create">
      <h4>生成分享链接</h4>
      <div class="shares-create__row">
        <input
          v-model="sharePasswordInput"
          type="password"
          class="form-input"
          placeholder="访问密码（可选）"
          style="width: 180px"
        />
        <input
          v-model="shareExpiryInput"
          type="datetime-local"
          class="form-input"
          style="width: 180px"
        />
        <button class="btn btn--sm btn--primary" @click="createShare">生成链接</button>
      </div>
      <p class="text-muted" style="margin: 4px 0 0; font-size: 11px">
        分享链接无需登录即可访问此文档（只读）。
      </p>
    </div>

    <div class="shares-list-section">
      <h4>已生成的分享链接</h4>
      <div v-if="sharesLoading" class="text-muted">加载中...</div>
      <div v-else-if="shares.length === 0" class="text-muted">暂无分享链接</div>
      <div v-else class="shares-list">
        <div
          v-for="(share, idx) in shares"
          :key="share.id"
          class="share-row"
          :class="{ 'share-row--inactive': !share.is_active }"
        >
          <div class="share-row__info">
            <span class="share-row__token">{{ share.token.slice(0, 12) }}…</span>
            <span class="share-row__status" :class="shareStatusClass(share)">{{ shareStatusLabel(share) }}</span>
            <span v-if="share.expires_at" class="share-row__expiry">
              过期：{{ new Date(share.expires_at).toLocaleString() }}
            </span>
          </div>
          <div class="share-row__actions">
            <button class="btn btn--sm btn--ghost" @click="copyShareUrl(share.token, idx)">
              {{ copiedShareIndex === idx ? '已复制' : '复制链接' }}
            </button>
            <button class="btn btn--sm" @click="toggleShareActive(share)">
              {{ share.is_active ? '停用' : '激活' }}
            </button>
            <button class="btn btn--sm btn--danger" @click="revokeShare(share)">吊销</button>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <button class="btn" @click="closeShareModal">关闭</button>
    </template>
  </AppModal>
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

/* ---- 分类 ---- */
.pages__doc-category {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  font-size: 13px;
}

.pages__doc-category-label {
  color: var(--text-tertiary);
  font-weight: 400;
  font-size: 12px;
}

.pages__doc-category-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 500;
}

.pages__doc-category-badge--empty {
  background: var(--surface-3);
  color: var(--text-tertiary);
}

.pages__doc-category-edit {
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-select {
  padding: 4px 8px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
}

/* ---- 版本历史 ---- */
.pages__versions {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.pages__versions h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.version-item:hover { background: var(--surface-2); }
.version-item--active { background: var(--brand-50); }

.version-item__number {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--brand-500);
  min-width: 36px;
}

.version-item__time {
  flex: 1;
  color: var(--text-tertiary);
  font-size: 11px;
}

.version-preview {
  margin-top: 12px;
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.version-preview__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-subtle);
}

.version-preview__content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
  max-height: 300px;
  overflow-y: auto;
}

/* ---- 关联工作项 ---- */
.pages__links {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.pages__links h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.links-add {
  display: flex;
  gap: 8px;
  align-items: center;
}

.form-input {
  padding: 5px 8px;
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

.links-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
}

.link-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.link-item:hover { background: var(--surface-2); }

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
  text-decoration: none;
  font-family: var(--font-mono);
  font-weight: 500;
}

.link-item__id:hover { text-decoration: underline; }

.link-item__remove {
  margin-left: auto;
  opacity: 0;
  transition: opacity 0.15s;
}

.link-item:hover .link-item__remove { opacity: 1; }

.btn--sm {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
  font-family: inherit;
}

.btn--ghost {
  background: none;
  border: none;
  color: var(--brand-500);
  padding: 2px 4px;
}

.btn--primary {
  background: var(--brand-500);
  color: var(--text-on-brand);
  border-color: var(--brand-500);
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 12px;
}

/* ---- 分享按钮 ---- */
.pages__share-btn {
  margin-left: auto;
}

/* ---- 模板列表 ---- */
.template-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.template-list__group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.template-list__cat {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 2px;
}

.template-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  transition: background 0.1s;
}

.template-item:hover {
  background: var(--surface-2);
}

.template-item__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
  margin-right: 12px;
}

.template-item__name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.template-item__desc {
  font-size: 11px;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---- 分享链接 ---- */
.shares-create {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.shares-create h4,
.shares-list-section h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.shares-create__row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.shares-list-section {
  margin-top: 8px;
}

.shares-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.share-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  gap: 8px;
}

.share-row--inactive {
  opacity: 0.6;
}

.share-row__info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  flex: 1;
  min-width: 0;
}

.share-row__token {
  font-family: var(--font-mono);
  color: var(--text-primary);
  min-width: 80px;
}

.share-row__status {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  font-weight: 500;
}

.share-row__status--active {
  background: var(--success-50, #e6f4ea);
  color: var(--success-600, #1e8e3e);
}

.share-row__status--inactive {
  background: var(--surface-3);
  color: var(--text-tertiary);
}

.share-row__status--expired {
  background: var(--warning-50, #fef7e0);
  color: var(--warning-600, #b06000);
}

.share-row__expiry {
  color: var(--text-tertiary);
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-row__actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.btn--danger {
  color: var(--danger-500);
  border-color: var(--danger-500);
}

.btn--danger:hover {
  background: var(--danger-50);
}
</style>
