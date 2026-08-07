<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { versionApi, type Version } from "@/api/services/version";

const route = useRoute();
const router = useRouter();

const projectId = computed(() => Number(route.params.projectId));
const workspaceSlug = computed(() => String(route.params.workspaceSlug ?? ""));
const versionId = computed(() => Number(route.params.versionId));

const version = ref<Version | null>(null);
const loading = ref(true);
const releasing = ref(false);
const error = ref("");
const draftOverride = ref("");
const forceChecklist = ref(false);
const addKnown = ref(true);

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
  try {
    const wsId = await resolveWsId();
    version.value = await versionApi.getVersion(wsId, projectId.value, versionId.value);
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

async function release() {
  if (!version.value) return;
  releasing.value = true;
  error.value = "";
  try {
    const wsId = await resolveWsId();
    await versionApi.releaseVersion(wsId, projectId.value, versionId.value, {
      draft_override: draftOverride.value || undefined,
      force_checklist: forceChecklist.value,
      add_known_issues_to_notes: addKnown.value,
    });
    router.push({ name: "version-detail", params: { versionId: versionId.value } });
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "发布失败";
  } finally {
    releasing.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div v-if="error" class="text-red-500 mb-3">{{ error }}</div>
  <div v-if="loading">加载中…</div>
  <div v-else-if="version" class="space-y-5">
    <div>
      <h1 class="text-xl font-semibold">发布向导</h1>
      <p class="text-sm text-gray-500">{{ version.name }} v{{ version.semver }}</p>
    </div>

    <!-- Gate 1: Checklist -->
    <section class="rounded border p-4 space-y-2">
      <h2 class="font-medium">① 检查清单</h2>
      <ul class="space-y-1 text-sm">
        <li v-for="item in version.checklist" :key="item.id" class="flex items-center gap-2">
          <span v-if="item.checked" class="text-green-600">✓</span>
          <span v-else class="text-red-500">✗</span>
          <span>{{ item.label }}</span>
          <span v-if="item.required" class="text-xs text-red-500">（必填）</span>
        </li>
      </ul>
    </section>

    <!-- Gate 2: Quality -->
    <section class="rounded border p-4 space-y-2">
      <h2 class="font-medium">② 准出校验</h2>
      <div class="text-sm space-y-1">
        <div>
          致命/严重未关闭缺陷:
          <span :class="version.quality?.critical_bugs === 0 ? 'text-green-600' : 'text-red-600'">
            {{ version.quality?.critical_bugs ?? 0 }}
          </span>
        </div>
        <div>修复率: {{ Math.round((version.quality?.fix_rate ?? 0) * 100) }}%</div>
        <div>
          结果:
          <span :class="version.quality?.pass_quality_gate ? 'text-green-600' : 'text-red-600'">
            {{ version.quality?.pass_quality_gate ? "通过" : "不通过" }}
          </span>
        </div>
      </div>
    </section>

    <!-- Gate 3: Notes Preview -->
    <section class="rounded border p-4 space-y-2">
      <h2 class="font-medium">③ Release Notes 预览</h2>
      <label class="flex items-center gap-2 text-sm">
        <input type="checkbox" v-model="addKnown" />
        在 Notes 中包含已知未关闭问题
      </label>
      <textarea
        v-model="draftOverride"
        placeholder="留空则使用自动生成的 Release Notes；填写则覆盖为自定义内容"
        class="w-full rounded border p-2 text-sm font-mono"
        rows="8"
      ></textarea>
    </section>

    <label class="flex items-center gap-2 text-sm">
      <input type="checkbox" v-model="forceChecklist" />
      强制跳过清单校验（Admin 豁免）
    </label>

    <div class="flex gap-2">
      <button
        :disabled="releasing"
        class="rounded bg-green-600 px-4 py-2 text-white disabled:opacity-50"
        @click="release"
      >
        确认发布
      </button>
      <button class="rounded border px-4 py-2" @click="router.back()">取消</button>
    </div>
  </div>
</template>
