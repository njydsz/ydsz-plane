/**
 * Ydsz Plane SDK — 资源模块。
 */
import type { PlaneClient } from './client';
import type {
  Workspace,
  CreateWorkspaceInput,
  Project,
  CreateProjectInput,
  WorkItem,
  ListWorkItemsParams,
  CreateWorkItemInput,
  Cycle,
  CreateCycleInput,
  Module,
  CreateModuleInput,
  Version,
  PageDoc,
  Member,
  AIAgent,
  AIMessage,
  ChatResult,
  WorkspaceView,
  Initiative,
  Teamspace,
  Customer,
  Page,
} from './types';

/* ------------------------------------------------------------------ */
/* 工作空间                                                          */
/* ------------------------------------------------------------------ */

export class WorkspacesApi {
  constructor(private readonly client: PlaneClient) {}

  list() {
    return this.client.request<Page<Workspace>>('GET', '/workspaces');
  }

  create(input: CreateWorkspaceInput) {
    return this.client.request<Workspace>('POST', '/workspaces', { body: input });
  }

  get(id: number | string) {
    return this.client.request<Workspace>('GET', `/workspaces/${id}`);
  }

  getBySlug(slug: string) {
    return this.client.request<Workspace>('GET', `/workspaces/slug/${slug}`);
  }
}

/* ------------------------------------------------------------------ */
/* 项目                                                              */
/* ------------------------------------------------------------------ */

export class ProjectsApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<Project>>('GET', `/workspaces/${workspaceId}/projects`);
  }

  create(workspaceId: number | string, input: CreateProjectInput) {
    return this.client.request<Project>('POST', `/workspaces/${workspaceId}/projects`, { body: input });
  }

  get(workspaceId: number | string, projectId: number | string) {
    return this.client.request<Project>('GET', `/workspaces/${workspaceId}/projects/${projectId}`);
  }

  update(workspaceId: number | string, projectId: number | string, input: Partial<CreateProjectInput>) {
    return this.client.request<Project>('PATCH', `/workspaces/${workspaceId}/projects/${projectId}`, { body: input });
  }

  archive(workspaceId: number | string, projectId: number | string) {
    return this.client.request<Project>('POST', `/workspaces/${workspaceId}/projects/${projectId}/archive`);
  }
}

/* ------------------------------------------------------------------ */
/* 工作项                                                            */
/* ------------------------------------------------------------------ */

export class WorkItemsApi {
  constructor(private readonly client: PlaneClient) {}

  list(params: ListWorkItemsParams) {
    const { projectId, ...query } = params;
    return this.client.request<Page<WorkItem>>(
      'GET',
      `/workspaces/${params.workspaceId ?? ''}/projects/${projectId}/issues`,
      { query },
    );
  }

