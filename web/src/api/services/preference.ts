/**
 * 视图偏好 API — 对接后端 preference 域。
 *
 * 对标 Plane 的 view preferences：按 (workspace, project, user, view_type)
 * 持久化每个用户的布局/过滤/排序配置。
 */
import { http } from "../client";

/** 偏好视图类型（看板 / 列表 / 日历 / 甘特图 / 表格）。 */
export type PreferenceViewType = "kanban" | "list" | "calendar" | "gantt" | "spreadsheet";

/** 视图偏好实体 — 按 (workspace, project, user, view_type) 持久化的布局 / 过滤 / 排序配置。 */
export interface ViewPreference {
  id: number;
  workspace_id: number;
  project_id: number;
  user_id: number;
  view_type: PreferenceViewType;
  layout: string;
  columns: unknown;
  filters: unknown;
  sort: unknown;
  extra: unknown;
  created_at: string;
  updated_at: string;
}

/** 保存视图偏好入参（layout/columns/filters/sort/extra 字段部分更新）。 */
export interface SavePreferenceInput {
  layout?: string;
  columns?: unknown;
  filters?: unknown;
  sort?: unknown;
  extra?: unknown;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 视图偏好域 API — 按 (workspace, project, user, view_type) 读写布局 / 过滤 / 排序配置。 */
export const preferenceApi = {
  /** 获取指定视图偏好（无偏好时返回 null） */
  get: async (wsId: number, projectId: number, viewType: PreferenceViewType) => {
    const data = await wrap<Partial<ViewPreference>>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/preferences/${viewType}`),
    );
    return data && data.id ? (data as ViewPreference) : null;
  },

  /** 保存（upsert）视图偏好 */
  save: (wsId: number, projectId: number, viewType: PreferenceViewType, input: SavePreferenceInput) =>
    wrap<ViewPreference>(
      http.put(`/workspaces/${wsId}/projects/${projectId}/preferences/${viewType}`, input),
    ),

  /** 列出全部视图偏好 */
  list: (wsId: number, projectId: number) =>
    wrap<ViewPreference[]>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/preferences`),
    ),
};

// ========== 命名视图（Saved Views） ==========

/** 视图范围 */
export type ViewScope = "personal" | "team" | "default";

/** 命名视图实体 */
export interface SavedView {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  type: PreferenceViewType;
  scope: ViewScope;
  config: Record<string, unknown>;
  owner_id: number;
  is_shared: boolean;
  created_at: string;
  updated_at: string;
}

/** 创建命名视图入参 */
export interface CreateViewInput {
  name: string;
  type: PreferenceViewType;
  scope?: ViewScope;
  config: Record<string, unknown>;
  is_shared?: boolean;
}

/** 更新命名视图入参（部分更新） */
export interface UpdateViewInput {
  name?: string;
  type?: PreferenceViewType;
  scope?: ViewScope;
  config?: Record<string, unknown>;
  is_shared?: boolean;
}

/** 命名视图 API — 管理保存的过滤/排序/分组/列配置。 */
export const viewsApi = {
  /** 列出项目下视图 */
  list: (wsId: number, projectId: number, scope?: ViewScope) =>
    wrap<{ results: SavedView[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/views`, {
        params: scope ? { scope } : undefined,
      }),
    ).then((r) => r.results),

  /** 获取单个视图 */
  get: (wsId: number, projectId: number, viewId: number) =>
    wrap<SavedView>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/views/${viewId}`),
    ),

  /** 创建视图 */
  create: (wsId: number, projectId: number, input: CreateViewInput) =>
    wrap<SavedView>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/views`, input),
    ),

  /** 更新视图（部分更新） */
  update: (wsId: number, projectId: number, viewId: number, input: UpdateViewInput) =>
    wrap<SavedView>(
      http.patch(`/workspaces/${wsId}/projects/${projectId}/views/${viewId}`, input),
    ),

  /** 删除视图 */
  delete: (wsId: number, projectId: number, viewId: number) =>
    http.delete(`/workspaces/${wsId}/projects/${projectId}/views/${viewId}`),

  /** 设为默认视图（管理员） */
  setDefault: (wsId: number, projectId: number, viewId: number) =>
    http.post(`/workspaces/${wsId}/projects/${projectId}/views/${viewId}/default`),
};
