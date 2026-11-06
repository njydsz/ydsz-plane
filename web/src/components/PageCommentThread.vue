<script setup lang="ts">
/**
 * PageCommentThread — 页面评论线程组件（页面级 / 行内评论）。
 *
 * 功能：
 *   - 顶部新建评论输入框（纯文本，Enter 提交 / Shift+Enter 换行）
 *   - 线程列表：根评论 + 嵌套回复（缩进），作者头像 / 昵称 / 时间
 *   - 每条评论支持回复、编辑（本人）、删除（本人）
 *   - 表情回应选择器（👍❤️😂🎉😮😢），点击切换，展示计数
 *   - 根评论线程的解决 / 取消解决（展示「已解决」徽标）
 *   - 行内评论（block_id 存在时）展示引用原文
 *   - 评论审核：具备 canModerate 时可隐藏 / 取消隐藏，隐藏评论展示「已隐藏」徽标与原因
 *   - 空状态、加载态、错误态
 *
 * Props:
 *   workspaceId / pageId / pageType — 定位评论所属页面
 *   compact                        — 紧凑模式（隐藏标题装饰，用于嵌入窄容器）
 *   canModerate                    — 当前用户是否具备评论审核权限（隐藏 / 取消隐藏）
 */
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  type ComponentPublicInstance,
} from "vue";

import {
  pageCommentApi,
  type PageComment,
  type PageCommentPageType,
  type PageCommentReaction,
} from "@/api/services/pagecomment";
import { useAuthStore } from "@/stores/auth";
import { useTranslation } from "@/composables/useTranslation";
import { useConfirm } from "@/composables/useConfirm";
import { toast } from "@/lib/toast";
import AppAvatar from "./AppAvatar.vue";
import AppBadge from "./AppBadge.vue";
import AppEmptyState from "./AppEmptyState.vue";
import AppLoadingState from "./AppLoadingState.vue";
import AppErrorState from "./AppErrorState.vue";
import AppButton from "./AppButton.vue";

const props = withDefaults(
  defineProps<{
    workspaceId: number;
    pageId: number;
    pageType?: PageCommentPageType;
    /** 紧凑模式：减少留白，用于嵌入窄容器 */
    compact?: boolean;
    /** 是否具备评论审核权限（隐藏 / 取消隐藏）。默认 false，由父组件按角色传入。 */
    canModerate?: boolean;
  }>(),
  {
    pageType: "project",
    compact: false,
    canModerate: false,
  },
);

const { t } = useTranslation();
const confirm = useConfirm();

const auth = useAuthStore();
const currentUserId = computed(() => auth.user?.id ?? 0);

/* ---- 预设表情 ---- */
const PRESET_EMOJIS = ["👍", "❤️", "😂", "🎉", "😮", "😢"] as const;

/* ---- 数据状态 ---- */
const comments = ref<PageComment[]>([]);
/** commentId -> 该评论的表情回应列表 */
const reactionsMap = ref<Map<number, PageCommentReaction[]>>(new Map());
const loading = ref(true);
const error = ref("");
const submitting = ref(false);

/* ---- 交互状态 ---- */
const newCommentText = ref("");
const editingComment = ref<PageComment | null>(null);
const editText = ref("");
const replyingTo = ref<PageComment | null>(null);
const replyText = ref("");
/** 当前打开表情选择器的评论 ID */
const pickerOpenFor = ref<number | null>(null);

/* 输入框引用（用于聚焦） */
const newInputRef = ref<HTMLTextAreaElement | null>(null);
const replyInputRef = ref<HTMLTextAreaElement | null>(null);

/**
 * 编辑框引用：编辑表单在根评论 / 回复中复用同一 ref 名，
 * 但同一时刻只会渲染一个（editingComment 唯一），故用函数式 ref 捕获。
 */
const editInputRef = ref<HTMLTextAreaElement | null>(null);
function setEditInput(el: Element | ComponentPublicInstance | null) {
  editInputRef.value = (el as HTMLTextAreaElement | null) ?? null;
}

