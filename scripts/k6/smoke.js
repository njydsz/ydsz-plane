import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

// --- 自定义指标 ---
const apiDuration = new Trend('ydsz_api_duration', true);
const apiErrors = new Rate('ydsz_api_errors');
const issueCreated = new Counter('ydsz_issue_created');

// --- 配置 ---
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // 预热：10 VU
    { duration: '1m',  target: 50 },   // 爬坡：50 VU
    { duration: '2m',  target: 100 },  // 峰值：100 VU
    { duration: '30s', target: 0 },    // 冷却
  ],
  thresholds: {
    'ydsz_api_duration': ['p(95)<200', 'p(99)<500'],
    'ydsz_api_errors':   ['rate<0.01'],
    'http_req_duration': ['p(95)<500'],
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080/api/v1';
const EMAIL = __ENV.TEST_EMAIL || 'admin@ydsz.dev';
const PASSWORD = __ENV.TEST_PASSWORD || 'Admin@123';

let accessToken = '';
let wsId = 0;
let projectId = 0;

export function setup() {
  // 1. 登录获取 token
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  }), { headers: { 'Content-Type': 'application/json' } });

  check(loginRes, { 'login: 200': (r) => r.status === 200 });
  const loginBody = JSON.parse(loginRes.body);
  accessToken = loginBody.access_token;
  const authHeader = { headers: { 'Authorization': `Bearer ${accessToken}` } };

  // 2. 获取工作空间
  const wsRes = http.get(`${BASE_URL}/workspaces`, authHeader);
  const wsBody = JSON.parse(wsRes.body);
  wsId = wsBody[0]?.id || 1;

  // 3. 获取项目
  const projRes = http.get(`${BASE_URL}/workspaces/${wsId}/projects`, authHeader);
  const projBody = JSON.parse(projRes.body);
  projectId = projBody.results?.[0]?.id || projBody[0]?.id || 1;

  console.log(`Setup complete: wsId=${wsId}, projectId=${projectId}`);
  return { accessToken, wsId, projectId };
}

export default function (data) {
  const headers = {
    'Authorization': `Bearer ${data.accessToken}`,
    'Content-Type': 'application/json',
  };
  const { wsId, projectId } = data;

  group('Health & Me', () => {
    // 健康检查
    const health = http.get(`${BASE_URL.replace('/api/v1', '')}/healthz`);
    check(health, { 'healthz: 200': (r) => r.status === 200 });

    // 当前用户
    const me = http.get(`${BASE_URL}/me`, { headers });
    check(me, { 'me: 200': (r) => r.status === 200 });
  });

  group('Issue 列表查询', () => {
    const listRes = http.get(
      `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=20`,
      { headers }
    );
    apiDuration.add(listRes.timings.duration);
    const ok = listRes.status === 200;
    if (!ok) apiErrors.add(1);
    check(listRes, { 'issue-list: 200': () => ok });
  });

  group('Issue 创建与详情', () => {
    // 创建工作项
    const createRes = http.post(
      `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
      JSON.stringify({
        type: 'task',
        name: `[k6] 压测任务 ${Date.now()}`,
        priority: 'medium',
      }),
      { headers }
    );
    apiDuration.add(createRes.timings.duration);
    const created = createRes.status === 201;
    if (created) {
      issueCreated.add(1);
      const issue = JSON.parse(createRes.body);
      // 查询详情
      const detailRes = http.get(
        `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
        { headers }
      );
      apiDuration.add(detailRes.timings.duration);
      check(detailRes, { 'issue-detail: 200': (r) => r.status === 200 });
    } else {
      apiErrors.add(1);
    }
    check(createRes, { 'issue-create: 201': () => created });
  });

  group('Sprint 查询', () => {
    const sprintRes = http.get(
      `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/sprints?limit=10`,
      { headers }
    );
    apiDuration.add(sprintRes.timings.duration);
    if (sprintRes.status !== 200) apiErrors.add(1);
    check(sprintRes, { 'sprint-list: 200': (r) => r.status === 200 });
  });

  group('Version 查询', () => {
    const verRes = http.get(
      `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/versions?limit=10`,
      { headers }
    );
    apiDuration.add(verRes.timings.duration);
    if (verRes.status !== 200) apiErrors.add(1);
    check(verRes, { 'version-list: 200': (r) => r.status === 200 });
  });

  group('States & 看板数据', () => {
    const statesRes = http.get(
      `${BASE_URL}/workspaces/${wsId}/projects/${projectId}/states`,
      { headers }
    );
    apiDuration.add(statesRes.timings.duration);
    check(statesRes, { 'states: 200': (r) => r.status === 200 });
  });

  sleep(1);
}

export function teardown(data) {
  console.log(`Test complete. Issues created: ${issueCreated?.value || 0}`);
}
