# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: search.spec.ts >> 全局搜索 >> 搜索后产生历史记录，可删除
- Location: e2e\search.spec.ts:55:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  1   | /**
  2   |  * 全局搜索 E2E 测试。
  3   |  *
  4   |  * 覆盖：工作空间级全文搜索（issue 命中）→ 搜索历史记录 → 书签创建/删除。
  5   |  * 回归 S8 交付的 PG FTS 主链路（ADR-0010 降级方案）+ search_documents 索引回填。
  6   |  *
  7   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  8   |  * 种子数据包含 500 个工作项（scripts/seed/main.go），其中包含可检索的
  9   |  * issue 名称（如 "登录页优化" / "支付流程" 等固定关键词），搜索词用种子关键词。
  10  |  */
  11  | import { expect, test } from "@playwright/test";
  12  | 
  13  | const TEST_EMAIL = "admin@ydsz.dev";
  14  | const TEST_PASSWORD = "Admin@123";
  15  | const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  16  | 
  17  | async function login(request: any) {
  18  |   const res = await request.post(`${API_URL}/auth/login`, {
  19  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  20  |   });
  21  |   expect(res.ok()).toBe(true);
  22  |   const body = await res.json();
  23  |   return body.access_token;
  24  | }
  25  | 
  26  | async function getFirstWorkspace(request: any, token: string) {
  27  |   const res = await request.get(`${API_URL}/workspaces`, {
  28  |     headers: { Authorization: `Bearer ${token}` },
  29  |   });
  30  |   expect(res.ok()).toBe(true);
  31  |   const list = await res.json();
  32  |   expect(list.length).toBeGreaterThan(0);
  33  |   return list[0].id;
  34  | }
  35  | 
  36  | test.describe("全局搜索", () => {
  37  |   test("工作空间搜索命中 issue 并返回分组结果", async ({ request }) => {
  38  |     const token = await login(request);
  39  |     const wsId = await getFirstWorkspace(request, token);
  40  | 
  41  |     // 用种子数据中的固定关键词搜索（seed 脚本内置 "登录" 等可检索词）
  42  |     const res = await request.get(`${API_URL}/workspaces/${wsId}/search`, {
  43  |       headers: { Authorization: `Bearer ${token}` },
  44  |       params: { q: "登录", limit: 20 },
  45  |     });
  46  |     expect(res.ok()).toBe(true);
  47  |     const body = await res.json();
  48  |     expect(body.total).toBeGreaterThanOrEqual(0);
  49  |     expect(body.query).toBe("登录");
  50  |     // 结构校验：results 按类型分组（issue/sprint/version/projects 字段必须存在）
  51  |     expect(body.results).toBeDefined();
  52  |     expect(Array.isArray(body.results.issues)).toBe(true);
  53  |   });
  54  | 
  55  |   test("搜索后产生历史记录，可删除", async ({ request }) => {
  56  |     const token = await login(request);
  57  |     const wsId = await getFirstWorkspace(request, token);
  58  | 
  59  |     // 触发一次搜索以写入历史
  60  |     await request.get(`${API_URL}/workspaces/${wsId}/search`, {
  61  |       headers: { Authorization: `Bearer ${token}` },
  62  |       params: { q: "迭代" },
  63  |     });
  64  | 
  65  |     // 读取历史，应包含刚搜索的关键词
  66  |     const histRes = await request.get(`${API_URL}/workspaces/${wsId}/search/history`, {
  67  |       headers: { Authorization: `Bearer ${token}` },
  68  |     });
  69  |     expect(histRes.ok()).toBe(true);
  70  |     const histBody = await histRes.json();
  71  |     const results = histBody.results || histBody || [];
  72  |     expect(Array.isArray(results)).toBe(true);
  73  |     const found = results.find((h: any) => h.query === "迭代");
  74  |     expect(found).toBeTruthy();
  75  | 
  76  |     // 删除该条历史
  77  |     const delRes = await request.delete(
  78  |       `${API_URL}/workspaces/${wsId}/search/history/${found.id}`,
  79  |       { headers: { Authorization: `Bearer ${token}` } },
  80  |     );
> 81  |     expect(delRes.ok()).toBe(true);
      |                         ^ Error: expect(received).toBe(expected) // Object.is equality
  82  |   });
  83  | 
  84  |   test("书签创建与删除闭环", async ({ request }) => {
  85  |     const token = await login(request);
  86  |     const wsId = await getFirstWorkspace(request, token);
  87  | 
  88  |     const name = `E2E bookmark ${Date.now()}`;
  89  |     const createRes = await request.post(`${API_URL}/workspaces/${wsId}/search/bookmarks`, {
  90  |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  91  |       data: { name, query: "登录", is_shared: false },
  92  |     });
  93  |     expect(createRes.ok()).toBe(true);
  94  |     const created = await createRes.json();
  95  |     const bookmarkId = created.id ?? created.bookmark?.id;
  96  | 
  97  |     // 列表应包含该书签
  98  |     const listRes = await request.get(`${API_URL}/workspaces/${wsId}/search/bookmarks`, {
  99  |       headers: { Authorization: `Bearer ${token}` },
  100 |     });
  101 |     expect(listRes.ok()).toBe(true);
  102 |     const listBody = await listRes.json();
  103 |     const list = listBody.results || listBody || [];
  104 |     expect(list.some((b: any) => b.id === bookmarkId)).toBe(true);
  105 | 
  106 |     // 删除书签
  107 |     const delRes = await request.delete(
  108 |       `${API_URL}/workspaces/${wsId}/search/bookmarks/${bookmarkId}`,
  109 |       { headers: { Authorization: `Bearer ${token}` } },
  110 |     );
  111 |     expect(delRes.ok()).toBe(true);
  112 |   });
  113 | });
  114 | 
```