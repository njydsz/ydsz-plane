<script setup lang="ts">
/**
 * RichTextEditor — 基于 TipTap 3 的通用富文本编辑器组件 v3
 *
 * v3 增强：
 *   — 斜杠命令浮层（slash commands）
 *   — Callout 提示框工具
 *   — 表格插入 + 列/行操作工具栏
 *   — 颜色选择器（文本颜色 + 高亮）
 *
 * 特性：
 *   - v-model:contentHTML / v-model:contentJSON 双向绑定
 *   - 菜单栏（粗体/斜体/链接/图片/代码/标题/列表/表格/颜色/callout）
 *   - @提及 扩展（需传入 mentionSuggestions）
 *   - 图片粘贴回调 @paste-image
 */

import { useEditor, EditorContent } from "@tiptap/vue-3"
import {
  completeExtensions,
  commentExtensions,
  compactExtensions,
  slashItems,
  type SlashCommandItem,
} from "@/lib/editor/extensions"
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from "vue"
import Mention from "@tiptap/extension-mention"

const props = withDefaults(
  defineProps<{
    contentHtml?: string
    contentJson?: string
    placeholder?: string
    editable?: boolean
    mentionSuggestions?: Array<{ id: number | string; label: string }>
    minHeight?: string
    /** 编辑器变体 — full(默认) / comment / compact */
    variant?: "full" | "comment" | "compact"
    /** 精简模式：隐藏菜单栏（用于紧凑场景，已废弃，用 variant='compact' 替代） */
    compact?: boolean
  }>(),
  {
    contentHtml: "",
    contentJson: "",
    placeholder: "输入内容...",
    editable: true,
    mentionSuggestions: () => [],
    minHeight: "120px",
    variant: "full",
    compact: false,
  },
)

const emit = defineEmits<{
  "update:contentHtml": [value: string]
  "update:contentJson": [value: string]
  "paste-image": [file: File]
}>()

/* ---- 配置 mention 扩展的 suggestion ---- */
const mentionItems = computed(() => {
  const { mentionSuggestions } = props
  return ({ query }: { query: string }) => {
    if (!query) return mentionSuggestions.slice(0, 8)
    return mentionSuggestions
      .filter((i) => i.label.toLowerCase().includes(query.toLowerCase()))
      .slice(0, 8)
  }
})

/* ---- 根据 variant 选择扩展 ---- */
const editorExtensions = computed(() => {
  let exts: any[]
  switch (props.variant) {
    case "comment":
      exts = commentExtensions(props.placeholder)
      break
    case "compact":
      exts = compactExtensions(props.placeholder)
      break
    case "full":
    default:
      exts = completeExtensions(props.placeholder)
  }
  // 注入 mention（如提供）
  if (props.mentionSuggestions.length > 0) {
    exts = [
      ...exts.filter((e) => e?.name !== "mention"),
      Mention.configure({ suggestion: { items: mentionItems.value } }),
    ]
  }
  return exts
})

/* ---- 初始化编辑器 ---- */
const editor = useEditor({
  content: props.contentHtml || props.contentJson || "",
  editable: props.editable,
  extensions: editorExtensions.value,
  editorProps: {
    handlePaste(_view, event) {
      const items = event.clipboardData?.items
      if (!items) return false
      for (const item of Array.from(items)) {
        if (item.type.startsWith("image/")) {
          const file = item.getAsFile()
          if (file) {
            emit("paste-image", file)
            return true
          }
        }
      }
      return false
    },
  },
  onUpdate({ editor: ed }) {
    const html = ed.getHTML()
    const json = ed.getJSON()
    emit("update:contentHtml", html)
    emit("update:contentJson", JSON.stringify(json))
  },
})

/* ---- 外部内容变化时同步到编辑器 ---- */
watch(
  () => props.contentHtml,
  (val) => {
    if (!editor.value) return
    const current = editor.value.getHTML()
    if (val !== current) {
      editor.value.commands.setContent(val || "")
    }
  },
)

