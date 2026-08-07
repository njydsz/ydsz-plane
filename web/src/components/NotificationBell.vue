<template>
  <div class="notification-bell" ref="bellRef">
    <button class="bell-btn" @click="toggle" :class="{ active: open }" :title="`${unreadCount} 条未读通知`">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
        <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
      </svg>
      <span v-if="unreadCount > 0" class="badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>

    <div v-if="open" class="dropdown" @click.stop>
      <div class="dropdown-header">
        <span class="title">通知</span>
        <button v-if="unreadCount > 0" class="mark-read-btn" @click="handleMarkAllRead">全部已读</button>
      </div>
      <div class="dropdown-body">
        <div v-if="loading" class="state-text">加载中...</div>
        <div v-else-if="items.length === 0" class="state-text">暂无通知</div>
        <div v-else class="notif-list">
          <div
            v-for="item in items"
            :key="item.id"
            class="notif-item"
            :class="{ unread: !item.is_read }"
            @click="handleClick(item)"
          >
            <div class="notif-actor">{{ item.actor_name || '系统' }}</div>
            <div class="notif-title">{{ item.title }}</div>
            <div class="notif-body" v-if="item.body">{{ item.body }}</div>
            <div class="notif-time">{{ formatTime(item.created_at) }}</div>
            <button class="notif-archive" @click.stop="handleArchive(item.id)" title="归档">×</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import { useWorkspaceStore } from '@/stores/workspace'

const router = useRouter()
const notifStore = useNotificationStore()
const wsStore = useWorkspaceStore()

const open = ref(false)
const bellRef = ref<HTMLElement | null>(null)

const items = computed(() => notifStore.items)
const unreadCount = computed(() => notifStore.unreadCount)
const loading = computed(() => notifStore.loading)

function toggle() {
  open.value = !open.value
  if (open.value && wsStore.current) {
    notifStore.fetchList(wsStore.current.id)
  }
}

function handleClick(item: any) {
  if (!item.is_read && wsStore.current) {
    notifStore.markRead(wsStore.current.id, item.id)
  }
  if (item.action_url) {
    router.push(item.action_url)
  }
  open.value = false
}

function handleMarkAllRead() {
  if (wsStore.current) {
    notifStore.markAllRead(wsStore.current.id)
  }
}

function handleArchive(id: number) {
  if (wsStore.current) {
    notifStore.archive(wsStore.current.id, id)
  }
}

function formatTime(ts: string): string {
  const d = new Date(ts)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return d.toLocaleDateString('zh-CN')
}

function handleClickOutside(e: MouseEvent) {
  if (bellRef.value && !bellRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (wsStore.current) {
    notifStore.fetchUnreadCount(wsStore.current.id)
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

watch(() => wsStore.current, (ws) => {
  if (ws) {
    notifStore.fetchUnreadCount(ws.id)
    // 每 30s 轮询未读数
    const timer = setInterval(() => notifStore.fetchUnreadCount(ws.id), 30000)
    onUnmounted(() => clearInterval(timer))
  }
})
</script>

<style scoped>
.notification-bell { position: relative; }
.bell-btn {
  background: none; border: none; cursor: pointer;
  color: #64748b; padding: 6px; border-radius: 6px;
  display: flex; align-items: center; position: relative;
}
.bell-btn:hover, .bell-btn.active { color: #1e293b; background: #f1f5f9; }
.badge {
  position: absolute; top: 0; right: 0;
  background: #ef4444; color: white; font-size: 10px; font-weight: 700;
  min-width: 18px; height: 18px; line-height: 18px; text-align: center;
  border-radius: 9px; padding: 0 4px;
  transform: translate(25%, -25%);
}
.dropdown {
  position: absolute; top: 100%; right: 0; margin-top: 8px;
  width: 380px; max-height: 480px;
  background: white; border: 1px solid #e2e8f0; border-radius: 8px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.1); z-index: 1000;
  display: flex; flex-direction: column;
}
.dropdown-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; border-bottom: 1px solid #f1f5f9;
}
.title { font-weight: 600; font-size: 15px; }
.mark-read-btn { background: none; border: none; color: #3b82f6; font-size: 13px; cursor: pointer; }
.mark-read-btn:hover { text-decoration: underline; }
.dropdown-body { flex: 1; overflow-y: auto; }
.state-text { padding: 32px 16px; text-align: center; color: #94a3b8; font-size: 14px; }
.notif-list { padding: 4px 0; }
.notif-item {
  padding: 10px 16px; border-bottom: 1px solid #f8fafc;
  cursor: pointer; position: relative;
}
.notif-item:hover { background: #f8fafc; }
.notif-item.unread { background: #eff6ff; }
.notif-item.unread:hover { background: #dbeafe; }
.notif-actor { font-size: 12px; color: #64748b; margin-bottom: 2px; }
.notif-title { font-size: 14px; color: #1e293b; font-weight: 500; }
.notif-body { font-size: 13px; color: #64748b; margin-top: 2px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.notif-time { font-size: 11px; color: #94a3b8; margin-top: 4px; }
.notif-archive {
  position: absolute; top: 8px; right: 12px; background: none; border: none;
  color: #cbd5e1; font-size: 16px; cursor: pointer; padding: 0; line-height: 1;
}
.notif-archive:hover { color: #ef4444; }
</style>
