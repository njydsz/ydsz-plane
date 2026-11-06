/**
 * Wails v3 生成的绑定声明。
 *
 * 实际绑定由 `wails3 dev` / `wails3 build` 在 frontend/bindings/ 下生成。
 * 此文件仅提供类型占位，避免浏览器环境下的类型错误。
 */

declare module "../bindings/github.com/njydsz/ydsz-plane/desktop/wails" {
  export interface NotificationAction {
    Label: string
    OnClick: () => void
  }

  export interface Notification {
    Title: string
    Message: string
    Actions?: NotificationAction[]
  }

  /** 平台服务 */
  export const PlatformService: {
    GetPlatform(): Promise<string>
    GetArch(): Promise<string>
    GetVersion(): Promise<string>
    IsDesktop(): Promise<boolean>
    GetHostname(): Promise<string>
    GetHomeDir(): Promise<string>
    GetAppDataDir(): Promise<string>
    Log(level: string, message: string): Promise<void>
  }

  /** 通知服务 */
  export const NotificationService: {
    Notify(title: string, body: string, icon?: string): Promise<void>
    NotifyWithActions(title: string, body: string, actions: string[]): Promise<void>
    RequestPermission(): Promise<boolean>
    HasPermission(): Promise<boolean>
  }

  /** 窗口服务 */
  export const WindowService: {
    SetTitle(title: string): Promise<void>
    Minimize(): Promise<void>
    Maximize(): Promise<void>
    Close(): Promise<void>
    Quit(): Promise<void>
    SetAlwaysOnTop(on: boolean): Promise<void>
    IsMaximized(): Promise<boolean>
    OpenURL(url: string): Promise<void>
    ShowAndFocus(): Promise<void>
  }

  /** 更新服务 */
  export const UpdaterService: {
    CheckForUpdate(): Promise<string>
    IsAutoUpdateEnabled(): Promise<boolean>
    SetAutoUpdateEnabled(enabled: boolean): Promise<void>
    InstallUpdate(): Promise<void>
    GetCurrentVersion(): Promise<string>
  }
}
