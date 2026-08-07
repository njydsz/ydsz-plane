<!--
  全局搜索弹层组件。
  通过 Ctrl/Cmd+K 快捷键唤起，在工作空间内对工作项/迭代/版本执行全文搜索，
  支持键盘上下键导航选中项、Enter 打开、ESC 关闭，结果按实体类型分组展示。
  搜索请求做 200ms 防抖，避免每次按键都触发后端请求。
  注：highlight 由服务端 PostgreSQL ts_headline 生成（内容已转义，仅包裹 <b>），v-html 属受信输出。
-->
<template>
  <Teleport to="body">
    <div v-if="open" class="search-overlay" @click.self="close">
      <div class="search-dialog">
        <div class="search-input-wrap">
          <svg class="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" /> <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            ref="inputRef"
            v-model="q"
            class="search-input"
            placeholder="搜索工作项、迭代、版本... (Ctrl+K)"
            @keydown.esc="close"
            @keydown.enter="handleEnter"
            @keydown.down.prevent="moveDown"
            @keydown.up.prevent="moveUp"
            @input="handleInput"
          />
          <kbd class="shortcut-hint">ESC</kbd>
        </div>
        <div v-if="q.trim()" class="search-body">
          <div v-if="store.loading" class="search-status">搜索中...</div>
          <template v-else-if="store.results">
            <!-- issues -->
            <div v-if="store.results.results.issues?.length" class="result-group">
              <div class="group-title">工作项</div>
              <div
                v-for="(item, idx) in store.results.results.issues"
                :key="item.id"
                class="result-item"
                :class="{ selected: selectedIdx === idx }"
                @click="goIssue(item)"
              >
                <span class="ri-type">#</span>
                <span class="ri-name" v-html="item.highlight || item.name"></span>
                <span class="ri-meta">{{ item.project_name }}</span>
              </div>
            </div>
            <!-- sprints -->
            <div v-if="store.results.results.sprints?.length" class="result-group">
              <div class="group-title">迭代</div>
              <div
                v-for="(item, idx) in store.results.results.sprints"
                :key="item.id"
                class="result-item"
                :class="{ selected: selectedIdx === (store.results.results.issues?.length || 0) + idx }"
                @click="goSprint(item)"
              >
                <span class="ri-type type-sprint">🏃</span>
                <span class="ri-name" v-html="item.highlight || item.name"></span>
                <span class="ri-meta">{{ item.project_name }}</span>
              </div>
            </div>
            <!-- versions -->
            <div v-if="store.results.results.versions?.length" class="result-group">
              <div class="group-title">版本</div>
              <div
                v-for="item in store.results.results.versions"
                :key="item.id"
                class="result-item"
                @click="goVersion(item)"
              >
                <span class="ri-type type-version">🚀</span>
                <span class="ri-name" v-html="item.highlight || item.name"></span>
                <span class="ri-meta">{{ item.project_name }}</span>
              </div>
            </div>
            <div v-if="isEmpty" class="search-status">未找到结果</div>
          </template>
        </div>
        <div class="search-footer">
          <span>↑↓ 导航</span>
          <span>↵ 打开</span>
          <span>ESC 关闭</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/* eslint-disable vue/no-v-html -- highlight 由服务端 PostgreSQL ts_headline 生成（内容已转义，仅包裹 <b> 标签），属受信输出，无 XSS 注入面。 */
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSearchStore } from '@/stores/search'
import { useWorkspaceStore } from '@/stores/workspace'

const router = useRouter()
const store = useSearchStore()
const wsStore = useWorkspaceStore()

const open = computed(() => store.open)
const q = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const selectedIdx = ref(-1)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

/** 结果是否为空（用于展示"未找到结果"提示） */
const isEmpty = computed(() => {
  const r = store.results
  if (!r) return true
  return (!r.results.issues?.length) && (!r.results.sprints?.length) && (!r.results.versions?.length)
})

/** 输入变化处理：重置选中索引，并对搜索请求做 200ms 防抖 */
function handleInput() {
  selectedIdx.value = -1
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (q.value.trim() && wsStore.current) {
      store.search(wsStore.current.id, q.value)
    }
  }, 200)
}

/** 回车键处理：若当前有选中项则打开它，否则保留默认行为 */
function handleEnter() {
  // 如果有选中项，打开它
  const allItems = getAllItems()
  if (selectedIdx.value >= 0 && selectedIdx.value < allItems.length) {
    const item = allItems[selectedIdx.value]
    navigateTo(item)
  }
  // 否则默认行为：跳转搜索结果页
}

