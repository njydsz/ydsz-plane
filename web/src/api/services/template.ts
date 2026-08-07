/**
 * 项目模板 API — 对接预置模板接口（agile/waterfall/generic）。
 */
import { apiClient } from '../client';

/** 项目模板 */
export interface ProjectTemplate {
  code: 'agile' | 'waterfall' | 'generic';
  name: string;
  description: string;
  apply_dev_flow: boolean;
  apply_defect_flow: boolean;
  apply_requirement_flow: boolean;
}

export const templateApi = {
  /** 获取全部预置模板列表 */
  async listTemplates(wsId: number): Promise<ProjectTemplate[]> {
    const { data } = await apiClient.get<ProjectTemplate[]>(
      `/api/v1/workspaces/${wsId}/templates`,
    );
    return data;
  },
};