watch(
  () => props.editable,
  (val) => {
    editor.value?.setEditable(val)
  },
)

onBeforeUnmount(() => {
  editor.value?.destroy()
})

/* ---- 菜单命令 ---- */
function toggleBold() {
  editor.value?.chain().focus().toggleBold().run()
}
function toggleItalic() {
  editor.value?.chain().focus().toggleItalic().run()
}
function toggleUnderline() {
  editor.value?.chain().focus().toggleUnderline().run()
}
function toggleStrike() {
  editor.value?.chain().focus().toggleStrike().run()
}
function toggleCode() {
  editor.value?.chain().focus().toggleCode().run()
}
function setHeading(level: 2 | 3) {
  editor.value?.chain().focus().toggleHeading({ level }).run()
}
function toggleBulletList() {
  editor.value?.chain().focus().toggleBulletList().run()
}
function toggleOrderedList() {
  editor.value?.chain().focus().toggleOrderedList().run()
}
function toggleBlockquote() {
  editor.value?.chain().focus().toggleBlockquote().run()
}
function toggleTaskList() {
  editor.value?.chain().focus().toggleTaskList().run()
}
function toggleCodeBlock() {
  editor.value?.chain().focus().toggleCodeBlock().run()
}
function addLink() {
  const url = prompt("输入链接 URL:")
  if (url) {
    editor.value?.chain().focus().setLink({ href: url }).run()
  }
}
function insertImage() {
  const url = prompt("输入图片 URL:")
  if (url) {
    editor.value?.chain().focus().setImage({ src: url }).run()
  }
}

/* ---- 表格 ---- */
function insertTable() {
  editor.value?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
}
function addColumnBefore() {
  editor.value?.chain().focus().addColumnBefore().run()
}
function addColumnAfter() {
  editor.value?.chain().focus().addColumnAfter().run()
}
function deleteColumn() {
  editor.value?.chain().focus().deleteColumn().run()
}
function addRowBefore() {
  editor.value?.chain().focus().addRowBefore().run()
}
function addRowAfter() {
  editor.value?.chain().focus().addRowAfter().run()
}
function deleteRow() {
  editor.value?.chain().focus().deleteRow().run()
}
function deleteTable() {
  editor.value?.chain().focus().deleteTable().run()
}
function toggleHeaderColumn() {
  editor.value?.chain().focus().toggleHeaderColumn().run()
}
function toggleHeaderRow() {
  editor.value?.chain().focus().toggleHeaderRow().run()
}

/* ---- Callout ---- */
function toggleCallout(type: "info" | "warning" | "error" | "success" = "info") {
  editor.value?.chain().focus().toggleWrap("callout", { type }).run()
}

/* ---- 文本对齐 ---- */
function alignLeft() {
  editor.value?.chain().focus().setTextAlign("left").run()
}
function alignCenter() {
  editor.value?.chain().focus().setTextAlign("center").run()
}
function alignRight() {
  editor.value?.chain().focus().setTextAlign("right").run()
}

/* ---- 激活状态 ---- */
const isActiveBold = computed(() => editor.value?.isActive("bold") ?? false)
const isActiveItalic = computed(() => editor.value?.isActive("italic") ?? false)
const isActiveUnderline = computed(() => editor.value?.isActive("underline") ?? false)
const isActiveStrike = computed(() => editor.value?.isActive("strike") ?? false)
const isActiveCode = computed(() => editor.value?.isActive("code") ?? false)
const isActiveHeading2 = computed(() => editor.value?.isActive("heading", { level: 2 }) ?? false)
const isActiveHeading3 = computed(() => editor.value?.isActive("heading", { level: 3 }) ?? false)
const isActiveBulletList = computed(() => editor.value?.isActive("bulletList") ?? false)
const isActiveOrderedList = computed(() => editor.value?.isActive("orderedList") ?? false)
const isActiveTaskList = computed(() => editor.value?.isActive("taskList") ?? false)
const isActiveBlockquote = computed(() => editor.value?.isActive("blockquote") ?? false)
const isActiveCodeBlock = computed(() => editor.value?.isActive("codeBlock") ?? false)
const isActiveLink = computed(() => editor.value?.isActive("link") ?? false)
const isActiveTable = computed(() => editor.value?.isActive("table") ?? false)
const isActiveCallout = computed(() => editor.value?.isActive("callout") ?? false)

