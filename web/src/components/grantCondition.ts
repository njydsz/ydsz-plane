/**
 * GrantConditionBuilder — 条件节点模型。
 *
 * 从 GrantConditionBuilder.vue 提取为独立 .ts 文件：
 * vue-tsc 对「在 .vue 中导出 interface 并在同文件模板中使用」的类型推断
 * 存在已知限制（模板中变量被错误解析为 props 类型），提取到 .ts 后消除歧义。
 *
 * 类型语义：
 *   - op: 操作符（逻辑类 and/or/not 或比较类 eq/ne/in/contains/gt/lt）
 *   - field: 比较字段（空串表示逻辑节点）
 *   - value: 比较值（单值）
 *   - valuesText: 比较值（多值，逗号分隔文本）
 *   - children: 子节点（and/or 使用）
 *   - inner: 内层节点（not 使用）
 */
import type { GrantField, GrantOp } from "@/api/services/grant";

export interface ConditionNode {
  key: string;
  op: GrantOp;
  field: GrantField | "";
  value: string;
  valuesText: string;
  children: ConditionNode[];
  inner: ConditionNode | null;
}
