<script setup lang="ts">
/**
 * ProjectStateView — 项目状态与更新。
 *
 * 功能：
 *  - 查看/编辑项目状态（on_track / at_risk / off_track / completed / paused）
 *  - 设置优先级与起止日期
 *  - 发布状态更新（含进度快照）
 *  - 更新流：表情回应 + 线程评论
 */
import { computed, onMounted, ref, watch } from "vue";

import {
  projectStateApi,
  type ProjectState,
  type ProjectUpdate,
  type ProjectUpdateReaction,
  type ProjectUpdateComment,
  type ProjectStatus,
  type ProjectPriority,
} from "@/api/services/projectstate";
import { AppBadge, AppEmptyState, AppLoadingState, AppModal, AppPageContainer } from "@/components";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const props = defineProps<{
  workspaceId: number;
  projectId: number;
}>();

const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

/* ---- 状态 ---- */

const state = ref<ProjectState | null>(null);
const updates = ref<ProjectUpdate[]>([]);
const loading = ref(true);
const error = ref("");

/* ---- 状态编辑 ---- */

const showStateForm = ref(false);
const savingState = ref(false);
const stateForm = ref({
  status: "on_track" as ProjectStatus,
  priority: "medium" as ProjectPriority,
  start_date: "",
  target_date: "",
  description: "",
});

/* ---- 更新发布 ---- */

const showUpdateForm = ref(false);
const savingUpdate = ref(false);
const updateForm = ref({
  status: "on_track" as ProjectStatus,
  content: "",
  progress_percent: 0,
});

/* ---- 反应与评论 ---- */

const reactionMap = ref<Record<number, ProjectUpdateReaction[]>>({});
const commentMap = ref<Record<number, ProjectUpdateComment[]>>({});
const showComments = ref<Record<number, boolean>>({});
const commentDraft = ref<Record<number, string>>({});

const emojiPresets = ["👍", "❤️", "🎉", "🚀", "👀", "🙏"];

/* ---- 枚举 ---- */

const statusOptions: Array<{ value: ProjectStatus; label: string }> = [
  { value: "on_track", label: "正常推进" },
  { value: "at_risk", label: "存在风险" },
  { value: "off_track", label: "偏离轨道" },
  { value: "completed", label: "已完成" },
  { value: "paused", label: "已暂停" },
];
const priorityOptions: Array<{ value: ProjectPriority; label: string }> = [
  { value: "urgent", label: "紧急" },
  { value: "high", label: "高" },
  { value: "medium", label: "中" },
  { value: "low", label: "低" },
  { value: "none", label: "无" },
];

function statusLabel(s: string): string {
  return statusOptions.find((o) => o.value === s)?.label ?? s;
}
function statusVariant(s: string): "default" | "success" | "info" | "warning" | "brand" | "danger" {
  switch (s) {
    case "on_track":
      return "success";
    case "at_risk":
      return "warning";
    case "off_track":
      return "danger";
    case "completed":
      return "brand";
    default:
      return "default";
  }
}
function priorityVariant(p: string): "default" | "success" | "info" | "warning" | "brand" | "danger" {
  switch (p) {
    case "urgent":
      return "danger";
    case "high":
      return "warning";
    default:
      return "default";
  }
}

function formatDate(s: string | null): string {
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" });
}

