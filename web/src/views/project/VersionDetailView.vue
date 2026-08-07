<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version } from "@/api/services/version";
import DefectPanel from "./DefectPanel.vue";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const versionId = computed(() => Number(route.params.versionId));
const activeTab = ref<"overview" | "sprints" | "defects" | "notes">("overview");

const version = ref<Version | null>(null);
const loading = ref(true);
const error = ref("");
const actionError = ref("");
const showAddSprint = ref(false);
const availableSprints = ref<{ id: number; name: string }[]>([]);

let wsIdVal = 0;

async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.getBySlug(workspaceSlug.value);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    version.value = await versionApi.getVersion(wsId, projectId.value, versionId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载版本详情失败";
  } finally {
    loading.value = false;
  }
}

async function transition(action: "activate" | "release" | "archive") {
  actionError.value = "";
  try {
    const wsId = await resolveWsId();
    if (action === "activate") {
      await versionApi.activateVersion(wsId, projectId.value, versionId.value);
    } else if (action === "archive") {
      await versionApi.archiveVersion(wsId, projectId.value, versionId.value);
    } else if (action === "release") {
      router.push({ name: "version-release", params: { versionId: versionId.value } });
      return;
    }
    await load();
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "操作失败";
  }
}

async function toggleChecklist(itemId: string) {
  if (!version.value) return;
  const v = version.value;
  const newList = (v.checklist ?? []).map((c) =>
    c.id === itemId ? { ...c, checked: !c.checked } : c,
  );
  const wsId = await resolveWsId();
  try {
    await versionApi.updateVersion(wsId, projectId.value, versionId.value, {
      checklist: newList,
      version: v.version ?? 0,
    } as any);
    version.value = await versionApi.getVersion(wsId, projectId.value, versionId.value);
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : "保存失败";
  }
}

async function refreshAvailableSprints() {
  try {
    const { sprintApi } = await import("@/api/services/sprint");
    const wsId = await resolveWsId();
    const res = await sprintApi.listSprints(wsId, projectId.value, { status: "active" });
    const attachedIds = new Set((version.value?.sprints ?? []).map((s) => s.sprint_id));
    availableSprints.value = res.results
      .filter((s: any) => !attachedIds.has(s.id))
      .map((s: any) => ({ id: s.id, name: s.name }));
  } catch {
    availableSprints.value = [];
  }
}

async function addSprint(sprintId: number) {
  const wsId = await resolveWsId();
  await versionApi.addSprint(wsId, projectId.value, versionId.value, { sprint_id: sprintId });
  showAddSprint.value = false;
  await load();
}

async function removeSprint(sprintId: number) {
  const wsId = await resolveWsId();
  await versionApi.removeSprint(wsId, projectId.value, versionId.value, sprintId);
  await load();
}

function goWizard() {
  router.push({ name: "version-release", params: { versionId: versionId.value } });
}

onMounted(async () => {
  await load();
  await refreshAvailableSprints();
});
</script>

