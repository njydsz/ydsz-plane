<script setup lang="ts">
/* eslint-disable vue/no-v-html -- highlight 由服务端 PostgreSQL ts_headline 生成：内容已被 ts_headline 自动 HTML 转义，仅包裹 <b> 高亮标签，属受信输出，无 XSS 注入面。 */
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
  workspaceId: number;
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
const showJqlHelp = ref(false);
const jqlHoveredField = ref<string | null>(null);

/** JQL 语法定义（与后端 pkg/searchql 保持一致） */
const jqlFields = [
  { key: "project", label: "项目", example: "project:YD", desc: "按项目 key 过滤" },
  { key: "type", label: "类型", example: "type:defect", desc: "requirement / task / defect" },
  { key: "status", label: "状态", example: "status:todo", desc: "按状态名过滤" },
  { key: "priority", label: "优先级", example: "priority:high", desc: "urgent / high / medium / low" },
  { key: "severity", label: "严重程度", example: "severity>=3", desc: "数值比较 (1-5)" },
  { key: "assignee", label: "指派给", example: "assignee:me()", desc: "me() / currentUser() / 用户名" },
  { key: "reporter", label: "报告人", example: "reporter:me()", desc: "同 assignee" },
  { key: "label", label: "标签", example: "label:前端", desc: "按标签名过滤" },
  { key: "module", label: "模块", example: "module:支付", desc: "按模块名过滤" },
  { key: "sprint", label: "迭代", example: "sprint:当前", desc: "迭代名或 ID" },
  { key: "version", label: "版本", example: "version:v1.0", desc: "按版本名过滤" },
  { key: "due", label: "截止日期", example: "due<now()", desc: "日期比较，支持 now(-7d)" },
  { key: "created", label: "创建时间", example: "created>now(-30d)", desc: "日期范围" },
  { key: "updated", label: "更新时间", example: "updated>now(-7d)", desc: "日期范围" },
];

const jqlOperators = [
  { op: ":", label: "等于", example: "status:todo" },
  { op: "=", label: "等于（同 :)", example: "type=defect" },
  { op: "!=", label: "不等于", example: "status!=done" },
  { op: ">", label: "大于", example: "severity>2" },
  { op: ">=", label: "大于等于", example: "severity>=3" },
  { op: "<", label: "小于", example: "due<now()" },
  { op: "<=", label: "小于等于", example: "priority<=2" },
  { op: "in", label: "包含于", example: "status in (todo, doing)" },
];

const jqlExamples = [
  { q: "登录页 闪退", desc: "全文检索（关键词之间隐式 AND）" },
  { q: "project:YD status:todo assignee:me()", desc: "结构化多条件组合" },
  { q: "type:defect severity>=3 created>now(-7d)", desc: "严重缺陷 + 近 7 天" },
  { q: '"支付回调" AND module:支付', desc: "短语 + 字段组合" },
  { q: "type:task status in (todo, doing) -assignee:me()", desc: "排除已指派给我的" },
];

/** 检测输入是否包含 JQL field: 语法 */
const hasJqlSyntax = computed(() => /\b[\w]+\s*[:=<>]/.test(query.value));

// ---- History & Bookmarks ----
const showHistory = ref(false);
const history = ref<SearchHistoryItem[]>([]);
const bookmarks = ref<SearchBookmark[]>([]);

// ---- Computed ----
const hasResults = computed(() => (results.value?.total ?? 0) > 0);