/** 回复框引用：回复表单在 v-for 内渲染，用函数式 ref 捕获单例元素 */
function setReplyInput(el: Element | ComponentPublicInstance | null) {
  replyInputRef.value = (el as HTMLTextAreaElement | null) ?? null;
}

/* ---- 线程分组 ---- */
const topLevelComments = computed(() =>
  comments.value.filter((c) => !c.parent_id),
);

const repliesByParent = computed(() => {
  const map = new Map<number, PageComment[]>();
  for (const c of comments.value) {
    if (c.parent_id) {
      const arr = map.get(c.parent_id) ?? [];
      arr.push(c);
      map.set(c.parent_id, arr);
    }
  }
  return map;
});

/* ---- 工具 ---- */
function formatTime(iso: string): string {
  const d = new Date(iso);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return t("pageComment.time.justNow");
  if (mins < 60) return t("pageComment.time.minutesAgo", { n: mins });
  const hours = Math.floor(mins / 60);
  if (hours < 24) return t("pageComment.time.hoursAgo", { n: hours });
  const days = Math.floor(hours / 24);
  if (days < 30) return t("pageComment.time.daysAgo", { n: days });
  return d.toLocaleDateString("zh-CN");
}

function isMine(comment: PageComment): boolean {
  return comment.author_id === currentUserId.value;
}

function isRoot(comment: PageComment): boolean {
  return !comment.parent_id;
}

function isResolved(comment: PageComment): boolean {
  return !!comment.resolved_at;
}

/** 评论是否被审核隐藏 */
function isHidden(comment: PageComment): boolean {
  return !!comment.hidden;
}

/** 隐藏原因文案（缺省兜底） */
function hiddenReason(comment: PageComment): string {
  return comment.moderation_reason?.trim() || t("pageComment.hiddenNoReason");
}

/** 当前用户是否对某评论使用过某 emoji */
function hasMyReaction(commentId: number, emoji: string): boolean {
  const list = reactionsMap.value.get(commentId) ?? [];
  return list.some((r) => r.user_id === currentUserId.value && r.emoji === emoji);
}

/** 某评论的回应摘要（去重 emoji + 计数） */
function reactionSummary(commentId: number): Array<{ emoji: string; count: number }> {
  const list = reactionsMap.value.get(commentId) ?? [];
  const counts = new Map<string, number>();
  for (const r of list) {
    counts.set(r.emoji, (counts.get(r.emoji) ?? 0) + 1);
  }
  return Array.from(counts.entries()).map(([emoji, count]) => ({ emoji, count }));
}

/* ---- 数据加载 ---- */
async function loadComments() {
  if (!props.workspaceId || !props.pageId) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await pageCommentApi.list(props.workspaceId, props.pageId, {
      page_type: props.pageType,
      // 具备审核权限时同时拉取被隐藏的评论，便于在列表内展示「已隐藏」徽标与原因
      include_hidden: props.canModerate,
    });
    comments.value = res.results;
    await loadAllReactions();
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t("pageComment.loadFailed");
  } finally {
    loading.value = false;
  }
}

/** 拉取所有评论的表情回应（逐条请求，后端未内嵌） */
async function loadAllReactions() {
  const ids = comments.value.map((c) => c.id);
  if (ids.length === 0) {
    reactionsMap.value = new Map();
    return;
  }
  const results = await Promise.allSettled(
    ids.map((id) =>
      pageCommentApi
        .listReactions(props.workspaceId, id)
        .then((res) => ({ id, reactions: res.results })),
    ),
  );
  const next = new Map<number, PageCommentReaction[]>();
  for (const r of results) {
    if (r.status === "fulfilled") {
      next.set(r.value.id, r.value.reactions);
    }
  }
  reactionsMap.value = next;
}

/* ---- 新建评论 ---- */
async function submitNewComment() {
  const content = newCommentText.value.trim();
  if (!content || submitting.value) return;
  submitting.value = true;
  try {
    await pageCommentApi.create(
      props.workspaceId,
      props.pageId,
      props.pageType,
      { content },
    );
    newCommentText.value = "";
    toast.success(t("pageComment.createSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.createFailed"));
  } finally {
    submitting.value = false;
  }
}

