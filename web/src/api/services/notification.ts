import { apiClient } from '../client'

export interface AppNotification {
  id: number
  workspace_id: number
  recipient_id: number
  event_type: string
  entity_type: string
  entity_id: number
  title: string
  body: string
  action_url: string
  actor_id: number | null
  actor_name: string
  is_read: boolean
  is_archived: boolean
  read_at: string | null
  channel: string
  payload: Record<string, any>
  created_at: string
}

export interface NotificationListResult {
  items: AppNotification[]
  total: number
  limit: number
  offset: number
}

/** 通知偏好 */
export interface NotificationPreference {
  id: number
  user_id: number
  workspace_id: number
  event_types: string[]
  channels: string[]
  digest: 'realtime' | 'daily' | 'weekly' | 'off'
  dnd_enabled: boolean
  dnd_start: string
  dnd_end: string
  is_enabled: boolean
  created_at: string
  updated_at: string
}

/** 更新通知偏好入参 */
export interface PreferenceUpdateInput {
  event_types?: string[]
  channels?: string[]
  digest?: 'realtime' | 'daily' | 'weekly' | 'off'
  dnd_enabled?: boolean
  dnd_start?: string
  dnd_end?: string
  is_enabled?: boolean
}

export const notificationApi = {
  /** 查询通知列表 */
  async list(wsId: number | string, params?: {
    limit?: number
    offset?: number
    is_read?: boolean
    event_type?: string
  }): Promise<NotificationListResult> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/notifications`, { params })
    return data
  },

  /** 获取未读数量 */
  async unreadCount(wsId: number | string): Promise<number> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/notifications/unread-count`)
    return data.count ?? 0
  },

  /** 标记已读 */
  async markRead(wsId: number | string, id: number): Promise<void> {
    await apiClient.put(`/workspaces/${wsId}/notifications/${id}/read`)
  },

  /** 全部已读 */
  async markAllRead(wsId: number | string): Promise<{ count: number }> {
    const { data } = await apiClient.put(`/workspaces/${wsId}/notifications/read-all`)
    return data
  },

  /** 归档通知 */
  async archive(wsId: number | string, id: number): Promise<void> {
    await apiClient.put(`/workspaces/${wsId}/notifications/${id}/archive`)
  },

  /** 获取通知偏好 */
  async getPreference(wsId: number | string): Promise<NotificationPreference> {
    const { data } = await apiClient.get(`/workspaces/${wsId}/notification-preferences`)
    return data
  },

  /** 更新通知偏好 */
  async updatePreference(wsId: number | string, input: PreferenceUpdateInput): Promise<NotificationPreference> {
    const { data } = await apiClient.put(`/workspaces/${wsId}/notification-preferences`, input)
    return data
  }
}
