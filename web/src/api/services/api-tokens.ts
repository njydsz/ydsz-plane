/**
 * 个人 API Token 域 API — axios 调用 + 类型化返回。
 *
 * 令牌是用户级资源（与工作空间无关），用于脚本/集成访问；
 * 原始 token 仅在创建响应中出现一次，之后只能吊销。
 */
import { http } from "../client";

/** 令牌视图模型（不含明文） */
export interface ApiToken {
  id: number;
  name: string;
  scopes: string[];
  last_used_at?: string;
  expires_at?: string;
  created_at: string;
}

/** 创建响应：额外携带一次性明文 token */
export interface ApiTokenCreated extends ApiToken {
  token: string;
}

/** scope 常量（与后端 internal/application/apitoken 保持一致） */
export const API_TOKEN_SCOPES = [
  { value: "read:workspace", label: "读取空间", desc: "查看空间、成员与项目（默认）" },
  { value: "write:workspace", label: "管理空间", desc: "修改空间设置、管理成员、管理项目" },
  { value: "read:issues", label: "读取需求/任务/缺陷", desc: "查看需求与缺陷" },
  { value: "write:issues", label: "管理需求/任务/缺陷", desc: "创建/修改/删除需求与缺陷" },
  { value: "read:sprints", label: "读取迭代", desc: "查看迭代" },
  { value: "write:sprints", label: "管理迭代", desc: "管理迭代" },
  { value: "read:versions", label: "读取版本", desc: "查看版本" },
  { value: "write:versions", label: "管理版本", desc: "管理版本" },
  { value: "read:audit", label: "读取审计日志", desc: "查看空间审计记录" },
] as const;

/** API Token 权限域（与后端 internal/application/apitoken 的 scope 常量对齐）。 */
export type ApiTokenScope = (typeof API_TOKEN_SCOPES)[number]["value"];

/** 有效期预设（秒） */
export const TOKEN_EXPIRY_OPTIONS = [
  { value: 30 * 24 * 3600, label: "30 天" },
  { value: 90 * 24 * 3600, label: "90 天" },
  { value: 365 * 24 * 3600, label: "365 天" },
  { value: 0, label: "永不过期" },
] as const;

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 个人 API Token API */
export const apiTokenApi = {
  list: () => wrap<ApiToken[]>(http.get("/me/api-tokens")),
  create: (input: { name: string; scopes: string[]; expires_in_seconds?: number }) =>
    wrap<ApiTokenCreated>(http.post("/me/api-tokens", input)),
  revoke: (tokenId: number) => wrap<void>(http.delete(`/me/api-tokens/${tokenId}`)),
};

/** scope 显示名（列表标签用） */
export function scopeLabel(value: string): string {
  const hit = API_TOKEN_SCOPES.find((s) => s.value === value);
  return hit ? hit.label : value;
}
