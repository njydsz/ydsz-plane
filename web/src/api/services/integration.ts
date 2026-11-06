/**
 * 集成（Marketplace + 第三方连接器）域 API — 对接后端集成 REST 接口。
 *
 * 路由前缀: /api/v1/workspaces/:workspace_id/integrations
 *
 * 能力：
 *   - 集成实例 CRUD（list / get / create / update / delete）；
 *   - 测试连接 / 立即同步；
 *   - 暂停 / 恢复 / 断开；
 *   - 事件日志查询；
 *   - 可用 provider 元数据（Marketplace）。
 *
 * 注意：凭证（credentials）为加密存储，服务端永不下发，
 * 因此 Integration 类型不含 credentials 字段。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 支持的 provider 标识 */
export type IntegrationProvider = "github" | "gitlab" | "slack" | "sentry" | "jira";

/** 集成实例生命周期状态 */
export type IntegrationStatus = "active" | "paused" | "error" | "pending";

/** 集成事件方向 */
export type IntegrationEventDirection = "inbound" | "outbound";

/** 集成事件处理状态 */
export type IntegrationEventStatus = "success" | "failed" | "ignored";

/** 集成实例（凭证永不下发前端） */
export interface Integration {
  id: number;
  workspace_id: number;
  provider: IntegrationProvider;
  name: string;
  status: IntegrationStatus;
  /** 明文配置 JSON 字符串（repo / channel / project 等） */
  config?: string;
  last_sync_at?: string | null;
  last_error?: string | null;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 集成事件日志 */
export interface IntegrationEvent {
  id: number;
  integration_id: number;
  direction: IntegrationEventDirection;
  event_type: string;
  status: IntegrationEventStatus;
  payload?: string;
  error?: string;
  created_at: string;
}

/** provider 的一个凭证字段元数据 */
export interface ProviderField {
  key: string;
  label: string;
  required: boolean;
  /** true = 密码 / Token 类，前端以密文输入 */
  secret: boolean;
  placeholder?: string;
}

/** 可用 provider 元数据（GET /providers） */
export interface ProviderMeta {
  provider: IntegrationProvider;
  name: string;
  icon: string;
  description: string;
  fields: ProviderField[];
  docs_url: string;
  supports_sync: boolean;
  supports_notify: boolean;
}

/** 创建 / 连接集成入参 */
export interface CreateIntegrationInput {
  provider: IntegrationProvider;
  name: string;
  credentials: Record<string, string>;
  config?: Record<string, string>;
}

/** 更新集成入参 */
export interface UpdateIntegrationInput {
  name?: string;
  credentials?: Record<string, string>;
  config?: Record<string, string>;
}

/** 同步结果摘要 */
export interface SyncResult {
  provider: string;
  items_fetched: number;
  items_imported: number;
  items_skipped: number;
  message?: string;
  fetched_at: string;
}

/** 测试连接结果 */
export interface TestConnectionResult {
  ok: boolean;
  message?: string;
}

/** 列表查询参数 */
export interface ListIntegrationsParams {
  provider?: IntegrationProvider;
  status?: IntegrationStatus;
  limit?: number;
  offset?: number;
}

/** 事件日志查询参数 */
export interface ListIntegrationEventsParams {
  direction?: IntegrationEventDirection;
  status?: IntegrationEventStatus;
  event_type?: string;
  limit?: number;
  offset?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 集成域 API */
export const integrationApi = {
  /* --- Marketplace --- */

  /** 可用 provider 列表 */
  listProviders: (workspaceId: number | string) =>
    wrap<{ items: ProviderMeta[] }>(
      http.get(`/workspaces/${workspaceId}/integrations/providers`),
    ),

  /* --- CRUD --- */

  /** 列出工作空间下全部集成 */
  list: (workspaceId: number | string, params?: ListIntegrationsParams) =>
    wrap<{ items: Integration[]; total: number }>(
      http.get(`/workspaces/${workspaceId}/integrations`, { params }),
    ),

  /** 获取集成详情 */
  get: (workspaceId: number | string, id: number | string) =>
    wrap<Integration>(http.get(`/workspaces/${workspaceId}/integrations/${id}`)),

  /** 创建 / 连接集成 */
  create: (workspaceId: number | string, input: CreateIntegrationInput) =>
    wrap<Integration>(
      http.post(`/workspaces/${workspaceId}/integrations`, input),
    ),

  /** 更新集成（名称 / 凭证 / 配置） */
  update: (
    workspaceId: number | string,
    id: number | string,
    input: UpdateIntegrationInput,
  ) =>
    wrap<Integration>(
      http.patch(`/workspaces/${workspaceId}/integrations/${id}`, input),
    ),

  /** 删除集成 */
  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/integrations/${id}`)),

  /* --- 操作 --- */

  /** 测试连接 */
  test: (workspaceId: number | string, id: number | string) =>
    wrap<TestConnectionResult>(
      http.post(`/workspaces/${workspaceId}/integrations/${id}/test`),
    ),

  /** 立即同步 */
  sync: (workspaceId: number | string, id: number | string) =>
    wrap<SyncResult>(
      http.post(`/workspaces/${workspaceId}/integrations/${id}/sync`),
    ),

  /** 暂停 */
  pause: (workspaceId: number | string, id: number | string) =>
    wrap<Integration>(
      http.post(`/workspaces/${workspaceId}/integrations/${id}/pause`),
    ),

  /** 恢复 */
  resume: (workspaceId: number | string, id: number | string) =>
    wrap<Integration>(
      http.post(`/workspaces/${workspaceId}/integrations/${id}/resume`),
    ),

  /** 断开（删除凭证，状态置为 pending） */
  disconnect: (workspaceId: number | string, id: number | string) =>
    wrap<Integration>(
      http.post(`/workspaces/${workspaceId}/integrations/${id}/disconnect`),
    ),

  /* --- 事件日志 --- */

  /** 事件日志 */
  listEvents: (
    workspaceId: number | string,
    id: number | string,
    params?: ListIntegrationEventsParams,
  ) =>
    wrap<{ items: IntegrationEvent[]; total: number }>(
      http.get(`/workspaces/${workspaceId}/integrations/${id}/events`, {
        params,
      }),
    ),
};
