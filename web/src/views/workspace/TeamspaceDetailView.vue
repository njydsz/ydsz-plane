<script setup lang="ts">
/**
 * TeamspaceDetailView — 团队空间详情。
 *
 * 展示：
 *  - 基本信息（名称/描述/负责人/颜色）
 *  - 进度汇总（完成率、逾期、进行中、项目数、活跃迭代）
 *  - 成员工作负载
 *  - 成员管理（添加/移除）
 *  - 关联项目管理（添加/移除）
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  teamspaceApi,
  type Teamspace,
  type TeamspaceMember,
  type TeamspaceProgress,
  type TeamspaceMemberWorkload,
} from "@/api/services/teamspace";
import { projectApi } from "@/api/services/project";
import { workspaceApi, type Member } from "@/api/services/workspace";
import { AppBadge, AppEmptyState, AppErrorState, AppLoadingState, AppPageContainer } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const router = useRouter();

const wsId = computed(() => Number(route.params.workspaceId));
const teamspaceId = computed(() => Number(route.params.id));

const teamspace = ref<Teamspace | null>(null);
const progress = ref<TeamspaceProgress | null>(null);
const members = ref<TeamspaceMember[]>([]);
const workloads = ref<TeamspaceMemberWorkload[]>([]);
const linkedProjectIds = ref<number[]>([]);
const workspaceMembers = ref<Member[]>([]);
const projects = ref<Array<{ id: number; name: string }>>([]);

const loading = ref(true);
const error = ref("");

const showAddMember = ref(false);
const showAddProject = ref(false);
const selectedUserId = ref<number>(0);
const selectedProjectId = ref<number>(0);

async function load() {
  if (!wsId.value || !teamspaceId.value) return;
  loading.value = true;
  error.value = "";
  try {
    const [ts, p, m, w, projIds] = await Promise.all([
      teamspaceApi.get(wsId.value, teamspaceId.value),
      teamspaceApi.progress(wsId.value, teamspaceId.value),
      teamspaceApi.listMembers(wsId.value, teamspaceId.value),
      teamspaceApi.workloads(wsId.value, teamspaceId.value),
      teamspaceApi.listProjects(wsId.value, teamspaceId.value),
    ]);
    teamspace.value = ts as unknown as Teamspace;
    progress.value = p;
    members.value = ((m as { items?: TeamspaceMember[] }).items ?? []) as TeamspaceMember[];
    workloads.value = ((w as { items?: TeamspaceMemberWorkload[] }).items ?? []) as TeamspaceMemberWorkload[];
    linkedProjectIds.value = ((projIds as { items?: number[] }).items ?? []) as number[];

    const [wm, pj] = await Promise.all([
      workspaceApi.listMembers(wsId.value),
      projectApi.listProjects(wsId.value),
    ]);
    workspaceMembers.value = (wm as unknown as Member[]) ?? [];
    projects.value = ((pj as { items?: Array<{ id: number; name: string }> }).items ?? []) as Array<{ id: number; name: string }>;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载团队空间详情失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch(() => teamspaceId.value, load);

/* ---- 成员操作 ---- */

async function addMember() {
  if (!selectedUserId.value) {
    toast.error("请选择成员");
    return;
  }
  try {
    await teamspaceApi.addMember(wsId.value, teamspaceId.value, selectedUserId.value);
    toast.success("成员已添加");
    showAddMember.value = false;
    selectedUserId.value = 0;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "添加成员失败");
  }
}

async function removeMember(userId: number) {
  try {
    await teamspaceApi.removeMember(wsId.value, teamspaceId.value, userId);
    toast.success("成员已移除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "移除成员失败");
  }
}

/* ---- 项目操作 ---- */

async function addProject() {
  if (!selectedProjectId.value) {
    toast.error("请选择项目");
    return;
  }
  try {
    await teamspaceApi.addProject(wsId.value, teamspaceId.value, selectedProjectId.value);
    toast.success("项目已关联");
    showAddProject.value = false;
    selectedProjectId.value = 0;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "关联项目失败");
  }
}

