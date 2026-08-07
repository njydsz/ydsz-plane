import { defineStore } from 'pinia'
import { ref } from 'vue'
import { notificationApi, type AppNotification } from '@/api/services/notification'

export const useNotificationStore = defineStore('notification', () => {
  const items = ref<AppNotification[]>([])
  const unreadCount = ref(0)
  const loading = ref(false)

  async function fetchUnreadCount(wsId: number | string) {
    try {
      unreadCount.value = await notificationApi.unreadCount(wsId)
    } catch { /* silent */ }
  }

  async function fetchList(wsId: number | string, params?: { limit?: number; is_read?: boolean }) {
    loading.value = true
    try {
      const result = await notificationApi.list(wsId, { limit: params?.limit ?? 20, is_read: params?.is_read })
      items.value = result.items
      await fetchUnreadCount(wsId)
    } catch { /* silent */ }
    finally { loading.value = false }
  }

  async function markRead(wsId: number | string, id: number) {
    await notificationApi.markRead(wsId, id)
    const item = items.value.find(n => n.id === id)
    if (item) {
      item.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  }

  async function markAllRead(wsId: number | string) {
    await notificationApi.markAllRead(wsId)
    items.value.forEach(n => { n.is_read = true })
    unreadCount.value = 0
  }

  async function archive(wsId: number | string, id: number) {
    await notificationApi.archive(wsId, id)
    items.value = items.value.filter(n => n.id !== id)
    await fetchUnreadCount(wsId)
  }

  function clear() {
    items.value = []
    unreadCount.value = 0
  }

  return { items, unreadCount, loading, fetchUnreadCount, fetchList, markRead, markAllRead, archive, clear }
})
