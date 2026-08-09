# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: issue-workflow.spec.ts >> 工作项状态机流转 >> 看板拖拽流转调用 transition API
- Location: e2e\issue-workflow.spec.ts:195:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 201
Received: 403
```

# Test source

```ts
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
  167 |     const me = await meRes.json();
  168 | 
  169 |     // 创建指给自己的工作项（触发 issue.assigned 通知）
  170 |     const createRes = await request.post(
  171 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
  172 |       {
  173 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  174 |         data: {
  175 |           type: "task",
  176 |           name: `E2E assigned issue ${Date.now()}`,
  177 |           assignees: [me.id],
  178 |           labels: [],
  179 |           modules: [],
  180 |         },
  181 |       },
  182 |     );
  183 |     expect(createRes.status()).toBe(201);
  184 | 
  185 |     // 自我豁免：操作者不接收自己触发的通知（S7 默认规则）。
  186 |     // 但因为当前 seeded admin 可能是 owner/creator 而非 assignees 逻辑中被排除的人，
  187 |     // 我们只验证：创建操作返回成功 + 通知列表能加载
  188 |     const notifRes = await request.get(
  189 |       `${API_URL}/workspaces/${wsId}/notifications?limit=10`,
  190 |       { headers: { Authorization: `Bearer ${token}` } },
  191 |     );
  192 |     expect(notifRes.ok()).toBe(true);
  193 |   });
  194 | 
  195 |   test("看板拖拽流转调用 transition API", async ({ request }) => {
  196 |     const token = await login(request);
  197 |     const wsId = await getFirstWorkspace(request, token);
  198 |     const projectId = await getFirstProject(request, token, wsId);
  199 | 
  200 |     // 创建任务
  201 |     const createRes = await request.post(
  202 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
  203 |       {
  204 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  205 |         data: {
  206 |           type: "task",
  207 |           name: `E2E kanban drag ${Date.now()}`,
  208 |           assignees: [],
  209 |           labels: [],
  210 |           modules: [],
  211 |         },
  212 |       },
  213 |     );
> 214 |     expect(createRes.status()).toBe(201);
      |                                ^ Error: expect(received).toBe(expected) // Object.is equality
  215 |     const issue = await createRes.json();
  216 | 
  217 |     // reorder API 是看板拖拽排序，测试可用性
  218 |     const reorderRes = await request.patch(
  219 |       `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}/reorder`,
  220 |       {
  221 |         headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
  222 |         data: { sort_order: 1.5 },
  223 |       },
  224 |     );
  225 |     expect(reorderRes.ok()).toBe(true);
  226 |   });
  227 | });
  228 | 
```