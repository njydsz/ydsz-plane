/**
 * 条件授权（ABAC）共享类型。
 */
import type { GrantField, GrantOp } from "@/api/services/grant";

/** 条件构造器节点（递归）。 */
export interface ConditionNode {
  key: string;
  op: GrantOp;
  field: GrantField | "";
  value: string;
  valuesText: string;
  children: ConditionNode[];
  inner: ConditionNode | null;
}

export function createComparisonNode(op: GrantOp = "eq"): ConditionNode {
  return {
    key: `n${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`,
    op,
    field: "",
    value: "",
    valuesText: "",
    children: [],
    inner: null,
  };
}

export function createLogicalNode(op: "and" | "or" | "not" = "and"): ConditionNode {
  const n = createComparisonNode(op);
  if (op === "not") {
    n.inner = createComparisonNode();
  } else {
    n.children = [createComparisonNode()];
  }
  return n;
}
