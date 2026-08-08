/**
 * complete.ts — 预组合的完整扩展集
 * 根据不同场景提供开箱即用的扩展组合。
 */
import { baseWithPlaceholder } from "./base"
import { formattingBasic, formattingExtensions } from "./formatting"
import { blockBasic, blockExtensions } from "./block"
import { inlineExtensions } from "./inline"
import { tableExtensions } from "./table"
import { mediaExtensions } from "./media"
import Callout from "./callout"
import SlashCommand, { slashItems } from "./slash-command"

/**
 * 完整编辑器扩展集 — 用于 Issue 描述、文档编辑等富文本场景
 * 包含格式、块级、表格、任务列表、文本对齐、颜色、高亮等全套能力
 */
export function completeExtensions(placeholder = "输入内容...") {
  return [
    ...baseWithPlaceholder(placeholder),
    ...formattingExtensions,
    ...blockExtensions,
    ...inlineExtensions,
    ...tableExtensions,
    ...mediaExtensions,
    Callout,
    SlashCommand,
  ]
}

/** 导出斜杠命令项列表（供 UI 浮层使用） */
export { slashItems }

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
    ...formattingBasic(),
  ]
}

/**
 * 只读模式扩展（仅渲染，无编辑历史）
 */
export function readonlyExtensions() {
  return [
    ...baseWithPlaceholder("").slice(0, 4), // Document, Paragraph, Text, HardBreak（不含 History）
    ...formattingExtensions,
    ...blockExtensions,
    ...inlineExtensions,
  ]
}
