<script setup lang="ts">
/**
 * PublicPageView — 公开分享页面的只读视图。
 *
 * 路由：/public/page/:token
 *
 * 特性：
 *   - 无需登录，通过 token 访问文档
 *   - 支持可选密码保护（401 + require_password 时弹出密码输入框）
 *   - 纯只读，无侧边栏、无导航、无编辑控件
 */
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { publicPagesApi, type PublicSharePage } from "@/api/services/pages";
import { AppEmptyState, AppErrorState, AppLoadingState } from "@/components";

const route = useRoute();

const loading = ref(true);
const error = ref("");
const requirePassword = ref(false);
const passwordInput = ref("");
const submittingPassword = ref(false);

const page = ref<PublicSharePage | null>(null);

const token = ref(String(route.params.token));

async function load(password?: string) {
  loading.value = true;
  error.value = "";
  try {
    const result = await publicPagesApi.getShared(token.value, password);
    page.value = result;
    requirePassword.value = false;
  } catch (e: unknown) {
    const err = e as { response?: { data?: { require_password?: boolean }; status?: number } };
    if (err?.response?.data?.require_password || err?.response?.status === 401) {
      requirePassword.value = true;
      error.value = err?.response?.data?.require_password
        ? "该分享链接需要访问密码"
        : "访问密码错误或链接已失效";
    } else if (err?.response?.status === 404) {
      error.value = "分享链接不存在或已被删除";
    } else if (err?.response?.status === 403) {
      error.value = "分享链接已过期或已被吊销";
    } else {
      error.value = e instanceof Error ? e.message : "加载失败";
    }
  } finally {
    loading.value = false;
  }
}

async function submitPassword() {
  if (!passwordInput.value.trim()) return;
  submittingPassword.value = true;
  try {
    await load(passwordInput.value.trim());
  } finally {
    submittingPassword.value = false;
  }
}

onMounted(() => load());
</script>

<template>
  <div class="public-page">
    <AppLoadingState v-if="loading" message="加载文档中..." />

    <AppErrorState
      v-else-if="error && !requirePassword"
      :message="error"
      @retry="load"
    />

    <!-- 密码输入 -->
    <div v-else-if="requirePassword" class="public-page__auth">
      <div class="public-page__auth-card">
        <h2>🔒 受密码保护的文档</h2>
        <p class="text-muted">请输入访问密码查看此文档。</p>
        <div class="public-page__auth-form">
          <input
            v-model="passwordInput"
            type="password"
            class="form-input"
            placeholder="访问密码"
            @keydown.enter="submitPassword"
          />
          <button
            class="btn btn--primary"
            :disabled="submittingPassword || !passwordInput.trim()"
            @click="submitPassword"
          >
            {{ submittingPassword ? '验证中...' : '查看文档' }}
          </button>
        </div>
        <p v-if="error" class="public-page__auth-error">{{ error }}</p>
      </div>
    </div>

    <!-- 文档内容 -->
    <article v-else-if="page" class="public-page__content">
      <header class="public-page__header">
        <h1>{{ page.name }}</h1>
        <div class="public-page__meta">
          <span>更新时间：{{ new Date(page.updated_at).toLocaleString() }}</span>
        </div>
      </header>
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div
        v-if="page.description_html"
        class="public-page__html"
        v-html="page.description_html"
      />
      <AppEmptyState
        v-else
        icon="📄"
        title="文档暂无内容"
        description="该文档尚未添加任何内容。"
      />
    </article>

    <AppEmptyState
      v-else
      icon="❓"
      title="文档未找到"
      description="该分享链接无效或已被删除。"
    />
  </div>
</template>

<style scoped>
.public-page {
  min-height: 100vh;
  background: var(--surface-1);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 16px;
}

.public-page__auth {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 60vh;
  width: 100%;
}

.public-page__auth-card {
  background: var(--surface-2);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 32px;
  width: 100%;
  max-width: 380px;
  text-align: center;
}

.public-page__auth-card h2 {
  font-size: 18px;
  margin: 0 0 8px;
  color: var(--text-primary);
}

.public-page__auth-form {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}

.public-page__auth-form .form-input {
  flex: 1;
  padding: 8px 12px;
  font-size: 14px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
}

.public-page__auth-form .form-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}

.public-page__auth-error {
  color: var(--danger-500);
  font-size: 12px;
  margin: 8px 0 0;
}

.public-page__content {
  width: 100%;
  max-width: 800px;
  padding: 40px 0;
}

.public-page__header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 16px;
  margin-bottom: 24px;
}

.public-page__header h1 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.public-page__meta {
  font-size: 12px;
  color: var(--text-tertiary);
}

.public-page__html {
  font-size: 15px;
  line-height: 1.7;
  color: var(--text-primary);
  word-wrap: break-word;
}

.public-page__html :deep(h1),
.public-page__html :deep(h2),
.public-page__html :deep(h3) {
  margin-top: 24px;
  margin-bottom: 12px;
  line-height: 1.3;
}

.public-page__html :deep(p) {
  margin: 0 0 12px;
}

.public-page__html :deep(ul),
.public-page__html :deep(ol) {
  padding-left: 24px;
  margin: 0 0 12px;
}

.public-page__html :deep(li) {
  margin-bottom: 4px;
}

.public-page__html :deep(code) {
  background: var(--surface-3);
  padding: 1px 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 13px;
}

.public-page__html :deep(pre) {
  background: var(--surface-3);
  padding: 12px;
  border-radius: var(--radius-sm);
  overflow-x: auto;
  margin: 0 0 12px;
}

.public-page__html :deep(blockquote) {
  border-left: 3px solid var(--border-default);
  padding-left: 12px;
  margin: 0 0 12px;
  color: var(--text-secondary);
}

.text-muted {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0;
}

.btn--primary {
  padding: 8px 16px;
  background: var(--brand-500);
  color: var(--text-on-brand);
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  white-space: nowrap;
}

.btn--primary:hover {
  background: var(--brand-600);
}

.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
