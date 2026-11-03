<script setup lang="ts">
/**
 * 角色与权限 — 权限矩阵展示（动态加载）+ 成员角色管理。
 *
 * 数据来源优先级：
 *   1. 后端 GET /workspaces/:wsId/rbac/matrix（真实授权状态）
 *   2. 静态 RBAC_MATRIX（后端不可用时的 fallback）
 *
 * 权限编辑：PUT /workspaces/:wsId/rbac/roles/:roleSlug/permissions
 *          全量覆盖语义，仅系统管理员可调用。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  RBAC_MATRIX,
  rbacApi,
  type PermissionMatrixResponse,
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

/** 后端矩阵数据；null 表示回退到静态矩阵。 */
const matrix = ref<PermissionMatrixResponse | null>(null);
/** 当前是否处于编辑模式。 */
const editing = ref(false);
/** 编辑中的权限授权：role_slug → Set<permission_code>。 */
const draftGrants = ref<Record<string, Set<string>>>({});
const saving = ref(false);

/** 角色列表：后端返回优先，否则用静态定义。 */
const roles = computed(() => {
  if (matrix.value) {
    return matrix.value.roles.map((r) => ({
      key: r.slug as RoleKey,
      label: r.name,
      icon: r.icon ?? "👤",
      level: r.level,
      description: r.description ?? "",
      isSystem: r.is_system,
    }));
  }
  return RBAC_MATRIX.roles;
});

/** 权限码分组：后端返回的权限码按前缀分组。 */
const groupedPermissions = computed(() => {
  if (!matrix.value) {
    return RBAC_MATRIX.resourceGroups.map((g) => ({
      name: g.name,
      icon: g.icon,
      permissions: g.permissions.map((p) => ({
        code: p.key,
        label: p.label,
      })),
    }));
  }

  const groups = new Map<string, { name: string; icon: string; codes: Set<string> }>();
  for (const code of matrix.value.permissions) {
    const [namespace] = code.split(":");
    const label = namespace.charAt(0).toUpperCase() + namespace.slice(1);
    if (!groups.has(namespace)) {
      groups.set(namespace, { name: label, icon: "🔹", codes: new Set() });
    }
    groups.get(namespace)!.codes.add(code);
  }

  return Array.from(groups.entries()).map(([ns, g]) => ({
    name: g.name,
    icon: g.icon,
    permissions: Array.from(g.codes).map((code) => ({
      code,
      label: code,
    })),
    namespace: ns,
  }));
});

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [memberList, matrixData] = await Promise.all([
      workspaceApi.listMembers(workspaceId.value),
      rbacApi.getPermissionMatrix(workspaceId.value).catch(() => null),
    ]);
    members.value = memberList;
    matrix.value = matrixData;
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

/** 进入编辑模式：把当前授权复制到草稿。 */
function startEditing() {
  if (!matrix.value) {
    toast.error("权限矩阵未加载，无法编辑");
    return;
  }
  const draft: Record<string, Set<string>> = {};
  for (const role of matrix.value.roles) {
    draft[role.slug] = new Set(Object.keys(matrix.value.grants[role.slug] ?? {}));
  }
  draftGrants.value = draft;
  editing.value = true;
}

function cancelEditing() {
  editing.value = false;
  draftGrants.value = {};
}

function togglePermission(roleSlug: string, code: string) {
  const set = draftGrants.value[roleSlug];
  if (!set) return;
  if (set.has(code)) {
    set.delete(code);
  } else {
    set.add(code);
  }
}

function isChecked(roleSlug: string, code: string): boolean {
  return draftGrants.value[roleSlug]?.has(code) ?? false;
}

function checkedCount(roleSlug: string): number {
  return draftGrants.value[roleSlug]?.size ?? 0;
}

