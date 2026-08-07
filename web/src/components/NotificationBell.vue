<!--
  通知铃铛组件 —— GHP P0-1d 通知中心下拉面板。

  功能：
   - 未读计数 badge（红色圆角，99+ 上限）
   - 点击展开下拉面板（380px 宽，最大 520px 高）
   - 最多 10 条通知，按时间倒序
   - 每条通知：领域图标 + 触发者 + 摘要 + 相对时间 + 已读/未读高亮
   - 点击通知 → 标记已读 + 跳转 action_url
   - "全部已读" 按钮（仅当未读数 > 0 时显示）
   - "查看全部" 链接 → /:slug/notifications
   - 空态 / 加载态友好提示
   - 点击外部关闭 + ESC 键关闭
   - 挂载或工作空间切换时拉取未读数
   - 每 30s 轻量轮询未读数（count-only 接口）
   - 主题感知（亮/暗）通过 CSS 变量
-->
<template>
  <div ref="bellRef" class="notification-bell">
    <!-- 铃铛按钮 -->
    <button
      class="bell-btn"
      :class="{ active: open }"
      :title="`${unreadCount} 条未读通知`"
      aria-label="通知"
      aria-haspopup="true"
      :aria-expanded="open"
      @click="toggle"
    >
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
        <path d="M13.73 21a2 2 0 0 1-3.46 0" />
      </svg>
      <span v-if="unreadCount > 0" class="badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
    </button>

    <!-- 下拉面板 -->
    <Transition name="dropdown">
      <div v-if="open" class="dropdown" role="menu" @click.stop>
        <!-- 头部：标题 + 全部已读 -->
        <div class="dropdown-header">
          <span class="title">通知</span>
          <button
            v-if="unreadCount > 0"
            class="mark-read-btn"
            :disabled="markingAll"
            @click="handleMarkAllRead"
          >
            {{ markingAll ? '处理中...' : '全部已读' }}
          </button>
        </div>

        <!-- 内容区 -->
        <div ref="scrollRef" class="dropdown-body">
          <!-- 加载中 -->
          <div v-if="loading" class="state-wrapper">
            <div class="spinner" />
            <span class="state-text">加载中...</span>
          </div>

          <!-- 空态 -->
          <div v-else-if="displayItems.length === 0" class="state-wrapper empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
            <span class="state-text">暂无通知</span>
            <span class="state-hint">新通知将在这里显示</span>
          </div>

          <!-- 通知列表 -->
          <div v-else class="notif-list">
            <div
              v-for="item in displayItems"
              :key="item.id"
              class="notif-item"
              :class="{ unread: !item.is_read }"
              role="menuitem"
              tabindex="0"
              @click="handleClick(item)"
              @keydown.enter="handleClick(item)"
            >
              <!-- 领域类型图标 -->
              <div class="notif-icon" :class="`icon-${item.entity_type}`">
                <component :is="getEventIcon(item.event_type)" />
              </div>

              <!-- 内容 -->
              <div class="notif-content">
                <div class="notif-line1">
                  <span class="notif-actor">{{ item.actor_name || '系统' }}</span>
                  <span class="notif-action">{{ getEventLabel(item.event_type) }}</span>
                </div>
                <div class="notif-title" :title="item.title">{{ truncate(item.title, 60) }}</div>
                <div v-if="item.body" class="notif-body">{{ truncate(item.body, 80) }}</div>
                <div class="notif-time">{{ formatRelativeTime(item.created_at) }}</div>
              </div>

              <!-- 操作区 -->
              <div class="notif-actions">
                <span v-if="!item.is_read" class="unread-dot" title="未读" />
                <button
                  class="notif-archive"
                  title="归档"
                  @click.stop="handleArchive(item.id)"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 底部：查看全部 -->
        <div class="dropdown-footer">
          <router-link :to="`/${wsStore.currentSlug}/notifications`" class="view-all-link" @click="open = false">
            查看全部通知
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9 18 15 12 9 6" />
            </svg>
          </router-link>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatRelativeTime } from '@/lib/formatTime'

// ===== 通知类型图标组件（内联 SVG 组件）=====

const IconIssue = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="6"/></svg>` }
const IconComment = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>` }
const IconAssign = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>` }
const IconStatus = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>` }
const IconSprint = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>` }
const IconVersion = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>` }
const IconMember = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>` }
const IconInvite = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>` }
const IconDefault = { template: `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>` }

/** 事件类型 → 图标组件映射 */
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

/** 事件类型 → 中文动作描述 */
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

// ===== 核心逻辑 =====

const router = useRouter()
const notifStore = useNotificationStore()
const wsStore = useWorkspaceStore()

const open = ref(false)
const bellRef = ref<HTMLElement | null>(null)
const scrollRef = ref<HTMLElement | null>(null)
const markingAll = ref(false)

/** 单钟下拉最多展示 10 条 */
const MAX_DROPDOWN_ITEMS = 10

const items = computed(() => notifStore.items)
const unreadCount = computed(() => notifStore.unreadCount)
const loading = computed(() => notifStore.loading)

/** 截断后的展示列表（最多 10 条） */
const displayItems = computed(() => items.value.slice(0, MAX_DROPDOWN_ITEMS))

/** 根据事件类型获取图标组件 */
function getEventIcon(eventType: string) {
  return eventIconMap[eventType] ?? IconDefault
}

/** 根据事件类型获取中文动作描述 */
function getEventLabel(eventType: string) {
  return eventLabelMap[eventType] ?? '触发了通知'
}

/** 截断文本 */
function truncate(text: string, maxLen: number): string {
  if (!text) return ''
  return text.length > maxLen ? text.slice(0, maxLen) + '...' : text
}

