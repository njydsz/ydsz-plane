# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: workspace.spec.ts >> 项目域 >> 工作空间下项目列表 + 创建项目
- Location: e2e\workspace.spec.ts:243:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
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
> 267 |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  268 |     const project = await createRes.json();
  269 |     expect(project.id).toBeGreaterThan(0);
  270 |     expect(project.identifier).toBe(identifier);
  271 |     expect(project.name).toContain("E2E");
  272 | 
  273 |     // 清理：归档项目
  274 |     const archiveRes = await request.delete(
  275 |       `${API_URL}/workspaces/${ws.id}/projects/${project.id}`,
  276 |       { headers: { Authorization: `Bearer ${token}` } },
  277 |     );
  278 |     expect(archiveRes.ok()).toBe(true);
  279 |   });
  280 | 
  281 |   test("项目 identifier 唯一性（同空间下重复应失败）", async ({ request }) => {
  282 |     const token = await login(request);
  283 |     const ws = await getFirstWorkspace(request, token);
  284 | 
  285 |     const identifier = `E2E${Date.now().toString().slice(-6)}`;
  286 |     // 先创建
  287 |     const firstRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
  288 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  289 |       data: { name: "First", identifier },
  290 |     });
  291 |     expect(firstRes.status()).toBe(201);
  292 |     const projectId = (await firstRes.json()).id;
  293 | 
  294 |     // 重复 identifier
  295 |     const dupRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
  296 |       headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  297 |       data: { name: "Duplicate", identifier },
  298 |     });
  299 |     expect([400, 409, 422]).toContain(dupRes.status());
  300 | 
  301 |     // 清理
  302 |     await request.delete(`${API_URL}/workspaces/${ws.id}/projects/${projectId}`, {
  303 |       headers: { Authorization: `Bearer ${token}` },
  304 |     });
  305 |   });
  306 | });
  307 | 
```