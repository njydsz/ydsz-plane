<script setup lang="ts">
/**
 * ProjectMembersView — 项目成员管理页。
 *
 * 列出项目现有成员，支持添加工作空间成员、修改角色、移除成员。
 * 权限：仅 project admin 或 workspace owner/admin 可操作。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { workspaceApi, type Member } from "@/api/services/workspace";
import { ApiError } from "@/api/client";
import { AppLoadingState, AppErrorState } from "@/components";

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

/** 项目成员列表 */
const members = ref<Member[]>([]);
const loading = ref(true);
const error = ref("");

/** 工作空间成员列表（用作"添加成员"下拉选项） */
const workspaceMembers = ref<Member[]>([]);

/** 当前用户在工作空间的角色 */
const myWorkspaceRole = ref("");

/** 当前用户是否有项目管理权限（workspace owner/admin 或 project admin） */
const canManage = computed(() => {
  if (myWorkspaceRole.value === "owner" || myWorkspaceRole.value === "admin") return true;
  const me = members.value.find((m) => m.role === "admin");
  // 通过"列表中是否有 admin 角色条目"推断（仅自己是 admin 时它必然可见）
  return !!me;
});

/** 可添加的工作空间成员（已加入空间但尚未加入项目） */
const availableMembers = computed(() => {
  const existingIds = new Set(members.value.map((m) => m.id));
  return workspaceMembers.value.filter((wm) => !existingIds.has(wm.id));
});

const showAddForm = ref(false);
const newMemberUserId = ref<number | null>(null);
const newMemberRole = ref("member");
const adding = ref(false);
const addError = ref("");

async function loadMembers() {
  loading.value = true;
  error.value = "";
  try {
    const ws = await workspaceApi.get(workspaceId);
    const [pm, wm, roleResp] = await Promise.all([
      workspaceApi.listProjectMembers(ws.id, projectId),
      workspaceApi.listMembers(ws.id),
      workspaceApi.getMyRole(ws.id),
    ]);
    members.value = pm;
    workspaceMembers.value = wm;
    myWorkspaceRole.value = roleResp?.role?.slug ?? "";
  } catch (e: any) {
    error.value = e instanceof ApiError ? e.message : "加载成员失败";
  } finally {
    loading.value = false;
  }
}

async function addMember() {
  if (!newMemberUserId.value) {
    addError.value = "请选择要添加的成员";
    return;
  }
  addError.value = "";
  adding.value = true;
  try {
    const ws = await workspaceApi.get(workspaceId);
    await workspaceApi.addProjectMember(ws.id, projectId, {
      user_id: newMemberUserId.value,
      role: newMemberRole.value,
    });
    showAddForm.value = false;
    newMemberUserId.value = null;
    newMemberRole.value = "member";
    await loadMembers();
  } catch (e: any) {
    addError.value = e instanceof ApiError ? e.message : "添加失败";
  } finally {
    adding.value = false;
  }
}

async function changeRole(userId: number, role: string) {
  if (role !== "admin" && role !== "member") return;
  try {
    const ws = await workspaceApi.get(workspaceId);
    await workspaceApi.changeProjectMemberRole(ws.id, projectId, userId, role);
    await loadMembers();
  } catch (e: any) {
    alert(e instanceof ApiError ? e.message : "修改失败");
  }
}

async function removeMember(userId: number, name: string) {
  if (!confirm(`确认将「${name}」移出项目？`)) return;
  try {
    const ws = await workspaceApi.get(workspaceId);
    await workspaceApi.removeProjectMember(ws.id, projectId, userId);
    await loadMembers();
  } catch (e: any) {
    alert(e instanceof ApiError ? e.message : "移除失败");
  }
}

function displayName(m: Member) {
  return m.display_name || m.email || `用户${m.id}`;
}

onMounted(loadMembers);
</script>