function formatTime(s: string): string {
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString("zh-CN", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
}

/* ---- 加载 ---- */

async function load() {
  if (!props.workspaceId || !props.projectId) return;
  loading.value = true;
  error.value = "";
  try {
    const [s, u] = await Promise.all([
      projectStateApi.getState(props.workspaceId, props.projectId).catch(() => null),
      projectStateApi.listUpdates(props.workspaceId, props.projectId),
    ]);
    state.value = s;
    updates.value = (u as { items?: ProjectUpdate[] }).items ?? [];

    // 加载每条更新的反应与评论
    await Promise.allSettled(
      updates.value.map(async (upd) => {
        const [r, c] = await Promise.all([
          projectStateApi.listReactions(props.workspaceId, props.projectId, upd.id),
          projectStateApi.listComments(props.workspaceId, props.projectId, upd.id),
        ]);
        reactionMap.value = {
          ...reactionMap.value,
          [upd.id]: ((r as { items?: ProjectUpdateReaction[] }).items ?? []),
        };
        commentMap.value = {
          ...commentMap.value,
          [upd.id]: ((c as { items?: ProjectUpdateComment[] }).items ?? []),
        };
      }),
    );
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载项目状态失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch(() => props.projectId, load);

/* ---- 状态操作 ---- */

function openStateForm() {
  stateForm.value = {
    status: state.value?.status ?? "on_track",
    priority: state.value?.priority ?? "medium",
    start_date: state.value?.start_date ?? "",
    target_date: state.value?.target_date ?? "",
    description: state.value?.description ?? "",
  };
  showStateForm.value = true;
}

async function saveState() {
  savingState.value = true;
  try {
    const res = await projectStateApi.upsertState(props.workspaceId, props.projectId, {
      status: stateForm.value.status,
      priority: stateForm.value.priority,
      start_date: stateForm.value.start_date || undefined,
      target_date: stateForm.value.target_date || undefined,
      description: stateForm.value.description,
    });
    state.value = res;
    toast.success("项目状态已更新");
    showStateForm.value = false;
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    savingState.value = false;
  }
}

/* ---- 更新操作 ---- */

function openUpdateForm() {
  updateForm.value = {
    status: state.value?.status ?? "on_track",
    content: "",
    progress_percent: 0,
  };
  showUpdateForm.value = true;
}

async function saveUpdate() {
  if (!updateForm.value.content.trim()) {
    toast.error("请输入更新内容");
    return;
  }
  savingUpdate.value = true;
  try {
    await projectStateApi.createUpdate(props.workspaceId, props.projectId, updateForm.value);
    toast.success("更新已发布");
    showUpdateForm.value = false;
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "发布失败");
  } finally {
    savingUpdate.value = false;
  }
}

async function removeUpdate(upd: ProjectUpdate) {
  if (!await confirm({ title: "删除更新", message: "确认删除这条更新？", confirmText: "删除", variant: "danger" })) return;
  try {
    await projectStateApi.removeUpdate(props.workspaceId, props.projectId, upd.id);
    toast.success("已删除");
    await load();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

/* ---- 反应 ---- */

async function toggleReaction(upd: ProjectUpdate, emoji: string) {
  const existing = (reactionMap.value[upd.id] ?? []).find(
    (r) => r.user_id === currentUserId.value && r.emoji === emoji,
  );
  try {
    if (existing) {
      await projectStateApi.removeReaction(props.workspaceId, props.projectId, upd.id, emoji);
      reactionMap.value = {
        ...reactionMap.value,
        [upd.id]: (reactionMap.value[upd.id] ?? []).filter((r) => r.id !== existing.id),
      };
    } else {
      await projectStateApi.addReaction(props.workspaceId, props.projectId, upd.id, emoji);
      const res = await projectStateApi.listReactions(props.workspaceId, props.projectId, upd.id);
      reactionMap.value = {
        ...reactionMap.value,
        [upd.id]: ((res as { items?: ProjectUpdateReaction[] }).items ?? []),
      };
    }
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "操作失败");
  }
}

function hasReacted(upd: ProjectUpdate, emoji: string): boolean {
  return (reactionMap.value[upd.id] ?? []).some(
    (r) => r.user_id === currentUserId.value && r.emoji === emoji,
  );
}

function reactionCount(upd: ProjectUpdate, emoji: string): number {
  return (reactionMap.value[upd.id] ?? []).filter((r) => r.emoji === emoji).length;
}

/* ---- 评论 ---- */

async function toggleComments(upd: ProjectUpdate) {
  showComments.value = { ...showComments.value, [upd.id]: !showComments.value[upd.id] };
  if (showComments.value[upd.id] && !commentMap.value[upd.id]) {
    try {
      const res = await projectStateApi.listComments(props.workspaceId, props.projectId, upd.id);
      commentMap.value = {
        ...commentMap.value,
        [upd.id]: ((res as { items?: ProjectUpdateComment[] }).items ?? []),
      };
    } catch {
      /* 忽略 */
    }
  }
}

async function addComment(upd: ProjectUpdate) {
  const content = (commentDraft.value[upd.id] ?? "").trim();
  if (!content) return;
  try {
    await projectStateApi.addComment(props.workspaceId, props.projectId, upd.id, content);
    commentDraft.value = { ...commentDraft.value, [upd.id]: "" };
    const res = await projectStateApi.listComments(props.workspaceId, props.projectId, upd.id);
    commentMap.value = {
      ...commentMap.value,
      [upd.id]: ((res as { items?: ProjectUpdateComment[] }).items ?? []),
    };
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "评论失败");
  }
}

async function removeComment(upd: ProjectUpdate, commentId: number) {
  try {
    await projectStateApi.removeComment(props.workspaceId, props.projectId, commentId);
    commentMap.value = {
      ...commentMap.value,
      [upd.id]: (commentMap.value[upd.id] ?? []).filter((c) => c.id !== commentId),
    };
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}
</script>

<template>
  
  <AppPageContainer max="900px">
    <div class="pstate">
    <!-- 状态卡片 -->
    <section class="pstate__card">
      <div class="pstate__card-head">
        <h2 class="pstate__card-title">项目状态</h2>
        <button class="pstate__action" @click="openStateForm">编辑状态</button>
      </div>

      <AppLoadingState v-if="loading && !state" />
      <div v-else-if="state" class="pstate__state">
        <div class="pstate__state-badges">
          <AppBadge :variant="statusVariant(state.status)">
            {{ statusLabel(state.status) }}
          </AppBadge>
          <AppBadge :variant="priorityVariant(state.priority)">
            {{ priorityOptions.find((o) => o.value === state!.priority)?.label }}
          </AppBadge>
        </div>
        <p v-if="state.description" class="pstate__state-desc">{{ state.description }}</p>
        <div class="pstate__state-dates">
          <span>开始：{{ formatDate(state.start_date) }}</span>
          <span>目标：{{ formatDate(state.target_date) }}</span>
          <span>更新：{{ formatTime(state.updated_at) }}</span>
        </div>
      </div>
      <div v-else class="pstate__state-empty">尚未设置项目状态</div>
    </section>

    <!-- 发布更新 -->
    <section class="pstate__card">
      <div class="pstate__card-head">
        <h2 class="pstate__card-title">状态更新</h2>
        <button class="pstate__action" @click="openUpdateForm">发布更新</button>
      </div>

      <AppLoadingState v-if="loading" />
      <AppEmptyState
        v-else-if="!updates.length"
        title="暂无更新"
        description="发布项目状态更新，记录进展与风险"
      />
      <div v-else class="pstate__timeline">
        <div v-for="upd in updates" :key="upd.id" class="pstate__update">
          <div class="pstate__update-head">
            <span class="pstate__avatar">{{ upd.author_id === currentUserId ? "我" : "#" + upd.author_id }}</span>
            <div class="pstate__update-head-main">
              <div class="pstate__update-badges">
                <AppBadge :variant="statusVariant(upd.status)">
                  {{ statusLabel(upd.status) }}
                </AppBadge>
                <span class="pstate__update-progress">
                  进度 {{ Math.round(upd.progress_percent) }}%
                </span>
              </div>
              <span class="pstate__update-time">{{ formatTime(upd.created_at) }}</span>
            </div>
            <button
              v-if="upd.author_id === currentUserId"
              class="pstate__link pstate__link--danger"
              @click="removeUpdate(upd)"
            >
              删除
            </button>
          </div>

          <p class="pstate__update-content">{{ upd.content }}</p>

          <!-- 反应 -->
          <div class="pstate__reactions">
            <button
              v-for="emoji in emojiPresets"
              :key="emoji"
              class="pstate__reaction"
              :class="{ 'pstate__reaction--active': hasReacted(upd, emoji) }"
              @click="toggleReaction(upd, emoji)"
            >
              {{ emoji }}<span v-if="reactionCount(upd, emoji)">{{ reactionCount(upd, emoji) }}</span>
            </button>
            <button class="pstate__link" @click="toggleComments(upd)">
              {{ showComments[upd.id] ? "收起评论" : "评论" }}
              <template v-if="(commentMap[upd.id] ?? []).length">
                ({{ (commentMap[upd.id] ?? []).length }})
              </template>
            </button>
          </div>

          <!-- 评论 -->
          <div v-if="showComments[upd.id]" class="pstate__comments">
            <div
              v-for="c in commentMap[upd.id] ?? []"
              :key="c.id"
              class="pstate__comment"
            >
              <span class="pstate__avatar pstate__avatar--sm">{{ c.author_id === currentUserId ? "我" : "#" + c.author_id }}</span>
              <div class="pstate__comment-main">
                <span class="pstate__comment-meta">
                  {{ c.author_id === currentUserId ? "我" : `用户 #${c.author_id}` }}
                  · {{ formatTime(c.created_at) }}
                </span>
                <span class="pstate__comment-content">{{ c.content }}</span>
              </div>
              <button
                v-if="c.author_id === currentUserId"
                class="pstate__link pstate__link--danger"
                @click="removeComment(upd, c.id)"
              >
                删除
              </button>
            </div>

            <div class="pstate__comment-input">
              <input
                v-model="commentDraft[upd.id]"
                class="pstate__input"
                placeholder="写下你的评论..."
                @keyup.enter="addComment(upd)"
              />
              <button class="pstate__primary" @click="addComment(upd)">发送</button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 状态编辑弹窗 -->
    <AppModal
      :visible="showStateForm"
      title="编辑项目状态"
      @close="showStateForm = false"
    >
      <div class="pstate__form">
        <label class="pstate__label">状态</label>
        <select v-model="stateForm.status" class="pstate__input">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">
            {{ o.label }}
          </option>
        </select>

        <label class="pstate__label">优先级</label>
        <select v-model="stateForm.priority" class="pstate__input">
          <option v-for="o in priorityOptions" :key="o.value" :value="o.value">
            {{ o.label }}
          </option>
        </select>

        <div class="pstate__grid">
          <div>
            <label class="pstate__label">开始日期</label>
            <input v-model="stateForm.start_date" type="date" class="pstate__input" />
          </div>
          <div>
            <label class="pstate__label">目标日期</label>
            <input v-model="stateForm.target_date" type="date" class="pstate__input" />
          </div>
        </div>

        <label class="pstate__label">描述</label>
        <textarea
          v-model="stateForm.description"
          class="pstate__textarea"
          rows="3"
          placeholder="项目当前的整体情况"
        />
      </div>
      <template #footer>
        <button class="pstate__ghost" @click="showStateForm = false">取消</button>
        <button class="pstate__primary" :disabled="savingState" @click="saveState">
          {{ savingState ? "保存中..." : "保存" }}
        </button>
      </template>
    </AppModal>

    <!-- 发布更新弹窗 -->
    <AppModal
      :visible="showUpdateForm"
      title="发布项目更新"
      @close="showUpdateForm = false"
    >
      <div class="pstate__form">
        <label class="pstate__label">当前状态</label>
        <select v-model="updateForm.status" class="pstate__input">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">
            {{ o.label }}
          </option>
        </select>

        <label class="pstate__label">进度（%）</label>
        <input
          v-model.number="updateForm.progress_percent"
          type="number"
          min="0"
          max="100"
          class="pstate__input"
        />

        <label class="pstate__label">更新内容</label>
        <textarea
          v-model="updateForm.content"
          class="pstate__textarea"
          rows="5"
          placeholder="本期进展、风险与下一步计划..."
        />
      </div>
      <template #footer>
        <button class="pstate__ghost" @click="showUpdateForm = false">取消</button>
        <button class="pstate__primary" :disabled="savingUpdate" @click="saveUpdate">
          {{ savingUpdate ? "发布中..." : "发布" }}
        </button>
      </template>
    </AppModal>
    </div>
  </AppPageContainer>
</template>

<style scoped>
.pstate {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  
  
}

.pstate__card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: var(--bg-surface);
}

.pstate__card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.pstate__card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--txt-primary);
}

