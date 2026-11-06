/**
 * 实时协作 API — 编辑会话 + 乐观锁 + 操作广播。
 *
 * 路由: /api/v1/workspaces/:ws/collaboration/*
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 文档类型 */
export type CollaborationDocumentType =
  | "project_page"
  | "knowledge_page"
  | "issue_description"
  | "sprint_goal";

/** 会话状态 */
export type SessionStatus = "active" | "idle" | "closed";

/** 协作会话（在线编辑者） */
export interface CollaborationSession {
  id: number;
  workspace_id: number;
  document_type: CollaborationDocumentType;
  document_id: number;
  user_id: number;
  user_name: string;
  display_name: string;
  status: SessionStatus;
  base_version: number;
  last_seen_at: string;
  created_at: string;
}

/** 协作操作 */
export interface CollaborationOperation {
  id: number;
  workspace_id: number;
  document_type: CollaborationDocumentType;
  document_id: number;
  user_id: number;
  user_name: string;
  op_type: "join" | "leave" | "edit" | "lock" | "unlock" | "cursor" | string;
  base_version: number;
  target_version: number;
  payload?: string;
  created_at: string;
}

/** 操作提交结果 */
export interface SubmitOperationResult {
  accepted: boolean;
  new_version: number;
  reason?: string;
  operation?: CollaborationOperation;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 实时协作 API */
export const collaborationApi = {
  /** 加入文档编辑会话 */
  joinSession: (
    workspaceId: number | string,
    documentType: CollaborationDocumentType,
    documentId: number | string,
  ) =>
    wrap<CollaborationSession>(
      http.post(`/workspaces/${workspaceId}/collaboration/sessions`, {
        document_type: documentType,
        document_id: documentId,
      }),
    ),

  /** 离开文档编辑会话 */
  leaveSession: (
    workspaceId: number | string,
    documentType: CollaborationDocumentType,
    documentId: number | string,
  ) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/collaboration/sessions`, {
        params: { document_type: documentType, document_id: documentId },
      }),
    ),

  /** 心跳 */
  heartbeat: (
    workspaceId: number | string,
    documentType: CollaborationDocumentType,
    documentId: number | string,
    baseVersion: number,
  ) =>
    wrap<void>(
      http.post(`/workspaces/${workspaceId}/collaboration/heartbeat`, {
        document_type: documentType,
        document_id: documentId,
        base_version: baseVersion,
      }),
    ),

  /** 在线编辑者列表 */
  listSessions: (
    workspaceId: number | string,
    documentType: CollaborationDocumentType,
    documentId: number | string,
  ) =>
    wrap<{ items: CollaborationSession[] }>(
      http.get(`/workspaces/${workspaceId}/collaboration/sessions`, {
        params: { document_type: documentType, document_id: documentId },
      }),
    ),

  /** 文档当前版本 */
  getVersion: (
    workspaceId: number | string,
    documentType: CollaborationDocumentType,
    documentId: number | string,
  ) =>
    wrap<{ version: number }>(
      http.get(`/workspaces/${workspaceId}/collaboration/version`, {
        params: { document_type: documentType, document_id: documentId },
      }),
    ),

  /** 提交协作操作 */
  submitOperation: (
    workspaceId: number | string,
    data: {
      document_type: CollaborationDocumentType;
      document_id: number | string;
      op_type: string;
      base_version: number;
      payload?: string;
    },
  ) =>
    wrap<SubmitOperationResult>(
      http.post(`/workspaces/${workspaceId}/collaboration/operations`, data),
    ),
};
