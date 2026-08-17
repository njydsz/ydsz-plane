<script setup lang="ts">
/**
 * 工作空间设置 — 基本信息、Logo、成员邀请管理。
 * 数据来源：workspaceApi。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  workspaceApi,
  type Workspace,
  type Invitation,
} from "@/api/services/workspace";
import { AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";
import { useBrandColor, BRAND_COLOR_PRESETS } from "@/composables/useBrandColor";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const ws = ref<Workspace | null>(null);
const invitations = ref<Invitation[]>([]);
const saving = ref(false);

const form = ref({ name: "", timezone: "Asia/Shanghai", language: "zh-CN", brand_color: "" });
const invite = ref({ email: "", role: "member" });

// 品牌色
const workspaceBrandColor = ref<string | undefined>(undefined);
const brand = useBrandColor(workspaceBrandColor);
const brandSaving = ref(false);

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    const [w, inv] = await Promise.all([
      workspaceApi.get(workspaceId.value),
      workspaceApi.listInvitations(workspaceId.value).catch(() => [] as Invitation[]),
    ]);
    ws.value = w;
    form.value = { name: w.name, timezone: w.timezone, language: w.language, brand_color: "" };
    workspaceBrandColor.value = w.brand_color || undefined;
    invitations.value = inv;
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (!form.value.name.trim()) { toast.warning("请输入名称"); return; }
  saving.value = true;
  try {
    await workspaceApi.update(workspaceId.value, form.value);
    toast.success("已保存");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function saveBrandColor(color: string) {
  brandSaving.value = true;
  try {
    await workspaceApi.update(workspaceId.value, { brand_color: color || undefined });
    workspaceBrandColor.value = color || undefined;
    toast.success("品牌色已更新");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "保存失败");
  } finally {
    brandSaving.value = false;
  }
}

async function sendInvite() {
  if (!invite.value.email.trim()) { toast.warning("请输入邮箱"); return; }
  try {
    await workspaceApi.sendInvitation(workspaceId.value, invite.value);
    invite.value.email = "";
    toast.success("邀请已发送");
    invitations.value = await workspaceApi.listInvitations(workspaceId.value);
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "发送失败");
  }
}

async function revoke(inv: Invitation) {
  try {
    await workspaceApi.revokeInvitation(workspaceId.value, inv.id);
    invitations.value = invitations.value.filter((i) => i.id !== inv.id);
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "撤销失败");
  }
}

onMounted(load);
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">工作空间设置</h1>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 基本信息 -->
      <section class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">基本信息</h2>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">名称</label>
          <input v-model="form.name" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">时区</label>
            <input v-model="form.timezone" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="text-xs text-[var(--text-tertiary)]">语言</label>
            <select v-model="form.language" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm">
              <option value="zh-CN">简体中文</option>
              <option value="en">English</option>
            </select>
          </div>
        </div>
        <button
          :disabled="saving"
          class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
          @click="save"
        >
          {{ saving ? "保存中…" : "保存" }}
        </button>
      </section>

      <!-- 品牌定制 -->
      <section class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">品牌定制</h2>
        <p class="text-xs text-[var(--text-tertiary)]">自定义工作空间主题色，全局应用</p>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">品牌色</label>
          <div class="mt-2 flex items-center gap-3">
            <input
              v-model="brand.currentColor"
              type="color"
              class="h-9 w-12 cursor-pointer rounded border border-[var(--border-subtle)] p-0.5"
            />
            <input
              v-model="brand.currentColor"
              type="text"
              placeholder="#2563eb"
              class="w-28 rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-sm uppercase"
            />
            <button
              :disabled="brandSaving"
              class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-xs font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
              @click="saveBrandColor(brand.currentColor)"
            >
              {{ brandSaving ? "保存中…" : "应用" }}
            </button>
            <button
              class="rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-xs text-[var(--text-secondary)] hover:bg-[var(--bg-secondary)]"
              @click="brand.resetBrandColor(); saveBrandColor('')"
            >
              重置
            </button>
          </div>
        </div>
        <div class="mt-2">
          <label class="text-xs text-[var(--text-tertiary)]">预设色板</label>
          <div class="mt-2 flex flex-wrap gap-2">
            <button
              v-for="preset in brand.presets"
              :key="preset.value"
              class="flex items-center gap-1.5 rounded-md border border-[var(--border-subtle)] px-2.5 py-1.5 text-xs transition-all hover:border-[var(--brand-400)] hover:shadow-sm"
              :class="{ 'ring-2 ring-[var(--brand-500)] ring-offset-1': brand.currentColor === preset.value }"
              @click="brand.setBrandColor(preset.value); saveBrandColor(preset.value)"
            >
              <span
                class="inline-block h-3.5 w-3.5 rounded-full border border-[var(--border-subtle)]"
                :style="{ backgroundColor: preset.value }"
              ></span>
              {{ preset.name }}
            </button>
          </div>
        </div>
      </section>

      <!-- 成员邀请 -->
      <section class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">邀请成员</h2>
        <div class="flex gap-2">
          <input
            v-model="invite.email"
            type="email"
            placeholder="member@example.com"
            class="flex-1 rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
          />
          <select v-model="invite.role" class="rounded-md border border-[var(--border-subtle)] px-2 py-2 text-sm">
            <option value="member">成员</option>
            <option value="admin">管理员</option>
            <option value="guest">访客</option>
          </select>
          <button
            class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
            @click="sendInvite"
          >
            邀请
          </button>
        </div>

        <div v-if="invitations.length" class="space-y-2">
          <div
            v-for="inv in invitations"
            :key="inv.id"
            class="flex items-center justify-between rounded-md border border-[var(--border-subtle)] px-3 py-2"
          >
            <div class="text-sm">
              <span class="text-[var(--text-primary)]">{{ inv.email }}</span>
              <span class="ml-2 text-xs text-[var(--text-tertiary)]">
                {{ inv.role }} · {{ inv.status }}
              </span>
            </div>
            <button
              v-if="inv.status === 'pending'"
              class="text-xs text-[var(--danger, #ef4444)]"
              @click="revoke(inv)"
            >
              撤销
            </button>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
