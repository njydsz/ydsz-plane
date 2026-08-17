<script setup lang="ts">
/**
 * 全局搜索 — 工作空间级全文搜索（PostgreSQL FTS）。
 * 支持结果分组（需求/任务/缺陷/迭代/版本/项目）、搜索历史与保存的书签。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import {
  searchApi,
  type SearchResponse,
  type SearchHistoryItem,
  type SearchBookmark,
} from "@/api/services/search";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const query = ref("");
const searching = ref(false);
const error = ref("");
const result = ref<SearchResponse | null>(null);
const history = ref<SearchHistoryItem[]>([]);
const bookmarks = ref<SearchBookmark[]>([]);

async function runSearch(q: string) {
  if (!q.trim() || !workspaceId.value) return;
  query.value = q;
  searching.value = true;
  error.value = "";
  result.value = null;
  try {
    result.value = await searchApi.searchWorkspace(workspaceId.value, { q });
    await loadHistory();
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "搜索失败";
  } finally {
    searching.value = false;
  }
}

async function loadHistory() {
  try {
    const r = await searchApi.getHistory(workspaceId.value);
    history.value = r.results ?? [];
  } catch {
    history.value = [];
  }
}

async function loadBookmarks() {
  try {
    const r = await searchApi.getBookmarks(workspaceId.value);
    bookmarks.value = r.results ?? [];
  } catch {
    bookmarks.value = [];
  }
}

function openItem(url?: string) {
  if (url) {
    router.push(url);
    return;
  }
}

const GROUPS: { key: keyof SearchResponse["results"]; label: string }[] = [
  { key: "issues", label: "需求/任务/缺陷" },
  { key: "sprints", label: "迭代" },
  { key: "versions", label: "版本" },
  { key: "projects", label: "项目" },
];

onMounted(() => {
  loadHistory();
  loadBookmarks();
});
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">全局搜索</h1>
    </div>

    <!-- 搜索框 -->
    <form
      class="flex gap-2"
      @submit.prevent="runSearch(query)"
    >
      <input
        v-model="query"
        type="text"
        placeholder="搜索需求/任务/缺陷、迭代、版本、项目…"
        class="flex-1 rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus:border-[var(--brand-500)]"
      />
      <button
        type="submit"
        :disabled="searching || !query.trim()"
        class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
      >
        {{ searching ? "搜索中…" : "搜索" }}
      </button>
    </form>

    <!-- 书签 -->
    <div v-if="bookmarks.length" class="flex flex-wrap gap-2">
      <button
        v-for="b in bookmarks"
        :key="b.id"
        class="rounded-md border border-[var(--border-subtle)] px-3 py-1 text-xs text-[var(--text-secondary)] hover:bg-[var(--surface-2)]"
        @click="runSearch(b.query)"
      >
        🔖 {{ b.name }}
      </button>
    </div>

    <!-- 搜索历史 -->
    <div v-if="!result && history.length" class="flex flex-wrap gap-2">
      <span class="text-xs text-[var(--text-tertiary)]">历史：</span>
      <button
        v-for="h in history.slice(0, 10)"
        :key="h.id"
        class="rounded-md px-2 py-1 text-xs text-[var(--text-secondary)] hover:bg-[var(--surface-2)]"
        @click="runSearch(h.query)"
      >
        {{ h.query }}
      </button>
    </div>

    <!-- 加载中 -->
    <div v-if="searching" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-12 w-full" />
    </div>

    <!-- 错误 -->
    <AppErrorState v-else-if="error" :message="error" @retry="runSearch(query)" />

    <!-- 结果 -->
    <template v-else-if="result">
      <div class="text-xs text-[var(--text-tertiary)]">
        共 {{ result.total }} 条结果（{{ result.time_ms }}ms）
      </div>

      <AppEmptyState
        v-if="result.total === 0"
        title="未找到匹配结果"
        description="尝试更换关键词或减少过滤条件。"
      />

      <div v-else class="space-y-6">
        <section
          v-for="g in GROUPS"
          :key="g.key"
        >
          <template v-if="result.results[g.key]?.length">
            <h2 class="mb-2 text-sm font-semibold text-[var(--text-secondary)]">{{ g.label }}</h2>
            <div class="space-y-1">
              <div
                v-for="item in result.results[g.key]"
                :key="`${item.type}-${item.id}`"
                class="flex cursor-pointer items-center gap-3 rounded-md border border-[var(--border-subtle)] px-3 py-2 hover:bg-[var(--surface-2)]"
                @click="openItem(item.url)"
              >
                <span class="font-mono text-xs text-[var(--brand-600)]">
                  {{ item.identifier ?? item.id }}
                </span>
                <div class="flex-1 min-w-0">
                  <div class="truncate text-sm text-[var(--text-primary)]">
                    <span v-html="item.highlight ?? item.name" />
                  </div>
                  <div v-if="item.description" class="truncate text-xs text-[var(--text-tertiary)]">
                    {{ item.description }}
                  </div>
                </div>
                <span class="text-xs text-[var(--text-tertiary)]">{{ item.project_name ?? "" }}</span>
              </div>
            </div>
          </template>
        </section>
      </div>
    </template>

    <!-- 初始空态 -->
    <AppEmptyState
      v-else
      title="搜索工作空间内容"
      description="输入关键词搜索需求/任务/缺陷、迭代、版本和项目。"
    />
  </div>
</template>
