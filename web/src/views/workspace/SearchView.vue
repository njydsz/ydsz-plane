<script setup lang="ts">
/**
 * SearchView — 独立搜索结果页。
 *
 * 设计对标: Plane / Linear / Jira 搜索页
 *  - 顶栏搜索框（Cmd+K 全局一致性）
 *  - 左侧过滤器（类型/项目/状态/优先级/时间）
 *  - 右侧按类型分组结果（每组可展开"查看更多"）
 *  - 关键词 <mark> 高亮
 */

import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import AppEmptyState from "@/components/AppEmptyState.vue";
import AppLoadingState from "@/components/AppLoadingState.vue";
import AppErrorState from "@/components/AppErrorState.vue";
import {
  searchApi,
  type SearchResponse,
  type SearchResultItem,
  type SearchHistoryItem,
  type SearchBookmark,
} from "@/api/services/search";
import { workspaceApi, type Workspace } from "@/api/services/workspace";

const route = useRoute();
const router = useRouter();

// ---- Route params ----
const props = defineProps<{
  workspaceSlug: string;
}>();

// ---- State ----
const ws = ref<Workspace | null>(null);
const query = ref(String(route.query.q ?? ""));
const loading = ref(false);
const error = ref("");
const results = ref<SearchResponse | null>(null);
const activeTab = ref<"all" | "issue" | "sprint" | "version">("all");

// ---- Filter states ----
const filterOpen = ref(true);

// ---- History & Bookmarks ----
const showHistory = ref(false);
const history = ref<SearchHistoryItem[]>([]);
const bookmarks = ref<SearchBookmark[]>([]);

// ---- Computed ----
const allHits = computed((): SearchResultItem[] => {
  if (!results.value) return [];
  const r = value.results;
  return [...r.issues, ...r.sprints, ...r.versions];
});

const filteredHits = computed((): SearchResultItem[] => {
  if (activeTab.value === "all") return allHits.value;
  return allHits.value.filter((h) => h.type === activeTab.value);
});

const hasResults = computed(() => (results.value?.total ?? 0) > 0);

// ---- Actions ----
async function load() {
  ws.value = await workspaceApi.getBySlug(props.workspaceSlug);
  if (query.value.trim()) {
    await runSearch(query.value);
  }
  await Promise.all([loadHistory(), loadBookmarks()]);
}

async function runSearch(q: string) {
  if (!q.trim() || !ws.value) return;
  loading.value = true;
  error.value = "";
  try {
    // 回写 URL（replaceState 避免历史堆栈污染）
    router.replace({
      path: `/${props.workspaceSlug}/search`,
      query: { q },
    });
    const resp = await searchApi.searchWorkspace(ws.value.id, {
      q,
      limit: 20,
    });
    results.value = resp;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "搜索失败";
  } finally {
    loading.value = false;
  }
}

function submitSearch() {
  if (query.value.trim()) {
    runSearch(query.value);
    showHistory.value = false;
  }
}

function applyHistory(h: SearchHistoryItem) {
  query.value = h.query;
  submitSearch();
}

function navigateTo(item: SearchResultItem) {
  if (item.url) {
    router.push(item.url);
  }
}

async function loadHistory() {
  if (!ws.value) return;
  try {
    const resp = await searchApi.getHistory(ws.value.id);
    history.value = resp.results ?? [];
  } catch {
    // 忽略历史加载失败
  }
}

async function loadBookmarks() {
  if (!ws.value) return;
  try {
    const resp = await searchApi.getBookmarks(ws.value.id);
    bookmarks.value = resp.results ?? [];
  } catch {
    // 忽略书签加载失败
  }
}

async function clearHistory() {
  if (!ws.value) return;
  await searchApi.clearHistory(ws.value.id);
  history.value = [];
}

// ---- Watchers ----
watch(
  () => route.query.q,
  (q) => {
    if (q && q !== query.value) {
      query.value = String(q);
      runSearch(query.value);
    }
  },
);

onMounted(load);
</script>

