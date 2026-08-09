# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: intake.spec.ts >> 收件箱（Intake）域 >> 公开提交 → 审核 → 转正 → 跟踪闭环
- Location: e2e\intake.spec.ts:30:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  1   | /**
  2   |  * 收件箱（Intake）域 E2E 测试（S10 出口：门户提交→审核→转正→跟踪闭环）。
  3   |  *
  4   |  * 覆盖完整业务闭环：
  5   |  *  1. 管理员创建公开频道（slug 唯一）
  6   |  *  2. 外部用户免登录经公开门户提交工单 → 获得 tracking_id
  7   |  *  3. 管理员审核（accepted）→ 转正（convert）→ 生成正式工作项并双向关联
  8   |  *  4. 提交者凭 tracking_id + email 跟踪处理状态（脱敏只读视图）
  9   |  *  5. 拒绝路径：rejected 后状态同步
  10  |  *
  11  |  * 全链路 API 驱动（公开接口免登录、管理接口走 Bearer），不依赖 UI 渲染。
  12  |  * 运行前提：后端已启动（make migrate && make seed）。
  13  |  */
  14  | import { expect, test } from "@playwright/test";
  15  | 
  16  | const TEST_EMAIL = "admin@ydsz.dev";
  17  | const TEST_PASSWORD = "Admin@123";
  18  | const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  19  | 
  20  | async function login(request: any) {
  21  |   const res = await request.post(`${apiURL}/auth/login`, {
  22  |     data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  23  |   });
  24  |   expect(res.ok()).toBe(true);
  25  |   const { access_token: token } = await res.json();
  26  |   return token as string;
  27  | }
  28  | 
  29  | test.describe("收件箱（Intake）域", () => {
  30  |   test("公开提交 → 审核 → 转正 → 跟踪闭环", async ({ request }) => {
  31  |     const token = await login(request);
  32  |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  33  |       headers: { Authorization: `Bearer ${token}` },
  34  |     });
  35  |     expect(wsRes.ok()).toBe(true);
  36  |     const wsList = await wsRes.json();
  37  |     const wsId = wsList[0]?.id ?? 1;
  38  |     const wsSlug = wsList[0]?.slug ?? "acme";
  39  | 
  40  |     // 取项目（转正目标项目）
  41  |     const projRes = await request.get(
  42  |       `${apiURL}/workspaces/${wsId}/projects?limit=5`,
  43  |       { headers: { Authorization: `Bearer ${token}` } },
  44  |     );
  45  |     const projList = projRes.ok() ? await projRes.json() : { items: [] };
  46  |     const items = Array.isArray(projList) ? projList : projList.items ?? [];
  47  |     const projectId = items[0]?.id ?? 1;
  48  | 
  49  |     const adminBase = `${apiURL}/workspaces/${wsId}/intake`;
  50  |     const slug = `e2e-${Date.now().toString(36)}`;
  51  |     const stamp = Date.now();
  52  | 
  53  |     // 1. 创建公开频道
  54  |     const chRes = await request.post(`${adminBase}/channels`, {
  55  |       headers: { Authorization: `Bearer ${token}` },
  56  |       data: {
  57  |         slug,
  58  |         name: `E2E 频道 ${stamp}`,
  59  |         description: "Playwright 创建",
  60  |         is_public: true,
  61  |         default_issue_type: "requirement",
  62  |         rate_limit_per_min: 50,
  63  |         require_captcha: false,
  64  |       },
  65  |     });
> 66  |     expect(chRes.ok()).toBe(true);
      |                        ^ Error: expect(received).toBe(expected) // Object.is equality
  67  |     const channel = await chRes.json();
  68  |     expect(channel.id).toBeGreaterThan(0);
  69  | 
  70  |     // 2. 外部用户免登录公开提交
  71  |     const submitRes = await request.post(
  72  |       `${apiURL}/public/intake/channels/${wsSlug}/${slug}/submit`,
  73  |       {
  74  |         data: {
  75  |           title: `E2E 门户工单 ${stamp}`,
  76  |           description: "来自公开门户的缺陷/需求描述",
  77  |           submitter_name: "张三",
  78  |           submitter_email: "zhangsan@example.com",
  79  |           issue_type: "requirement",
  80  |         },
  81  |       },
  82  |     );
  83  |     expect(submitRes.ok()).toBe(true);
  84  |     const submitted = await submitRes.json();
  85  |     expect(submitted.tracking_id).toBeTruthy();
  86  |     expect(submitted.status).toBe("open");
  87  |     const trackingId = submitted.tracking_id;
  88  | 
  89  |     // 3. 管理员查询收件队列并定位工单
  90  |     const queueRes = await request.get(`${adminBase}/issues?status=open&limit=20`, {
  91  |       headers: { Authorization: `Bearer ${token}` },
  92  |     });
  93  |     expect(queueRes.ok()).toBe(true);
  94  |     const queueBody = await queueRes.json();
  95  |     const queue = Array.isArray(queueBody) ? queueBody : queueBody.items ?? [];
  96  |     const ticket = queue.find((t: any) => t.tracking_id === trackingId);
  97  |     expect(ticket).toBeTruthy();
  98  |     const ticketId = ticket.id;
  99  | 
  100 |     // 4. 审核通过（accepted）
  101 |     const reviewRes = await request.post(`${adminBase}/issues/${ticketId}/review`, {
  102 |       headers: { Authorization: `Bearer ${token}` },
  103 |       data: {
  104 |         action: "accepted",
  105 |         target_issue_type: "requirement",
  106 |         target_project_id: projectId,
  107 |       },
  108 |     });
  109 |     expect(reviewRes.ok()).toBe(true);
  110 | 
  111 |     // 5. 转正 → 生成正式工作项并双向关联
  112 |     const convertRes = await request.post(`${adminBase}/issues/${ticketId}/convert`, {
  113 |       headers: { Authorization: `Bearer ${token}` },
  114 |       data: {
  115 |         target_project_id: projectId,
  116 |         target_issue_type: "requirement",
  117 |       },
  118 |     });
  119 |     expect(convertRes.ok()).toBe(true);
  120 |     const converted = await convertRes.json();
  121 |     expect(converted.converted_issue_id).toBeGreaterThan(0);
  122 |     const convertedIssueId = converted.converted_issue_id;
  123 | 
  124 |     // 6. 提交者跟踪门户（tracking_id + email）→ 应展示脱敏只读视图
  125 |     const trackRes = await request.get(
  126 |       `${apiURL}/public/intake/track?tracking_id=${trackingId}&email=zhangsan@example.com`,
  127 |     );
  128 |     expect(trackRes.ok()).toBe(true);
  129 |     const view = await trackRes.json();
  130 |     expect(view.status).toMatch(/accepted|converted/);
  131 |     expect(view.tracking_id).toBe(trackingId);
  132 |     // 脱敏：不应泄露内部字段（submitter_user_id 等）
  133 |     expect(view.submitter_user_id).toBeUndefined();
  134 | 
  135 |     // 7. 转正后正式工作项存在（管理员视角）
  136 |     const issueRes = await request.get(
  137 |       `${apiURL}/workspaces/${wsId}/projects/${projectId}/issues/${convertedIssueId}`,
  138 |       { headers: { Authorization: `Bearer ${token}` } },
  139 |     );
  140 |     expect(issueRes.ok()).toBe(true);
  141 | 
  142 |     // 8. 拒绝路径：新建一条并拒绝
  143 |     const submit2 = await request.post(
  144 |       `${apiURL}/public/intake/channels/${wsSlug}/${slug}/submit`,
  145 |       {
  146 |         data: {
  147 |           title: `E2E 拒绝工单 ${stamp}`,
  148 |           submitter_name: "李四",
  149 |           submitter_email: "lisi@example.com",
  150 |         },
  151 |       },
  152 |     );
  153 |     expect(submit2.ok()).toBe(true);
  154 |     const t2 = await submit2.json();
  155 |     const q2Res = await request.get(`${adminBase}/issues?status=open&limit=50`, {
  156 |       headers: { Authorization: `Bearer ${token}` },
  157 |     });
  158 |     const q2Body = await q2Res.json();
  159 |     const q2 = Array.isArray(q2Body) ? q2Body : q2Body.items ?? [];
  160 |     const t2row = q2.find((t: any) => t.tracking_id === t2.tracking_id);
  161 |     expect(t2row).toBeTruthy();
  162 |     const rejectRes = await request.post(`${adminBase}/issues/${t2row.id}/review`, {
  163 |       headers: { Authorization: `Bearer ${token}` },
  164 |       data: { action: "rejected", reason: "重复提交，无需处理" },
  165 |     });
  166 |     expect(rejectRes.ok()).toBe(true);
```