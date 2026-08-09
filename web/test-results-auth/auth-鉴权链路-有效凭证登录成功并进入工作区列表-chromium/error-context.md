# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: auth.spec.ts >> 鉴权链路 >> 有效凭证登录成功并进入工作区列表
- Location: e2e\auth.spec.ts:21:3

# Error details

```
Error: expect(page).not.toHaveURL(expected) failed

Expected pattern: not /\/login/
Received string: "http://localhost:5173/login"
Timeout: 15000ms

Call log:
  - Expect "not toHaveURL" with timeout 15000ms
    33 × locator resolved to <html lang="zh-CN" data-theme="light">…</html>
       - unexpected value "http://localhost:5173/login"

```

```yaml
- text: YD
- heading "Ydsz Plane" [level=1]
- paragraph: 面向中国软件团队的开源项目管理平台
- text: 邮箱
- textbox "邮箱":
  - /placeholder: you@example.com
- text: 密码
- textbox "密码":
  - /placeholder: 至少 8 位
- button "登录"
- link "注册账号":
  - /url: /register
- link "忘记密码？":
  - /url: /forgot-password
- button "🇺🇸 English":
  - text: 🇺🇸 English
  - img
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
  12 |     await page.locator('input[type="email"]').fill("admin@ydsz.dev");
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
> 28 |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
     |                            ^ Error: expect(page).not.toHaveURL(expected) failed
  29 |     await expect(page.locator("body")).toBeVisible();
  30 |   });
  31 | });
  32 | 
```