async function removeProject(projectId: number) {
  try {
    await teamspaceApi.removeProject(wsId.value, teamspaceId.value, projectId);
    toast.success("项目已解除关联");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "解除关联失败");
  }
}

/* ---- 工具 ---- */

function memberName(userId: number): string {
  return workspaceMembers.value.find((m) => m.id === userId)?.display_name
    || workspaceMembers.value.find((m) => m.id === userId)?.email
    || `用户 #${userId}`;
}

function projectName(projectId: number): string {
  return projects.value.find((p) => p.id === projectId)?.name || `项目 #${projectId}`;
}

function workloadOf(userId: number): TeamspaceMemberWorkload | undefined {
  return workloads.value.find((w) => w.user_id === userId);
}

function availableMembers(): Member[] {
  const ids = new Set(members.value.map((m) => m.user_id));
  return workspaceMembers.value.filter((m) => !ids.has(m.id));
}

function availableProjects() {
  const ids = new Set(linkedProjectIds.value);
  return projects.value.filter((p) => !ids.has(p.id));
}

function back() {
  router.push(`/${wsId.value}/teamspaces`);
}
</script>

<template>
  
  <AppPageContainer max="1200px">
    <div class="ts-detail">
    <!-- 头部 -->
    <div class="ts-detail__header">
      <button class="ts-detail__back" @click="back">← 返回</button>
      <div class="ts-detail__head-main">
        <div class="ts-detail__title-row">
          <span
            class="ts-detail__dot"
            :style="{ background: teamspace?.color || 'var(--brand-500)' }"
          />
          <h1 class="ts-detail__title">{{ teamspace?.name || "团队空间" }}</h1>
          <AppBadge v-if="teamspace?.archived_at" variant="default">已归档</AppBadge>
        </div>
        <p v-if="teamspace?.description" class="ts-detail__desc">
          {{ teamspace.description }}
        </p>
      </div>
    </div>

    <AppLoadingState v-if="loading" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState
      v-else-if="!teamspace"
      title="团队空间不存在"
      description="该团队空间可能已被删除"
    />

    <template v-else>
      <!-- 进度汇总 -->
      <section class="ts-detail__section">
        <h2 class="ts-detail__section-title">进度汇总</h2>
        <div class="ts-detail__metrics">
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.member_count ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">成员</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.project_count ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">项目</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.total_work_items ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">工作项</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.completed_work_items ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">已完成</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.in_progress_work_items ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">进行中</span>
          </div>
          <div class="ts-detail__metric">
            <span
              class="ts-detail__metric-value"
              :class="{ 'ts-detail__metric-value--danger': (progress?.overdue_work_items ?? 0) > 0 }"
            >
              {{ progress?.overdue_work_items ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">逾期</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ progress?.active_cycles ?? 0 }}
            </span>
            <span class="ts-detail__metric-label">活跃迭代</span>
          </div>
          <div class="ts-detail__metric">
            <span class="ts-detail__metric-value">
              {{ Math.round(progress?.completion_rate ?? 0) }}%
            </span>
            <span class="ts-detail__metric-label">完成率</span>
          </div>
        </div>
      </section>

      <div class="ts-detail__columns">
        <!-- 成员 -->
        <section class="ts-detail__section">
          <div class="ts-detail__section-head">
            <h2 class="ts-detail__section-title">成员</h2>
            <button class="ts-detail__action" @click="showAddMember = !showAddMember">
              {{ showAddMember ? "取消" : "添加成员" }}
            </button>
          </div>

          <div v-if="showAddMember" class="ts-detail__add">
            <select v-model="selectedUserId" class="ts-detail__select">
              <option :value="0" disabled>选择成员</option>
              <option v-for="m in availableMembers()" :key="m.id" :value="m.id">
                {{ m.display_name || m.email }}
              </option>
            </select>
            <button
              class="ts-detail__primary"
              :disabled="!selectedUserId"
              @click="addMember"
            >
              添加
            </button>
          </div>

          <div class="ts-detail__list">
            <div
              v-for="m in members"
              :key="m.user_id"
              class="ts-detail__list-item"
            >
              <div class="ts-detail__list-item-main">
                <span class="ts-detail__list-item-name">{{ memberName(m.user_id) }}</span>
                <AppBadge v-if="m.role === 'lead'" variant="brand">负责人</AppBadge>
              </div>
              <span class="ts-detail__list-item-meta">
                {{ workloadOf(m.user_id)?.total_work_items ?? 0 }} 工作项 ·
                {{ Math.round(workloadOf(m.user_id)?.completion_rate ?? 0) }}%
              </span>
              <button
                class="ts-detail__link ts-detail__link--danger"
                @click="removeMember(m.user_id)"
              >
                移除
              </button>
            </div>
            <div v-if="!members.length" class="ts-detail__empty">暂无成员</div>
          </div>
        </section>

        <!-- 项目 -->
        <section class="ts-detail__section">
          <div class="ts-detail__section-head">
            <h2 class="ts-detail__section-title">关联项目</h2>
            <button class="ts-detail__action" @click="showAddProject = !showAddProject">
              {{ showAddProject ? "取消" : "关联项目" }}
            </button>
          </div>

          <div v-if="showAddProject" class="ts-detail__add">
            <select v-model="selectedProjectId" class="ts-detail__select">
              <option :value="0" disabled>选择项目</option>
              <option v-for="p in availableProjects()" :key="p.id" :value="p.id">
                {{ p.name }}
              </option>
            </select>
            <button
              class="ts-detail__primary"
              :disabled="!selectedProjectId"
              @click="addProject"
            >
              关联
            </button>
          </div>

          <div class="ts-detail__list">
            <div
              v-for="pid in linkedProjectIds"
              :key="pid"
              class="ts-detail__list-item"
            >
              <div class="ts-detail__list-item-main">
                <span class="ts-detail__list-item-name">{{ projectName(pid) }}</span>
              </div>
              <button
                class="ts-detail__link ts-detail__link--danger"
                @click="removeProject(pid)"
              >
                解除关联
              </button>
            </div>
            <div v-if="!linkedProjectIds.length" class="ts-detail__empty">
              暂无关联项目
            </div>
          </div>
        </section>
      </div>
    </template>
    </div>
  </AppPageContainer>
</template>


<style scoped>
.ts-detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.ts-detail__header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.ts-detail__back {
  padding: 6px 12px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.ts-detail__head-main {
  flex: 1;
}

.ts-detail__title-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ts-detail__dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.ts-detail__title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary);
}

.ts-detail__desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.ts-detail__section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
}

