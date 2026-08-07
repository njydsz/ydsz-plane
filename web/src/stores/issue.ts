/**
 * Issue 域 Pinia store — 状态管理（当前工作空间 + 项目维度的工作项缓存）。
 */
import { defineStore } from "pinia";
import { computed, ref } from "vue";
import { issueApi, type Issue, type ListIssuesParams, type State } from "@/api/services/issue";

/** Issue 域 Pinia store —— 管理当前工作空间/项目维度的工作项状态、缓存与变更操作 */
export const useIssueStore = defineStore("issue", () => {
  // --- State ---
  const states = ref<State[]>([]);
  const issues = ref<Issue[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const currentIssue = ref<Issue | null>(null);
  const error = ref<string | null>(null);

  // --- Getters ---
  const statesByGroup = computed(() => {
    const groups: Record<string, State[]> = { backlog: [], started: [], completed: [], cancelled: [] };
    for (const s of states.value) {
      if (groups[s.group]) groups[s.group].push(s);
    }
    return groups;
  });

  const issuesByState = computed(() => {
    const map: Record<number, Issue[]> = {};
    for (const s of states.value) map[s.id] = [];
    for (const iss of issues.value) {
      if (map[iss.state_id]) map[iss.state_id].push(iss);
    }
    return map;
  });

  // --- Actions ---
  async function fetchStates(wsId: number, projectId: number) {
    states.value = await issueApi.listStates(wsId, projectId);
  }

  async function fetchIssues(wsId: number, projectId: number, params: ListIssuesParams = {}) {
    loading.value = true;
    error.value = null;
    try {
      const res = await issueApi.listIssues(wsId, projectId, params);
      issues.value = res.results;
      total.value = res.total;
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : "加载工作项失败";
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function fetchIssue(wsId: number, projectId: number, issueId: number) {
    loading.value = true;
    try {
      currentIssue.value = await issueApi.getIssue(wsId, projectId, issueId);
    } finally {
      loading.value = false;
    }
  }

  async function createIssue(wsId: number, projectId: number, input: Parameters<typeof issueApi.createIssue>[2]) {
    const iss = await issueApi.createIssue(wsId, projectId, input);
    issues.value.unshift(iss);
    total.value += 1;
    return iss;
  }

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

  async function transitionIssue(wsId: number, projectId: number, issueId: number, toStateId: number) {
    const iss = await issueApi.transition(wsId, projectId, issueId, toStateId);
    const idx = issues.value.findIndex((i) => i.id === issueId);
    if (idx >= 0) issues.value[idx] = iss;
    if (currentIssue.value?.id === issueId) currentIssue.value = iss;
    return iss;
  }

  async function deleteIssue(wsId: number, projectId: number, issueId: number) {
    await issueApi.deleteIssue(wsId, projectId, issueId);
    issues.value = issues.value.filter((i) => i.id !== issueId);
    total.value = Math.max(0, total.value - 1);
  }

  // --- State helpers ---
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