// ---- Actions ----
async function load() {
  ws.value = await workspaceApi.get(props.workspaceId);
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
      path: `/${props.workspaceId}/search`,
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

// ---- CSV 导出 ----

/** CSV 单元格转义（处理逗号、引号、换行） */
function csvCell(v: unknown): string {
  if (v == null) return "";
  const s = String(v);
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

/** 将搜索结果下载为 CSV */
async function exportResultsToCsv() {
  if (!ws.value || !query.value.trim()) return;
  try {
    // 导出时使用更大 limit 以覆盖更多匹配项
    const resp = await searchApi.searchWorkspace(ws.value.id, {
      q: query.value.trim(),
      limit: 200,
    });
    const rows: (string | number)[][] = [
      ["类型", "标识符", "名称", "描述", "项目", "排名分", "链接"],
    ];
    const allItems: SearchResultItem[] = [
      ...resp.results.issues.map((i) => ({ ...i, type: "工作项" })),
      ...resp.results.sprints.map((s) => ({ ...s, type: "迭代" })),
      ...resp.results.versions.map((v) => ({ ...v, type: "版本" })),
      ...resp.results.projects.map((p) => ({ ...p, type: "项目" })),
    ];
    for (const item of allItems) {
      rows.push([
        item.type,
        item.identifier ?? "",
        item.name,
        (item.description ?? "").replace(/\n/g, " "),
        item.project_name ?? "",
        item.rank.toFixed(3),
        item.url ? `${window.location.origin}${item.url}` : "",
      ]);
    }
    const csv = rows.map((r) => r.map(csvCell).join(",")).join("\r\n");
    // BOM 头确保 Excel 正确识别 UTF-8 中文
    const blob = new Blob(["﻿" + csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "search-" + query.value.trim().slice(0, 30) + "-" + new Date().toISOString().slice(0, 10) + ".csv";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "导出失败";
  }
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
          :class="{ 'search-bar__input--jql': hasJqlSyntax }"
          placeholder='搜索... 支持 JQL 语法，如 type:defect assignee:me()'
          @focus="showHistory = true"
          @keydown.enter="submitSearch"
        />
        <button class="search-bar__btn" @click="submitSearch">搜索</button>
        <button
          class="search-bar__help-btn"
          :class="{ 'search-bar__help-btn--active': showJqlHelp }"
          title="JQL 语法帮助"
          @click="showJqlHelp = !showJqlHelp"
        >?</button>
      </div>
      <div v-if="hasJqlSyntax" class="search-bar__jql-hint">
        <span class="jql-badge">JQL</span>
        <span class="jql-hint__text">已检测到结构化查询语法</span>
      </div>

      <!-- JQL 语法帮助面板 -->
      <div v-if="showJqlHelp" class="jql-help-panel">
        <div class="jql-help-panel__header">
          <h3>JQL 搜索语法</h3>
          <span class="jql-help-panel__note">支持类 Jira 查询语言，无语法时自动降级全文检索</span>
        </div>

        <div class="jql-help-panel__body">
          <!-- 示例 -->
          <section class="jql-section">
            <h4 class="jql-section__title">快速示例 <small>（点击插入）</small></h4>
            <div class="jql-examples">
              <button
                v-for="(ex, i) in jqlExamples"
                :key="i"
                class="jql-example-chip"
                @click="query = ex.q; showJqlHelp = false; submitSearch()"
              >
                <code>{{ ex.q }}</code>
                <span class="jql-example-chip__desc">{{ ex.desc }}</span>
              </button>
            </div>
          </section>

          <!-- 字段 -->
          <section class="jql-section">
            <h4 class="jql-section__title">字段</h4>
            <div class="jql-fields">
              <div
                v-for="f in jqlFields"
                :key="f.key"
                class="jql-field"
                @mouseenter="jqlHoveredField = f.key"
                @mouseleave="jqlHoveredField = null"
              >
                <code class="jql-field__key" @click="query += ' ' + f.key + ':'; showJqlHelp = false">{{ f.key }}</code>
                <span class="jql-field__label">{{ f.label }}</span>
                <span class="jql-field__desc">{{ f.desc }}</span>
                <code v-if="jqlHoveredField === f.key" class="jql-field__ex">{{ f.example }}</code>
              </div>
            </div>
          </section>

          <!-- 操作符 -->
          <section class="jql-section">
            <h4 class="jql-section__title">操作符</h4>
            <div class="jql-operators">
              <div v-for="op in jqlOperators" :key="op.op" class="jql-operator">
                <code>{{ op.op }}</code>
                <span>{{ op.label }}</span>
                <code class="jql-operator__ex">{{ op.example }}</code>
              </div>
            </div>
          </section>

          <!-- 逻辑 -->
          <section class="jql-section">
            <h4 class="jql-section__title">逻辑连接</h4>
            <div class="jql-logic">
              <code>AND</code> <span>空格隐式 AND</span>
              <code>OR</code> <span>或关系</span>
              <code>NOT / -</code> <span>排除（前缀 -）</span>
              <code>( ... )</code> <span>分组</span>
            </div>
          </section>
        </div>
      </div>

      <!-- Content -->
      <AppLoadingState v-if="loading" text="搜索中..." />
      <AppErrorState v-else-if="error" :message="error" @retry="submitSearch" />

      <template v-else-if="results">
        <div class="search-meta__row">
          <p v-if="!hasResults" class="search-meta">
            未找到与"<strong>{{ query }}</strong>"相关的结果
          </p>
          <p v-else class="search-meta">
            找到 <strong>{{ results.total }}</strong> 条结果 · {{ results.time_ms }}ms
          </p>
          <button
            v-if="hasResults"
            class="export-csv-btn"
            title="导出全部结果为 CSV（最多 200 条）"
            @click="exportResultsToCsv"
          >
            ⬇ 导出 CSV
          </button>
        </div>

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
                <span v-if="item.highlight" class="result-card__name" v-html="item.highlight"></span><span v-else class="result-card__name">{{ item.name }}</span>
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
                <span v-if="item.highlight" class="result-card__name" v-html="item.highlight"></span><span v-else class="result-card__name">{{ item.name }}</span>
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
                <span v-if="item.highlight" class="result-card__name" v-html="item.highlight"></span><span v-else class="result-card__name">{{ item.name }}</span>
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
  position: relative;
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
  padding-right: 36px;
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

/* ---- JQL ---- */
.search-bar__help-btn {
  position: absolute;
  right: 80px;
  top: 50%;
  transform: translateY(-50%);
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1px solid var(--border-default, #e5e7eb);
  background: var(--surface-1, #fff);
  color: var(--text-tertiary, #9ca3af);
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  font-family: var(--font-mono, monospace);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.search-bar__help-btn:hover {
  border-color: var(--brand-500, #3b82f6);
  color: var(--brand-500, #3b82f6);
}
.search-bar__help-btn--active {
  background: var(--brand-500, #3b82f6);
  color: #fff;
  border-color: var(--brand-500, #3b82f6);
}
.search-bar__input--jql {
  border-color: var(--brand-300, #93c5fd) !important;
  background: var(--brand-50, #eff6ff);
}
.search-bar__jql-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}
.jql-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--brand-500, #3b82f6);
  color: #fff;
  letter-spacing: 0.5px;
}
.jql-hint__text { color: var(--brand-600, #2563eb); }

/* ---- JQL Help Panel ---- */
.jql-help-panel {
  margin-top: 12px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-md, 8px);
  background: var(--surface-1, #fff);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}
.jql-help-panel__header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
  background: var(--surface-2, #f9fafb);
}
.jql-help-panel__header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}
.jql-help-panel__note {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}
.jql-help-panel__body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  padding: 16px;
}
.jql-section__title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #6b7280);
  margin: 0 0 8px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.jql-section__title small {
  font-weight: 400;
  text-transform: none;
  color: var(--text-tertiary, #9ca3af);
}
/* Examples */
.jql-examples {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.jql-example-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-1, #fff);
  cursor: pointer;
  text-align: left;
  font-family: inherit;
  transition: border-color 0.15s, background 0.15s;
}
.jql-example-chip:hover {
  border-color: var(--brand-400, #60a5fa);
  background: var(--brand-50, #eff6ff);
}
.jql-example-chip code {
  font-size: 11px;
  color: var(--brand-700, #1d4ed8);
  background: var(--brand-50, #eff6ff);
  padding: 1px 5px;
  border-radius: 3px;
  white-space: nowrap;
  flex-shrink: 0;
}
.jql-example-chip__desc {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
/* Fields */
.jql-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
}
.jql-field {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  padding: 4px 6px;
  border-radius: var(--radius-sm, 6px);
  cursor: pointer;
  position: relative;
}
.jql-field:hover { background: var(--surface-2, #f9fafb); }
.jql-field__key {
  font-size: 11px;
  color: var(--brand-600, #2563eb);
  cursor: pointer;
  font-weight: 600;
}
.jql-field__key:hover { text-decoration: underline; }
.jql-field__label {
  font-size: 11px;
  color: var(--text-secondary, #6b7280);
}
.jql-field__desc {
  font-size: 10px;
  color: var(--text-tertiary, #9ca3af);
  width: 100%;
  padding-left: 2px;
}
.jql-field__ex {
  font-size: 10px;
  color: var(--success-600, #059669);
  background: var(--success-50, #ecfdf5);
  padding: 1px 4px;
  border-radius: 2px;
}
/* Operators */
.jql-operators {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
}
.jql-operator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary, #6b7280);
  padding: 3px 6px;
}
.jql-operator code {
  font-size: 11px;
  color: var(--text-primary, #1f2937);
  background: var(--surface-3, #e5e7eb);
  padding: 1px 5px;
  border-radius: 3px;
  min-width: 20px;
  text-align: center;
}
.jql-operator__ex {
  color: var(--text-tertiary, #9ca3af);
  margin-left: auto;
}
/* Logic */
.jql-logic {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 4px 10px;
  font-size: 11px;
  color: var(--text-secondary, #6b7280);
}
.jql-logic code {
  font-size: 11px;
  color: var(--warning-700, #b45309);
  background: var(--warning-50, #fffbeb);
  padding: 1px 5px;
  border-radius: 3px;
  text-align: center;
}

@media (max-width: 768px) {
  .jql-help-panel__body { grid-template-columns: 1fr; }
  .jql-fields { grid-template-columns: 1fr; }
  .search-bar__help-btn { right: 70px; }
}

.search-meta__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  gap: 12px;
}

.search-meta {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
  margin: 0;
}

.search-meta strong {
  color: var(--text-primary, #1f2937);
}

.export-csv-btn {
  padding: 6px 12px;
  border: 1px solid var(--border-default, #e5e7eb);
  border-radius: var(--radius-sm, 4px);
  background: var(--surface-2, #f9fafb);
  color: var(--text-secondary, #6b7280);
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.export-csv-btn:hover {
  background: var(--brand-50);
  border-color: var(--brand-400);
  color: var(--brand-600);
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
