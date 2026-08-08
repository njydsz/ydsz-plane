<script setup lang="ts">
/**
 * KnowledgePageTreeNode — 知识库文档树节点（递归组件）。
 *
 * 通过 children 递归渲染无限层级文档树；支持展开/折叠、选中、
 * 新建子文档、重命名（InlineEdit）、删除。
 */
import { computed } from "vue";

import type { KnowledgePageNode } from "@/api/services/knowledge";
import InlineEdit from "@/components/InlineEdit.vue";

/** 显式命名以支持模板内递归自引用（KnowledgePageTreeNode） */
defineOptions({ name: "KnowledgePageTreeNode" });

const props = defineProps<{
  node: KnowledgePageNode;
  /** 折叠的节点 ID 集合 */
  collapsed: Set<number>;
  /** 当前选中的文档 ID */
  selectedId: number | null;
}>();

const emit = defineEmits<{
  (e: "select", id: number): void;
  (e: "toggle", id: number): void;
  (e: "create-child", id: number): void;
  (e: "rename", node: KnowledgePageNode, title: string): void;
  (e: "delete", node: KnowledgePageNode): void;
}>();

const isExpanded = computed(() => !props.collapsed.has(props.node.id));
const hasChildren = computed(() => props.node.children.length > 0);
</script>

<template>
  <li class="tree-node">
    <div
      class="tree-row"
      :class="{ 'tree-row--active': selectedId === node.id }"
      @click="emit('select', node.id)"
    >
      <span
        v-if="hasChildren"
        class="tree-caret"
        :class="{ 'tree-caret--open': isExpanded }"
        @click.stop="emit('toggle', node.id)"
      >▶</span>
      <span v-else class="tree-caret tree-caret--placeholder" />
      <span class="tree-file">📄</span>
      <InlineEdit
        class="tree-name"
        :model-value="node.title"
        placeholder="未命名文档"
        :max-length="512"
        @submit="(v) => emit('rename', node, v)"
      />
      <span class="tree-actions">
        <button
          class="tree-action"
          title="新建子文档"
          @click.stop="emit('create-child', node.id)"
        >＋</button>
        <button
          class="tree-action tree-action--danger"
          title="删除文档"
          @click.stop="emit('delete', node)"
        >×</button>
      </span>
    </div>

    <ul v-if="hasChildren && isExpanded" class="tree-children">
      <KnowledgePageTreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :collapsed="collapsed"
        :selected-id="selectedId"
        @select="(id) => emit('select', id)"
        @toggle="(id) => emit('toggle', id)"
        @create-child="(id) => emit('create-child', id)"
        @rename="(n, t) => emit('rename', n, t)"
        @delete="(n) => emit('delete', n)"
      />
    </ul>
  </li>
</template>

<style scoped>
.tree-node {
  margin: 1px 0;
}

.tree-children {
  list-style: none;
  margin: 0;
  padding: 0 0 0 16px;
}

.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 6px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.1s;
}

.tree-row:hover {
  background: var(--surface-3);
}

.tree-row--active {
  background: var(--brand-50);
}

.tree-caret {
  font-size: 9px;
  color: var(--text-tertiary);
  width: 14px;
  text-align: center;
  flex-shrink: 0;
  transition: transform 0.12s;
}

.tree-caret--open {
  transform: rotate(90deg);
}

.tree-caret--placeholder {
  visibility: hidden;
}

.tree-file {
  font-size: 12px;
  flex-shrink: 0;
}

.tree-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-primary);
}

.tree-actions {
  display: none;
  gap: 2px;
  flex-shrink: 0;
}

.tree-row:hover .tree-actions {
  display: inline-flex;
}

.tree-action {
  width: 18px;
  height: 18px;
  border: none;
  border-radius: 3px;
  background: none;
  color: var(--text-tertiary);
  font-size: 12px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
}

.tree-action:hover {
  background: var(--surface-3);
  color: var(--text-primary);
}

.tree-action--danger:hover {
  background: var(--danger-50);
  color: var(--danger-500);
}
</style>