.ts-detail__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ts-detail__section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary);
}

.ts-detail__action {
  font-size: 12px;
  color: var(--accent);
  background: none;
  border: none;
  cursor: pointer;
}

.ts-detail__metrics {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 12px;
}

.ts-detail__metric {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px;
  border-radius: 8px;
  background: var(--bg-surface-2);
}

.ts-detail__metric-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--txt-primary);
}

.ts-detail__metric-value--danger {
  color: var(--danger);
}

.ts-detail__metric-label {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.ts-detail__columns {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 20px;
}

.ts-detail__add {
  display: flex;
  gap: 8px;
}

.ts-detail__select {
  flex: 1;
  padding: 6px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.ts-detail__primary {
  padding: 6px 14px;
  font-size: 13px;
  color: var(--text-on-color);
  background: var(--accent);
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.ts-detail__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ts-detail__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ts-detail__list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
}

.ts-detail__list-item-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.ts-detail__list-item-name {
  font-size: 13px;
  color: var(--txt-primary);
}

.ts-detail__list-item-meta {
  font-size: 12px;
  color: var(--txt-tertiary);
}

.ts-detail__link {
  font-size: 12px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
}

.ts-detail__link--danger:hover {
  color: var(--danger);
}

.ts-detail__empty {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--txt-tertiary);
}
</style>
