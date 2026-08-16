<script setup lang="ts">
/**
 * 邀请预览 — 公开可读的邀请接受页（无需登录）。
 * 展示工作空间信息，登录后可接受邀请。
 */
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type InvitationPreview } from "@/api/services/workspace";
import { AppErrorState, AppSkeleton } from "@/components";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const loading = ref(true);
const error = ref("");
const preview = ref<InvitationPreview | null>(null);
const accepting = ref(false);

async function load() {
  const token = String(route.params.token ?? "");
  if (!token) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    preview.value = await workspaceApi.previewInvitation(token);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "邀请无效或已过期";
  } finally {
    loading.value = false;
  }
}

async function accept() {
  const token = String(route.params.token ?? "");
  accepting.value = true;
  try {
    const inv = await workspaceApi.acceptInvitation(token);
    toast.success("已加入工作空间");
    router.push(`/${inv.workspace_id}/workbench`);
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "接受失败");
  } finally {
    accepting.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="flex min-h-screen items-center justify-center p-4">
    <div class="w-full max-w-md">
      <div v-if="loading" class="space-y-3">
        <AppSkeleton v-for="i in 3" :key="i" class="h-12 w-full" />
      </div>

      <AppErrorState v-else-if="error" :message="error" />

      <div
        v-else-if="preview"
        class="space-y-4 rounded-lg border border-[var(--border-subtle)] p-6 text-center"
      >
        <div
          class="mx-auto flex h-14 w-14 items-center justify-center rounded-md bg-[var(--brand-600)] text-2xl font-bold text-white"
        >
          {{ preview.workspace_name.slice(0, 1).toUpperCase() }}
        </div>
        <div>
          <h1 class="text-xl font-bold text-[var(--text-primary)]">
            加入「{{ preview.workspace_name }}」
          </h1>
          <p class="mt-1 text-sm text-[var(--text-secondary)]">
            {{ preview.inviter_name }} 邀请你以
            <span class="font-medium text-[var(--text-primary)]">{{ preview.role }}</span>
            角色加入该工作空间
          </p>
          <p class="mt-0.5 text-xs text-[var(--text-tertiary)]">
            {{ preview.email }}
          </p>
        </div>

        <template v-if="auth.user">
          <button
            :disabled="accepting"
            class="w-full rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
            @click="accept"
          >
            {{ accepting ? "接受中…" : "接受邀请" }}
          </button>
        </template>
        <template v-else>
          <button
            class="w-full rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
            @click="router.push('/login')"
          >
            登录后接受邀请
          </button>
        </template>
      </div>
    </div>
  </div>
</template>
