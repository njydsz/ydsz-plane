<script setup lang="ts">
/**
 * 死信队列监控 — 查看/重试/清理领域事件消费失败的消息。
 * 数据来源：dlqApi（后端 dlq_events 表）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { dlqApi, type DLQItem } from "@/api/services/dlq";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const items = ref<DLQItem[]>([]);
const total = ref(0);

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    const r = await dlqApi.list(workspaceId.value, { limit: 100, unresolved_only: true });
    items.value = r.items ?? [];
    total.value = r.total ?? 0;
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function retry(item: DLQItem) {
  try {
    await dlqApi.retry(workspaceId.value, item.id);
    toast.success("已重新投递");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "重试失败");
  }
}

async function remove(item: DLQItem) {
  try {
    await dlqApi.remove(workspaceId.value, item.id);
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "清理失败");
  }
}

function truncate(s: string, n = 120): string {
  return s.length > n ? s.slice(0, n) + "…" : s;
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">死信队列</h1>
      <button class="text-sm text-[var(--brand-600)] hover:underline" @click="load">刷新</button>
    </div>

    <div v-if="total > 0" class="text-xs text-[var(--text-tertiary)]">
      共 {{ total }} 条未解决消息
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-20 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <AppEmptyState
      v-else-if="items.length === 0"
      title="死信队列为空"
      description="领域事件消费失败的消息会进入死信队列。"
    />

    <div v-else class="space-y-3">
      <div
        v-for="item in items"
        :key="item.id"
        class="rounded-md border border-[var(--border-subtle)] p-4"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="font-mono text-xs text-[var(--brand-600)]">#{{ item.id }}</span>
            <span class="text-sm font-medium text-[var(--text-primary)]">{{ item.queue }}</span>
            <span class="text-xs text-[var(--text-tertiary)]">{{ item.routing_key }}</span>
          </div>
          <div class="flex gap-2">
            <button class="text-xs text-[var(--brand-600)]" @click="retry(item)">重试</button>
            <button class="text-xs text-[var(--danger, #ef4444)]" @click="remove(item)">清理</button>
          </div>
        </div>
        <div class="mt-1 text-xs text-[var(--danger, #ef4444)]">{{ item.error_reason }}</div>
        <pre class="mt-2 overflow-x-auto rounded bg-[var(--surface-2)] p-2 text-xs text-[var(--text-secondary)]">{{ truncate(JSON.stringify(item.payload, null, 2)) }}</pre>
      </div>
    </div>
  </div>
</template>
