/**
 * Version 域 API — 对接后端 Version 域 REST 接口。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

export type VersionStatus = "planning" | "active" | "released" | "archived";

/** 发布检查清单条目 */
export interface ChecklistItem {
  id: string;
  label: string;
  required: boolean;
  checked: boolean;
}

/** 迭代进度摘要（版本聚合视图用） */
export interface SprintProgressRef {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
}

/** 版本关联的迭代摘要 */
export interface SprintRef {
  sprint_id: number;
  name: string;
  status: string;
  start_date?: string;
  end_date?: string;
  completed_at?: string;
  progress?: SprintProgressRef;
}

/** 版本实时进度聚合 */
export interface VersionProgress {
  total_points: number;
  done_points: number;
  total_issues: number;
  done_issues: number;
  completion_rate: number;
  by_state_group?: Record<string, number>;
  sprint_count: number;
}

/** 版本质量指标（发布准出校验用） */
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

/** 版本交付报告（发布前快照） */
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

/** 版本聚合根 */
export interface Version {
  id: number;
  workspace_id: number;
  project_id: number;
  name: string;
  semver: string;
  description?: string;
  status: VersionStatus;
  start_date?: string;
  end_date?: string;
  target_date?: string;
  checklist?: ChecklistItem[];
  release_notes?: string;
  delivered_at?: string;
  archived_at?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  sprints?: SprintRef[];
  progress?: VersionProgress;
  quality?: QualityMetrics;
  delivery_report?: DeliveryReport;
}

/** 缺陷面板中的缺陷视图投影 */
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

/** 创建版本入参 */
export interface CreateVersionInput {
  name: string;
  semver: string;
  description?: string;
  start_date?: string;
  end_date?: string;
  target_date?: string;
  checklist?: ChecklistItem[];
}

/** 更新版本入参（可选字段 + 乐观锁 version） */
export interface UpdateVersionInput {
  name?: string;
  description?: string;
  start_date?: string;
  end_date?: string;
  semver?: string;
  target_date?: string;
  checklist?: ChecklistItem[];
  version: number;
}

/** 发布版本入参（草稿覆盖 / 强制清单 / 已知缺陷写入发布说明） */
export interface ReleaseVersionInput {
  draft_override?: string;
  force_checklist?: boolean;
  add_known_issues_to_notes?: boolean;
}

/** 将迭代归属到版本的入参 */
export interface AddSprintInput {
  sprint_id: number;
}

/** 版本列表查询参数 */
export interface ListVersionsParams {
  status?: VersionStatus;
  limit?: number;
  offset?: number;
}

/* ------------------------------------------------------------------ */
/* API calls                                                          */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/** 版本域 API：CRUD / 生命周期 / 进度质量 / 交付报告 / 缺陷面板 / 迭代聚合 */
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

  // --- 版本生命周期 ---
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
