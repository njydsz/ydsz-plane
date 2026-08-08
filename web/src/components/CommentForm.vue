<script setup lang="ts">
/**
 * CommentForm — 评论输入表单（基于 TipTap 富文本编辑器）
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
import RichTextEditor from "./RichTextEditor.vue";

const props = defineProps<{
  loading?: boolean;
  initialValue?: string;
  parentId?: number | null;
  workspaceId?: number | string;
  projectId?: number | string;
}>();

const emit = defineEmits<{
  submit: [payload: { content_html: string; content_json: string; content_stripped: string; parent_id?: number | null }];
  cancel: [];
}>();

const contentHtml = ref("");
const contentJson = ref("{}");
const editorRef = ref<InstanceType<typeof RichTextEditor> | null>(null);

const isEditing = computed(() => !!props.initialValue);
const isReplying = computed(() => props.parentId != null);

const placeholder = computed(() => {
  if (isEditing.value) return "编辑评论...";
  if (isReplying.value) return "输入回复...";
  return "添加评论... 支持富文本格式";
});

const canSubmit = computed(() => {
  // 从 HTML 中提取纯文本判断是否有内容
  const tmp = document.createElement("div");
  tmp.innerHTML = contentHtml.value;
  return (tmp.textContent?.trim().length ?? 0) > 0 && !props.loading;
});

watch(
  () => props.initialValue,
  (val) => {
    if (val) {
      contentHtml.value = val;
    }
  },
  { immediate: true },
);

function handleSubmit() {
  if (!canSubmit.value) return;

  const tmp = document.createElement("div");
  tmp.innerHTML = contentHtml.value;
  const stripped = tmp.textContent?.trim() || "";

  emit("submit", {
    content_html: contentHtml.value,
    content_json: contentJson.value,
    content_stripped: stripped,
    parent_id: props.parentId,
  });
  contentHtml.value = "";
  contentJson.value = "{}";
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
  editorRef.value?.editor?.chain().focus().run();
}

defineExpose({ focus });
</script>

<template>
  <div class="comment-form" @keydown="handleKeydown">
    <RichTextEditor
      ref="editorRef"
      v-model:content-html="contentHtml"
      v-model:content-json="contentJson"
      :placeholder="placeholder"
      :workspace-id="workspaceId"
      :project-id="projectId"
      compact
    />

    <div class="comment-form__footer">
      <span class="comment-form__hint">
        <template v-if="isEditing">Esc 取消 · </template>
        <template v-if="isReplying">Esc 取消 · </template>
        Ctrl+Enter 提交
      </span>

      <div class="comment-form__actions">
        <button
          v-if="isEditing || isReplying"
          class="btn btn--sm"
          :disabled="loading"
          @click="emit('cancel')"
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
