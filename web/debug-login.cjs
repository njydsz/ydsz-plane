// 调试脚本：捕获浏览器控制台与网络错误
const { chromium } = require("@playwright/test");

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  page.on("console", (msg) => {
    if (msg.type() === "error" || msg.type() === "warning") {
      console.log(`[console.${msg.type()}]`, msg.text());
    }
  });
  page.on("requestfailed", (req) => {
    console.log("[requestfailed]", req.url(), req.failure()?.errorText);
  });
  page.on("response", (res) => {
    if (res.status() >= 400) {
      console.log("[bad-response]", res.status(), res.url());
    }
  });
  page.on("pageerror", (err) => {
    console.log("[pageerror]", err.message);
  });

  await page.goto("http://localhost:5173/login");
  await page.waitForSelector('input[type="email"]', { timeout: 10000 });
  await page.locator('input[type="email"]').fill("admin@ydsz.dev");
  await page.locator('input[type="password"]').fill("Admin@123");
  await page.locator("button.submit").click();
  await page.waitForTimeout(5000);
  console.log("[final-url]", page.url());

  await browser.close();
})();