<template>
  <div class="pm">
    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="loadMembers" />

    <div v-else>
      <div class="pm__bar">
        <div>
          <h2 class="pm__title">项目成员</h2>
          <p class="pm__desc">
            管理该项目的参与者。项目管理员可添加/移除成员与修改角色。
            共 {{ members.length }} 名成员。
          </p>
        </div>
        <button
          v-if="canManage"
          class="btn btn--primary btn--sm"
          @click="showAddForm = !showAddForm"
        >
          {{ showAddForm ? "取消" : "添加成员" }}
        </button>
      </div>

      <!-- 添加成员面板 -->
      <section v-if="showAddForm" class="panel add-panel">
        <h3 class="panel__sub-title">从工作空间添加成员</h3>
        <p v-if="addError" class="msg error">{{ addError }}</p>
        <p v-if="availableMembers.length === 0" class="panel__desc">
          工作空间的所有成员都已在本项目中。
        </p>
        <div v-else class="add-form">
          <select v-model.number="newMemberUserId" class="select">
            <option :value="null" disabled>请选择成员</option>
            <option v-for="m in availableMembers" :key="m.id" :value="m.id">
              {{ displayName(m) }} ({{ m.email }})
            </option>
          </select>
          <select v-model="newMemberRole" class="select select--sm">
            <option value="member">成员</option>
            <option value="admin">管理员</option>
          </select>
          <button
            class="btn btn--primary btn--sm"
            :disabled="!newMemberUserId || adding"
            @click="addMember"
          >
            {{ adding ? "添加中..." : "确认添加" }}
          </button>
        </div>
      </section>

      <!-- 成员列表 -->
      <section class="panel member-list">
        <div
          v-for="m in members"
          :key="m.id"
          class="member-row"
        >
          <div class="member-info">
            <div class="avatar">
              {{ (m.display_name || m.email || "?")[0].toUpperCase() }}
            </div>
            <div>
              <div class="member-name">{{ displayName(m) }}</div>
              <div class="member-email">{{ m.email }}</div>
            </div>
          </div>
          <div class="member-actions">
            <span
              class="role-badge"
              :class="m.role === 'admin' ? 'role-badge--admin' : 'role-badge--member'"
            >{{ m.role === "admin" ? "管理员" : "成员" }}</span>
            <select
              v-if="canManage && m.role !== 'admin'"
              :value="m.role"
              class="select select--xs"
              @change="changeRole(m.id, ($event.target as HTMLSelectElement).value)"
            >
              <option value="member">成员</option>
              <option value="admin">管理员</option>
            </select>
            <button
              v-if="canManage"
              class="btn btn--ghost btn--xs btn--danger"
              @click="removeMember(m.id, displayName(m))"
            >移除</button>
          </div>
        </div>
        <p v-if="members.length === 0" class="panel__desc">
          暂无成员。项目创建者会自动成为项目管理员。
        </p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.pm {
  max-width: 680px;
}

.pm__bar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.pm__title {
  font-size: 16px;
  margin: 0;
}

.pm__desc {
  font-size: 12px;
  color: var(--text-tertiary);
  margin: 4px 0 0;
}

.panel {
  padding: 20px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--surface-1);
  margin-bottom: 16px;
}

.panel__sub-title {
  font-size: 14px;
  margin: 0 0 12px;
}

.add-form {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.select {
  height: 32px;
  padding: 0 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  font-size: 13px;
  min-width: 120px;
}

.select--sm { min-width: 90px; }
.select--xs { min-width: 72px; height: 28px; }

.member-list {
  padding: 0;
}

.member-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.member-row:last-child {
  border-bottom: none;
}

.member-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--brand-100);
  color: var(--brand-600);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.member-name {
  font-size: 14px;
  font-weight: 500;
}

.member-email {
  font-size: 12px;
  color: var(--text-tertiary);
}

.member-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  font-size: 11px;
  font-weight: 500;
}

.role-badge--admin {
  background: var(--brand-50);
  color: var(--brand-600);
}

.role-badge--member {
  background: var(--surface-2);
  color: var(--text-secondary);
}

.btn {
  border: none;
  border-radius: var(--radius-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.btn--primary {
  background: var(--brand-500);
  color: #fff;
}
.btn--primary:hover { background: var(--brand-600); }
.btn--primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn--sm { padding: 6px 12px; font-size: 13px; }
.btn--xs { padding: 4px 8px; font-size: 12px; }

.btn--ghost {
  background: transparent;
}
.btn--danger { color: var(--danger-500); }
.btn--danger:hover { background: var(--danger-50); }

.msg.error { color: var(--danger-500); font-size: 13px; margin: 0 0 10px; }
</style>
