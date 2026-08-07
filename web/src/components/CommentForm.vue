<script setup lang="ts">
/**
 * CommentForm — 评论输入表单（支持编辑模式）
 *
 * Props:
 *   loading      — 是否提交中
 *   initialValue — 编辑模式的初始内容（HTML），非空时进入编辑模式
 *   parentId     — 回复模式下的父评论 ID
 *
 * Events:
 *   @submit — content_html, content_stripped, parent_id
 *   @cancel — 取消编辑/回复
 */
import { computed, ref, watch } from "vue";

const props = defineProps<{
  loading?: boolean;
  initialValue?: string;
  parentId?: number | null;
}>();

const emit = defineEmits<{
  submit: [payload: { content_html: string; content_stripped: string; parent_id?: number | null }];
  cancel: [];
}>();

const content = ref("");
const textareaRef = ref<HTMLTextAreaElement | null>(null);

const isEditing = computed(() => !!props.initialValue);
const isReplying = computed(() => props.parentId != null);

const placeholder = computed(() => {
  if (isEditing.value) return "编辑评论...";
  if (isReplying.value) return "输入回复...";
  return "添加评论... 支持 Markdown";
});

const canSubmit = computed(() => content.value.trim().length > 0 && !props.loading);

watch(
  () => props.initialValue,
  (val) => {
    if (val) {
      content.value = stripBasicHtml(val);
    }
  },
  { immediate: true },
);

function stripBasicHtml(html: string): string {
  return html.replace(/<br\s*\/?>/gi, "\n").replace(/<[^>]+>/g, "");
}

function handleSubmit() {
  const trimmed = content.value.trim();
  if (!trimmed || props.loading) return;

  // 简单 Markdown → HTML 转换
  const html = simpleMarkdownToHtml(trimmed);

  emit("submit", {
    content_html: html,
    content_stripped: trimmed,
    parent_id: props.parentId,
  });
  content.value = "";
}

function simpleMarkdownToHtml(text: string): string {
  return text
    .split("\n\n")
    .map((p) => {
      let line = p
        .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
        .replace(/\*(.+?)\*/g, "<em>$1</em>")
        .replace(/`([^`]+)`/g, "<code>$1</code>")
        .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');
      return `<p>${line}</p>`;
    })
    .join("");
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
    e.preventDefault();
    handleSubmit();
  }
  if (e.key === "Escape") {
    if (isEditing.value || isReplying.value) {
      e.preventDefault();
      emit("cancel");
    }
  }
}

function focus() {
  textareaRef.value?.focus();
}

defineExpose({ focus });
</script>

<template>
  <div class="comment-form">
    <textarea
      ref="textareaRef"
      v-model="content"
      class="comment-form__textarea"
      :placeholder="placeholder"
      rows="3"
      @keydown="handleKeydown"
    ></textarea>

    <div class="comment-form__footer">
      <span class="comment-form__hint">
        <template v-if="isEditing">Esc 取消 · </template>
        <template v-if="isReplying">Esc 取消 · </template>
        Ctrl+Enter 提交 · 支持 **粗体** `代码` [链接](url)
      </span>

      <div class="comment-form__actions">
        <button
          v-if="isEditing || isReplying"
          class="btn btn--sm"
          @click="emit('cancel')"
          :disabled="loading"
        >
          取消
        </button>
        <button
          class="btn btn--sm btn--primary"
          :disabled="!canSubmit"
          @click="handleSubmit"
        >
          {{ loading ? "提交中..." : isEditing ? "保存" : "评论" }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.comment-form {
  margin-bottom: 16px;
}

.comment-form__textarea {
  width: 100%;
  min-height: 80px;
  padding: 10px 12px;
  font-size: 13px;
  font-family: inherit;
  line-height: 1.5;
  color: var(--text-primary, #1f2937);
  background: var(--surface-2, #f9fafb);
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-md, 8px);
  outline: none;
  resize: vertical;
  transition: border-color 0.15s;
}

.comment-form__textarea:focus {
  border-color: var(--brand-500, #3b82f6);
  background: var(--surface-1, #fff);
}

.comment-form__textarea::placeholder {
  color: var(--text-tertiary, #9ca3af);
}

.comment-form__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 8px;
  gap: 12px;
}

.comment-form__hint {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  flex: 1;
}

.comment-form__actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.btn--sm {
  padding: 4px 12px;
  font-size: 12px;
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
  border: 1px solid var(--border-default, #d1d5db);
  background: var(--surface-1, #fff);
  color: var(--text-secondary, #4b5563);
  font-family: inherit;
}

.btn--sm:hover:not(:disabled) {
  background: var(--surface-3, #f3f4f6);
}

.btn--sm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--primary {
  background: var(--brand-500, #3b82f6);
  color: #fff;
  border-color: var(--brand-500, #3b82f6);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600, #2563eb);
}
</style>
