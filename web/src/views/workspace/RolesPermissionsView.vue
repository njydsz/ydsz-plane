<script setup lang="ts">
/**
 * 角色权限管理页 — 工作空间 RBAC 矩阵展示 + 成员角色变更。
 *
 * UI 参考：GitHub Repository Roles、GitLab Members、Notion Permissions Matrix。
 * 三部分：角色概览卡片 / 成员表格 / 权限矩阵。
 */

import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { RBAC_MATRIX, countMembersByRole, type RoleKey } from "@/api/services/rbac";
import { workspaceApi, type Member, type Workspace } from "@/api/services/workspace";
import { AppBadge } from "@/components";
import AppButton from "@/components/AppButton.vue";
import AppLoadingState from "@/components/AppLoadingState.vue";
import AppEmptyState from "@/components/AppEmptyState.vue";
import AppErrorState from "@/components/AppErrorState.vue";

const route = useRoute();
const wsId = computed(() => Number(route.params.workspaceId));

const ws = ref<Workspace | null>(null);
const members = ref<Member[]>([]);
const loading = ref(true);
const error = ref("");

// 权限矩阵视图模式：'member' 简洁视图 | 'full' 完整视图
const viewMode = ref<"member" | "full">("full");
// 折叠的资源组
const collapsedGroups = ref<Set<string>>(new Set());
// 角色变更状态
const changingMembers = ref<Set<number>>(new Set());

// --- 角色概览卡片数据 ---
const roleCards = computed(() => {
  return RBAC_MATRIX.roles.map((role) => ({
    ...role,
    count: countMembersByRole(members.value, role.key),
  }));
});

/** 简洁视图下仅显示 member 列 */
const visibleRoles = computed<RoleKey[]>(() => {
  if (viewMode.value === "member") return ["member"];
  return RBAC_MATRIX.roles.map((r) => r.key);
});

function toggleGroup(name: string) {
  if (collapsedGroups.value.has(name)) {
    collapsedGroups.value.delete(name);
  } else {
    collapsedGroups.value.add(name);
  }
}

function isGroupCollapsed(name: string): boolean {
  return collapsedGroups.value.has(name);
}

async function loadMembers() {
  loading.value = true;
  error.value = "";
  try {
    ws.value = await workspaceApi.get(wsId.value);
    members.value = await workspaceApi.listMembers(wsId.value);
  } catch (e: any) {
    error.value = e.message ?? "加载成员列表失败";
  } finally {
    loading.value = false;
  }
}

async function changeRole(member: Member, newRole: RoleKey) {
  if (member.role === newRole) return;
  changingMembers.value.add(member.id);
  try {
    await workspaceApi.updateMemberRole(wsId.value, member.id, newRole);
    // 更新本地数据
    member.role = newRole;
  } catch (e: any) {
    error.value = `角色变更失败: ${e.message ?? "未知错误"}`;
  } finally {
    changingMembers.value.delete(member.id);
  }
}

/** 判断当前登录用户是否有权变更某目标成员角色 */
function canChangeRole(member: Member): boolean {
  const role = ws.value?.role ?? "";
  // owner 可以变更 admin/member/guest
  if (role === "owner") return member.role !== "owner";
  // admin 可以变更 member/guest，不能改 owner 和 admin
  if (role === "admin") return member.role === "member" || member.role === "guest";
  return false;
}

/** 可选角色下拉 */
function availableRoles(member: Member): RoleKey[] {
  const role = ws.value?.role ?? "";
  if (role === "owner") {
    return member.role === "owner" ? (["owner"] as RoleKey[]) : (["admin", "member", "guest"] as RoleKey[]);
  }
  if (role === "admin") {
    return ["member", "guest"] as RoleKey[];
  }
  return [member.role as RoleKey];
}

onMounted(loadMembers);
</script>

