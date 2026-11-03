/**
 * DataJob 异步数据任务 API。
 *
 * 后端路由：/api/v1/workspaces/:wsId/data-jobs
 * 权限：读 workspace:read；写 project:update。
 */
import { apiClient } from "../client";

/** 任务状态（与后端 data_job_status 枚举对齐）。 */
export type DataJobStatus = "pending" | "running" | "success" | "failed" | "cancelled";

/** 任务类型常量。 */
export type DataJobType = "import" | "export" | "sync" | "report" | "reindex";

/** 数据任务实体。 */
export interface DataJob {
  id: number;
  tenant_id: number;
  workspace_id: number;
  project_id?: number;
  job_type: DataJobType | string;
  name: string;
  parameters?: Record<string, unknown>;
  progress: number;
  status: DataJobStatus;
  error_message?: string;
  scheduled_at?: string;
  executed_at?: string;
  completed_at?: string;
  duration_ms?: number;
  triggered_by?: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建任务入参。 */
export interface CreateDataJobInput {
  job_type: DataJobType | string;
  name: string;
  parameters?: Record<string, unknown>;
}

/** 列表查询参数。 */
export interface ListDataJobsParams {
  job_type?: string;
  status?: DataJobStatus;
  limit?: number;
  offset?: number;
}

/** 状态中文标签。 */
export const DATA_JOB_STATUS_LABELS: Record<DataJobStatus, string> = {
  pending: "待执行",
  running: "执行中",
  success: "成功",
  failed: "失败",
  cancelled: "已取消",
};

/** 状态对应的颜色（用于进度条与徽章）。 */
export const DATA_JOB_STATUS_COLORS: Record<DataJobStatus, string> = {
  pending: "#94a3b8",
  running: "#2563eb",
  success: "#16a34a",
  failed: "#dc2626",
  cancelled: "#64748b",
};

/** 任务类型中文标签。 */
export const DATA_JOB_TYPE_LABELS: Record<string, string> = {
  import: "数据导入",
  export: "数据导出",
  sync: "数据同步",
  report: "报表生成",
  reindex: "索引重建",
};

export const dataJobApi = {
  /** 任务列表。 */
  async list(wsId: number, params: ListDataJobsParams = {}): Promise<DataJob[]> {
    const query = new URLSearchParams();
    if (params.job_type) query.set("job_type", params.job_type);
    if (params.status) query.set("status", params.status);
    if (params.limit) query.set("limit", String(params.limit));
    if (params.offset) query.set("offset", String(params.offset));

    const qs = query.toString();
    const { data } = await apiClient.get<{ results: DataJob[] }>(
      `/workspaces/${wsId}/data-jobs${qs ? `?${qs}` : ""}`,
    );
    return data.results ?? [];
  },

  /** 任务详情。 */
  async get(wsId: number, jobId: number): Promise<DataJob> {
    const { data } = await apiClient.get<DataJob>(`/workspaces/${wsId}/data-jobs/${jobId}`);
    return data;
  },

  /** 创建任务。 */
  async create(wsId: number, input: CreateDataJobInput): Promise<DataJob> {
    const { data } = await apiClient.post<DataJob>(`/workspaces/${wsId}/data-jobs`, input);
    return data;
  },

  /** 更新进度（0-100）。 */
  async updateProgress(wsId: number, jobId: number, progress: number): Promise<void> {
    await apiClient.patch(`/workspaces/${wsId}/data-jobs/${jobId}/progress`, { progress });
  },

  /** 标记开始执行。 */
  async markRunning(wsId: number, jobId: number): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/data-jobs/${jobId}/run`);
  },

  /** 标记成功。 */
  async markSuccess(wsId: number, jobId: number): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/data-jobs/${jobId}/success`);
  },

  /** 标记失败。 */
  async markFailed(wsId: number, jobId: number, errorMessage: string): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/data-jobs/${jobId}/fail`, {
      error_message: errorMessage,
    });
  },

  /** 取消任务。 */
  async cancel(wsId: number, jobId: number): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/data-jobs/${jobId}/cancel`);
  },

  /** 删除任务（软删除）。 */
  async remove(wsId: number, jobId: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/data-jobs/${jobId}`);
  },
};
