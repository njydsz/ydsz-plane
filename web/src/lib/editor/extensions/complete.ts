/**
 * complete.ts — 预组合的完整扩展集
 * 根据不同场景提供开箱即用的扩展组合。
 */
import { baseWithPlaceholder } from "./base"
import { formattingBasic, formattingExtensions } from "./formatting"
import { blockBasic, blockExtensions } from "./block"
import { inlineExtensions } from "./inline"

/**
 * 完整编辑器扩展集 — 用于 Issue 描述、文档编辑等富文本场景
 */
export function completeExtensions(placeholder = "输入内容...") {
  return [
    ...baseWithPlaceholder(placeholder),
    ...formattingExtensions,
    ...blockExtensions,
    ...inlineExtensions,
  ]
}

/**
 * 评论编辑器扩展集 — 精简版，仅基础格式
 */
export function commentExtensions(placeholder = "写下你的评论...") {
  return [
    ...baseWithPlaceholder(placeholder),
    ...formattingBasic(),
    ...blockBasic(),
    ...inlineExtensions,
  ]
}

/**
 * 紧凑编辑器扩展集 — 用于标题等单行/少行场景
 */
export function compactExtensions(placeholder = "输入...") {
  return [
    ...baseWithPlaceholder(placeholder),
    formattingBasic(),
  ]
}

/**
 * 只读模式扩展（仅渲染，无编辑历史）
 */
export function readonlyExtensions(placeholder?: string) {
  const base = placeholder
    ? baseWithPlaceholder(placeholder)
    : [
        baseWithPlaceholder("")[0], // Document
        baseWithPlaceholder("")[1], // Paragraph
        baseWithPlaceholder("")[2], // Text
        baseWithPlaceholder("")[3], // HardBreak
      ]
  return [
    ...base,
    ...formattingExtensions,
    ...blockExtensions,
    ...inlineExtensions,
  ]
}
