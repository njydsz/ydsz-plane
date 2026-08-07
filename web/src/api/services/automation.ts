/**
 * 自动化规则 API — 规则 CRUD、模板库、执行历史、干跑测试。
 *
 * 路由: /api/v1/workspaces/:ws/projects/:pid/automation
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type RuleStatus = "draft" | "active" | "disabled" | "error";
/** 自动化触发器类型枚举（与后端 TriggerType 常量对齐）。 */
export type TriggerType = "issue.created" | "issue.updated" | "issue.status_changed"
  | "issue.assigned" | "issue.commented" | "sprint.started" | "sprint.completed"
  | "version.released" | "cron";

/** DSL 触发器定义 — 描述规则触发事件类型与可选前置条件。 */
export interface RuleDSLTrigger {
  type: TriggerType;
  conditions?: Record<string, any>;
}

/** DSL 动作定义 — 描述触发后执行的操作（类型 + 参数）。 */
export interface RuleDSLAction {
  type: string;
  params?: Record<string, any>;
}

/** 自动化规则 DSL 完整结构 — 触发器 + 条件列表 + 动作列表。 */
export interface RuleDSL {
  trigger: RuleDSLTrigger;
  conditions?: Array<{ field: string; operator: string; value: any }>;
  actions: RuleDSLAction[];
}

/** 自动化规则 */
export interface AutomationRule {
  id: number;
  workspace_id: number;
  project_id?: number | null;
  name: string;
  description: string;
  dsl: RuleDSL;
  status: RuleStatus;
  version: number;
  last_triggered_at?: string;
  failure_count: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 规则执行记录 */
export interface RuleExecution {
  id: number;
  rule_id: number;
  rule_name: string;
  trigger_event: string;
  status: "success" | "failed" | "skipped";
  actions_executed: string[];
  error_message?: string;
  duration_ms?: number;
  created_at: string;
}

/** 内置模板 */
export interface AutomationTemplate {
  slug: string;
  name: string;
  description: string;
  category: string;
  icon: string;
  dsl: RuleDSL;
}

/** 规则列表查询参数（分页 + 触发器类型过滤 + 状态过滤）。 */
export interface ListRulesParams {
  status?: RuleStatus;
  trigger_type?: string;
  limit?: number;
  offset?: number;
}

/** 创建规则入参（name + dsl 必填；status 默认 draft）。 */
export interface CreateRuleInput {
  name: string;
  description?: string;
  dsl: RuleDSL;
  status?: RuleStatus;
}

/** 更新规则入参（可选字段 + 乐观锁 version 防并发覆盖）。 */
export interface UpdateRuleInput {
  name?: string;
  description?: string;
  dsl?: RuleDSL;
  status?: string;
  version: number;
}

/** 干跑（dry-run）测试结果 — 返回校验信息而非真实执行。 */
export interface DryRunResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
  actions_count: number;
  trigger_type: string;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 自动化规则域 API：CRUD、模板预置、执行历史、dry-run 测试、启用/禁用开关。 */
export const automationApi = {
  // --- Rules CRUD ---
  list: (wsId: number, projectId: number, params?: ListRulesParams) =>
    wrap<{ total: number; items: AutomationRule[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/automation`, { params }),
    ),
  get: (wsId: number, projectId: number, ruleId: number) =>
    wrap<AutomationRule>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/automation/${ruleId}`),
    ),
  create: (wsId: number, projectId: number, input: CreateRuleInput) =>
    wrap<AutomationRule>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/automation`, input),
    ),
  update: (wsId: number, projectId: number, ruleId: number, input: UpdateRuleInput) =>
    wrap<AutomationRule>(
      http.patch(`/workspaces/${wsId}/projects/${projectId}/automation/${ruleId}`, input),
    ),
  delete: (wsId: number, projectId: number, ruleId: number) =>
    http.delete(`/workspaces/${wsId}/projects/${projectId}/automation/${ruleId}`),

  // --- Actions ---
  toggle: (wsId: number, projectId: number, ruleId: number, enable: boolean, version: number) =>
    wrap<AutomationRule>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/automation/${ruleId}/toggle`, {
        enable, version,
      }),
    ),

  // --- Templates ---
  listTemplates: (wsId: number, projectId: number) =>
    wrap<AutomationTemplate[]>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/automation/templates`),
    ),
  createFromTemplate: (wsId: number, projectId: number, templateSlug: string, name?: string) =>
    wrap<AutomationRule>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/automation/from-template`, {
        template_slug: templateSlug,
        name,
      }),
    ),

  // --- Dry run ---
  dryRun: (wsId: number, projectId: number, dsl: RuleDSL) =>
    wrap<DryRunResult>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/automation/dry-run`, { dsl }),
    ),

  // --- Executions ---
  listExecutions: (wsId: number, projectId: number, limit?: number) =>
    wrap<RuleExecution[]>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/automation/executions`, {
        params: { limit: limit ?? 20 },
      }),
    ),
};