  get(workspaceId: number | string, projectId: number | string, type: string, id: number | string) {
    return this.client.request<WorkItem>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/issues/${id}`,
      { query: { type } },
    );
  }

  create(input: CreateWorkItemInput) {
    return this.client.request<WorkItem>(
      'POST',
      `/workspaces/${input.workspaceId ?? ''}/projects/${input.projectId}/issues`,
      { body: input },
    );
  }

  update(
    workspaceId: number | string,
    projectId: number | string,
    type: string,
    id: number | string,
    input: Partial<CreateWorkItemInput>,
  ) {
    return this.client.request<WorkItem>(
      'PATCH',
      `/workspaces/${workspaceId}/projects/${projectId}/issues/${id}`,
      { body: input, query: { type } },
    );
  }

  remove(workspaceId: number | string, projectId: number | string, type: string, id: number | string) {
    return this.client.request<void>(
      'DELETE',
      `/workspaces/${workspaceId}/projects/${projectId}/issues/${id}`,
      { query: { type } },
    );
  }

  addComment(
    workspaceId: number | string,
    projectId: number | string,
    type: string,
    id: number | string,
    content: string,
  ) {
    return this.client.request<unknown>(
      'POST',
      `/workspaces/${workspaceId}/projects/${projectId}/issues/${id}/comments`,
      { body: { content }, query: { type } },
    );
  }
}

/* ------------------------------------------------------------------ */
/* 迭代                                                              */
/* ------------------------------------------------------------------ */

export class CyclesApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string, projectId: number | string, status?: string) {
    return this.client.request<Page<Cycle>>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/sprints`,
      { query: { status } },
    );
  }

  create(input: CreateCycleInput) {
    return this.client.request<Cycle>(
      'POST',
      `/workspaces/${input.workspaceId ?? ''}/projects/${input.projectId}/sprints`,
      { body: input },
    );
  }

  start(workspaceId: number | string, projectId: number | string, id: number | string) {
    return this.client.request<Cycle>(
      'POST',
      `/workspaces/${workspaceId}/projects/${projectId}/sprints/${id}/start`,
    );
  }

  complete(workspaceId: number | string, projectId: number | string, id: number | string) {
    return this.client.request<Cycle>(
      'POST',
      `/workspaces/${workspaceId}/projects/${projectId}/sprints/${id}/complete`,
    );
  }

  analytics(workspaceId: number | string, projectId: number | string, id: number | string) {
    return this.client.request<unknown>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/sprints/${id}/analytics`,
    );
  }
}

/* ------------------------------------------------------------------ */
/* 模块                                                              */
/* ------------------------------------------------------------------ */

export class ModulesApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string, projectId: number | string) {
    return this.client.request<Page<Module>>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/modules`,
    );
  }

  create(input: CreateModuleInput) {
    return this.client.request<Module>(
      'POST',
      `/workspaces/${input.workspaceId ?? ''}/projects/${input.projectId}/modules`,
      { body: input },
    );
  }

  update(
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
    input: Partial<CreateModuleInput>,
  ) {
    return this.client.request<Module>(
      'PATCH',
      `/workspaces/${workspaceId}/projects/${projectId}/modules/${id}`,
      { body: input },
    );
  }

  remove(workspaceId: number | string, projectId: number | string, id: number | string) {
    return this.client.request<void>(
      'DELETE',
      `/workspaces/${workspaceId}/projects/${projectId}/modules/${id}`,
    );
  }
}

/* ------------------------------------------------------------------ */
/* 版本                                                              */
/* ------------------------------------------------------------------ */

export class VersionsApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<Version>>('GET', `/workspaces/${workspaceId}/versions`);
  }

  create(workspaceId: number | string, input: { name: string; description?: string }) {
    return this.client.request<Version>('POST', `/workspaces/${workspaceId}/versions`, { body: input });
  }

  release(workspaceId: number | string, id: number | string) {
    return this.client.request<Version>('POST', `/workspaces/${workspaceId}/versions/${id}/release`);
  }
}

/* ------------------------------------------------------------------ */
/* 文档                                                              */
/* ------------------------------------------------------------------ */

