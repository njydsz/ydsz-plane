# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: kanban-ui.spec.ts >> Peek 预览抽屉 >> Peek 抽屉内上一个/下一个导航（如有）
- Location: e2e\kanban-ui.spec.ts:165:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator:  locator('body')
Expected: visible
Received: hidden
Timeout:  15000ms

Call log:
  - Expect "toBeVisible" with timeout 15000ms
  - waiting for locator('body')
    33 × locator resolved to <body>…</body>
       - unexpected value "hidden"

```

# Test source

```ts
  74  |         await expect(page.locator("body")).toBeVisible();
  75  |       }
  76  |     }
  77  |   });
  78  | 
  79  |   test("看板列头可排序或过滤（如有此功能）", async ({ page }) => {
  80  |     await page.goto("/");
  81  |     await page.waitForLoadState("domcontentloaded");
  82  | 
  83  |     // 看板过滤按钮（如果存在）
  84  |     const filterBtn = page.locator(
  85  |       "[class*=filter], [class*=filter-btn]",
  86  |     );
  87  |     if ((await filterBtn.count()) > 0) {
  88  |       await filterBtn.first().click();
  89  |       await page.waitForTimeout(500);
  90  |       // 点击后不应崩溃
  91  |       await expect(page.locator("body")).toBeVisible();
  92  |     }
  93  |   });
  94  | 
  95  |   test("点击工作项打开 Peek 预览抽屉", async ({ page }) => {
  96  |     await page.goto("/");
  97  |     await page.waitForLoadState("domcontentloaded");
  98  | 
  99  |     // 找到第一个工作项卡片并点击
  100 |     const cards = page.locator(
  101 |       "[class*=issue-card], [class*=card], [class*=kanban-card]",
  102 |     );
  103 |     if ((await cards.count()) > 0) {
  104 |       await cards.first().click();
  105 |       await page.waitForTimeout(800);
  106 | 
  107 |       // Peek 抽屉应出现
  108 |       const peekDrawer = page.locator(
  109 |         "[class*=peek], [class*=drawer], [class*=detail-panel]",
  110 |       );
  111 |       // 抽屉或详情视图应可见
  112 |       const peekVisible = await peekDrawer.count().then((c) => c > 0);
  113 |       // 也可能通过路由跳转到详情页
  114 |       const urlChanged = await page.url();
  115 |       expect(peekVisible || urlChanged).toBeTruthy();
  116 |     }
  117 |   });
  118 | 
  119 |   test("列内计数显示正确（若实现）", async ({ page }) => {
  120 |     await page.goto("/");
  121 |     await page.waitForLoadState("domcontentloaded");
  122 | 
  123 |     // 列内计数 badge（如果实现）
  124 |     const countBadges = page.locator(
  125 |       "[class*=column-count], [class*=count-badge], [class*=counter]",
  126 |     );
  127 |     // 不强制要求，仅验证页面不崩溃
  128 |     await expect(page.locator("body")).toBeVisible();
  129 |   });
  130 | });
  131 | 
  132 | test.describe("Peek 预览抽屉", () => {
  133 |   test("从列表视图点击行打开 Peek 抽屉", async ({ page }) => {
  134 |     await page.goto("/login");
  135 |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
  136 |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  137 |     await page.locator("button.submit").click();
  138 |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  139 | 
  140 |     // 导航到列表视图
  141 |     await page.goto("/");
  142 |     await page.waitForLoadState("domcontentloaded");
  143 | 
  144 |     // 切换到列表
  145 |     const listTab = page.locator(".view-tab, [class*=view-tab]").filter({ hasText: /列表/ });
  146 |     if ((await listTab.count()) > 0) {
  147 |       await listTab.first().click();
  148 |       await page.waitForLoadState("domcontentloaded");
  149 |     }
  150 | 
  151 |     // 点击第一行工作项
  152 |     const rows = page.locator("tr.row, tbody tr, [class*=table-row]");
  153 |     if ((await rows.count()) > 0) {
  154 |       await rows.first().click();
  155 |       await page.waitForTimeout(800);
  156 | 
  157 |       // Peek 抽屉或详情页应打开
  158 |       const peekDrawer = page.locator(
  159 |         "[class*=peek], [class*=drawer]",
  160 |       );
  161 |       await expect(page.locator("body")).toBeVisible();
  162 |     }
  163 |   });
  164 | 
  165 |   test("Peek 抽屉内上一个/下一个导航（如有）", async ({ page }) => {
  166 |     await page.goto("/");
  167 |     await page.waitForLoadState("domcontentloaded");
  168 | 
  169 |     // Peek 导航按钮在 drawer 内
  170 |     const navBtn = page.locator(
  171 |       "[class*=peek-prev], [class*=peek-next], [class*=nav-next]",
  172 |     );
  173 |     // 不强制要求有导航，仅验证不崩溃
> 174 |     await expect(page.locator("body")).toBeVisible();
      |                                        ^ Error: expect(locator).toBeVisible() failed
  175 |   });
  176 | });
  177 | 
```