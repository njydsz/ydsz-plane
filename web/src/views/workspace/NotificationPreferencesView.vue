<script setup lang="ts">
/**
 * 通知偏好 — 渠道、汇总频率、免打扰时段配置。
 * 数据来源：notificationApi.getPreference / updatePreference。
 */
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import {
  notificationApi,
  type NotificationPreference,
} from "@/api/services/notification";
import { AppErrorState, AppSkeleton } from "@/components";
import { toast } from "@/lib/toast";

const route = useRoute();
const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));

const loading = ref(true);
const error = ref("");
const pref = ref<NotificationPreference | null>(null);
const saving = ref(false);

const CHANNELS = [
  { key: "in_app", label: "站内通知" },
  { key: "email", label: "邮件" },
  { key: "im", label: "IM（飞书/钉钉等）" },
];

const DIGESTS = [
  { key: "realtime", label: "实时" },
  { key: "daily", label: "每日汇总" },
  { key: "weekly", label: "每周汇总" },
  { key: "off", label: "关闭" },
] as const;

async function load() {
  if (!workspaceId.value) { loading.value = false; return; }
  loading.value = true;
  error.value = "";
  try {
    pref.value = await notificationApi.getPreference(workspaceId.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : "加载失败";
  } finally {
    loading.value = false;
  }
}

function toggleChannel(key: string) {
  if (!pref.value) return;
  const i = pref.value.channels.indexOf(key);
  if (i >= 0) pref.value.channels.splice(i, 1);
  else pref.value.channels.push(key);
}

async function save() {
  if (!pref.value) return;
  saving.value = true;
  try {
    await notificationApi.updatePreference(workspaceId.value, {
      channels: pref.value.channels,
      digest: pref.value.digest,
      dnd_enabled: pref.value.dnd_enabled,
      dnd_start: pref.value.dnd_start,
      dnd_end: pref.value.dnd_end,
    });
    toast.success("已保存");
  } catch (err: unknown) {
    toast.error(err instanceof Error ? err.message : "保存失败");
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold tracking-tight">通知偏好</h1>
      <button
        :disabled="saving || !pref"
        class="rounded-md bg-[var(--brand-600)] px-3 py-1.5 text-sm font-medium text-white hover:bg-[var(--brand-700)] disabled:opacity-50"
        @click="save"
      >
        {{ saving ? "保存中…" : "保存" }}
      </button>
    </div>

    <div v-if="loading" class="space-y-3">
      <AppSkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
    </div>

    <AppErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else-if="pref">
      <!-- 渠道 -->
      <section>
        <h2 class="mb-2 text-sm font-semibold text-[var(--text-secondary)]">通知渠道</h2>
        <div class="space-y-2">
          <label
            v-for="c in CHANNELS"
            :key="c.key"
            class="flex items-center gap-2 text-sm text-[var(--text-primary)]"
          >
            <input
              type="checkbox"
              :checked="pref.channels.includes(c.key)"
              @change="toggleChannel(c.key)"
            />
            {{ c.label }}
          </label>
        </div>
      </section>

      <!-- 汇总频率 -->
      <section>
        <h2 class="mb-2 text-sm font-semibold text-[var(--text-secondary)]">汇总频率</h2>
        <div class="flex gap-2">
          <button
            v-for="d in DIGESTS"
            :key="d.key"
            class="rounded-md px-3 py-1.5 text-sm"
            :class="pref.digest === d.key
              ? 'bg-[var(--brand-600)] text-white'
              : 'border border-[var(--border-subtle)] text-[var(--text-secondary)]'"
            @click="pref.digest = d.key"
          >
            {{ d.label }}
          </button>
        </div>
      </section>

      <!-- 免打扰 -->
      <section>
        <h2 class="mb-2 text-sm font-semibold text-[var(--text-secondary)]">免打扰时段</h2>
        <label class="flex items-center gap-2 text-sm text-[var(--text-primary)]">
          <input type="checkbox" v-model="pref.dnd_enabled" />
          启用免打扰
        </label>
        <div v-if="pref.dnd_enabled" class="mt-2 flex items-center gap-2">
          <input
            v-model="pref.dnd_start"
            type="time"
            class="rounded-md border border-[var(--border-subtle)] px-2 py-1.5 text-sm"
          />
          <span class="text-sm text-[var(--text-tertiary)]">至</span>
          <input
            v-model="pref.dnd_end"
            type="time"
            class="rounded-md border border-[var(--border-subtle)] px-2 py-1.5 text-sm"
          />
        </div>
      </section>
    </template>
  </div>
</template>
