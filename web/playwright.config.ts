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
  reporter: "html",
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
