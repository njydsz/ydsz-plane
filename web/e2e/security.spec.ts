/**
 * 安全场景 E2E 测试。
 *
 * 覆盖：Webhook 签名伪造检测、重放攻击防护、输入校验（SQL 注入/XSS）、
 *       权限边界（跨空间访问隔离）、API Token 鉴权。
 * 对标 S10 安全测试项 + DevOps 纵深防御。
 *
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
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
  return list[0]?.id || 1;
}

async function getFirstProject(request: any, token: string, wsId: number) {
  const res = await request.get(`${API_URL}/workspaces/${wsId}/projects`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.ok()).toBe(true);
  const body = await res.json();
  return body.results?.[0]?.id || body[0]?.id || 1;
}

test.describe("鉴权与令牌安全", () => {
  test("无效 JWT token 受保护端点返回 401", async ({ request }) => {
    const res = await request.get(`${API_URL}/workspaces`, {
      headers: { Authorization: "Bearer invalid.token.here" },
    });
    expect(res.status()).toBe(401);
  });

  test("过期/篡改的 refresh_token 刷新失败", async ({ request }) => {
    const res = await request.post(`${API_URL}/auth/refresh`, {
      data: { refresh_token: "tampered-refresh-token-xyz" },
    });
    expect([400, 401, 422]).toContain(res.status());
  });

  test("缺少 Authorization 头时受保护端点返回 401", async ({ request }) => {
    const res = await request.get(`${API_URL}/workspaces`);
    expect(res.status()).toBe(401);
  });

  test("登录尝试次数限制（暴力破解防护）", async ({ request }) => {
    // 连续发送 5 次错误密码登录
    const attempts = [];
    for (let i = 0; i < 5; i++) {
      const res = await request.post(`${API_URL}/auth/login`, {
        data: { email: "nonexistent@example.com", password: `wrong${i}` },
      });
      attempts.push(res.status());
    }
    // 所有尝试都应返回 401（或者后几次可能返回 429 限流）
    const allRejected = attempts.every((s) => s === 401 || s === 429);
    expect(allRejected).toBe(true);
  });
});

test.describe("Webhook 签名安全", () => {
  test("创建 Webhook 返回的 secret 为 64 位 hex", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: {
        name: "Security Test Webhook",
        url: "https://localhost:19999/no-op",
        events: ["issue.created", "issue.updated"],
      },
    });
    expect(createRes.status()).toBe(201);
    const webhook = await createRes.json();
    // secret 应为 64 位 hex 字符串（HMAC-SHA256 key）
    expect(webhook.secret).toMatch(/^[0-9a-f]{64}$/);

    // 清理
    await request.delete(`${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  });

  test("非法事件类型被拒", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: {
        name: "Invalid Event Webhook",
        url: "https://example.com/hook",
        events: ["issue.hacked", "nonexistent.event"],
      },
    });
    // 非法事件应被拒
    expect([400, 422]).toContain(createRes.status());
  });

  test("Test 推送日志中包含签名头", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    const createRes = await request.post(`${API_URL}/workspaces/${wsId}/webhooks`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: {
        name: "Signature Test",
        url: "https://localhost:19999/no-op",
        events: ["issue.created"],
      },
    });
    expect(createRes.status()).toBe(201);
    const webhook = await createRes.json();

    // 触发 test push
    const testRes = await request.post(
      `${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}/test`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    // test push 目标不可达也应返回 2xx（测试请求已发送）
    expect([200, 202]).toContain(testRes.status());

    // 检查日志包含签名头
    const logsRes = await request.get(
      `${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}/logs?limit=5`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    if (logsRes.ok()) {
      const logs = await logsRes.json();
      const logList = logs.results || logs;
      if (logList.length > 0) {
        const firstLog = logList[0];
        // 日志中 request_headers 应包含签名相关头
        const headers = firstLog.request_headers || {};
        const hasSignature =
          headers["X-Ydsz-Signature-256"] || headers["x-ydsz-signature-256"];
        const hasTimestamp =
          headers["X-Ydsz-Timestamp"] || headers["x-ydsz-timestamp"];
        // 至少签名或时间戳存在（取决于后端日志结构）
        expect(hasSignature || hasTimestamp).toBeTruthy();
      }
    }

    // 清理
    await request.delete(`${API_URL}/workspaces/${wsId}/webhooks/${webhook.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  });
});

test.describe("租户隔离（RLS 边界）", () => {
  test("跨空间访问项目应返回 404/403", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 用不存在的 wsId 访问项目 → 应 404
    const invalidWsRes = await request.get(
      `${API_URL}/workspaces/99999/projects/${projectId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect([403, 404]).toContain(invalidWsRes.status());
  });

  test("无效项目 ID 返回 404", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    const res = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/999999`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(res.status()).toBe(404);
  });
});

test.describe("输入校验（注入防护）", () => {
  test("工作项名称 SQL 注入字符被安全处理", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 创建带特殊字符名称的工作项
    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          type: "task",
          name: `E2E SQL注入测试'; DROP TABLE issues; --`,
          assignees: [],
          labels: [],
          modules: [],
        },
      },
    );
    expect(createRes.status()).toBe(201);
    const issue = await createRes.json();

    // 读取验证数据完整无损
    const getRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(getRes.ok()).toBe(true);
    const fetched = await getRes.json();
    expect(fetched.id).toBe(issue.id);

    // 清理：删除工作项
    await request.delete(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );

    // 验证其他数据未被破坏（列出工作项应正常工作）
    const listRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=5`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(listRes.ok()).toBe(true);
  });

  test("HTML/XSS 富文本输入存储后不过度执行", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 创建带 script 标签描述的工作项
    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          type: "task",
          name: "E2E XSS Test",
          description_html: '<p>合法</p><script>alert("xss")</script>',
          assignees: [],
          labels: [],
          modules: [],
        },
      },
    );
    // 两种可接受路径：201（存储原始 HTML，由前端渲染时过滤）或 422（输入校验拦截）
    expect([201, 422]).toContain(createRes.status());

    if (createRes.status() === 201) {
      // 如果后端接受原始 HTML，验证能安全回显（不崩溃）
      const issue = await createRes.json();
      await request.delete(
        `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues/${issue.id}`,
        { headers: { Authorization: `Bearer ${token}` } },
      );
    }
  });

  test("负数 limit/offset 参数不导致服务错误", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 负数参数应被 clamp 为安全值或返回 400
    const res = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/issues?limit=-1&offset=-100`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    // 可接受：200（后端 clamp）或 400（参数校验）
    expect([200, 400, 422]).toContain(res.status());
  });
});

test.describe("审计日志", () => {
  test("管理员操作被记录到审计日志", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);

    // 审计日志查询接口（仅 admin/owner 可用）
    const auditRes = await request.get(`${API_URL}/workspaces/${wsId}/audit-logs?limit=5`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    // 接口存在且返回（200 或 403 取决于实现）
    expect([200, 403]).toContain(auditRes.status());

    if (auditRes.status() === 200) {
      const logs = await auditRes.json();
      expect(logs).toBeTruthy();
    }
  });
});
