<script setup lang="ts">
/**
 * SSOSettingsView — 工作空间单点登录（SSO）集成管理。
 *
 * 支持 OIDC / SAML Provider 的增删改查。
 * 注意：client_secret 不回显（后端只返回 has_client_secret）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  ssoApi,
  type ProviderModifyInput,
  type SSOProtocol,
  type SSOProviderListItem,
} from "@/api/services/sso";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const providers = ref<SSOProviderListItem[]>([]);

/** 编辑弹窗状态。 */
const showModal = ref(false);
const editingId = ref<number | null>(null);
const form = ref<ProviderModifyInput>({
  name: "",
  protocol: "oidc",
  issuer_url: "",
  client_id: "",
  client_secret: "",
  redirect_uri: "",
  auth_url: "",
  token_url: "",
  userinfo_url: "",
  jwks_url: "",
  scopes: "openid email profile",
  auto_create_user: true,
  default_role: "member",
});
const saving = ref(false);

const isEditing = computed(() => editingId.value !== null);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    providers.value = await ssoApi.list(workspaceId.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingId.value = null;
  form.value = {
    name: "",
    protocol: "oidc",
    issuer_url: "",
    client_id: "",
    client_secret: "",
    redirect_uri: "",
    auth_url: "",
    token_url: "",
    userinfo_url: "",
    jwks_url: "",
    scopes: "openid email profile",
    auto_create_user: true,
    default_role: "member",
  };
  showModal.value = true;
}

function openEdit(p: SSOProviderListItem) {
  editingId.value = p.id;
  form.value = {
    name: p.name,
    protocol: p.protocol,
    issuer_url: p.issuer_url,
    client_id: p.client_id,
    client_secret: "",
    redirect_uri: "",
    auth_url: p.auth_url,
    token_url: "",
    userinfo_url: "",
    jwks_url: "",
    scopes: p.scopes,
    auto_create_user: p.auto_createUser,
    default_role: p.default_role,
  };
  showModal.value = true;
}

