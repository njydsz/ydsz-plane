/**
 * 通知服务 —— 封装站内通知与通知偏好的 REST API 调用。
 *
 * 覆盖两类资源：
 *  - /notifications：通知列表、未读数、已读、全部已读、归档
 *  - /notification-preferences：通知偏好（渠道、汇总频率、免打扰时段）的读写
 */
import { apiClient } from '../client'

/** 站内通知实体，对应后端 notifications 表 */
export interface AppNotification {
  /** 通知唯一 ID */
  id: number
  /** 所属工作空间 ID */
  workspace_id: number
  /** 接收者用户 ID */
  recipient_id: number
  /** 触发该通知的领域事件类型（如 issue.created） */
  event_type: string
  /** 关联实体类型（issue / sprint / version / workspace 等） */
  entity_type: string
  /** 关联实体 ID */
  entity_id: number
  /** 通知标题 */
  title: string
  /** 通知正文（可含富文本片段） */
  body: string
  /** 点击通知后跳转的前端路由 */
  action_url: string
  /** 触发者用户 ID，系统通知为 null */
  actor_id: number | null
  /** 触发者展示名 */
  actor_name: string
  /** 是否已读 */
  is_read: boolean
  /** 是否已归档 */
  is_archived: boolean
  /** 已读时间，未读为 null */
  read_at: string | null
  /** 投递渠道（in_app / email 等） */
  channel: string
  /** 附加的业务载荷（结构化数据） */
  payload: Record<string, any>
  /** 创建时间 */
  created_at: string
}

/** 通知分页列表结果 */
export interface NotificationListResult {
  /** 当前页通知项 */
  items: AppNotification[]
  /** 符合条件的总条数 */
  total: number
  /** 分页大小 */
  limit: number
  /** 偏移量 */
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

/** 通知域 API — 通知列表、未读数、已读 / 归档、偏好读写。 */
export const notificationApi = {
  /** 查询通知列表 */
  async list(wsId: number | string, params?: {
    limit?: number
    offset?: number
    is_read?: boolean
    event_type?: string
    /** 按时间戳过滤，仅返回此时间之后的通知（用于 WS 断线重连补偿） */
    since?: number
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
