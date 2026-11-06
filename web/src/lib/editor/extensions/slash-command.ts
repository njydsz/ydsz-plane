/**
 * slash-command.ts — 斜杠命令菜单扩展
 *
 * 输入 "/" 弹出命令菜单，支持插入：
 *  - 标题 (H2/H3)
 *  - 无序/有序列表、任务列表
 *  - 引用、代码块、分割线
 *  - 表格
 *  - Callout 提示框（4 种语义）
 *
 * 采用轻量内联浮层，无需依赖第三方 suggestion 库
 */

import { Extension } from "@tiptap/core"
import { Plugin, PluginKey } from "@tiptap/pm/state"

export interface SlashCommandItem {
  id: string
  label: string
  icon: string
  description: string
  category: "基础" | "媒体" | "块" | "布局" | "AI"
  execute: (editor: any) => void
}

/** AI 命令上下文 — 由 RichTextEditor 在 workspace/project 变化时设置 */
export interface AICommandContext {
  wsId: number
  projectId: number
}

let _aiCtx: AICommandContext | null = null

/** 设置 AI 命令上下文 */
export function setAICommandContext(ctx: AICommandContext | null) {
  _aiCtx = ctx
}

/** 获取当前 AI 命令上下文 */
export function getAICommandContext(): AICommandContext | null {
  return _aiCtx
}

