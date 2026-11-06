/**
 * Ydsz Plane Node.js SDK
 *
 * 类型化的 API 客户端，覆盖工作空间 / 项目 / 工作项 / 迭代 / 模块 /
 * 版本 / 文档 / 成员 / 搜索 / AI / 视图 等域。
 *
 * 安装：
 *   npm install ydsz-plane
 *
 * 快速上手：
 *   import { PlaneClient } from 'ydsz-plane';
 *   const plane = new PlaneClient({ baseURL: 'https://your-host', token: 'ydz_xxx' });
 *   const items = await plane.workItems.list({ projectId: 1, type: 'task' });
 */
import {
  WorkspacesApi,
  ProjectsApi,
  WorkItemsApi,
  CyclesApi,
  ModulesApi,
  VersionsApi,
  PagesApi,
  MembersApi,
  SearchApi,
  AIAgentsApi,
  ViewsApi,
  InitiativesApi,
  TeamspacesApi,
  CustomersApi,
} from './resources';

/** 客户端配置 */
export interface PlaneClientOptions {
  /** API 基地址，例如 https://plane.example.com */
  baseURL: string;
  /** API Token（ydz_ 前缀）或 OAuth Access Token（oat_ 前缀） */
  token: string;
  /** 自定义 fetch（默认使用全局 fetch） */
  fetch?: typeof fetch;
  /** 请求超时（毫秒），默认 30000 */
  timeout?: number;
}

/** 请求选项 */
export interface RequestOptions {
  /** 覆盖超时 */
  timeout?: number;
  /** 自定义 header */
  headers?: Record<string, string>;
}

/** API 错误 */
export class PlaneError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: string,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = 'PlaneError';
  }
}

/** Ydsz Plane 客户端 */
export class PlaneClient {
  readonly workspaces: WorkspacesApi;
  readonly projects: ProjectsApi;
  readonly workItems: WorkItemsApi;
  readonly cycles: CyclesApi;
  readonly modules: ModulesApi;
  readonly versions: VersionsApi;
  readonly pages: PagesApi;
  readonly members: MembersApi;
  readonly search: SearchApi;
  readonly aiAgents: AIAgentsApi;
  readonly views: ViewsApi;
  readonly initiatives: InitiativesApi;
  readonly teamspaces: TeamspacesApi;
  readonly customers: CustomersApi;

  private readonly baseURL: string;
  private readonly token: string;
  private readonly fetchImpl: typeof fetch;
  private readonly timeout: number;

  constructor(options: PlaneClientOptions) {
    this.baseURL = options.baseURL.replace(/\/+$/, '');
    this.token = options.token;
    this.fetchImpl = options.fetch ?? globalThis.fetch;
    this.timeout = options.timeout ?? 30_000;

    this.workspaces = new WorkspacesApi(this);
    this.projects = new ProjectsApi(this);
    this.workItems = new WorkItemsApi(this);
    this.cycles = new CyclesApi(this);
    this.modules = new ModulesApi(this);
    this.versions = new VersionsApi(this);
    this.pages = new PagesApi(this);
    this.members = new MembersApi(this);
    this.search = new SearchApi(this);
    this.aiAgents = new AIAgentsApi(this);
    this.views = new ViewsApi(this);
    this.initiatives = new InitiativesApi(this);
    this.teamspaces = new TeamspacesApi(this);
    this.customers = new CustomersApi(this);
  }

  /** 发送 HTTP 请求（供各资源模块调用） */
  async request<T>(
    method: string,
    path: string,
    options?: {
      query?: Record<string, unknown>;
      body?: unknown;
      workspaceId?: number | string;
    } & RequestOptions,
  ): Promise<T> {
    const url = new URL(`${this.baseURL}/api/v1${path}`);
    if (options?.query) {
      for (const [k, v] of Object.entries(options.query)) {
        if (v !== undefined && v !== null && v !== '') {
          url.searchParams.set(k, String(v));
        }
      }
    }

    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), options?.timeout ?? this.timeout);

    let response: Response;
    try {
      response = await this.fetchImpl(url.toString(), {
        method,
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${this.token}`,
          ...(options?.headers ?? {}),
        },
        body: options?.body !== undefined ? JSON.stringify(options.body) : undefined,
      });
    } catch (err) {
      clearTimeout(timer);
      if (err instanceof Error && err.name === 'AbortError') {
        throw new PlaneError('Request timeout', 0, 'TIMEOUT');
      }
      throw err;
    }
    clearTimeout(timer);

    const text = await response.text();
    let data: unknown = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = text;
      }
    }

    if (!response.ok) {
      const d = data as { error?: string; code?: string; details?: unknown } | null;
      throw new PlaneError(
        d?.error ?? `HTTP ${response.status}`,
        response.status,
        d?.code,
        d?.details,
      );
    }
    return data as T;
  }
}

export default PlaneClient;
