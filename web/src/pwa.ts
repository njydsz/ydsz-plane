/**
 * PWA — Service Worker 注册与更新管理
 *
 * 功能:
 *   - 注册 Service Worker（vite-plugin-pwa 生成）
 *   - 自动检测更新并提示用户刷新
 *   - Web Push 通知订阅（预留接口）
 *   - 离线状态检测与 UI 提示
 *
 * 参考:
 *   - Workbox (Google)
 *   - vite-plugin-pwa auto update mode
 */

/** 注册 Service Worker 并监听更新 */
export function registerSW(): void {
  // 开发模式下禁用 PWA，避免 Service Worker 与 HMR 冲突导致页面闪烁/循环刷新
  if (import.meta.env.DEV) {
    // 开发模式下确保清理已有的 Service Worker，避免缓存干扰
    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.getRegistrations().then((regs) => {
        for (const reg of regs) {
          reg.unregister().catch(() => {});
        }
      });
    }
    return;
  }

  if (!("serviceWorker" in navigator)) {
    return;
  }

  // vite-plugin-pwa 会在构建时注入 sw.js
  const swUrl = "/sw.js";

  navigator.serviceWorker
    .register(swUrl, { scope: "/" })
    .then((registration) => {
      console.debug("[PWA] Service Worker registered:", registration.scope);

      // 监听更新
      registration.addEventListener("updatefound", () => {
        const installingWorker = registration.installing;
        if (!installingWorker) return;

        installingWorker.addEventListener("statechange", () => {
          if (
            installingWorker.state === "installed" &&
            navigator.serviceWorker.controller
          ) {
            // 新版本已就绪
            showUpdatePrompt();
          }
        });
      });

      // 定期检查更新（每小时）
      setInterval(() => {
        registration.update().catch(() => {});
      }, 60 * 60 * 1000);
    })
    .catch((err) => {
      console.warn("[PWA] Service Worker registration failed:", err);
    });

  // 监听 controller 变更（新 SW 接管后自动刷新）
  let refreshing = false;
  navigator.serviceWorker.addEventListener("controllerchange", () => {
    if (refreshing) return;
    refreshing = true;
    window.location.reload();
  });

  // 离线状态监听
  window.addEventListener("online", () => {
    document.body.classList.remove("app-offline");
  });
  window.addEventListener("offline", () => {
    document.body.classList.add("app-offline");
  });
}

/** 显示更新提示 */
function showUpdatePrompt(): void {
  // 简单的 toast 提示（避免依赖组件库）
  const banner = document.createElement("div");
  banner.className = "pwa-update-banner";
  banner.innerHTML = `
    <span>🔄 新版本已就绪</span>
    <button id="pwa-update-btn">立即刷新</button>
  `;

  // 样式
  const style = document.createElement("style");
  style.textContent = `
    .pwa-update-banner {
      position: fixed;
      bottom: 20px;
      left: 50%;
      transform: translateX(-50%);
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 12px 24px;
      background: #1f2937;
      color: #fff;
      border-radius: 12px;
      box-shadow: 0 8px 24px rgba(0,0,0,0.2);
      z-index: 99999;
      font-size: 14px;
      animation: pwa-slide-up 0.3s ease;
    }
    .pwa-update-banner button {
      padding: 6px 16px;
      border: none;
      border-radius: 6px;
      background: #3b82f6;
      color: #fff;
      font-size: 13px;
      cursor: pointer;
      transition: background 0.2s;
    }
    .pwa-update-banner button:hover {
      background: #2563eb;
    }
    @keyframes pwa-slide-up {
      from { transform: translateX(-50%) translateY(20px); opacity: 0; }
      to { transform: translateX(-50%) translateY(0); opacity: 1; }
    }
    .app-offline::before {
      content: "离线模式 — 部分功能不可用";
      display: block;
      padding: 6px;
      text-align: center;
      background: #f59e0b;
      color: #fff;
      font-size: 13px;
      position: sticky;
      top: 0;
      z-index: 9999;
    }
  `;
  document.head.appendChild(style);
  document.body.appendChild(banner);

  document.getElementById("pwa-update-btn")?.addEventListener("click", () => {
    banner.remove();
    // 通知 SW 跳过等待并接管
    navigator.serviceWorker.ready.then((reg) => {
      reg.waiting?.postMessage({ type: "SKIP_WAITING" });
    });
  });

  // 10 秒后自动消失
  setTimeout(() => banner.remove(), 10000);
}

/** 请求 Web Push 通知权限（预留接口） */
export async function requestPushPermission(): Promise<boolean> {
  if (!("Notification" in window)) {
    return false;
  }

  const permission = await Notification.requestPermission();
  return permission === "granted";
}

/** 订阅 Web Push（需要后端 VAPID 公钥支持） */
export async function subscribePush(
  vapidPublicKey: string
): Promise<PushSubscription | null> {
  if (!("serviceWorker" in navigator)) {
    return null;
  }

  try {
    const registration = await navigator.serviceWorker.ready;
    const subscription = await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(vapidPublicKey),
    });

    // 将订阅信息发送到后端
    // await pushService.register(subscription);

    return subscription;
  } catch (err) {
    console.warn("[PWA] Push subscription failed:", err);
    return null;
  }
}

/** 取消 Web Push 订阅 */
export async function unsubscribePush(): Promise<void> {
  const registration = await navigator.serviceWorker.ready;
  const subscription = await registration.pushManager.getSubscription();
  if (subscription) {
    await subscription.unsubscribe();
  }
}

/**
 * 将 Base64 URL-safe 字符串转换为 Uint8Array
 * （Web Push applicationServerKey 需要此格式）
 */
function urlBase64ToUint8Array(base64String: string): Uint8Array {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}
