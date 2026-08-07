<script setup lang="ts">
/**
 * 工作空间设置页 — 成员管理、邀请、角色变更与空间信息编辑。
 * 另含「访问令牌」tab：管理个人 API Token（用户级资源，用于脚本/集成）。
 */

import { computed, onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";

import { ApiError } from "@/api/client";
import {
  apiTokenApi,
  scopeLabel,
  API_TOKEN_SCOPES,
  TOKEN_EXPIRY_OPTIONS,
  type ApiToken,
  type ApiTokenCreated,
} from "@/api/services/api-tokens";
import { workspaceApi, type Invitation, type Member, type Workspace } from "@/api/services/workspace";
import { AppLoadingState, AppErrorState } from "@/components";
import { useAuthStore } from "@/stores/auth";

const route = useRoute();
const auth = useAuthStore();

const wsSlug = computed(() => String(route.params.workspaceSlug));
const activeTab = ref<"info" | "members" | "invitations" | "api-tokens">("info");

const wsId = ref(0); // 拿到 ID 后设置
const ws = ref<Workspace | null>(null);
const members = ref<Member[]>([]);
const invitations = ref<Invitation[]>([]);
const loading = ref(true);
const error = ref("");

// 编辑状态
const editing = ref(false);
const editForm = reactive({
  name: "",
  timezone: "Asia/Shanghai",
  language: "zh-CN",
});
const editSaving = ref(false);
const editError = ref("");
const editSuccess = ref("");

const timezoneOptions = [
  { value: "Asia/Shanghai", label: "中国标准时间 (UTC+8)" },
  { value: "Asia/Tokyo", label: "日本标准时间 (UTC+9)" },
  { value: "America/Los_Angeles", label: "太平洋时间 (UTC-8)" },
  { value: "America/New_York", label: "东部时间 (UTC-5)" },
  { value: "Europe/London", label: "格林威治时间 (UTC+0)" },
];

const languageOptions = [
  { value: "zh-CN", label: "简体中文" },
  { value: "en-US", label: "English (US)" },
];

// 邀请表单
const inviteEmail = ref("");
const inviteRole = ref("member");
const inviteMessage = ref("");
const inviteSending = ref(false);
const inviteError = ref("");
const inviteSuccess = ref("");

// 当前用户角色（计算属性，决定 UI 是否显示管理操作）
const myRole = computed(() => ws.value?.role ?? "");
const canManageMembers = computed(() => ["owner", "admin"].includes(myRole.value));

async function loadAll() {
  loading.value = true;
  error.value = "";
  try {
    // 先根据 slug 拿 ID
    const wsData = await workspaceApi.getBySlug(wsSlug.value);
    ws.value = wsData;
    wsId.value = wsData.id;
    editForm.name = wsData.name;
    editForm.timezone = wsData.timezone;
    editForm.language = wsData.language;

    const [mems, tokens] = await Promise.all([
      workspaceApi.listMembers(wsId.value),
      apiTokenApi.list(),
    ]);
    members.value = mems;
    apiTokens.value = tokens;
    if (canManageMembers.value) {
      invitations.value = await workspaceApi.listInvitations(wsId.value);
    }
  } catch (e: any) {
    error.value = e.message ?? "加载失败";
  } finally {
    loading.value = false;
  }
}

// === API Token 状态与操作 ===
const apiTokens = ref<ApiToken[]>([]);
const tokenLoading = ref(false);
const tokenError = ref("");

// 创建表单
const tokenForm = reactive({
  name: "",
  scopes: ["read:workspace"] as string[],
  expiresIn: 90 * 24 * 3600, // 秒；0 = 永不过期
});
const tokenCreating = ref(false);
const tokenCreateError = ref("");
// 创建成功后的一次性明文展示
const createdToken = ref<ApiTokenCreated | null>(null);
const copied = ref(false);

async function loadTokens() {
  tokenLoading.value = true;
  tokenError.value = "";
  try {
    apiTokens.value = await apiTokenApi.list();
  } catch (e: any) {
    tokenError.value = e.message ?? "加载失败";
  } finally {
    tokenLoading.value = false;
  }
}

function toggleScope(scope: string) {
  const idx = tokenForm.scopes.indexOf(scope);
  if (idx >= 0) {
    tokenForm.scopes.splice(idx, 1);
  } else {
    tokenForm.scopes.push(scope);
  }
}

async function createToken() {
  tokenCreateError.value = "";
  createdToken.value = null;
  copied.value = false;
  if (!tokenForm.name.trim()) {
    tokenCreateError.value = "请输入令牌名称";
    return;
  }
  if (tokenForm.scopes.length === 0) {
    tokenCreateError.value = "请至少选择一个权限范围";
    return;
  }
  tokenCreating.value = true;
  try {
    const created = await apiTokenApi.create({
      name: tokenForm.name.trim(),
      scopes: tokenForm.scopes,
      expires_in_seconds: tokenForm.expiresIn > 0 ? tokenForm.expiresIn : undefined,
    });
    createdToken.value = created;
    tokenForm.name = "";
    tokenForm.scopes = ["read:workspace"];
    await loadTokens();
  } catch (e: any) {
    tokenCreateError.value = e.message ?? "创建失败";
  } finally {
    tokenCreating.value = false;
  }
}

async function copyToken() {
  if (!createdToken.value) return;
  try {
    await navigator.clipboard.writeText(createdToken.value.token);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2000);
  } catch {
    // 剪贴板不可用时提示手动复制
    copied.value = false;
  }
}

