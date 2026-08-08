<script setup lang="ts">
/**
 * KnowledgePageDetailView — 单篇文档详情 + Tiptap Markdown 编辑器。
 *
 * 功能：
 *   - Tiptap Markdown 编辑器（@tiptap/extension-markdown）+ 实时预览切换
 *   - 标题行内编辑、状态 badge（draft / published / archived）
 *   - 版本历史 tab：快照列表 + 查看 / 回滚
 *   - 关联工作项 tab：搜索工作项 + 增删关联
 *   - 乐观锁 version 字段（PATCH 时携带当前 version）
 */
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useEditor, EditorContent } from "@tiptap/vue-3";

import { completeExtensions } from "@/lib/editor/extensions";
import {
  knowledgeApi,
  type KnowledgePage,
  type KnowledgePageRelation,
  type KnowledgePageVersion,
  type PageStatus,
} from "@/api/services/knowledge";
import { searchApi } from "@/api/services/search";
import { renderMarkdown } from "@/lib/markdown";
import { toast } from "@/lib/toast";

/* ===== Props / Emit ===== */
const props = defineProps<{
  page: KnowledgePage;
  workspaceId: number;
  spaceId: number;
}>();

const emit = defineEmits<{
  (e: "deleted", id: number): void;
  (e: "saved", page: KnowledgePage): void;
}>();

/* ===== 编辑器状态 ===== */
const editorMd = ref(props.page.content_md ?? "");
const dirty = ref(false);
const saving = ref(false);
const previewMode = ref(false);

const editor = useEditor({
  extensions: completeExtensions("使用 Markdown 撰写文档..."),
  content: editorMd.value,
  editable: true,
  editorProps: {
    attributes: {
      class: "km-editor__content",
    },
  },
  onUpdate({ editor: ed }) {
    editorMd.value = ed.getHTML();
    dirty.value = true;
  },
});

/* 同步外部 content_md 变化 */
watch(
  () => props.page.content_md,
  (val) => {
    if (editor.value && val !== undefined && val !== editor.value.getHTML()) {
      editor.value.commands.setContent(val, { emitUpdate: false });
      editorMd.value = val;
      dirty.value = false;
    }
  },
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});

/* ===== 状态 ===== */
const statusLabels: Record<PageStatus, string> = {
  draft: "草稿",
  published: "已发布",
  archived: "已归档",
};

/* ===== 保存 ===== */
async function savePage() {
  if (!dirty.value) return;
  saving.value = true;
  try {
    const updated = await knowledgeApi.updatePage(
      props.workspaceId, props.spaceId, props.page.id,
      {
        title: props.page.title,
        content_md: editorMd.value,
        version: props.page.version,
      },
    );
    dirty.value = false;
    toast.success("保存成功，版本 v" + updated.version);
    emit("saved", updated);
    await loadVersions();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

/* ===== 状态切换 ===== */
async function changeStatus(status: PageStatus) {
  if (props.page.status === status) return;
  const confirms: Partial<Record<PageStatus, string>> = {
    published: `确定发布文档「${props.page.title}」？`,
    archived: `确定归档文档「${props.page.title}」？归档后仅可浏览。`,
    draft: `确定将文档「${props.page.title}」转为草稿？`,
  };
  if (confirms[status] && !window.confirm(confirms[status])) return;
  try {
    const updated = await knowledgeApi.updatePage(
      props.workspaceId, props.spaceId, props.page.id,
      { status, version: props.page.version },
    );
    emit("saved", updated);
    toast.success("状态已更新为「" + statusLabels[status] + "」");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "状态更新失败");
  }
}

/* ===== 版本历史 ===== */
const versions = ref<KnowledgePageVersion[]>([]);
const versionsLoading = ref(false);
const viewingVersion = ref<KnowledgePageVersion | null>(null);

async function loadVersions() {
  versionsLoading.value = true;
  try {
    versions.value = await knowledgeApi.listVersions(
      props.workspaceId, props.spaceId, props.page.id,
    );
  } catch {
    versions.value = [];
  } finally {
    versionsLoading.value = false;
  }
}

function viewVersion(v: KnowledgePageVersion) {
  viewingVersion.value = v;
  previewMode.value = true;
}

function exitVersionPreview() {
  viewingVersion.value = null;
  previewMode.value = false;
}

async function revertToVersion(v: KnowledgePageVersion) {
  if (!window.confirm(`确定回滚到版本 v${v.version}？当前未保存的更改将丢失。`)) return;
  try {
    const updated = await knowledgeApi.revertVersion(
      props.workspaceId, props.spaceId, props.page.id, v.version,
    );
    editorMd.value = updated.content_md ?? "";
    if (editor.value) {
      editor.value.commands.setContent(updated.content_md ?? "", { emitUpdate: false });
    }
    dirty.value = false;
    exitVersionPreview();
    emit("saved", updated);
    toast.success(`已回滚到版本 v${v.version}`);
    await loadVersions();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "回滚失败");
  }
}

