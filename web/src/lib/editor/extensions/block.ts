/**
 * block.ts — 块级扩展
 * 标题、引用、代码块、分割线、列表
 */
import Heading from "@tiptap/extension-heading"
import Blockquote from "@tiptap/extension-blockquote"
import CodeBlock from "@tiptap/extension-code-block"
import HorizontalRule from "@tiptap/extension-horizontal-rule"
import BulletList from "@tiptap/extension-bullet-list"
import OrderedList from "@tiptap/extension-ordered-list"
import ListItem from "@tiptap/extension-list-item"

/** 全量块级扩展 */
export const blockExtensions = [
  Heading,
  Blockquote,
  CodeBlock,
  HorizontalRule,
  BulletList,
  OrderedList,
  ListItem,
]

/** 精简块级（仅基础列表+标题） */
export function blockBasic() {
  return [
    Heading,
    BulletList,
    OrderedList,
    ListItem,
  ]
}

export {
  Heading,
  Blockquote,
  CodeBlock,
  HorizontalRule,
  BulletList,
  OrderedList,
  ListItem,
}
