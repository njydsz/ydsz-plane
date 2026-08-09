# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: version-lifecycle.spec.ts >> 版本全生命周期 >> 版本发布：检查清单校验 + Release Notes 生成
- Location: e2e\version-lifecycle.spec.ts:137:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
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
  74  |     expect(createRes.status()).toBe(201);
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
> 158 |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
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
  230 |     expect(createRes.status()).toBe(201);
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
```