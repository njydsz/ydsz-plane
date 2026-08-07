<script setup lang="ts">
/**
 * EmptyStateWidget - 无 widget 时的空态占位。
 * 显示引导文本和"添加 Widget"按钮。
 */
const emit = defineEmits<{
  add: [];
}>();
</script>

<template>
  <div class="empty-state">
    <div class="empty-state__icon">📊</div>
    <h3 class="empty-state__title">暂无仪表盘</h3>
    <p class="empty-state__hint">添加 widget 来追踪项目进度、迭代燃尽与风险告警</p>
    <button class="empty-state__btn" type="button" @click="emit('add')">
      添加第一个 Widget
    </button>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  text-align: center;
}

.empty-state__icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.empty-state__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0 0 6px;
}

.empty-state__hint {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
  margin: 0 0 20px;
}

.empty-state__btn {
  padding: 8px 18px;
  border-radius: var(--radius-sm, 6px);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  background: var(--brand-500, #3f63f1);
  color: var(--text-on-brand, #fff);
  border: 1px solid var(--brand-500, #3f63f1);
  transition: background 0.15s;
}

.empty-state__btn:hover {
  background: var(--brand-600, #2f4fd0);
}
</style>
