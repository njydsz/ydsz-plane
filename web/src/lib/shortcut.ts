/**
 * 全局快捷键管理器。
 *
 * 支持的快捷键：
 *   Ctrl+K / Cmd+K → 全局搜索
 *   C → 快速创建（TODO: 根据当前视图上下文）
 *   Esc → 关闭弹窗/侧面板
 *
 * 使用方式：
 *   import { shortcuts } from '@/lib/shortcut'
 *   shortcuts.register({ key: 'c', ctrlKey: false, handler: () => createIssue() })
 *   shortcuts.unregister('c')
 */
type ShortcutDef = {
  key: string
  ctrlKey?: boolean
  metaKey?: boolean
  shiftKey?: boolean
  handler: () => void
  scope?: 'global' | 'input-excluded'
}

class ShortcutManager {
  private registry: Map<string, ShortcutDef> = new Map()

  register(def: ShortcutDef) {
    this.registry.set(def.key, def)
  }

  unregister(key: string) {
    this.registry.delete(key)
  }

  handle(e: KeyboardEvent) {
    // 在输入框中不触发快捷键（Ctrl+K 除外）
    const target = e.target as HTMLElement
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable

    for (const [_, def] of this.registry) {
      if (isInput && def.scope !== 'global') continue
      const keyMatch = e.key.toLowerCase() === def.key.toLowerCase()
      const ctrlMatch = !!def.ctrlKey === (e.ctrlKey || e.metaKey)
      const shiftMatch = !!def.shiftKey === e.shiftKey
      if (keyMatch && ctrlMatch && shiftMatch) {
        e.preventDefault()
        def.handler()
        return
      }
    }
    // ESC 全局处理
    if (e.key === 'Escape') {
      // 触发自定义事件，各组件可监听
      window.dispatchEvent(new CustomEvent('shortcut:escape'))
    }
  }
}

export const shortcuts = new ShortcutManager()

// 初始化全局监听
if (typeof window !== 'undefined') {
  document.addEventListener('keydown', (e) => shortcuts.handle(e))
}
