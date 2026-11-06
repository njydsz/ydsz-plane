/**
 * usePwa — PWA 能力封装。
 *
 * 提供：
 *  - 安装提示（beforeinstallprompt 事件捕获与触发）；
 *  - 在线/离线状态检测；
 *  - SW 更新就绪提示。
 */
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

/** 可被安装的提示事件（浏览器相关类型未在 TS DOM lib 中定义） */
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

const deferredPrompt = ref<BeforeInstallPromptEvent | null>(null);
const isInstalled = ref(false);
const isOnline = ref(typeof navigator !== "undefined" ? navigator.onLine : true);
const swUpdated = ref(false);
const swRegistration = ref<ServiceWorkerRegistration | null>(null);

let listenersBound = false;

function bindGlobalListeners() {
  if (listenersBound || typeof window === "undefined") return;
  listenersBound = true;

  window.addEventListener("beforeinstallprompt", (e) => {
    e.preventDefault();
    deferredPrompt.value = e as BeforeInstallPromptEvent;
  });

  window.addEventListener("appinstalled", () => {
    isInstalled.value = true;
    deferredPrompt.value = null;
  });

  window.addEventListener("online", () => {
    isOnline.value = true;
  });
  window.addEventListener("offline", () => {
    isOnline.value = false;
  });

  // 检测是否已以 standalone 模式运行
  if (window.matchMedia && window.matchMedia("(display-mode: standalone)").matches) {
    isInstalled.value = true;
  }
}

export function usePwa() {
  onMounted(() => {
    bindGlobalListeners();

    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.ready.then((reg) => {
        swRegistration.value = reg;
        reg.addEventListener("updatefound", () => {
          const installing = reg.installing;
          if (!installing) return;
          installing.addEventListener("statechange", () => {
            if (installing.state === "installed" && navigator.serviceWorker.controller) {
              swUpdated.value = true;
            }
          });
        });
      });
    }
  });

  onBeforeUnmount(() => {
    // 全局监听保留（PWA 生命周期与应用一致）
  });

  /** 是否可以弹出安装提示 */
  const canInstall = computed(() => deferredPrompt.value !== null && !isInstalled.value);

  /** 触发安装 */
  async function promptInstall(): Promise<"accepted" | "dismissed" | "unavailable"> {
    if (!deferredPrompt.value) return "unavailable";
    await deferredPrompt.value.prompt();
    const choice = await deferredPrompt.value.userChoice;
    deferredPrompt.value = null;
    return choice.outcome;
  }

  /** 应用待处理的 SW 更新 */
  function applyUpdate() {
    const reg = swRegistration.value;
    if (reg?.waiting) {
      reg.waiting.postMessage({ type: "SKIP_WAITING" });
      window.location.reload();
    }
  }

  return {
    isInstalled,
    isOnline,
    canInstall,
    swUpdated,
    promptInstall,
    applyUpdate,
  };
}
