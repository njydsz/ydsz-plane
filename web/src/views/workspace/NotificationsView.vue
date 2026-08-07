/**
 * NotificationsView — 通知中心全屏页。
 * 当前工作空间全部通知列表（分页加载），支持筛选、单条归档、全部已读操作。
 */
<template>
    <!-- 顶部操作栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <h1 class="page-title">通知中心</h1>
        <span v-if="unreadCount > 0" class="unread-badge">{{ unreadCount }} 条未读</span>
      </div>
      <div class="toolbar-right">
        <!-- 筛选 -->
        <div class="filter-group">
          <button
            class="filter-btn"
            :class="{ active: filter === 'all' }"
            @click="filter = 'all'"
          >
            全部
          </button>
          <button
            class="filter-btn"
            :class="{ active: filter === 'unread' }"
            @click="filter = 'unread'"
          >
            未读
          </button>
        </div>
        <!-- 全部已读 -->
        <button
          v-if="unreadCount > 0"
          class="mark-all-btn"
          :disabled="markingAll"
          @click="handleMarkAllRead"
        >
          {{ markingAll ? '处理中...' : '全部已读' }}
        </button>
      </div>
    </div>

    <!-- 加载中 -->
    <AppLoadingState v-if="loading && items.length === 0" />

    <!-- 错误态 -->
    <AppErrorState
      v-else-if="notifStore.error"
      :message="notifStore.error"
      @retry="reload"
    />

    <!-- 空态 -->
    <AppEmptyState
      v-else-if="filteredItems.length === 0"
      icon="🔔"
      :title="filter === 'unread' ? '没有未读通知' : '暂无通知'"
      description="新通知将在这里显示"
    />

    <!-- 通知列表 -->
    <div v-else class="notif-list">
      <div
        v-for="item in filteredItems"
        :key="item.id"
        class="notif-item"
        :class="{ unread: !item.is_read }"
        @click="handleClick(item)"
      >
        <!-- 领域图标 -->
        <div class="notif-icon" :class="`icon-${item.entity_type}`">
          <component :is="getEventIcon(item.event_type)" />
        </div>

        <!-- 内容 -->
        <div class="notif-content">
          <div class="notif-line1">
            <span class="notif-actor">{{ item.actor_name || '系统' }}</span>
            <span class="notif-action">{{ getEventLabel(item.event_type) }}</span>
          </div>
          <div class="notif-title">{{ item.title }}</div>
          <div v-if="item.body" class="notif-body">{{ truncate(item.body, 120) }}</div>
          <div class="notif-time">{{ formatRelativeTime(item.created_at) }}</div>
        </div>

        <!-- 操作区 -->
        <div class="notif-actions">
          <span v-if="!item.is_read" class="unread-dot" title="未读" />
          <button class="notif-archive" title="归档" @click.stop="handleArchive(item.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
            </svg>
          </button>
        </div>
      </div>

      <!-- 加载更多 -->
      <div v-if="hasMore" class="load-more">
        <button class="load-more-btn" :disabled="loadingMore" @click="loadMore">
          {{ loadingMore ? '加载中...' : '加载更多' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatRelativeTime } from '@/lib/formatTime'
import type { AppNotification } from '@/api/services/notification'

// ===== 图标组件 =====
const IconIssue = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="6"/></svg>` }
const IconComment = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>` }
const IconAssign = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>` }
const IconStatus = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>` }
const IconSprint = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>` }
const IconVersion = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>` }
const IconMember = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>` }
const IconInvite = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>` }
const IconDefault = { template: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>` }

const eventIconMap: Record<string, any> = {
  'issue.created': IconIssue,
  'issue.assigned': IconAssign,
  'issue.status_changed': IconStatus,
  'issue.deleted': IconDefault,
  'comment.created': IconComment,
  'sprint.started': IconSprint,
  'sprint.completed': IconSprint,
  'version.released': IconVersion,
  'member.added': IconMember,
  'member.removed': IconMember,
  'member.role_changed': IconMember,
  'invitation.sent': IconInvite,
}

const eventLabelMap: Record<string, string> = {
  'issue.created': '创建了工作项',
  'issue.assigned': '将工作项分配给你',
  'issue.status_changed': '变更了工作项状态',
  'issue.deleted': '删除了工作项',
  'comment.created': '评论了工作项',
  'sprint.started': '启动了迭代',
  'sprint.completed': '完成了迭代',
  'version.released': '发布了版本',
  'member.added': '添加了成员',
  'member.removed': '移除了成员',
  'member.role_changed': '变更了成员角色',
  'invitation.sent': '发送了邀请',
}

// ===== 状态 =====
const router = useRouter()
const notifStore = useNotificationStore()
const wsStore = useWorkspaceStore()

const filter = ref<'all' | 'unread'>('all')
const markingAll = ref(false)
const loadingMore = ref(false)
const offset = ref(0)
const PAGE_SIZE = 20

const items = computed(() => notifStore.items)
const unreadCount = computed(() => notifStore.unreadCount)
const loading = computed(() => notifStore.loading)

/** 筛选后列表 */
const filteredItems = computed(() => {
  if (filter.value === 'unread') {
    return items.value.filter((n: AppNotification) => !n.is_read)
  }
  return items.value
})

/** 是否有更多可加载 */
const hasMore = computed(() => items.value.length > 0 && items.value.length % PAGE_SIZE === 0 && items.value.length === offset.value + PAGE_SIZE)

function getEventIcon(eventType: string) {
  return eventIconMap[eventType] ?? IconDefault
}

function getEventLabel(eventType: string) {
  return eventLabelMap[eventType] ?? '触发了通知'
}

function truncate(text: string, maxLen: number): string {
  if (!text) return ''
  return text.length > maxLen ? text.slice(0, maxLen) + '...' : text
}

/** 点击通知：标记已读 + 跳转 */
async function handleClick(item: AppNotification) {
  if (!item.is_read && wsStore.current) {
    await notifStore.markRead(wsStore.current.id, item.id)
  }
  if (item.action_url) {
    router.push(item.action_url)
  }
}

/** 全部已读 */
async function handleMarkAllRead() {
  if (wsStore.current) {
    markingAll.value = true
    try {
      await notifStore.markAllRead(wsStore.current.id)
    } finally {
      markingAll.value = false
    }
  }
}

/** 归档 */
async function handleArchive(id: number) {
  if (wsStore.current) {
    await notifStore.archive(wsStore.current.id, id)
  }
}

/** 加载更多 */
async function loadMore() {
  if (!wsStore.current || loadingMore.value) return
  loadingMore.value = true
  offset.value += PAGE_SIZE
  try {
    await notifStore.fetchList(wsStore.current.id, { limit: PAGE_SIZE })
  } finally {
    loadingMore.value = false
  }
}

/** 重新加载 */
async function reload() {
  offset.value = 0
  if (wsStore.current) {
    await notifStore.fetchList(wsStore.current.id, { limit: PAGE_SIZE })
  }
}

onMounted(reload)

watch(() => wsStore.current?.id, () => {
  reload()
})
</script>

<style scoped>
.notifications-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 8px 0;
}

/* ===== 工具栏 ===== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-subtle);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.unread-badge {
  font-size: 12px;
  font-weight: 500;
  color: var(--brand-500);
  background: var(--brand-50);
  padding: 2px 8px;
  border-radius: 10px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-group {
  display: flex;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.filter-btn {
  background: var(--surface-1);
  border: none;
  color: var(--text-secondary);
  font-size: 13px;
  padding: 5px 12px;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.1s, color 0.1s;
}
.filter-btn:hover {
  background: var(--surface-2);
}
.filter-btn.active {
  background: var(--brand-500);
  color: #fff;
}

.mark-all-btn {
  background: var(--brand-500);
  color: #fff;
  border: none;
  border-radius: var(--radius-sm);
  padding: 5px 12px;
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s, opacity 0.15s;
}
.mark-all-btn:hover:not(:disabled) {
  background: var(--brand-600);
}
.mark-all-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== 状态容器 ===== */
.state-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  gap: 8px;
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.state-text {
  color: var(--text-tertiary);
  font-size: 15px;
}

.state-hint {
  color: var(--text-placeholder);
  font-size: 13px;
}

.empty-icon {
  color: var(--text-placeholder);
  margin-bottom: 8px;
}

/* ===== 通知列表 ===== */
.notif-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 12px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background 0.1s;
}
.notif-item:hover {
  background: var(--surface-2);
}
.notif-item.unread {
  background: var(--brand-50);
}
.notif-item.unread:hover {
  background: var(--surface-3);
}

/* ===== 领域图标 ===== */
.notif-icon {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 2px;
}
.icon-issue { background: #e0e7ff; color: #4f46e5; }
.icon-comment { background: #fef3c7; color: #d97706; }
.icon-assign { background: #dbeafe; color: #2563eb; }
.icon-sprint { background: #d1fae5; color: #059669; }
.icon-version { background: #ede9fe; color: #7c3aed; }
.icon-project { background: #ffe4e6; color: #e11d48; }
.icon-workspace { background: #e0f2fe; color: #0284c7; }
.icon-member { background: #fce7f3; color: #db2777; }
[data-theme="dark"] .icon-issue { background: #312e81; color: #a5b4fc; }
[data-theme="dark"] .icon-comment { background: #78350f; color: #fcd34d; }
[data-theme="dark"] .icon-assign { background: #1e3a8a; color: #93c5fd; }
[data-theme="dark"] .icon-sprint { background: #064e3b; color: #6ee7b7; }
[data-theme="dark"] .icon-version { background: #4c1d95; color: #c4b5fd; }
[data-theme="dark"] .icon-project { background: #881337; color: #fda4af; }
[data-theme="dark"] .icon-workspace { background: #0c4a6e; color: #7dd3fc; }
[data-theme="dark"] .icon-member { background: #831843; color: #f9a8d4; }

/* ===== 内容区 ===== */
.notif-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.notif-line1 {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
}

.notif-actor {
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 120px;
}

.notif-action {
  color: var(--text-tertiary);
  font-size: 12px;
}

.notif-title {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.5;
}

.notif-body {
  font-size: 13px;
  color: var(--text-tertiary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.notif-time {
  font-size: 12px;
  color: var(--text-placeholder);
  margin-top: 2px;
}

/* ===== 操作区 ===== */
.notif-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--brand-500);
  flex-shrink: 0;
}

.notif-archive {
  background: none;
  border: none;
  color: var(--text-placeholder);
  cursor: pointer;
  padding: 4px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.15s, color 0.15s, background 0.15s;
}
.notif-item:hover .notif-archive {
  opacity: 1;
}
.notif-archive:hover {
  color: var(--danger-500);
  background: var(--surface-3);
}

/* ===== 加载更多 ===== */
.load-more {
  display: flex;
  justify-content: center;
  padding: 16px;
}

.load-more-btn {
  background: var(--surface-2);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  font-size: 13px;
  padding: 8px 24px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-family: inherit;
  transition: background 0.1s, color 0.1s;
}
.load-more-btn:hover:not(:disabled) {
  background: var(--surface-3);
  color: var(--text-primary);
}
.load-more-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== 窄屏 ===== */
@media (max-width: 640px) {
  .notifications-page {
    padding: 8px;
  }
  .toolbar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
