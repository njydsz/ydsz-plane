<template>
  <div class="workbench">
    <div class="wb-header">
      <h2>工作台</h2>
      <span class="wb-greeting">{{ greeting }}</span>
    </div>

    <AppLoadingState v-if="loading" text="正在加载工作台..." />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <!-- 快捷操作 -->
      <section class="wb-section">
        <h3>快捷操作</h3>
        <div class="quick-actions">
          <button v-for="act in quickActions" :key="act.type" class="qa-btn" @click="handleAction(act)">
            <span class="qa-icon">{{ act.icon }}</span>
            <span class="qa-label">{{ act.label }}</span>
          </button>
        </div>
      </section>

      <div class="wb-grid">
        <!-- 我的任务 -->
        <section class="wb-section">
          <h3>我的任务</h3>
          <div class="task-section">
            <div class="task-group" v-if="overdueTasks.length">
              <div class="task-group-title overdue">逾期 ({{ overdueTasks.length }})</div>
              <div v-for="t in overdueTasks.slice(0,5)" :key="t.id" class="task-item" @click="goIssue(t.project_id, t.sequence_id)">
                <span class="task-id">#{{ t.sequence_id }}</span>
                <span class="task-name">{{ t.name }}</span>
                <span class="task-meta">{{ t.project_name }}</span>
              </div>
            </div>
            <div class="task-group" v-if="todayTasks.length">
              <div class="task-group-title today">今日 ({{ todayTasks.length }})</div>
              <div v-for="t in todayTasks.slice(0,5)" :key="t.id" class="task-item" @click="goIssue(t.project_id, t.sequence_id)">
                <span class="task-id">#{{ t.sequence_id }}</span>
                <span class="task-name">{{ t.name }}</span>
                <span class="task-meta">{{ t.project_name }}</span>
              </div>
            </div>
            <div class="task-group" v-if="inProgressTasks.length">
              <div class="task-group-title in-progress">进行中 ({{ inProgressTasks.length }})</div>
              <div v-for="t in inProgressTasks.slice(0,5)" :key="t.id" class="task-item" @click="goIssue(t.project_id, t.sequence_id)">
                <span class="task-id">#{{ t.sequence_id }}</span>
                <span class="task-name">{{ t.name }}</span>
                <span class="task-meta">{{ t.project_name }}</span>
              </div>
            </div>
          </div>
          <AppEmptyState v-if="noTasks" text="暂无待处理任务" />
        </section>

        <!-- 迭代概览 -->
        <section class="wb-section">
          <h3>迭代概览</h3>
          <AppEmptyState v-if="!currentSprint && !nextSprint" text="暂无迭代数据" />
          <div v-else class="sprint-cards">
            <div v-if="currentSprint" class="sprint-card active">
              <div class="sc-status active">进行中</div>
              <div class="sc-name">{{ currentSprint.name }}</div>
              <div class="sc-project">{{ currentSprint.project_name }}</div>
              <ProgressBar :percent="currentSprint.progress" />
              <div class="sc-stats">{{ currentSprint.completed_count }}/{{ currentSprint.issue_count }} 已完成</div>
              <router-link :to="`/${wsSlug}/projects/${currentSprint.project_id}/sprints/${currentSprint.id}`" class="sc-link">查看详情 →</router-link>
            </div>
            <div v-if="nextSprint" class="sprint-card">
              <div class="sc-status planned">规划中</div>
              <div class="sc-name">{{ nextSprint.name }}</div>
              <div class="sc-project">{{ nextSprint.project_name }}</div>
              <div class="sc-date" v-if="nextSprint.start_date">{{ formatDate(nextSprint.start_date) }} 开始</div>
            </div>
          </div>
        </section>
      </div>

      <!-- 最近访问 -->
      <section class="wb-section" v-if="recentItems.length">
        <h3>最近访问</h3>
        <div class="recent-list">
          <div v-for="item in recentItems.slice(0,10)" :key="item.entity_id" class="recent-item" @click="goRecent(item)">
            <span class="ri-type">{{ typeLabel(item.entity_type) }}</span>
            <span class="ri-name">{{ item.name }}</span>
            <span class="ri-meta">{{ item.project_name }}</span>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { workbenchApi, type WorkbenchSummary, type QuickAction } from '@/api/services/workbench'
import { useWorkspaceStore } from '@/stores/workspace'
import AppLoadingState from '@/components/AppLoadingState.vue'
import AppErrorState from '@/components/AppErrorState.vue'
import AppEmptyState from '@/components/AppEmptyState.vue'
import ProgressBar from '@/components/ProgressBar.vue'

const router = useRouter()
const wsStore = useWorkspaceStore()

const loading = ref(true)
const error = ref<string | null>(null)
const summary = ref<WorkbenchSummary | null>(null)

const wsSlug = computed(() => wsStore.currentSlug)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return '早上好 ☀️'
  if (h < 18) return '下午好 🌤'
  return '晚上好 🌙'
})

