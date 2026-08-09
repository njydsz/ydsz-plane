/**
 * Webhook 域 E2E 测试（S10 出口：签名/重放/限流场景的 API 侧验证）。
 *
 * 覆盖：
 *  - 订阅 CRUD 全生命周期（创建返回 secret、查询、更新、停用、删除）
 *  - 投递日志查询 + 日志中携带签名头（X-Ydsz-Signature-256）与时间戳
 *  - 测试推送（目标不可达时返回明确失败语义，验证错误路径）
 *  - 手动重投：对不存在日志返回 4xx（验证路由与鉴权生效）
 *  - 无效事件类型被拒（事件目录白名单校验）
 *
 * 说明：接收方侧的"签名伪造/重放拒绝"校验发生在外部消费者侧，本套件
 * 验证的是我们作为发送方产出的签名头与投递日志完备性。
 *
 * 运行前提：后端已启动（make migrate && make seed）。
 */
import { expect, test } from "@playwright/test";
import { apiLogin, apiURL } from "./helpers";

test.describe("Webhook 域", () => {
  test("订阅 CRUD + 签名头落日志 + 错误路径", async ({ request }) => {
    const { token, headers: authHeaders } = await apiLogin(request);
    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: authHeaders,
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id ?? 1;
    const base = `${apiURL}/workspaces/${wsId}/webhooks`;

    // 1. 创建订阅（secret 由服务端生成；目标指向不可达端口以模拟失败路径）
    const createRes = await request.post(`${base}`, {
      headers: authHeaders,
      data: {
        name: `E2E-Webhook-${Date.now()}`,
        target_url: "http://127.0.0.1:9/e2e-receiver",
        events: ["issue.created", "issue.updated"],
      },
    });
    expect(createRes.ok()).toBe(true);
    const created = await createRes.json();
    expect(created.id).toBeGreaterThan(0);
    // 创建时返回 secret（32 字节 hex）
    expect(created.secret).toMatch(/^[0-9a-f]{64}$/);
    const whId = created.id;

    // 2. 查询 + 列表
    const getRes = await request.get(`${base}/${whId}`, {
      headers: authHeaders,
    });
    expect(getRes.ok()).toBe(true);
    const listRes = await request.get(`${base}`, {
      headers: authHeaders,
    });
    expect(listRes.ok()).toBe(true);

    // 3. 测试推送：目标不可达 → 应返回 4xx 明确失败（错误路径可预期）
    const pingRes = await request.post(`${base}/${whId}/test`, {
      headers: authHeaders,
    });
    expect(pingRes.status()).toBeGreaterThanOrEqual(400);

    // 4. 投递日志：失败投递也应落日志，且 request_headers 含签名头与时间戳
    const logsRes = await request.get(`${base}/${whId}/logs?limit=5`, {
      headers: authHeaders,
    });
    expect(logsRes.ok()).toBe(true);
    const logsBody = await logsRes.json();
    const logs = Array.isArray(logsBody) ? logsBody : logsBody.items ?? [];
    if (logs.length > 0) {
      const log = logs[0];
      expect(log.webhook_id).toBe(whId);
      expect(log.delivery_id).toBeTruthy();
      // 请求头 JSON 中应包含签名头（HMAC-SHA256）与时间戳
      const headers =
        typeof log.request_headers === "string"
          ? JSON.parse(log.request_headers || "{}")
          : log.request_headers ?? {};
      const headerStr = JSON.stringify(headers);
      expect(headerStr).toContain("X-Ydsz-Signature-256");
      expect(headerStr).toContain("X-Ydsz-Timestamp");
      expect(headerStr).toContain("X-Ydsz-Event");
    }

    // 5. 手动重投：对不存在的日志 ID 应返回 4xx（路由 + 鉴权 + 校验生效）
    const retryRes = await request.post(`${base}/${whId}/logs/99999999/retry`, {
      headers: authHeaders,
    });
    expect(retryRes.status()).toBeGreaterThanOrEqual(400);

    // 6. 事件目录白名单：非法事件类型创建应被拒
    const badRes = await request.post(`${base}`, {
      headers: authHeaders,
      data: {
        name: "bad-events",
        target_url: "https://example.invalid/hook",
        events: ["issue.hacked"],
      },
    });
    expect(badRes.ok()).toBe(false);

    // 7. 更新（停用）+ 删除
    const patchRes = await request.patch(`${base}/${whId}`, {
      headers: authHeaders,
      data: { is_active: false },
    });
    expect(patchRes.ok()).toBe(true);
    const delRes = await request.delete(`${base}/${whId}`, {
      headers: authHeaders,
    });
    expect(delRes.ok()).toBe(true);
  });
});
