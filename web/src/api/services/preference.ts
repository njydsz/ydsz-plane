/**
 * 视图偏好 API — 对接后端 preference 域 REST 接口。
 */
import { http } from "../client";

export type ViewType = "kanban" | "list" | "calendar" | "gantt";

export interface ViewPreference {
  id?: number;
  workspace_id: number;
  project_id: number;
  user_id: number;
  view_type: ViewType;
  layout: string;
  columns: unknown[];
  filters: Record<string, unknown>;
  sort: { field?: string; desc?: boolean };
  extra: Record<string, unknown>;
}

/** 视图偏好域 API */
export const preferenceApi = {
  get: (wsId: number, projectId: number, viewType: ViewType) =>
    http
      .get<ViewPreference | Record<string, never>>(
        `/workspaces/${wsId}/projects/${projectId}/preferences/${viewType}`,
      )
      .then((r) => r.data),

  save: (wsId: number, projectId: number, viewType: ViewType, input: Partial<ViewPreference>) =>
    http
      .put<ViewPreference>(`/workspaces/${wsId}/projects/${projectId}/preferences/${viewType}`, input)
      .then((r) => r.data),

  list: (wsId: number, projectId: number) =>
    http
      .get<{ results: ViewPreference[] }>(`/workspaces/${wsId}/projects/${projectId}/preferences`)
      .then((r) => r.data.results),
};
