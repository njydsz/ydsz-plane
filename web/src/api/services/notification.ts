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
  }
}
