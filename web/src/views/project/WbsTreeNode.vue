<script setup lang="ts">
/**
 * WbsTreeNode — WBS 工作项树节点（递归组件）。
 *
 * 通过 children 递归渲染三层 WBS 层级：
 *   - 需求：Epic → Feature → Story
 *   - 任务：主任务 → 子任务 → 子子任务
 *   - 缺陷：主缺陷 → 子缺陷 → 子子缺陷
 *
 * 支持：展开/折叠、选中、行内重命名、新建子工作项、删除。
 */
import { computed } from "vue";

import { type Issue, type IssuePriority } from "@/api/services/issue";
import InlineEdit from "@/components/InlineEdit.vue";

/** 显式命名以支持模板内递归自引用 */
defineOptions({ name: "WbsTreeNode" });

const props = defineProps<{
  node: WbsNode;
  /** 折叠的节点 ID 集合 */
  collapsed: Set<number>;
  /** 当前选中的工作项 ID */
  selectedId: number | null;
}>();

const emit = defineEmits<{
  (e: "select", id: number): void;
  (e: "toggle", id: number): void;
  (e: "create-child", parentId: number): void;
  (e: "rename", issue: Issue, name: string): void;
  (e: "delete", issue: Issue): void;
}>();

const isExpanded = computed(() => !props.collapsed.has(props.node.issue.id));
const hasChildren = computed(() => props.node.children.length > 0);

/** 工作项类型图标 */
const typeIcon = computed(() => {
  switch (props.node.issue.type_code) {
    case "epic": return "🎯";
    case "requirement": return "📋";
    case "task": return "✅";
    case "defect": return "🐛";
    default: return "📄";
  }
});

/** 类型中文标签 */
const typeLabel = computed(() => {
  switch (props.node.issue.type_code) {
    case "epic": return "史诗";
    case "requirement": return "需求";
    case "task": return "任务";
    case "defect": return "缺陷";
    default: return "工作项";
  }
});

/** 优先级色块 */
const priorityColor = computed(() => {
  const map: Record<IssuePriority, string> = {
    urgent: "#ef4444",
    high: "#f97316",
    medium: "#eab308",
    low: "#22c55e",
    none: "#94a3b8",
  };
  return map[props.node.issue.priority] || "#94a3b8";
});

/** 优先级标签 */
const priorityLabel = computed(() => {
  const map: Record<IssuePriority, string> = {
    urgent: "紧急",
    high: "高",
    medium: "中",
    low: "低",
    none: "无",
  };
  return map[props.node.issue.priority] || "-";
});

/** 进度百分比文本 */
const progressText = computed(() => {
  const p = props.node.issue.progress ?? 0;
  return p > 0 ? `${p}%` : "";
});

/** 深度缩进中线色（不同层级用不同颜色标示） */
const depthBorderStyle = computed(() => {
  const colors = ["", "var(--brand-300)", "var(--brand-200)", "var(--brand-100)"];
  return {
    borderLeftColor: colors[props.node.issue.depth] || "transparent",
  };
});

/** 子工作项数量徽章 */
const childCount = computed(() => props.node.children.length);
</script>

<script lang="ts">
/** 树形节点：Issue + 子节点数组 */
export interface WbsNode {
  issue: Issue;
  children: WbsNode[];
}
</script>

