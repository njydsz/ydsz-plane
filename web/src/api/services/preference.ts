/**
 * 视图偏好 API — 对接后端 preference 域。
 *
 * 对标 Plane 的 view preferences：按 (workspace, project, user, view_type)
 * 持久化每个用户的布局/过滤/排序配置。
 */
import { http } from "../client";

/** 偏好视图类型（看板 / 列表 / 日历 / 甘特图）。 */
export type PreferenceViewType = "kanban" | "list" | "calendar" | "gantt";

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

export interface SavePreferenceInput {
  layout?: string;
  columns?: unknown;
  filters?: unknown;
  sort?: unknown;
  extra?: unknown;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

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
