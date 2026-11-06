/**
 * database.ts — 页内 Database 块
 *
 * 对标 Plane Enterprise 的 Database 功能：在文档中嵌入一个轻量表格数据库，
 * 支持类型化列（text/number/select/multiselect/checkbox/date/people/url）与行数据。
 *
* 数据以 JSON 形式存储在节点的 data-attributes 中（编辑器内直接编辑，
 * 不额外建表；后续如需跨文档查询可迁移到独立表）。
 *
 * HTML 结构:
 * <div data-type="database"
 *      data-db-columns='[{"id":"c1","name":"名称","type":"text"}]'
 *      data-db-rows='[{"id":"r1","cells":{"c1":"值"}}]'
 *      data-db-title="数据库">
 */
import { Node, mergeAttributes } from "@tiptap/core"

export type DatabaseColumnType =
  | "text"
  | "number"
  | "select"
  | "multiselect"
  | "checkbox"
  | "date"
  | "people"
  | "url"

export interface DatabaseColumn {
  id: string
  name: string
  type: DatabaseColumnType
  options?: string[]
  width?: number
}

export interface DatabaseRow {
  id: string
  cells: Record<string, any>
}

export interface DatabaseOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    database: {
      setDatabase: (opts?: { title?: string; columns?: DatabaseColumn[] }) => ReturnType
      unsetDatabase: () => ReturnType
    }
  }
}

/** 默认列定义 */
export const defaultDatabaseColumns = (): DatabaseColumn[] => [
  { id: "col_name", name: "名称", type: "text", width: 200 },
  { id: "col_status", name: "状态", type: "select", options: ["待办", "进行中", "已完成"], width: 140 },
  { id: "col_owner", name: "负责人", type: "people", width: 140 },
  { id: "col_date", name: "日期", type: "date", width: 140 },
]

export const Database = Node.create<DatabaseOptions>({
  name: "database",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return { HTMLAttributes: {} }
  },

  addAttributes() {
    return {
      title: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-db-title") || "",
        renderHTML: (attrs) => (attrs.title ? { "data-db-title": attrs.title } : {}),
      },
      columns: {
        default: null,
        parseHTML: (el) => el.getAttribute("data-db-columns"),
        renderHTML: (attrs) => (attrs.columns ? { "data-db-columns": attrs.columns } : {}),
      },
      rows: {
        default: null,
        parseHTML: (el) => el.getAttribute("data-db-rows"),
        renderHTML: (attrs) => (attrs.rows ? { "data-db-rows": attrs.rows } : {}),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="database"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "database" }, HTMLAttributes),
    ]
  },

  addCommands() {
    return {
      setDatabase:
        (opts = {}) =>
        ({ commands }) =>
          commands.insertContent({
            type: this.name,
            attrs: {
              title: opts.title || "数据库",
              columns: JSON.stringify(opts.columns || defaultDatabaseColumns()),
              rows: JSON.stringify([]),
            },
          }),
      unsetDatabase:
        () =>
        ({ commands }) =>
          commands.deleteNode(this.name),
    }
  },
})

export default Database
