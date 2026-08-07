/**
 * formatting.ts — 文本格式化扩展
 * 行内样式：粗体、斜体、下划线、删除线、行内代码
 */
import Bold from "@tiptap/extension-bold"
import Italic from "@tiptap/extension-italic"
import Underline from "@tiptap/extension-underline"
import Strike from "@tiptap/extension-strike"
import Code from "@tiptap/extension-code"

/** 全量格式化扩展 */
export const formattingExtensions = [
  Bold,
  Italic,
  Underline,
  Strike,
  Code,
]

/** 精简格式化 */
export function formattingBasic() {
  return [Bold, Italic, Underline, Strike, Code]
}

export { Bold, Italic, Underline, Strike, Code }
