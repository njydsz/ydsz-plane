<script setup lang="ts">
/**
 * DynamicIcon — 根据图标名称渲染对应 lucide 图标组件。
 *
 * 使用静态导入映射确保 build 时 tree-shaking 只保留实际使用的图标。
 * 图标名采用 lucide 命名规范（PascalCase），与 permission.ts 中 MenuItem.icon 对齐。
 */
import {
  LayoutDashboard,
  UserCircle,
  FolderKanban,
  BookOpen,
  Clock,
  Tag,
  BarChart3,
  Zap,
  ScrollText,
  Inbox,
  Link,
  Users,
  Settings,
} from "@lucide/vue";
import { computed } from "vue";

interface Props {
  name: string;
  size?: number;
  strokeWidth?: number;
}

const props = withDefaults(defineProps<Props>(), {
  size: 18,
  strokeWidth: 2,
});

const iconMap: Record<string, (typeof LayoutDashboard)> = {
  LayoutDashboard,
  UserCircle,
  FolderKanban,
  BookOpen,
  Clock,
  Tag,
  BarChart3,
  Zap,
  ScrollText,
  Inbox,
  Link,
  Users,
  Settings,
};

const IconComponent = computed(() => iconMap[props.name] ?? Settings);
</script>

<template>
  <component
    :is="IconComponent"
    :size="props.size"
    :stroke-width="props.strokeWidth"
  />
</template>