async function savePermissions() {
  if (!matrix.value) return;
  saving.value = true;
  try {
    // 逐角色覆盖：仅提交有变化的角色，减少无谓写入。
    for (const role of matrix.value.roles) {
      const next = Array.from(draftGrants.value[role.slug] ?? []);
      const current = Object.keys(matrix.value.grants[role.slug] ?? {});
      const changed =
        next.length !== current.length || next.some((p) => !current.includes(p));
      if (!changed) continue;

      // eslint-disable-next-line no-await-in-loop -- 顺序写入避免并发覆盖冲突
      const updated = await rbacApi.replaceRolePermissions(
        workspaceId.value,
        role.slug,
        next,
      );
      matrix.value = updated;
    }
    editing.value = false;
    draftGrants.value = {};
    toast.success("权限已保存");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "保存失败");
  } finally {
    saving.value = false;
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
                  <option v-for="r in roles" :key="r.key" :value="r.key">{{ r.label }}</option>
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
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">
          权限矩阵
          <span v-if="matrix" class="ml-2 text-xs font-normal text-[var(--text-tertiary)]">
            来自后端 · {{ matrix.permissions.length }} 项权限
          </span>
          <span v-else class="ml-2 text-xs font-normal text-[var(--text-tertiary)]">
            静态矩阵（后端不可用）
          </span>
        </h2>
        <div class="flex items-center gap-2">
          <template v-if="!editing">
            <button
              class="rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-xs hover:bg-[var(--surface-2)]"
              @click="startEditing"
            >
              编辑权限
            </button>
          </template>
          <template v-else>
            <button
              class="rounded-md px-3 py-1.5 text-xs hover:bg-[var(--surface-2)]"
              :disabled="saving"
              @click="cancelEditing"
            >
              取消
            </button>
            <button
              class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-xs text-white disabled:opacity-50"
              :disabled="saving"
              @click="savePermissions"
            >
              {{ saving ? "保存中…" : "保存" }}
            </button>
          </template>
        </div>
      </div>

      <div class="space-y-4">
        <div
          v-for="group in groupedPermissions"
          :key="group.name"
          class="overflow-hidden rounded-md border border-[var(--border-subtle)]"
        >
          <div class="bg-[var(--surface-2)] px-4 py-2 text-sm font-medium text-[var(--text-primary)]">
            {{ group.icon }} {{ group.name }}
          </div>
          <table class="w-full text-sm">
            <thead>
              <tr class="text-xs text-[var(--text-tertiary)]">
                <th class="px-4 py-1.5 text-left font-medium">权限</th>
                <th v-for="r in roles" :key="r.key" class="px-2 py-1.5 text-center font-medium">
                  {{ r.label }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--border-subtle)]">
              <tr v-for="p in group.permissions" :key="p.code">
                <td class="px-4 py-1.5 text-[var(--text-secondary)]">{{ p.label }}</td>
                <td
                  v-for="r in roles"
                  :key="r.key"
                  class="px-2 py-1.5 text-center"
                >
                  <input
                    v-if="editing"
                    type="checkbox"
                    class="h-3.5 w-3.5 cursor-pointer accent-[var(--brand-600)]"
                    :checked="isChecked(r.key, p.code)"
                    @change="togglePermission(r.key, p.code)"
                  />
                  <span
                    v-else
                    :class="
                      matrix?.grants[r.key]?.[p.code] !== undefined
                        ? 'text-[var(--brand-600)]'
                        : 'text-[var(--text-tertiary)]'
                    "
                  >
                    {{ matrix?.grants[r.key]?.[p.code] !== undefined ? "✓" : "—" }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 编辑模式下的汇总 -->
      <div v-if="editing" class="mt-4 rounded-md border border-[var(--border-subtle)] p-3 text-xs text-[var(--text-secondary)]">
        <div v-for="r in roles" :key="r.key" class="mr-4 inline-block">
          {{ r.label }}：{{ checkedCount(r.key) }} 项
        </div>
      </div>
    </section>
  </div>
</template>
