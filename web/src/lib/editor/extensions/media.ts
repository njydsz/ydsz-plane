/**
 * media.ts — 媒体与辅助功能扩展
 * 任务列表、文本对齐、文本颜色、高亮
 */
import TaskList from "@tiptap/extension-task-list"
import TaskItem from "@tiptap/extension-task-item"
import TextAlign from "@tiptap/extension-text-align"
import { TextStyle } from "@tiptap/extension-text-style"
import { Color } from "@tiptap/extension-color"
import Highlight from "@tiptap/extension-highlight"

/** 辅助编辑能力集合 */
export const mediaExtensions = [
  TaskList,
  TaskItem.configure({ nested: true }),
  TextAlign.configure({ types: ["heading", "paragraph"] }),
  TextStyle,
  Color,
  Highlight.configure({ multicolor: true }),
]

export { TaskList, TaskItem, TextAlign, TextStyle, Color, Highlight }