/* ---- 回复 ---- */
function startReply(comment: PageComment) {
  replyingTo.value = comment;
  editingComment.value = null;
  replyText.value = "";
  nextTick(() => replyInputRef.value?.focus());
}

function cancelReply() {
  replyingTo.value = null;
  replyText.value = "";
}

async function submitReply(rootComment: PageComment) {
  const content = replyText.value.trim();
  if (!content || submitting.value) return;
  submitting.value = true;
  try {
    await pageCommentApi.create(
      props.workspaceId,
      props.pageId,
      props.pageType,
      { content, parent_id: rootComment.id },
    );
    replyText.value = "";
    replyingTo.value = null;
    toast.success(t("pageComment.replySuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.replyFailed"));
  } finally {
    submitting.value = false;
  }
}

/* ---- 编辑 ---- */
function startEdit(comment: PageComment) {
  editingComment.value = comment;
  replyingTo.value = null;
  editText.value = comment.content;
  nextTick(() => {
    const el = editInputRef.value;
    el?.focus();
    // 光标移到末尾
    const len = el?.value.length ?? 0;
    el?.setSelectionRange(len, len);
  });
}

function cancelEdit() {
  editingComment.value = null;
  editText.value = "";
}

async function submitEdit() {
  if (!editingComment.value || submitting.value) return;
  const content = editText.value.trim();
  if (!content) return;
  submitting.value = true;
  try {
    await pageCommentApi.update(props.workspaceId, editingComment.value.id, {
      content,
    });
    editingComment.value = null;
    editText.value = "";
    toast.success(t("pageComment.updateSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.updateFailed"));
  } finally {
    submitting.value = false;
  }
}

/* ---- 删除 ---- */
async function handleDelete(comment: PageComment) {
  if (!await confirm({ title: "删除评论", message: t("pageComment.deleteConfirm"), confirmText: "删除", variant: "danger" })) return;
  try {
    await pageCommentApi.remove(props.workspaceId, comment.id);
    toast.success(t("pageComment.deleteSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.deleteFailed"));
  }
}

/* ---- 解决 / 取消解决 ---- */
async function handleResolve(comment: PageComment) {
  try {
    await pageCommentApi.resolve(props.workspaceId, comment.id);
    toast.success(t("pageComment.resolveSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.resolveFailed"));
  }
}

async function handleUnresolve(comment: PageComment) {
  try {
    await pageCommentApi.unresolve(props.workspaceId, comment.id);
    toast.success(t("pageComment.unresolveSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.unresolveFailed"));
  }
}

/* ---- 审核：隐藏 / 取消隐藏 ---- */
async function handleModerate(comment: PageComment, action: "hide" | "unhide") {
  if (action === "hide") {
    const reason = window.prompt(t("pageComment.hideReasonPrompt"));
    // 用户取消输入框时不做任何变更
    if (reason === null) return;
    try {
      await pageCommentApi.moderateComment(
        props.workspaceId,
        comment.id,
        "hide",
        reason.trim(),
      );
      toast.success(t("pageComment.hideSuccess"));
      await loadComments();
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : t("pageComment.hideFailed"));
    }
    return;
  }

  try {
    await pageCommentApi.moderateComment(props.workspaceId, comment.id, "unhide");
    toast.success(t("pageComment.unhideSuccess"));
    await loadComments();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.unhideFailed"));
  }
}

/* ---- 表情回应切换 ---- */
function togglePicker(commentId: number) {
  pickerOpenFor.value = pickerOpenFor.value === commentId ? null : commentId;
}

async function toggleReaction(comment: PageComment, emoji: string) {
  const already = hasMyReaction(comment.id, emoji);
  try {
    if (already) {
      await pageCommentApi.removeReaction(props.workspaceId, comment.id, emoji);
    } else {
      await pageCommentApi.addReaction(props.workspaceId, comment.id, emoji);
    }
    // 仅刷新该评论的回应，避免整列表重载
    const res = await pageCommentApi.listReactions(props.workspaceId, comment.id);
    reactionsMap.value = new Map(reactionsMap.value).set(comment.id, res.results);
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : t("pageComment.reactionFailed"));
  }
}

