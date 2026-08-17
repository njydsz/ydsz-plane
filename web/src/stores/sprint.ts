/**
 * Sprint 域 Pinia store — 迭代管理的全局状态核心。
 *
 * 职责：
 *  1) 迭代 CRUD 操作与列表缓存
 *  2) 当前迭代详情与生命周期管理
 *  3) 迭代内需求/任务/缺陷（Sprint Issues）与 Backlog 数据
 *  4) 燃尽图、复盘、速率统计等分析数据
 *  5) 乐观更新与错误回滚
 *
 * 设计参考互联网大厂标准（字节/阿里/腾讯前端状态管理规范）：
 *  - 单一数据源：所有 Sprint 视图共用此 store
 *  - 按需加载：列表/详情/分析数据独立加载，避免不必要请求
 *  - 乐观更新：拖拽排序、状态变更先更新 UI 再异步确认
 *  - 错误恢复：失败时回滚到 API 返回的最新数据
 */

import { defineStore } from "pinia";
import { computed, ref } from "vue";
import {
  sprintApi,
  type Sprint,
  type SprintProgress,
  type ReviewSnapshot,
  type BurndownPoint,
  type VelocityStats,
  type SprintIssueView,
  type BacklogItem,
  type CreateSprintInput,
  type UpdateSprintInput,
  type CompleteSprintInput,
  type ListSprintsParams,
  type SprintStatus,
} from "@/api/services/sprint";

/* ------------------------------------------------------------------ */
/* Store                                                               */
/* ------------------------------------------------------------------ */

