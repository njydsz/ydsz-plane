/**
 * CommandPalette (⌘K) — 全局命令面板。
 * 提供 Jump / Create / Search 三类命令源，支持模糊过滤、键盘导航、分组展示。
 */
<template>
  <Teleport to="body">
    <Transition name="cp-overlay">
      <div v-if="open" class="cp-overlay" @click.self="close">
        <div class="cp-dialog" role="dialog" aria-modal="true" aria-label="命令面板">
          <!-- 输入区 -->
          <div class="cp-input-wrap">
            <kbd class="cp-input-kbd">⌘K</kbd>
            <svg class="cp-search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            <input
              ref="inputRef"
              v-model="query"
              class="cp-input"
              :placeholder="mode === 'search' ? '搜索工作项、迭代、版本...' : '输入命令或搜索...'"
              @keydown.esc="close"
              @keydown.enter="handleEnter"
              @keydown.down.prevent="moveDown"
              @keydown.up.prevent="moveUp"
              @input="handleInput"
            />
            <kbd class="cp-input-kbd cp-input-kbd--esc">ESC</kbd>
          </div>

          <!-- 主体 -->
          <div class="cp-body">
            <!-- 命令模式 (无搜索词) -->
            <template v-if="!query.trim() && mode === 'command'">
              <div v-for="group in commandGroups" :key="group.title" class="cp-group">
                <div class="cp-group-title">{{ group.title }}</div>
                <div
                  v-for="(cmd, idx) in group.commands"
                  :key="cmd.id"
                  class="cp-item"
                  :class="{ 'cp-item--selected': isSelected(group, idx) }"
                  @click="execute(cmd)"
                  @mousemove="selectedGroup = group.title; selectedIdx = idx"
                >
                  <span class="cp-item-icon" :style="{ background: cmd.iconBg }">
                    {{ cmd.icon }}
                  </span>
                  <span class="cp-item-label">{{ cmd.label }}</span>
                  <span v-if="cmd.shortcut" class="cp-item-shortcut">{{ cmd.shortcut }}</span>
                </div>
              </div>
            </template>

            <!-- 搜索模式 -->
            <template v-else>
              <!-- Loading -->
              <div v-if="store.loading" class="cp-status">
                <span class="cp-spinner" />搜索中...
              </div>

              <!-- Results -->
              <template v-else-if="store.results">
                <!-- issues -->
                <div v-if="store.results.results.issues?.length" class="cp-group">
                  <div class="cp-group-title">工作项</div>
                  <div
                    v-for="(item, idx) in store.results.results.issues"
                    :key="'issue-' + item.id"
                    class="cp-item"
                    :class="{ 'cp-item--selected': isSelected('search', idx, 'issues') }"
                    @click="goTo('issue', item)"
                    @mousemove="selectItem('search', idx, 'issues')"
                  >
                    <span class="cp-item-icon cp-item-icon--hash">#</span>
                    <!-- eslint-disable-next-line vue/no-v-html -- highlight 由服务端 ts_headline 生成，内容已转义 -->
                    <span class="cp-item-label" v-html="item.highlight || item.name" />
                    <span class="cp-item-meta">{{ item.project_name }}</span>
                  </div>
                </div>

                <!-- sprints -->
                <div v-if="store.results.results.sprints?.length" class="cp-group">
                  <div class="cp-group-title">迭代</div>
                  <div
                    v-for="(item, idx) in store.results.results.sprints"
                    :key="'sprint-' + item.id"
                    class="cp-item"
                    :class="{ 'cp-item--selected': isSelected('search', offsetSprint + idx, 'sprints') }"
                    @click="goTo('sprint', item)"
                    @mousemove="selectItem('search', offsetSprint + idx, 'sprints')"
                  >
                    <span class="cp-item-icon" style="background: var(--extended-color-purple-50, #f3e8ff);">🏃</span>
                    <!-- eslint-disable-next-line vue/no-v-html -- highlight 由服务端 ts_headline 生成，内容已转义 -->
                    <span class="cp-item-label" v-html="item.highlight || item.name" />
                    <span class="cp-item-meta">{{ item.project_name }}</span>
                  </div>
                </div>

                <!-- versions -->
                <div v-if="store.results.results.versions?.length" class="cp-group">
                  <div class="cp-group-title">版本</div>
                  <div
                    v-for="(item, idx) in store.results.results.versions"
                    :key="'version-' + item.id"
                    class="cp-item"
                    :class="{ 'cp-item--selected': isSelected('search', offsetVersion + idx, 'versions') }"
                    @click="goTo('version', item)"
                    @mousemove="selectItem('search', offsetVersion + idx, 'versions')"
                  >
                    <span class="cp-item-icon" style="background: var(--extended-color-emerald-50, #ecfdf5);">🚀</span>
                    <!-- eslint-disable-next-line vue/no-v-html -- highlight 由服务端 ts_headline 生成，内容已转义 -->
                    <span class="cp-item-label" v-html="item.highlight || item.name" />
                    <span class="cp-item-meta">{{ item.project_name }}</span>
                  </div>
                </div>

                <!-- 空结果 -->
                <div v-if="isEmpty && !store.loading" class="cp-status">
                  <span class="cp-status-icon">🔍</span>
                  未找到与「{{ query }}」相关的结果
                </div>
              </template>

              <!-- 无搜索词提示 -->
              <div v-else-if="query.trim() && !store.loading && !store.results" class="cp-status">
                输入关键字搜索工作项、迭代、版本...
              </div>
            </template>
          </div>

          <!-- 底部 hint -->
          <div class="cp-footer">
            <span><kbd>↑</kbd><kbd>↓</kbd> 导航</span>
            <span><kbd>↵</kbd> 确认</span>
            <span><kbd>ESC</kbd> 关闭</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue"
