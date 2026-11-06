/**
 * useYjsCollab — TipTap 编辑器与 Y.js CRDT 的绑定。
 *
 * 能力：
 *   - 将编辑器的 ProseMirror 文档映射到 Y.XmlFragment；
 *   - 双向同步（本地编辑 → Y.Doc → 远端；远端更新 → Y.Doc → 编辑器）；
 *   - awareness（光标/选区/用户信息）广播；
 *   - 断线重连后自动重新同步；
 *   - 文档快照持久化（保存时）。
 *
 * 用法：
 *   const collab = useYjsCollab({
 *     docID: `project_page:42`,
 *     getEditor: () => editor.value,
 *     user: { id, name, color },
 *   })
 *   collab.destroy()
 */
import { onBeforeUnmount, ref, type Ref } from "vue"
import { yDocToProsemirror, ySyncPluginKey } from "y-prosemirror"

import { YjsProvider, MESSAGE_AWARENESS } from "@/lib/collab/provider"
import { createWsTransport } from "@/lib/collab/transport"
import { wsClient } from "@/lib/ws-client"

/** 协作者信息 */
export interface CollabUser {
  id: number
  name: string
  color: string
}

/** awareness 状态 */
export interface AwarenessState {
  user?: CollabUser
  cursor?: { anchor: number; head: number } | null
  selection?: unknown
  [key: string]: unknown
}

export interface UseYjsCollabOptions {
  /** 文档标识（工作空间内唯一，如 "project_page:42"） */
  docID: string
  /** 获取编辑器实例 */
  getEditor: () => any | null
  /** 当前用户 */
  user: Ref<CollabUser> | CollabUser
  /** WebSocket 基地址（默认从 ws-client 推断） */
  wsURL?: string
  /** 是否启用（默认 true） */
  enabled?: boolean
}

/** 协作者（来自 awareness） */
export interface Peer {
  clientID: number
  user?: CollabUser
  cursor?: { anchor: number; head: number } | null
}

const COLOR_PALETTE = [
  "#3b82f6", "#10b981", "#f59e0b", "#8b5cf6",
  "#ec4899", "#14b8a6", "#f97316", "#6366f1",
]

export function colorForUser(id: number): string {
  return COLOR_PALETTE[Math.abs(id) % COLOR_PALETTE.length]
}

/**
 * useYjsCollab — 绑定 TipTap 与 Y.js。
 */
