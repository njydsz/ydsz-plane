/**
 * 状态机配置 API — 封装状态与流转规则的 CRUD 接口。
 *
 * 所有路径挂载在 /api/v1/workspaces/:ws/projects/:pid 下。
 */
import { apiClient } from '../client';

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 状态分组 */
export type StateGroup = 'backlog' | 'started' | 'completed' | 'cancelled';

/** 需求/任务/缺陷类型 */
export type IssueTypeCode = 'epic' | 'requirement' | 'task' | 'defect';

/** 状态 */
export interface State {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  group: StateGroup;
  color: string;
  sequence: number;
  is_default: boolean;
  applicable_types: string[];
  created_at: string;
  updated_at: string;
}

/** 流转规则 */
export interface TransitionRule {
  id: number;
  from_state_id: number;
  from_state_name?: string;
  to_state_id: number;
  to_state_name?: string;
  type_code: string;
  required_fields: string[];
}

/** 创建状态请求 */
export interface CreateStateRequest {
  name: string;
  group?: StateGroup;
  color?: string;
  sequence?: number;
  is_default?: boolean;
  applicable_types?: string[];
}

/** 更新状态请求 */
export interface UpdateStateRequest {
  name?: string;
  group?: StateGroup;
  color?: string;
  sequence?: number;
  is_default?: boolean;
  applicable_types?: string[];
}

/** 添加流转规则请求 */
export interface AddTransitionRequest {
  from_state_id: number;
  to_state_id: number;
  type_code?: string;
  required_fields?: string[];
}

/** 状态机完整配置（状态 + 流转规则） */
export interface StateMachineConfig {
  states: State[];
  transitions: TransitionRule[];
}

/** 状态分组标签映射 */
export const STATE_GROUP_LABELS: Record<StateGroup, string> = {
  backlog: '待办',
  started: '进行中',
  completed: '已完成',
  cancelled: '已取消',
};

/** 状态分组颜色映射（用于 UI 默认值） */
export const STATE_GROUP_COLORS: Record<StateGroup, string> = {
  backlog: '#8DA2C2',
  started: '#F59E0B',
  completed: '#10B981',
  cancelled: '#9CA3AF',
};

/** 需求/任务/缺陷类型标签映射 */
export const TYPE_CODE_LABELS: Record<string, string> = {
  all: '全部类型',
  epic: '史诗',
  requirement: '需求',
  task: '任务',
  defect: '缺陷',
};

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

export const stateMachineApi = {
  /**
   * 获取项目全部状态（按 sequence 排序）。
   */
  async getStates(wsId: number, projectId: number): Promise<State[]> {
    const { data } = await apiClient.get<{ results: State[] }>(
      `/workspaces/${wsId}/projects/${projectId}/states`,
    );
    return data.results;
  },

  /**
   * 创建自定义状态。
   */
  async createState(wsId: number, projectId: number, req: CreateStateRequest): Promise<State> {
    const { data } = await apiClient.post<State>(
      `/workspaces/${wsId}/projects/${projectId}/states`,
      req,
    );
    return data;
  },

  /**
   * 更新状态属性。
   */
  async updateState(wsId: number, projectId: number, stateId: number, req: UpdateStateRequest): Promise<State> {
    const { data } = await apiClient.patch<State>(
      `/workspaces/${wsId}/projects/${projectId}/states/${stateId}`,
      req,
    );
    return data;
  },

  /**
   * 删除状态（软删除，若需求/任务/缺陷在使用则返回 409）。
   */
  async deleteState(wsId: number, projectId: number, stateId: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/projects/${projectId}/states/${stateId}`);
  },

  /**
   * 获取项目全部流转规则。
   */
  async getTransitions(wsId: number, projectId: number): Promise<TransitionRule[]> {
    const { data } = await apiClient.get<{ results: TransitionRule[] }>(
      `/workspaces/${wsId}/projects/${projectId}/state-transitions`,
    );
    return data.results;
  },

  /**
   * 添加流转规则。
   */
  async addTransition(wsId: number, projectId: number, req: AddTransitionRequest): Promise<TransitionRule> {
    const { data } = await apiClient.post<TransitionRule>(
      `/workspaces/${wsId}/projects/${projectId}/state-transitions`,
      req,
    );
    return data;
  },

  /**
   * 删除流转规则。
   */
  async removeTransition(wsId: number, projectId: number, transitionId: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/projects/${projectId}/state-transitions/${transitionId}`);
  },
};
