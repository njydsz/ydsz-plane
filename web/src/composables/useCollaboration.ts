/**
 * useCollaboration — 文档实时协作逻辑层。
 *
 * 能力：
 *  - 加入/离开编辑会话（在线编辑者可见）；
 *  - 心跳保活（30s 间隔）；
 *  - 乐观锁版本跟踪（编辑提交时校验 base_version）；
 *  - 操作广播（通过 WebSocket 订阅同文档事件）；
 *  - 冲突提示（版本不一致时通知调用方刷新）。
 *
 * 用法：
 *   const collab = useCollaboration({
 *     workspaceId: computed(() => ws.id),
 *     documentType: 'project_page',
 *     documentId: computed(() => pageId.value),
 *   })
 *   collab.join()
 *   // 保存时：
 *   const res = await collab.submitEdit(content)
 *   if (!res.accepted) { console.warn("冲突") }
 */
import { computed, onBeforeUnmount, ref, watch, type ComputedRef, type Ref } from "vue";

import {
  collaborationApi,
  type CollaborationDocumentType,
  type CollaborationOperation,
  type CollaborationSession,
} from "@/api/services/collaboration";
import { useAuthStore } from "@/stores/auth";
import { wsClient } from "@/lib/ws-client";

/** 远程协作者光标 */
export interface RemoteCursor {
  userId: number
  name: string
  color: string
  /** 文档内相对位置（0-1 比例，便于跨分辨率对齐） */
  anchorRatio: number
  headRatio: number
  updatedAt: number
}

/** 协作配置 */
export interface UseCollaborationOptions {
  workspaceId: ComputedRef<number> | Ref<number>;
  documentType: CollaborationDocumentType;
  documentId: ComputedRef<number> | Ref<number>;
  /** 心跳间隔（毫秒），默认 30s */
  heartbeatInterval?: number;
  /** 操作广播回调（用于同步其他协作者的变更） */
  onOperation?: (op: CollaborationOperation) => void;
}

/** 心跳间隔 */
const HEARTBEAT_INTERVAL = 30_000;

