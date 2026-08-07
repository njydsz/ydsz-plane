<script setup lang="ts">
/**
 * CommentList — 评论列表（含新建/回复/编辑/删除 + nested replies）
 *
 * Props:
 *   workspaceId / projectId / issueId — 路由参数，用以调用 API
 *
 * 自管理加载/错误状态，通过 issueApi 调用后端评论接口。
 */
import { computed, nextTick, ref, watch } from "vue";
import { issueApi, type IssueComment, type CreateCommentInput, type UpdateCommentInput } from "@/api/services/issue";
import { useAuthStore } from "@/stores/auth";
import { toast } from "@/lib/toast";
import CommentItem from "./CommentItem.vue";
import CommentForm from "./CommentForm.vue";
import AppEmptyState from "./AppEmptyState.vue";
import AppLoadingState from "./AppLoadingState.vue";
import AppErrorState from "./AppErrorState.vue";

const props = withDefaults(defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
  /** 紧凑模式：隐藏新建表单，仅展示评论列表 */
  compact?: boolean;
  /** 限制显示条数（紧凑模式下生效） */
  limit?: number;
}>(), {
  compact: false,
  limit: 0,
});

const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

const comments = ref<IssueComment[]>([]);
const loading = ref(true);
const error = ref("");
const submitting = ref(false);

// 编辑/回复状态
const editingComment = ref<IssueComment | null>(null);
const replyingTo = ref<IssueComment | null>(null);

// 表单 ref
const newFormRef = ref<InstanceType<typeof CommentForm> | null>(null);
const editFormRef = ref<InstanceType<typeof CommentForm> | null>(null);
const replyFormRef = ref<InstanceType<typeof CommentForm> | null>(null);

/* ---- 数据加载 ---- */

async function loadComments() {
  if (!props.workspaceId) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await issueApi.listComments(props.workspaceId, props.projectId, props.issueId);
    comments.value = res.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载评论失败";
  } finally {
    loading.value = false;
  }
}

/* ---- 评论分组 ---- */
// 顶级评论（parent_id 为空）
const topLevelComments = computed(() =>
  comments.value.filter((c) => !c.parent_id),
);

// 按 parent_id 分组子回复
const repliesByParent = computed(() => {
  const map = new Map<number, IssueComment[]>();
  for (const c of comments.value) {
    if (c.parent_id) {
      const arr = map.get(c.parent_id) ?? [];
      arr.push(c);
      map.set(c.parent_id, arr);
    }
  }
  return map;
});

// compact + limit 时截断顶级评论列表
const displayedComments = computed(() => {
  if (props.compact && props.limit > 0) {
    return topLevelComments.value.slice(0, props.limit);
  }
  return topLevelComments.value;
});

/* ---- 操作 ---- */

async function handleCreate(payload: { content_html: string; content_json: string; content_stripped: string; parent_id?: number | null }) {
  if (submitting.value) return;
  submitting.value = true;
  try {
    const input: CreateCommentInput = {
      content_json: payload.content_json || JSON.stringify({ type: "doc", content: [] }),
      content_html: payload.content_html,
      content_stripped: payload.content_stripped,
    };
    if (payload.parent_id) {
      input.parent_id = payload.parent_id;
    }
    await issueApi.createComment(props.workspaceId, props.projectId, props.issueId, input);

    // 重置状态
    replyingTo.value = null;
    toast.success("评论已发布");
    await loadComments();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "发表评论失败";
    toast.error(error.value);
  } finally {
    submitting.value = false;
  }
}

function handleStartReply(comment: IssueComment) {
  replyingTo.value = comment;
  editingComment.value = null;
  nextTick(() => replyFormRef.value?.focus());
}

function handleStartEdit(comment: IssueComment) {
  editingComment.value = comment;
  replyingTo.value = null;
}

async function handleUpdate(payload: { content_html: string; content_json: string; content_stripped: string }) {
  if (!editingComment.value || submitting.value) return;
  submitting.value = true;
  try {
    const input: UpdateCommentInput = {
      content_json: payload.content_json || JSON.stringify({ type: "doc", content: [] }),
      content_html: payload.content_html,
      content_stripped: payload.content_stripped,
    };
    await issueApi.updateComment(
      props.workspaceId, props.projectId, props.issueId,
      editingComment.value.id, input,
    );
    editingComment.value = null;
    toast.success("评论已更新");
    await loadComments();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "编辑评论失败";
    toast.error(error.value);
  } finally {
    submitting.value = false;
  }
}

async function handleDelete(comment: IssueComment) {
  try {
    await issueApi.deleteComment(props.workspaceId, props.projectId, props.issueId, comment.id);
    toast.success("评论已删除");
    await loadComments();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除评论失败";
    toast.error(error.value);
  }
}

function cancelEdit() {
  editingComment.value = null;
}

function cancelReply() {
  replyingTo.value = null;
}

// 监听 props 变化自动刷新
watch(
  () => [props.workspaceId, props.projectId, props.issueId],
  () => {
    if (props.workspaceId && props.projectId && props.issueId) {
      loadComments();
    }
  },
  { immediate: true },
);
</script>