/** 键盘 ↓：将选中索引下移一格（封顶到列表末尾） */
function moveDown() {
  const allItems = getAllItems()
  if (allItems.length === 0) return
  selectedIdx.value = Math.min(selectedIdx.value + 1, allItems.length - 1)
}

/** 键盘 ↑：将选中索引上移一格（下限为 -1，即无选中） */
function moveUp() {
  selectedIdx.value = Math.max(selectedIdx.value - 1, -1)
}

/** 将三类实体的结果合并为统一的扁平列表，并打上 _type 标记以便导航 */
function getAllItems(): any[] {
  const r = store.results
  if (!r) return []
  return [
    ...(r.results.issues || []).map(i => ({ ...i, _type: 'issue' })),
    ...(r.results.sprints || []).map(i => ({ ...i, _type: 'sprint' })),
    ...(r.results.versions || []).map(i => ({ ...i, _type: 'version' })),
  ]
}

/** 根据结果实体类型跳转到对应详情页，并关闭搜索面板 */
function navigateTo(item: any) {
  const slug = wsStore.currentSlug
  if (item._type === 'issue') {
    router.push(`/${slug}/projects/${item.project_id}/issues/${item.id}`)
  } else if (item._type === 'sprint') {
    router.push(`/${slug}/projects/${item.project_id}/sprints/${item.id}`)
  } else if (item._type === 'version') {
    router.push(`/${slug}/projects/${item.project_id}/versions/${item.id}`)
  }
  close()
}

/** 点击工作项结果跳转 */
function goIssue(item: any) { navigateTo({ ...item, _type: 'issue' }) }
/** 点击迭代结果跳转 */
function goSprint(item: any) { navigateTo({ ...item, _type: 'sprint' }) }
/** 点击版本结果跳转 */
function goVersion(item: any) { navigateTo({ ...item, _type: 'version' }) }

/** 关闭搜索面板并重置输入与选中状态 */
function close() {
  store.clear()
  q.value = ''
  selectedIdx.value = -1
}

/** 面板打开时自动聚焦输入框 */
watch(open, (val) => {
  if (val) {
    nextTick(() => inputRef.value?.focus())
  }
})

/** 全局 Ctrl/Cmd+K 快捷键处理：切换搜索面板开关 */
function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    store.toggle()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.search-overlay {
  position: fixed; inset: 0; z-index: 9999;
  display: flex; align-items: flex-start; justify-content: center;
  padding-top: 15vh;
  background: rgba(15, 23, 42, 0.4);
  backdrop-filter: blur(2px);
}
.search-dialog {
  width: 560px; max-width: 90vw;
  background: white; border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
  overflow: hidden;
}
.search-input-wrap {
  display: flex; align-items: center;
  padding: 14px 16px; border-bottom: 1px solid #f1f5f9;
  gap: 10px;
}
.search-icon { color: #94a3b8; flex-shrink: 0; }
.search-input {
  flex: 1; border: none; outline: none; font-size: 16px;
  color: #1e293b; background: transparent;
}
.search-input::placeholder { color: #cbd5e1; }
.shortcut-hint {
  font-size: 10px; padding: 2px 6px; border-radius: 4px;
  background: #f1f5f9; color: #94a3b8; font-family: monospace;
}
.search-body { max-height: 360px; overflow-y: auto; }
.search-status {
  padding: 24px; text-align: center; color: #94a3b8; font-size: 14px;
}
.result-group { padding: 4px 0; }
.group-title {
  padding: 6px 16px; font-size: 11px; font-weight: 600;
  color: #94a3b8; text-transform: uppercase; letter-spacing: 0.5px;
}
.result-item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 16px; cursor: pointer; font-size: 14px;
}
.result-item:hover, .result-item.selected { background: #f8fafc; }
.ri-type {
  font-size: 14px; color: #94a3b8; width: 20px; text-align: center;
}
.ri-name { flex: 1; color: #1e293b; }
.ri-meta { font-size: 12px; color: #cbd5e1; }
.search-footer {
  display: flex; gap: 16px; padding: 8px 16px;
  border-top: 1px solid #f1f5f9;
  font-size: 11px; color: #cbd5e1;
}
</style>
