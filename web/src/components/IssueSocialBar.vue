<script setup lang="ts">
/**
 * IssueSocialBar — 工作项关注栏（订阅动态通知）。
 */
import { computed, ref } from "vue";

import { issueApi } from "@/api/services/issue";
import { useAuthStore } from "@/stores/auth";
import { promiseToast } from "@/lib/toast";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  issueId: number;
  /** 当前用户是否已关注（由父级 Issue.watchers 推导） */
  initialWatching?: boolean;
}>();

const auth = useAuthStore();

const watching = ref(!!props.initialWatching);
const busy = ref(false);

const currentUserId = computed(() => auth.user?.id ?? 0);
const canInteract = computed(() => currentUserId.value > 0 && !busy.value);

async function toggleWatch() {
  if (!canInteract.value) return;
  busy.value = true;
  try {
    if (watching.value) {
      await promiseToast(
        issueApi.unwatch(props.workspaceId, props.projectId, props.issueId),
        { loading: "取消关注...", success: () => "已取消关注", error: () => "操作失败" },
      );
      watching.value = false;
    } else {
      await promiseToast(
        issueApi.watch(props.workspaceId, props.projectId, props.issueId),
        { loading: "关注中...", success: () => "已关注，变更将通知你", error: () => "操作失败" },
      );
      watching.value = true;
    }
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="social-bar" aria-label="关注">
    <!-- 关注 -->
    <button
      class="watch-btn"
      :class="{ 'watch-btn--active': watching }"
      :disabled="!canInteract"
      @click="toggleWatch"
    >
      <span class="watch-icon" aria-hidden="true">👁</span>
      <span>{{ watching ? "已关注" : "关注" }}</span>
    </button>
  </div>
</template>

<style scoped>
.social-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 2px 0;
}

/* ---- 关注 ---- */
.watch-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--bg-surface-1, #fff);
  color: var(--txt-secondary, #6b7280);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}
.watch-btn:hover:not(:disabled) {
  border-color: var(--brand-300, #93c5fd);
  color: var(--brand-600, #2563eb);
}
.watch-btn--active {
  border-color: var(--brand-400, #60a5fa);
  background: var(--brand-50, #eff6ff);
  color: var(--brand-600, #2563eb);
}
.watch-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
.watch-icon {
  font-size: 12px;
}
</style>
