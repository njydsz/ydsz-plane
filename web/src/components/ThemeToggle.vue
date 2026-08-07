<script setup lang="ts">
/**
 * ThemeToggle — 亮/暗主题切换按钮。
 * 点击在 light ↔ dark 间切换（首次点击会取消 system 跟随，改为显式选择）。
 */
import { computed } from "vue";
import { setThemeMode, useTheme } from "@/lib/theme";

const theme = useTheme();

const isDark = computed(() => theme.value === "dark");

function toggle() {
  // 显式选择：基于当前生效主题取反，退出 system 模式
  const next = theme.value === "dark" ? "light" : "dark";
  setThemeMode(next);
}

const label = computed(() => (isDark.value ? "切换到亮色模式" : "切换到暗色模式"));
</script>

<template>
  <button
    class="theme-toggle"
    :title="label"
    :aria-label="label"
    @click="toggle"
  >
    <!-- 太阳 / 月亮 图标（内联 SVG） -->
    <svg v-if="!isDark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
    </svg>
    <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
    </svg>
  </button>
</template>

<style scoped>
.theme-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-surface-2);
  color: var(--txt-secondary);
  cursor: pointer;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}

.theme-toggle:hover {
  background: var(--bg-layer-1-hover);
  color: var(--txt-primary);
}
</style>
