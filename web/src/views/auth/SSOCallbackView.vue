<script setup lang="ts">
/**
 * SSO 回调页 — 验证 SSO 登录是否成功并跳转。
 *
 * 后端 /api/v1/auth/oidc/callback 验证 IdP 响应后，
 * 设置 HTTP-only Cookie，然后 302 重定向到此页面。
 * 本页面调用 fetchMe() 验证 Cookie，成功则跳转工作空间。
 */
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const auth = useAuthStore();
const error = ref("");

onMounted(async () => {
  try {
    // 调用 fetchMe 验证后端设置的 Cookie 是否生效
    await auth.fetchMe();

    if (auth.isAuthenticated) {
      // 登录成功 → 跳转工作空间列表
      router.push("/");
    } else {
      error.value = "SSO 登录验证失败，请重试";
    }
  } catch {
    error.value = "SSO 登录验证失败，请重试";
  }
});
</script>

<template>
  <div class="sso-callback">
    <div class="sso-callback__card">
      <div v-if="!error" class="sso-callback__loading">
        <div class="spinner"></div>
        <p>SSO 登录中…</p>
      </div>
      <div v-else class="sso-callback__error">
        <p class="error-title">登录失败</p>
        <p class="error-msg">{{ error }}</p>
        <button class="retry-btn" @click="router.push('/login')">
          返回登录页
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sso-callback {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2);
}

.sso-callback__card {
  width: 360px;
  padding: 32px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  background: var(--surface-1);
  text-align: center;
}

.sso-callback__loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--danger-500);
  margin: 0 0 8px;
}

.error-msg {
  font-size: 13px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}

.retry-btn {
  height: 36px;
  padding: 0 20px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-size: 13px;
  cursor: pointer;
}
</style>