async function revokeToken(t: ApiToken) {
  if (!confirm(`确定要吊销令牌「${t.name}」吗？吊销后立即失效，且无法恢复。`)) return;
  try {
    await apiTokenApi.revoke(t.id);
    apiTokens.value = apiTokens.value.filter((x) => x.id !== t.id);
    if (createdToken.value?.id === t.id) createdToken.value = null;
  } catch (e: any) {
    alert(`吊销失败：${e.message}`);
  }
}

function tokenStatus(t: ApiToken): "active" | "expired" {
  if (!t.expires_at) return "active";
  return new Date(t.expires_at).getTime() > Date.now() ? "active" : "expired";
}

function expiryLabel(t: ApiToken): string {
  return t.expires_at ? formatDate(t.expires_at) : "永不过期";
}

function lastUsedLabel(t: ApiToken): string {
  return t.last_used_at ? formatDate(t.last_used_at) : "从未使用";
}

function startEdit() {
  if (!ws.value) return;
  editForm.name = ws.value.name;
  editForm.timezone = ws.value.timezone;
  editForm.language = ws.value.language;
  editError.value = "";
  editSuccess.value = "";
  editing.value = true;
}

function cancelEdit() {
  editing.value = false;
  editError.value = "";
}

async function saveEdit() {
  editError.value = "";
  editSuccess.value = "";
  if (!editForm.name.trim()) {
    editError.value = "名称不能为空";
    return;
  }

  editSaving.value = true;
  try {
    const updated = await workspaceApi.update(wsId.value, {
      name: editForm.name.trim(),
      timezone: editForm.timezone,
      language: editForm.language,
    });
    ws.value = updated;
    editSuccess.value = "保存成功";
    setTimeout(() => {
      editing.value = false;
      editSuccess.value = "";
    }, 1200);
  } catch (e) {
    editError.value = e instanceof ApiError ? e.message : "保存失败";
  } finally {
    editSaving.value = false;
  }
}

async function sendInvite() {
  inviteError.value = "";
  inviteSuccess.value = "";
  if (!inviteEmail.value.trim()) {
    inviteError.value = "请输入邮箱";
    return;
  }
  inviteSending.value = true;
  try {
    await workspaceApi.sendInvitation(wsId.value, {
      email: inviteEmail.value.trim(),
      role: inviteRole.value,
      message: inviteMessage.value || undefined,
    });
    inviteSuccess.value = "邀请已发送";
    inviteEmail.value = "";
    inviteMessage.value = "";
    // 刷新邀请列表
    invitations.value = await workspaceApi.listInvitations(wsId.value);
  } catch (e: any) {
    inviteError.value = e.message ?? "发送失败";
  } finally {
    inviteSending.value = false;
  }
}

