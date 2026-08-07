/**
 * 视图偏好持久化 — 基于 localStorage 的用户设置存储。
 *
 * 使用方式：
 *   import { prefs } from '@/lib/prefs'
 *   const filters = prefs.get('issue-filter', { priority: 'all' })
 *   prefs.set('issue-filter', { priority: 'high' })
 *   prefs.remove('issue-filter')
 */

const PREFIX = 'ydsz:'

export const prefs = {
  get<T>(key: string, fallback: T): T {
    try {
      const raw = localStorage.getItem(PREFIX + key)
      if (raw === null) return fallback
      return JSON.parse(raw) as T
    } catch {
      return fallback
    }
  },

  set<T>(key: string, value: T): void {
    try {
      localStorage.setItem(PREFIX + key, JSON.stringify(value))
    } catch { /* quota exceeded */ }
  },

  remove(key: string): void {
    localStorage.removeItem(PREFIX + key)
  },

  /** 获取上次访问的视图（用于板/列表切换） */
  getLastView(projectId: number | string, fallback: string = 'board'): string {
    return this.get(`view:${projectId}`, fallback)
  },

  /** 设置上次访问的视图 */
  setLastView(projectId: number | string, view: string): void {
    this.set(`view:${projectId}`, view)
  },

  /** 保存过滤器配置 */
  saveFilter(projectId: number | string, filter: Record<string, any>): void {
    this.set(`filter:${projectId}`, filter)
  },

  /** 读取过滤器配置 */
  getFilter(projectId: number | string): Record<string, any> | null {
    return this.get(`filter:${projectId}`, null)
  }
}
