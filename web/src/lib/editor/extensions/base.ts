/**
 * base.ts — 基础编辑扩展
 * 包含文档结构核心类型，所有场景都应引入。
 */
import Document from "@tiptap/extension-document"
import Paragraph from "@tiptap/extension-paragraph"
import Text from "@tiptap/extension-text"
import HardBreak from "@tiptap/extension-hard-break"
import History from "@tiptap/extension-history"
import Placeholder from "@tiptap/extension-placeholder"

/** 基础扩展集合 — 纯文档结构 */
export const baseExtensions = [
  Document,
  Paragraph,
  Text,
  HardBreak,
  History,
]

/** 含占位符的基础扩展 */
export function baseWithPlaceholder(placeholder: string) {
  return [
    Document,
    Paragraph,
    Text,
    HardBreak,
    History,
    Placeholder.configure({ placeholder }),
  ]
}

export { Document, Paragraph, Text, HardBreak, History, Placeholder }