/* 点击组件外部关闭表情选择器 */
function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null;
  if (target && target.closest(".page-comment__picker")) return;
  pickerOpenFor.value = null;
}

/* ---- 键盘快捷键 ---- */
function onNewKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    submitNewComment();
  }
}

function onReplyKeydown(e: KeyboardEvent, root: PageComment) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    submitReply(root);
  }
  if (e.key === "Escape") cancelReply();
}

function onEditKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    submitEdit();
  }
  if (e.key === "Escape") cancelEdit();
}

/* ---- 生命周期 ---- */
watch(
  () => [props.workspaceId, props.pageId, props.pageType],
  () => {
    if (props.workspaceId && props.pageId) {
      void loadComments();
    }
  },
  { immediate: true },
);

onMounted(() => {
  document.addEventListener("click", onDocClick);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", onDocClick);
});
</script>

<template>
  <section
    class="page-comment"
    :class="{ 'page-comment--compact': compact }"
    aria-label="page comments"
  >
    <!-- 标题 -->
    <header class="page-comment__header">
      <h3 class="page-comment__title">
        {{ t("pageComment.title") }}
        <span v-if="comments.length" class="page-comment__count">
          {{ comments.length }}
        </span>
      </h3>
    </header>

    <!-- 新建评论输入框 -->
    <div class="page-comment__new">
      <AppAvatar
        :name="auth.user?.display_name || 'U'"
        :src="auth.user?.avatar_url || ''"
        size="sm"
      />
      <div class="page-comment__new-body">
        <textarea
          ref="newInputRef"
          v-model="newCommentText"
          class="page-comment__textarea"
          rows="2"
          :placeholder="t('pageComment.newPlaceholder')"
          @keydown="onNewKeydown"
        />
        <div class="page-comment__new-footer">
          <span class="page-comment__hint">{{ t("pageComment.inputHint") }}</span>
          <AppButton
            size="sm"
            :loading="submitting"
            :disabled="!newCommentText.trim() || submitting"
            @click="submitNewComment"
          >
            {{ submitting ? t("pageComment.submitting") : t("pageComment.submit") }}
          </AppButton>
        </div>
      </div>
    </div>

    <!-- 错误态（无数据时） -->
    <AppErrorState
      v-if="error && comments.length === 0"
      :message="error"
      :retry="t('pageComment.retry')"
      @retry="loadComments"
    />

    <!-- 加载态 -->
    <AppLoadingState
      v-else-if="loading && comments.length === 0"
      :text="t('pageComment.loading')"
    />

    <!-- 空态 -->
    <AppEmptyState
      v-else-if="!loading && !error && comments.length === 0"
      icon="💬"
      :title="t('pageComment.emptyTitle')"
      :description="t('pageComment.emptyDesc')"
      compact
    />

    <!-- 线程列表 -->
    <div v-else class="page-comment__list">
      <template v-for="root in topLevelComments" :key="root.id">
        <article
          class="page-comment__item"
          :class="{
            'page-comment__item--resolved': isResolved(root),
            'page-comment__item--hidden': isHidden(root),
          }"
        >
          <!-- 根评论 -->
          <div class="page-comment__row">
            <AppAvatar
              :name="root.author_name || 'U'"
              :src="root.author_avatar || ''"
              size="sm"
            />
            <div class="page-comment__body">
              <div class="page-comment__meta">
                <span class="page-comment__author">
                  {{ root.author_name || t("pageComment.anonymous") }}
                </span>
                <span class="page-comment__time">{{ formatTime(root.created_at) }}</span>
                <span v-if="root.is_edited" class="page-comment__edited">
                  ({{ t("pageComment.edited") }})
                </span>
                <AppBadge v-if="isResolved(root)" variant="success">
                  {{ t("pageComment.resolved") }}
                </AppBadge>
                <AppBadge v-if="isHidden(root)" variant="danger">
                  {{ t("pageComment.hidden") }}
                </AppBadge>
              </div>

              <!-- 隐藏原因（仅审核者可见） -->
              <div v-if="isHidden(root) && canModerate" class="page-comment__hidden-reason">
                <span class="page-comment__hidden-reason-label">
                  {{ t("pageComment.hiddenReasonLabel") }}
                </span>
                <span class="page-comment__hidden-reason-text">
                  {{ hiddenReason(root) }}
                </span>
                <span v-if="root.moderated_at" class="page-comment__time">
                  {{ formatTime(root.moderated_at) }}
                </span>
              </div>

              <!-- 行内评论引用 -->
              <div v-if="root.block_id" class="page-comment__quote">
                <span class="page-comment__quote-label">
                  {{ t("pageComment.inlineQuote") }}
                </span>
                <span class="page-comment__quote-text">
                  {{ root.quote_text || root.block_id }}
                </span>
              </div>

              <!-- 内容（编辑态 / 展示态） -->
              <div v-if="editingComment?.id === root.id" class="page-comment__edit">
                <textarea
                  ref="editInputRef"
                  v-model="editText"
                  class="page-comment__textarea"
                  rows="3"
                  @keydown="onEditKeydown"
                />
                <div class="page-comment__edit-actions">
                  <AppButton
                    size="sm"
                    :loading="submitting"
                    :disabled="!editText.trim() || submitting"
                    @click="submitEdit"
                  >
                    {{ t("pageComment.save") }}
                  </AppButton>
                  <AppButton size="sm" variant="secondary" :disabled="submitting" @click="cancelEdit">
                    {{ t("pageComment.cancel") }}
                  </AppButton>
                </div>
              </div>
              <p v-else class="page-comment__content">{{ root.content }}</p>

              <!-- 操作栏 -->
              <div class="page-comment__actions">
                <button class="page-comment__action" @click="startReply(root)">
                  {{ t("pageComment.reply") }}
                </button>
                <template v-if="isMine(root)">
                  <button class="page-comment__action" @click="startEdit(root)">
                    {{ t("pageComment.edit") }}
                  </button>
                  <button
                    class="page-comment__action page-comment__action--danger"
                    @click="handleDelete(root)"
                  >
                    {{ t("pageComment.delete") }}
                  </button>
                </template>
                <button
                  v-if="canModerate && !isHidden(root)"
                  class="page-comment__action page-comment__action--danger"
                  @click="handleModerate(root, 'hide')"
                >
                  {{ t("pageComment.hide") }}
                </button>
                <button
                  v-if="canModerate && isHidden(root)"
                  class="page-comment__action"
                  @click="handleModerate(root, 'unhide')"
                >
                  {{ t("pageComment.unhide") }}
                </button>
                <button
                  v-if="isRoot(root) && !isResolved(root)"
                  class="page-comment__action page-comment__action--resolve"
                  @click="handleResolve(root)"
                >
                  {{ t("pageComment.resolve") }}
                </button>
                <button
                  v-if="isRoot(root) && isResolved(root)"
                  class="page-comment__action"
                  @click="handleUnresolve(root)"
                >
                  {{ t("pageComment.unresolve") }}
                </button>
              </div>

              <!-- 表情回应区 -->
              <div class="page-comment__reactions">
                <button
                  v-for="item in reactionSummary(root.id)"
                  :key="item.emoji"
                  class="page-comment__reaction"
                  :class="{
                    'page-comment__reaction--active': hasMyReaction(root.id, item.emoji),
                  }"
                  @click.stop="toggleReaction(root, item.emoji)"
                >
                  <span class="page-comment__reaction-emoji">{{ item.emoji }}</span>
                  <span class="page-comment__reaction-count">{{ item.count }}</span>
                </button>

                <!-- 表情选择器 -->
                <div class="page-comment__picker">
                  <button
                    class="page-comment__reaction-add"
                    :aria-label="t('pageComment.addReaction')"
                    @click.stop="togglePicker(root.id)"
                  >
                    😀
                  </button>
                  <div
                    v-if="pickerOpenFor === root.id"
                    class="page-comment__picker-panel"
                  >
                    <button
                      v-for="emoji in PRESET_EMOJIS"
                      :key="emoji"
                      class="page-comment__picker-item"
                      :class="{
                        'page-comment__picker-item--active': hasMyReaction(root.id, emoji),
                      }"
                      @click.stop="toggleReaction(root, emoji)"
                    >
                      {{ emoji }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 嵌套回复 -->
          <div
            v-if="repliesByParent.get(root.id)?.length"
            class="page-comment__replies"
          >
            <div
              v-for="reply in repliesByParent.get(root.id)"
              :key="reply.id"
              class="page-comment__row page-comment__row--reply"
              :class="{ 'page-comment__row--hidden': isHidden(reply) }"
            >
              <AppAvatar
                :name="reply.author_name || 'U'"
                :src="reply.author_avatar || ''"
                size="sm"
              />
              <div class="page-comment__body">
                <div class="page-comment__meta">
                  <span class="page-comment__author">
                    {{ reply.author_name || t("pageComment.anonymous") }}
                  </span>
                  <span class="page-comment__time">{{ formatTime(reply.created_at) }}</span>
                  <span v-if="reply.is_edited" class="page-comment__edited">
                    ({{ t("pageComment.edited") }})
                  </span>
                  <AppBadge v-if="isHidden(reply)" variant="danger">
                    {{ t("pageComment.hidden") }}
                  </AppBadge>
                </div>

                <!-- 隐藏原因（仅审核者可见） -->
                <div
                  v-if="isHidden(reply) && canModerate"
                  class="page-comment__hidden-reason"
                >
                  <span class="page-comment__hidden-reason-label">
                    {{ t("pageComment.hiddenReasonLabel") }}
                  </span>
                  <span class="page-comment__hidden-reason-text">
                    {{ hiddenReason(reply) }}
                  </span>
                  <span v-if="reply.moderated_at" class="page-comment__time">
                    {{ formatTime(reply.moderated_at) }}
                  </span>
                </div>

                <div v-if="reply.block_id" class="page-comment__quote">
                  <span class="page-comment__quote-label">
                    {{ t("pageComment.inlineQuote") }}
                  </span>
                  <span class="page-comment__quote-text">
                    {{ reply.quote_text || reply.block_id }}
                  </span>
                </div>

                <div
                  v-if="editingComment?.id === reply.id"
                  class="page-comment__edit"
                >
                  <textarea
                    :ref="setEditInput"
                    v-model="editText"
                    class="page-comment__textarea"
                    rows="3"
                    @keydown="onEditKeydown"
                  />
                  <div class="page-comment__edit-actions">
                    <AppButton
                      size="sm"
                      :loading="submitting"
                      :disabled="!editText.trim() || submitting"
                      @click="submitEdit"
                    >
                      {{ t("pageComment.save") }}
                    </AppButton>
                    <AppButton
                      size="sm"
                      variant="secondary"
                      :disabled="submitting"
                      @click="cancelEdit"
                    >
                      {{ t("pageComment.cancel") }}
                    </AppButton>
                  </div>
                </div>
                <p v-else class="page-comment__content">{{ reply.content }}</p>

                <div class="page-comment__actions">
                  <button class="page-comment__action" @click="startReply(root)">
                    {{ t("pageComment.reply") }}
                  </button>
                  <template v-if="isMine(reply)">
                    <button class="page-comment__action" @click="startEdit(reply)">
                      {{ t("pageComment.edit") }}
                    </button>
                    <button
                      class="page-comment__action page-comment__action--danger"
                      @click="handleDelete(reply)"
                    >
                      {{ t("pageComment.delete") }}
                    </button>
                  </template>
                  <button
                    v-if="canModerate && !isHidden(reply)"
                    class="page-comment__action page-comment__action--danger"
                    @click="handleModerate(reply, 'hide')"
                  >
                    {{ t("pageComment.hide") }}
                  </button>
                  <button
                    v-if="canModerate && isHidden(reply)"
                    class="page-comment__action"
                    @click="handleModerate(reply, 'unhide')"
                  >
                    {{ t("pageComment.unhide") }}
                  </button>
                </div>

                <div class="page-comment__reactions">
                  <button
                    v-for="item in reactionSummary(reply.id)"
                    :key="item.emoji"
                    class="page-comment__reaction"
                    :class="{
                      'page-comment__reaction--active': hasMyReaction(reply.id, item.emoji),
                    }"
                    @click.stop="toggleReaction(reply, item.emoji)"
                  >
                    <span class="page-comment__reaction-emoji">{{ item.emoji }}</span>
                    <span class="page-comment__reaction-count">{{ item.count }}</span>
                  </button>

                  <div class="page-comment__picker">
                    <button
                      class="page-comment__reaction-add"
                      :aria-label="t('pageComment.addReaction')"
                      @click.stop="togglePicker(reply.id)"
                    >
                      😀
                    </button>
                    <div
                      v-if="pickerOpenFor === reply.id"
                      class="page-comment__picker-panel"
                    >
                      <button
                        v-for="emoji in PRESET_EMOJIS"
                        :key="emoji"
                        class="page-comment__picker-item"
                        :class="{
                          'page-comment__picker-item--active': hasMyReaction(reply.id, emoji),
                        }"
                        @click.stop="toggleReaction(reply, emoji)"
                      >
                        {{ emoji }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 回复输入框（内联） -->
          <div
            v-if="replyingTo?.id === root.id"
            class="page-comment__reply-form"
          >
            <div class="page-comment__reply-hint">
              {{ t("pageComment.replyingTo", { name: root.author_name || t("pageComment.anonymous") }) }}
            </div>
            <textarea
              :ref="setReplyInput"
              v-model="replyText"
              class="page-comment__textarea"
              rows="2"
              :placeholder="t('pageComment.replyPlaceholder')"
              @keydown="onReplyKeydown($event, root)"
            />
            <div class="page-comment__edit-actions">
              <AppButton
                size="sm"
                :loading="submitting"
                :disabled="!replyText.trim() || submitting"
                @click="submitReply(root)"
              >
                {{ submitting ? t("pageComment.submitting") : t("pageComment.reply") }}
              </AppButton>
              <AppButton size="sm" variant="secondary" :disabled="submitting" @click="cancelReply">
                {{ t("pageComment.cancel") }}
              </AppButton>
            </div>
          </div>
        </article>
      </template>

      <!-- 已有数据时的错误横幅 -->
      <div v-if="error && comments.length > 0" class="page-comment__error-banner">
        {{ error }}
        <button class="page-comment__retry" @click="loadComments">
          {{ t("pageComment.retry") }}
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.page-comment {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-subtle);
}

