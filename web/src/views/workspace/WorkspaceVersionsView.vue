<script setup lang="ts">
/**
 * 工作空间级「版本」跳板视图。
 *
 * 后端无工作空间级版本聚合端点，所以这里渲染项目选择器 +
 * 每个项目的版本模块摘要。点击卡片跳转到项目级版本列表
 * （/:wsId/projects/:projectId/versions）。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { workspaceApi, type Project } from "@/api/services/workspace";
import { dashboardApi, type ProjectCompareItem } from "@/api/services/dashboard";
import { versionApi, type CreateVersionInput } from "@/api/services/version";
import { AppButton, AppEmptyState, AppErrorState, AppModal, AppSkeleton } from "@/components";

interface ProjectCard extends Project {
  compare?: ProjectCompareItem;
}

const route = useRoute();
const router = useRouter();
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

function versionEnabled(p: ProjectCard): boolean {
  return p.modules?.version !== false;
}

function percent(n: number | undefined): number {
  if (!n || Number.isNaN(n)) return 0;
  return Math.round(Math.max(0, Math.min(1, n)) * 100);
}

/* ---------- create version modal ---------- */

const showCreate = ref(false);
const creating = ref(false);
const createError = ref("");
const targetProjectId = ref<number | null>(null);
const form = ref<CreateVersionInput>({
  name: "",
  semver: "",
  description: "",
  start_date: "",
  end_date: "",
  target_date: "",
});

const enabledProjects = computed(() => projects.value.filter(versionEnabled));

function openCreateModal() {
  createError.value = "";
  if (enabledProjects.value.length === 1) {
    targetProjectId.value = enabledProjects.value[0].id;
  } else {
    targetProjectId.value = null;
  }
  form.value = { name: "", semver: "", description: "", start_date: "", end_date: "", target_date: "" };
  showCreate.value = true;
}

function closeCreateModal() {
  showCreate.value = false;
  targetProjectId.value = null;
  createError.value = "";
}

async function submitCreate() {
  if (!form.value.name.trim() || !form.value.semver.trim()) {
    createError.value = "版本名称和语义版本号不能为空";
    return;
  }
  const pid = targetProjectId.value;
  if (!pid) {
    createError.value = "请选择目标项目";
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    await versionApi.createVersion(workspaceId.value, pid, {
      name: form.value.name.trim(),
      semver: form.value.semver.trim(),
      description: form.value.description || undefined,
      start_date: form.value.start_date || undefined,
      end_date: form.value.end_date || undefined,
      target_date: form.value.target_date || undefined,
    });
    showCreate.value = false;
    // 刷新工作空间版本聚合数据
    await load();
    // 创建成功后跳转到该项目版本列表
    router.push(`/${workspaceId.value}/projects/${pid}/versions`);
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    creating.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">版本</h1>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          本页面汇总工作空间下所有项目的版本信息。点击项目卡片进入详细视图。
        </p>
      </div>
      <button
        class="text-sm font-medium text-[var(--bg-accent-primary)] hover:underline disabled:opacity-40 disabled:no-underline"
        :disabled="!projects.some(versionEnabled)"
        @click="openCreateModal"
      >
        新建版本
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
          description="请先创建一个项目，再来汇总版本信息。"
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

            <div class="mt-3 grid grid-cols-3 gap-2 text-xs">
              <div>
                <div class="text-[var(--text-tertiary)]">需求/任务/缺陷</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare?.total_issues ?? "—" }}
                </div>
              </div>
              <div>
                <div class="text-[var(--text-tertiary)]">完成率</div>
                <div class="mt-0.5 text-sm font-medium text-[var(--text-primary)]">
                  {{ p.compare ? `${percent(p.compare.completion_rate)}%` : "—" }}
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
              <div v-if="!versionEnabled(p)" class="text-xs text-[var(--text-tertiary)]">
                该项目未启用版本模块
              </div>
              <router-link
                v-else
                :to="`/${workspaceId}/projects/${p.id}/versions`"
                class="inline-flex items-center gap-1 rounded-md bg-[var(--bg-accent-primary)] px-3 py-1.5 text-sm font-medium text-white hover:bg-[var(--bg-accent-primary-hover)] active:bg-[var(--bg-accent-primary-active)] disabled:opacity-50"
              >
                版本 <span aria-hidden="true">→</span>
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </template>

    <!-- 新建版本 Modal -->
    <AppModal :visible="showCreate" title="新建版本" width="560px" @close="closeCreateModal">
      <form @submit.prevent="submitCreate">
        <!-- 多项目时选择目标项目 -->
        <label v-if="enabledProjects.length > 1" class="create-form__field">
          <span class="create-form__label">
            目标项目 <span class="create-form__required">*</span>
          </span>
          <select v-model="targetProjectId" class="create-form__input">
            <option :value="null" disabled>请选择项目</option>
            <option v-for="p in enabledProjects" :key="p.id" :value="p.id">
              {{ p.name }} ({{ p.identifier }})
            </option>
          </select>
        </label>

        <div class="create-form__row">
          <label class="create-form__field">
            <span class="create-form__label">
              版本名称 <span class="create-form__required">*</span>
            </span>
            <input
              v-model="form.name"
              placeholder="例如：v1.0 正式版"
              maxlength="120"
              class="create-form__input"
              autofocus
            />
          </label>
          <label class="create-form__field">
            <span class="create-form__label">
              语义版本号 <span class="create-form__required">*</span>
            </span>
            <input
              v-model="form.semver"
              placeholder="例如：1.0.0"
              maxlength="50"
              class="create-form__input create-form__input--mono"
            />
          </label>
        </div>

        <label class="create-form__field">
          <span class="create-form__label">目标日期</span>
          <input v-model="form.target_date" type="date" class="create-form__input" />
        </label>

        <div class="create-form__row">
          <label class="create-form__field">
            <span class="create-form__label">开始时间</span>
            <input v-model="form.start_date" type="date" class="create-form__input" />
          </label>
          <label class="create-form__field">
            <span class="create-form__label">结束时间</span>
            <input v-model="form.end_date" type="date" class="create-form__input" />
          </label>
        </div>

        <label class="create-form__field">
          <span class="create-form__label">描述（可选）</span>
          <textarea
            v-model="form.description"
            placeholder="版本目标与范围简述"
            maxlength="2000"
            rows="2"
            class="create-form__input"
          ></textarea>
        </label>

        <div v-if="createError" class="create-form__error">{{ createError }}</div>
      </form>

      <template #footer>
        <AppButton variant="secondary" size="sm" @click="closeCreateModal">取消</AppButton>
        <AppButton variant="primary" size="sm" :loading="creating" @click="submitCreate">创建</AppButton>
      </template>
    </AppModal>
  </div>
</template>

<style scoped>
/* 复用 VersionListView 中 create-form 的样式 */
.create-form__field {
  display: block;
  margin-bottom: 14px;
}
.create-form__label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  color: var(--text-secondary);
}
.create-form__required { color: var(--danger-500); }
.create-form__input {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--surface-1);
  color: var(--text-primary);
  font-size: 13px;
  transition: border-color 0.15s;
}
textarea.create-form__input {
  height: auto;
  padding: 8px 12px;
  resize: vertical;
  font-family: inherit;
}
.create-form__input--mono { font-family: var(--font-mono); }
.create-form__input:focus {
  outline: none;
  border-color: var(--brand-500);
}
.create-form__input::placeholder { color: var(--text-tertiary); }
.create-form__row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.create-form__error {
  color: var(--danger-500);
  font-size: 12px;
  margin-bottom: 8px;
}
</style>
