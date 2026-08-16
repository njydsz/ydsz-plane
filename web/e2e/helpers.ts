/**
 * E2E 共享工具：统一登录与会话头构造。
 *
 * 后端对「携带会话 Cookie 的状态变更请求」启用双提交 Cookie CSRF 校验
 * （见 internal/interfaces/middleware/csrf.go）。登录响应通过 Set-Cookie
 * 下发非 HttpOnly 的 X-CSRF-TOKEN；Playwright 的 request context 会自动
 * 持久化该 Cookie，导致后续 PUT/PATCH/POST/DELETE 触发 CSRF 校验。
 *
 * 真实浏览器 SPA 会读取该 Cookie 并以 X-CSRF-Token 头回传；这里在测试中
 * 登录后从存储中读取同一令牌并注入请求头，模拟真实浏览器的 CSRF 双提交行为。
 */
import { APIRequestContext } from "@playwright/test";

export const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
export const apiURL = API_URL;
export const TEST_EMAIL = "admin@njydsz.com";
export const TEST_PASSWORD = "Admin@1020";

export interface AuthSession {
  token: string;
  headers: Record<string, string>;
}

export async function apiLogin(request: APIRequestContext): Promise<AuthSession> {
  const res = await request.post(`${API_URL}/auth/login`, {
    data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  });
  if (!res.ok()) {
    throw new Error(`login failed: ${res.status()}`);
  }
  const { access_token } = await res.json();
  const state = await request.storageState();
  const csrf = state.cookies.find((c) => c.name === "X-CSRF-TOKEN")?.value;
  const headers: Record<string, string> = {
    Authorization: `Bearer ${access_token}`,
  };
  if (csrf) headers["X-CSRF-Token"] = csrf;
  return { token: access_token as string, headers };
}
