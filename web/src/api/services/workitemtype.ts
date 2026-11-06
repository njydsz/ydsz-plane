/**
 * 自定义工作项类型 API — 封装 WorkItemType / Property / CustomFieldValue 的 REST 接口。
 *
 * 路由: /api/v1/workspaces/:ws/projects/:pid/work-item-types
 *
 * 说明：
 *  - 类型（WorkItemType）为项目级，code 在同一项目内唯一；
 *  - 类型可归档（archived_at），归档后不再出现在列表；
 *  - 属性（WorkItemTypeProperty）的 code 在同一类型内唯一；
 *  - 属性值（WorkItemCustomFieldValue）采用 EAV 模型，value 统一为 TEXT。
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 自定义属性字段类型枚举 */
export type FieldType =
  | "text"
  | "number"
  | "select"
  | "multiselect"
  | "date"
  | "checkbox"
  | "url";

/** 工作项类型层级：0=工作项级，1=Epic 级 */
export type WorkItemTypeLevel = 0 | 1;

/** 自定义工作项类型（项目级） */
export interface WorkItemType {
  id: number;
  project_id: number;
  name: string;
  code: string;
  color: string;
  icon: string;
  description: string;
  level: WorkItemTypeLevel;
  is_active: boolean;
  is_default: boolean;
  created_at: string;
  updated_at: string;
  archived_at?: string | null;
}

/** 类型的自定义属性定义 */
export interface WorkItemTypeProperty {
  id: number;
  type_id: number;
  name: string;
  code: string;
  field_type: FieldType;
  /** select / multiselect 的候选项，后端以 JSON 字符串存储 */
  options: string;
  required: boolean;
  sort_order: number;
  created_at: string;
}

/** 工作项的自定义属性值（EAV） */
export interface WorkItemCustomFieldValue {
  id: number;
  property_id: number;
  work_item_id: number;
  work_item_type: string;
  value: string;
  created_at: string;
  updated_at: string;
}

/** 创建工作项类型入参 */
export interface CreateWorkItemTypeInput {
  name: string;
  code: string;
  color?: string;
  icon?: string;
  description?: string;
  level?: WorkItemTypeLevel;
  is_default?: boolean;
}

/** 更新工作项类型入参 */
export interface UpdateWorkItemTypeInput {
  name?: string;
  code?: string;
  color?: string;
  icon?: string;
  description?: string;
  level?: WorkItemTypeLevel;
  is_active?: boolean;
  is_default?: boolean;
}

/** 新增自定义属性入参 */
export interface CreatePropertyInput {
  name: string;
  code: string;
  field_type: FieldType;
  /** select / multiselect 的候选项（JSON 数组字符串） */
  options?: string;
  required?: boolean;
  sort_order?: number;
}

/** 写入自定义属性值入参 */
export interface SetCustomFieldValueInput {
  property_id: number;
  work_item_id: number;
  work_item_type: string;
  value: string;
}

/** 列表响应（后端返回 { items: [...] }） */
export interface WorkItemTypeListResult {
  items: WorkItemType[];
}

/** 属性列表响应 */
export interface WorkItemTypePropertyListResult {
  items: WorkItemTypeProperty[];
}

/** 自定义属性值列表响应 */
export interface WorkItemCustomFieldValueListResult {
  items: WorkItemCustomFieldValue[];
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** 自定义工作项类型域 API */
export const workItemTypeApi = {
  /* --- 类型 CRUD --- */

  /** 列出项目下全部工作项类型 */
  list: (workspaceId: number | string, projectId: number | string) =>
    wrap<WorkItemTypeListResult>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types`,
      ),
    ),

  /** 获取单个工作项类型详情 */
  get: (workspaceId: number | string, projectId: number | string, id: number | string) =>
    wrap<WorkItemType>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${id}`,
      ),
    ),

  /** 创建工作项类型 */
  create: (
    workspaceId: number | string,
    projectId: number | string,
    data: CreateWorkItemTypeInput,
  ) =>
    wrap<WorkItemType>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types`,
        data,
      ),
    ),

  /** 更新工作项类型 */
  update: (
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
    data: UpdateWorkItemTypeInput,
  ) =>
    wrap<WorkItemType>(
      http.patch(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${id}`,
        data,
      ),
    ),

  /** 删除工作项类型 */
  remove: (
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${id}`,
      ),
    ),

  /** 归档工作项类型 */
  archive: (
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
  ) =>
    wrap<WorkItemType>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${id}/archive`,
      ),
    ),

  /** 恢复已归档的工作项类型 */
  restore: (
    workspaceId: number | string,
    projectId: number | string,
    id: number | string,
  ) =>
    wrap<WorkItemType>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${id}/restore`,
      ),
    ),

  /* --- 自定义属性 --- */

  /** 列出某类型的全部自定义属性 */
  listProperties: (
    workspaceId: number | string,
    projectId: number | string,
    typeId: number | string,
  ) =>
    wrap<WorkItemTypePropertyListResult>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${typeId}/properties`,
      ),
    ),

  /** 为类型新增自定义属性 */
  addProperty: (
    workspaceId: number | string,
    projectId: number | string,
    typeId: number | string,
    data: CreatePropertyInput,
  ) =>
    wrap<WorkItemTypeProperty>(
      http.post(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${typeId}/properties`,
        data,
      ),
    ),

  /** 删除类型下的自定义属性 */
  removeProperty: (
    workspaceId: number | string,
    projectId: number | string,
    typeId: number | string,
    propertyId: number | string,
  ) =>
    wrap<void>(
      http.delete(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${typeId}/properties/${propertyId}`,
      ),
    ),

  /* --- 自定义属性值 --- */

  /** 获取某工作项在某类型下的全部自定义属性值 */
  getValues: (
    workspaceId: number | string,
    projectId: number | string,
    typeId: number | string,
    workItemId: number | string,
  ) =>
    wrap<WorkItemCustomFieldValueListResult>(
      http.get(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${typeId}/values/${workItemId}`,
      ),
    ),

  /** 写入工作项自定义属性值 */
  setValue: (
    workspaceId: number | string,
    projectId: number | string,
    typeId: number | string,
    data: SetCustomFieldValueInput,
  ) =>
    wrap<WorkItemCustomFieldValue>(
      http.put(
        `/workspaces/${workspaceId}/projects/${projectId}/work-item-types/${typeId}/values`,
        data,
      ),
    ),
};
