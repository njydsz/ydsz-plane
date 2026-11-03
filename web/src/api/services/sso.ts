/**
 * SSO Provider API —— 工作空间单点登录集成配置。
 *
 * 后端路由：/api/v1/workspaces/:wsId/sso/providers
 * 权限：读 workspace:read；写 workspace:update。
 */
import { apiClient } from "../client";

/** SSO 协议类型。 */
export type SSOProtocol = "oidc" | "saml";

/** SSO Provider 列表项（不含 secret）。 */
export interface SSOProviderListItem {
  id: number;
  name: string;
  protocol: SSOProtocol;
  issuer_url: string;
  client_id: string;
  auth_url: string;
  scopes: string;
  auto_createUser: boolean;
  default_role: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

/** SSO Provider 详情（不含明文 secret）。 */
export interface SSOProviderDetail {
  id: number;
  name: string;
  protocol: SSOProtocol;
  issuer_url: string;
  client_id: string;
  has_client_secret: boolean;
  redirect_uri: string;
  auth_url: string;
  token_url: string;
  userinfo_url: string;
  jwks_url: string;
  scopes: string;
  auto_create_user: boolean;
  default_role: string;
  attribute_mapping: Record<string, string>;
}

/** 创建/更新 Provider 入参。 */
export interface ProviderModifyInput {
  name: string;
  protocol: SSOProtocol;
  issuer_url: string;
  client_id: string;
  client_secret?: string;
  redirect_uri?: string;
  auth_url?: string;
  token_url?: string;
  userinfo_url?: string;
  jwks_url?: string;
  scopes?: string;
  auto_create_user?: boolean;
  default_role?: string;
  attribute_mapping?: Record<string, string>;
}

export const ssoApi = {
  /** 列出工作空间 SSO Providers。 */
  async list(wsId: number): Promise<SSOProviderListItem[]> {
    const { data } = await apiClient.get<{ items: SSOProviderListItem[] }>(
      `/workspaces/${wsId}/sso/providers`,
    );
    return data.items ?? [];
  },

  /** 获取单个 Provider 详情。 */
  async get(wsId: number, id: number): Promise<SSOProviderDetail> {
    const { data } = await apiClient.get<SSOProviderDetail>(
      `/workspaces/${wsId}/sso/providers/${id}`,
    );
    return data;
  },

  /** 创建 Provider。 */
  async create(wsId: number, input: ProviderModifyInput): Promise<SSOProviderDetail> {
    const { data } = await apiClient.post<SSOProviderDetail>(
      `/workspaces/${wsId}/sso/providers`,
      input,
    );
    return data;
  },

  /** 更新 Provider。 */
  async update(wsId: number, id: number, input: ProviderModifyInput): Promise<SSOProviderDetail> {
    const { data } = await apiClient.patch<SSOProviderDetail>(
      `/workspaces/${wsId}/sso/providers/${id}`,
      input,
    );
    return data;
  },

  /** 删除 Provider。 */
  async remove(wsId: number, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/sso/providers/${id}`);
  },
};
