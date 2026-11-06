/**
 * 全局快捷键管理器。
 *
 * 支持两种形态：
 *   1. 单键 / 修饰键组合：S（状态）、P（优先级）、A（指派）、Cmd+K（搜索）…
 *   2. 序列键（chord）：先按 G 再按 H/X/Y…，或先按 N 再按 I/D/V/C…
 *
 * 序列键实现：
 *   - 首个键命中时进入 pending 状态并启动超时（默认 800ms）；
 *   - 超时前命中后续键则触发并清空；
 *   - 超时未命中则清空 pending（不触发）。
 *
 * 使用方式：
 *   shortcuts.register({ key: 's', handler: () => openStatePanel() })
 *   shortcuts.registerChord({ prefix: 'g', key: 'h', handler: () => goHome() })
 */
type ShortcutDef = {
  key: string
  ctrlKey?: boolean
  metaKey?: boolean
  shiftKey?: boolean
  handler: () => void
  scope?: 'global' | 'input-excluded'
  /** 是否为序列键的第二段 */
  chord?: boolean
  /** 序列前缀（仅 chord=true 时有效） */
  prefix?: string
}

/** 序列键超时（毫秒） */
const CHORD_TIMEOUT = 800

class ShortcutManager {
  private registry: Map<string, ShortcutDef> = new Map()
  /** 当前待定的序列前缀 */
  private pendingPrefix: string | null = null
  private chordTimer: ReturnType<typeof setTimeout> | null = null

  register(def: ShortcutDef) {
    const id = this.defId(def)
    this.registry.set(id, def)
  }

  unregister(key: string) {
    this.registry.delete(key)
  }

  /** 注册序列键（等价于 register({ chord: true, prefix, key })） */
  registerChord(opts: { prefix: string; key: string; handler: () => void; scope?: 'global' | 'input-excluded' }) {
    this.register({
      key: opts.key,
      chord: true,
      prefix: opts.prefix,
      handler: opts.handler,
      scope: opts.scope ?? 'input-excluded',
    })
  }

  /** 注销序列键 */
  unregisterChord(prefix: string, key: string) {
    this.registry.delete(this.defId({ key, chord: true, prefix, handler: () => {} }))
  }

  private defId(def: ShortcutDef): string {
    if (def.chord) {
      return `chord:${def.prefix}:${def.key}`
    }
    const mod = [def.ctrlKey || def.metaKey ? 'mod' : '', def.shiftKey ? 'shift' : ''].filter(Boolean).join('+')
    return `${mod}${mod ? '+' : ''}${def.key}`
  }

  handle(e: KeyboardEvent) {
    const target = e.target as HTMLElement
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable
    const key = e.key.toLowerCase()

    // 序列键第二段优先匹配
    if (this.pendingPrefix) {
      for (const [, def] of this.registry) {
        if (!def.chord || def.prefix !== this.pendingPrefix) continue
        if (isInput && def.scope !== 'global') continue
        if (key === def.key.toLowerCase() && !e.ctrlKey && !e.metaKey && !e.altKey) {
          e.preventDefault()
          this.clearPending()
          def.handler()
          return
        }
      }
    }

    for (const [, def] of this.registry) {
      if (isInput && def.scope !== 'global') continue

      // 序列键第一段：无修饰键的单个字母，且存在以该键为前缀的 chord
      if (!def.chord) {
        const hasChordWithPrefix = this.hasChordPrefix(def.key)
        if (hasChordWithPrefix && !e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
          if (key === def.key.toLowerCase()) {
            e.preventDefault()
            this.setPending(def.key)
            return
          }
          continue
        }
      }

      // 序列键第二段已在上面处理，跳过普通匹配
      if (def.chord) continue

      const keyMatch = key === def.key.toLowerCase()
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
      this.clearPending()
      window.dispatchEvent(new CustomEvent('shortcut:escape'))
    }
  }

  /** 是否存在以给定键为前缀的序列键 */
  private hasChordPrefix(key: string): boolean {
    for (const [, def] of this.registry) {
      if (def.chord && def.prefix?.toLowerCase() === key.toLowerCase()) {
        return true
      }
    }
    return false
  }

  private setPending(prefix: string) {
    this.clearPending()
    this.pendingPrefix = prefix
    this.chordTimer = setTimeout(() => this.clearPending(), CHORD_TIMEOUT)
  }

  private clearPending() {
    this.pendingPrefix = null
    if (this.chordTimer) {
      clearTimeout(this.chordTimer)
      this.chordTimer = null
    }
  }

  /** 清空全部注册（主要用于测试） */
  reset() {
    this.registry.clear()
    this.clearPending()
  }
}

/** 全局快捷键管理器单例。 */
export const shortcuts = new ShortcutManager()

// 初始化全局监听
if (typeof window !== 'undefined') {
  document.addEventListener('keydown', (e) => shortcuts.handle(e))
}