<template>
  <li class="wbs-node">
    <!-- 节点行 -->
    <div
      class="wbs-row"
      :class="{
        'wbs-row--active': selectedId === node.issue.id,
        [`wbs-row--depth-${node.issue.depth}`]: true,
      }"
      :style="depthBorderStyle"
      @click="emit('select', node.issue.id)"
    >
      <!-- 展开/折叠按钮 -->
      <span
        v-if="hasChildren"
        class="wbs-caret"
        :class="{ 'wbs-caret--open': isExpanded }"
        :title="isExpanded ? '折叠' : '展开'"
        @click.stop="emit('toggle', node.issue.id)"
      >▶</span>
      <span v-else class="wbs-caret wbs-caret--placeholder" />

      <!-- 类型图标 -->
      <span class="wbs-type-icon" :title="typeLabel">{{ typeIcon }}</span>

      <!-- 工作项标识符 -->
      <span class="wbs-identifier">{{ node.issue.identifier }}</span>

      <!-- 名称（可内联编辑） -->
      <InlineEdit
        class="wbs-name"
        :model-value="node.issue.name"
        placeholder="未命名工作项"
        :max-length="200"
        @submit="(v) => emit('rename', node.issue, v)"
      />

      <!-- 优先级色块标记 -->
      <span
        class="wbs-priority-dot"
        :style="{ backgroundColor: priorityColor }"
        :title="`优先级: ${priorityLabel}`"
      />

      <!-- 进度 -->
      <span v-if="progressText" class="wbs-progress">{{ progressText }}</span>

      <!-- 子项数量徽章 -->
      <span v-if="childCount > 0" class="wbs-badge" :title="`${childCount} 个子工作项`">
        {{ childCount }}
      </span>

      <!-- 操作按钮（hover 显示） -->
      <span class="wbs-actions">
        <button
          v-if="node.issue.depth < 3"
          class="wbs-action"
          title="添加子工作项"
          @click.stop="emit('create-child', node.issue.id)"
        >＋</button>
        <button
          class="wbs-action wbs-action--danger"
          title="删除工作项"
          @click.stop="emit('delete', node.issue)"
        >×</button>
      </span>
    </div>

    <!-- 递归渲染子节点 -->
    <ul v-if="hasChildren && isExpanded" class="wbs-children">
      <WbsTreeNode
        v-for="child in node.children"
        :key="child.issue.id"
        :node="child"
        :collapsed="collapsed"
        :selected-id="selectedId"
        @select="(id) => emit('select', id)"
        @toggle="(id) => emit('toggle', id)"
        @create-child="(id) => emit('create-child', id)"
        @rename="(iss, name) => emit('rename', iss, name)"
        @delete="(iss) => emit('delete', iss)"
      />
    </ul>
  </li>
</template>

<style scoped>
.wbs-node {
  margin: 1px 0;
}

.wbs-children {
  list-style: none;
  margin: 0;
  padding: 0 0 0 20px;
}

.wbs-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.1s;
  border-left: 2px solid transparent;
}

.wbs-row:hover {
  background: var(--surface-3);
}

.wbs-row--active {
  background: var(--brand-50);
  border-left-color: var(--brand-500) !important;
}

.wbs-row--depth-0 {
  font-weight: 500;
}
.wbs-row--depth-1 {
  padding-left: 14px;
}
.wbs-row--depth-2 {
  padding-left: 28px;
}
.wbs-row--depth-3 {
  padding-left: 42px;
}

.wbs-caret {
  font-size: 8px;
  color: var(--text-tertiary);
  width: 14px;
  text-align: center;
  flex-shrink: 0;
  transition: transform 0.12s;
  user-select: none;
}

.wbs-caret--open {
  transform: rotate(90deg);
}

.wbs-caret--placeholder {
  visibility: hidden;
}

.wbs-type-icon {
  font-size: 13px;
  flex-shrink: 0;
  width: 18px;
  text-align: center;
}

.wbs-identifier {
  font-size: 11px;
  color: var(--text-tertiary);
  font-family: var(--font-mono, monospace);
  flex-shrink: 0;
  min-width: 50px;
}

.wbs-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-primary);
}

.wbs-priority-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.wbs-progress {
  font-size: 11px;
  color: var(--text-secondary);
  flex-shrink: 0;
  min-width: 32px;
  text-align: right;
}

.wbs-badge {
  font-size: 10px;
  background: var(--surface-4);
  color: var(--text-tertiary);
  border-radius: 8px;
  padding: 0 5px;
  line-height: 16px;
  flex-shrink: 0;
  min-width: 18px;
  text-align: center;
}

.wbs-actions {
  display: none;
  gap: 2px;
  flex-shrink: 0;
}

.wbs-row:hover .wbs-actions {
  display: inline-flex;
}

.wbs-action {
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 3px;
  background: none;
  color: var(--text-tertiary);
  font-size: 13px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.wbs-action:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.wbs-action--danger:hover {
  background: var(--danger-50);
  color: var(--danger-500);
}
</style>
