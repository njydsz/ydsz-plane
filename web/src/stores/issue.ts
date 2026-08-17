/**
 * Issue 域 Pinia store — 状态管理（当前工作空间 + 项目维度的需求/任务/缺陷缓存）。
 */
import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { issueApi, type Issue, type ListIssuesParams, type State } from "@/api/services/issue";

/** Issue 域 Pinia store —— 管理当前工作空间/项目维度的需求/任务/缺陷状态、缓存与变更操作 */
export const useIssueStore = defineStore("issue", () => {
  // --- State ---
  /** 项目内全部状态定义（含分组） */
  const states = ref<State[]>([]);
  /** 当前查询条件下的需求/任务/缺陷列表 */
  const issues = ref<Issue[]>([]);
  /** 符合条件的总条数（用于分页） */
  const total = ref(0);
  /** 列表/详情请求进行中的标志位 */
  const loading = ref(false);
  /** 当前正在查看的需求/任务/缺陷详情 */
  const currentIssue = ref<Issue | null>(null);
  /** 最近一次请求的错误信息 */
  const error = ref<string | null>(null);

  // --- Getters ---
  /** 按状态分组（backlog/started/completed/cancelled）索引各状态定义 */
  const statesByGroup = computed(() => {
    const groups: Record<string, State[]> = { backlog: [], started: [], completed: [], cancelled: [] };
    for (const s of states.value) {
      if (groups[s.group]) groups[s.group].push(s);
    }
    return groups;
  });

  /** 按状态 ID 索引需求/任务/缺陷，便于看板按列渲染 */
  const issuesByState = computed(() => {
    const map: Record<number, Issue[]> = {};
    for (const s of states.value) map[s.id] = [];
    for (const iss of issues.value) {
      if (map[iss.state_id]) map[iss.state_id].push(iss);
    }
    return map;
  });

  // --- Actions ---
  /** 拉取指定项目的状态定义列表 */
  async function fetchStates(wsId: number, projectId: number) {
    states.value = await issueApi.listStates(wsId, projectId);
  }

  /**
   * 分页拉取需求/任务/缺陷列表并缓存。
   * 失败时记录 error 并向上抛出，由调用方决定是否展示错误态。
   */
  async function fetchIssues(wsId: number, projectId: number, params: ListIssuesParams = {}) {
    loading.value = true;
    error.value = null;
    try {
      const res = await issueApi.listIssues(wsId, projectId, params);
      issues.value = res.results;
      total.value = res.total;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加载需求/任务/缺陷失败";
      throw e;
    } finally {
      loading.value = false;
    }
  }

  /** 拉取单个需求/任务/缺陷详情并写入 currentIssue */
  async function fetchIssue(wsId: number, projectId: number, issueId: number) {
    loading.value = true;
    try {
      currentIssue.value = await issueApi.getIssue(wsId, projectId, issueId);
    } finally {
      loading.value = false;
    }
  }

  /** 创建需求/任务/缺陷：插入列表头部并累加总数，返回新需求/任务/缺陷 */
  async function createIssue(wsId: number, projectId: number, input: Parameters<typeof issueApi.createIssue>[2]) {
    const iss = await issueApi.createIssue(wsId, projectId, input);
    issues.value.unshift(iss);
    total.value += 1;
    return iss;
  }

  /**
   * 更新需求/任务/缺陷：同时同步列表项与当前详情（若命中）。
   * @returns 更新后的需求/任务/缺陷
   */
  async function updateIssue(
    wsId: number,
    projectId: number,
    issueId: number,
    input: Parameters<typeof issueApi.updateIssue>[3],
  ) {
    const iss = await issueApi.updateIssue(wsId, projectId, issueId, input);
    const idx = issues.value.findIndex((i) => i.id === issueId);
    if (idx >= 0) issues.value[idx] = iss;
    if (currentIssue.value?.id === issueId) currentIssue.value = iss;
    return iss;
  }

  /**
   * 流转需求/任务/缺陷状态：后端执行后同步本地列表与详情。
   * @param toStateId 目标状态 ID
   * @returns 流转后的需求/任务/缺陷
   */
  async function transitionIssue(wsId: number, projectId: number, issueId: number, toStateId: number) {
    const iss = await issueApi.transition(wsId, projectId, issueId, toStateId);
    const idx = issues.value.findIndex((i) => i.id === issueId);
    if (idx >= 0) issues.value[idx] = iss;
    if (currentIssue.value?.id === issueId) currentIssue.value = iss;
    return iss;
  }

  /** 删除需求/任务/缺陷：从列表移除并递减总数（下限为 0） */
  async function deleteIssue(wsId: number, projectId: number, issueId: number) {
    await issueApi.deleteIssue(wsId, projectId, issueId);
    issues.value = issues.value.filter((i) => i.id !== issueId);
    total.value = Math.max(0, total.value - 1);
  }

  // --- State helpers ---
  /** 重置全部状态（常用于切换项目/工作空间时） */
  function clear() {
    states.value = [];
    issues.value = [];
    total.value = 0;
    currentIssue.value = null;
    error.value = null;
  }

  return {
    // state
    states,
    issues,
    total,
    loading,
    currentIssue,
    error,
    // getters
    statesByGroup,
    issuesByState,
    // actions
    fetchStates,
    fetchIssues,
    fetchIssue,
    createIssue,
    updateIssue,
    transitionIssue,
    deleteIssue,
    clear,
  };
});
