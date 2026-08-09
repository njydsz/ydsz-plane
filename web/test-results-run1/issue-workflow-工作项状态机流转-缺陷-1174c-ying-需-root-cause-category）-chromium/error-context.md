# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: issue-workflow.spec.ts >> 工作项状态机流转 >> 缺陷流转 required_fields 校验（Fixed→Verifying 需 root_cause_category）
- Location: e2e\issue-workflow.spec.ts:43:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
  1   | /**
  2   |  * 工作项状态机流转 E2E 测试。
  3   |  *
  4   |  * 覆盖：缺陷创建 → 确认 → 处理中 → 修复 → 待验（required_fields 校验）→ 关闭。
  5   |  * 回归 S4 修复的 required_fields 校验 Bug。
  6   |  *
  7   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  8   |  */
  9   | import { expect, test } from "@playwright/test";
  10  | 
  11  | const TEST_EMAIL = "admin@ydsz.dev";
  12  | const TEST_PASSWORD = "Admin@123";
  13  | const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  14  | 
  15  | async function login(request: any) {
  16  |   const res = await request.post(`${API_URL}/auth/login`, {
  17  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  18  |   });
  19  |   expect(res.ok()).toBe(true);
  20  |   const body = await res.json();
  21  |   return body.access_token;
  22  | }
  23  | 
  24  | async function getFirstWorkspace(request: any, token: string) {
  25  |   const res = await request.get(`${API_URL}/workspaces`, {
  26  |     headers: { Authorization: `Bearer ${token}` },
  27  |   });
  28  |   expect(res.ok()).toBe(true);
  29  |   const list = await res.json();
  30  |   return list[0]?.id || 1;
  31  | }
  32  | 
  33  | async function getFirstProject(request: any, token: string, wsId: number) {
  34  |   const res = await request.get(`${API_URL}/workspaces/${wsId}/projects`, {
  35  |     headers: { Authorization: `Bearer ${token}` },
  36  |   });
  37  |   expect(res.ok()).toBe(true);
  38  |   const body = await res.json();
  39  |   return body.results?.[0]?.id || body[0]?.id || 1;
  40  | }
  41  | 
  42  | test.describe("工作项状态机流转", () => {
  43  |   test("缺陷流转 required_fields 校验（Fixed→Verifying 需 root_cause_category）", async ({ request }) => {
  44  |     const token = await login(request);
  45  |     const wsId = await getFirstWorkspace(request, token);
  46  |     const projectId = await getFirstProject(request, token, wsId);
  47  | 
  48  |     // 创建缺陷
  49  |     const createRes = await request.post(
  50  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
  51  |       {
  52  |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  53  |         data: {
  54  |           type: "defect",
  55  |           name: `E2E defect ${Date.now()}`,
  56  |           severity: 3,
  57  |           found_phase: "integration",
  58  |           state_id: 1, // 新建
  59  |           assignees: [],
  60  |           labels: [],
  61  |           modules: [],
  62  |           description_html: "<p>E2E test defect for state machine</p>",
  63  |         },
  64  |       },
  65  |     );
> 66  |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  67  |     const defect = await createRes.json();
  68  |     expect(defect.id).toBeGreaterThan(0);
  69  | 
  70  |     // 获取状态列表以获取合法流转目标
  71  |     const statesRes = await request.get(
  72  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/states`,
  73  |       { headers: { Authorization: `Bearer ${token}` } },
  74  |     );
  75  |     expect(statesRes.ok()).toBe(true);
  76  |     const states = await statesRes.json();
  77  |     const stateList = states.results || states;
  78  | 
  79  |     // 辅助：按状态名查找 id
  80  |     const findStateId = (name: string) =>
  81  |       stateList.find((s: any) => s.name === name)?.id || null;
  82  | 
  83  |     // 尝试流转：新建 → 处理中（不传 required fields），应失败
  84  |     // 通常缺陷的状态机要求 Fixed→Verifying 需要 root_cause_category，
  85  |     // 但 "新建→处理中" 多数没有必填要求，所以我们测试反向场景
  86  |     // 实际回归重点是：required_fields 校验确实生效
  87  | 
  88  |     // 查找所有 transitions 验证 "Trying to skip required fields triggers 422"
  89  |     // 策略：列出 possible transitions，尝试无 context 流转
  90  |     const transitionsRes = await request.get(
  91  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${defect.id}`,
  92  |       { headers: { Authorization: `Bearer ${token}` } },
  93  |     );
  94  |     expect(transitionsRes.ok()).toBe(true);
  95  |     const issueDetail = await transitionsRes.json();
  96  | 
  97  |     // 当前状态应为新建
  98  |     expect(issueDetail.state_id || issueDetail.state?.id).toBeTruthy();
  99  | 
  100 |     // 尝试不带 required_fields 去做所有可能的流转，至少应有一个返回 422
  101 |     // 因为 API 未提供 状态流转的具体目标接口，我们测试最短路径
  102 |     // 实际端点: POST /issues/:id/transition  body: { to_state_id: X, context?: {} }
  103 | 
  104 |     const fixedStateId = findStateId("Fixed") || findStateId("修复中");
  105 |     const verifyingStateId = findStateId("Verifying") || findStateId("待验证");
  106 | 
  107 |     if (fixedStateId && verifyingStateId) {
  108 |       // 先流转到 Fixed（假设无需强校验）
  109 |       const toFixedRes = await request.post(
  110 |         `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${defect.id}/transition`,
  111 |         {
  112 |           headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  113 |           data: { to_state_id: fixedStateId },
  114 |         },
  115 |       );
  116 | 
  117 |       // 成功或失败取决于初始状态是否允许直接跳到 Fixed。
  118 |       // 多数字段允许 "新建→处理中→修复" 的路径，但我们的测试核心是验证
  119 |       // 后续 Verifying 时必须 root_cause_category
  120 | 
  121 |       if (toFixedRes.ok()) {
  122 |         // 尝试 Fixed → Verifying，不带 root_cause_category → 必须 422
  123 |         const toVerifyingRes = await request.post(
  124 |           `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${defect.id}/transition`,
  125 |           {
  126 |             headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  127 |             data: { to_state_id: verifyingStateId },
  128 |           },
  129 |         );
  130 |         expect(toVerifyingRes.status()).toBe(422);
  131 |         const errBody = await toVerifyingRes.json();
  132 |         expect(errBody.error?.code || errBody.error).toBeTruthy();
  133 | 
  134 |         // 带上 root_cause_category 再次流转，必须成功
  135 |         const toVerifyingWithCtxRes = await request.post(
  136 |           `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${defect.id}/transition`,
  137 |           {
  138 |             headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  139 |             data: {
  140 |               to_state_id: verifyingStateId,
  141 |               context: { root_cause_category: "logic_error" },
  142 |             },
  143 |           },
  144 |         );
  145 |         expect(toVerifyingWithCtxRes.ok()).toBe(true);
  146 |       } else {
  147 |         // 状态机不允许跳到 Fixed，此回归跳过
  148 |         console.log(
  149 |           `Cannot transition to Fixed from initial state (${toFixedRes.status()}) — skipping required_fields regression`,
  150 |         );
  151 |       }
  152 |     } else {
  153 |       console.log("States 'Fixed'/'Verifying' not found — skipping transition regression");
  154 |     }
  155 |   });
  156 | 
  157 |   test("工作项创建后指派人收到通知", async ({ request }) => {
  158 |     const token = await login(request);
  159 |     const wsId = await getFirstWorkspace(request, token);
  160 |     const projectId = await getFirstProject(request, token, wsId);
  161 | 
  162 |     // 获取当前用户信息
  163 |     const meRes = await request.get(`${API_URL}/me`, {
  164 |       headers: { Authorization: `Bearer ${token}` },
  165 |     });
  166 |     expect(meRes.ok()).toBe(true);
```