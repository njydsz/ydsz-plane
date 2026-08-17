/**
 * 附件域 API — 对接后端 Attachment 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 附件模型 */
export interface Attachment {
  id: number;
  workspace_id: number;
  project_id: number;
  entity_type: string;
  entity_id: number;
  file_name: string;
  file_size: number;
  content_type: string;
  storage_key: string;
  storage_url?: string;
  uploaded_by: number;
  created_at: string;
  updated_at: string;
}

/** 预签名上传入参 */
export interface PresignedUploadInput {
  file_name: string;
  content_type: string;
  entity_type: string;
  entity_id: number;
}

/** 预签名上传响应（presign 阶段不含 attachment，需 confirm 后才有） */
export interface PresignedUploadResult {
  upload_url: string;
  storage_key: string;
}

/** 上传确认入参 */
export interface ConfirmUploadInput {
  file_name: string;
  content_type: string;
  file_size: number;
  entity_type: string;
  entity_id: number;
  storage_key: string;
}

/** 上传确认响应 */
export interface ConfirmUploadResult {
  attachment: Attachment;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 附件域 API — 附件列表、预签名上传、分片上传、下载链接生成。 */
export const attachmentApi = {
  /** 获取实体的附件列表 */
  listAttachments: (wsId: number, projectId: number, entityType: string, entityId: number) =>
    wrap<{ results: Attachment[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/attachments`, {
        params: { entity_type: entityType, entity_id: entityId },
      }),
    ),

  /** 获取预签名上传 URL */
  getPresignedUploadURL: (wsId: number, projectId: number, input: PresignedUploadInput) =>
    wrap<PresignedUploadResult>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/attachments/presigned-upload`, input),
    ),

  /** 上传确认：PUT 成功后调用，写入 DB 附件记录 */
  confirmUpload: (wsId: number, projectId: number, input: ConfirmUploadInput) =>
    wrap<ConfirmUploadResult>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/attachments/confirm`, input),
    ),

  /** 删除附件 */
  deleteAttachment: (wsId: number, projectId: number, attachmentId: number, entityType?: string) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/attachments/${attachmentId}`, {
      params: entityType ? { entity_type: entityType } : undefined,
    })),
};
