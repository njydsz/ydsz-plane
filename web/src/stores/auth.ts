/**
 * 认证 Pinia store — 管理当前登录用户状态、登录/注册/登出/会话恢复动作。
 */
import { defineStore } from "pinia";

import { authApi } from "@/api/services/auth";
import type { TokenPair } from "@/api/services/auth";

/** 用户简要信息（列表/头像场景使用，不含敏感字段） */
export interface UserBrief {
  id: number;
  email: string;
  display_name: string;
  avatar_url: string;
}

/** 前端持久化 access_token 的 localStorage key */
const TOKEN_KEY = "ydsz_access_token";

/** 认证 store：user 为当前登录用户，loaded 表示是否已从后端恢复会话 */
export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as UserBrief | null,
    loaded: false,
    // 从 localStorage 恢复（页面刷新后保持登录态）
    accessToken: (typeof localStorage !== "undefined" ? localStorage.getItem(TOKEN_KEY) : null) as string | null,
  }),
  getters: {
    /** 是否已登录（user 非空） */
    isAuthenticated: (s) => s.user !== null,
  },
  actions: {
    /** 将令牌对写入状态并持久化到 localStorage */
    setSession(data: TokenPair) {
      this.user = data.user;
      this.accessToken = data.access_token;
      this.loaded = true;
      if (typeof localStorage !== "undefined") {
        localStorage.setItem(TOKEN_KEY, data.access_token);
      }
    },
    /** 邮箱+密码登录，成功后写入当前用户与令牌 */
    async login(email: string, password: string) {
      const { data } = await authApi.login({ email, password });
      this.setSession(data);
    },
    /** 注册新账号，成功后写入当前用户并返回令牌对 */
    async register(input: { email: string; password: string; display_name: string }) {
      const { data } = await authApi.register(input);
      this.setSession(data);
      return data as TokenPair;
    },
    /** 通过 refresh_token 换发新令牌对（响应拦截器 401 单飞刷新时调用） */
    async refresh() {
      const { data } = await authApi.refresh();
      this.setSession(data);
    },
    /** 拉取当前用户信息（页面刷新时恢复会话）；失败则视为未登录 */
    async fetchMe() {
      try {
        const { data } = await authApi.me();
        this.user = data;
      } catch {
        this.user = null;
        this.accessToken = null;
        if (typeof localStorage !== "undefined") {
          localStorage.removeItem(TOKEN_KEY);
        }
      } finally {
        this.loaded = true;
      }
    },
    /** 登出：清除状态与本地令牌，跳转登录页 */
    logout() {
      this.user = null;
      this.accessToken = null;
      this.loaded = true;
      if (typeof localStorage !== "undefined") {
        localStorage.removeItem(TOKEN_KEY);
      }
      window.location.assign("/login");
    },
  },
});
