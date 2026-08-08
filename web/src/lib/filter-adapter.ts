/**
 * FilterAdapter — 多视图过滤/排序/显示属性的统一序列化层。
 *
 * 对标 Plane 的 FilterAdapter 模式：提供 toInternal / toExternal 双向转换，
 * 将"URL 查询参数 ↔ 后端 ListIssuesParams ↔ 前端过滤状态"三套表示统一为
 * 一份强类型的 FilterState。
 *
 * 设计目标：
 *   - 看板、列表、日历、甘特图四视图共用同一份过滤定义
 *   - 服务端持久化（view_preferences.filters jsonb）与 URL 深链共享同一 schema
 *   - 消除当前 `unknown` + 双份 localStorage 的混乱
 */

import type { ListIssuesParams } from "@/api/services/issue";

/** 状态分组（看板列、列表快速过滤共用） */
export type StateGroup = "backlog" | "started" | "completed" | "cancelled";

/** 工作项类型 */
export type IssueTypeCode = "requirement" | "task" | "defect" | "epic";

/** 优先级 */
export type IssuePriority = "urgent" | "high" | "medium" | "low" | "none";

/** 视图类型 */
export type ViewType = "kanban" | "list" | "calendar" | "gantt" | "spreadsheet";

// =====================================================================
// 强类型过滤状态（替代 ViewPreference.filters 的 unknown）
// =====================================================================

/** 视图过滤器 — 所有视图共有的基础过滤条件 */
export interface FilterState {
  /** 全文搜索关键词 */
  search?: string;
  /** 状态分组过滤 */
  group?: StateGroup;
  /** 工作项类型过滤 */
  type?: IssueTypeCode;
  /** 优先级过滤 */
  priority?: IssuePriority;
  /** 指派人 ID */
  assignee_id?: number;
  /** 标签 ID */
  label_id?: number;
  /** 模块 ID */
  module_id?: number;
  /** 迭代 ID */
  sprint_id?: number;
  /** 父工作项 ID */
  parent_id?: number;
  /** 严重级别下限（含） */
  severity_from?: number;
  /** 发现版本 ID */
  found_version_id?: number;
  /** 修复版本 ID */
  fix_version_id?: number;
  /** 起始日期下限 */
  start_date_from?: string;
  /** 截止日期上限 */
  target_date_to?: string;
}

/** 排序状态 */
export interface SortState {
  /** 排序字段，前缀 "-" 表示降序 */
  field: string;
  /** 是否降序 */
  desc: boolean;
}

/** 显示属性配置 */
export interface DisplayState {
  /** 看板：列宽度 px */
  column_width?: number;
  /** 列表：可见列 ID 数组 */
  visible_columns?: string[];
  /** 看板/列表：分组维度 */
  group_by?: "state" | "priority" | "assignee" | "type" | "label";
}

/** 视图偏好完整的结构化载荷（替代当前 SavePreferenceInput 的 unknown 字段） */
export interface PreferencePayload {
  filters?: FilterState;
  sort?: SortState;
  display?: DisplayState;
  extra?: Record<string, unknown>;
}

// =====================================================================
// FilterAdapter — 核心转换逻辑
// =====================================================================

/**
 * 将 URL 查询参数反序列化为内部 FilterState。
 *
 * 示例：?type=defect&priority=high&group=started
 *   → { type: "defect", priority: "high", group: "started" }
 */
export function urlToFilter(urlParams: Record<string, string | undefined>): FilterState {
  const f: FilterState = {};
  const set = (key: keyof FilterState, transform?: (v: string) => unknown) => {
    const raw = urlParams[key];
    if (raw != null !== undefined && raw !== "") {
      (f as Record<string, unknown>)[key] = transform ? transform(raw) : raw;
    }
  };

  if (urlParams.search) f.search = urlParams.search;
  if (isValidGroup(urlParams.group)) f.group = urlParams.group;
  if (isValidType(urlParams.type)) f.type = urlParams.type;
  if (isValidPriority(urlParams.priority)) f.priority = urlParams.priority;
  set("assignee_id", (v) => Number(v));
  set("label_id", (v) => Number(v));
  set("module_id", (v) => Number(v));
  set("sprint_id", (v) => Number(v));
  set("parent_id", (v) => Number(v));
  set("severity_from", (v) => Number(v));
  set("found_version_id", (v) => Number(v));
  set("fix_version_id", (v) => Number(v));
  if (urlParams.start_date_from) f.start_date_from = urlParams.start_date_from;
  if (urlParams.target_date_to) f.target_date_to = urlParams.target_date_to;

  return f;
}

/**
 * 将内部 FilterState 序列化为 URL 查询参数（仅包含有值的字段）。
 */
export function filterToUrl(f: FilterState): Record<string, string> {
  const params: Record<string, string> = {};
  for (const [key, val] of Object.entries(f)) {
    if (val === undefined || val === null || val === "") continue;
    if (typeof val === "number" && isNaN(val)) continue;
    params[key] = String(val);
  }
  return params;
}

/**
 * 将内部 FilterState 转为后端 ListIssuesParams（查询字符串）。
 */
