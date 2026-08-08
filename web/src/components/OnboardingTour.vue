<script setup lang="ts">
/**
 * OnboardingTour — 新用户引导流程（工作空间创建后触发）。
 *
 * 步骤:
 *   1. 欢迎 + 创建第一个项目
 *   2. 在工作项板中创建第一个工作项
 *   3. 邀请团队成员（可选）
 *   4. 完成引导
 *
 * 特性:
 *   - 步骤指示器 + 跳过按钮
 *   - 持久化完成进度（localStorage，跨会话安全）
 *   - 不侵入式：用户随时可跳过，无强制操作
 *   - 首次进入与后续进入均可展示
 */
import { computed, ref } from "vue";

const props = defineProps<{
  workspaceId: string;
  workspaceName?: string;
}>();

const emit = defineEmits<{
  close: [];
  createProject: [];
}>();

/* ---- 步骤 ---- */
interface Step {
  title: string;
  subtitle: string;
  icon: string;
  action?: string;
}

const steps: Step[] = [
  {
    title: `欢迎来到「${props.workspaceName || "新工作空间"}」`,
    subtitle: "让我们花一分钟快速了解核心功能，助您高效启动项目管理。",
    icon: "🎉",
    action: "开始使用",
  },
  {
    title: "第一步：创建您的第一个项目",
    subtitle: "项目是工作项的容器。建议按产品或团队维度创建，如「移动端 App」「后端服务」。",
    icon: "📁",
    action: "创建项目",
  },
  {
    title: "第二步：创建第一个工作项",
    subtitle: "在看板或列表页点击「+ 创建工作项」。您可以选择需求、任务或缺陷类型来启动工作流。",
    icon: "📋",
    action: "了解看板",
  },
  {
    title: "第三步：邀请团队成员（可选）",
    subtitle: "通过工作空间设置 → 成员管理邀请团队协作。支持链接邀请和邮件邀请。",
    icon: "👥",
    action: "前往邀请",
  },
  {
    title: "准备就绪！",
    subtitle: "您已掌握基础操作。如需更深入了解，可查阅帮助文档或随时打开引导再次回顾。",
    icon: "✅",
    action: "完成",
  },
];

/* ---- 状态 ---- */
const currentStep = ref(0);
const skipped = ref(false);

const step = computed(() => steps[currentStep.value]);
const stepCount = computed(() => steps.length);
const isLastStep = computed(() => currentStep.value === steps.length - 1);
const isFirstStep = computed(() => currentStep.value === 0);
const progressPct = computed(() => ((currentStep.value + 1) / steps.length) * 100);

/* ---- 操作 ---- */
function nextStep() {
  if (isLastStep.value) {
    finish();
    return;
  }
  if (currentStep.value === 1) {
    // "创建项目" 按钮触发项目创建
    emit("createProject");
    return;
  }
  currentStep.value++;
}

function prevStep() {
  if (currentStep.value > 0) currentStep.value--;
}

function skip() {
  skipped.value = true;
  finish();
}

function finish() {
  // 持久化标记（按工作空间隔离）
  try {
    localStorage.setItem(`onboarding-done:${props.workspaceId}`, "1");
  } catch {
    // localStorage 不可用则静默
  }
  emit("close");
}

/* ---- 持久化检查（导出供父组件判断是否需要展示） ---- */
function hasCompleted(): boolean {
  try {
    return localStorage.getItem(`onboarding-done:${props.workspaceId}`) === "1";
  } catch {
    return false;
  }
}

defineExpose({ hasCompleted });
</script>

<template>
  <div class="onboarding-tour">
    <div class="onboarding-tour__backdrop" @click="skip" />
    <div class="onboarding-tour__card">
      <!-- 关闭按钮 -->
      <button class="onboarding-tour__close" title="跳过引导" @click="skip">✕</button>

      <!-- 进度指示 -->
      <div class="onboarding-tour__progress">
        <div class="onboarding-tour__progress-bar">
          <div class="onboarding-tour__progress-fill" :style="{ width: progressPct + '%' }" />
        </div>
        <span class="onboarding-tour__progress-label">{{ currentStep + 1 }} / {{ stepCount }}</span>
      </div>

      <!-- 步骤内容 -->
      <div class="onboarding-tour__body">
        <div class="onboarding-tour__icon">{{ step.icon }}</div>
        <h2 class="onboarding-tour__title">{{ step.title }}</h2>
        <p class="onboarding-tour__subtitle">{{ step.subtitle }}</p>
      </div>

      <!-- 操作区 -->
      <div class="onboarding-tour__actions">
        <button
          v-if="!isFirstStep"
          class="btn btn--ghost"
          @click="prevStep"
        >
          上一步
        </button>
        <button
          class="btn btn--primary"
          @click="nextStep"
        >
          {{ step.action || "下一步" }}
        </button>
        <button
          v-if="!isLastStep"
          class="btn btn--text"
          @click="skip"
        >
          跳过引导
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.onboarding-tour {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}
.onboarding-tour__backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(2px);
}
.onboarding-tour__card {
  position: relative;
  background: var(--bg-surface-1);
  border-radius: var(--radius-lg);
  padding: 32px 28px 24px;
  max-width: 480px;
  width: calc(100% - 40px);
  box-shadow: var(--shadow-overlay-200);
  z-index: 1;
}
.onboarding-tour__close {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--txt-tertiary);
  font-size: 14px;
  cursor: pointer;
  border-radius: 50%;
}
.onboarding-tour__close:hover {
  background: var(--bg-surface-3);
}
.onboarding-tour__progress {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}
.onboarding-tour__progress-bar {
  flex: 1;
  height: 4px;
  background: var(--bg-surface-3);
  border-radius: 2px;
  overflow: hidden;
}
.onboarding-tour__progress-fill {
  height: 100%;
  background: var(--brand-500);
  transition: width 0.3s ease;
}
.onboarding-tour__progress-label {
  font-size: 11px;
  color: var(--txt-tertiary);
  white-space: nowrap;
}
.onboarding-tour__body {
  text-align: center;
  margin-bottom: 24px;
}
.onboarding-tour__icon {
  font-size: 36px;
  margin-bottom: 12px;
}
.onboarding-tour__title {
  font-size: 17px;
  font-weight: 600;
  margin: 0 0 8px;
  color: var(--txt-primary);
}
.onboarding-tour__subtitle {
  font-size: 13px;
  color: var(--txt-secondary);
  line-height: 1.6;
  margin: 0;
}
.onboarding-tour__actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
}
.btn--ghost {
  padding: 6px 14px;
  border: 1px solid var(--border-default);
  background: transparent;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--txt-primary);
  font-family: inherit;
}
.btn--primary {
  padding: 7px 20px;
  background: var(--brand-500);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  font-weight: 500;
}
.btn--primary:hover {
  background: var(--brand-600);
}
.btn--text {
  padding: 6px 10px;
  background: none;
  border: none;
  color: var(--txt-tertiary);
  cursor: pointer;
  font-size: 12px;
  font-family: inherit;
}
.btn--text:hover {
  color: var(--txt-primary);
  text-decoration: underline;
}
</style>
