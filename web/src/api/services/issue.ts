/**
 * 工作项域 API — 对接后端 Issue 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type IssueType = "requirement" | "task" | "defect";
/** 工作项优先级（降序） */
export type IssuePriority = "urgent" | "high" | "medium" | "low" | "none";
/** 状态分组：backlog / started / completed / cancelled */
export type StateGroup = "backlog" | "started" | "completed" | "cancelled";

/** 工作项状态定义（含所属分组与展示色） */
export interface State {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  group: StateGroup;
  color: string;
  sequence: number;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

/** 工作项（需求/任务/缺陷统一模型），与后端 issue.Issue 对齐 */
export interface Issue {
  id: number;
  public_id: string;
  workspace_id: number;
  project_id: number;
  sequence_id: number;
  identifier: string;
  type_code: IssueType;
  parent_id?: number;
  depth: number;
  name: string;
  description_json?: Record<string, unknown>;
  description_html?: string;
  state_id: number;
  state?: State;
  priority: IssuePriority;
  severity?: number;
  found_phase?: string;
  root_cause_category?: string;
  verifier_id?: number;
  environment?: Record<string, unknown>;
  reproduce_steps?: Record<string, unknown>;
  category?: string;
  actual_effort?: number;
  remaining_effort?: number;
  delay_reason?: string;
  source?: string;
  point?: number;
  progress: number;
  start_date?: string;
  target_date?: string;
  completed_at?: string;
  is_draft: boolean;
  version: number;
  sort_order?: number;
  sprint_id?: number;
  found_version_id?: number;
  fix_version_id?: number;
  release_version_id?: number;
  assignees: number[];
  labels: number[];
  modules: number[];
  watchers: number[];
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 工作项活动日志条目（谁在何时改了什么字段） */
export interface IssueActivity {
  id: number;
  workspace_id: number;
  project_id: number;
  issue_id: number;
  verb: string;
  field?: string;
  old_value?: string;
  new_value?: string;
  old_ref?: Record<string, unknown>;
  new_ref?: Record<string, unknown>;
  actor_id?: number;
  actor_email: string;
  actor_name: string;
  created_at: string;
}

/** 工作项工时记录（分钟粒度） */
export interface TimeLog {
  id: number;
  workspace_id: number;
  project_id: number;
  issue_id: number;
  user_id: number;
  spent_date: string;
  duration_minutes: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

/** 工作项关联关系（如关联/被关联） */
export interface IssueRelation {
  id: number;
  workspace_id: number;
  project_id: number;
  source_issue_id: number;
  target_issue_id: number;
  relation_type: string;
  created_by: number;
  created_at: string;
}

/** 评论（富文本 + @提及 + 嵌套回复） */
export interface IssueComment {
  id: number;
  workspace_id: number;
  project_id: number;
  issue_id: number;
  content_json: Record<string, unknown>;
  content_html: string;
  content_stripped: string;
  created_by: number;
  creator_name: string;
  creator_avatar: string;
  mentions: number[];
  parent_id: number | null;
  is_edited: boolean;
  edited_at: string | null;
  created_at: string;
  updated_at: string;
}

/** 创建评论入参 */
export interface CreateCommentInput {
  content_json: string;
  content_html: string;
  content_stripped: string;
  mentions?: number[];
  parent_id?: number | null;
}

/** 更新评论入参 */
export interface UpdateCommentInput {
  content_json: string;
  content_html: string;
  content_stripped: string;
  mentions?: number[];
}

/** 工作项依赖关系（前置/后继 + 滞后天数） */
export interface IssueDependency {
  id: number;
  workspace_id: number;
  project_id: number;
  predecessor_id: number;
  successor_id: number;
  dependency_type: string;
  lag_days: number;
  created_by: number;
  created_at: string;
}

/** 创建工作项入参 */
export interface CreateIssueInput {
  type: IssueType;
  name: string;
  description_html?: string;
  state_id?: number;
  priority?: IssuePriority;
  parent_id?: number;
  severity?: number;
  found_phase?: string;
  reproduce_steps?: Record<string, unknown>;
  environment?: Record<string, unknown>;
  category?: string;
  source?: string;
  assignees?: number[];
  labels?: number[];
  modules?: number[];
  point?: number;
  is_draft?: boolean;
  found_version_id?: number;
  fix_version_id?: number;
}

/** 更新工作项入参（可选字段 + 乐观锁 version） */
export interface UpdateIssueInput {
  name?: string;
  description_html?: string;
  state_id?: number;
  priority?: IssuePriority;
  parent_id?: number;
  severity?: number;
  found_phase?: string;
  root_cause_category?: string;
  category?: string;
  assignees?: number[];
  labels?: number[];
  modules?: number[];
  source?: string;
  version: number;
  found_version_id?: number;
  fix_version_id?: number;
  release_version_id?: number;
}

/** 工作项列表查询参数（过滤/搜索/分页） */
export interface ListIssuesParams {
  state_id?: number;
  group?: StateGroup;
  type?: IssueType;
  priority?: IssuePriority;
  parent_id?: number;
  search?: string;
  sort?: string;
  limit?: number;
  offset?: number;
  assignee_id?: number;
  label_id?: number;
  module_id?: number;
  sprint_id?: number;
  start_date_from?: string;
  target_date_to?: string;
  severity_from?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 工作项域 API：状态 / CRUD / 流转 / 活动 / 工时 / 关联 / 依赖 */
export const issueApi = {
  // --- 状态 ---
  listStates: (wsId: number, projectId: number) =>
    wrap<State[]>(http.get(`/workspaces/${wsId}/projects/${projectId}/states`)),

  // --- 工作项 CRUD ---
  listIssues: (wsId: number, projectId: number, params?: ListIssuesParams) =>
    wrap<{ results: Issue[]; total: number; limit: number; offset: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues`, { params }),
    ),
  getIssue: (wsId: number, projectId: number, issueId: number) =>
    wrap<Issue>(http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}`)),
  createIssue: (wsId: number, projectId: number, input: CreateIssueInput) =>
    wrap<Issue>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues`, input)),
  updateIssue: (wsId: number, projectId: number, issueId: number, input: UpdateIssueInput) =>
    wrap<Issue>(http.patch(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}`, input)),
  deleteIssue: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}`)),

  // --- 状态流转 ---
  transition: (wsId: number, projectId: number, issueId: number, toStateId: number) =>
    wrap<Issue>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/transition`, {
      to_state_id: toStateId,
    })),

  // --- 看板排序 ---
  reorder: (wsId: number, projectId: number, issueId: number, prevSortOrder?: number | null, nextSortOrder?: number | null) =>
    wrap<Issue>(http.patch(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reorder`, {
      prev_sort_order: prevSortOrder ?? null,
      next_sort_order: nextSortOrder ?? null,
    })),

  // --- 批量操作 ---
  batch: (
    wsId: number,
    projectId: number,
    input: {
      issue_ids: number[];
      to_state_id?: number;
      assignee_id?: number;
      priority?: string;
      delete?: boolean;
    },
  ) =>
    wrap<{ succeeded: number; failed: number }>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/batch`, input),
    ),

  // --- 导出 CSV ---
  exportUrl: (wsId: number, projectId: number, params?: ListIssuesParams) => {
    const qs = new URLSearchParams();
    if (params?.type) qs.set("type", params.type);
    if (params?.state_id) qs.set("state_id", String(params.state_id));
    if (params?.search) qs.set("search", params.search);
    const q = qs.toString();
    return `/api/v1/workspaces/${wsId}/projects/${projectId}/issues/export${q ? "?" + q : ""}`;
  },

  // --- 活动日志 ---
  listActivities: (wsId: number, projectId: number, issueId: number, limit = 50, offset = 0) =>
    wrap<{ results: IssueActivity[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/activities`, {
        params: { limit, offset },
      }),
    ),

  // --- 工时 ---
  listTimeLogs: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ results: TimeLog[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/time-logs`),
    ),
  createTimeLog: (
    wsId: number,
    projectId: number,
    issueId: number,
    input: { spent_date: string; duration_minutes: number; description?: string },
  ) =>
    wrap<TimeLog>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/time-logs`, input)),
  updateTimeLog: (
    wsId: number,
    projectId: number,
    issueId: number,
    logId: number,
    input: { spent_date: string; duration_minutes: number; description?: string },
  ) =>
    wrap<TimeLog>(http.patch(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/time-logs/${logId}`, input)),
  deleteTimeLog: (wsId: number, projectId: number, issueId: number, logId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/time-logs/${logId}`)),

  // --- 关联关系 ---
  listRelations: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ results: IssueRelation[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/relations`),
    ),
  createRelation: (
    wsId: number,
    projectId: number,
    issueId: number,
    input: { target_issue_id: number; relation_type: string },
  ) =>
    wrap<IssueRelation>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/relations`, input)),
  deleteRelation: (wsId: number, projectId: number, issueId: number, relationId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/relations/${relationId}`)),

  // --- 评论 ---
  listComments: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ results: IssueComment[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/comments`),
    ),
  createComment: (wsId: number, projectId: number, issueId: number, input: CreateCommentInput) =>
    wrap<IssueComment>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/comments`, input)),
  updateComment: (wsId: number, projectId: number, issueId: number, commentId: number, input: UpdateCommentInput) =>
    wrap<IssueComment>(http.patch(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/comments/${commentId}`, input)),
  deleteComment: (wsId: number, projectId: number, issueId: number, commentId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/comments/${commentId}`)),

  // --- 依赖关系 ---
  listDependencies: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ predecessors: IssueDependency[]; successors: IssueDependency[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/dependencies`),
    ),
  createDependency: (
    wsId: number,
    projectId: number,
    issueId: number,
    input: { predecessor_id: number; successor_id: number; dependency_type: string; lag_days?: number },
  ) =>
    wrap<IssueDependency>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/dependencies`, input)),
  deleteDependency: (wsId: number, projectId: number, issueId: number, depId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/dependencies/${depId}`)),
};
