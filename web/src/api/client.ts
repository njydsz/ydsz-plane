/**
 * Axios 实例：统一 baseURL、错误壳解析、401 自动刷新重放（单飞）。
 * 类型在 S2 起由 openapi-typescript 生成后替换手写类型。
 */
import axios, { AxiosError, type AxiosRequestConfig } from "axios";

export interface ApiErrorBody {
  error?: {
    code: string;
    message: string;
    details?: Array<{ field: string; reason: string }>;
    request_id?: string;
  };
}

export class ApiError extends Error {
  code: string;
  status: number;
  details?: Array<{ field: string; reason: string }>;

  constructor(status: number, body: ApiErrorBody | undefined) {
    super(body?.error?.message ?? `请求失败 (${status})`);
    this.status = status;
    this.code = body?.error?.code ?? "UNKNOWN";
    this.details = body?.error?.details;
  }
}

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 30_000,
  withCredentials: true, // cookie 会话
  headers: { "Content-Type": "application/json" },
});

let refreshing: Promise<void> | null = null;

async function refreshSession(): Promise<void> {
  await http.post("/auth/refresh", {});
}

http.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<ApiErrorBody>) => {
    const status = error.response?.status ?? 0;
    const config = error.config as (AxiosRequestConfig & { _retried?: boolean }) | undefined;

    // 401 且非认证端点：单飞刷新后重放一次
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
        window.location.assign("/login");
      }
    }
    return Promise.reject(new ApiError(status, error.response?.data));
  },
);
