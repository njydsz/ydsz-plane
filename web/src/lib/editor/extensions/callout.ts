/**
 * callout.ts — Callout 块级提示框扩展
 *
 * 基于 TipTap Node 扩展，实现类似 Notion / Plane 的信息框样式。
 * 支持 4 种语义类型：info、warning、error、success
 * 输出 HTML: <div data-type="callout" data-callout-type="info">...</div>
 */

import { Node, mergeAttributes } from "@tiptap/core"

export type CalloutType = "info" | "warning" | "error" | "success"

export interface CalloutOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    callout: {
      setCallout: (type?: CalloutType) => ReturnType
      toggleCallout: (type?: CalloutType) => ReturnType
      unsetCallout: () => ReturnType
    }
  }
}

export const Callout = Node.create<CalloutOptions>({
  name: "callout",

  group: "block",
  content: "block+",
  defining: true,

  addAttributes() {
    return {
      type: {
        default: "info",
        parseHTML: (element) => element.getAttribute("data-callout-type") || "info",
        renderHTML: (attributes) => ({
          "data-callout-type": attributes.type,
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="callout"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "callout" }, HTMLAttributes),
      0,
    ]
  },

  addCommands() {
    return {
      setCallout:
        (type = "info") =>
        ({ commands }) =>
          commands.wrapIn(this.name, { type }),
      toggleCallout:
        (type = "info") =>
        ({ commands }) =>
          commands.toggleWrap(this.name, { type }),
      unsetCallout:
        () =>
        ({ commands }) =>
          commands.lift(this.name),
    }
  },
})

export default Callout