.page-comment--compact {
  margin-top: 0;
  padding-top: 0;
  border-top: none;
}

.page-comment__header {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.page-comment__title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.page-comment__count {
  padding: 1px 7px;
  border-radius: 10px;
  background: var(--surface-3);
  color: var(--text-tertiary);
  font-size: 12px;
  font-weight: 500;
}

/* ---- 新建输入 ---- */
.page-comment__new {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.page-comment__new-body {
  flex: 1;
  min-width: 0;
}

.page-comment__textarea {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--surface-1);
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  box-sizing: border-box;
}

.page-comment__textarea:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.page-comment__textarea::placeholder {
  color: var(--text-tertiary);
}

.page-comment__new-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  gap: 12px;
}

.page-comment__hint {
  flex: 1;
  font-size: 11px;
  color: var(--text-tertiary);
}

/* ---- 列表 ---- */
.page-comment__list {
  display: flex;
  flex-direction: column;
}

.page-comment__item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.page-comment__item:last-child {
  border-bottom: none;
}

.page-comment__item--resolved {
  opacity: 0.72;
}

/* 被审核隐藏的评论：降透明度 + 红色虚线提示，与普通评论区分 */
.page-comment__item--hidden,
.page-comment__row--hidden {
  opacity: 0.6;
}

.page-comment__item--hidden .page-comment__content,
.page-comment__row--hidden .page-comment__content {
  color: var(--text-tertiary);
  font-style: italic;
}

