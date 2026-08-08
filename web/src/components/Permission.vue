<script setup lang="ts">
/**
 * <Permission> 包裹组件 — 基于后端返回的权限码集合决定是否渲染插槽内容。
 *
 * 用法：
 *   <Permission permission="issue:edit_all">
 *     <button>编辑</button>
 *   </Permission>
 *
 *   <Permission :permissions="['issue:edit_all', 'issue:edit_own']" match="any">
 *     <button>编辑</button>
 *   </Permission>
 *
 *   <Permission :menu="{ permissions: ['sprint:read'], minLevel: 30 }">
 *     <SidebarLink to="/sprints" />
 *   </Permission>
 */
import { computed } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";

interface Props {
  /** 单一权限码 */
  permission?: string;
  /** 多个权限码 */
  permissions?: string[];
  /** 匹配策略：all = 全部满足，any = 满足其一 */
  match?: "all" | "any";
  /** 菜单项定义（含 permissions / minLevel） */
  menu?: { permissions?: string[]; minLevel?: number };
}

const props = withDefaults(defineProps<Props>(), {
  permission: undefined,
  permissions: () => [],
  match: "all",
  menu: undefined,
});

const wsStore = useWorkspaceStore();

const visible = computed(() => {
  // 菜单模式：直接用 canSeeMenu 判断
  if (props.menu) {
    return wsStore.canSeeMenu(props.menu);
  }

  const required = props.permission ? [props.permission] : props.permissions ?? [];
  if (required.length === 0) return true;

  if (props.match === "any") {
    return required.some((p) => wsStore.hasPermission(p));
  }
  return required.every((p) => wsStore.hasPermission(p));
});
</script>

<template>
  <slot v-if="visible" />
</template>
