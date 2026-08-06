<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { ApiError } from "@/api/client";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const form = reactive({ email: "", password: "" });
const loading = ref(false);
const errorMsg = ref("");

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
</style>
