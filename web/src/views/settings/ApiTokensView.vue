<script setup lang="ts">
/**
 * ApiTokensView — 个人 API Token 管理能力。
 *
 * 能力：
 *  - 列表展示（名称、Scopes、有效期、最后使用时间）
 *  - 创建（弹窗表单：名称 + 权限 + 有效期）
 *  - 吊销（确认对话框）
 *  - 创建后一次性展示 Token（带复制到剪贴板）
 *
 * 注意：API Token 是用户级资源（与工作空间无关），
 * 从 /settings/api-tokens 进入，不走工作空间布局。
 */

import { computed, onMounted, reactive, ref } from "vue";

import {
  apiTokenApi,
  type ApiToken,
  type ApiTokenCreated,
  API_TOKEN_SCOPES,
  TOKEN_EXPIRY_OPTIONS,
  scopeLabel,
} from "@/api/services/api-tokens";
import { AppLoadingState, AppErrorState, AppEmptyState, AppButton, AppModal } from "@/components";

/* ------------------------------------------------------------------ */
/* State                                                              */
/* ------------------------------------------------------------------ */

const loading = ref(true);
const error = ref("");
const tokens = ref<ApiToken[]>([]);

// Create form
const showCreate = ref(false);
const formSaving = ref(false);
const formError = ref("");
const form = reactive({
  name: "",
  scopes: ["read:workspace"],
  expirySeconds: 90 * 24 * 3600,
});

// Created token (one-time display)
const createdToken = ref<ApiTokenCreated | null>(null);
const tokenCopied = ref(false);

// Revoke confirmation
const revokeTarget = ref<ApiToken | null>(null);
const revoking = ref(false);

/* ------------------------------------------------------------------ */
/* Derived                                                            */
/* ------------------------------------------------------------------ */

const scopeLabels = computed(() => {
  return API_TOKEN_SCOPES.map((s) => ({
    value: s.value,
    label: s.label,
    desc: s.desc,
    checked: form.scopes.includes(s.value),
  }));
});

/* ------------------------------------------------------------------ */
/* Actions                                                            */
/* ------------------------------------------------------------------ */

async function load() {
  loading.value = true;
  error.value = "";
  try {
    tokens.value = await apiTokenApi.list();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载 API Token 列表失败";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  form.name = "";
  form.scopes = ["read:workspace"];
  form.expirySeconds = 90 * 24 * 3600;
  formError.value = "";
  tokenCopied.value = false;
  showCreate.value = true;
}

function toggleScope(value: string) {
  const idx = form.scopes.indexOf(value);
  if (idx >= 0) {
    if (form.scopes.length > 1) {
      form.scopes.splice(idx, 1);
    }
  } else {
    form.scopes.push(value);
  }
}

async function submitCreate() {
  if (!form.name.trim()) {
    formError.value = "请填写 Token 名称";
    return;
  }
  if (form.scopes.length === 0) {
    formError.value = "请至少选择一个权限";
    return;
  }
  formSaving.value = true;
  formError.value = "";
  try {
    const result = await apiTokenApi.create({
      name: form.name.trim(),
      scopes: form.scopes,
      expires_in_seconds: form.expirySeconds || undefined,
    });
    createdToken.value = result;
    await load();
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    formSaving.value = false;
  }
}

async function copyToken() {
  if (!createdToken.value) return;
  try {
    await navigator.clipboard.writeText(createdToken.value.token);
    tokenCopied.value = true;
    setTimeout(() => { tokenCopied.value = false; }, 2000);
  } catch {
    // Clipboard API 可能不可用
  }
}

function closeCreated() {
  createdToken.value = null;
  showCreate.value = false;
}

function confirmRevoke(token: ApiToken) {
  revokeTarget.value = token;
}

async function submitRevoke() {
  if (!revokeTarget.value) return;
  revoking.value = true;
  try {
    await apiTokenApi.revoke(revokeTarget.value.id);
    revokeTarget.value = null;
    await load();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "吊销失败";
  } finally {
    revoking.value = false;
  }
}

function fmtDate(s?: string) {
  if (!s) return "—";
  return s.slice(0, 10);
}

onMounted(() => void load());
</script>

<template>
  <div class="api-tokens-view">
    <!-- Header -->
    <header class="header">
      <div>
        <h1>API Tokens</h1>
        <p class="subtitle">
          创建个人访问令牌用于脚本或第三方集成。令牌拥有您账户的全部权限，请妥善保管。
        </p>
      </div>
      <AppButton variant="primary" @click="openCreate">
        + 创建 Token
      </AppButton>
    </header>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <div v-else class="content">
      <!-- Token 列表 -->
      <AppEmptyState
        v-if="tokens.length === 0"
        title="还没有 API Token"
        description="创建 Token 用于脚本自动化或集成第三方工具"
      />

      <div v-else class="token-list">
        <div v-for="tok in tokens" :key="tok.id" class="token-card">
          <div class="token-card__main">
            <div class="token-card__info">
              <span class="token-name">{{ tok.name }}</span>
              <div class="token-scopes">
                <span v-for="sc in tok.scopes" :key="sc" class="scope-badge">{{ scopeLabel(sc) }}</span>
              </div>
            </div>
            <div class="token-card__meta">
              <span class="meta-item">创建于 {{ fmtDate(tok.created_at) }}</span>
              <span class="meta-item">最后使用: {{ fmtDate(tok.last_used_at) }}</span>
              <span class="meta-item">
                {{ tok.expires_at ? `过期于 ${fmtDate(tok.expires_at)}` : "永不过期" }}
              </span>
            </div>
          </div>
          <button class="revoke-btn" @click="confirmRevoke(tok)">吊销</button>
        </div>
      </div>

      <!-- 安全提示 -->
      <div class="security-tip">
        <strong>安全提示：</strong>
        API Token 创建后仅展示一次。请立即复制到安全位置（如密码管理器）。
        如怀疑泄漏，立即吊销并重新生成。
      </div>
    </div>

    <!-- 创建 Token Modal -->
    <AppModal :visible="showCreate" title="创建 API Token" @close="showCreate = false">
      <div class="create-form">
        <!-- Created token one-time display -->
        <div v-if="createdToken" class="created-banner">
          <div class="created-banner__title">Token 已创建 — 请立即复制！</div>
          <div class="created-banner__warn">此 Token 仅显示一次，关闭后无法再次查看。</div>
          <div class="token-display">
            <code>{{ createdToken.token }}</code>
            <button class="copy-btn" @click="copyToken">
              {{ tokenCopied ? "已复制!" : "复制" }}
            </button>
          </div>
          <AppButton variant="primary" block @click="closeCreated">我已保存，关闭</AppButton>
        </div>

        <!-- Create form -->
        <template v-else>
          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="form-group">
            <label class="form-label">名称 *</label>
            <input
              v-model="form.name"
              class="form-input"
              placeholder="如：CI/CD Pipeline"
              maxlength="64"
            />
          </div>

          <div class="form-group">
            <label class="form-label">权限范围 *</label>
            <div class="scopes-grid">
              <label
                v-for="opt in scopeLabels"
                :key="opt.value"
                class="scope-checkbox"
                :class="{ 'scope-checkbox--checked': opt.checked }"
              >
                <input
                  type="checkbox"
                  :checked="opt.checked"
                  @change="toggleScope(opt.value)"
                />
                <div>
                  <strong>{{ opt.label }}</strong>
                  <span>{{ opt.desc }}</span>
                </div>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">有效期</label>
            <select v-model.number="form.expirySeconds" class="form-select">
              <option v-for="opt in TOKEN_EXPIRY_OPTIONS" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>

          <div class="form-actions">
            <AppButton variant="secondary" @click="showCreate = false">取消</AppButton>
            <AppButton variant="primary" :loading="formSaving" @click="submitCreate">
              创建 Token
            </AppButton>
          </div>
        </template>
      </div>
    </AppModal>

    <!-- Revoke confirmation -->
    <AppModal :visible="!!revokeTarget" :title="`吊销 Token`" @close="revokeTarget = null">
      <div class="revoke-confirm">
        <p>确定要吊销 <strong>{{ revokeTarget?.name }}</strong> 吗？</p>
        <p class="warn">此操作不可逆，正在使用此 Token 的集成将立即失效。</p>
        <div class="form-actions">
          <AppButton variant="secondary" :disabled="revoking" @click="revokeTarget = null">取消</AppButton>
          <AppButton variant="danger" :loading="revoking" @click="submitRevoke">确认吊销</AppButton>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<style scoped>
.api-tokens-view {
  max-width: 800px;
  margin: 0 auto;
  padding: 32px 24px;
}

.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
  gap: 16px;
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

/* Token list */
.token-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.token-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  gap: 12px;
}

