<script setup lang="ts">
import { onMounted, ref } from "vue";

import { versionApi, type BugVersionView } from "@/api/services/version";

const props = defineProps<{
  workspaceSlug: string;
  projectId: number;
  versionId: number;
}>();

const defects = ref<BugVersionView[]>([]);
const loading = ref(true);
const error = ref("");

const severityLabels: Record<number, string> = {
  0: "致命",
  1: "严重",
  2: "一般",
  3: "轻微",
  4: "建议",
};

let wsIdVal = 0;
async function resolveWsId(): Promise<number> {
  if (wsIdVal) return wsIdVal;
  const { workspaceApi } = await import("@/api/services/workspace");
  const ws = await workspaceApi.getBySlug(props.workspaceSlug);
  wsIdVal = ws.id;
  return wsIdVal;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    const res = await versionApi.getDefectPanel(wsId, props.projectId, props.versionId);
    defects.value = res.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载缺陷面板失败";
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-2">
    <p v-if="error" class="text-red-500 text-sm">{{ error }}</p>
    <p v-if="loading">加载中…</p>
    <div v-if="!loading && defects.length === 0" class="text-gray-400 text-sm">
      该版本暂无关联缺陷
    </div>
    <table v-if="defects.length" class="w-full text-sm">
      <thead class="text-left text-xs text-gray-500">
        <tr>
          <th class="py-1">标识</th>
          <th>名称</th>
          <th>严重度</th>
          <th>状态</th>
          <th>发现版本</th>
          <th>修复版本</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in defects" :key="d.issue_id" class="border-t">
          <td class="py-1 font-mono text-xs">{{ d.identifier }}</td>
          <td>{{ d.name }}</td>
          <td>{{ d.severity != null ? severityLabels[d.severity] ?? d.severity : "-" }}</td>
          <td>{{ d.state_name }}</td>
          <td>{{ d.found_version ?? "-" }}</td>
          <td>{{ d.fix_version ?? "-" }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