/* ---- Emoji Picker ---- */
const emojiPickerOpen = ref(false)

const emojiCategories: { name: string; emojis: string[] }[] = [
  { name: "常用", emojis: ["😀", "😂", "😊", "🥰", "😎", "🤔", "👍", "👎", "❤️", "🔥", "✅", "❌", "⭐", "💡", "📝", "🚀"] },
  { name: "工作项", emojis: ["🐛", "✨", "🔧", "📋", "🎯", "📌", "🔴", "🟡", "🟢", "🔵", "⚫", "⚪", "🟣", "🟠", "📁", "📂"] },
  { name: "符号", emojis: ["⭐", "💫", "💥", "💯", "❗", "❓", "⚠️", "⛔", "✅", "❌", "🔄", "➡️", "⬆️", "➕", "➖", "📎"] },
  { name: "面孔", emojis: ["😀", "😃", "😄", "😁", "😆", "😅", "🤣", "😂", "🙂", "😉", "😊", "😇", "🥰", "😍", "🤩", "😘"] },
]

function toggleEmojiPicker() {
  emojiPickerOpen.value = !emojiPickerOpen.value
  if (emojiPickerOpen.value) slashMenuOpen.value = false
}

function insertEmoji(emoji: string) {
  if (!editor.value) return
  editor.value.chain().focus().insertContent(emoji).run()
  emojiPickerOpen.value = false
}

/* ---- 工具栏配置（按 variant） ---- */
const toolbarGroups = computed(() => {
  if (props.variant === "compact") return []
  if (props.variant === "comment") {
    return [
      ["bold", "italic", "underline", "strike", "code"],
      ["bulletList", "orderedList"],
      ["link"],
    ]
  }
  // full
  return [
    ["bold", "italic", "underline", "strike", "code"],
    ["heading2", "heading3"],
    ["bulletList", "orderedList", "taskList", "blockquote", "codeBlock"],
    ["emoji"],
    ["alignLeft", "alignCenter", "alignRight"],
    ["calloutInfo", "calloutWarning", "calloutError", "calloutSuccess"],
    ["link", "image", "table"],
  ]
})

/* ---- Slash Command 浮层 ---- */
const slashMenuOpen = ref(false)
const slashFilter = ref("")
const slashSelectedIdx = ref(0)

const filteredSlashItems = computed(() => {
  const q = slashFilter.value.trim().toLowerCase()
  if (!q) return slashItems
  return slashItems.filter(
    (item) =>
      item.label.toLowerCase().includes(q) ||
      item.description.toLowerCase().includes(q) ||
      item.id.toLowerCase().includes(q),
  )
})

function openSlashMenu() {
  slashMenuOpen.value = true
  slashFilter.value = ""
  slashSelectedIdx.value = 0
}

function closeSlashMenu() {
  slashMenuOpen.value = false
  slashFilter.value = ""
}

function executeSlashCommand(item: SlashCommandItem) {
  if (!editor.value) return
  item.execute(editor.value)
  closeSlashMenu()
}

