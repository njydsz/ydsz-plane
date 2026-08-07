/**
 * inline.ts — 行内元素扩展
 * 链接、图片、提及、行内图片
 */
import Link from "@tiptap/extension-link"
import Image from "@tiptap/extension-image"
import Mention from "@tiptap/extension-mention"

/** 链接配置（点击不跳转，由应用层控制） */
const LinkExt = Link.configure({
  openOnClick: false,
  HTMLAttributes: {
    class: "editor-link",
    rel: "noopener noreferrer",
  },
})

/** 行内图片配置 */
const ImageExt = Image.configure({
  inline: true,
  allowBase64: true,
  HTMLAttributes: {
    class: "editor-image",
  },
})

/** 全量行内扩展 */
export const inlineExtensions = [LinkExt, ImageExt]

/** 带提及的行内提及配置工厂 */
export function withMention(suggestionConfig: any) {
  return [LinkExt, ImageExt, Mention.configure({ suggestion: suggestionConfig })]
}

export { LinkExt as Link, ImageExt as Image, Mention }
