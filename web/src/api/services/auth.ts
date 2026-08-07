/**
 * 认证域 API — 调用 axios 实例并提供类型化返回 + ApiResponse envelope。
 */
import { http, type ApiResponse } from "../client";
import type { UserBrief } from "@/stores/auth";

/** 登录/刷新接口返回的令牌对（与后端 auth.TokenPair 对齐） */
export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_at: string;
  user: UserBrief;
}

/** 登录请求参数 */
export interface LoginInput {
  email: string;
  password: string;
}

/** 注册请求参数 */
export interface RegisterInput {
  email: string;
  password: string;
  display_name: string;
}

/** 认证域 API：登录 / 注册 / 刷新 / 当前用户 / 密码找回与重置 */
export const authApi = {
  /** 使用邮箱密码登录，返回令牌对与用户信息 */
  login(input: LoginInput): Promise<ApiResponse<TokenPair>> {
    return http.post("/auth/login", input).then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  /** 注册新账号，成功后直接返回令牌对完成自动登录 */
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

  /** 获取当前登录用户信息 */
  me(): Promise<ApiResponse<UserBrief>> {
    return http.get<UserBrief>("/me").then((r) => ({
      data: r.data,
      requestId: r.headers["x-request-id"],
    }));
  },

  /** 发起密码重置（向邮箱发送重置链接） */
  forgotPassword(email: string): Promise<void> {
    return http.post("/auth/forgot-password", { email });
  },

  /** 使用重置令牌设置新密码 */
  resetPassword(token: string, newPassword: string): Promise<void> {
    return http.post("/auth/reset-password", { token, new_password: newPassword });
  },
};
