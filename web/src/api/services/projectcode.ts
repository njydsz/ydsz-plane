/**
 * Projects-as-Code API — YAML 声明式定义项目（状态 / 标签 / 模块 / 成员 / 设置）。
 *
 * 路由: /api/v1/workspaces/:ws/project-code
 *   - GET    /configs            列出已保存配置
 *   - POST   /configs            保存 YAML 配置
 *   - GET    /configs/:id        配置详情
 *   - DELETE /configs/:id        删除配置
 *   - POST   /parse              校验并返回解析后的规范（不保存）
 *   - POST   /diff               规范 vs 既有项目 diff
 *   - POST   /apply              应用规范（创建/更新项目）
 *   - GET    /export/:project_id 导出既有项目为 YAML
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 状态分组（与后端 issue.StateGroup 对齐） */
export type StateGroup = "backlog" | "started" | "completed" | "cancelled";

/** 项目角色（与 project_role_enum 对齐） */
export type ProjectMemberRole = "admin" | "member";

/** 状态定义 */
export interface StateSpec {
  name: string;
  /** 状态分组：backlog / started / completed / cancelled（后端仅校验取值，不区分大小写） */
  group: StateGroup;
  color?: string;
}

/** 标签定义 */
export interface LabelSpec {
  name: string;
  color?: string;
}

/** 模块定义 */
export interface ModuleSpec {
  name: string;
  /** 负责人邮箱（可选） */
  lead?: string;
}

/** 成员定义 */
export interface MemberSpec {
  email: string;
  /** 项目角色：admin / member（默认 member） */
  role: ProjectMemberRole;
}

/** 项目设置 */
export interface SpecSettings {
  default_assignee?: string;
  estimate_points?: number[];
}

/** 项目 YAML 规范（解析后的结构） */
export interface ProjectSpec {
  version: string;
  name: string;
  key: string;
  color?: string;
  icon?: string;
  description?: string;
  states: StateSpec[];
  labels: LabelSpec[];
  modules: ModuleSpec[];
  members: MemberSpec[];
  settings: SpecSettings;
}

/** 已保存的配置记录 */
export interface ProjectCodeConfig {
  id: number;
  workspace_id: number;
  /** 关联的既有项目 ID（0 / null 表示未绑定） */
  project_id?: number | null;
  name: string;
  spec_key?: string;
  version: string;
  yaml: string;
  spec?: ProjectSpec | null;
  sha256: string;
  config_version: number;
  created_by: number;
  updated_by: number;
  created_at: string;
  updated_at: string;
}

/** 保存配置入参 */
export interface SaveProjectCodeInput {
  yaml: string;
  name?: string;
  project_id?: number;
}

/** 差异类型 */
export type DiffChangeKind = "add" | "update" | "remove" | "noop";

/** 单条结构化差异 */
export interface ProjectCodeDiffItem {
  kind: DiffChangeKind;
  /** project / state / label / module / member / setting */
  resource: string;
  name?: string;
  field?: string;
  old?: unknown;
  new?: unknown;
}

/** 差异汇总结果 */
export interface ProjectCodeDiffResult {
  project_id?: number;
  spec_name: string;
  has_changes: boolean;
  items: ProjectCodeDiffItem[];
  add_count: number;
  update_count: number;
  remove_count: number;
}

/** 应用结果 */
export interface ProjectCodeApplyResult {
  project_id: number;
  /** true=新建；false=更新 */
  created: boolean;
  project_slug: string;
  identifier: string;
}

/** 导出结果 */
export interface ProjectCodeExportResult {
  spec: ProjectSpec;
  yaml: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** Projects-as-Code 域 API */
export const projectCodeApi = {
  /** 列出工作空间下已保存的配置 */
  listConfigs: (workspaceId: number | string, projectId?: number) =>
    wrap<{ items: ProjectCodeConfig[] }>(
      http.get(`/workspaces/${workspaceId}/project-code/configs`, {
        params: projectId ? { project_id: projectId } : {},
      }),
    ).then((r) => r.items ?? []),

  /** 获取配置详情 */
  getConfig: (workspaceId: number | string, id: number | string) =>
    wrap<ProjectCodeConfig>(
      http.get(`/workspaces/${workspaceId}/project-code/configs/${id}`),
    ),

  /** 保存 YAML 配置 */
  saveConfig: (
    workspaceId: number | string,
    input: SaveProjectCodeInput,
  ) =>
    wrap<ProjectCodeConfig>(
      http.post(`/workspaces/${workspaceId}/project-code/configs`, {
        yaml: input.yaml,
        name: input.name || undefined,
        project_id: input.project_id || undefined,
      }),
    ),

  /** 删除配置（软删除） */
  deleteConfig: (workspaceId: number | string, id: number | string) =>
    wrap<void>(
      http.delete(`/workspaces/${workspaceId}/project-code/configs/${id}`),
    ),

  /** 校验并解析 YAML（不保存） */
  parse: (workspaceId: number | string, yaml: string) =>
    wrap<ProjectSpec>(
      http.post(`/workspaces/${workspaceId}/project-code/parse`, { yaml }),
    ),

  /** 规范 vs 既有项目 diff */
  diff: (workspaceId: number | string, projectId: number, yaml: string) =>
    wrap<ProjectCodeDiffResult>(
      http.post(`/workspaces/${workspaceId}/project-code/diff`, {
        project_id: projectId,
        yaml,
      }),
    ),

  /** 应用规范（创建 / 更新项目） */
  apply: (
    workspaceId: number | string,
    yaml: string,
    projectId?: number,
  ) =>
    wrap<ProjectCodeApplyResult>(
      http.post(`/workspaces/${workspaceId}/project-code/apply`, {
        yaml,
        project_id: projectId || undefined,
      }),
    ),

  /** 导出既有项目为 YAML 规范 */
  exportProject: (workspaceId: number | string, projectId: number | string) =>
    wrap<ProjectCodeExportResult>(
      http.get(
        `/workspaces/${workspaceId}/project-code/export/${projectId}`,
      ),
    ),
};
