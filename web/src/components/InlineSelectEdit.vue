<script setup lang="ts">
/**
 * InlineSelectEdit — 内联选择编辑组件（对标 Plane 的 inline select）。
 *
 * 交互：单击当前值弹出浮层选项列表，点击选项即保存；Esc/点击外部关闭。
 *
 * Props:
 *   modelValue: 当前值（任意基础类型）
 *   options:    { value, label, color?, icon? }[]
 *   placeholder: 空值占位
 *   disabled:    禁用编辑
 *   align:       浮层对齐方向 'left' | 'right'
 *
 * Emits:
 *   submit: (value) => void  选择新值（值与旧值不同才触发）
 */
import { computed, onMounted, onUnmounted, ref } from "vue";

/** 内联选择编辑器的选项定义（value + label + 可选颜色/图标/title）。 */
export interface SelectOption<T = unknown> {
  value: T;
  label: string;
  color?: string;
  icon?: string;
  title?: string;
}

const props = withDefaults(defineProps<{
  modelValue: unknown;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
  align?: "left" | "right";
  emptyValue?: unknown;
}>(), {
  placeholder: "未设置",
  disabled: false,
  align: "left",
  emptyValue: undefined,
});

const emit = defineEmits<{
  (e: "submit", value: unknown): void;
}>();

const open = ref(false);
const rootRef = ref<HTMLElement | null>(null);

const current = computed(() =>
  props.options.find((o) => o.value === props.modelValue) ?? null,
);

function toggle() {
  if (props.disabled) return;
  open.value = !open.value;
}

function pick(opt: SelectOption) {
  open.value = false;
  if (opt.value !== props.modelValue) {
    emit("submit", opt.value);
  }
}

function onDocClick(e: MouseEvent) {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) {
    open.value = false;
  }
}

onMounted(() => document.addEventListener("mousedown", onDocClick));
onUnmounted(() => document.removeEventListener("mousedown", onDocClick));
</script>

<template>
  <span ref="rootRef" class="inline-select" :class="{ 'inline-select--open': open }">
      <span
        class="inline-select__trigger"
        :class="{ 'inline-select__trigger--editable': !disabled, 'inline-select__trigger--empty': !current }"
        role="button"
        tabindex="0"
        @click="toggle"
        @keydown.enter="toggle"
        @keydown.esc="open = false"
      >
        <slot name="trigger">
          <span v-if="current?.color" class="inline-select__dot" :style="{ background: current.color }" />
          <span v-if="current?.icon" class="inline-select__icon">{{ current.icon }}</span>
          <span class="inline-select__label">{{ current?.label ?? placeholder }}</span>
        </slot>
      </span>

    <Transition name="ise-pop">
      <div
        v-if="open"
        class="inline-select__pop"
        :class="`inline-select__pop--${align}`"
      >
        <button
          v-for="opt in options"
          :key="String(opt.value)"
          class="inline-select__opt"
          :class="{ 'inline-select__opt--active': opt.value === modelValue }"
          :title="opt.title"
          @click="pick(opt)"
        >
          <span v-if="opt.color" class="inline-select__dot" :style="{ background: opt.color }" />
          <span v-if="opt.icon" class="inline-select__icon">{{ opt.icon }}</span>
          <span class="inline-select__label">{{ opt.label }}</span>
        </button>
      </div>
    </Transition>
  </span>
</template>

<style scoped>
.inline-select {
  position: relative;
  display: inline-flex;
  max-width: 100%;
}

.inline-select__trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 1px 6px;
  margin: -1px -6px;
  border-radius: 4px;
  font-size: inherit;
  line-height: 1.5;
  color: inherit;
  cursor: default;
  max-width: 100%;
  transition: background 0.1s;
}

.inline-select__trigger--editable {
  cursor: pointer;
}

.inline-select__trigger--editable:hover {
  background: var(--bg-layer-1-hover);
}

.inline-select__trigger--empty {
  color: var(--txt-placeholder);
}

.inline-select__trigger--open {
  background: var(--bg-layer-1-hover);
}

.inline-select__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.inline-select__icon {
  font-size: 11px;
  flex-shrink: 0;
}

.inline-select__label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inline-select__pop {
  position: absolute;
  top: calc(100% + 4px);
  z-index: 300;
  min-width: 140px;
  max-width: 240px;
  max-height: 260px;
  overflow-y: auto;
  padding: 4px;
  background: var(--bg-surface-1);
  border: 1px solid var(--border-subtle-1);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
}

.inline-select__pop--left { left: 0; }
.inline-select__pop--right { right: 0; }

.inline-select__opt {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 5px 8px;
  border: none;
  border-radius: var(--radius-sm);
  background: none;
  font-size: 12px;
  font-family: inherit;
  color: var(--txt-primary);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
}

.inline-select__opt:hover {
  background: var(--bg-layer-1-hover);
}

.inline-select__opt--active {
  background: var(--bg-accent-subtle);
  color: var(--txt-accent-primary);
  font-weight: 500;
}

/* Transition */
.ise-pop-enter-active,
.ise-pop-leave-active {
  transition: opacity 0.12s ease, transform 0.12s ease;
}
.ise-pop-enter-from,
.ise-pop-leave-to {
  opacity: 0;
  transform: translateY(-3px);
}
</style>