import { useRouter, useRoute } from "vue-router"
import { useSearchStore } from "@/stores/search"
import { useWorkspaceStore } from "@/stores/workspace"

// ---------------------------------------------------------------------------
//  Router / Store
// ---------------------------------------------------------------------------
const router = useRouter()
const route = useRoute()
const store = useSearchStore()
const wsStore = useWorkspaceStore()

const open = computed(() => store.open)

// ---------------------------------------------------------------------------
//  Local state
// ---------------------------------------------------------------------------
const query = ref("")
const inputRef = ref<HTMLInputElement | null>(null)
const selectedGroup = ref("Jump")
const selectedIdx = ref(-1)
const mode = ref<"command" | "search">("command")

let debounceTimer: ReturnType<typeof setTimeout> | null = null

// ---------------------------------------------------------------------------
//  Command definitions (Jump + Create)
// ---------------------------------------------------------------------------
interface CommandDef {
  id: string
  label: string
  icon: string
  iconBg: string
  shortcut?: string
  action: () => void
}

interface CommandGroupDef {
  title: string
  commands: CommandDef[]
}

const commandGroups = computed<CommandGroupDef[]>(() => {
  const slug = wsStore.currentSlug

  return [
    {
      title: "跳转",
      commands: [
        {
          id: "go-home",
          label: "工作台",
          icon: "🏠",
          iconBg: "var(--brand-50, #eef2fe)",
          action: () => router.push(`/${slug}`),
        },
        {
          id: "go-projects",
          label: "项目列表",
          icon: "📁",
          iconBg: "var(--brand-50)",
          action: () => router.push(`/${slug}/projects`),
        },
        {
          id: "go-search",
          label: "全局搜索",
          icon: "🔍",
          iconBg: "var(--neutral-200)",
          shortcut: "Ctrl+K",
          action: () => { mode.value = "search" },
        },
        {
          id: "go-notifications",
          label: "通知中心",
          icon: "🔔",
          iconBg: "var(--amber-50, #fffbeb)",
          action: () => router.push(`/${slug}/notifications`),
        },
        {
          id: "go-workbench",
          label: "我的工作台",
          icon: "📋",
          iconBg: "var(--extended-color-indigo-50, #eef2fe)",
          action: () => router.push(`/${slug}/workbench`),
        },
        {
          id: "go-settings",
          label: "工作空间设置",
          icon: "⚙️",
          iconBg: "var(--neutral-200)",
          action: () => router.push(`/${slug}/settings`),
        },
      ],
    },
    {
      title: "创建",
      commands: [
        {
          id: "create-issue",
          label: "创建工作项",
          icon: "✏️",
          iconBg: "var(--brand-50)",
          shortcut: "C",
          action: () => {
            // 如果在项目内，打开创建弹窗；否则先跳项目列表
            const projectId = route.params.projectId
            if (projectId) {
              window.dispatchEvent(new CustomEvent("command:create-issue", { detail: { projectId } }))
            }
            router.push(`/${slug}/projects`)
          },
        },
        {
          id: "create-project",
          label: "创建项目",
          icon: "📂",
          iconBg: "var(--extended-color-emerald-50, #ecfdf5)",
          action: () => window.dispatchEvent(new CustomEvent("command:create-project")),
        },
        {
          id: "create-sprint",
          label: "创建迭代",
          icon: "🏃",
          iconBg: "var(--extended-color-purple-50, #f3e8ff)",
          action: () => window.dispatchEvent(new CustomEvent("command:create-sprint")),
        },
      ],
    },
  ]
})

