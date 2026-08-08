/**
 * 工作空间（Workspace）+ 成员 + 邀请 + 项目 E2E 测试。
 *
 * 覆盖：空间列表/详情/更新、成员列表/角色切换/移除、邀请发送/列表/撤销、
 *       RBAC 角色查询、项目 CRUD。
 * 回归 S2 IAM 与工作空间域。
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
  return list[0];
}

test.describe("工作空间域", () => {
  test("列表 → 详情 → 更新基本信息", async ({ request }) => {
    const token = await login(request);

    // 列表
    const listRes = await request.get(`${API_URL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(listRes.ok()).toBe(true);
    const list = await listRes.json();
    expect(list.length).toBeGreaterThanOrEqual(1);

    // 详情
    const wsId = list[0].id;
    const detailRes = await request.get(`${API_URL}/workspaces/${wsId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(detailRes.ok()).toBe(true);
    const ws = await detailRes.json();
    expect(ws.id).toBe(wsId);
    expect(ws.slug).toBeTruthy();

    // 更新（timezone / language）
    const updateRes = await request.patch(`${API_URL}/workspaces/${wsId}`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { timezone: "Asia/Shanghai", language: "zh-CN" },
    });
    expect(updateRes.ok()).toBe(true);
    const updated = await updateRes.json();
    expect(updated.timezone).toBe("Asia/Shanghai");
    expect(updated.language).toBe("zh-CN");
  });

  test("slug 唯一性校验", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 尝试创建与已有空间同 slug 的空间，应返回 409
    const createRes = await request.post(`${API_URL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { name: "Dup Slug Test", slug: ws.slug },
    });
    // slug 唯一性冲突应为 409
    expect(createRes.status()).toBe(409);
  });

  test("当前用户角色 + 权限查询", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    const roleRes = await request.get(`${API_URL}/workspaces/${ws.id}/role`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(roleRes.ok()).toBe(true);
    const roleData = await roleRes.json();
    expect(roleData.role).toBeTruthy();
    expect(roleData.role.slug).toBeTruthy();
    expect(Array.isArray(roleData.permissions)).toBe(true);
  });

  test("角色定义列表", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    const rolesRes = await request.get(`${API_URL}/workspaces/${ws.id}/roles`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(rolesRes.ok()).toBe(true);
    const roles = await rolesRes.json();
    expect(roles.length).toBeGreaterThanOrEqual(4); // Owner/Admin/Member/Guest
  });
});

test.describe("成员管理", () => {
  test("成员列表包含 admin 自身", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    const membersRes = await request.get(`${API_URL}/workspaces/${ws.id}/members`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(membersRes.ok()).toBe(true);
    const members = await membersRes.json();
    expect(members.length).toBeGreaterThanOrEqual(1);
    const adminMember = members.find((m: any) => m.email === TEST_EMAIL);
    expect(adminMember).toBeTruthy();
  });

  test("成员切换角色（admin 不能降级自己，应 403 或 400）", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 获取 admin 自身 ID
    const meRes = await request.get(`${API_URL}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(meRes.ok()).toBe(true);
    const me = await meRes.json();

    // 尝试把自己降级为 guest（应被拒绝）
    const changeRes = await request.patch(
      `${API_URL}/workspaces/${ws.id}/members/${me.id}`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: { role: "guest" },
      },
    );
    // Owner 自降应被拒绝（403/400/422 取决于后端实现）
    expect([400, 403, 422]).toContain(changeRes.status());
  });

  test("Owner 不可被移除（移除自己应失败）", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    const meRes = await request.get(`${API_URL}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const me = await meRes.json();

    const removeRes = await request.delete(
      `${API_URL}/workspaces/${ws.id}/members/${me.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    // Owner 不能移除自己（最后一个 owner）
    expect([400, 403, 422]).toContain(removeRes.status());
  });
});

test.describe("邀请域", () => {
  test("发送邀请 → 列表查询 → 撤销", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 发送邀请
    const inviteEmail = `invite-${Date.now()}@example.com`;
    const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { email: inviteEmail, role: "member", message: "E2E test invitation" },
    });
    expect(inviteRes.status()).toBe(201);
    const invitation = await inviteRes.json();
    expect(invitation.id).toBeGreaterThan(0);
    expect(invitation.status).toBe("pending");
    expect(invitation.email).toBe(inviteEmail);

    // 列出邀请
    const listRes = await request.get(`${API_URL}/workspaces/${ws.id}/invitations?status=pending`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(listRes.ok()).toBe(true);
    const invitations = await listRes.json();
    const found = invitations.find((inv: any) => inv.id === invitation.id);
    expect(found).toBeTruthy();

    // 撤销邀请
    const revokeRes = await request.delete(
      `${API_URL}/workspaces/${ws.id}/invitations/${invitation.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(revokeRes.ok()).toBe(true);
  });

  test("重复邀请已存在的成员应失败", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 邀请已存在的 admin 自身
    const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { email: TEST_EMAIL, role: "member" },
    });
    // 已存在成员应返回 409 或 400
    expect([400, 409, 422]).toContain(inviteRes.status());
  });

  test("邀请预览接口（无鉴权）返回邀请详情", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 发送邀请获取 token
    const inviteEmail = `preview-${Date.now()}@example.com`;
    const inviteRes = await request.post(`${API_URL}/workspaces/${ws.id}/invitations`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { email: inviteEmail, role: "member" },
    });
    expect(inviteRes.status()).toBe(201);
    const invitation = await inviteRes.json();

    // 注意：后端的邀请 preview 需要 token（邀请 token），不是 invitation.id
    // 这里仅验证接口存在且返回结构合理
    // 实际 preview endpoint: GET /invitations/:token
    // invitation.token 可能在返回中，取决于后端实现
    if (invitation.token) {
      const previewRes = await request.get(`${API_URL}/invitations/${invitation.token}`);
      expect(previewRes.ok()).toBe(true);
      const preview = await previewRes.json();
      expect(preview.workspace_id).toBe(ws.id);
      expect(preview.email).toBe(inviteEmail);
    }

    // 清理
    await request.delete(`${API_URL}/workspaces/${ws.id}/invitations/${invitation.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  });
});

test.describe("项目域", () => {
  test("工作空间下项目列表 + 创建项目", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    // 列表
    const listRes = await request.get(`${API_URL}/workspaces/${ws.id}/projects`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(listRes.ok()).toBe(true);
    const body = await listRes.json();
    const projects = body.results || body;
    expect(Array.isArray(projects)).toBe(true);

    // 创建项目
    const identifier = `E2E${Date.now().toString().slice(-6)}`;
    const createRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: {
        name: `E2E Project ${Date.now()}`,
        identifier,
        description: "E2E test project",
        network: "public",
      },
    });
    expect(createRes.status()).toBe(201);
    const project = await createRes.json();
    expect(project.id).toBeGreaterThan(0);
    expect(project.identifier).toBe(identifier);
    expect(project.name).toContain("E2E");

    // 清理：归档项目
    const archiveRes = await request.delete(
      `${API_URL}/workspaces/${ws.id}/projects/${project.id}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(archiveRes.ok()).toBe(true);
  });

  test("项目 identifier 唯一性（同空间下重复应失败）", async ({ request }) => {
    const token = await login(request);
    const ws = await getFirstWorkspace(request, token);

    const identifier = `E2E${Date.now().toString().slice(-6)}`;
    // 先创建
    const firstRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { name: "First", identifier },
    });
    expect(firstRes.status()).toBe(201);
    const projectId = (await firstRes.json()).id;

    // 重复 identifier
    const dupRes = await request.post(`${API_URL}/workspaces/${ws.id}/projects`, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      data: { name: "Duplicate", identifier },
    });
    expect([400, 409, 422]).toContain(dupRes.status());

    // 清理
    await request.delete(`${API_URL}/workspaces/${ws.id}/projects/${projectId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  });
});
