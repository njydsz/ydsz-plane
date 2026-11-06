<script setup lang="ts">
/**
 * DataJobsView — 数据任务监控页。
 *
 * 展示异步任务（导入/导出/同步/报表）的进度、状态与错误信息。
 * 支持创建任务、手动标记成功/失败/取消，以及自动刷新运行中任务。
 */
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  DATA_JOB_STATUS_COLORS,
  DATA_JOB_STATUS_LABELS,
  DATA_JOB_TYPE_LABELS,
  dataJobApi,
  type CreateDataJobInput,
  type DataJob,
  type DataJobStatus,
} from "@/api/services/datajob";
import { AppEmptyState, AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";
import { useConfirm } from "@/composables/useConfirm";

const confirm = useConfirm();
const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const jobs = ref<DataJob[]>([]);
const filterStatus = ref<DataJobStatus | "">("");

/** 新建任务弹窗。 */
const showCreateModal = ref(false);
const createForm = ref<CreateDataJobInput>({
  job_type: "export",
  name: "",
});
const saving = ref(false);

/** 自动刷新：运行中/待执行的任务每 5 秒拉取一次。 */
let refreshTimer: ReturnType<typeof setInterval> | null = null;

const hasActiveJobs = computed(() =>
  jobs.value.some((j) => j.status === "pending" || j.status === "running"),
);

const filteredJobs = computed(() => {
  if (!filterStatus.value) return jobs.value;
  return jobs.value.filter((j) => j.status === filterStatus.value);
});

async function load() {
  if (!workspaceId.value) {
    loading.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const params: { status?: DataJobStatus } = {};
    if (filterStatus.value) params.status = filterStatus.value;
    jobs.value = await dataJobApi.list(workspaceId.value, { limit: 200, ...params });
    startAutoRefresh();
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function startAutoRefresh() {
  stopAutoRefresh();
  if (hasActiveJobs.value) {
    refreshTimer = setInterval(() => {
      void load();
    }, 5000);
  }
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

async function openCreateModal() {
  createForm.value = { job_type: "export", name: "" };
  showCreateModal.value = true;
}

async function submitCreate() {
  const f = createForm.value;
  if (!f.name.trim()) {
    toast.error("请填写任务名称");
    return;
  }
  saving.value = true;
  try {
    await dataJobApi.create(workspaceId.value, f);
    showCreateModal.value = false;
    toast.success("任务已创建");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "创建失败");
  } finally {
    saving.value = false;
  }
}

async function markRunning(job: DataJob) {
  try {
    await dataJobApi.markRunning(workspaceId.value, job.id);
    toast.success("任务已开始执行");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function markSuccess(job: DataJob) {
  try {
    await dataJobApi.markSuccess(workspaceId.value, job.id);
    toast.success("任务已标记成功");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function markFailed(job: DataJob) {
  const message = prompt("请输入失败原因：");
  if (!message) return;
  try {
    await dataJobApi.markFailed(workspaceId.value, job.id, message);
    toast.success("任务已标记失败");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function cancelJob(job: DataJob) {
  if (!await confirm({ title: "取消任务", message: `确定取消任务「${job.name}」？`, confirmText: "取消", variant: "danger" })) return;
  try {
    await dataJobApi.cancel(workspaceId.value, job.id);
    toast.success("任务已取消");
    await load();
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

async function removeJob(job: DataJob) {
  if (!await confirm({ title: "删除任务", message: `确定删除任务「${job.name}」？`, confirmText: "删除", variant: "danger" })) return;
  try {
    await dataJobApi.remove(workspaceId.value, job.id);
    jobs.value = jobs.value.filter((j) => j.id !== job.id);
    toast.success("任务已删除");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "操作失败");
  }
}

function formatDuration(ms?: number): string {
  if (ms == null || ms <= 0) return "-";
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.round(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

onMounted(load);
onUnmounted(stopAutoRefresh);
</script>

<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">数据任务</h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          监控异步作业（导入 / 导出 / 同步 / 报表）的执行进度与状态
          <span v-if="hasActiveJobs" class="ml-2 inline-flex items-center gap-1 text-[var(--brand-600)]">
            <span class="inline-block h-2 w-2 animate-pulse rounded-full bg-[var(--brand-500)]" />
            实时刷新中
          </span>
        </p>
      </div>
      <button class="rounded-md bg-[var(--brand-500)] px-3 py-1.5 text-sm text-white hover:opacity-90" @click="openCreateModal">
        + 新建任务
      </button>
    </header>

    <!-- 过滤器 -->
    <div class="flex items-center gap-2">
      <button
        :class="['rounded-md border px-3 py-1 text-xs', filterStatus === '' ? 'border-[var(--brand-500)] bg-[var(--surface-2)]' : 'border-[var(--border-subtle)]']"
        @click="filterStatus = ''; load()"
      >
        全部
      </button>
      <button
        v-for="(label, status) in DATA_JOB_STATUS_LABELS"
        :key="status"
        :class="['rounded-md border px-3 py-1 text-xs', filterStatus === status ? 'border-[var(--brand-500)] bg-[var(--surface-2)]' : 'border-[var(--border-subtle)]']"
        @click="filterStatus = status; load()"
      >
        {{ label }}
      </button>
    </div>

    <AppSkeleton v-if="loading" variant="table" :rows="4" />
    <AppErrorState v-else-if="error" :message="error" @retry="load" />
    <AppEmptyState v-else-if="filteredJobs.length === 0" title="暂无数据任务" description="点击右上角「新建任务」创建一个异步作业" />

    <div v-else class="overflow-hidden rounded-lg border border-[var(--border-subtle)]">
      <table class="w-full text-sm">
        <thead class="bg-[var(--surface-2)] text-xs uppercase tracking-wider text-[var(--text-tertiary)]">
          <tr>
            <th class="px-4 py-2.5 text-left font-medium">任务</th>
            <th class="px-4 py-2.5 text-left font-medium">类型</th>
            <th class="px-4 py-2.5 text-left font-medium">进度</th>
            <th class="px-4 py-2.5 text-left font-medium">状态</th>
            <th class="px-4 py-2.5 text-left font-medium">耗时</th>
            <th class="px-4 py-2.5 text-right font-medium">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border-subtle)]">
          <tr v-for="job in filteredJobs" :key="job.id" class="hover:bg-[var(--surface-2)]">
            <td class="px-4 py-3">
              <div class="font-medium text-[var(--text-primary)]">{{ job.name }}</div>
              <div v-if="job.error_message" class="mt-0.5 text-xs text-[var(--danger)]">
                {{ job.error_message }}
              </div>
            </td>
            <td class="px-4 py-3 text-[var(--text-secondary)]">
              {{ DATA_JOB_TYPE_LABELS[job.job_type] ?? job.job_type }}
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-[var(--surface-3)]">
                  <div
                    class="h-full rounded-full transition-all"
                    :style="{
                      width: `${job.progress}%`,
                      backgroundColor: DATA_JOB_STATUS_COLORS[job.status],
                    }"
                  />
                </div>
                <span class="w-8 text-right text-xs tabular-nums text-[var(--text-tertiary)]">
                  {{ job.progress }}%
                </span>
              </div>
            </td>
            <td class="px-4 py-3">
              <span
                class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium"
                :style="{ color: DATA_JOB_STATUS_COLORS[job.status], backgroundColor: DATA_JOB_STATUS_COLORS[job.status] + '1a' }"
              >
                {{ DATA_JOB_STATUS_LABELS[job.status] ?? job.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-xs tabular-nums text-[var(--text-secondary)]">
              {{ formatDuration(job.duration_ms) }}
            </td>
            <td class="px-4 py-3 text-right">
              <div class="flex justify-end gap-2">
                <button
                  v-if="job.status === 'pending'"
                  class="text-xs text-[var(--brand-600)] hover:underline"
                  @click="markRunning(job)"
                >
                  开始
                </button>
                <button
                  v-if="job.status === 'running'"
                  class="text-xs text-[var(--success-500)] hover:underline"
                  @click="markSuccess(job)"
                >
                  成功
                </button>
                <button
                  v-if="job.status === 'running'"
                  class="text-xs text-[var(--danger)] hover:underline"
                  @click="markFailed(job)"
                >
                  失败
                </button>
                <button
                  v-if="job.status === 'pending' || job.status === 'running'"
                  class="text-xs text-[var(--text-tertiary)] hover:underline"
                  @click="cancelJob(job)"
                >
                  取消
                </button>
                <button
                  class="text-xs text-[var(--text-tertiary)] hover:underline"
                  @click="removeJob(job)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 新建任务弹窗 -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-6" @click.self="showCreateModal = false">
      <div class="w-full max-w-md rounded-xl border border-[var(--border-subtle)] bg-[var(--surface-1)] shadow-xl">
        <h2 class="m-0 border-b border-[var(--border-subtle)] px-5 py-4 text-base font-semibold">新建数据任务</h2>
        <div class="space-y-3 px-5 py-4">
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>任务名称 *</span>
            <input
              v-model="createForm.name"
              type="text"
              placeholder="如：导出 Q4 需求清单"
              class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]"
            />
          </label>
          <label class="flex flex-col gap-1 text-sm text-[var(--text-secondary)]">
            <span>任务类型</span>
            <select
              v-model="createForm.job_type"
              class="rounded-md border border-[var(--border-subtle)] bg-[var(--surface-1)] px-2.5 py-1.5 text-sm text-[var(--text-primary)]"
            >
              <option v-for="(label, key) in DATA_JOB_TYPE_LABELS" :key="key" :value="key">{{ label }}</option>
            </select>
          </label>
        </div>
        <div class="flex justify-end gap-2 border-t border-[var(--border-subtle)] px-5 py-3">
          <button class="rounded-md border border-[var(--border-subtle)] px-3 py-1.5 text-sm hover:bg-[var(--surface-2)]" @click="showCreateModal = false">
            取消
          </button>
          <button
            class="rounded-md bg-[var(--brand-500)] px-3 py-1.5 text-sm text-white disabled:opacity-50"
            :disabled="saving"
            @click="submitCreate"
          >
            {{ saving ? "创建中…" : "创建" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
