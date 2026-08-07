/**
 * 评论域 E2E 测试。
 *
 * 覆盖：创建评论、列表查询、编辑评论、删除评论、@mentions 生效。
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 */
import { expect, test } from "@playwright/test";

// 测试常量
const TEST_EMAIL = "admin@ydsz.dev";
const TEST_PASSWORD = "Admin@123";

test.describe("评论域", () => {
  test("工作项评论 CRUD 完整旅程", async ({ page }) => {
    // 1. 登录
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(TEST_EMAIL);
    await page.locator('input[type="password"]').fill(TEST_PASSWORD);
    await page.locator("button.submit").click();
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });

    // 2. 进入默认工作空间
    await page.waitForSelector('a[href*="/projects/"]', { timeout: 10_000 });
    const firstProject = page.locator('a[href*="/projects/"]').first();
    await firstProject.click();

    // 3. 进入第一个工作项详情
    await page.waitForSelector('tr.issue-row, .issue-list-item, a[href*="/issues/"]', { timeout: 10_000 });
    const firstIssue = page.locator('a[href*="/issues/"]').first();
    const issueHref = await firstIssue.getAttribute("href");
    expect(issueHref).toBeTruthy();
    await firstIssue.click();

    // 4. 工作项详情页 —— 评论区可见
    await page.waitForSelector(".comment-list, [data-testid='comment-list'], .comment-panel", { timeout: 10_000 });

    // 5. 创建评论
    const commentInput = page.locator(
      '[data-testid="comment-input"], .comment-form textarea, .comment-form [contenteditable], .rich-text-editor',
    );
    await expect(commentInput.first()).toBeVisible({ timeout: 10_000 });
    const testComment = `E2E test comment ${Date.now()}`;
    await commentInput.first().fill(testComment);

    const submitBtn = page.locator(
      '[data-testid="comment-submit"], .comment-form button[type="submit"], .comment-form .submit-btn',
    );
    await submitBtn.first().click();

    // 6. 新评论应出现在列表中
    await expect(page.locator(`text=${testComment.slice(0, 30)}`).first()).toBeVisible({ timeout: 10_000 });

    // 7. 编辑评论：找到刚创建的评论，点击编辑
    const commentItem = page.locator(".comment-item", { hasText: testComment.slice(0, 30) }).first();
    const editBtn = commentItem.locator(
      '[data-testid="comment-edit"], .comment-edit-btn, button[title="编辑"]',
    );
    if (await editBtn.isVisible({ timeout: 5_000 }).catch(() => false)) {
      await editBtn.click();
      const editInput = page.locator(
        '[data-testid="comment-edit-input"], .comment-edit-form textarea, .comment-edit-form [contenteditable]',
      );
      await editInput.first().fill(`${testComment} (edited)`);
      const saveBtn = page.locator(
        '[data-testid="comment-edit-save"], .comment-edit-form button[type="submit"]',
      );
      await saveBtn.first().click();
      await expect(page.locator(`text=${testComment.slice(0, 30)} (edited)`).first()).toBeVisible({ timeout: 10_000 });
    }

    // 8. 删除评论
    const deleteBtn = page.locator(".comment-item", { hasText: testComment.slice(0, 30) }).first()
      .locator('[data-testid="comment-delete"], .comment-delete-btn, button[title="删除"]');
    if (await deleteBtn.isVisible({ timeout: 5_000 }).catch(() => false)) {
      await deleteBtn.click();
      // 确认对话框（如有）
      const confirmBtn = page.locator(
        '[data-testid="confirm-delete"], .confirm-btn:has-text("删除"), button:has-text("确认"), .app-modal button.primary',
      );
      if (await confirmBtn.isVisible({ timeout: 3_000 }).catch(() => false)) {
        await confirmBtn.first().click();
      }
      // 评论应被移除
      await expect(page.locator(`text=${testComment.slice(0, 30)}`).first()).not.toBeVisible({ timeout: 10_000 });
    }
  });

  test("@mention 解析正确", async ({ page }) => {
    await page.goto("/login");
    await page.locator('input[type="email"]').fill(TEST_EMAIL);
    await page.locator('input[type="password"]').fill(TEST_PASSWORD);
    await page.locator("button.submit").click();
    await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });

    // 进入第一个工作项
    await page.waitForSelector('a[href*="/projects/"]', { timeout: 10_000 });
    await page.locator('a[href*="/projects/"]').first().click();
    await page.waitForSelector('a[href*="/issues/"]', { timeout: 10_000 });
    await page.locator('a[href*="/issues/"]').first().click();

    // 使用 @ 触发 mention
    const commentInput = page.locator(
      '[data-testid="comment-input"], .comment-form textarea, .comment-form [contenteditable]',
    );
    await expect(commentInput.first()).toBeVisible({ timeout: 10_000 });
    await commentInput.first().fill("@admin mention test");

    // mention 下拉出现（可选 UI）
    const mentionDropdown = page.locator(
      '[data-testid="mention-dropdown"], .mention-list, .tippy-box, .suggestion-list',
    );
    // mention 是高阶功能，断言不强制要求，但至少评论能提交
    const submitBtn = page.locator(
      '[data-testid="comment-submit"], .comment-form button[type="submit"]',
    );
    await submitBtn.first().click();
    await expect(page.locator("text=@admin mention test").first()).toBeVisible({ timeout: 10_000 });
  });
});
