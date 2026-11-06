/**
 * 客户需求管理 API — 客户记录 + 客户请求。
 *
 * 路由: /api/v1/workspaces/:ws/customers, /customer-requests
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 客户阶段 */
export type CustomerStage = "lead" | "prospect" | "customer" | "churned";
/** 客户等级 */
export type CustomerTier = "free" | "pro" | "enterprise";
/** 请求状态 */
export type RequestStatus =
  | "new"
  | "acknowledged"
  | "in_progress"
  | "resolved"
  | "closed";
/** 请求优先级 */
export type RequestPriority = "urgent" | "high" | "medium" | "low";
/** 请求来源 */
export type RequestSource = "email" | "form" | "in_app" | "phone";

/** 客户记录 */
export interface Customer {
  id: number;
  workspace_id: number;
  name: string;
  email: string;
  website: string;
  industry: string;
  stage: CustomerStage;
  tier: CustomerTier;
  region: string;
  employees: number;
  revenue: string;
  account_manager_id: number;
  notes: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  request_count?: number;
}

/** 客户请求 */
export interface CustomerRequest {
  id: number;
  workspace_id: number;
  customer_id: number;
  title: string;
  description: string;
  status: RequestStatus;
  priority: RequestPriority;
  source: RequestSource;
  work_item_id: number | null;
  work_item_type: string;
  assigned_to: number;
  created_at: string;
  updated_at: string;
  customer_name?: string;
}

/** 客户统计 */
export interface CustomerStats {
  customer_id: number;
  total_requests: number;
  by_status: Record<string, number>;
}

/** 创建客户入参 */
export interface CreateCustomerInput {
  name: string;
  email?: string;
  website?: string;
  industry?: string;
  stage?: CustomerStage;
  tier?: CustomerTier;
  region?: string;
  employees?: number;
  revenue?: string;
  account_manager_id?: number;
  notes?: string;
}

/** 更新客户入参 */
export interface UpdateCustomerInput {
  name?: string;
  email?: string;
  website?: string;
  industry?: string;
  stage?: CustomerStage;
  tier?: CustomerTier;
  region?: string;
  employees?: number;
  revenue?: string;
  account_manager_id?: number;
  notes?: string;
}

/** 创建客户请求入参 */
export interface CreateCustomerRequestInput {
  customer_id: number;
  title: string;
  description?: string;
  status?: RequestStatus;
  priority?: RequestPriority;
  source?: RequestSource;
  assigned_to?: number;
}

/** 更新客户请求入参 */
export interface UpdateCustomerRequestInput {
  title?: string;
  description?: string;
  status?: RequestStatus;
  priority?: RequestPriority;
  source?: RequestSource;
  assigned_to?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 客户需求域 API */
export const customerApi = {
  /* --- 客户 --- */

  list: (
    workspaceId: number | string,
    params?: {
      stage?: string;
      tier?: string;
      industry?: string;
      q?: string;
      limit?: number;
      offset?: number;
    },
  ) => wrap<{ items: Customer[] }>(http.get(`/workspaces/${workspaceId}/customers`, { params })),

  search: (workspaceId: number | string, q: string) =>
    wrap<{ items: Customer[] }>(
      http.get(`/workspaces/${workspaceId}/customers/search`, { params: { q } }),
    ),

  get: (workspaceId: number | string, id: number | string) =>
    wrap<Customer>(http.get(`/workspaces/${workspaceId}/customers/${id}`)),

  create: (workspaceId: number | string, data: CreateCustomerInput) =>
    wrap<Customer>(http.post(`/workspaces/${workspaceId}/customers`, data)),

  update: (
    workspaceId: number | string,
    id: number | string,
    data: UpdateCustomerInput,
  ) =>
    wrap<Customer>(http.patch(`/workspaces/${workspaceId}/customers/${id}`, data)),

  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/customers/${id}`)),

  stats: (workspaceId: number | string, id: number | string) =>
    wrap<CustomerStats>(http.get(`/workspaces/${workspaceId}/customers/${id}/stats`)),

  /* --- 客户请求 --- */

  listRequests: (
    workspaceId: number | string,
    params?: {
      customer_id?: number;
      status?: string;
      priority?: string;
      limit?: number;
      offset?: number;
    },
  ) =>
    wrap<{ items: CustomerRequest[] }>(
      http.get(`/workspaces/${workspaceId}/customer-requests`, { params }),
    ),

  getRequest: (workspaceId: number | string, id: number | string) =>
    wrap<CustomerRequest>(
      http.get(`/workspaces/${workspaceId}/customer-requests/${id}`),
    ),

  createRequest: (
    workspaceId: number | string,
    data: CreateCustomerRequestInput,
  ) =>
    wrap<CustomerRequest>(
      http.post(`/workspaces/${workspaceId}/customer-requests`, data),
    ),

  updateRequest: (
    workspaceId: number | string,
    id: number | string,
    data: UpdateCustomerRequestInput,
  ) =>
    wrap<CustomerRequest>(
      http.patch(`/workspaces/${workspaceId}/customer-requests/${id}`, data),
    ),

  removeRequest: (workspaceId: number | string, id: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/customer-requests/${id}`),
    ),

  linkToWorkItem: (
    workspaceId: number | string,
    id: number | string,
    workItemId: number,
    workItemType: string,
  ) =>
    wrap<CustomerRequest>(
      http.post(`/workspaces/${workspaceId}/customer-requests/${id}/link`, {
        work_item_id: workItemId,
        work_item_type: workItemType,
      }),
    ),
};
