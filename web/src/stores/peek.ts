/**
 * Peek Overview 状态管理 — 右侧抽屉预览工作项。
 *
 * 职责：
 *  - 维护当前 peek 的工作项标识（workspace/project/issue 三元组）
 *  - 提供 open/close/toggle 方法供列表/看板/表格调用
 *  - 支持上下文列表（当前视图可见的工作项 ID 顺序），
 *    用于 Peek 内的上一个/下一个连续预览
 *  - 持久化展开状态到 URL query（peek=issueId），刷新后恢复
 */
import { defineStore } from "pinia"
import { computed, ref } from "vue"

/** Peek 抽屉目标（workspaceId + projectId + issueId 三元组）。 */
export interface PeekTarget {
  workspaceId: number
  projectId: number
  issueId: number
}

/** Peek Overview 状态管理 — 右侧抽屉预览工作项的打开/关闭/目标。 */
export const usePeekStore = defineStore("peek", () => {
  const target = ref<PeekTarget | null>(null)
  const visible = ref(false)

  /** 当前视图上下文中的工作项 ID 列表（按视图展示顺序，供连续预览导航）。 */
  const contextList = ref<number[]>([])

  /** 当前工作项在上下文列表中的索引（不存在时为 -1）。 */
  const currentIndex = computed(() => {
    if (!target.value) return -1
    return contextList.value.indexOf(target.value.issueId)
  })

  const hasPrev = computed(() => currentIndex.value > 0)
  const hasNext = computed(() => currentIndex.value >= 0 && currentIndex.value < contextList.value.length - 1)

  function open(workspaceId: number, projectId: number, issueId: number) {
    target.value = { workspaceId, projectId, issueId }
    visible.value = true
  }

  /** 打开 Peek 并携带当前视图的上下文列表（用于连续预览）。 */
  function openWithContext(workspaceId: number, projectId: number, issueId: number, list: number[]) {
    contextList.value = list ?? []
    open(workspaceId, projectId, issueId)
  }

  /** 在上下文列表中导航：dir=1 下一个，dir=-1 上一个。 */
  function navigate(dir: 1 | -1): PeekTarget | null {
    if (!target.value) return null
    const idx = currentIndex.value
    const nextIdx = idx + dir
    if (nextIdx < 0 || nextIdx >= contextList.value.length) return null
    const nextIssueId = contextList.value[nextIdx]
    target.value = { ...target.value, issueId: nextIssueId }
    return target.value
  }

  function close() {
    visible.value = false
    target.value = null
    contextList.value = []
  }

  function toggle(workspaceId: number, projectId: number, issueId: number) {
    if (visible.value && target.value?.issueId === issueId) {
      close()
    } else {
      open(workspaceId, projectId, issueId)
    }
  }

  return { target, visible, contextList, currentIndex, hasPrev, hasNext, open, openWithContext, navigate, close, toggle }
})