async function changeRole(m: Member, role: string) {
  if (m.role === role) return;
  const oldRole = m.role;
  m.role = role; // 乐观更新
  try {
    await workspaceApi.changeRole(wsId.value, m.id, role);
  } catch (e: any) {
    m.role = oldRole; // 回滚
    alert(`角色修改失败：${e.message}`);
  }
}

async function removeMember(m: Member) {
  if (!confirm(`确定要移除成员 ${m.display_name} (${m.email}) 吗？`)) return;
  try {
    await workspaceApi.removeMember(wsId.value, m.id);
    members.value = members.value.filter((x) => x.id !== m.id);
  } catch (e: any) {
    alert(`移除失败：${e.message}`);
  }
}

async function revokeInvitation(inv: Invitation) {
  if (!confirm(`撤销对 ${inv.email} 的邀请？`)) return;
  try {
    await workspaceApi.revokeInvitation(wsId.value, inv.id);
    invitations.value = invitations.value.filter((x) => x.id !== inv.id);
  } catch (e: any) {
    alert(`撤销失败：${e.message}`);
  }
}

function roleLabel(role: string | undefined): string {
  return { owner: "所有者", admin: "管理员", member: "成员", guest: "访客" }[role ?? ""] ?? role ?? "-";
}
function formatDate(s: string | undefined): string {
  return s ? s.slice(0, 10) : "-";
}
function invStatusLabel(s: string | undefined): string {
  return { pending: "待处理", accepted: "已接受", revoked: "已撤销", expired: "已过期" }[s ?? ""] ?? s ?? "-";
}

onMounted(loadAll);
</script>

