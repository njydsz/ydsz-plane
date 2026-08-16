/**
 * 工作项域 API — 对接后端 Issue 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type IssueType = "epic" | "requirement" | "task" | "defect";
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

/**
 * 依赖类型（后端校验 oneof=FS SS FF SF，必须大写）。
 * FS = Finish→Start（完成→开始），SS = Start→Start，
 * FF = Finish→Finish，SF = Start→Finish。
 */
export type DependencyType = "FS" | "SS" | "FF" | "SF";

/** 依赖类型中文标签 */
export const DEPENDENCY_TYPE_LABELS: Record<DependencyType, string> = {
  FS: "完成→开始 (FS)",
  SS: "开始→开始 (SS)",
  FF: "完成→完成 (FF)",
  SF: "开始→完成 (SF)",
};

/** 工作项依赖关系（前置/后继 + 滞后天数） */
export interface IssueDependency {
  id: number;
  workspace_id: number;
  project_id: number;
  predecessor_id: number;
  successor_id: number;
  dependency_type: DependencyType;
  lag_days: number;
  created_by: number;
  created_at: string;
}

/** 表情反应聚合（单种表情的计数 + 当前用户是否已反应） */
export interface ReactionSummary {
  reaction_type: string;
  count: number;
  reacted: boolean;
}

/** 投票聚合统计 */
export interface VoteSummary {
  upvotes: number;
  downvotes: number;
  score: number;
  voted?: number | null; // 1=赞成 -1=反对 null=未投
}

/** 创建工作项入参 */
export interface CreateIssueInput {
  type: IssueType;
  name: string;
  description_html?: string;
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
  root_cause_category?: string;
  verifier_id?: number;
}
// 注意：创建时state_id不需要传，后端会默认使用项目的初始状态

/** 更新工作项入参（可选字段 + 乐观锁 version） */
export interface UpdateIssueInput {
  name?: string;
  description_html?: string;
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
  verifier_id?: number;
  reproduce_steps?: {
    steps?: string;
    expected?: string;
    actual?: string;
  };
}
// 注意：state_id不允许通过更新接口修改，所有状态变更必须调用transition接口

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
  found_version_id?: number;
  fix_version_id?: number;
}

/** 单条列映射：CSV 表头 -> Plane 字段名 */
export interface ImportColumnMapping {
  column_name: string; // 原始 CSV 表头（如 "名称"、"优先级"）
  field: string;       // 目标工作项字段（name / priority / external_id / ...）
}

/** 批量导入结果 */
export interface ImportResult {
  total: number;
  succeeded: number;
  created: number;
  updated: number;
  skipped: number;
  failed: number;
  errors: ImportError[];
}

/** 单行导入错误 */
export interface ImportError {
  row: number;
  field: string;
  message: string;
}

/* ------------------------------------------------------------------ */
/* 导入字段白名单（后端同俗，前端用于下拉选项）                            */
/* ------------------------------------------------------------------ */

/** 字段 key -> 中文标签 */
export const IMPORT_FIELD_LABELS: Record<string, string> = {
  name: "名称 *",
  description: "描述",
  priority: "优先级 (urgent/high/medium/low/none)",
  severity: "严重级别 (1-5)",
  found_phase: "发现阶段",
  root_cause_category: "根因分类",
  category: "分类",
  point: "点数",
  state_name: "状态名 (如: 进行中)",
  module_names: "模块名 (逗号分隔)",
  label_names: "标签名 (逗号分隔)",
  assignee_emails: "指派人口逗号分隔)",
  external_id: "外部 ID (用于增量同步)",
  source: "来源",
  found_version: "发现版本 (名称)",
  fix_version: "修复版本 (名称)",
  parent_identifier: "父工作项编号 (如: YD-123)",
};

