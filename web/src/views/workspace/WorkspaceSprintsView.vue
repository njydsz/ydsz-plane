<script setup lang="ts">
/**
 * 工作空间级「迭代」跳板视图。
 *
 * 渲染项目选择器 + 每个项目的迭代模块摘要。点击卡片跳转到
 * 项目级迭代列表（/:wsId/projects/:projectId/sprints）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { workspaceApi, type Project } from "@/api/services/workspace";
import { dashboardApi, type ProjectCompareItem } from "@/api/services/dashboard";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

interface ProjectCard extends Project {
  compare?: ProjectCompareItem;
}

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const projects = ref<ProjectCard[]>([]);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const [list, compare] = await Promise.all([
      workspaceApi.listProjects(workspaceId.value),
      dashboardApi.getProjectCompare(workspaceId.value).catch(() => [] as ProjectCompareItem[]),
    ]);
    const compareMap = new Map<number, ProjectCompareItem>(
      (compare ?? []).map((c) => [c.project_id, c]),
    );
    projects.value = (list ?? []).map((p) => ({
      ...p,
      compare: compareMap.get(p.id),
    }));
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function sprintEnabled(p: ProjectCard): boolean {
  return p.modules?.sprint !== false;
}

function percent(n: number | undefined): number {
  if (!n || Number.isNaN(n)) return 0;
  return Math.round(Math.max(0, Math.min(1, n)) * 100);
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">迭代</h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          本页面汇总工作空间下所有项目的迭代信息。点击项目卡片进入详细视图。
        </p>
      </div>
      <button
        class="text-sm text-[var(--brand-500)] hover:underline"
        @click="load"
      >
        刷新
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 6" :key="i" variant="card" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <section v-if="projects.length === 0">
        <AppEmptyState
          scenario="projects"
          title="工作空间下暂无项目"
          description="请先创建一个项目，再来汇总迭代信息。"
        />
      </section>

      <section v-else>
        <h2 class="mb-3 text-sm font-semibold text-[var(--text-secondary)]">
          项目（按创建顺序，共 {{ projects.length }} 个）
        </h2>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          <div
            v-for="p in projects"
            :key="p.id"
            class="flex flex-col rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] p-4 transition hover:border-[var(--brand-500)]"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <span class="font-mono text-xs text-[var(--brand-500)]">
                  {{ p.identifier }}
                </span>
                <div class="mt-0.5 truncate text-sm font-medium text-[var(--text-primary)]">
                  {{ p.name }}
                </div>
              </div>
            </div>

            <div class="mt-3 grid grid-cols-2 gap-2 text-xs">
              <div>
                <div class="text-[var(--text-tertiary)]">进行中迭代</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare?.active_sprint_count ?? 0 }}
                </div>
              </div>
              <div>
                <div class="text-[var(--text-tertiary)]">完成率</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare ? `${percent(p.compare.completion_rate)}%` : "—" }}
                </div>
              </div>
              <div>
                <div class="text-[var(--text-tertiary)]">工作项</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare ? `${p.compare.done_issues}/${p.compare.total_issues}` : "—" }}
                </div>
              </div>
              <div>
                <div class="text-[var(--text-tertiary)]">缺陷数</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare?.defect_count ?? "—" }}
                </div>
              </div>
            </div>

            <div class="mt-4 border-t border-[var(--border-subtle)] pt-3">
              <div v-if="!sprintEnabled(p)" class="text-xs text-[var(--text-tertiary)]">
                该项目未启用迭代模块
              </div>
              <router-link
                v-else
                :to="`/${workspaceId}/projects/${p.id}/sprints`"
                class="inline-flex items-center gap-1 rounded-sm px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
                style="background: var(--brand-500)"
              >
                迭代 <span aria-hidden="true">→</span>
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
