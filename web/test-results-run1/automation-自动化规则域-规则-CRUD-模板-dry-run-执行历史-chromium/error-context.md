# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: automation.spec.ts >> 自动化规则域 >> 规则 CRUD + 模板 + dry-run + 执行历史
- Location: e2e\automation.spec.ts:50:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  1   | /**
  2   |  * 自动化域 E2E 测试（S11 出口：自动化并发竞争用例）。
  3   |  *
  4   |  * 覆盖：
  5   |  *  - 规则 CRUD（创建/查询/更新/启停/删除）
  6   |  *  - 模板列表可用
  7   |  *  - Dry-Run 试运行（条件不命中时返回 matched=false 而非执行动作）
  8   |  *  - 执行历史端点返回结构化记录
  9   |  *  - 并发场景：同规则多次触发 → 执行记录逐条落库（依赖 worker + RabbitMQ，
  10  |  *    若事件链路未启动则跳过，不影响其余断言）
  11  |  *
  12  |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  13  |  */
  14  | import { expect, test } from "@playwright/test";
  15  | 
  16  | const TEST_EMAIL = "admin@ydsz.dev";
  17  | const TEST_PASSWORD = "Admin@123";
  18  | const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  19  | 
  20  | async function login(request: any) {
  21  |   const res = await request.post(`${apiURL}/auth/login`, {
  22  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  23  |   });
  24  |   expect(res.ok()).toBe(true);
  25  |   const { access_token: token } = await res.json();
  26  |   return token as string;
  27  | }
  28  | 
  29  | async function wsAndProject(request: any, token: string) {
  30  |   const wsRes = await request.get(`${apiURL}/workspaces`, {
  31  |     headers: { Authorization: `Bearer ${token}` },
  32  |   });
  33  |   expect(wsRes.ok()).toBe(true);
  34  |   const wsList = await wsRes.json();
  35  |   const wsId = wsList[0]?.id ?? 1;
  36  |   const slug = wsList[0]?.slug ?? "acme";
  37  | 
  38  |   // 取项目列表（seed 提供）
  39  |   const projRes = await request.get(
  40  |     `${apiURL}/workspaces/${wsId}/projects?limit=5`,
  41  |     { headers: { Authorization: `Bearer ${token}` } },
  42  |   );
  43  |   const projList = projRes.ok() ? await projRes.json() : { items: [] };
  44  |   const items = Array.isArray(projList) ? projList : projList.items ?? [];
  45  |   const projectId = items[0]?.id ?? 1;
  46  |   return { wsId, slug, projectId };
  47  | }
  48  | 
  49  | test.describe("自动化规则域", () => {
  50  |   test("规则 CRUD + 模板 + dry-run + 执行历史", async ({ request }) => {
  51  |     const token = await login(request);
  52  |     const { wsId, projectId } = await wsAndProject(request, token);
  53  |     const base = `${apiURL}/workspaces/${wsId}/projects/${projectId}/automation`;
  54  | 
  55  |     // 1. 模板列表（内置 15 条，开箱可用）
  56  |     const tplRes = await request.get(`${base}/templates`, {
  57  |       headers: { Authorization: `Bearer ${token}` },
  58  |     });
> 59  |     expect(tplRes.ok()).toBe(true);
      |                         ^ Error: expect(received).toBe(expected) // Object.is equality
  60  |     const templates = await tplRes.json();
  61  |     expect(Array.isArray(templates.items ?? templates)).toBe(true);
  62  | 
  63  |     // 2. 创建规则（状态流转通知：issue.status_changed → 通知创建者）
  64  |     const createRes = await request.post(`${base}`, {
  65  |       headers: { Authorization: `Bearer ${token}` },
  66  |       data: {
  67  |         name: `E2E-自动化-${Date.now()}`,
  68  |         description: "Playwright 创建的临时规则",
  69  |         project_id: projectId,
  70  |         status: "active",
  71  |         dsl: {
  72  |           trigger: { type: "issue.status_changed", filter: {} },
  73  |           conditions: [{ field: "issue.type_code", op: "eq", value: "defect" }],
  74  |           actions: [
  75  |             {
  76  |               type: "notify",
  77  |               config: {
  78  |                 channel: "in_app",
  79  |                 target: "${issue.created_by}",
  80  |                 template: "缺陷 {{issue.identifier}} 状态已变更",
  81  |               },
  82  |             },
  83  |           ],
  84  |         },
  85  |       },
  86  |     });
  87  |     expect(createRes.ok()).toBe(true);
  88  |     const rule = await createRes.json();
  89  |     expect(rule.id).toBeGreaterThan(0);
  90  | 
  91  |     // 3. 查询单个规则
  92  |     const getRes = await request.get(`${base}/${rule.id}`, {
  93  |       headers: { Authorization: `Bearer ${token}` },
  94  |     });
  95  |     expect(getRes.ok()).toBe(true);
  96  | 
  97  |     // 4. dry-run：DSL 校验器应判定规则合法（valid=true）且动作数正确
  98  |     const dryRunRes = await request.post(`${base}/dry-run`, {
  99  |       headers: { Authorization: `Bearer ${token}` },
  100 |       data: {
  101 |         dsl: {
  102 |           trigger: { type: "issue.status_changed", filter: {} },
  103 |           conditions: [{ field: "issue.type_code", op: "eq", value: "defect" }],
  104 |           actions: [
  105 |             {
  106 |               type: "notify",
  107 |               config: { channel: "in_app", target: "${issue.created_by}" },
  108 |             },
  109 |           ],
  110 |         },
  111 |       },
  112 |     });
  113 |     expect(dryRunRes.ok()).toBe(true);
  114 |     const dry = await dryRunRes.json();
  115 |     expect(dry.valid).toBe(true);
  116 |     expect(dry.actions).toBe(1);
  117 |     expect(dry.trigger_type).toBe("issue.status_changed");
  118 | 
  119 |     // dry-run 负例：非法动作类型应被校验器拒绝
  120 |     const badDry = await request.post(`${base}/dry-run`, {
  121 |       headers: { Authorization: `Bearer ${token}` },
  122 |       data: {
  123 |         dsl: {
  124 |           trigger: { type: "issue.created" },
  125 |           actions: [{ type: "explode" }],
  126 |         },
  127 |       },
  128 |     });
  129 |     expect(badDry.ok()).toBe(true);
  130 |     const badDryBody = await badDry.json();
  131 |     expect(badDryBody.valid).toBe(false);
  132 | 
  133 |     // 5. 执行历史端点返回结构化记录（可能为空，但结构正确）
  134 |     const execRes = await request.get(`${base}/executions?limit=10`, {
  135 |       headers: { Authorization: `Bearer ${token}` },
  136 |     });
  137 |     expect(execRes.ok()).toBe(true);
  138 |     const execBody = await execRes.json();
  139 |     const execs = Array.isArray(execBody) ? execBody : execBody.items ?? [];
  140 |     for (const e of execs) {
  141 |       expect(e.rule_id).toBeGreaterThan(0);
  142 |       expect(["success", "failed", "skipped", "dry_run"]).toContain(e.status);
  143 |     }
  144 | 
  145 |     // 6. 更新 + 停用 + 删除
  146 |     const toggleRes = await request.post(`${base}/${rule.id}/toggle`, {
  147 |       headers: { Authorization: `Bearer ${token}` },
  148 |     });
  149 |     expect(toggleRes.ok()).toBe(true);
  150 |     const delRes = await request.delete(`${base}/${rule.id}`, {
  151 |       headers: { Authorization: `Bearer ${token}` },
  152 |     });
  153 |     expect(delRes.ok()).toBe(true);
  154 |   });
  155 | 
  156 |   test("并发触发：同规则多条执行记录落库（事件链路就绪时生效）", async ({ request }) => {
  157 |     test.setTimeout(60_000);
  158 |     const token = await login(request);
  159 |     const { wsId, projectId } = await wsAndProject(request, token);
```