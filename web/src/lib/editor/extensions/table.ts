/**
 * table.ts — 表格扩展集成
 * 包含表格、行、单元格、表头及可调整列宽能力
 */
import { Table } from "@tiptap/extension-table"
import TableRow from "@tiptap/extension-table-row"
import TableCell from "@tiptap/extension-table-cell"
import TableHeader from "@tiptap/extension-table-header"

/** 表格扩展集合（可调整列宽） */
export const tableExtensions = [
  Table.configure({ resizable: true }),
  TableRow,
  TableHeader,
  TableCell,
]

export { Table, TableRow, TableCell, TableHeader }
