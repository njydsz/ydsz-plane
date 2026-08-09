# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: comment.spec.ts >> 评论域 >> 工作项评论 CRUD 完整旅程
- Location: e2e\comment.spec.ts:14:3

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
  2   |  * 评论域 E2E 测试。
  3   |  *
  4   |  * 覆盖：创建评论、列表查询、编辑评论、删除评论、@mentions 生效。
  5   |  * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
  6   |  */
  7   | import { expect, test } from "@playwright/test";
  8   | 
  9   | // 测试常量
  10  | const TEST_EMAIL = "admin@ydsz.dev";
  11  | const TEST_PASSWORD = "Admin@123";
  12  | 
  13  | test.describe("评论域", () => {
  14  |   test("工作项评论 CRUD 完整旅程", async ({ page }) => {
  15  |     // 1. 登录
  16  |     await page.goto("/login");
> 17  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
      |                                               ^ Error: locator.fill: Test timeout of 60000ms exceeded.
  18  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  19  |     await page.locator("button.submit").click();
  20  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  21  | 
  22  |     // 2. 进入默认工作空间
  23  |     await page.waitForSelector('a[href*="/projects/"]', { timeout: 10_000 });
  24  |     const firstProject = page.locator('a[href*="/projects/"]').first();
  25  |     await firstProject.click();
  26  | 
  27  |     // 3. 进入第一个工作项详情
  28  |     await page.waitForSelector('tr.issue-row, .issue-list-item, a[href*="/issues/"]', { timeout: 10_000 });
  29  |     const firstIssue = page.locator('a[href*="/issues/"]').first();
  30  |     const issueHref = await firstIssue.getAttribute("href");
  31  |     expect(issueHref).toBeTruthy();
  32  |     await firstIssue.click();
  33  | 
  34  |     // 4. 工作项详情页 —— 评论区可见
  35  |     await page.waitForSelector(".comment-list, [data-testid='comment-list'], .comment-panel", { timeout: 10_000 });
  36  | 
  37  |     // 5. 创建评论
  38  |     const commentInput = page.locator(
  39  |       '[data-testid="comment-input"], .comment-form textarea, .comment-form [contenteditable], .rich-text-editor',
  40  |     );
  41  |     await expect(commentInput.first()).toBeVisible({ timeout: 10_000 });
  42  |     const testComment = `E2E test comment ${Date.now()}`;
  43  |     await commentInput.first().fill(testComment);
  44  | 
  45  |     const submitBtn = page.locator(
  46  |       '[data-testid="comment-submit"], .comment-form button[type="submit"], .comment-form .submit-btn',
  47  |     );
  48  |     await submitBtn.first().click();
  49  | 
  50  |     // 6. 新评论应出现在列表中
  51  |     await expect(page.locator(`text=${testComment.slice(0, 30)}`).first()).toBeVisible({ timeout: 10_000 });
  52  | 
  53  |     // 7. 编辑评论：找到刚创建的评论，点击编辑
  54  |     const commentItem = page.locator(".comment-item", { hasText: testComment.slice(0, 30) }).first();
  55  |     const editBtn = commentItem.locator(
  56  |       '[data-testid="comment-edit"], .comment-edit-btn, button[title="编辑"]',
  57  |     );
  58  |     if (await editBtn.isVisible({ timeout: 5_000 }).catch(() => false)) {
  59  |       await editBtn.click();
  60  |       const editInput = page.locator(
  61  |         '[data-testid="comment-edit-input"], .comment-edit-form textarea, .comment-edit-form [contenteditable]',
  62  |       );
  63  |       await editInput.first().fill(`${testComment} (edited)`);
  64  |       const saveBtn = page.locator(
  65  |         '[data-testid="comment-edit-save"], .comment-edit-form button[type="submit"]',
  66  |       );
  67  |       await saveBtn.first().click();
  68  |       await expect(page.locator(`text=${testComment.slice(0, 30)} (edited)`).first()).toBeVisible({ timeout: 10_000 });
  69  |     }
  70  | 
  71  |     // 8. 删除评论
  72  |     const deleteBtn = page.locator(".comment-item", { hasText: testComment.slice(0, 30) }).first()
  73  |       .locator('[data-testid="comment-delete"], .comment-delete-btn, button[title="删除"]');
  74  |     if (await deleteBtn.isVisible({ timeout: 5_000 }).catch(() => false)) {
  75  |       await deleteBtn.click();
  76  |       // 确认对话框（如有）
  77  |       const confirmBtn = page.locator(
  78  |         '[data-testid="confirm-delete"], .confirm-btn:has-text("删除"), button:has-text("确认"), .app-modal button.primary',
  79  |       );
  80  |       if (await confirmBtn.isVisible({ timeout: 3_000 }).catch(() => false)) {
  81  |         await confirmBtn.first().click();
  82  |       }
  83  |       // 评论应被移除
  84  |       await expect(page.locator(`text=${testComment.slice(0, 30)}`).first()).not.toBeVisible({ timeout: 10_000 });
  85  |     }
  86  |   });
  87  | 
  88  |   test("@mention 解析正确", async ({ page }) => {
  89  |     await page.goto("/login");
  90  |     await page.locator('input[type="email"]').fill(TEST_EMAIL);
  91  |     await page.locator('input[type="password"]').fill(TEST_PASSWORD);
  92  |     await page.locator("button.submit").click();
  93  |     await expect(page).not.toHaveURL(/\/login/, { timeout: 15_000 });
  94  | 
  95  |     // 进入第一个工作项
  96  |     await page.waitForSelector('a[href*="/projects/"]', { timeout: 10_000 });
  97  |     await page.locator('a[href*="/projects/"]').first().click();
  98  |     await page.waitForSelector('a[href*="/issues/"]', { timeout: 10_000 });
  99  |     await page.locator('a[href*="/issues/"]').first().click();
  100 | 
  101 |     // 使用 @ 触发 mention
  102 |     const commentInput = page.locator(
  103 |       '[data-testid="comment-input"], .comment-form textarea, .comment-form [contenteditable]',
  104 |     );
  105 |     await expect(commentInput.first()).toBeVisible({ timeout: 10_000 });
  106 |     await commentInput.first().fill("@admin mention test");
  107 | 
  108 |     // mention 下拉出现（可选 UI）
  109 |     const mentionDropdown = page.locator(
  110 |       '[data-testid="mention-dropdown"], .mention-list, .tippy-box, .suggestion-list',
  111 |     );
  112 |     // mention 是高阶功能，断言不强制要求，但至少评论能提交
  113 |     const submitBtn = page.locator(
  114 |       '[data-testid="comment-submit"], .comment-form button[type="submit"]',
  115 |     );
  116 |     await submitBtn.first().click();
  117 |     await expect(page.locator("text=@admin mention test").first()).toBeVisible({ timeout: 10_000 });
```