export function useCollaboration(options: UseCollaborationOptions) {
  const auth = useAuthStore();

  const workspaceId = computed(() => Number(options.workspaceId.value) || 0);
  const documentId = computed(() => Number(options.documentId.value) || 0);
  const documentType = options.documentType;
  const heartbeatInterval = options.heartbeatInterval ?? HEARTBEAT_INTERVAL;

  /* ---- 状态 ---- */

  const sessions = ref<CollaborationSession[]>([]);
  const version = ref(0);
  const joined = ref(false);
  const loading = ref(false);
  const error = ref("");

  /** 远程协作者光标（不含自己） */
  const remoteCursors = ref<RemoteCursor[]>([]);
  /** 本地光标节流计时器 */
  let cursorTimer: ReturnType<typeof setTimeout> | null = null;
  let unsubscribeCursor: (() => void) | null = null;

  /** 当前用户 ID */
  const currentUserId = computed(() => auth.user?.id ?? 0);
  /** 除自己外的在线编辑者 */
  const otherEditors = computed(() =>
    sessions.value.filter((s) => s.user_id !== currentUserId.value),
  );
  /** 所有在线编辑者 */
  const allEditors = computed(() => sessions.value);

  /* ---- 定时器 ---- */

  let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  let refreshTimer: ReturnType<typeof setInterval> | null = null;

  /* ---- 会话 ---- */

  async function join() {
    if (!workspaceId.value || !documentId.value || joined.value) return;
    loading.value = true;
    error.value = "";
    try {
      await collaborationApi.joinSession(workspaceId.value, documentType, documentId.value);
      joined.value = true;
      await refresh();
      startHeartbeat();
      startRefresh();
      subscribeCursors();
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加入协作会话失败";
    } finally {
      loading.value = false;
    }
  }

  async function leave() {
    stopHeartbeat();
    stopRefresh();
    if (!workspaceId.value || !documentId.value || !joined.value) {
      joined.value = false;
      return;
    }
    try {
      await collaborationApi.leaveSession(workspaceId.value, documentType, documentId.value);
    } catch {
      /* 忽略离开失败 */
    } finally {
      joined.value = false;
      sessions.value = [];
    }
  }

  async function refresh() {
    if (!workspaceId.value || !documentId.value) return;
    try {
      const [sessRes, verRes] = await Promise.all([
        collaborationApi.listSessions(workspaceId.value, documentType, documentId.value),
        collaborationApi.getVersion(workspaceId.value, documentType, documentId.value),
      ]);
      sessions.value = (sessRes as { items?: CollaborationSession[] }).items ?? [];
      version.value = (verRes as { version: number }).version ?? 0;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "刷新协作状态失败";
    }
  }

  /* ---- 心跳 ---- */

  function startHeartbeat() {
    stopHeartbeat();
    heartbeatTimer = setInterval(async () => {
      if (!joined.value) return;
      try {
        await collaborationApi.heartbeat(
          workspaceId.value,
          documentType,
          documentId.value,
          version.value,
        );
      } catch {
        /* 心跳失败不阻塞编辑 */
      }
    }, heartbeatInterval);
  }

  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer);
      heartbeatTimer = null;
    }
  }

  function startRefresh() {
    stopRefresh();
    refreshTimer = setInterval(refresh, 15_000);
  }

  function stopRefresh() {
    if (refreshTimer) {
      clearInterval(refreshTimer);
      refreshTimer = null;
    }
  }

  /* ---- 操作提交 ---- */

  /**
   * 提交编辑操作（带乐观锁校验）。
   * @returns 提交结果；accepted=false 表示版本冲突
   */
  async function submitEdit(payload?: string) {
    if (!workspaceId.value || !documentId.value) {
      return { accepted: false, new_version: 0, reason: "缺少工作空间或文档 ID" };
    }
    try {
      const res = await collaborationApi.submitOperation(workspaceId.value, {
        document_type: documentType,
        document_id: documentId.value,
        op_type: "edit",
        base_version: version.value,
        payload,
      });
      if (res.accepted) {
        version.value = res.new_version;
        options.onOperation?.(res.operation as CollaborationOperation);
      } else {
        // 冲突：拉取最新版本
        await refresh();
      }
      return res;
    } catch (e: unknown) {
      return {
        accepted: false,
        new_version: version.value,
        reason: e instanceof Error ? e.message : "提交失败",
      };
    }
  }

  /**
   * 提交非编辑操作（join/leave/cursor/lock 等），不校验版本。
   */
  async function submitOperation(opType: string, payload?: string) {
    if (!workspaceId.value || !documentId.value) return null;
    try {
      const res = await collaborationApi.submitOperation(workspaceId.value, {
        document_type: documentType,
        document_id: documentId.value,
        op_type: opType,
        base_version: version.value,
        payload,
      });
      if (res.operation) {
        options.onOperation?.(res.operation);
      }
      if (res.new_version) {
        version.value = res.new_version;
      }
      return res;
    } catch {
      return null;
    }
  }

  /* ---- 光标广播与接收 ---- */

  /**
   * 广播本地光标位置（节流 50ms，避免高频请求）。
   * anchorRatio/headRatio 为 0-1 相对文档高度的比例。
   */
  function broadcastCursor(anchorRatio: number, headRatio: number) {
    if (!workspaceId.value || !documentId.value) return;
    if (cursorTimer) return;
    cursorTimer = setTimeout(() => { cursorTimer = null; }, 50);
    void collaborationApi
      .submitOperation(workspaceId.value, {
        document_type: documentType,
        document_id: documentId.value,
        op_type: "cursor",
        base_version: version.value,
        payload: JSON.stringify({ anchorRatio, headRatio }),
      })
      .catch(() => { /* 光标失败不影响编辑 */ });
  }

  /** 订阅远程光标（通过 WebSocket 实时推送） */
  function subscribeCursors() {
    if (unsubscribeCursor) return;
    const handler = (data: { userId: number; name?: string; anchorRatio: number; headRatio: number }) => {
      if (!data || data.userId === currentUserId.value) return;
      const list = remoteCursors.value.filter((c) => c.userId !== data.userId);
      list.push({
        userId: data.userId,
        name: data.name || `用户 #${data.userId}`,
        color: colorOf(data.userId),
        anchorRatio: data.anchorRatio,
        headRatio: data.headRatio,
        updatedAt: Date.now(),
      });
      remoteCursors.value = list;
    };
    wsClient.on("collab.cursor", handler);
    unsubscribeCursor = () => wsClient.off("collab.cursor", handler);
  }

  /** 取消订阅 */
  function unsubscribeRemoteCursors() {
    unsubscribeCursor?.();
    unsubscribeCursor = null;
  }

  /** 清理过期光标（超过 10s 未更新视为离线） */
  function pruneCursors() {
    const now = Date.now();
    remoteCursors.value = remoteCursors.value.filter((c) => now - c.updatedAt < 10_000);
  }

  /** 稳定的协作者颜色 */
  function colorOf(userId: number): string {
    const palette = ["#3b82f6", "#10b981", "#f59e0b", "#8b5cf6", "#ec4899", "#14b8a6", "#f97316"];
    return palette[Math.abs(userId) % palette.length];
  }

  /* ---- 监听文档切换 ---- */

  watch(documentId, (next, prev) => {
    if (next !== prev && joined.value) {
      // 切换文档：先离开旧会话，再加入新会话
      void leave();
      void join();
    }
  });

  /* ---- 清理 ---- */

  onBeforeUnmount(() => {
    unsubscribeRemoteCursors();
    void leave();
  });

  return {
    // 状态
    sessions,
    allEditors,
    otherEditors,
    version,
    joined,
    loading,
    error,
    currentUserId,
    remoteCursors,
    // 操作
    join,
    broadcastCursor,
    pruneCursors,
    leave,
    refresh,
    submitEdit,
    submitOperation,
  };
}
