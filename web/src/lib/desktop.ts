/**
 * desktop.ts — 桌面壳（Wails v3）桥接层。
 *
 * 职责：
 *   - 检测是否运行在 Wails 桌面壳内；
 *   - 封装平台服务调用（通知 / 窗口 / 文件系统 / 更新）；
 *   - 提供与浏览器环境一致的回退实现。
 *
 * 生成的绑定位于 frontend/bindings/，由 wails3 dev/build 自动生成。
 * 此处通过动态导入避免在浏览器环境下报错。
 */
import { ref } from "vue"

/** 是否运行在桌面壳内 */
export const isDesktop = ref(false)
/** 平台信息 */
export const platformInfo = ref<{
  platform: string
  arch: string
  version: string
  hostname: string
  homeDir: string
  appDataDir: string
} | null>(null)

/** 服务绑定缓存 */
let bindings: any = null

/**
 * 初始化桌面桥接。
 *
 * 通过检查 window.go 或 Wails 注入的全局对象判断环境。
 */
export async function initDesktop(): Promise<boolean> {
  try {
    // Wails v3 注入 window.runtime 或绑定模块
    const w = window as any
    if (w.runtime?.Platform || w.go) {
      isDesktop.value = true
    }

    // 尝试加载生成的绑定（位于 frontend/bindings/）
    if (isDesktop.value) {
      try {
        // 绑定由 wails3 dev/build 生成于 frontend/bindings/
        // 使用动态路径 + 宽松类型，避免浏览器环境下类型检查失败
        const bindingPath = "../bindings/github.com/njydsz/ydsz-plane/desktop/wails"
        const mod = await import(/* @vite-ignore */ bindingPath)
        bindings = mod as unknown as Record<string, unknown>
        await loadPlatformInfo()
      } catch {
        bindings = null
      }
    }
  } catch {
    isDesktop.value = false
  }
  return isDesktop.value
}

async function loadPlatformInfo() {
  if (!bindings) return
  try {
    const svc = bindings.PlatformService
    if (!svc) return
    platformInfo.value = {
      platform: await svc.GetPlatform(),
      arch: await svc.GetArch(),
      version: await svc.GetVersion(),
      hostname: await svc.GetHostname(),
      homeDir: await svc.GetHomeDir(),
      appDataDir: await svc.GetAppDataDir(),
    }
  } catch {
    /* 忽略 */
  }
}

// ----------------------------------------------------------------------
// 平台服务封装
// ----------------------------------------------------------------------

/** 发送原生通知 */
export async function notify(title: string, body: string, icon = ""): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try {
    await bindings.NotificationService.Notify(title, body, icon)
  } catch {
    /* 降级到浏览器通知 */
    if (typeof Notification !== "undefined" && Notification.permission === "granted") {
      new Notification(title, { body })
    }
  }
}

/** 设置窗口标题 */
export async function setWindowTitle(title: string): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.SetTitle(title) } catch { /* */ }
}

/** 最小化窗口 */
export async function minimizeWindow(): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.Minimize() } catch { /* */ }
}

/** 最大化/还原 */
export async function maximizeWindow(): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.Maximize() } catch { /* */ }
}

/** 关闭窗口（隐藏到托盘） */
export async function closeWindow(): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.Close() } catch { /* */ }
}

/** 退出应用 */
export async function quitApp(): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.Quit() } catch { /* */ }
}

/** 置顶窗口 */
export async function setAlwaysOnTop(on: boolean): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.WindowService.SetAlwaysOnTop(on) } catch { /* */ }
}

/** 用系统浏览器打开链接 */
export async function openExternal(url: string): Promise<void> {
  if (isDesktop.value && bindings) {
    try { await bindings.WindowService.OpenURL(url); return } catch { /* */ }
  }
  window.open(url, "_blank", "noopener,noreferrer")
}

/** 检查更新 */
export async function checkForUpdate(): Promise<string> {
  if (!isDesktop.value || !bindings) return ""
  try { return await bindings.UpdaterService.CheckForUpdate() } catch { return "" }
}

/** 是否启用自动更新 */
export async function isAutoUpdateEnabled(): Promise<boolean> {
  if (!isDesktop.value || !bindings) return false
  try { return await bindings.UpdaterService.IsAutoUpdateEnabled() } catch { return false }
}

/** 获取应用数据目录（用于导出文件） */
export function getAppDataDir(): string {
  return platformInfo.value?.appDataDir ?? ""
}

/** 转发日志到桌面端 */
export async function logToDesktop(level: string, message: string): Promise<void> {
  if (!isDesktop.value || !bindings) return
  try { await bindings.PlatformService.Log(level, message) } catch { /* */ }
}
