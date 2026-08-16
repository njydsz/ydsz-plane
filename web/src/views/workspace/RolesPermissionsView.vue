<script setup lang="ts">
/**
 * 角色与权限 — 权限矩阵展示 + 成员角色管理。
 * 数据来源：RBAC_MATRIX（静态矩阵）+ workspaceApi 成员接口。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  RBAC_MATRIX,
  type RoleKey,
} from "@/api/services/rbac";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const members = ref<Member[]>([]);

const ROLES: { key: RoleKey; label: string }[] = [
  { key: "owner", label: "所有者" },
  { key: "admin", label: "管理员" },
  { key: "member", label: "成员" },
  { key: "guest", label: "访客" },
];

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    members.value = await workspaceApi.listMembers(workspaceId.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function changeRole(m: Member, role: string) {
  try {
    await workspaceApi.updateMemberRole(workspaceId.value, m.id, role);
    m.role = role;
    toast.success("角色已更新");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "更新失败");
  }
}

async function removeMember(m: Member) {
  if (!confirm(`确定移除成员 ${m.display_name || m.email}？`)) return;
  try {
    await workspaceApi.removeMember(workspaceId.value, m.id);
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "移除失败");
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold tracking-tight">角色与权限</h1>

    <!-- 成员管理 -->
    <section>
      <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">成员</h2>
      <div v-if="loading" class="space-y-3">
        <AppSkeleton v-for="i in 3" :key="i" class="h-12 w-full" />
      </div>
      <AppErrorState v-else-if="error" :message="error" @retry="load" />
      <AppEmptyState v-else-if="members.length === 0" title="暂无成员" />
      <div v-else class="overflow-hidden rounded-md border border-[var(--border-subtle)]">
        <table class="w-full text-sm">
          <thead class="bg-[var(--surface-2)] text-xs uppercase tracking-wider text-[var(--text-tertiary)]">
            <tr>
              <th class="px-4 py-2 text-left font-medium">成员</th>
              <th class="px-4 py-2 text-left font-medium">邮箱</th>
              <th class="px-4 py-2 text-left font-medium">角色</th>
              <th class="px-4 py-2 text-left font-medium">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--border-subtle)]">
            <tr v-for="m in members" :key="m.id" class="hover:bg-[var(--surface-2)]">
              <td class="px-4 py-2.5 text-[var(--text-primary)]">{{ m.display_name ?? "-" }}</td>
              <td class="px-4 py-2.5 text-[var(--text-secondary)]">{{ m.email }}</td>
              <td class="px-4 py-2.5">
                <select
                  :value="m.role"
                  class="rounded-md border border-[var(--border-subtle)] px-2 py-1 text-sm"
                  @change="changeRole(m, ($event.target as HTMLSelectElement).value)"
                >
                  <option v-for="r in ROLES" :key="r.key" :value="r.key">{{ r.label }}</option>
                </select>
              </td>
              <td class="px-4 py-2.5">
                <button class="text-xs text-[var(--danger, #ef4444)]" @click="removeMember(m)">移除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- 权限矩阵 -->
    <section>
      <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">权限矩阵</h2>
      <div class="space-y-4">
        <div
          v-for="group in RBAC_MATRIX.resourceGroups"
          :key="group.name"
          class="rounded-md border border-[var(--border-subtle)] overflow-hidden"
        >
          <div class="bg-[var(--surface-2)] px-4 py-2 text-sm font-medium text-[var(--text-primary)]">
            {{ group.icon }} {{ group.name }}
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-xs text-[var(--text-tertiary)]">
                <th class="px-4 py-1.5 text-left font-medium">权限</th>
                <th v-for="r in ROLES" :key="r.key" class="px-2 py-1.5 text-center font-medium">
                  {{ r.label }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--border-subtle)]">
              <tr v-for="p in group.permissions" :key="p.key">
                <td class="px-4 py-1.5 text-[var(--text-secondary)]">{{ p.label }}</td>
                <td
                  v-for="r in ROLES"
                  :key="r.key"
                  class="px-2 py-1.5 text-center"
                >
                  <span :class="p[r.key] ? 'text-[var(--brand-600)]' : 'text-[var(--text-tertiary)]'">
                    {{ p[r.key] ? "✓" : "—" }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>