/* ===== 侧边栏 Tab ===== */
const activeTab = ref<"history" | "relations">("history");

/* ===== 关联工作项 ===== */
const relations = ref<KnowledgePageRelation[]>([]);
const relationsLoading = ref(false);
const relationSearch = ref("");
const relationResults = ref<
  Array<{ id: number; identifier: string; name: string; project_id: number; project_name: string }>
>([]);
const relationSearchOpen = ref(false);
const addingRelation = ref(false);

async function loadRelations() {
  relationsLoading.value = true;
  try {
    relations.value = await knowledgeApi.listRelations(
      props.workspaceId, props.spaceId, props.page.id,
    );
  } catch {
    relations.value = [];
  } finally {
    relationsLoading.value = false;
  }
}

async function searchIssues() {
  if (!relationSearch.value.trim()) {
    relationResults.value = [];
    return;
  }
  try {
    const res = await searchApi.searchWorkspace(props.workspaceId, {
      q: relationSearch.value.trim(),
      types: "issue",
      limit: 8,
    });
    relationResults.value = (res.results.issues ?? []).map((i) => ({
      id: i.id,
      identifier: i.identifier ?? `#${i.id}`,
      name: i.name,
      project_id: i.project_id,
      project_name: i.project_name ?? "",
    }));
  } catch {
    relationResults.value = [];
  }
}

async function addRelation(issueId: number) {
  addingRelation.value = true;
  try {
    await knowledgeApi.addRelation(props.workspaceId, props.spaceId, props.page.id, issueId);
    relationSearch.value = "";
    relationResults.value = [];
    relationSearchOpen.value = false;
    toast.success("关联已添加");
    await loadRelations();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "添加关联失败");
  } finally {
    addingRelation.value = false;
  }
}

async function removeRelation(rel: KnowledgePageRelation) {
  if (!window.confirm("确定删除该关联？")) return;
  try {
    await knowledgeApi.removeRelation(props.workspaceId, props.spaceId, props.page.id, rel.id);
    relations.value = relations.value.filter((r) => r.id !== rel.id);
    toast.success("关联已删除");
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除关联失败");
  }
}

