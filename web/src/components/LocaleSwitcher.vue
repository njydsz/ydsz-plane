<script setup lang="ts">
/**
 * LocaleSwitcher — 语言切换器
 *
 * 展示当前语言并提供下拉切换选项。
 * 切换后立即生效（所有使用 $t() 的文本实时更新）。
 */
import { computed } from "vue";
import { SUPPORTED_LOCALES, setLocale, getLocale, type SupportedLocale } from "../locales";

const currentLocale = computed(() => getLocale());

function switchTo(locale: SupportedLocale) {
  setLocale(locale);
}
</script>

<template>
  <div class="locale-switcher">
    <button
      v-for="loc in SUPPORTED_LOCALES"
      :key="loc.code"
      :class="['locale-btn', { active: currentLocale === loc.code }]"
      :title="loc.name"
      @click="switchTo(loc.code)"
    >
      <span class="locale-flag">{{ loc.flag }}</span>
      <span class="locale-label">{{ loc.code === 'zh-CN' ? '中' : 'EN' }}</span>
    </button>
  </div>
</template>

<style scoped>
.locale-switcher {
  display: flex;
  gap: 4px;
  align-items: center;
}

.locale-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 6px;
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.locale-btn:hover {
  background: var(--color-bg-hover, #f3f4f6);
  color: var(--color-text, #111827);
}

.locale-btn.active {
  border-color: var(--color-primary, #3b82f6);
  background: var(--color-primary-light, #eff6ff);
  color: var(--color-primary, #3b82f6);
}

.locale-flag {
  font-size: 14px;
  line-height: 1;
}

.locale-label {
  font-weight: 500;
}
</style>
