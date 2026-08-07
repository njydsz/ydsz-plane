/**
 * useWorkspaceContext — 工作空间与项目上下文解析 composable。
 *
 * 消除 4+ 个 Sprint 视图中重复的 resolveWsId / resolveProjectId 逻辑。
 * 设计参考：VueUse 的 createSharedComposable 模式（但保持独立以避免额外依赖）。
 */

import { computed, ref, watchEffect, type Ref } from "vue";
import { useRoute } from "vue-router";
import { useWorkspaceStore } from "@/stores/workspace";

/** 工作空间 + 项目上下文 */
export interface WorkspaceContext {
  /** 工作空间 ID（数值） */
  wsId: Ref<number>;
  /** 项目 ID（数值） */
  projectId: Ref<number>;
  /** 工作空间 slug */
  workspaceSlug: Ref<string>;
  /** 是否正在解析中 */
  loading: Ref<boolean>;
  /** 解析错误信息 */
  error: Ref<string | null>;
  /** 上下文是否已就绪（两个 ID 均为正数） */
  ready: Ref<boolean>;
}

/**
 * 从路由参数解析工作空间与项目上下文。
 *
 * 用法：
 * ```ts
 * const { wsId, projectId, workspaceSlug, ready } = useWorkspaceContext()
 * watchEffect(() => { if (ready.value) loadData(wsId.value, projectId.value) })
 * ```
 */
export function useWorkspaceContext(): WorkspaceContext {
  const route = useRoute();
  const wsStore = useWorkspaceStore();

  const workspaceSlug = ref<string>("");
  const wsId = ref(0);
  const projectId = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const ready = computed(() => wsId.value > 0 && projectId.value > 0);

  watchEffect(async () => {
    const slug = (route.params.workspaceSlug as string) ?? "";
    const pidRaw = route.params.projectId as string | undefined;

    workspaceSlug.value = slug;

    if (!slug || !pidRaw) {
      wsId.value = 0;
      projectId.value = 0;
      return;
    }

    const pid = Number(pidRaw);
    projectId.value = Number.isNaN(pid) ? 0 : pid;

    if (!Number.isNaN(pid) && pid > 0) {
      // 解析 workspace slug -> id
      if (slug) {
        loading.value = true;
        try {
          const ws = await wsStore.resolveBySlug(slug);
          wsId.value = ws?.id ?? 0;
          error.value = ws ? null : `工作空间 "${slug}" 未找到`;
        } catch (e: unknown) {
          error.value = e instanceof Error ? e.message : "工作空间解析失败";
          wsId.value = 0;
        } finally {
          loading.value = false;
        }
      }
    }
  });

  return { wsId, projectId, workspaceSlug, loading, error, ready };
}
