/**
 * provider.ts — CRDT 协同传输层（Y.js over WebSocket）。
 *
 * 协议：y-websocket 兼容（messageSync = 0 / messageAwareness = 1）。
 *
 * 帧格式（Uint8Array）：
 *   [0]     消息类型（0=sync, 1=awareness, 2=auth）
 *   [1..]   载荷
 *
 * 与现有 ws-client 集成：复用其连接与重连，二进制帧经 transport 分发。
 */
import * as Y from "yjs"
import * as encoding from "lib0/encoding"
import * as decoding from "lib0/decoding"
import { Awareness, applyAwarenessUpdate } from "y-protocols/awareness"

/** 消息类型 */
export const MESSAGE_SYNC = 0
export const MESSAGE_AWARENESS = 1
export const MESSAGE_AUTH = 2

/** Sync 子类型 */
export const SYNC_STEP1 = 0
export const SYNC_STEP2 = 1
export const SYNC_UPDATE = 2

/** 传输回调 */
export interface CollabTransport {
  send(data: Uint8Array): void
  onMessage(cb: (data: Uint8Array) => void): () => void
  onStatus(cb: (connected: boolean) => void): () => void
}

/**
 * YjsProvider — 管理 Y.Doc 与 WebSocket 的双向同步。
 *
 * 职责：
 *  - 本地变更 → 编码 → 发送（增量 update）
 *  - 接收远端更新 → 应用到 Y.Doc（origin 标记避免回环）
 *  - 首次连接 → 交换 state vector 完成同步
 *  - awareness 广播与接收
 */
export class YjsProvider {
  readonly doc: Y.Doc
  readonly awareness: Awareness
  private readonly transport: CollabTransport

  private connected = false
  private synced = false

  onSync?: (synced: boolean) => void
  onConnected?: (connected: boolean) => void

  private unsubscribeDoc: (() => void) | null = null
  private unsubscribeAwareness: (() => void) | null = null
  private detachMessage: (() => void) | null = null
  private detachStatus: (() => void) | null = null

  constructor(docID: string, transport: CollabTransport) {
    void docID
    this.transport = transport
    this.doc = new Y.Doc({ guid: docID })
    this.awareness = new Awareness(this.doc)

    this.setupDocSync()
    this.setupAwarenessSync()
    this.detachMessage = transport.onMessage((data) => this.handleMessage(data))
    this.detachStatus = transport.onStatus((connected) => {
      this.connected = connected
      if (connected) {
        this.requestSync()
      } else {
        this.synced = false
        this.onSync?.(false)
      }
      this.onConnected?.(connected)
    })
  }

  isSynced(): boolean { return this.synced }
  isConnected(): boolean { return this.connected }

  // ------------------------------------------------------------------
  // 本地 → 远端
  // ------------------------------------------------------------------

  private setupDocSync() {
    const onUpdate = (update: Uint8Array, origin: unknown) => {
      if (origin === "remote") return
      this.sendUpdate(update)
    }
    this.doc.on("update", onUpdate)
    this.unsubscribeDoc = () => this.doc.off("update", onUpdate)
  }

  private sendUpdate(update: Uint8Array) {
    if (!this.connected) return
    const encoder = encoding.createEncoder()
    encoding.writeVarUint(encoder, MESSAGE_SYNC)
    encoding.writeVarUint(encoder, SYNC_UPDATE)
    encoding.writeVarUint8Array(encoder, update)
    this.transport.send(encoding.toUint8Array(encoder))
  }

  private requestSync() {
    if (!this.connected) return
    const encoder = encoding.createEncoder()
    encoding.writeVarUint(encoder, MESSAGE_SYNC)
    encoding.writeVarUint(encoder, SYNC_STEP1)
    encoding.writeVarUint8Array(encoder, Y.encodeStateVector(this.doc))
    this.transport.send(encoding.toUint8Array(encoder))
  }

  private setupAwarenessSync() {
    const onAwareness = (
      { added, updated, removed }: { added: number[]; updated: number[]; removed: number[] },
      origin: unknown,
    ) => {
      if (origin === "remote") return
      this.sendAwareness(added.concat(updated, removed))
    }
    this.awareness.on("update", onAwareness)
    this.unsubscribeAwareness = () => this.awareness.off("update", onAwareness)
  }

  private sendAwareness(clients: number[]) {
    if (!this.connected) return
    const encoder = encoding.createEncoder()
    encoding.writeVarUint(encoder, MESSAGE_AWARENESS)
    encoding.writeVarUint8Array(encoder, encodeAwareness(this.awareness, clients))
    this.transport.send(encoding.toUint8Array(encoder))
  }

  // ------------------------------------------------------------------
  // 远端 → 本地
  // ------------------------------------------------------------------

  private handleMessage(data: Uint8Array) {
    const decoder = decoding.createDecoder(data)
    const messageType = decoding.readVarUint(decoder)

    switch (messageType) {
      case MESSAGE_SYNC:
        this.handleSync(decoder)
        break
      case MESSAGE_AWARENESS:
        this.handleAwareness(decoder)
        break
      default:
        break
    }
  }

  private handleSync(decoder: decoding.Decoder) {
    const subtype = decoding.readVarUint(decoder)
    switch (subtype) {
      case SYNC_STEP1: {
        const remoteSV = decoding.readVarUint8Array(decoder)
        const diff = Y.encodeStateAsUpdate(this.doc, remoteSV)
        this.sendUpdate(diff)
        break
      }
      case SYNC_STEP2:
      case SYNC_UPDATE: {
        const update = decoding.readVarUint8Array(decoder)
        Y.applyUpdate(this.doc, update, "remote")
        if (!this.synced) {
          this.synced = true
          this.onSync?.(true)
        }
        break
      }
    }
  }

  private handleAwareness(decoder: decoding.Decoder) {
    const update = decoding.readVarUint8Array(decoder)
    applyAwarenessUpdate(this.awareness, update, "remote")
  }

  // ------------------------------------------------------------------
  // 生命周期
  // ------------------------------------------------------------------

  destroy() {
    this.unsubscribeDoc?.()
    this.unsubscribeAwareness?.()
    this.detachMessage?.()
    this.detachStatus?.()
    this.awareness.destroy()
    this.doc.destroy()
    this.synced = false
    this.connected = false
  }

  /** 获取文档二进制快照（用于持久化） */
  encodeState(): Uint8Array {
    return Y.encodeStateAsUpdate(this.doc)
  }

  /** 从快照恢复 */
  applyState(state: Uint8Array) {
    Y.applyUpdate(this.doc, state, "remote")
  }
}

// ----------------------------------------------------------------------
// awareness 编解码
// ----------------------------------------------------------------------

function encodeAwareness(awareness: Awareness, clients: number[]): Uint8Array {
  const encoder = encoding.createEncoder()
  encoding.writeVarUint(encoder, clients.length)
  for (const clientID of clients) {
    const state = awareness.getStates().get(clientID)
    const clock = awareness.meta.get(clientID)?.clock ?? 0
    encoding.writeVarUint(encoder, clientID)
    encoding.writeVarUint(encoder, clock)
    encoding.writeVarString(encoder, state ? JSON.stringify(state) : "{}")
  }
  return encoding.toUint8Array(encoder)
}

export { Y }
