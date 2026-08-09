import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright E2E 配置。
 *
 * 运行前提：
 *   1. 后端 API 已启动（默认 http://localhost:8080）
 *   2. 前端 dev server 已启动（默认 http://localhost:5173）
 *   3. 已执行 make migrate && make seed（含种子账号 admin@ydsz.dev）
 *
 * 运行：pnpm exec playwright test
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  // 本地默认使用 line reporter（即时输出、无需清理 HTML 报告目录）；
  // CI 仍可通过环境变量覆盖为 html 报告。
  reporter: process.env.CI
    ? [["html", { outputFolder: "playwright-report" }]]
    : "line",
  // 每次运行写入带时间戳的独立输出目录，避免 Playwright 启动期清理
  // 已有 test-results 触发的批量删除拦截（本地环境限制）。
  outputDir: `test-results/run-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}`,
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  // E2E 依赖真实后端，本地环境可能未启动，允许配置超时
  timeout: 60_000,
  expect: { timeout: 15_000 },
});