/** 字段下拉选项数组 */
export const IMPORT_FIELD_OPTIONS = Object.entries(IMPORT_FIELD_LABELS).map(
  ([id, label]) => ({ id, label }),
);

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
  listIssues: (wsId: number, projectId: number, params?: ListIssuesParams, fields?: string[]) =>
    wrap<{ results: Issue[]; total: number; limit: number; offset: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues`, {
        params: fields && fields.length > 0 ? { ...params, fields: fields.join(",") } : params,
      }),
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

  // --- 看板排序（带乐观锁 version，冲突返回 409） ---
  reorder: (wsId: number, projectId: number, issueId: number, prevSortOrder?: number | null, nextSortOrder?: number | null, version?: number) =>
    wrap<Issue>(http.patch(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reorder`, {
      prev_sort_order: prevSortOrder ?? null,
      next_sort_order: nextSortOrder ?? null,
      version: version ?? null,
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

  // --- 导出 CSV / xlsx ---
  exportUrl: (wsId: number, projectId: number, params?: ListIssuesParams, format?: string) => {
    const qs = new URLSearchParams();
    if (params?.type) qs.set("type", params.type);
    if (params?.state_id) qs.set("state_id", String(params.state_id));
    if (params?.search) qs.set("search", params.search);
    if (format) qs.set("format", format);
    const q = qs.toString();
    return `/api/v1/workspaces/${wsId}/projects/${projectId}/issues/export${q ? "?" + q : ""}`;
  },

  // --- 导入 CSV / XLSX ---
  /**
   * 导入工作项。
   * @param mappings 字段映射数组（可选；为空则按 header 自动识别）
   * @param incremental 增量导入（按 external_id 更新已有项）
   */
  importIssues: (
    wsId: number,
    projectId: number,
    file: File,
    mappings?: ImportColumnMapping[],
    incremental?: boolean,
  ) => {
    const form = new FormData();
    form.append("file", file);
    if (mappings && mappings.length > 0) {
      form.append("mappings", JSON.stringify(mappings));
    }
    if (incremental) {
      form.append("incremental", "true");
    }
    return wrap<ImportResult>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/import`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      }),
    );
  },

  /** 预览导入文件的表头与前几行（CSV/XLSX 通用，XLSX 由后端解析） */
  previewImport: (wsId: number, projectId: number, file: File) => {
    const form = new FormData();
    form.append("file", file);
    return wrap<{ headers: string[]; preview_rows: string[][] }>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/import/preview`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      }),
    );
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
    input: { predecessor_id: number; successor_id: number; dependency_type: DependencyType; lag_days?: number },
  ) =>
    wrap<IssueDependency>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/dependencies`, input)),
  deleteDependency: (wsId: number, projectId: number, issueId: number, depId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/dependencies/${depId}`)),

  // --- 表情反应 ---
  listReactions: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ results: ReactionSummary[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reactions`),
    ),
  addReaction: (wsId: number, projectId: number, issueId: number, reactionType: string) =>
    wrap<{ id: number; reaction_type: string }>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reactions`, {
        reaction_type: reactionType,
      }),
    ),
  removeReaction: (wsId: number, projectId: number, issueId: number, reactionType: string) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reactions/${encodeURIComponent(reactionType)}`)),

  // --- 投票 ---
  voteSummary: (wsId: number, projectId: number, issueId: number) =>
    wrap<VoteSummary>(http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/vote`)),
  vote: (wsId: number, projectId: number, issueId: number, vote: 1 | -1) =>
    wrap<{ id: number; vote: number }>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/vote`, { vote }),
    ),
  removeVote: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/vote`)),

  // --- 关注（watchers）---
  watch: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/watch`)),
  unwatch: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/watch`)),

  // --- 回收站 ---
  listTrash: (wsId: number, projectId: number) =>
    wrap<TrashItem[]>(http.get(`/workspaces/${wsId}/projects/${projectId}/issues/trash`)),
  restoreIssue: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/restore`)),
  permanentDelete: (wsId: number, projectId: number, issueId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/permanent`)),

  // --- 需求评审工作流 ---
  listReviews: (wsId: number, projectId: number, issueId: number) =>
    wrap<{ results: ReviewRecord[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/reviews`),
    ),
  submitReview: (wsId: number, projectId: number, issueId: number, input: { name?: string; reviewers?: number[] }) =>
    wrap<ReviewRecord>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/review`, input),
    ),
  decideReview: (wsId: number, projectId: number, issueId: number, decision: "approved" | "rejected") =>
    wrap<{ ok: boolean }>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/issues/${issueId}/review/decision`, { decision }),
    ),
};

/** 评审活动记录 */
export interface ReviewRecord {
  id: number;
  workspace_id: number;
  project_id?: number | null;
  name: string;
  review_type: string;
  entity_type: string;
  entity_id?: number | null;
  status: string;
  description?: string;
  due_date?: string | null;
  created_date?: string | null;
  completed_date?: string | null;
  created_by: number;
  created_at: string;
  reviewers?: number[];
}

export interface TrashItem {
  id: number;
  project_id: number;
  sequence_id: number;
  type_code: IssueType;
  name: string;
  state_id: number;
  priority: IssuePriority;
  deleted_at: string;
}
