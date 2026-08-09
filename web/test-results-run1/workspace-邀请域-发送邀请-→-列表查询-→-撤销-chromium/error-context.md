# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace.spec.ts >> 邀请域 >> 发送邀请 → 列表查询 → 撤销
- Location: e2e\workspace.spec.ts:164:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
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
  103 |     expect(roles.length).toBeGreaterThanOrEqual(4); // Owner/Admin/Member/Guest
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
> 174 |     expect(inviteRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
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
  204 |       data: { email: TEST_EMAIL, role: "member" },
  205 |     });
  206 |     // 已存在成员应返回 409 或 400
  207 |     expect([400, 409, 422]).toContain(inviteRes.status());
  208 |   });
  209 | 
  210 |   test("邀请预览接口（无鉴权）返回邀请详情", async ({ request }) => {
  211 |     const token = await login(request);
  212 |     const ws = await getFirstWorkspace(request, token);
  213 | 
  214 |     // 发送邀请获取 token
  215 |     const inviteEmail = `preview-${Date.now()}@example.com`;
  216 |     const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
  217 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  218 |       data: { email: inviteEmail, role: "member" },
  219 |     });
  220 |     expect(inviteRes.status()).toBe(201);
  221 |     const invitation = await inviteRes.json();
  222 | 
  223 |     // 注意：后端的邀请 preview 需要 token（邀请 token），不是 invitation.id
  224 |     // 这里仅验证接口存在且返回结构合理
  225 |     // 实际 preview endpoint: GET /invitations/:token
  226 |     // invitation.token 可能在返回中，取决于后端实现
  227 |     if (invitation.token) {
  228 |       const previewRes = await request.get(`${API_URL}/invitations/${invitation.token}`);
  229 |       expect(previewRes.ok()).toBe(true);
  230 |       const preview = await previewRes.json();
  231 |       expect(preview.workspace_id).toBe(ws.id);
  232 |       expect(preview.email).toBe(inviteEmail);
  233 |     }
  234 | 
  235 |     // 清理
  236 |     await request.delete(`${API_URL}/workspaces/${ws.id}/invitations/${invitation.id}`, {
  237 |       headers: { Authorization: `Bearer ${token}` },
  238 |     });
  239 |   });
  240 | });
  241 | 
  242 | test.describe("项目域", () => {
  243 |   test("工作空间下项目列表 + 创建项目", async ({ request }) => {
  244 |     const token = await login(request);
  245 |     const ws = await getFirstWorkspace(request, token);
  246 | 
  247 |     // 列表
  248 |     const listRes = await request.get(`${API_URL}/workspaces/${ws.id}/projects`, {
  249 |       headers: { Authorization: `Bearer ${token}` },
  250 |     });
  251 |     expect(listRes.ok()).toBe(true);
  252 |     const body = await listRes.json();
  253 |     const projects = body.results || body;
  254 |     expect(Array.isArray(projects)).toBe(true);
  255 | 
  256 |     // 创建项目
  257 |     const identifier = `E2E${Date.now().toString().slice(-6)}`;
  258 |     const createRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
  259 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  260 |       data: {
  261 |         name: `E2E Project ${Date.now()}`,
  262 |         identifier,
  263 |         description: "E2E test project",
  264 |         network: "public",
  265 |       },
  266 |     });
  267 |     expect(createRes.status()).toBe(201);
  268 |     const project = await createRes.json();
  269 |     expect(project.id).toBeGreaterThan(0);
  270 |     expect(project.identifier).toBe(identifier);
  271 |     expect(project.name).toContain("E2E");
  272 | 
  273 |     // 清理：归档项目
  274 |     const archiveRes = await request.delete(
```