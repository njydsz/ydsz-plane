/**
 * 认证域 API — 调用 axios 实例并提供类型化返回 + ApiResponse envelope。
 */
import { http, type ApiResponse } from "../client";
import type { UserBrief } from "@/stores/auth";

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_at: string;
  user: UserBrief;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface RegisterInput {
  email: string;
  password: string;
  display_name: string;
}

export const authApi = {
  login(input: LoginInput): Promise<ApiResponse<TokenPair>> {
    return http.post("/auth/login", input).then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  register(input: RegisterInput): Promise<ApiResponse<TokenPair>> {
    return http.post("/auth/register", input).then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  /** 通过 cookie 中的 refresh_token 换发新的 access/refresh 对 */
  refresh(): Promise<ApiResponse<TokenPair>> {
    return http.post("/auth/refresh", {}).then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  me(): Promise<ApiResponse<UserBrief>> {
    return http.get<UserBrief>("/me").then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  forgotPassword(email: string): Promise<void> {
    return http.post("/auth/forgot-password", { email });
  },

  resetPassword(token: string, newPassword: string): Promise<void> {
    return http.post("/auth/reset-password", { token, new_password: newPassword });
  },
};