/* ===== 工具 ===== */
function fmtTime(iso: string): string {
  const d = new Date(iso);
  return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, "0")}/${String(
    d.getDate(),
  ).padStart(2, "0")} ${String(d.getHours()).padStart(2, "0")}:${String(
    d.getMinutes(),
  ).padStart(2, "0")}`;
}

onMounted(() => {
  loadVersions();
  loadRelations();
});
</script>

<template>
  <div class="kpage">
    <!-- 标题栏 -->
    <header class="kpage__header">
      <span class="kpage__title">{{ page.title }}</span>
      <div class="kpage__meta">
        <span
          class="status-badge"
          :class="`status-badge--${page.status}`"
        >{{ statusLabels[page.status] }}</span>
        <span class="kpage__version">v{{ page.version }}</span>
        <span v-if="dirty" class="kpage__dirty-dot" title="未保存">*</span>
      </div>
    </header>

    <!-- 状态操作栏 -->
    <div class="kpage__toolbar">
      <div class="kpage__status-actions">
        <button
          v-if="page.status !== 'published'"
          class="btn btn--sm"
          @click="changeStatus('published')"
        >发布</button>
        <button
          v-if="page.status !== 'draft'"
          class="btn btn--sm"
          @click="changeStatus('draft')"
        >转为草稿</button>
        <button
          v-if="page.status !== 'archived'"
          class="btn btn--sm"
          @click="changeStatus('archived')"
        >归档</button>
        <button
          v-if="page.status === 'archived'"
          class="btn btn--sm"
          @click="changeStatus('published')"
        >重新发布</button>
      </div>
      <div class="kpage__actions">
        <span v-if="saving" class="kpage__saving">保存中...</span>
        <button
          class="btn btn--sm btn--primary"
          :disabled="saving || !dirty"
          @click="savePage"
        >{{ saving ? "保存中..." : "保存" }}</button>
      </div>
    </div>

    <!-- 编辑器 / 预览 -->
    <div class="kpage__editor-wrap">
      <!-- 编辑工具栏 -->
      <div v-if="!previewMode && editor" class="km-toolbar">
        <button
          class="km-toolbar__btn"
          title="粗体"
          @click="editor.chain().focus().toggleBold().run()"
        ><strong>B</strong></button>
        <button
          class="km-toolbar__btn"
          title="斜体"
          @click="editor.chain().focus().toggleItalic().run()"
        ><em>I</em></button>
        <button
          class="km-toolbar__btn"
          title="标题 2"
          @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
        >H2</button>
        <button
          class="km-toolbar__btn"
          title="标题 3"
          @click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
        >H3</button>
        <button
          class="km-toolbar__btn"
          title="无序列表"
          @click="editor.chain().focus().toggleBulletList().run()"
        >•≡</button>
        <button
          class="km-toolbar__btn"
          title="有序列表"
          @click="editor.chain().focus().toggleOrderedList().run()"
        >1≡</button>
        <button
          class="km-toolbar__btn"
          title="任务列表"
          @click="editor.chain().focus().toggleTaskList().run()"
        >☑</button>
        <button
          class="km-toolbar__btn"
          title="代码块"
          @click="editor.chain().focus().toggleCodeBlock().run()"
        >&lt;/&gt;</button>
        <button
          class="km-toolbar__btn"
          title="引用"
          @click="editor.chain().focus().toggleBlockquote().run()"
        >❝</button>
      </div>

      <!-- 预览模式标题栏 -->
      <div v-if="previewMode" class="km-preview-bar">
        <span v-if="viewingVersion">
          查看版本 v{{ viewingVersion.version }}
          <template v-if="viewingVersion.change_summary"> · {{ viewingVersion.change_summary }}</template>
          · {{ fmtTime(viewingVersion.created_at) }}
        </span>
        <span v-else>预览模式</span>
        <button class="btn btn--sm" @click="exitVersionPreview">返回编辑</button>
      </div>

      <!-- 编辑区 / 预览区 -->
      <div class="km-editor">
        <EditorContent
          v-if="!previewMode && editor"
          :editor="editor"
          class="km-editor__tiptap"
        />
        <div
          v-else
          class="km-editor__preview"
          v-html="renderMarkdown(viewingVersion?.content_md ?? editorMd)"
        />
      </div>
    </div>

    <!-- 侧边栏 tabs -->
    <div class="kpage__sidebar">
      <div class="kpage__tabs">
        <button
          class="kpage__tab"
          :class="{ 'kpage__tab--active': activeTab === 'history' }"
          @click="activeTab = 'history'"
        >历史版本</button>
        <button
          class="kpage__tab"
          :class="{ 'kpage__tab--active': activeTab === 'relations' }"
          @click="activeTab = 'relations'"
        >关联工作项</button>
      </div>

      <!-- 历史版本 -->
      <div v-if="activeTab === 'history'" class="kpage__tab-content">
        <div v-if="versionsLoading" class="text-muted">加载中...</div>
        <div v-else-if="versions.length === 0" class="text-muted">暂无历史版本</div>
        <div v-else class="versions-list">
          <div
            v-for="v in versions"
            :key="v.id"
            class="version-item"
            :class="{ 'version-item--active': viewingVersion?.id === v.id }"
          >
            <span class="version-item__number">v{{ v.version }}</span>
            <span class="version-item__time">{{ fmtTime(v.created_at) }}</span>
            <span class="version-item__summary">{{ v.change_summary || "—" }}</span>
            <button class="btn btn--sm btn--ghost" @click="viewVersion(v)">查看</button>
            <button
              class="btn btn--sm btn--ghost"
              @click="revertToVersion(v)"
            >回滚</button>
          </div>
        </div>
      </div>

      <!-- 关联工作项 -->
      <div v-if="activeTab === 'relations'" class="kpage__tab-content">
        <div class="rel-search">
          <input
            v-model="relationSearch"
            type="text"
            class="form-input"
            placeholder="搜索工作项关键字"
            @input="searchIssues"
            @focus="relationSearchOpen = true"
          />
        </div>

        <!-- 搜索结果 -->
        <div v-if="relationSearchOpen && relationResults.length > 0" class="rel-results">
          <div
            v-for="r in relationResults"
            :key="r.id"
            class="rel-result"
          >
            <div class="rel-result__info">
              <span class="rel-result__id">{{ r.identifier }}</span>
              <span class="rel-result__name" :title="r.name">{{ r.name }}</span>
              <span class="rel-result__proj">{{ r.project_name }}</span>
            </div>
            <button
              class="btn btn--sm btn--primary"
              :disabled="addingRelation"
              @click="addRelation(r.id)"
            >＋</button>
          </div>
        </div>

        <!-- 已关联列表 -->
        <div v-if="relationsLoading" class="text-muted" style="margin-top: 8px">加载中...</div>
        <div v-else-if="relations.length === 0" class="text-muted" style="margin-top: 8px">暂无关联</div>
        <div v-else class="rel-list">
          <div v-for="rel in relations" :key="rel.id" class="rel-item">
            <span class="rel-item__type">工作项 #{{ rel.issue_id }}</span>
            <button class="btn btn--sm btn--ghost rel-item__remove" @click="removeRelation(rel)">✕</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ---- 标题栏 ---- */
.kpage__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}
.kpage__title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
}
.kpage__meta {
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
.kpage__version {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  color: var(--text-tertiary);
}
.kpage__dirty-dot {
  font-size: 16px;
  color: var(--warning-500, #f59e0b);
}

/* ---- 工具栏 ---- */
.kpage__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.kpage__status-actions {
  display: flex;
  gap: 6px;
}
.kpage__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.kpage__saving {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* ---- 编辑器区域 ---- */
.kpage__editor-wrap {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
  margin-bottom: 16px;
}

/* toolbar */
.km-toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px 8px;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}
.km-toolbar__btn {
  min-width: 26px;
  height: 26px;
  padding: 0 5px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-family: inherit;
}
.km-toolbar__btn:hover {
  background: var(--surface-3);
  border-color: var(--border-default);
}

/* preview bar */
.km-preview-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--warning-50, #fef7e0);
  border-bottom: 1px solid var(--warning-300, #fcd34d);
  font-size: 12px;
  color: var(--warning-700, #b45309);
}

/* editor */
.km-editor {
  min-height: 40vh;
}
.km-editor__tiptap {
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.75;
}
.km-editor__tiptap :deep(.ProseMirror) {
  outline: none;
  color: var(--text-primary);
}
.km-editor__tiptap :deep(.ProseMirror h2) {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 8px;
}
.km-editor__tiptap :deep(.ProseMirror h3) {
  font-size: 16px;
  font-weight: 600;
  margin: 12px 0 6px;
}
.km-editor__tiptap :deep(.ProseMirror p) {
  margin: 0 0 8px;
}
.km-editor__tiptap :deep(.ProseMirror ul),
.km-editor__tiptap :deep(.ProseMirror ol) {
  padding-left: 24px;
  margin: 4px 0 8px;
}
.km-editor__tiptap :deep(.ProseMirror blockquote) {
  border-left: 3px solid var(--brand-300, #9ec5ff);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--text-tertiary);
}
.km-editor__tiptap :deep(.ProseMirror code) {
  background: var(--surface-3);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: var(--font-mono, monospace);
}
.km-editor__tiptap :deep(.ProseMirror pre) {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 12px;
  margin: 8px 0;
  overflow-x: auto;
}
.km-editor__tiptap :deep(.ProseMirror pre code) {
  background: none;
  padding: 0;
}
.km-editor__tiptap :deep(.ProseMirror ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;
}
.km-editor__tiptap :deep(.ProseMirror ul[data-type="taskList"] li) {
  display: flex;
  gap: 6px;
  align-items: flex-start;
}
.km-editor__tiptap :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: var(--text-tertiary);
  pointer-events: none;
  height: 0;
}

/* preview */
.km-editor__preview {
  min-height: 40vh;
  max-height: 60vh;
  overflow-y: auto;
  padding: 14px 16px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-secondary);
  word-break: break-word;
}
.km-editor__preview :deep(h1),
.km-editor__preview :deep(h2),
.km-editor__preview :deep(h3) {
  color: var(--text-primary);
  margin: 1.2em 0 0.5em;
  line-height: 1.35;
}
.km-editor__preview :deep(h1) { font-size: 22px; border-bottom: 1px solid var(--border-subtle); padding-bottom: 8px; }
.km-editor__preview :deep(h2) { font-size: 18px; }
.km-editor__preview :deep(h3) { font-size: 16px; }
.km-editor__preview :deep(p) { margin: 0.6em 0; }
.km-editor__preview :deep(a) { color: var(--brand-500); }
.km-editor__preview :deep(code) {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  background: var(--surface-3);
  padding: 1px 5px;
  border-radius: 4px;
}
.km-editor__preview :deep(pre) {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 12px;
  overflow-x: auto;
}
.km-editor__preview :deep(pre code) { background: none; padding: 0; font-size: 12px; line-height: 1.6; }
.km-editor__preview :deep(blockquote) {
  margin: 0.8em 0;
  padding: 4px 14px;
  border-left: 3px solid var(--brand-300, #9ec5ff);
  color: var(--text-tertiary);
  background: var(--surface-2);
}
.km-editor__preview :deep(ul),
.km-editor__preview :deep(ol) { padding-left: 24px; margin: 0.6em 0; }
.km-editor__preview :deep(table) { border-collapse: collapse; margin: 0.8em 0; width: 100%; font-size: 13px; }
.km-editor__preview :deep(th),
.km-editor__preview :deep(td) { border: 1px solid var(--border-default); padding: 6px 10px; }
.km-editor__preview :deep(th) { background: var(--surface-2); font-weight: 600; }
.km-editor__preview :deep(img) { max-width: 100%; border-radius: var(--radius-sm); }
.km-editor__preview :deep(hr) { border: none; border-top: 1px solid var(--border-subtle); margin: 1.2em 0; }

/* ---- 侧边栏 ---- */
.kpage__sidebar {
  border-top: 1px solid var(--border-subtle);
  padding-top: 16px;
}
.kpage__tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}
.kpage__tab {
  padding: 6px 14px;
  font-size: 12px;
  font-family: inherit;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}
.kpage__tab--active {
  color: var(--brand-600);
  border-bottom-color: var(--brand-500);
  font-weight: 500;
}
.kpage__tab:hover {
  color: var(--text-primary);
}

.kpage__tab-content {
  max-height: 240px;
  overflow-y: auto;
}

/* 版本列表 */
.versions-list,
.rel-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.version-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}
.version-item:hover { background: var(--surface-2); }
.version-item--active { background: var(--warning-50, #fef7e0); }
.version-item__number {
  font-family: var(--font-mono, monospace);
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

/* 关联 */
.rel-search { margin-bottom: 8px; }
.rel-results {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}
.rel-result {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 12px;
}
.rel-result:hover { background: var(--surface-2); }
.rel-result__info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}
.rel-result__id {
  font-family: var(--font-mono, monospace);
  font-weight: 500;
  color: var(--brand-500);
  flex-shrink: 0;
}
.rel-result__name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary);
}
.rel-result__proj {
  font-size: 11px;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.rel-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
}
.rel-item:hover { background: var(--surface-2); }
.rel-item__type {
  font-family: var(--font-mono, monospace);
  color: var(--brand-500);
  font-weight: 500;
}
.rel-item__remove { opacity: 0; transition: opacity 0.15s; }
.rel-item:hover .rel-item__remove { opacity: 1; }

/* ---- 通用 ---- */
.text-muted {
  color: var(--text-tertiary);
  font-size: 12px;
}
.form-input {
  width: 100%;
  padding: 6px 10px;
  font-size: 12px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  box-sizing: border-box;
}
.form-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

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
.btn:hover { border-color: var(--brand-500); color: var(--brand-600); }
.btn--sm { padding: 4px 10px; font-size: 12px; }
.btn--ghost { background: none; border: none; color: var(--brand-500); padding: 2px 4px; }
.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 500;
}
.btn--primary:hover { background: var(--brand-600); color: var(--text-on-brand); }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
