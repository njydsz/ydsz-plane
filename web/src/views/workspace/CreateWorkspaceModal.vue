<script setup lang="ts">
/**
 * 创建工作空间弹窗 — 名称/slug/时区/语言表单。
 */

import { computed, ref, watch } from "vue";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const emit = defineEmits<{
  (e: "close"): void;
  (e: "created", ws: Workspace): void;
}>();

const name = ref("");
const slug = ref("");
const submitting = ref(false);
const error = ref("");

// 自动从名称派生 slug
watch(name, (v) => {
  if (!slugTouched.value) {
    slug.value = autoSlug(v);
  }
});

const slugTouched = ref(false);
function autoSlug(s: string): string {
  return s
    .toLowerCase()
    .trim()
    .replace(/[^\p{L}\p{N}\u4e00-\u9fff]+/gu, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60);
}

const slugHint = computed(() => (slug.value ? `链接标识：${slug.value}` : ""));

async function submit() {
  error.value = "";
  if (!name.value.trim()) {
    error.value = "请输入工作空间名称";
    return;
  }
  submitting.value = true;
  try {
    const ws = await workspaceApi.create({
      name: name.value.trim(),
      slug: slug.value || undefined,
      language: "zh-CN",
    });
    emit("created", ws);
  } catch (e: any) {
    error.value = e.message ?? "创建失败";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <header class="modal__header">
        <h2>创建工作空间</h2>
        <button class="modal__close" @click="emit('close')">×</button>
      </header>

      <div class="modal__body">
        <p class="desc">
          工作空间是团队共享的项目和成员的容器。创建后自动成为该空间所有者。
        </p>

        <label class="field">
          <span class="field__label">名称</span>
          <input
            v-model="name"
            class="field__input"
            placeholder="例如：核心产品团队"
            maxlength="80"
            autofocus
          />
        </label>

        <label class="field">
          <span class="field__label">链接标识（可选）</span>
          <input
            v-model="slug"
            class="field__input"
            placeholder="例如：core（会自动从名称生成）"
            maxlength="60"
            @input="slugTouched = true"
          />
          <span v-if="slugHint" class="field__hint">{{ slugHint }}</span>
        </label>

        <div v-if="error" class="modal__error">{{ error }}</div>
      </div>

      <footer class="modal__footer">
        <button class="btn" @click="emit('close')">取消</button>
        <button class="btn btn--primary" :disabled="submitting" @click="submit">
          {{ submitting ? "创建中..." : "创建" }}
        </button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  width: 480px;
  max-width: 95vw;
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
  overflow: hidden;
}

.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-subtle);
}

.modal__header h2 {
  font-size: 16px;
  margin: 0;
}

.modal__close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-tertiary);
  cursor: pointer;
  line-height: 1;
}

.modal__body {
  padding: 24px;
}

.desc {
  color: var(--text-tertiary);
  font-size: 13px;
  margin: 0 0 20px;
}

.field {
  display: block;
  margin-bottom: 16px;
}

.field__label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.field__input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 14px;
  font-family: inherit;
}

.field__input:focus {
  outline: none;
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50);
}

.field__hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.modal__error {
  color: var(--danger-500);
  font-size: 13px;
  padding: 8px 0;
}

.modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 16px 24px;
  border-top: 1px solid var(--border-subtle);
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border-default);
  background: var(--surface-1);
  color: var(--text-secondary);
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
