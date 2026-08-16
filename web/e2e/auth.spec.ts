/**
 * 鉴权 E2E 冒烟链路。
 *
 * 覆盖：登录成功跳转、错误凭证提示。
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 */
import { expect, test } from "@playwright/test";

test.describe("鉴权链路", () => {
  test("错误凭证时展示错误提示", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill("admin@njydsz.com");
    await page.locator('input[type="password"]').fill("wrong-password-123");
    await page.locator("button.submit").click();

    // 登录失败不应跳转，应停留在登录页并展示错误
    await expect(page.locator("form.login-card")).toBeVisible();
    await expect(page.locator(".error")).toBeVisible({ timeout: 15_000 });
  });

  test("有效凭证登录成功并进入工作区列表", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill("admin@njydsz.com");
    await page.locator('input[type="password"]').fill("Admin@1020");
    await page.locator("button.submit").click();

    // 登录后应离开登录页（URL 变化），进入应用主体
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
    await expect(page.locator("body")).toBeVisible();
  });
});