<template>
  <AppLoadingState v-if="loading" text="正在加载工作空间设置..." />
  <AppErrorState
    v-else-if="error"
    :message="error"
    @retry="loadAll"
  />
  <div v-else-if="ws" class="ws-settings">
    <header class="ws-settings__header">
      <h1>{{ ws.name }}</h1>
      <p class="slug">/{{ ws.slug }}</p>
    </header>

    <nav class="tabs">
      <button :class="{ active: activeTab === 'info' }" @click="activeTab = 'info'">基本信息</button>
      <button :class="{ active: activeTab === 'members' }" @click="activeTab = 'members'">
        成员 ({{ members.length }})
      </button>
      <button
        v-if="canManageMembers"
        :class="{ active: activeTab === 'invitations' }"
        @click="activeTab = 'invitations'"
      >
        邀请 ({{ invitations.length }})
      </button>
      <button :class="{ active: activeTab === 'api-tokens' }" @click="activeTab = 'api-tokens'">
        访问令牌
      </button>
    </nav>

    <!-- === 基本信息 === -->
    <section v-if="activeTab === 'info'" class="tab-panel">
      <!-- 只读模式 -->
      <template v-if="!editing">
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">名称</span>
            <span class="info-value">{{ ws.name }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">链接标识</span>
            <span class="info-value">{{ ws.slug }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">时区</span>
            <span class="info-value">{{ ws.timezone }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">语言</span>
            <span class="info-value">{{ ws.language }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">我的角色</span>
            <span class="info-value">
              <span class="role-badge" :data-role="ws.role">{{ roleLabel(ws.role) }}</span>
            </span>
          </div>
          <div class="info-item">
            <span class="info-label">成员数</span>
            <span class="info-value">{{ members.length }}</span>
          </div>
        </div>
        <button
          v-if="canManageMembers"
          class="btn btn--primary"
          style="margin-top: 20px"
          @click="startEdit"
        >
          编辑信息
        </button>
      </template>

      <!-- 编辑模式 -->
      <template v-else>
        <div class="info-grid edit-mode">
          <label class="info-item">
            <span class="info-label">名称</span>
            <input v-model="editForm.name" type="text" maxlength="128" class="info-input" />
          </label>
          <div class="info-item">
            <span class="info-label">链接标识</span>
            <span class="info-value mono">{{ ws.slug }}</span>
          </div>
          <label class="info-item">
            <span class="info-label">时区</span>
            <select v-model="editForm.timezone" class="info-input">
              <option v-for="tz in timezoneOptions" :key="tz.value" :value="tz.value">
                {{ tz.label }}
              </option>
            </select>
          </label>
          <label class="info-item">
            <span class="info-label">语言</span>
            <select v-model="editForm.language" class="info-input">
              <option v-for="lang in languageOptions" :key="lang.value" :value="lang.value">
                {{ lang.label }}
              </option>
            </select>
          </label>
        </div>

        <p v-if="editError" class="form-error">{{ editError }}</p>
        <p v-if="editSuccess" class="form-success">{{ editSuccess }}</p>

        <div class="edit-actions">
          <button class="btn btn--primary" :disabled="editSaving" @click="saveEdit">
            {{ editSaving ? "保存中..." : "保存" }}
          </button>
          <button class="btn" :disabled="editSaving" @click="cancelEdit">取消</button>
        </div>
      </template>
    </section>

    <!-- === 成员管理 === -->
    <section v-if="activeTab === 'members'" class="tab-panel">
      <table class="member-table">
        <thead>
          <tr>
            <th>成员</th>
            <th>邮箱</th>
            <th>加入时间</th>
            <th>角色</th>
            <th v-if="canManageMembers">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in members" :key="m.id">
            <td>
              <div class="member-cell">
                <span class="avatar">{{ m.display_name.charAt(0) }}</span>
                <span>{{ m.display_name }}</span>
              </div>
            </td>
            <td>{{ m.email }}</td>
            <td class="meta">{{ formatDate(m.joined_at) }}</td>
            <td>
              <select
                v-if="canManageMembers && m.role !== 'owner'"
                :value="m.role"
                @change="changeRole(m, ($event.target as HTMLSelectElement).value)"
              >
                <option value="admin">管理员</option>
                <option value="member">成员</option>
                <option value="guest">访客</option>
              </select>
              <span v-else class="role-badge" :data-role="m.role">{{ roleLabel(m.role) }}</span>
            </td>
            <td v-if="canManageMembers">
              <button
                v-if="m.role !== 'owner' && m.id !== auth.user?.id"
                class="btn-link danger"
                @click="removeMember(m)"
              >
                移除
              </button>
              <span v-else class="meta">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- === 邀请管理 === -->
    <section v-if="activeTab === 'invitations' && canManageMembers" class="tab-panel">
      <div class="invite-form">
        <h3>邀请新成员</h3>
        <form @submit.prevent="sendInvite">
          <input
            v-model="inviteEmail"
            class="field__input"
            placeholder="输入邮箱地址"
            type="email"
          />
          <select v-model="inviteRole" class="field__select">
            <option value="admin">管理员</option>
            <option value="member">成员</option>
            <option value="guest">访客</option>
          </select>
          <input
            v-model="inviteMessage"
            class="field__input"
            placeholder="附言（可选）"
            maxlength="500"
          />
          <button class="btn btn--primary" :disabled="inviteSending" type="submit">
            {{ inviteSending ? "发送中..." : "发送邀请" }}
          </button>
        </form>
        <p v-if="inviteError" class="form-error">{{ inviteError }}</p>
        <p v-if="inviteSuccess" class="form-success">{{ inviteSuccess }}</p>
      </div>

      <div v-if="invitations.length" class="invitation-list">
        <h3>邀请记录</h3>
        <table class="member-table">
          <thead>
            <tr>
              <th>邮箱</th>
              <th>角色</th>
              <th>状态</th>
              <th>过期时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="inv in invitations" :key="inv.id">
              <td>{{ inv.email }}</td>
              <td>{{ roleLabel(inv.role) }}</td>
              <td>
                <span class="status-badge" :data-status="inv.status">{{ invStatusLabel(inv.status) }}</span>
              </td>
              <td class="meta">{{ formatDate(inv.expires_at) }}</td>
              <td>
                <button
                  v-if="inv.status === 'pending'"
                  class="btn-link danger"
                  @click="revokeInvitation(inv)"
                >
                  撤销
                </button>
                <span v-else class="meta">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="muted">暂无邀请记录</p>
    </section>

    <!-- === 访问令牌（个人 API Token） === -->
    <section v-if="activeTab === 'api-tokens'" class="tab-panel">
      <div class="invite-form">
        <h3>创建个人访问令牌</h3>
        <p class="token-hint">
          令牌用于脚本/集成访问 API（<span class="mono">X-Api-Key: ydz_...</span>），属于你的个人资源，与空间角色无关。
        </p>

        <!-- 一次性明文展示 -->
        <div v-if="createdToken" class="token-reveal" data-testid="token-reveal">
          <p class="token-reveal__warn">
            ⚠️ 令牌只显示这一次，请立即复制保存。关闭后无法再次查看。
          </p>
          <div class="token-reveal__row">
            <code class="token-reveal__value mono">{{ createdToken.token }}</code>
            <button class="btn btn--primary" data-testid="copy-token" @click="copyToken">
              {{ copied ? "已复制 ✓" : "复制" }}
            </button>
          </div>
        </div>

        <form @submit.prevent="createToken">
          <input
            v-model="tokenForm.name"
            data-testid="token-name"
            class="field__input"
            placeholder="令牌名称（如：CI 部署脚本）"
            maxlength="80"
          />
          <div class="scope-grid">
            <label v-for="s in API_TOKEN_SCOPES" :key="s.value" class="scope-option">
              <input
                type="checkbox"
                :checked="tokenForm.scopes.includes(s.value)"
                :value="s.value"
                @change="toggleScope(s.value)"
              />
              <span class="scope-option__text">
                <span class="scope-option__name mono">{{ s.value }}</span>
                <span class="scope-option__desc">{{ s.label }} — {{ s.desc }}</span>
              </span>
            </label>
          </div>
          <div class="expiry-row">
            <label class="expiry-label">有效期</label>
            <select v-model="tokenForm.expiresIn" class="field__select" data-testid="token-expiry">
              <option v-for="o in TOKEN_EXPIRY_OPTIONS" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </div>
          <button class="btn btn--primary" data-testid="create-token" :disabled="tokenCreating" type="submit">
            {{ tokenCreating ? "创建中..." : "生成令牌" }}
          </button>
        </form>
        <p v-if="tokenCreateError" class="form-error">{{ tokenCreateError }}</p>
      </div>

      <div class="token-list">
        <h3>我的令牌 ({{ apiTokens.length }})</h3>
        <p v-if="tokenError" class="form-error">{{ tokenError }}</p>
        <table v-if="apiTokens.length" class="member-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>权限范围</th>
              <th>创建时间</th>
              <th>最后使用</th>
              <th>过期时间</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in apiTokens" :key="t.id">
              <td>{{ t.name }}</td>
              <td>
                <div class="scope-tags">
                  <span v-for="s in t.scopes" :key="s" class="scope-tag mono" :title="scopeLabel(s)">
                    {{ s }}
                  </span>
                </div>
              </td>
              <td class="meta">{{ formatDate(t.created_at) }}</td>
              <td class="meta">{{ lastUsedLabel(t) }}</td>
              <td class="meta">{{ expiryLabel(t) }}</td>
              <td>
                <span class="status-badge" :data-status="tokenStatus(t)">
                  {{ tokenStatus(t) === "active" ? "有效" : "已过期" }}
                </span>
              </td>
              <td>
                <button class="btn-link danger" data-testid="revoke-token" @click="revokeToken(t)">
                  吊销
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-else class="muted">暂无令牌。创建令牌后可用于脚本/集成调用 API。</p>
      </div>
    </section>
  </div>
</template>

<style scoped>
.ws-settings__header { margin-bottom: 20px; }
.ws-settings__header h1 { font-size: 20px; margin: 0; }
.slug { color: var(--text-tertiary); font-size: 13px; margin: 4px 0 0; }

.tabs {
  display: flex;
  gap: 2px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 24px;
}

.tabs button {
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-tertiary);
  font-size: 13px;
  cursor: pointer;
}

.tabs button.active {
  color: var(--brand-600);
  border-bottom-color: var(--brand-500);
  font-weight: 500;
}

.tab-panel { max-width: 720px; }

.info-grid {
  display: grid;
  gap: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.info-label {
  width: 100px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.info-value {
  font-size: 14px;
  color: var(--text-primary);
}

.role-badge,
.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  background: var(--surface-3);
  color: var(--text-secondary);
}

.role-badge[data-role="owner"] {
  background: var(--brand-50);
  color: var(--brand-600);
}

.status-badge[data-status="pending"] { background: var(--brand-50); color: var(--brand-600); }
.status-badge[data-status="accepted"] { background: rgba(15, 194, 123, 0.12); color: var(--success-500); }
.status-badge[data-status="revoked"] { background: rgba(220, 47, 47, 0.12); color: var(--danger-500); }
.status-badge[data-status="expired"] { background: var(--surface-3); color: var(--text-tertiary); }

.member-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.member-table th {
  text-align: left;
  padding: 8px 12px;
  color: var(--text-tertiary);
  font-weight: 500;
  font-size: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.member-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-primary);
}

.member-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--brand-50);
  color: var(--brand-600);
  font-weight: 600;
  font-size: 12px;
}

.meta { color: var(--text-tertiary); font-size: 12px; }

select {
  padding: 4px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--surface-1);
  color: var(--text-primary);
}

