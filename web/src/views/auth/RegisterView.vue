<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";

import { ApiError } from "@/api/client";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const router = useRouter();

const form = reactive({
  displayName: "",
  email: "",
  password: "",
  confirmPassword: "",
});
const loading = ref(false);
const errorMsg = ref("");
const fieldErrors = ref<Record<string, string>>({});

function validate(): boolean {
  fieldErrors.value = {};

  if (!form.displayName.trim()) {
    fieldErrors.value.displayName = "请输入显示名称";
  }
  if (!form.email.trim()) {
    fieldErrors.value.email = "请输入邮箱";
  }
  if (form.password.length < 8) {
    fieldErrors.value.password = "密码至少 8 位";
  }
  if (form.password !== form.confirmPassword) {
    fieldErrors.value.confirmPassword = "两次密码不一致";
  }

  return Object.keys(fieldErrors.value).length === 0;
}

async function onSubmit() {
  errorMsg.value = "";
  fieldErrors.value = {};
  if (!validate()) return;

  loading.value = true;
  try {
    await auth.register({
      email: form.email.trim(),
      password: form.password,
      display_name: form.displayName.trim(),
    });
    await router.push("/");
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
    <form class="auth-card" @submit.prevent="onSubmit">
      <div class="brand">
        <span class="brand__mark">YD</span>
        <h1 class="brand__name">创建账号</h1>
        <p class="brand__slogan">加入 Ydsz Plane，开始高效协作</p>
      </div>

      <label class="field">
        <span class="field__label">显示名称</span>
        <input
          v-model.trim="form.displayName"
          type="text"
          required
          placeholder="你的名字"
          autocomplete="name"
          :class="{ error: fieldErrors.displayName }"
        />
        <span v-if="fieldErrors.displayName" class="field__error">{{ fieldErrors.displayName }}</span>
      </label>

      <label class="field">
        <span class="field__label">邮箱</span>
        <input
          v-model.trim="form.email"
          type="email"
          required
          placeholder="you@example.com"
          autocomplete="email"
          :class="{ error: fieldErrors.email }"
        />
        <span v-if="fieldErrors.email" class="field__error">{{ fieldErrors.email }}</span>
      </label>

      <label class="field">
        <span class="field__label">密码</span>
        <input
          v-model="form.password"
          type="password"
          required
          minlength="8"
          placeholder="至少 8 位"
          autocomplete="new-password"
          :class="{ error: fieldErrors.password }"
        />
        <span v-if="fieldErrors.password" class="field__error">{{ fieldErrors.password }}</span>
      </label>

      <label class="field">
        <span class="field__label">确认密码</span>
        <input
          v-model="form.confirmPassword"
          type="password"
          required
          minlength="8"
          placeholder="再次输入密码"
          autocomplete="new-password"
          :class="{ error: fieldErrors.confirmPassword }"
        />
        <span v-if="fieldErrors.confirmPassword" class="field__error">{{ fieldErrors.confirmPassword }}</span>
      </label>

      <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

      <button class="submit" type="submit" :disabled="loading">
        {{ loading ? "注册中…" : "注册" }}
      </button>

      <p class="footer-link">
        已有账号？<router-link to="/login">立即登录</router-link>
      </p>
    </form>
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
  gap: 14px;
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

.footer-link {
  margin: 0;
  font-size: 13px;
  color: var(--text-tertiary);
  text-align: center;
}
</style>