/** 切换面板开合；首次打开时拉取当前工作空间的通知列表 */
async function toggle() {
  open.value = !open.value
  if (open.value && wsStore.current) {
    await notifStore.fetchList(wsStore.current.id, { limit: MAX_DROPDOWN_ITEMS })
  }
}

/** 点击单条通知：未读则标记已读，有 action_url 则跳转，随后关闭面板 */
async function handleClick(item: any) {
  if (!item.is_read && wsStore.current) {
    await notifStore.markRead(wsStore.current.id, item.id)
  }
  if (item.action_url) {
    router.push(item.action_url)
  }
  open.value = false
}

/** 将当前工作空间全部通知标记为已读 */
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

/** 归档指定通知并从列表移除 */
async function handleArchive(id: number) {
  if (wsStore.current) {
    await notifStore.archive(wsStore.current.id, id)
  }
}

/** 点击组件外部区域时关闭下拉面板 */
function handleClickOutside(e: MouseEvent) {
  if (bellRef.value && !bellRef.value.contains(e.target as Node)) {
    open.value = false
  }
}

/** ESC 键关闭面板 */
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && open.value) {
    open.value = false
  }
}

// ===== 生命周期 =====

let pollingTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
  if (wsStore.current) {
    void notifStore.fetchUnreadCount(wsStore.current.id)
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
  if (pollingTimer) clearInterval(pollingTimer)
})

/* 工作空间切换：刷新未读数 + 30s 轮询 */
watch(
  () => wsStore.current?.id,
  (wsId) => {
    if (pollingTimer) {
      clearInterval(pollingTimer)
      pollingTimer = null
    }
    if (wsId != null) {
      void notifStore.fetchUnreadCount(wsId)
      pollingTimer = setInterval(() => {
        void notifStore.fetchUnreadCount(wsId)
      }, 30_000)
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.notification-bell {
  position: relative;
  display: inline-flex;
}

/* ===== 铃铛按钮 ===== */
.bell-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--text-tertiary);
  padding: 6px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  position: relative;
  transition: color 0.15s, background 0.15s;
}
.bell-btn:hover,
.bell-btn.active {
  color: var(--text-primary);
  background: var(--surface-3);
}
.bell-btn:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: 1px;
}

/* ===== 角标 ===== */
.badge {
  position: absolute;
  top: 0;
  right: 0;
  background: var(--danger-500);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  border-radius: 9px;
  padding: 0 4px;
  transform: translate(25%, -25%);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

/* ===== 下拉面板 ===== */
.dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  width: 380px;
  max-height: 520px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-popover);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 进入/离开动画 */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* ===== 头部 ===== */
.dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.title {
  font-weight: 600;
  font-size: 15px;
  color: var(--text-primary);
}

.mark-read-btn {
  background: none;
  border: none;
  color: var(--brand-500);
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  padding: 2px 4px;
  border-radius: var(--radius-sm);
  transition: background 0.1s;
}
.mark-read-btn:hover:not(:disabled) {
  background: var(--brand-50);
}
.mark-read-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ===== 内容体 ===== */
.dropdown-body {
  flex: 1;
  overflow-y: auto;
  min-height: 120px;
}

/* ===== 状态容器（loading / empty）===== */
.state-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  gap: 8px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--brand-500);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.state-text {
  color: var(--text-tertiary);
  font-size: 14px;
}

.state-hint {
  color: var(--text-placeholder);
  font-size: 12px;
}

.empty-icon {
  color: var(--text-placeholder);
  margin-bottom: 4px;
}

/* ===== 通知列表 ===== */
.notif-list {
  padding: 4px 0;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  position: relative;
  transition: background 0.1s;
}
.notif-item:last-child {
  border-bottom: none;
}
.notif-item:hover {
  background: var(--surface-2);
}
.notif-item:focus-visible {
  outline: 2px solid var(--brand-500);
  outline-offset: -2px;
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
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  margin-top: 2px;
}
/* 图标色盘 —— 按 entity_type 分类 */
.icon-issue { background: #e0e7ff; color: #4f46e5; }
.icon-comment { background: #fef3c7; color: #d97706; }
.icon-assign { background: #dbeafe; color: #2563eb; }
.icon-sprint { background: #d1fae5; color: #059669; }
.icon-version { background: #ede9fe; color: #7c3aed; }
.icon-project { background: #ffe4e6; color: #e11d48; }
.icon-workspace { background: #e0f2fe; color: #0284c7; }
.icon-member { background: #fce7f3; color: #db2777; }
/* 暗主题下图标保持原色（确保可读） */
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
  gap: 1px;
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
  max-width: 80px;
}

.notif-action {
  color: var(--text-tertiary);
  font-size: 12px;
  white-space: nowrap;
}

.notif-title {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.notif-body {
  font-size: 12px;
  color: var(--text-tertiary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

.notif-time {
  font-size: 11px;
  color: var(--text-placeholder);
  margin-top: 2px;
}

/* ===== 操作区 ===== */
.notif-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
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
  padding: 2px;
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

/* ===== 底部 ===== */
.dropdown-footer {
  flex-shrink: 0;
  padding: 8px 16px;
  border-top: 1px solid var(--border-subtle);
  text-align: center;
}

.view-all-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--brand-500);
  font-size: 13px;
  cursor: pointer;
  text-decoration: none;
  font-family: inherit;
  transition: color 0.15s;
}
.view-all-link:hover {
  color: var(--brand-600);
  text-decoration: underline;
}

/* ===== 窄屏适配 ===== */
@media (max-width: 480px) {
  .dropdown {
    width: calc(100vw - 16px);
    right: -8px;
  }
}
</style>
