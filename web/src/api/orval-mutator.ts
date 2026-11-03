/**
 * Orval mutator — 类型安全的 HTTP 请求封装。
 *
 * 由 orval 生成器复用，所有生成的 API 方法通过此函数发起请求，
 * 从而保留 axios 实例的 interceptors（token 注入 / 错误统一处理 / 限流）。
 *
 * 说明：
 *   - 此文件专供 orval 生成流水线使用，orval 要求导出「函数声明」形式；
 *   - 运行时实现委托给 src/api/client.ts 的 http 实例，避免重复创建实例。
 */
import { http } from "./client";

/** Orval mutator 签名：与 axios 请求配置对齐。 */
export type RequestOptions = {
  url: string;
  // orval 生成器输出大写 method（"GET"/"POST"/...），axios 要求小写，
  // 在下方做归一化。
  method: "get" | "post" | "put" | "patch" | "delete" | "head" | "options" | Uppercase<string>;
  params?: Record<string, unknown>;
  data?: unknown;
  headers?: Record<string, unknown>;
  responseType?: string;
  signal?: AbortSignal;
};

/**
 * 发起 HTTP 请求。
 *
 * 统一走 client.ts 的 http 实例，确保拦截器链（Authorization 注入、
 * 401 刷新、429 限流、错误规范化）对所有生成方法生效。
 */
export async function customInstance<T>(options: RequestOptions): Promise<T> {
  const response = await http.request({
    url: options.url,
    method: (options.method as string).toLowerCase() as never,
    params: options.params,
    data: options.data,
    headers: options.headers as never,
    responseType: options.responseType as never,
    signal: options.signal,
  });
  return response.data as T;
}
