/**
 * 里程碑 API — 项目级里程碑实体的 CRUD、进度查询与工作项关联管理。
 *
 * 对标 Plane Milestones：把项目拆成若干里程碑节点，按目标日期推进，
 * 并可挂载需求/任务/缺陷以统计完成进度。
 *
 * 路由（后端 internal/application/milestone/handler.go）:
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/milestones
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/milestones
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId
 *   PATCH  /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId
 *   DELETE /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId/progress
 *   GET    /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId/work-items
 *   POST   /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId/work-items
 *   DELETE /api/v1/workspaces/:ws/projects/:projectId/milestones/:milestoneId/work-items/:workItemId
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 里程碑状态（与后端 MilestoneStatus 常量对齐） */
export type MilestoneStatus = "not_started" | "in_progress" | "completed" | "missed";

/** 里程碑实体 */
export interface Milestone {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  description?: string;
  /** 目标日期（YYYY-MM-DD），无则为 null */
  target_date?: string | null;
  status: MilestoneStatus;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 里程碑关联的工作项 */
export interface MilestoneWorkItem {
  milestone_id: number;
  work_item_id: number;
  work_item_type: string;
  created_at: string;
}

/** 里程碑进度聚合 */
export interface MilestoneProgress {
  /** 关联工作项总数 */
  total: number;
  /** 已完成（状态 group=done）数量 */
  completed: number;
  /** 进行中数量 */
  in_progress: number;
  /** 未开始数量 */
  not_started: number;
  /** 完成百分比 0-100 */
  percent: number;
  computed_at: string;
}

/** 创建里程碑入参 */
export interface CreateMilestoneInput {
  name: string;
  description?: string;
  target_date?: string | null;
  status?: MilestoneStatus;
}

/** 更新里程碑入参 */
export interface UpdateMilestoneInput {
  name?: string;
  description?: string;
  target_date?: string | null;
  status?: MilestoneStatus;
}

/** 里程碑列表查询参数 */
export interface ListMilestonesParams {
  status?: MilestoneStatus;
  limit?: number;
  offset?: number;
}

/** 里程碑列表响应 */
export interface ListMilestonesResult {
  results: Milestone[];
}

/** 里程碑工作项列表响应 */
export interface ListMilestoneWorkItemsResult {
  results: MilestoneWorkItem[];
}

/** 关联工作项入参 */
export interface LinkWorkItemInput {
  work_item_id: number;
  work_item_type?: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 里程碑域 API */
export const milestoneApi = {
  /** 列出项目下全部里程碑 */
  list: (
    workspaceId: number | string,
    projectId: number | string,
    params?: ListMilestonesParams,
  ) =>
    wrap<ListMilestonesResult>(
      http.get(`/workspaces/${workspaceId}/projects/${projectId}/milestones`, {
        params,
      }),
    ),

  /** 获取单个里程碑详情 */
  get: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
  ) =>
    wrap<Milestone>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}`,
      ),
    ),

  /** 创建里程碑 */
  create: (
    workspaceId: number | string,
    projectId: number | string,
    input: CreateMilestoneInput,
  ) =>
    wrap<Milestone>(
      http.post(`/workspaces/${workspaceId}/projects/${projectId}/milestones`, input),
    ),

  /** 更新里程碑 */
  update: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
    input: UpdateMilestoneInput,
  ) =>
    wrap<Milestone>(
      http.patch(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}`,
        input,
      ),
    ),

  /** 删除里程碑（关联关系一并解除，204 无内容） */
  remove: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}`,
      ),
    ),

  /** 获取里程碑进度聚合 */
  getProgress: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
  ) =>
    wrap<MilestoneProgress>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}/progress`,
      ),
    ),

  /** 列出里程碑关联的工作项 */
  listWorkItems: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
  ) =>
    wrap<ListMilestoneWorkItemsResult>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}/work-items`,
      ),
    ),

  /** 关联工作项到里程碑（204 无内容） */
  linkWorkItem: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
    input: LinkWorkItemInput,
  ) =>
    wrap<void>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}/work-items`,
        input,
      ),
    ),

  /** 从里程碑解除工作项关联（204 无内容） */
  unlinkWorkItem: (
    workspaceId: number | string,
    projectId: number | string,
    milestoneId: number | string,
    workItemId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/milestones/${milestoneId}/work-items/${workItemId}`,
      ),
    ),
};
