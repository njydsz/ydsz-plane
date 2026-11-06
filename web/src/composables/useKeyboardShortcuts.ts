/**
 * 全局快捷键（对齐 Plane）。
 *
 * 单键（工作项上下文）：
 *   S → 状态    P → 优先级    A → 指派    I → 指派给我
 *   L → 标签    E → 估算      M → 模块    C → 迭代
 *   Shift+S → 订阅    Cmd+. → 复制 ID    Cmd+Shift+' → 复制标题
 *   Cmd+Backspace → 删除
 *
 * 序列键（导航）：
 *   G H → 首页 / 工作台
 *   G X → 收件箱（通知）
 *   G Y → 我的工作项
 *   G I → 工作项列表
 *   G D → 文档
 *   G C → 迭代
 *   G M → 模块
 *   G V → 视图
 *   G K → 搜索
 *   G S → 设置
 *   G R → 报表/分析
 *   G A → 自动化
 *   G J → 里程碑
 *   G T → 团队空间
 *   G N → 倡议
 *   G B → 知识库
 *   G U → 项目
 *   G P → 项目（同 U）
 *
 * 序列键（创建）：
 *   N I → 新建工作项    N D → 新建文档
 *   N C → 新建迭代      N M → 新建模块
 *   N V → 新建视图      N P → 新建项目
 *   N T → 新建团队空间  N B → 新建知识库页面
 *
 * 其他：
 *   ? → 帮助面板      / → 聚焦搜索      Cmd+K → 命令面板
 *   Cmd+B → 侧边栏折叠   Cmd+/ → 帮助
 */
import { onBeforeUnmount, onMounted, ref } from "vue"

import { shortcuts } from "@/lib/shortcut"

export type Shortcut = {
  key: string
  label: string
  description: string
  ctrl?: boolean
  meta?: boolean
  shift?: boolean
  /** 序列键前缀（如 'G' → 'H'） */
  prefix?: string
  scope?: "global" | "input-excluded"
}

/** Plane 风格快捷键帮助列表 */
const HELP_LIST: Shortcut[] = [
  // ---- 通用 ----
  { key: "/", label: "/", description: "聚焦全局搜索" },
  { key: "Cmd+K", label: "⌘K", description: "命令面板（搜索 / 创建 / 跳转）" },
  { key: "?", label: "?", description: "打开 / 关闭快捷键帮助面板" },
  { key: "Cmd+/", label: "⌘/", description: "快捷键帮助" },
  { key: "Cmd+B", label: "⌘B", description: "折叠 / 展开侧边栏" },
  { key: "Esc", label: "Esc", description: "关闭当前弹窗 / 侧面板" },

  // ---- 导航（G 前缀）----
  { key: "H", label: "G H", description: "回到工作台", prefix: "g" },
  { key: "X", label: "G X", description: "通知 / 收件箱", prefix: "g" },
  { key: "Y", label: "G Y", description: "我的工作项", prefix: "g" },
  { key: "I", label: "G I", description: "工作项列表", prefix: "g" },
  { key: "D", label: "G D", description: "文档", prefix: "g" },
  { key: "C", label: "G C", description: "迭代", prefix: "g" },
  { key: "M", label: "G M", description: "模块", prefix: "g" },
  { key: "V", label: "G V", description: "视图", prefix: "g" },
  { key: "K", label: "G K", description: "搜索", prefix: "g" },
  { key: "S", label: "G S", description: "设置", prefix: "g" },
  { key: "R", label: "G R", description: "报表 / 分析", prefix: "g" },
  { key: "A", label: "G A", description: "自动化", prefix: "g" },
  { key: "J", label: "G J", description: "里程碑", prefix: "g" },
  { key: "T", label: "G T", description: "团队空间", prefix: "g" },
  { key: "N", label: "G N", description: "倡议", prefix: "g" },
  { key: "B", label: "G B", description: "知识库", prefix: "g" },
  { key: "U", label: "G U", description: "项目", prefix: "g" },

  // ---- 创建（N 前缀）----
  { key: "I", label: "N I", description: "新建工作项", prefix: "n" },
  { key: "D", label: "N D", description: "新建文档", prefix: "n" },
  { key: "C", label: "N C", description: "新建迭代", prefix: "n" },
  { key: "M", label: "N M", description: "新建模块", prefix: "n" },
  { key: "V", label: "N V", description: "新建视图", prefix: "n" },
  { key: "P", label: "N P", description: "新建项目", prefix: "n" },
  { key: "T", label: "N T", description: "新建团队空间", prefix: "n" },
  { key: "B", label: "N B", description: "新建知识库页面", prefix: "n" },

  // ---- 工作项操作 ----
  { key: "S", label: "S", description: "修改状态" },
  { key: "P", label: "P", description: "修改优先级" },
  { key: "A", label: "A", description: "修改指派人" },
  { key: "I", label: "I", description: "指派给我" },
  { key: "L", label: "L", description: "修改标签" },
  { key: "E", label: "E", description: "修改估算" },
  { key: "M", label: "M", description: "修改模块" },
  { key: "C", label: "C", description: "修改迭代" },
  { key: "S", label: "⇧S", description: "订阅 / 取消订阅", shift: true },
  { key: "Cmd+.", label: "⌘.", description: "复制工作项 ID", meta: true },
  { key: "Cmd+Backspace", label: "⌘⌫", description: "删除工作项", meta: true },
]

/** 工作项上下文快捷键（用于帮助面板分组） */
export const WORK_ITEM_SHORTCUTS = HELP_LIST.filter((s) =>
  ["S", "P", "A", "I", "L", "E", "M", "C"].includes(s.label),
)

/** 导航快捷键（G 前缀） */
export const NAVIGATION_SHORTCUTS = HELP_LIST.filter((s) => s.prefix === "g")

/** 创建快捷键（N 前缀） */
export const CREATE_SHORTCUTS = HELP_LIST.filter((s) => s.prefix === "n")

/**
 * 注册 Plane 风格快捷键。
 *
 * @param handlers 各快捷键的处理函数；未提供的项不会注册。
 */
export function useKeyboardShortcuts(handlers: Partial<Record<string, () => void>> = {}) {
  const helpOpen = ref(false)
  const list = ref<Shortcut[]>([...HELP_LIST])

  function toggleHelp() {
    helpOpen.value = !helpOpen.value
  }

  // 帮助面板
  shortcuts.register({ key: "?", scope: "input-excluded", handler: toggleHelp })

  // 单键工作项操作
  for (const s of HELP_LIST) {
    if (s.prefix) continue
    if (s.ctrl || s.meta) continue
    if (s.key === "?") continue
    const h = handlers[s.key.toLowerCase()]
    if (h) {
      shortcuts.register({
        key: s.key,
        shiftKey: s.shift,
        scope: "input-excluded",
        handler: h,
      })
    }
  }

  // 序列键（G / N 前缀）
  for (const s of HELP_LIST) {
    if (!s.prefix) continue
    const h = handlers[`${s.prefix.toLowerCase()}${s.key.toLowerCase()}`]
    if (h) {
      shortcuts.registerChord({
        prefix: s.prefix,
        key: s.key,
        handler: h,
        scope: "input-excluded",
      })
    }
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

  return {
    helpOpen,
    shortcuts: list,
    openHelp: () => (helpOpen.value = true),
    closeHelp: () => (helpOpen.value = false),
    register: shortcuts.register.bind(shortcuts),
    unregister: shortcuts.unregister.bind(shortcuts),
    registerChord: shortcuts.registerChord.bind(shortcuts),
  }
}
