# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: kanban-ui.spec.ts >> 看板 UI >> 看板页面加载并显示状态列
- Location: e2e\kanban-ui.spec.ts:25:3

# Error details

```
Test timeout of 60000ms exceeded while running "beforeEach" hook.
```

```
Error: locator.fill: Test timeout of 60000ms exceeded.
Call log:
  - waiting for locator('input[type="email"]')
    - waiting for "http://localhost:5173/login" navigation to finish...
    - navigated to "http://localhost:5173/login"

```

# Test source

```ts
  1   | /**
  2   |  * 看板 UI 交互 E2E 测试。
  3   |  *
  4   |  * 覆盖：看板加载状态列、工作项卡片展示、状态流转 UI 交互、
  5   |  *       内联编辑优先级/指派人、列过滤。
  6   |  * 回归 S4 前端视图就绪度（API 层已在 issue-workflow.spec.ts 覆盖，此处聚焦 UI）。
  7   |  *
  8   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  9   |  */
  10  | import { expect, test } from "@playwright/test";
  11  | 
  12  | const TEST_EMAIL = "admin@ydsz.dev";
  13  | const TEST_PASSWORD = "Admin@123";
  14  | 
  15  | test.describe("看板 UI", () => {
  16  |   test.beforeEach(async ({ page }) => {
  17  |     // 登录
  18  |     await page.goto("/login");
> 19  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
      |                                               ^ Error: locator.fill: Test timeout of 60000ms exceeded.
  20  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  21  |     await page.locator("button.submit").click();
  22  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  23  |   });
  24  | 
  25  |   test("看板页面加载并显示状态列", async ({ page }) => {
  26  |     // 导航到第一个项目的看板
  27  |     await page.goto("/"); // 根据路由配置，重定向到默认项目看板
  28  |     await page.waitForLoadState("domcontentloaded");
  29  | 
  30  |     // 看板加载后应看到列头（状态名）或空态提示
  31  |     const columnHeaders = page.locator(".kanban-column__header, .board-column__header");
  32  |     const emptyState = page.locator(
  33  |       ".app-empty-state, .empty-state, [class*=empty]",
  34  |     );
  35  | 
  36  |     // 至少列头或空态之一可见
  37  |     const hasColumns = await columnHeaders.count().then((c) => c > 0);
  38  |     const hasEmpty = await emptyState.count().then((c) => c > 0);
  39  |     expect(hasColumns || hasEmpty).toBe(true);
  40  |   });
  41  | 
  42  |   test("看板中存在工作项卡片或空态", async ({ page }) => {
  43  |     await page.goto("/");
  44  |     await page.waitForLoadState("domcontentloaded");
  45  | 
  46  |     // 工作项卡片 class 因实现而异，使用通用选择器
  47  |     const cards = page.locator(
  48  |       "[class*=issue-card], [class*=card], [class*=kanban-card]",
  49  |     );
  50  |     const emptyState = page.locator(
  51  |       ".app-empty-state, .empty-state, [class*=empty]",
  52  |     );
  53  | 
  54  |     const hasCards = await cards.count().then((c) => c > 0);
  55  |     const hasEmpty = await emptyState.count().then((c) => c > 0);
  56  |     expect(hasCards || hasEmpty).toBe(true);
  57  |   });
  58  | 
  59  |   test("视图切换器可在看板与列表间切换", async ({ page }) => {
  60  |     await page.goto("/");
  61  |     await page.waitForLoadState("domcontentloaded");
  62  | 
  63  |     // 查找视图切换 tab
  64  |     const viewTab = page.locator(
  65  |       ".view-tab, [class*=view-switcher] a, [class*=view-tab]",
  66  |     );
  67  |     if ((await viewTab.count()) >= 2) {
  68  |       // 切换到列表视图
  69  |       const listTab = viewTab.filter({ hasText: /列表/ });
  70  |       if ((await listTab.count()) > 0) {
  71  |         await listTab.first().click();
  72  |         await page.waitForLoadState("domcontentloaded");
  73  |         // URL 或页面内容应变化
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
```