/**
 * 工作空间视图 API — 工作空间级保存视图（跨项目）的 CRUD、复制与收藏 / 公开管理。
 *
 * 对标 Plane Workspace Views：把一组筛选 + 布局 + 排序配置保存为命名视图，
 * 可跨项目复用，支持收藏与公开共享。
 *
 * 路由（后端 internal/application/workspaceview/handler.go）:
 *   GET    /api/v1/workspaces/:ws/views
 *   POST   /api/v1/workspaces/:ws/views
 *   GET    /api/v1/workspaces/:ws/views/:viewId
 *   PATCH  /api/v1/workspaces/:ws/views/:viewId
 *   DELETE /api/v1/workspaces/:ws/views/:viewId
 *   POST   /api/v1/workspaces/:ws/views/:viewId/duplicate
 *   POST   /api/v1/workspaces/:ws/views/:viewId/favorite
 *   POST   /api/v1/workspaces/:ws/views/:viewId/public
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 视图布局形态（与后端 Layout 常量对齐） */
export type WorkspaceViewLayout = "board" | "list" | "spreadsheet" | "gantt" | "calendar";

/** 工作空间视图实体 */
export interface WorkspaceView {
  id: number;
  workspace_id: number;
  /** 视图名称 */
  name: string;
  description?: string;
  /** 布局形态 */
  layout: WorkspaceViewLayout;
  /** 筛选条件（JSON 对象） */
  filters: Record<string, unknown>;
  /** 展示选项（JSON 对象） */
  display_options: Record<string, unknown>;
  /** 排序配置（JSON 对象） */
  sort: Record<string, unknown>;
  is_public: boolean;
  is_favorite: boolean;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建视图入参（与后端 create 请求体对齐） */
export interface CreateWorkspaceViewInput {
  name: string;
  description?: string;
  layout: WorkspaceViewLayout;
  filters?: Record<string, unknown>;
  display_options?: Record<string, unknown>;
  sort?: Record<string, unknown>;
  is_public?: boolean;
  is_favorite?: boolean;
}

/** 更新视图入参（与后端 update 请求体对齐，全部可选） */
export interface UpdateWorkspaceViewInput {
  name?: string;
  description?: string;
  layout?: WorkspaceViewLayout;
  filters?: Record<string, unknown>;
  display_options?: Record<string, unknown>;
  sort?: Record<string, unknown>;
  is_public?: boolean;
  is_favorite?: boolean;
}

/** 视图列表查询参数 */
export interface ListWorkspaceViewsParams {
  /** 仅返回当前用户创建的视图 */
  mine?: boolean;
  /** 仅返回公开视图 */
  public?: boolean;
  /** 仅返回收藏的视图 */
  favorite?: boolean;
  limit?: number;
  offset?: number;
}

/** 视图列表响应 */
export interface ListWorkspaceViewsResult {
  results: WorkspaceView[];
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 工作空间视图域 API */
export const workspaceViewApi = {
  /** 列出工作空间下全部视图（含个人与公开） */
  list: (workspaceId: number | string, params?: ListWorkspaceViewsParams) =>
    wrap<ListWorkspaceViewsResult>(http.get(`/workspaces/${workspaceId}/views`, { params })),

  /** 获取单个视图 */
  get: (workspaceId: number | string, viewId: number | string) =>
    wrap<WorkspaceView>(http.get(`/workspaces/${workspaceId}/views/${viewId}`)),

  /** 创建视图 */
  create: (workspaceId: number | string, input: CreateWorkspaceViewInput) =>
    wrap<WorkspaceView>(http.post(`/workspaces/${workspaceId}/views`, input)),

  /** 更新视图 */
  update: (
    workspaceId: number | string,
    viewId: number | string,
    input: UpdateWorkspaceViewInput,
  ) =>
    wrap<WorkspaceView>(
      http.patch(`/workspaces/${workspaceId}/views/${viewId}`, input),
    ),

  /** 删除视图（204 无内容） */
  remove: (workspaceId: number | string, viewId: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/views/${viewId}`)),

  /** 复制视图（生成一份独立副本并返回，201） */
  duplicate: (workspaceId: number | string, viewId: number | string) =>
    wrap<WorkspaceView>(
      http.post(`/workspaces/${workspaceId}/views/${viewId}/duplicate`),
    ),

  /**
   * 切换收藏状态。
   * 后端为 Toggle 语义（不带请求体），返回切换后的 { is_favorite }。
   */
  toggleFavorite: (workspaceId: number | string, viewId: number | string) =>
    wrap<{ is_favorite: boolean }>(
      http.post(`/workspaces/${workspaceId}/views/${viewId}/favorite`),
    ),

  /**
   * 切换公开状态。
   * 后端为 Toggle 语义（不带请求体），返回切换后的 { is_public }。
   */
  togglePublic: (workspaceId: number | string, viewId: number | string) =>
    wrap<{ is_public: boolean }>(
      http.post(`/workspaces/${workspaceId}/views/${viewId}/public`),
    ),
};