<template>
  <div class="rbac">
    <!-- 页头 -->
    <header class="rbac__header">
      <div>
        <h1>角色与权限</h1>
        <p class="hint">工作空间 RBAC 权限管理 — 仅 owner / admin 可操作成员角色</p>
      </div>
      <div class="actions">
        <AppButton variant="ghost" size="sm" @click="loadMembers">刷新</AppButton>
      </div>
    </header>

    <!-- 加载态 -->
    <AppLoadingState v-if="loading" text="加载角色成员数据..." />

    <!-- 错误态 -->
    <AppErrorState v-else-if="error && members.length === 0" :message="error" @retry="loadMembers" />

    <template v-else>
      <!-- 错误提示 -->
      <div v-if="error" class="inline-error">
        <span>{{ error }}</span>
        <button class="inline-error__close" @click="error = ''">×</button>
      </div>

      <!-- 一、角色概览 -->
      <section class="section">
        <h2 class="section__title">角色概览</h2>
        <div role="list" class="role-cards">
          <div
            v-for="role in roleCards"
            :key="role.key"
            class="role-card"
            :class="`role-card--${role.key}`"
          >
            <div class="role-card__header">
              <span class="role-card__icon">{{ role.icon }}</span>
              <div>
                <div class="role-card__name">{{ role.label }}</div>
                <div class="role-card__role">{{ role.key }}</div>
              </div>
            </div>
            <div class="role-card__level">等级 {{ role.level }}</div>
            <p class="role-card__desc">{{ role.description }}</p>
            <div class="role-card__count">
              <span class="count-num">{{ role.count }}</span>
              <span class="count-label">位成员</span>
            </div>
          </div>
        </div>
      </section>

      <!-- 二、角色成员 -->
      <section class="section">
        <h2 class="section__title">角色成员</h2>

        <div v-if="members.length > 0" class="table-wrapper">
          <table class="member-table">
            <thead>
              <tr>
                <th class="col-user">成员</th>
                <th class="col-email">邮箱</th>
                <th class="col-role">当前角色</th>
                <th class="col-joined">加入时间</th>
                <th class="col-action">变更角色</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="member in members" :key="member.id">
                <td class="col-user">
                  <div class="user-cell">
                    <span class="avatar">{{ member.display_name?.[0]?.toUpperCase() ?? "?" }}</span>
                    <span class="user-name">{{ member.display_name || `用户 #${member.id}` }}</span>
                  </div>
                </td>
                <td class="col-email">
                  <span class="email-text">{{ member.email }}</span>
                </td>
                <td class="col-role">
                  <AppBadge :variant="member.role === 'owner' ? 'brand' : member.role === 'admin' ? 'warning' : member.role === 'member' ? 'success' : 'default'">
                    {{ RBAC_MATRIX.roles.find(r => r.key === member.role)?.label ?? member.role }}
                  </AppBadge>
                </td>
                <td class="col-joined">
                  <span class="time-text">{{ new Date(member.joined_at).toLocaleDateString("zh-CN") }}</span>
                </td>
                <td class="col-action">
                  <div class="role-select-wrap">
                    <select
                      v-if="canChangeRole(member)"
                      class="role-select"
                      :value="member.role"
                      :disabled="changingMembers.has(member.id)"
                      @change="(e) => changeRole(member, (e.target as HTMLSelectElement).value as RoleKey)"
                    >
                      <option
                        v-for="r in availableRoles(member)"
                        :key="r"
                        :value="r"
                      >
                        {{ RBAC_MATRIX.roles.find(rd => rd.key === r)?.label ?? r }}
                      </option>
                    </select>
                    <span v-else class="no-perm">无权限</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <AppEmptyState
          v-if="!loading && members.length === 0"
          icon="👥"
          title="暂无成员"
          description="该工作空间还没有成员"
        />
      </section>

      <!-- 三、权限矩阵 -->
      <section class="section">
        <div class="section__header">
          <h2 class="section__title">权限矩阵</h2>
          <div class="view-toggle">
            <button
              class="toggle-btn"
              :class="{ active: viewMode === 'member' }"
              @click="viewMode = 'member'"
            >
              简洁视图
            </button>
            <button
              class="toggle-btn"
              :class="{ active: viewMode === 'full' }"
              @click="viewMode = 'full'"
            >
              完整视图
            </button>
          </div>
        </div>
        <p v-if="viewMode === 'member'" class="view-hint">
          简洁视图仅展示「成员」角色的权限（普通用户可访问的功能）
        </p>

        <div class="matrix-wrapper">
          <table class="matrix-table">
            <thead>
              <tr>
                <th class="col-perm">权限</th>
                <th v-for="role in visibleRoles" :key="role" class="col-role-head">
                  <span class="role-head-content">
                    <span class="role-head-icon">{{ RBAC_MATRIX.roles.find(r => r.key === role)?.icon }}</span>
                    <span>{{ RBAC_MATRIX.roles.find(r => r.key === role)?.label }}</span>
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <template v-for="group in RBAC_MATRIX.resourceGroups" :key="group.name">
                <!-- 资源组标题行 -->
                <tr class="group-header" @click="toggleGroup(group.name)">
                  <td :colspan="visibleRoles.length + 1" class="group-cell">
                    <span class="group-toggle">{{ isGroupCollapsed(group.name) ? "▶" : "▼" }}</span>
                    <span class="group-icon">{{ group.icon }}</span>
                    <span class="group-name">{{ group.name }}</span>
                    <span class="group-count">{{ group.permissions.length }} 项</span>
                  </td>
                </tr>
                <template v-if="!isGroupCollapsed(group.name)">
                  <tr v-for="perm in group.permissions" :key="perm.key" class="perm-row">
                    <td class="col-perm">
                      <span class="perm-label">{{ perm.label }}</span>
                      <code class="perm-key">{{ perm.key }}</code>
                    </td>
                    <td
                      v-for="role in visibleRoles"
                      :key="role"
                      class="matrix-cell"
                      :class="{ allowed: perm[role] }"
                    >
                      <span class="matrix-icon">{{ perm[role] ? "✓" : "—" }}</span>
                    </td>
                  </tr>
                </template>
              </template>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.rbac__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
}

