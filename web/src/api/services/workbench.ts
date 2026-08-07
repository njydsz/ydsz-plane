/**
 * 工作台服务 —— 封装工作空间级工作台汇总、配置、最近访问的 API 调用。
 */
import { apiClient } from '../client'

/** 工作台汇总数据：聚合个人任务、迭代概览、最近访问与快捷动作 */
export interface WorkbenchSummary {
  /** 按日期分组我的任务 */
  my_tasks: {
    /** 今日到期任务 */
    today: WorkbenchTask[]
    /** 进行中的任务 */
    in_progress: WorkbenchTask[]
    /** 已逾期任务 */
    overdue: WorkbenchTask[]
  }
  /** 迭代概览 */
  iteration_overview: {
    /** 当前进行中的迭代 */
    current: SprintBrief | null
    /** 下一个即将开始的迭代 */
    next: SprintBrief | null
  }
  /** 最近访问的实体列表 */
  recent_items: RecentItem[]
  /** 可供前端渲染的快捷动作 */
  quick_actions: QuickAction[]
}

/** 工作台中的单条任务（工作项）摘要 */
export interface WorkbenchTask {
  /** 工作项 ID */
  id: number
  /** 项目内序列号（展示用编号） */
  sequence_id: number
  /** 工作项名称 */
  name: string
  /** 类型（task / bug / story 等） */
  type: string
  /** 优先级 */
  priority: string
  /** 状态显示名 */
  state_name: string
  /** 状态分组（todo / in_progress / done 等） */
  state_group: string
  /** 所属项目 ID */
  project_id: number
  /** 所属项目名称 */
  project_name: string
  /** 目标完成日期，未设置为 null */
  target_date: string | null
  /** 指派者用户 ID 列表 */
  assignee_ids: number[]
}

/** 迭代的简要信息 */
export interface SprintBrief {
  /** 迭代 ID */
  id: number
  /** 迭代名称 */
  name: string
  /** 所属项目 ID */
  project_id: number
  /** 所属项目名称 */
  project_name: string
  /** 迭代状态（planned / in_progress / completed） */
  status: string
  /** 开始日期 */
  start_date: string | null
  /** 结束日期 */
  end_date: string | null
  /** 完成进度（0-100） */
  progress: number
  /** 迭代内工作项总数 */
  issue_count: number
  /** 已完成工作项数 */
  completed_count: number
}

/** 最近访问的实体记录 */
export interface RecentItem {
  /** 记录 ID */
  id: number
  /** 实体类型（issue / project 等） */
  entity_type: string
  /** 实体 ID */
  entity_id: number
  /** 实体名称 */
  name: string
  /** 所属项目 ID */
  project_id: number
  /** 所属项目名称 */
  project_name: string
  /** 最近访问时间 */
  accessed_at: string
}

/** 工作台快捷动作 */
export interface QuickAction {
  /** 动作类型标识 */
  type: string
  /** 展示文案 */
  label: string
  /** 图标标识 */
  icon: string
  /** 点击后跳转的前端路由 */
  route: string
}

/** 工作台布局与可见组件配置 */
export interface WorkbenchConfig {
  /** 布局配置（任意结构，由前端解释） */
  layout: any
  /** 当前可见的小组件 ID 列表 */
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
