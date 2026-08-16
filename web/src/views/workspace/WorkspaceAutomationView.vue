<script setup lang="ts">
/**
 * 工作空间级「自动化」跳板视图。
 *
 * 后端无工作空间级自动化聚合端点，且 ProjectCompareItem
 * 不包含自动化指标。这里只渲染项目基本信息 + 跳转按钮，
 * 跳转目标为项目级自动化页（/:wsId/projects/:projectId/automation）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { workspaceApi, type Project } from "@/api/services/workspace";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const projects = ref<Project[]>([]);

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const list = await workspaceApi.listProjects(workspaceId.value);
    projects.value = list ?? [];
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">自动化</h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          本页面汇总工作空间下所有项目的自动化信息。点击项目卡片进入详细视图。
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
      <AppSkeleton v-for="i in 4" :key="i" variant="card" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else>
      <section v-if="projects.length === 0">
        <AppEmptyState
          scenario="automation"
          title="工作空间下暂无项目"
          description="请先创建一个项目，再来配置自动化规则。"
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
                <div v-if="p.description" class="mt-1 line-clamp-2 text-xs text-[var(--text-secondary)]">
                  {{ p.description }}
                </div>
              </div>
            </div>

            <div class="mt-4 border-t border-[var(--border-subtle)] pt-3">
              <router-link
                :to="`/${workspaceId}/projects/${p.id}/automation`"
                class="inline-flex items-center gap-1 rounded-sm px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
                style="background: var(--brand-500)"
              >
                自动化规则 <span aria-hidden="true">→</span>
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
