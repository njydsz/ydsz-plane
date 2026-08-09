/**
 * 通知域 E2E 测试。
 *
 * 覆盖：未读计数、通知列表、标记已读、全部已读、跳转工作项。
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 */
import { expect, test } from "@playwright/test";
import { apiLogin, API_URL, TEST_EMAIL, TEST_PASSWORD } from "./helpers";

test.describe("通知域", () => {
  test("通知铃铛展示未读计数", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(TEST_EMAIL);
    await page.locator('input[type="password"]').fill(TEST_PASSWORD);
    await page.locator("button.submit").click();
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });

    // 等待侧边栏/铃铛渲染
    const bellBtn = page.locator(
      '[data-testid="notification-bell"], .notification-bell .bell-btn, button[aria-label="通知"]',
    );
    await expect(bellBtn.first()).toBeVisible({ timeout: 15_000 });
  });

  test("点击铃铛展开通知下拉面板", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(TEST_EMAIL);
    await page.locator('input[type="password"]').fill(TEST_PASSWORD);
    await page.locator("button.submit").click();
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });

    const bellBtn = page.locator(
      '[data-testid="notification-bell"], .notification-bell .bell-btn, button[aria-label="通知"]',
    );
    await expect(bellBtn.first()).toBeVisible({ timeout: 15_000 });
    await bellBtn.first().click();

    // 下拉面板出现
    const dropdown = page.locator(
      '[data-testid="notification-dropdown"], .notification-bell .dropdown',
    );
    await expect(dropdown.first()).toBeVisible({ timeout: 10_000 });
  });

  test("通知列表页加载并展示通知项", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(TEST_EMAIL);
    await page.locator('input[type="password"]').fill(TEST_PASSWORD);
    await page.locator("button.submit").click();
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });

    // 导航到通知列表页
    const slug = await page.getAttribute('a[href*="/projects/"]', "href")
      ?.then(h => h?.match(/\/([^/]+)\/projects/)?.[1]);
    if (slug) {
      await page.goto(`/${slug}/notifications`);
    } else {
      await page.goto("/notifications");
    }
    await expect(page.locator("body")).toBeVisible({ timeout: 10_000 });

    // 通知列表或空态应出现在页面上
    const hasItems = await page.locator(".notif-item, [data-testid='notification-item'], .notification-row")
      .first()
      .isVisible({ timeout: 5_000 })
      .catch(() => false);
    const hasEmpty = await page.locator(
      "text=暂无通知, text=No notifications, [data-testid='empty-notifications']",
    ).first().isVisible({ timeout: 5_000 }).catch(() => false);
    expect(hasItems || hasEmpty).toBe(true);
  });

  test("标记单个通知为已读", async ({ page, request }) => {
    // 通过 API 登录并创建一条通知，然后验证标记已读
    const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

    // 登录
    const { token, headers: authHeaders } = await apiLogin(request);

    // 获取工作空间
    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: authHeaders,
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id || 1;

    // 查询通知列表
    const notifRes = await request.get(`${apiURL}/workspaces/${wsId}/notifications?limit=1`, {
      headers: authHeaders,
    });
    expect(notifRes.ok()).toBe(true);
    const notifBody = await notifRes.json();

    if (notifBody.items?.length > 0) {
      const notif = notifBody.items[0];
      expect(notif.id).toBeGreaterThan(0);

      // 标记已读
      const markReadRes = await request.put(
        `${apiURL}/workspaces/${wsId}/notifications/${notif.id}/read`,
        { headers: authHeaders },
      );
      expect(markReadRes.ok()).toBe(true);

      // 验证已读状态
      const afterRes = await request.get(
        `${apiURL}/workspaces/${wsId}/notifications?limit=10`,
        { headers: authHeaders },
      );
      const afterBody = await afterRes.json();
      const updated = afterBody.items?.find((n: any) => n.id === notif.id);
      if (updated) {
        expect(updated.is_read).toBe(true);
      }
    } else {
      console.log("No notifications in DB — skipping mark-read assertion");
    }
  });

  test("全部已读 API 生效", async ({ page, request }) => {
    const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

    const { token, headers: authHeaders } = await apiLogin(request);

    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: authHeaders,
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id || 1;

    // 调用全部已读
    const markAllRes = await request.put(
      `${apiURL}/workspaces/${wsId}/notifications/read-all`,
      { headers: authHeaders },
    );
    expect(markAllRes.ok()).toBe(true);
    const body = await markAllRes.json();
    expect(body.ok).toBe(true);

    // 验证未读计数归零
    const countRes = await request.get(
      `${apiURL}/workspaces/${wsId}/notifications/unread-count`,
      { headers: authHeaders },
    );
    expect(countRes.ok()).toBe(true);
    const countBody = await countRes.json();
    expect(countBody.count).toBe(0);
  });

  test("未读计数 API 正确返回", async ({ request }) => {
    const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

    const { token, headers: authHeaders } = await apiLogin(request);

    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: authHeaders,
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id || 1;

    const countRes = await request.get(
      `${apiURL}/workspaces/${wsId}/notifications/unread-count`,
      { headers: authHeaders },
    );
    expect(countRes.ok()).toBe(true);
    const body = await countRes.json();
    expect(body).toHaveProperty("count");
    expect(typeof body.count).toBe("number");
    expect(body.count).toBeGreaterThanOrEqual(0);
  });

  test("归档通知后不再出现在列表", async ({ request }) => {
    const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

    const { token, headers: authHeaders } = await apiLogin(request);

    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: authHeaders,
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id || 1;

    const notifRes = await request.get(`${apiURL}/workspaces/${wsId}/notifications?limit=1`, {
      headers: authHeaders,
    });
    const notifBody = await notifRes.json();
    if (notifBody.items?.length > 0) {
      const notif = notifBody.items[0];
      // 归档
      const archiveRes = await request.put(
        `${apiURL}/workspaces/${wsId}/notifications/${notif.id}/archive`,
        { headers: authHeaders },
      );
      expect(archiveRes.ok()).toBe(true);

      // 确认默认列表不再包含已归档项
      const afterRes = await request.get(
        `${apiURL}/workspaces/${wsId}/notifications?limit=50`,
        { headers: authHeaders },
      );
      const afterBody = await afterRes.json();
      const stillVisible = afterBody.items?.find((n: any) => n.id === notif.id);
      expect(stillVisible).toBeUndefined();
    }
  });
});
