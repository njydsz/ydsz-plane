# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: version-lifecycle.spec.ts >> 版本全生命周期 >> 版本关联迭代 + 缺陷面板查询
- Location: e2e\version-lifecycle.spec.ts:214:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
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
  175 |       },
  176 |     );
  177 |     // 准出失败时应返回 422，或者返回 200 但 version.status 仍为 active
  178 |     if (releaseRes.status() === 422) {
  179 |       // 校验拦截符合预期
  180 |       expect(releaseRes.status()).toBe(422);
  181 |     } else {
  182 |       // 如果后端使用 force_checklist=false 直接返回未通过状态
  183 |       expect(releaseRes.ok()).toBe(true);
  184 |     }
  185 | 
  186 |     // 生成 Release Notes
  187 |     const notesRes = await request.get(
  188 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release-notes`,
  189 |       { headers: { Authorization: `Bearer ${token}` } },
  190 |     );
  191 |     expect(notesRes.ok()).toBe(true);
  192 |     const notes = await notesRes.json();
  193 |     expect(notes).toHaveProperty("release_notes");
  194 | 
  195 |     // 强制发布（绕过检查清单）
  196 |     const forceReleaseRes = await request.post(
  197 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release`,
  198 |       {
  199 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  200 |         data: { force_checklist: true },
  201 |       },
  202 |     );
  203 |     expect(forceReleaseRes.ok()).toBe(true);
  204 |     const released = await forceReleaseRes.json();
  205 |     expect(released.status).toBe("released");
  206 | 
  207 |     // 清理：删除
  208 |     await request.delete(
  209 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
  210 |       { headers: { Authorization: `Bearer ${token}` } },
  211 |     );
  212 |   });
  213 | 
  214 |   test("版本关联迭代 + 缺陷面板查询", async ({ request }) => {
  215 |     const token = await login(request);
  216 |     const wsId = await getFirstWorkspace(request, token);
  217 |     const projectId = await getFirstProject(request, token, wsId);
  218 | 
  219 |     // 创建版本
  220 |     const createRes = await request.post(
  221 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
  222 |       {
  223 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  224 |         data: {
  225 |           name: `E2E Sprint Version ${Date.now()}`,
  226 |           semver: "1.5.0",
  227 |         },
  228 |       },
  229 |     );
> 230 |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  231 |     const version = await createRes.json();
  232 |     const versionId = version.id;
  233 | 
  234 |     // 尝试获取第一个迭代并关联
  235 |     const sprintId = await getFirstSprint(request, token, wsId, projectId);
  236 |     if (sprintId) {
  237 |       const addRes = await request.post(
  238 |         `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`,
  239 |         {
  240 |           headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  241 |           data: { sprint_id: sprintId },
  242 |         },
  243 |       );
  244 |       expect(addRes.ok()).toBe(true);
  245 | 
  246 |       // 列出关联的迭代
  247 |       const listRes = await request.get(
  248 |         `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`,
  249 |         { headers: { Authorization: `Bearer ${token}` } },
  250 |       );
  251 |       expect(listRes.ok()).toBe(true);
  252 |       const sprints = await listRes.json();
  253 |       expect(sprints.results.length).toBeGreaterThanOrEqual(1);
  254 | 
  255 |       // 移除迭代关联
  256 |       const removeRes = await request.delete(
  257 |         `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints/${sprintId}`,
  258 |         { headers: { Authorization: `Bearer ${token}` } },
  259 |       );
  260 |       expect(removeRes.ok()).toBe(true);
  261 |     }
  262 | 
  263 |     // 缺陷面板查询（空版本，但接口应可用）
  264 |     const defectsRes = await request.get(
  265 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/defects`,
  266 |       { headers: { Authorization: `Bearer ${token}` } },
  267 |     );
  268 |     expect(defectsRes.ok()).toBe(true);
  269 |     const defects = await defectsRes.json();
  270 |     expect(defects).toHaveProperty("results");
  271 |     expect(defects).toHaveProperty("total");
  272 | 
  273 |     // 跨版本缺陷过滤接口
  274 |     const filterRes = await request.get(
  275 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/defects?found_version_id=${versionId}&limit=10`,
  276 |       { headers: { Authorization: `Bearer ${token}` } },
  277 |     );
  278 |     expect(filterRes.ok()).toBe(true);
  279 | 
  280 |     // 清理
  281 |     await request.delete(
  282 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
  283 |       { headers: { Authorization: `Bearer ${token}` } },
  284 |     );
  285 |   });
  286 | 
  287 |   test("乐观锁：带 version 字段更新版本", async ({ request }) => {
  288 |     const token = await login(request);
  289 |     const wsId = await getFirstWorkspace(request, token);
  290 |     const projectId = await getFirstProject(request, token, wsId);
  291 | 
  292 |     const createRes = await request.post(
  293 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
  294 |       {
  295 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  296 |         data: {
  297 |           name: `E2E Optimistic ${Date.now()}`,
  298 |           semver: "3.0.0",
  299 |         },
  300 |       },
  301 |     );
  302 |     expect(createRes.status()).toBe(201);
  303 |     const version = await createRes.json();
  304 |     const versionId = version.id;
  305 |     const currentVersion = version.version;
  306 | 
  307 |     // 更新时必须携带 version 字段
  308 |     const updateRes = await request.patch(
  309 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
  310 |       {
  311 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  312 |         data: {
  313 |           name: `E2E Optimistic Updated ${Date.now()}`,
  314 |           version: currentVersion,
  315 |         },
  316 |       },
  317 |     );
  318 |     expect(updateRes.ok()).toBe(true);
  319 |     const updated = await updateRes.json();
  320 |     expect(updated.name).toContain("Updated");
  321 |     expect(updated.version).toBeGreaterThan(currentVersion);
  322 | 
  323 |     // 清理
  324 |     await request.delete(
  325 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
  326 |       { headers: { Authorization: `Bearer ${token}` } },
  327 |     );
  328 |   });
  329 | });
  330 | 
```