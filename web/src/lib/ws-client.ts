/**
 * WebSocket 客户端 — 自动重连、心跳、事件订阅、断线补偿
 *
 * 使用方式：
 *   import { wsClient } from '@/lib/ws-client'
 *   wsClient.connect(workspaceId)
 *   wsClient.on('issue.updated', (data) => { ... })
 *   wsClient.off('issue.updated', handler)
 *   wsClient.onReconnect(() => { fetchLatest() })
 *   wsClient.disconnect()
 */
type WSMessage = {
  type: string
  data: any
}

type Listener = (data: any) => void

class WSClient {
  private ws: WebSocket | null = null
  private url: string = ''
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private listeners: Map<string, Set<Listener>> = new Map()
  private reconnectCallbacks: Set<() => void> = new Set()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private intentionalClose = false
  private workspaceId: number | null = null
  private userId: number | undefined
  /** 上次断开时的时间戳，用于断线补偿 since 参数 */
  get lastDisconnectTimestamp(): number {
    return this.lastDisconnectTs
  }

  /** 注册断线重连补偿回调（在每次重连成功后调用） */
  onReconnect(callback: () => void) {
    this.reconnectCallbacks.add(callback)
  }

  /** 取消断线重连补偿回调 */
  offReconnect(callback: () => void) {
    this.reconnectCallbacks.delete(callback)
  }

  /** 当前连接的 workspaceId（未连接时为 null） */
  get currentWorkspaceId(): number | null {
    return this.workspaceId
  }

  /** 当前连接的用户 ID（未连接时为 undefined） */
  get currentUserId(): number | undefined {
    return this.userId
  }

  /** 建立 WebSocket 连接 */
  connect(workspaceId: number, userId?: number) {
    this.intentionalClose = false
    this.workspaceId = workspaceId
    this.userId = userId

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = location.host
    this.url = `${protocol}//${host}/ws/${workspaceId}`

    this.createSocket()
  }

  private createSocket() {
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
    }

    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      this.startHeartbeat()
      console.log('[WS] connected')

      // 断线重连补偿：通知所有已注册的回调进行数据补齐
      if (this.lastDisconnectTs > 0) {
        console.log(`[WS] reconnect detected, triggering compensation (since ${new Date(this.lastDisconnectTs).toISOString()})`)
        this.reconnectCallbacks.forEach(cb => {
          try { cb() } catch (e) { console.error('[WS] reconnect callback error:', e) }
        })
      }
    }

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        if (msg.type === 'pong') return
        this.emit(msg.type, msg.data)
      } catch {
        // 忽略解析失败的消息
      }
    }

    this.ws.onclose = () => {
      this.stopHeartbeat()
      if (!this.intentionalClose) {
        this.lastDisconnectTs = Date.now()
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      // onclose 会紧接着触发，由 onclose 处理重连
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('[WS] max reconnect attempts reached')
      return
    }
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)
    this.reconnectAttempts++
    console.log(`[WS] reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
    this.reconnectTimer = setTimeout(() => this.createSocket(), delay)
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, 25000)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  /** 断开连接 */
  disconnect() {
    this.intentionalClose = true
    this.stopHeartbeat()
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.listeners.clear()
    this.reconnectCallbacks.clear()
    this.lastDisconnectTs = 0
  }

  /** 订阅事件 */
  on(eventType: string, listener: Listener) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set())
    }
    this.listeners.get(eventType)!.add(listener)
  }

  /** 取消订阅 */
  off(eventType: string, listener: Listener) {
    this.listeners.get(eventType)?.delete(listener)
  }

  /** 触发事件 */
  private emit(eventType: string, data: any) {
    this.listeners.get(eventType)?.forEach(fn => {
      try {
        fn(data)
      } catch (e) {
        console.error(`[WS] listener error for ${eventType}:`, e)
      }
    })
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

export const wsClient = new WSClient()
