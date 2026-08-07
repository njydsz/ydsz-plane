/**
 * block.ts — 块级元素扩展
 * 标题、引用、代码块、分隔线、列表、任务列表
 */
import Heading from "@tiptap/extension-heading"
import Blockquote from "@tiptap/extension-blockquote"
import CodeBlock from "@tiptap/extension-code-block"
import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight"
import HorizontalRule from "@tiptap/extension-horizontal-rule"
import BulletList from "@tiptap/extension-bullet-list"
import OrderedList from "@tiptap/extension-ordered-list"
import ListItem from "@tiptap/extension-list-item"
import TaskList from "@tiptap/extension-task-list"
import TaskItem from "@tiptap/extension-task-item"

/** 标题配置（仅 H2、H3，避免 H1 与页面标题冲突） */
const HeadingExt = Heading.configure({ levels: [2, 3] })

/** 全量块级扩展 */
export const blockExtensions = [
  HeadingExt,
  Blockquote,
  CodeBlock,
  HorizontalRule,
  BulletList,
  OrderedList,
  ListItem,
  TaskList,
  TaskItem.configure({ nested: true }),
]

/** 精简块级（不含任务列表和代码块） */
export function blockBasic() {
  return [HeadingExt, Blockquote, HorizontalRule, BulletList, OrderedList, ListItem]
}

export {
  HeadingExt as Heading,
  Blockquote,
  CodeBlock,
  CodeBlockLowlight,
  HorizontalRule,
  BulletList,
  OrderedList,
  ListItem,
  TaskList,
  TaskItem,
}