<template>
  <div v-if="error" class="text-red-500">{{ error }}</div>
  <div v-else-if="loading">加载中…</div>
  <div v-else-if="version" class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold">{{ version.name }}</h1>
        <p class="text-sm text-gray-500">semver: {{ version.semver }} · status: {{ version.status }}</p>
      </div>
      <div class="flex gap-2">
        <button
          v-if="version.status === 'planning'"
          class="rounded bg-amber-600 px-3 py-1 text-sm text-white"
          @click="transition('activate')"
        >
          启动
        </button>
        <button
          v-if="version.status === 'active'"
          class="rounded bg-green-600 px-3 py-1 text-sm text-white"
          @click="goWizard"
        >
          发布…
        </button>
        <button
          v-if="version.status !== 'archived'"
          class="rounded border px-3 py-1 text-sm"
          @click="transition('archive')"
        >
          归档
        </button>
      </div>
    </div>
    <p v-if="actionError" class="text-red-500 text-sm">{{ actionError }}</p>

    <div class="flex gap-4 border-b text-sm">
      <button
        v-for="t in (['overview', 'sprints', 'defects', 'notes'] as const)"
        :key="t"
        class="px-3 py-2"
        :class="activeTab === t ? 'border-b-2 border-blue-600 font-medium' : 'text-gray-500'"
        @click="activeTab = t"
      >
        {{ { overview: "概览", sprints: "迭代", defects: "缺陷", notes: "Release Notes" }[t] }}
      </button>
    </div>

    <!-- Overview -->
    <div v-if="activeTab === 'overview'" class="space-y-4">
      <div class="grid grid-cols-2 gap-4">
        <div class="rounded border p-3">
          <div class="text-xs text-gray-500">进度</div>
          <div class="text-lg font-semibold">
            {{ Math.round((version.progress?.completion_rate ?? 0) * 100) }}%
          </div>
          <div class="text-xs text-gray-400">
            {{ version.progress?.done_issues }}/{{ version.progress?.total_issues }} 工作项 ·
            {{ version.progress?.done_points }}/{{ version.progress?.total_points }} 点
          </div>
        </div>
        <div class="rounded border p-3">
          <div class="text-xs text-gray-500">质量</div>
          <div class="text-lg font-semibold">
            {{ version.quality?.pass_quality_gate ? "通过" : "未通过" }}
          </div>
          <div class="text-xs text-gray-400">
            致命/严重未关闭 {{ version.quality?.critical_bugs ?? 0 }} ·
            修复率 {{ Math.round((version.quality?.fix_rate ?? 0) * 100) }}%
          </div>
        </div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500 mb-2">发布检查清单</div>
        <ul class="space-y-1 text-sm">
          <li v-for="item in version.checklist" :key="item.id" class="flex items-center gap-2">
            <input
              type="checkbox"
              :checked="item.checked"
              :disabled="version.status === 'released' || version.status === 'archived'"
              @change="toggleChecklist(item.id)"
            />
            <span :class="{ 'text-gray-400': item.checked }">
              {{ item.label }}
              <span v-if="item.required" class="text-red-500">*</span>
            </span>
          </li>
          <li v-if="!version.checklist?.length" class="text-gray-400">无检查项</li>
        </ul>
      </div>
    </div>

    <!-- Sprints -->
    <div v-if="activeTab === 'sprints'" class="space-y-2">
      <div class="flex justify-between items-center">
        <span class="text-sm text-gray-500">{{ version.sprints?.length ?? 0 }} 个迭代</span>
        <button
          v-if="version.status === 'planning' || version.status === 'active'"
          class="rounded border px-2 py-1 text-xs"
          @click="showAddSprint = !showAddSprint"
        >
          + 挂入迭代
        </button>
      </div>
      <div v-if="showAddSprint" class="rounded border bg-gray-50 p-2">
        <select class="rounded border px-2 py-1 text-sm" @change="(e) => addSprint(Number((e.target as HTMLSelectElement).value))">
          <option value="">选择迭代…</option>
          <option v-for="s in availableSprints" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>
      <ul class="space-y-1">
        <li
          v-for="s in version.sprints"
          :key="s.sprint_id"
          class="flex items-center justify-between rounded border px-3 py-2"
        >
          <div>
            <span class="font-medium">{{ s.name }}</span>
            <span class="ml-2 text-xs text-gray-500">{{ s.status }}</span>
          </div>
          <button
            v-if="version.status === 'planning' || version.status === 'active'"
            class="text-xs text-red-500"
            @click="removeSprint(s.sprint_id)"
          >
            解绑
          </button>
        </li>
      </ul>
    </div>

    <!-- Defects -->
    <div v-if="activeTab === 'defects'">
      <DefectPanel
        :workspace-slug="workspaceSlug"
        :project-id="projectId"
        :version-id="versionId"
      />
    </div>

    <!-- Notes -->
    <div v-if="activeTab === 'notes'" class="space-y-2">
      <div class="text-sm text-gray-500">
        <button class="rounded border px-2 py-1 text-xs" @click="regenerate">重新生成</button>
      </div>
      <pre class="whitespace-pre-wrap rounded border bg-gray-50 p-3 text-sm font-mono">{{ version.release_notes || "（尚未生成 Release Notes）" }}</pre>
    </div>
  </div>
</template>