.page-comment__hidden-reason {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin: 4px 0 6px;
  padding: 4px 8px;
  border: 1px dashed var(--danger-300);
  border-radius: var(--radius-sm, 4px);
  background: var(--danger-50, rgba(239, 68, 68, 0.05));
  font-size: 11px;
}

.page-comment__hidden-reason-label {
  font-weight: 600;
  color: var(--danger-500);
}

.page-comment__hidden-reason-text {
  color: var(--text-secondary);
  word-break: break-word;
}

.page-comment__row {
  display: flex;
  gap: 10px;
}

.page-comment__body {
  flex: 1;
  min-width: 0;
}

.page-comment__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.page-comment__author {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.page-comment__time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.page-comment__edited {
  font-size: 11px;
  font-style: italic;
  color: var(--text-tertiary);
}

.page-comment__quote {
  margin: 4px 0 6px;
  padding: 6px 10px;
  border-left: 3px solid var(--brand-300);
  background: var(--surface-2);
  border-radius: var(--radius-sm, 4px);
  font-size: 12px;
}

.page-comment__quote-label {
  margin-right: 6px;
  color: var(--brand-600);
  font-weight: 500;
}

.page-comment__quote-text {
  color: var(--text-tertiary);
  word-break: break-word;
}

.page-comment__content {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-word;
}

.page-comment__edit {
  margin-top: 4px;
}

.page-comment__edit-actions {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}

.page-comment__actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.page-comment__action {
  padding: 0;
  border: none;
  background: none;
  font-family: inherit;
  font-size: 11px;
  color: var(--text-tertiary);
  cursor: pointer;
}

.page-comment__action:hover {
  color: var(--brand-500);
}

.page-comment__action--danger:hover {
  color: var(--danger-500);
}

.page-comment__action--resolve:hover {
  color: var(--success-600);
}

/* ---- 表情回应 ---- */
.page-comment__reactions {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.page-comment__reaction {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border: 1px solid var(--border-subtle);
  border-radius: 999px;
  background: var(--surface-1);
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.page-comment__reaction:hover {
  border-color: var(--brand-300);
  background: var(--brand-50, rgba(59, 130, 246, 0.06));
}

.page-comment__reaction--active {
  border-color: var(--brand-500);
  background: var(--brand-50, rgba(59, 130, 246, 0.1));
}

.page-comment__reaction-emoji {
  font-size: 13px;
  line-height: 1;
}

.page-comment__reaction-count {
  font-size: 11px;
  color: var(--text-tertiary);
}

.page-comment__reaction--active .page-comment__reaction-count {
  color: var(--brand-600);
  font-weight: 600;
}

.page-comment__picker {
  position: relative;
}

.page-comment__reaction-add {
  padding: 2px 8px;
  border: 1px dashed var(--border-default);
  border-radius: 999px;
  background: none;
  cursor: pointer;
  font-size: 13px;
  line-height: 1;
  opacity: 0.6;
  transition: opacity 0.15s;
}

.page-comment__reaction-add:hover {
  opacity: 1;
}

.page-comment__picker-panel {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  z-index: 20;
  display: flex;
  gap: 4px;
  padding: 6px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.page-comment__picker-item {
  padding: 4px;
  border: none;
  border-radius: 4px;
  background: none;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  transition: background 0.12s, transform 0.12s;
}

.page-comment__picker-item:hover {
  background: var(--surface-2);
  transform: scale(1.15);
}

.page-comment__picker-item--active {
  background: var(--brand-50, rgba(59, 130, 246, 0.12));
}

/* ---- 回复 ---- */
.page-comment__replies {
  margin-left: 34px;
  margin-top: 4px;
  padding-left: 12px;
  border-left: 2px solid var(--border-subtle);
}

.page-comment__row--reply {
  padding: 8px 0;
}

.page-comment__reply-form {
  margin-left: 34px;
  margin-top: 10px;
  padding: 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-2);
}

.page-comment__reply-hint {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 500;
  color: var(--brand-500);
}

/* ---- 错误横幅 ---- */
.page-comment__error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: var(--radius-sm, 4px);
  background: var(--danger-50, rgba(239, 68, 68, 0.06));
  color: var(--danger-500);
  font-size: 12px;
}

.page-comment__retry {
  padding: 0;
  border: none;
  background: none;
  color: var(--brand-500);
  font-family: inherit;
  font-size: 12px;
  text-decoration: underline;
  cursor: pointer;
}
</style>
