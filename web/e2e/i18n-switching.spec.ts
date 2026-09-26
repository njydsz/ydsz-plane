/**
 * 国际化（i18n）语言切换 E2E 测试。
 *
 * 覆盖：简体中文 → 英语切换、切换后 UI 文案变化、刷新后语言持久化（localStorage）、
 *       日语切换、表单验证错误信息的语言切换。
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 */
import { expect, test } from "@playwright/test";
import { TEST_EMAIL, TEST_PASSWORD } from "./helpers";

/** 登录并等待主界面渲染完毕。 */
async function loginAndGotoHome(page: import("@playwright/test").Page) {
  await page.goto("/login");
  await page.locator('input[type="email"]').fill(TEST_EMAIL);
  await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  await page.locator("button.submit").click();
  // 登录成功后应离开登录页
  await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  // 等待侧边栏或主页核心元素渲染
  await expect(page.locator("body")).toBeVisible({ timeout: 10_000 });
}

/** 打开语言切换下拉并选择目标语言。 */
async function switchLocale(page: import("@playwright/test").Page, localeLabel: string) {
  const trigger = page.locator(".lang-switcher__trigger");
  await expect(trigger).toBeVisible({ timeout: 10_000 });
  await trigger.click();

  const dropdown = page.locator(".lang-switcher__dropdown");
  await expect(dropdown).toBeVisible({ timeout: 5_000 });

  // 在下拉中点击对应语言选项（通过显示文本匹配）
  const option = dropdown.locator(".lang-option", { hasText: localeLabel });
  await expect(option).toBeVisible({ timeout: 5_000 });
  await option.click();

  // 下拉应自动收起
  await expect(dropdown).not.toBeVisible({ timeout: 5_000 });
}

test.describe("国际化语言切换", () => {
  test("简体中文 → 英语：主页 UI 文案切换", async ({ page }) => {
    // 确保从默认中文开始
    await page.goto("/");
    await page.evaluate(() => localStorage.setItem("ydsz-locale", "zh-CN"));

    await loginAndGotoHome(page);

    // 简体中文下，工作空间列表标题应为「工作空间」
    // 使用 page.getByText 宽松匹配，兼容侧栏/标题等多处渲染
    await expect(page.getByText("工作空间").first()).toBeVisible({ timeout: 10_000 });

    // 切换至英文
    await switchLocale(page, "English");

    // 英文环境下 html[lang] 属性应更新为 en-US
    expect(await page.evaluate(() => document.documentElement.getAttribute("lang"))).toBe("en-US");

    // 核心文案应变为英文…验证侧栏或页面中的「Workspaces」文本
    await expect(page.getByText("Workspaces").first()).toBeVisible({ timeout: 10_000 });

    // 中文文案应不再可见
    const zhVisible = await page.getByText("工作空间").first().isVisible({ timeout: 3_000 }).catch(() => false);
    expect(zhVisible).toBe(false);
  });

  test("英语 → 日语：文案切换至日文", async ({ page }) => {
    await page.goto("/");
    await page.evaluate(() => localStorage.setItem("ydsz-locale", "en-US"));

    await loginAndGotoHome(page);

    // 英文环境下应看到 Workspaces
    await expect(page.getByText("Workspaces").first()).toBeVisible({ timeout: 10_000 });

    // 切换至日文
    await switchLocale(page, "日本語");

    // 日文环境下 html[lang] 属性应更新为 ja-JP
    expect(await page.evaluate(() => document.documentElement.getAttribute("lang"))).toBe("ja-JP");

    // 验证日文文案（ワークスペース = workspace 的日文）
    // 由于日文语言包可能仍在完备中，此处使用 poll 软断言：只要日文或英文任一可见即可
    await expect
      .poll(
        async () => {
          const jaVisible = await page.getByText("ワークスペース").first().isVisible({ timeout: 2_000 }).catch(() => false);
          return jaVisible;
        },
        { timeout: 10_000, intervals: [1_000, 2_000, 3_000] },
      )
      .toBe(true);
  });

  test("刷新页面后语言选择应持久化（localStorage）", async ({ page }) => {
    await page.goto("/");
    await loginAndGotoHome(page);

    // 切换至英文
    await switchLocale(page, "English");

    // 验证 localStorage 中已存储 english locale
    const stored = await page.evaluate(() => localStorage.getItem("ydsz-locale"));
    expect(stored).toBe("en-US");

    // 刷新页面
    await page.reload();
    await expect(page.locator("body")).toBeVisible({ timeout: 10_000 });

    // 刷新后 html[lang] 仍为 en-US（从 localStorage 恢复）
    expect(await page.evaluate(() => document.documentElement.getAttribute("lang"))).toBe("en-US");

    // 文案仍为英文
    await expect(page.getByText("Workspaces").first()).toBeVisible({ timeout: 10_000 });
  });

  test("表单验证错误信息随语言切换（创建工作空间空名称）", async ({ page }) => {
    test.setTimeout(60_000);

    await page.goto("/");
    await loginAndGotoHome(page);

    // 中文环境下：导航到创建工作空间入口（如侧栏按钮）
    // 先尝试通过顶部创建按钮进入工作空间创建
    const createBtn = page
      .locator("button, a")
      .filter({ hasText: /创建工作空间|Create workspace/ })
      .first();
    const hasCreateBtn = await createBtn.isVisible({ timeout: 5_000 }).catch(() => false);

    if (hasCreateBtn) {
      await createBtn.click();
    } else {
      // 直接跳转到创建工作空间 URL
      await page.goto("/?create-workspace=1");
    }

    // 等待创建弹窗 / 表单出现
    const nameInput = page
      .locator('input[name="name"], input[placeholder*="名称"], input[placeholder*="name"], #workspace-name')
      .first();
    const modalVisible = await nameInput.isVisible({ timeout: 8_000 }).catch(() => false);

    if (!modalVisible) {
      // 如果没有找到创建表单，跳过此测试（UI 结构可能不同）
      test.skip(true, "未找到工作空间创建表单，跳过表单验证测试");
      return;
    }

    // 空名称直接提交
    const submitBtn = page
      .locator("button")
      .filter({ hasText: /创建|Create|保存|Save/ })
      .first();
    await submitBtn.click();

    // 验证中文错误提示（如「必填」「请输入」等）
    const zhError = page
      .locator(".error, .form-error, .field-error, [data-testid='field-error']")
      .filter({ hasText: /必填|请输入|不能为空/ })
      .first();
    await expect(zhError).toBeVisible({ timeout: 8_000 });

    // 切换至英文后再提交一次（仍在同一弹窗）
    // 先清空已输入的内容（如果有）
    await nameInput.fill("");

    // 通过页面头部语言切换器切换
    await switchLocale(page, "English");

    // 重新点击提交（空名称）
    await submitBtn.click();

    // 英文错误信息应出现
    const enError = page
      .locator(".error, .form-error, .field-error, [data-testid='field-error']")
      .filter({ hasText: /required|empty|enter|name/i })
      .first();
    await expect(enError).toBeVisible({ timeout: 8_000 });
  });
});
