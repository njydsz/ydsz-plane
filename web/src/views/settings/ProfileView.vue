<script setup lang="ts">
/**
 * ProfileView — 个人设置页（资料编辑）。
 *
 * 允许当前登录用户修改：
 *  - 显示名称（display_name）
 *  - 头像 URL（avatar_url，附带预览）
 *  - 时区（timezone）
 *  - 语言偏好（language）
 *
 * 提交按钮在表单未变更时禁用（dirty check）。
 */

import { computed, reactive, ref } from "vue";

import { userApi, type UserProfileInput } from "@/api/services/user";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";

/* ------------------------------------------------------------------ */
/* 时区 / 语言选项                                                     */
/* ------------------------------------------------------------------ */

const TIMEZONE_OPTIONS = [
  { value: "Asia/Shanghai", label: "(UTC+8) 上海 / 北京" },
  { value: "Asia/Tokyo", label: "(UTC+9) 东京" },
  { value: "UTC", label: "(UTC+0) UTC" },
  { value: "America/New_York", label: "(UTC-5/-4) 纽约" },
  { value: "Europe/London", label: "(UTC+0/+1) 伦敦" },
] as const;

const LANGUAGE_OPTIONS = [
  { value: "zh-CN", label: "简体中文" },
  { value: "en-US", label: "English" },
] as const;

/* ------------------------------------------------------------------ */
/* 初始数据 & 表单状态                                                  */
/* ------------------------------------------------------------------ */

const auth = useAuthStore();

interface ProfileForm {
  display_name: string;
  avatar_url: string;
  timezone: string;
  language: string;
}

/** 从 auth store 同步的原始快照，用于 dirty 比较 */
const original = reactive<ProfileForm>({
  display_name: auth.user?.display_name ?? "",
  avatar_url: auth.user?.avatar_url ?? "",
  timezone: (auth.user as Record<string, unknown>)?.timezone as string ?? "Asia/Shanghai",
  language: (auth.user as Record<string, unknown>)?.language as string ?? "zh-CN",
});

const form = reactive<ProfileForm>({ ...original });

const saving = ref(false);
const formError = ref("");

/* ------------------------------------------------------------------ */
/* Derived                                                             */
/* ------------------------------------------------------------------ */

/** 表单是否已变更（用于禁用提交按钮） */
const isDirty = computed(() => {
  return (
    form.display_name !== original.display_name ||
    form.avatar_url !== original.avatar_url ||
    form.timezone !== original.timezone ||
    form.language !== original.language
  );
});

/* ------------------------------------------------------------------ */
/* Actions                                                             */
/* ------------------------------------------------------------------ */

async function submit() {
  if (!form.display_name.trim()) {
    formError.value = "请填写显示名称";
    return;
  }

  // 仅提交有变更的字段
  const payload: UserProfileInput = {};
  if (form.display_name !== original.display_name) {
    payload.display_name = form.display_name.trim();
  }
  if (form.avatar_url !== original.avatar_url) {
    payload.avatar_url = form.avatar_url.trim();
  }
  if (form.timezone !== original.timezone) {
    payload.timezone = form.timezone;
  }
  if (form.language !== original.language) {
    payload.language = form.language;
  }

  saving.value = true;
  formError.value = "";
  try {
    const { data } = await userApi.update(payload);
    // 更新本地 auth store
    auth.user = { ...auth.user!, ...data };
    // 同步 original 快照
    original.display_name = form.display_name;
    original.avatar_url = form.avatar_url;
    original.timezone = form.timezone;
    original.language = form.language;
    toast.success("个人资料已更新");
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "保存失败，请稍后重试";
    formError.value = msg;
    toast.error(msg);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="profile-view">
    <!-- Header -->
    <header class="header">
      <div>
        <h1>个人资料</h1>
        <p class="subtitle">
          管理您的个人信息，包括显示名称、头像、时区和语言偏好。
        </p>
      </div>
    </header>

    <!-- Avatar preview card -->
    <section class="avatar-card">
      <div class="avatar-card__preview">
        <img
          v-if="form.avatar_url"
          :src="form.avatar_url"
          alt="头像预览"
          class="avatar-img"
          @error="(e: Event) => ((e.target as HTMLImageElement).style.display = 'none')"
        />
        <span v-else class="avatar-img avatar-img--placeholder">
          {{ (form.display_name || "? ").charAt(0).toUpperCase() }}
        </span>
      </div>
      <div class="avatar-card__info">
        <span class="avatar-card__name">{{ form.display_name || "未设置名称" }}</span>
        <span class="avatar-card__email">{{ auth.user?.email ?? "" }}</span>
      </div>
    </section>

    <!-- Edit form -->
    <form class="form" @submit.prevent="submit">
      <div v-if="formError" class="form-error">{{ formError }}</div>

      <div class="form-group">
        <label class="form-label">显示名称 *</label>
        <input
          v-model="form.display_name"
          class="form-input"
          placeholder="请输入您的显示名称"
          maxlength="64"
        />
      </div>

      <div class="form-group">
        <label class="form-label">头像 URL</label>
        <input
          v-model="form.avatar_url"
          class="form-input"
          placeholder="https://example.com/avatar.png"
        />
        <span class="form-hint">填写图片链接以设置头像。留空则使用默认头像。</span>
      </div>

      <div class="form-group">
        <label class="form-label">时区</label>
        <select v-model="form.timezone" class="form-select">
          <option v-for="opt in TIMEZONE_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label class="form-label">语言</label>
        <select v-model="form.language" class="form-select">
          <option v-for="opt in LANGUAGE_OPTIONS" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div class="form-actions">
        <button
          type="submit"
          class="btn btn--primary"
          :disabled="!isDirty || saving"
        >
          {{ saving ? "保存中..." : "保存修改" }}
        </button>
      </div>
    </form>

    <!-- Tip -->
    <div class="tip">
      <strong>提示：</strong>
      修改后的信息将在您下次刷新页面或重新登录后在所有工作空间生效。
    </div>
  </div>
</template>

<style scoped>
.profile-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 32px 24px;
}

.header {
  margin-bottom: 24px;
}

.header h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
}

.subtitle {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 500px;
}

/* Avatar card */
.avatar-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  margin-bottom: 24px;
}

.avatar-card__preview {
  flex-shrink: 0;
}

.avatar-img {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border-subtle);
}

.avatar-img--placeholder {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 700;
  font-size: 20px;
}

.avatar-card__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.avatar-card__name {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.avatar-card__email {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Form */
.form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input,
.form-select {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--text-primary);
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  outline: none;
}

.form-input:focus,
.form-select:focus {
  border-color: var(--brand-500);
}

.form-hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.form-error {
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: var(--danger-50);
  color: var(--danger-600);
  font-size: 12px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn {
  padding: 8px 20px;
  font-size: 13px;
  font-family: inherit;
  font-weight: 500;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn--primary {
  background: var(--brand-500);
  color: #fff;
  border-color: var(--brand-500);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
  border-color: var(--brand-600);
}

.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Tip */
.tip {
  margin-top: 20px;
  padding: 12px 16px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}
</style>
