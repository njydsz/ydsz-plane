/**
 * 内容模板 API — 需求/任务/缺陷模板库的 CRUD 操作。
 *
 * 路径：/api/v1/workspaces/:ws/projects/:pid/content-templates
 */
import { apiClient } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 内容模板实体 */
export interface ContentTemplate {
  id: number;
  tenant_id: number;
  workspace_id: number;
  project_id?: number | null;
  name: string;
  template_type: "requirement" | "task" | "defect";
  content_json: Record<string, any>;
  content_html?: string;
  is_default: boolean;
  status: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建模板请求 */
export interface CreateTemplateRequest {
  name: string;
  template_type: "requirement" | "task" | "defect";
  content_json: Record<string, any>;
  content_html?: string;
  is_default?: boolean;
}

/** 更新模板请求 */
export interface UpdateTemplateRequest {
  name?: string;
  content_json?: Record<string, any>;
  content_html?: string;
  is_default?: boolean;
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 内容模板 API */
export const contentTemplateApi = {
  /**
   * 列出内容模板（按类型筛选）。
   */
  async list(wsId: number, projectId: number, templateType?: string): Promise<ContentTemplate[]> {
    const qs = new URLSearchParams();
    if (templateType) qs.set("template_type", templateType);
    const q = qs.toString();
    const { data } = await apiClient.get<{ results: ContentTemplate[] }>(
      `/workspaces/${wsId}/projects/${projectId}/content-templates${q ? `?${q}` : ""}`,
    );
    return data.results;
  },

  /**
   * 创建内容模板。
   */
  async create(wsId: number, projectId: number, input: CreateTemplateRequest): Promise<ContentTemplate> {
    const { data } = await apiClient.post<ContentTemplate>(
      `/workspaces/${wsId}/projects/${projectId}/content-templates`,
      input,
    );
    return data;
  },

  /**
   * 更新内容模板。
   */
  async update(wsId: number, projectId: number, templateId: number, input: UpdateTemplateRequest): Promise<ContentTemplate> {
    const { data } = await apiClient.patch<ContentTemplate>(
      `/workspaces/${wsId}/projects/${projectId}/content-templates/${templateId}`,
      input,
    );
    return data;
  },

  /**
   * 删除内容模板。
   */
  async delete(wsId: number, projectId: number, templateId: number): Promise<void> {
    await apiClient.delete(
      `/workspaces/${wsId}/projects/${projectId}/content-templates/${templateId}`,
    );
  },
};
