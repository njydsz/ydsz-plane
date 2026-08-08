<script setup lang="ts">
/**
 * ModuleFormModal — 新增 / 编辑模块的统一表单弹窗。
 *
 * 字段：模块名称（必填）、描述、负责人（工作空间成员下拉）。
 *
 * 采用受控模式：父组件传入 module（非空为编辑，空为新建），
 * 提交后通过 submit 事件把干净的 CreateModuleInput / UpdateModuleInput
 * 交给父组件决定如何调用 API。
 */
import { computed, reactive, ref, watch } from "vue";

import {
  type CreateModuleInput,
  type Module,
  type UpdateModuleInput,
} from "@/api/services/module";
import type { Member } from "@/api/services/workspace";

const props = defineProps<{
  /** 是否可见 */
  visible: boolean;
  /** 编辑的模块对象；null/undefined 表示新建 */
  module?: Module | null;
  /** 工作空间成员列表（负责人下拉数据源） */
  members: Member[];
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "submit", payload: CreateModuleInput | UpdateModuleInput): void;
}>();

/** 是否为编辑模式 */
const isEdit = computed(() => !!props.module?.id);

const submitting = ref(false);
const errorMsg = reactive({ text: "" });

// ---- 表单态 ----
const form = reactive({
  name: "",
  description: "",
  lead_id: undefined as number | undefined,
});

/** visible 变化时同步表单初始值 */
watch(
  () => props.visible,
  (vis) => {
    if (vis) {
      errorMsg.text = "";
      if (props.module) {
        form.name = props.module.name;
        form.description = props.module.description ?? "";
        form.lead_id = props.module.lead_id;
      } else {
        form.name = "";
        form.description = "";
        form.lead_id = undefined;
      }
    }
  },
);

function handleSubmit() {
  if (!form.name.trim()) {
    errorMsg.text = "模块名称不能为空";
    return;
  }
  submitting.value = false; // reset
  submitting.value = true;
  try {
    if (isEdit.value) {
      const payload: UpdateModuleInput = {
        name: form.name.trim(),
      };
      if (form.description.trim()) {
        payload.description = form.description.trim();
      } else {
        payload.description = "";
      }
      payload.lead_id = form.lead_id;
      emit("submit", payload);
    } else {
      const payload: CreateModuleInput = {
        name: form.name.trim(),
      };
      if (form.description.trim()) payload.description = form.description.trim();
      if (form.lead_id) payload.lead_id = form.lead_id;
      emit("submit", payload);
    }
  } finally {
    submitting.value = false;
  }
}

function onMemberSelect(e: Event) {
  const val = (e.target as HTMLSelectElement).value;
  form.lead_id = val ? Number(val) : undefined;
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="modal-overlay"
      @click.self="emit('close')"
    >
      <div
        class="modal"
        role="dialog"
        aria-modal="true"
        :aria-label="isEdit ? '编辑模块' : '新建模块'"
      >
        <header class="modal__header">
          <h2 class="modal__title">{{ isEdit ? "编辑模块" : "新建模块" }}</h2>
          <button
            class="modal__close"
            aria-label="关闭"
            @click="emit('close')"
          >
            &times;
          </button>
        </header>

        <div class="modal__body">
          <p v-if="errorMsg.text" class="modal__error">{{ errorMsg.text }}</p>

          <form class="form" @submit.prevent="handleSubmit">
            <!-- 模块名称 -->
            <label class="form__field">
              <span class="form__label">
                模块名称 <span class="form__required">*</span>
              </span>
              <input
                v-model="form.name"
                type="text"
                maxlength="128"
                placeholder="例如：用户模块、支付模块"
                autocomplete="off"
                class="form__input"
                autofocus
              />
            </label>

            <!-- 描述 -->
            <label class="form__field">
              <span class="form__label">描述</span>
              <textarea
                v-model="form.description"
                maxlength="500"
                rows="3"
                placeholder="模块用途简介（可选）"
                class="form__input form__textarea"
              ></textarea>
            </label>

            <!-- 负责人 -->
            <label class="form__field">
              <span class="form__label">负责人</span>
              <select
                class="form__input form__select"
                :value="form.lead_id ?? ''"
                @change="onMemberSelect"
              >
                <option value="">未指定</option>
                <option
                  v-for="m in members"
                  :key="m.id"
                  :value="m.id"
                >
                  {{ m.display_name }} ({{ m.email }})
                </option>
              </select>
            </label>
          </form>
        </div>

        <footer class="modal__footer">
          <button
            type="button"
            class="btn btn--secondary"
            @click="emit('close')"
          >
            取消
          </button>
          <button
            type="submit"
            class="btn btn--primary"
            :disabled="submitting || !form.name.trim()"
            @click="handleSubmit"
          >
            {{ submitting ? "保存中..." : isEdit ? "保存修改" : "创建" }}
          </button>
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* ---------- overlay ---------- */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
  animation: fadeIn 0.15s ease;
}

.modal {
  width: calc(100% - 32px);
  max-width: 480px;
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
  animation: slideUp 0.2s ease;
}

/* ---------- header ---------- */
.modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.modal__title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.modal__close {
  background: none;
  border: none;
  font-size: 22px;
  color: var(--text-tertiary);
  cursor: pointer;
  line-height: 1;
  padding: 0;
}

.modal__close:hover {
  color: var(--text-primary);
}

/* ---------- body ---------- */
.modal__body {
  padding: 20px 24px;
}

.modal__error {
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--danger-500);
}

/* ---------- form ---------- */
.form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form__label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.form__required {
  color: var(--danger-500);
}

.form__input {
  padding: 8px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  background: var(--surface-1);
  outline: none;
  font-family: inherit;
  transition: border-color 0.15s;
}

.form__input:focus {
  border-color: var(--brand-500);
  box-shadow: 0 0 0 3px var(--brand-50, rgba(59, 130, 246, 0.1));
}

.form__textarea {
  resize: vertical;
  min-height: 60px;
}

.form__select {
  appearance: auto;
}

/* ---------- footer ---------- */
.modal__footer {
  padding: 0 24px 20px;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* ---------- buttons ---------- */
.btn {
  padding: 7px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: background 0.15s, border-color 0.15s, opacity 0.15s;
  font-family: inherit;
}

.btn--primary {
  background: var(--brand-500);
  border-color: var(--brand-500);
  color: var(--text-on-brand);
}

.btn--primary:hover:not(:disabled) {
  background: var(--brand-600);
}

.btn--primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn--secondary {
  background: var(--surface-1);
  border-color: var(--border-default);
  color: var(--text-secondary);
}

.btn--secondary:hover {
  background: var(--surface-hover, #f9fafb);
}

/* ---------- animations ---------- */
@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to   { opacity: 1; transform: translateY(0); }
}
</style>
