<script setup lang="ts">
/**
 * ActivityTimelineWidget — 最近活动时间线。
 * 展示 actor + verb + issue identifier + 相对时间。
 */
import type { RecentActivityData } from "@/api/services/dashboard";
import { formatRelativeTime } from "@/lib/formatTime";

defineProps<{
  data?: RecentActivityData;
}>();

function verbLabel(verb: string): string {
  const map: Record<string, string> = {
    created: "创建了",
    updated: "更新了",
    transitioned: "流转了",
    commented: "评论了",
    assigned: "指派了",
    unassigned: "取消指派了",
  };
  return map[verb] ?? verb;
}
</script>

<template>
  <div class="activity-timeline">
    <ul v-if="data?.items?.length" class="activity-timeline__list">
      <li v-for="item in data.items" :key="item.id" class="activity-item">
        <div class="activity-item__avatar">
          {{ item.actor_name.charAt(0).toUpperCase() }}
        </div>
        <div class="activity-item__body">
          <span class="activity-item__actor">{{ item.actor_name }}</span>
          <span class="activity-item__verb">{{ verbLabel(item.verb) }}</span>
          <span class="activity-item__issue">{{ item.issue_identifier }}</span>
          <span class="activity-item__time">{{ formatRelativeTime(item.created_at) }}</span>
        </div>
      </li>
    </ul>
    <div v-else class="activity-timeline__empty">暂无动态</div>
  </div>
</template>

<style scoped>
.activity-timeline {
  display: flex;
  flex-direction: column;
}

.activity-timeline__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-subtle, #e5e7eb);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-item__avatar {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: var(--brand-50, #eef2fe);
  color: var(--brand-600, #2f4fd0);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
}

.activity-item__body {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  line-height: 1.5;
}

.activity-item__actor {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

.activity-item__verb {
  color: var(--text-secondary, #4b5563);
}

.activity-item__issue {
  color: var(--brand-500, #3f63f1);
  font-weight: 500;
  font-family: var(--font-mono, monospace);
  cursor: pointer;
}

.activity-item__issue:hover {
  text-decoration: underline;
}

.activity-item__time {
  color: var(--text-tertiary, #9ca3af);
  font-size: 11px;
  margin-left: auto;
}

.activity-timeline__empty {
  padding: 24px 0;
  text-align: center;
  color: var(--text-tertiary, #9ca3af);
  font-size: 13px;
}
</style>
