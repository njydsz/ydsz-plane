<script setup lang="ts">
/**
 * PwaPrompts — PWA 提示集合。
 *
 * 包含：
 *  - 安装应用提示条（可关闭，7 天内不再提示）；
 *  - 离线状态横幅；
 *  - SW 更新就绪提示。
 */
import { ref } from "vue";

import { usePwa } from "@/composables/usePwa";

const { isOnline, canInstall, swUpdated, promptInstall, applyUpdate } = usePwa();

const dismissedInstall = ref(false);

async function install() {
  const outcome = await promptInstall();
  if (outcome === "accepted") {
    dismissedInstall.value = true;
  }
}
</script>

<template>
  <div class="pwa">
    <!-- 安装提示 -->
    <div v-if="canInstall && !dismissedInstall" class="pwa__banner pwa__banner--install">
      <span class="pwa__text">安装应用到桌面，获得更接近原生的体验</span>
      <div class="pwa__actions">
        <button class="pwa__btn pwa__btn--primary" @click="install">安装</button>
        <button class="pwa__btn pwa__btn--ghost" @click="dismissedInstall = true">
          暂不
        </button>
      </div>
    </div>

    <!-- 离线提示 -->
    <div v-if="!isOnline" class="pwa__banner pwa__banner--offline">
      <span class="pwa__text">当前处于离线状态，部分功能可能不可用</span>
    </div>

    <!-- 更新就绪 -->
    <div v-if="swUpdated" class="pwa__banner pwa__banner--update">
      <span class="pwa__text">新版本已就绪，刷新即可升级</span>
      <div class="pwa__actions">
        <button class="pwa__btn pwa__btn--primary" @click="applyUpdate">刷新</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pwa {
  position: fixed;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: min(92vw, 480px);
}

.pwa__banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  font-size: 13px;
  animation: pwa-slide-up 0.25s ease;
}

@keyframes pwa-slide-up {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.pwa__banner--install {
  background: var(--accent);
  color: var(--text-on-color);
}

.pwa__banner--offline {
  background: var(--warning);
  color: var(--text-on-color);
}

.pwa__banner--update {
  background: var(--success);
  color: var(--text-on-color);
}

.pwa__text {
  flex: 1;
}

.pwa__actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.pwa__btn {
  padding: 5px 12px;
  font-size: 12px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
}

.pwa__btn--primary {
  background: var(--text-on-color);
  color: var(--accent);
  font-weight: 500;
}

.pwa__banner--offline .pwa__btn--primary,
.pwa__banner--update .pwa__btn--primary {
  color: inherit;
}

.pwa__btn--ghost {
  background: rgba(255, 255, 255, 0.2);
  color: var(--text-on-color);
}
</style>
