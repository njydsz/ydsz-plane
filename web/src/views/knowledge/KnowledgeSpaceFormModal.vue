<script setup lang="ts">
/**
 * KnowledgeSpaceFormModal — 知识库空间新建/编辑弹窗。
 *
 * 字段：name（必填）、slug（auto from name）、description、
 * is_private 开关、default_permission 下拉、cover_image URL。
 *
 * emit submit / cancel。
 */
import { ref, watch } from "vue";

import type { SpacePermission, UpdateSpaceInput } from "@/api/services/knowledge";
import { AppModal } from "@/components";

/* ===== Props / Emit ===== */
const props = defineProps<{
  visible: boolean;
  /** 编辑模式：传入已存在的空间数据 */
  space?: {
    id: number;
    name: string;
    slug: string;
    description?: string;
    is_private: boolean;
    default_permission: SpacePermission;
    cover_image?: string;
  };
}>();

const emit = defineEmits<{
  (e: "submit", input: UpdateSpaceInput & { slug: string }): void;
  (e: "cancel"): void;
}>();

/* ===== 表单状态 ===== */
const formName = ref("");
const formSlug = ref("");
const slugTouched = ref(false);
const formDescription = ref("");
const formIsPrivate = ref(true);
const formPermission = ref<SpacePermission>("viewer");
const formCoverImage = ref("");
const submitting = ref(false);

const permissionOptions: Array<{ value: SpacePermission; label: string }> = [
  { value: "viewer", label: "只读" },
  { value: "editor", label: "可编辑" },
  { value: "admin", label: "管理员" },
  { value: "owner", label: "所有者" },
];

/** slugify */
function slugify(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

/* 初始化表单 */
watch(() => props.visible, (v) => {
  if (v) {
    submitting.value = false;
    if (props.space) {
      formName.value = props.space.name;
      formSlug.value = props.space.slug;
      formDescription.value = props.space.description ?? "";
      formIsPrivate.value = props.space.is_private;
      formPermission.value = props.space.default_permission;
      formCoverImage.value = props.space.cover_image ?? "";
      slugTouched.value = true;
    } else {
      formName.value = "";
      formSlug.value = "";
      slugTouched.value = false;
      formDescription.value = "";
      formIsPrivate.value = true;
      formPermission.value = "viewer";
      formCoverImage.value = "";
    }
  }
});

/* 名称变化时自动推导 slug */
watch(formName, (val) => {
  if (!slugTouched.value) {
    formSlug.value = slugify(val);
  }
});

function onSlugInput() {
  slugTouched.value = true;
}

function onSubmit() {
  if (!formName.value.trim()) return;
  const slug = formSlug.value.trim() || slugify(formName.value);
  if (!slug) return;

  submitting.value = true;
  emit("submit", {
    name: formName.value.trim(),
    slug,
    description: formDescription.value.trim(),
    is_private: formIsPrivate.value,
    default_permission: formPermission.value,
    cover_image: formCoverImage.value.trim() || undefined,
  });
}

function onCancel() {
  emit("cancel");
}
</script>

<template>
  <AppModal
    :visible="visible"
    :title="space ? '编辑空间' : '新建空间'"
    width="520px"
    @close="onCancel"
  >
    <div class="space-form">
      <div class="form-field">
        <label class="form-label">
          空间名称 <span class="form-label__required">*</span>
        </label>
        <input
          v-model="formName"
          class="form-input"
          placeholder="如：产品文档中心"
          @keydown.enter.prevent="onSubmit"
        />
      </div>

      <div class="form-field">
        <label class="form-label">链接标识 (slug)</label>
        <input
          v-model="formSlug"
          class="form-input form-input--mono"
          placeholder="自动根据名称生成"
          @input="onSlugInput"
        />
        <span class="form-hint">用于 URL 路径，仅允许小写字母、数字与连字符</span>
      </div>

      <div class="form-field">
        <label class="form-label">描述</label>
        <textarea
          v-model="formDescription"
          class="form-textarea"
          rows="3"
          placeholder="简要描述该空间的用途..."
        />
      </div>

      <div class="form-field">
        <label class="form-label">默认权限</label>
        <select v-model="formPermission" class="form-input">
          <option v-for="opt in permissionOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
        <span class="form-hint">新成员访问此空间时的默认权限级别</span>
      </div>

      <div class="form-field">
        <label class="form-label">封面图片 URL（可选）</label>
        <input
          v-model="formCoverImage"
          class="form-input"
          placeholder="https://..."
        />
      </div>

      <label class="form-check">
        <input v-model="formIsPrivate" type="checkbox" />
        <span class="form-check__label">私有空间（仅邀请成员可见）</span>
      </label>
    </div>

    <template #footer>
      <button class="btn" :disabled="submitting" @click="onCancel">取消</button>
      <button
        class="btn btn--primary"
        :disabled="submitting || !formName.trim()"
        @click="onSubmit"
      >
        {{ submitting ? "提交中..." : space ? "保存修改" : "创建空间" }}
      </button>
    </template>
  </AppModal>
</template>

<style scoped>
.space-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.form-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}
.form-label__required {
  color: var(--danger-500, #ef4444);
}
.form-input {
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.form-input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}
.form-input--mono {
  font-family: var(--font-mono, monospace);
}
.form-textarea {
  padding: 8px 10px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  outline: none;
  resize: vertical;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.form-textarea:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 2px var(--brand-50);
}
.form-hint {
  font-size: 11px;
  color: var(--text-tertiary);
}
.form-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}
.form-check__label {
  user-select: none;
}

.btn {
  padding: 8px 16px;
  font-size: 13px;
  font-family: inherit;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.btn:hover {
  border-color: var(--brand-500);
  color: var(--brand-600);
}
.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
  font-weight: 500;
}
.btn--primary:hover {
  background: var(--brand-600);
  color: var(--text-on-brand);
}
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
