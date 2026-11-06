/**
 * video.ts — 独立视频块
 *
 * 对标 Plane 的 Video 块：支持 YouTube / Bilibili / Vimeo / 直接视频 URL，
 * 自动识别平台并渲染对应嵌入（iframe 或 <video>）。
 *
 * HTML 结构:
 * <div data-type="video" data-video-url="..." data-video-platform="youtube" data-video-title="...">
 */
import { Node, mergeAttributes } from "@tiptap/core"

export type VideoPlatform = "youtube" | "bilibili" | "vimeo" | "direct"

export interface VideoOptions {
  HTMLAttributes: Record<string, any>
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    video: {
      setVideo: (url: string, title?: string) => ReturnType
      unsetVideo: () => ReturnType
    }
  }
}

/** 识别视频平台 */
export function detectVideoPlatform(url: string): VideoPlatform {
  const u = url.toLowerCase()
  if (u.includes("youtube.com") || u.includes("youtu.be")) return "youtube"
  if (u.includes("bilibili.com") || u.includes("b23.tv")) return "bilibili"
  if (u.includes("vimeo.com")) return "vimeo"
  return "direct"
}

/** 提取 YouTube 视频 ID */
export function extractYouTubeID(url: string): string {
  const m = url.match(/(?:youtube\.com\/(?:watch\?v=|embed\/|shorts\/)|youtu\.be\/)([\w-]{6,})/)
  return m ? m[1] : ""
}

/** 提取 Bilibili BV 号 */
export function extractBilibiliID(url: string): string {
  const m = url.match(/(?:bilibili\.com\/video\/(BV[\w]+)|b23\.tv\/(BV[\w]+))/i)
  return m ? (m[1] || m[2] || "") : ""
}

/** 提取 Vimeo 视频 ID */
export function extractVimeoID(url: string): string {
  const m = url.match(/vimeo\.com\/(\d+)/)
  return m ? m[1] : ""
}

/** 生成嵌入 URL */
export function buildEmbedURL(url: string): { platform: VideoPlatform; embedURL: string } {
  const platform = detectVideoPlatform(url)
  switch (platform) {
    case "youtube": {
      const id = extractYouTubeID(url)
      return { platform, embedURL: id ? `https://www.youtube-nocookie.com/embed/${id}` : url }
    }
    case "bilibili": {
      const id = extractBilibiliID(url)
      return {
        platform,
        embedURL: id
          ? `https://player.bilibili.com/player.html?bvid=${id}&autoplay=0`
          : url,
      }
    }
    case "vimeo": {
      const id = extractVimeoID(url)
      return { platform, embedURL: id ? `https://player.vimeo.com/video/${id}` : url }
    }
    default:
      return { platform, embedURL: url }
  }
}

export const Video = Node.create<VideoOptions>({
  name: "video",

  group: "block",
  atom: true,
  defining: true,

  addOptions() {
    return { HTMLAttributes: {} }
  },

  addAttributes() {
    return {
      url: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-video-url") || "",
        renderHTML: (attrs) => ({ "data-video-url": attrs.url || "" }),
      },
      platform: {
        default: "direct",
        parseHTML: (el) => el.getAttribute("data-video-platform") || "direct",
        renderHTML: (attrs) => ({ "data-video-platform": attrs.platform || "direct" }),
      },
      title: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-video-title") || "",
        renderHTML: (attrs) => (attrs.title ? { "data-video-title": attrs.title } : {}),
      },
      embedURL: {
        default: "",
        parseHTML: (el) => el.getAttribute("data-video-embed") || "",
        renderHTML: (attrs) => (attrs.embedURL ? { "data-video-embed": attrs.embedURL } : {}),
      },
    }
  },

  parseHTML() {
    return [{ tag: 'div[data-type="video"]' }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, { "data-type": "video" }, HTMLAttributes),
    ]
  },

  addCommands() {
    return {
      setVideo:
        (url, title = "") =>
        ({ commands }) => {
          const { platform, embedURL } = buildEmbedURL(url)
          return commands.insertContent({
            type: this.name,
            attrs: { url, platform, embedURL, title },
          })
        },
      unsetVideo:
        () =>
        ({ commands }) =>
          commands.deleteNode(this.name),
    }
  },
})

export default Video
