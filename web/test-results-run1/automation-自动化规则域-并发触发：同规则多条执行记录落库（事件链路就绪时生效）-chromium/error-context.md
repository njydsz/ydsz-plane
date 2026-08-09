# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: automation.spec.ts >> 自动化规则域 >> 并发触发：同规则多条执行记录落库（事件链路就绪时生效）
- Location: e2e\automation.spec.ts:156:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  86  |     });
  87  |     expect(createRes.ok()).toBe(true);
  88  |     const rule = await createRes.json();
  89  |     expect(rule.id).toBeGreaterThan(0);
  90  | 
  91  |     // 3. 查询单个规则
  92  |     const getRes = await request.get(`${base}/${rule.id}`, {
  93  |       headers: { Authorization: `Bearer ${token}` },
  94  |     });
  95  |     expect(getRes.ok()).toBe(true);
  96  | 
  97  |     // 4. dry-run：DSL 校验器应判定规则合法（valid=true）且动作数正确
  98  |     const dryRunRes = await request.post(`${base}/dry-run`, {
  99  |       headers: { Authorization: `Bearer ${token}` },
  100 |       data: {
  101 |         dsl: {
  102 |           trigger: { type: "issue.status_changed", filter: {} },
  103 |           conditions: [{ field: "issue.type_code", op: "eq", value: "defect" }],
  104 |           actions: [
  105 |             {
  106 |               type: "notify",
  107 |               config: { channel: "in_app", target: "${issue.created_by}" },
  108 |             },
  109 |           ],
  110 |         },
  111 |       },
  112 |     });
  113 |     expect(dryRunRes.ok()).toBe(true);
  114 |     const dry = await dryRunRes.json();
  115 |     expect(dry.valid).toBe(true);
  116 |     expect(dry.actions).toBe(1);
  117 |     expect(dry.trigger_type).toBe("issue.status_changed");
  118 | 
  119 |     // dry-run 负例：非法动作类型应被校验器拒绝
  120 |     const badDry = await request.post(`${base}/dry-run`, {
  121 |       headers: { Authorization: `Bearer ${token}` },
  122 |       data: {
  123 |         dsl: {
  124 |           trigger: { type: "issue.created" },
  125 |           actions: [{ type: "explode" }],
  126 |         },
  127 |       },
  128 |     });
  129 |     expect(badDry.ok()).toBe(true);
  130 |     const badDryBody = await badDry.json();
  131 |     expect(badDryBody.valid).toBe(false);
  132 | 
  133 |     // 5. 执行历史端点返回结构化记录（可能为空，但结构正确）
  134 |     const execRes = await request.get(`${base}/executions?limit=10`, {
  135 |       headers: { Authorization: `Bearer ${token}` },
  136 |     });
  137 |     expect(execRes.ok()).toBe(true);
  138 |     const execBody = await execRes.json();
  139 |     const execs = Array.isArray(execBody) ? execBody : execBody.items ?? [];
  140 |     for (const e of execs) {
  141 |       expect(e.rule_id).toBeGreaterThan(0);
  142 |       expect(["success", "failed", "skipped", "dry_run"]).toContain(e.status);
  143 |     }
  144 | 
  145 |     // 6. 更新 + 停用 + 删除
  146 |     const toggleRes = await request.post(`${base}/${rule.id}/toggle`, {
  147 |       headers: { Authorization: `Bearer ${token}` },
  148 |     });
  149 |     expect(toggleRes.ok()).toBe(true);
  150 |     const delRes = await request.delete(`${base}/${rule.id}`, {
  151 |       headers: { Authorization: `Bearer ${token}` },
  152 |     });
  153 |     expect(delRes.ok()).toBe(true);
  154 |   });
  155 | 
  156 |   test("并发触发：同规则多条执行记录落库（事件链路就绪时生效）", async ({ request }) => {
  157 |     test.setTimeout(60_000);
  158 |     const token = await login(request);
  159 |     const { wsId, projectId } = await wsAndProject(request, token);
  160 |     const base = `${apiURL}/workspaces/${wsId}/projects/${projectId}/automation`;
  161 | 
  162 |     // 创建"新缺陷通知创建者"规则
  163 |     const createRes = await request.post(`${base}`, {
  164 |       headers: { Authorization: `Bearer ${token}` },
  165 |       data: {
  166 |         name: `E2E-并发-${Date.now()}`,
  167 |         description: "并发竞争用例",
  168 |         project_id: projectId,
  169 |         status: "active",
  170 |         dsl: {
  171 |           trigger: { type: "issue.created", filter: { type_code: "defect" } },
  172 |           conditions: [],
  173 |           actions: [
  174 |             {
  175 |               type: "notify",
  176 |               config: {
  177 |                 channel: "in_app",
  178 |                 target: "${issue.created_by}",
  179 |                 template: "新缺陷 {{issue.identifier}} 已创建",
  180 |               },
  181 |             },
  182 |           ],
  183 |         },
  184 |       },
  185 |     });
> 186 |     expect(createRes.ok()).toBe(true);
      |                            ^ Error: expect(received).toBe(expected) // Object.is equality
  187 |     const rule = await createRes.json();
  188 | 
  189 |     // 并发创建 3 个缺陷（走 issue API；事件经 Outbox→RabbitMQ→自动化消费者）
  190 |     const issueBase = `${apiURL}/workspaces/${wsId}/projects/${projectId}/issues`;
  191 |     const created = await Promise.all(
  192 |       [1, 2, 3].map((i) =>
  193 |         request.post(`${issueBase}`, {
  194 |           headers: { Authorization: `Bearer ${token}` },
  195 |           data: {
  196 |             name: `E2E-并发缺陷-${Date.now()}-${i}`,
  197 |             type_code: "defect",
  198 |             priority: "high",
  199 |           },
  200 |         }),
  201 |       ),
  202 |     );
  203 |     const createdIds = created.map((r, i) => (r.ok() ? i : -1)).filter((i) => i >= 0);
  204 | 
  205 |     // 等待事件链路消费（Outbox relay → automation.evaluate），轮询执行记录
  206 |     let matched = 0;
  207 |     const deadline = Date.now() + 30_000;
  208 |     while (Date.now() < deadline) {
  209 |       const execRes = await request.get(`${base}/executions?limit=20`, {
  210 |         headers: { Authorization: `Bearer ${token}` },
  211 |       });
  212 |       if (execRes.ok()) {
  213 |         const body = await execRes.json();
  214 |         const execs = Array.isArray(body) ? body : body.items ?? [];
  215 |         matched = execs.filter((e: any) => e.rule_id === rule.id).length;
  216 |         if (matched >= createdIds.length) break;
  217 |       }
  218 |       await new Promise((r) => setTimeout(r, 2_000));
  219 |     }
  220 | 
  221 |     // 事件链路未启动（worker/RabbitMQ 缺失）时 matched=0：记录警告而非失败
  222 |     if (matched === 0) {
  223 |       console.log(
  224 |         "[skip] 事件链路（worker + RabbitMQ）未就绪，跳过执行记录断言",
  225 |       );
  226 |     } else {
  227 |       expect(matched).toBeGreaterThanOrEqual(1);
  228 |       expect(matched).toBeLessThanOrEqual(10); // 防循环保护生效（无爆炸式重复执行）
  229 |     }
  230 | 
  231 |     // 清理
  232 |     await request.delete(`${base}/${rule.id}`, {
  233 |       headers: { Authorization: `Bearer ${token}` },
  234 |     });
  235 |   });
  236 | });
  237 | 
```