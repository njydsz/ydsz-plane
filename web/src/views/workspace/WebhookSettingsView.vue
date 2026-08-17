<script setup lang="ts">
/**
 * Webhook 管理 — 订阅列表 + 创建/启停/测试/删除。
 * 数据来源：webhookApi（后端 webhook 域）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  webhookApi,
  type Webhook,
  type WebhookEvent,
} from "@/api/services/webhook";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const webhooks = ref<Webhook[]>([]);
const showForm = ref(false);
const saving = ref(false);
const secret = ref<string | null>(null);

const ALL_EVENTS: WebhookEvent[] = [
  "issue.created", "issue.updated", "issue.status_changed", "issue.commented",
  "issue.assigned", "sprint.started", "sprint.completed", "version.released",
  "project.created", "member.joined", "member.removed", "comment.created",
  "attachment.uploaded",
];

const form = ref({
  name: "",
  url: "",
  events: [] as WebhookEvent[],
});

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    const r = await webhookApi.list(workspaceId.value, { limit: 100 });
    webhooks.value = r.items ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function toggleEvent(ev: WebhookEvent) {
  const i = form.value.events.indexOf(ev);
  if (i >= 0) form.value.events.splice(i, 1);
  else form.value.events.push(ev);
}

async function create() {
  if (!form.value.name.trim() || !form.value.url.trim() || form.value.events.length === 0) {
    toast.warning("请填写名称、URL 并至少选择一个事件");
    return;
  }
  saving.value = true;
  try {
    const r = await webhookApi.create(workspaceId.value, form.value);
    secret.value = r.secret;
    form.value = { name: "", url: "", events: [] };
    showForm.value = false;
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

async function toggle(w: Webhook) {
  try {
    if (w.status === "paused") await webhookApi.resume(workspaceId.value, w.id);
    else await webhookApi.pause(workspaceId.value, w.id);
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function test(w: Webhook) {
  try {
    const r = await webhookApi.test(workspaceId.value, w.id);
    if (r.success) toast.success("测试投递成功");
    else toast.error(`测试失败${r.error ? `：${r.error}` : ""}`);
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "测试失败");
  }
}

async function remove(w: Webhook) {
  if (!confirm(`确定删除 Webhook「${w.name}」？`)) return;
  try {
    await webhookApi.delete(workspaceId.value, w.id);
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "删除失败");
  }
}

function eventLabel(ev: WebhookEvent): string {
  const map: Record<string, string> = {
    "issue.created": "需求/任务/缺陷创建", "issue.updated": "需求/任务/缺陷更新",
    "issue.status_changed": "状态流转", "issue.commented": "评论",
    "issue.assigned": "指派", "sprint.started": "迭代启动",
    "sprint.completed": "迭代完成", "version.released": "版本发布",
    "project.created": "项目创建", "member.joined": "成员加入",
    "member.removed": "成员移除", "comment.created": "评论创建",
    "attachment.uploaded": "附件上传",
  };
  return map[ev] ?? ev;
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">Webhook</h1>
      <button
        class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
        @click="showForm = !showForm"
      >
        {{ showForm ? "取消" : "新建 Webhook" }}
      </button>
    </div>

    <!-- 创建后回显 secret -->
    <div
      v-if="secret"
      class="rounded-md border border-[var(--brand-400)] bg-[var(--surface-2)] px-4 py-3 text-sm"
    >
      <div class="font-medium text-[var(--text-primary)]">Webhook 已创建</div>
      <div class="mt-1 text-[var(--text-secondary)]">
        签名密钥（仅显示一次，请妥善保存）：
        <code class="font-mono text-[var(--brand-600)]">{{ secret }}</code>
      </div>
      <button class="mt-1 text-xs text-[var(--brand-600)]" @click="secret = null">关闭</button>
    </div>

    <!-- 创建表单 -->
    <div
      v-if="showForm"
      class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4"
    >
      <div>
        <label class="text-xs text-[var(--text-tertiary)]">名称</label>
        <input
          v-model="form.name"
          type="text"
          class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
          placeholder="如：CI 流水线通知"
        />
      </div>
      <div>
        <label class="text-xs text-[var(--text-tertiary)]">URL</label>
        <input
          v-model="form.url"
          type="url"
          class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm"
          placeholder="https://example.com/hook"
        />
      </div>
      <div>
        <label class="text-xs text-[var(--text-tertiary)]">订阅事件</label>
        <div class="mt-1 grid grid-cols-2 gap-1 sm:grid-cols-3">
          <label
            v-for="ev in ALL_EVENTS"
            :key="ev"
            class="flex items-center gap-1.5 text-xs text-[var(--text-secondary)]"
          >
            <input
              type="checkbox"
              :checked="form.events.includes(ev)"
              @change="toggleEvent(ev)"
            />
            {{ eventLabel(ev) }}
          </label>
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
      v-else-if="webhooks.length === 0"
      title="暂无 Webhook"
      description="创建 Webhook 以便在事件发生时向外部系统推送通知。"
    />

    <div v-else class="space-y-3">
      <div
        v-for="w in webhooks"
        :key="w.id"
        class="rounded-md border border-[var(--border-subtle)] p-4"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium text-[var(--text-primary)]">{{ w.name }}</span>
            <span
              class="rounded-full px-2 py-0.5 text-xs"
              :class="w.status === 'active'
                ? 'bg-[var(--brand-50)] text-[var(--brand-600)]'
                : 'bg-[var(--surface-2)] text-[var(--text-tertiary)]'"
            >
              {{ w.status === "active" ? "运行中" : w.status === "paused" ? "已暂停" : "异常" }}
            </span>
          </div>
          <div class="flex gap-2">
            <button class="text-xs text-[var(--brand-600)]" @click="test(w)">测试</button>
            <button class="text-xs text-[var(--brand-600)]" @click="toggle(w)">
              {{ w.status === "paused" ? "启用" : "暂停" }}
            </button>
            <button class="text-xs text-[var(--danger, #ef4444)]" @click="remove(w)">删除</button>
          </div>
        </div>
        <div class="mt-1 truncate font-mono text-xs text-[var(--text-tertiary)]">{{ w.url }}</div>
        <div class="mt-2 flex flex-wrap gap-1">
          <span
            v-for="ev in w.events"
            :key="ev"
            class="rounded bg-[var(--surface-2)] px-1.5 py-0.5 text-xs text-[var(--text-secondary)]"
          >
            {{ eventLabel(ev) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
