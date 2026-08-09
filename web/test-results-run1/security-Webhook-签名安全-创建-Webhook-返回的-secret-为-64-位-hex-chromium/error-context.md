# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: security.spec.ts >> Webhook 签名安全 >> 创建 Webhook 返回的 secret 为 64 位 hex
- Location: e2e\security.spec.ts:79:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
  1   | /**
  2   |  * 安全场景 E2E 测试。
  3   |  *
  4   |  * 覆盖：Webhook 签名伪造检测、重放攻击防护、输入校验（SQL 注入/XSS）、
  5   |  *       权限边界（跨空间访问隔离）、API Token 鉴权。
  6   |  * 对标 S10 安全测试项 + DevOps 纵深防御。
  7   |  *
  8   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  9   |  */
  10  | import { expect, test } from "@playwright/test";
  11  | 
  12  | const TEST_EMAIL = "admin@ydsz.dev";
  13  | const TEST_PASSWORD = "Admin@123";
  14  | const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  15  | 
  16  | async function login(request: any) {
  17  |   const res = await request.post(`${API_URL}/auth/login`, {
  18  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  19  |   });
  20  |   expect(res.ok()).toBe(true);
  21  |   const body = await res.json();
  22  |   return body.access_token;
  23  | }
  24  | 
  25  | async function getFirstWorkspace(request: any, token: string) {
  26  |   const res = await request.get(`${API_URL}/workspaces`, {
  27  |     headers: { Authorization: `Bearer ${token}` },
  28  |   });
  29  |   expect(res.ok()).toBe(true);
  30  |   const list = await res.json();
  31  |   return list[0]?.id || 1;
  32  | }
  33  | 
  34  | async function getFirstProject(request: any, token: string, wsId: number) {
  35  |   const res = await request.get(`${API_URL}/workspaces/${wsId}/projects`, {
  36  |     headers: { Authorization: `Bearer ${token}` },
  37  |   });
  38  |   expect(res.ok()).toBe(true);
  39  |   const body = await res.json();
  40  |   return body.results?.[0]?.id || body[0]?.id || 1;
  41  | }
  42  | 
  43  | test.describe("鉴权与令牌安全", () => {
  44  |   test("无效 JWT token 受保护端点返回 401", async ({ request }) => {
  45  |     const res = await request.get(`${API_URL}/workspaces`, {
  46  |       headers: { Authorization: "Bearer invalid.token.here" },
  47  |     });
  48  |     expect(res.status()).toBe(401);
  49  |   });
  50  | 
  51  |   test("过期/篡改的 refresh_token 刷新失败", async ({ request }) => {
  52  |     const res = await request.post(`${API_URL}/auth/refresh`, {
  53  |       data: { refresh_token: "tampered-refresh-token-xyz" },
  54  |     });
  55  |     expect([400, 401, 422]).toContain(res.status());
  56  |   });
  57  | 
  58  |   test("缺少 Authorization 头时受保护端点返回 401", async ({ request }) => {
  59  |     const res = await request.get(`${API_URL}/workspaces`);
  60  |     expect(res.status()).toBe(401);
  61  |   });
  62  | 
  63  |   test("登录尝试次数限制（暴力破解防护）", async ({ request }) => {
  64  |     // 连续发送 5 次错误密码登录
  65  |     const attempts = [];
  66  |     for (let i = 0; i < 5; i++) {
  67  |       const res = await request.post(`${API_URL}/auth/login`, {
  68  |         data: { email: "nonexistent@example.com", password: `wrong${i}` },
  69  |       });
  70  |       attempts.push(res.status());
  71  |     }
  72  |     // 所有尝试都应返回 401（或者后几次可能返回 429 限流）
  73  |     const allRejected = attempts.every((s) => s === 401 || s === 429);
  74  |     expect(allRejected).toBe(true);
  75  |   });
  76  | });
  77  | 
  78  | test.describe("Webhook 签名安全", () => {
  79  |   test("创建 Webhook 返回的 secret 为 64 位 hex", async ({ request }) => {
  80  |     const token = await login(request);
  81  |     const wsId = await getFirstWorkspace(request, token);
  82  | 
  83  |     const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
  84  |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  85  |       data: {
  86  |         name: "Security Test Webhook",
  87  |         url: "https://localhost:19999/no-op",
  88  |         events: ["issue.created", "issue.updated"],
  89  |       },
  90  |     });
> 91  |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  92  |     const webhook = await createRes.json();
  93  |     // secret 应为 64 位 hex 字符串（HMAC-SHA256 key）
  94  |     expect(webhook.secret).toMatch(/^[0-9a-f]{64}$/);
  95  | 
  96  |     // 清理
  97  |     await request.delete(`${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}`, {
  98  |       headers: { Authorization: `Bearer ${token}` },
  99  |     });
  100 |   });
  101 | 
  102 |   test("非法事件类型被拒", async ({ request }) => {
  103 |     const token = await login(request);
  104 |     const wsId = await getFirstWorkspace(request, token);
  105 | 
  106 |     const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
  107 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  108 |       data: {
  109 |         name: "Invalid Event Webhook",
  110 |         url: "https://example.com/hook",
  111 |         events: ["issue.hacked", "nonexistent.event"],
  112 |       },
  113 |     });
  114 |     // 非法事件应被拒
  115 |     expect([400, 422]).toContain(createRes.status());
  116 |   });
  117 | 
  118 |   test("Test 推送日志中包含签名头", async ({ request }) => {
  119 |     const token = await login(request);
  120 |     const wsId = await getFirstWorkspace(request, token);
  121 | 
  122 |     const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
  123 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  124 |       data: {
  125 |         name: "Signature Test",
  126 |         url: "https://localhost:19999/no-op",
  127 |         events: ["issue.created"],
  128 |       },
  129 |     });
  130 |     expect(createRes.status()).toBe(201);
  131 |     const webhook = await createRes.json();
  132 | 
  133 |     // 触发 test push
  134 |     const testRes = await request.post(
  135 |       `${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}/test`,
  136 |       { headers: { Authorization: `Bearer ${token}` } },
  137 |     );
  138 |     // test push 目标不可达也应返回 2xx（测试请求已发送）
  139 |     expect([200, 202]).toContain(testRes.status());
  140 | 
  141 |     // 检查日志包含签名头
  142 |     const logsRes = await request.get(
  143 |       `${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}/logs?limit=5`,
  144 |       { headers: { Authorization: `Bearer ${token}` } },
  145 |     );
  146 |     if (logsRes.ok()) {
  147 |       const logs = await logsRes.json();
  148 |       const logList = logs.results || logs;
  149 |       if (logList.length > 0) {
  150 |         const firstLog = logList[0];
  151 |         // 日志中 request_headers 应包含签名相关头
  152 |         const headers = firstLog.request_headers || {};
  153 |         const hasSignature =
  154 |           headers["X-Ydsz-Signature-256"] || headers["x-ydsz-signature-256"];
  155 |         const hasTimestamp =
  156 |           headers["X-Ydsz-Timestamp"] || headers["x-ydsz-timestamp"];
  157 |         // 至少签名或时间戳存在（取决于后端日志结构）
  158 |         expect(hasSignature || hasTimestamp).toBeTruthy();
  159 |       }
  160 |     }
  161 | 
  162 |     // 清理
  163 |     await request.delete(`${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}`, {
  164 |       headers: { Authorization: `Bearer ${token}` },
  165 |     });
  166 |   });
  167 | });
  168 | 
  169 | test.describe("租户隔离（RLS 边界）", () => {
  170 |   test("跨空间访问项目应返回 404/403", async ({ request }) => {
  171 |     const token = await login(request);
  172 |     const wsId = await getFirstWorkspace(request, token);
  173 |     const projectId = await getFirstProject(request, token, wsId);
  174 | 
  175 |     // 用不存在的 wsId 访问项目 → 应 404
  176 |     const invalidWsRes = await request.get(
  177 |       `${API_URL}/workspaces/99999/projects/${projectId}`,
  178 |       { headers: { Authorization: `Bearer ${token}` } },
  179 |     );
  180 |     expect([403, 404]).toContain(invalidWsRes.status());
  181 |   });
  182 | 
  183 |   test("无效项目 ID 返回 404", async ({ request }) => {
  184 |     const token = await login(request);
  185 |     const wsId = await getFirstWorkspace(request, token);
  186 | 
  187 |     const res = await request.get(
  188 |       `${API_URL}/workspaces/${wsId}/projects/999999`,
  189 |       { headers: { Authorization: `Bearer ${token}` } },
  190 |     );
  191 |     expect(res.status()).toBe(404);
```