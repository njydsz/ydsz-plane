/**
 * drawio.ts — draw.io 图表嵌入块
 *
 * 允许在文档中嵌入 draw.io 图表：
 *   - 支持直接粘贴 draw.io XML（mxGraphModel）；
 *   - 支持嵌入 iframe（通过 draw.io 在线查看器或自托管实例）；
 *   - 渲染为占位卡片，点击可打开编辑/查看。
 *
 * HTML 结构:
 * <div data-type="drawio" data-xml="..." data-viewer-url="..." data-title="...">
 */
import { Node, mergeAttributes } from "@tiptap/core"

export interface DrawioOptions {
  HTMLAttributes: Record<string, any>
  /** draw.io 在线查看器地址（自托管时可替换） */
  viewerURL: string
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    drawio: {
      setDrawio: (opts: { xml?: string; title?: string; viewerURL?: string }) => ReturnType
      unsetDrawio: () => ReturnType
    }
  }
}

export const Drawio = Node.create<DrawioOptions>({
  name: "drawio",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
      viewerURL: "https://embed.diagrams.net/?embed=1",
    }
  },

  addAttributes() {
    return {
      xml: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-xml") || "",
        renderHTML: (attrs) => ({ "data-xml": attrs.xml || "" }),
      },
      title: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-title") || "",
        renderHTML: (attrs) => (attrs.title ? { "data-title": attrs.title } : {}),
      },
      viewerURL: {
        default: null,
        parseHTML: (el) => el.getAttribute("data-viewer-url"),
        renderHTML: (attrs) => (attrs.viewerURL ? { "data-viewer-url": attrs.viewerURL } : {}),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="drawio"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "drawio" }, HTMLAttributes),
    ]
  },

  addCommands() {
    return {
      setDrawio:
        (opts) =>
        ({ commands }) =>
          commands.insertContent({
            type: this.name,
            attrs: {
              xml: opts.xml || "",
              title: opts.title || "",
              viewerURL: opts.viewerURL || this.options.viewerURL,
            },
          }),
      unsetDrawio:
        () =>
        ({ commands }) =>
          commands.deleteNode(this.name),
    }
  },
})

export default Drawio