.btn-link {
  background: none;
  border: none;
  color: var(--brand-500);
  cursor: pointer;
  font-size: 13px;
  padding: 0;
}

.btn-link.danger { color: var(--danger-500); }

.invite-form {
  margin-bottom: 32px;
  padding: 20px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--surface-2);
}

.invite-form h3 { font-size: 14px; margin: 0 0 12px; }

.invite-form form {
  display: grid;
  grid-template-columns: 1fr 120px 1fr auto;
  gap: 10px;
  align-items: center;
}

.invitation-list h3 { font-size: 14px; margin: 0 0 12px; }

.muted { color: var(--text-tertiary); font-size: 13px; }

.form-error { color: var(--danger-500); font-size: 12px; margin: 8px 0 0; }
.form-success { color: var(--success-500); font-size: 12px; margin: 8px 0 0; }

.field__input,
.field__select {
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
}

.btn {
  padding: 8px 14px;
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

@media (max-width: 600px) {
  .invite-form form { grid-template-columns: 1fr; }
}

/* === API Token === */
.token-hint {
  margin: 0 0 14px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.token-reveal {
  margin: 0 0 16px;
  padding: 14px;
  border: 1px solid var(--warning-500, #f59e0b);
  border-radius: var(--radius-md);
  background: rgba(245, 158, 11, 0.08);
}

.token-reveal__warn {
  margin: 0 0 10px;
  font-size: 12px;
  color: var(--warning-500, #f59e0b);
  font-weight: 500;
}

.token-reveal__row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.token-reveal__value {
  flex: 1;
  padding: 8px 10px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 12px;
  word-break: break-all;
  color: var(--text-primary);
}

.scope-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 6px 16px;
  margin: 12px 0;
  max-height: 260px;
  overflow-y: auto;
}

.scope-option {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
}

.scope-option:hover {
  background: var(--surface-3);
}

.scope-option input {
  margin-top: 3px;
  accent-color: var(--brand-500);
}

.scope-option__text {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.scope-option__name {
  font-weight: 500;
  color: var(--text-primary);
}

.scope-option__desc {
  font-size: 12px;
  color: var(--text-tertiary);
}

.expiry-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 12px 0 14px;
}

.expiry-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.token-list h3 {
  font-size: 14px;
  margin: 0 0 12px;
}

.scope-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  max-width: 320px;
}

.scope-tag {
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--surface-3);
  color: var(--text-secondary);
  font-size: 11px;
}

.status-badge[data-status="expired"] {
  background: var(--surface-3);
  color: var(--text-tertiary);
}

/* 编辑模式 */
.edit-mode .info-item {
  padding: 10px 0;
}

.info-input {
  padding: 6px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
  flex: 1;
  max-width: 360px;
}

.info-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.mono { font-family: var(--font-mono); }

.edit-actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
}
</style>
