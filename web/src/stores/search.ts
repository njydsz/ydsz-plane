/**
 * 搜索 Pinia store —— 管理全局搜索弹层与搜索结果的集中状态。
 *
 * 职责：
 *  - 持有搜索关键字（query）与当前结果（results）
 *  - 维护搜索面板开关状态（open），供全局快捷键/按钮统一触发
 *  - 封装调用 searchApi 的加载流程，失败时回退为空结果
 *
 * 设计说明：
 *  - open 与 query/results 耦合管理：关闭面板时自动清空上次搜索，
 *    避免下次打开时残留旧结果。
 *  - results 与 SearchResponse 结构对齐：{ results: { issues, sprints, versions }, total, time_ms }
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { searchApi, type SearchResponse } from '@/api/services/search'

/** 默认空结果结构（对齐后端 SearchResponse，避免模板空引用） */
const emptyResults: SearchResponse = {
  query: '',
  total: 0,
  results: { issues: [], sprints: [], versions: [] },
  time_ms: 0,
  suggestions: [],
}

export const useSearchStore = defineStore('search', () => {
  /** 当前搜索关键字 */
  const query = ref('')
  /** 最近一次搜索结果（空查询或失败时为 null） */
  const results = ref<SearchResponse | null>(null)
  /** 搜索请求进行中的标志位 */
  const loading = ref(false)
  /** 搜索面板是否展开 */
  const open = ref(false)

  /**
   * 在指定工作空间内执行搜索（最多返回 10 条）。
   * @param wsId 工作空间 ID
   * @param q 搜索关键字
   */
  async function search(wsId: number | string, q: string) {
    // 空白关键字直接清空结果，避免无意义请求
    if (!q.trim()) {
      results.value = null
      return
    }
    query.value = q
    loading.value = true
    try {
      results.value = await searchApi.searchWorkspace(wsId, { q, limit: 10 })
    } catch {
      // 失败时回退为默认空结果，保证 UI 有稳定的空态（避免模板崩溃）
      results.value = { ...emptyResults, query: q }
    } finally {
      loading.value = false
    }
  }

  /** 切换搜索面板开合；关闭时清空关键字与结果 */
  function toggle() {
    open.value = !open.value
    if (!open.value) {
      results.value = null
      query.value = ''
    }
  }

  /** 重置整个搜索状态（关键字、结果、面板） */
  function clear() {
    query.value = ''
    results.value = null
    open.value = false
  }

  return { query, results, loading, open, search, toggle, clear }
})
