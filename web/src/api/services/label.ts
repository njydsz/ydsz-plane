/**
 * 标签域 API — 项目级标签独立 CRUD（对标 Plane / Linear 的 Label 概念）。
 *
 * 标签是需求/任务/缺陷的轻量分类维度（M2M 经 issue_labels 关联），
 * 本模块提供标签实体本身的 创建/列表/详情/更新/删除。
 */
import { http } from "../client";

/** 标签实体 */
export interface Label {
  id: number;
  code?: string;
  workspace_id: number;
  project_id: number;
  name: string;
  color: string;
  description?: string;
  status?: string;
  issue_count?: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建标签入参 */
export interface CreateLabelInput {
  name: string;
  color?: string;
  description?: string;
}

/** 更新标签入参 */
export interface UpdateLabelInput {
  name?: string;
  color?: string;
  description?: string;
}

interface LabelList {
  results: Label[];
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 标签域 API — 标签 CRUD */
export const labelApi = {
  /** 列出项目下全部标签 */
  list: (wsId: number, projectId: number, status?: string) =>
    wrap<LabelList>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/labels`, {
        params: status ? { status } : undefined,
      }),
    ),

  /** 创建标签 */
  create: (wsId: number, projectId: number, input: CreateLabelInput) =>
    wrap<Label>(http.post(`/workspaces/${wsId}/projects/${projectId}/labels`, input)),

  /** 获取单个标签 */
  get: (wsId: number, projectId: number, labelId: number) =>
    wrap<Label>(http.get(`/workspaces/${wsId}/projects/${projectId}/labels/${labelId}`)),

  /** 更新标签 */
  update: (wsId: number, projectId: number, labelId: number, input: UpdateLabelInput) =>
    wrap<Label>(http.patch(`/workspaces/${wsId}/projects/${projectId}/labels/${labelId}`, input)),

  /** 删除标签 */
  remove: (wsId: number, projectId: number, labelId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/labels/${labelId}`)),
};
