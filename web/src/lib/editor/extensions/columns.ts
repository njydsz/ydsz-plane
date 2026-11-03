/**
 * columns.ts — 多列布局扩展
 *
 * 基于 TipTap Node 扩展，实现 Notion/Plane 风格的多列布局。
 * 支持 2 列 / 3 列布局，每列内可嵌套任意 block 内容。
 *
 * HTML 结构:
 * <div data-type="columns" data-column-count="2">
 *   <div data-type="column">...</div>
 *   <div data-type="column">...</div>
 * </div>
 */

import { Node, mergeAttributes } from "@tiptap/core"

export interface ColumnsOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    columns: {
      setColumns: (count?: 2 | 3) => ReturnType
      toggleColumns: (count?: 2 | 3) => ReturnType
      unsetColumns: () => ReturnType
    }
  }
}

export const Columns = Node.create<ColumnsOptions>({
  name: "columns",

  group: "block",
  content: "column{2,3}",
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      count: {
        default: 2,
        parseHTML: (element) => {
          const val = element.getAttribute("data-column-count")
          return val ? parseInt(val, 10) : 2
        },
        renderHTML: (attributes) => ({
          "data-column-count": attributes.count,
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="columns"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "columns" }, HTMLAttributes),
      0,
    ]
  },

  addCommands() {
    return {
      setColumns:
        (count = 2) =>
        ({ commands }) =>
          commands.wrapIn(this.name, { count }),
      toggleColumns:
        (count = 2) =>
        ({ commands }) =>
          commands.toggleWrap(this.name, { count }),
      unsetColumns:
        () =>
        ({ commands }) =>
          commands.lift(this.name),
    }
  },
})

export const Column = Node.create({
  name: "column",

  group: "block",
  content: "block+",
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      width: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-width"),
        renderHTML: (attributes) => {
          if (!attributes.width) return {}
          return { "data-width": attributes.width }
        },
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="column"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "column" }, HTMLAttributes),
      0,
    ]
  },
})

export const columnsExtensions = [Columns, Column]

export { Columns as ColumnLayout, Column as ColumnItem }
