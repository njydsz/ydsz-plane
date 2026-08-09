# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: webhook.spec.ts >> Webhook 域 >> 订阅 CRUD + 签名头落日志 + 错误路径
- Location: e2e\webhook.spec.ts:32:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  1   | /**
  2   |  * Webhook 域 E2E 测试（S10 出口：签名/重放/限流场景的 API 侧验证）。
  3   |  *
  4   |  * 覆盖：
  5   |  *  - 订阅 CRUD 全生命周期（创建返回 secret、查询、更新、停用、删除）
  6   |  *  - 投递日志查询 + 日志中携带签名头（X-Ydsz-Signature-256）与时间戳
  7   |  *  - 测试推送（目标不可达时返回明确失败语义，验证错误路径）
  8   |  *  - 手动重投：对不存在日志返回 4xx（验证路由与鉴权生效）
  9   |  *  - 无效事件类型被拒（事件目录白名单校验）
  10  |  *
  11  |  * 说明：接收方侧的"签名伪造/重放拒绝"校验发生在外部消费者侧，本套件
  12  |  * 验证的是我们作为发送方产出的签名头与投递日志完备性。
  13  |  *
  14  |  * 运行前提：后端已启动（make migrate && make seed）。
  15  |  */
  16  | import { expect, test } from "@playwright/test";
  17  | 
  18  | const TEST_EMAIL = "admin@ydsz.dev";
  19  | const TEST_PASSWORD = "Admin@123";
  20  | const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  21  | 
  22  | async function login(request: any) {
  23  |   const res = await request.post(`${apiURL}/auth/login`, {
  24  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  25  |   });
  26  |   expect(res.ok()).toBe(true);
  27  |   const { access_token: token } = await res.json();
  28  |   return token as string;
  29  | }
  30  | 
  31  | test.describe("Webhook 域", () => {
  32  |   test("订阅 CRUD + 签名头落日志 + 错误路径", async ({ request }) => {
  33  |     const token = await login(request);
  34  |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  35  |       headers: { Authorization: `Bearer ${token}` },
  36  |     });
  37  |     expect(wsRes.ok()).toBe(true);
  38  |     const wsList = await wsRes.json();
  39  |     const wsId = wsList[0]?.id ?? 1;
  40  |     const base = `${apiURL}/workspaces/${wsId}/webhooks`;
  41  | 
  42  |     // 1. 创建订阅（secret 由服务端生成；目标指向不可达端口以模拟失败路径）
  43  |     const createRes = await request.post(`${base}`, {
  44  |       headers: { Authorization: `Bearer ${token}` },
  45  |       data: {
  46  |         name: `E2E-Webhook-${Date.now()}`,
  47  |         target_url: "http://127.0.0.1:9/e2e-receiver",
  48  |         events: ["issue.created", "issue.updated"],
  49  |       },
  50  |     });
> 51  |     expect(createRes.ok()).toBe(true);
      |                            ^ Error: expect(received).toBe(expected) // Object.is equality
  52  |     const created = await createRes.json();
  53  |     expect(created.id).toBeGreaterThan(0);
  54  |     // 创建时返回 secret（32 字节 hex）
  55  |     expect(created.secret).toMatch(/^[0-9a-f]{64}$/);
  56  |     const whId = created.id;
  57  | 
  58  |     // 2. 查询 + 列表
  59  |     const getRes = await request.get(`${base}/${whId}`, {
  60  |       headers: { Authorization: `Bearer ${token}` },
  61  |     });
  62  |     expect(getRes.ok()).toBe(true);
  63  |     const listRes = await request.get(`${base}`, {
  64  |       headers: { Authorization: `Bearer ${token}` },
  65  |     });
  66  |     expect(listRes.ok()).toBe(true);
  67  | 
  68  |     // 3. 测试推送：目标不可达 → 应返回 4xx 明确失败（错误路径可预期）
  69  |     const pingRes = await request.post(`${base}/${whId}/test`, {
  70  |       headers: { Authorization: `Bearer ${token}` },
  71  |     });
  72  |     expect(pingRes.status()).toBeGreaterThanOrEqual(400);
  73  | 
  74  |     // 4. 投递日志：失败投递也应落日志，且 request_headers 含签名头与时间戳
  75  |     const logsRes = await request.get(`${base}/${whId}/logs?limit=5`, {
  76  |       headers: { Authorization: `Bearer ${token}` },
  77  |     });
  78  |     expect(logsRes.ok()).toBe(true);
  79  |     const logsBody = await logsRes.json();
  80  |     const logs = Array.isArray(logsBody) ? logsBody : logsBody.items ?? [];
  81  |     if (logs.length > 0) {
  82  |       const log = logs[0];
  83  |       expect(log.webhook_id).toBe(whId);
  84  |       expect(log.delivery_id).toBeTruthy();
  85  |       // 请求头 JSON 中应包含签名头（HMAC-SHA256）与时间戳
  86  |       const headers =
  87  |         typeof log.request_headers === "string"
  88  |           ? JSON.parse(log.request_headers || "{}")
  89  |           : log.request_headers ?? {};
  90  |       const headerStr = JSON.stringify(headers);
  91  |       expect(headerStr).toContain("X-Ydsz-Signature-256");
  92  |       expect(headerStr).toContain("X-Ydsz-Timestamp");
  93  |       expect(headerStr).toContain("X-Ydsz-Event");
  94  |     }
  95  | 
  96  |     // 5. 手动重投：对不存在的日志 ID 应返回 4xx（路由 + 鉴权 + 校验生效）
  97  |     const retryRes = await request.post(`${base}/${whId}/logs/99999999/retry`, {
  98  |       headers: { Authorization: `Bearer ${token}` },
  99  |     });
  100 |     expect(retryRes.status()).toBeGreaterThanOrEqual(400);
  101 | 
  102 |     // 6. 事件目录白名单：非法事件类型创建应被拒
  103 |     const badRes = await request.post(`${base}`, {
  104 |       headers: { Authorization: `Bearer ${token}` },
  105 |       data: {
  106 |         name: "bad-events",
  107 |         target_url: "https://example.invalid/hook",
  108 |         events: ["issue.hacked"],
  109 |       },
  110 |     });
  111 |     expect(badRes.ok()).toBe(false);
  112 | 
  113 |     // 7. 更新（停用）+ 删除
  114 |     const patchRes = await request.patch(`${base}/${whId}`, {
  115 |       headers: { Authorization: `Bearer ${token}` },
  116 |       data: { is_active: false },
  117 |     });
  118 |     expect(patchRes.ok()).toBe(true);
  119 |     const delRes = await request.delete(`${base}/${whId}`, {
  120 |       headers: { Authorization: `Bearer ${token}` },
  121 |     });
  122 |     expect(delRes.ok()).toBe(true);
  123 |   });
  124 | });
  125 | 
```