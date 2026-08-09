<script setup lang="ts">
/**
 * 我的工作项 — 跨项目聚合当前登录用户被全部分配的工作项。
 *
 * 由于后端 Issue API 限定在项目内，本组件先拉取工作空间下所有项目，
 * 再对每个项目按 assignee_id 过滤并聚合展示，按更新时间倒序。
 *
 * MVP 阶段：最多加载前 200 条，不做虚拟滚动；后续可升级为后端工作空间级聚合接口。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { type Issue, type ListIssuesParams, issueApi } from "@/api/services/issue";
import { workspaceApi, type Project } from "@/api/services/workspace";
import { useAuthStore } from "@/stores/auth";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
const loading = ref(true);
const error = ref("");
const issues = ref<Issue[]>([]);
const projects = ref<Project[]>([]);

/** 当前用户 ID — 仅已登录有效 */
const currentUserId = computed(() => auth.user?.id ?? 0);

/** 生效的 workspaceId：路由无效时直接用 0 提示空态 */
const effectiveWsId = computed(() => workspaceId.value);

/** 加载工作空间内所有项目 + 各自分配给当前用户的工作项 */
async function load() {
  if (!effectiveWsId.value || !currentUserId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  issues.value = [];
  try {
    // 1) 拉取空间内所有项目
    projects.value = await workspaceApi.listProjects(effectiveWsId.value);

    // 2) 对每个项目并发拉取分配给当前用户的工作项
    const LIST_VIEW_FIELDS = ["id", "identifier", "name", "state_id", "priority", "type_code", "assignees", "project_id", "updated_at"];
    const params: ListIssuesParams = {
      assignee_id: currentUserId.value,
      sort: "-updated_at",
      limit: 50,
    };

    const settled = await Promise.allSettled(
      projects.value.map((p) =>
        issueApi.listIssues(effectiveWsId.value, p.id, params, LIST_VIEW_FIELDS),
      ),
    );

    const all: Issue[] = [];
    for (const r of settled) {
      if (r.status === "fulfilled") {
        all.push(...r.value.results);
      }
    }
    // 按 updated_at 倒序 + 截断
    all.sort((a, b) => (b.updated_at ?? "").localeCompare(a.updated_at ?? ""));
    issues.value = all.slice(0, 200);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

/** 跳转到工作项详情页 */
function openIssue(issue: Issue) {
  router.push(`/${effectiveWsId.value}/projects/${issue.project_id}/issues/${issue.id}`);
}

/** 工作项类型中文映射 */
function typeLabel(code?: string): string {
  const map: Record<string, string> = {
    epic: "史诗",
    requirement: "需求",
    task: "任务",
    defect: "缺陷",
  };
  return map[code ?? ""] ?? code ?? "";
}

/** 优先级中文映射 */
function priorityLabel(p?: string): string {
  const map: Record<string, string> = {
    urgent: "紧急",
    high: "高",
    medium: "中",
    low: "低",
    none: "-",
  };
  return map[p ?? ""] ?? p ?? "";
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">{{ $t("menu.myIssues") }}</h1>
      <span v-if="issues.length > 0" class="text-xs text-muted-foreground">
        {{ issues.length }} 项
      </span>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 5" :key="i" class="h-12 w-full" />
    </div>

    <!-- 错误 -->
    <AppErrorState
      v-else-if="error"
      :message="error"
      @retry="load"
    />

    <!-- 无效工作空间 ID -->
    <AppEmptyState
      v-else-if="!effectiveWsId"
      title="请先选择工作空间"
      description="选择一个工作空间后即可查看分配给您的工作项。"
    />

    <!-- 空数据 -->
    <AppEmptyState
      v-else-if="issues.length === 0"
      title="暂无分配给您的工作项"
      description="当有工作项被指派给您时，将在此处显示。"
    />

    <!-- 工作项列表 -->
    <div v-else class="rounded-md border border-[var(--border-subtle)] overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-[var(--surface-2)] text-[var(--text-tertiary)] text-xs uppercase tracking-wider">
          <tr>
            <th class="px-4 py-2 text-left font-medium">编号</th>
            <th class="px-4 py-2 text-left font-medium">标题</th>
            <th class="hidden sm:table-cell px-4 py-2 text-left font-medium">类型</th>
            <th class="hidden sm:table-cell px-4 py-2 text-left font-medium">优先级</th>
            <th class="hidden md:table-cell px-4 py-2 text-left font-medium">更新时间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border-subtle)]">
          <tr
            v-for="issue in issues"
            :key="issue.id"
            class="hover:bg-[var(--surface-2)] cursor-pointer transition-colors"
            @click="openIssue(issue)"
          >
            <td class="px-4 py-2.5 font-mono text-xs text-[var(--brand-600)]">
              {{ issue.identifier }}
            </td>
            <td class="px-4 py-2.5 text-[var(--text-primary)] truncate max-w-[400px]">
              {{ issue.name }}
            </td>
            <td class="hidden sm:table-cell px-4 py-2.5 text-[var(--text-secondary)]">
              {{ typeLabel(issue.type_code) }}
            </td>
            <td class="hidden sm:table-cell px-4 py-2.5 text-[var(--text-secondary)]">
              {{ priorityLabel(issue.priority) }}
            </td>
            <td class="hidden md:table-cell px-4 py-2.5 text-[var(--text-tertiary)] text-xs">
              {{ issue.updated_at?.slice(0, 10) ?? "-" }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