function handleSlashKeydown(e: KeyboardEvent) {
  if (!slashMenuOpen.value) return
  const items = filteredSlashItems.value
  if (e.key === "ArrowDown") {
    e.preventDefault()
    slashSelectedIdx.value = (slashSelectedIdx.value + 1) % items.length
  } else if (e.key === "ArrowUp") {
    e.preventDefault()
    slashSelectedIdx.value = (slashSelectedIdx.value - 1 + items.length) % items.length
  } else if (e.key === "Enter") {
    e.preventDefault()
    if (items[slashSelectedIdx.value]) {
      executeSlashCommand(items[slashSelectedIdx.value])
    }
  } else if (e.key === "Escape") {
    closeSlashMenu()
  }
}

function handleEditorInput() {
  if (!editor.value) return
  const { from } = editor.value.state.selection
  const textBefore = editor.value.state.doc.textBetween(
    Math.max(0, from - 50),
    from,
    undefined,
    "\ufffc",
  )
  // 检测 "/" 前面是空格或段首
  const match = textBefore.match(/(?:^|\s)\/$/)
  if (match && !slashMenuOpen.value) {
    openSlashMenu()
  }
}

function onSlashCommandOpen() {
  openSlashMenu()
}

function onEmojiShortcutOpen() {
  toggleEmojiPicker()
}

onMounted(() => {
  // slash-command 插件派发 window 级事件
  window.addEventListener("slash-command:open", onSlashCommandOpen)
  window.addEventListener("slash-command:close", closeSlashMenu)
  window.addEventListener("rich-editor:open-emoji", onEmojiShortcutOpen)
})

onUnmounted(() => {
  window.removeEventListener("slash-command:open", onSlashCommandOpen)
  window.removeEventListener("slash-command:close", closeSlashMenu)
  window.removeEventListener("rich-editor:open-emoji", onEmojiShortcutOpen)
})

defineExpose({ editor })
</script>

