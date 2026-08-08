/**
 * 知识库（Knowledge Base）域 API — 对接后端 knowledge 域 REST 接口。
 *
 * 路由前缀：/api/v1/workspaces/:workspace_id/knowledge
 * 提供空间 CRUD、文档树（嵌套层级）、版本快照、文档与工作项关联。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* 枚举                                                                */
/* ------------------------------------------------------------------ */

/** 空间默认权限 */
export type SpacePermission = "viewer" | "editor" | "admin" | "owner";

/** 文档状态 */
export type PageStatus = "draft" | "published" | "archived";

/** 文档与工作项关联类型 */
export type PageRelationType = "referenced" | "referencing";

/* ------------------------------------------------------------------ */
/* 实体                                                                */
/* ------------------------------------------------------------------ */

/** 知识库空间。 */
export interface KnowledgeSpace {
  id: number;
  workspace_id: number;
  project_id?: number | null;
  name: string;
  slug: string;
  description?: string;
  owner_id?: number | null;
  default_permission: SpacePermission;
  is_private: boolean;
  cover_image?: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

/** 知识库文档。 */
export interface KnowledgePage {
  id: number;
  workspace_id: number;
  space_id: number;
  parent_id?: number | null;
  lft: number;
  rgt: number;
  depth: number;
  title: string;
  path?: string;
  content_md?: string;
  content_html?: string;
  version: number;
  status: PageStatus;
  sort_order: number;
  is_pinned: boolean;
  is_featured: boolean;
  view_count: number;
  created_by?: number | null;
  updated_by?: number | null;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

/** 文档树形节点（children 递归包含子节点）。 */
export interface KnowledgePageNode extends KnowledgePage {
  children: KnowledgePageNode[];
}

/** 文档版本快照。 */
export interface KnowledgePageVersion {
  id: number;
  page_id: number;
  version: number;
  title: string;
  content_md?: string;
  content_html?: string;
  change_summary?: string;
  created_by?: number | null;
  created_at: string;
}

/** 文档与工作项关联关系。 */
export interface KnowledgePageRelation {
  id: number;
  page_id: number;
  issue_id: number;
  relation_type: PageRelationType;
  created_at: string;
}

/* ------------------------------------------------------------------ */
/* 输入                                                                */
/* ------------------------------------------------------------------ */

/** 创建空间入参。 */
export interface CreateSpaceInput {
  name: string;
  slug: string;
  description?: string;
  default_permission?: SpacePermission;
  is_private?: boolean;
  cover_image?: string;
}

/** 更新空间入参（指针字段为 undefined 时不更新）。 */
export interface UpdateSpaceInput {
  name?: string;
  description?: string;
  default_permission?: SpacePermission;
  is_private?: boolean;
  cover_image?: string;
}

/** 创建文档入参。 */
export interface CreatePageInput {
  title: string;
  content_md?: string;
  content_html?: string;
  parent_id?: number | null;
  status?: PageStatus;
  sort_order?: number;
}

/** 更新文档入参（乐观锁 version 必传）。 */
export interface UpdatePageInput {
  title?: string;
  content_md?: string;
  content_html?: string;
  parent_id?: number | null;
  status?: PageStatus;
  sort_order?: number;
  is_pinned?: boolean;
  is_featured?: boolean;
  version: number;
  change_summary?: string;
}

/* ------------------------------------------------------------------ */
/* API 封装                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

const spacePath = (wsId: number, sid: number) =>
  `/workspaces/${wsId}/knowledge/spaces/${sid}`;

const pagePath = (wsId: number, sid: number, pid: number) =>
  `${spacePath(wsId, sid)}/pages/${pid}`;

/** 知识库域 API — 空间 / 文档 / 版本 / 关联。 */
export const knowledgeApi = {
  /* ===== 空间 ===== */

  /** 列出工作空间下全部知识库空间 */
  listSpaces: async (wsId: number) => {
    const data = await wrap<{ results?: KnowledgeSpace[]; total?: number }>(
      http.get(`/workspaces/${wsId}/knowledge/spaces`),
    );
    return data?.results ?? [];
  },

  /** 获取单个空间详情 */
  getSpace: (wsId: number, sid: number) =>
    wrap<KnowledgeSpace>(http.get(spacePath(wsId, sid))),

  /** 创建知识库空间 */
  createSpace: (wsId: number, input: CreateSpaceInput) =>
    wrap<KnowledgeSpace>(http.post(`/workspaces/${wsId}/knowledge/spaces`, input)),

  /** 更新知识库空间 */
  updateSpace: (wsId: number, sid: number, input: UpdateSpaceInput) =>
    wrap<KnowledgeSpace>(http.patch(spacePath(wsId, sid), input)),

  /** 删除知识库空间（软删除） */
  deleteSpace: (wsId: number, sid: number) =>
    wrap<void>(http.delete(spacePath(wsId, sid))),

  /* ===== 文档 ===== */

  /** 获取空间下全部文档的树形结构 */
  getPageTree: async (wsId: number, sid: number) => {
    const data = await wrap<{ results?: KnowledgePageNode[] }>(
      http.get(`${spacePath(wsId, sid)}/pages`),
    );
    return data?.results ?? [];
  },

  /** 获取单个文档详情 */
  getPage: (wsId: number, sid: number, pid: number) =>
    wrap<KnowledgePage>(http.get(pagePath(wsId, sid, pid))),

  /** 创建文档（parent_id 为空 = 根文档） */
  createPage: (wsId: number, sid: number, input: CreatePageInput) =>
    wrap<KnowledgePage>(http.post(`${spacePath(wsId, sid)}/pages`, input)),

  /** 更新文档（乐观锁，版本号自动递增并生成快照） */
  updatePage: (wsId: number, sid: number, pid: number, input: UpdatePageInput) =>
    wrap<KnowledgePage>(http.patch(pagePath(wsId, sid, pid), input)),

  /** 删除文档（软删除，子文档级联） */
  deletePage: (wsId: number, sid: number, pid: number) =>
    wrap<void>(http.delete(pagePath(wsId, sid, pid))),

  /* ===== 版本 ===== */

  /** 列出文档的版本快照列表 */
  listVersions: async (wsId: number, sid: number, pid: number) => {
    const data = await wrap<{ results?: KnowledgePageVersion[] }>(
      http.get(`${pagePath(wsId, sid, pid)}/versions`),
    );
    return data?.results ?? [];
  },

  /** 回滚到指定版本（复制快照内容为最新版本，版本号 +1） */
  revertVersion: (wsId: number, sid: number, pid: number, version: number) =>
    wrap<KnowledgePage>(http.post(`${pagePath(wsId, sid, pid)}/revert`, { version })),

  /* ===== 关联工作项 ===== */

  /** 列出文档关联的工作项 */
  listRelations: async (wsId: number, sid: number, pid: number) => {
    const data = await wrap<{ results?: KnowledgePageRelation[] }>(
      http.get(`${pagePath(wsId, sid, pid)}/relations`),
    );
    return data?.results ?? [];
  },

  /** 添加文档与工作项的关联 */
  addRelation: (wsId: number, sid: number, pid: number, issueId: number) =>
    wrap<KnowledgePageRelation>(
      http.post(`${pagePath(wsId, sid, pid)}/relations`, { issue_id: issueId }),
    ),

  /** 移除文档与工作项的关联 */
  removeRelation: (wsId: number, sid: number, pid: number, rid: number) =>
    wrap<void>(http.delete(`${pagePath(wsId, sid, pid)}/relations/${rid}`)),

  /* ===== 全文检索 ===== */

  /** 全文检索（simple tsvector）。可选 space_id 限定空间 */
  search: async (wsId: number, q: string, spaceId?: number) => {
    const params: Record<string, string | number> = { q };
    if (spaceId != null) params.space_id = spaceId;
    const data = await wrap<{ results?: KnowledgePage[]; total?: number }>(
      http.get(`/workspaces/${wsId}/knowledge/search`, { params }),
    );
    return data?.results ?? [];
  },
};
