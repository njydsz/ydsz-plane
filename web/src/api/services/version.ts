/**
 * Version 域 API — 对接后端 Version 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type VersionStatus = "planning" | "active" | "released" | "archived";

export interface ChecklistItem {
  id: string;
  label: string;
  required: boolean;
  checked: boolean;
}

export interface SprintProgressRef {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
}

export interface SprintRef {
  sprint_id: number;
  name: string;
  status: string;
  start_date?: string;
  end_date?: string;
  completed_at?: string;
  progress?: SprintProgressRef;
}

export interface VersionProgress {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
  completion_rate: number;
  by_state_group?: Record<string, number>;
  sprint_count: number;
}

export interface QualityMetrics {
  total_bugs: number;
  open_bugs: number;
  critical_bugs: number;
  major_bugs: number;
  bug_by_severity?: Record<string, number>;
  found_bug_count: number;
  fixed_bug_count: number;
  fix_rate: number;
  pass_quality_gate: boolean;
}

export interface DeliveryReport {
  generated_at: string;
  sprint_count: number;
  total_points: number;
  completed_points: number;
  total_issues: number;
  completed_issues: number;
  bug_count: number;
  fixed_bug_count: number;
  pass_rate: number;
  eligible_to_release: boolean;
}

export interface Version {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  semver: string;
  description?: string;
  status: VersionStatus;
  checklist?: ChecklistItem[];
  release_notes?: string;
  delivered_at?: string;
  target_date?: string;
  archived_at?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  sprints?: SprintRef[];
  progress?: VersionProgress;
  quality?: QualityMetrics;
  delivery_report?: DeliveryReport;
}

export interface BugVersionView {
  issue_id: number;
  identifier: string;
  name: string;
  severity?: number;
  found_phase?: string;
  state_name: string;
  state_group: string;
  found_version?: string;
  fix_version?: string;
  root_cause_category?: string;
}

export interface CreateVersionInput {
  name: string;
  semver: string;
  description?: string;
  target_date?: string;
  checklist?: ChecklistItem[];
}

export interface UpdateVersionInput {
  name?: string;
  description?: string;
  semver?: string;
  target_date?: string;
  checklist?: ChecklistItem[];
  version: number;
}

export interface ReleaseVersionInput {
  draft_override?: string;
  force_checklist?: boolean;
  add_known_issues_to_notes?: boolean;
}

export interface AddSprintInput {
  sprint_id: number;
}

export interface ListVersionsParams {
  status?: VersionStatus;
  limit?: number;
  offset?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

export const versionApi = {
  listVersions: (wsId: number, projectId: number, params?: ListVersionsParams) =>
    wrap<{ results: Version[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions`, { params }),
    ),
  getVersion: (wsId: number, projectId: number, versionId: number) =>
    wrap<Version>(http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`)),
  createVersion: (wsId: number, projectId: number, input: CreateVersionInput) =>
    wrap<Version>(http.post(`/workspaces/${wsId}/projects/${projectId}/versions`, input)),
  updateVersion: (wsId: number, projectId: number, versionId: number, input: UpdateVersionInput) =>
    wrap<Version>(
      http.patch(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`, input),
    ),
  deleteVersion: (wsId: number, projectId: number, versionId: number) =>
    wrap<void>(http.delete(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`)),

  // --- 版本日生命周期 ---
  activateVersion: (wsId: number, projectId: number, versionId: number) =>
    wrap<Version>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/activate`),
    ),
  releaseVersion: (
    wsId: number,
    projectId: number,
    versionId: number,
    input: ReleaseVersionInput,
  ) =>
    wrap<Version>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release`, input),
    ),
  archiveVersion: (wsId: number, projectId: number, versionId: number) =>
    wrap<Version>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/archive`),
    ),

  // --- 进度 / 质量 / 交付报告 ---
  getVersionProgress: (wsId: number, projectId: number, versionId: number) =>
    wrap<VersionProgress>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/progress`),
    ),
  getQuality: (wsId: number, projectId: number, versionId: number) =>
    wrap<QualityMetrics>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/quality`),
    ),
  getDeliveryReport: (wsId: number, projectId: number, versionId: number) =>
    wrap<DeliveryReport>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/delivery-report`),
    ),
  getReleaseNotes: (wsId: number, projectId: number, versionId: number) =>
    wrap<{ version_id: number; release_notes: string }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release-notes`),
    ),
  regenerateNotes: (
    wsId: number,
    projectId: number,
    versionId: number,
    addKnown = true,
  ) =>
    wrap<{ release_notes: string }>(
      http.post(
        `/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release-notes/regenerate`,
        { add_known_issues_to_notes: addKnown },
      ),
    ),

  // --- 缺陷面板 ---
  getDefectPanel: (wsId: number, projectId: number, versionId: number) =>
    wrap<{ results: BugVersionView[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/defects`),
    ),

  // --- 跨版本缺陷过滤 ---
  filterDefects: (
    wsId: number,
    projectId: number,
    params: {
      found_version_id?: number;
      fix_version_id?: number;
      state_group?: string;
      severity?: number;
      limit?: number;
      offset?: number;
    },
  ) =>
    wrap<{ results: BugVersionView[]; total: number }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/defects`, { params }),
    ),

  // --- 迭代聚合 ---
  listVersionSprints: (wsId: number, projectId: number, versionId: number) =>
    wrap<{ results: SprintRef[] }>(
      http.get(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`),
    ),
  addSprint: (wsId: number, projectId: number, versionId: number, input: AddSprintInput) =>
    wrap<void>(
      http.post(`/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`, input),
    ),
  removeSprint: (wsId: number, projectId: number, versionId: number, sprintId: number) =>
    wrap<void>(
      http.delete(
        `/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints/${sprintId}`,
      ),
    ),
};