// ---------------------------------------------------------------------------
//  Selection / Navigation
// ---------------------------------------------------------------------------
const allCommandFlat = computed(() => {
  const groups = commandGroups.value
  return groups.flatMap((g, gi) => g.commands.map((cmd, ci) => ({ group: g.title, idx: ci, gi, cmd })))
})

const offsetSprint = computed(() => store.results?.results.issues?.length ?? 0)
const offsetVersion = computed(() => offsetSprint.value + (store.results?.results.sprints?.length ?? 0))

function isSelected(group: Pick<CommandGroupDef, "title"> | string, idx: number, _entityType?: string) {
  const g = typeof group === "string" ? group : group.title
  return selectedGroup.value === g && selectedIdx.value === idx
}

function selectItem(groupTitle: string, idx: number, _entityType?: string) {
  selectedGroup.value = groupTitle
  selectedIdx.value = idx
}

function moveDown() {
  if (mode.value === "command") {
    const flat = allCommandFlat.value
    if (flat.length === 0) return
    const currentGlobalIdx = flat.findIndex(f => f.group === selectedGroup.value && f.idx === selectedIdx.value)
    const next = currentGlobalIdx + 1
    if (next >= flat.length) return
    selectedGroup.value = flat[next].group
    selectedIdx.value = flat[next].idx
  } else {
    // Search mode: navigate through search results
    const allItems = getAllSearchItems()
    if (allItems.length === 0) return
    const totalFlatIdx = getTotalFlatIdx()
    selectedIdx.value = Math.min(totalFlatIdx + 1, allItems.length - 1)
    // Recalculate selectedGroup based on position
    const issueCount = store.results?.results.issues?.length ?? 0
    const sprintCount = store.results?.results.sprints?.length ?? 0
    if (selectedIdx.value < issueCount) {
      selectedGroup.value = "search-issues"
    } else if (selectedIdx.value < issueCount + sprintCount) {
      selectedGroup.value = "search-sprints"
    } else {
      selectedGroup.value = "search-versions"
    }
  }
}

function moveUp() {
  if (mode.value === "command") {
    const flat = allCommandFlat.value
    if (flat.length === 0) return
    const currentGlobalIdx = flat.findIndex(f => f.group === selectedGroup.value && f.idx === selectedIdx.value)
    const prev = currentGlobalIdx - 1
    if (prev < 0) return
    selectedGroup.value = flat[prev].group
    selectedIdx.value = flat[prev].idx
  } else {
    const totalFlatIdx = getTotalFlatIdx()
    selectedIdx.value = Math.max(totalFlatIdx - 1, 0)
    // Recalculate selectedGroup
    const issueCount = store.results?.results.issues?.length ?? 0
    const sprintCount = store.results?.results.sprints?.length ?? 0
    if (selectedIdx.value < issueCount) {
      selectedGroup.value = "search-issues"
    } else if (selectedIdx.value < issueCount + sprintCount) {
      selectedGroup.value = "search-sprints"
    } else {
      selectedGroup.value = "search-versions"
    }
  }
}

