/**
 * v-permission 自定义指令 — 基于后端权限集合控制元素显示/隐藏。
 *
 * 用法：
 *   <button v-permission="'issue:edit_all'">编辑</button>
 *   <button v-permission:any="['issue:edit_all', 'issue:edit_own']">编辑</button>
 *
 * 如果当前用户没有相应权限，元素会从 DOM 中移除（与 v-if 类似）。
 */
import type { Directive } from "vue";

import { useWorkspaceStore } from "@/stores/workspace";

export const vPermission: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    applyPermission(el, binding);
  },
  updated(el, binding) {
    applyPermission(el, binding);
  },
};

function applyPermission(el: HTMLElement, binding: { value: string | string[]; arg?: string }) {
  const wsStore = useWorkspaceStore();
  const required = Array.isArray(binding.value) ? binding.value : [binding.value];

  let ok: boolean;
  if (binding.arg === "any") {
    ok = required.some((p) => wsStore.hasPermission(p));
  } else {
    ok = required.every((p) => wsStore.hasPermission(p));
  }

  if (!ok && el.parentElement) {
    el.parentElement.removeChild(el);
  }
}
