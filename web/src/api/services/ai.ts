/**
 * AI 域 API — 对接后端 AI 智能辅助 REST 接口。
 *
 * 对应后端：internal/application/ai
 * 路由前缀：/api/v1/workspaces/:wsId/projects/:projectId/ai
 */

import { apiClient } from "../client";

/* ------------------------------------------------------------------ */
/* Types (mirror backend models)                                      */
/* ------------------------------------------------------------------ */

/** AI 服务当前状态 */
export interface AIStatus {
  enabled: boolean;
  provider: string;
}

/** 智能指派候选人 */
export interface AssignCandidate {
  user_id: number;
  display_name: string;
  score: number;
  reason: string;
}

/** 智能指派输入 */
export interface SmartAssignInput {
  issue_title: string;
  issue_description?: string;
  type_code?: string;
  top_n?: number;
}

/** 疑似重复工作项 */
export interface DuplicateCandidate {
  issue_id: number;
  identifier: string;
  title: string;
  similarity: number;
}

/** 智能分类结果 */
export interface ClassifyResult {
  type_code: "requirement" | "task" | "defect";
  priority: "critical" | "high" | "medium" | "low";
  confidence: number;
}

/** 摘要结果 */
export interface SummarizeResult {
  summary: string;
  key_points: string[];
  word_count: number;
}

/** 摘要输入 */
export interface SummarizeInput {
  content_type: "issue" | "sprint" | "version";
  title: string;
  content: string;
  max_words?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** AI 域 API — 智能指派、重复检测、分类、摘要。 */
export const aiApi = {
  /** 获取 AI 功能状态 */
  getStatus: (wsId: number | string, projectId: number | string) =>
    wrap<AIStatus>(apiClient.get(`/workspaces/${wsId}/projects/${projectId}/ai/status`)),

  /** 智能推荐指派人 */
  smartAssign: (wsId: number | string, projectId: number | string, input: SmartAssignInput) =>
    wrap<AssignCandidate[]>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/ai/smart-assign`, input),
    ),

  /** 检测重复工作项 */
  detectDuplicates: (wsId: number | string, projectId: number | string, title: string, description?: string) =>
    wrap<DuplicateCandidate[]>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/ai/detect-duplicates`, {
        title,
        description: description || "",
      }),
    ),

  /** 智能分类工作项（自动推荐类型和优先级） */
  smartClassify: (wsId: number | string, projectId: number | string, title: string, description?: string) =>
    wrap<ClassifyResult>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/ai/classify`, {
        title,
        description: description || "",
      }),
    ),

  /** 生成文字摘要 */
  summarize: (wsId: number | string, projectId: number | string, input: SummarizeInput) =>
    wrap<SummarizeResult>(
      apiClient.post(`/workspaces/${wsId}/projects/${projectId}/ai/summarize`, input),
    ),
};
