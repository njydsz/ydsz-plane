/**
 * useWorkspaceContext — 从路由解析 workspace/project 上下文的通用 composable。
 *
 * 解决 4 个 Sprint 视图（及其他项目域视图）重复实现 resolveWsId() 的问题：
 *  1) 从路由参数直接读取 workspaceId（无需 slug→ID 解析）
 *  2) 从路由参数解析 projectId
 *  3) 暴露 ready 信号，供 watchEffect 驱动按需加载
 */

import { computed, ref, watchEffect } from "vue";
import { useRoute } from "vue-router";
import { workspaceApi } from "@/api/services/workspace";

/** Workspace/project 上下文结构体（含 wsId / projectId / ready / resolve）。 */
export interface WorkspaceContext {
  /** 工作空间 ID ref（解析成功后 > 0） */
  wsId: Readonly<import("vue").Ref<number>>;
  /** 项目 ID ref（从路由解析） */
  projectId: Readonly<import("vue").Ref<number>>;
  /** 上下文是否已就绪（workspaceId 有效且 projectId 有效） */
  ready: Readonly<import("vue").Ref<boolean>>;
  /** 解析失败的错误信息 */
  error: Readonly<import("vue").Ref<string | null>;
  /** 手动强制重新解析 */
  resolve: () => Promise<void>;
}

let cache = new Map<number, number>();

/** 从路由解析 workspace/project 上下文的 composable。 */
export function useWorkspaceContext(): WorkspaceContext {
  const route = useRoute();

  const workspaceId = computed(() => Number(route.params.workspaceId ?? 0));
  const projectId = computed(() => Number(route.params.projectId));

  const wsId = ref(0);
  const error = ref<string | null>(null);
  const resolving = ref(false);

  const ready = computed(() => wsId.value > 0 && projectId.value > 0);

  async function resolve() {
    if (!workspaceId.value || projectId.value <= 0) {
      error.value = "路由缺少 workspaceId 或 projectId 参数";
      return;
    }
    if (resolving.value) return;
    resolving.value = true;
    error.value = null;
    try {
      const id = workspaceId.value;
      const cached = cache.get(id);
      if (cached) {
        wsId.value = cached;
      } else {
        const ws = await workspaceApi.get(id);
        cache.set(id, ws.id);
        wsId.value = ws.id;
      }
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "工作空间解析失败";
    } finally {
      resolving.value = false;
    }
  }

  // 路由参数变化时自动重解析
  watchEffect(() => {
    if (workspaceId.value && projectId.value > 0) {
      void resolve();
    }
  });

  return { wsId, projectId, ready, error, resolve };
}