.rbac__header h1 {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 4px;
  color: var(--txt-primary);
}

.hint {
  color: var(--txt-tertiary);
  font-size: var(--text-13);
  margin: 0;
}

/* --- 错误提示 --- */
.inline-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  margin-bottom: 12px;
  border-radius: var(--radius-sm);
  background: var(--bg-danger-subtle, rgba(220, 47, 47, 0.06));
  border: 1px solid var(--border-danger-subtle);
  color: var(--txt-danger-primary);
  font-size: var(--text-13);
}

.inline-error__close {
  background: none;
  border: none;
  font-size: 18px;
  color: inherit;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

/* --- Section --- */
.section {
  margin-bottom: 32px;
}

.section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.section__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--txt-primary);
  margin: 0 0 12px;
}

.section__header .section__title {
  margin: 0;
}

.view-hint {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
  margin: 0 0 8px;
}

/* --- 视图切换 --- */
.view-toggle {
  display: inline-flex;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.toggle-btn {
  padding: 5px 14px;
  border: none;
  background: var(--bg-surface-1);
  color: var(--txt-secondary);
  font-size: var(--text-12);
  cursor: pointer;
  font-family: inherit;
}

.toggle-btn.active {
  background: var(--bg-accent-primary);
  color: var(--text-on-brand);
}

.toggle-btn:hover:not(.active) {
  background: var(--bg-surface-2);
}

/* --- 角色卡片 --- */
.role-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.role-card {
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
  transition: box-shadow 0.15s;
}

.role-card:hover {
  box-shadow: var(--shadow-card);
}

.role-card--owner { border-top: 3px solid var(--brand-500); }
.role-card--admin { border-top: 3px solid var(--warning-500); }
.role-card--member { border-top: 3px solid var(--success-500); }
.role-card--guest { border-top: 3px solid var(--border-strong); }

.role-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.role-card__icon {
  font-size: 24px;
}

.role-card__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary);
}

