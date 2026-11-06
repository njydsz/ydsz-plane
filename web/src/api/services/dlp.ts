/**
 * DLP（数据防泄漏）API。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type DlpCategory =
  | "id_card" | "bank_card" | "phone" | "email"
  | "password" | "api_key" | "address" | "passport" | "license" | "custom";

export type DlpSeverity = "critical" | "high" | "medium" | "low";
export type DlpAction = "flagged" | "blocked" | "resolved";
export type DlpScope = "project" | "page" | "issue" | "global";

/** 扫描发现 */
export interface DlpFinding {
  id: number;
  job_id: number;
  workspace_id: number;
  scope: DlpScope;
  target_id: number;
  target_title: string;
  category: DlpCategory;
  severity: DlpSeverity;
  rule_name: string;
  masked_value: string;
  offset_pos: number;
  confidence: number;
  action: DlpAction;
  actioned_by?: number;
  resolved_at?: string;
  created_at: string;
}

/** 扫描任务 */
export interface DlpJob {
  id: number;
  workspace_id: number;
  scope: DlpScope;
  target_id?: number;
  status: "pending" | "running" | "completed" | "failed";
  total_items: number;
  scanned_items: number;
  findings: number;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  created_by: number;
  created_at: string;
}

/** 扫描报告 */
export interface DlpReport {
  job_id: number;
  workspace_id: number;
  total_items: number;
  findings: number;
  by_category: Record<string, number>;
  by_severity: Record<string, number>;
  completed_at: string;
}

/** 检测规则 */
export interface DlpRule {
  name: string;
  category: DlpCategory;
  severity: DlpSeverity;
  confidence: number;
  enabled: boolean;
}

/** 检查结果 */
export interface DlpCheckResult {
  action: "allow" | "warn" | "block";
  matches: Array<{
    category: DlpCategory;
    severity: DlpSeverity;
    value: string;
    raw_length: number;
    offset: number;
    rule_name: string;
    confidence: number;
  }>;
  count: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

export const dlpApi = {
  listFindings: (
    workspaceId: number | string,
    params?: { category?: string; severity?: string; action?: string; limit?: number },
  ) =>
    wrap<{ items: DlpFinding[]; count: number }>(
      http.get(`/workspaces/${workspaceId}/dlp/findings`, { params }),
    ),

  listJobs: (workspaceId: number | string) =>
    wrap<{ items: DlpJob[] }>(http.get(`/workspaces/${workspaceId}/dlp/jobs`)),

  getJob: (workspaceId: number | string, jobId: number | string) =>
    wrap<DlpJob>(http.get(`/workspaces/${workspaceId}/dlp/jobs/${jobId}`)),

  getReport: (workspaceId: number | string, jobId: number | string) =>
    wrap<DlpReport>(http.get(`/workspaces/${workspaceId}/dlp/jobs/${jobId}/report`)),

  createJob: (workspaceId: number | string, scope: DlpScope, targetId?: number) =>
    wrap<DlpJob>(http.post(`/workspaces/${workspaceId}/dlp/jobs`, { scope, target_id: targetId })),

  runJob: (workspaceId: number | string, jobId: number | string) =>
    wrap<DlpJob>(http.post(`/workspaces/${workspaceId}/dlp/jobs/${jobId}/run`)),

  resolveFinding: (workspaceId: number | string, id: number | string) =>
    wrap<{ ok: boolean }>(http.post(`/workspaces/${workspaceId}/dlp/findings/${id}/resolve`)),

  /** 写入前检查 */
  check: (workspaceId: number | string, text: string) =>
    wrap<DlpCheckResult>(http.post(`/workspaces/${workspaceId}/dlp/check`, { text })),

  /** 规则列表 */
  listRules: (workspaceId: number | string) =>
    wrap<{ items: DlpRule[] }>(http.get(`/workspaces/${workspaceId}/dlp/rules`)),
};
