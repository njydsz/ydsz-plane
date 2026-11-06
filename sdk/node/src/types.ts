/**
 * Ydsz Plane SDK — 共享类型定义。
 */

/* ------------------------------------------------------------------ */
/* 通用                                                              */
/* ------------------------------------------------------------------ */

/** 分页列表响应 */
export interface Page<T> {
  items: T[];
  total?: number;
}

/** 工作项类型 */
export type WorkItemType = 'requirement' | 'task' | 'defect';

/** 优先级 */
export type Priority = 'urgent' | 'high' | 'medium' | 'low' | 'none';

/** 状态分组 */
export type StateGroup = 'backlog' | 'started' | 'completed' | 'cancelled';

/* ------------------------------------------------------------------ */
/* 工作空间 / 项目                                                    */
/* ------------------------------------------------------------------ */

export interface Workspace {
  id: number;
  name: string;
  slug: string;
  logo_url?: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
}

export interface CreateWorkspaceInput {
  name: string;
  slug?: string;
  logo_url?: string;
}

export interface Project {
  id: number;
  workspace_id: number;
  name: string;
  slug: string;
  description?: string;
  network?: string;
  icon?: string;
  color?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateProjectInput {
  name: string;
  slug?: string;
  description?: string;
  network?: string;
  icon?: string;
  color?: string;
}

/* ------------------------------------------------------------------ */
/* 工作项                                                            */
/* ------------------------------------------------------------------ */

export interface WorkItem {
  id: number;
  project_id: number;
  type: WorkItemType;
  title: string;
  description?: string;
  priority: Priority;
  state_id: number;
  state?: string;
  state_group?: StateGroup;
  estimate_point?: number;
  assignees?: Array<{ id: number; name: string }>;
  labels?: Array<{ id: number; name: string }>;
  module_id?: number;
  sprint_id?: number;
  created_at: string;
  updated_at: string;
}

export interface ListWorkItemsParams {
  workspaceId?: number | string;
  projectId: number | string;
  type?: WorkItemType;
  state?: string;
  assignee?: string;
  limit?: number;
  offset?: number;
}

export interface CreateWorkItemInput {
  workspaceId?: number | string;
  projectId: number | string;
  type: WorkItemType;
  title: string;
  description?: string;
  priority?: Priority;
  assigneeId?: number;
  stateId?: number;
}

/* ------------------------------------------------------------------ */
/* 迭代                                                              */
/* ------------------------------------------------------------------ */

export interface Cycle {
  id: number;
  project_id: number;
  name: string;
  goal?: string;
  status: 'planned' | 'active' | 'completed';
  start_date?: string;
  end_date?: string;
  created_at: string;
}

export interface CreateCycleInput {
  workspaceId?: number | string;
  projectId: number | string;
  name: string;
  goal?: string;
  startDate?: string;
  endDate?: string;
}

/* ------------------------------------------------------------------ */
/* 模块                                                              */
/* ------------------------------------------------------------------ */

export interface Module {
  id: number;
  project_id: number;
  name: string;
  description?: string;
  lead_id?: number;
  status: string;
  workflow_status?: 'backlog' | 'planned' | 'in_progress' | 'paused' | 'completed' | 'cancelled';
  start_date?: string;
  target_date?: string;
}

export interface CreateModuleInput {
  workspaceId?: number | string;
  projectId: number | string;
  name: string;
  description?: string;
  leadId?: number;
  workflowStatus?: Module['workflow_status'];
}

/* ------------------------------------------------------------------ */
/* 版本                                                              */
/* ------------------------------------------------------------------ */

export interface Version {
  id: number;
  workspace_id: number;
  name: string;
  description?: string;
  status: 'planning' | 'active' | 'released' | 'archived';
  start_date?: string;
  target_date?: string;
  released_at?: string;
}

/* ------------------------------------------------------------------ */
/* 文档                                                              */
/* ------------------------------------------------------------------ */

export interface PageDoc {
  id: number;
  project_id: number;
  title: string;
  description_html?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/* ------------------------------------------------------------------ */
/* 成员                                                              */
/* ------------------------------------------------------------------ */

export interface Member {
  id: number;
  user_id: number;
  email: string;
  display_name?: string;
  role: string;
  status: string;
}

/* ------------------------------------------------------------------ */
/* AI Agent                                                          */
/* ------------------------------------------------------------------ */

export interface AIAgent {
  id: number;
  workspace_id: number;
  name: string;
  description?: string;
  instructions?: string;
  model: string;
  temperature: number;
  daily_credit_limit: number;
  used_credits_today: number;
  status: string;
  triggers?: string[];
  skills?: string[];
}

export interface AIMessage {
  id: number;
  conversation_id: number;
  role: 'user' | 'assistant' | 'system';
  content: string;
  created_at: string;
}

export interface ChatResult {
  user_message: AIMessage;
  assistant_message: AIMessage;
  llm_enabled: boolean;
  tool_calls?: Array<{ name: string; args?: unknown; result?: unknown; error?: string }>;
  usage?: { prompt_tokens: number; completion_tokens: number; total_tokens: number };
  run?: { run_id: number; status: string };
}

/* ------------------------------------------------------------------ */
/* 视图 / 倡议 / 团队空间 / 客户                                      */
/* ------------------------------------------------------------------ */

export interface WorkspaceView {
  id: number;
  workspace_id: number;
  name: string;
  layout: 'board' | 'list' | 'spreadsheet' | 'gantt' | 'calendar';
  filters?: unknown;
  is_public: boolean;
  is_favorite: boolean;
}

export interface Initiative {
  id: number;
  workspace_id: number;
  title: string;
  status: 'on_track' | 'at_risk' | 'off_track' | 'completed';
  priority: Priority;
}

export interface Teamspace {
  id: number;
  workspace_id: number;
  name: string;
  description?: string;
  color?: string;
  lead_id: number;
}

export interface Customer {
  id: number;
  workspace_id: number;
  name: string;
  stage: 'lead' | 'prospect' | 'customer' | 'churned';
  tier: 'free' | 'pro' | 'enterprise';
  email?: string;
  industry?: string;
}
