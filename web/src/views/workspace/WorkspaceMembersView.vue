<script setup lang="ts">
/**
 * 工作空间成员管理页 — 成员列表 + 角色变更 + 移除成员。
 *
 *  数据来源：workspaceApi.listMembers / updateMemberRole / removeMember。
 *  权限矩阵展示另有 RolesPermissionsView.vue 负责，本页不再重复。
 *
 *  权限策略：
 *    - 仅当前用户的角色为 owner/admin 时才显示操作列
 *    - 不能修改/移除自己（避免误操作把自己降级掉）
 *    - 不能修改/移除 owner（owner 在团队中通常只有 1 个，交给后端校验兜底）
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import { workspaceApi, type Member } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { formatRelativeTime } from "@/lib/formatTime";
import { toast } from "@/lib/toast";

const route = useRoute();
const auth = useAuthStore();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const currentUserId = computed(() => auth.user?.id ?? 0);

const loading = ref(true);
const error = ref("");
const members = ref<Member[]>([]);

/** 当前打开角色下拉的成员 ID（null = 全部关闭），点击空白处关闭 */
const openRoleMenuFor = ref<number | null>(null);
/** 正在变更/移除中的成员 ID 集合（用于 disable 按钮 + 显示 loading 文案） */
const pendingIds = ref<Set<number>>(new Set());

/** 角色枚举 + 中文 label（与 RolesPermissionsView 对齐） */
const ROLES = [
  { key: "owner", label: "所有者" },
  { key: "admin", label: "管理员" },
  { key: "member", label: "成员" },
  { key: "guest", label: "访客" },
] as const;

/** 当前用户在该工作空间的成员条目 */
const me = computed<Member | undefined>(() =>
  members.value.find((m) => m.id === currentUserId.value),
);

/** 当前用户是否具备成员管理权限（owner/admin） */
const canManageMembers = computed(
  () => me.value?.role === "owner" || me.value?.role === "admin",
);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
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

function toggleRoleMenu(memberId: number) {
  openRoleMenuFor.value = openRoleMenuFor.value === memberId ? null : memberId;
}

function closeRoleMenu() {
  openRoleMenuFor.value = null;
}

/** 是否允许对某成员执行变更/移除（不能动自己、不能动 owner） */
function canModify(m: Member): boolean {
  if (!canManageMembers.value) return false;
  if (m.id === currentUserId.value) return false;
  if (m.role === "owner") return false;
  return true;
}

function addPending(id: number) {
  const next = new Set(pendingIds.value);
  next.add(id);
  pendingIds.value = next;
}

function clearPending(id: number) {
  const next = new Set(pendingIds.value);
  next.delete(id);
  pendingIds.value = next;
}

function roleLabel(role: string): string {
  return ROLES.find((r) => r.key === role)?.label ?? role;
}

/** 头像首字符 — display_name 优先，回退 email，去空白后取首字符并大写 */
function avatarChar(m: Member): string {
  const seed = (m.display_name || m.email || "").trim();
  return (seed.charAt(0) || "?").toUpperCase();
}

/** 切换角色 — optimistic update + rollback on error */
async function changeRole(m: Member, role: string) {
  if (m.role === role) {
    closeRoleMenu();
    return;
  }
  const previous = m.role;
  m.role = role; // 先更新本地状态
  closeRoleMenu();
  addPending(m.id);
  try {
    await workspaceApi.updateMemberRole(workspaceId.value, m.id, role);
    toast.success(`${m.display_name || m.email} 已设为「${roleLabel(role)}」`);
  } catch (err: unknown) {
    m.role = previous; // 回滚
    toast.error(err instanceof Error ? err.message : "更新失败");
  } finally {
    clearPending(m.id);
  }
}

