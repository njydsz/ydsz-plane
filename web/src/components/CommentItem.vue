<script setup lang="ts">
/**
 * CommentItem — 单条评论展示（含编辑/删除操作）
 *
 * Props:
 *   comment      — 评论数据
 *   currentUserId — 当前登录用户 ID（用于判断编辑/删除权限）
 *   replyHint     — 展示 "回复 @xxx" 提示行
 *
 * Events:
 *   @edit   — 用户点击「编辑」
 *   @delete — 用户点击「删除」（组件内 confirm 确认）
 *   @reply  — 用户点击「回复」
 */
import { computed } from "vue";
import AppAvatar from "./AppAvatar.vue";
import type { IssueComment } from "@/api/services/issue";

const props = defineProps<{
  comment: IssueComment;
  currentUserId: number;
  replyHint?: string;
}>();

const emit = defineEmits<{
  edit: [comment: IssueComment];
  delete: [comment: IssueComment];
  reply: [comment: IssueComment];
}>();

const isMine = computed(() => props.comment.created_by === props.currentUserId);

const timeAgo = computed(() => {
  const d = new Date(props.comment.created_at);
  const now = Date.now();
  const diff = now - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "刚刚";
  if (mins < 60) return `${mins}分钟前`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}小时前`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}天前`;
  return d.toLocaleDateString("zh-CN");
});

function handleDelete() {
  if (confirm("确定删除该评论？")) {
    emit("delete", props.comment);
  }
}
</script>

<template>
  <div class="comment-item" :class="{ 'comment-item--mine': isMine }">
    <AppAvatar
      :name="comment.creator_name || 'U'"
      :src="comment.creator_avatar || undefined"
      size="sm"
    />

    <div class="comment-item__body">
      <div class="comment-item__header">
        <span class="comment-item__author">{{ comment.creator_name || "匿名用户" }}</span>
        <span class="comment-item__time">{{ timeAgo }}</span>
        <span v-if="comment.is_edited" class="comment-item__edited">(已编辑)</span>
      </div>

      <div v-if="replyHint" class="comment-item__reply-hint">
        回复 @{{ replyHint }}
      </div>

      <div
        class="comment-item__content"
        v-html="comment.content_html"
      ></div>

      <div class="comment-item__actions">
        <button class="comment-item__action" @click="emit('reply', comment)">回复</button>
        <template v-if="isMine">
          <button class="comment-item__action" @click="emit('edit', comment)">编辑</button>
          <button class="comment-item__action comment-item__action--danger" @click="handleDelete">删除</button>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comment-item {
  display: flex;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-item--mine {
  background: var(--brand-50, rgba(59,130,246,0.04));
  margin: 0 -12px;
  padding: 12px;
  border-radius: var(--radius-md, 8px);
}

.comment-item__body {
  flex: 1;
  min-width: 0;
}

.comment-item__header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.comment-item__author {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.comment-item__time {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.comment-item__edited {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  font-style: italic;
}

.comment-item__reply-hint {
  font-size: 11px;
  color: var(--brand-500, #3b82f6);
  margin-bottom: 4px;
}

.comment-item__content {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary, #4b5563);
  word-break: break-word;
}

.comment-item__content :deep(p) {
  margin: 0 0 8px;
}

.comment-item__content :deep(p:last-child) {
  margin-bottom: 0;
}

.comment-item__content :deep(a) {
  color: var(--brand-500, #3b82f6);
}

.comment-item__content :deep(code) {
  background: var(--surface-3, #f3f4f6);
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
  font-family: var(--font-mono, 'Consolas', monospace);
}

.comment-item__content :deep(pre) {
  background: var(--surface-3, #f3f4f6);
  padding: 8px 12px;
  border-radius: var(--radius-sm, 4px);
  overflow-x: auto;
  font-size: 12px;
}

.comment-item__actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.comment-item__action {
  background: none;
  border: none;
  padding: 0;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  font-family: inherit;
}

.comment-item__action:hover {
  color: var(--brand-500, #3b82f6);
}

.comment-item__action--danger:hover {
  color: var(--danger-500, #ef4444);
}
</style>