async function submit() {
  const f = form.value;
  if (!f.name.trim()) {
    toast.error("请填写 Provider 名称");
    return;
  }
  if (!f.issuer_url.trim()) {
    toast.error("请填写 Issuer URL");
    return;
  }
  if (!f.client_id.trim()) {
    toast.error("请填写 Client ID");
    return;
  }

  saving.value = true;
  try {
    if (isEditing.value) {
      await ssoApi.update(workspaceId.value, editingId.value!, f);
      toast.success("SSO Provider 已更新");
    } else {
      await ssoApi.create(workspaceId.value, f);
      toast.success("SSO Provider 已创建");
    }
    showModal.value = false;
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function remove(p: SSOProviderListItem) {
  if (!confirm(`确定删除 SSO Provider「${p.name}」？`)) return;
  try {
    await ssoApi.remove(workspaceId.value, p.id);
    providers.value = providers.value.filter((x) => x.id !== p.id);
    toast.success("已删除");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "删除失败");
  }
}

function protocolLabel(p: SSOProtocol): string {
  return p === "oidc" ? "OIDC" : "SAML";
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">单点登录（SSO）</h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          配置 OIDC / SAML 身份提供商，实现工作空间统一登录
        </p>
      </div>
      <button
        class="rounded-md bg-[var(--brand-500)] px-3 py-1.5 text-sm text-white hover:opacity-90"
        @click="openCreate"
      >
        + 新增 Provider
      </button>
    </header>

    <AppSkeleton v-if="loading" variant="table" :rows="3" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="providers.length === 0"
      title="暂无 SSO Provider"
      description="点击右上角「新增 Provider」配置单点登录"
    />

    <div v-else class="overflow-hidden rounded-lg border border-[var(--border-subtle)]">
      <table class="w-full text-sm">
        <thead class="bg-[var(--surface-2)] text-xs uppercase tracking-wider text-[var(--text-tertiary)]">
          <tr>
            <th class="px-4 py-2.5 text-left font-medium">名称</th>
            <th class="px-4 py-2.5 text-left font-medium">协议</th>
            <th class="px-4 py-2.5 text-left font-medium">Issuer</th>
            <th class="px-4 py-2.5 text-left font-medium">Client ID</th>
            <th class="px-4 py-2.5 text-left font-medium">默认角色</th>
            <th class="px-4 py-2.5 text-left font-medium">状态</th>
            <th class="px-4 py-2.5 text-right font-medium">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border-subtle)]">
          <tr v-for="p in providers" :key="p.id" class="hover:bg-[var(--surface-2)]">
            <td class="px-4 py-3 font-medium text-[var(--text-primary)]">{{ p.name }}</td>
            <td class="px-4 py-3">
              <span class="inline-flex rounded-full bg-[var(--surface-3)] px-2 py-0.5 text-xs text-[var(--text-secondary)]">
                {{ protocolLabel(p.protocol) }}
              </span>
            </td>
            <td class="max-w-[220px] truncate px-4 py-3 text-xs text-[var(--text-secondary)]" :title="p.issuer_url">
              {{ p.issuer_url }}
            </td>
            <td class="px-4 py-3 font-mono text-xs text-[var(--text-secondary)]">{{ p.client_id }}</td>
            <td class="px-4 py-3 text-[var(--text-secondary)]">{{ p.default_role }}</td>
            <td class="px-4 py-3">
              <span
                class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                :class="p.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'"
              >
                {{ p.enabled ? "已启用" : "已停用" }}
              </span>
            </td>
            <td class="px-4 py-3 text-right">
              <div class="flex justify-end gap-3">
                <button class="text-xs text-[var(--brand-600)] hover:underline" @click="openEdit(p)">编辑</button>
                <button class="text-xs text-[var(--danger, #ef4444)] hover:underline" @click="remove(p)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 编辑弹窗 -->
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-6"
      @click.self="showModal = false"
    >
      <div class="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-1)] shadow-xl">
        <h2 class="m-0 border-b border-[var(--border-subtle)] px-5 py-4 text-base font-semibold">
          {{ isEditing ? "编辑 SSO Provider" : "新增 SSO Provider" }}
        </h2>
        <div class="space-y-3 px-5 py-4">
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>名称 *</span>
            <input v-model="form.name" type="text" placeholder="如：企业微信 OIDC" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>协议</span>
            <select v-model="form.protocol" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]">
              <option value="oidc">OIDC</option>
              <option value="saml">SAML</option>
            </select>
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>Issuer URL *</span>
            <input v-model="form.issuer_url" type="text" placeholder="https://sso.example.com" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>Client ID *</span>
            <input v-model="form.client_id" type="text" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>
              Client Secret
              <span v-if="isEditing" class="text-xs text-[var(--text-tertiary)]">（留空表示不修改）</span>
            </span>
            <input v-model="form.client_secret" type="password" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>Auth URL</span>
            <input v-model="form.auth_url" type="text" placeholder="可选（OIDC 可留空自动发现）" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>Scopes</span>
            <input v-model="form.scopes" type="text" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>默认角色</span>
            <input v-model="form.default_role" type="text" placeholder="member" class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]" />
          </label>
          <label class="flex items-center gap-2 text-sm text-[var(--text-secondary)]">
            <input v-model="form.auto_create_user" type="checkbox" class="h-3.5 w-3.5" />
            <span>自动创建用户（首次登录时自动注册）</span>
          </label>
        </div>
        <div class="flex justify-end gap-2 border-t border-[var(--border-subtle)] px-5 py-3">
          <button class="rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-sm hover:bg-[var(--surface-2)]" @click="showModal = false">
            取消
          </button>
          <button
            class="rounded-md bg-[var(--brand-500)] px-3 py-1.5 text-sm text-white disabled:opacity-50"
            :disabled="saving"
            @click="submit"
          >
            {{ saving ? "保存中…" : "保存" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
