<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import { versionApi, type DeliveryReport, type Version } from "@/api/services/version";

const route = useRoute();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const versionId = computed(() => Number(route.params.versionId));

const version = ref<Version | null>(null);
const report = ref<DeliveryReport | null>(null);
const loading = ref(true);
const error = ref("");

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
    const [v, r] = await Promise.all([
      versionApi.getVersion(wsId, projectId.value, versionId.value),
      versionApi.getDeliveryReport(wsId, projectId.value, versionId.value),
    ]);
    version.value = v;
    report.value = r;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div v-if="error" class="text-red-500">{{ error }}</div>
  <div v-if="loading">加载中…</div>
  <div v-else-if="report && version" class="space-y-4">
    <div>
      <h1 class="text-xl font-semibold">交付报告</h1>
      <p class="text-sm text-gray-500">{{ version.name }} v{{ version.semver }}</p>
    </div>

    <div class="grid grid-cols-2 gap-4 md:grid-cols-3">
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">迭代数</div>
        <div class="text-2xl font-semibold">{{ report.sprint_count }}</div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">总工作项</div>
        <div class="text-2xl font-semibold">{{ report.total_issues }}</div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">已完成项</div>
        <div class="text-2xl font-semibold">{{ report.completed_issues }}</div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">完成点</div>
        <div class="text-2xl font-semibold">{{ Math.round(report.completed_points) }}/{{ Math.round(report.total_points) }}</div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">缺陷 / 已修复</div>
        <div class="text-2xl font-semibold">{{ report.fixed_bug_count }}/{{ report.bug_count }}</div>
      </div>
      <div class="rounded border p-3">
        <div class="text-xs text-gray-500">通过率</div>
        <div class="text-2xl font-semibold">{{ Math.round(report.pass_rate * 100) }}%</div>
      </div>
    </div>

    <div
      class="rounded border p-3 text-sm"
      :class="report.eligible_to_release ? 'bg-green-50' : 'bg-amber-50'"
    >
      {{ report.eligible_to_release ? "✅ 该版本符合发布准出条件" : "⚠️ 该版本尚未满足准出条件（通过率 &lt;80% 或存在严重致命未关闭缺陷）" }}
    </div>
  </div>
</template>
