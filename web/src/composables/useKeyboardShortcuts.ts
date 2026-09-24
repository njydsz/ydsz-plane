/**
 * 全局快捷键（Register-all + 帮助面板）。
 *
 * 已通过 CommandPalette 自管 Ctrl/Cmd+K，本模块不再重复注册 Ctrl/K；
 * 仅接管：
 *   ?           → 切换快捷键帮助面板
 *   c           → 快速创建（在项目 / 空间视图下触发 issue-create 事件）
 *
 * escape 让给各弹窗自己的 Esc 处理；本模块只读、不拦截。
 *
 * 使用方式（App.vue onMounted 调用一次）：
 *   const { helpOpen, shortcuts, register } = useKeyboardShortcuts()
 *   register({ key: 'c', handler: () => router.push({ name: 'issue-create' })})
 */
import { onBeforeUnmount, onMounted, ref } from "vue"

import { shortcuts } from "@/lib/shortcut"

export type Shortcut = {
  key: string
  label: string
  description: string
  ctrl?: boolean
  meta?: boolean
  scope?: "global" | "input-excluded"
}

const HELP_LIST: Shortcut[] = [
  { key: "/", label: "/", description: "聚焦全局搜索" },
  { key: "?", label: "?", description: "打开 / 关闭快捷键帮助面板" },
  { key: "c", label: "C", description: "快速创建（根据当前视图创建需求 / 任务 / 缺陷）" },
  { key: "Esc", label: "Esc", description: "关闭当前弹窗 / 侧面板" },
]

export function useKeyboardShortcuts(initial: Shortcut[] = []) {
  const helpOpen = ref(false)
  const list = ref<Shortcut[]>([...HELP_LIST, ...initial])

  function toggleHelp() {
    helpOpen.value = !helpOpen.value
  }

  // 将 "帮助面板" 注册到底层 ShortcutManager（'?' 切换，'/' 聚焦搜索，'c' 创建为占位）。
  shortcuts.register({ key: "?", scope: "input-excluded", handler: toggleHelp })
  for (const s of initial) {
    shortcuts.register({
      key: s.key,
      ctrlKey: s.ctrl,
      metaKey: s.meta,
      scope: s.scope ?? "input-excluded",
      handler: () => {}, // 实际 handler 由各视图自行 register；此处仅占位列帮助
    })
  }

  function onWindowKeydown(e: KeyboardEvent) {
    shortcuts.handle(e)
    if (e.key === "Escape" && helpOpen.value) {
      e.stopPropagation()
      helpOpen.value = false
    }
  }

  onMounted(() => window.addEventListener("keydown", onWindowKeydown, true))
  onBeforeUnmount(() => window.removeEventListener("keydown", onWindowKeydown, true))

  return { helpOpen, shortcuts: list, openHelp: () => (helpOpen.value = true), closeHelp: () => (helpOpen.value = false), register: shortcuts.register.bind(shortcuts), unregister: shortcuts.unregister.bind(shortcuts) }
}