/** 移除成员 — 按规范走 confirm + removeMember + 重新拉取以同步服务端事实 */
async function removeMember(m: Member) {
  if (!confirm(`确定移除成员 ${m.display_name || m.email}？此操作无法撤销。`)) return;
  addPending(m.id);
  try {
    await workspaceApi.removeMember(workspaceId.value, m.id);
    toast.success(`已移除 ${m.display_name || m.email}`);
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "移除失败");
  } finally {
    clearPending(m.id);
  }
}

onMounted(load);
watch(workspaceId, () => {
  openRoleMenuFor.value = null;
  load();
});

/** 关闭角色下拉的全局监听 — 仅当有菜单打开时才挂载，性能 OK */
function onDocumentClick(e: MouseEvent) {
  if (openRoleMenuFor.value === null) return;
  const target = e.target as HTMLElement | null;
  if (target && target.closest("[data-role-menu]")) return;
  openRoleMenuFor.value = null;
}
if (typeof document !== "undefined") {
  document.addEventListener("click", onDocumentClick);
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">成员管理</h1>
      <p class="mt-1 text-sm text-[var(--text-secondary)]">
        管理本工作空间的成员及其角色。权限矩阵可在
        <RouterLink
          :to="{ name: 'workspace-rbac', params: { workspaceId } }"
          class="text-[var(--brand-600)] hover:underline"
        >角色与权限</RouterLink>
        页面查看。
      </p>
    </div>

    <section>
      <!-- 三态：loading / error / empty / data -->
      <div v-if="loading" class="space-y-3">
        <AppSkeleton v-for="i in 4" :key="i" class="h-14 w-full" />
      </div>
      <AppErrorState v-else-if="error" :message="error" @retry="load" />
      <AppEmptyState
        v-else-if="members.length === 0"
        title="暂无成员"
        description="还没有任何人加入这个工作空间。"
      />
      <div
        v-else
        class="overflow-hidden rounded-md border border-[var(--border-subtle)]"
      >
        <table class="w-full text-sm">
          <thead class="bg-[var(--surface-2)] text-xs uppercase tracking-wider text-[var(--text-tertiary)]">
            <tr>
              <th class="px-4 py-2 text-left font-medium">成员</th>
              <th class="px-4 py-2 text-left font-medium">邮箱</th>
              <th class="px-4 py-2 text-left font-medium">角色</th>
              <th class="px-4 py-2 text-left font-medium">加入时间</th>
              <th v-if="canManageMembers" class="px-4 py-2 text-right font-medium">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--border-subtle)]">
            <tr
              v-for="m in members"
              :key="m.id"
              class="hover:bg-[var(--surface-2)]"
            >
              <!-- 头像 + 显示名 -->
              <td class="px-4 py-2.5">
                <div class="flex items-center gap-3">
                  <span class="member-avatar" :aria-label="m.display_name || m.email">
                    <img
                      v-if="m.avatar_url"
                      :src="m.avatar_url"
                      :alt="m.display_name || m.email"
                      class="member-avatar__img"
                    />
                    <span v-else class="member-avatar__fallback">{{ avatarChar(m) }}</span>
                  </span>
                  <div class="min-w-0">
                    <div class="truncate text-[var(--text-primary)]">
                      {{ m.display_name || "—" }}
                      <span
                        v-if="m.id === currentUserId"
                        class="ml-1 text-xs text-[var(--text-tertiary)]"
                      >（你）</span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- 邮箱 -->
              <td class="px-4 py-2.5 text-[var(--text-secondary)]">
                {{ m.email }}
              </td>

              <!-- 角色徽章 + 可点击的下拉触发器 -->
              <td class="px-4 py-2.5">
                <div class="flex items-center gap-1">
                  <span class="role-badge" :class="`role-badge--${m.role}`">
                    {{ roleLabel(m.role) }}
                  </span>
                  <button
                    v-if="canModify(m)"
                    type="button"
                    class="role-trigger"
                    :disabled="pendingIds.has(m.id)"
                    data-role-menu
                    :aria-expanded="openRoleMenuFor === m.id"
                    aria-haspopup="menu"
                    title="切换角色"
                    @click.stop="toggleRoleMenu(m.id)"
                  >
                    <span aria-hidden="true">▾</span>
                  </button>

                  <!-- 角色下拉菜单 -->
                  <div
                    v-if="canModify(m) && openRoleMenuFor === m.id"
                    class="role-menu"
                    role="menu"
                    data-role-menu
                    @click.stop
                  >
                    <button
                      v-for="r in ROLES.filter((x) => x.key !== 'owner')"
                      :key="r.key"
                      type="button"
                      role="menuitem"
                      class="role-menu__item"
                      :class="{ 'role-menu__item--active': m.role === r.key }"
                      :disabled="pendingIds.has(m.id)"
                      @click="changeRole(m, r.key)"
                    >
                      <span>{{ r.label }}</span>
                      <span v-if="m.role === r.key" aria-hidden="true">✓</span>
                    </button>
                  </div>
                </div>
              </td>

              <!-- 加入时间 -->
              <td
                class="px-4 py-2.5 text-[var(--text-tertiary)]"
                :title="m.joined_at"
              >
                {{ formatRelativeTime(m.joined_at) }}
              </td>

              <!-- 操作 -->
              <td v-if="canManageMembers" class="px-4 py-2.5 text-right">
                <button
                  v-if="canModify(m)"
                  type="button"
                  class="member-action-danger"
                  :disabled="pendingIds.has(m.id)"
                  @click="removeMember(m)"
                >
                  {{ pendingIds.has(m.id) ? "处理中…" : "移除" }}
                </button>
                <span v-else class="text-xs text-[var(--text-tertiary)]">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* -------------------------------------------------------------------------
 *  头像 — 圆形 fallback 背景 var(--brand-500)，白字
 * ------------------------------------------------------------------------- */
