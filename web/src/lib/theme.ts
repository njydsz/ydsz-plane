/**
 * ThemeManager — 主题切换（light / dark / high-contrast）。
 *
 * high-contrast 对标信创无障碍需求：
 *   - 监听 localStorage('ydsz-theme') 与系统 prefers-contrast: more
 *   - 应用 data-theme 属性到 document.documentElement
 *   - 暴露 currentTheme ref 供 UI 主题切换器使用
 */
export type ThemeName = 'light' | 'dark' | 'high-contrast'

const STORAGE_KEY = 'ydsz-theme'

/** 当前活动主题（响应式入口）。 */
import { ref } from 'vue'

export const currentTheme = ref<ThemeName>('light')

/** 解析系统/持久化的初始主题。 */
function resolveInitial(): ThemeName {
  try {
    const saved = localStorage.getItem(STORAGE_KEY) as ThemeName | null
    if (saved === 'light' || saved === 'dark' || saved === 'high-contrast') {
      return saved
    }
  } catch {
    /* localStorage 不可用（隐私模式）回退 */
  }
  // 系统级高对比偏好
  if (window.matchMedia?.('(prefers-contrast: more)').matches) {
    return 'high-contrast'
  }
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

/** 应用主题到 document。 */
export function applyTheme(theme: ThemeName) {
  const root = document.documentElement
  root.setAttribute('data-theme', theme)
  // 为 Tailwind v4 提供 class 触发（dark 变体）
  root.classList.toggle('dark', theme === 'dark')
  root.classList.toggle('high-contrast', theme === 'high-contrast')
  try { localStorage.setItem(STORAGE_KEY, theme) } catch { /* ignore */ }
  currentTheme.value = theme
}

/** 在应用启动时同步一次主题。 */
export function initTheme(): ThemeName {
  const theme = resolveInitial()
  applyTheme(theme)
  return theme
}

/** 监听系统主题变化（仅限 light/dark 切换，不覆盖 high-contrast）。 */
export function watchSystemTheme() {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener?.('change', () => {
    if (currentTheme.value === 'high-contrast') return
    applyTheme(mq.matches ? 'dark' : 'light')
  })
  const hc = window.matchMedia('(prefers-contrast: more)')
  hc.addEventListener?.('change', () => {
    if (hc.matches && currentTheme.value !== 'high-contrast') {
      applyTheme('high-contrast')
    }
  })
}
