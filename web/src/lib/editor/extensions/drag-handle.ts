/**
 * drag-handle.ts — 块级拖放支持
 *
 * 基于 TipTap Extension 实现 Notion/Plane 风格的块拖放。
 * 每个块级元素可通过拖动手柄进行排序和移动。
 *
 * 实现原理：
 * 1. 在编辑器内容区域监听拖放事件
 * 2. 检测块级元素（paragraph/heading/list/table/callout 等）
 * 3. 通过 ProseMirror 事务移动节点
 * 4. 提供视觉反馈（拖放指示线）
 *
 * 使用方式：
 *   - 在编辑器容器上添加 draggable="true"
 *   - 通过快捷键或拖动手柄启动拖放
 *   - 使用 data-block-id 标识每个块
 */

import { Extension } from "@tiptap/core"
import { Plugin, PluginKey } from "@tiptap/pm/state"
import { Node as PMNode } from "@tiptap/pm/model"

export interface DragHandleOptions {
  /** 拖动手柄的选择器（用于触发拖放的元素） */
  handleSelector: string
  /** 可拖放的块类型列表 */
  draggableBlocks: string[]
  /** 是否显示拖放指示线 */
  showIndicator: boolean
  /** 指示线 CSS 类名 */
  indicatorClass: string
}

/**
 * 查找包含指定位置的最近块级节点
 */
function findBlockAtPos(doc: PMNode, pos: number): { node: PMNode; pos: number } | null {
  let blockNode: PMNode | null = null
  let blockPos = pos

  // 遍历文档查找包含该位置的块
  doc.descendants((node, nodePos) => {
    if (nodePos > pos) return false
    if (node.isBlock && node.type.name !== "doc") {
      blockNode = node
      blockPos = nodePos
    }
    return true
  })

  return blockNode ? { node: blockNode, pos: blockPos } : null
}

export const DragHandle = Extension.create<DragHandleOptions>({
  name: "dragHandle",

  addOptions(): DragHandleOptions {
    return {
      handleSelector: ".block-drag-handle",
      draggableBlocks: [
        "paragraph",
        "heading",
        "bulletList",
        "orderedList",
        "taskList",
        "blockquote",
        "codeBlock",
        "horizontalRule",
        "table",
        "callout",
        "columns",
        "tabs",
        "embed",
        "workItemEmbed",
        "pageLink",
      ],
      showIndicator: true,
      indicatorClass: "drag-indicator",
      ...this.parent?.(),
    }
  },

  addProseMirrorPlugins() {
    const options = this.options

    return [
      new Plugin({
        key: new PluginKey("dragHandle"),
        props: {
          handleDOMEvents: {
            /**
             * 鼠标按下时检测是否点击了拖动手柄
             */
            mousedown(view, event) {
              const target = event.target as HTMLElement
              if (!target || !target.closest) return false

              const handle = target.closest(options.handleSelector)
              if (!handle) return false

              // 找到手柄对应的块元素
              const blockEl = handle.closest("[data-block-id]") ||
                handle.parentElement?.closest("[data-block-id]")
              if (!blockEl) return false

              // 获取块在文档中的位置
              const blockPos = view.posAtDOM(blockEl, 0)
              const blockInfo = findBlockAtPos(view.state.doc, blockPos)

              if (!blockInfo) return false
              if (!options.draggableBlocks.includes(blockInfo.node.type.name)) {
                return false
              }

              // 阻止默认行为，准备拖放
              event.preventDefault()
              return true
            },

            /**
             * 拖放开始时设置数据
             */
            dragstart(_view, event) {
              if (!event.dataTransfer) return false

              const target = event.target as HTMLElement
              const blockEl = target?.closest?.("[data-block-id]")
              if (!blockEl) return false

              // 设置拖放数据
              event.dataTransfer.effectAllowed = "move"
              event.dataTransfer.setData(
                "application/x-block",
                JSON.stringify({
                  blockId: blockEl.getAttribute("data-block-id"),
                  blockType: blockEl.getAttribute("data-block-type") || "paragraph",
                }),
              )

              // 添加拖放样式
              blockEl.classList.add("block-dragging")
              return false
            },

            /**
             * 拖放经过时显示指示线
             */
            dragover(_view, event) {
              if (!options.showIndicator) return false
              if (!event.dataTransfer?.types.includes("application/x-block")) return false

              event.preventDefault()
              event.dataTransfer.dropEffect = "move"
              return false
            },

            /**
             * 放置时移动块
             */
            drop(view, event) {
              if (!event.dataTransfer) return false

              const data = event.dataTransfer.getData("application/x-block")
              if (!data) return false

              event.preventDefault()

              const { blockId } = JSON.parse(data)
              if (!blockId) return false

              // 找到目标位置
              const dropPos = view.posAtCoords({
                left: event.clientX,
                top: event.clientY,
              })
              if (!dropPos) return false

              // 找到源块和目标块
              const doc = view.state.doc
              let sourcePos = -1
              let sourceSize = 0
              let targetPos = dropPos.pos

              doc.descendants((node, pos) => {
                if (node.attrs?.blockId === blockId) {
                  sourcePos = pos
                  sourceSize = node.nodeSize
                  return false
                }
                return true
              })

              if (sourcePos === -1) return false

              // 计算目标位置（在目标块之后插入）
              const tr = view.state.tr

              // 先删除源块
              tr.delete(sourcePos, sourcePos + sourceSize)

              // 调整目标位置（如果删除点在目标之前）
              if (sourcePos < targetPos) {
                targetPos -= sourceSize
              }

              // 在目标位置插入
              const slice = tr.doc.slice(sourcePos, sourcePos + sourceSize)
              tr.insert(targetPos, slice.content)

              view.dispatch(tr)
              return true
            },

            /**
             * 拖放结束时清理样式
             */
            dragend(_view, event) {
              const target = event.target as HTMLElement
              const blockEl = target?.closest?.("[data-block-id]")
              blockEl?.classList.remove("block-dragging")
              return false
            },
          },
        },
      }),
    ]
  },
})

export default DragHandle