<template>
  <div class="search-page">
    <!-- Left Filters (collapsible) -->
    <aside v-if="filterOpen" class="search-filters">
      <div class="filter-group">
        <h3 class="filter-group__title">类型</h3>
        <label
          v-for="tab in [
            { key: 'all', label: '全部' },
            { key: 'issue', label: '工作项' },
            { key: 'sprint', label: '迭代' },
            { key: 'version', label: '版本' },
          ]"
          :key="tab.key"
          class="filter-option"
          :class="{ 'filter-option--active': activeTab === tab.key }"
        >
          <input
            v-model="activeTab"
            type="radio"
            name="activeTab"
            :value="tab.key"
            class="visually-hidden"
          />
          {{ tab.label }}
        </label>
      </div>

      <!-- History -->
      <div v-if="history.length" class="filter-group">
        <div class="filter-group__head">
          <h3 class="filter-group__title">最近搜索</h3>
          <button class="filter-group__clear" @click="clearHistory">清空</button>
        </div>
        <div
          v-for="h in history.slice(0, 8)"
          :key="h.id"
          class="history-item"
          @click="applyHistory(h)"
        >
          <span class="history-item__query">{{ h.query }}</span>
          <span class="history-item__meta">{{ h.result_count }} 条</span>
        </div>
      </div>

      <!-- Bookmarks -->
      <div v-if="bookmarks.length" class="filter-group">
        <h3 class="filter-group__title">收藏</h3>
        <div
          v-for="bm in bookmarks"
          :key="bm.id"
          class="bookmark-item"
          @click="applyHistory({ id: bm.id!, query: bm.query } as SearchHistoryItem)"
        >
          <span class="bookmark-item__icon">📌</span>
          <span class="bookmark-item__name">{{ bm.name }}</span>
        </div>
      </div>
    </aside>

    <!-- Main Results -->
    <main class="search-main">
      <!-- Search Bar -->
      <div class="search-bar">
        <svg class="search-bar__icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          v-model="query"
          class="search-bar__input"
          placeholder="搜索工作项、迭代、版本..."
          @focus="showHistory = true"
          @keydown.enter="submitSearch"
        />
        <button class="search-bar__btn" @click="submitSearch">搜索</button>
      </div>

      <!-- Content -->
      <AppLoadingState v-if="loading" text="搜索中..." />
      <AppErrorState v-else-if="error" :message="error" @retry="submitSearch" />

      <template v-else-if="results">
        <p v-if="!hasResults" class="search-meta">
          未找到与"<strong>{{ query }}</strong>"相关的结果
        </p>
        <p v-else class="search-meta">
          找到 <strong>{{ results.total }}</strong> 条结果 · {{ results.time_ms }}ms
        </p>

        <div v-if="hasResults" class="result-groups">
          <!-- Issue Group -->
          <div
            v-if="activeTab === 'all' || activeTab === 'issue'"
            v-show="results.results.issues.length"
            class="result-group"
          >
            <h2 class="result-group__title">
              工作项 <span class="result-group__count">{{ results.results.issues.length }}</span>
            </h2>
            <div
              v-for="item in results.results.issues"
              :key="`issue-${item.id}`"
              class="result-card"
              @click="navigateTo(item)"
            >
              <div class="result-card__top">
                <span class="result-card__identifier">{{ item.identifier }}</span>
                <span class="result-card__name" v-html="item.highlight || item.name"></span>
              </div>
              <div class="result-card__meta">
                <span>{{ item.project_name }}</span>
                <span>· 优先级 {{ item.rank.toFixed(2) }}</span>
              </div>
            </div>
          </div>

          <!-- Sprint Group -->
          <div
            v-if="activeTab === 'all' || activeTab === 'sprint'"
            v-show="results.results.sprints.length"
            class="result-group"
          >
            <h2 class="result-group__title">
              迭代 <span class="result-group__count">{{ results.results.sprints.length }}</span>
            </h2>
            <div
              v-for="item in results.results.sprints"
              :key="`sprint-${item.id}`"
              class="result-card"
              @click="navigateTo(item)"
            >
              <div class="result-card__top">
                <span class="result-card__icon">🏃</span>
                <span class="result-card__name" v-html="item.highlight || item.name"></span>
              </div>
              <div class="result-card__meta">
                <span>{{ item.project_name }}</span>
              </div>
            </div>
          </div>

          <!-- Version Group -->
          <div
            v-if="activeTab === 'all' || activeTab === 'version'"
            v-show="results.results.versions.length"
            class="result-group"
          >
            <h2 class="result-group__title">
              版本 <span class="result-group__count">{{ results.results.versions.length }}</span>
            </h2>
            <div
              v-for="item in results.results.versions"
              :key="`version-${item.id}`"
              class="result-card"
              @click="navigateTo(item)"
            >
              <div class="result-card__top">
                <span class="result-card__icon">🚀</span>
                <span class="result-card__name" v-html="item.highlight || item.name"></span>
              </div>
              <div class="result-card__meta">
                <span>{{ item.project_name }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Empty state: no query yet -->
      <AppEmptyState
        v-else-if="!query.trim()"
        title="输入关键字开始搜索"
        description="支持工作项标识符、标题、描述全文检索"
      />
    </main>
  </div>
</template>

<style scoped>
.search-page {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 24px;
  max-width: 1100px;
  margin: 0 auto;
}

/* ---- Left Filters ---- */
.search-filters {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.filter-group__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-group__title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #9ca3af);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0 0 8px;
}

.filter-group__clear {
  border: none;
  background: none;
  font-size: 11px;
  color: var(--brand-500, #3b82f6);
  cursor: pointer;
  padding: 0;
}

.filter-option {
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-secondary, #4b5563);
  cursor: pointer;
  transition: background 0.15s;
}

.filter-option:hover {
  background: var(--surface-3, #f3f4f6);
}

.filter-option--active {
  background: var(--brand-50, #eff6ff);
  color: var(--brand-600, #2563eb);
  font-weight: 500;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
}

/* ---- History & Bookmarks ---- */
.history-item,
.bookmark-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: background 0.15s;
}

.history-item:hover,
.bookmark-item:hover {
  background: var(--surface-3, #f3f4f6);
}

.history-item__query,
.bookmark-item__name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary, #1f2937);
}

.history-item__meta {
  color: var(--text-tertiary, #9ca3af);
  font-size: 11px;
  flex-shrink: 0;
  margin-left: 8px;
}

.bookmark-item__icon {
  margin-right: 6px;
}

/* ---- Main Search Area ---- */
.search-main {
  min-width: 0;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--border-default, #d1d5db);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1, #fff);
  margin-bottom: 20px;
  transition: border-color 0.15s;
}

.search-bar:focus-within {
  border-color: var(--brand-500, #3b82f6);
  box-shadow: 0 0 0 3px var(--brand-50, #eff6ff);
}

.search-bar__icon {
  color: var(--text-tertiary, #9ca3af);
  flex-shrink: 0;
}

.search-bar__input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  background: transparent;
}

.search-bar__input::placeholder {
  color: var(--text-tertiary, #9ca3af);
}

.search-bar__btn {
  padding: 4px 12px;
  border: none;
  border-radius: var(--radius-sm, 6px);
  background: var(--brand-500, #3b82f6);
  color: white;
  font-size: 13px;
  cursor: pointer;
  flex-shrink: 0;
}

.search-bar__btn:hover {
  background: var(--brand-600, #2563eb);
}

.search-meta {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
  margin: 0 0 16px;
}

.search-meta strong {
  color: var(--text-primary, #1f2937);
}

/* ---- Result Groups ---- */
.result-groups {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.result-group__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0 0 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.result-group__count {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary, #9ca3af);
  background: var(--surface-3, #f3f4f6);
  padding: 1px 8px;
  border-radius: 10px;
}

.result-card {
  padding: 12px 14px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  margin-bottom: 8px;
  cursor: pointer;
  transition: border-color 0.15s, box-shadow 0.15s;
}

.result-card:hover {
  border-color: var(--brand-200, #bfdbfe);
  box-shadow: var(--shadow-sm, 0 1px 2px rgba(0, 0, 0, 0.05));
}

.result-card__top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.result-card__identifier {
  font-family: var(--font-mono, "Consolas", monospace);
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  flex-shrink: 0;
}

.result-card__icon {
  font-size: 14px;
  flex-shrink: 0;
}

.result-card__name {
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-card__name :deep(mark),
.result-card__name :deep(b) {
  background: var(--warning-100, #fef3c7);
  color: var(--warning-800, #92400e);
  padding: 0 2px;
  border-radius: 2px;
}

.result-card__meta {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  display: flex;
  gap: 4px;
}

@media (max-width: 900px) {
  .search-page {
    grid-template-columns: 1fr;
  }
  .search-filters {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 16px;
  }
}
</style>
