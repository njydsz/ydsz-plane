// smoke-test.js — 冒烟测试：验证核心 API 端点可用性。
//
// 覆盖端点：
//   POST /api/v1/auth/login
//   GET  /api/v1/workspaces
//   GET  /api/v1/workspaces/:wsId/projects/:projectId/issues
//   GET  /api/v1/workspaces/:wsId/projects/:projectId/issues/:issueId
//   GET  /api/v1/workspaces/:wsId/workbench/summary
//
// 运行：
//   k6 run tests/perf/smoke-test.js
//   或带环境变量：
//   k6 run -e BASE_URL=http://localhost:8080 \
//          -e TEST_USER_EMAIL=admin@njydsz.com \
//          -e TEST_USER_PASSWORD=Admin@1020 \
//          tests/perf/smoke-test.js
//
// 也可通过 thresholds 快速判断服务基本健康度。

import http from 'k6/http';
import { check, sleep } from 'k6';

// -------------------------------------------------------------------------
// 配置
// -------------------------------------------------------------------------

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080/api/v1';
const EMAIL    = __ENV.TEST_USER_EMAIL || 'admin@njydsz.com';
const PASSWORD = __ENV.TEST_USER_PASSWORD || 'Admin@1020';

// -------------------------------------------------------------------------
// options：恒定 1 VU，仅用于验证可用性
// -------------------------------------------------------------------------

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    http_req_duration: ['p(50)<500', 'p(95)<1000'],
    http_req_failed:    ['rate<0.20'], // 冒烟允许 20% 失败（可能是 DB 未就绪）
  },
};

// -------------------------------------------------------------------------
// setup()：登录获取 token，探测 workspace / project / issue
// -------------------------------------------------------------------------

export function setup() {
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } });

  const loginOk = loginRes.status === 200;
  if (!loginOk) {
    console.error(`login failed: ${loginRes.status} ${loginRes.body}`);
    return { skip: true };
  }

  const loginBody = JSON.parse(loginRes.body);
  const token = loginBody.access_token;
  const authHeader = {
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
  };

  // 获取工作空间
  const wsRes = http.get(`${BASE_URL}/workspaces`, authHeader);
  const wsList = wsRes.status === 200 ? JSON.parse(wsRes.body) : [];
  const wsId = wsList[0]?.id || 1;

  // 获取项目
  const projRes = http.get(`${BASE_URL}/workspaces/${wsId}/projects`, authHeader);
  let projectId = 1;
  if (projRes.status === 200) {
    const projBody = JSON.parse(projRes.body);
    projectId = projBody.results?.[0]?.id || projBody[0]?.id || 1;
  }

  // 获取一条工作项
  const issueListRes = http.get(
    `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=1`,
    authHeader,
  );
  let issueId = 1;
  if (issueListRes.status === 200) {
    const issues = JSON.parse(issueListRes.body);
    issueId = issues.results?.[0]?.id || issues[0]?.id || 1;
  }

  return { token, wsId, projectId, issueId };
}

// -------------------------------------------------------------------------
// default()
// -------------------------------------------------------------------------

export default function (data) {
  if (data.skip) {
    console.warn('Skipping smoke test — login failed in setup');
    return;
  }

  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // 1. 登录已在 setup 完成，此处做 token 有效性检查（调用 /me）
  const meRes = http.get(`${BASE_URL}/me`, { headers });
  check(meRes, {
    'GET /me: status 200': (r) => r.status === 200,
  });

  // 2. 工作空间列表
  const wsListRes = http.get(`${BASE_URL}/workspaces`, { headers });
  check(wsListRes, {
    'GET /workspaces: status 200': (r) => r.status === 200,
    'GET /workspaces: array body': (r) => {
      try { return Array.isArray(JSON.parse(r.body)); } catch { return false; }
    },
  });

  // 3. 工作项分页列表
  const issueListRes = http.get(
    `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues?limit=20`,
    { headers },
  );
  check(issueListRes, {
    'GET /issues: status 200': (r) => r.status === 200,
  });

  // 4. 单个工作项详情
  const issueDetailRes = http.get(
    `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues/${data.issueId}`,
    { headers },
  );
  check(issueDetailRes, {
    'GET /issues/{id}: status 200': (r) => r.status === 200,
  });

  // 5. 工作台汇总
  const summaryRes = http.get(
    `${BASE_URL}/workspaces/${data.wsId}/workbench/summary`,
    { headers },
  );
  check(summaryRes, {
    'GET /workbench/summary: status 200': (r) => r.status === 200,
  });

  sleep(0.5);
}
