/**
 * 重复工作项 API — 项目级「周期性自动生成工作项」规则的 CRUD 与手动执行。
 *
 * 对标 Plane Recurring Work Items：按 Cron 表达式在固定节奏下自动生成
 * 需求/任务/缺陷，避免每周/每月例行工作项需要人工重复创建。
 *
 * 路由（后端 internal/application/recurring/handler.go）:
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/recurring-rules
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/recurring-rules
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId
 *   PATCH  /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId
 *   DELETE /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId/enable
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId/disable
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/recurring-rules/:ruleId/run-now
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 重复生成的工作项类型（与后端 RecurringRule.WorkItemType 对齐） */
export type RecurringWorkItemType = "requirement" | "task" | "defect";

/**
 * 重复规则。
 *
 * 说明：
 *  - title_template 支持占位符 {date}（如「周报 {date}」），生成时替换为当天日期；
 *  - cron_expr 为 5 段标准 Cron（分 时 日 月 周），由后端调度器解析；
 *  - assignee_ids / label_ids 为 ID 数组；state_id 为生成工作项的初始状态。
 */
export interface RecurringRule {
  id: number;
  workspace_id: number;
  project_id: number;
  /** 工作项类型：task | requirement | defect */
  work_item_type: RecurringWorkItemType;
  /** 工作项标题模板，支持 {date} 占位符 */
  title_template: string;
  /** 工作项描述模板（可选） */
  description_template?: string;
  /** 5 段 Cron 表达式（分 时 日 月 周） */
  cron_expr: string;
  /** 下次计划生成时间 */
  next_run_at: string;
  /** 上次生成时间 */
  last_run_at?: string | null;
  /** 指派人 ID 列表 */
  assignee_ids?: number[];
  /** 标签 ID 列表 */
  label_ids?: number[];
  /** 生成工作项的初始状态 ID（0 表示不指定） */
  state_id?: number;
  /** 是否启用 */
  enabled: boolean;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建重复规则入参（与后端 createRule 请求体对齐） */
export interface CreateRecurringRuleInput {
  work_item_type: RecurringWorkItemType;
  title_template: string;
  description_template?: string;
  cron_expr: string;
  assignee_ids?: number[];
  label_ids?: number[];
  state_id?: number;
  enabled?: boolean;
}

/** 更新重复规则入参（与后端 updateRule 请求体对齐，全部可选） */
export interface UpdateRecurringRuleInput {
  title_template?: string;
  description_template?: string;
  cron_expr?: string;
  assignee_ids?: number[];
  label_ids?: number[];
  state_id?: number;
  enabled?: boolean;
}

/** 重复规则列表查询参数 */
export interface ListRecurringRulesParams {
  enabled?: boolean;
  limit?: number;
  offset?: number;
}

/** 重复规则列表响应 */
export interface ListRecurringRulesResult {
  results: RecurringRule[];
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 重复工作项规则域 API */
export const recurringApi = {
  /** 列出项目下全部重复规则 */
  list: (
    workspaceId: number | string,
    projectId: number | string,
    params?: ListRecurringRulesParams,
  ) =>
    wrap<ListRecurringRulesResult>(
      http.get(`/workspaces/${workspaceId}/projects/${projectId}/recurring-rules`, {
        params,
      }),
    ),

  /** 获取单条重复规则 */
  get: (workspaceId: number | string, projectId: number | string, ruleId: number | string) =>
    wrap<RecurringRule>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}`,
      ),
    ),

  /** 创建重复规则 */
  create: (
    workspaceId: number | string,
    projectId: number | string,
    input: CreateRecurringRuleInput,
  ) =>
    wrap<RecurringRule>(
      http.post(`/workspaces/${workspaceId}/projects/${projectId}/recurring-rules`, input),
    ),

  /** 更新重复规则 */
  update: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
    input: UpdateRecurringRuleInput,
  ) =>
    wrap<RecurringRule>(
      http.patch(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}`,
        input,
      ),
    ),

  /** 删除重复规则（204 无内容） */
  remove: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}`,
      ),
    ),

  /** 启用重复规则 */
  enable: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
  ) =>
    wrap<RecurringRule>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}/enable`,
      ),
    ),

  /** 停用重复规则 */
  disable: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
  ) =>
    wrap<RecurringRule>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}/disable`,
      ),
    ),

  /** 启用 / 停用切换（按目标状态选择对应端点） */
  setEnabled: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
    enabled: boolean,
  ) =>
    enabled
      ? recurringApi.enable(workspaceId, projectId, ruleId)
      : recurringApi.disable(workspaceId, projectId, ruleId),

  /** 立即生成一次（不等 Cron 触发），用于验证规则配置 */
  runNow: (
    workspaceId: number | string,
    projectId: number | string,
    ruleId: number | string,
  ) =>
    wrap<{ ok: boolean }>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/recurring-rules/${ruleId}/run-now`,
      ),
    ),
};