<template>
  <div
    class="rich-editor"
    :class="{
      'rich-editor--compact': variant === 'compact' || compact,
      'rich-editor--comment': variant === 'comment',
      'rich-editor--readonly': !editable,
    }"
  >
    <!-- 菜单栏 -->
    <div v-if="editable && variant !== 'compact' && !compact" class="rich-editor__toolbar">
      <template v-for="(group, gi) in toolbarGroups" :key="gi">
        <template v-for="action in group" :key="action">
          <!-- 格式按钮 (bold/italic/underline/strike/code) -->
          <button
            v-if="action === 'bold'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveBold }"
            title="粗体 (Ctrl+B)"
            @click="toggleBold"
          >
            <strong>B</strong>
          </button>
          <button
            v-else-if="action === 'italic'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveItalic }"
            title="斜体 (Ctrl+I)"
            @click="toggleItalic"
          >
            <em>I</em>
          </button>
          <button
            v-else-if="action === 'underline'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveUnderline }"
            title="下划线 (Ctrl+U)"
            @click="toggleUnderline"
          >
            <u>U</u>
          </button>
          <button
            v-else-if="action === 'strike'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveStrike }"
            title="删除线"
            @click="toggleStrike"
          >
            <s>S</s>
          </button>
          <button
            v-else-if="action === 'code'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveCode }"
            title="行内代码"
            @click="toggleCode"
          >
            {"{ }"}
          </button>

          <!-- 标题 -->
          <button
            v-else-if="action === 'heading2'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveHeading2 }"
            title="标题 2"
            @click="setHeading(2)"
          >
            H2
          </button>
          <button
            v-else-if="action === 'heading3'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveHeading3 }"
            title="标题 3"
            @click="setHeading(3)"
          >
            H3
          </button>

          <!-- 列表 -->
          <button
            v-else-if="action === 'bulletList'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveBulletList }"
            title="无序列表"
            @click="toggleBulletList"
          >
            •≡
          </button>
          <button
            v-else-if="action === 'orderedList'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveOrderedList }"
            title="有序列表"
            @click="toggleOrderedList"
          >
            1≡
          </button>
          <button
            v-else-if="action === 'taskList'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveTaskList }"
            title="任务列表 (Ctrl+Shift+9)"
            @click="toggleTaskList"
          >
            ☑
          </button>
          <button
            v-else-if="action === 'blockquote'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveBlockquote }"
            title="引用"
            @click="toggleBlockquote"
          >
            ❝
          </button>
          <button
            v-else-if="action === 'codeBlock'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveCodeBlock }"
            title="代码块"
            @click="toggleCodeBlock"
          >
            &lt;/&gt;
          </button>

          <!-- Emoji -->
          <button
            v-else-if="action === 'emoji'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': emojiPickerOpen }"
            title="插入 Emoji"
            @click="toggleEmojiPicker"
          >
            😊
          </button>

          <!-- 对齐 -->
          <button
            v-else-if="action === 'alignLeft'"
            class="rich-editor__btn"
            title="左对齐"
            @click="alignLeft"
          >
            ≡‹
          </button>
          <button
            v-else-if="action === 'alignCenter'"
            class="rich-editor__btn"
            title="居中"
            @click="alignCenter"
          >
            ≡≡
          </button>
          <button
            v-else-if="action === 'alignRight'"
            class="rich-editor__btn"
            title="右对齐"
            @click="alignRight"
          >
            ›≡
          </button>

          <!-- Callout 提示框 -->
          <button
            v-else-if="action === 'calloutInfo'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveCallout }"
            title="信息提示框 💡"
            @click="toggleCallout('info')"
          >
            💡
          </button>
          <button
            v-else-if="action === 'calloutWarning'"
            class="rich-editor__btn"
            title="警告提示框 ⚠️"
            @click="toggleCallout('warning')"
          >
            ⚠
          </button>
          <button
            v-else-if="action === 'calloutError'"
            class="rich-editor__btn"
            title="错误提示框 🚫"
            @click="toggleCallout('error')"
          >
            🚫
          </button>
          <button
            v-else-if="action === 'calloutSuccess'"
            class="rich-editor__btn"
            title="成功提示框 ✅"
            @click="toggleCallout('success')"
          >
            ✅
          </button>

          <!-- 链接/图片 -->
          <button
            v-else-if="action === 'link'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveLink }"
            title="链接"
            @click="addLink"
          >
            🔗
          </button>
          <button
            v-else-if="action === 'image'"
            class="rich-editor__btn"
            title="图片"
            @click="insertImage"
          >
            🖼
          </button>

          <!-- 表格 -->
          <button
            v-else-if="action === 'table'"
            class="rich-editor__btn"
            :class="{ 'rich-editor__btn--active': isActiveTable }"
            title="插入表格"
            @click="insertTable"
          >
            ⊞
          </button>
        </template>

        <span v-if="gi < toolbarGroups.length - 1" class="rich-editor__divider"></span>
      </template>

      <!-- Table operations dropdown (shown when in table) -->
      <template v-if="editable && isActiveTable">
        <span class="rich-editor__divider"></span>
        <span class="rich-editor__btn-group">
          <button class="rich-editor__btn" title="在前面插入列" @click="addColumnBefore">+←</button>
          <button class="rich-editor__btn" title="在后面插入列" @click="addColumnAfter">+→</button>
          <button class="rich-editor__btn" title="删除列" @click="deleteColumn">−⊞</button>
          <button class="rich-editor__btn" title="在前面插入行" @click="addRowBefore">+↑</button>
          <button class="rich-editor__btn" title="在后面插入行" @click="addRowAfter">+↓</button>
          <button class="rich-editor__btn" title="删除行" @click="deleteRow">−≡</button>
          <button class="rich-editor__btn" title="表头行" @click="toggleHeaderRow">H≡</button>
          <button class="rich-editor__btn" title="表头列" @click="toggleHeaderColumn">V≡</button>
          <button class="rich-editor__btn rich-editor__btn--danger" title="删除表格" @click="deleteTable">🗑⊞</button>
        </span>
      </template>
    </div>

    <!-- Slash 命令浮层 -->
    <div v-if="slashMenuOpen" class="rich-editor__slash-menu" @keydown="handleSlashKeydown">
      <div class="rich-editor__slash-header">快捷命令</div>
      <div
        v-for="(item, idx) in filteredSlashItems"
        :key="item.id"
        class="rich-editor__slash-item"
        :class="{ 'rich-editor__slash-item--selected': idx === slashSelectedIdx }"
        @click="executeSlashCommand(item)"
        @mouseenter="slashSelectedIdx = idx"
      >
        <span class="rich-editor__slash-icon">{{ item.icon }}</span>
        <div class="rich-editor__slash-info">
          <span class="rich-editor__slash-label">{{ item.label }}</span>
          <span class="rich-editor__slash-desc">{{ item.description }}</span>
        </div>
        <span class="rich-editor__slash-cat">{{ item.category }}</span>
      </div>
      <div v-if="filteredSlashItems.length === 0" class="rich-editor__slash-empty">
        没有匹配的命令
      </div>
    </div>

    <!-- Emoji 选择面板 -->
    <div v-if="emojiPickerOpen" class="rich-editor__emoji-panel">
      <div
        v-for="cat in emojiCategories"
        :key="cat.name"
        class="rich-editor__emoji-cat"
      >
        <div class="rich-editor__emoji-cat-name">{{ cat.name }}</div>
        <div class="rich-editor__emoji-grid">
          <button
            v-for="e in cat.emojis"
            :key="e"
            class="rich-editor__emoji-item"
            @click="insertEmoji(e)"
          >
            {{ e }}
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑区域 -->
    <EditorContent
      :editor="editor"
      class="rich-editor__content"
      :style="{ minHeight: variant === 'compact' ? '40px' : variant === 'comment' ? '80px' : minHeight }"
      @input="handleEditorInput"
    />
  </div>
