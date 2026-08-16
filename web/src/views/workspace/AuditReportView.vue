<script setup lang="ts">
/**
 * 审计日志 — 工作空间操作审计记录。
 * 数据来源：auditApi（后端 audit_logs 表）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { auditApi, type AuditLogEntry } from "@/api/services/audit";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const logs = ref<AuditLogEntry[]>([]);

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    logs.value = await auditApi.list(workspaceId.value, 200);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function actionLabel(action: string): string {
  const map: Record<string, string> = {
    "workspace.create": "创建工作空间",
    "workspace.update": "更新工作空间",
    "member.invite": "邀请成员",
    "member.remove": "移除成员",
    "member.role_change": "变更角色",
    "project.create": "创建项目",
    "project.archive": "归档项目",
  };
  return map[action] ?? action;
}

function timeLabel(iso: string): string {
  return new Date(iso).toLocaleString();
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">审计日志</h1>
      <button class="text-sm text-[var(--brand-600)] hover:underline" @click="load">刷新</button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <AppEmptyState
      v-else-if="logs.length === 0"
      title="暂无审计记录"
      description="工作空间内的关键操作将被记录在此。"
    />

    <div v-else class="overflow-hidden rounded-md border border-[var(--border-subtle)]">
      <table class="w-full text-sm">
        <thead class="bg-[var(--surface-2)] text-xs uppercase tracking-wider text-[var(--text-tertiary)]">
          <tr>
            <th class="px-4 py-2 text-left font-medium">操作</th>
            <th class="px-4 py-2 text-left font-medium">操作人</th>
            <th class="px-4 py-2 text-left font-medium">对象</th>
            <th class="hidden md:table-cell px-4 py-2 text-left font-medium">IP</th>
            <th class="px-4 py-2 text-left font-medium">时间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border-subtle)]">
          <tr v-for="l in logs" :key="l.id" class="hover:bg-[var(--surface-2)]">
            <td class="px-4 py-2.5 text-[var(--text-primary)]">{{ actionLabel(l.action) }}</td>
            <td class="px-4 py-2.5 text-[var(--text-secondary)]">{{ l.actor_name ?? l.actor_id }}</td>
            <td class="px-4 py-2.5 text-[var(--text-secondary)]">{{ l.target ?? "-" }}</td>
            <td class="hidden md:table-cell px-4 py-2.5 text-[var(--text-tertiary)]">{{ l.ip ?? "-" }}</td>
            <td class="px-4 py-2.5 text-xs text-[var(--text-tertiary)]">{{ timeLabel(l.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
