// load-test.js — 负载测试：渐进式 VU 增长，模拟典型工作日负载。
//
// VU 曲线：
//   0 – 30s   预热（0 → 10 VU）
//   30s – 2m  爬坡（10 → 100 VU）
//   2m – 4m   稳态（100 VU）
//   4m – 5m   冷却（100 → 0 VU）
//
// 断言目标（P95）：
//   http_req_duration  P95 < 200ms
//   http_req_failed    错误率 < 1%
//
// 运行：
//   k6 run tests/perf/load-test.js
//   或：
//   k6 run -e BASE_URL=http://localhost:8080 \
//          -e TEST_USER_EMAIL=admin@ydsz.dev \
//          -e TEST_USER_PASSWORD=Admin@123 \
//          tests/perf/load-test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate } from 'k6/metrics';

// -------------------------------------------------------------------------
// 自定义指标
// -------------------------------------------------------------------------

const issueListDuration = new Trend('ydsz_issue_list_duration', true);
const issueDetailDuration = new Trend('ydsz_issue_detail_duration', true);
const wbSummaryDuration = new Trend('ydsz_wb_summary_duration', true);
const apiErrors = new Rate('ydsz_api_errors');

// -------------------------------------------------------------------------
// 配置
// -------------------------------------------------------------------------

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080/api/v1';
const EMAIL    = __ENV.TEST_USER_EMAIL || 'admin@ydsz.dev';
const PASSWORD = __ENV.TEST_USER_PASSWORD || 'Admin@123';

// -------------------------------------------------------------------------
// options：渐进 VU + 核心断言
// -------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '30s',  target: 10 },   // 预热
    { duration: '90s',  target: 100 },  // 爬坡
    { duration: '2m',   target: 100 },  // 稳态
    { duration: '1m',   target: 0 },    // 冷却
  ],
  thresholds: {
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    http_req_failed:    ['rate<0.01'],
    ydsz_api_errors:    ['rate<0.01'],
    ydsz_issue_list_duration:    ['p(95)<150'],
    ydsz_issue_detail_duration:  ['p(95)<100'],
    ydsz_wb_summary_duration:    ['p(95)<200'],
  },
};

// -------------------------------------------------------------------------
// setup()：登录获取 token
// -------------------------------------------------------------------------

export function setup() {
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } });

  if (loginRes.status !== 200) {
    console.error(`load-test setup: login failed (${loginRes.status}): ${loginRes.body}`);
    return { skip: true };
  }

  const loginBody = JSON.parse(loginRes.body);
  const token = loginBody.access_token;

  // 探测 workspace / project
  const authHeader = {
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
  };
  const wsRes = http.get(`${BASE_URL}/workspaces`, authHeader);
  const wsList = wsRes.status === 200 ? JSON.parse(wsRes.body) : [];
  const wsId = wsList[0]?.id || 1;

  const projRes = http.get(`${BASE_URL}/workspaces/${wsId}/projects`, authHeader);
  let projectId = 1;
  if (projRes.status === 200) {
    const projBody = JSON.parse(projRes.body);
    projectId = projBody.results?.[0]?.id || projBody[0]?.id || 1;
  }

  // 获取一条有效 issueId
  const issueRes = http.get(
    `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=1`,
    authHeader,
  );
  let issueId = 1;
  if (issueRes.status === 200) {
    const issues = JSON.parse(issueRes.body);
    issueId = issues.results?.[0]?.id || issues[0]?.id || 1;
  }

  return { token, wsId, projectId, issueId };
}

// -------------------------------------------------------------------------
// default()
// -------------------------------------------------------------------------

export default function (data) {
  if (data.skip) return;

  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // 60% 流量：浏览工作项列表（分页查询）
  group('Issue List', () => {
    const page = Math.floor(Math.random() * 5) + 1;
    const listRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues?limit=20&offset=${(page - 1) * 20}`,
      { headers },
    );
    issueListDuration.add(listRes.timings.duration);
    const ok = listRes.status === 200;
    if (!ok) apiErrors.add(1);
    check(listRes, { 'issue-list: 200': () => ok });
  });

  // 20% 流量：查看工作项详情
  group('Issue Detail', () => {
    const issueId = data.issueId + Math.floor(Math.random() * 100);
    const detailRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues/${issueId}`,
      { headers },
    );
    issueDetailDuration.add(detailRes.timings.duration);
    // 404 也算正常（ID 随机），只要服务端不 5xx
    const ok = detailRes.status === 200 || detailRes.status === 404;
    if (!ok) apiErrors.add(1);
    check(detailRes, { 'issue-detail: 200 or 404': () => ok });
  });

  // 10% 流量：工作台汇总
  group('Workbench Summary', () => {
    const wbRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/workbench/summary`,
      { headers },
    );
    wbSummaryDuration.add(wbRes.timings.duration);
    const ok = wbRes.status === 200;
    if (!ok) apiErrors.add(1);
    check(wbRes, { 'wb-summary: 200': () => ok });
  });

  // 10% 流量：工作空间列表
  group('Workspace List', () => {
    const wsRes = http.get(`${BASE_URL}/workspaces`, { headers });
    const ok = wsRes.status === 200;
    if (!ok) apiErrors.add(1);
    check(wsRes, { 'workspaces: 200': () => ok });
  });

  sleep(Math.random() * 0.5 + 0.5); // 0.5 – 1.0s 思考时间
}
