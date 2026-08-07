/**
 * 审计日志 API — 对接 /:ws/audit-logs 端点。
 */
import { apiClient } from '../client';

/** 审计日志条目 */
export interface AuditLogEntry {
  id: number;
  actor_id: number;
  actor_name?: string;
  action: string;
  target?: string;
  detail?: Record<string, unknown>;
  ip?: string;
  created_at: string;
}

/** 审计日志汇总统计 */
export interface AuditLogSummary {
  total: number;
  by_action: Record<string, number>;
  by_actor: { actor_id: number; actor_name: string; count: number }[];
  recent_24h: number;
}

/** 审计日志 API — 对接 /:ws/audit-logs 端点，读取操作审计记录与汇总统计。 */
export const auditApi = {
  /** 获取审计日志列表 */
  async list(wsId: number, limit = 50): Promise<AuditLogEntry[]> {
    const { data } = await apiClient.get<AuditLogEntry[]>(
      `/api/v1/workspaces/${wsId}/audit-logs`,
      { params: { limit } },
    );
    return data;
  },
};
