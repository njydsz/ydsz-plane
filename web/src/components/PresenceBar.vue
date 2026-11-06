<script setup lang="ts">
/**
 * PresenceBar — 文档在线编辑者指示器。
 *
 * 显示当前文档的在线协作者头像组、版本号与连接状态。
 * 与 useCollaboration 配合使用。
 */
import { computed } from "vue";

import type { CollaborationSession } from "@/api/services/collaboration";

const props = withDefaults(defineProps<{
  sessions: CollaborationSession[];
  currentUserId: number;
  version: number;
  joined: boolean;
  /** 是否显示版本号 */
  showVersion?: boolean;
  /** 最大展示头像数 */
  maxAvatars?: number;
}>(), {
  showVersion: true,
  maxAvatars: 5,
});

const visible = computed(() => props.sessions.slice(0, props.maxAvatars));
const overflow = computed(() => Math.max(0, props.sessions.length - props.maxAvatars));

function initials(name: string): string {
  if (!name) return "?";
  return name.slice(0, 2).toUpperCase();
}

/** 稳定的头像底色（基于 user_id） */
const palette = ["var(--brand-500)", "var(--success-500)", "var(--warning-500)", "var(--brand-500)", "var(--brand-500)", "var(--info-500)", "var(--warning-500)"];
function colorOf(userId: number): string {
  return palette[Math.abs(userId) % palette.length];
}
</script>

<template>
  <div class="presence">
    <!-- 连接状态 -->
    <span
      class="presence__dot"
      :class="joined ? 'presence__dot--on' : 'presence__dot--off'"
      :title="joined ? '协作已连接' : '协作未连接'"
    />

    <!-- 头像组 -->
    <div v-if="visible.length" class="presence__avatars">
      <span
        v-for="s in visible"
        :key="s.user_id"
        class="presence__avatar"
        :style="{ background: colorOf(s.user_id) }"
        :title="`${s.display_name || s.user_name || ('用户 #' + s.user_id)}${s.user_id === currentUserId ? '（我）' : ''}`"
      >
        {{ initials(s.display_name || s.user_name || String(s.user_id)) }}
      </span>
      <span v-if="overflow" class="presence__avatar presence__avatar--more">
        +{{ overflow }}
      </span>
    </div>

    <span v-else class="presence__empty">仅你一人</span>

    <!-- 版本 -->
    <span v-if="showVersion" class="presence__version">v{{ version }}</span>
  </div>
</template>

<style scoped>
.presence {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--txt-tertiary);
}

.presence__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.presence__dot--on {
  background: var(--success);
  box-shadow: 0 0 0 3px var(--success-bg, rgba(16, 185, 129, 0.15));
}

.presence__dot--off {
  background: var(--txt-tertiary);
}

.presence__avatars {
  display: flex;
  align-items: center;
}

.presence__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: var(--text-on-color);
  font-size: 10px;
  font-weight: 600;
  border: 2px solid var(--bg-surface);
  margin-left: -6px;
}

.presence__avatar:first-child {
  margin-left: 0;
}

.presence__avatar--more {
  background: var(--bg-surface-2);
  color: var(--txt-secondary);
}

.presence__empty {
  font-size: 12px;
}

.presence__version {
  font-variant-numeric: tabular-nums;
  color: var(--txt-tertiary);
}
</style>
