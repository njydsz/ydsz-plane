<script setup lang="ts">
/**
 * RichTextEditor — 基于 TipTap 3 的通用富文本编辑器组件。
 *
 * 特性：
 *   - v-model:contentHTML / v-model:contentJSON 双向绑定
 *   - 菜单栏（粗体/斜体/链接/图片/代码/标题/列表）
 *   - @提及 扩展（需传入 mentionSuggestions）
 *   - 图片粘贴回调 @paste-image
 *
 * 使用：
 *   <RichTextEditor v-model:content-html="html" v-model:content-json="json" />
 */

import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Image from "@tiptap/extension-image";
import Placeholder from "@tiptap/extension-placeholder";
import Mention from "@tiptap/extension-mention";
import { computed, onBeforeUnmount, watch } from "vue";

const props = withDefaults(
  defineProps<{
    contentHtml?: string;
    contentJson?: string;
    placeholder?: string;
    editable?: boolean;
    mentionSuggestions?: Array<{ id: number | string; label: string }>;
    minHeight?: string;
    /** 精简模式：隐藏菜单栏（用于评论等紧凑场景） */
    compact?: boolean;
  }>(),
  {
    contentHtml: "",
    contentJson: "",
    placeholder: "输入内容...",
    editable: true,
    mentionSuggestions: () => [],
    minHeight: "120px",
    compact: false,
  },
);

const emit = defineEmits<{
  "update:contentHtml": [value: string];
  "update:contentJson": [value: string];
  "paste-image": [file: File];
}>();

/* ---- 配置 mention 扩展的 suggestion ---- */
const mentionItems = computed(() => {
  const { mentionSuggestions } = props;
  return ({ query }: { query: string }) => {
    if (!query) return mentionSuggestions.slice(0, 8);
    return mentionSuggestions
      .filter((i) => i.label.toLowerCase().includes(query.toLowerCase()))
      .slice(0, 8);
  };
});

/* ---- 初始化编辑器 ---- */
const editor = useEditor({
  content: props.contentHtml || props.contentJson || "",
  editable: props.editable,
  extensions: [
    StarterKit.configure({
      heading: { levels: [2, 3] },
    }),
    Underline,
    Link.configure({ openOnClick: false }),
    Image.configure({ inline: true, allowBase64: true }),
    Placeholder.configure({ placeholder: props.placeholder }),
    ...(props.mentionSuggestions.length > 0
      ? [Mention.configure({ suggestion: { items: mentionItems.value } })]
      : []),
  ],
  editorProps: {
    handlePaste(_view, event) {
      const items = event.clipboardData?.items;
      if (!items) return false;
      for (const item of Array.from(items)) {
        if (item.type.startsWith("image/")) {
          const file = item.getAsFile();
          if (file) {
            emit("paste-image", file);
            return true;
          }
        }
      }
      return false;
    },
  },
  onUpdate({ editor: ed }) {
    const html = ed.getHTML();
    const json = ed.getJSON();
    emit("update:contentHtml", html);
    emit("update:contentJson", JSON.stringify(json));
  },
});

/* ---- 外部内容变化时同步到编辑器 ---- */
watch(
  () => props.contentHtml,
  (val) => {
    if (!editor.value) return;
    const current = editor.value.getHTML();
    if (val !== current) {
      editor.value.commands.setContent(val || "");
    }
  },
);

watch(
  () => props.editable,
  (val) => {
    editor.value?.setEditable(val);
  },
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});

/* ---- 菜单命令 ---- */
function toggleBold() {
  editor.value?.chain().focus().toggleBold().run();
}
function toggleItalic() {
  editor.value?.chain().focus().toggleItalic().run();
}
function toggleUnderline() {
  editor.value?.chain().focus().toggleUnderline().run();
}
function toggleStrike() {
  editor.value?.chain().focus().toggleStrike().run();
}
function toggleCode() {
  editor.value?.chain().focus().toggleCode().run();
}
function setHeading(level: 2 | 3) {
  editor.value?.chain().focus().toggleHeading({ level }).run();
}
function toggleBulletList() {
  editor.value?.chain().focus().toggleBulletList().run();
}
function toggleOrderedList() {
  editor.value?.chain().focus().toggleOrderedList().run();
}
function toggleBlockquote() {
  editor.value?.chain().focus().toggleBlockquote().run();
}
function addLink() {
  const url = prompt("输入链接 URL:");
  if (url) {
    editor.value?.chain().focus().setLink({ href: url }).run();
  }
}
function insertImage() {
  const url = prompt("输入图片 URL:");
  if (url) {
    editor.value?.chain().focus().setImage({ src: url }).run();
  }
}

/* ---- 激活状态 ---- */
const isActiveBold = computed(() => editor.value?.isActive("bold") ?? false);
const isActiveItalic = computed(() => editor.value?.isActive("italic") ?? false);
const isActiveUnderline = computed(() => editor.value?.isActive("underline") ?? false);
const isActiveStrike = computed(() => editor.value?.isActive("strike") ?? false);
const isActiveCode = computed(() => editor.value?.isActive("code") ?? false);
const isActiveHeading2 = computed(() => editor.value?.isActive("heading", { level: 2 }) ?? false);
const isActiveHeading3 = computed(() => editor.value?.isActive("heading", { level: 3 }) ?? false);
const isActiveBulletList = computed(() => editor.value?.isActive("bulletList") ?? false);
const isActiveOrderedList = computed(() => editor.value?.isActive("orderedList") ?? false);
const isActiveBlockquote = computed(() => editor.value?.isActive("blockquote") ?? false);
const isActiveLink = computed(() => editor.value?.isActive("link") ?? false);

