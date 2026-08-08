/**
 * Intake 收件箱 API — 通道管理、工单审核、公开提交。
 *
 * 管理路由: /api/v1/workspaces/:ws/intake/channels + /issues
 * 公开路由: /api/v1/public/intake/:slug/submit + /track
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type IssueType = "requirement" | "bug" | "task";

/** Intake 收件箱通道实体（对外开放提交入口的配置）。 */
export interface IntakeChannel {
  id: number;
  workspace_id: number;
  project_id?: number | null;
  slug: string;
  name: string;
  description: string;
  is_public: boolean;
  is_active: boolean;
  default_issue_type: IssueType;
  default_priority: number;
  auto_assign_rules: any;
  rate_limit_per_min: number;
  require_captcha: boolean;
  custom_fields: any;
  branding: any;
  notify_on_submit: boolean;
  notify_users: number[];
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** Intake 工单状态：待审核 / 已审核 / 已转换 / 已拒绝 / 已归档。 */
export type IntakeStatus = "pending" | "reviewed" | "converted" | "rejected" | "archived";

/** Intake 工单实体（外部用户提交的问题追踪）。 */
export interface IntakeIssue {
  id: number;
  workspace_id: number;
  channel_id: number;
  tracking_id: string;
  title: string;
  description: string;
  submitter_name: string;
  submitter_email: string;
  issue_type: IssueType;
  priority: number;
  status: IntakeStatus;
  custom_fields: any;
  attachment_ids: number[];
  converted_issue_id?: number | null;
  reviewer_id?: number | null;
  review_reason?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
}

/** 创建 Intake 通道入参（slug、配置、限流、通知规则等）。 */
export interface CreateChannelInput {
  slug: string;
  name: string;
  description?: string;
  is_public?: boolean;
  default_issue_type?: IssueType;
  default_priority?: number;
  auto_assign_rules?: any;
  rate_limit_per_min?: number;
  require_captcha?: boolean;
  custom_fields?: any;
  branding?: any;
  notify_on_submit?: boolean;
  notify_users?: number[];
  project_id?: number;
}

/** 更新 Intake 通道入参（仅包含可变字段）。 */
export interface UpdateChannelInput {
  name?: string;
  description?: string;
  is_public?: boolean;
  is_active?: boolean;
  auto_assign_rules?: any;
}

/** 通道列表查询参数（分页 + 项目过滤）。 */
export interface ListChannelsParams {
  project_id?: number;
  limit?: number;
  offset?: number;
}

/** Intake 工单列表查询参数（分页 + 通道 / 状态过滤）。 */
export interface ListIssuesParams {
  channel_id?: number;
  status?: IntakeStatus;
  limit?: number;
  offset?: number;
}

/** 工单审核入参（approve / reject / archive + 目标类型与原因）。 */
export interface ReviewInput {
  action: "approve" | "reject" | "archive";
  target_issue_type?: string;
  target_project_id?: number;
  reason?: string;
}

/** 工单转换为正式工作项的入参（目标项目 + 工作项类型）。 */
export interface ConvertInput {
  target_project_id: number;
  target_issue_type: string;
}

/** 转换结果 — 新建的工作项 ID 与标识符。 */
export interface ConvertResult {
  converted_issue_id: number;
  identifier: string;
  type: string;
  project_id: number;
}

/** 公开提交入参（标题、描述、提交者信息、自定义字段）。 */
export interface SubmitInput {
  title: string;
  description?: string;
  submitter_name: string;
  submitter_email: string;
  issue_type?: string;
  custom_fields?: any;
  attachment_ids?: number[];
}

/** 公开提交结果 — tracking_id + 状态 + 提交时间 + 提示信息。 */
export interface SubmitResult {
  tracking_id: string;
  status: string;
  submitted_at: string;
  message: string;
}

/** 工单追踪结果（状态 + 提交时间 + 审核时间 + 转换项 + 优先级/描述/原因）。 */
export interface TrackResult {
  tracking_id: string;
  status: string;
  title: string;
  description?: string;
  status_text?: string;
  status_reason?: string;
  priority?: number;
  issue_type?: string;
  submitted_at: string;
  reviewed_at?: string;
  converted_issue_id?: number;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** Intake 收件箱域 API — 通道管理、工单审核、公开提交、追踪查询。 */
export const intakeApi = {
  // --- Channel CRUD (admin) ---
  listChannels: (wsId: number, params?: ListChannelsParams) =>
    wrap<{ items: IntakeChannel[]; total: number; limit: number; offset: number }>(
      http.get(`/workspaces/${wsId}/intake/channels`, { params }),
    ),
  getChannel: (wsId: number, channelId: number) =>
    wrap<IntakeChannel>(http.get(`/workspaces/${wsId}/intake/channels/${channelId}`)),
  createChannel: (wsId: number, input: CreateChannelInput) =>
    wrap<IntakeChannel>(http.post(`/workspaces/${wsId}/intake/channels`, input)),
  updateChannel: (wsId: number, channelId: number, input: UpdateChannelInput) =>
    wrap<IntakeChannel>(http.patch(`/workspaces/${wsId}/intake/channels/${channelId}`, input)),
  deleteChannel: (wsId: number, channelId: number) =>
    wrap<{ ok: true }>(http.delete(`/workspaces/${wsId}/intake/channels/${channelId}`)),

  // --- Issue review (admin) ---
  listIssues: (wsId: number, params?: ListIssuesParams) =>
    wrap<{ items: IntakeIssue[]; total: number; limit: number; offset: number }>(
      http.get(`/workspaces/${wsId}/intake/issues`, { params }),
    ),
  getIssue: (wsId: number, issueId: number) =>
    wrap<IntakeIssue>(http.get(`/workspaces/${wsId}/intake/issues/${issueId}`)),
  reviewIssue: (wsId: number, issueId: number, input: ReviewInput) =>
    wrap<IntakeIssue>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/review`, input)),
  convertIssue: (wsId: number, issueId: number, input: ConvertInput) =>
    wrap<ConvertResult>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/convert`, input)),

  // --- Public routes ---
  getPublicChannel: (wsId: number, slug: string) =>
    wrap<{
      id: number; slug: string; name: string; description: string;
      default_issue_type: string; require_captcha: boolean;
      custom_fields: any; branding: any;
    }>(http.get(`/public/intake/${wsId}/${slug}`)),

  submitIssue: (wsId: number, slug: string, input: SubmitInput) =>
    wrap<SubmitResult>(http.post(`/public/intake/${wsId}/${slug}/submit`, input)),

  trackIssue: (trackingId: string, email: string) =>
    wrap<TrackResult>(http.get("/public/intake/track", {
      params: { tracking_id: trackingId, email },
    })),
};
