/**
 * 模块域 API — 对接后端 Module 域 REST 接口。
 *
 * 模块是项目下的一级分类维度，用于分组管理工作项（如"用户模块"、"支付模块"）。
 * 后端 modules 表含 id/project_id/name/description/lead_id/status/sort_order/created_at/updated_at。
 */
import { http } from "../client";

/** 模块实体 */
export interface Module {
  id: number;
  project_id: number;
  name: string;
  description?: string;
  lead_id?: number;
  lead_name?: string;
  status: string;
  sort_order: number;
  issue_count?: number;
  created_at: string;
  updated_at: string;
}

/** 创建模块入参 */
export interface CreateModuleInput {
  name: string;
  description?: string;
  lead_id?: number;
  sort_order?: number;
}

/** 更新模块入参 */
export interface UpdateModuleInput {
  name?: string;
  description?: string;
  lead_id?: number;
  status?: string;
  sort_order?: number;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 模块域 API — 模块 CRUD */
export const moduleApi = {
  /** 列出项目下全部模块 */
  list: (wsId: number, projectId: number) =>
    wrap<Module[]>(http.get(`/workspaces/${wsId}/projects/${projectId}/modules`)),

  /** 创建模块 */
  create: (wsId: number, projectId: number, input: CreateModuleInput) =>
    wrap<Module>(http.post(`/workspaces/${wsId}/projects/${projectId}/modules`, input)),

  /** 更新模块 */
  update: (wsId: number, projectId: number, moduleId: number, input: UpdateModuleInput) =>
    wrap<Module>(http.patch(`/workspaces/${wsId}/projects/${projectId}/modules/${moduleId}`, input)),

  /** 删除模块 */
  remove: (wsId: number, projectId: number, moduleId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/modules/${moduleId}`)),
};
