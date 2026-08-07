// stress-test.js — 压力测试：恒定高并发，找出系统拐点。
//
// 策略：
//   恒定 200 VU 持续 3 分钟；
//   通过输出观察错误率飙升拐点，判断系统容量上限。
//
// 不设严格阈值（因为是探索性测试），仅输出统计。
//
// 运行：
//   k6 run tests/perf/stress-test.js
//   或：
//   k6 run -e BASE_URL=http://localhost:8080 \
//          -e TEST_USER_EMAIL=admin@ydsz.dev \
//          -e TEST_USER_PASSWORD=Admin@123 \
//          -e STRESS_VUS=200 \
//          -e STRESS_DURATION=3m \
//          tests/perf/stress-test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// -------------------------------------------------------------------------
// 自定义指标
// -------------------------------------------------------------------------

const issueListDuration = new Trend('ydsz_issue_list_duration', true);
const issueDetailDuration = new Trend('ydsz_issue_detail_duration', true);
const apiErrors = new Rate('ydsz_api_errors');
const serverErrors = new Counter('ydsz_5xx_errors');
const totalReqs   = new Counter('ydsz_total_requests');

// -------------------------------------------------------------------------
// 配置
// -------------------------------------------------------------------------

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080/api/v1';
const EMAIL    = __ENV.TEST_USER_EMAIL || 'admin@ydsz.dev';
const PASSWORD = __ENV.TEST_USER_PASSWORD || 'Admin@123';
const VUS      = Number(__ENV.STRESS_VUS) || 200;
const DURATION = __ENV.STRESS_DURATION || '3m';

// -------------------------------------------------------------------------
// options
// -------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '30s',  target: VUS },
    { duration: DURATION, target: VUS },
    { duration: '30s',  target: 0 },
  ],
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(75)', 'p(90)', 'p(95)', 'p(99)'],
};

// -------------------------------------------------------------------------
// setup()
// -------------------------------------------------------------------------

export function setup() {
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } });

  if (loginRes.status !== 200) {
    console.error(`login failed: ${loginRes.status}`);
    return { skip: true };
  }

  const body = JSON.parse(loginRes.body);
  const token = body.access_token;
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

  const issueRes = http.get(
    `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=1`,
    authHeader,
  );
  let issueId = 1;
  if (issueRes.status === 200) {
    const issues = JSON.parse(issueRes.body);
    issueId = issues.results?.[0]?.id || issues[0]?.id || 1;
  }

  console.log(`Stress test setup: wsId=${wsId}, projectId=${projectId}, issueId=${issueId}, VUS=${VUS}, duration=${DURATION}`);

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

  // 工作项列表查询（50%）
  group('Issue List', () => {
    const listRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues?limit=20`,
      { headers, tags: { name: 'issue_list' } },
    );
    totalReqs.add(1);
    issueListDuration.add(listRes.timings.duration);

    const ok = listRes.status === 200;
    if (!ok) apiErrors.add(1);
    if (listRes.status >= 500) serverErrors.add(1);
    check(listRes, { 'issue-list: 200': () => ok });
  });

  // 工作项详情（30%）
  group('Issue Detail', () => {
    const issueId = data.issueId + Math.floor(Math.random() * 500);
    const detailRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/projects/${data.projectId}/issues/${issueId}`,
      { headers, tags: { name: 'issue_detail' } },
    );
    totalReqs.add(1);
    issueDetailDuration.add(detailRes.timings.duration);

    const ok = detailRes.status === 200 || detailRes.status === 404;
    if (!ok) apiErrors.add(1);
    if (detailRes.status >= 500) serverErrors.add(1);
    check(detailRes, { 'issue-detail: ok': () => ok });
  });

  // 工作台汇总（20%）
  group('Workbench Summary', () => {
    const wbRes = http.get(
      `${BASE_URL}/workspaces/${data.wsId}/workbench/summary`,
      { headers, tags: { name: 'wb_summary' } },
    );
    totalReqs.add(1);

    const ok = wbRes.status === 200;
    if (!ok) apiErrors.add(1);
    if (wbRes.status >= 500) serverErrors.add(1);
    check(wbRes, { 'wb-summary: 200': () => ok });
  });

  sleep(Math.random() * 0.3 + 0.1);
}

// -------------------------------------------------------------------------
// teardown()
// -------------------------------------------------------------------------

export function teardown(data) {
  if (data.skip) return;
  // 自定义指标由 k6 在 exit summary 中自动输出，此处无需重复打印。
  console.log('Stress test complete. Check stdout for ydsz_* metrics.');
}