.token-card__main {
  flex: 1;
  min-width: 0;
}

.token-card__info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 6px;
}

.token-name {
  font-weight: 600;
  font-size: 14px;
}

.token-scopes {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.scope-badge {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 3px;
  background: var(--surface-2);
  color: var(--text-secondary);
}

.token-card__meta {
  display: flex;
  gap: 16px;
  font-size: 11px;
  color: var(--text-tertiary);
}

.meta-item {
  white-space: nowrap;
}

.revoke-btn {
  padding: 4px 12px;
  font-size: 12px;
  color: var(--danger-500);
  background: none;
  border: 1px solid var(--danger-200);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-family: inherit;
  flex-shrink: 0;
}

.revoke-btn:hover {
  background: var(--danger-50);
}

/* Security tip */
.security-tip {
  margin-top: 20px;
  padding: 12px 16px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* Create form */
.create-form {
  padding: 4px 0;
}

.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 6px;
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

.form-error {
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: var(--danger-50);
  color: var(--danger-600);
  font-size: 12px;
  margin-bottom: 12px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
}

/* Scopes */
.scopes-grid {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.scope-checkbox {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.15s;
}

.scope-checkbox:hover {
  border-color: var(--brand-300);
}

.scope-checkbox--checked {
  border-color: var(--brand-500);
  background: var(--brand-50, rgba(59, 130, 246, 0.04));
}

.scope-checkbox input {
  margin-top: 2px;
  accent-color: var(--brand-500);
}

.scope-checkbox div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.scope-checkbox strong {
  font-size: 12px;
  color: var(--text-primary);
}

.scope-checkbox span {
  font-size: 11px;
  color: var(--text-tertiary);
}

/* Created token banner */
.created-banner {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.created-banner__title {
  font-weight: 600;
  font-size: 14px;
  color: var(--success-600);
}

.created-banner__warn {
  font-size: 12px;
  color: var(--warning-600);
}

.token-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.token-display code {
  flex: 1;
  font-family: var(--font-mono, monospace);
  font-size: 11px;
  word-break: break-all;
  color: var(--text-primary);
}

.copy-btn {
  padding: 4px 12px;
  font-size: 12px;
  background: var(--brand-500);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  flex-shrink: 0;
  font-family: inherit;
}

.copy-btn:hover {
  background: var(--brand-600);
}

/* Revoke confirm */
.revoke-confirm {
  padding: 4px 0;
}

.revoke-confirm p {
  margin: 0 0 8px;
  font-size: 13px;
}

.revoke-confirm .warn {
  color: var(--warning-600);
  font-size: 12px;
}
</style>
