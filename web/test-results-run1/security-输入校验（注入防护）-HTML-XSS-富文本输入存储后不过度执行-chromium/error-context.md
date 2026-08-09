# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: security.spec.ts >> 输入校验（注入防护） >> HTML/XSS 富文本输入存储后不过度执行
- Location: e2e\security.spec.ts:241:3

# Error details

```
Error: expect(received).toContain(expected) // indexOf

Expected value: 403
Received array: [201, 422]
```

# Test source

```ts
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
  192 |   });
  193 | });
  194 | 
  195 | test.describe("输入校验（注入防护）", () => {
  196 |   test("工作项名称 SQL 注入字符被安全处理", async ({ request }) => {
  197 |     const token = await login(request);
  198 |     const wsId = await getFirstWorkspace(request, token);
  199 |     const projectId = await getFirstProject(request, token, wsId);
  200 | 
  201 |     // 创建带特殊字符名称的工作项
  202 |     const createRes = await request.post(
  203 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
  204 |       {
  205 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  206 |         data: {
  207 |           type: "task",
  208 |           name: `E2E SQL注入测试'; DROP TABLE issues; --`,
  209 |           assignees: [],
  210 |           labels: [],
  211 |           modules: [],
  212 |         },
  213 |       },
  214 |     );
  215 |     expect(createRes.status()).toBe(201);
  216 |     const issue = await createRes.json();
  217 | 
  218 |     // 读取验证数据完整无损
  219 |     const getRes = await request.get(
  220 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
  221 |       { headers: { Authorization: `Bearer ${token}` } },
  222 |     );
  223 |     expect(getRes.ok()).toBe(true);
  224 |     const fetched = await getRes.json();
  225 |     expect(fetched.id).toBe(issue.id);
  226 | 
  227 |     // 清理：删除工作项
  228 |     await request.delete(
  229 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
  230 |       { headers: { Authorization: `Bearer ${token}` } },
  231 |     );
  232 | 
  233 |     // 验证其他数据未被破坏（列出工作项应正常工作）
  234 |     const listRes = await request.get(
  235 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=5`,
  236 |       { headers: { Authorization: `Bearer ${token}` } },
  237 |     );
  238 |     expect(listRes.ok()).toBe(true);
  239 |   });
  240 | 
  241 |   test("HTML/XSS 富文本输入存储后不过度执行", async ({ request }) => {
  242 |     const token = await login(request);
  243 |     const wsId = await getFirstWorkspace(request, token);
  244 |     const projectId = await getFirstProject(request, token, wsId);
  245 | 
  246 |     // 创建带 script 标签描述的工作项
  247 |     const createRes = await request.post(
  248 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
  249 |       {
  250 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  251 |         data: {
  252 |           type: "task",
  253 |           name: "E2E XSS Test",
  254 |           description_html: '<p>合法</p><script>alert("xss")</script>',
  255 |           assignees: [],
  256 |           labels: [],
  257 |           modules: [],
  258 |         },
  259 |       },
  260 |     );
  261 |     // 两种可接受路径：201（存储原始 HTML，由前端渲染时过滤）或 422（输入校验拦截）
> 262 |     expect([201, 422]).toContain(createRes.status());
      |                        ^ Error: expect(received).toContain(expected) // indexOf
  263 | 
  264 |     if (createRes.status() === 201) {
  265 |       // 如果后端接受原始 HTML，验证能安全回显（不崩溃）
  266 |       const issue = await createRes.json();
  267 |       await request.delete(
  268 |         `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
  269 |         { headers: { Authorization: `Bearer ${token}` } },
  270 |       );
  271 |     }
  272 |   });
  273 | 
  274 |   test("负数 limit/offset 参数不导致服务错误", async ({ request }) => {
  275 |     const token = await login(request);
  276 |     const wsId = await getFirstWorkspace(request, token);
  277 |     const projectId = await getFirstProject(request, token, wsId);
  278 | 
  279 |     // 负数参数应被 clamp 为安全值或返回 400
  280 |     const res = await request.get(
  281 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=-1&offset=-100`,
  282 |       { headers: { Authorization: `Bearer ${token}` } },
  283 |     );
  284 |     // 可接受：200（后端 clamp）或 400（参数校验）
  285 |     expect([200, 400, 422]).toContain(res.status());
  286 |   });
  287 | });
  288 | 
  289 | test.describe("审计日志", () => {
  290 |   test("管理员操作被记录到审计日志", async ({ request }) => {
  291 |     const token = await login(request);
  292 |     const wsId = await getFirstWorkspace(request, token);
  293 | 
  294 |     // 审计日志查询接口（仅 admin/owner 可用）
  295 |     const auditRes = await request.get(`${API_URL}/workspaces/${wsId}/audit-logs?limit=5`, {
  296 |       headers: { Authorization: `Bearer ${token}` },
  297 |     });
  298 |     // 接口存在且返回（200 或 403 取决于实现）
  299 |     expect([200, 403]).toContain(auditRes.status());
  300 | 
  301 |     if (auditRes.status() === 200) {
  302 |       const logs = await auditRes.json();
  303 |       expect(logs).toBeTruthy();
  304 |     }
  305 |   });
  306 | });
  307 | 
```