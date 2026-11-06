/**
 * PQL (Plane Query Language) API。
 *
 * 路由: /api/v1/workspaces/:ws/pql/*
 */
import { http } from "../client";

/* ------------------------------------------------------------------ */
/* Types                                                              */
/* ------------------------------------------------------------------ */

/** 字段元数据 */
export interface PQLField {
  name: string;
  column: string;
  type: "text" | "number" | "date" | "enum" | "user";
  values?: string[];
}

/** PQL 解析/编译结果 */
export interface PQLParseResult {
  valid: boolean;
  where_sql?: string;
  args?: unknown[];
  sort_column?: string;
  sort_dir?: string;
  limit: number;
  offset: number;
  error?: string;
}

/* ------------------------------------------------------------------ */
/* Helpers                                                            */
/* ------------------------------------------------------------------ */

const wrap = <T>(p: Promise<{ data: T }>) => p.then((r) => r.data);

/* ------------------------------------------------------------------ */
/* API                                                                */
/* ------------------------------------------------------------------ */

/** PQL API */
export const pqlApi = {
  /** 解析查询（含编译后的 SQL，用于调试） */
  parse: (workspaceId: number | string, query: string) =>
    wrap<PQLParseResult>(http.post(`/workspaces/${workspaceId}/pql/parse`, { query })),

  /** 校验查询 */
  validate: (workspaceId: number | string, query: string) =>
    wrap<PQLParseResult>(
      http.post(`/workspaces/${workspaceId}/pql/validate`, { query }),
    ),

  /** 解释查询 */
  explain: (workspaceId: number | string, query: string) =>
    wrap<PQLParseResult>(
      http.post(`/workspaces/${workspaceId}/pql/explain`, { query }),
    ),

  /** 可用字段列表 */
  listFields: (workspaceId: number | string) =>
    wrap<{ fields: PQLField[]; sort_fields: Record<string, boolean> }>(
      http.get(`/workspaces/${workspaceId}/pql/fields`),
    ),
};