export function filterToListParams(f: FilterState): ListIssuesParams {
  const p: ListIssuesParams = {};
  if (f.search?.trim()) p.search = f.search.trim();
  if (f.group) p.group = f.group;
  if (f.type) p.type = f.type as ListIssuesParams["type"];
  if (f.priority) p.priority = f.priority as ListIssuesParams["priority"];
  if (f.assignee_id != null && !isNaN(f.assignee_id)) p.assignee_id = f.assignee_id;
  if (f.label_id != null && !isNaN(f.label_id)) p.label_id = f.label_id;
  if (f.module_id != null && !isNaN(f.module_id)) p.module_id = f.module_id;
  if (f.sprint_id != null && !isNaN(f.sprint_id)) p.sprint_id = f.sprint_id;
  if (f.parent_id != null && !isNaN(f.parent_id)) p.parent_id = f.parent_id;
  if (f.severity_from != null && !isNaN(f.severity_from)) p.severity_from = f.severity_from;
  if (f.found_version_id != null !== undefined && !isNaN(f.found_version_id)) p.found_version_id = f.found_version_id;
  if (f.fix_version_id != null && !isNaN(f.fix_version_id)) p.fix_version_id = f.fix_version_id;
  if (f.start_date_from) p.start_date_from = f.start_date_from;
  if (f.target_date_to) p.target_date_to = f.target_date_to;
  return p;
}

/**
 * 将后端 ListIssuesParams 合并回 FilterState（用于从服务端恢复时补全前端不支持的字段）。
 */
export function listParamsToFilter(p: ListIssuesParams): FilterState {
  return {
    search: p.search,
    group: p.group,
    type: p.type,
    priority: p.priority,
    assignee_id: p.assignee_id,
    label_id: p.label_id,
    module_id: p.module_id,
    sprint_id: p.sprint_id,
    parent_id: p.parent_id,
    severity_from: p.severity_from,
    found_version_id: p.found_version_id,
    fix_version_id: p.fix_version_id,
    start_date_from: p.start_date_from,
    target_date_to: p.target_date_to,
  };
}

// =====================================================================
// SortAdapter
// =====================================================================

/** 将 SortState 转为后端 sort 查询字符串，如 "-updated_at" */
export function sortToQuery(s: SortState): string {
  return s.desc ? `-${s.field}` : s.field;
}

/** 将后端 sort 查询字符串解析为 SortState */
export function queryToSort(q: string | undefined | null, defaultField = "updated_at"): SortState {
  if (!q) return { field: defaultField, desc: true };
  const desc = q.startsWith("-");
  return { field: desc ? q.slice(1) : q, desc };
}

// =====================================================================
// 判断是否"有活跃过滤"
// =====================================================================

/** 检查 FilterState 是否有任何有效过滤条件 */
export function hasActiveFilter(f: FilterState): boolean {
  return Object.values(f).some(
    (v) => v !== undefined && v !== null && v !== "" && !(typeof v === "number" && isNaN(v)),
  );
}

/** 返回活跃过滤条件数量（供 UI badge 展示） */
export function activeFilterCount(f: FilterState): number {
  return Object.values(f).filter(
    (v) => v !== undefined && v !== null && v !== "" && !(typeof v === "number" && isNaN(v)),
  ).length;
}

/**
 * 从"服务端 view_preferences.filters jsonb"恢复 FilterState。
 * 服务端存储的 filters 可能是旧格式的 Record<string, unknown>，需要安全转型。
 */
export function safeParseFilters(raw: unknown): FilterState {
  if (!raw || typeof raw !== "object") return {};
  const f: FilterState = {};
  const obj = raw as Record<string, unknown>;

  if (typeof obj.search === "string") f.search = obj.search;
  if (isValidGroup(obj.group)) f.group = obj.group;
  if (isValidType(obj.type)) f.type = obj.type;
  if (isValidPriority(obj.priority)) f.priority = obj.priority;
  if (typeof obj.assignee_id === "number") f.assignee_id = obj.assignee_id;
  if (typeof obj.label_id === "number") f.label_id = obj.label_id;
  if (typeof obj.module_id === "number") f.module_id = obj.module_id;
  if (typeof obj.sprint_id === "number") f.sprint_id = obj.sprint_id;
  if (typeof obj.parent_id === "number") f.parent_id = obj.parent_id;
  if (typeof obj.severity_from === "number") f.severity_from = obj.severity_from;
  if (typeof obj.found_version_id === "number") f.found_version_id = obj.found_version_id;
  if (typeof obj.fix_version_id === "number") f.fix_version_id = obj.fix_version_id;
  if (typeof obj.start_date_from === "string") f.start_date_from = obj.start_date_from;
  if (typeof obj.target_date_to === "string") f.target_date_to = obj.target_date_to;
  return f;
}

// =====================================================================
// 工具函数
// =====================================================================

function isValidGroup(v: unknown): v is StateGroup {
  return v === "backlog" || v === "started" || v === "completed" || v === "cancelled";
}

function isValidType(v: unknown): v is IssueTypeCode {
  return v === "requirement" || v === "task" || v === "defect" || v === "epic";
}

function isValidPriority(v: unknown): v is IssuePriority {
  return v === "urgent" || v === "high" || v === "medium" || v === "low" || v === "none";
}
