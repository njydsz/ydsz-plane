/**
 * embed.ts — 嵌入内容扩展
 *
 * 支持嵌入外部内容：视频 (YouTube/Bilibili)、iframe、代码沙箱等。
 * 通过输入 URL + 选择类型来插入。
 *
 * HTML 结构:
 * <div data-type="embed" data-embed-type="video" data-embed-url="...">
 */

import { Node, mergeAttributes } from "@tiptap/core"

export type EmbedType = "video" | "iframe" | "link-preview"

export interface EmbedOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    embed: {
      setEmbed: (type: EmbedType, url: string) => ReturnType
      toggleEmbed: () => ReturnType
      unsetEmbed: () => ReturnType
    }
  }
}

export const Embed = Node.create<EmbedOptions>({
  name: "embed",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      type: {
        default: "video",
        parseHTML: (element) => element.getAttribute("data-embed-type") || "video",
        renderHTML: (attributes) => ({
          "data-embed-type": attributes.type,
        }),
      },
      url: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-embed-url") || "",
        renderHTML: (attributes) => ({
          "data-embed-url": attributes.url,
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="embed"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "embed" }, HTMLAttributes),
    ]
  },

  addCommands() {
    return {
      setEmbed:
        (type, url) =>
        ({ commands }) =>
          commands.insertContent({
            type: this.name,
            attrs: { type, url },
          }),
      toggleEmbed:
        () =>
        ({ commands }) =>
          commands.toggleWrap(this.name),
      unsetEmbed:
        () =>
        ({ commands }) =>
          commands.lift(this.name),
    }
  },
})

/**
 * WorkItemEmbed — 内嵌工作项引用块
 *
 * 在文档中嵌入引用一个工作项（Issue/Requirement/Task/Defect）
 * 显示为卡片样式，可点击查看完整信息。
 *
 * HTML 结构:
 * <div data-type="work-item-embed" data-issue-id="123" data-issue-type="task">
 */
export interface WorkItemEmbedOptions {
  HTMLAttributes: Record<string, any>
}

export const WorkItemEmbed = Node.create<WorkItemEmbedOptions>({
  name: "workItemEmbed",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      issueId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-issue-id"),
        renderHTML: (attributes) => ({
          "data-issue-id": attributes.issueId,
        }),
      },
      issueType: {
        default: "task",
        parseHTML: (element) => element.getAttribute("data-issue-type") || "task",
        renderHTML: (attributes) => ({
          "data-issue-type": attributes.issueType,
        }),
      },
      title: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-title") || "",
        renderHTML: (attributes) => ({
          "data-title": attributes.title,
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="work-item-embed"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(
        this.options.HTMLAttributes,
        { "data-type": "work-item-embed" },
        HTMLAttributes,
      ),
    ]
  },
})

/**
 * PageLink — 页面链接块
 *
 * 在文档中嵌入引用另一个知识库页面/项目页面。
 * 显示为带图标的链接卡片。
 *
 * HTML 结构:
 * <div data-type="page-link" data-page-id="456" data-page-space="kb">
 */
export interface PageLinkOptions {
  HTMLAttributes: Record<string, any>
}

export const PageLink = Node.create<PageLinkOptions>({
  name: "pageLink",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return {
      HTMLAttributes: {},
    }
  },

  addAttributes() {
    return {
      pageId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-page-id"),
        renderHTML: (attributes) => ({
          "data-page-id": attributes.pageId,
        }),
      },
      space: {
        default: "project",
        parseHTML: (element) => element.getAttribute("data-page-space") || "project",
        renderHTML: (attributes) => ({
          "data-page-space": attributes.space,
        }),
      },
      title: {
        default: "",
        parseHTML: (element) => element.getAttribute("data-title") || "",
        renderHTML: (attributes) => ({
          "data-title": attributes.title,
        }),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="page-link"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(
        this.options.HTMLAttributes,
        { "data-type": "page-link" },
        HTMLAttributes,
      ),
    ]
  },
})

export const embedExtensions = [Embed, WorkItemEmbed, PageLink]

export { Embed as EmbedBlock, WorkItemEmbed as WorkItemEmbedBlock, PageLink as PageLinkBlock }
