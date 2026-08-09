# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: auth.spec.ts >> 鉴权链路 >> 错误凭证时展示错误提示
- Location: e2e\auth.spec.ts:10:3

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
  1  | /**
  2  |  * 鉴权 E2E 冒烟链路。
  3  |  *
  4  |  * 覆盖：登录成功跳转、错误凭证提示。
  5  |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  6  |  */
  7  | import { expect, test } from "@playwright/test";
  8  | 
  9  | test.describe("鉴权链路", () => {
  10 |   test("错误凭证时展示错误提示", async ({ page }) => {
  11 |     await page.goto("/login");
> 12 |     await page.locator('input[type="email"]').fill("admin@ydsz.dev");
     |                                               ^ Error: locator.fill: Test timeout of 60000ms exceeded.
  13 |     await page.locator('input[type="password"]').fill("wrong-password-123");
  14 |     await page.locator("button.submit").click();
  15 | 
  16 |     // 登录失败不应跳转，应停留在登录页并展示错误
  17 |     await expect(page.locator("form.login-card")).toBeVisible();
  18 |     await expect(page.locator(".error")).toBeVisible({ timeout: 15_000 });
  19 |   });
  20 | 
  21 |   test("有效凭证登录成功并进入工作区列表", async ({ page }) => {
  22 |     await page.goto("/login");
  23 |     await page.locator('input[type="email"]').fill("admin@ydsz.dev");
  24 |     await page.locator('input[type="password"]').fill("Admin@123");
  25 |     await page.locator("button.submit").click();
  26 | 
  27 |     // 登录后应离开登录页（URL 变化），进入应用主体
  28 |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  29 |     await expect(page.locator("body")).toBeVisible();
  30 |   });
  31 | });
  32 | 
```