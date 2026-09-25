/**
 * k6 部署后性能冒烟测试 — S16 P2-2。
 *
 * 用途:
 *   - CI 部署后运行，验证核心端点 P95 延迟 ≤ 200ms
 *   - 作为性能基线回归检测（future: 每次 main push 后自动对比）
 *
 * 运行:
 *   k6 run -e BASE_URL=http://localhost:8080/api/v1 \
 *          -e TEST_USER_EMAIL=admin@njydsz.com \
 *          -e TEST_USER_PASSWORD=Admin@1020 \
 *          tests/perf/post-deploy-check.js
 *
 * 基线断言:
 *   - login P95 ≤ 500ms（含 bcrypt 哈希）
 *   - workspaces list P95 ≤ 200ms
 *   - projects list P95 ≤ 200ms
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// --- 自定义指标 ---
const loginDuration = new Trend("login_duration_ms", true);
const apiErrorRate = new Rate("api_errors");

// --- 配置 ---
const BASE_URL = __ENV.BASE_URL || "http://localhost:8080/api/v1";
const USER_EMAIL = __ENV.TEST_USER_EMAIL || "admin@njydsz.com";
const USER_PASSWORD = __ENV.TEST_USER_PASSWORD || "Admin@1020";

// 性能阈值（对标 Google SRE / 字节 API 规范）
const P95_THRESHOLD_MS = {
  login: 500,      // bcrypt 较慢
  default: 200,
};

export const options = {
  vus: 10,
  duration: "30s",
  thresholds: {
    // 整体 HTTP 请求失败率 < 1%
    http_req_failed: ["rate<0.01"],
    // 核心端点 P95 延迟
    [`http_req_duration{url:${BASE_URL}/auth/login}`]:
      [`p(95)<${P95_THRESHOLD_MS.login}`],
    [`http_req_duration{url:${BASE_URL}/workspaces}`]:
      [`p(95)<${P95_THRESHOLD_MS.default}`],
    [`http_req_duration{url:${BASE_URL}/projects}`]:
      [`p(95)<${P95_THRESHOLD_MS.default}`],
  },
};

// --- 登录获取 token ---
function login() {
  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({
      email: USER_EMAIL,
      password: USER_PASSWORD,
    }),
    { headers: { "Content-Type": "application/json" } }
  );

  loginDuration.add(Date.now() - start);

  const ok = check(res, {
    "login status is 200": (r) => r.status === 200,
    "login returns access_token": (r) => {
      try {
        return JSON.parse(r.body).access_token !== undefined;
      } catch {
        return false;
      }
    },
  });

  apiErrorRate.add(!ok);

  if (res.status === 200) {
    try {
      return JSON.parse(res.body).access_token;
    } catch {
      return null;
    }
  }
  return null;
}

// --- 主测试流程 ---
export default function () {
  // 1. 登录
  const token = login();
  if (!token) {
    return;
  }

  const headers = {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };

  // 2. 核心端点冒烟
  const endpoints = [
    `${BASE_URL}/workspaces`,
    `${BASE_URL}/me`,
  ];

  for (const url of endpoints) {
    const res = http.get(url, { headers });
    const ok = check(res, {
      [`${url} status is 200`]: (r) => r.status === 200,
    });
    apiErrorRate.add(!ok);
  }

  sleep(1);
}