.role-card__role {
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  margin-top: 1px;
}

.role-card__level {
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  margin-bottom: 6px;
}

.role-card__desc {
  font-size: var(--text-12);
  color: var(--txt-secondary);
  margin: 0 0 12px;
  line-height: 1.5;
  min-height: 36px;
}

.role-card__count {
  display: flex;
  align-items: baseline;
  gap: 4px;
  padding-top: 8px;
  border-top: 1px solid var(--border-subtle);
}

.count-num {
  font-size: 24px;
  font-weight: 600;
  color: var(--txt-primary);
}

.count-label {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
}

/* --- 成员表格 --- */
.table-wrapper {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
  overflow-x: auto;
}

.member-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--text-13);
}

.member-table thead th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 500;
  color: var(--txt-tertiary);
  font-size: var(--text-12);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.member-table tbody td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: middle;
}

.member-table tbody tr:hover {
  background: var(--bg-surface-2);
}

.col-user { min-width: 180px; }
.col-email { min-width: 160px; }
.col-role { width: 100px; }
.col-joined { width: 110px; }
.col-action { width: 140px; }

.user-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--bg-accent-subtle);
  color: var(--txt-accent-primary);
  font-size: var(--text-12);
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  font-size: var(--text-13);
  color: var(--txt-primary);
}

.email-text {
  font-family: var(--font-mono);
  font-size: var(--text-12);
  color: var(--txt-tertiary);
}

.time-text {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
}

.role-select-wrap {
  display: flex;
  align-items: center;
}

.role-select {
  padding: 4px 8px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-surface-1);
  color: var(--txt-primary);
  font-size: var(--text-12);
  font-family: inherit;
  cursor: pointer;
}

.role-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.no-perm {
  font-size: var(--text-12);
  color: var(--txt-tertiary);
}

/* --- 权限矩阵 --- */
.matrix-wrapper {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface-1);
  overflow-x: auto;
  max-height: 700px;
  overflow-y: auto;
}

.matrix-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--text-13);
}

.matrix-table thead th {
  position: sticky;
  top: 0;
  background: var(--bg-surface-2);
  z-index: 1;
  padding: 10px 12px;
  text-align: left;
  font-weight: 500;
  color: var(--txt-tertiary);
  font-size: var(--text-12);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.role-head-content {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.role-head-icon {
  font-size: 14px;
}

.col-perm {
  min-width: 200px;
}

.col-role-head {
  text-align: center;
  min-width: 80px;
}

.group-header {
  cursor: pointer;
  background: var(--bg-surface-2);
  user-select: none;
}

.group-header:hover {
  background: var(--bg-layer-1-hover);
}

.group-cell {
  padding: 8px 12px;
  font-weight: 500;
  color: var(--txt-primary);
  font-size: var(--text-13);
  border-bottom: 1px solid var(--border-subtle);
}

.group-toggle {
  margin-right: 6px;
  font-size: 10px;
  color: var(--txt-tertiary);
}

.group-icon {
  margin-right: 6px;
}

.group-name {
  font-weight: 600;
}

.group-count {
  margin-left: 8px;
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  font-weight: 400;
}

.perm-row:hover {
  background: var(--bg-surface-2);
}

.perm-row td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle);
  vertical-align: middle;
}

.perm-label {
  font-size: var(--text-13);
  color: var(--txt-primary);
}

.perm-key {
  display: block;
  font-family: var(--font-mono);
  font-size: var(--text-11);
  color: var(--txt-tertiary);
  margin-top: 1px;
}

.matrix-cell {
  text-align: center;
  width: 80px;
}

.matrix-icon {
  font-size: 14px;
  font-weight: 600;
}

.matrix-cell.allowed .matrix-icon {
  color: var(--success-500);
}

.matrix-cell:not(.allowed) .matrix-icon {
  color: var(--txt-tertiary);
  opacity: 0.5;
}
</style>