export const useSprintStore = defineStore("sprint", () => {
  /* ================================================================ */
  /* State                                                              */
  /* ================================================================ */

  /** 迭代列表（当前项目） */
  const sprints = ref<Sprint[]>([]);
  const total = ref(0);

  /** 当前查看/操作的迭代详情 */
  const currentSprint = ref<Sprint | null>(null);

  /** 迭代内需求/任务/缺陷 */
  const sprintIssues = ref<SprintIssueView[]>([]);
  const sprintIssuesTotal = ref(0);

  /** Backlog 需求/任务/缺陷 */
  const backlogItems = ref<BacklogItem[]>([]);
  const backlogTotal = ref(0);

  /** 分析数据 */
  const burndownData = ref<BurndownPoint[]>([]);
  const reviewData = ref<ReviewSnapshot | null>(null);
  const velocityStats = ref<VelocityStats | null>(null);

  /** 通用状态 */
  const loading = ref(false);
  const error = ref<string | null>(null);

  /** 当前操作的 workspace_id / project_id（由视图注入） */
  const wsId = ref(0);
  const projectId = ref(0);

  /* ================================================================ */
  /* Getters                                                            */
  /* ================================================================ */

  /** 当前活跃迭代 */
  const activeSprint = computed<Sprint | undefined>(() =>
    sprints.value.find((s) => s.status === "active"),
  );

  /** 按状态分组的迭代 */
  const sprintsByStatus = computed<Record<SprintStatus, Sprint[]>>(() => {
    const groups: Record<SprintStatus, Sprint[]> = {
      planned: [],
      active: [],
      completed: [],
    };
    for (const s of sprints.value) {
      if (groups[s.status]) groups[s.status].push(s);
    }
    return groups;
  });

  /** 是否有进行中的操作 */
  const isOperating = ref(false);

  /* ================================================================ */
  /* Actions: 上下文注入                                                  */
  /* ================================================================ */

  /** 设置当前工作空间和项目上下文 */
  function setContext(w: number, p: number) {
    wsId.value = w;
    projectId.value = p;
  }

  /* ================================================================ */
  /* Actions: 迭代列表 CRUD                                              */
  /* ================================================================ */

  /** 加载迭代列表 */
  async function fetchSprints(params: ListSprintsParams = {}) {
    loading.value = true;
    error.value = null;
    try {
      const res = await sprintApi.listSprints(wsId.value, projectId.value, params);
      sprints.value = res.results;
      total.value = res.total;
      return res;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加载迭代列表失败";
      throw e;
    } finally {
      loading.value = false;
    }
  }

  /** 创建迭代 */
  async function createSprint(input: CreateSprintInput): Promise<Sprint> {
    isOperating.value = true;
    error.value = null;
    try {
      const sprint = await sprintApi.createSprint(wsId.value, projectId.value, input);
      sprints.value.unshift(sprint);
      total.value += 1;
      return sprint;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "创建迭代失败";
      throw e;
    } finally {
      isOperating.value = false;
    }
  }

  /** 加载单个迭代详情 */
  async function fetchSprint(sprintId: number): Promise<Sprint> {
    loading.value = true;
    error.value = null;
    try {
      const sprint = await sprintApi.getSprint(wsId.value, projectId.value, sprintId);
      currentSprint.value = sprint;
      // 同步更新列表中的缓存
      const idx = sprints.value.findIndex((s) => s.id === sprintId);
      if (idx >= 0) sprints.value[idx] = sprint;
      return sprint;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加载迭代详情失败";
      throw e;
    } finally {
      loading.value = false;
    }
  }

  /** 更新迭代 */
  async function updateSprint(sprintId: number, input: UpdateSprintInput): Promise<Sprint> {
    isOperating.value = true;
    error.value = null;
    try {
      const sprint = await sprintApi.updateSprint(wsId.value, projectId.value, sprintId, input);
      const idx = sprints.value.findIndex((s) => s.id === sprintId);
      if (idx >= 0) sprints.value[idx] = sprint;
      if (currentSprint.value?.id === sprintId) currentSprint.value = sprint;
      return sprint;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "更新迭代失败";
      throw e;
    } finally {
      isOperating.value = false;
    }
  }

  /** 删除（归档）迭代 */
  async function deleteSprint(sprintId: number): Promise<void> {
    isOperating.value = true;
    error.value = null;
    try {
      await sprintApi.deleteSprint(wsId.value, projectId.value, sprintId);
      sprints.value = sprints.value.filter((s) => s.id !== sprintId);
      total.value = Math.max(0, total.value - 1);
      if (currentSprint.value?.id === sprintId) currentSprint.value = null;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "删除迭代失败";
      throw e;
    } finally {
      isOperating.value = false;
    }
  }

  /* ================================================================ */
  /* Actions: 生命周期                                                    */
  /* ================================================================ */

  /** 启动迭代 */
  async function startSprint(sprintId: number): Promise<Sprint> {
    isOperating.value = true;
    try {
      const sprint = await sprintApi.startSprint(wsId.value, projectId.value, sprintId);
      const idx = sprints.value.findIndex((s) => s.id === sprintId);
      if (idx >= 0) sprints.value[idx] = sprint;
      if (currentSprint.value?.id === sprintId) currentSprint.value = sprint;
      return sprint;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "启动迭代失败";
      throw e;
    } finally {
      isOperating.value = false;
    }
  }

  /** 结束迭代 */
  async function completeSprint(sprintId: number, input: CompleteSprintInput): Promise<Sprint> {
    isOperating.value = true;
    try {
      const sprint = await sprintApi.completeSprint(wsId.value, projectId.value, sprintId, input);
      const idx = sprints.value.findIndex((s) => s.id === sprintId);
      if (idx >= 0) sprints.value[idx] = sprint;
      if (currentSprint.value?.id === sprintId) currentSprint.value = sprint;
      return sprint;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "结束迭代失败";
      throw e;
    } finally {
      isOperating.value = false;
    }
  }

  /* ================================================================ */
  /* Actions: 迭代进度                                                  */
  /* ================================================================ */

  /** 加载迭代实时进度 */
  async function fetchProgress(sprintId: number): Promise<SprintProgress> {
    const progress = await sprintApi.getSprintProgress(wsId.value, projectId.value, sprintId);
    if (currentSprint.value?.id === sprintId) {
      currentSprint.value = { ...currentSprint.value, progress };
    }
    return progress;
  }

  /* ================================================================ */
  /* Actions: 规划（Sprint Issues / Backlog）                           */
  /* ================================================================ */

  /** 加载迭代内需求/任务/缺陷 */
  async function fetchSprintIssues(sprintId: number, limit = 50, offset = 0) {
    loading.value = true;
    try {
      const res = await sprintApi.listSprintIssues(wsId.value, projectId.value, sprintId, limit, offset);
      sprintIssues.value = res.results;
      sprintIssuesTotal.value = res.total;
      return res;
    } finally {
      loading.value = false;
    }
  }

  /** 添加需求/任务/缺陷到迭代（乐观更新） */
  async function addIssueToSprint(sprintId: number, issueId: number, sortOrder = 65535) {
    await sprintApi.addIssue(wsId.value, projectId.value, sprintId, issueId, sortOrder);
    // 从 Backlog 中移除
    backlogItems.value = backlogItems.value.filter((b) => b.issue_id !== issueId);
    backlogTotal.value = Math.max(0, backlogTotal.value - 1);
    // 重新加载迭代需求/任务/缺陷以获取最新状态
    await fetchSprintIssues(sprintId);
  }

  /** 从迭代移除需求/任务/缺陷（乐观更新） */
  async function removeIssueFromSprint(sprintId: number, issueId: number) {
    // 乐观移除
    const backup = [...sprintIssues.value];
    sprintIssues.value = sprintIssues.value.filter((i) => i.issue_id !== issueId);
    sprintIssuesTotal.value = Math.max(0, sprintIssuesTotal.value - 1);

    try {
      await sprintApi.removeIssue(wsId.value, projectId.value, sprintId, issueId);
    } catch (e: unknown) {
      // 失败回滚
      sprintIssues.value = backup;
      sprintIssuesTotal.value = backup.length;
      throw e;
    }
  }

  /** 加载 Backlog */
  async function fetchBacklog(limit = 50, offset = 0) {
    loading.value = true;
    try {
      const res = await sprintApi.getBacklog(wsId.value, projectId.value, limit, offset);
      backlogItems.value = res.results;
      backlogTotal.value = res.total;
      return res;
    } finally {
      loading.value = false;
    }
  }

  /* ================================================================ */
  /* Actions: 分析数据                                                  */
  /* ================================================================ */

  /** 加载燃尽图数据 */
  async function fetchBurndown(sprintId: number) {
    const res = await sprintApi.burndown(wsId.value, projectId.value, sprintId);
    burndownData.value = res.points;
    return res;
  }

  /** 加载复盘数据 */
  async function fetchReview(sprintId: number) {
    const data = await sprintApi.getReview(wsId.value, projectId.value, sprintId);
    reviewData.value = data;
    return data;
  }

  /** 加载速率建议 */
  async function fetchVelocityStats() {
    const stats = await sprintApi.suggestCapacity(wsId.value, projectId.value);
    velocityStats.value = stats;
    return stats;
  }

  /* ================================================================ */
  /* Actions: 状态重置                                                  */
  /* ================================================================ */

  function clear() {
    sprints.value = [];
    total.value = 0;
    currentSprint.value = null;
    sprintIssues.value = [];
    sprintIssuesTotal.value = 0;
    backlogItems.value = [];
    backlogTotal.value = 0;
    burndownData.value = [];
    reviewData.value = null;
    velocityStats.value = null;
    loading.value = false;
    error.value = null;
    isOperating.value = false;
  }

  /* ================================================================ */
  /* Export                                                             */
  /* ================================================================ */

  return {
    // State
    sprints,
    total,
    currentSprint,
    sprintIssues,
    sprintIssuesTotal,
    backlogItems,
    backlogTotal,
    burndownData,
    reviewData,
    velocityStats,
    loading,
    error,
    wsId,
    projectId,
    isOperating,

    // Getters
    activeSprint,
    sprintsByStatus,

    // Actions: Context
    setContext,

    // Actions: CRUD
    fetchSprints,
    createSprint,
    fetchSprint,
    updateSprint,
    deleteSprint,

    // Actions: Lifecycle
    startSprint,
    completeSprint,

    // Actions: Progress
    fetchProgress,

    // Actions: Planning
    fetchSprintIssues,
    addIssueToSprint,
    removeIssueFromSprint,
    fetchBacklog,

    // Actions: Analytics
    fetchBurndown,
    fetchReview,
    fetchVelocityStats,

    // Actions: Reset
    clear,
  };
});
