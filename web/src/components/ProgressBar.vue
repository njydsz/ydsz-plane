<script setup lang="ts">
/**
 * ProgressBar — 通用进度条组件，支持多尺寸和颜色变体。
 *
 * Props:
 *   percent:  完成百分比 0-100
 *   size:     'sm' | 'md' | 'lg'
 *   color:    自定义填充色（默认 --success-500）
 *   showLabel: 是否显示百分比文字
 *   animated:  是否显示动画
 *   label:     自定义标签文字（覆盖百分比）
 *   striped:   条纹效果
 */
withDefaults(defineProps<{
  percent: number;
  size?: "sm" | "md" | "lg";
  color?: string;
  showLabel?: boolean;
  animated?: boolean;
  label?: string;
  striped?: boolean;
}>(), {
  size: "md",
  showLabel: true,
  animated: true,
});
</script>

<template>
  <div class="progress" :class="[`progress--${size}`]">
    <div v-if="showLabel" class="progress__header">
      <span class="progress__label">{{ label ?? `${Math.round(percent)}%` }}</span>
      <slot name="header" />
    </div>
    <div class="progress__track">
      <div
        class="progress__bar"
        :class="{ 'progress__bar--animated': animated, 'progress__bar--striped': striped }"
        :style="{
          width: `${Math.min(Math.max(percent, 0), 100)}%`,
          background: color ?? 'var(--success-500)',
        }"
      ></div>
    </div>
    <div v-if="!showLabel" class="progress__inline-label">
      {{ label ?? `${Math.round(percent)}%` }}
    </div>
  </div>
</template>

<style scoped>
.progress {
  width: 100%;
}

.progress__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.progress__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.progress__inline-label {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono);
  margin-top: 2px;
  text-align: right;
}

.progress__track {
  width: 100%;
  background: var(--surface-3);
  border-radius: 999px;
  overflow: hidden;
}

.progress--sm .progress__track { height: 4px; border-radius: 2px; }
.progress--md .progress__track { height: 8px; border-radius: 4px; }
.progress--lg .progress__track { height: 16px; border-radius: 8px; }

.progress__bar {
  height: 100%;
  border-radius: inherit;
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  min-width: 0;
}

.progress__bar--animated {
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.progress__bar--striped {
  background-image: linear-gradient(
    45deg,
    rgba(255, 255, 255, 0.15) 25%,
    transparent 25%,
    transparent 50%,
    rgba(255, 255, 255, 0.15) 50%,
    rgba(255, 255, 255, 0.15) 75%,
    transparent 75%,
    transparent
  );
  background-size: 20px 20px;
  animation: progress-stripe 0.6s linear infinite;
}

@keyframes progress-stripe {
  from { background-position: 20px 0; }
  to { background-position: 0 0; }
}
</style>
