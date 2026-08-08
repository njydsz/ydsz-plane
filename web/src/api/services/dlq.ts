/**
 * DLQ（死信队列）监控 API —— 封装 DLQ 消息查询、重试、清理的 REST 调用。
 *
 * 覆盖两类资源：
 *  - GET  /workspaces/:wsId/admin/dlq             — 死信列表（分页）
 *  - POST /workspaces/:wsId/admin/dlq/:id/retry  — 重试指定死信事件
 *  - DELETE /workspaces/:wsId/admin/dlq/:id      — 清理指定死信
 *  - POST /workspaces/:wsId/admin/dlq/cleanup    — 批量清理（按 ID 或全部已解决）
 */
import { apiClient } from "../client";

/** 死信消息实体 */
export interface DLQItem {
  id: number;
  event_id: number;
  queue: string;
  exchange: string;
  routing_key: string;
  payload: Record<string, any>;
  error_reason: string;
  resolved_at: string | null;
  created_at: string;
}

/** DLQ 列表返回 */
export interface DLQListResult {
  items: DLQItem[];
  total: number;
}

/** DLQ 列表查询参数 */
export interface DLQListParams {
  /** 分页偏移 */
  offset?: number;
  /** 分页大小 */
  limit?: number;
  /** 仅显示未解决 */
  unresolved_only?: boolean;
}

/** DLQ 域 API */
export const dlqApi = {
  /** 查询死信消息列表 */
  async list(wsId: number, params?: DLQListParams): Promise<DLQListResult> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/admin/dlq`, { params });
    return data;
  },

  /** 重试指定死信事件 */
  async retry(wsId: number, id: number): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/admin/dlq/${id}/retry`);
  },

  /** 清理（标记 resolved）指定死信 */
  async remove(wsId: number, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/admin/dlq/${id}`);
  },

  /** 批量清理死信 */
  async cleanup(wsId: number, body: { event_ids: number[] } | { resolved_all: true }): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/admin/dlq/cleanup`, body);
  },
};
