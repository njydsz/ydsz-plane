/**
 * 搜索服务 —— 封装项目级/工作空间级全文搜索、搜索历史与书签的 API 调用。
 */
import { apiClient } from '../client'

/** 单条搜索结果项（issue / sprint / version 共用的归一化结构） */
export interface SearchResultItem {
  /** 实体 ID */
  id: number
  /** 实体类型（issue / sprint / version） */
  type: string
  /** 实体名称 */
  name: string
  /** 简要描述 */
  description: string
  /** 命中的高亮片段 */
  highlight: string
  /** 所属项目 ID（工作空间级搜索时可为空） */
  project_id?: number
  /** 所属项目名称 */
  project_name?: string
  /** 所属工作空间 ID */
  workspace_id?: number
  /** 创建时间 */
  created_at?: string
  /** 相关性排序分数 */
  rank?: number
}

/** 一次搜索请求的响应，按实体类型分组结果 */
export interface SearchResponse {
  /** 按实体类型分组的搜索结果 */
  results: {
    /** 命中的工作项 */
    issues: SearchResultItem[]
    /** 命中的迭代 */
    sprints: SearchResultItem[]
    /** 命中的版本 */
    versions: SearchResultItem[]
  }
  /** 结果总数 */
  total: number
  /** 服务端耗时（毫秒） */
  took_ms: number
}

/** 单条搜索历史记录 */
export interface SearchHistoryItem {
  /** 历史记录 ID */
  id: number
  /** 历史搜索关键字 */
  query: string
  /** 记录创建时间 */
  created_at: string
}

/** 保存的搜索书签，便于一键复用查询 */
export interface SearchBookmark {
  /** 书签 ID */
  id: number
  /** 书签名称 */
  name: string
  /** 保存的查询关键字 */
  query: string
  /** 可选的额外过滤条件 */
  filters?: Record<string, any>
  /** 创建时间 */
  created_at: string
}

export const searchApi = {
  /** 项目级搜索 */
  async searchProject(wsId: number | string, projectId: number | string, params: {
    q: string
    types?: string
    limit?: number
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
  async getHistory(wsId: number | string): Promise<SearchHistoryItem[]> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/history`)
    return data
  },

  /** 清空当前用户的搜索历史 */
  async clearHistory(wsId: number | string): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/history`)
  },

  /** 查询已保存的搜索书签列表 */
  async getBookmarks(wsId: number | string): Promise<SearchBookmark[]> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/bookmarks`)
    return data
  },

  /** 创建一个搜索书签 */
  async createBookmark(wsId: number | string, input: { name: string; query: string }): Promise<SearchBookmark> {
    const { data } = await apiClient.post(`/workspaces/${wsId}/search/bookmarks`, input)
    return data
  },

  /** 删除一个搜索书签 */
  async deleteBookmark(wsId: number | string, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/bookmarks/${id}`)
  }
}
