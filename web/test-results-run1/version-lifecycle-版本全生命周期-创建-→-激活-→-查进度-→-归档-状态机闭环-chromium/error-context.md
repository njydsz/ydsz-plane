# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: version-lifecycle.spec.ts >> 版本全生命周期 >> 创建 → 激活 → 查进度 → 归档 状态机闭环
- Location: e2e\version-lifecycle.spec.ts:53:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
  1   | /**
  2   |  * 版本（Version）全生命周期 E2E 测试。
  3   |  *
  4   |  * 覆盖：创建版本 → 激活 → 关联迭代 → 进度查询 → 发布（含检查清单准出）→ 归档。
  5   |  * 回归 S6 版本状态机 + S4 缺陷面板联动。
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
  42  | async function getFirstSprint(request: any, token: string, wsId: number, projectId: number) {
  43  |   const res = await request.get(`${API_URL}/workspaces/${wsId}/projects/${projectId}/sprints`, {
  44  |     headers: { Authorization: `Bearer ${token}` },
  45  |   });
  46  |   if (!res.ok()) return null;
  47  |   const body = await res.json();
  48  |   const list = body.results || body;
  49  |   return list[0]?.id || null;
  50  | }
  51  | 
  52  | test.describe("版本全生命周期", () => {
  53  |   test("创建 → 激活 → 查进度 → 归档 状态机闭环", async ({ request }) => {
  54  |     const token = await login(request);
  55  |     const wsId = await getFirstWorkspace(request, token);
  56  |     const projectId = await getFirstProject(request, token, wsId);
  57  | 
  58  |     // 1. 创建版本
  59  |     const uniqueName = `E2E Version ${Date.now()}`;
  60  |     const createRes = await request.post(
  61  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
  62  |       {
  63  |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  64  |         data: {
  65  |           name: uniqueName,
  66  |           semver: "1.0.0",
  67  |           description: "E2E test version",
  68  |           start_date: "2026-01-01",
  69  |           end_date: "2026-06-30",
  70  |           target_date: "2026-06-15",
  71  |         },
  72  |       },
  73  |     );
> 74  |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  75  |     const version = await createRes.json();
  76  |     expect(version.id).toBeGreaterThan(0);
  77  |     expect(version.status).toBe("planning");
  78  |     expect(version.version).toBeGreaterThanOrEqual(0);
  79  | 
  80  |     const versionId = version.id;
  81  | 
  82  |     // 2. 激活版本 planning → active
  83  |     const activateRes = await request.post(
  84  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/activate`,
  85  |       { headers: { Authorization: `Bearer ${token}` } },
  86  |     );
  87  |     expect(activateRes.ok()).toBe(true);
  88  |     const activated = await activateRes.json();
  89  |     expect(activated.status).toBe("active");
  90  | 
  91  |     // 3. 查询进度聚合接口
  92  |     const progressRes = await request.get(
  93  |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/progress`,
  94  |       { headers: { Authorization: `Bearer ${token}` } },
  95  |     );
  96  |     expect(progressRes.ok()).toBe(true);
  97  |     const progress = await progressRes.json();
  98  |     expect(progress).toHaveProperty("total_issues");
  99  |     expect(progress).toHaveProperty("completion_rate");
  100 | 
  101 |     // 4. 查询质量指标接口
  102 |     const qualityRes = await request.get(
  103 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/quality`,
  104 |       { headers: { Authorization: `Bearer ${token}` } },
  105 |     );
  106 |     expect(qualityRes.ok()).toBe(true);
  107 |     const quality = await qualityRes.json();
  108 |     expect(quality).toHaveProperty("total_bugs");
  109 |     expect(quality).toHaveProperty("pass_quality_gate");
  110 | 
  111 |     // 5. 交付报告接口
  112 |     const reportRes = await request.get(
  113 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/delivery-report`,
  114 |       { headers: { Authorization: `Bearer ${token}` } },
  115 |     );
  116 |     expect(reportRes.ok()).toBe(true);
  117 |     const report = await reportRes.json();
  118 |     expect(report).toHaveProperty("eligible_to_release");
  119 | 
  120 |     // 6. 归档版本 active → archived
  121 |     const archiveRes = await request.post(
  122 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/archive`,
  123 |       { headers: { Authorization: `Bearer ${token}` } },
  124 |     );
  125 |     expect(archiveRes.ok()).toBe(true);
  126 |     const archived = await archiveRes.json();
  127 |     expect(archived.status).toBe("archived");
  128 | 
  129 |     // 7. 清理：删除版本
  130 |     const deleteRes = await request.delete(
  131 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
  132 |       { headers: { Authorization: `Bearer ${token}` } },
  133 |     );
  134 |     expect(deleteRes.ok()).toBe(true);
  135 |   });
  136 | 
  137 |   test("版本发布：检查清单校验 + Release Notes 生成", async ({ request }) => {
  138 |     const token = await login(request);
  139 |     const wsId = await getFirstWorkspace(request, token);
  140 |     const projectId = await getFirstProject(request, token, wsId);
  141 | 
  142 |     // 创建带检查清单的版本
  143 |     const createRes = await request.post(
  144 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
  145 |       {
  146 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  147 |         data: {
  148 |           name: `E2E Release ${Date.now()}`,
  149 |           semver: "2.0.0",
  150 |           checklist: [
  151 |             { id: "c1", label: "代码冻结", required: true, checked: false },
  152 |             { id: "c2", label: "回归测试通过", required: true, checked: false },
  153 |             { id: "c3", label: "文档更新", required: false, checked: true },
  154 |           ],
  155 |         },
  156 |       },
  157 |     );
  158 |     expect(createRes.status()).toBe(201);
  159 |     const version = await createRes.json();
  160 |     const versionId = version.id;
  161 | 
  162 |     // 激活
  163 |     const activateRes = await request.post(
  164 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/activate`,
  165 |       { headers: { Authorization: `Bearer ${token}` } },
  166 |     );
  167 |     expect(activateRes.ok()).toBe(true);
  168 | 
  169 |     // 尝试发布（有必填检查项未勾选，应失败或返回 eligible_to_release=false）
  170 |     const releaseRes = await request.post(
  171 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release`,
  172 |       {
  173 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  174 |         data: {},
```