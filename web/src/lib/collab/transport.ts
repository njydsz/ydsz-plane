/**
 * transport.ts — 基于现有 ws-client 的 CRDT 二进制传输适配。
 *
 * 复用 ws-client 的连接/重连/鉴权，仅扩展二进制帧通道：
 *   - 发送：通过 wsClient.send({ binary }) 或原始 socket.send()
 *   - 接收：拦截 onmessage 的 ArrayBuffer/Blob 类型
 *
 * 注意：现有 ws-client 主要处理 JSON 消息；此处新增独立的二进制分发通道，
 * 与 JSON 消息互不干扰（按数据类型分流）。
 */
import type { CollabTransport } from "./provider"

export interface WsTransportOptions {
  /** 原始 WebSocket 实例（或工厂函数） */
  getSocket: () => WebSocket | null
  /** 是否已连接 */
  isConnected: () => boolean
  /** 连接状态变化回调注册 */
  onOpen: (cb: () => void) => () => void
  onClose: (cb: () => void) => () => void
}

/**
 * createWsTransport — 创建 CRDT 传输层。
 *
 * 设计取舍：
 *   - 不复用 ws-client 的 JSON 分发（避免二进制帧被 JSON.parse 破坏）；
 *   - 直接操作底层 socket 的 message 事件，按数据类型分流；
 *   - 连接状态通过 onClose/onOpen 感知。
 */
export function createWsTransport(opts: WsTransportOptions): CollabTransport {
  const messageHandlers = new Set<(data: Uint8Array) => void>()
  const statusHandlers = new Set<(connected: boolean) => void>()

  let attached = false
  let currentSocket: WebSocket | null = null

  const handleMessage = (event: MessageEvent) => {
    const data = event.data
    // 二进制帧：ArrayBuffer
    if (data instanceof ArrayBuffer) {
      const bytes = new Uint8Array(data)
      messageHandlers.forEach((fn) => fn(bytes))
      return
    }
    // 二进制帧：Blob（部分浏览器实现）
    if (data instanceof Blob) {
      data.arrayBuffer().then((buf) => {
        const bytes = new Uint8Array(buf)
        messageHandlers.forEach((fn) => fn(bytes))
      })
    }
  }

  const attach = (socket: WebSocket) => {
    if (attached && currentSocket === socket) return
    detach()
    currentSocket = socket
    socket.addEventListener("message", handleMessage)
    attached = true
  }

  const detach = () => {
    if (currentSocket && attached) {
      currentSocket.removeEventListener("message", handleMessage)
    }
    attached = false
    currentSocket = null
  }

  // 轮询 socket 变化（ws-client 内部重连会替换 socket 实例）
  const pollInterval = setInterval(() => {
    const socket = opts.getSocket()
    if (socket && socket.readyState === WebSocket.OPEN) {
      attach(socket)
      statusHandlers.forEach((fn) => fn(true))
    } else if (!socket || socket.readyState === WebSocket.CLOSED) {
      detach()
      statusHandlers.forEach((fn) => fn(false))
    }
  }, 500)

  // 监听连接事件以立即响应
  opts.onOpen(() => {
    const socket = opts.getSocket()
    if (socket) {
      attach(socket)
      statusHandlers.forEach((fn) => fn(true))
    }
  })
  opts.onClose(() => {
    detach()
    statusHandlers.forEach((fn) => fn(false))
  })

  return {
    send(data: Uint8Array) {
      const socket = currentSocket ?? opts.getSocket()
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(data)
      }
    },
    onMessage(cb) {
      messageHandlers.add(cb)
      return () => { messageHandlers.delete(cb) }
    },
    onStatus(cb) {
      statusHandlers.add(cb)
      return () => { statusHandlers.delete(cb) }
    },
    /** 清理（内部使用） */
    destroy() {
      clearInterval(pollInterval)
      detach()
      messageHandlers.clear()
      statusHandlers.clear()
    },
  } as CollabTransport & { destroy: () => void }
}
