<script setup lang="ts">
/**
 * 通知中心 — 站内通知列表 + 未读/已读/归档操作。
 * 数据来源：notificationApi（后端 notification 域）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  notificationApi,
  type AppNotification,
} from "@/api/services/notification";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const items = ref<AppNotification[]>([]);
const unread = ref(0);

const filter = ref<"all" | "unread">("all");

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [list, count] = await Promise.all([
      notificationApi.list(workspaceId.value, {
        limit: 100,
        is_read: filter.value === "unread" ? false : undefined,
      }),
      notificationApi.unreadCount(workspaceId.value),
    ]);
    items.value = list.items ?? [];
    unread.value = count;
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function markRead(n: AppNotification) {
  if (n.is_read) return;
  try {
    await notificationApi.markRead(workspaceId.value, n.id);
    n.is_read = true;
    unread.value = Math.max(0, unread.value - 1);
  } catch {
    /* 忽略单条失败 */
  }
}

async function markAllRead() {
  try {
    await notificationApi.markAllRead(workspaceId.value);
    items.value.forEach((n) => (n.is_read = true));
    unread.value = 0;
  } catch {
    /* 忽略 */
  }
}

function open(n: AppNotification) {
  if (!n.is_read) markRead(n);
  if (n.action_url) router.push(n.action_url);
}

function eventLabel(eventType: string): string {
  const map: Record<string, string> = {
"issue.created": "新建需求/任务/缺陷",
"issue.updated": "更新需求/任务/缺陷",
"issue.deleted": "删除需求/任务/缺陷",
    "issue.commented": "评论",
    "issue.assigned": "指派",
    "issue.status_changed": "状态流转",
    "sprint.started": "迭代启动",
    "sprint.completed": "迭代完成",
  };
  return map[eventType] ?? eventType;
}

function timeLabel(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  const now = Date.now();
  const diff = now - d.getTime();
  if (diff < 60_000) return "刚刚";
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
  return iso.slice(0, 10);
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">通知</h1>
      <button
        v-if="unread > 0"
        class="text-sm text-[var(--brand-600)] hover:underline"
        @click="markAllRead"
      >
        全部已读（{{ unread }}）
      </button>
    </div>

    <!-- 过滤 -->
    <div class="flex gap-2">
      <button
        v-for="f in [{ k: 'all', l: '全部' }, { k: 'unread', l: '未读' }] as const"
        :key="f.k"
        class="rounded-md px-3 py-1.5 text-sm"
        :class="filter === f.k
          ? 'bg-[var(--brand-600)] text-white'
          : 'border border-[var(--border-subtle)] text-[var(--text-secondary)]'"
        @click="filter = f.k; load()"
      >
        {{ f.l }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 5" :key="i" class="h-14 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <AppEmptyState
      v-else-if="items.length === 0"
      title="暂无通知"
      description="当有新的需求/任务/缺陷动态时，通知会显示在这里。"
    />

    <div v-else class="space-y-2">
      <div
        v-for="n in items"
        :key="n.id"
        class="flex cursor-pointer items-start gap-3 rounded-md border px-4 py-3 transition-colors"
        :class="n.is_read ? 'border-[var(--border-subtle)]' : 'border-[var(--brand-400)] bg-[var(--surface-2)]'"
        @click="open(n)"
      >
        <span
          v-if="!n.is_read"
          class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[var(--brand-600)]"
        />
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span class="text-xs text-[var(--brand-600)]">{{ eventLabel(n.event_type) }}</span>
            <span class="text-sm font-medium text-[var(--text-primary)]">{{ n.title }}</span>
          </div>
          <div class="mt-0.5 truncate text-sm text-[var(--text-secondary)]">{{ n.body }}</div>
          <div class="mt-1 text-xs text-[var(--text-tertiary)]">
            {{ n.actor_name ? `${n.actor_name} · ` : "" }}{{ timeLabel(n.created_at) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