defineExpose({ editor });
</script>

<template>
  <div class="rich-editor" :class="{ 'rich-editor--compact': compact, 'rich-editor--readonly': !editable }">
    <!-- 菜单栏 -->
    <div v-if="editable && !compact" class="rich-editor__toolbar">
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveBold }"
        title="粗体 (Ctrl+B)"
        @click="toggleBold"
      >B</button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveItalic }"
        title="斜体 (Ctrl+I)"
        @click="toggleItalic"
      ><em>I</em></button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveUnderline }"
        title="下划线 (Ctrl+U)"
        @click="toggleUnderline"
      ><u>U</u></button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveStrike }"
        title="删除线"
        @click="toggleStrike"
      ><s>S</s></button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveCode }"
        title="行内代码"
        @click="toggleCode"
      >{ }</button>

      <span class="rich-editor__divider"></span>

      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveHeading2 }"
        title="标题 2"
        @click="setHeading(2)"
      >H2</button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveHeading3 }"
        title="标题 3"
        @click="setHeading(3)"
      >H3</button>

      <span class="rich-editor__divider"></span>

      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveBulletList }"
        title="无序列表"
        @click="toggleBulletList"
      >•≡</button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveOrderedList }"
        title="有序列表"
        @click="toggleOrderedList"
      >1≡</button>
      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveBlockquote }"
        title="引用"
        @click="toggleBlockquote"
      >❝</button>

      <span class="rich-editor__divider"></span>

      <button
        class="rich-editor__btn"
        :class="{ 'rich-editor__btn--active': isActiveLink }"
        title="链接"
        @click="addLink"
      >🔗</button>
      <button
        class="rich-editor__btn"
        title="图片"
        @click="insertImage"
      >🖼</button>
    </div>

    <!-- 编辑区域 -->
    <EditorContent
      :editor="editor"
      class="rich-editor__content"
      :style="{ minHeight: compact ? '80px' : minHeight }"
    />
  </div>
</template>

<style scoped>
.rich-editor {
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1, #fff);
  overflow: hidden;
  transition: border-color 0.15s;
}

.rich-editor:focus-within {
  border-color: var(--brand-500, #3b82f6);
  box-shadow: 0 0 0 3px var(--brand-50, rgba(59,130,246,0.1));
}

.rich-editor--readonly {
  border: none;
  background: transparent;
}

.rich-editor--readonly:focus-within {
  box-shadow: none;
}

/* ---- Toolbar ---- */
.rich-editor__toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  background: var(--surface-2, #f9fafb);
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  flex-wrap: wrap;
}

.rich-editor__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #4b5563);
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.1s;
}

.rich-editor__btn:hover {
  background: var(--surface-3, #f3f4f6);
  border-color: var(--border-default, #d1d5db);
}

.rich-editor__btn--active {
  background: var(--brand-50, rgba(59,130,246,0.08));
  color: var(--brand-500, #3b82f6);
  border-color: var(--brand-200, #bfdbfe);
}

.rich-editor__divider {
  width: 1px;
  height: 20px;
  background: var(--border-subtle, #e5e7eb);
  margin: 0 4px;
}

/* ---- Content ---- */
.rich-editor__content {
  padding: 12px;
}

.rich-editor__content :deep(.ProseMirror) {
  outline: none;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary, #1f2937);
}

.rich-editor--readonly .rich-editor__content {
  padding: 0;
}

.rich-editor--readonly .rich-editor__content :deep(.ProseMirror) {
  padding: 0;
}

.rich-editor__content :deep(.ProseMirror p) {
  margin: 0 0 8px;
}

.rich-editor__content :deep(.ProseMirror p:last-child) {
  margin-bottom: 0;
}

.rich-editor__content :deep(.ProseMirror h2) {
  font-size: 18px;
  font-weight: 600;
  margin: 16px 0 8px;
  color: var(--text-primary, #1f2937);
}

.rich-editor__content :deep(.ProseMirror h3) {
  font-size: 16px;
  font-weight: 600;
  margin: 12px 0 6px;
}

.rich-editor__content :deep(.ProseMirror ul),
.rich-editor__content :deep(.ProseMirror ol) {
  padding-left: 24px;
  margin: 4px 0 8px;
}

.rich-editor__content :deep(.ProseMirror li) {
  margin-bottom: 4px;
}

.rich-editor__content :deep(.ProseMirror blockquote) {
  border-left: 3px solid var(--brand-200, #bfdbfe);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--text-secondary, #4b5563);
}

.rich-editor__content :deep(.ProseMirror code) {
  background: var(--surface-3, #f3f4f6);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: var(--font-mono, 'Consolas', monospace);
}

.rich-editor__content :deep(.ProseMirror a) {
  color: var(--brand-500, #3b82f6);
  text-decoration: underline;
}

.rich-editor__content :deep(.ProseMirror img) {
  max-width: 100%;
  border-radius: var(--radius-sm, 4px);
  margin: 8px 0;
}

.rich-editor__content :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: var(--text-tertiary, #9ca3af);
  pointer-events: none;
  height: 0;
}

/* Mention */
.rich-editor__content :deep(.ProseMirror .mention) {
  background: var(--brand-50, rgba(59,130,246,0.08));
  color: var(--brand-600, #2563eb);
  padding: 1px 4px;
  border-radius: 3px;
  font-weight: 500;
  font-size: 13px;
}
</style>
