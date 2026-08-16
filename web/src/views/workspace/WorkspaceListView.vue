<script setup lang="ts">
/**
 * 工作空间列表 — 首页，展示当前用户加入的所有工作空间 + 新建入口。
 * 数据来源：workspaceApi.list / create。
 */
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import { workspaceApi, type Workspace } from "@/api/services/workspace";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const router = useRouter();

const loading = ref(true);
const error = ref("");
const workspaces = ref<Workspace[]>([]);
const showForm = ref(false);
const saving = ref(false);
const form = ref({ name: "", slug: "" });

async function load() {
  loading.value = true;
  error.value = "";
  try {
    workspaces.value = await workspaceApi.list();
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function create() {
  if (!form.value.name.trim()) {
    toast.warning("请输入工作空间名称");
    return;
  }
  saving.value = true;
  try {
    const ws = await workspaceApi.create(form.value);
    form.value = { name: "", slug: "" };
    showForm.value = false;
    await load();
    router.push(`/${ws.id}/workbench`);
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

function open(ws: Workspace) {
  router.push(`/${ws.id}/workbench`);
}

onMounted(load);
</script>

<template>
  <div class="mx-auto max-w-4xl space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">工作空间</h1>
      <button
        class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
        @click="showForm = !showForm"
      >
        {{ showForm ? "取消" : "新建工作空间" }}
      </button>
    </div>

    <!-- 新建表单 -->
    <div v-if="showForm" class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">名称</label>
          <input v-model="form.name" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="我的团队" />
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">Slug（可选）</label>
          <input v-model="form.slug" type="text" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="my-team" />
        </div>
      </div>
      <button
        :disabled="saving"
        class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
        @click="create"
      >
        {{ saving ? "创建中…" : "创建" }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <AppEmptyState
      v-else-if="workspaces.length === 0"
      title="暂无工作空间"
      description="创建你的第一个工作空间，开始管理需求、任务与缺陷。"
    />

    <div v-else class="grid gap-4 sm:grid-cols-2">
      <button
        v-for="ws in workspaces"
        :key="ws.id"
        class="flex items-center gap-3 rounded-lg border border-[var(--border-subtle)] p-4 text-left hover:bg-[var(--surface-2)] transition-colors"
        @click="open(ws)"
      >
        <div
          class="flex h-10 w-10 items-center justify-center rounded-md bg-[var(--brand-600)] text-lg font-bold text-white"
        >
          {{ ws.name.slice(0, 1).toUpperCase() }}
        </div>
        <div class="min-w-0">
          <div class="truncate text-sm font-medium text-[var(--text-primary)]">{{ ws.name }}</div>
          <div class="mt-0.5 text-xs text-[var(--text-tertiary)]">
            {{ ws.slug }}<span v-if="ws.member_count != null"> · {{ ws.member_count }} 成员</span>
          </div>
        </div>
      </button>
    </div>
  </div>
</template>
