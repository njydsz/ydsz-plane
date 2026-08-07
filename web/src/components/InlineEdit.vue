<script setup lang="ts">
/**
 * InlineEdit — 内联编辑组件（对标 Plane 的 inline edit 交互）。
 *
 * 交互：
 *  - 单击文本进入编辑态（可配置 trigger: 'click' | 'dblclick'）
 *  - Enter 提交保存，Esc 取消，失焦提交保存
 *  - 编辑态自动聚焦并全选文本
 *  - 可选校验函数，校验失败时禁止提交并保留编辑态
 *
 * Props:
 *   modelValue: 当前文本值
 *   trigger:    'click' | 'dblclick'
 *   placeholder: 空值占位文案
 *   maxLength:   最大长度
 *   validate:    (value) => string | null，返回错误信息或 null
 *   disabled:    禁用编辑
 *
 * Emits:
 *   submit: (value: string) => void  提交保存
 *   cancel: () => void               取消编辑
 */
import { nextTick, ref } from "vue";

const props = withDefaults(defineProps<{
  modelValue: string;
  trigger?: "click" | "dblclick";
  placeholder?: string;
  maxLength?: number | string;
  validate?: (value: string) => string | null;
  disabled?: boolean;
  align?: "left" | "right" | "center";
}>(), {
  trigger: "click",
  placeholder: "点击编辑",
  maxLength: 200,
  disabled: false,
  align: "left",
});

const emit = defineEmits<{
  (e: "submit", value: string): void;
  (e: "cancel"): void;
}>();

const editing = ref(false);
const draft = ref("");
const inputRef = ref<HTMLInputElement | null>(null);
const error = ref<string | null>(null);

function startEdit(e: MouseEvent) {
  if (props.disabled) return;
  e.stopPropagation();
  draft.value = props.modelValue ?? "";
  error.value = null;
  editing.value = true;
  nextTick(() => {
    inputRef.value?.focus();
    inputRef.value?.select();
  });
}

function commit() {
  const val = draft.value.trim();
  if (props.validate) {
    const err = props.validate(val);
    if (err) {
      error.value = err;
      return;
    }
  }
  editing.value = false;
  error.value = null;
  if (val !== (props.modelValue ?? "")) {
    emit("submit", val);
  } else {
    emit("cancel");
  }
}

function cancel() {
  editing.value = false;
  error.value = null;
  emit("cancel");
}
</script>

<template>
  <span class="inline-edit" :class="`inline-edit--${align}`">
    <template v-if="editing">
      <input
        ref="inputRef"
        v-model="draft"
        class="inline-edit__input"
        :maxlength="Number(maxLength) || undefined"
        @keydown.enter.prevent="commit"
        @keydown.esc.prevent="cancel"
        @blur="commit"
        @click.stop
      />
      <span v-if="error" class="inline-edit__error">{{ error }}</span>
    </template>
    <span
      v-else
      class="inline-edit__text"
      :class="{ 'inline-edit__text--empty': !modelValue, 'inline-edit__text--editable': !disabled }"
      :title="disabled ? '' : placeholder"
      @click="trigger === 'click' && startEdit($event)"
      @dblclick="trigger === 'dblclick' && startEdit($event)"
    >
      <slot>{{ modelValue || placeholder }}</slot>
    </span>
  </span>
</template>

<style scoped>
.inline-edit {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
}

.inline-edit--right { justify-content: flex-end; }
.inline-edit--center { justify-content: center; }

.inline-edit__text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: 4px;
  padding: 1px 4px;
  margin: -1px -4px;
  transition: background 0.1s, box-shadow 0.1s;
}

.inline-edit__text--editable {
  cursor: text;
}

.inline-edit__text--editable:hover {
  background: var(--bg-layer-1-hover);
  box-shadow: 0 0 0 1px var(--border-subtle-1);
}

.inline-edit__text--empty {
  color: var(--txt-placeholder);
  font-style: italic;
}

.inline-edit__input {
  width: 100%;
  min-width: 60px;
  max-width: 100%;
  padding: 1px 6px;
  font-size: inherit;
  font-family: inherit;
  color: var(--txt-primary);
  background: var(--bg-surface-1);
  border: 1px solid var(--border-accent-strong);
  border-radius: 4px;
  outline: none;
  box-shadow: 0 0 0 3px var(--bg-accent-subtle);
}

.inline-edit__error {
  display: block;
  margin-top: 2px;
  font-size: 11px;
  color: var(--txt-danger-primary);
}
</style>