</template>

<style scoped>
.rich-editor {
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-md, 8px);
  background: var(--bg-surface-1, var(--surface-1, #fff));
  overflow: hidden;
  position: relative;
  transition: border-color 0.15s;
}

.rich-editor:focus-within {
  border-color: var(--border-accent-strong, var(--brand-500, #3b82f6));
  box-shadow: 0 0 0 3px var(--bg-accent-subtle, var(--brand-50, rgba(59, 130, 246, 0.1)));
}

.rich-editor--readonly {
  border: none;
  background: transparent;
}

.rich-editor--readonly:focus-within {
  box-shadow: none;
}

.rich-editor--comment {
  border-radius: 8px;
}

.rich-editor--comment .rich-editor__content {
  padding: 10px 12px;
}

.rich-editor--compact {
  border: none;
  border-bottom: 1px solid var(--border-subtle);
  border-radius: 0 0 var(--radius-sm, 4px) var(--radius-sm, 4px);
  box-shadow: none;
}

.rich-editor--compact .rich-editor__content {
  padding: 6px 0;
}

.rich-editor--compact.rich-editor:focus-within {
  border-color: var(--border-accent-strong, var(--brand-500));
  box-shadow: none;
}

/* ---- Toolbar ---- */
.rich-editor__toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  background: var(--bg-surface-2, var(--surface-2, #f9fafb));
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  flex-wrap: wrap;
}

.rich-editor--comment .rich-editor__toolbar {
  padding: 4px 6px;
  background: var(--bg-layer-transparent);
  border-bottom-style: dashed;
}

.rich-editor__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 26px;
  height: 26px;
  padding: 0 5px;
  font-size: 12px;
  font-weight: 600;
  color: var(--txt-secondary, var(--text-secondary, #4b5563));
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.1s;
}

.rich-editor__btn:hover {
  background: var(--bg-layer-3-hover, var(--surface-3, #f3f4f6));
  border-color: var(--border-default, #d1d5db);
}

.rich-editor__btn--active {
  background: var(--bg-accent-subtle, var(--brand-50, rgba(59, 130, 246, 0.08)));
  color: var(--txt-accent-primary, var(--brand-500, #3b82f6));
  border-color: var(--border-accent-subtle, var(--brand-200, #bfdbfe));
}

.rich-editor__btn--danger {
  color: var(--danger-500, #ef4444);
}

.rich-editor__btn--danger:hover {
  background: var(--bg-danger-subtle, var(--danger-50));
  border-color: var(--border-danger-subtle);
}

.rich-editor__divider {
  width: 1px;
  height: 18px;
  background: var(--border-subtle, #e5e7eb);
  margin: 0 4px;
}

.rich-editor__btn-group {
  display: inline-flex;
  gap: 1px;
}

/* ---- Slash Command Menu ---- */
.rich-editor__slash-menu {
  position: absolute;
  top: 40px;
  left: 8px;
  right: 8px;
  max-height: 300px;
  overflow-y: auto;
  background: var(--bg-surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-overlay-100);
  z-index: 100;
  padding: 4px;
}

.rich-editor__slash-header {
  padding: 6px 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.rich-editor__slash-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  transition: background 0.1s;
}

.rich-editor__slash-item:hover,
.rich-editor__slash-item--selected {
  background: var(--bg-accent-subtle, var(--brand-50, #eef2fe));
}

.rich-editor__slash-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  background: var(--bg-surface-2);
  border: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.rich-editor__slash-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
}

.rich-editor__slash-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--txt-primary);
}

.rich-editor__slash-desc {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.rich-editor__slash-cat {
  font-size: 10px;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.3px;
  background: var(--bg-layer-1);
  padding: 2px 6px;
  border-radius: 3px;
}

.rich-editor__slash-empty {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--txt-tertiary);
}

/* ---- Content ---- */
.rich-editor__content {
  padding: 12px;
}

.rich-editor__content :deep(.ProseMirror) {
  outline: none;
  font-size: 14px;
  line-height: 1.6;
  color: var(--txt-primary, var(--text-primary, #1f2937));
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
  color: var(--txt-primary);
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
  border-left: 3px solid var(--border-accent-subtle, var(--brand-200, #bfdbfe));
  padding-left: 12px;
  margin: 8px 0;
  color: var(--txt-secondary);
}

.rich-editor__content :deep(.ProseMirror code) {
  background: var(--bg-layer-3, var(--surface-3, #f3f4f6));
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: var(--font-mono, 'Consolas', monospace);
}

.rich-editor__content :deep(.ProseMirror pre) {
  background: var(--bg-layer-2, #f8fafc);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm, 6px);
  padding: 12px;
  margin: 8px 0;
  overflow-x: auto;
}

.rich-editor__content :deep(.ProseMirror pre code) {
  background: none;
  padding: 0;
  border-radius: 0;
  font-size: 13px;
}

.rich-editor__content :deep(.ProseMirror a) {
  color: var(--txt-accent-primary, var(--brand-500, #3b82f6));
  text-decoration: underline;
}

.rich-editor__content :deep(.ProseMirror img) {
  max-width: 100%;
  border-radius: var(--radius-sm, 4px);
  margin: 8px 0;
}

.rich-editor__content :deep(.ProseMirror ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;
}

.rich-editor__content :deep(.ProseMirror ul[data-type="taskList"] li) {
  display: flex;
  gap: 6px;
  align-items: flex-start;
}

.rich-editor__content :deep(.ProseMirror ul[data-type="taskList"] li input[type="checkbox"]) {
  margin-top: 3px;
  accent-color: var(--brand-default, var(--brand-500));
}

.rich-editor__content :deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: var(--txt-placeholder, var(--text-tertiary, #9ca3af));
  pointer-events: none;
  height: 0;
}

/* Mention */
.rich-editor__content :deep(.ProseMirror .mention) {
  background: var(--bg-accent-subtle, var(--brand-50));
  color: var(--txt-accent-secondary, var(--brand-600, #2563eb));
  padding: 1px 4px;
  border-radius: 3px;
  font-weight: 500;
  font-size: 13px;
}

/* Highlight */
.rich-editor__content :deep(.ProseMirror mark) {
  background: var(--bg-warning-subtle, var(--amber-50));
  color: var(--txt-warning-primary);
  padding: 1px 3px;
  border-radius: 3px;
}

/* Callout */
.rich-editor__content :deep(.ProseMirror div[data-type="callout"]) {
  border-radius: var(--radius-md, 8px);
  padding: 12px 16px;
  margin: 8px 0;
  border-left: 3px solid;
  display: flex;
  gap: 8px;
}

.rich-editor__content :deep(.ProseMirror div[data-type="callout"][data-callout-type="info"]) {
  background: var(--brand-50, #eef2fe);
  border-color: var(--brand-default, #3b82f6);
}

.rich-editor__content :deep(.ProseMirror div[data-type="callout"][data-callout-type="warning"]) {
  background: var(--warning-50, #fffbeb);
  border-color: var(--warning-500, #f59e0b);
}

.rich-editor__content :deep(.ProseMirror div[data-type="callout"][data-callout-type="error"]) {
  background: var(--danger-50, #fef2f2);
  border-color: var(--danger-500, #ef4444);
}

.rich-editor__content :deep(.ProseMirror div[data-type="callout"][data-callout-type="success"]) {
  background: var(--success-50, #ecfdf5);
  border-color: var(--success-500, #10b981);
}

/* Table */
.rich-editor__content :deep(.ProseMirror table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
  table-layout: fixed;
}

.rich-editor__content :deep(.ProseMirror th),
.rich-editor__content :deep(.ProseMirror td) {
  border: 1px solid var(--border-subtle);
  padding: 8px 12px;
  position: relative;
  vertical-align: top;
  min-width: 60px;
}

.rich-editor__content :deep(.ProseMirror th) {
  background: var(--bg-surface-2);
  font-weight: 600;
  text-align: left;
}

.rich-editor__content :deep(.ProseMirror .selectedCell:after) {
  content: "";
  position: absolute;
  inset: 0;
  background: var(--bg-accent-subtle);
  pointer-events: none;
  z-index: 2;
}

.rich-editor__content :deep(.ProseMirror .column-resize-handle) {
  position: absolute;
  top: 0;
  bottom: 0;
  right: -2px;
  width: 4px;
  background: var(--brand-default);
  pointer-events: none;
}

/* Text align */
.rich-editor__content :deep(.ProseMirror .is-editor-empty):first-child::before {
  color: var(--txt-placeholder);
}

/* Text color (TipTap Color extension uses mark) */
.rich-editor__content :deep(.ProseMirror span[style*="color"]) {
  display: inline;
}

/* ---- Emoji Picker ---- */
.rich-editor__emoji-panel {
  position: absolute;
  top: 40px;
  right: 8px;
  width: 280px;
  max-height: 200px;
  overflow-y: auto;
  background: var(--bg-surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 8px);
  box-shadow: var(--shadow-overlay-100);
  z-index: 100;
  padding: 8px;
}

.rich-editor__emoji-cat {
  margin-bottom: 6px;
}

.rich-editor__emoji-cat-name {
  font-size: 10px;
  font-weight: 600;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.3px;
  margin-bottom: 4px;
  padding: 0 2px;
}

.rich-editor__emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 2px;
}

.rich-editor__emoji-item {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  aspect-ratio: 1;
  font-size: 16px;
  background: none;
  border: none;
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
  transition: background 0.1s;
  padding: 0;
}

.rich-editor__emoji-item:hover {
  background: var(--bg-surface-3, var(--surface-3, #f3f4f6));
}
</style>
