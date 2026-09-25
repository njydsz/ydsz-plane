/**
 * ApiError — 类型化 API 错误。
 *
 * 与后端 pkg/errs.AppError 错误信封对齐：
 *   HTTP 错误响应体: {"error": {"code": "DOMAIN.SNAKE", "message": "中文提示", "details": [{"field":"f","reason":"r"}], "request_id": "..."}}
 *
 * 调用方可通过 `e instanceof ApiError && e.code === "CODE"` 区分错误类型；
 * 表单校验类错误可通过 `e.fields` 直接获取字段级错误映射（field → message）。
 */

/** 后端错误信封中的原始 details 条目（参数校验场景） */
export interface ApiErrorDetailItem {
  field: string;
  reason: string;
}

/** 后端错误信封直接结构（未经 interceptor 解包） */
export interface ApiEnvelope {
  error?: {
    code: string;
    message: string;
    details?: ApiErrorDetailItem[];
    request_id?: string;
  };
}

/** ApiError 构造参数 */
export interface ApiErrorInit {
  code: string;
  message: string;
  /** 字段级错误映射（由 details 数组转换而来，便于表单直接使用） */
  fields?: Record<string, string>;
  /** 原始 details 透传（保留原始 field/reason 结构） */
  details?: unknown;
}

export class ApiError extends Error {
  code: string;
  fields: Record<string, string>;
  details: unknown;
  statusCode: number;

  constructor(resp: ApiErrorInit, statusCode: number) {
    super(resp.message ?? `请求失败 (${statusCode})`);
    this.name = "ApiError";
    this.code = resp.code ?? "UNKNOWN";
    this.fields = resp.fields ?? {};
    this.details = resp.details;
    this.statusCode = statusCode;
  }

  get isValidation() {
    return this.statusCode === 422;
  }
  get isAuth() {
    return this.statusCode === 401;
  }
  get isForbidden() {
    return this.statusCode === 403;
  }
  get isNotFound() {
    return this.statusCode === 404;
  }
  get isRateLimited() {
    return this.statusCode === 429;
  }
  get isNetwork() {
    return this.statusCode === 0;
  }
}

/**
 * 将后端错误信封转换为 ApiErrorInit。
 *
 * 将 details 数组同时转换为 fields Record（便于表单使用）和保留原始 details。
 */
export function envelopeToInit(env: ApiEnvelope | undefined): ApiErrorInit {
  const err = env?.error;
  if (!err) {
    return { code: "UNKNOWN", message: "未知错误" };
  }
  const fields: Record<string, string> = {};
  if (Array.isArray(err.details)) {
    for (const d of err.details) {
      if (d?.field) fields[d.field] = d.reason ?? "无效";
    }
  }
  return {
    code: err.code ?? "UNKNOWN",
    message: err.message ?? "请求失败",
    fields,
    details: err.details,
  };
}

/**
 * 将 HTTP 响应错误体直接构造为 ApiError。
 * 由 axios 响应拦截器调用，统一拦截路径。
 */
export function makeApiError(
  status: number,
  body: ApiEnvelope | undefined,
): ApiError {
  return new ApiError(envelopeToInit(body), status);
}
