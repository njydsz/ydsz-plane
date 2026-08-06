import { defineStore } from "pinia";

import { http } from "@/api/client";

export interface UserBrief {
  id: number;
  email: string;
  display_name: string;
  avatar_url: string;
}

interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_at: string;
  user: UserBrief;
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
      const { data } = await http.post<TokenPair>("/auth/login", { email, password });
      this.user = data.user;
      this.loaded = true;
    },
    async fetchMe() {
      try {
        const { data } = await http.get<UserBrief>("/me");
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
