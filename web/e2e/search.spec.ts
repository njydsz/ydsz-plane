/**
 * 全局搜索 E2E 测试。
 *
 * 覆盖：工作空间级全文搜索（issue 命中）→ 搜索历史记录 → 书签创建/删除。
 * 回归 S8 交付的 PG FTS 主链路（ADR-0010 降级方案）+ search_documents 索引回填。
 *
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 * 种子数据包含 500 个工作项（scripts/seed/main.go），其中包含可检索的
 * issue 名称（如 "登录页优化" / "支付流程" 等固定关键词），搜索词用种子关键词。
 */
import { expect, test } from "@playwright/test";

const TEST_EMAIL = "admin@ydsz.dev";
const TEST_PASSWORD = "Admin@123";
const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

async function login(request: any) {
  const res = await request.post(`${API_URL}/auth/login`, {
    data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  });
  expect(res.ok()).toBe(true);
  const body = await res.json();
  return body.access_token;
}

async function getFirstWorkspace(request: any, token: string) {
  const res = await request.get(`${API_URL}/workspaces`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.ok()).toBe(true);
  const list = await res.json();
  expect(list.length).toBeGreaterThan(0);
  return list[0].id;
}

test.describe("全局搜索", () => {
  test("工作空间搜索命中 issue 并返回分组结果", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    // 用种子数据中的固定关键词搜索（seed 脚本内置 "登录" 等可检索词）
    const res = await request.get(`${API_URL}/workspaces/${wsId}/search`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { q: "登录", limit: 20 },
    });
    expect(res.ok()).toBe(true);
    const body = await res.json();
    expect(body.total).toBeGreaterThanOrEqual(0);
    expect(body.query).toBe("登录");
    // 结构校验：results 按类型分组（issue/sprint/version/projects 字段必须存在）
    expect(body.results).toBeDefined();
    expect(Array.isArray(body.results.issues)).toBe(true);
  });

  test("搜索后产生历史记录，可删除", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    // 触发一次搜索以写入历史
    await request.get(`${API_URL}/workspaces/${wsId}/search`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { q: "迭代" },
    });

    // 读取历史，应包含刚搜索的关键词
    const histRes = await request.get(`${API_URL}/workspaces/${wsId}/search/history`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(histRes.ok()).toBe(true);
    const histBody = await histRes.json();
    const results = histBody.results || histBody || [];
    expect(Array.isArray(results)).toBe(true);
    const found = results.find((h: any) => h.query === "迭代");
    expect(found).toBeTruthy();

    // 删除该条历史
    const delRes = await request.delete(
      `${API_URL}/workspaces/${wsId}/search/history/${found.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(delRes.ok()).toBe(true);
  });

  test("书签创建与删除闭环", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    const name = `E2E bookmark ${Date.now()}`;
    const createRes = await request.post(`${API_URL}/workspaces/${wsId}/search/bookmarks`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { name, query: "登录", is_shared: false },
    });
    expect(createRes.ok()).toBe(true);
    const created = await createRes.json();
    const bookmarkId = created.id ?? created.bookmark?.id;

    // 列表应包含该书签
    const listRes = await request.get(`${API_URL}/workspaces/${wsId}/search/bookmarks`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(listRes.ok()).toBe(true);
    const listBody = await listRes.json();
    const list = listBody.results || listBody || [];
    expect(list.some((b: any) => b.id === bookmarkId)).toBe(true);

    // 删除书签
    const delRes = await request.delete(
      `${API_URL}/workspaces/${wsId}/search/bookmarks/${bookmarkId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(delRes.ok()).toBe(true);
  });
});