export function useYjsCollab(options: UseYjsCollabOptions) {
  const enabled = options.enabled ?? true
  const provider = ref<YjsProvider | null>(null)
  const binding = ref<{ destroy: () => void } | null>(null)
  const peers = ref<Peer[]>([])
  const connected = ref(false)
  const synced = ref(false)
  const syncedOnce = ref(false)

  /** 当前 awareness 状态 */
  let localState: AwarenessState = {}

  // ------------------------------------------------------------------
  // 初始化
  // ------------------------------------------------------------------

  function init() {
    if (!enabled) return

    // 1. 创建传输层
    const transport = createWsTransport({
      getSocket: () => (wsClient as unknown as { socket?: WebSocket }).socket ?? null,
      isConnected: () => wsClient.isConnected,
      onOpen: (cb) => wsClient.on("open", cb) as unknown as () => void,
      onClose: (cb) => wsClient.on("close", cb) as unknown as () => void,
    })

    // 2. 创建 YjsProvider
    const p = new YjsProvider(options.docID, transport)
    provider.value = p

    p.onConnected = (c) => { connected.value = c }
    p.onSync = (s) => {
      synced.value = s
      if (s) syncedOnce.value = true
    }

    // 3. 设置本地 awareness
    const user = "value" in options.user ? options.user.value : options.user
    localState = { user: { ...user, color: user.color || colorForUser(user.id) } }
    p.awareness.setLocalState(localState)

    // 4. 监听 awareness 变化 → 更新 peers
    const onAwareness = () => {
      const states = p.awareness.getStates()
      const list: Peer[] = []
      states.forEach((state, clientID) => {
        if (clientID === p.awareness.clientID) return
        const s = state as AwarenessState
        if (!s.user) return
        list.push({ clientID, user: s.user, cursor: s.cursor ?? null })
      })
      peers.value = list
    }
    p.awareness.on("change", onAwareness)

    // 5. 绑定编辑器（延迟到编辑器可用）
    const tryBind = () => {
      const editor = options.getEditor()
      if (!editor || binding.value) return

      // ySync 插件由 y-prosemirror 提供；此处通过 yCursorPlugin/ySyncPlugin 绑定
      const yXmlFragment = p.doc.getXmlFragment("prosemirror")

      // 创建 ProseMirror 插件
      const { ySyncPlugin, yCursorPlugin, yUndoPlugin } = require("y-prosemirror")

      const syncPlugin = ySyncPlugin(yXmlFragment)
      const cursorPlugin = yCursorPlugin(p.awareness)
      const undoPlugin = yUndoPlugin()

      // 注入插件（通过 editor 的 extension 机制）
      editor.registerPlugin?.(syncPlugin)
      editor.registerPlugin?.(cursorPlugin)
      editor.registerPlugin?.(undoPlugin)

      binding.value = {
        destroy: () => {
          try {
            editor.unregisterPlugin?.(syncPlugin)
            editor.unregisterPlugin?.(cursorPlugin)
            editor.unregisterPlugin?.(undoPlugin)
          } catch {
            /* 忽略 */
          }
        },
      }
    }

    // 立即尝试 + 延迟重试
    tryBind()
    const retryTimer = setInterval(() => {
      if (binding.value) {
        clearInterval(retryTimer)
        return
      }
      tryBind()
    }, 300)

    // 6. 广播光标位置到 awareness
    const onSelectionChange = () => {
      const editor = options.getEditor()
      if (!editor) return
      const { state } = editor
      const { from, to } = state.selection
      localState = {
        ...localState,
        cursor: { anchor: from, head: to },
      }
      p.awareness.setLocalState(localState)
    }

    const editor = options.getEditor()
    editor?.on?.("selectionUpdate", onSelectionChange)

    // 清理
    cleanupFns.push(() => {
      clearInterval(retryTimer)
      editor?.off?.("selectionUpdate", onSelectionChange)
      p.awareness.off("change", onAwareness)
      binding.value?.destroy()
      p.destroy()
      ;(transport as unknown as { destroy?: () => void }).destroy?.()
    })
  }

  const cleanupFns: Array<() => void> = []

  function destroy() {
    cleanupFns.forEach((fn) => fn())
    cleanupFns.length = 0
    provider.value = null
    binding.value = null
    peers.value = []
  }

  // 延迟初始化（等待 ws-client 连接）
  const initTimer = setTimeout(init, 500)

  onBeforeUnmount(() => {
    clearTimeout(initTimer)
    destroy()
  })

  return {
    provider,
    peers,
    connected,
    synced,
    syncedOnce,
    destroy,
    /** 获取文档快照 */
    encodeState: () => provider.value?.encodeState() ?? new Uint8Array(),
    /** 应用快照 */
    applyState: (state: Uint8Array) => provider.value?.applyState(state),
  }
}

// ----------------------------------------------------------------------
// 服务端持久化辅助（前端侧）
// ----------------------------------------------------------------------

/** 将 Uint8Array 编码为 Base64（用于保存到服务端） */
export function encodeStateToBase64(bytes: Uint8Array): string {
  let binary = ""
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
  }
  return btoa(binary)
}

/** 将 Base64 解码为 Uint8Array */
export function decodeStateFromBase64(b64: string): Uint8Array {
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

/** 导出 y-prosemirror 的文档转换辅助 */
export { yDocToProsemirror, ySyncPluginKey, MESSAGE_AWARENESS }
