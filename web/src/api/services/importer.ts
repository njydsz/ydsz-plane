/**
 * Importer 外部数据导入 API。
 *
 * 路由: /api/v1/workspaces/:wsId/projects/:projectId/imports/*
 * 权限：读 project:read；写 project:update。
 *
 * 来源：
 *   - csv   通用 CSV（multipart 上传）
 *   - jira  Atlassian Jira Cloud / Data Center
 *   - linear Linear GraphQL API
 *   - asana  Asana REST API
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 导入来源（与后端 importer.Source 对齐）。 */
export type ImportSource = "csv" | "jira" | "linear" | "asana";

/** 任务状态。 */
export type ImportJobStatus = "pending" | "running" | "completed" | "failed";

/** 单条导入错误。 */
export interface ImportError {
  external_id?: string;
  title?: string;
  reason: string;
}

/** 导入结果统计。 */
export interface ImportResult {
  source: ImportSource | string;
  total_records: number;
  imported: number;
  skipped: number;
  failed: number;
  errors?: ImportError[];
  started_at: string;
  completed_at?: string;
}

/** 导入任务实体。 */
export interface ImportJob {
  id: number;
  workspace_id: number;
  project_id: number;
  source: ImportSource | string;
  status: ImportJobStatus | string;
  result?: ImportResult | null;
  created_by: number;
  created_at: string;
  completed_at?: string | null;
}

/** 支持的来源元数据。 */
export interface SourceInfo {
  id: ImportSource | string;
  name: string;
  description: string;
  needs_credentials: boolean;
  /** 该平台需要填写的凭证字段（如 base_url / email / api_token / project_key / team_id） */
  fields?: string[];
}

/** 外部平台凭证（字段按来源裁剪）。 */
export interface ImportCredentials {
  api_token?: string;
  email?: string;
  base_url?: string;
  project_key?: string;
  team_id?: string;
}

/** 平台导入选项。 */
export interface ImportOptionsInput {
  column_mapping?: Record<string, string>;
  default_type?: string;
  import_children?: boolean;
  batch_size?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 判断来源是否为受支持的平台来源（非 CSV）。 */
export function isPlatformSource(source: ImportSource | string): boolean {
  return source !== "csv";
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** Importer API。 */
export const importerApi = {
  /** 支持的导入来源列表。 */
  listSources: (wsId: number | string, projectId: number | string) =>
    wrap<{ sources: SourceInfo[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/imports/sources`),
    ).then((r) => r.sources),

  /** 当前项目的导入任务列表。 */
  listJobs: (wsId: number | string, projectId: number | string) =>
    wrap<{ items: ImportJob[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/imports/jobs`),
    ).then((r) => r.items),

  /** 创建导入任务。 */
  createJob: (wsId: number | string, projectId: number | string, source: ImportSource) =>
    wrap<ImportJob>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/imports/jobs`, { source }),
    ),

  /** 导入任务详情。 */
  getJob: (wsId: number | string, projectId: number | string, jobId: number | string) =>
    wrap<ImportJob>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/imports/jobs/${jobId}`),
    ),

  /**
   * CSV 导入（multipart/form-data）。
   * 字段：file（必填）+ default_type（可选，缺省类型）。
   */
  uploadCsv: (
    wsId: number | string,
    projectId: number | string,
    jobId: number | string,
    file: File,
    defaultType?: string,
  ) => {
    const form = new FormData();
    form.append("file", file);
    if (defaultType) form.append("default_type", defaultType);
    return wrap<ImportResult>(
      http.post(
        `/workspaces/${wsId}/projects/${projectId}/imports/jobs/${jobId}/csv`,
        form,
        { headers: { "Content-Type": "multipart/form-data" } },
      ),
    );
  },

  /** 平台导入（Jira / Linear / Asana）。 */
  importPlatform: (
    wsId: number | string,
    projectId: number | string,
    jobId: number | string,
    credentials: ImportCredentials,
    options: ImportOptionsInput = {},
  ) =>
    wrap<ImportResult>(
      http.post(
        `/workspaces/${wsId}/projects/${projectId}/imports/jobs/${jobId}/platform`,
        { credentials, options },
      ),
    ),
};
