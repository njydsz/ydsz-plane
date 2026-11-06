/**
 * OAuth 应用 API —— 工作空间级开发者应用（OAuth App）管理。
 *
 * 后端路由: /api/v1/workspaces/:workspace_id/oauth-apps
 * 权限：读 workspace:read；写 workspace:update。
 *
 * 注意：
 *   - client_secret 仅在「创建」与「轮换」时返回一次，列表/详情不回显；
 *   - redirect_uris 必须为 https:// 或 http://localhost（后端校验）。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 应用状态 */
export type OAuthAppStatus = "active" | "revoked";

/** OAuth 应用（不含明文 secret） */
export interface OAuthApp {
  id: number;
  workspace_id: number;
  name: string;
  description: string;
  /** 应用标识（公开） */
  client_id: string;
  /** 回调地址白名单 */
  redirect_uris: string[];
  /** 申请的权限范围 */
  scopes: string[];
  /** 图标 URL */
  icon_url?: string;
  status: OAuthAppStatus;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建应用入参 */
export interface CreateOAuthAppInput {
  name: string;
  description?: string;
  redirect_uris: string[];
  scopes?: string[];
  icon_url?: string;
}

/** 更新应用入参（字段可选，仅传需要修改的字段） */
export interface UpdateOAuthAppInput {
  name?: string;
  description?: string;
  redirect_uris?: string[];
  scopes?: string[];
  icon_url?: string;
}

/** 创建应用响应（含仅回显一次的 client_secret） */
export interface CreateOAuthAppResult extends OAuthApp {
  client_secret: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** OAuth 应用域 API */
export const oauthAppApi = {
  /** 列出工作空间下全部 OAuth 应用 */
  list: (workspaceId: number | string) =>
    wrap<{ items: OAuthApp[] }>(
      http.get(`/workspaces/${workspaceId}/oauth-apps`),
    ).then((r) => r.items ?? []),

  /** 获取单个应用详情 */
  get: (workspaceId: number | string, id: number | string) =>
    wrap<OAuthApp>(http.get(`/workspaces/${workspaceId}/oauth-apps/${id}`)),

  /** 创建应用（响应含仅显示一次的 client_secret） */
  create: (
    workspaceId: number | string,
    input: CreateOAuthAppInput,
  ) =>
    wrap<CreateOAuthAppResult>(
      http.post(`/workspaces/${workspaceId}/oauth-apps`, input),
    ),

  /** 更新应用 */
  update: (
    workspaceId: number | string,
    id: number | string,
    input: UpdateOAuthAppInput,
  ) =>
    wrap<OAuthApp>(
      http.patch(`/workspaces/${workspaceId}/oauth-apps/${id}`, input),
    ),

  /** 删除应用 */
  remove: (workspaceId: number | string, id: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/oauth-apps/${id}`),
    ),

  /** 轮换 client_secret（响应含仅显示一次的新 secret） */
  rotateSecret: (workspaceId: number | string, id: number | string) =>
    wrap<{ client_secret: string }>(
      http.post(
        `/workspaces/${workspaceId}/oauth-apps/${id}/rotate`,
        {},
      ),
    ).then((r) => r.client_secret),
};
