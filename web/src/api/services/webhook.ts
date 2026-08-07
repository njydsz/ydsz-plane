/**
 * Webhook 域 API — 对接后端 Webhook 管理 REST 接口。
 *
 * 路由前缀: /api/v1/workspaces/:workspace_id/webhooks
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** Webhook 事件类型（对齐后端 webhook.Event 常量） */
export type WebhookEvent =
  | "issue.created"
  | "issue.updated"
  | "issue.deleted"
  | "issue.status_changed"
  | "issue.commented"
  | "issue.assigned"
  | "sprint.started"
  | "sprint.completed"
  | "version.released"
  | "version.created"
  | "project.created"
  | "member.joined"
  | "member.removed"
  | "comment.created"
  | "attachment.uploaded"
  | "intake.submitted"
  | "automation.triggered"
  | "webhook.tested";

/** Webhook 订阅状态 */
export type WebhookStatus = "active" | "paused" | "unhealthy";

/** Webhook 订阅 */
export interface Webhook {
  id: number;
  workspace_id: number;
  project_id?: number | null;
  name: string;
  url: string;
  events: WebhookEvent[];
  is_active: boolean;
  status: WebhookStatus;
  last_triggered_at?: string | null;
  last_status?: string | null;
  last_error?: string | null;
  failure_count: number;
  description?: string | null;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建 Webhook 入参 */
export interface CreateWebhookInput {
  name: string;
  url: string;
  events: WebhookEvent[];
  project_id?: number | null;
  description?: string;
  is_active?: boolean;
}

/** 更新 Webhook 入参（全字段可选） */
export interface UpdateWebhookInput {
  name?: string;
  url?: string;
  events?: WebhookEvent[];
  project_id?: number | null;
  description?: string;
  is_active?: boolean;
}

/** Webhook 创建响应（含只回显一次的 secret） */
export interface WebhookCreated {
  webhook: Webhook;
  secret: string;
}

/** Webhook 投递日志条目 */
export interface WebhookLog {
  id: number;
  webhook_id: number;
  event: string;
  status_code?: number | null;
  success: boolean;
  attempt: number;
  error_message?: string | null;
  response_body?: string | null;
  duration_ms?: number;
  created_at: string;
}

/** 列表查询参数 */
export interface ListWebhooksParams {
  project_id?: number;
  status?: WebhookStatus;
  limit?: number;
  offset?: number;
}

/** Webhook 日志查询参数 */
export interface ListWebhookLogsParams {
  webhook_id: number;
  success?: boolean;
  limit?: number;
  offset?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

export const webhookApi = {
  // --- CRUD ---
  list: (wsId: number, params?: ListWebhooksParams) =>
    wrap<{ items: Webhook[]; total: number }>(
      http.get(`/workspaces/${wsId}/webhooks`, { params }),
    ),
  get: (wsId: number, webhookId: number) =>
    wrap<Webhook>(http.get(`/workspaces/${wsId}/webhooks/${webhookId}`)),
  create: (wsId: number, input: CreateWebhookInput) =>
    wrap<WebhookCreated>(http.post(`/workspaces/${wsId}/webhooks`, input)),
  update: (wsId: number, webhookId: number, input: UpdateWebhookInput) =>
    wrap<Webhook>(http.patch(`/workspaces/${wsId}/webhooks/${webhookId}`, input)),
  delete: (wsId: number, webhookId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/webhooks/${webhookId}`)),

  // --- 操作 ---
  pause: (wsId: number, webhookId: number) =>
    wrap<Webhook>(http.post(`/workspaces/${wsId}/webhooks/${webhookId}/pause`)),
  resume: (wsId: number, webhookId: number) =>
    wrap<Webhook>(http.post(`/workspaces/${wsId}/webhooks/${webhookId}/resume`)),
  test: (wsId: number, webhookId: number) =>
    wrap<{ success: boolean; status_code?: number; error?: string }>(
      http.post(`/workspaces/${wsId}/webhooks/${webhookId}/test`),
    ),

  // --- 日志 ---
  listLogs: (wsId: number, params: ListWebhookLogsParams) =>
    wrap<{ items: WebhookLog[]; total: number }>(
      http.get(`/workspaces/${wsId}/webhooks/${params.webhook_id}/logs`, {
        params: {
          success: params.success,
          limit: params.limit,
          offset: params.offset,
        },
      }),
    ),
};
