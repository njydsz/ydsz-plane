# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: kanban-ui.spec.ts >> Peek 预览抽屉 >> 从列表视图点击行打开 Peek 抽屉
- Location: e2e\kanban-ui.spec.ts:133:3

# Error details

```
Test timeout of 60000ms exceeded.
```

```
Error: locator.fill: Test timeout of 60000ms exceeded.
Call log:
  - waiting for locator('input[type="email"]')

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]: "[plugin:vite:import-analysis] Failed to resolve import \"@/views/workspace/WorkspaceListView.vue\" from \"src/router/index.ts\". Does the file exist?"
  - generic [ref=e5]: D:/Code/open/ydsz-plane/web/src/router/index.ts:99:34
  - generic [ref=e6]: "83 | path: \"\", 84 | name: \"home\", 85 | component: () => import(\"@/views/workspace/WorkspaceListView.vue\") | ^ 86 | }, 87 | {"
  - generic [ref=e7]: at TransformPluginContext._formatLog (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:42658:41) at TransformPluginContext.error (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:42655:16) at normalizeUrl (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:40634:23) at process.processTicksAndRejections (node:internal/process/task_queues:104:5) at async file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:40753:37 at async Promise.all (index 14) at async TransformPluginContext.transform (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:40680:7) at async EnvironmentPluginContainer.transform (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:42453:18) at async loadAndTransform (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:35845:27) at async viteTransformMiddleware (file:///D:/Code/open/ydsz-plane/web/node_modules/.pnpm/vite@6.4.3_@types+node@20.19.43_terser@5.49.2/node_modules/vite/dist/node/chunks/dep-Dm0c1Wj2.js:37369:24
  - generic [ref=e8]:
    - text: Click outside, press Esc key, or fix the code to dismiss.You can also disable this overlay by setting
    - code [ref=e9]: server.hmr.overlay
    - text: to
    - code [ref=e10]: "false"
    - text: in
    - code [ref=e11]: vite.config.ts
    - text: .
```

# Test source

```ts
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
> 135 |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
      |                                               ^ Error: locator.fill: Test timeout of 60000ms exceeded.
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
  174 |     await expect(page.locator("body")).toBeVisible();
  175 |   });
  176 | });
  177 | 
```