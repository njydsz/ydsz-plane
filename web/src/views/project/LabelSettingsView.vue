<script setup lang="ts">
/**
 * LabelSettingsView — 标签管理页。
 * 展示项目下全部标签，支持新建、行内编辑、删除，并显示各标签被使用次数。
 */
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { labelApi, type Label } from "@/api/services/label";
import { AppEmptyState, AppErrorState, AppLoadingState } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = Number(route.params.workspaceId);
const projectId = Number(route.params.projectId);

const labels = ref<Label[]>([]);
const loading = ref(true);
const error = ref("");

// 预设颜色
const COLOR_PRESETS = [
  "#ef4444", "#f97316", "#eab308", "#22c55e", "#14b8a6",
  "#3b82f6", "#6366f1", "#a855f7", "#ec4899", "#64748b",
];

// 新建
const showCreate = ref(false);
const createForm = ref({ name: "", color: "#3b82f6", description: "" });
const creating = ref(false);

// 行内编辑
const editingId = ref<number | null>(null);
const editForm = ref({ name: "", color: "#3b82f6", description: "" });
const saving = ref(false);

async function loadLabels() {
  loading.value = true;
  error.value = "";
  try {
    const r = await labelApi.list(workspaceId, projectId);
    labels.value = r.results ?? [];
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  showCreate.value = true;
  createForm.value = { name: "", color: "#3b82f6", description: "" };
}

async function doCreate() {
  if (!createForm.value.name.trim()) { toast.warning("请输入标签名称"); return; }
  creating.value = true;
  try {
    await labelApi.create(workspaceId, projectId, {
      name: createForm.value.name.trim(),
      color: createForm.value.color,
      description: createForm.value.description.trim() || undefined,
    });
    toast.success("标签创建成功");
    showCreate.value = false;
    await loadLabels();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "创建失败");
  } finally {
    creating.value = false;
  }
}

function startEdit(l: Label) {
  editingId.value = l.id;
  editForm.value = { name: l.name, color: l.color || "#3b82f6", description: l.description ?? "" };
}

async function saveEdit(l: Label) {
  if (!editForm.value.name.trim()) { toast.warning("标签名称不能为空"); return; }
  saving.value = true;
  try {
    await labelApi.update(workspaceId, projectId, l.id, {
      name: editForm.value.name.trim(),
      color: editForm.value.color,
      description: editForm.value.description.trim() || undefined,
    });
    toast.success("已保存");
    editingId.value = null;
    await loadLabels();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

async function remove(l: Label) {
  if (!confirm(`确定删除标签「${l.name}」？工作项上的标签关联会一并解除。`)) return;
  try {
    await labelApi.remove(workspaceId, projectId, l.id);
    toast.success("标签已删除");
    await loadLabels();
  } catch (e: unknown) {
    toast.error(e instanceof Error ? e.message : "删除失败");
  }
}

onMounted(loadLabels);
</script>

<template>
  <div class="mx-auto max-w-4xl space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">标签管理</h1>
        <p class="mt-1 text-sm text-[var(--text-tertiary)]">为需求/任务/缺陷提供轻量分类，可搭配颜色区分</p>
      </div>
      <button
        class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)]"
        @click="openCreate"
      >
        新建标签
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppLoadingState />
    </div>
    <AppErrorState v-else-if="error" :message="error" @retry="loadLabels" />

    <template v-else>
      <!-- 新建表单 -->
      <section v-if="showCreate" class="space-y-3 rounded-md border border-[var(--border-subtle)] p-4">
        <h2 class="text-sm font-semibold text-[var(--text-secondary)]">新建标签</h2>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">名称 *</label>
          <input v-model="createForm.name" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="如：前端 / 后端 / 待评审" />
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">颜色</label>
          <div class="mt-1 flex items-center gap-2">
            <input v-model="createForm.color" type="color" class="h-8 w-10 cursor-pointer rounded border border-[var(--border-subtle)] p-0.5" />
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="c in COLOR_PRESETS"
                :key="c"
                class="h-6 w-6 rounded-full border border-black/10"
                :class="{ 'ring-2 ring-[var(--brand-500)] ring-offset-1': createForm.color === c }"
                :style="{ backgroundColor: c }"
                @click="createForm.color = c"
              />
            </div>
          </div>
        </div>
        <div>
          <label class="text-xs text-[var(--text-tertiary)]">描述</label>
          <input v-model="createForm.description" class="mt-1 w-full rounded-md border border-[var(--border-subtle)] px-3 py-2 text-sm" placeholder="可选" />
        </div>
        <div class="flex gap-2">
          <button :disabled="creating" class="rounded-md bg-[var(--brand-600)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50" @click="doCreate">
            {{ creating ? "创建中…" : "创建" }}
          </button>
          <button class="rounded-md border border-[var(--border-subtle)] px-4 py-2 text-sm" @click="showCreate = false">取消</button>
        </div>
      </section>

      <!-- 标签列表 -->
      <AppEmptyState v-if="!labels.length" title="还没有标签" desc="创建标签后，可在工作项上快速标记分类" />
      <section v-else class="space-y-2">
        <div
          v-for="l in labels"
          :key="l.id"
          class="flex items-center justify-between rounded-md border border-[var(--border-subtle)] px-3 py-2.5"
        >
          <template v-if="editingId === l.id">
            <div class="flex flex-1 flex-wrap items-center gap-2">
              <input v-model="editForm.name" class="w-40 rounded-md border border-[var(--border-subtle)] px-2 py-1.5 text-sm" />
              <input v-model="editForm.color" type="color" class="h-7 w-9 cursor-pointer rounded border border-[var(--border-subtle)] p-0.5" />
              <input v-model="editForm.description" class="w-48 rounded-md border border-[var(--border-subtle)] px-2 py-1.5 text-sm" placeholder="描述（可选）" />
              <button :disabled="saving" class="rounded bg-[var(--brand-600)] px-3 py-1.5 text-xs text-white" @click="saveEdit(l)">保存</button>
              <button class="rounded border border-[var(--border-subtle)] px-3 py-1.5 text-xs" @click="editingId = null">取消</button>
            </div>
          </template>
          <template v-else>
            <div class="flex min-w-0 items-center gap-2">
              <span class="inline-block h-3 w-3 shrink-0 rounded-full" :style="{ backgroundColor: l.color || '#3b82f6' }" />
              <span class="truncate text-sm font-medium">{{ l.name }}</span>
              <span v-if="l.description" class="hidden truncate text-xs text-[var(--text-tertiary)] md:inline">{{ l.description }}</span>
            </div>
            <div class="flex shrink-0 items-center gap-3">
              <span class="text-xs text-[var(--text-tertiary)]">使用 {{ l.issue_count ?? 0 }}</span>
              <button class="text-xs text-[var(--text-secondary)] hover:text-[var(--brand-600)]" @click="startEdit(l)">编辑</button>
              <button class="text-xs text-[var(--danger,#ef4444)]" @click="remove(l)">删除</button>
            </div>
          </template>
        </div>
      </section>
    </template>
  </div>
</template>
