/**
 * recent-visited.ts — 最近访问记录 Store
 *
 * 职责：
 *  - 跟踪用户最近访问的需求/任务/缺陷 / 迭代 / 版本/ 项目
 *  - 持久化到 localStorage（最大 20 条）
 *  - 供 Command Palette 的"最近访问"分组使用
 */
import { defineStore } from "pinia"
import { ref, watch } from "vue"

export interface RecentItem {
  id: number
  type: "issue" | "sprint" | "version" | "project"
  name: string
  projectName?: string
  workspaceId: number
  /** 路由 path，用于跳转 */
  href: string
  /** 访问时间戳 */
  visitedAt: number
  /** 图标 emoji */
  icon: string
}

const STORAGE_KEY = "ydsz-recent-visited"
const MAX_ITEMS = 20

function loadFromStorage(): RecentItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed
    return []
  } catch {
    return []
  }
}

function saveToStorage(items: RecentItem[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items.slice(0, MAX_ITEMS)))
  } catch {
    // 忽略序列化错误（隐私模式等）
  }
}

export const useRecentVisitedStore = defineStore("recentVisited", () => {
  const items = ref<RecentItem[]>(loadFromStorage())

  // 持久化
  watch(items, (val) => saveToStorage(val), { deep: true })

  /**
   * 添加访问记录
   */
  function add(item: Omit<RecentItem, "visitedAt">) {
    // 去重：同类型+同 ID 则更新
    const existing = items.value.findIndex(
      (i) => i.type === item.type && i.id === item.id,
    )
    if (existing !== -1) {
      items.value.splice(existing, 1)
    }
    items.value.unshift({
      ...item,
      visitedAt: Date.now(),
    })
    // 超过上限截断
    if (items.value.length > MAX_ITEMS) {
      items.value = items.value.slice(0, MAX_ITEMS)
    }
  }

  /**
   * 清除所有记录
   */
  function clear() {
    items.value = []
  }

  /**
   * 移除指定记录
   */
  function remove(id: number, type: string) {
    items.value = items.value.filter((i) => !(i.id === id && i.type === type))
  }

  /**
   * 获取指定工作空间的最近访问
   */
  function getByWorkspace(wsId: number): RecentItem[] {
    return items.value.filter((i) => i.workspaceId === wsId)
  }

  return { items, add, clear, remove, getByWorkspace }
})
