<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type InvitationPreview } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const token = computed(() => String(route.params.token));
const preview = ref<InvitationPreview | null>(null);
const loading = ref(true);
const error = ref("");
const accepting = ref(false);

const roleText = computed(() => {
  const m: Record<string, string> = { admin: "管理员", member: "成员", guest: "访客" };
  return preview.value ? (m[preview.value.role] ?? preview.value.role) : "";
});

async function load() {
  try {
    preview.value = await workspaceApi.previewInvitation(token.value);
  } catch (e: any) {
    error.value = e.message ?? "加载邀请失败";
  } finally {
    loading.value = false;
  }
}

async function accept() {
  accepting.value = true;
  error.value = "";
  try {
    const inv = await workspaceApi.acceptInvitation(token.value);
    router.push(`/`);
    // 接受成功后刷新授权状态
  } catch (e: any) {
    error.value = e.message ?? "接受邀请失败";
  } finally {
    accepting.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="invite-page">
    <div v-if="loading" class="card">加载中...</div>
    <div v-else-if="error" class="card card--error">
      <h2>邀请不可用</h2>
      <p>{{ error }}</p>
      <button class="btn btn--primary" @click="router.push('/')">返回工作空间</button>
    </div>
    <div v-else-if="preview" class="card">
      <div class="card__header">
        <h1>加入「{{ preview.workspace_name }}」</h1>
        <p>
          {{ preview.inviter_name }} 邀请你以 <strong>{{ roleText }}</strong> 角色加入该工作空间。
        </p>
      </div>

      <dl class="info">
        <dt>邀请人</dt>
        <dd>{{ preview.inviter_name }}</dd>
        <dt>工作空间</dt>
        <dd>{{ preview.workspace_name }}</dd>
        <dt>受邀角色</dt>
        <dd>{{ roleText }}</dd>
        <dt>过期时间</dt>
        <dd>{{ preview.expires_at.slice(0, 10) }}</dd>
      </dl>

      <div v-if="error" class="form-error">{{ error }}</div>

      <div class="card__actions">
        <button
          v-if="!auth.isAuthenticated"
          class="btn btn--primary"
          @click="router.push({ name: 'login', query: { redirect: $route.fullPath } })"
        >
          登录后加入
        </button>
        <button
          v-else
          class="btn btn--primary"
          :disabled="accepting"
          @click="accept"
        >
          {{ accepting ? "加入中..." : "接受邀请" }}
        </button>
        <button class="btn" @click="router.push('/')">取消</button>
      </div>
    </div>
  </div>
</template>

.invite-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}

.card {
  width: 440px;
  max-width: 95vw;
  padding: 32px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  background: var(--surface-1);
  box-shadow: var(--shadow-card);
}

.card--error {
  text-align: center;
}

.card--error h2 { color: var(--danger-500); }

.card__header h1 {
  font-size: 20px;
  margin: 0 0 8px;
}

.card__header p {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 0 0 20px;
}

.info {
  display: grid;
  grid-template-columns: 80px 1fr;
  gap: 8px 16px;
  margin: 0 0 24px;
  font-size: 13px;
}

.info dt { color: var(--text-tertiary); }
.info dd { color: var(--text-primary); margin: 0; }

.card__actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }

.form-error { color: var(--danger-500); font-size: 13px; margin-bottom: 12px; }
</style>
