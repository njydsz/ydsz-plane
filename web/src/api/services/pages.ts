/**
 * Pages 文档域 API — 对接后端 Pages 域 REST 接口。
 *
 * 对标 Plane 的 Pages 功能：项目内文档树，支持富文本内容与嵌套层级。
 */
import { http } from "../client";

/** 文档页面实体（富文本内容 + 嵌套层级 + 乐观锁版本）。 */
export interface Page {
  id: number;
  public_id: string;
  workspace_id: number;
  project_id: number;
  name: string;
  description_json?: Record<string, unknown> | null;
  description_html?: string | null;
  description_stripped?: string | null;
  parent_id?: number | null;
  sort_order: number;
  created_by: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
  version: number;
}

/** 创建文档页入参（name 必填，描述 + 父节点可选）。 */
export interface CreatePageInput {
  name: string;
  description_json?: string;
  description_html?: string;
  description_stripped?: string;
  parent_id?: number | null;
  sort_order?: number;
}

/** 更新文档页入参（可选字段 + 乐观锁 version）。 */
export interface UpdatePageInput {
  name?: string;
  description_json?: string;
  description_html?: string;
  description_stripped?: string;
  parent_id?: number | null;
  sort_order?: number;
  version: number;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** Pages 文档域 API — 文档页 CRUD（列表、创建、更新、删除）。 */
export const pagesApi = {
  /** 列出项目全部文档页面（扁平列表，由前端组装树） */
  list: async (wsId: number, projectId: number) => {
    const data = await wrap<{ results?: Page[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/pages`),
    );
    return data?.results ?? (Array.isArray(data) ? (data as Page[]) : []);
  },

  /** 获取单个文档页面 */
  get: (wsId: number, projectId: number, pageId: number) =>
    wrap<Page>(http.get(`/workspaces/${wsId}/projects/${projectId}/pages/${pageId}`)),

  /** 创建文档页面 */
  create: (wsId: number, projectId: number, input: CreatePageInput) =>
    wrap<Page>(http.post(`/workspaces/${wsId}/projects/${projectId}/pages`, input)),

  /** 更新文档页面（乐观锁） */
  update: (wsId: number, projectId: number, pageId: number, input: UpdatePageInput) =>
    wrap<Page>(http.patch(`/workspaces/${wsId}/projects/${projectId}/pages/${pageId}`, input)),

  /** 删除文档页面（软删除） */
  remove: (wsId: number, projectId: number, pageId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/pages/${pageId}`)),
};
