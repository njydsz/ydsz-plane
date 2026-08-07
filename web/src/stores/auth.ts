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

/** 认证 store：user 为当前登录用户，loaded 表示是否已从后端恢复会话 */
export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as UserBrief | null,
    loaded: false,
  }),
  getters: {
    /** 是否已登录（user 非空） */
    isAuthenticated: (s) => s.user !== null,
  },
  actions: {
    /** 邮箱+密码登录，成功后写入当前用户 */
    async login(email: string, password: string) {
      const { data } = await authApi.login({ email, password });
      this.user = data.user;
      this.loaded = true;
    },
    /** 注册新账号，成功后写入当前用户并返回令牌对 */
    async register(input: { email: string; password: string; display_name: string }) {
      const { data } = await authApi.register(input);
      this.user = data.user;
      this.loaded = true;
      return data as TokenPair;
    },
    /** 拉取当前用户信息（页面刷新时恢复会话）；失败则视为未登录 */
    async fetchMe() {
      try {
        const { data } = await authApi.me();
        this.user = data;
      } catch {
        this.user = null;
      } finally {
        this.loaded = true;
      }
    },
    /** 登出：Cookie 由后端过期，前端仅清空状态并跳转登录页 */
    logout() {
      // Cookie 由后端过期；前端清空状态并跳转
      this.user = null;
      window.location.assign("/login");
    },
  },
});
