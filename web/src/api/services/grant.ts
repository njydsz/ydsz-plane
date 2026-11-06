/**
 * 条件授权（ABAC）API —— 工作空间级条件授权规则管理。
 *
 * 后端路由: /api/v1/workspaces/:workspace_id/grant-rules
 * 在角色权限（RBAC）之上叠加基于属性的约束：
 *   - 字段比较：eq / ne / gt / lt / in / contains
 *   - 逻辑组合：and / or / not
 *   - 受支持字段：projectId / type / stateGroup / priority / creatorId /
 *     assigneeId / labelId / workspaceId / moduleId / sprintId
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 条件字段操作符 */
export type GrantOp =
  | "and"
  | "or"
  | "not"
  | "eq"
  | "ne"
  | "in"
  | "contains"
  | "gt"
  | "lt";

/** 允许参与条件判断的字段白名单 */
export type GrantField =
  | "projectId"
  | "type"
  | "stateGroup"
  | "priority"
  | "creatorId"
  | "assigneeId"
  | "labelId"
  | "workspaceId"
  | "moduleId"
  | "sprintId";

/** 条件表达式节点（受限 JSON，仅识别白名单操作符与字段） */
export interface GrantExpr {
  op: GrantOp;
  /** 比较类操作符必填 */
  field?: GrantField;
  /** eq / ne / contains / gt / lt 使用 */
  value?: unknown;
  /** in 操作符使用 */
  values?: unknown[];
  /** not 操作符使用 */
  inner?: GrantExpr | null;
  /** and 操作符使用 */
  all?: GrantExpr[] | null;
  /** or 操作符使用 */
  any?: GrantExpr[] | null;
}

/** 条件授权规则 */
export interface GrantRule {
  id: number;
  workspace_id: number;
  name: string;
  description: string;
  role_slugs: string[];
  permissions: string[];
  scope: string;
  condition: GrantExpr | null;
  enabled: boolean;
  priority: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建规则入参 */
export interface CreateGrantRuleInput {
  name: string;
  description?: string;
  role_slugs: string[];
  permissions: string[];
  scope?: string;
  condition: GrantExpr;
  priority: number;
}

/** 更新规则入参（字段可选，仅传需要修改的字段） */
export interface UpdateGrantRuleInput {
  name?: string;
  description?: string;
  role_slugs?: string[];
  permissions?: string[];
  scope?: string;
  condition?: GrantExpr;
  enabled?: boolean;
  priority?: number;
}

/** 试算上下文（属性集合） */
export type GrantContext = Partial<Record<GrantField, unknown>>;

/** 试算响应 */
export interface GrantEvaluateResult {
  valid: boolean;
  result: boolean;
  error?: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 条件授权规则域 API */
export const grantApi = {
  /** 列出工作空间下全部条件授权规则 */
  list: (workspaceId: number | string) =>
    wrap<{ items: GrantRule[] }>(
      http.get(`/workspaces/${workspaceId}/grant-rules`),
    ).then((r) => r.items ?? []),

  /** 获取单条规则详情 */
  get: (workspaceId: number | string, id: number | string) =>
    wrap<GrantRule>(http.get(`/workspaces/${workspaceId}/grant-rules/${id}`)),

  /** 创建规则 */
  create: (workspaceId: number | string, input: CreateGrantRuleInput) =>
    wrap<GrantRule>(
      http.post(`/workspaces/${workspaceId}/grant-rules`, input),
    ),

  /** 更新规则 */
  update: (
    workspaceId: number | string,
    id: number | string,
    input: UpdateGrantRuleInput,
  ) =>
    wrap<GrantRule>(
      http.patch(`/workspaces/${workspaceId}/grant-rules/${id}`, input),
    ),

  /** 删除规则 */
  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/grant-rules/${id}`),
    ),

  /** 启用规则 */
  enable: (workspaceId: number | string, id: number | string) =>
    wrap<GrantRule>(
      http.post(`/workspaces/${workspaceId}/grant-rules/${id}/enable`, {}),
    ),

  /** 停用规则 */
  disable: (workspaceId: number | string, id: number | string) =>
    wrap<GrantRule>(
      http.post(`/workspaces/${workspaceId}/grant-rules/${id}/disable`, {}),
    ),

  /** 试算条件表达式（调试用） */
  evaluate: (
    workspaceId: number | string,
    condition: GrantExpr,
    context: GrantContext,
  ) =>
    wrap<GrantEvaluateResult>(
      http.post(`/workspaces/${workspaceId}/grant-rules/evaluate`, {
        condition,
        context,
      }),
    ),
};