.member-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--brand-500);
  color: #fff;
  font-weight: 600;
  font-size: 13px;
  user-select: none;
}
.member-avatar__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.member-avatar__fallback {
  line-height: 1;
}

/* -------------------------------------------------------------------------
 *  角色徽章 — 按角色取色
 *   owner: 紫色    admin: 蓝色    member: 灰色    guest: 浅灰
 * ------------------------------------------------------------------------- */
.role-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.4;
  white-space: nowrap;
}
.role-badge--owner {
  background: var(--label-purple-bg);
  color: var(--label-purple-text);
}
.role-badge--admin {
  background: rgba(23, 190, 233, 0.12);
  color: var(--info-500);
}
.role-badge--member {
  background: var(--surface-3);
  color: var(--text-secondary);
}
.role-badge--guest {
  background: var(--surface-2);
  color: var(--text-tertiary);
}

/* 徽章旁的下拉触发器（小箭头按钮） */
.role-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  color: var(--text-tertiary);
  font-size: 10px;
  line-height: 1;
  transition: background-color 120ms;
}
.role-trigger:hover:not(:disabled) {
  background: var(--surface-2);
  color: var(--text-primary);
}
.role-trigger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* -------------------------------------------------------------------------
 *  角色下拉菜单 — 绝对定位浮层（相对于父 td 内的 inline-flex 容器）
 * ------------------------------------------------------------------------- */
.role-menu {
  position: absolute;
  z-index: 30;
  margin-top: 4px;
  min-width: 120px;
  background: var(--surface-1);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  padding: 4px;
}
.role-menu__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-primary);
  text-align: left;
  transition: background-color 120ms;
}
.role-menu__item:hover:not(:disabled) {
  background: var(--surface-2);
}
.role-menu__item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.role-menu__item--active {
  color: var(--brand-600);
  font-weight: 500;
}

/* -------------------------------------------------------------------------
 *  操作列 — 移除按钮
 * ------------------------------------------------------------------------- */
.member-action-danger {
  font-size: 12px;
  color: var(--danger-500);
  background: transparent;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 120ms;
}
.member-action-danger:hover:not(:disabled) {
  background: rgba(220, 47, 47, 0.08);
}
.member-action-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
