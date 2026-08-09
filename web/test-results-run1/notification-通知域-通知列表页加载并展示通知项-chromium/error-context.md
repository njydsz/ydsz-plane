# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notification.spec.ts >> 通知域 >> 通知列表页加载并展示通知项
- Location: e2e\notification.spec.ts:47:3

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: locator.fill: Test timeout of 60000ms exceeded.
Call log:
  - waiting for locator('input[type="email"]')

```

# Test source

```ts
  1   | /**
  2   |  * 通知域 E2E 测试。
  3   |  *
  4   |  * 覆盖：未读计数、通知列表、标记已读、全部已读、跳转工作项。
  5   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  6   |  */
  7   | import { expect, test } from "@playwright/test";
  8   | 
  9   | const TEST_EMAIL = "admin@ydsz.dev";
  10  | const TEST_PASSWORD = "Admin@123";
  11  | 
  12  | test.describe("通知域", () => {
  13  |   test("通知铃铛展示未读计数", async ({ page }) => {
  14  |     await page.goto("/login");
  15  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
  16  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  17  |     await page.locator("button.submit").click();
  18  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  19  | 
  20  |     // 等待侧边栏/铃铛渲染
  21  |     const bellBtn = page.locator(
  22  |       '[data-testid="notification-bell"], .notification-bell .bell-btn, button[aria-label="通知"]',
  23  |     );
  24  |     await expect(bellBtn.first()).toBeVisible({ timeout: 15_000 });
  25  |   });
  26  | 
  27  |   test("点击铃铛展开通知下拉面板", async ({ page }) => {
  28  |     await page.goto("/login");
  29  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
  30  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  31  |     await page.locator("button.submit").click();
  32  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  33  | 
  34  |     const bellBtn = page.locator(
  35  |       '[data-testid="notification-bell"], .notification-bell .bell-btn, button[aria-label="通知"]',
  36  |     );
  37  |     await expect(bellBtn.first()).toBeVisible({ timeout: 15_000 });
  38  |     await bellBtn.first().click();
  39  | 
  40  |     // 下拉面板出现
  41  |     const dropdown = page.locator(
  42  |       '[data-testid="notification-dropdown"], .notification-bell .dropdown',
  43  |     );
  44  |     await expect(dropdown.first()).toBeVisible({ timeout: 10_000 });
  45  |   });
  46  | 
  47  |   test("通知列表页加载并展示通知项", async ({ page }) => {
  48  |     await page.goto("/login");
> 49  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
      |                                               ^ Error: locator.fill: Test timeout of 60000ms exceeded.
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
  148 |     expect(markAllRes.ok()).toBe(true);
  149 |     const body = await markAllRes.json();
```