export class PagesApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string, projectId: number | string) {
    return this.client.request<Page<PageDoc>>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/pages`,
    );
  }

  get(workspaceId: number | string, projectId: number | string, id: number | string) {
    return this.client.request<PageDoc>(
      'GET',
      `/workspaces/${workspaceId}/projects/${projectId}/pages/${id}`,
    );
  }

  create(workspaceId: number | string, projectId: number | string, input: { title: string; content?: string }) {
    return this.client.request<PageDoc>(
      'POST',
      `/workspaces/${workspaceId}/projects/${projectId}/pages`,
      { body: input },
    );
  }
}

/* ------------------------------------------------------------------ */
/* 成员                                                              */
/* ------------------------------------------------------------------ */

export class MembersApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<Member>>('GET', `/workspaces/${workspaceId}/members`);
  }

  updateRole(workspaceId: number | string, id: number | string, role: string) {
    return this.client.request<Member>('PATCH', `/workspaces/${workspaceId}/members/${id}`, {
      body: { role },
    });
  }

  remove(workspaceId: number | string, id: number | string) {
    return this.client.request<void>('DELETE', `/workspaces/${workspaceId}/members/${id}`);
  }
}

/* ------------------------------------------------------------------ */
/* 搜索                                                              */
/* ------------------------------------------------------------------ */

export class SearchApi {
  constructor(private readonly client: PlaneClient) {}

  query(workspaceId: number | string, q: string, limit = 20) {
    return this.client.request<Page<unknown>>('GET', `/workspaces/${workspaceId}/search`, {
      query: { q, limit },
    });
  }
}

/* ------------------------------------------------------------------ */
/* AI Agent                                                          */
/* ------------------------------------------------------------------ */

export class AIAgentsApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<AIAgent>>('GET', `/workspaces/${workspaceId}/ai-agents`);
  }

  create(workspaceId: number | string, input: Partial<AIAgent>) {
    return this.client.request<AIAgent>('POST', `/workspaces/${workspaceId}/ai-agents`, { body: input });
  }

  remove(workspaceId: number | string, id: number | string) {
    return this.client.request<void>('DELETE', `/workspaces/${workspaceId}/ai-agents/${id}`);
  }

  /** 发起对话（返回含 assistant 回复的结果） */
  chat(
    workspaceId: number | string,
    conversationId: number | string,
    content: string,
    mode?: 'ask' | 'build' | 'autopilot',
  ) {
    return this.client.request<ChatResult>(
      'POST',
      `/workspaces/${workspaceId}/ai/conversations/${conversationId}/messages`,
      { body: { content, mode } },
    );
  }

  messages(workspaceId: number | string, conversationId: number | string) {
    return this.client.request<{ conversation: unknown; messages: AIMessage[] }>(
      'GET',
      `/workspaces/${workspaceId}/ai/conversations/${conversationId}`,
    );
  }
}

/* ------------------------------------------------------------------ */
/* 视图 / 倡议 / 团队空间 / 客户                                      */
/* ------------------------------------------------------------------ */

export class ViewsApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<WorkspaceView>>('GET', `/workspaces/${workspaceId}/views`);
  }

  create(workspaceId: number | string, input: Partial<WorkspaceView>) {
    return this.client.request<WorkspaceView>('POST', `/workspaces/${workspaceId}/views`, { body: input });
  }

  duplicate(workspaceId: number | string, id: number | string) {
    return this.client.request<WorkspaceView>('POST', `/workspaces/${workspaceId}/views/${id}/duplicate`);
  }

  toggleFavorite(workspaceId: number | string, id: number | string) {
    return this.client.request<WorkspaceView>('POST', `/workspaces/${workspaceId}/views/${id}/favorite`);
  }
}

export class InitiativesApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<Initiative>>('GET', `/workspaces/${workspaceId}/initiatives`);
  }

  create(workspaceId: number | string, input: { title: string; description?: string; priority?: string }) {
    return this.client.request<Initiative>('POST', `/workspaces/${workspaceId}/initiatives`, { body: input });
  }

  progress(workspaceId: number | string, id: number | string) {
    return this.client.request<unknown>('GET', `/workspaces/${workspaceId}/initiatives/${id}/progress`);
  }
}

export class TeamspacesApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string) {
    return this.client.request<Page<Teamspace>>('GET', `/workspaces/${workspaceId}/teamspaces`);
  }

  create(workspaceId: number | string, input: { name: string; description?: string; color?: string }) {
    return this.client.request<Teamspace>('POST', `/workspaces/${workspaceId}/teamspaces`, { body: input });
  }

  progress(workspaceId: number | string, id: number | string) {
    return this.client.request<unknown>('GET', `/workspaces/${workspaceId}/teamspaces/${id}/progress`);
  }
}

export class CustomersApi {
  constructor(private readonly client: PlaneClient) {}

  list(workspaceId: number | string, params?: { stage?: string; tier?: string; q?: string }) {
    return this.client.request<Page<Customer>>('GET', `/workspaces/${workspaceId}/customers`, {
      query: params,
    });
  }

  create(workspaceId: number | string, input: { name: string; email?: string; stage?: string; tier?: string }) {
    return this.client.request<Customer>('POST', `/workspaces/${workspaceId}/customers`, { body: input });
  }
}
