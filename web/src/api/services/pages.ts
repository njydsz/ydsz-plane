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
  category?: string | null;
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
  category?: string;
}

/** 更新文档页入参（可选字段 + 乐观锁 version）。 */
export interface UpdatePageInput {
  name?: string;
  description_json?: string;
  description_html?: string;
  description_stripped?: string;
  parent_id?: number | null;
  sort_order?: number;
  category?: string;
  version: number;
}

/** 文档版本快照实体。 */
export interface DocumentVersion {
  id: number;
  page_id: number;
  version_number: number;
  content_md?: string;
  content_html?: string;
  created_by: number;
  created_at: string;
}

/** 文档关联实体。 */
export interface DocumentLink {
  id: number;
  page_id: number;
  linkable_type: string;
  linkable_id: number;
  created_by: number;
  created_at: string;
}

/** 文档模板实体。 */
export interface PageTemplate {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  description?: string | null;
  content_html?: string | null;
  category?: string | null;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建文档模板入参。 */
export interface CreateTemplateInput {
  name: string;
  description?: string;
  content_html?: string;
  category?: string;
}

/** 更新文档模板入参（可选字段）。 */
export interface UpdateTemplateInput {
  name?: string;
  description?: string;
  content_html?: string;
  category?: string;
}

/** 文档公开分享链接实体。 */
export interface PageShare {
  id: number;
  page_id: number;
  workspace_id: number;
  project_id: number;
  token: string;
  is_active: boolean;
  expires_at?: string | null;
  created_by: number;
  created_at: string;
  has_password?: boolean; // 前端推断：列表时由后端不返回 password_hash，但可单独标记
}

/** 创建分享链接入参。 */
export interface CreateShareInput {
  password?: string;
  expires_at?: string | null;
}

/** 更新分享链接入参。 */
export interface UpdateShareInput {
  is_active?: boolean;
  password?: string | null; // 空字符串 = 清除密码
  expires_at?: string | null;
}

/** 公开分享页面视图。 */
export interface PublicSharePage {
  page_id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  description_html?: string | null;
  description_json?: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
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

const basePath = (wsId: number, projectId: number, pageId: number) =>
  `/workspaces/${wsId}/projects/${projectId}/pages/${pageId}`;

/** 文档版本历史 API */
export const versionsApi = {
  /** 列出页面的所有历史版本 */
  list: async (wsId: number, projectId: number, pageId: number) => {
    const data = await wrap<{ results?: DocumentVersion[] }>(
      http.get(`${basePath(wsId, projectId, pageId)}/versions`),
    );
    return data?.results ?? [];
  },

  /** 获取指定版本快照 */
  get: (wsId: number, projectId: number, pageId: number, versionNumber: number) =>
    wrap<DocumentVersion>(
      http.get(`${basePath(wsId, projectId, pageId)}/versions/${versionNumber}`),
    ),

  /** 回滚到指定版本 */
  rollback: (wsId: number, projectId: number, pageId: number, versionNumber: number) =>
    wrap<Page>(
      http.post(`${basePath(wsId, projectId, pageId)}/versions/${versionNumber}/rollback`),
    ),
};

/** 文档关联 API */
export const linksApi = {
  /** 列出页面的所有关联 */
  list: async (wsId: number, projectId: number, pageId: number) => {
    const data = await wrap<{ results?: DocumentLink[] }>(
      http.get(`${basePath(wsId, projectId, pageId)}/links`),
    );
    return data?.results ?? [];
  },

  /** 创建关联 */
  create: (
    wsId: number,
    projectId: number,
    pageId: number,
    input: { linkable_type: string; linkable_id: number },
  ) =>
    wrap<DocumentLink>(
      http.post(`${basePath(wsId, projectId, pageId)}/links`, input),
    ),

  /** 删除关联 */
  remove: (wsId: number, projectId: number, pageId: number, linkId: number) =>
    wrap<void>(http.delete(`${basePath(wsId, projectId, pageId)}/links/${linkId}`)),
};

/** 文档模板 API */
export const templatesApi = {
  /** 列出项目可用模板 */
  list: async (wsId: number, projectId: number) => {
    const data = await wrap<{ results?: PageTemplate[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/templates`),
    );
    return data?.results ?? [];
  },

  /** 创建模板 */
  create: (wsId: number, projectId: number, input: CreateTemplateInput) =>
    wrap<PageTemplate>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/templates`, input),
    ),

  /** 更新模板 */
  update: (wsId: number, projectId: number, templateId: number, input: UpdateTemplateInput) =>
    wrap<PageTemplate>(
      http.patch(`/workspaces/${wsId}/projects/${projectId}/templates/${templateId}`, input),
    ),

  /** 删除模板 */
  remove: (wsId: number, projectId: number, templateId: number) =>
    wrap<void>(
      http.delete(`/workspaces/${wsId}/projects/${projectId}/templates/${templateId}`),
    ),
};

/** 文档公开分享 API */
export const sharesApi = {
  /** 创建分享链接 */
  create: (wsId: number, projectId: number, pageId: number, input: CreateShareInput) =>
    wrap<PageShare>(
      http.post(`${basePath(wsId, projectId, pageId)}/shares`, input),
    ),

  /** 列出页面分享链接 */
  list: async (wsId: number, projectId: number, pageId: number) => {
    const data = await wrap<{ results?: PageShare[] }>(
      http.get(`${basePath(wsId, projectId, pageId)}/shares`),
    );
    return data?.results ?? [];
  },

  /** 更新分享链接 */
  update: (wsId: number, projectId: number, pageId: number, shareId: number, input: UpdateShareInput) =>
    wrap<PageShare>(
      http.patch(`${basePath(wsId, projectId, pageId)}/shares/${shareId}`, input),
    ),

  /** 吊销分享链接 */
  revoke: (wsId: number, projectId: number, pageId: number, shareId: number) =>
    wrap<void>(
      http.delete(`${basePath(wsId, projectId, pageId)}/shares/${shareId}`),
    ),
};

/** 公开分享页面 — 免登录访问 */
export const publicPagesApi = {
  /** 获取公开分享页面内容
   * @param token 分享令牌
   * @param password 可选访问密码
   * @returns 页面内容；若需要密码，后端返回 require_password: true（HTTP 401）
   */
  getShared: (token: string, password?: string): Promise<PublicSharePage> => {
    const qs = password ? `?password=${encodeURIComponent(password)}` : "";
    return wrap<PublicSharePage>(http.get(`/public/pages/${token}${qs}`));
  },
};
