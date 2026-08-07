/**
 * Peek Overview 状态管理 — 右侧抽屉预览工作项。
 *
 * 职责：
 *  - 维护当前 peek 的工作项标识（workspace/project/issue 三元组）
 *  - 提供 open/close/toggle 方法供列表/看板/表格调用
 *  - 持久化展开状态到 URL query（peek=issueId），刷新后恢复
 */
import { defineStore } from "pinia"
import { ref } from "vue"

/** Peek 抽屉目标（workspaceSlug + projectId + issueId 三元组）。 */
export interface PeekTarget {
  workspaceSlug: string
  projectId: number
  issueId: number
}

/** Peek Overview 状态管理 — 右侧抽屉预览工作项的打开/关闭/目标。 */
export const usePeekStore = defineStore("peek", () => {
  const target = ref<PeekTarget | null>(null)
  const visible = ref(false)

  function open(wsSlug: string, projectId: number, issueId: number) {
    target.value = { workspaceSlug: wsSlug, projectId, issueId }
    visible.value = true
  }

  function close() {
    visible.value = false
    target.value = null
  }

  function toggle(wsSlug: string, projectId: number, issueId: number) {
    if (visible.value && target.value?.issueId === issueId) {
      close()
    } else {
      open(wsSlug, projectId, issueId)
    }
  }

  return { target, visible, open, close, toggle }
})
