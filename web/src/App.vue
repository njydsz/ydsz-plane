/** 应用根组件：挂载路由视图，并全局注入命令面板与消息提示。 */
<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";

import CommandPalette from "@/components/CommandPalette.vue";
import AppToast from "@/components/AppToast.vue";
import OnboardingTour from "@/components/OnboardingTour.vue";
import { useWorkspaceStore } from "@/stores/workspace";
import { applyBrandColor, clearBrandColor } from "@/composables/useBrandColor";

const route = useRoute();
const wsStore = useWorkspaceStore();
const onboardingWsId = ref<string | null>(null);
const onboardingWsName = ref<string>("");

/** 监听当前workspace变化，应用品牌色 */
watch(
  () => wsStore.current?.brand_color,
  (color) => {
    if (color) {
      applyBrandColor(color);
    } else {
      clearBrandColor();
    }
  },
  { immediate: true },
);

/** 检测是否需要显示新手引导（sessionStorage 中存在 pending-onboarding:{wsId} 且未完成过）。 */
function checkOnboarding() {
  onboardingWsId.value = null;
  onboardingWsName.value = "";

  // 从当前路由提取 workspace ID
  const wsId = route.params?.workspaceId;
  if (!wsId || Array.isArray(wsId)) return;

  try {
    // 已完成引导的不再展示
    if (localStorage.getItem(`onboarding-done:${wsId}`) === "1") return;
    // 仅当 sessionStorage 中存在 pending 标记时展示
    const flag = sessionStorage.getItem(`pending-onboarding:${wsId}`);
    if (flag === "1") {
      onboardingWsId.value = wsId;
    }
  } catch {
    // ignore
  }
}

function closeOnboarding() {
  // 标记已完成
  if (onboardingWsId.value) {
    try {
      sessionStorage.removeItem(`pending-onboarding:${onboardingWsId.value}`);
    } catch {
      // ignore
    }
  }
  onboardingWsId.value = null;
}

function onOnboardingCreateProject() {
  closeOnboarding();
  // 在引导「创建项目」步骤后关闭，用户可自主选择是否创建项目
  // 无需强制跳转，保持当前位置
}

// 每次路由切换后检查是否需要展示引导
watch(() => route.params.workspaceId, () => {
  checkOnboarding();
}, { immediate: true });

// 全局挂载：供 WorkspaceListView 跨路由也能触发
onMounted(() => {
  checkOnboarding();
});
</script>

<template>
  <router-view />
  <CommandPalette />
  <AppToast />
  <OnboardingTour
    v-if="onboardingWsId"
    :workspace-id="onboardingWsId"
    :workspace-name="onboardingWsName"
    @close="closeOnboarding"
    @create-project="onOnboardingCreateProject"
  />
</template>
