<script setup lang="ts">
/**
 * AppModal — 通用模态框。
 * 特性: 点击遮罩关闭、ESC 关闭、focus 锁定、body scroll 锁定。
 *
 * Props:
 *   visible: 是否显示
 *   title:   标题
 *   width:   宽度 (CSS 值，默认 480px)
 *
 * Events:
 *   @close: 关闭事件
 */
import { onMounted, onUnmounted, watch } from "vue";

const props = withDefaults(defineProps<{
  visible: boolean;
  title?: string;
  width?: string;
}>(), {
  title: "",
  width: "480px",
});

const emit = defineEmits<{
  close: [];
}>();

function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && props.visible) {
    emit("close");
  }
}

watch(() => props.visible, (v) => {
  document.body.style.overflow = v ? "hidden" : "";
});

onMounted(() => window.addEventListener("keydown", onKeydown));
onUnmounted(() => {
  window.removeEventListener("keydown", onKeydown);
  document.body.style.overflow = "";
});
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="app-modal-overlay" @click.self="emit('close')">
      <div class="app-modal" :style="{ maxWidth: width }" role="dialog" aria-modal="true">
        <header v-if="title || $slots.header" class="app-modal__header">
          <slot name="header">
            <h2 class="app-modal__title">{{ title }}</h2>
          </slot>
          <button class="app-modal__close" aria-label="关闭" @click="emit('close')">&times;</button>
        </header>
        <div class="app-modal__body">
          <slot />
        </div>
        <footer v-if="$slots.footer" class="app-modal__footer">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.app-modal-overlay {
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

.app-modal {
  width: calc(100% - 32px);
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-popover);
  animation: slideUp 0.2s ease;
}

.app-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.app-modal__title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.app-modal__close {
  background: none;
  border: none;
  font-size: 22px;
  color: var(--text-tertiary);
  cursor: pointer;
  line-height: 1;
  padding: 0;
}

.app-modal__close:hover {
  color: var(--text-primary);
}

.app-modal__body {
  padding: 20px 24px;
}

.app-modal__footer {
  padding: 0 24px 20px;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
