/**
 * useSprint — 迭代域业务逻辑 composable。
 *
 * 封装 Sprint 域常用的组合操作，作为 Pinia Store 的便利层：
 *  1) 上下文自动注入（wsId/projectId 从路由解析）
 *  2) 拖拽排序的乐观更新
 *  3) 燃尽图/复盘等分析数据的联动加载
 *  4) 迭代生命周期快捷方法
 *
 * 设计参考：Vue 3 Composition API + Pinia 最佳实践。
 */

import { computed, watchEffect } from "vue";
import { useSprintStore } from "@/stores/sprint";
import { useWorkspaceContext } from "./useWorkspaceContext";
import type {
  Sprint,
  CreateSprintInput,
  UpdateSprintInput,
  CompleteSprintInput,
  ListSprintsParams,
} from "@/api/services/sprint";

/* ------------------------------------------------------------------ */
/* 主 composable                                                        */
/* ------------------------------------------------------------------ */

export function useSprint() {
  const store = useSprintStore();
  const { wsId, projectId, ready } = useWorkspaceContext();

  // 上下文就绪时自动注入到 store
  watchEffect(() => {
    if (ready.value) {
      store.setContext(wsId.value, projectId.value);
    }
  });

  /* ================================================================ */
  /* 便捷导出（从 store 透传）                                          */
  /* ================================================================ */

  /** 迭代列表 */
  const sprints = computed(() => store.sprints);
  const total = computed(() => store.total);
  const loading = computed(() => store.loading);
  const error = computed(() => store.error);
  const activeSprint = computed(() => store.activeSprint);
  const sprintsByStatus = computed(() => store.sprintsByStatus);
  const isOperating = computed(() => store.isOperating);

  /** 当前迭代 */
  const currentSprint = computed(() => store.currentSprint);

  /** 迭代需求/任务/缺陷 */
  const sprintIssues = computed(() => store.sprintIssues);
  const sprintIssuesTotal = computed(() => store.sprintIssuesTotal);

  /** Backlog */
  const backlogItems = computed(() => store.backlogItems);
  const backlogTotal = computed(() => store.backlogTotal);

  /** 分析数据 */
  const burndownData = computed(() => store.burndownData);
  const reviewData = computed(() => store.reviewData);
  const velocityStats = computed(() => store.velocityStats);

  /* ================================================================ */
  /* 自动加载（首次就绪时）                                              */
  /* ================================================================ */

  let initialLoadDone = false;

  watchEffect(() => {
    if (ready.value && !initialLoadDone) {
      initialLoadDone = true;
      store.fetchVelocityStats();
    }
  });

  /* ================================================================ */
  /* 业务方法（含乐观更新与错误处理）                                    */
  /* ================================================================ */

  /** 加载迭代列表 */
  async function loadSprints(params?: ListSprintsParams) {
    ensureReady();
    return store.fetchSprints(params);
  }

  /** 创建迭代 */
  async function addSprint(input: CreateSprintInput): Promise<Sprint> {
    ensureReady();
    return store.createSprint(input);
  }

  /** 加载迭代详情（含进度/需求/任务/缺陷/燃尽图联动） */
  async function loadSprintDetail(sprintId: number) {
    ensureReady();
    const sprint = await store.fetchSprint(sprintId);
    // 并行加载关联数据
    await Promise.allSettled([
      store.fetchSprintIssues(sprintId),
      store.fetchBurndown(sprintId),
      store.fetchReview(sprintId),
    ]);
    return sprint;
  }

  /** 更新迭代 */
  async function editSprint(sprintId: number, input: UpdateSprintInput): Promise<Sprint> {
    ensureReady();
    return store.updateSprint(sprintId, input);
  }

  /** 删除迭代 */
  async function removeSprint(sprintId: number): Promise<void> {
    ensureReady();
    return store.deleteSprint(sprintId);
  }

  /** 启动迭代 */
  async function beginSprint(sprintId: number): Promise<Sprint> {
    ensureReady();
    return store.startSprint(sprintId);
  }

  /** 结束迭代 */
  async function finishSprint(sprintId: number, input: CompleteSprintInput): Promise<Sprint> {
    ensureReady();
    return store.completeSprint(sprintId, input);
  }

  /** 加载 Backlog */
  async function loadBacklog(limit = 50, offset = 0) {
    ensureReady();
    return store.fetchBacklog(limit, offset);
  }

  /** 加载迭代需求/任务/缺陷 */
  async function loadSprintIssues(sprintId: number, limit = 50, offset = 0) {
    ensureReady();
    return store.fetchSprintIssues(sprintId, limit, offset);
  }

  /**
   * 在迭代间移动需求/任务/缺陷（拖拽排序）。
   * 处理 Backlog → Sprint / Sprint → Backlog / Sprint 内重排。
   */
  async function moveIssue(params: {
    issueId: number;
    fromSprintId?: number;
    toSprintId?: number;
    sortOrder?: number;
  }) {
    ensureReady();
    const { issueId, fromSprintId, toSprintId, sortOrder } = params;

    // 从原迭代移除
    if (fromSprintId) {
      await store.removeIssueFromSprint(fromSprintId, issueId);
    }

    // 加入目标迭代
    if (toSprintId) {
      await store.addIssueToSprint(toSprintId, issueId, sortOrder);
    }
  }

  /** 刷新燃尽图 */
  async function refreshBurndown(sprintId: number) {
    ensureReady();
    return store.fetchBurndown(sprintId);
  }

  /** 刷新速率统计 */
  async function refreshVelocity() {
    ensureReady();
    return store.fetchVelocityStats();
  }

  /** 清空 sprint 状态（离开项目时调用） */
  function reset() {
    store.clear();
    initialLoadDone = false;
  }

  /* ================================================================ */
  /* 辅助                                                               */
  /* ================================================================ */

  function ensureReady() {
    if (!ready.value) {
      throw new Error("工作空间/项目上下文未就绪，请等待路由解析完成");
    }
  }

  return {
    // 状态
    sprints,
    total,
    loading,
    error,
    activeSprint,
    sprintsByStatus,
    isOperating,
    currentSprint,
    sprintIssues,
    sprintIssuesTotal,
    backlogItems,
    backlogTotal,
    burndownData,
    reviewData,
    velocityStats,
    ready,

    // 方法
    loadSprints,
    addSprint,
    loadSprintDetail,
    editSprint,
    removeSprint,
    beginSprint,
    finishSprint,
    loadBacklog,
    loadSprintIssues,
    moveIssue,
    refreshBurndown,
    refreshVelocity,
    reset,
  };
}
