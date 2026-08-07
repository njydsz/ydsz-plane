/**
 * Axios 实例 — 请求/响应拦截器提供以下能力：
 *  1) X-Request-ID 注入（后端 prometheus 链路打通）
 *  2) 401 单飞刷新 + 原请求重放
 *  3) 429 限流感知：读取 Retry-After，触发限流回调
 *  4) 离线 / 网络错误统一识别
 *  5) 类型化错误（ApiError）+ 字段级详情
 */

import axios, {
  AxiosError,
  type AxiosRequestConfig,
  type AxiosResponse,
} from "axios";

/* ------------------------------------------------------------------ */
/* 公共类型                                                             */
/* ------------------------------------------------------------------ */

/** 后端错误 envelope（与 pkg/errs.AppError 对齐） */
export interface ApiErrorBody {
  error?: {
    code: string;
    message: string;
    details?: Array<{ field: string; reason: string }>;
    request_id?: string;
  };
}

/** 类型化 API 响应 envelope */
export interface ApiResponse<T = unknown> {
  data: T;
  requestId?: string;
}

export { AxiosResponse };
export type { AxiosRequestConfig };

/* ------------------------------------------------------------------ */
/* ApiError                                                             */
/* ------------------------------------------------------------------ */

export class ApiError extends Error {
  code: string;
  status: number;
  details?: Array<{ field: string; reason: string }>;
  requestId?: string;

  constructor(status: number, body: ApiErrorBody | undefined) {
    super(body?.error?.message ?? `请求失败 (${status})`);
    this.name = "ApiError";
    this.status = status;
    this.code = body?.error?.code ?? "UNKNOWN";
    this.details = body?.error?.details;
    this.requestId = body?.error?.request_id;
  }

  get isValidation() {
    return this.status === 422;
  }
  get isAuth() {
    return this.status === 401;
  }
  get isForbidden() {
    return this.status === 403;
  }
  get isNotFound() {
    return this.status === 404;
  }
  get isRateLimited() {
    return this.status === 429;
  }
  get isNetwork() {
    return this.status === 0;
  }

  /** 字段级错误映射到 form key，便于表单组件直接使用 */
  fieldErrors(): Record<string, string> {
    const out: Record<string, string> = {};
    for (const d of this.details ?? []) out[d.field] = d.reason;
    return out;
  }
}

/* ------------------------------------------------------------------ */
/* 限流回调                                                              */
/* ------------------------------------------------------------------ */

type RateLimitHandler = (retryAfterSec: number) => void;
let onRateLimit: RateLimitHandler | null = null;

/** 注册全局限流回调；收到 429 且带 Retry-After 时触发（可空，用于清空） */
export function setRateLimitHandler(fn: RateLimitHandler | null) {
  onRateLimit = fn;
}

/* ------------------------------------------------------------------ */
/* axios 实例                                                            */
/* ------------------------------------------------------------------ */

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 30_000,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

/* ------------------------------------------------------------------ */
/* 请求拦截器：X-Request-ID                                             */
/* ------------------------------------------------------------------ */

http.interceptors.request.use((config) => {
  // 与后端 middleware.RequestID 对齐：若上游已带 ID（SSR 调用链场景）则复用
  const existing = config.headers["X-Request-ID"];
  if (!existing) {
    const rid = generateRequestId();
    config.headers.set("X-Request-ID", rid);
    (config as RequestMeta).__requestId = rid;
  }
  return config;
});

interface RequestMeta {
  _retried?: boolean;
  __requestId?: string;
}

/** 16 字节伪随机 ID（URL safe），兼容旧浏览器。 */
function generateRequestId(): string {
  const b = new Uint8Array(16);
  if (typeof crypto !== "undefined" && crypto.getRandomValues) {
    crypto.getRandomValues(b);
  } else {
    for (let i = 0; i < b.length; i++) b[i] = Math.floor(Math.random() * 256);
  }
  return Array.from(b, (x) => x.toString(16).padStart(2, "0")).join("");
}

/* ------------------------------------------------------------------ */
/* 响应拦截器                                                            */
/* ------------------------------------------------------------------ */

let refreshing: Promise<void> | null = null;

async function refreshSession(): Promise<void> {
  await http.post("/auth/refresh", {});
}

http.interceptors.response.use(
  (res) => {
    // 后端当前直接返回业务对象；如未来切到 envelope 格式：
    //   return res.data?.data ?? res.data
    return res;
  },
  async (error: AxiosError<ApiErrorBody>) => {
    const status = error.response?.status ?? 0;
    const respHeaders = error.response?.headers ?? {};
    const config = error.config as (AxiosRequestConfig & RequestMeta) | undefined;

    // 限流感知
    if (status === 429) {
      const retryAfter = Number(respHeaders["retry-after"]);
      if (!Number.isNaN(retryAfter) && retryAfter > 0) {
        onRateLimit?.(retryAfter);
      }
    }

    // 401 且非 /auth/ 端点：单飞刷新后重放一次
    if (
      status === 401 &&
      config &&
      !config._retried &&
      !String(config.url).startsWith("/auth/")
    ) {
      config._retried = true;
      try {
        refreshing ??= refreshSession().finally(() => {
          refreshing = null;
        });
        await refreshing;
        return http.request(config);
      } catch {
        window.location.assign("/login?reason=expired");
        return Promise.reject(new ApiError(401, { error: { code: "AUTH.EXPIRED", message: "登录已过期，请重新登录" } }));
      }
    }

    // 网络错误 / CORS / 断网
    if (!error.response) {
      return Promise.reject(
        new ApiError(0, { error: { code: "NETWORK", message: "网络不可用或请求被拦截" } }),
      );
    }

    return Promise.reject(new ApiError(status, error.response.data));
  },
);
