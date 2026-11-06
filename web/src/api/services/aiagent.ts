/**
 * AI Agent 域 API — 对接后端 Plane AI 风格的 Agent / 记忆 / 会话 / 动作接口。
 *
 * 后端路由前缀：/api/v1/workspaces/:workspace_id
 *   - /ai-agents            Agent CRUD
 *   - /ai-memory            记忆（scope: user / project / workspace）
 *   - /ai/conversations     会话与消息
 *   - /ai/actions           待确认动作（批准 / 拒绝 / 执行）
 *
 * 权限：workspace:read（后端在注册层统一校验）。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types（与后端 models.go 对齐）                                       */
/* ------------------------------------------------------------------ */

/** AI 交互模式：Ask 只读问答 / Build 生成待确认动作 / Autopilot 直接执行 */
export type AIMode = "ask" | "build" | "autopilot";

/** 记忆作用域 */
export type AIMemoryScope = "user" | "project" | "workspace";

/** Agent 状态（后端当前仅 active / inactive，预留扩展） */
export type AIAgentStatus = string;

/** Agent 触发器类型 */
export type AIAgentTrigger = string;

/** Agent 技能（可执行的动作类型） */
export type AIAgentSkill = string;

/** AI Agent 定义 */
export interface AIAgent {
  id: number;
  workspace_id: number;
  name: string;
  description: string;
  instructions: string;
  triggers: string[];
  skills: string[];
  scope_project_ids: number[];
  scope_work_item_types: string[];
  model: string;
  temperature: number;
  daily_credit_limit: number;
  used_credits_today: number;
  status: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建 / 更新 Agent 入参 */
export interface AIAgentInput {
  name: string;
  description?: string;
  instructions?: string;
  triggers?: string[];
  skills?: string[];
  scope_project_ids?: number[];
  scope_work_item_types?: string[];
  model?: string;
  temperature?: number;
  daily_credit_limit?: number;
}

/** AI 记忆条目 */
export interface AIWorkspaceMemory {
  id: number;
  workspace_id: number;
  scope: AIMemoryScope;
  scope_id: number;
  key: string;
  content: string;
  source: string;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 写入记忆入参 */
export interface AIWorkspaceMemoryInput {
  scope: AIMemoryScope;
  scope_id?: number;
  key?: string;
  content: string;
  source?: string;
}

/** AI 会话 */
export interface AIConversation {
  id: number;
  workspace_id: number;
  agent_id: number;
  user_id: number;
  mode: AIMode;
  title: string;
  created_at: string;
  updated_at: string;
}

/** 会话消息 */
export interface AIMessage {
  id: number;
  conversation_id: number;
  role: string;
  content: string;
  actions?: AIAction[];
  created_at: string;
}

/** 动作类型（后端定义的 11 种动作，未知类型按 string 兼容） */
export type AIActionType = string;

/** 动作状态 */
export type AIActionStatus = string;

/** AI 动作 */
export interface AIAction {
  id: number;
  message_id: number;
  type: AIActionType;
  /** JSON 字符串（后端 payload 为 text） */
  payload: string;
  status: AIActionStatus;
  result?: string;
  error_message?: string;
  executed_by?: number;
  executed_at?: string | null;
  created_at: string;
}

/** 会话详情响应 */
export interface AIConversationDetail {
  conversation: AIConversation;
  messages: AIMessage[];
}

/** 创建会话入参 */
export interface CreateAIConversationInput {
  agent_id: number;
  mode: AIMode;
  message: string;
  project_id?: number;
}

/** 发送消息入参 */
export interface SendAIMessageInput {
  content: string;
  mode?: AIMode;
  project_id?: number;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 解析动作 payload JSON（失败时返回 null）。 */
export function parseActionPayload(payload: string | null | undefined): Record<string, unknown> | null {
  if (!payload) return null;
  try {
    const parsed = JSON.parse(payload);
    return parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : null;
  } catch {
    return null;
  }
}

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** AI Agent 域 API */
export const aiAgentApi = {
  /* ---- Agent ---- */

  /** 列出工作空间下全部 Agent */
  listAgents: (workspaceId: number | string) =>
    wrap<{ items: AIAgent[] }>(
      http.get(`/workspaces/${workspaceId}/ai-agents`),
    ).then((r) => r.items ?? []),

  /** 获取单个 Agent */
  getAgent: (workspaceId: number | string, id: number | string) =>
    wrap<AIAgent>(http.get(`/workspaces/${workspaceId}/ai-agents/${id}`)),

  /** 创建 Agent */
  createAgent: (workspaceId: number | string, input: AIAgentInput) =>
    wrap<AIAgent>(http.post(`/workspaces/${workspaceId}/ai-agents`, input)),

  /** 更新 Agent */
  updateAgent: (workspaceId: number | string, id: number | string, input: AIAgentInput) =>
    wrap<AIAgent>(http.patch(`/workspaces/${workspaceId}/ai-agents/${id}`, input)),

  /** 删除 Agent */
  deleteAgent: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/ai-agents/${id}`)),

  /* ---- 记忆 ---- */

  /** 列出记忆（可按 scope / scope_id 过滤） */
  listMemory: (
    workspaceId: number | string,
    params?: { scope?: AIMemoryScope | ""; scope_id?: number },
  ) =>
    wrap<{ items: AIWorkspaceMemory[] }>(
      http.get(`/workspaces/${workspaceId}/ai-memory`, { params }),
    ).then((r) => r.items ?? []),

  /** 写入记忆（按 scope+scope_id+key 幂等 upsert） */
  setMemory: (
    workspaceId: number | string,
    input: AIWorkspaceMemoryInput,
  ) =>
    wrap<AIWorkspaceMemory>(
      http.post(`/workspaces/${workspaceId}/ai-memory`, input),
    ),

  /** 删除记忆 */
  deleteMemory: (workspaceId: number | string, id: number | string) =>
    wrap<void>(http.delete(`/workspaces/${workspaceId}/ai-memory/${id}`)),

  /* ---- 会话 ---- */

  /** 列出当前用户的会话 */
  listConversations: (workspaceId: number | string) =>
    wrap<{ items: AIConversation[] }>(
      http.get(`/workspaces/${workspaceId}/ai/conversations`),
    ).then((r) => r.items ?? []),

  /** 创建会话（含首条消息） */
  createConversation: (
    workspaceId: number | string,
    input: CreateAIConversationInput,
  ) =>
    wrap<AIConversation>(
      http.post(`/workspaces/${workspaceId}/ai/conversations`, input),
    ),

  /** 获取会话详情（含消息） */
  getConversation: (workspaceId: number | string, id: number | string) =>
    wrap<AIConversationDetail>(
      http.get(`/workspaces/${workspaceId}/ai/conversations/${id}`),
    ),

  /** 发送消息 */
  sendMessage: (
    workspaceId: number | string,
    conversationId: number | string,
    input: SendAIMessageInput,
  ) =>
    wrap<AIMessage>(
      http.post(
        `/workspaces/${workspaceId}/ai/conversations/${conversationId}/messages`,
        input,
      ),
    ),

  /* ---- 动作 ---- */

  /** 列出会话待处理动作 */
  listActions: (workspaceId: number | string, conversationId: number | string) =>
    wrap<{ items: AIAction[] }>(
      http.get(`/workspaces/${workspaceId}/ai/actions`, {
        params: { conversation_id: conversationId },
      }),
    ).then((r) => r.items ?? []),

  /** 批准动作 */
  approveAction: (workspaceId: number | string, actionId: number | string) =>
    wrap<AIAction>(
      http.post(`/workspaces/${workspaceId}/ai/actions/${actionId}/approve`, {}),
    ),

  /** 拒绝动作 */
  rejectAction: (workspaceId: number | string, actionId: number | string) =>
    wrap<AIAction>(
      http.post(`/workspaces/${workspaceId}/ai/actions/${actionId}/reject`, {}),
    ),

  /** 执行已批准的动作 */
  executeAction: (workspaceId: number | string, actionId: number | string) =>
    wrap<AIAction>(
      http.post(`/workspaces/${workspaceId}/ai/actions/${actionId}/execute`, {}),
    ),
};