.pstate__action {
  font-size: 12px;
  color: var(--accent);
  background: none;
  border: none;
  cursor: pointer;
}

.pstate__state {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pstate__state-badges {
  display: flex;
  gap: 8px;
}

.pstate__state-desc {
  margin: 0;
  font-size: 13px;
  color: var(--txt-secondary);
}

.pstate__state-dates {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--txt-tertiary);
}

.pstate__state-empty {
  font-size: 13px;
  color: var(--txt-tertiary);
}

.pstate__timeline {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pstate__update {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.pstate__update:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.pstate__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--bg-accent-subtle, rgba(59, 130, 246, 0.12));
  color: var(--accent);
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.pstate__avatar--sm {
  width: 24px;
  height: 24px;
  font-size: 10px;
}

.pstate__update-head {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.pstate__update-head-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pstate__update-badges {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pstate__update-progress {
  font-size: 12px;
  color: var(--txt-tertiary);
}

.pstate__update-time {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.pstate__update-content {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--txt-primary);
  white-space: pre-wrap;
}

.pstate__reactions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.pstate__reaction {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 16px;
  background: var(--bg-surface);
  cursor: pointer;
}

.pstate__reaction:hover {
  background: var(--bg-surface-2);
}

.pstate__reaction--active {
  border-color: var(--accent);
  background: var(--bg-accent-subtle, rgba(59, 130, 246, 0.08));
}

.pstate__link {
  font-size: 12px;
  color: var(--txt-secondary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.pstate__link:hover {
  color: var(--accent);
}

.pstate__link--danger:hover {
  color: var(--danger);
}

.pstate__comments {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  background: var(--bg-surface-2);
}

.pstate__comment {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.pstate__comment-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pstate__comment-meta {
  font-size: 11px;
  color: var(--txt-tertiary);
}

.pstate__comment-content {
  font-size: 13px;
  color: var(--txt-primary);
}

.pstate__comment-input {
  display: flex;
  gap: 8px;
}

.pstate__input {
  flex: 1;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
}

.pstate__primary {
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-on-color);
  background: var(--accent);
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.pstate__primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pstate__ghost {
  padding: 8px 14px;
  font-size: 13px;
  color: var(--txt-secondary);
  background: none;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  cursor: pointer;
}

.pstate__form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pstate__grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.pstate__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--txt-secondary);
}

.pstate__textarea {
  width: 100%;
  padding: 8px 10px;
  font-size: 13px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: var(--bg-surface);
  color: var(--txt-primary);
  resize: vertical;
}
</style>