function getTotalFlatIdx(): number {
  if (selectedGroup.value === "search-issues") return selectedIdx.value
  if (selectedGroup.value === "search-sprints") return (store.results?.results.issues?.length ?? 0) + selectedIdx.value
  if (selectedGroup.value === "search-versions") return (store.results?.results.issues?.length ?? 0) + (store.results?.results.sprints?.length ?? 0) + selectedIdx.value
  return selectedIdx.value
}

function getAllSearchItems(): any[] {
  const r = store.results
  if (!r) return []
  return [
    ...(r.results.issues || []).map(i => ({ ...i, _type: 'issue' })),
    ...(r.results.sprints || []).map(i => ({ ...i, _type: 'sprint' })),
    ...(r.results.versions || []).map(i => ({ ...i, _type: 'version' })),
  ]
}

function getSelectedSearchItem(): any | null {
  const items = getAllSearchItems()
  const flatIdx = getTotalFlatIdx()
  if (flatIdx >= 0 && flatIdx < items.length) return items[flatIdx]
  return null
}

// ---------------------------------------------------------------------------
//  Actions
// ---------------------------------------------------------------------------
function execute(cmd: CommandDef) {
  cmd.action()
  close()
}

function goTo(type: string, item: any) {
  const slug = wsStore.currentSlug
  if (type === 'issue') router.push(`/${slug}/projects/${item.project_id}/issues/${item.id}`)
  else if (type === 'sprint') router.push(`/${slug}/projects/${item.project_id}/sprints/${item.id}`)
  else if (type === 'version') router.push(`/${slug}/projects/${item.project_id}/versions/${item.id}`)
  close()
}

function handleEnter() {
  if (mode.value === "command") {
    const cmd = allCommandFlat.value.find(f => f.group === selectedGroup.value && f.idx === selectedIdx.value)
    if (cmd) execute(cmd.cmd)
  } else {
    const item = getSelectedSearchItem()
    if (item) goTo(item._type, item)
  }
}

// ---------------------------------------------------------------------------
//  Search debounce
// ---------------------------------------------------------------------------
function handleInput() {
  selectedIdx.value = -1
  selectedGroup.value = "Jump"

  // Mode switching based on query
  if (query.value.trim()) {
    mode.value = "search"
  } else {
    mode.value = "command"
    store.clear()
    return
  }

  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (query.value.trim() && wsStore.current) {
      store.search(wsStore.current.id, query.value)
    }
  }, 200)
}

// ---------------------------------------------------------------------------
//  Computed
// ---------------------------------------------------------------------------
const isEmpty = computed(() => {
  const r = store.results
  if (!r) return true
  return (!r.results.issues?.length) && (!r.results.sprints?.length) && (!r.results.versions?.length)
})

// ---------------------------------------------------------------------------
//  Open / Close
// ---------------------------------------------------------------------------
function close() {
  store.clear()
  query.value = ""
  selectedIdx.value = -1
  selectedGroup.value = "Jump"
  mode.value = "command"
}

watch(open, (val) => {
  if (val) {
    query.value = ""
    mode.value = "command"
    selectedIdx.value = -1
    selectedGroup.value = "Jump"
    nextTick(() => inputRef.value?.focus())
  }
})

