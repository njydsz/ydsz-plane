# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace.spec.ts >> 工作空间域 >> 角色定义列表
- Location: e2e\workspace.spec.ts:94:3

# Error details

```
Error: expect(received).toBeGreaterThanOrEqual(expected)

Matcher error: received value must be a number or bigint

Received has value: undefined
```

# Test source

```ts
  3   |  *
  4   |  * 覆盖：空间列表/详情/更新、成员列表/角色切换/移除、邀请发送/列表/撤销、
  5   |  *       RBAC 角色查询、项目 CRUD。
  6   |  * 回归 S2 IAM 与工作空间域。
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
  31  |   return list[0];
  32  | }
  33  | 
  34  | test.describe("工作空间域", () => {
  35  |   test("列表 → 详情 → 更新基本信息", async ({ request }) => {
  36  |     const token = await login(request);
  37  | 
  38  |     // 列表
  39  |     const listRes = await request.get(`${API_URL}/workspaces`, {
  40  |       headers: { Authorization: `Bearer ${token}` },
  41  |     });
  42  |     expect(listRes.ok()).toBe(true);
  43  |     const list = await listRes.json();
  44  |     expect(list.length).toBeGreaterThanOrEqual(1);
  45  | 
  46  |     // 详情
  47  |     const wsId = list[0].id;
  48  |     const detailRes = await request.get(`${API_URL}/workspaces/${wsId}`, {
  49  |       headers: { Authorization: `Bearer ${token}` },
  50  |     });
  51  |     expect(detailRes.ok()).toBe(true);
  52  |     const ws = await detailRes.json();
  53  |     expect(ws.id).toBe(wsId);
  54  |     expect(ws.slug).toBeTruthy();
  55  | 
  56  |     // 更新（timezone / language）
  57  |     const updateRes = await request.patch(`${API_URL}/workspaces/${wsId}`, {
  58  |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  59  |       data: { timezone: "Asia/Shanghai", language: "zh-CN" },
  60  |     });
  61  |     expect(updateRes.ok()).toBe(true);
  62  |     const updated = await updateRes.json();
  63  |     expect(updated.timezone).toBe("Asia/Shanghai");
  64  |     expect(updated.language).toBe("zh-CN");
  65  |   });
  66  | 
  67  |   test("slug 唯一性校验", async ({ request }) => {
  68  |     const token = await login(request);
  69  |     const ws = await getFirstWorkspace(request, token);
  70  | 
  71  |     // 尝试创建与已有空间同 slug 的空间，应返回 409
  72  |     const createRes = await request.post(`${API_URL}/workspaces`, {
  73  |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  74  |       data: { name: "Dup Slug Test", slug: ws.slug },
  75  |     });
  76  |     // slug 唯一性冲突应为 409
  77  |     expect(createRes.status()).toBe(409);
  78  |   });
  79  | 
  80  |   test("当前用户角色 + 权限查询", async ({ request }) => {
  81  |     const token = await login(request);
  82  |     const ws = await getFirstWorkspace(request, token);
  83  | 
  84  |     const roleRes = await request.get(`${API_URL}/workspaces/${ws.id}/role`, {
  85  |       headers: { Authorization: `Bearer ${token}` },
  86  |     });
  87  |     expect(roleRes.ok()).toBe(true);
  88  |     const roleData = await roleRes.json();
  89  |     expect(roleData.role).toBeTruthy();
  90  |     expect(roleData.role.slug).toBeTruthy();
  91  |     expect(Array.isArray(roleData.permissions)).toBe(true);
  92  |   });
  93  | 
  94  |   test("角色定义列表", async ({ request }) => {
  95  |     const token = await login(request);
  96  |     const ws = await getFirstWorkspace(request, token);
  97  | 
  98  |     const rolesRes = await request.get(`${API_URL}/workspaces/${ws.id}/roles`, {
  99  |       headers: { Authorization: `Bearer ${token}` },
  100 |     });
  101 |     expect(rolesRes.ok()).toBe(true);
  102 |     const roles = await rolesRes.json();
> 103 |     expect(roles.length).toBeGreaterThanOrEqual(4); // Owner/Admin/Member/Guest
      |                          ^ Error: expect(received).toBeGreaterThanOrEqual(expected)
  104 |   });
  105 | });
  106 | 
  107 | test.describe("成员管理", () => {
  108 |   test("成员列表包含 admin 自身", async ({ request }) => {
  109 |     const token = await login(request);
  110 |     const ws = await getFirstWorkspace(request, token);
  111 | 
  112 |     const membersRes = await request.get(`${API_URL}/workspaces/${ws.id}/members`, {
  113 |       headers: { Authorization: `Bearer ${token}` },
  114 |     });
  115 |     expect(membersRes.ok()).toBe(true);
  116 |     const members = await membersRes.json();
  117 |     expect(members.length).toBeGreaterThanOrEqual(1);
  118 |     const adminMember = members.find((m: any) => m.email === TEST_EMAIL);
  119 |     expect(adminMember).toBeTruthy();
  120 |   });
  121 | 
  122 |   test("成员切换角色（admin 不能降级自己，应 403 或 400）", async ({ request }) => {
  123 |     const token = await login(request);
  124 |     const ws = await getFirstWorkspace(request, token);
  125 | 
  126 |     // 获取 admin 自身 ID
  127 |     const meRes = await request.get(`${API_URL}/me`, {
  128 |       headers: { Authorization: `Bearer ${token}` },
  129 |     });
  130 |     expect(meRes.ok()).toBe(true);
  131 |     const me = await meRes.json();
  132 | 
  133 |     // 尝试把自己降级为 guest（应被拒绝）
  134 |     const changeRes = await request.patch(
  135 |       `${API_URL}/workspaces/${ws.id}/members/${me.id}`,
  136 |       {
  137 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  138 |         data: { role: "guest" },
  139 |       },
  140 |     );
  141 |     // Owner 自降应被拒绝（403/400/422 取决于后端实现）
  142 |     expect([400, 403, 422]).toContain(changeRes.status());
  143 |   });
  144 | 
  145 |   test("Owner 不可被移除（移除自己应失败）", async ({ request }) => {
  146 |     const token = await login(request);
  147 |     const ws = await getFirstWorkspace(request, token);
  148 | 
  149 |     const meRes = await request.get(`${API_URL}/me`, {
  150 |       headers: { Authorization: `Bearer ${token}` },
  151 |     });
  152 |     const me = await meRes.json();
  153 | 
  154 |     const removeRes = await request.delete(
  155 |       `${API_URL}/workspaces/${ws.id}/members/${me.id}`,
  156 |       { headers: { Authorization: `Bearer ${token}` } },
  157 |     );
  158 |     // Owner 不能移除自己（最后一个 owner）
  159 |     expect([400, 403, 422]).toContain(removeRes.status());
  160 |   });
  161 | });
  162 | 
  163 | test.describe("邀请域", () => {
  164 |   test("发送邀请 → 列表查询 → 撤销", async ({ request }) => {
  165 |     const token = await login(request);
  166 |     const ws = await getFirstWorkspace(request, token);
  167 | 
  168 |     // 发送邀请
  169 |     const inviteEmail = `invite-${Date.now()}@example.com`;
  170 |     const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
  171 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  172 |       data: { email: inviteEmail, role: "member", message: "E2E test invitation" },
  173 |     });
  174 |     expect(inviteRes.status()).toBe(201);
  175 |     const invitation = await inviteRes.json();
  176 |     expect(invitation.id).toBeGreaterThan(0);
  177 |     expect(invitation.status).toBe("pending");
  178 |     expect(invitation.email).toBe(inviteEmail);
  179 | 
  180 |     // 列出邀请
  181 |     const listRes = await request.get(`${API_URL}/workspaces/${ws.id}/invitations?status=pending`, {
  182 |       headers: { Authorization: `Bearer ${token}` },
  183 |     });
  184 |     expect(listRes.ok()).toBe(true);
  185 |     const invitations = await listRes.json();
  186 |     const found = invitations.find((inv: any) => inv.id === invitation.id);
  187 |     expect(found).toBeTruthy();
  188 | 
  189 |     // 撤销邀请
  190 |     const revokeRes = await request.delete(
  191 |       `${API_URL}/workspaces/${ws.id}/invitations/${invitation.id}`,
  192 |       { headers: { Authorization: `Bearer ${token}` } },
  193 |     );
  194 |     expect(revokeRes.ok()).toBe(true);
  195 |   });
  196 | 
  197 |   test("重复邀请已存在的成员应失败", async ({ request }) => {
  198 |     const token = await login(request);
  199 |     const ws = await getFirstWorkspace(request, token);
  200 | 
  201 |     // 邀请已存在的 admin 自身
  202 |     const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
  203 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
```