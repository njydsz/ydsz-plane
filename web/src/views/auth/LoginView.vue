<script setup lang="ts">
/**
 * 登录页 — 邮箱+密码登录 + SSO/OIDC 企业认证入口。
 *
 * S13 SSO 流程：
 *   1. 页面加载已认证后拉取当前工作空间的 SSO Providers
 *   2. 用户点击 "使用 <Provider> 登录" → 浏览器重定向到后端 /sso/:wid/providers/:pid/login
 *   3. 后端发起 state + code_challenge → 302 重定向到 IdP
 *   4. IdP 回调 /api/v1/auth/oidc/callback → 设置 Cookie 302 → /sso/callback
 */

import { onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { ApiError, apiClient } from "@/api/client";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const form = reactive({ email: "", password: "" });
const loading = ref(false);
const errorMsg = ref("");

// S13: SSO Providers
interface SSOProvider {
  id: number;
  name: string;
  protocol: string;
  client_id: string;
}
const ssoProviders = ref<SSOProvider[]>([]);
const ssoLoading = ref(false);

onMounted(async () => {
  // 仅未登录时加载 SSO Providers
  if (auth.isAuthenticated) return;
  ssoLoading.value = true;
  try {
    const wsId = getWorkspaceId();
    if (wsId > 0) {
      const res = await apiClient.get<{ items: SSOProvider[] }>(
        `/api/v1/workspaces/${wsId}/sso/providers`
      );
      ssoProviders.value = res.data?.items ?? [];
    }
  } catch {
    // SSO 未配置时静默忽略
  } finally {
    ssoLoading.value = false;
  }
});

function getWorkspaceId(): number {
  const redirect = route.query.redirect as string;
  if (redirect) {
    const m = redirect.match(/\/workspaces\/(\d+)/);
    if (m) return parseInt(m[1], 10);
  }
  return 0;
}

function onSSOLogin(provider: SSOProvider) {
  if (provider.protocol !== "oidc") return;
  const wsId = getWorkspaceId();
  const redirect = (route.query.redirect as string) || "/";
  window.location.href =
    `/api/v1/auth/sso/${wsId}/providers/${provider.id}/login` +
    `?redirect_to=${encodeURIComponent(redirect)}`;
}

async function onSubmit() {
  errorMsg.value = "";
  loading.value = true;
  try {
    await auth.login(form.email, form.password);
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
    await router.push(redirect);
  } catch (e) {
    errorMsg.value = e instanceof ApiError ? e.message : "网络异常，请稍后再试";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="onSubmit">
      <div class="brand">
        <span class="brand__mark">YD</span>
        <h1 class="brand__name">Ydsz Plane</h1>
        <p class="brand__slogan">面向中国软件团队的开源项目管理平台</p>
      </div>

      <label class="field">
        <span class="field__label">邮箱</span>
        <input v-model.trim="form.email" type="email" required placeholder="you@example.com" autocomplete="email" />
      </label>

      <label class="field">
        <span class="field__label">密码</span>
        <input v-model="form.password" type="password" required minlength="8" placeholder="至少 8 位" autocomplete="current-password" />
      </label>

      <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

      <button class="submit" type="submit" :disabled="loading">
        {{ loading ? "登录中…" : "登录" }}
      </button>

      <div class="footer-links">
        <router-link to="/register">注册账号</router-link>
        <router-link to="/forgot-password">忘记密码？</router-link>
      </div>

      <!-- S13 SSO divider -->
      <div v-if="ssoProviders.length > 0" class="sso-divider">
        <span class="sso-divider__line"></span>
        <span class="sso-divider__text">或使用企业账号登录</span>
        <span class="sso-divider__line"></span>
      </div>

      <!-- S13 SSO Providers -->
      <div v-if="ssoProviders.length > 0" class="sso-providers">
        <button
          v-for="p in ssoProviders"
          :key="p.id"
          class="sso-btn"
          :disabled="ssoLoading"
          @click="onSSOLogin(p)"
        >
          <span class="sso-btn__label">{{ p.name }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2);
}

.login-card {
  width: 360px;
  padding: 32px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  background: var(--surface-1);
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.brand {
  text-align: center;
  margin-bottom: 8px;
}

.brand__mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 700;
}

.brand__name {
  margin: 12px 0 4px;
  font-size: 18px;
  color: var(--text-primary);
}

.brand__slogan {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field__label {
  font-size: 13px;
  color: var(--text-secondary);
}

.field input {
  height: 38px;
  padding: 0 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  font-size: 14px;
}

.field input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.error {
  margin: 0;
  font-size: 13px;
  color: var(--danger-500);
}

.submit {
  height: 40px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.submit:hover:not(:disabled) {
  background: var(--brand-600);
}

.submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.footer-links {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 13px;
}

.footer-links a {
  color: var(--brand-500);
}

/* S13: SSO section */
.sso-divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 4px 0;
}

.sso-divider__line {
  flex: 1;
  height: 1px;
  background: var(--border-subtle);
}

.sso-divider__text {
  font-size: 12px;
  color: var(--text-tertiary);
  white-space: nowrap;
}

.sso-providers {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sso-btn {
  height: 40px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.15s, border-color 0.15s;
}

.sso-btn:hover:not(:disabled) {
  background: var(--surface-2);
  border-color: var(--border-hover);
}

.sso-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