// ---------------------------------------------------------------------------
//  Global hotkey
// ---------------------------------------------------------------------------
function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    store.toggle()
  }
  // 'c' for quick create (when not in input)
  if (e.key === 'c' && !e.ctrlKey && !e.metaKey && !e.altKey) {
    const target = e.target as HTMLElement
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable
    if (!isInput && wsStore.currentSlug) {
      e.preventDefault()
      mode.value = "command"
      // Set filter to show create commands
      selectedGroup.value = "创建"
      selectedIdx.value = 0
      // Open palette if not open
      if (!store.open) store.toggle()
      // If palette is already open, just switch mode
      query.value = ""
    }
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
/* ---- Overlay ---- */
.cp-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 12vh;
  background: var(--bg-backdrop, rgba(15, 23, 42, 0.4));
  backdrop-filter: blur(4px);
}

/* ---- Dialog ---- */
.cp-dialog {
  width: 600px;
  max-width: 92vw;
  max-height: 70vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-surface-1, #fff);
  border-radius: var(--radius-lg, 12px);
  box-shadow: var(--shadow-overlay-200);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

/* ---- Input ---- */
.cp-input-wrap {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  gap: 10px;
}

.cp-input-kbd {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--bg-layer-1);
  color: var(--txt-tertiary);
  font-family: var(--font-mono);
  border: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.cp-input-kbd--esc {
  margin-left: auto;
}

.cp-search-icon {
  color: var(--txt-tertiary);
  flex-shrink: 0;
}

.cp-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 15px;
  color: var(--txt-primary);
  background: transparent;
  min-width: 0;
}

.cp-input::placeholder {
  color: var(--txt-placeholder);
}

/* ---- Body ---- */
.cp-body {
  max-height: 380px;
  overflow-y: auto;
  padding: 8px 0;
  flex: 1;
}

.cp-group {
  padding: 4px 0;
}

.cp-group-title {
  padding: 6px 16px;
  font-size: 11px;
  font-weight: 600;
  color: var(--txt-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.cp-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
  border-radius: 4px;
  margin: 0 4px;
  transition: background 0.1s;
}

.cp-item:hover,
.cp-item--selected {
  background: var(--bg-accent-subtle, var(--brand-50, #eef2fe));
}

.cp-item-icon {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  flex-shrink: 0;
}

.cp-item-icon--hash {
  background: var(--neutral-200);
  color: var(--txt-tertiary);
  font-weight: 600;
  font-size: 12px;
}

.cp-item-label {
  flex: 1;
  color: var(--txt-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cp-item-label :deep(b) {
  font-weight: 600;
  color: var(--txt-accent-primary);
}

.cp-item-meta {
  font-size: 12px;
  color: var(--txt-tertiary);
  flex-shrink: 0;
}

.cp-item-shortcut {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--bg-layer-1);
  color: var(--txt-tertiary);
  font-family: var(--font-mono);
  border: 1px solid var(--border-subtle);
}

/* ---- Status ---- */
.cp-status {
  padding: 24px;
  text-align: center;
  color: var(--txt-tertiary);
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.cp-status-icon {
  font-size: 20px;
}

.cp-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--brand-default);
  border-radius: 50%;
  animation: cp-spin 0.6s linear infinite;
  display: inline-block;
}

@keyframes cp-spin {
  to { transform: rotate(360deg); }
}

/* ---- Footer ---- */
.cp-footer {
  display: flex;
  gap: 16px;
  padding: 8px 16px;
  border-top: 1px solid var(--border-subtle);
  font-size: 11px;
  color: var(--txt-tertiary);
}

.cp-footer kbd {
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--bg-layer-1);
  border: 1px solid var(--border-subtle);
  font-family: var(--font-mono);
  font-size: 10px;
}

/* ---- Transition ---- */
.cp-overlay-enter-active,
.cp-overlay-leave-active {
  transition: opacity 0.15s ease;
}

.cp-overlay-enter-active .cp-dialog,
.cp-overlay-leave-active .cp-dialog {
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.cp-overlay-enter-from,
.cp-overlay-leave-to {
  opacity: 0;
}

.cp-overlay-enter-from .cp-dialog,
.cp-overlay-leave-to .cp-dialog {
  transform: translateY(-8px) scale(0.98);
  opacity: 0;
}
</style>
