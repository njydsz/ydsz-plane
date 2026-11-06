/**
 * 项目状态与更新 API。
 *
 * 路由: /api/v1/workspaces/:ws/projects/:projectId/project-state
 *       /api/v1/workspaces/:ws/projects/:projectId/project-updates
 *       /api/v1/workspaces/:ws/initiative-updates
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 项目状态 */
export type ProjectStatus =
  | "on_track"
  | "at_risk"
  | "off_track"
  | "completed"
  | "paused";
/** 优先级 */
export type ProjectPriority = "urgent" | "high" | "medium" | "low" | "none";

/** 项目状态记录 */
export interface ProjectState {
  id: number;
  project_id: number;
  status: ProjectStatus;
  priority: ProjectPriority;
  start_date: string | null;
  target_date: string | null;
  owner_id: number;
  description: string;
  updated_at: string;
}

/** 项目更新 */
export interface ProjectUpdate {
  id: number;
  workspace_id: number;
  project_id: number;
  author_id: number;
  status: ProjectStatus;
  content: string;
  progress_percent: number;
  initiative_id: number | null;
  created_at: string;
}

/** 更新表情回应 */
export interface ProjectUpdateReaction {
  id: number;
  update_id: number;
  user_id: number;
  emoji: string;
  created_at: string;
}

/** 更新评论 */
export interface ProjectUpdateComment {
  id: number;
  update_id: number;
  author_id: number;
  parent_id: number | null;
  content: string;
  created_at: string;
  deleted_at: string | null;
}

/** 创建更新入参 */
export interface CreateProjectUpdateInput {
  status: ProjectStatus;
  content: string;
  progress_percent?: number;
  initiative_id?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 项目状态与更新 API */
export const projectStateApi = {
  /* --- 状态 --- */

  getState: (workspaceId: number | string, projectId: number | string) =>
    wrap<ProjectState>(
      http.get(`/workspaces/${workspaceId}/projects/${projectId}/project-state`),
    ),

  upsertState: (
    workspaceId: number | string,
    projectId: number | string,
    data: {
      status: ProjectStatus;
      priority?: ProjectPriority;
      start_date?: string;
      target_date?: string;
      owner_id?: number;
      description?: string;
    },
  ) =>
    wrap<ProjectState>(
      http.put(`/workspaces/${workspaceId}/projects/${projectId}/project-state`, data),
    ),

  /* --- 更新 --- */

  listUpdates: (
    workspaceId: number | string,
    projectId: number | string,
    params?: { limit?: number; offset?: number },
  ) =>
    wrap<{ items: ProjectUpdate[] }>(
      http.get(`/workspaces/${workspaceId}/projects/${projectId}/project-updates`, {
        params,
      }),
    ),

  listInitiativeUpdates: (
    workspaceId: number | string,
    initiativeId: number,
    params?: { limit?: number; offset?: number },
  ) =>
    wrap<{ items: ProjectUpdate[] }>(
      http.get(`/workspaces/${workspaceId}/initiative-updates`, {
        params: { initiative_id: initiativeId, ...params },
      }),
    ),

  createUpdate: (
    workspaceId: number | string,
    projectId: number | string,
    data: CreateProjectUpdateInput,
  ) =>
    wrap<ProjectUpdate>(
      http.post(`/workspaces/${workspaceId}/projects/${projectId}/project-updates`, data),
    ),

  removeUpdate: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}`,
      ),
    ),

  /* --- 反应 --- */

  listReactions: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
  ) =>
    wrap<{ items: ProjectUpdateReaction[] }>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}/reactions`,
      ),
    ),

  addReaction: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
    emoji: string,
  ) =>
    wrap<void>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}/reactions`,
        { emoji },
      ),
    ),

  removeReaction: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
    emoji: string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}/reactions/${encodeURIComponent(emoji)}`,
      ),
    ),

  /* --- 评论 --- */

  listComments: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
  ) =>
    wrap<{ items: ProjectUpdateComment[] }>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}/comments`,
      ),
    ),

  addComment: (
    workspaceId: number | string,
    projectId: number | string,
    updateId: number | string,
    content: string,
    parentId?: number,
  ) =>
    wrap<ProjectUpdateComment>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/${updateId}/comments`,
        { content, parent_id: parentId },
      ),
    ),

  removeComment: (
    workspaceId: number | string,
    projectId: number | string,
    commentId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/project-updates/comments/${commentId}`,
      ),
    ),
};
