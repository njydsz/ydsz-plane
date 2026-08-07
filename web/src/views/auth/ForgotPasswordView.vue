<script setup lang="ts">
import { ref } from "vue";

import { authApi } from "@/api/services/auth";
import { ApiError } from "@/api/client";

const email = ref("");
const loading = ref(false);
const sent = ref(false);
const errorMsg = ref("");

async function onSubmit() {
  errorMsg.value = "";
  if (!email.value.trim()) {
    errorMsg.value = "请输入邮箱地址";
    return;
  }

  loading.value = true;
  try {
    await authApi.forgotPassword(email.value.trim());
    sent.value = true;
  } catch (e) {
    errorMsg.value = e instanceof ApiError ? e.message : "网络异常，请稍后再试";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="brand">
        <span class="brand__mark">YD</span>
        <h1 class="brand__name">找回密码</h1>
        <p class="brand__slogan">输入注册邮箱，我们将发送重置链接</p>
      </div>

      <!-- 发送成功 -->
      <div v-if="sent" class="success-state">
        <div class="success-icon">✓</div>
        <p class="success-text">
          如果该邮箱已注册，重置密码链接已发送至 <strong>{{ email }}</strong>
        </p>
        <p class="success-hint">请检查收件箱（含垃圾邮件），链接 15 分钟内有效</p>
        <router-link to="/login" class="back-link">返回登录</router-link>
      </div>

      <!-- 表单 -->
      <form v-else @submit.prevent="onSubmit">
        <label class="field">
          <span class="field__label">注册邮箱</span>
          <input
            v-model.trim="email"
            type="email"
            required
            placeholder="you@example.com"
            autocomplete="email"
          />
        </label>

        <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

        <button class="submit" type="submit" :disabled="loading">
          {{ loading ? "发送中…" : "发送重置链接" }}
        </button>
      </form>

      <p v-if="!sent" class="footer-link">
        想起密码了？<router-link to="/login">返回登录</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2);
}

.auth-card {
  width: 380px;
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
  margin-bottom: 4px;
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
  margin: 10px 0 4px;
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
  width: 100%;
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

.success-state {
  text-align: center;
  padding: 16px 0;
}

.success-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: rgba(15, 194, 123, 0.12);
  color: var(--success-500);
  font-size: 24px;
  margin-bottom: 12px;
}

.success-text {
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.6;
  margin: 0 0 8px;
}

.success-hint {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 0 0 16px;
}

.back-link {
  font-size: 13px;
  color: var(--brand-500);
}

.footer-link {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: center;
}
</style>
