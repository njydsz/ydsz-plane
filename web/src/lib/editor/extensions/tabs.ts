/**
 * tabs.ts — 标签页块扩展
 *
 * 基于 TipTap Node 扩展，实现 Notion/Plane 风格的可折叠标签页。
 * 支持多个标签页，每个标签页有标题和 block 内容。
 *
 * HTML 结构:
 * <div data-type="tabs">
 *   <div data-type="tab" data-tab-title="Tab 1" data-tab-active="true">...</div>
 *   <div data-type="tab" data-tab-title="Tab 2">...</div>
 * </div>
 */

import { Node, mergeAttributes } from "@tiptap/core"

export interface TabsOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    tabs: {
      setTabs: () => ReturnType
      toggleTabs: () => ReturnType
      unsetTabs: () => ReturnType
      addTab: () => ReturnType
      removeTab: () => ReturnType
    }
  }
}

export const Tabs = Node.create<TabsOptions>({
  name: "tabs",

  group: "block",
  content: "tab+",
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {}
  },

  parseHTML() {
    return [{ tag: 'div[data-type="tabs"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "tabs" }, HTMLAttributes),
      0,
    ]
  },

  addCommands() {
    return {
      setTabs:
        () =>
        ({ commands }) =>
          commands.wrapIn(this.name),
      toggleTabs:
        () =>
        ({ commands }) =>
          commands.toggleWrap(this.name),
      unsetTabs:
        () =>
        ({ commands }) =>
          commands.lift(this.name),
      addTab:
        () =>
        ({ state, chain }) => {
          const { $to } = state.selection
          const tabStart = $to.end()
          return chain()
            .focus()
            .insertContentAt(tabStart, {
              type: "tab",
              attrs: { title: "新标签页", active: false },
              content: [{ type: "paragraph" }],
            })
            .run()
        },
      removeTab:
        () =>
        ({ state, chain }) => {
          const { $from } = state.selection
          const tabNode = $from.node($from.depth - 1)
          if (tabNode?.type?.name === "tab") {
            return chain().focus().deleteNode("tab").run()
          }
          return chain().focus().run()
        },
    }
  },
})

export const Tab = Node.create({
  name: "tab",

  group: "block",
  content: "block+",
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      title: {
        default: "标签页",
        parseHTML: (element) => element.getAttribute("data-tab-title") || "标签页",
        renderHTML: (attributes) => ({
          "data-tab-title": attributes.title,
        }),
      },
      active: {
        default: false,
        parseHTML: (element) => element.getAttribute("data-tab-active") === "true",
        renderHTML: (attributes) => ({
          "data-tab-active": attributes.active ? "true" : "false",
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="tab"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "tab" }, HTMLAttributes),
      0,
    ]
  },
})

export const tabsExtensions = [Tabs, Tab]

export { Tabs as TabContainer, Tab as TabItem }
