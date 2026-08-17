/**
 * 收件箱（Intake）域 API — 匿名提报渠道与工单管理。
 *
 * 覆盖：
 *  - 渠道 CRUD（认证）：/workspaces/:ws/intake/channels
 *  - 工单管理（认证）：列表 / 详情 / 接受 / 拒绝 / 归档 / 转正
 *  - 公开端点（免登）：/public/intake/channels/:slug、提交、跟踪
 */
import { http } from "../client";

/** 入口渠道 */
export interface IntakeChannel {
  id: number;
  code?: string;
  name: string;
  slug: string;
  workspace_id: number;
  project_id?: number;
  description?: string;
  is_active: boolean;
  config?: Record<string, unknown>;
  status: string;
  issue_count?: number;
  created_at: string;
  updated_at: string;
}

/** 入口工单 */
export interface IntakeIssue {
  id: number;
  code?: string;
  name: string;
  workspace_id: number;
  project_id?: number;
  channel_id: number;
  channel_name?: string;
  tracking_id: string;
  submitter_name?: string;
  submitter_email?: string;
  description?: string;
  priority: string;
  status: "open" | "accepted" | "rejected" | "archived";
  linked_entity_type?: string;
  linked_entity_id?: number;
  linked_entity_identifier?: string;
  resolved_at?: string;
  resolved_by?: number;
  created_at: string;
  updated_at: string;
}

/** 创建渠道入参 */
export interface CreateChannelInput {
  name: string;
  description?: string;
  slug?: string;
  project_id?: number;
  config?: Record<string, unknown>;
}

/** 更新渠道入参 */
export interface UpdateChannelInput {
  name?: string;
  description?: string;
  slug?: string;
  is_active?: boolean;
  project_id?: number;
}

/** 提交工单入参（公开/认证共用） */
export interface SubmitIssueInput {
  channel_id?: number;
  channel_slug?: string;
  name: string;
  description?: string;
  submitter_name?: string;
  submitter_email: string;
  priority?: string;
}

/** 转正入参 */
export interface PromoteIssueInput {
  type_code?: string;
  severity?: number;
  found_phase?: string;
  project_id?: number;
}

interface Paged<T> {
  results: T[];
  total: number;
}

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 收件箱 API */
export const intakeApi = {
  // ---------- 渠道（认证） ----------
  listChannels: (wsId: number, projectId?: number) =>
    wrap<Paged<IntakeChannel>>(
      http.get(`/workspaces/${wsId}/intake/channels`, {
        params: projectId ? { project_id: projectId } : undefined,
      }),
    ),
  getChannel: (wsId: number, channelId: number) =>
    wrap<IntakeChannel>(http.get(`/workspaces/${wsId}/intake/channels/${channelId}`)),
  createChannel: (wsId: number, input: CreateChannelInput) =>
    wrap<IntakeChannel>(http.post(`/workspaces/${wsId}/intake/channels`, input)),
  updateChannel: (wsId: number, channelId: number, input: UpdateChannelInput) =>
    wrap<IntakeChannel>(http.patch(`/workspaces/${wsId}/intake/channels/${channelId}`, input)),
  removeChannel: (wsId: number, channelId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/intake/channels/${channelId}`)),

  // ---------- 工单（认证） ----------
  listIssues: (wsId: number, params?: { status?: string; channel_id?: number; project_id?: number; limit?: number; offset?: number }) =>
    wrap<Paged<IntakeIssue>>(http.get(`/workspaces/${wsId}/intake/issues`, { params })),
  getIssue: (wsId: number, issueId: number) =>
    wrap<IntakeIssue>(http.get(`/workspaces/${wsId}/intake/issues/${issueId}`)),
  acceptIssue: (wsId: number, issueId: number) =>
    wrap<IntakeIssue>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/accept`)),
  rejectIssue: (wsId: number, issueId: number) =>
    wrap<IntakeIssue>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/reject`)),
  archiveIssue: (wsId: number, issueId: number) =>
    wrap<IntakeIssue>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/archive`)),
  promoteIssue: (wsId: number, issueId: number, input: PromoteIssueInput = {}) =>
    wrap<IntakeIssue>(http.post(`/workspaces/${wsId}/intake/issues/${issueId}/promote`, input)),

  // ---------- 公开（免登） ----------
  publicGetChannel: (slug: string) =>
    wrap<IntakeChannel>(http.get(`/public/intake/channels/${slug}`)),
  publicSubmitIssue: (input: SubmitIssueInput) =>
    wrap<IntakeIssue>(http.post("/public/intake/issues", input)),
  publicTrackIssue: (trackingId: string, email: string) =>
    wrap<IntakeIssue>(http.post("/public/intake/track", { tracking_id: trackingId, submitter_email: email })),
};
