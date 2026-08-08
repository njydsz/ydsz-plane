/**
 * 搜索服务 —— 封装项目级/工作空间级全文搜索、搜索历史与书签的 API 调用。
 * 后端：internal/application/search (PostgreSQL FTS 引擎)
 */
import { apiClient } from '../client'

/** 单条搜索结果项（issue / sprint / version 共用的归一化结构，对齐后端 SearchHit） */
export interface SearchResultItem {
  /** 实体 ID */
  id: number
  /** 实体类型（issue / sprint / version） */
  type: string
  /** 实体名称/标题 */
  name: string
  /** 可读标识符（如 YD-123） */
  identifier?: string
  /** 简要描述（原始内容） */
  description?: string
  /** 命中的高亮片段（含 <b> 标签） */
  highlight?: string
  /** 所属项目 ID */
  project_id: number
  /** 所属项目名称 */
  project_name?: string
  /** 相关性排序分数 */
  rank: number
  /** 前端跳转 URL */
  url?: string
}

/** 按实体类型分组的搜索结果 */
export interface SearchResultsGroup {
  issues: SearchResultItem[]
  sprints: SearchResultItem[]
  versions: SearchResultItem[]
  projects: SearchResultItem[]
}

/** 一次搜索请求的响应（对齐后端 SearchResponse） */
export interface SearchResponse {
  /** 原始搜索词 */
  query: string
  /** 按实体类型分组的搜索结果 */
  results: SearchResultsGroup
  /** 结果总数 */
  total: number
  /** 服务端耗时（毫秒） */
  time_ms: number
  /** 查询建议（可选） */
  suggestions?: string[]
}

/** 单条搜索历史记录 */
export interface SearchHistoryItem {
  /** 历史记录 ID */
  id: number
  /** 工作空间 ID */
  workspace_id?: number
  /** 用户 ID */
  user_id?: number
  /** 历史搜索关键字 */
  query: string
  /** 过滤条件快照 */
  filters?: Record<string, any>
  /** 结果数 */
  result_count?: number
  /** 搜索时间（ISO 8601） */
  searched_at: string
}

/** 保存的搜索书签，便于一键复用查询 */
export interface SearchBookmark {
  /** 书签 ID */
  id: number
  /** 工作空间 ID */
  workspace_id?: number
  /** 项目 ID（null 表示全局） */
  project_id?: number | null
  /** 用户 ID */
  user_id?: number
  /** 书签名称 */
  name: string
  /** 保存的查询关键字 */
  query: string
  /** 可选的额外过滤条件 */
  filters?: Record<string, any>
  /** 是否共享 */
  is_shared?: boolean
  /** 排序权重 */
  sort_order?: number
  /** 创建时间 */
  created_at: string
  /** 更新时间 */
  updated_at?: string
}

/** 搜索域 API — 项目级 / 工作空间级全文搜索 + 搜索历史与书签管理。 */
export const searchApi = {
  /** 项目级搜索 */
  async searchProject(wsId: number | string, projectId: number | string, params: {
    q: string
    types?: string
    limit?: number
    offset?: number
  }): Promise<SearchResponse> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/projects/${projectId}/search`, { params })
    return data
  },

  /** 工作空间级全局搜索 */
  async searchWorkspace(wsId: number | string, params: {
    q: string
    types?: string
    limit?: number
    offset?: number
  }): Promise<SearchResponse> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search`, { params })
    return data
  },

  /** 查询当前用户的搜索历史 */
  async getHistory(wsId: number | string): Promise<{ results: SearchHistoryItem[] }> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/history`)
    return data
  },

  /** 删除单条搜索历史 */
  async deleteHistory(wsId: number | string, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/history/${id}`)
  },

  /** 清空当前用户的搜索历史 */
  async clearHistory(wsId: number | string): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/history`)
  },

  /** 查询已保存的搜索书签列表 */
  async getBookmarks(wsId: number | string): Promise<{ results: SearchBookmark[] }> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/bookmarks`)
    return data
  },

  /** 创建一个搜索书签 */
  async createBookmark(wsId: number | string, input: {
    name: string
    query: string
    filters?: Record<string, any>
    is_shared?: boolean
  }): Promise<SearchBookmark> {
    const { data } = await apiClient.post(`/workspaces/${wsId}/search/bookmarks`, input)
    return data
  },

  /** 更新搜索书签 */
  async updateBookmark(wsId: number | string, id: number, input: {
    name?: string
    query?: string
    filters?: Record<string, any>
    is_shared?: boolean
  }): Promise<SearchBookmark> {
    const { data } = await apiClient.patch(`/workspaces/${wsId}/search/bookmarks/${id}`, input)
    return data
  },

  /** 删除一个搜索书签 */
  async deleteBookmark(wsId: number | string, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/bookmarks/${id}`)
  }
}
