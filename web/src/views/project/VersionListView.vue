<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version, type VersionStatus } from "@/api/services/version";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));

const versions = ref<Version[]>([]);
const loading = ref(true);
const error = ref("");
const total = ref(0);
const filterStatus = ref<VersionStatus | "">("");
const showCreate = ref(false);

const form = ref({
  name: "",
  semver: "",
  description: "",
  target_date: "",
});
const creating = ref(false);
const createError = ref("");

let wsIdVal = 0;

const statusLabels: Record<VersionStatus, string> = {
  planning: "规划中",
  active: "进行中",
  released: "已发布",
  archived: "已归档",
};

const statusClass: Record<VersionStatus, string> = {
  planning: "bg-blue-100 text-blue-800",
  active: "bg-amber-100 text-amber-800",
  released: "bg-green-100 text-green-800",
  archived: "bg-gray-100 text-gray-600",
};

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
    const res = await versionApi.listVersions(wsId, projectId.value, {
      status: filterStatus.value || undefined,
      limit: 50,
      offset: 0,
    });
    versions.value = res.results;
    total.value = res.total;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载版本列表失败";
  } finally {
    loading.value = false;
  }
}

async function createVersion() {
  if (!form.value.name.trim() || !form.value.semver.trim()) {
    createError.value = "版本名和 semver 不能为空";
    return;
  }
  creating.value = true;
  createError.value = "";
  try {
    const wsId = await resolveWsId();
    await versionApi.createVersion(wsId, projectId.value, {
      name: form.value.name,
      semver: form.value.semver,
      description: form.value.description || undefined,
      target_date: form.value.target_date || undefined,
    });
    showCreate.value = false;
    form.value = { name: "", semver: "", description: "", target_date: "" };
    await load();
  } catch (e: unknown) {
    createError.value = e instanceof Error ? e.message : "创建失败";
  } finally {
    creating.value = false;
  }
}

function openDetail(v: Version) {
  router.push({
    name: "version-detail",
    params: { versionId: v.id },
  });
}

onMounted(load);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-semibold">版本日</h1>
        <p class="text-sm text-gray-500">管理项目的版本发布计划</p>
      </div>
      <div class="flex items-center gap-3">
        <select
          v-model="filterStatus"
          class="rounded border px-2 py-1 text-sm"
          @change="load"
        >
          <option value="">全部状态</option>
          <option value="planning">规划中</option>
          <option value="active">进行中</option>
          <option value="released">已发布</option>
          <option value="archived">已归档</option>
        </select>
        <button
          class="rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700"
          @click="showCreate = !showCreate"
        >
          + 新建版本
        </button>
      </div>
    </div>

    <div v-if="showCreate" class="rounded border bg-gray-50 p-4 space-y-3">
      <div class="grid grid-cols-2 gap-3">
        <input
          v-model="form.name"
          placeholder="版本名 (如 v1.0 正式版)"
          class="rounded border px-2 py-1 text-sm"
        />
        <input
          v-model="form.semver"
          placeholder="semver (如 1.0.0)"
          class="rounded border px-2 py-1 text-sm"
        />
        <input
          v-model="form.target_date"
          type="date"
          placeholder="目标日期"
          class="rounded border px-2 py-1 text-sm"
        />
        <input
          v-model="form.description"
          placeholder="描述 (可选)"
          class="rounded border px-2 py-1 text-sm col-span-2"
        />
      </div>
      <p v-if="createError" class="text-sm text-red-600">{{ createError }}</p>
      <div class="flex gap-2">
        <button
          :disabled="creating"
          class="rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
          @click="createVersion"
        >
          创建
        </button>
        <button class="rounded border px-3 py-1 text-sm" @click="showCreate = false">取消</button>
      </div>
    </div>

    <p v-if="error" class="text-red-500">{{ error }}</p>
    <p v-if="loading">加载中…</p>

    <div v-if="!loading && versions.length === 0" class="text-gray-500 text-sm">
      暂无版本日，点击"新建版本"开始规划。
    </div>

    <ul class="space-y-2">
      <li
        v-for="v in versions"
        :key="v.id"
        class="cursor-pointer rounded border p-3 hover:bg-gray-50"
        @click="openDetail(v)"
      >
        <div class="flex items-center justify-between">
          <div>
            <span class="font-medium">{{ v.name }}</span>
            <span class="ml-2 text-sm text-gray-500">{{ v.semver }}</span>
            <span
              class="ml-2 inline-block rounded px-2 py-0.5 text-xs"
              :class="statusClass[v.status]"
            >
              {{ statusLabels[v.status] }}
            </span>
          </div>
          <span v-if="v.progress" class="text-sm text-gray-500">
            进度 {{ Math.round((v.progress.completion_rate ?? 0) * 100) }}%
          </span>
        </div>
        <div v-if="v.target_date" class="mt-1 text-xs text-gray-400">
          目标: {{ v.target_date }}
        </div>
      </li>
    </ul>
  </div>
</template>
