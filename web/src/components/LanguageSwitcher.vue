<script setup lang="ts">
/**
 * 语言切换器组件 — S13-P2 多语言支持。
 *
 * 功能：
 *  - 显示当前语言名称，点击展开下拉列表
 *  - 选择语言调用 vue-i18n setLocale 并持久化到 localStorage
 *  - 支持 dropdown / icon 两种变体
 *  - 移动端触控友好 (≥44x44px)
 *
 * 使用方式：
 *  <LanguageSwitcher variant="dropdown" />  // header 栏 / 登录页
 *  <LanguageSwitcher variant="icon" />      // 紧凑场景
 */
import { computed, onMounted, onUnmounted, ref } from "vue";

import {
  SUPPORTED_LOCALES,
  getLocale,
  setLocale,
  type SupportedLocale,
} from "@/locales";
import DynamicIcon from "@/components/DynamicIcon.vue";

const props = withDefaults(
  defineProps<{
    /** 展示形式：dropdown = 文字+箭头 icon = 仅图标 */
    variant?: "dropdown" | "icon";
  }>(),
  { variant: "dropdown" }
);

interface LanguageOption {
  code: string;
  label: string;
  flag: string;
}

const options: LanguageOption[] = SUPPORTED_LOCALES.map((l) => ({
  code: l.code,
  label: l.name,
  flag: l.flag ?? "",
}));

const currentLocale = ref<SupportedLocale>(getLocale());
const open = ref(false);
const dropdownRef = ref<HTMLElement | null>(null);

const currentLabel = computed(() => {
  const cur = options.find((o) => o.code === currentLocale.value);
  return cur?.label ?? "中文";
});
const currentFlag = computed(() => {
  const cur = options.find((o) => o.code === currentLocale.value);
  return cur?.flag ?? "";
});

function selectLocale(code: string) {
  setLocale(code as SupportedLocale);
  currentLocale.value = code as SupportedLocale;
  open.value = false;
}

function toggle() {
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function onClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    close();
  }
}

onMounted(() => document.addEventListener("mousedown", onClickOutside));
onUnmounted(() => document.removeEventListener("mousedown", onClickOutside));
</script>

<template>
  <div ref="dropdownRef" class="lang-switcher">
    <button
      class="lang-switcher__trigger"
      :class="[`lang-switcher__trigger--${props.variant}`]"
      type="button"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
    >
      <span v-if="props.variant === 'dropdown'" class="lang-flag">{{ currentFlag }}</span>
      <span v-if="props.variant === 'dropdown'" class="lang-label">{{ currentLabel }}</span>
      <span v-if="props.variant === 'icon'" class="lang-icon" title="{{ currentLabel }}">{{ currentFlag }}</span>
      <DynamicIcon
        v-if="props.variant === 'dropdown'"
        class="lang-chevron"
        :class="{ 'lang-chevron--open': open }"
        name="ChevronDown"
        :size="12"
      />
    </button>

    <transition name="lang-fade">
      <ul v-if="open" class="lang-switcher__dropdown" role="listbox">
        <li
          v-for="opt in options"
          :key="opt.code"
          class="lang-option"
          :class="{ 'lang-option--active': opt.code === currentLocale }"
          role="option"
          :aria-selected="opt.code === currentLocale"
          @click="selectLocale(opt.code)"
        >
          <span class="lang-flag">{{ opt.flag }}</span>
          <span class="lang-option__label">{{ opt.label }}</span>
          <DynamicIcon
            v-if="opt.code === currentLocale"
            class="lang-option__check"
            name="Check"
            :size="14"
          />
        </li>
      </ul>
    </transition>
  </div>
</template>

<style scoped>
.lang-switcher {
  position: relative;
  display: inline-flex;
  font-size: 13px;
}

.lang-switcher__trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 10px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm, 6px);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
  min-height: 44px;
  min-width: 44px;
}

.lang-switcher__trigger:hover {
  background: var(--surface-2);
  border-color: var(--border-default);
}

.lang-switcher__trigger--icon {
  padding: 0 6px;
  width: 44px;
  justify-content: center;
}

.lang-flag {
  font-size: 16px;
  line-height: 1;
}

.lang-label {
  font-weight: 500;
  white-space: nowrap;
}

.lang-icon {
  font-size: 18px;
  line-height: 1;
}

.lang-chevron {
  transition: transform 0.2s;
}

.lang-chevron--open {
  transform: rotate(180deg);
}

.lang-switcher__dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 200;
  min-width: 160px;
  padding: 4px;
  margin: 0;
  list-style: none;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1);
  box-shadow: var(--shadow-card, 0 4px 12px rgba(0, 0, 0, 0.08));
}

.lang-option {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.12s;
  color: var(--text-primary);
}

.lang-option:hover {
  background: var(--surface-2);
}

.lang-option--active {
  color: var(--brand-600);
  font-weight: 500;
}

.lang-option__label {
  flex: 1;
}

.lang-option__check {
  color: var(--brand-500);
}

.lang-fade-enter-active,
.lang-fade-leave-active {
  transition: opacity 0.15s, transform 0.15s;
}

.lang-fade-enter-from,
.lang-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
