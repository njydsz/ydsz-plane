<script setup lang="ts">
/**
 * 重置密码页 — 通过邮件中的 token 设置新密码。
 */

import { reactive, ref } from "vue";
import { useRoute } from "vue-router";

import { authApi } from "@/api/services/auth";
import { ApiError } from "@/api/client";

const route = useRoute();

const token = ref(String(route.query.token ?? ""));
const form = reactive({
  newPassword: "",
  confirmPassword: "",
});
const loading = ref(false);
const success = ref(false);
const errorMsg = ref("");
const fieldErrors = ref<Record<string, string>>({});

// 无 token 自动跳转
if (!token.value) {
  errorMsg.value = "无效的重置链接：缺少 token";
}

function validate(): boolean {
  fieldErrors.value = {};

  if (form.newPassword.length < 8) {
    fieldErrors.value.newPassword = "密码至少 8 位";
  }
  if (form.newPassword !== form.confirmPassword) {
    fieldErrors.value.confirmPassword = "两次密码不一致";
  }

  return Object.keys(fieldErrors.value).length === 0;
}

async function onSubmit() {
  errorMsg.value = "";
  fieldErrors.value = {};
  if (!token.value) {
    errorMsg.value = "缺少 token，请从邮件链接打开此页面";
    return;
  }
  if (!validate()) return;

  loading.value = true;
  try {
    await authApi.resetPassword(token.value, form.newPassword);
    success.value = true;
  } catch (e) {
    if (e instanceof ApiError) {
      errorMsg.value = e.message;
      if (e.isValidation && e.details) {
        for (const d of e.details) {
          fieldErrors.value[d.field] = d.reason;
        }
      }
    } else {
      errorMsg.value = "网络异常，请稍后再试";
    }
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
        <h1 class="brand__name">重置密码</h1>
        <p class="brand__slogan">请设置你的新密码</p>
      </div>

      <!-- token 缺失 -->
      <div v-if="!token" class="error-state">
        <p class="error-text">{{ errorMsg }}</p>
        <router-link to="/forgot-password" class="back-link">重新发送重置链接</router-link>
      </div>

      <!-- 重置成功 -->
      <div v-else-if="success" class="success-state">
        <div class="success-icon">✓</div>
        <p class="success-text">密码重置成功</p>
        <router-link to="/login" class="btn-link">立即登录</router-link>
      </div>

      <!-- 重置表单 -->
      <form v-else @submit.prevent="onSubmit">
        <label class="field">
          <span class="field__label">新密码</span>
          <input
            v-model="form.newPassword"
            type="password"
            required
            minlength="8"
            placeholder="至少 8 位"
            autocomplete="new-password"
            :class="{ error: fieldErrors.newPassword }"
          />
          <span v-if="fieldErrors.newPassword" class="field__error">{{ fieldErrors.newPassword }}</span>
        </label>

        <label class="field">
          <span class="field__label">确认新密码</span>
          <input
            v-model="form.confirmPassword"
            type="password"
            required
            minlength="8"
            placeholder="再次输入新密码"
            autocomplete="new-password"
            :class="{ error: fieldErrors.confirmPassword }"
          />
          <span v-if="fieldErrors.confirmPassword" class="field__error">{{ fieldErrors.confirmPassword }}</span>
        </label>

        <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

        <button class="submit" type="submit" :disabled="loading">
          {{ loading ? "重置中…" : "重置密码" }}
        </button>
      </form>
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

.field input.error {
  border-color: var(--danger-500);
}

.field input.error:focus {
  box-shadow: 0 0 0 3px rgba(220, 47, 47, 0.12);
}

.field__error {
  font-size: 12px;
  color: var(--danger-500);
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

.error-state,
.success-state {
  text-align: center;
  padding: 16px 0;
}

.error-text {
  font-size: 14px;
  color: var(--danger-500);
  margin: 0 0 12px;
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
  margin: 0 0 12px;
}

.btn-link {
  display: inline-block;
  height: 40px;
  line-height: 40px;
  padding: 0 24px;
  border-radius: var(--radius-sm);
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
}

.back-link {
  font-size: 13px;
  color: var(--brand-500);
}
</style>
