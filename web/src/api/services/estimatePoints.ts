/**
 * 估算点数 API — 项目级估算点数配置 CRUD。
 *
 * 估算点数是工作量对标参照系（斐波那契数列或 T 恤码等），
 * 工作项发布时用于评估规模。
 */
import { apiClient } from "../client";

/** 单条估算点数 */
export interface EstimatePointDef {
  label: string;
  value: number;
  color?: string;
}

/** 估算点数实体 */
export interface EstimatePoint {
  id: number;
  code?: string;
  workspace_id: number;
  project_id: number;
  name: string;
  description?: string;
  points: EstimatePointDef[];
  is_default: boolean;
  status?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建估算点数入参 */
export interface CreateEstimatePointInput {
  name: string;
  description?: string;
  points?: EstimatePointDef[];
  is_default?: boolean;
}

/** 更新估算点数入参 */
export interface UpdateEstimatePointInput {
  name?: string;
  description?: string;
  points?: EstimatePointDef[];
  is_default?: boolean;
}

/** 估算点数 API */
export const estimatePointApi = {
  /** 列出项目估算点数 */
  list: (workspaceId: number, projectId: number, status?: string) =>
    apiClient
      .get<{ results: EstimatePoint[] }>(
        `/workspaces/${workspaceId}/projects/${projectId}/estimate-points${status ? `?status=${status}` : ""}`,
      )
      .then((r) => r.data),

  /** 获取单个估算点数 */
  get: (workspaceId: number, projectId: number, pointId: number) =>
    apiClient
      .get<EstimatePoint>(
        `/workspaces/${workspaceId}/projects/${projectId}/estimate-points/${pointId}`,
      )
      .then((r) => r.data),

  /** 创建估算点数 */
  create: (workspaceId: number, projectId: number, input: CreateEstimatePointInput) =>
    apiClient
      .post<EstimatePoint>(
        `/workspaces/${workspaceId}/projects/${projectId}/estimate-points`,
        input,
      )
      .then((r) => r.data),

  /** 更新估算点数 */
  update: (
    workspaceId: number,
    projectId: number,
    pointId: number,
    input: UpdateEstimatePointInput,
  ) =>
    apiClient
      .patch<EstimatePoint>(
        `/workspaces/${workspaceId}/projects/${projectId}/estimate-points/${pointId}`,
        input,
      )
      .then((r) => r.data),

  /** 删除估算点数 */
  remove: (workspaceId: number, projectId: number, pointId: number) =>
    apiClient
      .delete(
        `/workspaces/${workspaceId}/projects/${projectId}/estimate-points/${pointId}`,
      )
      .then(() => {}),
};
