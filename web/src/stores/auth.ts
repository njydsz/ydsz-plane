import { defineStore } from "pinia";

import { authApi } from "@/api/services/auth";
import type { TokenPair } from "@/api/services/auth";

export interface UserBrief {
  id: number;
  email: string;
  display_name: string;
  avatar_url: string;
}

export const useAuthStore = defineStore("auth", {
  state: () => ({
    user: null as UserBrief | null,
    loaded: false,
  }),
  getters: {
    isAuthenticated: (s) => s.user !== null,
  },
  actions: {
    async login(email: string, password: string) {
      const { data } = await authApi.login({ email, password });
      this.user = data.user;
      this.loaded = true;
    },
    async register(input: { email: string; password: string; display_name: string }) {
      const { data } = await authApi.register(input);
      this.user = data.user;
      this.loaded = true;
      return data as TokenPair;
    },
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
    logout() {
      // Cookie 由后端过期；前端清空状态并跳转
      this.user = null;
      window.location.assign("/login");
    },
  },
});