const quickActions = computed<QuickAction[]>(() => summary.value?.quick_actions ?? [
  { type: 'create_issue', label: '新建工作项', icon: '➕', route: `/${wsSlug.value}/projects` },
  { type: 'create_sprint', label: '新建迭代', icon: '🏃', route: `/${wsSlug.value}/projects` },
])

const overdueTasks = computed(() => summary.value?.my_tasks?.overdue ?? [])
const todayTasks = computed(() => summary.value?.my_tasks?.today ?? [])
const inProgressTasks = computed(() => summary.value?.my_tasks?.in_progress ?? [])
const noTasks = computed(() => overdueTasks.value.length === 0 && todayTasks.value.length === 0 && inProgressTasks.value.length === 0)

const currentSprint = computed(() => summary.value?.iteration_overview?.current ?? null)
const nextSprint = computed(() => summary.value?.iteration_overview?.next ?? null)
const recentItems = computed(() => summary.value?.recent_items ?? [])

async function load() {
  loading.value = true
  error.value = null
  try {
    summary.value = await workbenchApi.getSummary(wsStore.current!.id)
  } catch (e: any) {
    error.value = e?.message || '加载失败'
  } finally {
    loading.value = false
  }
}

function goIssue(projectId: number, seqId: number) {
  router.push(`/${wsSlug.value}/projects/${projectId}/issues/${seqId}`)
}

function goRecent(item: any) {
  if (item.entity_type === 'issue') {
    router.push(`/${wsSlug.value}/projects/${item.project_id}/issues/${item.entity_id}`)
  }
}

function handleAction(act: QuickAction) {
  if (act.route) router.push(act.route)
}

function typeLabel(t: string): string {
  const map: Record<string, string> = { issue: '工作项', sprint: '迭代', version: '版本' }
  return map[t] || t
}

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

onMounted(() => {
  if (wsStore.current) load()
})
</script>

<style scoped>
.workbench { max-width: 960px; margin: 0 auto; padding: 24px; }
.wb-header { margin-bottom: 24px; display: flex; align-items: baseline; gap: 16px; }
.wb-header h2 { font-size: 22px; font-weight: 700; color: #1e293b; margin: 0; }
.wb-greeting { color: #94a3b8; font-size: 14px; }
.wb-section { margin-bottom: 24px; }
.wb-section h3 { font-size: 14px; font-weight: 600; color: #64748b; text-transform: uppercase; margin-bottom: 12px; letter-spacing: 0.5px; }
.wb-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
@media (max-width: 768px) { .wb-grid { grid-template-columns: 1fr; } }

.quick-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.qa-btn { display: flex; align-items: center; gap: 6px; padding: 8px 14px; background: white; border: 1px solid #e2e8f0; border-radius: 8px; cursor: pointer; font-size: 14px; color: #334155; }
.qa-btn:hover { background: #f8fafc; border-color: #cbd5e1; }
.qa-icon { font-size: 16px; }
.qa-label { font-weight: 500; }

.task-group { margin-bottom: 12px; }
.task-group-title { font-size: 12px; font-weight: 600; padding: 2px 8px; border-radius: 4px; display: inline-block; margin-bottom: 6px; }
.task-group-title.overdue { color: #dc2626; background: #fef2f2; }
.task-group-title.today { color: #d97706; background: #fffbeb; }
.task-group-title.in-progress { color: #2563eb; background: #eff6ff; }

.task-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 6px; cursor: pointer; font-size: 14px; }
.task-item:hover { background: #f8fafc; }
.task-id { color: #94a3b8; font-family: monospace; font-size: 12px; min-width: 28px; }
.task-name { flex: 1; color: #1e293b; }
.task-meta { color: #94a3b8; font-size: 12px; }

.sprint-cards { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.sprint-card { background: white; border: 1px solid #e2e8f0; border-radius: 8px; padding: 14px; }
.sprint-card.active { border-color: #bfdbfe; background: #f8fafc; }
.sc-status { font-size: 11px; font-weight: 600; border-radius: 4px; padding: 1px 8px; display: inline-block; margin-bottom: 6px; }
.sc-status.active { background: #dbeafe; color: #2563eb; }
.sc-status.planned { background: #f1f5f9; color: #64748b; }
.sc-name { font-size: 15px; font-weight: 600; color: #1e293b; margin-bottom: 4px; }
.sc-project { font-size: 12px; color: #94a3b8; margin-bottom: 8px; }
.sc-stats { font-size: 12px; color: #64748b; margin-top: 6px; }
.sc-link { font-size: 13px; color: #3b82f6; text-decoration: none; margin-top: 6px; display: inline-block; }
.sc-date { font-size: 13px; color: #64748b; }

.recent-list { display: flex; flex-wrap: wrap; gap: 8px; }
.recent-item { display: flex; align-items: center; gap: 6px; padding: 6px 12px; background: white; border: 1px solid #e2e8f0; border-radius: 6px; cursor: pointer; font-size: 13px; }
.recent-item:hover { background: #f8fafc; }
.ri-type { font-size: 11px; color: #94a3b8; background: #f1f5f9; padding: 1px 6px; border-radius: 3px; }
.ri-name { color: #1e293b; }
.ri-meta { color: #94a3b8; font-size: 11px; }
</style>
