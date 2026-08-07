import { apiClient } from '../client'

export interface SearchResultItem {
  id: number
  type: string
  name: string
  description: string
  highlight: string
  project_id?: number
  project_name?: string
  workspace_id?: number
  created_at?: string
  rank?: number
}

export interface SearchResponse {
  results: {
    issues: SearchResultItem[]
    sprints: SearchResultItem[]
    versions: SearchResultItem[]
  }
  total: number
  took_ms: number
}

export interface SearchHistoryItem {
  id: number
  query: string
  created_at: string
}

export interface SearchBookmark {
  id: number
  name: string
  query: string
  filters?: Record<string, any>
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

  /** 搜索历史 */
  async getHistory(wsId: number | string): Promise<SearchHistoryItem[]> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/history`)
    return data
  },

  async clearHistory(wsId: number | string): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/history`)
  },

  /** 搜索书签 */
  async getBookmarks(wsId: number | string): Promise<SearchBookmark[]> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/search/bookmarks`)
    return data
  },

  async createBookmark(wsId: number | string, input: { name: string; query: string }): Promise<SearchBookmark> {
    const { data } = await apiClient.post(`/workspaces/${wsId}/search/bookmarks`, input)
    return data
  },

  async deleteBookmark(wsId: number | string, id: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/search/bookmarks/${id}`)
  }
}
