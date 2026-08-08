/**
 * 收件箱（Intake）域 E2E 测试（S10 出口：门户提交→审核→转正→跟踪闭环）。
 *
 * 覆盖完整业务闭环：
 *  1. 管理员创建公开频道（slug 唯一）
 *  2. 外部用户免登录经公开门户提交工单 → 获得 tracking_id
 *  3. 管理员审核（accepted）→ 转正（convert）→ 生成正式工作项并双向关联
 *  4. 提交者凭 tracking_id + email 跟踪处理状态（脱敏只读视图）
 *  5. 拒绝路径：rejected 后状态同步
 *
 * 全链路 API 驱动（公开接口免登录、管理接口走 Bearer），不依赖 UI 渲染。
 * 运行前提：后端已启动（make migrate && make seed）。
 */
import { expect, test } from "@playwright/test";

const TEST_EMAIL = "admin@ydsz.dev";
const TEST_PASSWORD = "Admin@123";
const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

async function login(request: any) {
  const res = await request.post(`${apiURL}/auth/login`, {
    data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  });
  expect(res.ok()).toBe(true);
  const { access_token: token } = await res.json();
  return token as string;
}

test.describe("收件箱（Intake）域", () => {
  test("公开提交 → 审核 → 转正 → 跟踪闭环", async ({ request }) => {
    const token = await login(request);
    const wsRes = await request.get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(wsRes.ok()).toBe(true);
    const wsList = await wsRes.json();
    const wsId = wsList[0]?.id ?? 1;
    const wsSlug = wsList[0]?.slug ?? "acme";

    // 取项目（转正目标项目）
    const projRes = await request.get(
      `${apiURL}/workspaces/${wsId}/projects?limit=5`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    const projList = projRes.ok() ? await projRes.json() : { items: [] };
    const items = Array.isArray(projList) ? projList : projList.items ?? [];
    const projectId = items[0]?.id ?? 1;

    const adminBase = `${apiURL}/workspaces/${wsId}/intake`;
    const slug = `e2e-${Date.now().toString(36)}`;
    const stamp = Date.now();

    // 1. 创建公开频道
    const chRes = await request.post(`${adminBase}/channels`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        slug,
        name: `E2E 频道 ${stamp}`,
        description: "Playwright 创建",
        is_public: true,
        default_issue_type: "requirement",
        rate_limit_per_min: 50,
        require_captcha: false,
      },
    });
    expect(chRes.ok()).toBe(true);
    const channel = await chRes.json();
    expect(channel.id).toBeGreaterThan(0);

    // 2. 外部用户免登录公开提交
    const submitRes = await request.post(
      `${apiURL}/public/intake/channels/${wsSlug}/${slug}/submit`,
      {
        data: {
          title: `E2E 门户工单 ${stamp}`,
          description: "来自公开门户的缺陷/需求描述",
          submitter_name: "张三",
          submitter_email: "zhangsan@example.com",
          issue_type: "requirement",
        },
      },
    );
    expect(submitRes.ok()).toBe(true);
    const submitted = await submitRes.json();
    expect(submitted.tracking_id).toBeTruthy();
    expect(submitted.status).toBe("open");
    const trackingId = submitted.tracking_id;

    // 3. 管理员查询收件队列并定位工单
    const queueRes = await request.get(`${adminBase}/issues?status=open&limit=20`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(queueRes.ok()).toBe(true);
    const queueBody = await queueRes.json();
    const queue = Array.isArray(queueBody) ? queueBody : queueBody.items ?? [];
    const ticket = queue.find((t: any) => t.tracking_id === trackingId);
    expect(ticket).toBeTruthy();
    const ticketId = ticket.id;

    // 4. 审核通过（accepted）
    const reviewRes = await request.post(`${adminBase}/issues/${ticketId}/review`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        action: "accepted",
        target_issue_type: "requirement",
        target_project_id: projectId,
      },
    });
    expect(reviewRes.ok()).toBe(true);

    // 5. 转正 → 生成正式工作项并双向关联
    const convertRes = await request.post(`${adminBase}/issues/${ticketId}/convert`, {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        target_project_id: projectId,
        target_issue_type: "requirement",
      },
    });
    expect(convertRes.ok()).toBe(true);
    const converted = await convertRes.json();
    expect(converted.converted_issue_id).toBeGreaterThan(0);
    const convertedIssueId = converted.converted_issue_id;

    // 6. 提交者跟踪门户（tracking_id + email）→ 应展示脱敏只读视图
    const trackRes = await request.get(
      `${apiURL}/public/intake/track?tracking_id=${trackingId}&email=zhangsan@example.com`,
    );
    expect(trackRes.ok()).toBe(true);
    const view = await trackRes.json();
    expect(view.status).toMatch(/accepted|converted/);
    expect(view.tracking_id).toBe(trackingId);
    // 脱敏：不应泄露内部字段（submitter_user_id 等）
    expect(view.submitter_user_id).toBeUndefined();

    // 7. 转正后正式工作项存在（管理员视角）
    const issueRes = await request.get(
      `${apiURL}/workspaces/${wsId}/projects/${projectId}/issues/${convertedIssueId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(issueRes.ok()).toBe(true);

    // 8. 拒绝路径：新建一条并拒绝
    const submit2 = await request.post(
      `${apiURL}/public/intake/channels/${wsSlug}/${slug}/submit`,
      {
        data: {
          title: `E2E 拒绝工单 ${stamp}`,
          submitter_name: "李四",
          submitter_email: "lisi@example.com",
        },
      },
    );
    expect(submit2.ok()).toBe(true);
    const t2 = await submit2.json();
    const q2Res = await request.get(`${adminBase}/issues?status=open&limit=50`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const q2Body = await q2Res.json();
    const q2 = Array.isArray(q2Body) ? q2Body : q2Body.items ?? [];
    const t2row = q2.find((t: any) => t.tracking_id === t2.tracking_id);
    expect(t2row).toBeTruthy();
    const rejectRes = await request.post(`${adminBase}/issues/${t2row.id}/review`, {
      headers: { Authorization: `Bearer ${token}` },
      data: { action: "rejected", reason: "重复提交，无需处理" },
    });
    expect(rejectRes.ok()).toBe(true);

    // 清理频道（软删）
    await request.delete(`${adminBase}/channels/${channel.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  });
});
