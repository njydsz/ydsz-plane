/**
 * 工作流与审批 API — 封装工作流（状态机组）与审批流的 REST 接口。
 *
 * 路由: /api/v1/workspaces/:ws/projects/:pid/workflows
 *       /api/v1/workspaces/:ws/approvals
 *
 * 说明：每个工作流是一组独立的「状态机」配置（状态 + 流转规则），
 * 流转规则可配置是否需要审批；审批提交后由指定角色/人员审批。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 工作流（独立状态机组） */
export interface Workflow {
  id: number;
  project_id: number;
  name: string;
  description: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
  transitions?: WorkflowTransition[];
}

/** 工作流流转规则 */
export interface WorkflowTransition {
  id: number;
  workflow_id: number;
  from_state_id: number;
  to_state_id: number;
  name: string;
  requires_approval: boolean;
  approver_role: string;
  approver_ids: number[];
  conditions: string;
  created_at: string;
}

/** 审批记录状态 */
export type ApprovalStatus = "pending" | "approved" | "rejected";

/** 审批记录 */
export interface WorkflowApproval {
  id: number;
  transition_id: number;
  work_item_id: number;
  work_item_type: string;
  requester_id: number;
  approver_id: number;
  status: ApprovalStatus;
  comment: string;
  created_at: string;
  resolved_at: string | null;
}

/** 创建工作流入参 */
export interface CreateWorkflowInput {
  name: string;
  description?: string;
  is_default?: boolean;
}

/** 更新工作流入参 */
export interface UpdateWorkflowInput {
  name?: string;
  description?: string;
  is_default?: boolean;
}

/** 添加流转规则入参 */
export interface AddTransitionInput {
  from_state_id: number;
  to_state_id: number;
  name: string;
  requires_approval?: boolean;
  approver_role?: string;
  approver_ids?: number[];
}

/** 审批列表查询参数 */
export interface ListApprovalsParams {
  status?: ApprovalStatus;
  page?: number;
}

/** 审批列表响应 */
export interface ListApprovalsResult {
  items: WorkflowApproval[];
  total: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 工作流与审批域 API */
export const workflowApi = {
  /* --- 工作流 CRUD --- */

  /** 列出项目下全部工作流 */
  list: (workspaceId: number | string, projectId: number | string) =>
    wrap<Workflow[]>(http.get(`/workspaces/${workspaceId}/projects/${projectId}/workflows`)),

  /** 获取单个工作流详情（含流转规则） */
  get: (workspaceId: number | string, projectId: number | string, id: number | string) =>
    wrap<Workflow>(http.get(`/workspaces/${workspaceId}/projects/${projectId}/workflows/${id}`)),

  /** 创建工作流 */
  create: (workspaceId: number | string, projectId: number | string, data: CreateWorkflowInput) =>
    wrap<Workflow>(http.post(`/workspaces/${workspaceId}/projects/${projectId}/workflows`, data)),

  /** 更新工作流 */
  update: (
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
    data: UpdateWorkflowInput,
  ) =>
    wrap<Workflow>(
      http.patch(`/workspaces/${workspaceId}/projects/${projectId}/workflows/${id}`, data),
    ),

  /** 删除工作流 */
  remove: (workspaceId: number | string, projectId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/projects/${projectId}/workflows/${id}`)),

  /* --- 流转规则 --- */

  /** 为工作流添加流转规则 */
  addTransition: (
    workspaceId: number | string,
    projectId: number | string,
    workflowId: number | string,
    data: AddTransitionInput,
  ) =>
    wrap<WorkflowTransition>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/workflows/${workflowId}/transitions`,
        data,
      ),
    ),

  /** 删除工作流中的流转规则 */
  removeTransition: (
    workspaceId: number | string,
    projectId: number | string,
    workflowId: number | string,
    transitionId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/workflows/${workflowId}/transitions/${transitionId}`,
      ),
    ),

  /** 校验某工作项能否执行该流转（前置条件检查） */
  validateTransition: (
    workspaceId: number | string,
    projectId: number | string,
    workflowId: number | string,
    transitionId: number | string,
    workItemId: number | string,
  ) =>
    wrap<{ valid: boolean; errors?: string[] }>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/workflows/${workflowId}/transitions/${transitionId}/validate`,
        { work_item_id: workItemId },
      ),
    ),

  /* --- 审批 --- */

  /** 提交审批（发起一次状态流转审批） */
  submitApproval: (
    workspaceId: number | string,
    data: { transition_id: number; work_item_id: number; work_item_type: string },
  ) =>
    wrap<WorkflowApproval>(http.post(`/workspaces/${workspaceId}/approvals`, data)),

  /** 列出工作空间下的审批记录 */
  listApprovals: (workspaceId: number | string, params?: ListApprovalsParams) =>
    wrap<ListApprovalsResult>(
      http.get(`/workspaces/${workspaceId}/approvals`, { params }),
    ),

  /** 通过审批 */
  approve: (workspaceId: number | string, id: number | string, comment?: string) =>
    wrap<WorkflowApproval>(
      http.post(`/workspaces/${workspaceId}/approvals/${id}/approve`, { comment }),
    ),

  /** 驳回审批 */
  reject: (workspaceId: number | string, id: number | string, comment?: string) =>
    wrap<WorkflowApproval>(
      http.post(`/workspaces/${workspaceId}/approvals/${id}/reject`, { comment }),
    ),
};