const slashItems: SlashCommandItem[] = [
  // ---- 基础 ----
  {
    id: "heading-2",
    label: "标题 2",
    icon: "H2",
    description: "大号章节标题",
    category: "基础",
    execute: (editor) => editor.chain().focus().toggleHeading({ level: 2 }).run(),
  },
  {
    id: "heading-3",
    label: "标题 3",
    icon: "H3",
    description: "中号段落标题",
    category: "基础",
    execute: (editor) => editor.chain().focus().toggleHeading({ level: 3 }).run(),
  },
  {
    id: "bullet-list",
    label: "无序列表",
    icon: "•≡",
    description: "创建简单的项目列表",
    category: "基础",
    execute: (editor) => editor.chain().focus().toggleBulletList().run(),
  },
  {
    id: "ordered-list",
    label: "有序列表",
    icon: "1≡",
    description: "创建带编号的列表",
    category: "基础",
    execute: (editor) => editor.chain().focus().toggleOrderedList().run(),
  },
  {
    id: "task-list",
    label: "任务列表",
    icon: "☑",
    description: "创建可勾选的任务列表",
    category: "基础",
    execute: (editor) => editor.chain().focus().toggleTaskList().run(),
  },
  // ---- 块 ----
  {
    id: "blockquote",
    label: "引用",
    icon: "❝",
    description: "插入引用块",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleBlockquote().run(),
  },
  {
    id: "code-block",
    label: "代码块",
    icon: "<>",
    description: "插入多行代码块",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleCodeBlock().run(),
  },
  {
    id: "horizontal-rule",
    label: "分割线",
    icon: "―",
    description: "插入水平分隔线",
    category: "块",
    execute: (editor) => editor.chain().focus().setHorizontalRule().run(),
  },
  {
    id: "callout-info",
    label: "提示框 – 信息",
    icon: "💡",
    description: "蓝色信息提示框",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleWrap("callout", { type: "info" }).run(),
  },
  {
    id: "callout-warning",
    label: "提示框 – 警告",
    icon: "⚠️",
    description: "黄色警告提示框",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleWrap("callout", { type: "warning" }).run(),
  },
  {
    id: "callout-error",
    label: "提示框 – 错误",
    icon: "🚫",
    description: "红色错误提示框",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleWrap("callout", { type: "error" }).run(),
  },
  {
    id: "callout-success",
    label: "提示框 – 成功",
    icon: "✅",
    description: "绿色成功提示框",
    category: "块",
    execute: (editor) => editor.chain().focus().toggleWrap("callout", { type: "success" }).run(),
  },
  // ---- 媒体 ----
  {
    id: "table",
    label: "表格",
    icon: "⊞",
    description: "插入 3×3 表格",
    category: "媒体",
    execute: (editor) =>
      editor
        .chain()
        .focus()
        .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
        .run(),
  },
  {
    id: "image",
    label: "图片",
    icon: "🖼",
    description: "插入图片（输入 URL）",
    category: "媒体",
    execute: (editor) => {
      const url = prompt("输入图片 URL:")
      if (url) editor.chain().focus().setImage({ src: url }).run()
    },
  },
  {
    id: "link",
    label: "链接",
    icon: "🔗",
    description: "插入超链接",
    category: "媒体",
    execute: (editor) => {
      const url = prompt("输入链接 URL:")
      if (url) editor.chain().focus().setLink({ href: url }).run()
    },
  },
  {
    id: "emoji",
    label: "Emoji",
    icon: "😊",
    description: "插入表情符号",
    category: "媒体",
    execute: () => {
      // 触发自定义事件 — 由 RichTextEditor 监听并打开 Emoji 选择面板
      window.dispatchEvent(new CustomEvent("rich-editor:open-emoji"))
    },
  },
  {
    id: "video",
    label: "视频嵌入",
    icon: "▶",
    description: "嵌入 YouTube / Bilibili 等视频",
    category: "媒体",
    execute: (editor) => {
      const url = prompt("输入视频 URL (YouTube/Bilibili):")
      if (url) {
        editor.chain().focus().setEmbed("video", url).run()
      }
    },
  },
  {
    id: "embed",
    label: "嵌入网页",
    icon: "🔲",
    description: "嵌入任意 iframe 内容",
    category: "媒体",
    execute: (editor) => {
      const url = prompt("输入嵌入 URL:")
      if (url) {
        editor.chain().focus().setEmbed("iframe", url).run()
      }
    },
  },
  {
    id: "page-link",
    label: "页面链接",
    icon: "📄",
    description: "引用知识库或项目页面",
    category: "媒体",
    execute: (editor) => {
      const pageId = prompt("输入页面 ID:")
      const title = prompt("输入页面标题:") || "页面"
      if (pageId) {
        editor.chain().focus().insertContent({
          type: "pageLink",
          attrs: { pageId, space: "project", title },
        }).run()
      }
    },
  },
  {
    id: "drawio",
    label: "draw.io 图表",
    icon: "◇",
    description: "嵌入 draw.io 流程图/架构图",
    category: "媒体",
    execute: (editor) => {
      const title = prompt("图表标题（可选）:")
      editor.chain().focus().setDrawio({ title: title || "" }).run()
    },
  },
  {
    id: "database",
    label: "Database 表格",
    icon: "▦",
    description: "嵌入页内数据库（类型化列 + 行数据）",
    category: "布局",
    execute: (editor) => {
      editor.chain().focus().setDatabase({ title: "数据库" }).run()
    },
  },
  {
    id: "work-item-embed",
    label: "工作项引用",
    icon: "📋",
    description: "嵌入引用一个工作项",
    category: "媒体",
    execute: (editor) => {
      const issueId = prompt("输入工作项 ID:")
      const title = prompt("输入工作项标题:") || "工作项"
      if (issueId) {
        editor.chain().focus().insertContent({
          type: "workItemEmbed",
          attrs: { issueId, issueType: "task", title },
        }).run()
      }
    },
  },
  // ---- 布局 ----
  {
    id: "columns-2",
    label: "双列布局",
    icon: "▐▌",
    description: "创建左右两列布局",
    category: "布局",
    execute: (editor) => {
      editor.chain().focus().setColumns(2).run()
    },
  },
  {
    id: "columns-3",
    label: "三列布局",
    icon: "▌▐▌",
    description: "创建三列布局",
    category: "布局",
    execute: (editor) => {
      editor.chain().focus().setColumns(3).run()
    },
  },
  {
    id: "tabs",
    label: "标签页",
    icon: "⊞",
    description: "创建可折叠的标签页块",
    category: "布局",
    execute: (editor) => {
      editor.chain().focus().setTabs().run()
    },
  },
  // ---- AI ----
  {
    id: "ai-assist",
    label: "AI 续写",
    icon: "✨",
    description: "根据上下文智能续写文本",
    category: "AI",
    execute: () => {
      window.dispatchEvent(new CustomEvent("rich-editor:ai-assist"))
    },
  },
  {
    id: "ai-rewrite",
    label: "AI 改写",
    icon: "🔄",
    description: "改写选中文本（正式/精简/流畅/扩写）",
    category: "AI",
    execute: () => {
      window.dispatchEvent(new CustomEvent("rich-editor:ai-rewrite"))
    },
  },
  {
    id: "ai-fix-grammar",
    label: "AI 纠错",
    icon: "🔍",
    description: "检测并修正语法、拼写和标点问题",
    category: "AI",
    execute: () => {
      window.dispatchEvent(new CustomEvent("rich-editor:ai-fix-grammar"))
    },
  },
]

export { slashItems }

export const SlashCommand = Extension.create({
  name: "slashCommand",

  addOptions() {
    return {
      suggestion: {
        char: "/",
      },
    }
  },

  addProseMirrorPlugins() {
    return [
      new Plugin({
        key: new PluginKey("slashCommand"),
        props: {
          handleKeyDown(view, event) {
            // 当输入 "/" 且前面为空时自动选中第一个命令
            if (event.key === "/") {
              const { $from } = view.state.selection
              const before = $from.parent.textBetween(
                Math.max(0, $from.parentOffset - 20),
                $from.parentOffset,
                undefined,
                "\ufffc",
              )
              // 只有在行首或空格后才触发
              if (before === "" || before.endsWith(" ")) {
                // 触发自定义事件打开斜杠菜单（无需传递 editor，已有监听方自行持有）
                const evt = new CustomEvent("slash-command:open", {
                  detail: { items: slashItems },
                })
                window.dispatchEvent(evt)
              }
            }
            // ESC 关闭菜单
            if (event.key === "Escape") {
              window.dispatchEvent(new CustomEvent("slash-command:close"))
            }
            return false
          },
        },
      }),
    ]
  },
})

export default SlashCommand
