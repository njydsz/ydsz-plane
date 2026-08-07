/**
 * 通知 Pinia store —— 集中管理工作空间内站内通知的状态。
 *
 * 职责：
 *  - 缓存当前工作空间的通知列表（items）
 *  - 维护未读数量（unreadCount），驱动顶部铃铛角标
 *  - 提供已读、全部已读、归档等操作，并在本地同步变更后的状态
 *
 * 设计说明：
 *  - 所有写操作（markRead / markAllRead / archive）成功后都会同步更新
 *    本地内存中的 items / unreadCount，避免依赖后端返回再做二次请求。
 *  - 读取类异常静默吞掉（catch 为空），保证组件层拿到稳定空状态而非抛错。
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notificationApi, type AppNotification } from '@/api/services/notification'

export const useNotificationStore = defineStore('notification', () => {
  /** 当前工作空间的通知列表 */
  const items = ref<AppNotification[]>([])
  /** 未读通知数量，用于铃铛角标展示 */
  const unreadCount = ref(0)
  /** 列表加载中的标志位 */
  const loading = ref(false)

  /** 拉取指定工作空间的未读数量并更新角标 */
  async function fetchUnreadCount(wsId: number | string) {
    try {
      unreadCount.value = await notificationApi.unreadCount(wsId)
    } catch { /* 静默失败：保持当前角标不变 */ }
  }

  /**
   * 拉取指定工作空间的通知列表（分页），刷新后同步未读数。
   * @param wsId 工作空间 ID
   * @param params 可选过滤参数：limit（每页条数）、is_read（按已读状态过滤）、since（时间戳过滤，用于断线补偿）
   */
  async function fetchList(wsId: number | string, params?: { limit?: number; is_read?: boolean; since?: number }) {
    loading.value = true
    try {
      const result = await notificationApi.list(wsId, { limit: params?.limit ?? 20, is_read: params?.is_read, since: params?.since })
      items.value = result.items
      await fetchUnreadCount(wsId)
    } catch { /* 静默失败：列表保持旧值 */ }
    finally { loading.value = false }
  }

  /**
   * 将单条通知标记为已读，并在本地同步 items 与未读数。
   * @param wsId 工作空间 ID
   * @param id 目标通知 ID
   */
  async function markRead(wsId: number | string, id: number) {
    await notificationApi.markRead(wsId, id)
    const item = items.value.find(n => n.id === id)
    if (item) {
      item.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  }

  /** 将当前工作空间全部通知标记为已读，未读数清零 */
  async function markAllRead(wsId: number | string) {
    await notificationApi.markAllRead(wsId)
    items.value.forEach(n => { n.is_read = true })
    unreadCount.value = 0
  }

  /**
   * 归档单条通知：从列表中移除并刷新未读数。
   * @param wsId 工作空间 ID
   * @param id 目标通知 ID
   */
  async function archive(wsId: number | string, id: number) {
    await notificationApi.archive(wsId, id)
    items.value = items.value.filter(n => n.id !== id)
    await fetchUnreadCount(wsId)
  }

  /** 清空本地通知状态（常用于切换工作空间时重置） */
  function clear() {
    items.value = []
    unreadCount.value = 0
  }

  return { items, unreadCount, loading, fetchUnreadCount, fetchList, markRead, markAllRead, archive, clear }
})
