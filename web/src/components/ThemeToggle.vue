<script setup lang="ts">
/**
 * ThemeToggle — 亮/暗主题切换按钮。
 * 点击在 light ↔ dark 间切换（首次点击会取消 system 跟随，改为显式选择）。
 */
import { computed } from "vue";
import { applyTheme, currentTheme } from "@/lib/theme";
import DynamicIcon from "@/components/DynamicIcon.vue";

const theme = currentTheme;

const isDark = computed(() => theme.value === "dark");

function toggle() {
  // 显式选择：基于当前生效主题取反，退出 system 模式
  const next = theme.value === "dark" ? "light" : "dark";
  applyTheme(next);
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
    <!-- 太阳 / 月亮 图标 -->
    <DynamicIcon :name="isDark ? 'Moon' : 'Sun'" :size="16" />
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
