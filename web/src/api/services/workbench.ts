import { apiClient } from '../client'

export interface WorkbenchSummary {
  my_tasks: {
    today: WorkbenchTask[]
    in_progress: WorkbenchTask[]
    overdue: WorkbenchTask[]
  }
  iteration_overview: {
    current: SprintBrief | null
    next: SprintBrief | null
  }
  recent_items: RecentItem[]
  quick_actions: QuickAction[]
}

export interface WorkbenchTask {
  id: number
  sequence_id: number
  name: string
  type: string
  priority: string
  state_name: string
  state_group: string
  project_id: number
  project_name: string
  target_date: string | null
  assignee_ids: number[]
}

export interface SprintBrief {
  id: number
  name: string
  project_id: number
  project_name: string
  status: string
  start_date: string | null
  end_date: string | null
  progress: number
  issue_count: number
  completed_count: number
}

export interface RecentItem {
  id: number
  entity_type: string
  entity_id: number
  name: string
  project_id: number
  project_name: string
  accessed_at: string
}

export interface QuickAction {
  type: string
  label: string
  icon: string
  route: string
}

export interface WorkbenchConfig {
  layout: any
  visible_widgets: string[]
}

export const workbenchApi = {
  /** 获取工作空间级工作台汇总 */
  async getSummary(wsId: number | string): Promise<WorkbenchSummary> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/workbench/summary`)
    return data
  },

  /** 获取工作台配置 */
  async getConfig(wsId: number | string): Promise<WorkbenchConfig> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/workbench/config`)
    return data
  },

  /** 保存工作台配置 */
  async saveConfig(wsId: number | string, config: WorkbenchConfig): Promise<void> {
    await apiClient.put(`/workspaces/${wsId}/workbench/config`, config)
  },

  /** 获取最近访问 */
  async getRecent(wsId: number | string): Promise<RecentItem[]> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/workbench/recent`)
    return data
  },

  /** 记录访问 */
  async recordRecent(wsId: number | string, item: { entity_type: string; entity_id: number; name: string; project_id: number }): Promise<void> {
    await apiClient.post(`/workspaces/${wsId}/workbench/recent`, item)
  }
}
