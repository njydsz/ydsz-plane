# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notification.spec.ts >> 通知域 >> 全部已读 API 生效
- Location: e2e\notification.spec.ts:127:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Test source

```ts
  48  |     await page.goto("/login");
  49  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
  50  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  51  |     await page.locator("button.submit").click();
  52  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  53  | 
  54  |     // 导航到通知列表页
  55  |     const slug = await page.getAttribute('a[href*="/projects/"]', "href")
  56  |       ?.then(h => h?.match(/\/([^/]+)\/projects/)?.[1]);
  57  |     if (slug) {
  58  |       await page.goto(`/${slug}/notifications`);
  59  |     } else {
  60  |       await page.goto("/notifications");
  61  |     }
  62  |     await expect(page.locator("body")).toBeVisible({ timeout: 10_000 });
  63  | 
  64  |     // 通知列表或空态应出现在页面上
  65  |     const hasItems = await page.locator(".notif-item, [data-testid='notification-item'], .notification-row")
  66  |       .first()
  67  |       .isVisible({ timeout: 5_000 })
  68  |       .catch(() => false);
  69  |     const hasEmpty = await page.locator(
  70  |       "text=暂无通知, text=No notifications, [data-testid='empty-notifications']",
  71  |     ).first().isVisible({ timeout: 5_000 }).catch(() => false);
  72  |     expect(hasItems || hasEmpty).toBe(true);
  73  |   });
  74  | 
  75  |   test("标记单个通知为已读", async ({ page, request }) => {
  76  |     // 通过 API 登录并创建一条通知，然后验证标记已读
  77  |     const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  78  | 
  79  |     // 登录
  80  |     const loginRes = await request.post(`${apiURL}/auth/login`, {
  81  |       data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  82  |     });
  83  |     expect(loginRes.ok()).toBe(true);
  84  |     const { access_token: token } = await loginRes.json();
  85  | 
  86  |     // 获取工作空间
  87  |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  88  |       headers: { Authorization: `Bearer ${token}` },
  89  |     });
  90  |     expect(wsRes.ok()).toBe(true);
  91  |     const wsList = await wsRes.json();
  92  |     const wsId = wsList[0]?.id || 1;
  93  | 
  94  |     // 查询通知列表
  95  |     const notifRes = await request.get(`${apiURL}/workspaces/${wsId}/notifications?limit=1`, {
  96  |       headers: { Authorization: `Bearer ${token}` },
  97  |     });
  98  |     expect(notifRes.ok()).toBe(true);
  99  |     const notifBody = await notifRes.json();
  100 | 
  101 |     if (notifBody.items?.length > 0) {
  102 |       const notif = notifBody.items[0];
  103 |       expect(notif.id).toBeGreaterThan(0);
  104 | 
  105 |       // 标记已读
  106 |       const markReadRes = await request.put(
  107 |         `${apiURL}/workspaces/${wsId}/notifications/${notif.id}/read`,
  108 |         { headers: { Authorization: `Bearer ${token}` } },
  109 |       );
  110 |       expect(markReadRes.ok()).toBe(true);
  111 | 
  112 |       // 验证已读状态
  113 |       const afterRes = await request.get(
  114 |         `${apiURL}/workspaces/${wsId}/notifications?limit=10`,
  115 |         { headers: { Authorization: `Bearer ${token}` } },
  116 |       );
  117 |       const afterBody = await afterRes.json();
  118 |       const updated = afterBody.items?.find((n: any) => n.id === notif.id);
  119 |       if (updated) {
  120 |         expect(updated.is_read).toBe(true);
  121 |       }
  122 |     } else {
  123 |       console.log("No notifications in DB — skipping mark-read assertion");
  124 |     }
  125 |   });
  126 | 
  127 |   test("全部已读 API 生效", async ({ page, request }) => {
  128 |     const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  129 | 
  130 |     const loginRes = await request.post(`${apiURL}/auth/login`, {
  131 |       data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  132 |     });
  133 |     expect(loginRes.ok()).toBe(true);
  134 |     const { access_token: token } = await loginRes.json();
  135 | 
  136 |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  137 |       headers: { Authorization: `Bearer ${token}` },
  138 |     });
  139 |     expect(wsRes.ok()).toBe(true);
  140 |     const wsList = await wsRes.json();
  141 |     const wsId = wsList[0]?.id || 1;
  142 | 
  143 |     // 调用全部已读
  144 |     const markAllRes = await request.put(
  145 |       `${apiURL}/workspaces/${wsId}/notifications/read-all`,
  146 |       { headers: { Authorization: `Bearer ${token}` } },
  147 |     );
> 148 |     expect(markAllRes.ok()).toBe(true);
      |                             ^ Error: expect(received).toBe(expected) // Object.is equality
  149 |     const body = await markAllRes.json();
  150 |     expect(body.ok).toBe(true);
  151 | 
  152 |     // 验证未读计数归零
  153 |     const countRes = await request.get(
  154 |       `${apiURL}/workspaces/${wsId}/notifications/unread-count`,
  155 |       { headers: { Authorization: `Bearer ${token}` } },
  156 |     );
  157 |     expect(countRes.ok()).toBe(true);
  158 |     const countBody = await countRes.json();
  159 |     expect(countBody.count).toBe(0);
  160 |   });
  161 | 
  162 |   test("未读计数 API 正确返回", async ({ request }) => {
  163 |     const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  164 | 
  165 |     const loginRes = await request.post(`${apiURL}/auth/login`, {
  166 |       data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  167 |     });
  168 |     expect(loginRes.ok()).toBe(true);
  169 |     const { access_token: token } = await loginRes.json();
  170 | 
  171 |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  172 |       headers: { Authorization: `Bearer ${token}` },
  173 |     });
  174 |     expect(wsRes.ok()).toBe(true);
  175 |     const wsList = await wsRes.json();
  176 |     const wsId = wsList[0]?.id || 1;
  177 | 
  178 |     const countRes = await request.get(
  179 |       `${apiURL}/workspaces/${wsId}/notifications/unread-count`,
  180 |       { headers: { Authorization: `Bearer ${token}` } },
  181 |     );
  182 |     expect(countRes.ok()).toBe(true);
  183 |     const body = await countRes.json();
  184 |     expect(body).toHaveProperty("count");
  185 |     expect(typeof body.count).toBe("number");
  186 |     expect(body.count).toBeGreaterThanOrEqual(0);
  187 |   });
  188 | 
  189 |   test("归档通知后不再出现在列表", async ({ request }) => {
  190 |     const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  191 | 
  192 |     const loginRes = await request.post(`${apiURL}/auth/login`, {
  193 |       data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  194 |     });
  195 |     expect(loginRes.ok()).toBe(true);
  196 |     const { access_token: token } = await loginRes.json();
  197 | 
  198 |     const wsRes = await request.get(`${apiURL}/workspaces`, {
  199 |       headers: { Authorization: `Bearer ${token}` },
  200 |     });
  201 |     expect(wsRes.ok()).toBe(true);
  202 |     const wsList = await wsRes.json();
  203 |     const wsId = wsList[0]?.id || 1;
  204 | 
  205 |     const notifRes = await request.get(`${apiURL}/workspaces/${wsId}/notifications?limit=1`, {
  206 |       headers: { Authorization: `Bearer ${token}` },
  207 |     });
  208 |     const notifBody = await notifRes.json();
  209 |     if (notifBody.items?.length > 0) {
  210 |       const notif = notifBody.items[0];
  211 |       // 归档
  212 |       const archiveRes = await request.put(
  213 |         `${apiURL}/workspaces/${wsId}/notifications/${notif.id}/archive`,
  214 |         { headers: { Authorization: `Bearer ${token}` } },
  215 |       );
  216 |       expect(archiveRes.ok()).toBe(true);
  217 | 
  218 |       // 确认默认列表不再包含已归档项
  219 |       const afterRes = await request.get(
  220 |         `${apiURL}/workspaces/${wsId}/notifications?limit=50`,
  221 |         { headers: { Authorization: `Bearer ${token}` } },
  222 |       );
  223 |       const afterBody = await afterRes.json();
  224 |       const stillVisible = afterBody.items?.find((n: any) => n.id === notif.id);
  225 |       expect(stillVisible).toBeUndefined();
  226 |     }
  227 |   });
  228 | });
  229 | 
```