<template>
  <div class="comment-list" :class="{ 'comment-list--compact': compact }">
    <div class="comment-list__header">
      <h3 class="comment-list__title">
        评论
        <span v-if="comments.length" class="comment-list__count">{{ comments.length }}</span>
      </h3>
    </div>

    <!-- 新建评论表单（compact 模式下隐藏） -->
    <div v-if="!compact && !editingComment && !replyingTo" class="comment-list__new">
      <CommentForm
        ref="newFormRef"
        :loading="submitting"
        @submit="handleCreate"
      />
    </div>

    <!-- 编辑表单 -->
    <div v-if="editingComment" class="comment-list__edit-banner">
      <span class="comment-list__edit-label">正在编辑评论</span>
    </div>
    <div v-if="editingComment" class="comment-list__edit-form">
      <CommentForm
        ref="editFormRef"
        :loading="submitting"
        :initial-value="editingComment.content_html || editingComment.content_stripped"
        @submit="handleUpdate"
        @cancel="cancelEdit"
      />
    </div>

    <!-- 错误 -->
    <AppErrorState
      v-if="error && comments.length === 0"
      :message="error"
      @retry="loadComments"
    />

    <!-- 加载中 -->
    <AppLoadingState v-if="loading && comments.length === 0" text="加载评论..." />

    <!-- 空态 -->
    <AppEmptyState
      v-if="!loading && !error && comments.length === 0 && !editingComment"
      title="暂无评论"
      description="成为第一个评论的人"
    />

    <!-- 评论列表（compact 模式下按 limit 截断，隐藏回复表单） -->
    <div v-if="comments.length > 0" class="comment-list__items">
      <template v-for="topComment in displayedComments" :key="topComment.id">
        <CommentItem
          :comment="topComment"
          :current-user-id="currentUserId"
          @edit="handleStartEdit"
          @delete="handleDelete"
          @reply="handleStartReply"
        />

        <!-- 嵌套回复 -->
        <div
          v-if="repliesByParent.get(topComment.id)?.length"
          class="comment-list__replies"
        >
          <CommentItem
            v-for="reply in repliesByParent.get(topComment.id)"
            :key="reply.id"
            :comment="reply"
            :current-user-id="currentUserId"
            :reply-hint="topComment.creator_name || '匿名用户'"
            @edit="handleStartEdit"
            @delete="handleDelete"
            @reply="handleStartReply"
          />
        </div>

        <!-- 回复表单（内联） -->
        <div
          v-if="replyingTo?.id === topComment.id"
          class="comment-list__reply-form"
        >
          <div class="comment-list__reply-hint">
            回复 @{{ topComment.creator_name || "匿名用户" }}
          </div>
          <CommentForm
            ref="replyFormRef"
            :loading="submitting"
            :parent-id="topComment.id"
            @submit="handleCreate"
            @cancel="cancelReply"
          />
        </div>
      </template>

      <!-- 页面级错误横幅（已有数据时） -->
      <div v-if="error && comments.length > 0" class="comment-list__error-banner">
        {{ error }}
        <button class="btn--link" @click="loadComments">重试</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comment-list {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-subtle, #e5e7eb);
}

.comment-list__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.comment-list__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.comment-list__count {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-tertiary, #9ca3af);
  background: var(--surface-3, #f3f4f6);
  padding: 1px 7px;
  border-radius: 10px;
}

.comment-list__new {
  margin-bottom: 16px;
}

.comment-list__edit-banner,
.comment-list__edit-form {
  margin-bottom: 12px;
}

.comment-list__edit-label {
  font-size: 12px;
  color: var(--brand-500, #3b82f6);
  font-weight: 500;
}

.comment-list__items {
  display: flex;
  flex-direction: column;
}

.comment-list__replies {
  margin-left: 36px;
  padding-left: 12px;
  border-left: 2px solid var(--border-subtle, #e5e7eb);
  margin-top: 0;
}

.comment-list__reply-form {
  margin-left: 36px;
  margin-top: 12px;
  margin-bottom: 12px;
  padding: 12px;
  background: var(--surface-2, #f9fafb);
  border-radius: var(--radius-md, 8px);
}

.comment-list__reply-hint {
  font-size: 12px;
  color: var(--brand-500, #3b82f6);
  margin-bottom: 8px;
  font-weight: 500;
}

.comment-list__error-banner {
  margin-top: 12px;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--danger-500, #ef4444);
  background: var(--danger-50, rgba(239,68,68,0.06));
  border-radius: var(--radius-sm, 4px);
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn--link {
  background: none;
  border: none;
  padding: 0;
  font-size: 12px;
  color: var(--brand-500, #3b82f6);
  cursor: pointer;
  font-family: inherit;
  text-decoration: underline;
}

/* ===== Compact 模式（用于 Peek Overview） ===== */
.comment-list--compact {
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}
.comment-list--compact .comment-list__header {
  margin-bottom: 8px;
}
.comment-list--compact .comment-list__title {
  font-size: 12px;
}
.comment-list--compact .comment-list__items {
  gap: 0;
}
</style>
