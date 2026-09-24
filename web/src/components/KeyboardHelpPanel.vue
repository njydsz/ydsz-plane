<!-- 快捷键帮助弹窗（"?" 触发，Esc 关闭）。 -->
<script setup lang="ts">
defineProps<{ shortcuts: { key: string; label: string; description: string }[] }>()
const emit = defineEmits<{ (e: "close"): void }>()
</script>

<template>
  <Teleport to="body">
    <div class="kh-overlay" role="dialog" aria-modal="true" aria-label="键盘快捷键帮助" @click.self="emit('close')">
      <div class="kh-card">
        <div class="kh-head">
          <h2 class="kh-title">键盘快捷键</h2>
          <button class="kh-close" aria-label="关闭" @click="emit('close')">✕</button>
        </div>
        <ul class="kh-list">
          <li v-for="s in shortcuts" :key="s.key" class="kh-row">
            <kbd class="kh-key">{{ s.label }}</kbd>
            <span class="kh-desc">{{ s.description }}</span>
          </li>
        </ul>
        <p class="kh-tip">按 <kbd class="kh-key">?</kbd> 随时打开此面板</p>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.kh-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
}
.kh-card {
  width: min(560px, 92vw);
  max-height: 80vh;
  overflow-y: auto;
  background: var(--bg-elevated, #fff);
  border-radius: 12px;
  padding: 20px 22px;
  box-shadow: 0 18px 60px rgba(0, 0, 0, 0.28);
}
.kh-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.kh-title {
  font-size: 16px;
  font-weight: 600;
}
.kh-close {
  border: none;
  background: transparent;
  font-size: 18px;
  cursor: pointer;
  color: var(--text-secondary);
}
.kh-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.kh-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-subtle, #eee);
}
.kh-key {
  min-width: 32px;
  text-align: center;
  padding: 3px 8px;
  border-radius: 5px;
  border: 1px solid var(--border-subtle, #ddd);
  background: var(--bg-secondary, #f5f5f5);
  font-family: ui-monospace, monospace;
  font-size: 12px;
}
.kh-desc {
  font-size: 13px;
  color: var(--text-primary);
}
.kh-tip {
  margin-top: 14px;
  font-size: 12px;
  color: var(--text-tertiary);
}
